package url

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/trackfy/fy-analysis/internal/models"
)

// Analyzer maneja el análisis de URLs
type Analyzer struct {
	maliciousDomains  map[string]bool
	shortenerDomains  map[string]bool
	suspiciousTLDs    map[string]bool
	phishingKeywords  []string
}

// NewAnalyzer crea una nueva instancia del analizador de URLs
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		maliciousDomains: loadMaliciousDomains(),
		shortenerDomains: loadShortenerDomains(),
		suspiciousTLDs:   loadSuspiciousTLDs(),
		phishingKeywords: loadPhishingKeywords(),
	}
}

// Analyze realiza el análisis completo de una URL
func (a *Analyzer) Analyze(rawURL string, context string) models.URLAnalysisResponse {
	rawURL = strings.TrimSpace(rawURL)

	response := models.URLAnalysisResponse{
		URL: rawURL,
		Analysis: models.AnalysisResult{
			IsMalicious: false,
			ThreatLevel: models.ThreatLevelSafe,
			ThreatTypes: []models.ThreatType{},
			Confidence:  0.0,
			Reasons:     []string{},
			AnalyzedAt:  time.Now().UTC(),
		},
		Recommendations: []string{},
	}

	// Parsear URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		response.Analysis.ThreatLevel = models.ThreatLevelMedium
		response.Analysis.Reasons = append(response.Analysis.Reasons, "URL con formato inválido")
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeSuspicious)
		response.Analysis.Confidence = 0.8
		return response
	}

	// Si no tiene scheme, añadir http
	if parsedURL.Scheme == "" {
		rawURL = "http://" + rawURL
		parsedURL, _ = url.Parse(rawURL)
	}

	// Extraer información de la URL
	urlInfo := a.extractURLInfo(parsedURL)
	response.URLInfo = urlInfo

	// Verificar si es un acortador
	if urlInfo.IsShortened {
		response.Analysis.ThreatLevel = models.ThreatLevelLow
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeSuspicious)
		response.Analysis.Reasons = append(response.Analysis.Reasons, "URL acortada - destino desconocido")
		response.Analysis.Confidence = 0.6
		response.Recommendations = append(response.Recommendations, "Expandir la URL antes de hacer clic")
	}

	// Verificar dominio malicioso
	domain := strings.ToLower(parsedURL.Hostname())
	if a.isMaliciousDomain(domain) {
		response.Analysis.IsMalicious = true
		response.Analysis.ThreatLevel = models.ThreatLevelCritical
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeMalware)
		response.Analysis.Reasons = append(response.Analysis.Reasons, "Dominio conocido como malicioso")
		response.Analysis.Confidence = 0.99
		response.Recommendations = append(response.Recommendations, "NO visitar esta URL bajo ninguna circunstancia")
		return response
	}

	// Análisis heurístico
	heuristicResult := a.heuristicAnalysis(parsedURL, urlInfo)

	if heuristicResult.score > 0.7 {
		response.Analysis.IsMalicious = true
		response.Analysis.ThreatLevel = models.ThreatLevelHigh
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypePhishing)
		response.Analysis.Reasons = append(response.Analysis.Reasons, heuristicResult.reasons...)
		response.Analysis.Confidence = heuristicResult.score
		response.Recommendations = append(response.Recommendations, "Verificar la autenticidad del sitio antes de ingresar datos")
	} else if heuristicResult.score > 0.4 {
		response.Analysis.ThreatLevel = models.ThreatLevelMedium
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeSuspicious)
		response.Analysis.Reasons = append(response.Analysis.Reasons, heuristicResult.reasons...)
		response.Analysis.Confidence = heuristicResult.score
		response.Recommendations = append(response.Recommendations, "Proceder con precaución")
	}

	// Verificar HTTPS
	if parsedURL.Scheme != "https" {
		response.Analysis.Reasons = append(response.Analysis.Reasons, "No usa conexión segura (HTTPS)")
		if response.Analysis.ThreatLevel == models.ThreatLevelSafe {
			response.Analysis.ThreatLevel = models.ThreatLevelLow
		}
		response.Recommendations = append(response.Recommendations, "No ingresar información sensible en sitios sin HTTPS")
	}

	// Si no hay amenazas
	if len(response.Analysis.ThreatTypes) == 0 {
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeNone)
		response.Analysis.Confidence = 0.85
	}

	return response
}

func (a *Analyzer) extractURLInfo(parsedURL *url.URL) *models.URLInfo {
	domain := strings.ToLower(parsedURL.Hostname())

	info := &models.URLInfo{
		Domain:              domain,
		Scheme:              parsedURL.Scheme,
		Path:                parsedURL.Path,
		HasSuspiciousParams: a.hasSuspiciousParams(parsedURL),
		IsShortened:         a.shortenerDomains[domain],
		SSLValid:            parsedURL.Scheme == "https",
	}

	return info
}

func (a *Analyzer) hasSuspiciousParams(parsedURL *url.URL) bool {
	suspiciousParams := []string{"redirect", "url", "link", "goto", "return", "next", "target"}
	query := strings.ToLower(parsedURL.RawQuery)

	for _, param := range suspiciousParams {
		if strings.Contains(query, param+"=") {
			return true
		}
	}
	return false
}

func (a *Analyzer) isMaliciousDomain(domain string) bool {
	// Verificar dominio exacto
	if a.maliciousDomains[domain] {
		return true
	}

	// Verificar subdominios
	parts := strings.Split(domain, ".")
	for i := range parts {
		subdomain := strings.Join(parts[i:], ".")
		if a.maliciousDomains[subdomain] {
			return true
		}
	}

	return false
}

type heuristicResult struct {
	score   float64
	reasons []string
}

func (a *Analyzer) heuristicAnalysis(parsedURL *url.URL, urlInfo *models.URLInfo) heuristicResult {
	result := heuristicResult{
		score:   0.0,
		reasons: []string{},
	}

	domain := urlInfo.Domain
	fullURL := parsedURL.String()

	// 1. Verificar TLD sospechoso
	tld := extractTLD(domain)
	if a.suspiciousTLDs[tld] {
		result.score += 0.2
		result.reasons = append(result.reasons, "TLD frecuentemente usado en sitios maliciosos")
	}

	// 2. Verificar IP en lugar de dominio
	ipRegex := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	if ipRegex.MatchString(domain) {
		result.score += 0.4
		result.reasons = append(result.reasons, "URL usa dirección IP en lugar de dominio")
	}

	// 3. Verificar caracteres sospechosos en el dominio
	if strings.Contains(domain, "-") && countOccurrences(domain, "-") > 2 {
		result.score += 0.15
		result.reasons = append(result.reasons, "Dominio con múltiples guiones (posible typosquatting)")
	}

	// 4. Buscar palabras clave de phishing
	for _, keyword := range a.phishingKeywords {
		if strings.Contains(strings.ToLower(fullURL), keyword) {
			result.score += 0.15
			result.reasons = append(result.reasons, "Contiene palabras clave asociadas a phishing: "+keyword)
			break
		}
	}

	// 5. Verificar longitud excesiva del dominio
	if len(domain) > 50 {
		result.score += 0.2
		result.reasons = append(result.reasons, "Dominio excesivamente largo")
	}

	// 6. Verificar subdominios excesivos
	if strings.Count(domain, ".") > 3 {
		result.score += 0.25
		result.reasons = append(result.reasons, "Demasiados subdominios")
	}

	// 7. Verificar caracteres unicode/homógrafos
	if hasHomographChars(domain) {
		result.score += 0.5
		result.reasons = append(result.reasons, "Posible ataque de homógrafos (caracteres similares)")
	}

	// 8. Verificar parámetros sospechosos
	if urlInfo.HasSuspiciousParams {
		result.score += 0.15
		result.reasons = append(result.reasons, "Parámetros de redirección sospechosos")
	}

	// 9. Verificar @ en URL (técnica de ofuscación)
	if strings.Contains(fullURL, "@") {
		result.score += 0.4
		result.reasons = append(result.reasons, "URL contiene @ (técnica de ofuscación)")
	}

	// 10. Verificar codificación sospechosa
	if strings.Contains(fullURL, "%") && strings.Count(fullURL, "%") > 5 {
		result.score += 0.2
		result.reasons = append(result.reasons, "Codificación URL excesiva")
	}

	// Limitar score máximo
	if result.score > 1.0 {
		result.score = 1.0
	}

	return result
}

func extractTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func countOccurrences(s, substr string) int {
	return strings.Count(s, substr)
}

func hasHomographChars(domain string) bool {
	// Detectar caracteres cirílicos u otros que parecen latinos
	for _, r := range domain {
		// Caracteres fuera del rango ASCII básico (excepto guión y punto)
		if r > 127 {
			return true
		}
	}
	return false
}

func loadMaliciousDomains() map[string]bool {
	return map[string]bool{
		"malware-distribution.com":  true,
		"phishing-site.net":         true,
		"evil-domain.org":           true,
		"fake-login.com":            true,
		"credential-stealer.net":    true,
		"ransomware-download.com":   true,
		"scam-prize.org":            true,
		"fake-antivirus.com":        true,
		"drive-by-download.net":     true,
		"cryptominer-inject.com":    true,
	}
}

func loadShortenerDomains() map[string]bool {
	return map[string]bool{
		"bit.ly":      true,
		"tinyurl.com": true,
		"t.co":        true,
		"goo.gl":      true,
		"ow.ly":       true,
		"is.gd":       true,
		"buff.ly":     true,
		"adf.ly":      true,
		"bl.ink":      true,
		"lnkd.in":     true,
		"rebrand.ly":  true,
		"short.io":    true,
	}
}

func loadSuspiciousTLDs() map[string]bool {
	return map[string]bool{
		"tk":      true,
		"ml":      true,
		"ga":      true,
		"cf":      true,
		"gq":      true,
		"xyz":     true,
		"top":     true,
		"work":    true,
		"click":   true,
		"link":    true,
		"download": true,
	}
}

func loadPhishingKeywords() []string {
	return []string{
		"login", "signin", "account", "verify", "secure", "update",
		"confirm", "password", "credential", "bank", "paypal", "apple",
		"microsoft", "google", "facebook", "amazon", "netflix", "support",
		"helpdesk", "suspended", "locked", "unusual", "activity",
	}
}
