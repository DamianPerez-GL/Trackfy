package importer

import (
	"context"
	"time"
)

// Importer interface para sincronizar datos de amenazas a PostgreSQL
type Importer interface {
	// Sync descarga e importa los datos directamente a PostgreSQL
	Sync(ctx context.Context) error
	// Name retorna el nombre del importer
	Name() string
	// GetStats retorna estadísticas de la última sincronización
	GetStats() ImportStats
}

// ImportStats estadísticas de importación
type ImportStats struct {
	LastImport   time.Time
	TotalRecords int64
	Errors       int64
	Duration     time.Duration
}
