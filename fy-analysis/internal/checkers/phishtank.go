package checkers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// PhishTankChecker verifica URLs contra la base de datos de PhishTank
type PhishTankChecker struct {
	enabled     bool
	weight      float64
	dbPath      string
	urlDB       map[string]*PhishTankEntry
	domainDB    map[string]*PhishTankEntry
	mu          sync.RWMutex
	lastUpdate  time.Time
	downloadURL string
	apiKey      string // Opcional, para mayor rate limit
}

// PhishTankEntry representa una entrada en la DB de PhishTank
type PhishTankEntry struct {
	PhishID           int64    `json:"phish_id"`
	URL               string   `json:"url"`
	PhishDetailURL    string   `json:"phish_detail_url"`
	SubmissionTime    string   `json:"submission_time"`
	Verified          string   `json:"verified"`
	VerifiedTime      string   `json:"verification_time"`
	Online            string   `json:"online"`
	Target            string   `json:"target"`
	Details           []Detail `json:"details,omitempty"`
}

// Detail detalles adicionales de PhishTank
type Detail struct {
	IPAddress   string `json:"ip_address"`
	CIDRBlock   string `json:"cidr_block"`
	AnnouncedBy string `json:"announcing_network"`
	Country     string `json:"country"`
}

// NewPhishTankChecker crea un nuevo checker de PhishTank
func NewPhishTankChecker(dbPath string, apiKey string) *PhishTankChecker {
	checker := &PhishTankChecker{
		enabled:     true,
		weight:      0.20,
		dbPath:      dbPath,
		urlDB:       make(map[string]*PhishTankEntry),
		domainDB:    make(map[string]*PhishTankEntry),
		downloadURL: "http://data.phishtank.com/data/online-valid.json",
		apiKey:      apiKey,
	}

	// Si hay API key, usar la URL con autenticación
	if apiKey != "" {
		checker.downloadURL = fmt.Sprintf("http://data.phishtank.com/data/%s/online-valid.json", apiKey)
	}

	// Intentar cargar DB existente
	if err := checker.LoadDB(); err != nil {
		log.Warn().Err(err).Msg("[PhishTank] Failed to load existing DB, will download")
	}

	return checker
}

// Name retorna el nombre del checker
func (c *PhishTankChecker) Name() string {
	return "phishtank"
}

// Weight retorna el peso del checker
func (c *PhishTankChecker) Weight() float64 {
	return c.weight
}

// IsEnabled indica si el checker está habilitado
func (c *PhishTankChecker) IsEnabled() bool {
	return c.enabled
}

// SupportedTypes retorna los tipos soportados (solo URLs)
func (c *PhishTankChecker) SupportedTypes() []InputType {
	return []InputType{InputTypeURL}
}

// Check verifica los indicadores contra PhishTank
func (c *PhishTankChecker) Check(ctx context.Context, indicators *Indicators) (*CheckResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := &CheckResult{
		Source:  c.Name(),
		Found:   false,
		RawData: make(map[string]interface{}),
	}

	if len(c.urlDB) == 0 {
		log.Debug().Msg("[PhishTank] Database is empty")
		return result, nil
	}

	log.Debug().
		Str("url", indicators.FullURL).
		Int("db_size", len(c.urlDB)).
		Msg("[PhishTank] Checking indicators")

	// Normalizar URL para búsqueda
	searchURL := strings.ToLower(indicators.FullURL)

	// 1. Buscar URL exacta
	if entry, found := c.urlDB[searchURL]; found {
		return c.buildFoundResult(result, entry, "exact_url"), nil
	}

	// 2. Buscar variantes
	variants := c.generateURLVariants(searchURL)
	for _, variant := range variants {
		if entry, found := c.urlDB[variant]; found {
			result.RawData["matched_variant"] = variant
			return c.buildFoundResult(result, entry, "url_variant"), nil
		}
	}

	// 3. Buscar por dominio
	if entry, found := c.domainDB[indicators.Domain]; found {
		result.RawData["matched_url"] = entry.URL
		return c.buildFoundResult(result, entry, "domain"), nil
	}

	log.Debug().
		Str("url", indicators.FullURL).
		Msg("[PhishTank] Not found in database")

	return result, nil
}

// buildFoundResult construye el resultado cuando se encuentra una amenaza
func (c *PhishTankChecker) buildFoundResult(result *CheckResult, entry *PhishTankEntry, matchType string) *CheckResult {
	result.Found = true
	result.ThreatType = ThreatTypePhishing
	result.Confidence = c.calculateConfidence(entry)
	result.Tags = []string{"phishing", entry.Target}
	result.RawData["phish_id"] = entry.PhishID
	result.RawData["phish_detail_url"] = entry.PhishDetailURL
	result.RawData["target"] = entry.Target
	result.RawData["verified"] = entry.Verified
	result.RawData["online"] = entry.Online
	result.RawData["match_type"] = matchType

	log.Info().
		Int64("phish_id", entry.PhishID).
		Str("target", entry.Target).
		Str("verified", entry.Verified).
		Str("match_type", matchType).
		Msg("[PhishTank] PHISHING FOUND")

	return result
}

// calculateConfidence calcula la confianza basada en el estado de verificación
func (c *PhishTankChecker) calculateConfidence(entry *PhishTankEntry) float64 {
	confidence := 0.70 // Base

	// Si está verificado, mayor confianza
	if entry.Verified == "yes" {
		confidence = 0.90
	}

	// Si está online actualmente, mayor confianza
	if entry.Online == "yes" {
		confidence += 0.05
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// LoadDB carga la base de datos desde el archivo JSON
func (c *PhishTankChecker) LoadDB() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(c.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DB file: %w", err)
	}
	defer file.Close()

	return c.parseJSON(file)
}

// DownloadDB descarga la base de datos actualizada
func (c *PhishTankChecker) DownloadDB(ctx context.Context) error {
	log.Info().Str("url", c.downloadURL).Msg("[PhishTank] Downloading database...")

	req, err := http.NewRequestWithContext(ctx, "GET", c.downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// PhishTank requiere User-Agent
	req.Header.Set("User-Agent", "phishtank/fy-analysis")

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

	log.Info().Int64("bytes", written).Msg("[PhishTank] Database downloaded")

	// Recargar DB
	file.Seek(0, 0)

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.parseJSON(file); err != nil {
		return fmt.Errorf("failed to parse downloaded DB: %w", err)
	}

	c.lastUpdate = time.Now()
	log.Info().
		Int("urls", len(c.urlDB)).
		Int("domains", len(c.domainDB)).
		Msg("[PhishTank] Database loaded")

	return nil
}

// parseJSON parsea el archivo JSON de PhishTank
func (c *PhishTankChecker) parseJSON(reader io.Reader) error {
	// Limpiar DBs existentes
	c.urlDB = make(map[string]*PhishTankEntry)
	c.domainDB = make(map[string]*PhishTankEntry)

	decoder := json.NewDecoder(reader)

	// El JSON es un array de entries
	var entries []PhishTankEntry
	if err := decoder.Decode(&entries); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	for i := range entries {
		entry := &entries[i]

		// Normalizar URL
		normalizedURL := strings.ToLower(entry.URL)

		// Indexar por URL
		c.urlDB[normalizedURL] = entry

		// Extraer e indexar por dominio
		domain := c.extractDomain(entry.URL)
		if domain != "" {
			c.domainDB[domain] = entry
		}
	}

	log.Debug().Int("entries", len(entries)).Msg("[PhishTank] JSON parsed")
	return nil
}

// extractDomain extrae el dominio de una URL
func (c *PhishTankChecker) extractDomain(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return strings.ToLower(parsed.Hostname())
}

// generateURLVariants genera variantes de una URL para búsqueda
func (c *PhishTankChecker) generateURLVariants(urlStr string) []string {
	variants := []string{}

	// Con/sin trailing slash
	if strings.HasSuffix(urlStr, "/") {
		variants = append(variants, strings.TrimSuffix(urlStr, "/"))
	} else {
		variants = append(variants, urlStr+"/")
	}

	// HTTP/HTTPS
	if strings.HasPrefix(urlStr, "http://") {
		variants = append(variants, strings.Replace(urlStr, "http://", "https://", 1))
	} else if strings.HasPrefix(urlStr, "https://") {
		variants = append(variants, strings.Replace(urlStr, "https://", "http://", 1))
	}

	return variants
}

// GetStats retorna estadísticas de la DB
func (c *PhishTankChecker) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"urls":        len(c.urlDB),
		"domains":     len(c.domainDB),
		"last_update": c.lastUpdate,
	}
}
