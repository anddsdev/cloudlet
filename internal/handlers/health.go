package handlers

import (
	"net/http"

	"github.com/anddsdev/cloudlet/internal/utils"
)

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ok",
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
