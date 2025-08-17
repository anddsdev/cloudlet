package services

import (
	"fmt"
	"mime/multipart"
	"sync"
	"time"

	"github.com/anddsdev/cloudlet/config"
	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/transaction"
)

// MultipleUploadService handles multiple file uploads with hybrid strategy
type MultipleUploadService struct {
	fileService *FileService
	validator   *MultipleUploadValidator
	cfg         *config.Config
}

// NewMultipleUploadService creates a new multiple upload service
func NewMultipleUploadService(fileService *FileService, cfg *config.Config, storagePath string) *MultipleUploadService {
	return &MultipleUploadService{
		fileService: fileService,
		validator:   NewMultipleUploadValidator(cfg, storagePath),
		cfg:         cfg,
	}
}

// UploadMultipleFiles implements hybrid strategy: validate all first, then process sequentially with atomic transactions
func (s *MultipleUploadService) UploadMultipleFiles(files []*multipart.FileHeader, targetPath string) *models.MultipleUploadResponse {
	startTime := time.Now()
	
	response := &models.MultipleUploadResponse{
		TotalFiles:       len(files),
		SuccessfulFiles:  0,
		FailedFiles:      0,
		TotalSize:        0,
		ProcessedSize:    0,
		Files:            make([]models.FileUploadResult, 0, len(files)),
		Strategy:         "hybrid",
		ProcessingTimeMs: 0,
	}

	// Phase 1: Comprehensive validation of all files
	validation := s.validator.ValidateMultipleUpload(files, targetPath)
	response.TotalSize = validation.TotalSize

	if !validation.Valid {
		response.Success = false
		response.Message = s.validator.GenerateUploadSummary(validation)
		response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
		
		// Create failed results for all files
		for i, file := range files {
			result := models.FileUploadResult{
				Filename:     file.Filename,
				OriginalName: file.Filename,
				Size:         file.Size,
				Path:         targetPath,
				Success:      false,
				Error:        "Validation failed",
				Index:        i,
			}
			response.Files = append(response.Files, result)
		}
		response.FailedFiles = len(files)
		return response
	}

	// Phase 2: Sequential processing with transaction management
	if s.cfg.Server.Upload.CleanupOnFailure {
		return s.processWithFullRollback(files, targetPath, response, startTime)
	} else {
		return s.processWithBestEffort(files, targetPath, response, startTime)
	}
}

// processWithFullRollback processes files with complete rollback on any failure
func (s *MultipleUploadService) processWithFullRollback(files []*multipart.FileHeader, targetPath string, response *models.MultipleUploadResponse, startTime time.Time) *models.MultipleUploadResponse {
	// Create transaction manager for the entire batch
	tm := transaction.NewTransactionManager()
	uploadedFiles := make([]string, 0, len(files))

	// Process each file and add to transaction
	for i, file := range files {
		fileResult := s.processFileWithTransaction(file, targetPath, i, tm, &uploadedFiles)
		response.Files = append(response.Files, fileResult)
		
		if fileResult.Success {
			response.ProcessedSize += fileResult.Size
		}
	}

	// Execute all operations atomically
	if err := tm.Execute(); err != nil {
		response.Success = false
		response.Message = fmt.Sprintf("Batch upload failed: %v", err)
		response.FailedFiles = len(files)
		
		// Mark all files as failed since we're rolling back
		for i := range response.Files {
			response.Files[i].Success = false
			response.Files[i].Error = "Batch rollback: " + err.Error()
		}
	} else {
		response.Success = true
		response.SuccessfulFiles = len(files)
		response.Message = fmt.Sprintf("Successfully uploaded %d files", len(files))
	}

	response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	return response
}

// processWithBestEffort processes files individually, allowing partial success
func (s *MultipleUploadService) processWithBestEffort(files []*multipart.FileHeader, targetPath string, response *models.MultipleUploadResponse, startTime time.Time) *models.MultipleUploadResponse {
	// Determine processing strategy based on file sizes
	largeFiles := 0
	for _, file := range files {
		if file.Size > s.cfg.Server.Upload.StreamingThreshold {
			largeFiles++
		}
	}

	// Use concurrent processing for many small files, sequential for large files
	if largeFiles == 0 && len(files) > 5 && s.cfg.Server.Upload.MaxConcurrentUploads > 1 {
		return s.processConcurrently(files, targetPath, response, startTime)
	} else {
		return s.processSequentially(files, targetPath, response, startTime)
	}
}

// processSequentially processes files one by one
func (s *MultipleUploadService) processSequentially(files []*multipart.FileHeader, targetPath string, response *models.MultipleUploadResponse, startTime time.Time) *models.MultipleUploadResponse {
	for i, file := range files {
		result := s.processIndividualFile(file, targetPath, i)
		response.Files = append(response.Files, result)
		
		if result.Success {
			response.SuccessfulFiles++
			response.ProcessedSize += result.Size
		} else {
			response.FailedFiles++
		}
	}

	response.Success = response.SuccessfulFiles > 0
	if response.SuccessfulFiles == len(files) {
		response.Message = fmt.Sprintf("Successfully uploaded all %d files", len(files))
	} else if response.SuccessfulFiles > 0 {
		response.Message = fmt.Sprintf("Partial success: %d/%d files uploaded", response.SuccessfulFiles, len(files))
	} else {
		response.Message = "All uploads failed"
	}

	response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	return response
}

// processConcurrently processes files using worker pool pattern
func (s *MultipleUploadService) processConcurrently(files []*multipart.FileHeader, targetPath string, response *models.MultipleUploadResponse, startTime time.Time) *models.MultipleUploadResponse {
	maxWorkers := s.cfg.Server.Upload.MaxConcurrentUploads
	if maxWorkers <= 0 {
		maxWorkers = 3 // Default
	}

	// Create channels for work distribution
	fileJobs := make(chan fileJob, len(files))
	results := make(chan models.FileUploadResult, len(files))

	// Worker pool
	var wg sync.WaitGroup
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go s.uploadWorker(&wg, fileJobs, results, targetPath)
	}

	// Send jobs
	for i, file := range files {
		fileJobs <- fileJob{file: file, index: i}
	}
	close(fileJobs)

	// Wait for workers and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	resultSlice := make([]models.FileUploadResult, len(files))
	for result := range results {
		resultSlice[result.Index] = result
		
		if result.Success {
			response.SuccessfulFiles++
			response.ProcessedSize += result.Size
		} else {
			response.FailedFiles++
		}
	}

	response.Files = resultSlice
	response.Success = response.SuccessfulFiles > 0
	
	if response.SuccessfulFiles == len(files) {
		response.Message = fmt.Sprintf("Successfully uploaded all %d files concurrently", len(files))
	} else if response.SuccessfulFiles > 0 {
		response.Message = fmt.Sprintf("Concurrent upload: %d/%d files succeeded", response.SuccessfulFiles, len(files))
	} else {
		response.Message = "All concurrent uploads failed"
	}

	response.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	return response
}

// fileJob represents a file upload job for worker pool
type fileJob struct {
	file  *multipart.FileHeader
	index int
}

// uploadWorker processes file upload jobs
func (s *MultipleUploadService) uploadWorker(wg *sync.WaitGroup, jobs <-chan fileJob, results chan<- models.FileUploadResult, targetPath string) {
	defer wg.Done()
	
	for job := range jobs {
		result := s.processIndividualFile(job.file, targetPath, job.index)
		results <- result
	}
}

// processIndividualFile processes a single file upload
func (s *MultipleUploadService) processIndividualFile(file *multipart.FileHeader, targetPath string, index int) models.FileUploadResult {
	result := models.FileUploadResult{
		Filename:     file.Filename,
		OriginalName: file.Filename,
		Size:         file.Size,
		Path:         targetPath,
		Success:      false,
		Index:        index,
	}

	// Open the file
	f, err := file.Open()
	if err != nil {
		result.Error = fmt.Sprintf("Cannot open file: %v", err)
		return result
	}
	defer f.Close()

	// Determine upload method based on file size
	if file.Size > s.cfg.Server.Upload.StreamingThreshold {
		// Use streaming for large files
		err = s.fileService.SaveFileStream(file.Filename, targetPath, f, file.Size)
	} else {
		// Use traditional upload for small files
		data := make([]byte, file.Size)
		_, err = f.Read(data)
		if err != nil {
			result.Error = fmt.Sprintf("Cannot read file: %v", err)
			return result
		}
		
		err = s.fileService.SaveFile(file.Filename, targetPath, data)
	}

	if err != nil {
		result.Error = fmt.Sprintf("Upload failed: %v", err)
		return result
	}

	// Success
	result.Success = true
	result.MimeType = s.fileService.detectMimeType(file.Filename)
	return result
}

// processFileWithTransaction processes a file and adds operations to transaction manager
func (s *MultipleUploadService) processFileWithTransaction(file *multipart.FileHeader, targetPath string, index int, tm *transaction.TransactionManager, uploadedFiles *[]string) models.FileUploadResult {
	result := models.FileUploadResult{
		Filename:     file.Filename,
		OriginalName: file.Filename,
		Size:         file.Size,
		Path:         targetPath,
		Success:      false,
		Index:        index,
	}

	// Validate path for this specific file
	validatedPath, err := s.fileService.pathValidator.ValidateAndNormalizePath(targetPath)
	if err != nil {
		result.Error = fmt.Sprintf("Invalid path: %v", err)
		return result
	}

	fullPath := s.fileService.buildPath(validatedPath, file.Filename)
	
	// Open file
	f, err := file.Open()
	if err != nil {
		result.Error = fmt.Sprintf("Cannot open file: %v", err)
		return result
	}
	defer f.Close()

	// Read file data
	data := make([]byte, file.Size)
	_, err = f.Read(data)
	if err != nil {
		result.Error = fmt.Sprintf("Cannot read file: %v", err)
		return result
	}

	// Add file operation to transaction
	fileOperation := transaction.NewFileOperation(
		fmt.Sprintf("Upload file %s", file.Filename),
		func() error {
			return s.fileService.storage.SaveFile(fullPath, data)
		},
		func() error {
			// Rollback: delete the file
			return s.fileService.storage.DeleteFile(fullPath)
		},
	)
	tm.AddOperation(fileOperation)

	// Add database operation to transaction
	dbOperation := transaction.NewDatabaseOperation(
		fmt.Sprintf("Insert file record %s", file.Filename),
		func() error {
			fileInfo := &models.FileInfo{
				Name:        file.Filename,
				Path:        fullPath,
				Size:        file.Size,
				MimeType:    s.fileService.detectMimeType(file.Filename),
				IsDirectory: false,
				ParentPath:  validatedPath,
			}
			return s.fileService.repo.InsertFile(fileInfo)
		},
		func() error {
			// Rollback: delete from database
			return s.fileService.repo.DeleteFile(fullPath)
		},
	)
	tm.AddOperation(dbOperation)

	// Track uploaded file for cleanup
	*uploadedFiles = append(*uploadedFiles, fullPath)

	// Mark as success (will be validated when transaction executes)
	result.Success = true
	result.MimeType = s.fileService.detectMimeType(file.Filename)
	return result
}

// GetUploadProgress returns progress for batch uploads (placeholder for future implementation)
func (s *MultipleUploadService) GetUploadProgress(batchID string) (*models.BatchUploadProgress, error) {
	// This would be implemented with a progress tracking system
	// using Redis, in-memory cache, or database
	return nil, fmt.Errorf("progress tracking not implemented yet")
}

// CancelBatchUpload cancels an ongoing batch upload (placeholder for future implementation)
func (s *MultipleUploadService) CancelBatchUpload(batchID string) error {
	// This would be implemented with a cancellation system
	return fmt.Errorf("batch cancellation not implemented yet")
}