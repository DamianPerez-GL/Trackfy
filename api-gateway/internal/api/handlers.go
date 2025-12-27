package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/trackfy/api-gateway/internal/auth"
	"github.com/trackfy/api-gateway/internal/db"
	"github.com/trackfy/api-gateway/internal/middleware"
	"github.com/trackfy/api-gateway/internal/models"
	"github.com/trackfy/api-gateway/internal/services"
)

type Handler struct {
	postgres   *db.PostgresDB
	redis      *db.RedisDB
	jwtManager *auth.JWTManager
	fyEngine   *services.FyEngineClient
}

func NewHandler(postgres *db.PostgresDB, redis *db.RedisDB, jwtManager *auth.JWTManager, fyEngine *services.FyEngineClient) *Handler {
	return &Handler{
		postgres:   postgres,
		redis:      redis,
		jwtManager: jwtManager,
		fyEngine:   fyEngine,
	}
}

// ==================== AUTH ====================

type RegisterRequest struct {
	Phone     string `json:"phone"`
	Nombre    string `json:"nombre"`
	Apellidos string `json:"apellidos"`
}

type RegisterResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

// Register registra un nuevo usuario
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	// Validar campos
	phone := normalizePhone(req.Phone)
	if phone == "" {
		respondError(w, http.StatusBadRequest, "invalid_phone", "Invalid phone number")
		return
	}

	if strings.TrimSpace(req.Nombre) == "" {
		respondError(w, http.StatusBadRequest, "invalid_nombre", "Nombre is required")
		return
	}

	if strings.TrimSpace(req.Apellidos) == "" {
		respondError(w, http.StatusBadRequest, "invalid_apellidos", "Apellidos is required")
		return
	}

	// Verificar si ya existe
	exists, err := h.postgres.UserExists(r.Context(), phone)
	if err != nil {
		log.Error().Err(err).Msg("[Register] Database error")
		respondError(w, http.StatusInternalServerError, "db_error", "Database error")
		return
	}

	if exists {
		respondError(w, http.StatusConflict, "user_exists", "Phone already registered")
		return
	}

	// Crear usuario
	user, err := h.postgres.CreateUser(r.Context(), phone, strings.TrimSpace(req.Nombre), strings.TrimSpace(req.Apellidos))
	if err != nil {
		log.Error().Err(err).Msg("[Register] Failed to create user")
		respondError(w, http.StatusInternalServerError, "create_error", "Failed to create user")
		return
	}

	log.Info().Str("user_id", user.ID.String()).Str("phone", maskPhone(phone)).Msg("[Register] User created")

	respondJSON(w, http.StatusCreated, RegisterResponse{
		Message: "Usuario registrado. Verifica tu teléfono para activar la cuenta.",
		UserID:  user.ID.String(),
	})
}

type SendCodeRequest struct {
	Phone string `json:"phone"`
}

type SendCodeResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in"` // segundos
	// En desarrollo incluimos el código para pruebas
	Code string `json:"code,omitempty"`
}

// SendVerificationCode envía código SMS (simulado por ahora)
func (h *Handler) SendVerificationCode(w http.ResponseWriter, r *http.Request) {
	var req SendCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	phone := normalizePhone(req.Phone)
	if phone == "" {
		respondError(w, http.StatusBadRequest, "invalid_phone", "Invalid phone number")
		return
	}

	// Verificar que el usuario existe
	exists, err := h.postgres.UserExists(r.Context(), phone)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db_error", "Database error")
		return
	}

	if !exists {
		respondError(w, http.StatusNotFound, "user_not_found", "User not found")
		return
	}

	// Generar código
	code, err := h.postgres.GenerateVerificationCode(r.Context(), phone)
	if err != nil {
		log.Error().Err(err).Msg("[SendCode] Failed to generate code")
		respondError(w, http.StatusInternalServerError, "code_error", "Failed to generate code")
		return
	}

	log.Info().Str("phone", maskPhone(phone)).Str("code", code).Msg("[SendCode] Verification code generated")

	// TODO: Integrar servicio SMS real
	// Por ahora devolvemos el código en la respuesta para desarrollo
	respondJSON(w, http.StatusOK, SendCodeResponse{
		Message:   "Código enviado",
		ExpiresIn: 300, // 5 minutos
		Code:      code, // Solo en desarrollo
	})
}

type VerifyCodeRequest struct {
	Phone      string `json:"phone"`
	Code       string `json:"code"`
	DeviceID   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
	DeviceType string `json:"device_type,omitempty"` // ios, android
	AppVersion string `json:"app_version,omitempty"`
}

type VerifyCodeResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresAt    time.Time       `json:"expires_at"`
	TokenType    string          `json:"token_type"`
	User         models.UserPublic `json:"user"`
}

// VerifyCode verifica el código SMS y devuelve tokens
func (h *Handler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	var req VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	phone := normalizePhone(req.Phone)
	if phone == "" {
		respondError(w, http.StatusBadRequest, "invalid_phone", "Invalid phone number")
		return
	}

	if len(req.Code) != 6 {
		respondError(w, http.StatusBadRequest, "invalid_code", "Code must be 6 digits")
		return
	}

	// Verificar código
	valid, err := h.postgres.VerifyCode(r.Context(), phone, req.Code)
	if err != nil {
		log.Error().Err(err).Msg("[VerifyCode] Database error")
		respondError(w, http.StatusInternalServerError, "db_error", "Verification failed")
		return
	}

	if !valid {
		respondError(w, http.StatusUnauthorized, "invalid_code", "Invalid or expired code")
		return
	}

	// Obtener usuario
	user, err := h.postgres.GetUserByPhone(r.Context(), phone)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "user_error", "Failed to get user")
		return
	}

	// Crear sesión
	sessionID := uuid.New()
	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, sessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "token_error", "Failed to generate tokens")
		return
	}

	// Guardar sesión en PostgreSQL
	session := &models.Session{
		ID:         sessionID,
		UserID:     user.ID,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
		DeviceType: req.DeviceType,
		AppVersion: req.AppVersion,
		IPAddress:  getClientIP(r),
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}
	if err := h.postgres.CreateSession(r.Context(), session, auth.HashTokenBytes(tokens.AccessToken)); err != nil {
		log.Error().Err(err).Msg("[VerifyCode] Failed to create session in DB")
	}

	// Guardar sesión en Redis
	sessionData := &db.SessionData{
		SessionID:  sessionID,
		UserID:     user.ID,
		DeviceID:   req.DeviceID,
		DeviceType: req.DeviceType,
		CreatedAt:  time.Now(),
		ExpiresAt:  tokens.ExpiresAt,
	}
	if err := h.redis.StoreSession(r.Context(), auth.HashToken(tokens.AccessToken), sessionData, 15*time.Minute); err != nil {
		log.Error().Err(err).Msg("[VerifyCode] Failed to store session in Redis")
	}

	// Actualizar último login
	_ = h.postgres.UpdateUserLastLogin(r.Context(), user.ID)

	log.Info().
		Str("user_id", user.ID.String()).
		Str("session_id", sessionID.String()).
		Msg("[VerifyCode] Login successful")

	respondJSON(w, http.StatusOK, VerifyCodeResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		TokenType:    tokens.TokenType,
		User:         user.ToPublic(),
	})
}

// Logout invalida la sesión actual
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	sessionID, _ := middleware.GetSessionID(r.Context())

	// Obtener token del header para invalidar en Redis
	authHeader := r.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 {
		tokenHash := auth.HashToken(parts[1])
		_ = h.redis.DeleteSession(r.Context(), tokenHash, userID)
	}

	// Invalidar en PostgreSQL
	_ = h.postgres.InvalidateSession(r.Context(), sessionID, "logout")

	log.Info().Str("user_id", userID.String()).Msg("[Logout] Session invalidated")

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// LogoutAll invalida todas las sesiones del usuario
func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	// Invalidar todas las sesiones en Redis y PostgreSQL
	_ = h.redis.DeleteAllUserSessions(r.Context(), userID)
	_ = h.postgres.InvalidateAllUserSessions(r.Context(), userID)

	log.Info().Str("user_id", userID.String()).Msg("[LogoutAll] All sessions invalidated")

	respondJSON(w, http.StatusOK, map[string]string{"message": "All sessions logged out"})
}

// ==================== USER ====================

// GetMe devuelve el perfil del usuario actual
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	// Intentar cache
	if user, _ := h.redis.GetCachedUser(r.Context(), userID); user != nil {
		respondJSON(w, http.StatusOK, user.ToPublic())
		return
	}

	user, err := h.postgres.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user_not_found", "User not found")
		return
	}

	// Cachear
	_ = h.redis.CacheUser(r.Context(), user)

	respondJSON(w, http.StatusOK, user.ToPublic())
}

// GetMySessions devuelve las sesiones activas del usuario
func (h *Handler) GetMySessions(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	sessions, err := h.postgres.GetUserSessions(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db_error", "Failed to get sessions")
		return
	}

	respondJSON(w, http.StatusOK, sessions)
}

// GetMyStats devuelve las estadísticas del usuario
func (h *Handler) GetMyStats(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	stats, err := h.postgres.GetUserStats(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db_error", "Failed to get stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// ==================== CONVERSATIONS ====================

type CreateConversationRequest struct {
	Title string `json:"title,omitempty"`
}

// CreateConversation crea una nueva conversación
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req CreateConversationRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	title := req.Title
	if title == "" {
		title = "Conversación " + time.Now().Format("02/01/2006 15:04")
	}

	conv, err := h.postgres.CreateConversation(r.Context(), userID, title)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db_error", "Failed to create conversation")
		return
	}

	respondJSON(w, http.StatusCreated, conv)
}

// GetConversations lista las conversaciones del usuario
func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	conversations, err := h.postgres.GetUserConversations(r.Context(), userID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db_error", "Failed to get conversations")
		return
	}

	respondJSON(w, http.StatusOK, conversations)
}

// GetConversation obtiene una conversación específica
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	convIDStr := chi.URLParam(r, "id")

	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "Invalid conversation ID")
		return
	}

	conv, err := h.postgres.GetConversation(r.Context(), convID, userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Conversation not found")
		return
	}

	respondJSON(w, http.StatusOK, conv)
}

// GetConversationMessages obtiene los mensajes de una conversación
func (h *Handler) GetConversationMessages(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	convIDStr := chi.URLParam(r, "id")

	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "Invalid conversation ID")
		return
	}

	// Verificar propiedad
	_, err = h.postgres.GetConversation(r.Context(), convID, userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Conversation not found")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	messages, err := h.postgres.GetConversationMessages(r.Context(), convID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db_error", "Failed to get messages")
		return
	}

	respondJSON(w, http.StatusOK, messages)
}

// ==================== CHAT ====================

type ChatRequest struct {
	ConversationID string `json:"conversation_id,omitempty"`
	Message        string `json:"message"`
}

type ChatResponse struct {
	ConversationID string `json:"conversation_id"`
	Response       string `json:"response"`
	Mood           string `json:"mood"`
	Intent         string `json:"intent"`
}

// Chat envía un mensaje a Fy
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		respondError(w, http.StatusBadRequest, "empty_message", "Message is required")
		return
	}

	// Obtener o crear conversación
	var convID uuid.UUID
	if req.ConversationID != "" {
		var err error
		convID, err = uuid.Parse(req.ConversationID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid_conversation", "Invalid conversation ID")
			return
		}
		// Verificar propiedad
		if _, err := h.postgres.GetConversation(r.Context(), convID, userID); err != nil {
			respondError(w, http.StatusNotFound, "conversation_not_found", "Conversation not found")
			return
		}
	} else {
		// Crear nueva conversación
		conv, err := h.postgres.CreateConversation(r.Context(), userID, "")
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db_error", "Failed to create conversation")
			return
		}
		convID = conv.ID
	}

	// Obtener contexto de la conversación (últimos mensajes)
	var context []services.ContextMessage
	if memory, _ := h.redis.GetFyMemory(r.Context(), userID, convID); memory != nil {
		for _, msg := range memory.RecentMessages {
			context = append(context, services.ContextMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// Guardar mensaje del usuario
	userMsg := &models.Message{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           "user",
		Content:        req.Message,
		CreatedAt:      time.Now(),
	}
	_ = h.postgres.AddMessage(r.Context(), userMsg)

	// Enviar a Fy Engine
	fyResp, err := h.fyEngine.Chat(r.Context(), userID.String(), req.Message, context)
	if err != nil {
		log.Error().Err(err).Msg("[Chat] Fy Engine error")
		respondError(w, http.StatusServiceUnavailable, "fy_error", "Failed to process message")
		return
	}

	// Guardar respuesta de Fy
	fyMsg := &models.Message{
		ID:                uuid.New(),
		ConversationID:    convID,
		Role:              "assistant",
		Content:           fyResp.Response,
		Intent:            fyResp.Intent,
		Mood:              fyResp.Mood,
		AnalysisPerformed: fyResp.AnalysisPerformed,
		CreatedAt:         time.Now(),
	}
	_ = h.postgres.AddMessage(r.Context(), fyMsg)

	// Actualizar memoria corta de Fy
	recentMessages := append(context, services.ContextMessage{Role: "user", Content: req.Message})
	recentMessages = append(recentMessages, services.ContextMessage{Role: "assistant", Content: fyResp.Response})
	// Mantener solo los últimos 10 mensajes
	if len(recentMessages) > 10 {
		recentMessages = recentMessages[len(recentMessages)-10:]
	}
	var fyMemory []db.FyMemoryMessage
	for _, m := range recentMessages {
		fyMemory = append(fyMemory, db.FyMemoryMessage{Role: m.Role, Content: m.Content})
	}
	_ = h.redis.StoreFyMemory(r.Context(), userID, convID, fyMemory, fyResp.Intent, fyResp.Mood)

	// Actualizar estadísticas
	isThreat := fyResp.Mood == "danger" || fyResp.Mood == "warning"
	_ = h.postgres.UpdateUserStats(r.Context(), userID, fyResp.AnalysisPerformed, isThreat)

	respondJSON(w, http.StatusOK, ChatResponse{
		ConversationID: convID.String(),
		Response:       fyResp.Response,
		Mood:           fyResp.Mood,
		Intent:         fyResp.Intent,
	})
}

// ==================== HEALTH ====================

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	// Verificar dependencias
	fyHealth := h.fyEngine.Health(r.Context())

	status := "ok"
	if !fyHealth {
		status = "degraded"
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    status,
		"service":   "api-gateway",
		"fy_engine": fyHealth,
	})
}

// ==================== HELPERS ====================

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

var phoneRegex = regexp.MustCompile(`[^\d+]`)

func normalizePhone(phone string) string {
	// Limpiar todo excepto dígitos y +
	cleaned := phoneRegex.ReplaceAllString(phone, "")

	// Si empieza con 00, reemplazar por +
	if strings.HasPrefix(cleaned, "00") {
		cleaned = "+" + cleaned[2:]
	}

	// Si no tiene código y son 9 dígitos, añadir +34
	if !strings.HasPrefix(cleaned, "+") {
		if len(cleaned) == 9 && (cleaned[0] == '6' || cleaned[0] == '7' || cleaned[0] == '9') {
			cleaned = "+34" + cleaned
		} else {
			return "" // Inválido
		}
	}

	// Validar formato español
	if strings.HasPrefix(cleaned, "+34") && len(cleaned) == 12 {
		return cleaned
	}

	return ""
}

func maskPhone(phone string) string {
	if len(phone) > 6 {
		return phone[:4] + "***" + phone[len(phone)-3:]
	}
	return "***"
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	// Tomar solo la primera IP si hay múltiples
	if idx := strings.Index(ip, ","); idx != -1 {
		ip = ip[:idx]
	}
	return strings.TrimSpace(ip)
}
