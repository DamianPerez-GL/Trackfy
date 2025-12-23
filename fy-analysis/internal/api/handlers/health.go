package handlers

import (
	"net/http"
	"time"

	"github.com/trackfy/fy-analysis/internal/models"
)

const Version = "1.0.0"

// HealthCheck maneja las solicitudes de health check
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "healthy",
		Version:   Version,
		Timestamp: time.Now().UTC(),
	}
	respondWithJSON(w, http.StatusOK, response)
}
