package database

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSafeQueryBuilder_EscapeLikePattern(t *testing.T) {
	sqb := NewSafeQueryBuilder()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No special characters",
			input:    "simple_path",
			expected: "simple\\_path",
		},
		{
			name:     "With percent",
			input:    "path%with%percent",
			expected: "path\\%with\\%percent",
		},
		{
			name:     "With underscore",
			input:    "path_with_underscore",
			expected: "path\\_with\\_underscore",
		},
		{
			name:     "With backslash",
			input:    "path\\with\\backslash",
			expected: "path\\\\with\\\\backslash",
		},
		{
			name:     "All special characters",
			input:    "path%_\\test",
			expected: "path\\%\\_\\\\test",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sqb.EscapeLikePattern(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSafeQueryBuilder_BuildSafeLikePattern(t *testing.T) {
	sqb := NewSafeQueryBuilder()

	tests := []struct {
		name     string
		basePath string
		suffix   string
		expected string
	}{
		{
			name:     "Simple path",
			basePath: "/documents",
			suffix:   "/%",
			expected: "/documents/%",
		},
		{
			name:     "Path with special characters",
			basePath: "/docs%_test",
			suffix:   "/%",
			expected: "/docs\\%\\_test/%",
		},
		{
			name:     "Root path",
			basePath: "/",
			suffix:   "%",
			expected: "/%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sqb.BuildSafeLikePattern(tt.basePath, tt.suffix)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSafeQueryBuilder_ValidatePathForSQL(t *testing.T) {
	sqb := NewSafeQueryBuilder()

	tests := []struct {
		name        string
		path        string
		expectError bool
		errorText   string
	}{
		{
			name:        "Valid path",
			path:        "/documents/file.txt",
			expectError: false,
		},
		{
			name:        "Path with null byte",
			path:        "/documents/file\x00.txt",
			expectError: true,
			errorText:   "null byte",
		},
		{
			name:        "Too long path",
			path:        strings.Repeat("a", 5000),
			expectError: true,
			errorText:   "too long",
		},
		{
			name:        "Path with DROP keyword",
			path:        "/documents/DROP_table.txt",
			expectError: true,
			errorText:   "suspicious SQL keyword",
		},
		{
			name:        "Path with SELECT keyword",
			path:        "/documents/SELECT_data.txt",
			expectError: true,
			errorText:   "suspicious SQL keyword",
		},
		{
			name:        "Path with SQL injection attempt",
			path:        "/documents/file'; DROP TABLE files; --",
			expectError: true,
			errorText:   "suspicious SQL keyword",
		},
		{
			name:        "Path with UNION keyword",
			path:        "/UNION/file.txt",
			expectError: true,
			errorText:   "suspicious SQL keyword",
		},
		{
			name:        "Empty path",
			path:        "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sqb.ValidatePathForSQL(tt.path)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSafeQueryBuilder_BuildWhereClause(t *testing.T) {
	sqb := NewSafeQueryBuilder()

	tests := []struct {
		name       string
		conditions map[string]interface{}
		expectWhere string
		expectCount int
	}{
		{
			name:        "Empty conditions",
			conditions:  map[string]interface{}{},
			expectWhere: "",
			expectCount: 0,
		},
		{
			name: "Single condition",
			conditions: map[string]interface{}{
				"name": "test.txt",
			},
			expectWhere: "WHERE name = ?",
			expectCount: 1,
		},
		{
			name: "Multiple conditions",
			conditions: map[string]interface{}{
				"name":         "test.txt",
				"is_directory": false,
			},
			expectCount: 2,
		},
		{
			name: "Invalid column name",
			conditions: map[string]interface{}{
				"valid_name":   "test",
				"invalid-name": "should be ignored",
				"invalid name": "should be ignored",
			},
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whereClause, values := sqb.BuildWhereClause(tt.conditions)
			
			if tt.expectWhere != "" && whereClause != tt.expectWhere {
				t.Errorf("Expected WHERE clause %q, got %q", tt.expectWhere, whereClause)
			}
			
			if len(values) != tt.expectCount {
				t.Errorf("Expected %d values, got %d", tt.expectCount, len(values))
			}
			
			// Verify WHERE clause structure for multiple conditions
			if tt.expectCount > 1 {
				if !strings.HasPrefix(whereClause, "WHERE ") {
					t.Errorf("WHERE clause should start with 'WHERE ', got %q", whereClause)
				}
				if !strings.Contains(whereClause, " AND ") {
					t.Errorf("Multiple conditions should be joined with AND, got %q", whereClause)
				}
			}
		})
	}
}

func TestIsValidColumnName(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected bool
	}{
		{
			name:     "Valid simple name",
			column:   "name",
			expected: true,
		},
		{
			name:     "Valid with underscore",
			column:   "file_name",
			expected: true,
		},
		{
			name:     "Valid with numbers",
			column:   "file_name_2",
			expected: true,
		},
		{
			name:     "Invalid with dash",
			column:   "file-name",
			expected: false,
		},
		{
			name:     "Invalid with space",
			column:   "file name",
			expected: false,
		},
		{
			name:     "Invalid starting with number",
			column:   "2file_name",
			expected: false,
		},
		{
			name:     "Invalid empty",
			column:   "",
			expected: false,
		},
		{
			name:     "Invalid with special chars",
			column:   "name!@#",
			expected: false,
		},
		{
			name:     "Valid uppercase",
			column:   "FILE_NAME",
			expected: true,
		},
		{
			name:     "Valid mixed case",
			column:   "fileName",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidColumnName(tt.column)
			if result != tt.expected {
				t.Errorf("Expected %v for column %q, got %v", tt.expected, tt.column, result)
			}
		})
	}
}

// Integration tests with actual database
func TestSafeQueryBuilder_WithDatabase(t *testing.T) {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE files (
			id INTEGER PRIMARY KEY,
			name TEXT,
			path TEXT,
			parent_path TEXT,
			size INTEGER,
			is_directory BOOLEAN
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Insert test data
	testData := []struct {
		name       string
		path       string
		parentPath string
		size       int
		isDir      bool
	}{
		{"file1.txt", "/dir1/file1.txt", "/dir1", 100, false},
		{"file2.txt", "/dir1/file2.txt", "/dir1", 200, false},
		{"subdir", "/dir1/subdir", "/dir1", 0, true},
		{"file3.txt", "/dir1/subdir/file3.txt", "/dir1/subdir", 300, false},
		{"other.txt", "/dir2/other.txt", "/dir2", 150, false},
	}

	for _, data := range testData {
		_, err = db.Exec("INSERT INTO files (name, path, parent_path, size, is_directory) VALUES (?, ?, ?, ?, ?)",
			data.name, data.path, data.parentPath, data.size, data.isDir)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	sqb := NewSafeQueryBuilder()

	t.Run("CountChildrenSafely", func(t *testing.T) {
		count, err := sqb.CountChildrenSafely(db, "/dir1")
		if err != nil {
			t.Fatalf("CountChildrenSafely failed: %v", err)
		}
		if count != 3 { // file1.txt, file2.txt, subdir
			t.Errorf("Expected 3 children, got %d", count)
		}
	})

	t.Run("GetDirectoryStatsSafely", func(t *testing.T) {
		itemCount, totalSize, err := sqb.GetDirectoryStatsSafely(db, "/dir1")
		if err != nil {
			t.Fatalf("GetDirectoryStatsSafely failed: %v", err)
		}
		if itemCount != 3 {
			t.Errorf("Expected 3 items, got %d", itemCount)
		}
		if totalSize != 300 { // file1.txt (100) + file2.txt (200), subdir doesn't count
			t.Errorf("Expected total size 300, got %d", totalSize)
		}
	})

	t.Run("ExecuteInTransaction", func(t *testing.T) {
		err := sqb.ExecuteInTransaction(db, func(tx *sql.Tx) error {
			_, err := tx.Exec("UPDATE files SET size = ? WHERE name = ?", 999, "file1.txt")
			return err
		})
		if err != nil {
			t.Fatalf("ExecuteInTransaction failed: %v", err)
		}

		// Verify the change was committed
		var size int
		err = db.QueryRow("SELECT size FROM files WHERE name = ?", "file1.txt").Scan(&size)
		if err != nil {
			t.Fatalf("Failed to query updated size: %v", err)
		}
		if size != 999 {
			t.Errorf("Expected size 999, got %d", size)
		}
	})

	t.Run("ExecuteInTransaction_Rollback", func(t *testing.T) {
		originalSize := 999 // from previous test
		
		err := sqb.ExecuteInTransaction(db, func(tx *sql.Tx) error {
			_, err := tx.Exec("UPDATE files SET size = ? WHERE name = ?", 1111, "file1.txt")
			if err != nil {
				return err
			}
			// Force an error to trigger rollback
			return fmt.Errorf("intentional error for rollback test")
		})
		
		// Should have failed
		if err == nil {
			t.Fatal("Expected transaction to fail")
		}

		// Verify the change was rolled back
		var size int
		err = db.QueryRow("SELECT size FROM files WHERE name = ?", "file1.txt").Scan(&size)
		if err != nil {
			t.Fatalf("Failed to query size after rollback: %v", err)
		}
		if size != originalSize {
			t.Errorf("Expected size to be rolled back to %d, got %d", originalSize, size)
		}
	})

	t.Run("UpdateChildrenPaths", func(t *testing.T) {
		// Start transaction for testing UpdateChildrenPaths
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		err = sqb.UpdateChildrenPaths(tx, "/dir1", "/new_dir1")
		if err != nil {
			t.Fatalf("UpdateChildrenPaths failed: %v", err)
		}

		// Check if paths were updated correctly
		rows, err := tx.Query("SELECT path FROM files WHERE path LIKE '/new_dir1/%'")
		if err != nil {
			t.Fatalf("Failed to query updated paths: %v", err)
		}
		defer rows.Close()

		var paths []string
		for rows.Next() {
			var path string
			err := rows.Scan(&path)
			if err != nil {
				t.Fatalf("Failed to scan path: %v", err)
			}
			paths = append(paths, path)
		}

		expectedPaths := []string{
			"/new_dir1/file1.txt",
			"/new_dir1/file2.txt",
			"/new_dir1/subdir",
			"/new_dir1/subdir/file3.txt",
		}

		if len(paths) != len(expectedPaths) {
			t.Errorf("Expected %d updated paths, got %d", len(expectedPaths), len(paths))
		}
	})

	t.Run("DeleteDirectoryRecursive", func(t *testing.T) {
		// Start transaction for testing DeleteDirectoryRecursive
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		err = sqb.DeleteDirectoryRecursive(tx, "/dir1")
		if err != nil {
			t.Fatalf("DeleteDirectoryRecursive failed: %v", err)
		}

		// Check if directory and children were deleted
		var count int
		err = tx.QueryRow("SELECT COUNT(*) FROM files WHERE path = ? OR path LIKE ? ESCAPE '\\'", 
			"/dir1", "/dir1/%").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count remaining files: %v", err)
		}

		if count != 0 {
			t.Errorf("Expected 0 files after recursive delete, got %d", count)
		}
	})
}

func TestSafeQueryBuilder_PrepareStatementSafely(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create a dummy table for testing
	_, err = db.Exec("CREATE TABLE files (id INTEGER, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	sqb := NewSafeQueryBuilder()

	tests := []struct {
		name        string
		query       string
		expectError bool
		errorText   string
	}{
		{
			name:        "Valid query with parameters",
			query:       "SELECT * FROM files WHERE name = ?",
			expectError: false,
		},
		{
			name:        "Query with comment syntax",
			query:       "SELECT * FROM files -- comment",
			expectError: true,
			errorText:   "suspicious comment syntax",
		},
		{
			name:        "Query with block comment",
			query:       "SELECT * FROM files /* comment */",
			expectError: true,
			errorText:   "suspicious comment syntax",
		},
		{
			name:        "WHERE clause without parameters",
			query:       "SELECT * FROM files WHERE name = 'test'",
			expectError: true,
			errorText:   "should use parameters",
		},
		{
			name:        "Simple SELECT without WHERE",
			query:       "SELECT * FROM files",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := sqb.PrepareStatementSafely(db, tt.query)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					if stmt != nil {
						stmt.Close()
					}
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if stmt != nil {
					stmt.Close()
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkSafeQueryBuilder_EscapeLikePattern(b *testing.B) {
	sqb := NewSafeQueryBuilder()
	testPath := "/documents/with%special_chars\\test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sqb.EscapeLikePattern(testPath)
	}
}

func BenchmarkSafeQueryBuilder_ValidatePathForSQL(b *testing.B) {
	sqb := NewSafeQueryBuilder()
	testPath := "/documents/normal/path/file.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sqb.ValidatePathForSQL(testPath)
	}
}

func BenchmarkSafeQueryBuilder_BuildWhereClause(b *testing.B) {
	sqb := NewSafeQueryBuilder()
	conditions := map[string]interface{}{
		"name":         "test.txt",
		"size":         1024,
		"is_directory": false,
		"parent_path":  "/documents",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sqb.BuildWhereClause(conditions)
	}
}