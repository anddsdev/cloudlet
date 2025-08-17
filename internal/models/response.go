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
