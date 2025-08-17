package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/anddsdev/cloudlet/config"
	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/security"
	"github.com/anddsdev/cloudlet/internal/utils"
)

// MultipleUploadValidator handles validation for multiple file uploads
type MultipleUploadValidator struct {
	cfg           *config.Config
	pathValidator *security.PathValidator
}

// NewMultipleUploadValidator creates a new multiple upload validator
func NewMultipleUploadValidator(cfg *config.Config, storagePath string) *MultipleUploadValidator {
	return &MultipleUploadValidator{
		cfg:           cfg,
		pathValidator: security.NewPathValidator(storagePath),
	}
}

// ValidateMultipleUpload performs comprehensive validation for multiple file uploads
func (v *MultipleUploadValidator) ValidateMultipleUpload(files []*multipart.FileHeader, targetPath string) *models.UploadValidationResult {
	result := &models.UploadValidationResult{
		Valid:      true,
		TotalFiles: len(files),
		TotalSize:  0,
	}

	// Validate target path first
	if _, err := v.pathValidator.ValidateAndNormalizePath(targetPath); err != nil {
		result.Valid = false
		result.ValidationErrors = append(result.ValidationErrors, fmt.Sprintf("Invalid target path: %v", err))
		return result
	}

	// Check if number of files exceeds limit
	if len(files) > v.cfg.Server.Upload.MaxFilesPerRequest {
		result.Valid = false
		result.MaxFilesExceeded = true
		result.ValidationErrors = append(result.ValidationErrors, 
			fmt.Sprintf("Too many files: %d exceeds limit of %d", len(files), v.cfg.Server.Upload.MaxFilesPerRequest))
	}

	// Track file names to detect duplicates
	fileNames := make(map[string]bool)
	var totalSize int64

	for i, file := range files {
		// Calculate total size
		totalSize += file.Size
		result.TotalSize = totalSize

		// Validate individual file size
		if file.Size > v.cfg.Server.MaxFileSize {
			result.Valid = false
			result.OversizedFiles = append(result.OversizedFiles, file.Filename)
		}

		// Validate filename
		if !utils.IsValidFilename(file.Filename) {
			result.Valid = false
			result.InvalidFiles = append(result.InvalidFiles, file.Filename)
		}

		// Check for dangerous files
		if v.isDangerousFile(file.Filename) {
			result.Valid = false
			result.DangerousFiles = append(result.DangerousFiles, file.Filename)
		}

		// Check for duplicates
		normalizedName := strings.ToLower(file.Filename)
		if fileNames[normalizedName] {
			result.Valid = false
			result.DuplicateFiles = append(result.DuplicateFiles, file.Filename)
		}
		fileNames[normalizedName] = true

		// Validate MIME type
		if !v.isAllowedMimeType(file) {
			result.Valid = false
			result.ValidationErrors = append(result.ValidationErrors, 
				fmt.Sprintf("File %s: unsupported file type", file.Filename))
		}

		// Additional file-specific validations
		if err := v.validateFileContent(file, i); err != nil {
			result.Valid = false
			result.ValidationErrors = append(result.ValidationErrors, 
				fmt.Sprintf("File %s: %v", file.Filename, err))
		}
	}

	// Check total size limit
	if totalSize > v.cfg.Server.Upload.MaxTotalSizePerRequest {
		result.Valid = false
		result.MaxSizeExceeded = true
		result.ValidationErrors = append(result.ValidationErrors, 
			fmt.Sprintf("Total size %d exceeds limit of %d bytes", totalSize, v.cfg.Server.Upload.MaxTotalSizePerRequest))
	}

	return result
}

// ValidateUploadRate checks if upload rate limits are respected
func (v *MultipleUploadValidator) ValidateUploadRate(clientIP string, fileCount int) error {
	// In a real implementation, this would check against a rate limiter
	// For now, we'll do a simple check against the configured limit
	if fileCount > v.cfg.Server.Upload.RateLimitPerMinute {
		return fmt.Errorf("rate limit exceeded: %d files exceeds limit of %d per minute", 
			fileCount, v.cfg.Server.Upload.RateLimitPerMinute)
	}
	return nil
}

// isDangerousFile checks if a file has a dangerous extension or pattern
func (v *MultipleUploadValidator) isDangerousFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Dangerous extensions
	dangerousExtensions := map[string]bool{
		".exe": true, ".bat": true, ".cmd": true, ".com": true,
		".scr": true, ".pif": true, ".vbs": true, ".ps1": true,
		".jar": true, ".app": true, ".deb": true, ".rpm": true,
		".dmg": true, ".pkg": true, ".msi": true, ".msp": true,
		".gadget": true, ".application": true, ".lnk": true,
		".reg": true, ".inf": true, ".hta": true, ".cpl": true,
	}

	if dangerousExtensions[ext] {
		return true
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"autorun", "desktop.ini", "thumbs.db", ".htaccess", 
		"web.config", ".env", ".git", "id_rsa", "private",
	}

	lowerFilename := strings.ToLower(filename)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerFilename, pattern) {
			return true
		}
	}

	return false
}

// isAllowedMimeType validates the MIME type of the file
func (v *MultipleUploadValidator) isAllowedMimeType(file *multipart.FileHeader) bool {
	// For now, we'll use a simple extension-based check
	// In production, you might want to check the actual file content
	ext := strings.ToLower(filepath.Ext(file.Filename))
	
	// Common allowed extensions - this should be configurable
	allowedExtensions := map[string]bool{
		// Images
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, 
		".bmp": true, ".webp": true, ".svg": true, ".ico": true,
		".tiff": true, ".tif": true,
		
		// Documents
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, 
		".xlsx": true, ".ppt": true, ".pptx": true, ".odt": true,
		".ods": true, ".odp": true,
		
		// Text
		".txt": true, ".md": true, ".html": true, ".htm": true,
		".css": true, ".json": true, ".xml": true, ".csv": true,
		".yaml": true, ".yml": true, ".toml": true,
		
		// Archives
		".zip": true, ".rar": true, ".7z": true, ".tar": true,
		".gz": true, ".bz2": true, ".xz": true,
		
		// Audio
		".mp3": true, ".wav": true, ".flac": true, ".ogg": true,
		".m4a": true, ".aac": true,
		
		// Video
		".mp4": true, ".avi": true, ".mov": true, ".wmv": true,
		".flv": true, ".webm": true, ".mkv": true,
		
		// Code files
		".go": true, ".py": true, ".java": true, ".c": true,
		".cpp": true, ".h": true, ".hpp": true, ".php": true,
		".rb": true, ".js": true, ".ts": true, ".sql": true,
	}

	return allowedExtensions[ext]
}

// validateFileContent performs additional content-based validation
func (v *MultipleUploadValidator) validateFileContent(file *multipart.FileHeader, index int) error {
	// Open file to check content
	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("cannot open file for validation: %v", err)
	}
	defer f.Close()

	// Read first few bytes to detect actual file type
	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil && n == 0 {
		return fmt.Errorf("cannot read file content: %v", err)
	}

	// Check for executable file signatures
	if v.hasExecutableSignature(buffer[:n]) {
		return fmt.Errorf("executable file detected")
	}

	// Check for script content in text files
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == ".txt" || ext == ".md" || ext == ".html" || ext == ".htm" {
		if v.hasScriptContent(buffer[:n]) {
			return fmt.Errorf("potentially dangerous script content detected")
		}
	}

	return nil
}

// hasExecutableSignature checks for common executable file signatures
func (v *MultipleUploadValidator) hasExecutableSignature(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// PE executable (Windows)
	if len(data) >= 2 && data[0] == 0x4D && data[1] == 0x5A {
		return true
	}

	// ELF executable (Linux)
	if len(data) >= 4 && data[0] == 0x7F && data[1] == 0x45 && data[2] == 0x4C && data[3] == 0x46 {
		return true
	}

	// Mach-O executable (macOS)
	if len(data) >= 4 {
		// 32-bit
		if (data[0] == 0xFE && data[1] == 0xED && data[2] == 0xFA && data[3] == 0xCE) ||
		   (data[0] == 0xCE && data[1] == 0xFA && data[2] == 0xED && data[3] == 0xFE) {
			return true
		}
		// 64-bit
		if (data[0] == 0xFE && data[1] == 0xED && data[2] == 0xFA && data[3] == 0xCF) ||
		   (data[0] == 0xCF && data[1] == 0xFA && data[2] == 0xED && data[3] == 0xFE) {
			return true
		}
	}

	return false
}

// hasScriptContent checks for potentially dangerous script content
func (v *MultipleUploadValidator) hasScriptContent(data []byte) bool {
	content := strings.ToLower(string(data))
	
	dangerousPatterns := []string{
		"<script", "javascript:", "vbscript:", "data:text/html",
		"<?php", "<%", "#!/bin/", "#!/usr/bin/",
		"powershell", "cmd.exe", "bash", "sh ",
		"eval(", "exec(", "system(", "shell_exec(",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

// GenerateUploadSummary creates a summary of validation results
func (v *MultipleUploadValidator) GenerateUploadSummary(validation *models.UploadValidationResult) string {
	if validation.Valid {
		return fmt.Sprintf("All %d files passed validation (Total size: %d bytes)", 
			validation.TotalFiles, validation.TotalSize)
	}

	var issues []string
	if validation.MaxFilesExceeded {
		issues = append(issues, "too many files")
	}
	if validation.MaxSizeExceeded {
		issues = append(issues, "total size too large")
	}
	if len(validation.InvalidFiles) > 0 {
		issues = append(issues, fmt.Sprintf("%d invalid filenames", len(validation.InvalidFiles)))
	}
	if len(validation.DuplicateFiles) > 0 {
		issues = append(issues, fmt.Sprintf("%d duplicate files", len(validation.DuplicateFiles)))
	}
	if len(validation.DangerousFiles) > 0 {
		issues = append(issues, fmt.Sprintf("%d dangerous files", len(validation.DangerousFiles)))
	}

	return fmt.Sprintf("Validation failed: %s", strings.Join(issues, ", "))
}