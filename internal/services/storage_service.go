package services

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/anddsdev/cloudlet/internal/security"
	"github.com/anddsdev/cloudlet/internal/storage"
)

type StorageService struct {
	basePath      string
	pathValidator *security.PathValidator
	atomicOps     *storage.AtomicFileOperations
}

func NewStorageService(basePath string) *StorageService {
	os.MkdirAll(basePath, 0755)

	return &StorageService{
		basePath:      basePath,
		pathValidator: security.NewPathValidator(basePath),
		atomicOps:     storage.NewAtomicFileOperations(basePath),
	}
}

func (s *StorageService) CreateDirectory(relativePath string) error {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	return os.MkdirAll(fullPath, 0755)
}

func (s *StorageService) SaveFile(relativePath string, data []byte) error {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	return s.atomicOps.AtomicWriteFile(fullPath, data, 0644)
}

func (s *StorageService) ReadFile(relativePath string) ([]byte, error) {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	return s.atomicOps.SafeReadFile(fullPath)
}

func (s *StorageService) GetFileInfo(relativePath string) (fs.FileInfo, error) {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	return os.Stat(fullPath)
}

func (s *StorageService) MoveFile(oldPath, newPath string) error {
	fullOldPath, err := s.getSecureFullPath(oldPath)
	if err != nil {
		return fmt.Errorf("invalid old path: %w", err)
	}

	fullNewPath, err := s.getSecureFullPath(newPath)
	if err != nil {
		return fmt.Errorf("invalid new path: %w", err)
	}

	return s.atomicOps.AtomicMoveFile(fullOldPath, fullNewPath)
}

func (s *StorageService) DeleteFile(relativePath string) error {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	return s.atomicOps.AtomicDeleteFile(fullPath)
}

// SaveFileStream saves data from an io.Reader atomically (for large files)
func (s *StorageService) SaveFileStream(relativePath string, reader io.Reader) error {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	return s.atomicOps.AtomicWriteFileStream(fullPath, reader, 0644)
}

// OpenFile opens a file for reading safely
func (s *StorageService) OpenFile(relativePath string) (*os.File, error) {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	return s.atomicOps.SafeOpenFile(fullPath)
}

// GetStats returns statistics about storage operations
func (s *StorageService) GetStats() map[string]interface{} {
	stats := s.atomicOps.GetStats()
	stats["base_path"] = s.basePath
	return stats
}

// Close shuts down the storage service and cleans up resources
func (s *StorageService) Close() error {
	if s.atomicOps != nil {
		return s.atomicOps.Close()
	}
	return nil
}

// GetPhysicalPath returns the physical file system path for a given relative path
func (s *StorageService) GetPhysicalPath(relativePath string) string {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		// Return a safe fallback path if validation fails
		return s.basePath
	}
	return fullPath
}

// getSecureFullPath validates and returns the full path safely
func (s *StorageService) getSecureFullPath(relativePath string) (string, error) {
	return s.pathValidator.ValidateAndGetFullPath(relativePath)
}
