package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/trackfy/fy-analysis/internal/api"
	"github.com/trackfy/fy-analysis/internal/config"
	"github.com/trackfy/fy-analysis/internal/urlengine"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Configurar logger
	setupLogger(cfg)

	log.Info().
		Str("port", cfg.Port).
		Str("environment", cfg.Environment).
		Msg("Starting Fy-Analysis Service")

	// Inicializar URL Engine
	urlEngine := initURLEngine(cfg)

	// Crear router con URL Engine
	routerConfig := &api.RouterConfig{
		URLEngine: urlEngine,
	}
	router := api.NewRouterWithConfig(routerConfig)

	// Configurar servidor
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Iniciar servidor en goroutine
	go func() {
		log.Info().Msgf("Server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown con timeout de 30 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Detener URL Engine
	if urlEngine != nil {
		urlEngine.Stop()
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped gracefully")
}

func setupLogger(cfg *config.Config) {
	// Configurar output legible para desarrollo
	if cfg.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Configurar nivel de log
	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// initURLEngine inicializa el motor de verificación de URLs
func initURLEngine(cfg *config.Config) *urlengine.Engine {
	log.Info().Msg("Initializing URL Engine...")

	engineConfig := &urlengine.EngineConfig{
		CheckTimeout:      3 * time.Second,
		URLhausDBPath:     cfg.URLhausDBPath,
		PhishTankDBPath:   cfg.PhishTankDBPath,
		GoogleWebRiskKey:  cfg.GoogleWebRiskKey,
		URLScanKey:        cfg.URLScanKey,
		PhishTankKey:      cfg.PhishTankKey,
		EnableDBSync:      cfg.EnableDBSync,
		DatabaseURL:       cfg.DatabaseURL,
		EnableLocalDB:     cfg.EnableLocalDB,
		EnableUserReports: cfg.EnableUserReports,
	}

	engine := urlengine.NewEngine(engineConfig)

	// Iniciar sincronización de DBs en background
	ctx := context.Background()
	engine.Start(ctx)

	log.Info().
		Bool("db_sync_enabled", cfg.EnableDBSync).
		Bool("google_webrisk", cfg.GoogleWebRiskKey != "").
		Bool("urlscan", cfg.URLScanKey != "").
		Msg("URL Engine initialized and started")

	return engine
}
