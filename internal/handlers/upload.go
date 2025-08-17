package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/utils"
)

func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
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

	if !utils.IsValidFilename(header.Filename) {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Invalid filename")
		return
	}

	// Use streaming for files larger than 10MB to prevent memory leaks
	const streamingThreshold = 10 * 1024 * 1024 // 10MB
	
	if header.Size > streamingThreshold {
		// Use streaming upload for large files
		err = h.fileService.SaveFileStream(header.Filename, targetPath, file, header.Size)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to save file: "+err.Error())
			return
		}
	} else {
		// Use traditional upload for small files (better performance)
		data, err := io.ReadAll(file)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to read file")
			return
		}

		err = h.fileService.SaveFile(header.Filename, targetPath, data)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to save file: "+err.Error())
			return
		}
	}

	response := &models.UploadResponse{
		Success:  true,
		Filename: header.Filename,
		Size:     header.Size,
		Path:     targetPath,
		Message:  "File uploaded successfully",
	}

	utils.WriteJSON(w, http.StatusCreated, response)
}
