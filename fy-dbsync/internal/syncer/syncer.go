package syncer

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trackfy/fy-dbsync/internal/downloader"
)

// DBSyncer maneja la sincronización periódica de las bases de datos
type DBSyncer struct {
	urlhausDownloader   downloader.Downloader
	phishtankDownloader downloader.Downloader
	urlhausInterval     time.Duration
	phishtankInterval   time.Duration
	stopCh              chan struct{}
}

// NewDBSyncer crea un nuevo sincronizador de DBs
func NewDBSyncer(urlhaus, phishtank downloader.Downloader) *DBSyncer {
	return &DBSyncer{
		urlhausDownloader:   urlhaus,
		phishtankDownloader: phishtank,
		urlhausInterval:     5 * time.Minute,  // URLhaus se actualiza cada 5 min
		phishtankInterval:   1 * time.Hour,    // PhishTank cada 1 hora
		stopCh:              make(chan struct{}),
	}
}

// Start inicia la sincronización periódica en background
func (s *DBSyncer) Start(ctx context.Context) {
	log.Info().
		Dur("urlhaus_interval", s.urlhausInterval).
		Dur("phishtank_interval", s.phishtankInterval).
		Msg("[DBSyncer] Starting database synchronization")

	// Sincronización inicial
	go s.syncNow(ctx)

	// Goroutine para URLhaus
	go s.syncLoop(ctx, s.urlhausDownloader, s.urlhausInterval)

	// Goroutine para PhishTank
	go s.syncLoop(ctx, s.phishtankDownloader, s.phishtankInterval)
}

// Stop detiene la sincronización
func (s *DBSyncer) Stop() {
	log.Info().Msg("[DBSyncer] Stopping database synchronization")
	close(s.stopCh)
}

// syncNow sincroniza todas las DBs inmediatamente
func (s *DBSyncer) syncNow(ctx context.Context) {
	log.Info().Msg("[DBSyncer] Running initial sync...")

	// URLhaus
	if s.urlhausDownloader != nil {
		if err := s.urlhausDownloader.Download(ctx); err != nil {
			log.Error().Err(err).Msg("[DBSyncer] Failed to sync URLhaus")
		} else {
			log.Info().Msg("[DBSyncer] URLhaus synced successfully")
		}
	}

	// PhishTank
	if s.phishtankDownloader != nil {
		if err := s.phishtankDownloader.Download(ctx); err != nil {
			log.Error().Err(err).Msg("[DBSyncer] Failed to sync PhishTank")
		} else {
			log.Info().Msg("[DBSyncer] PhishTank synced successfully")
		}
	}

	log.Info().Msg("[DBSyncer] Initial sync completed")
}

// syncLoop ejecuta sincronización periódica para un downloader específico
func (s *DBSyncer) syncLoop(ctx context.Context, dl downloader.Downloader, interval time.Duration) {
	if dl == nil {
		return
	}

	name := dl.Name()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug().Str("db", name).Msg("[DBSyncer] Context cancelled, stopping sync loop")
			return
		case <-s.stopCh:
			log.Debug().Str("db", name).Msg("[DBSyncer] Stop signal received")
			return
		case <-ticker.C:
			log.Debug().Str("db", name).Msg("[DBSyncer] Running scheduled sync")

			syncCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			if err := dl.Download(syncCtx); err != nil {
				log.Error().Err(err).Str("db", name).Msg("[DBSyncer] Scheduled sync failed")
			} else {
				log.Info().Str("db", name).Msg("[DBSyncer] Scheduled sync completed")
			}
			cancel()
		}
	}
}

// GetStatus retorna el estado de las DBs
func (s *DBSyncer) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})

	if s.urlhausDownloader != nil {
		status["urlhaus"] = s.urlhausDownloader.GetStats()
	}

	if s.phishtankDownloader != nil {
		status["phishtank"] = s.phishtankDownloader.GetStats()
	}

	return status
}

// ForceSync fuerza una sincronización inmediata de una DB específica
func (s *DBSyncer) ForceSync(ctx context.Context, dbName string) error {
	switch dbName {
	case "urlhaus":
		if s.urlhausDownloader != nil {
			return s.urlhausDownloader.Download(ctx)
		}
	case "phishtank":
		if s.phishtankDownloader != nil {
			return s.phishtankDownloader.Download(ctx)
		}
	case "all":
		s.syncNow(ctx)
		return nil
	}
	return nil
}
