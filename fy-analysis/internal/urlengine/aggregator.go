package urlengine

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trackfy/fy-analysis/internal/checkers"
)

// Aggregator agrega resultados de múltiples checkers y calcula el score final
type Aggregator struct {
	// Pesos por defecto si el checker no especifica
	defaultWeights map[string]float64
}

// NewAggregator crea un nuevo aggregator
func NewAggregator() *Aggregator {
	return &Aggregator{
		defaultWeights: map[string]float64{
			"urlhaus":   0.40, // Base de datos de malware - alta confianza
			"webrisk":   0.30, // Google Web Risk - buena cobertura
			"phishtank": 0.20, // PhishTank - especializado en phishing
			"urlscan":   0.10, // URLScan.io - fallback/complementario
		},
	}
}

// Aggregate combina los resultados y genera la respuesta final
func (a *Aggregator) Aggregate(
	normalized *NormalizeResult,
	indicators *checkers.Indicators,
	results []*checkers.CheckResult,
	startTime time.Time,
) *URLCheckResponse {

	response := &URLCheckResponse{
		URL:           normalized.OriginalURL,
		NormalizedURL: normalized.NormalizedURL,
		Threats:       []ThreatDetail{},
		Sources:       []SourceResult{},
		CheckedAt:     time.Now().UTC(),
		Cached:        false,
	}

	// Si se expandió shortener, usar URL expandida
	if normalized.ExpandedURL != "" {
		response.NormalizedURL = normalized.ExpandedURL
	}

	// PRIMERO: Verificar si algún checker marcó el dominio como whitelisted
	for _, result := range results {
		if result.RawData != nil {
			if isSafe, ok := result.RawData["is_safe"].(bool); ok && isSafe {
				// Dominio está en whitelist - retornar como seguro inmediatamente
				response.RiskScore = 0
				response.RiskLevel = RiskLevelSafe

				brand := ""
				if b, ok := result.RawData["brand"].(string); ok {
					brand = b
				}

				if brand != "" {
					response.Explanation = fmt.Sprintf("✅ Este es el dominio oficial de %s. Es seguro.", brand)
				} else {
					response.Explanation = "✅ Este dominio está verificado como legítimo."
				}

				response.Action = "Puedes acceder con confianza."
				response.Latency = time.Since(startTime).String()

				// Añadir source info
				response.Sources = append(response.Sources, SourceResult{
					Name:    result.Source,
					Found:   false,
					Latency: result.Latency.String(),
					Weight:  a.getWeight(result.Source),
				})

				log.Info().
					Str("domain", normalized.NormalizedURL).
					Str("brand", brand).
					Msg("[Aggregator] Domain is WHITELISTED - returning safe")

				return response
			}
		}
	}

	// Procesar cada resultado
	var totalWeight float64
	var weightedScore float64
	var threatsFound int

	for _, result := range results {
		// Obtener peso del checker
		weight := a.getWeight(result.Source)

		// Añadir a sources
		sourceResult := SourceResult{
			Name:    result.Source,
			Found:   result.Found,
			Latency: result.Latency.String(),
			Weight:  weight,
		}

		if result.Error != nil {
			sourceResult.Error = result.Error.Error()
			// No contar errores en el peso total
			response.Sources = append(response.Sources, sourceResult)
			continue
		}

		response.Sources = append(response.Sources, sourceResult)
		totalWeight += weight

		// Si encontró amenaza
		if result.Found {
			threatsFound++

			// Añadir detalle de amenaza
			threat := ThreatDetail{
				Source:     result.Source,
				Type:       result.ThreatType,
				Confidence: result.Confidence,
				Tags:       result.Tags,
			}
			response.Threats = append(response.Threats, threat)

			// Calcular contribución al score
			// Score = peso * confianza * 100
			contribution := weight * result.Confidence * 100
			weightedScore += contribution

			log.Debug().
				Str("source", result.Source).
				Float64("weight", weight).
				Float64("confidence", result.Confidence).
				Float64("contribution", contribution).
				Msg("[Aggregator] Threat contribution calculated")
		}
	}

	// Calcular score final
	if totalWeight > 0 && threatsFound > 0 {
		// Normalizar por peso total
		response.RiskScore = int(weightedScore / totalWeight)

		// Aplicar boost si múltiples fuentes confirman
		if threatsFound >= 2 {
			boost := 1.0 + (float64(threatsFound-1) * 0.1) // +10% por cada fuente adicional
			response.RiskScore = int(float64(response.RiskScore) * boost)
			log.Debug().
				Int("threats_found", threatsFound).
				Float64("boost", boost).
				Int("boosted_score", response.RiskScore).
				Msg("[Aggregator] Applied multi-source boost")
		}

		// Limitar a 100
		if response.RiskScore > 100 {
			response.RiskScore = 100
		}
	} else if totalWeight == 0 {
		// Ningún checker respondió correctamente
		response.RiskScore = 50 // Incertidumbre
	} else {
		// Ninguna amenaza encontrada
		response.RiskScore = 0
	}

	// Determinar nivel de riesgo
	response.RiskLevel = GetRiskLevel(response.RiskScore)

	// Generar explicación y acción
	response.Explanation = a.generateExplanation(response, normalized)
	response.Action = GetRecommendedAction(response.RiskLevel)

	// Calcular latencia total
	response.Latency = time.Since(startTime).String()

	log.Info().
		Int("risk_score", response.RiskScore).
		Str("risk_level", string(response.RiskLevel)).
		Int("threats", len(response.Threats)).
		Int("sources_ok", len(response.Sources)).
		Msg("[Aggregator] Aggregation completed")

	return response
}

// getWeight obtiene el peso de un checker
func (a *Aggregator) getWeight(source string) float64 {
	if weight, exists := a.defaultWeights[source]; exists {
		return weight
	}
	return 0.1 // Peso default para checkers desconocidos
}

// generateExplanation genera una explicación detallada
func (a *Aggregator) generateExplanation(response *URLCheckResponse, normalized *NormalizeResult) string {
	var explanation string

	// Info sobre shortener
	if normalized.IsShortener {
		explanation += fmt.Sprintf("URL acortada detectada (expandida a: %s). ", normalized.ExpandedURL)
	}

	// Basado en nivel de riesgo
	switch response.RiskLevel {
	case RiskLevelSafe:
		explanation += "La URL parece segura. "
		if len(response.Sources) > 0 {
			explanation += fmt.Sprintf("Verificada contra %d fuentes de amenazas sin resultados positivos.", len(response.Sources))
		}

	case RiskLevelWarning:
		if len(response.Threats) > 0 {
			explanation += "Se encontraron indicios de posible amenaza. "
			for _, t := range response.Threats {
				explanation += fmt.Sprintf("[%s: %s] ", t.Source, t.Type)
			}
			explanation += "Se recomienda precaución."
		} else {
			explanation += "No se pudo verificar completamente. Algunas fuentes no respondieron. Proceda con precaución."
		}

	case RiskLevelDanger:
		explanation += "¡PELIGRO! Esta URL ha sido identificada como maliciosa. "
		for _, t := range response.Threats {
			explanation += fmt.Sprintf("[%s reporta: %s con %.0f%% confianza] ", t.Source, t.Type, t.Confidence*100)
		}
		explanation += "NO visite esta URL."
	}

	return explanation
}

// CalculateScoreFromThreats calcula un score basado solo en los tipos de amenaza
func (a *Aggregator) CalculateScoreFromThreats(threats []ThreatDetail) int {
	if len(threats) == 0 {
		return 0
	}

	// Pesos por tipo de amenaza
	threatWeights := map[string]int{
		checkers.ThreatTypeMalware:           90,
		checkers.ThreatTypePhishing:          85,
		checkers.ThreatTypeSocialEng:         80,
		checkers.ThreatTypePotentiallyHarmful: 60,
		checkers.ThreatTypeSpam:              40,
		checkers.ThreatTypeUnwanted:          30,
		checkers.ThreatTypeUnknown:           50,
	}

	maxScore := 0
	for _, t := range threats {
		if score, exists := threatWeights[t.Type]; exists {
			// Aplicar confianza
			adjustedScore := int(float64(score) * t.Confidence)
			if adjustedScore > maxScore {
				maxScore = adjustedScore
			}
		}
	}

	return maxScore
}
