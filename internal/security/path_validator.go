package security

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrPathTraversal     = errors.New("path traversal attempt detected")
	ErrInvalidPath       = errors.New("invalid path")
	ErrAbsolutePath      = errors.New("absolute paths not allowed")
	ErrEmptyPath         = errors.New("path cannot be empty")
	ErrPathTooLong       = errors.New("path exceeds maximum length")
	ErrInvalidCharacters = errors.New("path contains invalid characters")
)

const (
	MaxPathLength = 4096
)

// PathValidator provides secure path validation to prevent directory traversal attacks
type PathValidator struct {
	basePath string
}

// NewPathValidator creates a new path validator with the given base path
func NewPathValidator(basePath string) *PathValidator {
	return &PathValidator{
		basePath: filepath.Clean(basePath),
	}
}

// ValidateAndNormalizePath validates and normalizes a path to prevent directory traversal
// This is the main function that should be used for all path validation
func (pv *PathValidator) ValidateAndNormalizePath(inputPath string) (string, error) {
	// Step 1: Basic validation
	if err := pv.validateBasicPath(inputPath); err != nil {
		return "", err
	}

	// Step 2: Check for directory traversal BEFORE normalization
	if err := pv.checkPathTraversal(inputPath); err != nil {
		return "", err
	}

	// Step 3: Normalize the path
	normalized := pv.normalizePath(inputPath)

	// Step 4: Check for directory traversal AFTER normalization (in case normalization revealed issues)
	if err := pv.checkPathTraversal(normalized); err != nil {
		return "", err
	}

	// Step 5: Validate final path is within base directory
	if err := pv.validateWithinBase(normalized); err != nil {
		return "", err
	}

	return normalized, nil
}

// ValidateAndGetFullPath validates a relative path and returns the full system path
func (pv *PathValidator) ValidateAndGetFullPath(relativePath string) (string, error) {
	// Validate and normalize the relative path
	normalizedPath, err := pv.ValidateAndNormalizePath(relativePath)
	if err != nil {
		return "", err
	}

	// Convert to full path
	fullPath := pv.buildFullPath(normalizedPath)

	// Final security check: ensure the resolved path is still within base
	if err := pv.validateFinalPath(fullPath); err != nil {
		return "", err
	}

	return fullPath, nil
}

// validateBasicPath performs basic path validation
func (pv *PathValidator) validateBasicPath(path string) error {
	if path == "" {
		return ErrEmptyPath
	}

	if len(path) > MaxPathLength {
		return ErrPathTooLong
	}

	// Check for null bytes (can be used to bypass validation)
	if strings.Contains(path, "\x00") {
		return ErrInvalidCharacters
	}

	// Check for other dangerous characters
	dangerousChars := []string{"\r", "\n", "\t"}
	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return ErrInvalidCharacters
		}
	}

	return nil
}

// normalizePath normalizes the path safely
func (pv *PathValidator) normalizePath(path string) string {
	// Handle empty or root path
	if path == "" || path == "/" {
		return "/"
	}

	// Convert to forward slashes first for consistency
	path = strings.ReplaceAll(path, "\\", "/")

	// Clean the path (removes redundant separators, etc.)
	cleaned := filepath.Clean(path)

	// Ensure it starts with /
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}

	// Convert back to forward slashes (in case Clean added backslashes on Windows)
	cleaned = strings.ReplaceAll(cleaned, "\\", "/")

	// Remove any double slashes that might have been created
	for strings.Contains(cleaned, "//") {
		cleaned = strings.ReplaceAll(cleaned, "//", "/")
	}

	return cleaned
}

// checkPathTraversal checks for directory traversal attempts
func (pv *PathValidator) checkPathTraversal(path string) error {
	// Check for explicit path traversal sequences
	if strings.Contains(path, "..") {
		return ErrPathTraversal
	}

	// Check for encoded path traversal attempts
	encodedTraversals := []string{
		"%2e%2e",     // URL encoded ..
		"%252e%252e", // Double URL encoded ..
		"..%2f",      // Mixed encoding
		"%2e%2e%2f",  // URL encoded ../
		"..%5c",      // Backslash variant
		"%2e%2e%5c",  // URL encoded ..\
	}

	lowerPath := strings.ToLower(path)
	for _, encoded := range encodedTraversals {
		if strings.Contains(lowerPath, encoded) {
			return ErrPathTraversal
		}
	}

	// Check each path component individually
	components := strings.Split(path, "/")
	for _, component := range components {
		if component == ".." {
			return ErrPathTraversal
		}
		
		// Note: We allow "." (current directory) as it's safe after normalization
		
		// Check for Unicode normalization attacks
		if strings.Contains(component, "\u002e\u002e") { // Unicode dots
			return ErrPathTraversal
		}
	}

	return nil
}

// validateWithinBase ensures the normalized path is within the allowed base
func (pv *PathValidator) validateWithinBase(normalizedPath string) error {
	// For relative paths starting with /, remove the leading /
	if strings.HasPrefix(normalizedPath, "/") && normalizedPath != "/" {
		normalizedPath = normalizedPath[1:]
	}

	// Check if any component tries to go up
	components := strings.Split(normalizedPath, "/")
	level := 0
	
	for _, component := range components {
		if component == "" || component == "." {
			continue
		}
		if component == ".." {
			level--
			if level < 0 {
				return ErrPathTraversal
			}
		} else {
			level++
		}
	}

	return nil
}

// buildFullPath safely constructs the full system path
func (pv *PathValidator) buildFullPath(normalizedPath string) string {
	// Remove leading slash for joining
	if strings.HasPrefix(normalizedPath, "/") && normalizedPath != "/" {
		normalizedPath = normalizedPath[1:]
	}

	// Special case for root
	if normalizedPath == "/" {
		normalizedPath = ""
	}

	return filepath.Join(pv.basePath, normalizedPath)
}

// validateFinalPath performs final validation on the resolved full path
func (pv *PathValidator) validateFinalPath(fullPath string) error {
	// Clean the full path
	cleanedFull := filepath.Clean(fullPath)
	cleanedBase := filepath.Clean(pv.basePath)

	// Ensure the path is within the base directory
	relPath, err := filepath.Rel(cleanedBase, cleanedFull)
	if err != nil {
		return ErrInvalidPath
	}

	// Check if the relative path tries to escape
	if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, string(filepath.Separator)+"..") {
		return ErrPathTraversal
	}

	return nil
}

// IsValidFilename validates a filename (not a path)
func IsValidFilename(filename string) error {
	if filename == "" {
		return ErrEmptyPath
	}

	if len(filename) > 255 {
		return ErrPathTooLong
	}

	// Check for dangerous characters in filename
	dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return ErrInvalidCharacters
		}
	}

	// Check for reserved names (Windows)
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	upperFilename := strings.ToUpper(filename)
	for _, reserved := range reservedNames {
		if upperFilename == reserved || strings.HasPrefix(upperFilename, reserved+".") {
			return ErrInvalidCharacters
		}
	}

	// Check for names that are just dots
	if filename == "." || filename == ".." {
		return ErrInvalidCharacters
	}

	return nil
}

// SanitizePath removes or replaces dangerous characters from a path
func SanitizePath(path string) string {
	// Replace dangerous characters
	sanitized := strings.ReplaceAll(path, "..", "_dot_dot_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")
	
	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	
	// Remove other control characters
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\t", "")
	
	return sanitized
}