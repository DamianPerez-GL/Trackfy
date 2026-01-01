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

// PhishTankDownloader descarga la base de datos de PhishTank
type PhishTankDownloader struct {
	dbPath      string
	downloadURL string
	apiKey      string
	lastUpdate  time.Time
	lastSize    int64
}

// NewPhishTankDownloader crea un nuevo downloader de PhishTank
func NewPhishTankDownloader(dbPath string, apiKey string) *PhishTankDownloader {
	downloadURL := "http://data.phishtank.com/data/online-valid.json"
	if apiKey != "" {
		downloadURL = fmt.Sprintf("http://data.phishtank.com/data/%s/online-valid.json", apiKey)
	}

	return &PhishTankDownloader{
		dbPath:      dbPath,
		downloadURL: downloadURL,
		apiKey:      apiKey,
	}
}

// Name retorna el nombre del downloader
func (d *PhishTankDownloader) Name() string {
	return "phishtank"
}

// Download descarga la base de datos actualizada
func (d *PhishTankDownloader) Download(ctx context.Context) error {
	log.Info().Str("url", d.downloadURL).Msg("[PhishTank] Downloading database...")

	req, err := http.NewRequestWithContext(ctx, "GET", d.downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// PhishTank requiere User-Agent
	req.Header.Set("User-Agent", "phishtank/fy-dbsync")

	client := &http.Client{Timeout: 120 * time.Second} // PhishTank puede ser lento
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
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
		Msg("[PhishTank] Database downloaded successfully")

	return nil
}

// GetStats retorna estadísticas del downloader
func (d *PhishTankDownloader) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"last_sync": d.lastUpdate.Format(time.RFC3339),
		"last_size": d.lastSize,
	}
}
