package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/trackfy/fy-analysis/internal/checkers"
	"github.com/trackfy/fy-analysis/internal/urlengine"
)

// URLEngineHandler maneja las solicitudes del URL verification engine
type URLEngineHandler struct {
	engine *urlengine.Engine
}

// NewURLEngineHandler crea un nuevo handler para el URL engine
func NewURLEngineHandler(engine *urlengine.Engine) *URLEngineHandler {
	return &URLEngineHandler{
		engine: engine,
	}
}

// CheckURL maneja POST /api/v1/urlengine/check (legacy)
func (h *URLEngineHandler) CheckURL(w http.ResponseWriter, r *http.Request) {
	var req urlengine.URLCheckRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Error al parsear el JSON")
		return
	}

	if req.URL == "" {
		respondWithError(w, http.StatusBadRequest, "MISSING_URL", "El campo 'url' es requerido")
		return
	}

	// Ejecutar verificación
	result := h.engine.Check(r.Context(), req.URL)

	respondWithJSON(w, http.StatusOK, result)
}

// AnalyzeRequest estructura de la petición de análisis unificado
type AnalyzeRequest struct {
	Input   string `json:"input"`
	Type    string `json:"type"` // url, email, phone
	Context *struct {
		ClaimedSender string `json:"claimed_sender,omitempty"`
		MessageType   string `json:"message_type,omitempty"`
		OriginalText  string `json:"original_text,omitempty"`
	} `json:"context,omitempty"`
}

// Analyze maneja POST /api/v1/analyze - Endpoint unificado
func (h *URLEngineHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_JSON", "Error al parsear el JSON")
		return
	}

	if req.Input == "" {
		respondWithError(w, http.StatusBadRequest, "MISSING_INPUT", "El campo 'input' es requerido")
		return
	}

	// Validar tipo
	var inputType checkers.InputType
	switch req.Type {
	case "url", "":
		inputType = checkers.InputTypeURL
	case "email":
		inputType = checkers.InputTypeEmail
	case "phone":
		inputType = checkers.InputTypePhone
	default:
		respondWithError(w, http.StatusBadRequest, "INVALID_TYPE", "Tipo inválido. Usar: url, email, phone")
		return
	}

	// Construir request del engine
	engineReq := &urlengine.AnalysisRequest{
		Input: req.Input,
		Type:  inputType,
	}

	// Añadir contexto si existe
	if req.Context != nil {
		engineReq.Context = &checkers.AnalysisContext{
			ClaimedSender: req.Context.ClaimedSender,
			MessageType:   req.Context.MessageType,
			OriginalText:  req.Context.OriginalText,
		}
	}

	// Ejecutar análisis
	result := h.engine.Analyze(r.Context(), engineReq)

	respondWithJSON(w, http.StatusOK, result)
}

// GetStatus maneja GET /api/v1/urlengine/status
func (h *URLEngineHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.engine.GetStatus()
	respondWithJSON(w, http.StatusOK, status)
}

// SyncDB maneja POST /api/v1/urlengine/sync
func (h *URLEngineHandler) SyncDB(w http.ResponseWriter, r *http.Request) {
	dbName := r.URL.Query().Get("db")
	if dbName == "" {
		dbName = "all"
	}

	if err := h.engine.ForceDBSync(r.Context(), dbName); err != nil {
		respondWithError(w, http.StatusInternalServerError, "SYNC_FAILED", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"status": "sync_started",
		"db":     dbName,
	})
}
