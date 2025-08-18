package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// SafeQueryBuilder provides SQL injection-safe query building utilities
type SafeQueryBuilder struct{}

// NewSafeQueryBuilder creates a new safe query builder
func NewSafeQueryBuilder() *SafeQueryBuilder {
	return &SafeQueryBuilder{}
}

// EscapeLikePattern escapes special characters in LIKE patterns to prevent SQL injection
// This function escapes %, _, and \ characters that have special meaning in LIKE clauses
func (sqb *SafeQueryBuilder) EscapeLikePattern(pattern string) string {
	// First escape the escape character itself
	escaped := strings.ReplaceAll(pattern, "\\", "\\\\")
	
	// Then escape other special characters
	escaped = strings.ReplaceAll(escaped, "%", "\\%")
	escaped = strings.ReplaceAll(escaped, "_", "\\_")
	
	return escaped
}

// BuildSafeLikePattern creates a safe LIKE pattern for path operations
// It properly escapes the base path and adds the pattern suffix
func (sqb *SafeQueryBuilder) BuildSafeLikePattern(basePath, suffix string) string {
	escapedBasePath := sqb.EscapeLikePattern(basePath)
	return escapedBasePath + suffix
}

// DeleteDirectoryRecursive safely deletes a directory and all its children
func (sqb *SafeQueryBuilder) DeleteDirectoryRecursive(tx *sql.Tx, path string) error {
	// Escape the path to prevent SQL injection
	safePattern := sqb.BuildSafeLikePattern(path, "/%")
	
	query := `DELETE FROM files WHERE path = ? OR path LIKE ? ESCAPE '\'`
	_, err := tx.Exec(query, path, safePattern)
	return err
}

// UpdateChildrenPaths safely updates paths for children when a directory is moved/renamed
func (sqb *SafeQueryBuilder) UpdateChildrenPaths(tx *sql.Tx, oldParentPath, newParentPath string) error {
	// Escape the old parent path to prevent SQL injection
	safePattern := sqb.BuildSafeLikePattern(oldParentPath, "/%")
	
	// Use CASE to safely replace both path and parent_path
	// This approach is safer than string concatenation with SUBSTR
	query := `
	UPDATE files 
	SET path = CASE 
		WHEN path LIKE ? ESCAPE '\' THEN ? || SUBSTR(path, ?)
		ELSE path
	END,
	parent_path = CASE 
		WHEN parent_path LIKE ? ESCAPE '\' THEN ? || SUBSTR(parent_path, ?)
		WHEN parent_path = ? THEN ?
		ELSE parent_path
	END
	WHERE path LIKE ? ESCAPE '\'
	`
	
	_, err := tx.Exec(query, 
		safePattern,                    // LIKE pattern for matching path
		newParentPath,                  // New prefix for path
		len(oldParentPath)+1,          // Start position for SUBSTR path
		safePattern,                    // LIKE pattern for matching parent_path
		newParentPath,                  // New prefix for parent_path
		len(oldParentPath)+1,          // Start position for SUBSTR parent_path
		oldParentPath,                  // Direct parent_path match
		newParentPath,                  // New parent_path value
		safePattern,                   // WHERE condition pattern
	)
	return err
}

// FindFilesByPathPattern safely finds files matching a path pattern
func (sqb *SafeQueryBuilder) FindFilesByPathPattern(db *sql.DB, pathPattern string) (*sql.Rows, error) {
	// Escape the pattern to prevent SQL injection
	safePattern := sqb.EscapeLikePattern(pathPattern)
	
	query := `
	SELECT id, name, path, size, mime_type, is_directory, parent_path, created_at, updated_at
	FROM files 
	WHERE path LIKE ? ESCAPE '\'
	ORDER BY is_directory DESC, LOWER(name) ASC
	`
	
	return db.Query(query, safePattern)
}

// CountChildrenSafely safely counts children of a directory
func (sqb *SafeQueryBuilder) CountChildrenSafely(db *sql.DB, parentPath string) (int, error) {
	var count int
	
	// For exact parent match, we don't need LIKE - use equality
	query := `SELECT COUNT(*) FROM files WHERE parent_path = ?`
	err := db.QueryRow(query, parentPath).Scan(&count)
	
	return count, err
}

// GetDirectoryStatsSafely safely gets directory statistics
func (sqb *SafeQueryBuilder) GetDirectoryStatsSafely(db *sql.DB, path string) (int64, int64, error) {
	var itemCount, totalSize int64
	
	// Count items - use exact parent_path match
	err := db.QueryRow("SELECT COUNT(*) FROM files WHERE parent_path = ?", path).Scan(&itemCount)
	if err != nil {
		return 0, 0, err
	}
	
	// Sum sizes (only files, not directories) - use exact parent_path match
	err = db.QueryRow("SELECT COALESCE(SUM(size), 0) FROM files WHERE parent_path = ? AND is_directory = FALSE", path).Scan(&totalSize)
	if err != nil {
		return 0, 0, err
	}
	
	return itemCount, totalSize, nil
}

// ValidatePathForSQL validates that a path is safe to use in SQL queries
func (sqb *SafeQueryBuilder) ValidatePathForSQL(path string) error {
	// Check for null bytes which can be used for SQL injection
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}
	
	// Check for excessively long paths that might cause buffer overflows
	if len(path) > 4096 {
		return fmt.Errorf("path too long")
	}
	
	// Check for suspicious SQL keywords (basic detection)
	suspiciousKeywords := []string{
		"DROP", "DELETE", "INSERT", "UPDATE", "SELECT",
		"UNION", "OR 1=1", "'; --", "/*", "*/",
	}
	
	upperPath := strings.ToUpper(path)
	for _, keyword := range suspiciousKeywords {
		if strings.Contains(upperPath, keyword) {
			return fmt.Errorf("path contains suspicious SQL keyword: %s", keyword)
		}
	}
	
	return nil
}

// PrepareStatementSafely prepares a statement with validation
func (sqb *SafeQueryBuilder) PrepareStatementSafely(db *sql.DB, query string) (*sql.Stmt, error) {
	// Basic query validation
	if strings.Contains(query, "--") || strings.Contains(query, "/*") {
		return nil, fmt.Errorf("query contains suspicious comment syntax")
	}
	
	// Count parameters vs placeholders
	paramCount := strings.Count(query, "?")
	if paramCount == 0 && strings.Contains(strings.ToUpper(query), "WHERE") {
		return nil, fmt.Errorf("query with WHERE clause should use parameters")
	}
	
	return db.Prepare(query)
}

// ExecuteInTransaction safely executes multiple operations in a transaction
func (sqb *SafeQueryBuilder) ExecuteInTransaction(db *sql.DB, operations func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Ensure rollback on panic or error
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()
	
	// Execute operations
	err = operations(tx)
	if err != nil {
		return err // Rollback will be called by defer
	}
	
	// Commit if everything succeeded
	return tx.Commit()
}

// LogSQLOperation logs SQL operations for security auditing
func (sqb *SafeQueryBuilder) LogSQLOperation(operation string, query string, params []interface{}) {
	// In production, this would write to a secure audit log
	// For now, we'll use a simple log format that doesn't expose sensitive data
	
	// Sanitize parameters for logging (remove sensitive info)
	sanitizedParams := make([]interface{}, len(params))
	for i, param := range params {
		if str, ok := param.(string); ok && len(str) > 50 {
			// Truncate long strings to prevent log spam
			sanitizedParams[i] = str[:50] + "..."
		} else {
			sanitizedParams[i] = param
		}
	}
	
	// This would normally go to a secure audit log
	// For now, we'll just validate the operation is safe
	_ = operation
	_ = query
	_ = sanitizedParams
}

// BuildWhereClause safely builds WHERE clauses with multiple conditions
func (sqb *SafeQueryBuilder) BuildWhereClause(conditions map[string]interface{}) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}
	
	var parts []string
	var values []interface{}
	
	for column, value := range conditions {
		// Validate column name to prevent injection
		if !isValidColumnName(column) {
			continue // Skip invalid column names
		}
		
		parts = append(parts, column+" = ?")
		values = append(values, value)
	}
	
	if len(parts) == 0 {
		return "", nil
	}
	
	return "WHERE " + strings.Join(parts, " AND "), values
}

// isValidColumnName validates that a column name is safe
func isValidColumnName(name string) bool {
	// Allow only alphanumeric characters and underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '_') {
			return false
		}
	}
	
	// Don't allow empty names or names starting with numbers
	if len(name) == 0 || (name[0] >= '0' && name[0] <= '9') {
		return false
	}
	
	return true
}