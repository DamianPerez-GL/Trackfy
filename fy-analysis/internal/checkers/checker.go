package checkers

import (
	"context"
	"time"
)

// InputType define el tipo de entrada a analizar
type InputType string

const (
	InputTypeURL   InputType = "url"
	InputTypeEmail InputType = "email"
	InputTypePhone InputType = "phone"
)

// Indicators contiene los indicadores extraídos para verificación
type Indicators struct {
	// Común
	Original   string    // Valor original sin modificar
	Normalized string    // Valor normalizado
	Hash       string    // SHA256 del valor normalizado
	InputType  InputType // Tipo de entrada

	// URL específico
	FullURL    string // URL completa normalizada (alias de Normalized para URLs)
	Domain     string // Dominio extraído
	DomainHash string // SHA256 del dominio
	IP         string // IP si aplica (resolved o directo)
	TLD        string // Top Level Domain (.es, .com, etc)
	Scheme     string // http, https
	Path       string // Path de la URL
	URLHash    string // SHA256 del URL (alias de Hash para URLs)

	// Email específico
	EmailUser   string // Parte antes del @
	EmailDomain string // Dominio del email

	// Phone específico
	PhoneNumber  string // Número normalizado E.164
	CountryCode  string // Código de país (+34)
	NationalNum  string // Número sin código de país
	IsPremium    bool   // Si es número premium (806, 807, etc)
	CarrierHint  string // Pista del operador si disponible
}

// AnalysisContext información adicional que da el usuario
type AnalysisContext struct {
	ClaimedSender string `json:"claimed_sender,omitempty"` // "Dice ser de BBVA"
	MessageType   string `json:"message_type,omitempty"`   // "sms", "whatsapp", "email"
	OriginalText  string `json:"original_text,omitempty"`  // Texto completo del mensaje
}

// CheckResult representa el resultado de un checker individual
type CheckResult struct {
	Source     string                 // Nombre del checker (urlhaus, webrisk, etc)
	Found      bool                   // Si se encontró como amenaza
	ThreatType string                 // Tipo: malware, phishing, spam, unwanted, etc
	Confidence float64                // Confianza 0.0-1.0
	Tags       []string               // Tags adicionales
	RawData    map[string]interface{} // Datos crudos de la respuesta
	Latency    time.Duration          // Tiempo de respuesta
	Error      error                  // Error si falló
}

// ThreatChecker es la interface que deben implementar todos los checkers
type ThreatChecker interface {
	// Check verifica los indicadores contra la fuente de amenazas
	Check(ctx context.Context, indicators *Indicators) (*CheckResult, error)

	// Name retorna el nombre del checker
	Name() string

	// Weight retorna el peso para el cálculo del score (0.0-1.0)
	Weight() float64

	// IsEnabled indica si el checker está habilitado/configurado
	IsEnabled() bool

	// SupportedTypes retorna los tipos de entrada que soporta este checker
	SupportedTypes() []InputType
}

// SupportsType verifica si un checker soporta un tipo de entrada
func SupportsType(checker ThreatChecker, inputType InputType) bool {
	for _, t := range checker.SupportedTypes() {
		if t == inputType {
			return true
		}
	}
	return false
}

// ThreatType constantes para tipos de amenazas
const (
	ThreatTypeMalware      = "malware"
	ThreatTypePhishing     = "phishing"
	ThreatTypeSpam         = "spam"
	ThreatTypeUnwanted     = "unwanted"
	ThreatTypeSocialEng    = "social_engineering"
	ThreatTypePotentiallyHarmful = "potentially_harmful"
	ThreatTypeUnknown      = "unknown"
)

// CheckerConfig configuración común para checkers
type CheckerConfig struct {
	Enabled    bool
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
}