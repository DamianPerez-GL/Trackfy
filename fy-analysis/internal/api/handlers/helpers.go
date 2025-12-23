package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/trackfy/fy-analysis/internal/models"
)

// respondWithJSON envía una respuesta JSON
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

// respondWithError envía una respuesta de error
func respondWithError(w http.ResponseWriter, statusCode int, code string, message string) {
	response := models.ErrorResponse{
		Error: message,
		Code:  code,
	}
	respondWithJSON(w, statusCode, response)
}
