package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Refex for valid file names (avoid directory traversal)
	validFilenameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// Extensions considered dangerous
	dangerousExtensions = map[string]bool{
		".exe": true, ".bat": true, ".cmd": true, ".sh": true,
		".ps1": true, ".vbs": true, ".js": true, ".jar": true,
	}
)

func IsValidFilename(filename string) bool {
	if filename == "" || filename == "." || filename == ".." {
		return false
	}

	// Verify characters dangerous
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return false
	}

	// Veryfy extension is not dangerous
	ext := strings.ToLower(filepath.Ext(filename))
	if dangerousExtensions[ext] {
		return false
	}

	// Verify length
	if len(filename) > 255 {
		return false
	}

	return true
}
