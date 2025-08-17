package security

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathValidator_ValidateAndNormalizePath(t *testing.T) {
	tempDir := t.TempDir()
	validator := NewPathValidator(tempDir)

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errorType   error
	}{
		// Valid paths
		{
			name:        "Root path",
			input:       "/",
			expected:    "/",
			expectError: false,
		},
		{
			name:        "Simple file path",
			input:       "/file.txt",
			expected:    "/file.txt",
			expectError: false,
		},
		{
			name:        "Nested directory",
			input:       "/dir/subdir/file.txt",
			expected:    "/dir/subdir/file.txt",
			expectError: false,
		},
		{
			name:        "Path without leading slash",
			input:       "file.txt",
			expected:    "/file.txt",
			expectError: false,
		},
		{
			name:        "Path with redundant slashes",
			input:       "//dir///file.txt",
			expected:    "/dir/file.txt",
			expectError: false,
		},
		
		// Path traversal attempts - should all fail
		{
			name:        "Simple dot-dot traversal",
			input:       "../file.txt",
			expected:    "",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "Deep path traversal",
			input:       "../../etc/passwd",
			expected:    "",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "Path traversal in middle",
			input:       "/dir/../../../etc/passwd",
			expected:    "",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "URL encoded path traversal",
			input:       "/dir/%2e%2e/file.txt",
			expected:    "",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "Double URL encoded",
			input:       "/dir/%252e%252e/file.txt",
			expected:    "",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "Mixed encoding",
			input:       "/dir/..%2f../file.txt",
			expected:    "",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "Backslash traversal",
			input:       "/dir/..\\../file.txt",
			expected:    "",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "Unicode dot traversal",
			input:       "/dir/\u002e\u002e/file.txt",
			expected:    "",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		
		// Invalid characters
		{
			name:        "Null byte injection",
			input:       "/file.txt\x00.php",
			expected:    "",
			expectError: true,
			errorType:   ErrInvalidCharacters,
		},
		{
			name:        "Newline character",
			input:       "/file\n.txt",
			expected:    "",
			expectError: true,
			errorType:   ErrInvalidCharacters,
		},
		{
			name:        "Carriage return",
			input:       "/file\r.txt",
			expected:    "",
			expectError: true,
			errorType:   ErrInvalidCharacters,
		},
		
		// Edge cases
		{
			name:        "Empty path",
			input:       "",
			expected:    "",
			expectError: true,
			errorType:   ErrEmptyPath,
		},
		{
			name:        "Very long path",
			input:       "/" + string(make([]byte, MaxPathLength)),
			expected:    "",
			expectError: true,
			errorType:   ErrPathTooLong,
		},
		{
			name:        "Current directory reference",
			input:       "/./file.txt",
			expected:    "/file.txt",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateAndNormalizePath(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("Expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPathValidator_ValidateAndGetFullPath(t *testing.T) {
	tempDir := t.TempDir()
	validator := NewPathValidator(tempDir)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "Valid file path",
			input:       "/documents/file.txt",
			expectError: false,
		},
		{
			name:        "Path traversal attempt",
			input:       "../../../etc/passwd",
			expectError: true,
		},
		{
			name:        "Complex traversal",
			input:       "/docs/../../etc/passwd",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath, err := validator.ValidateAndGetFullPath(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify the full path is within the base directory
			if !filepath.HasPrefix(fullPath, tempDir) {
				t.Errorf("Full path %q is not within base directory %q", fullPath, tempDir)
			}

			// Verify no directory traversal in the resolved path
			relPath, err := filepath.Rel(tempDir, fullPath)
			if err != nil {
				t.Errorf("Failed to get relative path: %v", err)
				return
			}

			if filepath.IsAbs(relPath) || filepath.HasPrefix(relPath, "..") {
				t.Errorf("Resolved path %q attempts to escape base directory", relPath)
			}
		})
	}
}

func TestIsValidFilename(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		// Valid filenames
		{
			name:        "Simple filename",
			filename:    "file.txt",
			expectError: false,
		},
		{
			name:        "Filename with numbers",
			filename:    "file123.txt",
			expectError: false,
		},
		{
			name:        "Filename with underscores and hyphens",
			filename:    "my_file-v2.txt",
			expectError: false,
		},
		{
			name:        "Filename with dots",
			filename:    "file.backup.txt",
			expectError: false,
		},

		// Invalid filenames
		{
			name:        "Empty filename",
			filename:    "",
			expectError: true,
		},
		{
			name:        "Filename with forward slash",
			filename:    "dir/file.txt",
			expectError: true,
		},
		{
			name:        "Filename with backslash",
			filename:    "dir\\file.txt",
			expectError: true,
		},
		{
			name:        "Filename with colon",
			filename:    "file:stream.txt",
			expectError: true,
		},
		{
			name:        "Filename with asterisk",
			filename:    "file*.txt",
			expectError: true,
		},
		{
			name:        "Filename with question mark",
			filename:    "file?.txt",
			expectError: true,
		},
		{
			name:        "Filename with quotes",
			filename:    "file\".txt",
			expectError: true,
		},
		{
			name:        "Filename with angle brackets",
			filename:    "file<>.txt",
			expectError: true,
		},
		{
			name:        "Filename with pipe",
			filename:    "file|pipe.txt",
			expectError: true,
		},
		{
			name:        "Filename with null byte",
			filename:    "file\x00.txt",
			expectError: true,
		},
		{
			name:        "Dot filename",
			filename:    ".",
			expectError: true,
		},
		{
			name:        "Double dot filename",
			filename:    "..",
			expectError: true,
		},
		{
			name:        "Reserved name CON",
			filename:    "CON",
			expectError: true,
		},
		{
			name:        "Reserved name con.txt",
			filename:    "con.txt",
			expectError: true,
		},
		{
			name:        "Reserved name PRN",
			filename:    "PRN",
			expectError: true,
		},
		{
			name:        "Reserved name COM1",
			filename:    "COM1",
			expectError: true,
		},
		{
			name:        "Reserved name LPT1",
			filename:    "LPT1",
			expectError: true,
		},
		{
			name:        "Too long filename",
			filename:    string(make([]byte, 256)),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsValidFilename(tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Path with traversal",
			input:    "../file.txt",
			expected: "_dot_dot_/file.txt",
		},
		{
			name:     "Path with backslashes",
			input:    "dir\\file.txt",
			expected: "dir/file.txt",
		},
		{
			name:     "Path with null bytes",
			input:    "file\x00.txt",
			expected: "file.txt",
		},
		{
			name:     "Path with control characters",
			input:    "file\r\n\t.txt",
			expected: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePath(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Benchmark tests
func BenchmarkPathValidator_ValidateAndNormalizePath(b *testing.B) {
	tempDir := b.TempDir()
	validator := NewPathValidator(tempDir)
	testPath := "/documents/subdirectory/file.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateAndNormalizePath(testPath)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkIsValidFilename(b *testing.B) {
	filename := "test_file_name.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := IsValidFilename(filename)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// Test real-world attack vectors
func TestPathTraversalAttackVectors(t *testing.T) {
	tempDir := t.TempDir()
	validator := NewPathValidator(tempDir)

	// Common attack vectors from OWASP and security research
	attackVectors := []string{
		// Basic traversal
		"../",
		"../../",
		"../../../",
		"../../../etc/passwd",
		"../../../windows/system32/config/sam",
		
		// URL encoded
		"%2e%2e/",
		"%2e%2e%2f",
		"..%2f",
		"%2e%2e%2f%2e%2e%2f%2e%2e%2f",
		
		// Double URL encoded
		"%252e%252e/",
		"%252e%252e%252f",
		
		// UTF-8 encoded
		"..%c0%af",
		"..%c1%9c",
		
		// Unicode variations
		"\u002e\u002e/",
		"\u002e\u002e\u002f",
		
		// Mixed case (should be handled by normalization)
		"..%2F",
		"..%2f",
		
		// Null byte injection
		"../\x00",
		"file.txt\x00.php",
		
		// Backslash variations
		"..\\",
		"..\\..\\",
		"..\\..\\..\\windows\\system32\\",
		
		// Current directory bypass attempts
		"./../../",
		"dir/./../../",
		
		// Long path attempts
		strings.Repeat("../", 100) + "etc/passwd",
		
		// Combination attacks
		"/dir/../../../etc/passwd",
		"/./dir/../../../etc/passwd",
		"dir/subdir/../../../../../../etc/passwd",
	}

	for i, vector := range attackVectors {
		t.Run(fmt.Sprintf("AttackVector_%d", i), func(t *testing.T) {
			// All attack vectors should be rejected
			_, err := validator.ValidateAndNormalizePath(vector)
			if err == nil {
				t.Errorf("Attack vector %q was not blocked", vector)
			}

			// Also test with ValidateAndGetFullPath
			_, err = validator.ValidateAndGetFullPath(vector)
			if err == nil {
				t.Errorf("Attack vector %q was not blocked by ValidateAndGetFullPath", vector)
			}
		})
	}
}

// Test that legitimate paths still work after security hardening
func TestLegitimatePathsStillWork(t *testing.T) {
	tempDir := t.TempDir()
	validator := NewPathValidator(tempDir)

	legitimatePaths := []string{
		"/",
		"/file.txt",
		"/documents/file.txt",
		"/documents/subdirectory/file.txt",
		"/images/photo.jpg",
		"/backup/2023/data.zip",
		"/user-uploads/document.pdf",
		"/very/deep/directory/structure/file.txt",
	}

	for _, path := range legitimatePaths {
		t.Run("Legitimate_"+strings.ReplaceAll(path, "/", "_"), func(t *testing.T) {
			// Should not error
			normalized, err := validator.ValidateAndNormalizePath(path)
			if err != nil {
				t.Errorf("Legitimate path %q was blocked: %v", path, err)
				return
			}

			// Should return a clean path
			if normalized == "" {
				t.Errorf("Legitimate path %q returned empty result", path)
				return
			}

			// Should be able to get full path
			fullPath, err := validator.ValidateAndGetFullPath(path)
			if err != nil {
				t.Errorf("Legitimate path %q failed ValidateAndGetFullPath: %v", path, err)
				return
			}

			// Full path should be within base directory
			if !strings.HasPrefix(fullPath, tempDir) {
				t.Errorf("Full path %q is not within base directory %q", fullPath, tempDir)
			}
		})
	}
}