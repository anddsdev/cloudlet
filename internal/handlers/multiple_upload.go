package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/services"
	"github.com/anddsdev/cloudlet/internal/utils"
)

// UploadMultiple handles multiple file uploads using hybrid strategy
func (h *Handlers) UploadMultiple(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with configured memory limit
	err := r.ParseMultipartForm(int64(h.cfg.Server.MaxMemory))
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	// Get target path
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "/"
	}

	// Get upload strategy preference
	strategy := r.FormValue("strategy")
	if strategy == "" {
		strategy = "hybrid" // Default strategy
	}

	// Get files from form
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "No files provided")
		return
	}

	// Check basic limits
	if len(files) > h.cfg.Server.Upload.MaxFilesPerRequest {
		utils.WriteErrorJSON(w, http.StatusBadRequest, 
			fmt.Sprintf("Too many files: %d exceeds limit of %d", len(files), h.cfg.Server.Upload.MaxFilesPerRequest))
		return
	}

	// Check rate limiting (basic IP-based check)
	clientIP := getClientIP(r)
	if err := h.checkUploadRateLimit(clientIP, len(files)); err != nil {
		utils.WriteErrorJSON(w, http.StatusTooManyRequests, err.Error())
		return
	}

	// Create multiple upload service
	multipleUploadService := services.NewMultipleUploadService(h.fileService, h.cfg, h.cfg.Server.Storage.Path)

	// Process multiple uploads
	response := multipleUploadService.UploadMultipleFiles(files, targetPath)

	// Set appropriate HTTP status based on results
	status := http.StatusCreated
	if !response.Success {
		if response.SuccessfulFiles == 0 {
			status = http.StatusBadRequest
		} else {
			status = http.StatusMultiStatus // Partial success
		}
	}

	utils.WriteJSON(w, status, response)
}

// UploadMultipleValidate validates multiple files without actually uploading them
func (h *Handlers) UploadMultipleValidate(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(int64(h.cfg.Server.MaxMemory))
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	// Get target path
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "/"
	}

	// Get files from form
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "No files provided")
		return
	}

	// Create validator and run validation
	validator := services.NewMultipleUploadValidator(h.cfg, h.cfg.Server.Storage.Path)
	validation := validator.ValidateMultipleUpload(files, targetPath)

	// Return validation results
	utils.WriteJSON(w, http.StatusOK, validation)
}

// UploadBatch handles batch uploads with progress tracking
func (h *Handlers) UploadBatch(w http.ResponseWriter, r *http.Request) {
	// Check if batch processing is enabled
	if !h.cfg.Server.Upload.EnableBatchProcessing {
		utils.WriteErrorJSON(w, http.StatusNotImplemented, "Batch processing is disabled")
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(int64(h.cfg.Server.MaxMemory))
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	// Get target path
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "/"
	}

	// Get batch size
	batchSizeStr := r.FormValue("batch_size")
	batchSize := h.cfg.Server.Upload.BatchSize
	if batchSizeStr != "" {
		if parsed, err := strconv.Atoi(batchSizeStr); err == nil && parsed > 0 {
			batchSize = parsed
		}
	}

	// Get files from form
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "No files provided")
		return
	}

	// Process in batches
	response := h.processBatchUpload(files, targetPath, batchSize)
	
	status := http.StatusCreated
	if !response.Success {
		if response.SuccessfulFiles == 0 {
			status = http.StatusBadRequest
		} else {
			status = http.StatusMultiStatus
		}
	}

	utils.WriteJSON(w, status, response)
}

// UploadMultipleStream handles multiple file uploads using streaming for all files
func (h *Handlers) UploadMultipleStream(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with smaller memory limit for streaming
	streamingMemory := h.cfg.Server.MaxMemory / 4 // Use less memory for streaming
	err := r.ParseMultipartForm(int64(streamingMemory))
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	// Get target path
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "/"
	}

	// Get files from form
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "No files provided")
		return
	}

	// Force streaming for all files by temporarily lowering threshold
	originalThreshold := h.cfg.Server.Upload.StreamingThreshold
	h.cfg.Server.Upload.StreamingThreshold = 1 // Force streaming for all files

	// Create multiple upload service
	multipleUploadService := services.NewMultipleUploadService(h.fileService, h.cfg, h.cfg.Server.Storage.Path)

	// Process uploads
	response := multipleUploadService.UploadMultipleFiles(files, targetPath)
	response.Strategy = "streaming"

	// Restore original threshold
	h.cfg.Server.Upload.StreamingThreshold = originalThreshold

	status := http.StatusCreated
	if !response.Success {
		if response.SuccessfulFiles == 0 {
			status = http.StatusBadRequest
		} else {
			status = http.StatusMultiStatus
		}
	}

	utils.WriteJSON(w, status, response)
}

// processBatchUpload processes files in batches
func (h *Handlers) processBatchUpload(files []*multipart.FileHeader, targetPath string, batchSize int) *models.MultipleUploadResponse {
	totalFiles := len(files)
	allResults := make([]models.FileUploadResult, 0, totalFiles)
	totalSuccessful := 0
	totalFailed := 0
	var totalProcessedSize int64

	// Create multiple upload service
	multipleUploadService := services.NewMultipleUploadService(h.fileService, h.cfg, h.cfg.Server.Storage.Path)

	// Process in batches
	for i := 0; i < totalFiles; i += batchSize {
		end := i + batchSize
		if end > totalFiles {
			end = totalFiles
		}

		batch := files[i:end]
		batchResponse := multipleUploadService.UploadMultipleFiles(batch, targetPath)

		// Adjust indices for global context
		for j := range batchResponse.Files {
			batchResponse.Files[j].Index = i + j
		}

		allResults = append(allResults, batchResponse.Files...)
		totalSuccessful += batchResponse.SuccessfulFiles
		totalFailed += batchResponse.FailedFiles
		totalProcessedSize += batchResponse.ProcessedSize

		log.Printf("Processed batch %d-%d: %d successful, %d failed", 
			i, end-1, batchResponse.SuccessfulFiles, batchResponse.FailedFiles)
	}

	// Calculate total size
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	response := &models.MultipleUploadResponse{
		Success:         totalSuccessful > 0,
		TotalFiles:      totalFiles,
		SuccessfulFiles: totalSuccessful,
		FailedFiles:     totalFailed,
		TotalSize:       totalSize,
		ProcessedSize:   totalProcessedSize,
		Files:           allResults,
		Strategy:        "batch",
	}

	if totalSuccessful == totalFiles {
		response.Message = fmt.Sprintf("Successfully uploaded all %d files in batches", totalFiles)
	} else if totalSuccessful > 0 {
		response.Message = fmt.Sprintf("Batch upload completed: %d/%d files successful", totalSuccessful, totalFiles)
	} else {
		response.Message = "All batch uploads failed"
	}

	return response
}

// checkUploadRateLimit performs basic rate limiting check
func (h *Handlers) checkUploadRateLimit(clientIP string, fileCount int) error {
	// In a production environment, this would use Redis or another cache
	// For now, we'll do a simple check against configuration
	if fileCount > h.cfg.Server.Upload.RateLimitPerMinute {
		return fmt.Errorf("rate limit exceeded: %d files exceeds limit of %d per minute", 
			fileCount, h.cfg.Server.Upload.RateLimitPerMinute)
	}
	return nil
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// GetBatchProgress returns progress for a batch upload (placeholder)
func (h *Handlers) GetBatchProgress(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("batchId")
	if batchID == "" {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Batch ID is required")
		return
	}

	// This would be implemented with actual progress tracking
	progress := &models.BatchUploadProgress{
		BatchID:         batchID,
		Status:          "not_implemented",
		TotalFiles:      0,
		ProcessedFiles:  0,
		PercentComplete: 0,
	}

	utils.WriteJSON(w, http.StatusOK, progress)
}

// CancelBatchUpload cancels a batch upload (placeholder)
func (h *Handlers) CancelBatchUpload(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("batchId")
	if batchID == "" {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Batch ID is required")
		return
	}

	// This would be implemented with actual cancellation logic
	response := map[string]string{
		"status":  "not_implemented",
		"message": "Batch cancellation not implemented yet",
		"batchId": batchID,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}