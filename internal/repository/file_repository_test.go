package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/anddsdev/cloudlet/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// createTestDB creates an in-memory SQLite database for testing
func createTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Enable WAL mode and other pragmas for consistency with production
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA foreign_keys=ON",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			t.Fatalf("Failed to set pragma %s: %v", pragma, err)
		}
	}

	return db
}

// setupTestRepository creates a FileRepository with test database
func setupTestRepository(t *testing.T) *FileRepository {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	repo, err := NewFileRepository(dbPath, 5)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	
	return repo
}

func TestNewFileRepository(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	repo, err := NewFileRepository(dbPath, 5)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Verify tables were created
	var count int
	err = repo.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='files'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check for files table: %v", err)
	}
	if count != 1 {
		t.Error("Files table was not created")
	}
}

func TestFileRepository_InsertFile(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	file := &models.FileInfo{
		Name:        "test.txt",
		Path:        "/test.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/",
	}

	err := repo.InsertFile(file)
	if err != nil {
		t.Fatalf("Failed to insert file: %v", err)
	}

	if file.ID == 0 {
		t.Error("File ID was not set after insert")
	}

	// Verify file was inserted
	var count int
	err = repo.db.QueryRow("SELECT COUNT(*) FROM files WHERE path = ?", file.Path).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count files: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 file, got %d", count)
	}
}

func TestFileRepository_GetFileByPath(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Insert test file
	originalFile := &models.FileInfo{
		Name:        "test.txt",
		Path:        "/documents/test.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/documents",
	}
	
	err := repo.InsertFile(originalFile)
	if err != nil {
		t.Fatalf("Failed to insert test file: %v", err)
	}

	// Retrieve file
	retrievedFile, err := repo.GetFileByPath("/documents/test.txt")
	if err != nil {
		t.Fatalf("Failed to get file by path: %v", err)
	}

	// Verify file data
	if retrievedFile.Name != originalFile.Name {
		t.Errorf("Name mismatch: expected %s, got %s", originalFile.Name, retrievedFile.Name)
	}
	if retrievedFile.Path != originalFile.Path {
		t.Errorf("Path mismatch: expected %s, got %s", originalFile.Path, retrievedFile.Path)
	}
	if retrievedFile.Size != originalFile.Size {
		t.Errorf("Size mismatch: expected %d, got %d", originalFile.Size, retrievedFile.Size)
	}
	if retrievedFile.MimeType != originalFile.MimeType {
		t.Errorf("MimeType mismatch: expected %s, got %s", originalFile.MimeType, retrievedFile.MimeType)
	}
	if retrievedFile.IsDirectory != originalFile.IsDirectory {
		t.Errorf("IsDirectory mismatch: expected %v, got %v", originalFile.IsDirectory, retrievedFile.IsDirectory)
	}
}

func TestFileRepository_GetFileByPath_NotFound(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	_, err := repo.GetFileByPath("/nonexistent.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestFileRepository_GetFilesByPath(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Insert test files
	testFiles := []*models.FileInfo{
		{
			Name:        "file1.txt",
			Path:        "/documents/file1.txt",
			Size:        100,
			MimeType:    "text/plain",
			IsDirectory: false,
			ParentPath:  "/documents",
		},
		{
			Name:        "file2.pdf",
			Path:        "/documents/file2.pdf",
			Size:        200,
			MimeType:    "application/pdf",
			IsDirectory: false,
			ParentPath:  "/documents",
		},
		{
			Name:        "subdir",
			Path:        "/documents/subdir",
			Size:        0,
			MimeType:    "inode/directory",
			IsDirectory: true,
			ParentPath:  "/documents",
		},
	}

	for _, file := range testFiles {
		err := repo.InsertFile(file)
		if err != nil {
			t.Fatalf("Failed to insert test file %s: %v", file.Name, err)
		}
	}

	// Retrieve files
	files, err := repo.GetFilesByPath("/documents")
	if err != nil {
		t.Fatalf("Failed to get files by path: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}

	// Verify ordering (directories first, then alphabetical)
	if !files[0].IsDirectory {
		t.Error("First file should be a directory")
	}
	if files[0].Name != "subdir" {
		t.Errorf("Expected first file to be 'subdir', got %s", files[0].Name)
	}
}

func TestFileRepository_CreateDirectory(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Create parent directory first
	parentDir, err := repo.CreateDirectory("documents", "/")
	if err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}

	// Create subdirectory
	subDir, err := repo.CreateDirectory("subdir", "/documents")
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Verify directory properties
	if subDir.Name != "subdir" {
		t.Errorf("Expected name 'subdir', got %s", subDir.Name)
	}
	if subDir.Path != "/documents/subdir" {
		t.Errorf("Expected path '/documents/subdir', got %s", subDir.Path)
	}
	if !subDir.IsDirectory {
		t.Error("Created item should be a directory")
	}
	if subDir.MimeType != "inode/directory" {
		t.Errorf("Expected mime type 'inode/directory', got %s", subDir.MimeType)
	}
	if subDir.ParentPath != "/documents" {
		t.Errorf("Expected parent path '/documents', got %s", subDir.ParentPath)
	}

	// Verify both directories exist in database
	dirs, err := repo.GetFilesByPath("/")
	if err != nil {
		t.Fatalf("Failed to get root files: %v", err)
	}
	if len(dirs) != 1 || dirs[0].ID != parentDir.ID {
		t.Error("Parent directory not found in root")
	}

	subDirs, err := repo.GetFilesByPath("/documents")
	if err != nil {
		t.Fatalf("Failed to get documents files: %v", err)
	}
	if len(subDirs) != 1 || subDirs[0].ID != subDir.ID {
		t.Error("Subdirectory not found in documents")
	}
}

func TestFileRepository_CreateDirectory_AlreadyExists(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Create directory
	_, err := repo.CreateDirectory("testdir", "/")
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Try to create same directory again
	_, err = repo.CreateDirectory("testdir", "/")
	if err == nil {
		t.Error("Expected error when creating directory that already exists")
	}
}

func TestFileRepository_CreateDirectory_ParentNotExists(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Try to create directory with non-existent parent
	_, err := repo.CreateDirectory("subdir", "/nonexistent")
	if err == nil {
		t.Error("Expected error when parent directory does not exist")
	}
}

func TestFileRepository_RenameFile(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Insert test file
	file := &models.FileInfo{
		Name:        "oldname.txt",
		Path:        "/documents/oldname.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/documents",
	}
	
	err := repo.InsertFile(file)
	if err != nil {
		t.Fatalf("Failed to insert test file: %v", err)
	}

	// Rename file
	err = repo.RenameFile("/documents/oldname.txt", "newname.txt")
	if err != nil {
		t.Fatalf("Failed to rename file: %v", err)
	}

	// Verify old path doesn't exist
	_, err = repo.GetFileByPath("/documents/oldname.txt")
	if err == nil {
		t.Error("Old path should not exist after rename")
	}

	// Verify new path exists
	renamedFile, err := repo.GetFileByPath("/documents/newname.txt")
	if err != nil {
		t.Fatalf("Failed to get renamed file: %v", err)
	}

	if renamedFile.Name != "newname.txt" {
		t.Errorf("Expected name 'newname.txt', got %s", renamedFile.Name)
	}
	if renamedFile.Path != "/documents/newname.txt" {
		t.Errorf("Expected path '/documents/newname.txt', got %s", renamedFile.Path)
	}
}

func TestFileRepository_RenameDirectory_WithChildren(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Create directory structure
	dir := &models.FileInfo{
		Name:        "olddir",
		Path:        "/olddir",
		Size:        0,
		MimeType:    "inode/directory",
		IsDirectory: true,
		ParentPath:  "/",
	}
	err := repo.InsertFile(dir)
	if err != nil {
		t.Fatalf("Failed to insert directory: %v", err)
	}

	// Add child file
	childFile := &models.FileInfo{
		Name:        "child.txt",
		Path:        "/olddir/child.txt",
		Size:        100,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/olddir",
	}
	err = repo.InsertFile(childFile)
	if err != nil {
		t.Fatalf("Failed to insert child file: %v", err)
	}

	// Add nested directory and file
	nestedDir := &models.FileInfo{
		Name:        "nested",
		Path:        "/olddir/nested",
		Size:        0,
		MimeType:    "inode/directory",
		IsDirectory: true,
		ParentPath:  "/olddir",
	}
	err = repo.InsertFile(nestedDir)
	if err != nil {
		t.Fatalf("Failed to insert nested directory: %v", err)
	}

	nestedFile := &models.FileInfo{
		Name:        "nested.txt",
		Path:        "/olddir/nested/nested.txt",
		Size:        200,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/olddir/nested",
	}
	err = repo.InsertFile(nestedFile)
	if err != nil {
		t.Fatalf("Failed to insert nested file: %v", err)
	}

	// Rename directory
	err = repo.RenameFile("/olddir", "newdir")
	if err != nil {
		t.Fatalf("Failed to rename directory: %v", err)
	}

	// Verify old paths don't exist
	_, err = repo.GetFileByPath("/olddir")
	if err == nil {
		t.Error("Old directory path should not exist")
	}

	// Verify new paths exist
	newDir, err := repo.GetFileByPath("/newdir")
	if err != nil {
		t.Fatalf("Failed to get renamed directory: %v", err)
	}
	if newDir.Name != "newdir" {
		t.Errorf("Expected directory name 'newdir', got %s", newDir.Name)
	}

	// Verify child paths were updated
	newChildFile, err := repo.GetFileByPath("/newdir/child.txt")
	if err != nil {
		t.Fatalf("Failed to get child file after rename: %v", err)
	}
	if newChildFile.ParentPath != "/newdir" {
		t.Errorf("Expected child parent path '/newdir', got %s", newChildFile.ParentPath)
	}

	// Verify nested paths were updated
	newNestedFile, err := repo.GetFileByPath("/newdir/nested/nested.txt")
	if err != nil {
		t.Fatalf("Failed to get nested file after rename: %v", err)
	}
	if newNestedFile.ParentPath != "/newdir/nested" {
		t.Errorf("Expected nested file parent path '/newdir/nested', got %s", newNestedFile.ParentPath)
	}
}

func TestFileRepository_MoveFile(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Create source and destination directories
	sourceDir := &models.FileInfo{
		Name: "source", Path: "/source", MimeType: "inode/directory",
		IsDirectory: true, ParentPath: "/",
	}
	err := repo.InsertFile(sourceDir)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	destDir := &models.FileInfo{
		Name: "dest", Path: "/dest", MimeType: "inode/directory",
		IsDirectory: true, ParentPath: "/",
	}
	err = repo.InsertFile(destDir)
	if err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}

	// Create file to move
	file := &models.FileInfo{
		Name:        "moveme.txt",
		Path:        "/source/moveme.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/source",
	}
	err = repo.InsertFile(file)
	if err != nil {
		t.Fatalf("Failed to create file to move: %v", err)
	}

	// Move file
	err = repo.MoveFile("/source/moveme.txt", "/dest")
	if err != nil {
		t.Fatalf("Failed to move file: %v", err)
	}

	// Verify old path doesn't exist
	_, err = repo.GetFileByPath("/source/moveme.txt")
	if err == nil {
		t.Error("Old path should not exist after move")
	}

	// Verify new path exists
	movedFile, err := repo.GetFileByPath("/dest/moveme.txt")
	if err != nil {
		t.Fatalf("Failed to get moved file: %v", err)
	}

	if movedFile.Name != "moveme.txt" {
		t.Errorf("Expected name 'moveme.txt', got %s", movedFile.Name)
	}
	if movedFile.Path != "/dest/moveme.txt" {
		t.Errorf("Expected path '/dest/moveme.txt', got %s", movedFile.Path)
	}
	if movedFile.ParentPath != "/dest" {
		t.Errorf("Expected parent path '/dest', got %s", movedFile.ParentPath)
	}
}

func TestFileRepository_DeleteFile(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Insert test file
	file := &models.FileInfo{
		Name:        "delete_me.txt",
		Path:        "/delete_me.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/",
	}
	
	err := repo.InsertFile(file)
	if err != nil {
		t.Fatalf("Failed to insert test file: %v", err)
	}

	// Delete file
	err = repo.DeleteFile("/delete_me.txt")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file no longer exists
	_, err = repo.GetFileByPath("/delete_me.txt")
	if err == nil {
		t.Error("File should not exist after deletion")
	}
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestFileRepository_DeleteDirectory_Empty(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Create empty directory
	dir := &models.FileInfo{
		Name:        "empty_dir",
		Path:        "/empty_dir",
		Size:        0,
		MimeType:    "inode/directory",
		IsDirectory: true,
		ParentPath:  "/",
	}
	
	err := repo.InsertFile(dir)
	if err != nil {
		t.Fatalf("Failed to insert test directory: %v", err)
	}

	// Delete directory
	err = repo.DeleteFile("/empty_dir")
	if err != nil {
		t.Fatalf("Failed to delete empty directory: %v", err)
	}

	// Verify directory no longer exists
	_, err = repo.GetFileByPath("/empty_dir")
	if err == nil {
		t.Error("Directory should not exist after deletion")
	}
}

func TestFileRepository_DeleteDirectory_NotEmpty(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Create directory
	dir := &models.FileInfo{
		Name:        "nonempty_dir",
		Path:        "/nonempty_dir",
		Size:        0,
		MimeType:    "inode/directory",
		IsDirectory: true,
		ParentPath:  "/",
	}
	err := repo.InsertFile(dir)
	if err != nil {
		t.Fatalf("Failed to insert test directory: %v", err)
	}

	// Add file to directory
	file := &models.FileInfo{
		Name:        "child.txt",
		Path:        "/nonempty_dir/child.txt",
		Size:        100,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/nonempty_dir",
	}
	err = repo.InsertFile(file)
	if err != nil {
		t.Fatalf("Failed to insert child file: %v", err)
	}

	// Try to delete non-empty directory
	err = repo.DeleteFile("/nonempty_dir")
	if err == nil {
		t.Error("Should not be able to delete non-empty directory")
	}

	// Verify directory still exists
	_, err = repo.GetFileByPath("/nonempty_dir")
	if err != nil {
		t.Error("Directory should still exist after failed deletion")
	}
}

func TestFileRepository_DeleteDirectoryRecursive(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Create directory structure
	dir := &models.FileInfo{
		Name: "recursive_test", Path: "/recursive_test", MimeType: "inode/directory",
		IsDirectory: true, ParentPath: "/",
	}
	err := repo.InsertFile(dir)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Add files and subdirectories
	file1 := &models.FileInfo{
		Name: "file1.txt", Path: "/recursive_test/file1.txt", Size: 100,
		MimeType: "text/plain", IsDirectory: false, ParentPath: "/recursive_test",
	}
	err = repo.InsertFile(file1)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	subdir := &models.FileInfo{
		Name: "subdir", Path: "/recursive_test/subdir", MimeType: "inode/directory",
		IsDirectory: true, ParentPath: "/recursive_test",
	}
	err = repo.InsertFile(subdir)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	file2 := &models.FileInfo{
		Name: "file2.txt", Path: "/recursive_test/subdir/file2.txt", Size: 200,
		MimeType: "text/plain", IsDirectory: false, ParentPath: "/recursive_test/subdir",
	}
	err = repo.InsertFile(file2)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Delete recursively
	err = repo.DeleteDirectoryRecursive("/recursive_test")
	if err != nil {
		t.Fatalf("Failed to delete directory recursively: %v", err)
	}

	// Verify all files and directories are gone
	paths := []string{
		"/recursive_test",
		"/recursive_test/file1.txt",
		"/recursive_test/subdir",
		"/recursive_test/subdir/file2.txt",
	}

	for _, path := range paths {
		_, err = repo.GetFileByPath(path)
		if err == nil {
			t.Errorf("Path %s should not exist after recursive deletion", path)
		}
	}
}

func TestFileRepository_BuildPath(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	tests := []struct {
		parent   string
		name     string
		expected string
	}{
		{"/", "file.txt", "/file.txt"},
		{"/documents", "file.txt", "/documents/file.txt"},
		{"/documents/photos", "image.jpg", "/documents/photos/image.jpg"},
	}

	for _, tt := range tests {
		result := repo.buildPath(tt.parent, tt.name)
		if result != tt.expected {
			t.Errorf("buildPath(%q, %q) = %q, want %q", tt.parent, tt.name, result, tt.expected)
		}
	}
}

func TestFileRepository_PathExists(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Insert test file
	file := &models.FileInfo{
		Name:        "exists.txt",
		Path:        "/exists.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/",
	}
	
	err := repo.InsertFile(file)
	if err != nil {
		t.Fatalf("Failed to insert test file: %v", err)
	}

	// Test existing path
	if !repo.pathExists("/exists.txt") {
		t.Error("pathExists should return true for existing file")
	}

	// Test non-existing path
	if repo.pathExists("/nonexistent.txt") {
		t.Error("pathExists should return false for non-existing file")
	}
}

func TestFileRepository_IsDirectory(t *testing.T) {
	repo := setupTestRepository(t)
	defer repo.Close()

	// Insert test file
	file := &models.FileInfo{
		Name:        "file.txt",
		Path:        "/file.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/",
	}
	err := repo.InsertFile(file)
	if err != nil {
		t.Fatalf("Failed to insert test file: %v", err)
	}

	// Insert test directory
	dir := &models.FileInfo{
		Name:        "directory",
		Path:        "/directory",
		Size:        0,
		MimeType:    "inode/directory",
		IsDirectory: true,
		ParentPath:  "/",
	}
	err = repo.InsertFile(dir)
	if err != nil {
		t.Fatalf("Failed to insert test directory: %v", err)
	}

	// Test file
	if repo.isDirectory("/file.txt") {
		t.Error("isDirectory should return false for file")
	}

	// Test directory
	if !repo.isDirectory("/directory") {
		t.Error("isDirectory should return true for directory")
	}

	// Test non-existing path
	if repo.isDirectory("/nonexistent") {
		t.Error("isDirectory should return false for non-existing path")
	}
}

// Benchmark tests
func BenchmarkFileRepository_InsertFile(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "benchmark.db")
	
	repo, err := NewFileRepository(dbPath, 10)
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file := &models.FileInfo{
			Name:        fmt.Sprintf("file%d.txt", i),
			Path:        fmt.Sprintf("/file%d.txt", i),
			Size:        1024,
			MimeType:    "text/plain",
			IsDirectory: false,
			ParentPath:  "/",
		}
		
		err := repo.InsertFile(file)
		if err != nil {
			b.Fatalf("Failed to insert file: %v", err)
		}
	}
}

func BenchmarkFileRepository_GetFileByPath(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "benchmark.db")
	
	repo, err := NewFileRepository(dbPath, 10)
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Insert test file
	file := &models.FileInfo{
		Name:        "benchmark.txt",
		Path:        "/benchmark.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/",
	}
	err = repo.InsertFile(file)
	if err != nil {
		b.Fatalf("Failed to insert test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetFileByPath("/benchmark.txt")
		if err != nil {
			b.Fatalf("Failed to get file: %v", err)
		}
	}
}

func BenchmarkFileRepository_GetFilesByPath(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "benchmark.db")
	
	repo, err := NewFileRepository(dbPath, 10)
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Insert test files
	for i := 0; i < 100; i++ {
		file := &models.FileInfo{
			Name:        fmt.Sprintf("file%d.txt", i),
			Path:        fmt.Sprintf("/documents/file%d.txt", i),
			Size:        1024,
			MimeType:    "text/plain",
			IsDirectory: false,
			ParentPath:  "/documents",
		}
		err := repo.InsertFile(file)
		if err != nil {
			b.Fatalf("Failed to insert test file: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetFilesByPath("/documents")
		if err != nil {
			b.Fatalf("Failed to get files: %v", err)
		}
	}
}