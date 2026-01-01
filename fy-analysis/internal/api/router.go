package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/trackfy/fy-analysis/internal/api/handlers"
	customMiddleware "github.com/trackfy/fy-analysis/internal/api/middleware"
	"github.com/trackfy/fy-analysis/internal/urlengine"
)

// RouterConfig configuración para el router
type RouterConfig struct {
	URLEngine *urlengine.Engine
}

// NewRouter crea y configura el router de la API (versión legacy)
func NewRouter() *chi.Mux {
	return NewRouterWithConfig(nil)
}

// NewRouterWithConfig crea el router con configuración del URL engine
func NewRouterWithConfig(config *RouterConfig) *chi.Mux {
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
		// Endpoints de análisis básico (legacy)
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

		// URL Engine - Verificación avanzada con múltiples fuentes
		if config != nil && config.URLEngine != nil {
			urlEngineHandler := handlers.NewURLEngineHandler(config.URLEngine)
			reportsHandler := handlers.NewReportsHandler(config.URLEngine)

			// Endpoint unificado de análisis (recomendado)
			r.Post("/analyze", urlEngineHandler.Analyze)

			// Endpoints legacy para compatibilidad
			r.Route("/urlengine", func(r chi.Router) {
				r.Post("/check", urlEngineHandler.CheckURL)
				r.Get("/status", urlEngineHandler.GetStatus)
				r.Post("/sync", urlEngineHandler.SyncDB)
			})

			// Endpoints de reportes de usuarios
			r.Route("/reports", func(r chi.Router) {
				r.Post("/", reportsHandler.ReportURL)           // POST /api/v1/reports
				r.Get("/stats", reportsHandler.GetReportsStats) // GET /api/v1/reports/stats
			})
		}
	})

	// Endpoints para Fy-Engine (formato compatible con Python service)
	// Estos endpoints están en la raíz para compatibilidad con fy-engine
	if config != nil && config.URLEngine != nil {
		fyHandler := handlers.NewFyEngineHandler(config.URLEngine)
		r.Route("/analyze", func(r chi.Router) {
			r.Post("/url", fyHandler.AnalyzeURL)
			r.Post("/email", fyHandler.AnalyzeEmail)
			r.Post("/phone", fyHandler.AnalyzePhone)
		})
	}

	return r
}
