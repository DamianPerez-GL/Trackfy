package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/trackfy/fy-analysis/internal/analyzer/phone"
	"github.com/trackfy/fy-analysis/internal/models"
)

type PhoneHandler struct {
	analyzer *phone.Analyzer
}

func NewPhoneHandler() *PhoneHandler {
	return &PhoneHandler{
		analyzer: phone.NewAnalyzer(),
	}
}

// AnalyzePhone maneja las solicitudes de análisis de teléfono
func (h *PhoneHandler) AnalyzePhone(w http.ResponseWriter, r *http.Request) {
	var req models.PhoneAnalysisRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Error al parsear el JSON")
		return
	}

	if req.Phone == "" {
		respondWithError(w, http.StatusBadRequest, "MISSING_PHONE", "El campo 'phone' es requerido")
		return
	}

	result := h.analyzer.Analyze(req.Phone, req.CountryCode, req.Context)
	respondWithJSON(w, http.StatusOK, result)
}
