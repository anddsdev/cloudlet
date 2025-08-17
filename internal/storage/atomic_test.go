package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAtomicFileOperations_ConcurrentWrites(t *testing.T) {
	tempDir := t.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	testFile := filepath.Join(tempDir, "concurrent_test.txt")

	const numGoroutines = 10
	const numWrites = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numWrites)

	// Start multiple goroutines writing to the same file concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numWrites; j++ {
				data := fmt.Sprintf("Goroutine %d, Write %d\n", goroutineID, j)
				err := afo.AtomicWriteFile(testFile, []byte(data), 0644)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, write %d: %w", goroutineID, j, err)
				}

				// Small delay to increase chance of race conditions if they exist
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		t.Fatalf("Errors during concurrent writes: %v", allErrors)
	}

	// Verify the file exists and has content from the last write
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("File should exist after concurrent writes")
	}

	// Read the final content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read final file content: %v", err)
	}

	// The content should be from exactly one of the writes (atomic)
	contentStr := string(content)
	if !strings.Contains(contentStr, "Goroutine") || !strings.Contains(contentStr, "Write") {
		t.Errorf("File content doesn't match expected pattern: %s", contentStr)
	}

	t.Logf("Final file content: %s", contentStr)
}

func TestAtomicFileOperations_ConcurrentReadWrite(t *testing.T) {
	tempDir := t.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	testFile := filepath.Join(tempDir, "read_write_test.txt")

	// Initial write
	initialData := "Initial content for read/write test"
	err := afo.AtomicWriteFile(testFile, []byte(initialData), 0644)
	if err != nil {
		t.Fatalf("Failed initial write: %v", err)
	}

	const numReaders = 3
	const numWriters = 2
	const duration = 50 * time.Millisecond

	var wg sync.WaitGroup
	errors := make(chan error, numReaders+numWriters)
	readResults := make(chan string, numReaders*10) // Buffer for multiple reads

	// Start reader goroutines
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()

			endTime := time.Now().Add(duration)
			for time.Now().Before(endTime) {
				data, err := afo.SafeReadFile(testFile)
				if err != nil {
					errors <- fmt.Errorf("reader %d: %w", readerID, err)
					return
				}

				readResults <- string(data)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Start writer goroutines
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()

			endTime := time.Now().Add(duration)
			writeCount := 0
			for time.Now().Before(endTime) {
				data := fmt.Sprintf("Writer %d, Write %d, Time: %d", writerID, writeCount, time.Now().Unix())
				err := afo.AtomicWriteFile(testFile, []byte(data), 0644)
				if err != nil {
					errors <- fmt.Errorf("writer %d: %w", writerID, err)
					return
				}

				writeCount++
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	close(errors)
	close(readResults)

	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		t.Fatalf("Errors during concurrent read/write: %v", allErrors)
	}

	// Verify all reads returned valid data (not corrupted/partial)
	readCount := 0
	for result := range readResults {
		readCount++
		if result == "" {
			t.Error("Read returned empty data")
		}
		// Each read should return either the initial data or complete writer data
		if !strings.Contains(result, "Initial content") && !strings.Contains(result, "Writer") {
			t.Errorf("Read returned invalid data: %s", result)
		}
	}

	t.Logf("Completed %d reads successfully", readCount)
}

func TestAtomicFileOperations_ConcurrentMove(t *testing.T) {
	tempDir := t.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	const numFiles = 10
	const numMovers = 3

	// Create initial files
	sourceFiles := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		sourceFiles[i] = filepath.Join(tempDir, fmt.Sprintf("source_%d.txt", i))
		data := fmt.Sprintf("Content of file %d", i)
		err := afo.AtomicWriteFile(sourceFiles[i], []byte(data), 0644)
		if err != nil {
			t.Fatalf("Failed to create source file %d: %v", i, err)
		}
	}

	// Create target directory
	targetDir := filepath.Join(tempDir, "moved")
	os.MkdirAll(targetDir, 0755)

	var wg sync.WaitGroup
	errors := make(chan error, numFiles*numMovers)

	// Start multiple goroutines trying to move files concurrently
	for moverID := 0; moverID < numMovers; moverID++ {
		wg.Add(1)
		go func(mover int) {
			defer wg.Done()

			for i, sourceFile := range sourceFiles {
				targetFile := filepath.Join(targetDir, fmt.Sprintf("moved_%d_by_%d.txt", i, mover))

				err := afo.AtomicMoveFile(sourceFile, targetFile)
				if err != nil {
					// It's expected that only one mover will succeed per file
					// Other movers will get "source file does not exist" errors
					if !strings.Contains(err.Error(), "source file does not exist") {
						errors <- fmt.Errorf("mover %d, file %d: %w", mover, i, err)
					}
				}

				time.Sleep(time.Millisecond)
			}
		}(moverID)
	}

	wg.Wait()
	close(errors)

	// Check for unexpected errors
	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		t.Fatalf("Unexpected errors during concurrent moves: %v", allErrors)
	}

	// Verify exactly one target file exists for each source file
	movedFiles := 0
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		t.Fatalf("Failed to read target directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			movedFiles++
		}
	}

	if movedFiles != numFiles {
		t.Errorf("Expected %d moved files, got %d", numFiles, movedFiles)
	}

	// Verify no source files remain
	for i, sourceFile := range sourceFiles {
		if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
			t.Errorf("Source file %d should not exist after move: %s", i, sourceFile)
		}
	}
}

func TestAtomicFileOperations_StreamOperations(t *testing.T) {
	tempDir := t.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	testFile := filepath.Join(tempDir, "stream_test.txt")
	testData := "This is test data for streaming operations\nLine 2\nLine 3\n"

	// Test atomic write with stream
	reader := strings.NewReader(testData)
	err := afo.AtomicWriteFileStream(testFile, reader, 0644)
	if err != nil {
		t.Fatalf("Failed to write file with stream: %v", err)
	}

	// Verify content
	content, err := afo.SafeReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("Content mismatch. Expected: %q, Got: %q", testData, string(content))
	}

	// Test concurrent stream writes
	const numStreams = 5
	var wg sync.WaitGroup
	errors := make(chan error, numStreams)

	for i := 0; i < numStreams; i++ {
		wg.Add(1)
		go func(streamID int) {
			defer wg.Done()

			streamFile := filepath.Join(tempDir, fmt.Sprintf("stream_%d.txt", streamID))
			data := fmt.Sprintf("Stream %d data\nMultiple lines\nFor testing\n", streamID)
			reader := strings.NewReader(data)

			err := afo.AtomicWriteFileStream(streamFile, reader, 0644)
			if err != nil {
				errors <- fmt.Errorf("stream %d: %w", streamID, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		t.Fatalf("Errors during concurrent stream writes: %v", allErrors)
	}

	// Verify all stream files were created correctly
	for i := 0; i < numStreams; i++ {
		streamFile := filepath.Join(tempDir, fmt.Sprintf("stream_%d.txt", i))
		if _, err := os.Stat(streamFile); os.IsNotExist(err) {
			t.Errorf("Stream file %d was not created", i)
		}
	}
}

func TestAtomicFileOperations_CleanupTempFiles(t *testing.T) {
	tempDir := t.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	tempFileDir := filepath.Join(tempDir, ".cloudlet-tmp")

	// Create some old temp files manually
	oldTempFile := filepath.Join(tempFileDir, "old_file.tmp")
	os.MkdirAll(tempFileDir, 0755)
	err := os.WriteFile(oldTempFile, []byte("old temp content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test temp file: %v", err)
	}

	// Modify the file time to be old
	oldTime := time.Now().Add(-2 * time.Hour)
	os.Chtimes(oldTempFile, oldTime, oldTime)

	// Create a recent temp file
	recentTempFile := filepath.Join(tempFileDir, "recent_file.tmp")
	err = os.WriteFile(recentTempFile, []byte("recent temp content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create recent temp file: %v", err)
	}

	// Trigger cleanup
	afo.cleanupOrphanedTempFiles()

	// Old file should be gone
	if _, err := os.Stat(oldTempFile); !os.IsNotExist(err) {
		t.Error("Old temp file should have been cleaned up")
	}

	// Recent file should still exist
	if _, err := os.Stat(recentTempFile); os.IsNotExist(err) {
		t.Error("Recent temp file should not have been cleaned up")
	}
}

func TestAtomicFileOperations_UniqueTemporaryNames(t *testing.T) {
	tempDir := t.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	targetPath := filepath.Join(tempDir, "test.txt")

	// Generate multiple temp paths and verify they're unique
	tempPaths := make(map[string]bool)
	const numPaths = 100

	for i := 0; i < numPaths; i++ {
		tempPath := afo.generateUniqueTempPath(targetPath)

		if tempPaths[tempPath] {
			t.Errorf("Duplicate temp path generated: %s", tempPath)
		}
		tempPaths[tempPath] = true

		// Verify path format
		if !strings.HasSuffix(tempPath, ".tmp") {
			t.Errorf("Temp path doesn't end with .tmp: %s", tempPath)
		}

		if !strings.Contains(tempPath, "test.txt") {
			t.Errorf("Temp path doesn't contain original filename: %s", tempPath)
		}
	}

	if len(tempPaths) != numPaths {
		t.Errorf("Expected %d unique paths, got %d", numPaths, len(tempPaths))
	}
}

func TestAtomicFileOperations_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	// Test writing to invalid directory
	invalidPath := filepath.Join(tempDir, "nonexistent", "deep", "path", "file.txt")

	// This should still work because AtomicWriteFile creates directories
	err := afo.AtomicWriteFile(invalidPath, []byte("test"), 0644)
	if err != nil {
		t.Errorf("AtomicWriteFile should create directories: %v", err)
	}

	// Test moving non-existent file
	nonExistentFile := filepath.Join(tempDir, "does_not_exist.txt")
	targetFile := filepath.Join(tempDir, "target.txt")

	err = afo.AtomicMoveFile(nonExistentFile, targetFile)
	if err == nil {
		t.Error("Moving non-existent file should return error")
	}

	// Test deleting non-existent file (should not error)
	err = afo.AtomicDeleteFile(nonExistentFile)
	if err != nil {
		t.Errorf("Deleting non-existent file should not error: %v", err)
	}
}

func TestAtomicFileOperations_GetStats(t *testing.T) {
	tempDir := t.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	stats := afo.GetStats()

	if stats["active_locks"] == nil {
		t.Error("Stats should include active_locks")
	}

	if stats["temp_dir"] == nil {
		t.Error("Stats should include temp_dir")
	}

	if stats["temp_dir"] != afo.tempDir {
		t.Errorf("Stats temp_dir mismatch. Expected: %s, Got: %s", afo.tempDir, stats["temp_dir"])
	}
}

// Benchmark tests
func BenchmarkAtomicFileOperations_WriteFile(b *testing.B) {
	tempDir := b.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, fmt.Sprintf("benchmark_%d.txt", i))
		err := afo.AtomicWriteFile(testFile, data, 0644)
		if err != nil {
			b.Fatalf("Benchmark write failed: %v", err)
		}
	}
}

func BenchmarkAtomicFileOperations_ReadFile(b *testing.B) {
	tempDir := b.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	// Setup test file
	testFile := filepath.Join(tempDir, "benchmark_read.txt")
	data := []byte("benchmark test data for reading")
	err := afo.AtomicWriteFile(testFile, data, 0644)
	if err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := afo.SafeReadFile(testFile)
		if err != nil {
			b.Fatalf("Benchmark read failed: %v", err)
		}
	}
}

// Test that demonstrates the race condition fix
func TestRaceConditionFix(t *testing.T) {
	tempDir := t.TempDir()
	afo := NewAtomicFileOperations(tempDir)
	defer afo.Close()

	testFile := filepath.Join(tempDir, "race_test.txt")

	// This test simulates the exact scenario that would cause race conditions
	// Multiple goroutines trying to save the same file simultaneously
	const numGoroutines = 20
	const fileContent = "This is the final content that should be in the file"

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// All goroutines try to write the same content to the same file
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			data := fmt.Sprintf("%s (written by goroutine %d)", fileContent, id)
			err := afo.AtomicWriteFile(testFile, []byte(data), 0644)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		t.Fatalf("Race condition test failed with errors: %v", allErrors)
	}

	// Verify the file exists and has complete content from one of the writes
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read final file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, fileContent) {
		t.Errorf("File content is corrupted or incomplete: %s", contentStr)
	}

	// The content should be from exactly one goroutine (atomic write)
	// and should contain "written by goroutine X"
	if !strings.Contains(contentStr, "written by goroutine") {
		t.Errorf("File doesn't contain complete write from one goroutine: %s", contentStr)
	}

	t.Logf("Race condition test passed. Final content: %s", contentStr)
}
