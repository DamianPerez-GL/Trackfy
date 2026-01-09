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
	"github.com/trackfy/api-gateway/internal/api"
	"github.com/trackfy/api-gateway/internal/auth"
	"github.com/trackfy/api-gateway/internal/config"
	"github.com/trackfy/api-gateway/internal/db"
	"github.com/trackfy/api-gateway/internal/services"
)

func main() {
	// Configurar logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Starting Trackfy API Gateway...")

	// Cargar configuración
	cfg := config.Load()

	// Conectar a PostgreSQL
	postgres, err := db.NewPostgresDB(cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer postgres.Close()

	// Conectar a Redis
	redis, err := db.NewRedisDB(cfg.Redis.URL, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redis.Close()

	// Crear JWT Manager
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)

	// Crear cliente de Fy Engine
	fyEngine := services.NewFyEngineClient(cfg.FyEngine.URL, cfg.FyEngine.Timeout)

	// Crear cliente de Fy Analysis (para reportes)
	fyAnalysis := services.NewFyAnalysisClient(cfg.FyAnalysis.URL, cfg.FyAnalysis.Timeout)

	// Crear router
	router := api.NewRouter(postgres, redis, jwtManager, fyEngine, fyAnalysis)

	// Configurar servidor
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Iniciar servidor en goroutine
	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("API Gateway listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped")
}
