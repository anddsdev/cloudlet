package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AtomicFileOperations provides thread-safe, atomic file operations
type AtomicFileOperations struct {
	// Mutex map for file-level locking
	fileLocks sync.Map

	// Global operations mutex for critical sections
	opsMutex sync.RWMutex

	// Temp directory for atomic operations
	tempDir string

	// Cleanup ticker for orphaned temp files
	cleanupTicker *time.Ticker
	cleanupDone   chan bool
}

// FileLock represents a per-file mutex
type FileLock struct {
	mu   sync.RWMutex
	refs int64 // Reference count
}

// TempFileInfo contains info about temporary files for atomic operations
type TempFileInfo struct {
	TempPath   string
	TargetPath string
	CreatedAt  time.Time
	ProcessID  int
}

// NewAtomicFileOperations creates a new atomic file operations manager
func NewAtomicFileOperations(baseDir string) *AtomicFileOperations {
	tempDir := filepath.Join(baseDir, ".cloudlet-tmp")
	os.MkdirAll(tempDir, 0755)

	afo := &AtomicFileOperations{
		tempDir:       tempDir,
		cleanupTicker: time.NewTicker(5 * time.Minute), // Cleanup every 5 minutes
		cleanupDone:   make(chan bool),
	}

	// Start background cleanup routine
	go afo.cleanupRoutine()

	return afo
}

// getFileLock gets or creates a file-specific lock
func (afo *AtomicFileOperations) getFileLock(filePath string) *FileLock {
	lockInterface, _ := afo.fileLocks.LoadOrStore(filePath, &FileLock{})
	lock := lockInterface.(*FileLock)

	// Increment reference count
	afo.opsMutex.Lock()
	lock.refs++
	afo.opsMutex.Unlock()

	return lock
}

// releaseFileLock decrements ref count and removes if no longer needed
func (afo *AtomicFileOperations) releaseFileLock(filePath string, lock *FileLock) {
	afo.opsMutex.Lock()
	lock.refs--
	if lock.refs <= 0 {
		afo.fileLocks.Delete(filePath)
	}
	afo.opsMutex.Unlock()
}

// generateUniqueTempPath creates a unique temporary file path
func (afo *AtomicFileOperations) generateUniqueTempPath(targetPath string) string {
	// Generate UUID for uniqueness
	id := uuid.New().String()

	// Add timestamp for additional uniqueness
	timestamp := time.Now().Unix()

	// Get base filename
	baseName := filepath.Base(targetPath)

	// Create unique temp filename
	tempFileName := fmt.Sprintf("%s.%d.%s.tmp", baseName, timestamp, id)

	return filepath.Join(afo.tempDir, tempFileName)
}

// AtomicWriteFile writes data to a file atomically
func (afo *AtomicFileOperations) AtomicWriteFile(targetPath string, data []byte, perm os.FileMode) error {
	// Get file-specific lock
	lock := afo.getFileLock(targetPath)
	lock.mu.Lock()
	defer func() {
		lock.mu.Unlock()
		afo.releaseFileLock(targetPath, lock)
	}()

	// Generate unique temporary file path
	tempPath := afo.generateUniqueTempPath(targetPath)

	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Write to temporary file
	if err := os.WriteFile(tempPath, data, perm); err != nil {
		// Cleanup temp file if write failed
		os.Remove(tempPath)
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically move temp file to target location
	if err := os.Rename(tempPath, targetPath); err != nil {
		// Cleanup temp file if rename failed
		os.Remove(tempPath)
		return fmt.Errorf("failed to move temporary file to target: %w", err)
	}

	return nil
}

// AtomicWriteFileStream writes from an io.Reader to a file atomically
func (afo *AtomicFileOperations) AtomicWriteFileStream(targetPath string, reader io.Reader, perm os.FileMode) error {
	// Get file-specific lock
	lock := afo.getFileLock(targetPath)
	lock.mu.Lock()
	defer func() {
		lock.mu.Unlock()
		afo.releaseFileLock(targetPath, lock)
	}()

	// Generate unique temporary file path
	tempPath := afo.generateUniqueTempPath(targetPath)

	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create temporary file
	tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Copy data from reader to temp file
	_, err = io.Copy(tempFile, reader)
	closeErr := tempFile.Close()

	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to write data to temporary file: %w", err)
	}

	if closeErr != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temporary file: %w", closeErr)
	}

	// Atomically move temp file to target location
	if err := os.Rename(tempPath, targetPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to move temporary file to target: %w", err)
	}

	return nil
}

// AtomicMoveFile moves a file atomically with proper locking
func (afo *AtomicFileOperations) AtomicMoveFile(sourcePath, targetPath string) error {
	// Lock both source and target files (in consistent order to prevent deadlocks)
	var firstPath, secondPath string
	if sourcePath < targetPath {
		firstPath, secondPath = sourcePath, targetPath
	} else {
		firstPath, secondPath = targetPath, sourcePath
	}

	lock1 := afo.getFileLock(firstPath)
	lock2 := afo.getFileLock(secondPath)

	lock1.mu.Lock()
	defer func() {
		lock1.mu.Unlock()
		afo.releaseFileLock(firstPath, lock1)
	}()

	lock2.mu.Lock()
	defer func() {
		lock2.mu.Unlock()
		afo.releaseFileLock(secondPath, lock2)
	}()

	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Perform atomic move
	if err := os.Rename(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %w", sourcePath, targetPath, err)
	}

	return nil
}

// AtomicDeleteFile deletes a file atomically with proper locking
func (afo *AtomicFileOperations) AtomicDeleteFile(filePath string) error {
	// Get file-specific lock
	lock := afo.getFileLock(filePath)
	lock.mu.Lock()
	defer func() {
		lock.mu.Unlock()
		afo.releaseFileLock(filePath, lock)
	}()

	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil // Already deleted, consider it success
	}
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		// For directories, use RemoveAll for recursive deletion
		return os.RemoveAll(filePath)
	} else {
		// For files, use Remove
		return os.Remove(filePath)
	}
}

// SafeReadFile reads a file with read lock to prevent reading during writes
func (afo *AtomicFileOperations) SafeReadFile(filePath string) ([]byte, error) {
	// Get file-specific lock
	lock := afo.getFileLock(filePath)
	lock.mu.RLock() // Read lock - allows concurrent reads, blocks writes
	defer func() {
		lock.mu.RUnlock()
		afo.releaseFileLock(filePath, lock)
	}()

	return os.ReadFile(filePath)
}

// SafeOpenFile opens a file for reading with read lock
func (afo *AtomicFileOperations) SafeOpenFile(filePath string) (*os.File, error) {
	// For streaming reads, we don't hold the lock for the entire duration
	// Just check the file exists and can be opened
	lock := afo.getFileLock(filePath)
	lock.mu.RLock()

	file, err := os.Open(filePath)

	lock.mu.RUnlock()
	afo.releaseFileLock(filePath, lock)

	return file, err
}

// cleanupRoutine runs periodically to clean up orphaned temporary files
func (afo *AtomicFileOperations) cleanupRoutine() {
	for {
		select {
		case <-afo.cleanupTicker.C:
			afo.cleanupOrphanedTempFiles()
		case <-afo.cleanupDone:
			return
		}
	}
}

// cleanupOrphanedTempFiles removes temporary files older than 1 hour
func (afo *AtomicFileOperations) cleanupOrphanedTempFiles() {
	cutoff := time.Now().Add(-1 * time.Hour)

	entries, err := os.ReadDir(afo.tempDir)
	if err != nil {
		return // Silently fail - temp dir might not exist yet
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".tmp" {
			fullPath := filepath.Join(afo.tempDir, entry.Name())

			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				os.Remove(fullPath) // Best effort cleanup
			}
		}
	}
}

// Close shuts down the atomic file operations manager
func (afo *AtomicFileOperations) Close() error {
	if afo.cleanupTicker != nil {
		afo.cleanupTicker.Stop()
	}

	if afo.cleanupDone != nil {
		close(afo.cleanupDone)
	}

	// Final cleanup of temp files
	afo.cleanupOrphanedTempFiles()

	return nil
}

// GetStats returns statistics about the atomic operations
func (afo *AtomicFileOperations) GetStats() map[string]interface{} {
	afo.opsMutex.RLock()
	defer afo.opsMutex.RUnlock()

	activeLocks := 0
	afo.fileLocks.Range(func(key, value interface{}) bool {
		activeLocks++
		return true
	})

	return map[string]interface{}{
		"active_locks": activeLocks,
		"temp_dir":     afo.tempDir,
	}
}
