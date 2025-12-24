package sync

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trackfy/fy-analysis/internal/checkers"
)

// DBSyncer maneja la sincronización periódica de las bases de datos locales
type DBSyncer struct {
	urlhausChecker   *checkers.URLhausChecker
	phishtankChecker *checkers.PhishTankChecker
	urlhausInterval  time.Duration
	phishtankInterval time.Duration
	stopCh           chan struct{}
}

// NewDBSyncer crea un nuevo sincronizador de DBs
func NewDBSyncer(urlhaus *checkers.URLhausChecker, phishtank *checkers.PhishTankChecker) *DBSyncer {
	return &DBSyncer{
		urlhausChecker:   urlhaus,
		phishtankChecker: phishtank,
		urlhausInterval:  5 * time.Minute,  // URLhaus se actualiza cada 5 min
		phishtankInterval: 1 * time.Hour,   // PhishTank cada 1 hora
		stopCh:           make(chan struct{}),
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
	go s.syncLoop(ctx, "urlhaus", s.urlhausInterval, func(ctx context.Context) error {
		return s.urlhausChecker.DownloadDB(ctx)
	})

	// Goroutine para PhishTank
	go s.syncLoop(ctx, "phishtank", s.phishtankInterval, func(ctx context.Context) error {
		return s.phishtankChecker.DownloadDB(ctx)
	})
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
	if s.urlhausChecker != nil {
		if err := s.urlhausChecker.DownloadDB(ctx); err != nil {
			log.Error().Err(err).Msg("[DBSyncer] Failed to sync URLhaus")
		} else {
			stats := s.urlhausChecker.GetStats()
			log.Info().
				Int("urls", stats["urls"].(int)).
				Msg("[DBSyncer] URLhaus synced successfully")
		}
	}

	// PhishTank
	if s.phishtankChecker != nil {
		if err := s.phishtankChecker.DownloadDB(ctx); err != nil {
			log.Error().Err(err).Msg("[DBSyncer] Failed to sync PhishTank")
		} else {
			stats := s.phishtankChecker.GetStats()
			log.Info().
				Int("urls", stats["urls"].(int)).
				Msg("[DBSyncer] PhishTank synced successfully")
		}
	}

	log.Info().Msg("[DBSyncer] Initial sync completed")
}

// syncLoop ejecuta sincronización periódica para una DB específica
func (s *DBSyncer) syncLoop(ctx context.Context, name string, interval time.Duration, syncFunc func(context.Context) error) {
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
			if err := syncFunc(syncCtx); err != nil {
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

	if s.urlhausChecker != nil {
		status["urlhaus"] = s.urlhausChecker.GetStats()
	}

	if s.phishtankChecker != nil {
		status["phishtank"] = s.phishtankChecker.GetStats()
	}

	return status
}

// ForceSync fuerza una sincronización inmediata de una DB específica
func (s *DBSyncer) ForceSync(ctx context.Context, dbName string) error {
	switch dbName {
	case "urlhaus":
		if s.urlhausChecker != nil {
			return s.urlhausChecker.DownloadDB(ctx)
		}
	case "phishtank":
		if s.phishtankChecker != nil {
			return s.phishtankChecker.DownloadDB(ctx)
		}
	case "all":
		s.syncNow(ctx)
		return nil
	}
	return nil
}
