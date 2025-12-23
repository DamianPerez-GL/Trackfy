package phone

import (
	"regexp"
	"strings"
	"time"

	"github.com/trackfy/fy-analysis/internal/models"
)

// Analyzer maneja el análisis de números de teléfono
type Analyzer struct {
	premiumPrefixes    map[string][]string
	scamNumbers        map[string]bool
	countryPatterns    map[string]*regexp.Regexp
}

// NewAnalyzer crea una nueva instancia del analizador de teléfonos
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		premiumPrefixes: loadPremiumPrefixes(),
		scamNumbers:     loadScamNumbers(),
		countryPatterns: loadCountryPatterns(),
	}
}

// Analyze realiza el análisis completo de un número de teléfono
func (a *Analyzer) Analyze(phone string, countryCode string, context string) models.PhoneAnalysisResponse {
	// Limpiar número
	cleanPhone := cleanPhoneNumber(phone)

	response := models.PhoneAnalysisResponse{
		Phone: phone,
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

	// Extraer información del teléfono
	phoneInfo := a.extractPhoneInfo(cleanPhone, countryCode)
	response.PhoneInfo = phoneInfo

	// Verificar si es válido
	if !phoneInfo.IsValid {
		response.Analysis.ThreatLevel = models.ThreatLevelMedium
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeSuspicious)
		response.Analysis.Reasons = append(response.Analysis.Reasons, "Formato de número inválido")
		response.Analysis.Confidence = 0.7
		return response
	}

	// Verificar si es número conocido como scam
	if a.isKnownScam(cleanPhone) {
		response.Analysis.IsMalicious = true
		response.Analysis.ThreatLevel = models.ThreatLevelCritical
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeScam)
		response.Analysis.Reasons = append(response.Analysis.Reasons, "Número reportado en base de datos de estafas")
		response.Analysis.Confidence = 0.99
		response.Recommendations = append(response.Recommendations, "NO contestar ni devolver la llamada")
		return response
	}

	// Verificar si es número premium
	if phoneInfo.IsPremiumRate {
		response.Analysis.ThreatLevel = models.ThreatLevelMedium
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeFraud)
		response.Analysis.Reasons = append(response.Analysis.Reasons, "Número de tarificación especial (premium)")
		response.Analysis.Confidence = 0.9
		response.Recommendations = append(response.Recommendations, "Llamar a este número puede generar cargos elevados")
	}

	// Análisis heurístico
	heuristicResult := a.heuristicAnalysis(cleanPhone, phoneInfo, context)

	if heuristicResult.score > 0.6 {
		response.Analysis.ThreatLevel = models.ThreatLevelHigh
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeScam)
		response.Analysis.Reasons = append(response.Analysis.Reasons, heuristicResult.reasons...)
		response.Analysis.Confidence = heuristicResult.score
		response.Recommendations = append(response.Recommendations, "Verificar la identidad del llamante por otros medios")
	} else if heuristicResult.score > 0.3 {
		if response.Analysis.ThreatLevel == models.ThreatLevelSafe {
			response.Analysis.ThreatLevel = models.ThreatLevelLow
		}
		response.Analysis.Reasons = append(response.Analysis.Reasons, heuristicResult.reasons...)
	}

	// Si no hay amenazas
	if len(response.Analysis.ThreatTypes) == 0 {
		response.Analysis.ThreatTypes = append(response.Analysis.ThreatTypes, models.ThreatTypeNone)
		response.Analysis.Confidence = 0.8
	}

	return response
}

func cleanPhoneNumber(phone string) string {
	// Remover todo excepto dígitos y el símbolo +
	reg := regexp.MustCompile(`[^\d+]`)
	return reg.ReplaceAllString(phone, "")
}

func (a *Analyzer) extractPhoneInfo(phone string, countryCode string) *models.PhoneInfo {
	info := &models.PhoneInfo{
		CountryCode:   countryCode,
		IsValid:       a.isValidPhone(phone, countryCode),
		IsPremiumRate: a.isPremiumRate(phone, countryCode),
	}

	// Detectar país si no se especificó
	if countryCode == "" {
		info.CountryCode, info.Country = a.detectCountry(phone)
	} else {
		info.Country = getCountryName(countryCode)
	}

	// Detectar tipo de número
	info.Type = a.detectPhoneType(phone, info.CountryCode)

	return info
}

func (a *Analyzer) isValidPhone(phone string, countryCode string) bool {
	// Validación básica de longitud
	cleanPhone := strings.TrimPrefix(phone, "+")

	if len(cleanPhone) < 7 || len(cleanPhone) > 15 {
		return false
	}

	// Si hay patrón específico para el país, validar
	if pattern, exists := a.countryPatterns[countryCode]; exists {
		return pattern.MatchString(phone)
	}

	// Validación genérica
	return len(cleanPhone) >= 8
}

func (a *Analyzer) isPremiumRate(phone string, countryCode string) bool {
	prefixes, exists := a.premiumPrefixes[countryCode]
	if !exists {
		// Usar prefijos genéricos
		prefixes = a.premiumPrefixes["default"]
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(phone, prefix) || strings.HasPrefix(strings.TrimPrefix(phone, "+"), prefix) {
			return true
		}
	}

	return false
}

func (a *Analyzer) isKnownScam(phone string) bool {
	return a.scamNumbers[phone]
}

func (a *Analyzer) detectCountry(phone string) (string, string) {
	phone = strings.TrimPrefix(phone, "+")

	countryPrefixes := map[string]struct {
		code string
		name string
	}{
		"1":   {"US", "Estados Unidos/Canadá"},
		"34":  {"ES", "España"},
		"52":  {"MX", "México"},
		"54":  {"AR", "Argentina"},
		"55":  {"BR", "Brasil"},
		"56":  {"CL", "Chile"},
		"57":  {"CO", "Colombia"},
		"51":  {"PE", "Perú"},
		"58":  {"VE", "Venezuela"},
		"44":  {"GB", "Reino Unido"},
		"49":  {"DE", "Alemania"},
		"33":  {"FR", "Francia"},
		"39":  {"IT", "Italia"},
	}

	// Buscar prefijo más largo primero
	for i := 3; i >= 1; i-- {
		if len(phone) >= i {
			prefix := phone[:i]
			if country, exists := countryPrefixes[prefix]; exists {
				return country.code, country.name
			}
		}
	}

	return "UNKNOWN", "Desconocido"
}

func (a *Analyzer) detectPhoneType(phone string, countryCode string) string {
	phone = strings.TrimPrefix(phone, "+")

	// Reglas simples por país
	switch countryCode {
	case "ES":
		if strings.HasPrefix(phone, "346") || strings.HasPrefix(phone, "347") {
			return "mobile"
		}
		if strings.HasPrefix(phone, "349") {
			return "landline"
		}
		if strings.HasPrefix(phone, "34800") || strings.HasPrefix(phone, "34900") {
			return "toll_free"
		}
		if strings.HasPrefix(phone, "34803") || strings.HasPrefix(phone, "34806") || strings.HasPrefix(phone, "34807") {
			return "premium"
		}
	case "MX":
		if len(phone) >= 4 && (phone[2] == '1' || phone[3] == '1') {
			return "mobile"
		}
		if strings.HasPrefix(phone, "52800") || strings.HasPrefix(phone, "52900") {
			return "toll_free"
		}
	case "US":
		if strings.HasPrefix(phone, "1800") || strings.HasPrefix(phone, "1888") || strings.HasPrefix(phone, "1877") {
			return "toll_free"
		}
		if strings.HasPrefix(phone, "1900") {
			return "premium"
		}
	}

	return "unknown"
}

type heuristicResult struct {
	score   float64
	reasons []string
}

func (a *Analyzer) heuristicAnalysis(phone string, info *models.PhoneInfo, context string) heuristicResult {
	result := heuristicResult{
		score:   0.0,
		reasons: []string{},
	}

	// 1. País de alto riesgo para estafas
	highRiskCountries := map[string]bool{
		"UNKNOWN": true,
	}
	if highRiskCountries[info.CountryCode] {
		result.score += 0.2
		result.reasons = append(result.reasons, "País de origen no identificado")
	}

	// 2. Número VOIP (difícil de rastrear)
	if info.Type == "voip" {
		result.score += 0.2
		result.reasons = append(result.reasons, "Número VOIP (difícil de rastrear)")
	}

	// 3. Analizar contexto
	if context != "" {
		contextLower := strings.ToLower(context)
		scamPatterns := []string{
			"premio", "ganador", "lotería", "lottery", "winner", "prize",
			"urgente", "urgent", "banco", "bank", "bloqueo", "blocked",
			"suspendido", "suspended", "verificar", "verify", "confirmar",
			"tarjeta", "card", "cuenta", "account", "contraseña", "password",
			"soporte técnico", "tech support", "microsoft", "apple",
		}
		for _, pattern := range scamPatterns {
			if strings.Contains(contextLower, pattern) {
				result.score += 0.15
				result.reasons = append(result.reasons, "Contexto contiene palabras asociadas a estafas")
				break
			}
		}
	}

	// 4. Número con muchos ceros o patrones repetitivos
	if hasRepetitivePattern(phone) {
		result.score += 0.1
		result.reasons = append(result.reasons, "Patrón numérico inusual")
	}

	// Limitar score máximo
	if result.score > 1.0 {
		result.score = 1.0
	}

	return result
}

func hasRepetitivePattern(phone string) bool {
	digits := strings.TrimPrefix(phone, "+")
	if len(digits) < 6 {
		return false
	}

	// Verificar secuencias repetitivas
	for i := 0; i < len(digits)-3; i++ {
		if digits[i] == digits[i+1] && digits[i+1] == digits[i+2] && digits[i+2] == digits[i+3] {
			return true
		}
	}

	return false
}

func getCountryName(code string) string {
	countries := map[string]string{
		"US": "Estados Unidos",
		"ES": "España",
		"MX": "México",
		"AR": "Argentina",
		"BR": "Brasil",
		"CL": "Chile",
		"CO": "Colombia",
		"PE": "Perú",
		"VE": "Venezuela",
		"GB": "Reino Unido",
		"DE": "Alemania",
		"FR": "Francia",
		"IT": "Italia",
	}

	if name, exists := countries[code]; exists {
		return name
	}
	return "Desconocido"
}

func loadPremiumPrefixes() map[string][]string {
	return map[string][]string{
		"ES":      {"803", "806", "807", "905"},
		"MX":      {"900"},
		"US":      {"900", "976"},
		"GB":      {"09", "070"},
		"default": {"900", "901", "902"},
	}
}

func loadScamNumbers() map[string]bool {
	// Base de datos de números conocidos como estafas
	return map[string]bool{
		"+34900000000":  true,
		"+1234567890":   true,
		"+15551234567":  true,
	}
}

func loadCountryPatterns() map[string]*regexp.Regexp {
	return map[string]*regexp.Regexp{
		"ES": regexp.MustCompile(`^\+?34[6-9]\d{8}$`),
		"MX": regexp.MustCompile(`^\+?52\d{10}$`),
		"US": regexp.MustCompile(`^\+?1[2-9]\d{9}$`),
	}
}
