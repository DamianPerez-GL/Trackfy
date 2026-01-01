package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/trackfy/fy-dbsync/internal/config"
	"github.com/trackfy/fy-dbsync/internal/syncer"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Configurar logger
	setupLogger(cfg)

	log.Info().
		Str("environment", cfg.Environment).
		Msg("Starting Fy-DBSync Service")

	// Verificar DATABASE_URL
	if cfg.DatabaseURL == "" {
		log.Fatal().Msg("DATABASE_URL is required")
	}

	// Crear syncer
	dbSyncer, err := syncer.NewDBSyncer(&syncer.SyncerConfig{
		DatabaseURL:  cfg.DatabaseURL,
		PhishTankKey: cfg.PhishTankKey,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create syncer")
	}

	// Contexto con cancelación
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar sincronización
	dbSyncer.Start(ctx)

	// Servidor HTTP para health checks y status
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"fy-dbsync"}`))
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := dbSyncer.GetStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		source := r.URL.Query().Get("source")
		if source == "" {
			source = "all"
		}
		go dbSyncer.ForceSync(ctx, source)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Sync started",
			"source":  source,
		})
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Iniciar servidor
	go func() {
		log.Info().Str("port", cfg.Port).Msg("HTTP server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("HTTP server failed")
		}
	}()

	// Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down...")

	// Detener syncer
	dbSyncer.Stop()

	// Shutdown del servidor
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)

	log.Info().Msg("Fy-DBSync stopped")
}

func setupLogger(cfg *config.Config) {
	if cfg.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

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
