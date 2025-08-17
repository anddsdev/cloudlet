package utils

import (
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := WriteJSONData(w, data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func WriteErrorJSON(w http.ResponseWriter, status int, message string) {
	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  status,
	}
	WriteJSON(w, status, response)
}
