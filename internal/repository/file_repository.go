package repository

import (
	"database/sql"
	"fmt"

	"github.com/anddsdev/cloudlet/internal/database"
	"github.com/anddsdev/cloudlet/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type FileRepository struct {
	db          *sql.DB
	safeQueries *database.SafeQueryBuilder
}

func NewFileRepository(dsn string, maxConn int) (*FileRepository, error) {
	initializer := database.NewDatabaseInitializer(dsn)
	if err := initializer.InitializeDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(maxConn)
	db.SetMaxIdleConns(maxConn / 2)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Printf("Database connection established successfully: %s\n", dsn)

	return &FileRepository{
		db: db,
	}, nil
}

func (r *FileRepository) GetFilesByPath(parentPath string) ([]*models.FileInfo, error) {
	query := `
	SELECT id, name, path, size, mime_type, is_directory, parent_path, created_at, updated_at
	FROM files 
	WHERE parent_path = ? 
	ORDER BY is_directory DESC, LOWER(name) ASC
	`

	rows, err := r.db.Query(query, parentPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*models.FileInfo
	for rows.Next() {
		file := &models.FileInfo{}
		err := rows.Scan(
			&file.ID, &file.Name, &file.Path, &file.Size,
			&file.MimeType, &file.IsDirectory, &file.ParentPath,
			&file.CreatedAt, &file.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// If it's a directory, calculate statistics
		if file.IsDirectory {
			file.ItemCount, file.TotalSize = r.getDirectoryStats(file.Path)
		}

		files = append(files, file)
	}

	return files, nil
}

func (r *FileRepository) GetFileByPath(path string) (*models.FileInfo, error) {
	query := `
	SELECT id, name, path, size, mime_type, is_directory, parent_path, created_at, updated_at
	FROM files WHERE path = ?
	`

	file := &models.FileInfo{}
	err := r.db.QueryRow(query, path).Scan(
		&file.ID, &file.Name, &file.Path, &file.Size,
		&file.MimeType, &file.IsDirectory, &file.ParentPath,
		&file.CreatedAt, &file.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// If it's a directory, get statistics
	if file.IsDirectory {
		file.ItemCount, file.TotalSize = r.getDirectoryStats(file.Path)
	}

	return file, nil
}

func (r *FileRepository) InsertFile(file *models.FileInfo) error {
	query := `
	INSERT INTO files (name, path, size, mime_type, is_directory, parent_path)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		file.Name, file.Path, file.Size, file.MimeType,
		file.IsDirectory, file.ParentPath,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	file.ID = id
	return nil
}

func (r *FileRepository) CreateDirectory(name, parentPath string) (*models.FileInfo, error) {
	fullPath := r.buildPath(parentPath, name)

	if r.pathExists(fullPath) {
		return nil, fmt.Errorf("directory already exists: %s", fullPath)
	}

	if parentPath != "/" && !r.pathExists(parentPath) {
		return nil, fmt.Errorf("parent directory does not exist: %s", parentPath)
	}

	dir := &models.FileInfo{
		Name:        name,
		Path:        fullPath,
		Size:        0,
		MimeType:    "inode/directory",
		IsDirectory: true,
		ParentPath:  parentPath,
	}

	err := r.InsertFile(dir)
	if err != nil {
		return nil, err
	}

	return dir, nil
}

func (r *FileRepository) RenameFile(oldPath, newName string) error {
	file, err := r.GetFileByPath(oldPath)
	if err != nil {
		return err
	}

	newPath := r.buildPath(file.ParentPath, newName)

	if r.pathExists(newPath) {
		return fmt.Errorf("file already exists: %s", newPath)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the file/directory
	_, err = tx.Exec("UPDATE files SET name = ?, path = ? WHERE path = ?", newName, newPath, oldPath)
	if err != nil {
		return err
	}

	// If it's a directory, update all children recursively
	if file.IsDirectory {
		err = r.updateChildrenPaths(tx, oldPath, newPath)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *FileRepository) MoveFile(sourcePath, destinationPath string) error {

	sourceFile, err := r.GetFileByPath(sourcePath)
	if err != nil {
		return err
	}

	if destinationPath != "/" && !r.isDirectory(destinationPath) {
		return fmt.Errorf("destination is not a directory: %s", destinationPath)
	}

	newPath := r.buildPath(destinationPath, sourceFile.Name)

	if r.pathExists(newPath) {
		return fmt.Errorf("file already exists at destination: %s", newPath)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE files SET path = ?, parent_path = ? WHERE path = ?",
		newPath, destinationPath, sourcePath)
	if err != nil {
		return err
	}

	if sourceFile.IsDirectory {
		err = r.updateChildrenPaths(tx, sourcePath, newPath)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *FileRepository) DeleteFile(path string) error {

	file, err := r.GetFileByPath(path)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// If it's a directory, check if it's empty or delete recursively
	if file.IsDirectory {
		// Count children
		var count int
		err = tx.QueryRow("SELECT COUNT(*) FROM files WHERE parent_path = ?", path).Scan(&count)
		if err != nil {
			return err
		}

		// For MVP, only allow deleting empty directories
		// In the future, add recursive deletion
		if count > 0 {
			return fmt.Errorf("directory not empty: %s", path)
		}
	}

	_, err = tx.Exec("DELETE FROM files WHERE path = ?", path)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *FileRepository) DeleteDirectoryRecursive(path string) error {
	// Validate path for SQL safety
	if err := r.safeQueries.ValidatePathForSQL(path); err != nil {
		return fmt.Errorf("invalid path for SQL operation: %w", err)
	}

	// Use safe transaction execution
	return r.safeQueries.ExecuteInTransaction(r.db, func(tx *sql.Tx) error {
		return r.safeQueries.DeleteDirectoryRecursive(tx, path)
	})
}

func (r *FileRepository) buildPath(parent, name string) string {
	if parent == "/" {
		return "/" + name
	}
	return parent + "/" + name
}

func (r *FileRepository) pathExists(path string) bool {
	var count int
	r.db.QueryRow("SELECT COUNT(*) FROM files WHERE path = ?", path).Scan(&count)
	return count > 0
}

func (r *FileRepository) isDirectory(path string) bool {
	var isDir bool
	err := r.db.QueryRow("SELECT is_directory FROM files WHERE path = ?", path).Scan(&isDir)
	return err == nil && isDir
}

func (r *FileRepository) getDirectoryStats(path string) (int64, int64) {
	// Use safe queries to prevent SQL injection
	itemCount, totalSize, err := r.safeQueries.GetDirectoryStatsSafely(r.db, path)
	if err != nil {
		// Log error but return zero values for backwards compatibility
		return 0, 0
	}

	return itemCount, totalSize
}

func (r *FileRepository) updateChildrenPaths(tx *sql.Tx, oldParentPath, newParentPath string) error {
	// Validate paths for SQL safety
	if err := r.safeQueries.ValidatePathForSQL(oldParentPath); err != nil {
		return fmt.Errorf("invalid old parent path for SQL operation: %w", err)
	}
	if err := r.safeQueries.ValidatePathForSQL(newParentPath); err != nil {
		return fmt.Errorf("invalid new parent path for SQL operation: %w", err)
	}

	// Use safe query builder to prevent SQL injection
	return r.safeQueries.UpdateChildrenPaths(tx, oldParentPath, newParentPath)
}

func (r *FileRepository) Close() error {
	return r.db.Close()
}
