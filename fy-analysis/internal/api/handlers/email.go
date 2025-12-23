package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/trackfy/fy-analysis/internal/analyzer/email"
	"github.com/trackfy/fy-analysis/internal/models"
)

type EmailHandler struct {
	analyzer *email.Analyzer
}

func NewEmailHandler() *EmailHandler {
	return &EmailHandler{
		analyzer: email.NewAnalyzer(),
	}
}

// AnalyzeEmail maneja las solicitudes de análisis de email
func (h *EmailHandler) AnalyzeEmail(w http.ResponseWriter, r *http.Request) {
	var req models.EmailAnalysisRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Error al parsear el JSON")
		return
	}

	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "MISSING_EMAIL", "El campo 'email' es requerido")
		return
	}

	result := h.analyzer.Analyze(req.Email, req.Context)
	respondWithJSON(w, http.StatusOK, result)
}
