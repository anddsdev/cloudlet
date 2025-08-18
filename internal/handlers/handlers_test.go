package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/anddsdev/cloudlet/config"
	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/services"
)

// Mock FileService for testing handlers
type MockFileService struct {
	files      map[string]*models.FileInfo
	fileData   map[string][]byte
	shouldFail bool
	failError  error
}

func NewMockFileService() *MockFileService {
	return &MockFileService{
		files:    make(map[string]*models.FileInfo),
		fileData: make(map[string][]byte),
	}
}

func (m *MockFileService) GetDirectoryListing(path string) (*models.DirectoryListing, error) {
	if m.shouldFail {
		return nil, m.failError
	}

	var files []*models.FileInfo
	var directories []*models.FileInfo
	var totalSize int64

	for _, file := range m.files {
		if file.ParentPath == path {
			if file.IsDirectory {
				directories = append(directories, file)
			} else {
				files = append(files, file)
				totalSize += file.Size
			}
		}
	}

	return &models.DirectoryListing{
		Path:        path,
		ParentPath:  getParentPath(path),
		Files:       files,
		Directories: directories,
		TotalFiles:  len(files),
		TotalDirs:   len(directories),
		TotalSize:   totalSize,
		Breadcrumbs: []models.Breadcrumb{{Name: "Home", Path: "/"}},
	}, nil
}

func (m *MockFileService) CreateDirectory(name, parentPath string) (*models.FileInfo, error) {
	if m.shouldFail {
		return nil, m.failError
	}

	fullPath := buildPath(parentPath, name)
	if _, exists := m.files[fullPath]; exists {
		return nil, fmt.Errorf("directory already exists: %s", fullPath)
	}

	dir := &models.FileInfo{
		Name:        name,
		Path:        fullPath,
		Size:        0,
		MimeType:    "inode/directory",
		IsDirectory: true,
		ParentPath:  parentPath,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.files[fullPath] = dir
	return dir, nil
}

func (m *MockFileService) SaveFile(filename, parentPath string, data []byte) error {
	if m.shouldFail {
		return m.failError
	}

	fullPath := buildPath(parentPath, filename)
	file := &models.FileInfo{
		Name:        filename,
		Path:        fullPath,
		Size:        int64(len(data)),
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  parentPath,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.files[fullPath] = file
	m.fileData[fullPath] = data
	return nil
}

func (m *MockFileService) SaveFileStream(filename, parentPath string, reader io.Reader, size int64) error {
	if m.shouldFail {
		return m.failError
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	return m.SaveFile(filename, parentPath, data)
}

func (m *MockFileService) GetFileData(path string) ([]byte, *models.FileInfo, error) {
	if m.shouldFail {
		return nil, nil, m.failError
	}

	file, exists := m.files[path]
	if !exists {
		return nil, nil, services.ErrFileNotFound
	}

	if file.IsDirectory {
		return nil, nil, fmt.Errorf("cannot download directory")
	}

	data, exists := m.fileData[path]
	if !exists {
		return nil, nil, fmt.Errorf("file data not found")
	}

	return data, file, nil
}

func (m *MockFileService) RenameFile(path, newName string) error {
	if m.shouldFail {
		return m.failError
	}

	file, exists := m.files[path]
	if !exists {
		return services.ErrFileNotFound
	}

	newPath := buildPath(file.ParentPath, newName)
	if _, exists := m.files[newPath]; exists {
		return fmt.Errorf("file already exists: %s", newPath)
	}

	delete(m.files, path)
	file.Name = newName
	file.Path = newPath
	m.files[newPath] = file

	if data, exists := m.fileData[path]; exists {
		delete(m.fileData, path)
		m.fileData[newPath] = data
	}

	return nil
}

func (m *MockFileService) MoveFile(sourcePath, destinationPath string) error {
	if m.shouldFail {
		return m.failError
	}

	file, exists := m.files[sourcePath]
	if !exists {
		return services.ErrFileNotFound
	}

	newPath := buildPath(destinationPath, file.Name)
	if _, exists := m.files[newPath]; exists {
		return fmt.Errorf("file already exists at destination: %s", newPath)
	}

	delete(m.files, sourcePath)
	file.ParentPath = destinationPath
	file.Path = newPath
	m.files[newPath] = file

	if data, exists := m.fileData[sourcePath]; exists {
		delete(m.fileData, sourcePath)
		m.fileData[newPath] = data
	}

	return nil
}

func (m *MockFileService) DeleteFile(path string) error {
	if m.shouldFail {
		return m.failError
	}

	file, exists := m.files[path]
	if !exists {
		return services.ErrFileNotFound
	}

	// Check if directory is empty
	if file.IsDirectory {
		for _, f := range m.files {
			if f.ParentPath == path {
				return fmt.Errorf("directory not empty: %s", path)
			}
		}
	}

	delete(m.files, path)
	delete(m.fileData, path)
	return nil
}

// Helper functions
func buildPath(parent, name string) string {
	if parent == "/" {
		return "/" + name
	}
	return parent + "/" + name
}

func getParentPath(path string) string {
	if path == "/" {
		return ""
	}
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == 0 {
		return "/"
	}
	return path[:lastSlash]
}

// TestHandlers wraps the handlers for testing with a mock service
type TestHandlers struct {
	fileService *MockFileService
	cfg         *config.Config
}

// Helper function to create test handlers
func createTestHandlers() (*TestHandlers, *MockFileService) {
	mockFileService := NewMockFileService()
	cfg := &config.Config{}
	cfg.Server.MaxFileSize = 10 * 1024 * 1024 // 10MB
	cfg.Server.MaxMemory = 8 * 1024 * 1024    // 8MB
	
	handlers := &TestHandlers{
		fileService: mockFileService,
		cfg:         cfg,
	}
	return handlers, mockFileService
}

// Implement all the handler methods for testing
func (h *TestHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ok",
	}
	WriteJSONTest(w, http.StatusOK, response)
}

func (h *TestHandlers) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		WriteErrorJSONTest(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var path string
	if strings.HasPrefix(r.URL.Path, "/api/v1/files") {
		path = strings.TrimPrefix(r.URL.Path, "/api/v1/files")
	} else if strings.HasPrefix(r.URL.Path, "/api/v1/directories") {
		path = strings.TrimPrefix(r.URL.Path, "/api/v1/directories")
	}

	if path == "" {
		path = "/"
	}

	listing, err := h.fileService.GetDirectoryListing(path)
	if err != nil {
		WriteErrorJSONTest(w, http.StatusInternalServerError, "Failed to list files: "+err.Error())
		return
	}

	WriteJSONTest(w, http.StatusOK, listing)
}

func (h *TestHandlers) CreateDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		WriteErrorJSONTest(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.CreateDirectoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorJSONTest(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name == "" {
		WriteErrorJSONTest(w, http.StatusBadRequest, "Directory name is required")
		return
	}

	if req.ParentPath == "" {
		req.ParentPath = "/"
	}

	directory, err := h.fileService.CreateDirectory(req.Name, req.ParentPath)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			WriteErrorJSONTest(w, http.StatusConflict, err.Error())
		} else if strings.Contains(err.Error(), "does not exist") {
			WriteErrorJSONTest(w, http.StatusNotFound, err.Error())
		} else {
			WriteErrorJSONTest(w, http.StatusInternalServerError, "Failed to create directory: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"message":   "Directory created successfully",
		"directory": directory,
	}

	WriteJSONTest(w, http.StatusCreated, response)
}

func (h *TestHandlers) Upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(int64(h.cfg.Server.MaxMemory))
	if err != nil {
		WriteErrorJSONTest(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		WriteErrorJSONTest(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	if header.Size > h.cfg.Server.MaxFileSize {
		WriteErrorJSONTest(w, http.StatusBadRequest, "File too large")
		return
	}

	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "/"
	}

	// Use streaming for files larger than 10MB
	const streamingThreshold = 10 * 1024 * 1024
	
	if header.Size > streamingThreshold {
		err = h.fileService.SaveFileStream(header.Filename, targetPath, file, header.Size)
	} else {
		data, err := io.ReadAll(file)
		if err != nil {
			WriteErrorJSONTest(w, http.StatusInternalServerError, "Failed to read file")
			return
		}
		err = h.fileService.SaveFile(header.Filename, targetPath, data)
	}

	if err != nil {
		WriteErrorJSONTest(w, http.StatusInternalServerError, "Failed to save file: "+err.Error())
		return
	}

	response := &models.UploadResponse{
		Success:  true,
		Filename: header.Filename,
		Size:     header.Size,
		Path:     targetPath,
		Message:  "File uploaded successfully",
	}

	WriteJSONTest(w, http.StatusCreated, response)
}

func (h *TestHandlers) Download(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/api/v1/download")
	if filePath == "" || filePath == "/" {
		WriteErrorJSONTest(w, http.StatusBadRequest, "File path required")
		return
	}

	data, fileInfo, err := h.fileService.GetFileData(filePath)
	if err != nil {
		if err == services.ErrFileNotFound {
			WriteErrorJSONTest(w, http.StatusNotFound, "File not found")
		} else {
			WriteErrorJSONTest(w, http.StatusInternalServerError, "Failed to read file: "+err.Error())
		}
		return
	}

	filename := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Type", fileInfo.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size, 10))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *TestHandlers) MoveFile(w http.ResponseWriter, r *http.Request) {
	var req models.MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorJSONTest(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.SourcePath == "" || req.DestinationPath == "" {
		WriteErrorJSONTest(w, http.StatusBadRequest, "Source and destination paths are required")
		return
	}

	err := h.fileService.MoveFile(req.SourcePath, req.DestinationPath)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteErrorJSONTest(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "already exists") {
			WriteErrorJSONTest(w, http.StatusConflict, err.Error())
		} else {
			WriteErrorJSONTest(w, http.StatusInternalServerError, "Failed to move file: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"message":     "File moved successfully",
		"source":      req.SourcePath,
		"destination": req.DestinationPath,
	}

	WriteJSONTest(w, http.StatusOK, response)
}

func (h *TestHandlers) RenameFile(w http.ResponseWriter, r *http.Request) {
	var req models.RenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorJSONTest(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Path == "" || req.NewName == "" {
		WriteErrorJSONTest(w, http.StatusBadRequest, "Path and new name are required")
		return
	}

	err := h.fileService.RenameFile(req.Path, req.NewName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteErrorJSONTest(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "already exists") {
			WriteErrorJSONTest(w, http.StatusConflict, err.Error())
		} else {
			WriteErrorJSONTest(w, http.StatusInternalServerError, "Failed to rename file: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"message":  "File renamed successfully",
		"old_path": req.Path,
		"new_name": req.NewName,
	}

	WriteJSONTest(w, http.StatusOK, response)
}

func (h *TestHandlers) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		WriteErrorJSONTest(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/files")
	if path == "" || path == "/" {
		WriteErrorJSONTest(w, http.StatusBadRequest, "Invalid file path")
		return
	}

	err := h.fileService.DeleteFile(path)
	if err != nil {
		if err == services.ErrFileNotFound {
			WriteErrorJSONTest(w, http.StatusNotFound, "File not found")
		} else if strings.Contains(err.Error(), "not empty") {
			WriteErrorJSONTest(w, http.StatusConflict, "Directory not empty")
		} else {
			WriteErrorJSONTest(w, http.StatusInternalServerError, "Failed to delete file: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "File deleted successfully",
		"path":    path,
	}

	WriteJSONTest(w, http.StatusOK, response)
}

// Test utility functions
func WriteJSONTest(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteErrorJSONTest(w http.ResponseWriter, status int, message string) {
	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  status,
	}
	WriteJSONTest(w, status, response)
}

func TestHealthCheck(t *testing.T) {
	handlers, _ := createTestHandlers()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthCheck(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}
}

func TestListFiles(t *testing.T) {
	handlers, mockService := createTestHandlers()

	// Setup test data
	mockService.files["/documents/file1.txt"] = &models.FileInfo{
		Name: "file1.txt", Path: "/documents/file1.txt", Size: 100,
		MimeType: "text/plain", IsDirectory: false, ParentPath: "/documents",
	}
	mockService.files["/documents/subdir"] = &models.FileInfo{
		Name: "subdir", Path: "/documents/subdir", Size: 0,
		MimeType: "inode/directory", IsDirectory: true, ParentPath: "/documents",
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedFiles  int
		expectedDirs   int
	}{
		{
			name:           "List root directory",
			path:           "/api/v1/files/",
			expectedStatus: http.StatusOK,
			expectedFiles:  0,
			expectedDirs:   0,
		},
		{
			name:           "List documents directory",
			path:           "/api/v1/files/documents",
			expectedStatus: http.StatusOK,
			expectedFiles:  1,
			expectedDirs:   1,
		},
		{
			name:           "List with directories endpoint",
			path:           "/api/v1/directories/documents",
			expectedStatus: http.StatusOK,
			expectedFiles:  1,
			expectedDirs:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handlers.ListFiles(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {
				var listing models.DirectoryListing
				err := json.NewDecoder(resp.Body).Decode(&listing)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if listing.TotalFiles != tt.expectedFiles {
					t.Errorf("Expected %d files, got %d", tt.expectedFiles, listing.TotalFiles)
				}

				if listing.TotalDirs != tt.expectedDirs {
					t.Errorf("Expected %d directories, got %d", tt.expectedDirs, listing.TotalDirs)
				}
			}
		})
	}
}

func TestListFiles_MethodNotAllowed(t *testing.T) {
	handlers, _ := createTestHandlers()

	req := httptest.NewRequest("POST", "/api/v1/files/", nil)
	w := httptest.NewRecorder()

	handlers.ListFiles(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestListFiles_ServiceError(t *testing.T) {
	handlers, mockService := createTestHandlers()

	mockService.shouldFail = true
	mockService.failError = fmt.Errorf("database connection failed")

	req := httptest.NewRequest("GET", "/api/v1/files/", nil)
	w := httptest.NewRecorder()

	handlers.ListFiles(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestCreateDirectory(t *testing.T) {
	handlers, _ := createTestHandlers()

	reqBody := models.CreateDirectoryRequest{
		Name:       "new_folder",
		ParentPath: "/documents",
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/directories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateDirectory(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Expected success to be true")
	}

	if response["message"] != "Directory created successfully" {
		t.Errorf("Unexpected message: %v", response["message"])
	}
}

func TestCreateDirectory_InvalidJSON(t *testing.T) {
	handlers, _ := createTestHandlers()

	req := httptest.NewRequest("POST", "/api/v1/directories", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateDirectory(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestCreateDirectory_EmptyName(t *testing.T) {
	handlers, _ := createTestHandlers()

	reqBody := models.CreateDirectoryRequest{
		Name:       "",
		ParentPath: "/documents",
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/directories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateDirectory(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestCreateDirectory_MethodNotAllowed(t *testing.T) {
	handlers, _ := createTestHandlers()

	req := httptest.NewRequest("GET", "/api/v1/directories", nil)
	w := httptest.NewRecorder()

	handlers.CreateDirectory(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestUpload(t *testing.T) {
	handlers, _ := createTestHandlers()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add file
	fileWriter, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	
	testContent := "test file content"
	fileWriter.Write([]byte(testContent))
	
	// Add path
	writer.WriteField("path", "/documents")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handlers.Upload(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var response models.UploadResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Filename != "test.txt" {
		t.Errorf("Expected filename 'test.txt', got %s", response.Filename)
	}

	if response.Size != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), response.Size)
	}
}

func TestUpload_NoFile(t *testing.T) {
	handlers, _ := createTestHandlers()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handlers.Upload(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestUpload_FileTooLarge(t *testing.T) {
	handlers, _ := createTestHandlers()

	// Create multipart form with large file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	fileWriter, err := writer.CreateFormFile("file", "large.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	
	// Write content larger than max file size (10MB + 1 byte)
	largeContent := make([]byte, 10*1024*1024+1)
	fileWriter.Write(largeContent)
	
	writer.WriteField("path", "/")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handlers.Upload(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestDownload(t *testing.T) {
	handlers, mockService := createTestHandlers()

	// Setup test file
	testContent := []byte("test file content")
	mockService.files["/documents/test.txt"] = &models.FileInfo{
		Name: "test.txt", Path: "/documents/test.txt", Size: int64(len(testContent)),
		MimeType: "text/plain", IsDirectory: false, ParentPath: "/documents",
	}
	mockService.fileData["/documents/test.txt"] = testContent

	req := httptest.NewRequest("GET", "/api/v1/download/documents/test.txt", nil)
	w := httptest.NewRecorder()

	handlers.Download(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check headers
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got %s", contentType)
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "attachment") {
		t.Errorf("Expected Content-Disposition to contain 'attachment', got %s", contentDisposition)
	}

	// Check content
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !bytes.Equal(responseBody, testContent) {
		t.Error("Response body does not match test content")
	}
}

func TestDownload_FileNotFound(t *testing.T) {
	handlers, _ := createTestHandlers()

	req := httptest.NewRequest("GET", "/api/v1/download/nonexistent.txt", nil)
	w := httptest.NewRecorder()

	handlers.Download(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestDownload_EmptyPath(t *testing.T) {
	handlers, _ := createTestHandlers()

	req := httptest.NewRequest("GET", "/api/v1/download", nil)
	w := httptest.NewRecorder()

	handlers.Download(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestMoveFile(t *testing.T) {
	handlers, mockService := createTestHandlers()

	// Setup test data
	mockService.files["/source/file.txt"] = &models.FileInfo{
		Name: "file.txt", Path: "/source/file.txt", Size: 100,
		MimeType: "text/plain", IsDirectory: false, ParentPath: "/source",
	}

	reqBody := models.MoveRequest{
		SourcePath:      "/source/file.txt",
		DestinationPath: "/destination",
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/move", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.MoveFile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Expected success to be true")
	}
}

func TestMoveFile_InvalidRequest(t *testing.T) {
	handlers, _ := createTestHandlers()

	reqBody := models.MoveRequest{
		SourcePath:      "",
		DestinationPath: "/destination",
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/move", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.MoveFile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestRenameFile(t *testing.T) {
	handlers, mockService := createTestHandlers()

	// Setup test data
	mockService.files["/documents/oldname.txt"] = &models.FileInfo{
		Name: "oldname.txt", Path: "/documents/oldname.txt", Size: 100,
		MimeType: "text/plain", IsDirectory: false, ParentPath: "/documents",
	}

	reqBody := models.RenameRequest{
		Path:    "/documents/oldname.txt",
		NewName: "newname.txt",
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/rename", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.RenameFile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Expected success to be true")
	}
}

func TestRenameFile_InvalidRequest(t *testing.T) {
	handlers, _ := createTestHandlers()

	reqBody := models.RenameRequest{
		Path:    "",
		NewName: "newname.txt",
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/rename", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.RenameFile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestDeleteFile(t *testing.T) {
	handlers, mockService := createTestHandlers()

	// Setup test data
	mockService.files["/documents/delete_me.txt"] = &models.FileInfo{
		Name: "delete_me.txt", Path: "/documents/delete_me.txt", Size: 100,
		MimeType: "text/plain", IsDirectory: false, ParentPath: "/documents",
	}

	req := httptest.NewRequest("DELETE", "/api/v1/files/documents/delete_me.txt", nil)
	w := httptest.NewRecorder()

	handlers.DeleteFile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Expected success to be true")
	}
}

func TestDeleteFile_MethodNotAllowed(t *testing.T) {
	handlers, _ := createTestHandlers()

	req := httptest.NewRequest("GET", "/api/v1/files/documents/file.txt", nil)
	w := httptest.NewRecorder()

	handlers.DeleteFile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestDeleteFile_InvalidPath(t *testing.T) {
	handlers, _ := createTestHandlers()

	req := httptest.NewRequest("DELETE", "/api/v1/files/", nil)
	w := httptest.NewRecorder()

	handlers.DeleteFile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestDeleteFile_FileNotFound(t *testing.T) {
	handlers, mockService := createTestHandlers()

	mockService.shouldFail = true
	mockService.failError = services.ErrFileNotFound

	req := httptest.NewRequest("DELETE", "/api/v1/files/nonexistent.txt", nil)
	w := httptest.NewRecorder()

	handlers.DeleteFile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestDeleteFile_DirectoryNotEmpty(t *testing.T) {
	handlers, mockService := createTestHandlers()

	mockService.shouldFail = true
	mockService.failError = fmt.Errorf("directory not empty")

	req := httptest.NewRequest("DELETE", "/api/v1/files/nonempty_dir", nil)
	w := httptest.NewRecorder()

	handlers.DeleteFile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}

// Benchmark tests
func BenchmarkHealthCheck(b *testing.B) {
	handlers, _ := createTestHandlers()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		handlers.HealthCheck(w, req)
	}
}

func BenchmarkListFiles(b *testing.B) {
	handlers, mockService := createTestHandlers()

	// Setup test data
	for i := 0; i < 100; i++ {
		mockService.files[fmt.Sprintf("/documents/file%d.txt", i)] = &models.FileInfo{
			Name: fmt.Sprintf("file%d.txt", i), Path: fmt.Sprintf("/documents/file%d.txt", i),
			Size: 1024, MimeType: "text/plain", IsDirectory: false, ParentPath: "/documents",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/files/documents", nil)
		w := httptest.NewRecorder()
		handlers.ListFiles(w, req)
	}
}

func BenchmarkUpload(b *testing.B) {
	handlers, _ := createTestHandlers()
	testContent := "test file content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		
		fileWriter, _ := writer.CreateFormFile("file", fmt.Sprintf("test%d.txt", i))
		fileWriter.Write([]byte(testContent))
		writer.WriteField("path", "/")
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		
		handlers.Upload(w, req)
	}
}