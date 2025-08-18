package models

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestFileInfo_JSONMarshaling(t *testing.T) {
	now := time.Now()
	file := &FileInfo{
		ID:          123,
		Name:        "test.txt",
		Path:        "/documents/test.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/documents",
		CreatedAt:   now,
		UpdatedAt:   now,
		ItemCount:   0,
		TotalSize:   1024,
	}

	// Test marshaling
	data, err := json.Marshal(file)
	if err != nil {
		t.Fatalf("Failed to marshal FileInfo: %v", err)
	}

	// Test unmarshaling
	var unmarshaled FileInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal FileInfo: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != file.ID {
		t.Errorf("ID mismatch: expected %d, got %d", file.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != file.Name {
		t.Errorf("Name mismatch: expected %s, got %s", file.Name, unmarshaled.Name)
	}
	if unmarshaled.Path != file.Path {
		t.Errorf("Path mismatch: expected %s, got %s", file.Path, unmarshaled.Path)
	}
	if unmarshaled.Size != file.Size {
		t.Errorf("Size mismatch: expected %d, got %d", file.Size, unmarshaled.Size)
	}
	if unmarshaled.MimeType != file.MimeType {
		t.Errorf("MimeType mismatch: expected %s, got %s", file.MimeType, unmarshaled.MimeType)
	}
	if unmarshaled.IsDirectory != file.IsDirectory {
		t.Errorf("IsDirectory mismatch: expected %v, got %v", file.IsDirectory, unmarshaled.IsDirectory)
	}
	if unmarshaled.ParentPath != file.ParentPath {
		t.Errorf("ParentPath mismatch: expected %s, got %s", file.ParentPath, unmarshaled.ParentPath)
	}
}

func TestDirectoryListing_JSONMarshaling(t *testing.T) {
	files := []*FileInfo{
		{
			ID:          1,
			Name:        "file1.txt",
			Path:        "/test/file1.txt",
			Size:        100,
			MimeType:    "text/plain",
			IsDirectory: false,
			ParentPath:  "/test",
		},
		{
			ID:          2,
			Name:        "file2.pdf",
			Path:        "/test/file2.pdf",
			Size:        200,
			MimeType:    "application/pdf",
			IsDirectory: false,
			ParentPath:  "/test",
		},
	}

	directories := []*FileInfo{
		{
			ID:          3,
			Name:        "subdir",
			Path:        "/test/subdir",
			Size:        0,
			MimeType:    "inode/directory",
			IsDirectory: true,
			ParentPath:  "/test",
			ItemCount:   5,
			TotalSize:   1024,
		},
	}

	breadcrumbs := []Breadcrumb{
		{Name: "Home", Path: "/"},
		{Name: "test", Path: "/test"},
	}

	listing := &DirectoryListing{
		Path:        "/test",
		ParentPath:  "/",
		Files:       files,
		Directories: directories,
		TotalFiles:  2,
		TotalDirs:   1,
		TotalSize:   300,
		Breadcrumbs: breadcrumbs,
	}

	// Test marshaling
	data, err := json.Marshal(listing)
	if err != nil {
		t.Fatalf("Failed to marshal DirectoryListing: %v", err)
	}

	// Test unmarshaling
	var unmarshaled DirectoryListing
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal DirectoryListing: %v", err)
	}

	// Verify fields
	if unmarshaled.Path != listing.Path {
		t.Errorf("Path mismatch: expected %s, got %s", listing.Path, unmarshaled.Path)
	}
	if len(unmarshaled.Files) != len(listing.Files) {
		t.Errorf("Files count mismatch: expected %d, got %d", len(listing.Files), len(unmarshaled.Files))
	}
	if len(unmarshaled.Directories) != len(listing.Directories) {
		t.Errorf("Directories count mismatch: expected %d, got %d", len(listing.Directories), len(unmarshaled.Directories))
	}
	if unmarshaled.TotalFiles != listing.TotalFiles {
		t.Errorf("TotalFiles mismatch: expected %d, got %d", listing.TotalFiles, unmarshaled.TotalFiles)
	}
	if unmarshaled.TotalDirs != listing.TotalDirs {
		t.Errorf("TotalDirs mismatch: expected %d, got %d", listing.TotalDirs, unmarshaled.TotalDirs)
	}
	if unmarshaled.TotalSize != listing.TotalSize {
		t.Errorf("TotalSize mismatch: expected %d, got %d", listing.TotalSize, unmarshaled.TotalSize)
	}
}

func TestBreadcrumb_JSONMarshaling(t *testing.T) {
	breadcrumb := Breadcrumb{
		Name: "Documents",
		Path: "/documents",
	}

	data, err := json.Marshal(breadcrumb)
	if err != nil {
		t.Fatalf("Failed to marshal Breadcrumb: %v", err)
	}

	var unmarshaled Breadcrumb
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Breadcrumb: %v", err)
	}

	if unmarshaled.Name != breadcrumb.Name {
		t.Errorf("Name mismatch: expected %s, got %s", breadcrumb.Name, unmarshaled.Name)
	}
	if unmarshaled.Path != breadcrumb.Path {
		t.Errorf("Path mismatch: expected %s, got %s", breadcrumb.Path, unmarshaled.Path)
	}
}

func TestCreateDirectoryRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateDirectoryRequest
		valid   bool
	}{
		{
			name: "Valid request",
			request: CreateDirectoryRequest{
				Name:       "new_folder",
				ParentPath: "/documents",
			},
			valid: true,
		},
		{
			name: "Empty name",
			request: CreateDirectoryRequest{
				Name:       "",
				ParentPath: "/documents",
			},
			valid: false,
		},
		{
			name: "Root parent path",
			request: CreateDirectoryRequest{
				Name:       "root_folder",
				ParentPath: "/",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			var unmarshaled CreateDirectoryRequest
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal request: %v", err)
			}

			if unmarshaled.Name != tt.request.Name {
				t.Errorf("Name mismatch: expected %s, got %s", tt.request.Name, unmarshaled.Name)
			}
			if unmarshaled.ParentPath != tt.request.ParentPath {
				t.Errorf("ParentPath mismatch: expected %s, got %s", tt.request.ParentPath, unmarshaled.ParentPath)
			}
		})
	}
}

func TestMoveRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request MoveRequest
		valid   bool
	}{
		{
			name: "Valid move without rename",
			request: MoveRequest{
				SourcePath:      "/documents/file.txt",
				DestinationPath: "/backup",
			},
			valid: true,
		},
		{
			name: "Valid move with rename",
			request: MoveRequest{
				SourcePath:      "/documents/file.txt",
				DestinationPath: "/backup",
				NewName:         "renamed_file.txt",
			},
			valid: true,
		},
		{
			name: "Invalid - empty source",
			request: MoveRequest{
				SourcePath:      "",
				DestinationPath: "/backup",
			},
			valid: false,
		},
		{
			name: "Invalid - empty destination",
			request: MoveRequest{
				SourcePath:      "/documents/file.txt",
				DestinationPath: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			var unmarshaled MoveRequest
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal request: %v", err)
			}

			if unmarshaled.SourcePath != tt.request.SourcePath {
				t.Errorf("SourcePath mismatch: expected %s, got %s", tt.request.SourcePath, unmarshaled.SourcePath)
			}
			if unmarshaled.DestinationPath != tt.request.DestinationPath {
				t.Errorf("DestinationPath mismatch: expected %s, got %s", tt.request.DestinationPath, unmarshaled.DestinationPath)
			}
			if unmarshaled.NewName != tt.request.NewName {
				t.Errorf("NewName mismatch: expected %s, got %s", tt.request.NewName, unmarshaled.NewName)
			}
		})
	}
}

func TestRenameRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request RenameRequest
		valid   bool
	}{
		{
			name: "Valid rename",
			request: RenameRequest{
				Path:    "/documents/old_name.txt",
				NewName: "new_name.txt",
			},
			valid: true,
		},
		{
			name: "Invalid - empty path",
			request: RenameRequest{
				Path:    "",
				NewName: "new_name.txt",
			},
			valid: false,
		},
		{
			name: "Invalid - empty new name",
			request: RenameRequest{
				Path:    "/documents/old_name.txt",
				NewName: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			var unmarshaled RenameRequest
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal request: %v", err)
			}

			if unmarshaled.Path != tt.request.Path {
				t.Errorf("Path mismatch: expected %s, got %s", tt.request.Path, unmarshaled.Path)
			}
			if unmarshaled.NewName != tt.request.NewName {
				t.Errorf("NewName mismatch: expected %s, got %s", tt.request.NewName, unmarshaled.NewName)
			}
		})
	}
}

func TestFileInfo_DirectoryStats(t *testing.T) {
	// Test directory with stats
	dir := &FileInfo{
		ID:          1,
		Name:        "test_dir",
		Path:        "/test_dir",
		Size:        0,
		MimeType:    "inode/directory",
		IsDirectory: true,
		ParentPath:  "/",
		ItemCount:   10,
		TotalSize:   2048,
	}

	// Verify directory stats are included in JSON
	data, err := json.Marshal(dir)
	if err != nil {
		t.Fatalf("Failed to marshal directory: %v", err)
	}

	var unmarshaled FileInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal directory: %v", err)
	}

	if unmarshaled.ItemCount != dir.ItemCount {
		t.Errorf("ItemCount mismatch: expected %d, got %d", dir.ItemCount, unmarshaled.ItemCount)
	}
	if unmarshaled.TotalSize != dir.TotalSize {
		t.Errorf("TotalSize mismatch: expected %d, got %d", dir.TotalSize, unmarshaled.TotalSize)
	}
}

func TestFileInfo_EmptyValues(t *testing.T) {
	// Test with minimal values
	file := &FileInfo{
		Name:        "minimal.txt",
		Path:        "/minimal.txt",
		IsDirectory: false,
	}

	data, err := json.Marshal(file)
	if err != nil {
		t.Fatalf("Failed to marshal minimal file: %v", err)
	}

	var unmarshaled FileInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal minimal file: %v", err)
	}

	if unmarshaled.Name != file.Name {
		t.Errorf("Name mismatch: expected %s, got %s", file.Name, unmarshaled.Name)
	}
	if unmarshaled.Path != file.Path {
		t.Errorf("Path mismatch: expected %s, got %s", file.Path, unmarshaled.Path)
	}
	if unmarshaled.IsDirectory != file.IsDirectory {
		t.Errorf("IsDirectory mismatch: expected %v, got %v", file.IsDirectory, unmarshaled.IsDirectory)
	}
}

// Benchmark tests
func BenchmarkFileInfo_Marshal(b *testing.B) {
	file := &FileInfo{
		ID:          123,
		Name:        "benchmark.txt",
		Path:        "/benchmark/benchmark.txt",
		Size:        1024,
		MimeType:    "text/plain",
		IsDirectory: false,
		ParentPath:  "/benchmark",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(file)
		if err != nil {
			b.Fatalf("Marshal failed: %v", err)
		}
	}
}

func BenchmarkDirectoryListing_Marshal(b *testing.B) {
	files := make([]*FileInfo, 100)
	for i := 0; i < 100; i++ {
		files[i] = &FileInfo{
			ID:          int64(i),
			Name:        fmt.Sprintf("file%d.txt", i),
			Path:        fmt.Sprintf("/test/file%d.txt", i),
			Size:        1024,
			MimeType:    "text/plain",
			IsDirectory: false,
			ParentPath:  "/test",
		}
	}

	listing := &DirectoryListing{
		Path:        "/test",
		ParentPath:  "/",
		Files:       files,
		Directories: []*FileInfo{},
		TotalFiles:  100,
		TotalDirs:   0,
		TotalSize:   102400,
		Breadcrumbs: []Breadcrumb{{Name: "Home", Path: "/"}, {Name: "test", Path: "/test"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(listing)
		if err != nil {
			b.Fatalf("Marshal failed: %v", err)
		}
	}
}