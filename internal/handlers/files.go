package handlers

import (
	"net/http"
	"strings"

	"github.com/anddsdev/cloudlet/internal/services"
	"github.com/anddsdev/cloudlet/internal/utils"
)

func (h *Handlers) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		utils.WriteErrorJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var path string
	if strings.HasPrefix(r.URL.Path, "/api/v1/files") {
		path = strings.TrimPrefix(r.URL.Path, "/api/v1/files")
	} else if strings.HasPrefix(r.URL.Path, "/api/v1/directories") {
		path = strings.TrimPrefix(r.URL.Path, "/api/v1/directories")
	}

	if path == "" {
		path = "/"
	}

	listing, err := h.fileService.GetDirectoryListing(path)
	if err != nil {
		utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to list files: "+err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, listing)
}

func (h *Handlers) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		utils.WriteErrorJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/files")
	if path == "" || path == "/" {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Invalid file path")
		return
	}

	err := h.fileService.DeleteFile(path)
	if err != nil {
		if err == services.ErrFileNotFound {
			utils.WriteErrorJSON(w, http.StatusNotFound, "File not found")
		} else if strings.Contains(err.Error(), "not empty") {
			utils.WriteErrorJSON(w, http.StatusConflict, "Directory not empty")
		} else {
			utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to delete file: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "File deleted successfully",
		"path":    path,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
