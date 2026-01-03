package importer

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const openPhishURL = "https://openphish.com/feed.txt"

// OpenPhishImporter descarga e importa datos de OpenPhish directamente a PostgreSQL
type OpenPhishImporter struct {
	db        *sql.DB
	lastStats ImportStats
}

// NewOpenPhishImporter crea un nuevo importer de OpenPhish
func NewOpenPhishImporter(db *sql.DB) *OpenPhishImporter {
	return &OpenPhishImporter{db: db}
}

// Name retorna el nombre del importer
func (i *OpenPhishImporter) Name() string {
	return "openphish"
}

// GetStats retorna las estadísticas de la última importación
func (i *OpenPhishImporter) GetStats() ImportStats {
	return i.lastStats
}

// Sync descarga e importa los datos directamente a PostgreSQL
func (i *OpenPhishImporter) Sync(ctx context.Context) error {
	startTime := time.Now()
	stats := ImportStats{LastImport: startTime}

	log.Info().Str("url", openPhishURL).Msg("[OpenPhish] Downloading and importing to PostgreSQL...")

	req, err := http.NewRequestWithContext(ctx, "GET", openPhishURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Fy-DBSync/1.0")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	log.Info().Msg("[OpenPhish] Download complete, parsing and importing...")

	// Parsear línea por línea
	scanner := bufio.NewScanner(resp.Body)
	lineNum := 0
	inserted := int64(0)
	batchSize := 100
	var batch []phishEntry

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lineNum++
		stats.TotalRecords++

		parsedURL, err := url.Parse(line)
		if err != nil {
			stats.Errors++
			continue
		}

		domain := strings.ToLower(parsedURL.Hostname())
		if domain == "" || len(domain) < 3 || !strings.Contains(domain, ".") {
			stats.Errors++
			continue
		}

		// Saltar IPs
		if net.ParseIP(domain) != nil {
			continue
		}

		batch = append(batch, phishEntry{
			domain:   domain,
			path:     parsedURL.Path,
			sourceID: fmt.Sprintf("openphish-%d", lineNum),
			tld:      extractTLD(domain),
		})

		if len(batch) >= batchSize {
			n, errs := i.insertBatch(ctx, batch)
			inserted += n
			stats.Errors += errs
			batch = batch[:0]
		}
	}

	// Insertar último batch
	if len(batch) > 0 {
		n, errs := i.insertBatch(ctx, batch)
		inserted += n
		stats.Errors += errs
	}

	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Msg("[OpenPhish] Scanner error")
	}

	// Actualizar sync_status
	_, err = i.db.ExecContext(ctx, `
		INSERT INTO sync_status (source, last_sync, last_count)
		VALUES ('phishtank'::source_enum, NOW(), $1)
		ON CONFLICT (source) DO UPDATE SET
			last_sync = NOW(),
			last_count = $1
	`, inserted)
	if err != nil {
		log.Error().Err(err).Msg("[OpenPhish] Failed to update sync_status")
	}

	stats.Duration = time.Since(startTime)
	i.lastStats = stats

	log.Info().
		Int64("total", stats.TotalRecords).
		Int64("inserted", inserted).
		Int64("errors", stats.Errors).
		Dur("duration", stats.Duration).
		Msg("[OpenPhish] Import completed")

	return nil
}

type phishEntry struct {
	domain   string
	path     string
	sourceID string
	tld      string
}

// insertBatch inserta un batch de dominios de phishing
func (i *OpenPhishImporter) insertBatch(ctx context.Context, batch []phishEntry) (int64, int64) {
	var inserted, errors int64
	now := time.Now()

	for _, entry := range batch {
		_, err := i.db.ExecContext(ctx, `
			INSERT INTO threat_domains (domain_hash, domain, threat_type, severity, confidence, source, source_id, tld, first_seen, last_seen, flags)
			VALUES (sha256_bytea($1), $1, 'phishing'::threat_type_enum, 'high'::severity_enum, 90, 'phishtank'::source_enum, $2, $3, $4, $5, 1)
			ON CONFLICT (domain_hash) DO UPDATE SET
				last_seen = EXCLUDED.last_seen,
				hit_count = threat_domains.hit_count + 1,
				confidence = GREATEST(threat_domains.confidence, EXCLUDED.confidence)
		`, entry.domain, entry.sourceID, entry.tld, now, now)

		if err != nil {
			errors++
			continue
		}
		inserted++

		// Insertar path si existe
		if entry.path != "" && entry.path != "/" {
			fullPath := entry.domain + entry.path
			i.db.ExecContext(ctx, `
				INSERT INTO threat_paths (path_hash, domain_hash, path, threat_type, severity, confidence, source, first_seen, last_seen, flags)
				VALUES (sha256_bytea($1), sha256_bytea($2), $3, 'phishing'::threat_type_enum, 'high'::severity_enum, 90, 'phishtank'::source_enum, $4, $5, 1)
				ON CONFLICT (path_hash) DO UPDATE SET last_seen = EXCLUDED.last_seen
			`, fullPath, entry.domain, entry.path, now, now)
		}
	}

	return inserted, errors
}
