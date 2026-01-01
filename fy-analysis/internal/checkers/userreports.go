package checkers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// UserReportsChecker verifica URLs contra reportes de usuarios
// Usa el sistema de confianza agregada anti-spam
type UserReportsChecker struct {
	db      *sql.DB
	enabled bool
	weight  float64

	// Configuración de umbrales
	minScoreForWarning int // Score mínimo para considerar como warning (default: 40)
	minScoreForDanger  int // Score mínimo para considerar como danger (default: 70)
	minReportersForUse int // Mínimo de reportadores únicos para usar (default: 2)
}

// UserReportsConfig configuración para el checker de reportes
type UserReportsConfig struct {
	DatabaseURL        string
	Weight             float64
	MinScoreForWarning int
	MinScoreForDanger  int
	MinReportersForUse int
}

// NewUserReportsChecker crea un nuevo checker de reportes de usuarios
func NewUserReportsChecker(db *sql.DB, config *UserReportsConfig) *UserReportsChecker {
	if db == nil {
		log.Warn().Msg("[UserReports] No database provided, checker disabled")
		return &UserReportsChecker{
			enabled: false,
			weight:  0.10,
		}
	}

	weight := config.Weight
	if weight <= 0 {
		weight = 0.10 // Peso bajo por defecto (crowdsourced data)
	}

	minWarning := config.MinScoreForWarning
	if minWarning <= 0 {
		minWarning = 40
	}

	minDanger := config.MinScoreForDanger
	if minDanger <= 0 {
		minDanger = 70
	}

	minReporters := config.MinReportersForUse
	if minReporters <= 0 {
		minReporters = 2 // Requiere al menos 2 reportadores por defecto
	}

	log.Info().
		Float64("weight", weight).
		Int("min_warning_score", minWarning).
		Int("min_danger_score", minDanger).
		Int("min_reporters", minReporters).
		Msg("[UserReports] Checker initialized")

	return &UserReportsChecker{
		db:                 db,
		enabled:            true,
		weight:             weight,
		minScoreForWarning: minWarning,
		minScoreForDanger:  minDanger,
		minReportersForUse: minReporters,
	}
}

// Check verifica una URL contra los reportes de usuarios
func (c *UserReportsChecker) Check(ctx context.Context, indicators *Indicators) (*CheckResult, error) {
	if !c.enabled || c.db == nil {
		return &CheckResult{
			Source: c.Name(),
			Found:  false,
			Error:  fmt.Errorf("UserReports checker is disabled"),
		}, nil
	}

	startTime := time.Now()

	// Solo soportamos URLs por ahora
	if indicators.InputType != InputTypeURL {
		return &CheckResult{
			Source:  c.Name(),
			Found:   false,
			Latency: time.Since(startTime),
		}, nil
	}

	return c.checkURL(ctx, indicators, startTime)
}

// checkURL verifica una URL en los reportes de usuarios
func (c *UserReportsChecker) checkURL(ctx context.Context, indicators *Indicators, startTime time.Time) (*CheckResult, error) {
	result := &CheckResult{
		Source: c.Name(),
		Found:  false,
		RawData: map[string]interface{}{
			"reasons": []string{},
		},
	}

	var reasons []string
	url := strings.ToLower(indicators.FullURL)
	domain := strings.ToLower(indicators.Domain)

	// 1. Buscar URL exacta en reportes
	var threatType sql.NullString
	var aggregatedScore int16
	var totalReports, uniqueReporters int
	var status string
	var firstReported, lastReported time.Time

	err := c.db.QueryRowContext(ctx, `
		SELECT
			primary_threat_type,
			aggregated_score,
			total_reports,
			unique_reporters,
			status,
			first_reported_at,
			last_reported_at
		FROM reported_urls
		WHERE url_hash = sha256_bytea($1)
		  AND (flags & 1) = 1
		LIMIT 1
	`, url).Scan(&threatType, &aggregatedScore, &totalReports, &uniqueReporters, &status, &firstReported, &lastReported)

	if err == nil && uniqueReporters >= c.minReportersForUse {
		// Encontrada y con suficientes reportadores
		result.Found = true
		result.RawData["aggregated_score"] = aggregatedScore
		result.RawData["total_reports"] = totalReports
		result.RawData["unique_reporters"] = uniqueReporters
		result.RawData["review_status"] = status
		result.RawData["first_reported"] = firstReported.Format(time.RFC3339)
		result.RawData["last_reported"] = lastReported.Format(time.RFC3339)
		result.RawData["is_user_reported"] = true

		// Calcular confianza basada en score y status
		confidence := float64(aggregatedScore) / 100.0

		// Boost si ya fue confirmada por revisores
		if status == "confirmed" {
			confidence = confidence * 1.3
			if confidence > 1.0 {
				confidence = 1.0
			}
			reasons = append(reasons, fmt.Sprintf("URL confirmada como maliciosa por revisores (%d reportes de %d usuarios)",
				totalReports, uniqueReporters))
		} else {
			reasons = append(reasons, fmt.Sprintf("URL reportada por %d usuarios diferentes (%d reportes totales)",
				uniqueReporters, totalReports))
		}

		// Penalizar si fue rechazada
		if status == "rejected" {
			result.Found = false
			result.Confidence = 0
			reasons = []string{"URL fue revisada y marcada como segura"}
			result.RawData["reasons"] = reasons
			result.Latency = time.Since(startTime)
			return result, nil
		}

		result.Confidence = confidence

		// Determinar tipo de amenaza
		if threatType.Valid {
			result.ThreatType = threatType.String
		} else {
			result.ThreatType = "user_reported"
		}

		// Determinar severidad basada en score
		if int(aggregatedScore) >= c.minScoreForDanger {
			result.RawData["severity"] = "high"
			result.Tags = append(result.Tags, "high_community_score")
		} else if int(aggregatedScore) >= c.minScoreForWarning {
			result.RawData["severity"] = "medium"
			result.Tags = append(result.Tags, "moderate_community_score")
		} else {
			result.RawData["severity"] = "low"
		}

		log.Debug().
			Str("url", url).
			Int("score", int(aggregatedScore)).
			Int("reporters", uniqueReporters).
			Str("status", status).
			Msg("[UserReports] URL found in user reports")

	} else if err != nil && err != sql.ErrNoRows {
		log.Debug().Err(err).Msg("[UserReports] Error querying reported_urls")
	}

	// 2. Si no encontramos la URL exacta, buscar por dominio
	if !result.Found && domain != "" {
		var domainReports int
		var avgScore float64

		err := c.db.QueryRowContext(ctx, `
			SELECT
				COUNT(*) as domain_reports,
				COALESCE(AVG(aggregated_score), 0) as avg_score
			FROM reported_urls
			WHERE domain = $1
			  AND (flags & 1) = 1
			  AND unique_reporters >= $2
		`, domain, c.minReportersForUse).Scan(&domainReports, &avgScore)

		if err == nil && domainReports >= 3 && avgScore >= float64(c.minScoreForWarning) {
			// Múltiples URLs del mismo dominio reportadas
			result.RawData["domain_reports"] = domainReports
			result.RawData["domain_avg_score"] = avgScore
			result.RawData["domain_pattern"] = true

			reasons = append(reasons, fmt.Sprintf("Dominio con múltiples URLs reportadas (%d URLs, score promedio %.0f)",
				domainReports, avgScore))

			// No marcamos como Found pero añadimos contexto
			result.Tags = append(result.Tags, "suspicious_domain_pattern")

			log.Debug().
				Str("domain", domain).
				Int("reports", domainReports).
				Float64("avg_score", avgScore).
				Msg("[UserReports] Domain has multiple reported URLs")
		}
	}

	result.RawData["reasons"] = reasons
	result.Latency = time.Since(startTime)

	return result, nil
}

// Name retorna el nombre del checker
func (c *UserReportsChecker) Name() string {
	return "user_reports"
}

// Weight retorna el peso del checker
func (c *UserReportsChecker) Weight() float64 {
	return c.weight
}

// IsEnabled indica si el checker está habilitado
func (c *UserReportsChecker) IsEnabled() bool {
	return c.enabled
}

// SupportedTypes retorna los tipos de input soportados
func (c *UserReportsChecker) SupportedTypes() []InputType {
	return []InputType{InputTypeURL}
}

// GetStats retorna estadísticas de reportes de usuarios
func (c *UserReportsChecker) GetStats(ctx context.Context) (map[string]interface{}, error) {
	if !c.enabled || c.db == nil {
		return nil, fmt.Errorf("checker disabled")
	}

	stats := make(map[string]interface{})

	// Estadísticas generales de URLs reportadas
	var totalURLs, pendingReview, confirmed, rejected int64
	var highScore, mediumScore, lowScore int64
	var avgReporters float64
	var totalReports int64

	err := c.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'confirmed'),
			COUNT(*) FILTER (WHERE status = 'rejected'),
			COUNT(*) FILTER (WHERE aggregated_score >= 70),
			COUNT(*) FILTER (WHERE aggregated_score BETWEEN 40 AND 69),
			COUNT(*) FILTER (WHERE aggregated_score < 40),
			COALESCE(AVG(unique_reporters), 0),
			COALESCE(SUM(total_reports), 0)
		FROM reported_urls
		WHERE (flags & 1) = 1
	`).Scan(&totalURLs, &pendingReview, &confirmed, &rejected,
		&highScore, &mediumScore, &lowScore, &avgReporters, &totalReports)

	if err != nil {
		return nil, err
	}

	stats["reported_urls"] = map[string]interface{}{
		"total":            totalURLs,
		"pending_review":   pendingReview,
		"confirmed":        confirmed,
		"rejected":         rejected,
		"high_score":       highScore,
		"medium_score":     mediumScore,
		"low_score":        lowScore,
		"avg_reporters":    avgReporters,
		"total_reports":    totalReports,
	}

	// Estadísticas de usuarios reportadores
	var totalUsers, bannedUsers int64
	var avgTrust float64

	err = c.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE (flags & 1) = 1),
			COALESCE(AVG(trust_score), 50)
		FROM user_trust_scores
	`).Scan(&totalUsers, &bannedUsers, &avgTrust)

	if err == nil {
		stats["reporters"] = map[string]interface{}{
			"total_users":      totalUsers,
			"banned_users":     bannedUsers,
			"avg_trust_score":  avgTrust,
		}
	}

	return stats, nil
}

// ReportURL permite reportar una URL (llamado desde el API gateway)
func (c *UserReportsChecker) ReportURL(ctx context.Context, url, domain, userID string,
	threatType, description, reportContext string, userIP, userAgent string) (bool, string, int, error) {

	if !c.enabled || c.db == nil {
		return false, "Servicio no disponible", 0, fmt.Errorf("checker disabled")
	}

	var success bool
	var message string
	var score int16
	var isNew bool

	// Convertir threat_type string a enum válido
	validThreatTypes := map[string]bool{
		"phishing": true, "malware": true, "scam": true, "spam": true,
		"vishing": true, "smishing": true, "other": true,
	}
	if !validThreatTypes[threatType] {
		threatType = "other"
	}

	err := c.db.QueryRowContext(ctx, `
		SELECT success, message, url_score, is_new_report
		FROM report_url($1, $2, $3, $4::threat_type_enum, $5, $6, $7::inet, $8)
	`, url, domain, userID, threatType, description, reportContext, userIP, userAgent).Scan(&success, &message, &score, &isNew)

	if err != nil {
		log.Error().Err(err).
			Str("url", url).
			Str("user_id", userID).
			Msg("[UserReports] Error calling report_url function")
		return false, "Error al procesar reporte", 0, err
	}

	log.Info().
		Str("url", url).
		Str("user_id", userID).
		Bool("success", success).
		Int("new_score", int(score)).
		Msg("[UserReports] URL report processed")

	return success, message, int(score), nil
}
