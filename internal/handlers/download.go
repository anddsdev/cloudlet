package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anddsdev/cloudlet/internal/services"
	"github.com/anddsdev/cloudlet/internal/utils"
)

func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {

	filePath := strings.TrimPrefix(r.URL.Path, "/api/v1/download")
	if filePath == "" || filePath == "/" {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "File path required")
		return
	}

	data, fileInfo, err := h.fileService.GetFileData(filePath)
	if err != nil {
		if err == services.ErrFileNotFound {
			utils.WriteErrorJSON(w, http.StatusNotFound, "File not found")
		} else {
			utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to read file: "+err.Error())
		}
		return
	}

	// Header for download
	filename := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Type", fileInfo.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size, 10))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Write file data
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
