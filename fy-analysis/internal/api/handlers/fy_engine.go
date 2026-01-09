package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/trackfy/fy-analysis/internal/checkers"
	"github.com/trackfy/fy-analysis/internal/urlengine"
)

// FyEngineHandler maneja las solicitudes de fy-engine
// Formato compatible con el servicio Python de fy-engine
type FyEngineHandler struct {
	engine *urlengine.Engine
}

// NewFyEngineHandler crea un nuevo handler para fy-engine
func NewFyEngineHandler(engine *urlengine.Engine) *FyEngineHandler {
	return &FyEngineHandler{
		engine: engine,
	}
}

// FyEngineResponse formato de respuesta esperado por fy-engine
type FyEngineResponse struct {
	Type      string   `json:"type"`
	Content   string   `json:"content"`
	RiskScore int      `json:"risk_score"`
	Verdict   string   `json:"verdict"`
	Reasons   []string `json:"reasons"`
	// Campos adicionales para trazabilidad
	FoundInDB bool   `json:"found_in_db"` // Si se encontró en la DB local
	Source    string `json:"source"`      // Fuente principal que lo detectó
	LatencyMs int64  `json:"latency_ms"`  // Tiempo de análisis en ms
}

// URLRequest petición de análisis de URL
type URLRequest struct {
	URL string `json:"url"`
}

// EmailRequest petición de análisis de email
type EmailRequest struct {
	Email string `json:"email"`
}

// PhoneRequest petición de análisis de teléfono
type PhoneRequest struct {
	Phone string `json:"phone"`
}

// convertToFyEngineResponse convierte el resultado del engine al formato de fy-engine
func convertToFyEngineResponse(result *urlengine.AnalysisResponse, inputType string, content string) *FyEngineResponse {
	// Determinar verdict basado en risk_level
	verdict := "safe"
	switch result.RiskLevel {
	case "warning":
		verdict = "suspicious"
	case "danger", "critical":
		verdict = "dangerous"
	}

	// Extraer info de trazabilidad de las fuentes
	foundInDB := false
	mainSource := ""
	for _, src := range result.Sources {
		if src.Found {
			if mainSource == "" {
				mainSource = src.Name
			}
			// localdb es la DB de amenazas
			if src.Name == "localdb" {
				foundInDB = true
				mainSource = "localdb"
				break
			}
		}
	}

	return &FyEngineResponse{
		Type:      inputType,
		Content:   content,
		RiskScore: result.RiskScore,
		Verdict:   verdict,
		Reasons:   result.Reasons,
		FoundInDB: foundInDB,
		Source:    mainSource,
		LatencyMs: result.ResponseTimeMs,
	}
}

// AnalyzeURL maneja POST /analyze/url
func (h *FyEngineHandler) AnalyzeURL(w http.ResponseWriter, r *http.Request) {
	var req URLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "url", "", "Error al parsear el JSON")
		return
	}

	if req.URL == "" {
		h.respondError(w, "url", "", "El campo 'url' es requerido")
		return
	}

	// Ejecutar análisis
	engineReq := &urlengine.AnalysisRequest{
		Input: req.URL,
		Type:  checkers.InputTypeURL,
	}

	result := h.engine.Analyze(r.Context(), engineReq)
	response := convertToFyEngineResponse(result, "url", req.URL)

	respondWithJSON(w, http.StatusOK, response)
}

// AnalyzeEmail maneja POST /analyze/email
func (h *FyEngineHandler) AnalyzeEmail(w http.ResponseWriter, r *http.Request) {
	var req EmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "email", "", "Error al parsear el JSON")
		return
	}

	if req.Email == "" {
		h.respondError(w, "email", "", "El campo 'email' es requerido")
		return
	}

	// Ejecutar análisis
	engineReq := &urlengine.AnalysisRequest{
		Input: req.Email,
		Type:  checkers.InputTypeEmail,
	}

	result := h.engine.Analyze(r.Context(), engineReq)
	response := convertToFyEngineResponse(result, "email", req.Email)

	respondWithJSON(w, http.StatusOK, response)
}

// AnalyzePhone maneja POST /analyze/phone
func (h *FyEngineHandler) AnalyzePhone(w http.ResponseWriter, r *http.Request) {
	var req PhoneRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "phone", "", "Error al parsear el JSON")
		return
	}

	if req.Phone == "" {
		h.respondError(w, "phone", "", "El campo 'phone' es requerido")
		return
	}

	// Ejecutar análisis
	engineReq := &urlengine.AnalysisRequest{
		Input: req.Phone,
		Type:  checkers.InputTypePhone,
	}

	result := h.engine.Analyze(r.Context(), engineReq)
	response := convertToFyEngineResponse(result, "phone", req.Phone)

	respondWithJSON(w, http.StatusOK, response)
}

// respondError responde con un error en formato fy-engine
func (h *FyEngineHandler) respondError(w http.ResponseWriter, inputType, content, reason string) {
	response := &FyEngineResponse{
		Type:      inputType,
		Content:   content,
		RiskScore: 50,
		Verdict:   "unknown",
		Reasons:   []string{reason},
	}
	respondWithJSON(w, http.StatusOK, response)
}