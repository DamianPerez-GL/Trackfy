package importer

import (
	"bufio"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Stop Forum Spam - emails reportados por spam
const stopForumSpamEmailsURL = "https://www.stopforumspam.com/downloads/listed_email_365_all.gz"

// StopForumSpamImporter descarga e importa emails de spam desde Stop Forum Spam
type StopForumSpamImporter struct {
	db        *sql.DB
	lastStats ImportStats
}

// NewStopForumSpamImporter crea un nuevo importer de Stop Forum Spam
func NewStopForumSpamImporter(db *sql.DB) *StopForumSpamImporter {
	return &StopForumSpamImporter{db: db}
}

// Name retorna el nombre del importer
func (i *StopForumSpamImporter) Name() string {
	return "stopforumspam"
}

// GetStats retorna las estadísticas de la última importación
func (i *StopForumSpamImporter) GetStats() ImportStats {
	return i.lastStats
}

// Sync descarga e importa los emails de spam directamente a PostgreSQL
func (i *StopForumSpamImporter) Sync(ctx context.Context) error {
	startTime := time.Now()
	stats := ImportStats{LastImport: startTime}

	log.Info().Str("url", stopForumSpamEmailsURL).Msg("[StopForumSpam] Downloading emails and importing to PostgreSQL...")

	req, err := http.NewRequestWithContext(ctx, "GET", stopForumSpamEmailsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Fy-DBSync/1.0")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Descomprimir gzip
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	log.Info().Msg("[StopForumSpam] Download complete, parsing and importing...")

	// Parsear CSV: email,count,lastseen
	scanner := bufio.NewScanner(gzReader)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineNum := 0
	inserted := int64(0)
	batchSize := 500
	var batch []emailEntry

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Formato: email,count,lastseen
		parts := strings.Split(line, ",")
		if len(parts) < 1 {
			continue
		}

		email := strings.ToLower(strings.TrimSpace(parts[0]))
		if email == "" || !strings.Contains(email, "@") {
			continue
		}

		// Validar formato básico de email
		if len(email) < 5 || len(email) > 254 {
			continue
		}

		lineNum++
		stats.TotalRecords++

		// Extraer dominio del email
		atIndex := strings.LastIndex(email, "@")
		if atIndex < 1 {
			stats.Errors++
			continue
		}
		domain := email[atIndex+1:]

		// Calcular confianza basado en el count si está disponible
		confidence := int16(70)
		if len(parts) >= 2 {
			// Si tiene más de 10 reportes, mayor confianza
			if parts[1] != "" {
				var count int
				fmt.Sscanf(parts[1], "%d", &count)
				if count > 100 {
					confidence = 95
				} else if count > 50 {
					confidence = 90
				} else if count > 10 {
					confidence = 80
				}
			}
		}

		batch = append(batch, emailEntry{
			email:      email,
			domain:     domain,
			confidence: confidence,
			sourceID:   fmt.Sprintf("sfs-%d", lineNum),
		})

		if len(batch) >= batchSize {
			n, errs := i.insertEmailBatch(ctx, batch)
			inserted += n
			stats.Errors += errs
			batch = batch[:0]

			if lineNum%50000 == 0 {
				log.Info().Int("processed", lineNum).Int64("inserted", inserted).Msg("[StopForumSpam] Import progress")
			}
		}
	}

	// Insertar último batch
	if len(batch) > 0 {
		n, errs := i.insertEmailBatch(ctx, batch)
		inserted += n
		stats.Errors += errs
	}

	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Msg("[StopForumSpam] Scanner error")
	}

	// Actualizar sync_status - usamos 'osint' como source ya que stopforumspam no está en el enum
	_, err = i.db.ExecContext(ctx, `
		INSERT INTO sync_status (source, last_sync, last_count)
		VALUES ('osint'::source_enum, NOW(), $1)
		ON CONFLICT (source) DO UPDATE SET
			last_sync = NOW(),
			last_count = $1
	`, inserted)
	if err != nil {
		log.Error().Err(err).Msg("[StopForumSpam] Failed to update sync_status")
	}

	stats.Duration = time.Since(startTime)
	i.lastStats = stats

	log.Info().
		Int64("total", stats.TotalRecords).
		Int64("inserted", inserted).
		Int64("errors", stats.Errors).
		Dur("duration", stats.Duration).
		Msg("[StopForumSpam] Import completed")

	return nil
}

type emailEntry struct {
	email      string
	domain     string
	confidence int16
	sourceID   string
}

// insertEmailBatch inserta un batch de emails
func (i *StopForumSpamImporter) insertEmailBatch(ctx context.Context, batch []emailEntry) (int64, int64) {
	var inserted, errors int64
	now := time.Now()

	for _, entry := range batch {
		_, err := i.db.ExecContext(ctx, `
			INSERT INTO threat_emails (email_hash, email, domain_hash, threat_type, severity, confidence, source, first_seen, last_seen, flags)
			VALUES (sha256_bytea($1), $1, sha256_bytea($2), 'spam'::threat_type_enum, 'medium'::severity_enum, $3, 'osint'::source_enum, $4, $5, 1)
			ON CONFLICT (email_hash) DO UPDATE SET
				last_seen = EXCLUDED.last_seen,
				report_count = threat_emails.report_count + 1,
				confidence = GREATEST(threat_emails.confidence, EXCLUDED.confidence)
		`, entry.email, entry.domain, entry.confidence, now, now)

		if err != nil {
			errors++
			continue
		}
		inserted++
	}

	return inserted, errors
}
