package models

// EmailAnalysisRequest representa la solicitud para analizar un email
type EmailAnalysisRequest struct {
	Email   string `json:"email"`
	Context string `json:"context,omitempty"` // Contexto adicional (asunto, contenido, etc.)
}

// URLAnalysisRequest representa la solicitud para analizar una URL
type URLAnalysisRequest struct {
	URL     string `json:"url"`
	Context string `json:"context,omitempty"`
}

// PhoneAnalysisRequest representa la solicitud para analizar un número de teléfono
type PhoneAnalysisRequest struct {
	Phone       string `json:"phone"`
	CountryCode string `json:"country_code,omitempty"` // Código de país (ej: "ES", "MX")
	Context     string `json:"context,omitempty"`
}

// BatchAnalysisRequest representa una solicitud de análisis en lote
type BatchAnalysisRequest struct {
	Emails []string `json:"emails,omitempty"`
	URLs   []string `json:"urls,omitempty"`
	Phones []string `json:"phones,omitempty"`
}
