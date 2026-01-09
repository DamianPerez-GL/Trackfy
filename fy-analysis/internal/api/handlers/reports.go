package handlers

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/trackfy/fy-analysis/internal/urlengine"
)

// ReportsHandler maneja las solicitudes de reportes de usuarios
type ReportsHandler struct {
	engine *urlengine.Engine
}

// NewReportsHandler crea un nuevo handler de reportes
func NewReportsHandler(engine *urlengine.Engine) *ReportsHandler {
	return &ReportsHandler{
		engine: engine,
	}
}

// ReportURLRequest estructura de la petición de reporte
type ReportURLRequest struct {
	URL         string `json:"url"`
	UserID      string `json:"user_id"`
	ThreatType  string `json:"threat_type"`  // phishing, malware, scam, spam, other
	Description string `json:"description"`  // Descripción opcional
	Context     string `json:"context"`      // chat, manual, browser_extension
}

// ReportURL maneja POST /api/v1/report
func (h *ReportsHandler) ReportURL(w http.ResponseWriter, r *http.Request) {
	var req ReportURLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Error al parsear el JSON")
		return
	}

	// Validaciones
	if req.URL == "" {
		respondWithError(w, http.StatusBadRequest, "MISSING_URL", "El campo 'url' es requerido")
		return
	}

	if req.UserID == "" {
		respondWithError(w, http.StatusBadRequest, "MISSING_USER_ID", "El campo 'user_id' es requerido")
		return
	}

	// Validar threat_type
	validTypes := map[string]bool{
		"phishing": true, "malware": true, "scam": true,
		"spam": true, "vishing": true, "smishing": true, "other": true,
	}
	if req.ThreatType == "" {
		req.ThreatType = "other"
	} else if !validTypes[req.ThreatType] {
		respondWithError(w, http.StatusBadRequest, "INVALID_THREAT_TYPE",
			"Tipo de amenaza inválido. Usar: phishing, malware, scam, spam, vishing, smishing, other")
		return
	}

	// Context por defecto
	if req.Context == "" {
		req.Context = "manual"
	}

	// Obtener IP y User-Agent
	userIP := r.Header.Get("X-Forwarded-For")
	if userIP == "" {
		userIP = r.Header.Get("X-Real-IP")
	}
	if userIP == "" {
		userIP = r.RemoteAddr
	}
	// Quitar el puerto si existe (RemoteAddr viene como "ip:port")
	if host, _, err := net.SplitHostPort(userIP); err == nil {
		userIP = host
	}
	userAgent := r.Header.Get("User-Agent")

	// Llamar al engine
	engineReq := &urlengine.ReportURLRequest{
		URL:           req.URL,
		UserID:        req.UserID,
		ThreatType:    req.ThreatType,
		Description:   req.Description,
		ReportContext: req.Context,
		UserIP:        userIP,
		UserAgent:     userAgent,
	}

	result := h.engine.ReportURL(r.Context(), engineReq)

	if !result.Success {
		respondWithJSON(w, http.StatusOK, result) // 200 pero con success=false
		return
	}

	respondWithJSON(w, http.StatusCreated, result)
}

// GetReportsStats maneja GET /api/v1/reports/stats
func (h *ReportsHandler) GetReportsStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.engine.GetUserReportsStats(r.Context())
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}
