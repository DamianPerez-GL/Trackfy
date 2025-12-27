package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/trackfy/api-gateway/internal/auth"
	"github.com/trackfy/api-gateway/internal/db"
)

type contextKey string

const (
	ContextKeyUserID    contextKey = "user_id"
	ContextKeySessionID contextKey = "session_id"
	ContextKeyClaims    contextKey = "claims"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
	redis      *db.RedisDB
}

func NewAuthMiddleware(jwtManager *auth.JWTManager, redis *db.RedisDB) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		redis:      redis,
	}
}

// Authenticate middleware que requiere autenticación
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extraer token del header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing_token", "Authorization header required")
			return
		}

		// Verificar formato Bearer
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondError(w, http.StatusUnauthorized, "invalid_format", "Invalid authorization format")
			return
		}

		tokenString := parts[1]

		// Validar JWT
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			log.Debug().Err(err).Msg("[Auth] Token validation failed")
			respondError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
			return
		}

		// Verificar que sea un access token
		if claims.TokenType != "access" {
			respondError(w, http.StatusUnauthorized, "wrong_token_type", "Access token required")
			return
		}

		// Verificar sesión en Redis
		tokenHash := auth.HashToken(tokenString)
		session, err := m.redis.GetSession(r.Context(), tokenHash)
		if err != nil {
			log.Error().Err(err).Msg("[Auth] Redis error")
			respondError(w, http.StatusInternalServerError, "session_error", "Session verification failed")
			return
		}

		if session == nil {
			respondError(w, http.StatusUnauthorized, "session_expired", "Session not found or expired")
			return
		}

		// Actualizar actividad de sesión
		_ = m.redis.UpdateSessionActivity(r.Context(), tokenHash)

		// Añadir claims al contexto
		ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeySessionID, claims.SessionID)
		ctx = context.WithValue(ctx, ContextKeyClaims, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth middleware que permite acceso sin autenticación pero añade usuario si existe
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := m.jwtManager.ValidateToken(parts[1])
		if err == nil && claims.TokenType == "access" {
			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeySessionID, claims.SessionID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserID extrae el UserID del contexto
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return userID, ok
}

// GetSessionID extrae el SessionID del contexto
func GetSessionID(ctx context.Context) (uuid.UUID, bool) {
	sessionID, ok := ctx.Value(ContextKeySessionID).(uuid.UUID)
	return sessionID, ok
}

// GetClaims extrae los claims del contexto
func GetClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(*auth.Claims)
	return claims, ok
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + code + `","message":"` + message + `"}`))
}
