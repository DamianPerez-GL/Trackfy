package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/trackfy/fy-analysis/internal/analyzer/email"
	"github.com/trackfy/fy-analysis/internal/analyzer/phone"
	urlanalyzer "github.com/trackfy/fy-analysis/internal/analyzer/url"
	"github.com/trackfy/fy-analysis/internal/models"
)

type BatchHandler struct {
	emailAnalyzer *email.Analyzer
	urlAnalyzer   *urlanalyzer.Analyzer
	phoneAnalyzer *phone.Analyzer
}

func NewBatchHandler() *BatchHandler {
	return &BatchHandler{
		emailAnalyzer: email.NewAnalyzer(),
		urlAnalyzer:   urlanalyzer.NewAnalyzer(),
		phoneAnalyzer: phone.NewAnalyzer(),
	}
}

// AnalyzeBatch maneja las solicitudes de análisis en lote
func (h *BatchHandler) AnalyzeBatch(w http.ResponseWriter, r *http.Request) {
	var req models.BatchAnalysisRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Error al parsear el JSON")
		return
	}

	// Validar que hay algo que analizar
	if len(req.Emails) == 0 && len(req.URLs) == 0 && len(req.Phones) == 0 {
		respondWithError(w, http.StatusBadRequest, "EMPTY_REQUEST", "Debe proporcionar al menos un email, URL o teléfono para analizar")
		return
	}

	// Limitar el tamaño del lote
	maxBatchSize := 100
	totalItems := len(req.Emails) + len(req.URLs) + len(req.Phones)
	if totalItems > maxBatchSize {
		respondWithError(w, http.StatusBadRequest, "BATCH_TOO_LARGE", "El lote no puede exceder 100 elementos")
		return
	}

	response := models.BatchAnalysisResponse{
		Emails:  make([]models.EmailAnalysisResponse, 0, len(req.Emails)),
		URLs:    make([]models.URLAnalysisResponse, 0, len(req.URLs)),
		Phones:  make([]models.PhoneAnalysisResponse, 0, len(req.Phones)),
		Summary: models.BatchSummary{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Analizar emails en paralelo
	for _, emailAddr := range req.Emails {
		wg.Add(1)
		go func(e string) {
			defer wg.Done()
			result := h.emailAnalyzer.Analyze(e, "")
			mu.Lock()
			response.Emails = append(response.Emails, result)
			updateSummary(&response.Summary, result.Analysis)
			mu.Unlock()
		}(emailAddr)
	}

	// Analizar URLs en paralelo
	for _, url := range req.URLs {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			result := h.urlAnalyzer.Analyze(u, "")
			mu.Lock()
			response.URLs = append(response.URLs, result)
			updateSummary(&response.Summary, result.Analysis)
			mu.Unlock()
		}(url)
	}

	// Analizar teléfonos en paralelo
	for _, phoneNum := range req.Phones {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			result := h.phoneAnalyzer.Analyze(p, "", "")
			mu.Lock()
			response.Phones = append(response.Phones, result)
			updateSummary(&response.Summary, result.Analysis)
			mu.Unlock()
		}(phoneNum)
	}

	wg.Wait()

	response.Summary.TotalAnalyzed = len(response.Emails) + len(response.URLs) + len(response.Phones)

	respondWithJSON(w, http.StatusOK, response)
}

func updateSummary(summary *models.BatchSummary, analysis models.AnalysisResult) {
	if analysis.IsMalicious {
		summary.MaliciousCount++
	} else if analysis.ThreatLevel == models.ThreatLevelMedium || analysis.ThreatLevel == models.ThreatLevelLow {
		summary.SuspiciousCount++
	} else {
		summary.SafeCount++
	}
}
