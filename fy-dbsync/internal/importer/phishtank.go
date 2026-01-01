package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const phishtankBaseURL = "http://data.phishtank.com/data"

// PhishTankEntry representa una entrada de PhishTank
type PhishTankEntry struct {
	PhishID        int    `json:"phish_id"`
	URL            string `json:"url"`
	SubmissionTime string `json:"submission_time"`
	Verified       string `json:"verified"`
	Online         string `json:"online"`
	Target         string `json:"target"`
}

// PhishTankImporter descarga e importa datos de PhishTank directamente a PostgreSQL
type PhishTankImporter struct {
	db        *sql.DB
	apiKey    string
	lastStats ImportStats
}

// NewPhishTankImporter crea un nuevo importer de PhishTank
func NewPhishTankImporter(db *sql.DB, apiKey string) *PhishTankImporter {
	return &PhishTankImporter{
		db:     db,
		apiKey: apiKey,
	}
}

// Name retorna el nombre del importer
func (i *PhishTankImporter) Name() string {
	return "phishtank"
}

// GetStats retorna las estadísticas de la última importación
func (i *PhishTankImporter) GetStats() ImportStats {
	return i.lastStats
}

// getDownloadURL construye la URL de descarga
func (i *PhishTankImporter) getDownloadURL() string {
	if i.apiKey != "" {
		return fmt.Sprintf("%s/%s/online-valid.json", phishtankBaseURL, i.apiKey)
	}
	return fmt.Sprintf("%s/online-valid.json", phishtankBaseURL)
}

// Sync descarga e importa los datos directamente a PostgreSQL (sin archivos intermedios)
func (i *PhishTankImporter) Sync(ctx context.Context) error {
	startTime := time.Now()
	stats := ImportStats{LastImport: startTime}

	downloadURL := i.getDownloadURL()
	log.Info().Str("url", downloadURL).Msg("[PhishTank] Downloading and importing to PostgreSQL...")

	// Descargar datos
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "phishtank/fy-dbsync")

	client := &http.Client{Timeout: 180 * time.Second} // PhishTank puede ser lento
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	log.Info().Msg("[PhishTank] Download complete, parsing JSON...")

	// Parsear JSON directamente del body (streaming)
	var entries []PhishTankEntry
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&entries); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	log.Info().Int("entries", len(entries)).Msg("[PhishTank] Parsed entries, importing to PostgreSQL...")

	// Preparar transacción
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Preparar statements
	upsertDomainStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO threat_domains (domain_hash, domain, threat_type, severity, confidence, source, source_id, tld, first_seen, last_seen, flags)
		VALUES (sha256_bytea($1), $1, 'phishing'::threat_type_enum, $2::severity_enum, $3, 'phishtank'::source_enum, $4, $5, $6, $7, 1)
		ON CONFLICT (domain_hash) DO UPDATE SET
			last_seen = EXCLUDED.last_seen,
			hit_count = threat_domains.hit_count + 1,
			confidence = GREATEST(threat_domains.confidence, EXCLUDED.confidence)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare domain statement: %w", err)
	}
	defer upsertDomainStmt.Close()

	upsertPathStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO threat_paths (path_hash, domain_hash, path, threat_type, severity, confidence, source, target_brand, first_seen, last_seen, flags)
		VALUES (sha256_bytea($1), sha256_bytea($2), $3, 'phishing'::threat_type_enum, $4::severity_enum, $5, 'phishtank'::source_enum, $6, $7, $8, 1)
		ON CONFLICT (path_hash) DO UPDATE SET
			last_seen = EXCLUDED.last_seen,
			target_brand = COALESCE(EXCLUDED.target_brand, threat_paths.target_brand)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare path statement: %w", err)
	}
	defer upsertPathStmt.Close()

	// Procesar entradas
	for idx, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		stats.TotalRecords++

		// Solo importar URLs verificadas
		if entry.Verified != "yes" {
			continue
		}

		// Parsear URL
		parsedURL, err := url.Parse(entry.URL)
		if err != nil {
			stats.Errors++
			continue
		}

		domain := strings.ToLower(parsedURL.Hostname())
		if domain == "" {
			stats.Errors++
			continue
		}

		severity := "high"
		confidence := int16(90)
		if entry.Online == "yes" {
			confidence = 95
		}

		firstSeen, err := time.Parse("2006-01-02T15:04:05+00:00", entry.SubmissionTime)
		if err != nil {
			firstSeen = time.Now()
		}
		lastSeen := time.Now()

		tld := extractTLD(domain)
		sourceID := fmt.Sprintf("%d", entry.PhishID)

		// Insertar dominio
		_, err = upsertDomainStmt.ExecContext(ctx,
			domain, severity, confidence, sourceID, tld, firstSeen, lastSeen,
		)
		if err != nil {
			stats.Errors++
			continue
		}

		// Insertar path si existe
		path := parsedURL.Path
		if path != "" && path != "/" {
			fullPath := domain + path
			target := entry.Target
			if target == "" || target == "Other" {
				target = ""
			}

			_, err = upsertPathStmt.ExecContext(ctx,
				fullPath, domain, path, severity, confidence,
				sql.NullString{String: target, Valid: target != ""},
				firstSeen, lastSeen,
			)
			if err != nil {
				log.Debug().Err(err).Str("path", path).Msg("[PhishTank] Failed to upsert path")
			}
		}

		if (idx+1)%5000 == 0 {
			log.Info().Int("processed", idx+1).Int("total", len(entries)).Msg("[PhishTank] Import progress")
		}
	}

	// Actualizar sync_status
	_, _ = tx.ExecContext(ctx, `
		INSERT INTO sync_status (source, last_sync, last_count)
		VALUES ('phishtank'::source_enum, NOW(), $1)
		ON CONFLICT (source) DO UPDATE SET
			last_sync = NOW(),
			last_count = $1
	`, stats.TotalRecords)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	stats.Duration = time.Since(startTime)
	i.lastStats = stats

	log.Info().
		Int64("total", stats.TotalRecords).
		Int64("errors", stats.Errors).
		Dur("duration", stats.Duration).
		Msg("[PhishTank] Import completed")

	return nil
}
