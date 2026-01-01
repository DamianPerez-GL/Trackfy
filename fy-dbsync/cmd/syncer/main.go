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

	"github.com/trackfy/fy-dbsync/internal/config"
	"github.com/trackfy/fy-dbsync/internal/downloader"
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

	// Inicializar downloaders
	urlhausDownloader := downloader.NewURLhausDownloader(cfg.URLhausDBPath)
	phishtankDownloader := downloader.NewPhishTankDownloader(cfg.PhishTankDBPath, cfg.PhishTankKey)

	// Crear syncer
	dbSyncer := syncer.NewDBSyncer(urlhausDownloader, phishtankDownloader)

	// Contexto con cancelación
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar sincronización
	dbSyncer.Start(ctx)

	// Servidor HTTP mínimo para health checks
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"fy-dbsync"}`))
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := dbSyncer.GetStatus()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Simple status response
		w.Write([]byte(`{"urlhaus":` + formatStats(status["urlhaus"]) + `,"phishtank":` + formatStats(status["phishtank"]) + `}`))
	})
	mux.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db := r.URL.Query().Get("db")
		if db == "" {
			db = "all"
		}
		go dbSyncer.ForceSync(ctx, db)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"message":"Sync started","db":"` + db + `"}`))
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Iniciar servidor
	go func() {
		log.Info().Str("port", cfg.Port).Msg("Health server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Health server failed")
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

func formatStats(stats interface{}) string {
	if stats == nil {
		return `{"status":"not_initialized"}`
	}
	m, ok := stats.(map[string]interface{})
	if !ok {
		return `{"status":"unknown"}`
	}
	urls, _ := m["urls"].(int)
	lastSync, _ := m["last_sync"].(string)
	return `{"urls":` + intToStr(urls) + `,"last_sync":"` + lastSync + `"}`
}

func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
