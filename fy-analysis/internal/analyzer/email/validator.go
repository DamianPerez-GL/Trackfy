package email

import (
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/trackfy/fy-analysis/internal/models"
)

// Analyzer maneja el análisis de emails
type Analyzer struct {
	disposableDomains map[string]bool
	freemailDomains   map[string]bool
	blacklistedDomains map[string]bool
}

// NewAnalyzer crea una nueva instancia del analizador de emails
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		disposableDomains:  loadDisposableDomains(),
		freemailDomains:    loadFreemailDomains(),
		blacklistedDomains: loadBlacklistedDomains(),
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Analyze realiza el análisis completo de un email
func (a *Analyzer) Analyze(email string, context string) models.EmailAnalysisResponse {
	email = strings.ToLower(strings.TrimSpace(email))

	response := models.EmailAnalysisResponse{
		Email: email,
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

	// Validar formato
	if !isValidFormat(email) {
		response.Analysis.IsMalicious = true
		response.Analysis.ThreatLevel = models.ThreatLevelMedium
		response.Analysis.Confidence = 0.9
		response.Analysis.Reasons = append(response.Analysis.Reasons, "Formato de email inválido")
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeSuspicious)
		return response
	}

	// Extraer dominio
	parts := strings.Split(email, "@")
	domain := parts[1]

	// Obtener información del dominio
	domainInfo := a.analyzeDomain(domain)
	response.DomainInfo = domainInfo

	// Verificar si es dominio desechable
	if a.isDisposable(domain) {
		response.Analysis.ThreatLevel = models.ThreatLevelMedium
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeDisposable)
		response.Analysis.Reasons = append(response.Analysis.Reasons, "Email de dominio desechable/temporal")
		response.Analysis.Confidence = 0.95
		response.Recommendations = append(response.Recommendations, "Solicitar un email corporativo o personal permanente")
	}

	// Verificar lista negra
	if a.isBlacklisted(domain) {
		response.Analysis.IsMalicious = true
		response.Analysis.ThreatLevel = models.ThreatLevelHigh
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeSpam)
		response.Analysis.Reasons = append(response.Analysis.Reasons, "Dominio en lista negra de spam/phishing")
		response.Analysis.Confidence = 0.99
		response.Recommendations = append(response.Recommendations, "No interactuar con este remitente")
	}

	// Análisis heurístico del email
	heuristicScore := a.heuristicAnalysis(email, context)
	if heuristicScore > 0.7 {
		response.Analysis.IsMalicious = true
		response.Analysis.ThreatLevel = models.ThreatLevelHigh
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypePhishing)
		response.Analysis.Reasons = append(response.Analysis.Reasons, "Patrones sospechosos detectados en el email")
		if response.Analysis.Confidence < heuristicScore {
			response.Analysis.Confidence = heuristicScore
		}
	} else if heuristicScore > 0.4 {
		if response.Analysis.ThreatLevel == models.ThreatLevelSafe {
			response.Analysis.ThreatLevel = models.ThreatLevelLow
		}
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeSuspicious)
		response.Analysis.Reasons = append(response.Analysis.Reasons, "Algunos patrones inusuales detectados")
	}

	// Verificar registros MX
	if !domainInfo.HasMXRecords {
		response.Analysis.ThreatLevel = models.ThreatLevelMedium
		response.Analysis.Reasons = append(response.Analysis.Reasons, "El dominio no tiene registros MX válidos")
		response.Analysis.Confidence = 0.8
	}

	// Si no hay amenazas, marcar como seguro
	if len(response.Analysis.ThreatTypes) == 0 {
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeNone)
		response.Analysis.Confidence = 0.85
	}

	return response
}

func isValidFormat(email string) bool {
	return emailRegex.MatchString(email)
}

func (a *Analyzer) isDisposable(domain string) bool {
	return a.disposableDomains[domain]
}

func (a *Analyzer) isBlacklisted(domain string) bool {
	return a.blacklistedDomains[domain]
}

func (a *Analyzer) isFreemail(domain string) bool {
	return a.freemailDomains[domain]
}

func (a *Analyzer) analyzeDomain(domain string) *models.DomainInfo {
	info := &models.DomainInfo{
		Domain:       domain,
		IsDisposable: a.isDisposable(domain),
		IsFreemail:   a.isFreemail(domain),
		HasMXRecords: hasMXRecords(domain),
	}
	return info
}

func hasMXRecords(domain string) bool {
	mxRecords, err := net.LookupMX(domain)
	return err == nil && len(mxRecords) > 0
}

func (a *Analyzer) heuristicAnalysis(email string, context string) float64 {
	score := 0.0

	localPart := strings.Split(email, "@")[0]

	// Muchos números en la parte local
	numCount := 0
	for _, c := range localPart {
		if c >= '0' && c <= '9' {
			numCount++
		}
	}
	if float64(numCount)/float64(len(localPart)) > 0.5 {
		score += 0.2
	}

	// Patrones sospechosos en la parte local
	suspiciousPatterns := []string{
		"admin", "support", "security", "helpdesk", "verify", "confirm",
		"update", "account", "banking", "secure", "login", "password",
	}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(localPart, pattern) {
			score += 0.15
		}
	}

	// Analizar contexto si existe
	if context != "" {
		contextLower := strings.ToLower(context)
		urgentPatterns := []string{
			"urgente", "urgent", "inmediato", "immediate", "acción requerida",
			"action required", "suspendido", "suspended", "bloqueado", "blocked",
			"verify your", "confirma tu", "actualiza tu", "update your",
		}
		for _, pattern := range urgentPatterns {
			if strings.Contains(contextLower, pattern) {
				score += 0.1
			}
		}
	}

	// Limitar score máximo
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func loadDisposableDomains() map[string]bool {
	return map[string]bool{
		"tempmail.com":      true,
		"throwaway.email":   true,
		"guerrillamail.com": true,
		"10minutemail.com":  true,
		"mailinator.com":    true,
		"yopmail.com":       true,
		"trashmail.com":     true,
		"fakeinbox.com":     true,
		"tempail.com":       true,
		"getnada.com":       true,
		"temp-mail.org":     true,
		"disposablemail.com": true,
		"maildrop.cc":       true,
		"dispostable.com":   true,
		"sharklasers.com":   true,
	}
}

func loadFreemailDomains() map[string]bool {
	return map[string]bool{
		"gmail.com":     true,
		"yahoo.com":     true,
		"hotmail.com":   true,
		"outlook.com":   true,
		"aol.com":       true,
		"icloud.com":    true,
		"protonmail.com": true,
		"mail.com":      true,
		"zoho.com":      true,
		"yandex.com":    true,
	}
}

func loadBlacklistedDomains() map[string]bool {
	return map[string]bool{
		"malicious-domain.com":   true,
		"phishing-site.net":      true,
		"scam-emails.org":        true,
		"fraud-domain.com":       true,
		"spam-sender.net":        true,
		"fake-bank-alert.com":    true,
		"lottery-winner.net":     true,
		"prince-nigeria.com":     true,
		"free-iphone-winner.com": true,
	}
}
