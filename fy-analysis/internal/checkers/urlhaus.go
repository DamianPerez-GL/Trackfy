package checkers

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// URLhausChecker verifica URLs contra la base de datos de URLhaus (abuse.ch)
type URLhausChecker struct {
	enabled     bool
	weight      float64
	dbPath      string
	urlDB       map[string]*URLhausEntry // URL -> Entry
	domainDB    map[string]*URLhausEntry // Domain -> Entry
	mu          sync.RWMutex
	lastUpdate  time.Time
	downloadURL string
}

// URLhausEntry representa una entrada en la DB de URLhaus
type URLhausEntry struct {
	ID          string
	DateAdded   string
	URL         string
	URLStatus   string // online, offline
	Threat      string // malware_download, etc
	Tags        []string
	URLhausLink string
	Reporter    string
}

// NewURLhausChecker crea un nuevo checker de URLhaus
func NewURLhausChecker(dbPath string) *URLhausChecker {
	checker := &URLhausChecker{
		enabled:     true,
		weight:      0.40,
		dbPath:      dbPath,
		urlDB:       make(map[string]*URLhausEntry),
		domainDB:    make(map[string]*URLhausEntry),
		downloadURL: "https://urlhaus.abuse.ch/downloads/csv/",
	}

	// Intentar cargar DB existente
	if err := checker.LoadDB(); err != nil {
		log.Warn().Err(err).Msg("[URLhaus] Failed to load existing DB, will download")
	}

	return checker
}

// Name retorna el nombre del checker
func (c *URLhausChecker) Name() string {
	return "urlhaus"
}

// Weight retorna el peso del checker
func (c *URLhausChecker) Weight() float64 {
	return c.weight
}

// IsEnabled indica si el checker está habilitado
func (c *URLhausChecker) IsEnabled() bool {
	return c.enabled
}

// SupportedTypes retorna los tipos soportados (solo URLs)
func (c *URLhausChecker) SupportedTypes() []InputType {
	return []InputType{InputTypeURL}
}

// Check verifica los indicadores contra URLhaus
func (c *URLhausChecker) Check(ctx context.Context, indicators *Indicators) (*CheckResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := &CheckResult{
		Source:  c.Name(),
		Found:   false,
		RawData: make(map[string]interface{}),
	}

	if len(c.urlDB) == 0 {
		log.Debug().Msg("[URLhaus] Database is empty")
		return result, nil
	}

	log.Debug().
		Str("url", indicators.FullURL).
		Str("domain", indicators.Domain).
		Int("db_size", len(c.urlDB)).
		Msg("[URLhaus] Checking indicators")

	// 1. Buscar URL exacta
	if entry, found := c.urlDB[indicators.FullURL]; found {
		result.Found = true
		result.ThreatType = c.mapThreatType(entry.Threat)
		result.Confidence = 0.95
		result.Tags = entry.Tags
		result.RawData["urlhaus_id"] = entry.ID
		result.RawData["urlhaus_link"] = entry.URLhausLink
		result.RawData["status"] = entry.URLStatus
		result.RawData["match_type"] = "exact_url"

		log.Info().
			Str("url", indicators.FullURL).
			Str("threat", entry.Threat).
			Str("status", entry.URLStatus).
			Msg("[URLhaus] FOUND - Exact URL match")
		return result, nil
	}

	// 2. Buscar variantes de URL (con/sin trailing slash, http/https)
	variants := c.generateURLVariants(indicators.FullURL)
	for _, variant := range variants {
		if entry, found := c.urlDB[variant]; found {
			result.Found = true
			result.ThreatType = c.mapThreatType(entry.Threat)
			result.Confidence = 0.90
			result.Tags = entry.Tags
			result.RawData["urlhaus_id"] = entry.ID
			result.RawData["matched_variant"] = variant
			result.RawData["match_type"] = "url_variant"

			log.Info().
				Str("url", indicators.FullURL).
				Str("matched", variant).
				Msg("[URLhaus] FOUND - URL variant match")
			return result, nil
		}
	}

	// 3. Buscar por dominio
	if entry, found := c.domainDB[indicators.Domain]; found {
		result.Found = true
		result.ThreatType = c.mapThreatType(entry.Threat)
		result.Confidence = 0.75 // Menor confianza porque es match de dominio
		result.Tags = entry.Tags
		result.RawData["urlhaus_id"] = entry.ID
		result.RawData["matched_url"] = entry.URL
		result.RawData["match_type"] = "domain"

		log.Info().
			Str("domain", indicators.Domain).
			Str("matched_url", entry.URL).
			Msg("[URLhaus] FOUND - Domain match")
		return result, nil
	}

	log.Debug().
		Str("url", indicators.FullURL).
		Msg("[URLhaus] Not found in database")

	return result, nil
}

// LoadDB carga la base de datos desde el archivo CSV
func (c *URLhausChecker) LoadDB() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(c.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DB file: %w", err)
	}
	defer file.Close()

	return c.parseCSV(file)
}

// DownloadDB descarga la base de datos actualizada
func (c *URLhausChecker) DownloadDB(ctx context.Context) error {
	log.Info().Str("url", c.downloadURL).Msg("[URLhaus] Downloading database...")

	req, err := http.NewRequestWithContext(ctx, "GET", c.downloadURL, nil)
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

	// Guardar a archivo
	file, err := os.Create(c.dbPath)
	if err != nil {
		return fmt.Errorf("failed to create DB file: %w", err)
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write DB file: %w", err)
	}

	log.Info().Int64("bytes", written).Msg("[URLhaus] Database downloaded")

	// Recargar DB
	file.Seek(0, 0)

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.parseCSV(file); err != nil {
		return fmt.Errorf("failed to parse downloaded DB: %w", err)
	}

	c.lastUpdate = time.Now()
	log.Info().
		Int("urls", len(c.urlDB)).
		Int("domains", len(c.domainDB)).
		Msg("[URLhaus] Database loaded")

	return nil
}

// parseCSV parsea el archivo CSV de URLhaus
func (c *URLhausChecker) parseCSV(reader io.Reader) error {
	// Limpiar DBs existentes
	c.urlDB = make(map[string]*URLhausEntry)
	c.domainDB = make(map[string]*URLhausEntry)

	scanner := bufio.NewScanner(reader)

	// Saltear comentarios iniciales (líneas que empiezan con #)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			// Primera línea de datos
			break
		}
	}

	// Crear reader CSV desde la posición actual
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1 // Número variable de campos
	csvReader.LazyQuotes = true

	count := 0
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Saltar líneas mal formateadas
		}

		// Formato esperado: id,dateadded,url,url_status,last_online,threat,tags,urlhaus_link,reporter
		if len(record) < 8 {
			continue
		}

		entry := &URLhausEntry{
			ID:          record[0],
			DateAdded:   record[1],
			URL:         strings.TrimSpace(record[2]),
			URLStatus:   record[3],
			Threat:      record[5],
			Tags:        c.parseTags(record[6]),
			URLhausLink: record[7],
		}
		if len(record) > 8 {
			entry.Reporter = record[8]
		}

		// Indexar por URL
		c.urlDB[entry.URL] = entry

		// Extraer e indexar por dominio
		domain := c.extractDomain(entry.URL)
		if domain != "" {
			c.domainDB[domain] = entry
		}

		count++
	}

	log.Debug().Int("entries", count).Msg("[URLhaus] CSV parsed")
	return nil
}

// parseTags parsea los tags del formato URLhaus
func (c *URLhausChecker) parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return nil
	}
	// Tags vienen separados por coma
	tags := strings.Split(tagsStr, ",")
	var cleaned []string
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			cleaned = append(cleaned, t)
		}
	}
	return cleaned
}

// extractDomain extrae el dominio de una URL
func (c *URLhausChecker) extractDomain(urlStr string) string {
	// Remover scheme
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")

	// Tomar solo el host
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		urlStr = urlStr[:idx]
	}
	if idx := strings.Index(urlStr, ":"); idx != -1 {
		urlStr = urlStr[:idx]
	}

	return strings.ToLower(urlStr)
}

// mapThreatType mapea el tipo de amenaza de URLhaus a nuestros tipos
func (c *URLhausChecker) mapThreatType(urlhausThreat string) string {
	switch strings.ToLower(urlhausThreat) {
	case "malware_download":
		return ThreatTypeMalware
	case "phishing":
		return ThreatTypePhishing
	default:
		return ThreatTypeMalware // URLhaus es principalmente malware
	}
}

// generateURLVariants genera variantes de una URL para búsqueda
func (c *URLhausChecker) generateURLVariants(url string) []string {
	variants := []string{}

	// Con/sin trailing slash
	if strings.HasSuffix(url, "/") {
		variants = append(variants, strings.TrimSuffix(url, "/"))
	} else {
		variants = append(variants, url+"/")
	}

	// HTTP/HTTPS
	if strings.HasPrefix(url, "http://") {
		variants = append(variants, strings.Replace(url, "http://", "https://", 1))
	} else if strings.HasPrefix(url, "https://") {
		variants = append(variants, strings.Replace(url, "https://", "http://", 1))
	}

	return variants
}

// GetStats retorna estadísticas de la DB
func (c *URLhausChecker) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"urls":        len(c.urlDB),
		"domains":     len(c.domainDB),
		"last_update": c.lastUpdate,
	}
}
