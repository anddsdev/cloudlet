package models

type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type UploadResponse struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
	Message  string `json:"message"`
}

// FileUploadResult represents the result of uploading a single file
type FileUploadResult struct {
	Filename    string `json:"filename"`
	OriginalName string `json:"originalName"`
	Size        int64  `json:"size"`
	Path        string `json:"path"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	MimeType    string `json:"mimeType"`
	Index       int    `json:"index"` // Original index in the request
}

// MultipleUploadResponse represents the response for multiple file uploads
type MultipleUploadResponse struct {
	Success           bool               `json:"success"`
	TotalFiles        int                `json:"totalFiles"`
	SuccessfulFiles   int                `json:"successfulFiles"`
	FailedFiles       int                `json:"failedFiles"`
	TotalSize         int64              `json:"totalSize"`
	ProcessedSize     int64              `json:"processedSize"`
	Files             []FileUploadResult `json:"files"`
	Message           string             `json:"message"`
	ProcessingTimeMs  int64              `json:"processingTimeMs"`
	Strategy          string             `json:"strategy"` // "hybrid", "sequential", "parallel"
}

// UploadValidationResult contains validation results for multiple files
type UploadValidationResult struct {
	Valid             bool     `json:"valid"`
	TotalFiles        int      `json:"totalFiles"`
	TotalSize         int64    `json:"totalSize"`
	InvalidFiles      []string `json:"invalidFiles,omitempty"`
	OversizedFiles    []string `json:"oversizedFiles,omitempty"`
	DuplicateFiles    []string `json:"duplicateFiles,omitempty"`
	DangerousFiles    []string `json:"dangerousFiles,omitempty"`
	ValidationErrors  []string `json:"validationErrors,omitempty"`
	MaxFilesExceeded  bool     `json:"maxFilesExceeded"`
	MaxSizeExceeded   bool     `json:"maxSizeExceeded"`
}

// BatchUploadProgress represents progress information for batch uploads
type BatchUploadProgress struct {
	BatchID           string  `json:"batchId"`
	TotalFiles        int     `json:"totalFiles"`
	ProcessedFiles    int     `json:"processedFiles"`
	SuccessfulFiles   int     `json:"successfulFiles"`
	FailedFiles       int     `json:"failedFiles"`
	CurrentFile       string  `json:"currentFile"`
	PercentComplete   float64 `json:"percentComplete"`
	EstimatedTimeLeft int64   `json:"estimatedTimeLeftMs"`
	StartTime         int64   `json:"startTime"`
	Status            string  `json:"status"` // "processing", "completed", "failed", "cancelled"
}
