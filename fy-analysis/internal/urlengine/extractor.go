package urlengine

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/trackfy/fy-analysis/internal/checkers"
)

// Extractor extrae indicadores de una URL normalizada
type Extractor struct{}

// NewExtractor crea un nuevo extractor
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract extrae todos los indicadores de una URL normalizada
func (e *Extractor) Extract(normalized *NormalizeResult) *checkers.Indicators {
	indicators := &checkers.Indicators{}

	// URL a verificar (expandida si era shortener, sino la normalizada)
	targetURL := normalized.NormalizedURL
	if normalized.ExpandedURL != "" {
		targetURL = normalized.ExpandedURL
	}

	indicators.FullURL = targetURL

	// Extraer dominio de la URL final
	parsed, err := url.Parse(targetURL)
	if err == nil {
		indicators.Domain = strings.ToLower(parsed.Hostname())
	} else {
		indicators.Domain = normalized.Domain
	}

	// IP (puede ser de la URL expandida)
	if normalized.ExpandedURL != "" {
		// Re-resolver para la URL expandida si es necesario
		expandedParsed, _ := url.Parse(normalized.ExpandedURL)
		if expandedParsed != nil {
			indicators.Domain = strings.ToLower(expandedParsed.Hostname())
		}
	}
	indicators.IP = normalized.IP

	// Generar hashes
	indicators.URLHash = hashSHA256(indicators.FullURL)
	indicators.DomainHash = hashSHA256(indicators.Domain)

	log.Debug().
		Str("url", indicators.FullURL).
		Str("domain", indicators.Domain).
		Str("ip", indicators.IP).
		Str("url_hash", indicators.URLHash[:16]+"...").
		Msg("[Extractor] Indicators extracted")

	return indicators
}

// hashSHA256 genera el hash SHA256 de un string
func hashSHA256(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// ExtractDomainVariants genera variantes del dominio para búsqueda
func (e *Extractor) ExtractDomainVariants(domain string) []string {
	variants := []string{domain}

	// Sin www
	if strings.HasPrefix(domain, "www.") {
		variants = append(variants, domain[4:])
	} else {
		// Con www
		variants = append(variants, "www."+domain)
	}

	// Subdominios (para verificar dominio padre)
	parts := strings.Split(domain, ".")
	if len(parts) > 2 {
		// ejemplo.malware.com -> malware.com
		parentDomain := strings.Join(parts[1:], ".")
		variants = append(variants, parentDomain)
	}

	return variants
}

// ExtractURLVariants genera variantes de la URL para búsqueda
func (e *Extractor) ExtractURLVariants(fullURL string) []string {
	variants := []string{fullURL}

	parsed, err := url.Parse(fullURL)
	if err != nil {
		return variants
	}

	// Sin query string
	if parsed.RawQuery != "" {
		noQuery := *parsed
		noQuery.RawQuery = ""
		variants = append(variants, noQuery.String())
	}

	// Sin path (solo dominio)
	if parsed.Path != "" && parsed.Path != "/" {
		noPath := *parsed
		noPath.Path = "/"
		noPath.RawQuery = ""
		variants = append(variants, noPath.String())
	}

	// Con y sin trailing slash
	if strings.HasSuffix(fullURL, "/") {
		variants = append(variants, strings.TrimSuffix(fullURL, "/"))
	} else {
		variants = append(variants, fullURL+"/")
	}

	// HTTP/HTTPS variants
	if parsed.Scheme == "http" {
		httpsVariant := *parsed
		httpsVariant.Scheme = "https"
		variants = append(variants, httpsVariant.String())
	} else if parsed.Scheme == "https" {
		httpVariant := *parsed
		httpVariant.Scheme = "http"
		variants = append(variants, httpVariant.String())
	}

	return variants
}
