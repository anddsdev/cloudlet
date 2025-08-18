package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/security"
)

// Mock repository for testing
type MockFileRepository struct {
	files      map[string]*models.FileInfo
	nextID     int64
	shouldFail bool
	failError  error
}

func NewMockFileRepository() *MockFileRepository {
	return &MockFileRepository{
		files:  make(map[string]*models.FileInfo),
		nextID: 1,
	}
}

func (m *MockFileRepository) GetFilesByPath(parentPath string) ([]*models.FileInfo, error) {
	if m.shouldFail {
		return nil, m.failError
	}

	var files []*models.FileInfo
	for _, file := range m.files {
		if file.ParentPath == parentPath {
			files = append(files, file)
		}
	}
	return files, nil
}

func (m *MockFileRepository) GetFileByPath(path string) (*models.FileInfo, error) {
	if m.shouldFail {
		return nil, m.failError
	}

	file, exists := m.files[path]
	if !exists {
		return nil, fmt.Errorf("file not found")
	}
	return file, nil
}

func (m *MockFileRepository) InsertFile(file *models.FileInfo) error {
	if m.shouldFail {
		return m.failError
	}

	file.ID = m.nextID
	m.nextID++
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()
	m.files[file.Path] = file
	return nil
}

func (m *MockFileRepository) CreateDirectory(name, parentPath string) (*models.FileInfo, error) {
	if m.shouldFail {
		return nil, m.failError
	}

	fullPath := buildMockPath(parentPath, name)
	if _, exists := m.files[fullPath]; exists {
		return nil, fmt.Errorf("directory already exists")
	}

	dir := &models.FileInfo{
		ID:          m.nextID,
		Name:        name,
		Path:        fullPath,
		Size:        0,
		MimeType:    "inode/directory",
		IsDirectory: true,
		ParentPath:  parentPath,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.nextID++
	m.files[fullPath] = dir
	return dir, nil
}

func (m *MockFileRepository) RenameFile(oldPath, newName string) error {
	if m.shouldFail {
		return m.failError
	}

	file, exists := m.files[oldPath]
	if !exists {
		return fmt.Errorf("file not found")
	}

	delete(m.files, oldPath)
	file.Name = newName
	file.Path = buildMockPath(file.ParentPath, newName)
	file.UpdatedAt = time.Now()
	m.files[file.Path] = file
	return nil
}

func (m *MockFileRepository) MoveFile(sourcePath, destinationPath string) error {
	if m.shouldFail {
		return m.failError
	}

	file, exists := m.files[sourcePath]
	if !exists {
		return fmt.Errorf("file not found")
	}

	delete(m.files, sourcePath)
	file.ParentPath = destinationPath
	file.Path = buildMockPath(destinationPath, file.Name)
	file.UpdatedAt = time.Now()
	m.files[file.Path] = file
	return nil
}

func (m *MockFileRepository) DeleteFile(path string) error {
	if m.shouldFail {
		return m.failError
	}

	_, exists := m.files[path]
	if !exists {
		return fmt.Errorf("file not found")
	}

	delete(m.files, path)
	return nil
}

func buildMockPath(parent, name string) string {
	if parent == "/" {
		return "/" + name
	}
	return parent + "/" + name
}

// Mock storage service for testing
type MockStorageService struct {
	files      map[string][]byte
	shouldFail bool
	failError  error
}

func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		files: make(map[string][]byte),
	}
}

func (m *MockStorageService) SaveFile(path string, data []byte) error {
	if m.shouldFail {
		return m.failError
	}
	m.files[path] = data
	return nil
}

func (m *MockStorageService) SaveFileStream(path string, reader io.Reader) error {
	if m.shouldFail {
		return m.failError
	}

	// For mock, we'll just store empty data
	m.files[path] = []byte("stream_data")
	return nil
}

func (m *MockStorageService) ReadFile(path string) ([]byte, error) {
	if m.shouldFail {
		return nil, m.failError
	}

	data, exists := m.files[path]
	if !exists {
		return nil, fmt.Errorf("file not found in storage")
	}
	return data, nil
}

func (m *MockStorageService) CreateDirectory(path string) error {
	if m.shouldFail {
		return m.failError
	}
	// For mock, we don't need to actually create directories
	return nil
}

func (m *MockStorageService) MoveFile(sourcePath, destPath string) error {
	if m.shouldFail {
		return m.failError
	}

	data, exists := m.files[sourcePath]
	if !exists {
		return fmt.Errorf("source file not found in storage")
	}

	m.files[destPath] = data
	delete(m.files, sourcePath)
	return nil
}

func (m *MockStorageService) DeleteFile(path string) error {
	if m.shouldFail {
		return m.failError
	}

	_, exists := m.files[path]
	if !exists {
		return fmt.Errorf("file not found in storage")
	}

	delete(m.files, path)
	return nil
}

// TestableFileService implements the FileService methods needed for testing
type TestableFileService struct {
	repo          *MockFileRepository
	storage       *MockStorageService
	pathValidator *MockPathValidator
}

// Implement the key FileService methods for testing
func (t *TestableFileService) GetDirectoryListing(path string) (*models.DirectoryListing, error) {
	validatedPath, err := t.pathValidator.ValidateAndNormalizePath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	path = validatedPath

	files, err := t.repo.GetFilesByPath(path)
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

	breadcrumbs := t.generateBreadcrumbs(path)
	parentPath := t.getParentPath(path)

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

func (t *TestableFileService) CreateDirectory(name, parentPath string) (*models.FileInfo, error) {
	// Validate filename
	if err := security.IsValidFilename(name); err != nil {
		return nil, fmt.Errorf("invalid directory name: %w", err)
	}

	validatedParentPath, err := t.pathValidator.ValidateAndNormalizePath(parentPath)
	if err != nil {
		return nil, fmt.Errorf("invalid parent path: %w", err)
	}
	parentPath = validatedParentPath

	if !t.isValidName(name) {
		return nil, errors.New("invalid directory name")
	}

	dir, err := t.repo.CreateDirectory(name, parentPath)
	if err != nil {
		return nil, err
	}

	err = t.storage.CreateDirectory(dir.Path)
	if err != nil {
		t.repo.DeleteFile(dir.Path)
		return nil, err
	}

	return dir, nil
}

func (t *TestableFileService) SaveFile(filename, parentPath string, data []byte) error {
	// Validate filename
	if err := security.IsValidFilename(filename); err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}

	validatedParentPath, err := t.pathValidator.ValidateAndNormalizePath(parentPath)
	if err != nil {
		return fmt.Errorf("invalid parent path: %w", err)
	}
	parentPath = validatedParentPath
	fullPath := t.buildPath(parentPath, filename)

	if err := t.storage.SaveFile(fullPath, data); err != nil {
		return err
	}

	file := &models.FileInfo{
		Name:        filename,
		Path:        fullPath,
		Size:        int64(len(data)),
		MimeType:    t.detectMimeType(filename),
		IsDirectory: false,
		ParentPath:  parentPath,
	}

	return t.repo.InsertFile(file)
}

func (t *TestableFileService) SaveFileStream(filename, parentPath string, reader io.Reader, size int64) error {
	// Validate filename
	if err := security.IsValidFilename(filename); err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}

	validatedParentPath, err := t.pathValidator.ValidateAndNormalizePath(parentPath)
	if err != nil {
		return fmt.Errorf("invalid parent path: %w", err)
	}
	parentPath = validatedParentPath
	fullPath := t.buildPath(parentPath, filename)

	if err := t.storage.SaveFileStream(fullPath, reader); err != nil {
		return err
	}

	file := &models.FileInfo{
		Name:        filename,
		Path:        fullPath,
		Size:        size,
		MimeType:    t.detectMimeType(filename),
		IsDirectory: false,
		ParentPath:  parentPath,
	}

	return t.repo.InsertFile(file)
}

func (t *TestableFileService) GetFileData(path string) ([]byte, *models.FileInfo, error) {
	validatedPath, err := t.pathValidator.ValidateAndNormalizePath(path)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid path: %w", err)
	}
	path = validatedPath

	fileInfo, err := t.repo.GetFileByPath(path)
	if err != nil {
		return nil, nil, err
	}

	if fileInfo.IsDirectory {
		return nil, nil, errors.New("cannot download directory")
	}

	data, err := t.storage.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	return data, fileInfo, nil
}

func (t *TestableFileService) RenameFile(oldPath, newName string) error {
	return t.repo.RenameFile(oldPath, newName)
}

func (t *TestableFileService) MoveFile(sourcePath, destinationPath string) error {
	err := t.repo.MoveFile(sourcePath, destinationPath)
	if err != nil {
		return err
	}

	// Build new path for storage move
	fileInfo, err := t.repo.GetFileByPath(t.buildPath(destinationPath, t.getFilenameFromPath(sourcePath)))
	if err != nil {
		return err
	}

	return t.storage.MoveFile(sourcePath, fileInfo.Path)
}

func (t *TestableFileService) DeleteFile(path string) error {
	err := t.repo.DeleteFile(path)
	if err != nil {
		return err
	}
	return t.storage.DeleteFile(path)
}

// Helper methods copied from FileService
func (t *TestableFileService) buildPath(parent, name string) string {
	if parent == "/" {
		return "/" + name
	}
	return parent + "/" + name
}

func (t *TestableFileService) getParentPath(path string) string {
	if path == "/" {
		return ""
	}

	parent := filepath.Dir(path)
	if parent == "." {
		return "/"
	}

	// Convert backslashes to forward slashes for consistent path handling
	parent = strings.ReplaceAll(parent, "\\", "/")

	return parent
}

func (t *TestableFileService) generateBreadcrumbs(path string) []models.Breadcrumb {
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

func (t *TestableFileService) isValidName(name string) bool {
	// Simple validation for testing - reject names with special chars
	return !strings.ContainsAny(name, "/\\:*?\"<>|") && name != "" && 
		   !strings.Contains(name, "..") && name != "CON" && name != "PRN" &&
		   !strings.ContainsAny(name, "\x00")
}

func (t *TestableFileService) detectMimeType(filename string) string {
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

func (t *TestableFileService) getFilenameFromPath(path string) string {
	return filepath.Base(path)
}

// Helper function to create test file service with mocks
func createTestFileService() (*TestableFileService, *MockFileRepository, *MockStorageService) {
	mockRepo := NewMockFileRepository()
	mockStorage := NewMockStorageService()
	mockPathValidator := &MockPathValidator{}
	
	service := &TestableFileService{
		repo:          mockRepo,
		storage:       mockStorage,
		pathValidator: mockPathValidator,
	}
	
	return service, mockRepo, mockStorage
}

// Mock path validator for testing
type MockPathValidator struct{}

func (m *MockPathValidator) ValidateAndNormalizePath(path string) (string, error) {
	// Simple validation for tests - reject paths with ../
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path traversal detected")
	}
	return path, nil
}

func TestFileService_GetDirectoryListing(t *testing.T) {
	service, mockRepo, _ := createTestFileService()

	// Setup test data
	mockRepo.files["/documents/file1.txt"] = &models.FileInfo{
		ID: 1, Name: "file1.txt", Path: "/documents/file1.txt",
		Size: 100, MimeType: "text/plain", IsDirectory: false, ParentPath: "/documents",
	}
	mockRepo.files["/documents/file2.pdf"] = &models.FileInfo{
		ID: 2, Name: "file2.pdf", Path: "/documents/file2.pdf",
		Size: 200, MimeType: "application/pdf", IsDirectory: false, ParentPath: "/documents",
	}
	mockRepo.files["/documents/subdir"] = &models.FileInfo{
		ID: 3, Name: "subdir", Path: "/documents/subdir",
		Size: 0, MimeType: "inode/directory", IsDirectory: true, ParentPath: "/documents",
	}

	listing, err := service.GetDirectoryListing("/documents")
	if err != nil {
		t.Fatalf("Failed to get directory listing: %v", err)
	}

	if listing.Path != "/documents" {
		t.Errorf("Expected path '/documents', got %s", listing.Path)
	}

	if len(listing.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(listing.Files))
	}

	if len(listing.Directories) != 1 {
		t.Errorf("Expected 1 directory, got %d", len(listing.Directories))
	}

	if listing.TotalFiles != 2 {
		t.Errorf("Expected TotalFiles=2, got %d", listing.TotalFiles)
	}

	if listing.TotalDirs != 1 {
		t.Errorf("Expected TotalDirs=1, got %d", listing.TotalDirs)
	}

	if listing.TotalSize != 300 {
		t.Errorf("Expected TotalSize=300, got %d", listing.TotalSize)
	}
}

func TestFileService_GetDirectoryListing_InvalidPath(t *testing.T) {
	service, _, _ := createTestFileService()

	// Test with invalid path
	_, err := service.GetDirectoryListing("../../../etc/passwd")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestFileService_CreateDirectory(t *testing.T) {
	service, mockRepo, _ := createTestFileService()

	// Setup parent directory
	mockRepo.files["/documents"] = &models.FileInfo{
		ID: 1, Name: "documents", Path: "/documents",
		Size: 0, MimeType: "inode/directory", IsDirectory: true, ParentPath: "/",
	}

	dir, err := service.CreateDirectory("new_folder", "/documents")
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if dir.Name != "new_folder" {
		t.Errorf("Expected name 'new_folder', got %s", dir.Name)
	}

	if dir.Path != "/documents/new_folder" {
		t.Errorf("Expected path '/documents/new_folder', got %s", dir.Path)
	}

	if !dir.IsDirectory {
		t.Error("Created item should be a directory")
	}
}

func TestFileService_CreateDirectory_InvalidName(t *testing.T) {
	service, _, _ := createTestFileService()

	invalidNames := []string{
		"", // empty
		"folder/with/slash",
		"folder\x00with\x00null",
		"CON", // reserved name
		"folder*.txt",
	}

	for _, name := range invalidNames {
		_, err := service.CreateDirectory(name, "/")
		if err == nil {
			t.Errorf("Expected error for invalid directory name: %s", name)
		}
	}
}

func TestFileService_SaveFile(t *testing.T) {
	service, mockRepo, mockStorage := createTestFileService()
	data := []byte("test file content")

	err := service.SaveFile("test.txt", "/", data)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Verify file was saved in repository
	file, exists := mockRepo.files["/test.txt"]
	if !exists {
		t.Error("File was not saved in repository")
	}

	if file.Name != "test.txt" {
		t.Errorf("Expected name 'test.txt', got %s", file.Name)
	}

	if file.Size != int64(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), file.Size)
	}

	// Verify file was saved in storage
	storedData, exists := mockStorage.files["/test.txt"]
	if !exists {
		t.Error("File was not saved in storage")
	}

	if !bytes.Equal(storedData, data) {
		t.Error("Stored data does not match original")
	}
}

func TestFileService_SaveFile_InvalidFilename(t *testing.T) {
	service, _, _ := createTestFileService()
	data := []byte("test content")

	invalidNames := []string{
		"", // empty
		"file/with/slash.txt",
		"file\x00with\x00null.txt",
		"CON.txt", // reserved name
		"file*.txt",
	}

	for _, name := range invalidNames {
		err := service.SaveFile(name, "/", data)
		if err == nil {
			t.Errorf("Expected error for invalid filename: %s", name)
		}
	}
}

func TestFileService_SaveFileStream(t *testing.T) {
	service, mockRepo, _ := createTestFileService()
	reader := strings.NewReader("stream test content")

	err := service.SaveFileStream("stream.txt", "/", reader, 100)
	if err != nil {
		t.Fatalf("Failed to save file stream: %v", err)
	}

	// Verify file was saved in repository
	file, exists := mockRepo.files["/stream.txt"]
	if !exists {
		t.Error("File was not saved in repository")
	}

	if file.Name != "stream.txt" {
		t.Errorf("Expected name 'stream.txt', got %s", file.Name)
	}

	if file.Size != 100 {
		t.Errorf("Expected size 100, got %d", file.Size)
	}
}

func TestFileService_GetFileData(t *testing.T) {
	service, mockRepo, mockStorage := createTestFileService()

	// Setup test file
	testData := []byte("test file content")
	mockRepo.files["/test.txt"] = &models.FileInfo{
		ID: 1, Name: "test.txt", Path: "/test.txt",
		Size: int64(len(testData)), MimeType: "text/plain", IsDirectory: false, ParentPath: "/",
	}
	mockStorage.files["/test.txt"] = testData

	data, fileInfo, err := service.GetFileData("/test.txt")
	if err != nil {
		t.Fatalf("Failed to get file data: %v", err)
	}

	if !bytes.Equal(data, testData) {
		t.Error("Retrieved data does not match stored data")
	}

	if fileInfo.Name != "test.txt" {
		t.Errorf("Expected file name 'test.txt', got %s", fileInfo.Name)
	}
}

func TestFileService_GetFileData_Directory(t *testing.T) {
	service, mockRepo, _ := createTestFileService()

	// Setup test directory
	mockRepo.files["/testdir"] = &models.FileInfo{
		ID: 1, Name: "testdir", Path: "/testdir",
		Size: 0, MimeType: "inode/directory", IsDirectory: true, ParentPath: "/",
	}

	_, _, err := service.GetFileData("/testdir")
	if err == nil {
		t.Error("Expected error when trying to download directory")
	}
}

func TestFileService_RenameFile(t *testing.T) {
	service, mockRepo, mockStorage := createTestFileService()

	// Setup test file
	mockRepo.files["/oldname.txt"] = &models.FileInfo{
		ID: 1, Name: "oldname.txt", Path: "/oldname.txt",
		Size: 100, MimeType: "text/plain", IsDirectory: false, ParentPath: "/",
	}
	mockStorage.files["/oldname.txt"] = []byte("test content")

	err := service.RenameFile("/oldname.txt", "newname.txt")
	if err != nil {
		t.Fatalf("Failed to rename file: %v", err)
	}

	// Verify old file doesn't exist
	_, exists := mockRepo.files["/oldname.txt"]
	if exists {
		t.Error("Old file should not exist in repository after rename")
	}

	// Verify new file exists
	newFile, exists := mockRepo.files["/newname.txt"]
	if !exists {
		t.Error("New file should exist in repository after rename")
	}

	if newFile.Name != "newname.txt" {
		t.Errorf("Expected new name 'newname.txt', got %s", newFile.Name)
	}
}

func TestFileService_MoveFile(t *testing.T) {
	service, mockRepo, mockStorage := createTestFileService()

	// Setup test file and destination directory
	mockRepo.files["/source/file.txt"] = &models.FileInfo{
		ID: 1, Name: "file.txt", Path: "/source/file.txt",
		Size: 100, MimeType: "text/plain", IsDirectory: false, ParentPath: "/source",
	}
	mockRepo.files["/dest"] = &models.FileInfo{
		ID: 2, Name: "dest", Path: "/dest",
		Size: 0, MimeType: "inode/directory", IsDirectory: true, ParentPath: "/",
	}
	mockStorage.files["/source/file.txt"] = []byte("test content")

	err := service.MoveFile("/source/file.txt", "/dest")
	if err != nil {
		t.Fatalf("Failed to move file: %v", err)
	}

	// Verify old file doesn't exist
	_, exists := mockRepo.files["/source/file.txt"]
	if exists {
		t.Error("Old file should not exist in repository after move")
	}

	// Verify new file exists
	movedFile, exists := mockRepo.files["/dest/file.txt"]
	if !exists {
		t.Error("Moved file should exist in repository")
	}

	if movedFile.ParentPath != "/dest" {
		t.Errorf("Expected parent path '/dest', got %s", movedFile.ParentPath)
	}
}

func TestFileService_DeleteFile(t *testing.T) {
	service, mockRepo, mockStorage := createTestFileService()

	// Setup test file
	mockRepo.files["/delete_me.txt"] = &models.FileInfo{
		ID: 1, Name: "delete_me.txt", Path: "/delete_me.txt",
		Size: 100, MimeType: "text/plain", IsDirectory: false, ParentPath: "/",
	}
	mockStorage.files["/delete_me.txt"] = []byte("test content")

	err := service.DeleteFile("/delete_me.txt")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file doesn't exist in repository
	_, exists := mockRepo.files["/delete_me.txt"]
	if exists {
		t.Error("File should not exist in repository after deletion")
	}

	// Verify file doesn't exist in storage
	_, exists = mockStorage.files["/delete_me.txt"]
	if exists {
		t.Error("File should not exist in storage after deletion")
	}
}

func TestFileService_DetectMimeType(t *testing.T) {
	service, _, _ := createTestFileService()

	tests := []struct {
		filename string
		expected string
	}{
		{"document.txt", "text/plain"},
		{"image.jpg", "image/jpeg"},
		{"image.png", "image/png"},
		{"document.pdf", "application/pdf"},
		{"archive.zip", "application/zip"},
		{"script.js", "text/javascript"},
		{"data.json", "application/json"},
		{"style.css", "text/css"},
		{"page.html", "text/html"},
		{"unknown.xyz", "application/octet-stream"},
		{"dangerous.exe", "application/x-executable"},
		{"script.bat", "application/x-executable"},
	}

	for _, tt := range tests {
		result := service.detectMimeType(tt.filename)
		if result != tt.expected {
			t.Errorf("detectMimeType(%s) = %s, want %s", tt.filename, result, tt.expected)
		}
	}
}

func TestFileService_GenerateBreadcrumbs(t *testing.T) {
	service, _, _ := createTestFileService()

	tests := []struct {
		path     string
		expected []models.Breadcrumb
	}{
		{
			path: "/",
			expected: []models.Breadcrumb{
				{Name: "Home", Path: "/"},
			},
		},
		{
			path: "/documents",
			expected: []models.Breadcrumb{
				{Name: "Home", Path: "/"},
				{Name: "documents", Path: "/documents"},
			},
		},
		{
			path: "/documents/photos/vacation",
			expected: []models.Breadcrumb{
				{Name: "Home", Path: "/"},
				{Name: "documents", Path: "/documents"},
				{Name: "photos", Path: "/documents/photos"},
				{Name: "vacation", Path: "/documents/photos/vacation"},
			},
		},
	}

	for _, tt := range tests {
		result := service.generateBreadcrumbs(tt.path)
		if len(result) != len(tt.expected) {
			t.Errorf("generateBreadcrumbs(%s) returned %d breadcrumbs, want %d", 
				tt.path, len(result), len(tt.expected))
			continue
		}

		for i, breadcrumb := range result {
			if breadcrumb.Name != tt.expected[i].Name || breadcrumb.Path != tt.expected[i].Path {
				t.Errorf("generateBreadcrumbs(%s)[%d] = %+v, want %+v", 
					tt.path, i, breadcrumb, tt.expected[i])
			}
		}
	}
}

func TestFileService_BuildPath(t *testing.T) {
	service, _, _ := createTestFileService()

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
		result := service.buildPath(tt.parent, tt.name)
		if result != tt.expected {
			t.Errorf("buildPath(%s, %s) = %s, want %s", tt.parent, tt.name, result, tt.expected)
		}
	}
}

func TestFileService_GetParentPath(t *testing.T) {
	service, _, _ := createTestFileService()

	tests := []struct {
		path     string
		expected string
	}{
		{"/", ""},
		{"/file.txt", "/"},
		{"/documents/file.txt", "/documents"},
		{"/documents/photos/image.jpg", "/documents/photos"},
	}

	for _, tt := range tests {
		result := service.getParentPath(tt.path)
		if result != tt.expected {
			t.Errorf("getParentPath(%s) = %s, want %s", tt.path, result, tt.expected)
		}
	}
}

// Error handling tests
func TestFileService_ErrorHandling(t *testing.T) {
	service, mockRepo, mockStorage := createTestFileService()

	t.Run("Repository failure", func(t *testing.T) {
		mockRepo.shouldFail = true
		mockRepo.failError = errors.New("database connection failed")

		_, err := service.GetDirectoryListing("/")
		if err == nil {
			t.Error("Expected error when repository fails")
		}

		mockRepo.shouldFail = false
	})

	t.Run("Storage failure", func(t *testing.T) {
		mockStorage.shouldFail = true
		mockStorage.failError = errors.New("storage system unavailable")

		err := service.SaveFile("test.txt", "/", []byte("data"))
		if err == nil {
			t.Error("Expected error when storage fails")
		}

		mockStorage.shouldFail = false
	})
}

// Benchmark tests
func BenchmarkFileService_GetDirectoryListing(b *testing.B) {
	service, mockRepo, _ := createTestFileService()

	// Setup test data
	for i := 0; i < 100; i++ {
		mockRepo.files[fmt.Sprintf("/documents/file%d.txt", i)] = &models.FileInfo{
			ID: int64(i), Name: fmt.Sprintf("file%d.txt", i), 
			Path: fmt.Sprintf("/documents/file%d.txt", i),
			Size: 1024, MimeType: "text/plain", IsDirectory: false, ParentPath: "/documents",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetDirectoryListing("/documents")
		if err != nil {
			b.Fatalf("Failed to get directory listing: %v", err)
		}
	}
}

func BenchmarkFileService_DetectMimeType(b *testing.B) {
	service, _, _ := createTestFileService()
	filename := "document.pdf"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.detectMimeType(filename)
	}
}

func BenchmarkFileService_GenerateBreadcrumbs(b *testing.B) {
	service, _, _ := createTestFileService()
	path := "/documents/photos/vacation/summer/2023"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.generateBreadcrumbs(path)
	}
}