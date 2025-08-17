package services

import (
	"mime/multipart"
	"testing"

	"github.com/anddsdev/cloudlet/config"
)

func TestMultipleUploadValidator_DangerousFileDetection(t *testing.T) {
	cfg := &config.Config{}
	cfg.Server.Upload.MaxFilesPerRequest = 10
	cfg.Server.Upload.MaxTotalSizePerRequest = 1024
	cfg.Server.MaxFileSize = 100

	validator := NewMultipleUploadValidator(cfg, "/tmp/test")

	dangerousFiles := []string{
		"suspicious.exe",
		"script.bat",
		"malware.scr",
		"trojan.com",
		"autorun.inf",
		"shell.ps1",
		"hack.vbs",
		"backdoor.jar",
	}

	for _, filename := range dangerousFiles {
		t.Run(filename, func(t *testing.T) {
			if !validator.isDangerousFile(filename) {
				t.Errorf("File %s should be detected as dangerous", filename)
			}
		})
	}

	safeFiles := []string{
		"document.pdf",
		"image.jpg",
		"text.txt",
		"data.json",
		"archive.zip",
		"audio.mp3",
		"video.mp4",
	}

	for _, filename := range safeFiles {
		t.Run(filename, func(t *testing.T) {
			if validator.isDangerousFile(filename) {
				t.Errorf("File %s should not be detected as dangerous", filename)
			}
		})
	}
}

func TestMultipleUploadValidator_ScriptContentDetection(t *testing.T) {
	cfg := &config.Config{}
	validator := NewMultipleUploadValidator(cfg, "/tmp/test")

	tests := []struct {
		name      string
		content   string
		dangerous bool
	}{
		{
			name:      "JavaScript",
			content:   "<script>alert('test')</script>",
			dangerous: true,
		},
		{
			name:      "PHP code",
			content:   "<?php echo 'hello'; ?>",
			dangerous: true,
		},
		{
			name:      "PowerShell",
			content:   "powershell -exec bypass",
			dangerous: true,
		},
		{
			name:      "Plain text",
			content:   "This is just normal text content",
			dangerous: false,
		},
		{
			name:      "Markdown",
			content:   "# Header\nSome normal markdown content",
			dangerous: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.hasScriptContent([]byte(tt.content))
			if result != tt.dangerous {
				t.Errorf("Expected dangerous=%v for %s, got %v", tt.dangerous, tt.name, result)
			}
		})
	}
}

func TestMultipleUploadValidator_FileTypeValidation(t *testing.T) {
	cfg := &config.Config{}
	validator := NewMultipleUploadValidator(cfg, "/tmp/test")

	tests := []struct {
		name     string
		filename string
		allowed  bool
	}{
		{
			name:     "Text file",
			filename: "document.txt",
			allowed:  true,
		},
		{
			name:     "Image file",
			filename: "photo.jpg",
			allowed:  true,
		},
		{
			name:     "PDF document",
			filename: "document.pdf",
			allowed:  true,
		},
		{
			name:     "Unknown extension",
			filename: "file.unknown",
			allowed:  false,
		},
		{
			name:     "No extension",
			filename: "README",
			allowed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a proper multipart.FileHeader
			mockFile := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     100,
			}
			
			result := validator.isAllowedMimeType(mockFile)
			if result != tt.allowed {
				t.Errorf("Expected allowed=%v for %s, got %v", tt.allowed, tt.filename, result)
			}
		})
	}
}

func TestMimeTypeDetection(t *testing.T) {
	cfg := &config.Config{}
	validator := NewMultipleUploadValidator(cfg, "/tmp/test")

	testCases := []struct {
		filename string
		expected bool
	}{
		{"document.pdf", true},
		{"image.png", true},
		{"audio.mp3", true},
		{"video.mp4", true},
		{"code.go", true},
		{"unknown.xyz", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// Create test file header with proper structure
			file := &multipart.FileHeader{
				Filename: tc.filename,
				Size:     1024,
			}
			
			result := validator.isAllowedMimeType(file)
			if result != tc.expected {
				t.Errorf("File %s: expected %v, got %v", tc.filename, tc.expected, result)
			}
		})
	}
}