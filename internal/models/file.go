package models

import (
	"time"
)

type FileInfo struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Path        string    `json:"path" db:"path"`
	Size        int64     `json:"size" db:"size"`
	MimeType    string    `json:"mime_type" db:"mime_type"`
	IsDirectory bool      `json:"is_directory" db:"is_directory"`
	ParentPath  string    `json:"parent_path" db:"parent_path"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	ItemCount int64 `json:"item_count,omitempty"`
	TotalSize int64 `json:"total_size,omitempty"`
}

type DirectoryListing struct {
	Path        string       `json:"path"`
	ParentPath  string       `json:"parent_path"`
	Files       []*FileInfo  `json:"files"`
	Directories []*FileInfo  `json:"directories"`
	TotalFiles  int          `json:"total_files"`
	TotalDirs   int          `json:"total_directories"`
	TotalSize   int64        `json:"total_size"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs"`
}

type Breadcrumb struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type CreateDirectoryRequest struct {
	Name       string `json:"name"`
	ParentPath string `json:"parent_path"`
}

type MoveRequest struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
	NewName         string `json:"new_name,omitempty"`
}

type RenameRequest struct {
	Path    string `json:"path"`
	NewName string `json:"new_name"`
}
