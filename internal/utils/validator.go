package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Enhanced regex for valid file names - more permissive but still secure
	validFilenameRegex = regexp.MustCompile(`^[a-zA-Z0-9._\-\s\(\)\[\]]+$`)

	// Extensions considered dangerous
	dangerousExtensions = map[string]bool{
		".exe": true, ".bat": true, ".cmd": true, ".sh": true,
		".ps1": true, ".vbs": true, ".scr": true, ".com": true,
		".pif": true, ".application": true, ".gadget": true, ".msi": true,
		".msp": true, ".msc": true, ".jar": true,
	}

	// Reserved names that should not be used as filenames
	reservedNames = map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true,
		"COM5": true, "COM6": true, "COM7": true, "COM8": true,
		"COM9": true, "LPT1": true, "LPT2": true, "LPT3": true,
		"LPT4": true, "LPT5": true, "LPT6": true, "LPT7": true,
		"LPT8": true, "LPT9": true,
	}
)

func IsValidFilename(filename string) bool {
	if filename == "" || filename == "." || filename == ".." {
		return false
	}

	// Check length limit
	if len(filename) > 255 {
		return false
	}

	// Check for path separators and other dangerous characters
	dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return false
		}
	}

	// Check for reserved names (case-insensitive)
	nameWithoutExt := strings.ToUpper(strings.TrimSuffix(filename, filepath.Ext(filename)))
	if reservedNames[nameWithoutExt] {
		return false
	}

	// Check if filename matches our allowed pattern
	if !validFilenameRegex.MatchString(filename) {
		return false
	}

	// Check extension against dangerous list
	ext := strings.ToLower(filepath.Ext(filename))
	if dangerousExtensions[ext] {
		return false
	}

	// Additional checks for specific patterns
	if strings.HasPrefix(filename, ".") && len(filename) > 1 {
		// Allow dotfiles but check they're reasonable
		if !validFilenameRegex.MatchString(filename[1:]) {
			return false
		}
	}

	// Prevent leading/trailing spaces or dots (can cause issues on Windows)
	trimmed := strings.TrimSpace(filename)
	if trimmed != filename || strings.HasSuffix(filename, ".") {
		return false
	}

	return true
}
