package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// URLhausDownloader descarga la base de datos de URLhaus
type URLhausDownloader struct {
	dbPath      string
	downloadURL string
	lastUpdate  time.Time
	lastSize    int64
}

// NewURLhausDownloader crea un nuevo downloader de URLhaus
func NewURLhausDownloader(dbPath string) *URLhausDownloader {
	return &URLhausDownloader{
		dbPath:      dbPath,
		downloadURL: "https://urlhaus.abuse.ch/downloads/csv/",
	}
}

// Name retorna el nombre del downloader
func (d *URLhausDownloader) Name() string {
	return "urlhaus"
}

// Download descarga la base de datos actualizada
func (d *URLhausDownloader) Download(ctx context.Context) error {
	log.Info().Str("url", d.downloadURL).Msg("[URLhaus] Downloading database...")

	req, err := http.NewRequestWithContext(ctx, "GET", d.downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Guardar a archivo temporal primero
	tmpPath := d.dbPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	written, err := io.Copy(file, resp.Body)
	file.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write DB file: %w", err)
	}

	// Mover archivo temporal al destino final (atómico)
	if err := os.Rename(tmpPath, d.dbPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to move temp file: %w", err)
	}

	d.lastUpdate = time.Now()
	d.lastSize = written

	log.Info().
		Int64("bytes", written).
		Str("path", d.dbPath).
		Msg("[URLhaus] Database downloaded successfully")

	return nil
}

// GetStats retorna estadísticas del downloader
func (d *URLhausDownloader) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"last_sync": d.lastUpdate.Format(time.RFC3339),
		"last_size": d.lastSize,
	}
}
