package syncer

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/trackfy/fy-dbsync/internal/importer"
)

// DBSyncer maneja la sincronización periódica de las bases de datos de amenazas
type DBSyncer struct {
	db                    *sql.DB
	urlhausImporter       importer.Importer
	openphishImporter     importer.Importer
	stopforumspamImporter importer.Importer
	urlhausInterval       time.Duration
	openphishInterval     time.Duration
	emailsInterval        time.Duration
	stopCh                chan struct{}
}

// SyncerConfig configuración para el syncer
type SyncerConfig struct {
	DatabaseURL       string
	URLhausInterval   time.Duration
	OpenPhishInterval time.Duration
	EmailsInterval    time.Duration
}

// NewDBSyncer crea un nuevo sincronizador de DBs
func NewDBSyncer(cfg *SyncerConfig) (*DBSyncer, error) {
	if cfg.DatabaseURL == "" {
		return nil, nil
	}

	// Conectar a PostgreSQL
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Configurar pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verificar conexión
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	log.Info().Msg("[DBSyncer] Connected to PostgreSQL")

	// Valores por defecto
	urlhausInterval := cfg.URLhausInterval
	if urlhausInterval == 0 {
		urlhausInterval = 5 * time.Minute
	}
	openphishInterval := cfg.OpenPhishInterval
	if openphishInterval == 0 {
		openphishInterval = 1 * time.Hour
	}
	emailsInterval := cfg.EmailsInterval
	if emailsInterval == 0 {
		emailsInterval = 24 * time.Hour // Emails se actualizan cada 24h
	}

	return &DBSyncer{
		db:                    db,
		urlhausImporter:       importer.NewURLhausImporter(db),
		openphishImporter:     importer.NewOpenPhishImporter(db),
		stopforumspamImporter: importer.NewStopForumSpamImporter(db),
		urlhausInterval:       urlhausInterval,
		openphishInterval:     openphishInterval,
		emailsInterval:        emailsInterval,
		stopCh:                make(chan struct{}),
	}, nil
}

// Start inicia la sincronización periódica en background
func (s *DBSyncer) Start(ctx context.Context) {
	log.Info().
		Dur("urlhaus_interval", s.urlhausInterval).
		Dur("openphish_interval", s.openphishInterval).
		Dur("emails_interval", s.emailsInterval).
		Msg("[DBSyncer] Starting threat database synchronization")

	// Sincronización inicial
	go s.syncNow(ctx)

	// Goroutine para URLhaus
	go s.syncLoop(ctx, s.urlhausImporter, s.urlhausInterval)

	// Goroutine para OpenPhish
	go s.syncLoop(ctx, s.openphishImporter, s.openphishInterval)

	// Goroutine para StopForumSpam (emails)
	go s.syncLoop(ctx, s.stopforumspamImporter, s.emailsInterval)
}

// Stop detiene la sincronización
func (s *DBSyncer) Stop() {
	log.Info().Msg("[DBSyncer] Stopping database synchronization")
	close(s.stopCh)
	if s.db != nil {
		s.db.Close()
	}
}

// syncNow sincroniza todas las fuentes inmediatamente
func (s *DBSyncer) syncNow(ctx context.Context) {
	log.Info().Msg("[DBSyncer] Running initial sync...")

	// URLhaus
	if s.urlhausImporter != nil {
		if err := s.urlhausImporter.Sync(ctx); err != nil {
			log.Error().Err(err).Msg("[DBSyncer] Failed to sync URLhaus")
		}
	}

	// OpenPhish
	if s.openphishImporter != nil {
		if err := s.openphishImporter.Sync(ctx); err != nil {
			log.Error().Err(err).Msg("[DBSyncer] Failed to sync OpenPhish")
		}
	}

	// StopForumSpam (emails)
	if s.stopforumspamImporter != nil {
		if err := s.stopforumspamImporter.Sync(ctx); err != nil {
			log.Error().Err(err).Msg("[DBSyncer] Failed to sync StopForumSpam")
		}
	}

	log.Info().Msg("[DBSyncer] Initial sync completed")
}

// syncLoop ejecuta sincronización periódica para un importer
func (s *DBSyncer) syncLoop(ctx context.Context, imp importer.Importer, interval time.Duration) {
	if imp == nil {
		return
	}

	name := imp.Name()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug().Str("source", name).Msg("[DBSyncer] Context cancelled")
			return
		case <-s.stopCh:
			log.Debug().Str("source", name).Msg("[DBSyncer] Stop signal received")
			return
		case <-ticker.C:
			log.Info().Str("source", name).Msg("[DBSyncer] Running scheduled sync")

			syncCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
			if err := imp.Sync(syncCtx); err != nil {
				log.Error().Err(err).Str("source", name).Msg("[DBSyncer] Scheduled sync failed")
			} else {
				stats := imp.GetStats()
				log.Info().
					Str("source", name).
					Int64("records", stats.TotalRecords).
					Int64("errors", stats.Errors).
					Dur("duration", stats.Duration).
					Msg("[DBSyncer] Scheduled sync completed")
			}
			cancel()
		}
	}
}

// GetStatus retorna el estado de las sincronizaciones
func (s *DBSyncer) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})

	if s.urlhausImporter != nil {
		stats := s.urlhausImporter.GetStats()
		status["urlhaus"] = map[string]interface{}{
			"last_sync":     stats.LastImport.Format(time.RFC3339),
			"total_records": stats.TotalRecords,
			"errors":        stats.Errors,
			"duration_ms":   stats.Duration.Milliseconds(),
		}
	}

	if s.openphishImporter != nil {
		stats := s.openphishImporter.GetStats()
		status["openphish"] = map[string]interface{}{
			"last_sync":     stats.LastImport.Format(time.RFC3339),
			"total_records": stats.TotalRecords,
			"errors":        stats.Errors,
			"duration_ms":   stats.Duration.Milliseconds(),
		}
	}

	if s.stopforumspamImporter != nil {
		stats := s.stopforumspamImporter.GetStats()
		status["stopforumspam"] = map[string]interface{}{
			"last_sync":     stats.LastImport.Format(time.RFC3339),
			"total_records": stats.TotalRecords,
			"errors":        stats.Errors,
			"duration_ms":   stats.Duration.Milliseconds(),
		}
	}

	return status
}

// ForceSync fuerza una sincronización inmediata
func (s *DBSyncer) ForceSync(ctx context.Context, source string) error {
	switch source {
	case "urlhaus":
		if s.urlhausImporter != nil {
			return s.urlhausImporter.Sync(ctx)
		}
	case "openphish":
		if s.openphishImporter != nil {
			return s.openphishImporter.Sync(ctx)
		}
	case "stopforumspam", "emails":
		if s.stopforumspamImporter != nil {
			return s.stopforumspamImporter.Sync(ctx)
		}
	case "all":
		s.syncNow(ctx)
	}
	return nil
}
