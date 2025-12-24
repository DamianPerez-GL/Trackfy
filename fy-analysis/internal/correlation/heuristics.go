package correlation

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/trackfy/fy-analysis/internal/checkers"
)

// HeuristicEngine motor de análisis heurístico
type HeuristicEngine struct {
	// Dominios oficiales de bancos españoles
	spanishBanks map[string][]string
	// Dominios oficiales de telcos españolas
	spanishTelcos map[string][]string
	// TLDs sospechosos
	suspiciousTLDs map[string]int // TLD -> puntos de riesgo
	// Prefijos premium españoles
	premiumPrefixes []string
}

// NewHeuristicEngine crea un nuevo motor heurístico
func NewHeuristicEngine() *HeuristicEngine {
	return &HeuristicEngine{
		spanishBanks: map[string][]string{
			"bbva":       {"bbva.es", "bbva.com"},
			"santander":  {"santander.es", "santander.com", "bancosantander.es"},
			"caixabank":  {"caixabank.es", "caixabank.com", "lacaixa.es"},
			"sabadell":   {"bancsabadell.com", "sabadell.com"},
			"ing":        {"ing.es", "ingdirect.es"},
			"openbank":   {"openbank.es", "openbank.com"},
			"bankinter":  {"bankinter.com", "bankinter.es"},
			"evo":        {"evobanco.com"},
			"unicaja":    {"unicaja.es", "unicajabanco.es"},
			"kutxabank":  {"kutxabank.es", "kutxabank.com"},
			"abanca":     {"abanca.com", "abanca.es"},
			"ibercaja":   {"ibercaja.es"},
			"liberbank":  {"liberbank.es"},
			"deutsche":   {"deutsche-bank.es"},
			"bankia":     {"bankia.es", "bankia.com"}, // Ahora CaixaBank
		},
		spanishTelcos: map[string][]string{
			"movistar":  {"movistar.es", "movistar.com", "telefonica.es"},
			"vodafone":  {"vodafone.es", "vodafone.com"},
			"orange":    {"orange.es", "orange.com"},
			"yoigo":     {"yoigo.com", "yoigo.es"},
			"masmovil":  {"masmovil.es", "masmovil.com", "masmovil.com.es"},
			"pepephone": {"pepephone.com"},
			"lowi":      {"lowi.es"},
			"digi":      {"digimobil.es"},
			"simyo":     {"simyo.es"},
			"finetwork": {"finetwork.com"},
		},
		suspiciousTLDs: map[string]int{
			"xyz":     15,
			"top":     15,
			"tk":      20,
			"ml":      20,
			"ga":      20,
			"cf":      20,
			"gq":      20,
			"buzz":    10,
			"click":   15,
			"link":    10,
			"work":    10,
			"rest":    15,
			"cam":     15,
			"icu":     15,
			"monster": 15,
			"uno":     10,
		},
		premiumPrefixes: []string{"803", "806", "807", "905", "907"},
	}
}

// HeuristicResult resultado del análisis heurístico
type HeuristicResult struct {
	Score       int      // Puntos de riesgo acumulados
	Reasons     []string // Razones en español
	Flags       []string // Flags técnicos
	ContextHits []string // Coincidencias de contexto
}

// Analyze ejecuta el análisis heurístico completo
func (h *HeuristicEngine) Analyze(ctx context.Context, indicators *checkers.Indicators, analysisCtx *checkers.AnalysisContext) *HeuristicResult {
	result := &HeuristicResult{
		Score:       0,
		Reasons:     []string{},
		Flags:       []string{},
		ContextHits: []string{},
	}

	switch indicators.InputType {
	case checkers.InputTypeURL:
		h.analyzeURL(indicators, analysisCtx, result)
	case checkers.InputTypeEmail:
		h.analyzeEmail(indicators, analysisCtx, result)
	case checkers.InputTypePhone:
		h.analyzePhone(indicators, analysisCtx, result)
	}

	log.Debug().
		Int("score", result.Score).
		Strs("flags", result.Flags).
		Msg("[Heuristics] Analysis completed")

	return result
}

// analyzeURL analiza heurísticas específicas de URLs
func (h *HeuristicEngine) analyzeURL(indicators *checkers.Indicators, ctx *checkers.AnalysisContext, result *HeuristicResult) {
	domain := strings.ToLower(indicators.Domain)

	// 1. TLD sospechoso
	if points, isSuspicious := h.suspiciousTLDs[indicators.TLD]; isSuspicious {
		result.Score += points
		result.Flags = append(result.Flags, "suspicious_tld")
		result.Reasons = append(result.Reasons, fmt.Sprintf("El dominio usa un TLD sospechoso (.%s)", indicators.TLD))
	}

	// 2. Typosquatting de bancos
	for brand, domains := range h.spanishBanks {
		if h.isTyposquatting(domain, brand, domains) {
			result.Score += 35
			result.Flags = append(result.Flags, "typosquatting_bank")
			result.Reasons = append(result.Reasons, fmt.Sprintf("El dominio parece imitar a %s (posible suplantación)", brand))
			break
		}
	}

	// 3. Typosquatting de telcos
	for brand, domains := range h.spanishTelcos {
		if h.isTyposquatting(domain, brand, domains) {
			result.Score += 30
			result.Flags = append(result.Flags, "typosquatting_telco")
			result.Reasons = append(result.Reasons, fmt.Sprintf("El dominio parece imitar a %s (posible suplantación)", brand))
			break
		}
	}

	// 4. Context mismatch: dice ser X pero dominio no coincide
	if ctx != nil && ctx.ClaimedSender != "" {
		claimedLower := strings.ToLower(ctx.ClaimedSender)
		matched := false

		// Buscar en bancos
		for brand, domains := range h.spanishBanks {
			if strings.Contains(claimedLower, brand) {
				// Dice ser este banco, verificar si el dominio es legítimo
				for _, legitDomain := range domains {
					if domain == legitDomain || strings.HasSuffix(domain, "."+legitDomain) {
						matched = true
						break
					}
				}
				if !matched {
					result.Score += 40
					result.Flags = append(result.Flags, "context_mismatch")
					result.Reasons = append(result.Reasons, fmt.Sprintf("El remitente dice ser %s pero el dominio no coincide con los oficiales", brand))
					result.ContextHits = append(result.ContextHits, "claimed_bank_mismatch")
				}
				break
			}
		}

		// Buscar en telcos si no hubo match de banco
		if !matched {
			for brand, domains := range h.spanishTelcos {
				if strings.Contains(claimedLower, brand) {
					for _, legitDomain := range domains {
						if domain == legitDomain || strings.HasSuffix(domain, "."+legitDomain) {
							matched = true
							break
						}
					}
					if !matched {
						result.Score += 35
						result.Flags = append(result.Flags, "context_mismatch")
						result.Reasons = append(result.Reasons, fmt.Sprintf("El remitente dice ser %s pero el dominio no coincide", brand))
						result.ContextHits = append(result.ContextHits, "claimed_telco_mismatch")
					}
					break
				}
			}
		}
	}

	// 5. URL con IP directa (sospechoso)
	if indicators.IP != "" && indicators.Domain == indicators.IP {
		result.Score += 25
		result.Flags = append(result.Flags, "direct_ip")
		result.Reasons = append(result.Reasons, "La URL usa una dirección IP directa en lugar de un dominio")
	}

	// 6. Subdominios excesivos (más de 3 niveles)
	parts := strings.Split(domain, ".")
	if len(parts) > 4 {
		result.Score += 15
		result.Flags = append(result.Flags, "excessive_subdomains")
		result.Reasons = append(result.Reasons, "El dominio tiene muchos subdominios (táctica común de phishing)")
	}

	// 7. Palabras clave sospechosas en dominio
	suspiciousKeywords := []string{"secure", "login", "verify", "account", "update", "confirm", "banking", "seguro", "verificar", "cuenta", "actualizar"}
	for _, kw := range suspiciousKeywords {
		if strings.Contains(domain, kw) {
			result.Score += 10
			result.Flags = append(result.Flags, "suspicious_keyword")
			result.Reasons = append(result.Reasons, "El dominio contiene palabras que buscan generar urgencia o confianza falsa")
			break
		}
	}
}

// analyzeEmail analiza heurísticas específicas de emails
func (h *HeuristicEngine) analyzeEmail(indicators *checkers.Indicators, ctx *checkers.AnalysisContext, result *HeuristicResult) {
	domain := strings.ToLower(indicators.EmailDomain)

	// 1. TLD sospechoso
	if points, isSuspicious := h.suspiciousTLDs[indicators.TLD]; isSuspicious {
		result.Score += points
		result.Flags = append(result.Flags, "suspicious_tld")
		result.Reasons = append(result.Reasons, fmt.Sprintf("El email usa un dominio con TLD sospechoso (.%s)", indicators.TLD))
	}

	// 2. Typosquatting en dominio del email
	for brand, domains := range h.spanishBanks {
		if h.isTyposquatting(domain, brand, domains) {
			result.Score += 40
			result.Flags = append(result.Flags, "email_typosquatting")
			result.Reasons = append(result.Reasons, fmt.Sprintf("El dominio del email parece imitar a %s", brand))
			break
		}
	}

	// 3. Context mismatch
	if ctx != nil && ctx.ClaimedSender != "" {
		claimedLower := strings.ToLower(ctx.ClaimedSender)
		for brand, domains := range h.spanishBanks {
			if strings.Contains(claimedLower, brand) {
				isLegit := false
				for _, legitDomain := range domains {
					if domain == legitDomain {
						isLegit = true
						break
					}
				}
				if !isLegit {
					result.Score += 45
					result.Flags = append(result.Flags, "email_sender_mismatch")
					result.Reasons = append(result.Reasons, fmt.Sprintf("El email dice ser de %s pero el dominio no es oficial", brand))
				}
				break
			}
		}
	}

	// 4. Dominios de email desechables comunes
	disposableDomains := []string{
		"tempmail.com", "guerrillamail.com", "10minutemail.com", "mailinator.com",
		"throwaway.email", "temp-mail.org", "fakeinbox.com", "trashmail.com",
	}
	for _, disposable := range disposableDomains {
		if domain == disposable || strings.HasSuffix(domain, "."+disposable) {
			result.Score += 30
			result.Flags = append(result.Flags, "disposable_email")
			result.Reasons = append(result.Reasons, "Esta dirección de email parece ser temporal/desechable")
			break
		}
	}
}

// analyzePhone analiza heurísticas específicas de teléfonos
func (h *HeuristicEngine) analyzePhone(indicators *checkers.Indicators, ctx *checkers.AnalysisContext, result *HeuristicResult) {
	// 1. Número premium
	if indicators.IsPremium {
		// Determinar qué prefijo es
		prefix := ""
		for _, p := range h.premiumPrefixes {
			if strings.HasPrefix(indicators.NationalNum, p) {
				prefix = p
				break
			}
		}
		result.Score += 50
		result.Flags = append(result.Flags, "premium_number")
		result.Reasons = append(result.Reasons, fmt.Sprintf("Este es un número de tarificación adicional (prefijo %s). Las llamadas tienen coste elevado.", prefix))
	}

	// 2. Context mismatch: dice ser banco/empresa pero número no parece oficial
	if ctx != nil && ctx.ClaimedSender != "" {
		claimedLower := strings.ToLower(ctx.ClaimedSender)
		// Los bancos españoles no llaman desde móviles
		if strings.HasPrefix(indicators.NationalNum, "6") || strings.HasPrefix(indicators.NationalNum, "7") {
			for brand := range h.spanishBanks {
				if strings.Contains(claimedLower, brand) {
					result.Score += 35
					result.Flags = append(result.Flags, "bank_mobile_number")
					result.Reasons = append(result.Reasons, fmt.Sprintf("El remitente dice ser %s pero el número es un móvil. Los bancos no llaman desde móviles.", brand))
					break
				}
			}
		}
	}

	// 3. Número internacional no esperado
	if indicators.CountryCode != "+34" && indicators.CountryCode != "" {
		// Si dice ser empresa española pero número no es español
		if ctx != nil && ctx.ClaimedSender != "" {
			claimedLower := strings.ToLower(ctx.ClaimedSender)
			for brand := range h.spanishBanks {
				if strings.Contains(claimedLower, brand) {
					result.Score += 40
					result.Flags = append(result.Flags, "foreign_number_spanish_sender")
					result.Reasons = append(result.Reasons, fmt.Sprintf("El remitente dice ser %s (empresa española) pero el número es internacional (%s)", brand, indicators.CountryCode))
					break
				}
			}
		}
	}
}

// isTyposquatting detecta si un dominio parece typosquatting de una marca
func (h *HeuristicEngine) isTyposquatting(domain, brand string, legitDomains []string) bool {
	// Primero verificar si es un dominio legítimo
	for _, legit := range legitDomains {
		if domain == legit || strings.HasSuffix(domain, "."+legit) {
			return false
		}
	}

	// Verificar si contiene el nombre de la marca
	if strings.Contains(domain, brand) {
		// Contiene la marca pero no es dominio oficial = typosquatting
		return true
	}

	// Verificar distancia de Levenshtein contra dominios oficiales
	for _, legit := range legitDomains {
		// Solo comparar la parte antes del TLD
		legitBase := strings.Split(legit, ".")[0]
		domainBase := strings.Split(domain, ".")[0]

		if levenshteinDistance(domainBase, legitBase) <= 2 {
			return true
		}
	}

	return false
}

// levenshteinDistance calcula la distancia de Levenshtein entre dos strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Crear matriz
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Llenar matriz
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// ToCheckResult convierte el resultado heurístico a CheckResult
func (h *HeuristicEngine) ToCheckResult(result *HeuristicResult) *checkers.CheckResult {
	found := result.Score >= 20 // Umbral mínimo para considerar sospechoso

	// Determinar tipo de amenaza basado en flags
	threatType := checkers.ThreatTypeUnknown
	if contains(result.Flags, "typosquatting_bank") || contains(result.Flags, "email_typosquatting") {
		threatType = checkers.ThreatTypePhishing
	} else if contains(result.Flags, "premium_number") {
		threatType = "scam"
	} else if contains(result.Flags, "context_mismatch") {
		threatType = checkers.ThreatTypeSocialEng
	}

	// Calcular confianza basada en score
	confidence := float64(result.Score) / 100.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.3 && found {
		confidence = 0.3
	}

	return &checkers.CheckResult{
		Source:     "heuristics",
		Found:      found,
		ThreatType: threatType,
		Confidence: confidence,
		Tags:       result.Flags,
		RawData: map[string]interface{}{
			"score":        result.Score,
			"reasons":      result.Reasons,
			"context_hits": result.ContextHits,
		},
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
