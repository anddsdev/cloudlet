package services

import (
	"errors"
	"fmt"
	"io"

	"path/filepath"
	"strings"

	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/repository"
	"github.com/anddsdev/cloudlet/internal/security"
	"github.com/anddsdev/cloudlet/internal/transaction"
)

type FileService struct {
	repo          *repository.FileRepository
	storage       *StorageService
	pathValidator *security.PathValidator
}

func NewFileService(repo *repository.FileRepository, storage *StorageService, storagePath string) *FileService {
	return &FileService{
		repo:          repo,
		storage:       storage,
		pathValidator: security.NewPathValidator(storagePath),
	}
}

func (s *FileService) GetDirectoryListing(path string) (*models.DirectoryListing, error) {
	validatedPath, err := s.pathValidator.ValidateAndNormalizePath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	path = validatedPath

	files, err := s.repo.GetFilesByPath(path)
	if err != nil {
		return nil, err
	}

	var directories []*models.FileInfo
	var regularFiles []*models.FileInfo
	var totalSize int64

	for _, file := range files {
		if file.IsDirectory {
			directories = append(directories, file)
		} else {
			regularFiles = append(regularFiles, file)
			totalSize += file.Size
		}
	}

	breadcrumbs := s.generateBreadcrumbs(path)

	parentPath := s.getParentPath(path)

	return &models.DirectoryListing{
		Path:        path,
		ParentPath:  parentPath,
		Files:       regularFiles,
		Directories: directories,
		TotalFiles:  len(regularFiles),
		TotalDirs:   len(directories),
		TotalSize:   totalSize,
		Breadcrumbs: breadcrumbs,
	}, nil
}

func (s *FileService) CreateDirectory(name, parentPath string) (*models.FileInfo, error) {
	// Validate filename
	if err := security.IsValidFilename(name); err != nil {
		return nil, fmt.Errorf("invalid directory name: %w", err)
	}

	// Validate and normalize parent path
	validatedParentPath, err := s.pathValidator.ValidateAndNormalizePath(parentPath)
	if err != nil {
		return nil, fmt.Errorf("invalid parent path: %w", err)
	}
	parentPath = validatedParentPath

	if !s.isValidName(name) {
		return nil, errors.New("invalid directory name")
	}

	dir, err := s.repo.CreateDirectory(name, parentPath)
	if err != nil {
		return nil, err
	}

	err = s.storage.CreateDirectory(dir.Path)
	if err != nil {
		// Rollback: Delete from database if fails physically
		s.repo.DeleteFile(dir.Path)
		return nil, err
	}

	return dir, nil
}

func (s *FileService) SaveFile(filename, parentPath string, data []byte) error {
	// Validate filename
	if err := security.IsValidFilename(filename); err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}

	// Validate and normalize parent path
	validatedParentPath, err := s.pathValidator.ValidateAndNormalizePath(parentPath)
	if err != nil {
		return fmt.Errorf("invalid parent path: %w", err)
	}
	parentPath = validatedParentPath
	fullPath := s.buildPath(parentPath, filename)

	if err := s.storage.SaveFile(fullPath, data); err != nil {
		return err
	}

	file := &models.FileInfo{
		Name:        filename,
		Path:        fullPath,
		Size:        int64(len(data)),
		MimeType:    s.detectMimeType(filename),
		IsDirectory: false,
		ParentPath:  parentPath,
	}

	return s.repo.InsertFile(file)
}

// SaveFileStream saves a file from an io.Reader using streaming to prevent memory leaks
func (s *FileService) SaveFileStream(filename, parentPath string, reader io.Reader, size int64) error {
	// Validate filename
	if err := security.IsValidFilename(filename); err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}

	// Validate and normalize parent path
	validatedParentPath, err := s.pathValidator.ValidateAndNormalizePath(parentPath)
	if err != nil {
		return fmt.Errorf("invalid parent path: %w", err)
	}
	parentPath = validatedParentPath
	fullPath := s.buildPath(parentPath, filename)

	// Save file using streaming operations
	if err := s.storage.SaveFileStream(fullPath, reader); err != nil {
		return err
	}

	// Create file metadata
	file := &models.FileInfo{
		Name:        filename,
		Path:        fullPath,
		Size:        size,
		MimeType:    s.detectMimeType(filename),
		IsDirectory: false,
		ParentPath:  parentPath,
	}

	return s.repo.InsertFile(file)
}

func (s *FileService) GetFileData(path string) ([]byte, *models.FileInfo, error) {
	// Validate and normalize path
	validatedPath, err := s.pathValidator.ValidateAndNormalizePath(path)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid path: %w", err)
	}
	path = validatedPath

	fileInfo, err := s.repo.GetFileByPath(path)
	if err != nil {
		return nil, nil, ErrFileNotFound
	}

	// Only allow download of files, not directories
	// TODO: implement directory download (zip?)
	if fileInfo.IsDirectory {
		return nil, nil, errors.New("cannot download directory")
	}

	data, err := s.storage.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	return data, fileInfo, nil
}

func (s *FileService) RenameFile(path, newName string) error {
	// Validate new filename
	if err := security.IsValidFilename(newName); err != nil {
		return fmt.Errorf("invalid new name: %w", err)
	}

	// Validate and normalize path
	validatedPath, err := s.pathValidator.ValidateAndNormalizePath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	path = validatedPath

	fileInfo, err := s.repo.GetFileByPath(path)
	if err != nil {
		return ErrFileNotFound
	}

	newPath := s.buildPath(fileInfo.ParentPath, newName)

	// Create transaction manager for atomic operations
	tm := transaction.NewTransactionManager()

	// First operation: Rename physical file
	storageOperation := transaction.NewFileOperation(
		fmt.Sprintf("Rename file from %s to %s", path, newPath),
		func() error {
			return s.storage.MoveFile(path, newPath)
		},
		func() error {
			// Rollback: revert rename physically
			return s.storage.MoveFile(newPath, path)
		},
	)
	tm.AddOperation(storageOperation)

	// Second operation: Update database
	dbOperation := transaction.NewDatabaseOperation(
		fmt.Sprintf("Update database record for rename: %s to %s", path, newName),
		func() error {
			return s.repo.RenameFile(path, newName)
		},
		func() error {
			// Rollback: revert database changes
			return s.repo.RenameFile(newPath, fileInfo.Name)
		},
	)
	tm.AddOperation(dbOperation)

	// Execute all operations with automatic rollback on failure
	if err := tm.Execute(); err != nil {
		return fmt.Errorf("failed to rename file %s: %w", path, err)
	}

	return nil
}

func (s *FileService) MoveFile(sourcePath, destinationPath string) error {
	// Validate and normalize source path
	validatedSourcePath, err := s.pathValidator.ValidateAndNormalizePath(sourcePath)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	sourcePath = validatedSourcePath

	// Validate and normalize destination path
	validatedDestinationPath, err := s.pathValidator.ValidateAndNormalizePath(destinationPath)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}
	destinationPath = validatedDestinationPath

	sourceInfo, err := s.repo.GetFileByPath(sourcePath)
	if err != nil {
		return ErrFileNotFound
	}

	newPath := s.buildPath(destinationPath, sourceInfo.Name)

	// Create transaction manager for atomic operations
	tm := transaction.NewTransactionManager()

	// First operation: Move file physically
	storageOperation := transaction.NewFileOperation(
		fmt.Sprintf("Move file from %s to %s", sourcePath, newPath),
		func() error {
			return s.storage.MoveFile(sourcePath, newPath)
		},
		func() error {
			// Rollback: revert move physically
			return s.storage.MoveFile(newPath, sourcePath)
		},
	)
	tm.AddOperation(storageOperation)

	// Second operation: Update database
	dbOperation := transaction.NewDatabaseOperation(
		fmt.Sprintf("Update database record for move: %s to %s", sourcePath, destinationPath),
		func() error {
			return s.repo.MoveFile(sourcePath, destinationPath)
		},
		func() error {
			// Rollback: revert database changes
			return s.repo.MoveFile(newPath, sourceInfo.ParentPath)
		},
	)
	tm.AddOperation(dbOperation)

	// Execute all operations with automatic rollback on failure
	if err := tm.Execute(); err != nil {
		return fmt.Errorf("failed to move file %s: %w", sourcePath, err)
	}

	return nil
}

func (s *FileService) DeleteFile(path string, recursive ...bool) error {
	// Validate and normalize path
	validatedPath, err := s.pathValidator.ValidateAndNormalizePath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	path = validatedPath

	// Get file info for rollback purposes
	fileInfo, err := s.repo.GetFileByPath(path)
	if err != nil {
		return ErrFileNotFound
	}

	// Check if this is a directory and if recursive deletion is requested
	isRecursive := len(recursive) > 0 && recursive[0]
	
	if fileInfo.IsDirectory && !isRecursive {
		// Check if directory has children
		children, err := s.repo.GetFilesByPath(path)
		if err != nil {
			return err
		}
		
		if len(children) > 0 {
			return errors.New("directory not empty")
		}
	}

	// If it's a directory and recursive is true, delete all children first
	if fileInfo.IsDirectory && isRecursive {
		if err := s.deleteDirectoryRecursive(path); err != nil {
			return fmt.Errorf("failed to delete directory recursively: %w", err)
		}
		return nil
	}

	// Create transaction manager for atomic operations
	tm := transaction.NewTransactionManager()

	// First operation: Delete from database
	dbOperation := transaction.NewDatabaseOperation(
		fmt.Sprintf("Delete file from database: %s", path),
		func() error {
			return s.repo.DeleteFile(path)
		},
		func() error {
			// Rollback: Re-insert file into database
			return s.repo.InsertFile(fileInfo)
		},
	)
	tm.AddOperation(dbOperation)

	// Second operation: Delete physical file
	storageOperation := transaction.NewFileOperation(
		fmt.Sprintf("Delete physical file: %s", path),
		func() error {
			return s.storage.DeleteFile(path)
		},
		func() error {
			// Rollback: This is complex for file restoration
			// In practice, we'd need to have backed up the file content
			// For now, we'll log the inconsistency
			return fmt.Errorf("cannot restore deleted file %s - manual intervention required", path)
		},
	)
	tm.AddOperation(storageOperation)

	// Execute all operations with automatic rollback on failure
	if err := tm.Execute(); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}

	return nil
}

func (s *FileService) deleteDirectoryRecursive(path string) error {
	// Get all children of this directory
	children, err := s.repo.GetFilesByPath(path)
	if err != nil {
		return err
	}

	// First, delete all files and subdirectories recursively
	for _, child := range children {
		if child.IsDirectory {
			// Recursively delete subdirectory
			if err := s.deleteDirectoryRecursive(child.Path); err != nil {
				return fmt.Errorf("failed to delete subdirectory %s: %w", child.Path, err)
			}
		} else {
			// Delete file
			if err := s.DeleteFile(child.Path, false); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", child.Path, err)
			}
		}
	}

	// Finally, delete the directory itself
	if err := s.DeleteFile(path, false); err != nil {
		return fmt.Errorf("failed to delete directory %s: %w", path, err)
	}

	return nil
}

// normalizePath is now deprecated - use pathValidator.ValidateAndNormalizePath instead
// Kept for compatibility but should not be used
func (s *FileService) normalizePath(path string) string {
	validated, err := s.pathValidator.ValidateAndNormalizePath(path)
	if err != nil {
		// For backwards compatibility, return "/" for invalid paths
		// In production, this should return an error
		return "/"
	}
	return validated
}

func (s *FileService) buildPath(parent, name string) string {
	if parent == "/" {
		return "/" + name
	}
	return parent + "/" + name
}

func (s *FileService) getParentPath(path string) string {
	if path == "/" {
		return ""
	}

	parent := filepath.Dir(path)
	if parent == "." {
		return "/"
	}

	return parent
}

func (s *FileService) generateBreadcrumbs(path string) []models.Breadcrumb {
	if path == "/" {
		return []models.Breadcrumb{{Name: "Home", Path: "/"}}
	}

	var breadcrumbs []models.Breadcrumb
	breadcrumbs = append(breadcrumbs, models.Breadcrumb{Name: "Home", Path: "/"})

	parts := strings.Split(strings.Trim(path, "/"), "/")
	currentPath := ""

	for _, part := range parts {
		if part != "" {
			currentPath += "/" + part
			breadcrumbs = append(breadcrumbs, models.Breadcrumb{
				Name: part,
				Path: currentPath,
			})
		}
	}

	return breadcrumbs
}

// isValidName is now deprecated - use security.IsValidFilename instead
// Kept for compatibility but should not be used
func (s *FileService) isValidName(name string) bool {
	return security.IsValidFilename(name) == nil
}

func (s *FileService) detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// Comprehensive MIME type mapping
	mimeTypes := map[string]string{
		// Images
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".png":   "image/png",
		".gif":   "image/gif",
		".bmp":   "image/bmp",
		".webp":  "image/webp",
		".svg":   "image/svg+xml",
		".ico":   "image/x-icon",
		".tiff":  "image/tiff",
		".tif":   "image/tiff",

		// Documents
		".pdf":   "application/pdf",
		".doc":   "application/msword",
		".docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":   "application/vnd.ms-excel",
		".xlsx":  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":   "application/vnd.ms-powerpoint",
		".pptx":  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".odt":   "application/vnd.oasis.opendocument.text",
		".ods":   "application/vnd.oasis.opendocument.spreadsheet",
		".odp":   "application/vnd.oasis.opendocument.presentation",

		// Text
		".txt":   "text/plain",
		".md":    "text/markdown",
		".html":  "text/html",
		".htm":   "text/html",
		".css":   "text/css",
		".js":    "text/javascript",
		".json":  "application/json",
		".xml":   "text/xml",
		".csv":   "text/csv",
		".yaml":  "application/x-yaml",
		".yml":   "application/x-yaml",
		".toml":  "application/toml",

		// Archives
		".zip":   "application/zip",
		".rar":   "application/vnd.rar",
		".7z":    "application/x-7z-compressed",
		".tar":   "application/x-tar",
		".gz":    "application/gzip",
		".bz2":   "application/x-bzip2",
		".xz":    "application/x-xz",

		// Audio
		".mp3":   "audio/mpeg",
		".wav":   "audio/wav",
		".flac":  "audio/flac",
		".ogg":   "audio/ogg",
		".m4a":   "audio/mp4",
		".aac":   "audio/aac",

		// Video
		".mp4":   "video/mp4",
		".avi":   "video/x-msvideo",
		".mov":   "video/quicktime",
		".wmv":   "video/x-ms-wmv",
		".flv":   "video/x-flv",
		".webm":  "video/webm",
		".mkv":   "video/x-matroska",

		// Code files
		".go":    "text/x-go",
		".py":    "text/x-python",
		".java":  "text/x-java-source",
		".c":     "text/x-c",
		".cpp":   "text/x-c++",
		".h":     "text/x-c",
		".hpp":   "text/x-c++",
		".php":   "text/x-php",
		".rb":    "text/x-ruby",
		".sh":    "text/x-shellscript",
		".sql":   "application/sql",

		// Fonts
		".ttf":   "font/ttf",
		".otf":   "font/otf",
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".eot":   "application/vnd.ms-fontobject",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}

	// Additional validation - reject potentially dangerous files
	dangerousExtensions := map[string]bool{
		".exe": true, ".bat": true, ".cmd": true, ".com": true,
		".scr": true, ".pif": true, ".vbs": true, ".ps1": true,
		".jar": true, ".app": true, ".deb": true, ".rpm": true,
		".dmg": true, ".pkg": true, ".msi": true,
	}

	if dangerousExtensions[ext] {
		return "application/x-executable"
	}

	return "application/octet-stream"
}

var ErrFileNotFound = errors.New("file not found")
