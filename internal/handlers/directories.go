package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anddsdev/cloudlet/internal/models"
	"github.com/anddsdev/cloudlet/internal/utils"
)

func (h *Handlers) CreateDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.WriteErrorJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.CreateDirectoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Name == "" {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Directory name is required")
		return
	}

	if req.ParentPath == "" {
		req.ParentPath = "/"
	}

	directory, err := h.fileService.CreateDirectory(req.Name, req.ParentPath)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.WriteErrorJSON(w, http.StatusConflict, err.Error())
		} else if strings.Contains(err.Error(), "does not exist") {
			utils.WriteErrorJSON(w, http.StatusNotFound, err.Error())
		} else {
			utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to create directory: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"message":   "Directory created successfully",
		"directory": directory,
	}

	utils.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handlers) MoveFile(w http.ResponseWriter, r *http.Request) {
	var req models.MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.SourcePath == "" || req.DestinationPath == "" {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Source and destination paths are required")
		return
	}

	err := h.fileService.MoveFile(req.SourcePath, req.DestinationPath)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorJSON(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "already exists") {
			utils.WriteErrorJSON(w, http.StatusConflict, err.Error())
		} else {
			utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to move file: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"message":     "File moved successfully",
		"source":      req.SourcePath,
		"destination": req.DestinationPath,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handlers) RenameFile(w http.ResponseWriter, r *http.Request) {
	var req models.RenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Path == "" || req.NewName == "" {
		utils.WriteErrorJSON(w, http.StatusBadRequest, "Path and new name are required")
		return
	}

	err := h.fileService.RenameFile(req.Path, req.NewName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteErrorJSON(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "already exists") {
			utils.WriteErrorJSON(w, http.StatusConflict, err.Error())
		} else {
			utils.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to rename file: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"message":  "File renamed successfully",
		"old_path": req.Path,
		"new_name": req.NewName,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
