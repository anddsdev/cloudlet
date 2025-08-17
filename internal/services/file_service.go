package services

import (
	"errors"
	"fmt"

	"path/filepath"
	"strings"

	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/repository"
	"github.com/anddsdev/cloudlet/internal/security"
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

	// Rename physically
	err = s.storage.MoveFile(path, newPath)
	if err != nil {
		return err
	}

	// Update database first
	err = s.repo.RenameFile(path, newName)
	if err != nil {
		// Rollback: revert rename physically
		s.storage.MoveFile(newPath, path)
		return err
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

	// Move physically
	err = s.storage.MoveFile(sourcePath, newPath)
	if err != nil {
		return err
	}

	// Update database first
	err = s.repo.MoveFile(sourcePath, destinationPath)
	if err != nil {
		// Rollback: revert move physically
		s.storage.MoveFile(newPath, sourcePath)
		return err
	}

	return nil
}

func (s *FileService) DeleteFile(path string) error {
	// Validate and normalize path
	validatedPath, err := s.pathValidator.ValidateAndNormalizePath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	path = validatedPath

	_, err = s.repo.GetFileByPath(path)
	if err != nil {
		return ErrFileNotFound
	}

	// Delete from database first
	err = s.repo.DeleteFile(path)
	if err != nil {
		return err
	}

	// Delete physically
	err = s.storage.DeleteFile(path)
	if err != nil {
		// Rollback: this is complex, for now log the error
		// In production, implement a cleanup job

		fmt.Printf("Warning: failed to delete physical file %s: %v\n", path, err)
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

	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".md":   "text/markdown",
		".json": "application/json",
		".xml":  "text/xml",
		".zip":  "application/zip",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}

	return "application/octet-stream"
}

var ErrFileNotFound = errors.New("file not found")
