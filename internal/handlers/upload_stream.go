package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/utils"
)

// UploadStream handles file uploads using streaming to prevent memory leaks
// This handler should be used for large files or when memory conservation is important
func (h *Handlers) UploadStream(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(int64(h.cfg.Server.MaxMemory))
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	if header.Size > h.cfg.Server.MaxFileSize {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "File too large. Max size: "+strconv.FormatInt(h.cfg.Server.MaxFileSize, 10)+" bytes")
		return
	}

	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "/"
	}

	// Validate filename using security module
	if !utils.IsValidFilename(header.Filename) {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	// Use streaming upload to prevent memory leaks
	err = h.fileService.SaveFileStream(header.Filename, targetPath, file, header.Size)
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to save file: "+err.Error())
		return
	}

	response := &models.UploadResponse{
		Success:  true,
		Filename: header.Filename,
		Size:     header.Size,
		Path:     targetPath,
		Message:  "File uploaded successfully using streaming",
	}

	utils.WriteJSON(w, http.StatusCreated, response)
}

// UploadChunked handles chunked file uploads for very large files
// This allows uploading files larger than available memory by processing them in chunks
func (h *Handlers) UploadChunked(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(int64(h.cfg.Server.MaxMemory))
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	if header.Size > h.cfg.Server.MaxFileSize {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "File too large. Max size: "+strconv.FormatInt(h.cfg.Server.MaxFileSize, 10)+" bytes")
		return
	}

	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "/"
	}

	// Validate filename
	if !utils.IsValidFilename(header.Filename) {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	// Create a chunked reader that processes the file in small chunks
	chunkSize := 64 * 1024 // 64KB chunks
	chunkedReader := &ChunkedReader{
		reader:    file,
		chunkSize: chunkSize,
	}

	// Use streaming upload with chunked processing
	err = h.fileService.SaveFileStream(header.Filename, targetPath, chunkedReader, header.Size)
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to save file: "+err.Error())
		return
	}

	response := &models.UploadResponse{
		Success:  true,
		Filename: header.Filename,
		Size:     header.Size,
		Path:     targetPath,
		Message:  fmt.Sprintf("File uploaded successfully using %d KB chunks", chunkSize/1024),
	}

	utils.WriteJSON(w, http.StatusCreated, response)
}

// ChunkedReader wraps an io.Reader to process data in fixed-size chunks
// This helps prevent memory spikes when processing large files
type ChunkedReader struct {
	reader    io.Reader
	chunkSize int
	buffer    []byte
}

// Read implements io.Reader interface with chunked processing
func (cr *ChunkedReader) Read(p []byte) (n int, err error) {
	// Ensure we don't read more than our chunk size at once
	if len(p) > cr.chunkSize {
		p = p[:cr.chunkSize]
	}

	// Allocate buffer if needed
	if cr.buffer == nil {
		cr.buffer = make([]byte, cr.chunkSize)
	}

	// Read from underlying reader
	n, err = cr.reader.Read(p)
	return n, err
}

// UploadWithProgressTracking handles uploads with progress tracking
// This is useful for large files where clients need progress feedback
func (h *Handlers) UploadWithProgressTracking(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(int64(h.cfg.Server.MaxMemory))
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	if header.Size > h.cfg.Server.MaxFileSize {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "File too large. Max size: "+strconv.FormatInt(h.cfg.Server.MaxFileSize, 10)+" bytes")
		return
	}

	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = "/"
	}

	// Validate filename
	if !utils.IsValidFilename(header.Filename) {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	// Create a progress tracking reader
	progressReader := &ProgressTrackingReader{
		reader:    file,
		total:     header.Size,
		progress:  0,
		onProgress: func(bytesRead, total int64) {
			// In a real implementation, this could send progress updates
			// via WebSocket, Server-Sent Events, or store progress in a cache
			percentage := float64(bytesRead) / float64(total) * 100
			_ = percentage // For now, just calculate but don't send
		},
	}

	// Use streaming upload with progress tracking
	err = h.fileService.SaveFileStream(header.Filename, targetPath, progressReader, header.Size)
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to save file: "+err.Error())
		return
	}

	response := &models.UploadResponse{
		Success:  true,
		Filename: header.Filename,
		Size:     header.Size,
		Path:     targetPath,
		Message:  "File uploaded successfully with progress tracking",
	}

	utils.WriteJSON(w, http.StatusCreated, response)
}

// ProgressTrackingReader wraps an io.Reader to track upload progress
type ProgressTrackingReader struct {
	reader     io.Reader
	total      int64
	progress   int64
	onProgress func(bytesRead, total int64)
}

// Read implements io.Reader interface with progress tracking
func (ptr *ProgressTrackingReader) Read(p []byte) (n int, err error) {
	n, err = ptr.reader.Read(p)
	
	if n > 0 {
		ptr.progress += int64(n)
		if ptr.onProgress != nil {
			ptr.onProgress(ptr.progress, ptr.total)
		}
	}
	
	return n, err
}