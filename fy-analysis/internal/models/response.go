package models

import "time"

// ThreatLevel representa el nivel de amenaza
type ThreatLevel string

const (
	ThreatLevelSafe       ThreatLevel = "safe"
	ThreatLevelLow        ThreatLevel = "low"
	ThreatLevelMedium     ThreatLevel = "medium"
	ThreatLevelHigh       ThreatLevel = "high"
	ThreatLevelCritical   ThreatLevel = "critical"
	ThreatLevelUnknown    ThreatLevel = "unknown"
)

// ThreatType representa el tipo de amenaza detectada
type ThreatType string

const (
	ThreatTypePhishing     ThreatType = "phishing"
	ThreatTypeMalware      ThreatType = "malware"
	ThreatTypeSpam         ThreatType = "spam"
	ThreatTypeScam         ThreatType = "scam"
	ThreatTypeSuspicious   ThreatType = "suspicious"
	ThreatTypeDisposable   ThreatType = "disposable_email"
	ThreatTypeFraud        ThreatType = "fraud"
	ThreatTypeNone         ThreatType = "none"
)

// AnalysisResult representa el resultado base de cualquier análisis
type AnalysisResult struct {
	IsMalicious  bool          `json:"is_malicious"`
	ThreatLevel  ThreatLevel   `json:"threat_level"`
	ThreatTypes  []ThreatType  `json:"threat_types,omitempty"`
	Confidence   float64       `json:"confidence"` // 0.0 - 1.0
	Reasons      []string      `json:"reasons,omitempty"`
	AnalyzedAt   time.Time     `json:"analyzed_at"`
}

// EmailAnalysisResponse representa la respuesta del análisis de email
type EmailAnalysisResponse struct {
	Email          string         `json:"email"`
	Analysis       AnalysisResult `json:"analysis"`
	DomainInfo     *DomainInfo    `json:"domain_info,omitempty"`
	Recommendations []string      `json:"recommendations,omitempty"`
}

// DomainInfo información sobre el dominio del email
type DomainInfo struct {
	Domain       string `json:"domain"`
	IsDisposable bool   `json:"is_disposable"`
	IsFreemail   bool   `json:"is_freemail"`
	HasMXRecords bool   `json:"has_mx_records"`
	Age          string `json:"age,omitempty"` // Antigüedad del dominio
}

// URLAnalysisResponse representa la respuesta del análisis de URL
type URLAnalysisResponse struct {
	URL             string         `json:"url"`
	Analysis        AnalysisResult `json:"analysis"`
	URLInfo         *URLInfo       `json:"url_info,omitempty"`
	Recommendations []string       `json:"recommendations,omitempty"`
}

// URLInfo información detallada sobre la URL
type URLInfo struct {
	Domain          string   `json:"domain"`
	Scheme          string   `json:"scheme"`
	Path            string   `json:"path"`
	HasSuspiciousParams bool `json:"has_suspicious_params"`
	IsShortened     bool     `json:"is_shortened"`
	RedirectsTo     string   `json:"redirects_to,omitempty"`
	SSLValid        bool     `json:"ssl_valid"`
	Categories      []string `json:"categories,omitempty"`
}

// PhoneAnalysisResponse representa la respuesta del análisis de teléfono
type PhoneAnalysisResponse struct {
	Phone           string         `json:"phone"`
	Analysis        AnalysisResult `json:"analysis"`
	PhoneInfo       *PhoneInfo     `json:"phone_info,omitempty"`
	Recommendations []string       `json:"recommendations,omitempty"`
}

// PhoneInfo información sobre el número de teléfono
type PhoneInfo struct {
	CountryCode  string `json:"country_code"`
	Country      string `json:"country"`
	Carrier      string `json:"carrier,omitempty"`
	Type         string `json:"type"` // mobile, landline, voip, premium, toll_free
	IsValid      bool   `json:"is_valid"`
	IsPremiumRate bool  `json:"is_premium_rate"`
}

// BatchAnalysisResponse representa la respuesta de análisis en lote
type BatchAnalysisResponse struct {
	Emails  []EmailAnalysisResponse `json:"emails,omitempty"`
	URLs    []URLAnalysisResponse   `json:"urls,omitempty"`
	Phones  []PhoneAnalysisResponse `json:"phones,omitempty"`
	Summary BatchSummary            `json:"summary"`
}

// BatchSummary resumen del análisis en lote
type BatchSummary struct {
	TotalAnalyzed    int `json:"total_analyzed"`
	MaliciousCount   int `json:"malicious_count"`
	SuspiciousCount  int `json:"suspicious_count"`
	SafeCount        int `json:"safe_count"`
}

// ErrorResponse representa una respuesta de error
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// HealthResponse representa la respuesta del health check
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}
