package importer

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const urlhausDownloadURL = "https://urlhaus.abuse.ch/downloads/csv/"

// URLhausImporter descarga e importa datos de URLhaus directamente a PostgreSQL
type URLhausImporter struct {
	db        *sql.DB
	lastStats ImportStats
}

// NewURLhausImporter crea un nuevo importer de URLhaus
func NewURLhausImporter(db *sql.DB) *URLhausImporter {
	return &URLhausImporter{db: db}
}

// Name retorna el nombre del importer
func (i *URLhausImporter) Name() string {
	return "urlhaus"
}

// GetStats retorna las estadísticas de la última importación
func (i *URLhausImporter) GetStats() ImportStats {
	return i.lastStats
}

// Sync descarga e importa los datos directamente a PostgreSQL (sin archivos intermedios)
func (i *URLhausImporter) Sync(ctx context.Context) error {
	startTime := time.Now()
	stats := ImportStats{LastImport: startTime}

	log.Info().Str("url", urlhausDownloadURL).Msg("[URLhaus] Downloading and importing to PostgreSQL...")

	// Descargar datos
	req, err := http.NewRequestWithContext(ctx, "GET", urlhausDownloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	log.Info().Msg("[URLhaus] Download complete, parsing and importing...")

	// Preparar transacción
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Preparar statements
	upsertDomainStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO threat_domains (domain_hash, domain, threat_type, severity, confidence, source, source_id, tld, first_seen, last_seen, flags)
		VALUES (sha256_bytea($1), $1, $2::threat_type_enum, $3::severity_enum, $4, 'urlhaus'::source_enum, $5, $6, $7, $8, 1)
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
		INSERT INTO threat_paths (path_hash, domain_hash, path, threat_type, severity, confidence, source, first_seen, last_seen, flags)
		VALUES (sha256_bytea($1), sha256_bytea($2), $3, $4::threat_type_enum, $5::severity_enum, $6, 'urlhaus'::source_enum, $7, $8, 1)
		ON CONFLICT (path_hash) DO UPDATE SET
			last_seen = EXCLUDED.last_seen
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare path statement: %w", err)
	}
	defer upsertPathStmt.Close()

	// Parsear CSV directamente del body (streaming)
	reader := csv.NewReader(bufio.NewReader(resp.Body))
	reader.Comment = '#'
	reader.FieldsPerRecord = -1

	lineNum := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			stats.Errors++
			continue
		}

		lineNum++
		stats.TotalRecords++

		// Campos mínimos: id, dateadded, url, url_status, last_online, threat
		if len(record) < 6 {
			stats.Errors++
			continue
		}

		sourceID := record[0]
		dateAdded := record[1]
		rawURL := record[2]
		urlStatus := record[3]
		threat := record[5]

		// Solo importar URLs con estado válido
		if urlStatus != "online" && urlStatus != "offline" {
			continue
		}

		// Parsear URL
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			stats.Errors++
			continue
		}

		domain := strings.ToLower(parsedURL.Hostname())
		if domain == "" {
			stats.Errors++
			continue
		}

		threatType := mapURLhausThreat(threat)
		severity := "high"
		confidence := int16(85)

		firstSeen, err := time.Parse("2006-01-02 15:04:05", dateAdded)
		if err != nil {
			firstSeen = time.Now()
		}
		lastSeen := time.Now()

		tld := extractTLD(domain)

		// Insertar dominio
		_, err = upsertDomainStmt.ExecContext(ctx,
			domain, threatType, severity, confidence, sourceID, tld, firstSeen, lastSeen,
		)
		if err != nil {
			stats.Errors++
			continue
		}

		// Insertar path si existe
		path := parsedURL.Path
		if path != "" && path != "/" {
			fullPath := domain + path
			_, err = upsertPathStmt.ExecContext(ctx,
				fullPath, domain, path, threatType, severity, confidence, firstSeen, lastSeen,
			)
			if err != nil {
				log.Debug().Err(err).Str("path", path).Msg("[URLhaus] Failed to upsert path")
			}
		}

		if lineNum%10000 == 0 {
			log.Info().Int("processed", lineNum).Msg("[URLhaus] Import progress")
		}
	}

	// Actualizar sync_status
	_, _ = tx.ExecContext(ctx, `
		INSERT INTO sync_status (source, last_sync, last_count)
		VALUES ('urlhaus'::source_enum, NOW(), $1)
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
		Msg("[URLhaus] Import completed")

	return nil
}

func mapURLhausThreat(threat string) string {
	switch strings.ToLower(strings.TrimSpace(threat)) {
	case "malware_download":
		return "malware"
	case "phishing":
		return "phishing"
	case "cryptojacking":
		return "cryptojacking"
	case "ransomware":
		return "ransomware"
	default:
		return "malware"
	}
}

func extractTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}
