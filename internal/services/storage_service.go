package services

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/anddsdev/cloudlet/internal/security"
)

type StorageService struct {
	basePath      string
	pathValidator *security.PathValidator
}

func NewStorageService(basePath string) *StorageService {
	os.MkdirAll(basePath, 0755)

	return &StorageService{
		basePath:      basePath,
		pathValidator: security.NewPathValidator(basePath),
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

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tempPath := fullPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, fullPath)
}

func (s *StorageService) ReadFile(relativePath string) ([]byte, error) {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	return os.ReadFile(fullPath)
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

	dir := filepath.Dir(fullNewPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.Rename(fullOldPath, fullNewPath)
}

func (s *StorageService) DeleteFile(relativePath string) error {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return os.RemoveAll(fullPath)
	} else {
		return os.Remove(fullPath)
	}
}

// getSecureFullPath validates and returns the full path safely
func (s *StorageService) getSecureFullPath(relativePath string) (string, error) {
	return s.pathValidator.ValidateAndGetFullPath(relativePath)
}

// getFullPath is now deprecated - use getSecureFullPath instead
// Kept for compatibility but should not be used
func (s *StorageService) getFullPath(relativePath string) string {
	fullPath, err := s.getSecureFullPath(relativePath)
	if err != nil {
		// For backwards compatibility, return a safe default
		// In production, this should return an error
		return filepath.Join(s.basePath, "invalid")
	}
	return fullPath
}
