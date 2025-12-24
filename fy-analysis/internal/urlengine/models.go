package urlengine

import (
	"time"

	"github.com/trackfy/fy-analysis/internal/checkers"
)

// Re-exportar tipos de checkers para conveniencia
type InputType = checkers.InputType

const (
	InputTypeURL   = checkers.InputTypeURL
	InputTypeEmail = checkers.InputTypeEmail
	InputTypePhone = checkers.InputTypePhone
)

// AnalysisRequest representa la solicitud unificada de análisis
type AnalysisRequest struct {
	Input   string                   `json:"input" validate:"required"`
	Type    InputType                `json:"type" validate:"required,oneof=url email phone"`
	Context *checkers.AnalysisContext `json:"context,omitempty"`
}

// URLCheckRequest representa la solicitud de verificación (legacy, para compatibilidad)
type URLCheckRequest struct {
	URL string `json:"url" validate:"required"`
}

// URLCheckResponse representa la respuesta completa de verificación
type URLCheckResponse struct {
	URL           string         `json:"url"`
	NormalizedURL string         `json:"normalized_url"`
	RiskScore     int            `json:"risk_score"`      // 0-100
	RiskLevel     RiskLevel      `json:"risk_level"`      // safe/warning/danger
	Threats       []ThreatDetail `json:"threats"`
	Explanation   string         `json:"explanation"`     // Explicación para el usuario/Fy
	Action        string         `json:"recommended_action"`
	Sources       []SourceResult `json:"sources"`         // Resultados por fuente
	CheckedAt     time.Time      `json:"checked_at"`
	Cached        bool           `json:"cached"`          // Si vino de cache
	Latency       string         `json:"latency"`         // Tiempo total de verificación
}

// ThreatDetail detalle de una amenaza detectada
type ThreatDetail struct {
	Source     string   `json:"source"`
	Type       string   `json:"type"`        // malware, phishing, spam
	Confidence float64  `json:"confidence"`  // 0.0-1.0
	Tags       []string `json:"tags,omitempty"`
}

// SourceResult resultado individual de cada fuente
type SourceResult struct {
	Name    string  `json:"name"`
	Found   bool    `json:"found"`
	Latency string  `json:"latency"`
	Error   string  `json:"error,omitempty"`
	Weight  float64 `json:"weight"`
}

// RiskLevel niveles de riesgo
type RiskLevel string

const (
	RiskLevelSafe    RiskLevel = "safe"    // 0-20
	RiskLevelWarning RiskLevel = "warning" // 21-60
	RiskLevelDanger  RiskLevel = "danger"  // 61-100
)

// GetRiskLevel determina el nivel de riesgo basado en el score
func GetRiskLevel(score int) RiskLevel {
	switch {
	case score <= 20:
		return RiskLevelSafe
	case score <= 60:
		return RiskLevelWarning
	default:
		return RiskLevelDanger
	}
}

// GetExplanation genera una explicación basada en el resultado
func GetExplanation(level RiskLevel, threats []ThreatDetail) string {
	switch level {
	case RiskLevelSafe:
		return "La URL parece segura. No se encontraron amenazas en las fuentes consultadas."
	case RiskLevelWarning:
		if len(threats) > 0 {
			return "La URL presenta indicios sospechosos. Se recomienda precaución."
		}
		return "No se pudo verificar completamente. Proceda con precaución."
	case RiskLevelDanger:
		if len(threats) > 0 {
			return "¡PELIGRO! Esta URL ha sido identificada como maliciosa. NO la visite."
		}
		return "Esta URL presenta múltiples indicadores de riesgo. Se recomienda NO visitarla."
	default:
		return "No se pudo determinar el nivel de riesgo."
	}
}

// GetRecommendedAction genera la acción recomendada
func GetRecommendedAction(level RiskLevel) string {
	switch level {
	case RiskLevelSafe:
		return "ALLOW"
	case RiskLevelWarning:
		return "WARN"
	case RiskLevelDanger:
		return "BLOCK"
	default:
		return "WARN"
	}
}

// AggregatedResult resultado agregado de todos los checkers
type AggregatedResult struct {
	Indicators *checkers.Indicators
	Results    []*checkers.CheckResult
	TotalTime  time.Duration
}

// AnalysisResponse es la respuesta unificada de análisis
type AnalysisResponse struct {
	Input             string         `json:"input"`
	Type              InputType      `json:"type"`
	NormalizedInput   string         `json:"normalized_input"`
	RiskScore         int            `json:"risk_score"`
	RiskLevel         string         `json:"risk_level"`
	Threats           []ThreatDetail `json:"threats"`
	Reasons           []string       `json:"reasons"`
	RecommendedAction string         `json:"recommended_action"`
	Sources           []SourceResult `json:"sources"`
	CacheHit          bool           `json:"cache_hit"`
	ResponseTimeMs    int64          `json:"response_time_ms"`
	CheckedAt         time.Time      `json:"checked_at"`
}

// RecommendedAction constantes para acciones recomendadas
const (
	ActionSafe    = "SAFE"     // Seguro para proceder
	ActionCaution = "CAUTION"  // Proceder con precaución
	ActionNoClick = "NO_CLICK" // No hacer click/llamar
	ActionBlock   = "BLOCK"    // Bloquear completamente
)

// GetActionForLevel retorna la acción recomendada basada en el nivel de riesgo
func GetActionForLevel(level RiskLevel) string {
	switch level {
	case RiskLevelSafe:
		return ActionSafe
	case RiskLevelWarning:
		return ActionCaution
	case RiskLevelDanger:
		return ActionBlock
	default:
		return ActionCaution
	}
}

// Explicaciones en español para Fy
var ReasonsES = map[string]string{
	"urlhaus_malware":     "Esta URL está reportada en URLhaus como distribuidora de malware",
	"urlhaus_domain":      "El dominio está asociado con distribución de malware en URLhaus",
	"phishtank_phishing":  "Esta URL está confirmada como phishing en PhishTank",
	"phishtank_domain":    "El dominio está asociado con campañas de phishing",
	"webrisk_malware":     "Google Web Risk identifica esta URL como malware",
	"webrisk_social_eng":  "Google Web Risk identifica esto como ingeniería social",
	"webrisk_unwanted":    "Google Web Risk identifica software no deseado",
	"urlscan_malicious":   "URLScan.io detectó contenido malicioso",
	"domain_typosquatting": "El dominio parece imitar a %s (posible suplantación)",
	"domain_new":          "El dominio fue registrado hace menos de 30 días",
	"tld_suspicious":      "El dominio usa un TLD sospechoso (%s)",
	"context_mismatch":    "El remitente dice ser %s pero el dominio no coincide",
	"premium_number":      "Este es un número de tarificación adicional (empezando por %s)",
	"unknown_carrier":     "No se pudo verificar el operador de este número",
	"disposable_email":    "Esta dirección de email parece ser temporal/desechable",
	"no_threats_found":    "No se encontraron amenazas en las fuentes consultadas",
	"partial_check":       "Algunas fuentes no respondieron. Proceda con precaución",
}