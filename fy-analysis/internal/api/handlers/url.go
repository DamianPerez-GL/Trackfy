package handlers

import (
	"encoding/json"
	"net/http"

	urlanalyzer "github.com/trackfy/fy-analysis/internal/analyzer/url"
	"github.com/trackfy/fy-analysis/internal/models"
)

type URLHandler struct {
	analyzer *urlanalyzer.Analyzer
}

func NewURLHandler() *URLHandler {
	return &URLHandler{
		analyzer: urlanalyzer.NewAnalyzer(),
	}
}

// AnalyzeURL maneja las solicitudes de análisis de URL
func (h *URLHandler) AnalyzeURL(w http.ResponseWriter, r *http.Request) {
	var req models.URLAnalysisRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Error al parsear el JSON")
		return
	}

	if req.URL == "" {
		respondWithError(w, http.StatusBadRequest, "MISSING_URL", "El campo 'url' es requerido")
		return
	}

	result := h.analyzer.Analyze(req.URL, req.Context)
	respondWithJSON(w, http.StatusOK, result)
}
