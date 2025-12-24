package checkers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

// URLScanChecker verifica URLs contra URLScan.io API
// Documentación: https://urlscan.io/docs/api/
type URLScanChecker struct {
	enabled    bool
	weight     float64
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// URLScanSearchResponse respuesta de búsqueda
type URLScanSearchResponse struct {
	Results []URLScanResult `json:"results"`
	Total   int             `json:"total"`
}

// URLScanResult resultado individual
type URLScanResult struct {
	Task       URLScanTask `json:"task"`
	Stats      URLScanStats `json:"stats,omitempty"`
	Page       URLScanPage  `json:"page"`
	Verdicts   URLScanVerdicts `json:"verdicts,omitempty"`
}

// URLScanTask información de la tarea
type URLScanTask struct {
	UUID       string `json:"uuid"`
	URL        string `json:"url"`
	Domain     string `json:"domain"`
	Time       string `json:"time"`
	Visibility string `json:"visibility"`
}

// URLScanStats estadísticas
type URLScanStats struct {
	Malicious    int `json:"malicious"`
	SecureStatus int `json:"secureStatus"`
	UniqIPs      int `json:"uniqIPs"`
}

// URLScanPage información de la página
type URLScanPage struct {
	URL     string `json:"url"`
	Domain  string `json:"domain"`
	Country string `json:"country"`
	Server  string `json:"server"`
	IP      string `json:"ip"`
}

// URLScanVerdicts veredictos de la página
type URLScanVerdicts struct {
	Overall    URLScanVerdict `json:"overall"`
	Engines    URLScanVerdict `json:"engines"`
	Community  URLScanVerdict `json:"community"`
}

// URLScanVerdict veredicto individual
type URLScanVerdict struct {
	Score      int      `json:"score"`
	Malicious  bool     `json:"malicious"`
	Categories []string `json:"categories,omitempty"`
}

// NewURLScanChecker crea un nuevo checker de URLScan.io
func NewURLScanChecker(apiKey string) *URLScanChecker {
	checker := &URLScanChecker{
		enabled: apiKey != "",
		weight:  0.10,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://urlscan.io/api/v1",
	}

	if checker.enabled {
		log.Info().Msg("[URLScan] Checker enabled with API key")
	} else {
		log.Warn().Msg("[URLScan] Checker disabled - no API key provided")
	}

	return checker
}

// Name retorna el nombre del checker
func (c *URLScanChecker) Name() string {
	return "urlscan"
}

// Weight retorna el peso del checker
func (c *URLScanChecker) Weight() float64 {
	return c.weight
}

// IsEnabled indica si el checker está habilitado
func (c *URLScanChecker) IsEnabled() bool {
	return c.enabled
}

// SupportedTypes retorna los tipos soportados (solo URLs)
func (c *URLScanChecker) SupportedTypes() []InputType {
	return []InputType{InputTypeURL}
}

// Check verifica los indicadores contra URLScan.io
// Usa la API de búsqueda para encontrar scans previos de la URL/dominio
func (c *URLScanChecker) Check(ctx context.Context, indicators *Indicators) (*CheckResult, error) {
	result := &CheckResult{
		Source:  c.Name(),
		Found:   false,
		RawData: make(map[string]interface{}),
	}

	if !c.enabled {
		return result, fmt.Errorf("URLScan API key not configured")
	}

	log.Debug().
		Str("url", indicators.FullURL).
		Str("domain", indicators.Domain).
		Msg("[URLScan] Checking URL")

	// Buscar por dominio (más probable encontrar resultados)
	searchResult, err := c.searchDomain(ctx, indicators.Domain)
	if err != nil {
		log.Warn().Err(err).Msg("[URLScan] Search failed")
		return result, err
	}

	if searchResult == nil || len(searchResult.Results) == 0 {
		log.Debug().
			Str("domain", indicators.Domain).
			Msg("[URLScan] No previous scans found for domain")
		return result, nil
	}

	// Analizar resultados
	for _, scan := range searchResult.Results {
		if c.isMalicious(scan) {
			result.Found = true
			result.ThreatType = c.determineThreatType(scan)
			result.Confidence = c.calculateConfidence(scan)
			result.Tags = c.extractTags(scan)
			result.RawData["scan_uuid"] = scan.Task.UUID
			result.RawData["scan_url"] = "https://urlscan.io/result/" + scan.Task.UUID
			result.RawData["scan_time"] = scan.Task.Time
			result.RawData["page_ip"] = scan.Page.IP
			result.RawData["page_country"] = scan.Page.Country

			log.Info().
				Str("domain", indicators.Domain).
				Str("uuid", scan.Task.UUID).
				Str("threat_type", result.ThreatType).
				Msg("[URLScan] MALICIOUS FOUND")

			return result, nil
		}
	}

	log.Debug().
		Str("domain", indicators.Domain).
		Int("scans_checked", len(searchResult.Results)).
		Msg("[URLScan] No malicious verdicts found")

	return result, nil
}

// searchDomain busca scans previos de un dominio
func (c *URLScanChecker) searchDomain(ctx context.Context, domain string) (*URLScanSearchResponse, error) {
	// Construir query: domain:example.com
	query := fmt.Sprintf("domain:%s", domain)
	params := url.Values{}
	params.Add("q", query)
	params.Add("size", "10") // Limitar resultados

	requestURL := c.baseURL + "/search/?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Debug().
		Int("status", resp.StatusCode).
		Int("body_len", len(body)).
		Msg("[URLScan] Search response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var searchResp URLScanSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &searchResp, nil
}

// isMalicious determina si un scan indica contenido malicioso
func (c *URLScanChecker) isMalicious(scan URLScanResult) bool {
	// Verificar veredictos
	if scan.Verdicts.Overall.Malicious {
		return true
	}

	if scan.Verdicts.Engines.Malicious {
		return true
	}

	if scan.Verdicts.Community.Malicious {
		return true
	}

	// Verificar stats
	if scan.Stats.Malicious > 0 {
		return true
	}

	return false
}

// determineThreatType determina el tipo de amenaza
func (c *URLScanChecker) determineThreatType(scan URLScanResult) string {
	// Buscar en categorías
	categories := scan.Verdicts.Overall.Categories
	for _, cat := range categories {
		switch cat {
		case "phishing":
			return ThreatTypePhishing
		case "malware":
			return ThreatTypeMalware
		case "spam":
			return ThreatTypeSpam
		}
	}

	// Default basado en que es malicioso
	return ThreatTypeUnknown
}

// calculateConfidence calcula la confianza basada en el scan
func (c *URLScanChecker) calculateConfidence(scan URLScanResult) float64 {
	confidence := 0.60 // Base

	// Si múltiples engines coinciden
	if scan.Verdicts.Engines.Malicious && scan.Verdicts.Community.Malicious {
		confidence = 0.85
	} else if scan.Verdicts.Overall.Malicious {
		confidence = 0.75
	}

	// Boost por score alto
	if scan.Verdicts.Overall.Score > 50 {
		confidence += 0.10
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// extractTags extrae tags del scan
func (c *URLScanChecker) extractTags(scan URLScanResult) []string {
	tags := []string{}

	tags = append(tags, scan.Verdicts.Overall.Categories...)

	if scan.Page.Server != "" {
		tags = append(tags, "server:"+scan.Page.Server)
	}

	if scan.Page.Country != "" {
		tags = append(tags, "country:"+scan.Page.Country)
	}

	return tags
}

// SubmitScan envía una URL para escaneo (opcional, para URLs no encontradas)
func (c *URLScanChecker) SubmitScan(ctx context.Context, targetURL string) (string, error) {
	if !c.enabled {
		return "", fmt.Errorf("URLScan API key not configured")
	}

	// Esta función es para uso futuro - submite una URL para escaneo
	// POST /api/v1/scan/
	// Por ahora solo buscamos scans existentes

	return "", fmt.Errorf("submit scan not implemented")
}
