package checkers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// LocalDBChecker verifica amenazas contra la base de datos local PostgreSQL
type LocalDBChecker struct {
	db      *sql.DB
	enabled bool
	weight  float64
}

// LocalDBConfig configuración para el checker de DB local
type LocalDBConfig struct {
	DatabaseURL string
	MaxConns    int
	Weight      float64
}

// NewLocalDBChecker crea un nuevo checker de base de datos local
func NewLocalDBChecker(config *LocalDBConfig) *LocalDBChecker {
	if config == nil || config.DatabaseURL == "" {
		log.Warn().Msg("[LocalDB] No database URL provided, checker disabled")
		return &LocalDBChecker{
			enabled: false,
			weight:  0.5,
		}
	}

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Error().Err(err).Msg("[LocalDB] Failed to open database connection")
		return &LocalDBChecker{
			enabled: false,
			weight:  config.Weight,
		}
	}

	// Configurar pool de conexiones
	maxConns := config.MaxConns
	if maxConns <= 0 {
		maxConns = 10
	}
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns / 2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Error().Err(err).Msg("[LocalDB] Failed to ping database")
		return &LocalDBChecker{
			enabled: false,
			weight:  config.Weight,
		}
	}

	weight := config.Weight
	if weight <= 0 {
		weight = 0.5
	}

	log.Info().
		Str("url", maskDatabaseURL(config.DatabaseURL)).
		Int("max_conns", maxConns).
		Float64("weight", weight).
		Msg("[LocalDB] Checker initialized successfully")

	return &LocalDBChecker{
		db:      db,
		enabled: true,
		weight:  weight,
	}
}

// Check verifica un indicador contra la base de datos local
func (c *LocalDBChecker) Check(ctx context.Context, indicators *Indicators) (*CheckResult, error) {
	if !c.enabled || c.db == nil {
		return &CheckResult{
			Source: c.Name(),
			Found:  false,
			Error:  fmt.Errorf("LocalDB checker is disabled"),
		}, nil
	}

	startTime := time.Now()

	switch indicators.InputType {
	case InputTypeURL:
		return c.checkURL(ctx, indicators, startTime)
	case InputTypeEmail:
		return c.checkEmail(ctx, indicators, startTime)
	case InputTypePhone:
		return c.checkPhone(ctx, indicators, startTime)
	default:
		return &CheckResult{
			Source:  c.Name(),
			Found:   false,
			Latency: time.Since(startTime),
		}, nil
	}
}

// checkURL verifica una URL en la base de datos (Schema v2.0 optimizado)
func (c *LocalDBChecker) checkURL(ctx context.Context, indicators *Indicators, startTime time.Time) (*CheckResult, error) {
	result := &CheckResult{
		Source: c.Name(),
		Found:  false,
		RawData: map[string]interface{}{
			"reasons": []string{},
		},
	}

	var reasons []string
	domain := strings.ToLower(indicators.Domain)

	// 0. PRIMERO: Verificar whitelist (dominios legítimos conocidos)
	if domain != "" {
		var brand sql.NullString
		var category sql.NullString
		err := c.db.QueryRowContext(ctx, `
			SELECT brand, category
			FROM whitelist_domains
			WHERE domain = $1
			LIMIT 1
		`, domain).Scan(&brand, &category)

		if err == nil {
			// Dominio está en whitelist - marcar como seguro
			result.RawData["whitelisted"] = true
			result.RawData["is_safe"] = true
			if brand.Valid {
				result.RawData["brand"] = brand.String
				reasons = append(reasons, fmt.Sprintf("Dominio oficial verificado de %s", brand.String))
			} else {
				reasons = append(reasons, "Dominio verificado como legítimo")
			}
			if category.Valid {
				result.RawData["category"] = category.String
			}
			result.ThreatType = "safe"
			result.Confidence = 1.0

			log.Debug().
				Str("domain", domain).
				Str("brand", brand.String).
				Msg("[LocalDB] Domain found in whitelist - marked as SAFE")

			result.RawData["reasons"] = reasons
			result.Latency = time.Since(startTime)
			return result, nil
		}
	}

	// 1. Buscar dominio usando la función optimizada find_threat_domain
	if domain != "" {
		var domainHash []byte
		var domainStr, threatType, severity string
		var confidence int16
		var impersonates sql.NullString

		log.Debug().
			Str("domain", domain).
			Msg("[LocalDB] Searching domain with find_threat_domain")

		err := c.db.QueryRowContext(ctx, `
			SELECT domain_hash, domain, threat_type, severity, confidence, impersonates
			FROM find_threat_domain($1)
			LIMIT 1
		`, domain).Scan(&domainHash, &domainStr, &threatType, &severity, &confidence, &impersonates)

		if err == nil {
			result.Found = true
			result.ThreatType = threatType
			result.Confidence = float64(confidence) / 100.0 // Convertir 0-100 a 0.0-1.0
			if impersonates.Valid {
				reasons = append(reasons, fmt.Sprintf("Dominio malicioso que suplanta a %s", impersonates.String))
				result.RawData["impersonates"] = impersonates.String
			} else {
				reasons = append(reasons, fmt.Sprintf("Dominio en lista negra (%s)", threatType))
			}
			result.RawData["severity"] = severity

			// Obtener tags del dominio
			result.Tags = c.getDomainTags(ctx, domainHash)
		} else if err != sql.ErrNoRows {
			log.Debug().Err(err).Msg("[LocalDB] Error querying find_threat_domain")
		}
	}

	// 2. Si no encontramos el dominio, buscar path específico
	if !result.Found && indicators.Path != "" && domain != "" {
		var threatType, severity string
		var confidence int16

		err := c.db.QueryRowContext(ctx, `
			SELECT tp.threat_type, tp.severity, tp.confidence
			FROM threat_paths tp
			JOIN threat_domains td ON tp.domain_hash = td.domain_hash
			WHERE td.domain = $1
			  AND tp.path = $2
			  AND (tp.flags & 1) = 1
			LIMIT 1
		`, domain, indicators.Path).Scan(&threatType, &severity, &confidence)

		if err == nil {
			result.Found = true
			result.ThreatType = threatType
			result.Confidence = float64(confidence) / 100.0
			result.RawData["severity"] = severity
			reasons = append(reasons, fmt.Sprintf("Path malicioso encontrado (%s)", threatType))
		}
	}

	// 3. Detección heurística: verificar si el dominio intenta suplantar una marca conocida
	if !result.Found && domain != "" {
		if brand, official := c.detectBrandImpersonation(ctx, domain); brand != "" {
			result.Found = true
			result.ThreatType = "phishing"
			result.Confidence = 0.75 // Confianza media-alta por heurística
			result.RawData["severity"] = "high"
			result.RawData["impersonates"] = brand
			result.RawData["official_domain"] = official
			reasons = append(reasons, fmt.Sprintf("Dominio sospechoso que parece suplantar a %s (dominio oficial: %s)", brand, official))

			log.Info().
				Str("domain", domain).
				Str("brand", brand).
				Str("official", official).
				Msg("[LocalDB] Heuristic: URL brand impersonation detected")
		}
	}

	result.RawData["reasons"] = reasons
	result.Latency = time.Since(startTime)

	log.Debug().
		Str("url", indicators.FullURL).
		Bool("found", result.Found).
		Str("threat_type", result.ThreatType).
		Dur("latency", result.Latency).
		Msg("[LocalDB] URL check completed")

	return result, nil
}

// getDomainTags obtiene los tags de un dominio
func (c *LocalDBChecker) getDomainTags(ctx context.Context, domainHash []byte) []string {
	rows, err := c.db.QueryContext(ctx, `
		SELECT t.name
		FROM domain_tags dt
		JOIN tags t ON dt.tag_id = t.id
		WHERE dt.domain_hash = $1
	`, domainHash)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err == nil {
			tags = append(tags, tag)
		}
	}
	return tags
}

// checkEmail verifica un email en la base de datos (Schema v2.0 optimizado)
func (c *LocalDBChecker) checkEmail(ctx context.Context, indicators *Indicators, startTime time.Time) (*CheckResult, error) {
	result := &CheckResult{
		Source: c.Name(),
		Found:  false,
		RawData: map[string]interface{}{
			"reasons": []string{},
		},
	}

	var reasons []string
	email := strings.ToLower(indicators.Normalized)

	// 1. Buscar email exacto por hash BYTEA
	var threatType, severity string
	var confidence int16
	var impersonates sql.NullString
	var flags int16

	err := c.db.QueryRowContext(ctx, `
		SELECT threat_type, severity, confidence, impersonates, flags
		FROM threat_emails
		WHERE email_hash = sha256_bytea($1) AND (flags & 1) = 1
		LIMIT 1
	`, email).Scan(&threatType, &severity, &confidence, &impersonates, &flags)

	if err == nil {
		result.Found = true
		result.ThreatType = threatType
		result.Confidence = float64(confidence) / 100.0
		result.RawData["severity"] = severity
		if impersonates.Valid {
			reasons = append(reasons, fmt.Sprintf("Email fraudulento que suplanta a %s", impersonates.String))
			result.RawData["impersonates"] = impersonates.String
		} else {
			reasons = append(reasons, fmt.Sprintf("Email en lista negra (%s)", threatType))
		}
	} else if err != sql.ErrNoRows {
		log.Debug().Err(err).Msg("[LocalDB] Error querying threat_emails")
	}

	// 2. Verificar si el dominio del email está en threat_domains
	if !result.Found && indicators.EmailDomain != "" {
		emailDomain := strings.ToLower(indicators.EmailDomain)
		var domThreatType, domSeverity string
		var domConfidence int16

		err := c.db.QueryRowContext(ctx, `
			SELECT threat_type, severity, confidence
			FROM find_threat_domain($1)
			LIMIT 1
		`, emailDomain).Scan(&domThreatType, &domSeverity, &domConfidence)

		if err == nil {
			result.Found = true
			result.ThreatType = domThreatType
			result.Confidence = float64(domConfidence) / 100.0 * 0.9 // Ligeramente menor confianza
			result.RawData["severity"] = domSeverity
			reasons = append(reasons, "Dominio del email en lista negra")
		}
	}

	// 3. Detección heurística: verificar si el dominio intenta suplantar una marca conocida
	if !result.Found && indicators.EmailDomain != "" {
		if brand, official := c.detectBrandImpersonation(ctx, indicators.EmailDomain); brand != "" {
			result.Found = true
			result.ThreatType = "phishing"
			result.Confidence = 0.75 // Confianza media-alta por heurística
			result.RawData["severity"] = "high"
			result.RawData["impersonates"] = brand
			result.RawData["official_domain"] = official
			reasons = append(reasons, fmt.Sprintf("Dominio sospechoso que parece suplantar a %s (dominio oficial: %s)", brand, official))

			log.Info().
				Str("domain", indicators.EmailDomain).
				Str("brand", brand).
				Str("official", official).
				Msg("[LocalDB] Heuristic: Brand impersonation detected")
		}
	}

	result.RawData["reasons"] = reasons
	result.Latency = time.Since(startTime)

	log.Debug().
		Str("email", indicators.Normalized).
		Bool("found", result.Found).
		Dur("latency", result.Latency).
		Msg("[LocalDB] Email check completed")

	return result, nil
}

// checkPhone verifica un teléfono en la base de datos (Schema v2.0 optimizado)
func (c *LocalDBChecker) checkPhone(ctx context.Context, indicators *Indicators, startTime time.Time) (*CheckResult, error) {
	result := &CheckResult{
		Source: c.Name(),
		Found:  false,
		RawData: map[string]interface{}{
			"reasons": []string{},
		},
	}

	var reasons []string

	// Normalizar teléfono: extraer solo los últimos 9 dígitos (número nacional español)
	phoneNational := indicators.Normalized
	if len(phoneNational) > 9 {
		// Quitar prefijo internacional (+34, 0034, etc.)
		phoneNational = phoneNational[len(phoneNational)-9:]
	}

	// Buscar teléfono por número nacional (PK directa, muy eficiente)
	var threatType, severity string
	var confidence int16
	var description sql.NullString
	var flags int16

	err := c.db.QueryRowContext(ctx, `
		SELECT threat_type, severity, confidence, description, flags
		FROM threat_phones
		WHERE phone_national = $1 AND (flags & 1) = 1
		LIMIT 1
	`, phoneNational).Scan(&threatType, &severity, &confidence, &description, &flags)

	if err == nil {
		result.Found = true
		result.ThreatType = threatType
		result.Confidence = float64(confidence) / 100.0
		result.RawData["severity"] = severity

		// Check bit 1 (is_premium)
		isPremium := (flags & 2) == 2
		if isPremium {
			reasons = append(reasons, "Número premium fraudulento reportado")
			result.RawData["is_premium"] = true
		}
		if description.Valid {
			reasons = append(reasons, description.String)
		} else {
			reasons = append(reasons, fmt.Sprintf("Teléfono en lista negra (%s)", threatType))
		}
	} else if err != sql.ErrNoRows {
		log.Debug().Err(err).Msg("[LocalDB] Error querying threat_phones")
	}

	// Verificar prefijos de alto riesgo (806, 807, 905, 803, 804)
	if !result.Found && len(phoneNational) >= 3 {
		prefix := phoneNational[:3]
		highRiskPrefixes := map[string]string{
			"803": "Línea premium adultos",
			"806": "Línea premium tarot/ocio",
			"807": "Línea premium profesional",
			"905": "Línea premium masivo",
		}
		if desc, isHighRisk := highRiskPrefixes[prefix]; isHighRisk {
			result.RawData["high_risk_prefix"] = true
			result.RawData["prefix_description"] = desc
			reasons = append(reasons, fmt.Sprintf("Prefijo de alto riesgo: %s (%s)", prefix, desc))
		}
	}

	result.RawData["reasons"] = reasons
	result.Latency = time.Since(startTime)

	log.Debug().
		Str("phone", indicators.Normalized).
		Str("national", phoneNational).
		Bool("found", result.Found).
		Dur("latency", result.Latency).
		Msg("[LocalDB] Phone check completed")

	return result, nil
}

// Name retorna el nombre del checker
func (c *LocalDBChecker) Name() string {
	return "localdb"
}

// Weight retorna el peso del checker
func (c *LocalDBChecker) Weight() float64 {
	return c.weight
}

// IsEnabled indica si el checker está habilitado
func (c *LocalDBChecker) IsEnabled() bool {
	return c.enabled
}

// SupportedTypes retorna los tipos de input soportados
func (c *LocalDBChecker) SupportedTypes() []InputType {
	return []InputType{InputTypeURL, InputTypeEmail, InputTypePhone}
}

// Close cierra la conexión a la base de datos
func (c *LocalDBChecker) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// GetStats retorna estadísticas de la base de datos (Schema v2.0)
func (c *LocalDBChecker) GetStats(ctx context.Context) (map[string]interface{}, error) {
	if !c.enabled || c.db == nil {
		return nil, fmt.Errorf("checker disabled")
	}

	stats := make(map[string]interface{})

	// Usar la vista threat_stats optimizada
	rows, err := c.db.QueryContext(ctx, `SELECT type, total, active, phishing, malware, scam, size FROM threat_stats`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tableType, size string
			var total, active, phishing, malware, scam int64
			if err := rows.Scan(&tableType, &total, &active, &phishing, &malware, &scam, &size); err == nil {
				stats[tableType] = map[string]interface{}{
					"total":    total,
					"active":   active,
					"phishing": phishing,
					"malware":  malware,
					"scam":     scam,
					"size":     size,
				}
			}
		}
	} else {
		// Fallback: contar manualmente usando flags
		type tableQuery struct {
			name  string
			query string
		}
		queries := []tableQuery{
			{"threat_domains", "SELECT COUNT(*) FROM threat_domains WHERE (flags & 1) = 1"},
			{"threat_paths", "SELECT COUNT(*) FROM threat_paths WHERE (flags & 1) = 1"},
			{"threat_emails", "SELECT COUNT(*) FROM threat_emails WHERE (flags & 1) = 1"},
			{"threat_phones", "SELECT COUNT(*) FROM threat_phones WHERE (flags & 1) = 1"},
			{"whitelist_domains", "SELECT COUNT(*) FROM whitelist_domains"},
		}
		for _, q := range queries {
			var count int64
			if err := c.db.QueryRowContext(ctx, q.query).Scan(&count); err == nil {
				stats[q.name] = count
			}
		}
	}

	return stats, nil
}

// maskDatabaseURL oculta la contraseña de la URL de conexión
func maskDatabaseURL(url string) string {
	// Simplificación: mostrar solo el host
	if len(url) > 30 {
		return url[:30] + "..."
	}
	return url
}

// detectBrandImpersonation detecta si un dominio intenta suplantar una marca conocida
// Busca patrones como "bbva-algo.es", "santander-login.com", etc.
// Retorna (brand, official_domain) si detecta suplantación, o ("", "") si no
func (c *LocalDBChecker) detectBrandImpersonation(ctx context.Context, domain string) (string, string) {
	domain = strings.ToLower(domain)

	// Obtener todas las marcas de la whitelist
	rows, err := c.db.QueryContext(ctx, `
		SELECT LOWER(brand), domain
		FROM whitelist_domains
		WHERE brand IS NOT NULL AND brand != ''
	`)
	if err != nil {
		log.Debug().Err(err).Msg("[LocalDB] Error fetching whitelist brands")
		return "", ""
	}
	defer rows.Close()

	for rows.Next() {
		var brand, officialDomain string
		if err := rows.Scan(&brand, &officialDomain); err != nil {
			continue
		}

		brandLower := strings.ToLower(brand)

		// Si el dominio ya es el oficial, no es suplantación
		if domain == officialDomain {
			return "", ""
		}

		// Detectar patrones de suplantación:
		// 1. Dominio contiene la marca pero no es el oficial: bbva-compra.es, mi-santander.com
		// 2. Marca seguida de guión o punto: bbva-login, santander.verify
		// 3. Marca al inicio seguida de palabras sospechosas

		// Verificar si el dominio contiene el nombre de la marca
		if strings.Contains(domain, brandLower) {
			// Palabras sospechosas comunes en phishing
			suspiciousWords := []string{
				"login", "verify", "secure", "update", "confirm", "account",
				"seguro", "verificar", "actualizar", "confirmar", "cuenta",
				"compra", "pago", "factura", "incidencia", "soporte", "ayuda",
				"cliente", "acceso", "password", "clave", "tarjeta", "banco",
			}

			// Verificar si tiene alguna palabra sospechosa
			for _, word := range suspiciousWords {
				if strings.Contains(domain, word) {
					log.Debug().
						Str("domain", domain).
						Str("brand", brand).
						Str("suspicious_word", word).
						Msg("[LocalDB] Brand impersonation pattern detected")
					return brand, officialDomain
				}
			}

			// Si contiene la marca con guión o guión bajo, es sospechoso
			if strings.Contains(domain, brandLower+"-") ||
				strings.Contains(domain, brandLower+"_") ||
				strings.Contains(domain, "-"+brandLower) ||
				strings.Contains(domain, "_"+brandLower) {
				return brand, officialDomain
			}

			// Si el dominio empieza con la marca pero tiene más texto, verificar
			if strings.HasPrefix(domain, brandLower) && domain != officialDomain {
				// Extraer lo que viene después de la marca
				suffix := domain[len(brandLower):]
				// Si hay más texto que no es solo el TLD
				if len(suffix) > 4 && !strings.HasPrefix(suffix, ".es") &&
					!strings.HasPrefix(suffix, ".com") {
					return brand, officialDomain
				}
			}
		}
	}

	return "", ""
}
