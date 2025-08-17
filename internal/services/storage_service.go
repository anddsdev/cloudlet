package services

import (
	"io/fs"
	"os"
	"path/filepath"
)

type StorageService struct {
	basePath string
}

func NewStorageService(basePath string) *StorageService {
	os.MkdirAll(basePath, 0755)

	return &StorageService{
		basePath: basePath,
	}
}

func (s *StorageService) CreateDirectory(relativePath string) error {
	fullPath := s.getFullPath(relativePath)
	return os.MkdirAll(fullPath, 0755)
}

func (s *StorageService) SaveFile(relativePath string, data []byte) error {
	fullPath := s.getFullPath(relativePath)

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
	fullPath := s.getFullPath(relativePath)
	return os.ReadFile(fullPath)
}

func (s *StorageService) GetFileInfo(relativePath string) (fs.FileInfo, error) {
	fullPath := s.getFullPath(relativePath)
	return os.Stat(fullPath)
}

func (s *StorageService) MoveFile(oldPath, newPath string) error {
	fullOldPath := s.getFullPath(oldPath)
	fullNewPath := s.getFullPath(newPath)

	dir := filepath.Dir(fullNewPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.Rename(fullOldPath, fullNewPath)
}

func (s *StorageService) DeleteFile(relativePath string) error {
	fullPath := s.getFullPath(relativePath)

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

func (s *StorageService) getFullPath(relativePath string) string {
	relativePath = filepath.Clean(relativePath)
	if filepath.IsAbs(relativePath) {
		relativePath = relativePath[1:]
	}

	return filepath.Join(s.basePath, relativePath)
}
