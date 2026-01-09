package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/trackfy/api-gateway/internal/auth"
	"github.com/trackfy/api-gateway/internal/db"
	"github.com/trackfy/api-gateway/internal/middleware"
	"github.com/trackfy/api-gateway/internal/services"
)

func NewRouter(postgres *db.PostgresDB, redis *db.RedisDB, jwtManager *auth.JWTManager, fyEngine *services.FyEngineClient, fyAnalysis *services.FyAnalysisClient) http.Handler {
	r := chi.NewRouter()

	// Middleware global
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Device-ID"},
		ExposedHeaders:   []string{"X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Crear handler y middlewares
	h := NewHandler(postgres, redis, jwtManager, fyEngine)
	h.SetFyAnalysisClient(fyAnalysis)
	authMw := middleware.NewAuthMiddleware(jwtManager, redis)
	rateLimiter := middleware.NewRateLimiter(redis)

	// Health check (sin auth)
	r.Get("/health", h.Health)

	// Rutas públicas de autenticación
	r.Route("/auth", func(r chi.Router) {
		// Rate limit más estricto para auth
		r.Use(rateLimiter.Limit(10, time.Minute))

		r.Post("/register", h.Register)
		r.Post("/send-code", h.SendVerificationCode)
		r.Post("/verify", h.VerifyCode)
	})

	// Rutas protegidas
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Use(rateLimiter.LimitByUser(100, time.Minute))

		// Usuario
		r.Route("/me", func(r chi.Router) {
			r.Get("/", h.GetMe)
			r.Get("/sessions", h.GetMySessions)
			r.Get("/stats", h.GetMyStats)
			r.Post("/logout", h.Logout)
			r.Post("/logout-all", h.LogoutAll)
		})

		// Conversaciones
		r.Route("/conversations", func(r chi.Router) {
			r.Get("/", h.GetConversations)
			r.Post("/", h.CreateConversation)
			r.Get("/{id}", h.GetConversation)
			r.Get("/{id}/messages", h.GetConversationMessages)
		})

		// Chat con Fy
		r.Post("/chat", h.Chat)

		// Reportes de URLs sospechosas
		r.Post("/report", h.ReportURL)
	})

	return r
}
