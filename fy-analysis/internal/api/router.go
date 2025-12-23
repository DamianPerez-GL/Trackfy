package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"time"

	"github.com/trackfy/fy-analysis/internal/api/handlers"
	customMiddleware "github.com/trackfy/fy-analysis/internal/api/middleware"
)

// NewRouter crea y configura el router de la API
func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middleware global
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // En producción, especificar dominios
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting: 100 requests por minuto por IP
	r.Use(httprate.LimitByIP(100, time.Minute))

	// Health check
	r.Get("/health", handlers.HealthCheck)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/analyze", func(r chi.Router) {
			emailHandler := handlers.NewEmailHandler()
			urlHandler := handlers.NewURLHandler()
			phoneHandler := handlers.NewPhoneHandler()
			batchHandler := handlers.NewBatchHandler()

			r.Post("/email", emailHandler.AnalyzeEmail)
			r.Post("/url", urlHandler.AnalyzeURL)
			r.Post("/phone", phoneHandler.AnalyzePhone)
			r.Post("/batch", batchHandler.AnalyzeBatch)
		})
	})

	return r
}
