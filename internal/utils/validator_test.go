package utils

import (
	"strings"
	"testing"
)

func TestIsValidFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		// Valid filenames
		{
			name:     "Simple filename",
			filename: "document.txt",
			expected: true,
		},
		{
			name:     "Filename with numbers",
			filename: "file123.txt",
			expected: true,
		},
		{
			name:     "Filename with underscores",
			filename: "my_file.txt",
			expected: true,
		},
		{
			name:     "Filename with hyphens",
			filename: "my-file.txt",
			expected: true,
		},
		{
			name:     "Filename with dots",
			filename: "file.backup.txt",
			expected: true,
		},
		{
			name:     "Filename with spaces",
			filename: "my file.txt",
			expected: true,
		},
		{
			name:     "Filename with parentheses",
			filename: "file(1).txt",
			expected: true,
		},
		{
			name:     "Filename with brackets",
			filename: "file[backup].txt",
			expected: true,
		},
		{
			name:     "Long filename",
			filename: strings.Repeat("a", 250) + ".txt",
			expected: true,
		},
		{
			name:     "Dotfile",
			filename: ".gitignore",
			expected: true,
		},
		{
			name:     "Hidden file with extension",
			filename: ".hidden.txt",
			expected: true,
		},

		// Invalid filenames - Empty and special cases
		{
			name:     "Empty filename",
			filename: "",
			expected: false,
		},
		{
			name:     "Current directory",
			filename: ".",
			expected: false,
		},
		{
			name:     "Parent directory",
			filename: "..",
			expected: false,
		},

		// Invalid filenames - Path separators
		{
			name:     "Forward slash",
			filename: "dir/file.txt",
			expected: false,
		},
		{
			name:     "Backslash",
			filename: "dir\\file.txt",
			expected: false,
		},

		// Invalid filenames - Dangerous characters
		{
			name:     "Colon",
			filename: "file:stream.txt",
			expected: false,
		},
		{
			name:     "Asterisk",
			filename: "file*.txt",
			expected: false,
		},
		{
			name:     "Question mark",
			filename: "file?.txt",
			expected: false,
		},
		{
			name:     "Double quote",
			filename: "file\".txt",
			expected: false,
		},
		{
			name:     "Less than",
			filename: "file<.txt",
			expected: false,
		},
		{
			name:     "Greater than",
			filename: "file>.txt",
			expected: false,
		},
		{
			name:     "Pipe",
			filename: "file|.txt",
			expected: false,
		},
		{
			name:     "Null byte",
			filename: "file\x00.txt",
			expected: false,
		},

		// Invalid filenames - Reserved names (Windows)
		{
			name:     "CON",
			filename: "CON",
			expected: false,
		},
		{
			name:     "CON with extension",
			filename: "CON.txt",
			expected: false,
		},
		{
			name:     "PRN",
			filename: "PRN",
			expected: false,
		},
		{
			name:     "AUX",
			filename: "AUX",
			expected: false,
		},
		{
			name:     "NUL",
			filename: "NUL",
			expected: false,
		},
		{
			name:     "COM1",
			filename: "COM1",
			expected: false,
		},
		{
			name:     "COM9",
			filename: "COM9",
			expected: false,
		},
		{
			name:     "LPT1",
			filename: "LPT1",
			expected: false,
		},
		{
			name:     "LPT9",
			filename: "LPT9",
			expected: false,
		},
		{
			name:     "Case insensitive CON",
			filename: "con",
			expected: false,
		},
		{
			name:     "Case insensitive con.txt",
			filename: "con.txt",
			expected: false,
		},

		// Invalid filenames - Dangerous extensions
		{
			name:     "Executable",
			filename: "program.exe",
			expected: false,
		},
		{
			name:     "Batch file",
			filename: "script.bat",
			expected: false,
		},
		{
			name:     "Command file",
			filename: "script.cmd",
			expected: false,
		},
		{
			name:     "Shell script",
			filename: "script.sh",
			expected: false,
		},
		{
			name:     "PowerShell script",
			filename: "script.ps1",
			expected: false,
		},
		{
			name:     "VBS script",
			filename: "script.vbs",
			expected: false,
		},
		{
			name:     "Screen saver",
			filename: "malware.scr",
			expected: false,
		},
		{
			name:     "COM file",
			filename: "program.com",
			expected: false,
		},
		{
			name:     "PIF file",
			filename: "program.pif",
			expected: false,
		},
		{
			name:     "Application file",
			filename: "app.application",
			expected: false,
		},
		{
			name:     "Gadget file",
			filename: "widget.gadget",
			expected: false,
		},
		{
			name:     "MSI installer",
			filename: "installer.msi",
			expected: false,
		},
		{
			name:     "MSP patch",
			filename: "patch.msp",
			expected: false,
		},
		{
			name:     "MSC console",
			filename: "console.msc",
			expected: false,
		},
		{
			name:     "JAR file",
			filename: "program.jar",
			expected: false,
		},

		// Invalid filenames - Length
		{
			name:     "Too long filename",
			filename: strings.Repeat("a", 256),
			expected: false,
		},

		// Invalid filenames - Leading/trailing issues
		{
			name:     "Leading space",
			filename: " file.txt",
			expected: false,
		},
		{
			name:     "Trailing space",
			filename: "file.txt ",
			expected: false,
		},
		{
			name:     "Trailing dot",
			filename: "file.txt.",
			expected: false,
		},

		// Invalid filenames - Invalid regex patterns
		{
			name:     "Control characters",
			filename: "file\r\n.txt",
			expected: false,
		},
		{
			name:     "Tab character",
			filename: "file\t.txt",
			expected: false,
		},
		{
			name:     "Unicode control",
			filename: "file\u0001.txt",
			expected: false,
		},

		// Edge cases for dotfiles
		{
			name:     "Invalid dotfile with control chars",
			filename: ".\r\n",
			expected: false,
		},
		{
			name:     "Dotfile with invalid chars",
			filename: ".file*.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("IsValidFilename(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsValidFilename_ValidFiles(t *testing.T) {
	validFiles := []string{
		"document.pdf",
		"image.jpg",
		"image.png",
		"data.json",
		"style.css",
		"script.js",
		"README.md",
		"config.yaml",
		"data.csv",
		"presentation.pptx",
		"spreadsheet.xlsx",
		"archive.zip",
		"music.mp3",
		"video.mp4",
		"photo.jpeg",
		"icon.svg",
		"font.ttf",
		"My Document.docx",
		"Project (Final).zip",
		"backup_2024_01_15.sql",
		"user-profile.json",
		"temp.tmp",
		"cache_file_123.dat",
		"log_2024_01_15_10_30_45.log",
		"settings[production].ini",
		"data file (version 2).xlsx",
	}

	for _, filename := range validFiles {
		t.Run("Valid_"+filename, func(t *testing.T) {
			if !IsValidFilename(filename) {
				t.Errorf("Valid filename %q was rejected", filename)
			}
		})
	}
}

func TestIsValidFilename_InvalidFiles(t *testing.T) {
	invalidFiles := []string{
		"file/path.txt",
		"file\\path.txt",
		"file:name.txt",
		"file*name.txt",
		"file?name.txt",
		"file\"name.txt",
		"file<name.txt",
		"file>name.txt",
		"file|name.txt",
		"malware.exe",
		"virus.bat",
		"trojan.scr",
		"backdoor.com",
		"hack.vbs",
		"evil.ps1",
		"dangerous.msi",
		"CON",
		"PRN",
		"AUX",
		"NUL",
		"COM1",
		"LPT1",
		"con.txt",
		"prn.doc",
		"",
		".",
		"..",
		" file.txt",
		"file.txt ",
		"file.txt.",
		"file\x00.txt",
		"file\n.txt",
		"file\r.txt",
		"file\t.txt",
		strings.Repeat("a", 300),
	}

	for _, filename := range invalidFiles {
		t.Run("Invalid_"+filename, func(t *testing.T) {
			if IsValidFilename(filename) {
				t.Errorf("Invalid filename %q was accepted", filename)
			}
		})
	}
}

func TestIsValidFilename_CaseInsensitiveReservedNames(t *testing.T) {
	reservedTests := []struct {
		name     string
		filename string
	}{
		{"uppercase CON", "CON"},
		{"lowercase con", "con"},
		{"mixed case Con", "Con"},
		{"uppercase CON.txt", "CON.txt"},
		{"lowercase con.txt", "con.txt"},
		{"mixed case Con.txt", "Con.txt"},
		{"uppercase PRN.doc", "PRN.doc"},
		{"lowercase prn.doc", "prn.doc"},
		{"uppercase AUX.log", "AUX.log"},
		{"lowercase aux.log", "aux.log"},
		{"uppercase NUL.dat", "NUL.dat"},
		{"lowercase nul.dat", "nul.dat"},
		{"uppercase COM1.cfg", "COM1.cfg"},
		{"lowercase com1.cfg", "com1.cfg"},
		{"uppercase LPT1.txt", "LPT1.txt"},
		{"lowercase lpt1.txt", "lpt1.txt"},
	}

	for _, tt := range reservedTests {
		t.Run(tt.name, func(t *testing.T) {
			if IsValidFilename(tt.filename) {
				t.Errorf("Reserved filename %q should be invalid", tt.filename)
			}
		})
	}
}

func TestIsValidFilename_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "Just extension",
			filename: ".txt",
			expected: true, // Dotfiles are valid
		},
		{
			name:     "Multiple dots",
			filename: "file...txt",
			expected: true,
		},
		{
			name:     "Very long extension",
			filename: "file." + strings.Repeat("x", 10),
			expected: true,
		},
		{
			name:     "No extension",
			filename: "README",
			expected: true,
		},
		{
			name:     "Unicode filename",
			filename: "文件.txt",
			expected: false, // Non-ASCII not allowed in current regex
		},
		{
			name:     "Emoji filename",
			filename: "😀.txt",
			expected: false, // Emoji not allowed
		},
		{
			name:     "Mixed valid chars",
			filename: "Test-File_123 (Copy).txt",
			expected: true,
		},
	}

	for _, tt := range edgeCases {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("IsValidFilename(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsValidFilename_SecurityTests(t *testing.T) {
	// Test various security-related edge cases
	securityTests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "Path traversal attempt",
			filename: "../../../etc/passwd",
			expected: false,
		},
		{
			name:     "Hidden path traversal",
			filename: "..\\..\\Windows\\System32",
			expected: false,
		},
		{
			name:     "Null byte injection",
			filename: "safe.txt\x00.exe",
			expected: false,
		},
		{
			name:     "Control character injection",
			filename: "file\x01\x02.txt",
			expected: false,
		},
		{
			name:     "URL encoding attempt",
			filename: "file%2e%2e%2fpasswd",
			expected: false, // % not in allowed regex
		},
		{
			name:     "Mixed dangerous chars",
			filename: "file<script>alert('xss')</script>.txt",
			expected: false,
		},
		{
			name:     "Long malicious extension",
			filename: "innocent." + strings.Repeat("exe", 50),
			expected: true, // Long extensions are allowed if they don't match dangerous list
		},
	}

	for _, tt := range securityTests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("Security test %q: IsValidFilename() = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// Test the regex pattern directly
func TestValidFilenameRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Alphanumeric",
			input:    "abc123",
			expected: true,
		},
		{
			name:     "With dots",
			input:    "file.txt",
			expected: true,
		},
		{
			name:     "With underscores",
			input:    "file_name",
			expected: true,
		},
		{
			name:     "With hyphens",
			input:    "file-name",
			expected: true,
		},
		{
			name:     "With spaces",
			input:    "file name",
			expected: true,
		},
		{
			name:     "With parentheses",
			input:    "file(1)",
			expected: true,
		},
		{
			name:     "With brackets",
			input:    "file[1]",
			expected: true,
		},
		{
			name:     "Invalid characters",
			input:    "file@#$",
			expected: false,
		},
		{
			name:     "Path separators",
			input:    "dir/file",
			expected: false,
		},
		{
			name:     "Control characters",
			input:    "file\n",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validFilenameRegex.MatchString(tt.input)
			if result != tt.expected {
				t.Errorf("Regex test %q: MatchString() = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Test dangerous extensions map
func TestDangerousExtensions(t *testing.T) {
	dangerousExts := []string{
		".exe", ".bat", ".cmd", ".sh", ".ps1", ".vbs", ".scr", ".com",
		".pif", ".application", ".gadget", ".msi", ".msp", ".msc", ".jar",
	}

	for _, ext := range dangerousExts {
		t.Run("Dangerous_"+ext, func(t *testing.T) {
			if !dangerousExtensions[ext] {
				t.Errorf("Extension %q should be marked as dangerous", ext)
			}

			// Test case insensitivity by checking uppercase
			upperExt := strings.ToUpper(ext)
			filename := "file" + upperExt
			if IsValidFilename(filename) {
				t.Errorf("File with dangerous extension %q should be invalid", upperExt)
			}
		})
	}
}

// Test reserved names map
func TestReservedNames(t *testing.T) {
	reservedNamesList := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	for _, name := range reservedNamesList {
		t.Run("Reserved_"+name, func(t *testing.T) {
			if !reservedNames[name] {
				t.Errorf("Name %q should be marked as reserved", name)
			}

			// Test the name is rejected
			if IsValidFilename(name) {
				t.Errorf("Reserved name %q should be invalid", name)
			}

			// Test with extension
			filenameWithExt := name + ".txt"
			if IsValidFilename(filenameWithExt) {
				t.Errorf("Reserved name with extension %q should be invalid", filenameWithExt)
			}
		})
	}
}

// Benchmark tests
func BenchmarkIsValidFilename_Valid(b *testing.B) {
	filename := "normal_file_name.txt"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidFilename(filename)
	}
}

func BenchmarkIsValidFilename_Invalid(b *testing.B) {
	filename := "invalid/file\\name.exe"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidFilename(filename)
	}
}

func BenchmarkIsValidFilename_Long(b *testing.B) {
	filename := strings.Repeat("a", 200) + ".txt"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidFilename(filename)
	}
}

// Test concurrent access
func TestIsValidFilename_Concurrent(t *testing.T) {
	// Test concurrent access to ensure thread safety
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	
	testFiles := []string{
		"valid_file.txt",
		"invalid/file.txt",
		"malware.exe",
		"CON",
		"normal-file_123.pdf",
	}
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			
			for j := 0; j < 100; j++ {
				for _, filename := range testFiles {
					IsValidFilename(filename)
				}
			}
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}