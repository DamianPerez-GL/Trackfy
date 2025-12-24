package checkers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// WebRiskChecker verifica URLs contra Google Web Risk API
// Documentación: https://cloud.google.com/web-risk/docs/lookup-api
type WebRiskChecker struct {
	enabled    bool
	weight     float64
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// WebRiskResponse respuesta de la API de Google Web Risk
type WebRiskResponse struct {
	Threat *WebRiskThreat `json:"threat,omitempty"`
}

// WebRiskThreat detalle de amenaza
type WebRiskThreat struct {
	ThreatTypes      []string `json:"threatTypes"`
	ExpireTime       string   `json:"expireTime"`
	ThreatEntryTypes []string `json:"threatEntryTypes,omitempty"`
}

// NewWebRiskChecker crea un nuevo checker de Google Web Risk
func NewWebRiskChecker(apiKey string) *WebRiskChecker {
	checker := &WebRiskChecker{
		enabled: apiKey != "",
		weight:  0.30,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL: "https://webrisk.googleapis.com/v1/uris:search",
	}

	if checker.enabled {
		log.Info().Msg("[WebRisk] Checker enabled with API key")
	} else {
		log.Warn().Msg("[WebRisk] Checker disabled - no API key provided")
	}

	return checker
}

// Name retorna el nombre del checker
func (c *WebRiskChecker) Name() string {
	return "webrisk"
}

// Weight retorna el peso del checker
func (c *WebRiskChecker) Weight() float64 {
	return c.weight
}

// IsEnabled indica si el checker está habilitado
func (c *WebRiskChecker) IsEnabled() bool {
	return c.enabled
}

// SupportedTypes retorna los tipos soportados (solo URLs)
func (c *WebRiskChecker) SupportedTypes() []InputType {
	return []InputType{InputTypeURL}
}

// Check verifica los indicadores contra Google Web Risk
func (c *WebRiskChecker) Check(ctx context.Context, indicators *Indicators) (*CheckResult, error) {
	result := &CheckResult{
		Source:  c.Name(),
		Found:   false,
		RawData: make(map[string]interface{}),
	}

	if !c.enabled {
		return result, fmt.Errorf("Web Risk API key not configured")
	}

	log.Debug().
		Str("url", indicators.FullURL).
		Msg("[WebRisk] Checking URL")

	// Construir URL de la API
	// GET https://webrisk.googleapis.com/v1/uris:search?key=API_KEY&threatTypes=MALWARE&threatTypes=SOCIAL_ENGINEERING&uri=URL
	params := url.Values{}
	params.Add("key", c.apiKey)
	params.Add("uri", indicators.FullURL)
	// Tipos de amenazas a verificar
	params.Add("threatTypes", "MALWARE")
	params.Add("threatTypes", "SOCIAL_ENGINEERING")
	params.Add("threatTypes", "UNWANTED_SOFTWARE")

	requestURL := c.baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return result, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("[WebRisk] API request failed")
		return result, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response: %w", err)
	}

	log.Debug().
		Int("status", resp.StatusCode).
		Str("body", string(body)).
		Msg("[WebRisk] API response")

	// Verificar status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 400 {
			// Bad request - posiblemente URL inválida, no es un error crítico
			log.Debug().Str("url", indicators.FullURL).Msg("[WebRisk] URL not checkable (400)")
			return result, nil
		}
		return result, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parsear respuesta
	var webRiskResp WebRiskResponse
	if err := json.Unmarshal(body, &webRiskResp); err != nil {
		return result, fmt.Errorf("failed to parse response: %w", err)
	}

	// Si no hay threat, URL es segura
	if webRiskResp.Threat == nil || len(webRiskResp.Threat.ThreatTypes) == 0 {
		log.Debug().
			Str("url", indicators.FullURL).
			Msg("[WebRisk] URL is clean")
		return result, nil
	}

	// ¡Amenaza encontrada!
	result.Found = true
	result.ThreatType = c.mapThreatType(webRiskResp.Threat.ThreatTypes)
	result.Confidence = 0.95 // Google tiene alta confianza
	result.Tags = webRiskResp.Threat.ThreatTypes
	result.RawData["threat_types"] = webRiskResp.Threat.ThreatTypes
	result.RawData["expire_time"] = webRiskResp.Threat.ExpireTime

	log.Info().
		Str("url", indicators.FullURL).
		Strs("threat_types", webRiskResp.Threat.ThreatTypes).
		Msg("[WebRisk] THREAT FOUND")

	return result, nil
}

// mapThreatType mapea los tipos de amenaza de Google a nuestros tipos
func (c *WebRiskChecker) mapThreatType(googleTypes []string) string {
	// Priorizar por severidad
	for _, t := range googleTypes {
		switch strings.ToUpper(t) {
		case "MALWARE":
			return ThreatTypeMalware
		case "SOCIAL_ENGINEERING":
			return ThreatTypeSocialEng
		case "UNWANTED_SOFTWARE":
			return ThreatTypeUnwanted
		case "POTENTIALLY_HARMFUL_APPLICATION":
			return ThreatTypePotentiallyHarmful
		}
	}

	if len(googleTypes) > 0 {
		return ThreatTypeUnknown
	}
	return ""
}

// TestAPIKey verifica si la API key es válida
func (c *WebRiskChecker) TestAPIKey(ctx context.Context) error {
	if !c.enabled {
		return fmt.Errorf("API key not configured")
	}

	// Probar con una URL conocida como segura
	testURL := "https://www.google.com"
	params := url.Values{}
	params.Add("key", c.apiKey)
	params.Add("uri", testURL)
	params.Add("threatTypes", "MALWARE")

	requestURL := c.baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	log.Info().Msg("[WebRisk] API key is valid")
	return nil
}
