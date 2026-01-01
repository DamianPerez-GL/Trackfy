package urlengine

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trackfy/fy-analysis/internal/checkers"
	"github.com/trackfy/fy-analysis/internal/correlation"
	"github.com/trackfy/fy-analysis/internal/sync"
)

// Engine es el motor principal de verificación de URLs, emails y teléfonos
type Engine struct {
	orchestrator       *Orchestrator
	normalizer         *Normalizer
	aggregator         *Aggregator
	heuristics         *correlation.HeuristicEngine
	dbSyncer           *sync.DBSyncer
	userReportsChecker *checkers.UserReportsChecker
	config             *EngineConfig
}

// EngineConfig configuración del engine
type EngineConfig struct {
	CheckTimeout       time.Duration
	URLhausDBPath      string
	PhishTankDBPath    string
	GoogleWebRiskKey   string
	URLScanKey         string
	PhishTankKey       string
	EnableDBSync       bool
	DatabaseURL        string
	EnableLocalDB      bool
	EnableUserReports  bool // Habilitar checker de reportes de usuarios
}

// DefaultConfig retorna la configuración por defecto
func DefaultConfig() *EngineConfig {
	return &EngineConfig{
		CheckTimeout:      3 * time.Second,
		URLhausDBPath:     getEnv("URLHAUS_DB_PATH", "/app/data/urlhaus.csv"),
		PhishTankDBPath:   getEnv("PHISHTANK_DB_PATH", "/app/data/phishtank.json"),
		GoogleWebRiskKey:  getEnv("GOOGLE_WEBRISK_KEY", ""),
		URLScanKey:        getEnv("URLSCAN_KEY", ""),
		PhishTankKey:      getEnv("PHISHTANK_KEY", ""),
		EnableDBSync:      getEnv("ENABLE_DB_SYNC", "true") == "true",
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		EnableLocalDB:     getEnv("ENABLE_LOCAL_DB", "true") == "true",
		EnableUserReports: getEnv("ENABLE_USER_REPORTS", "true") == "true",
	}
}

// NewEngine crea un nuevo URL verification engine
func NewEngine(config *EngineConfig) *Engine {
	if config == nil {
		config = DefaultConfig()
	}

	log.Info().
		Dur("timeout", config.CheckTimeout).
		Str("urlhaus_db", config.URLhausDBPath).
		Str("phishtank_db", config.PhishTankDBPath).
		Bool("webrisk_enabled", config.GoogleWebRiskKey != "").
		Bool("urlscan_enabled", config.URLScanKey != "").
		Bool("db_sync", config.EnableDBSync).
		Msg("[Engine] Initializing URL Verification Engine")

	// Crear checkers
	var threatCheckers []checkers.ThreatChecker

	// URLhaus (local DB)
	urlhausChecker := checkers.NewURLhausChecker(config.URLhausDBPath)
	threatCheckers = append(threatCheckers, urlhausChecker)
	log.Info().Msg("[Engine] URLhaus checker initialized")

	// PhishTank (local DB)
	phishtankChecker := checkers.NewPhishTankChecker(config.PhishTankDBPath, config.PhishTankKey)
	threatCheckers = append(threatCheckers, phishtankChecker)
	log.Info().Msg("[Engine] PhishTank checker initialized")

	// Google Web Risk (API)
	if config.GoogleWebRiskKey != "" {
		webRiskChecker := checkers.NewWebRiskChecker(config.GoogleWebRiskKey)
		threatCheckers = append(threatCheckers, webRiskChecker)
		log.Info().Msg("[Engine] Google Web Risk checker initialized")
	}

	// URLScan.io (API)
	if config.URLScanKey != "" {
		urlscanChecker := checkers.NewURLScanChecker(config.URLScanKey)
		threatCheckers = append(threatCheckers, urlscanChecker)
		log.Info().Msg("[Engine] URLScan.io checker initialized")
	}

	// LocalDB (PostgreSQL) - Prioridad alta
	log.Debug().
		Bool("enable_local_db", config.EnableLocalDB).
		Bool("has_database_url", config.DatabaseURL != "").
		Msg("[Engine] Checking LocalDB configuration")

	// Variable para guardar la conexión DB para otros checkers
	var localDBChecker *checkers.LocalDBChecker

	if config.EnableLocalDB && config.DatabaseURL != "" {
		log.Info().Msg("[Engine] Initializing LocalDB checker...")
		localDBChecker = checkers.NewLocalDBChecker(&checkers.LocalDBConfig{
			DatabaseURL: config.DatabaseURL,
			MaxConns:    10,
			Weight:      0.50, // Peso alto para DB local
		})
		if localDBChecker.IsEnabled() {
			// Insertar al inicio para mayor prioridad
			threatCheckers = append([]checkers.ThreatChecker{localDBChecker}, threatCheckers...)
			log.Info().Msg("[Engine] LocalDB checker initialized (PostgreSQL)")
		} else {
			log.Warn().Msg("[Engine] LocalDB checker failed to initialize")
		}
	} else {
		log.Info().Msg("[Engine] LocalDB checker disabled or no DATABASE_URL")
	}

	// UserReports Checker - Reportes de usuarios con anti-spam
	var userReportsChecker *checkers.UserReportsChecker
	if config.EnableUserReports && config.DatabaseURL != "" {
		log.Info().Msg("[Engine] Initializing UserReports checker...")
		userReportsChecker = checkers.NewUserReportsChecker(
			localDBChecker.GetDB(),
			&checkers.UserReportsConfig{
				Weight:             0.10, // Peso bajo por ser crowdsourced
				MinScoreForWarning: 40,
				MinScoreForDanger:  70,
				MinReportersForUse: 2,
			},
		)
		if userReportsChecker.IsEnabled() {
			threatCheckers = append(threatCheckers, userReportsChecker)
			log.Info().Msg("[Engine] UserReports checker initialized")
		}
	}

	// Crear orchestrator
	orchestrator := NewOrchestrator(threatCheckers, config.CheckTimeout)

	// Crear syncer para DBs locales
	var dbSyncer *sync.DBSyncer
	if config.EnableDBSync {
		dbSyncer = sync.NewDBSyncer(urlhausChecker, phishtankChecker)
	}

	engine := &Engine{
		orchestrator:       orchestrator,
		normalizer:         NewNormalizer(),
		aggregator:         NewAggregator(),
		heuristics:         correlation.NewHeuristicEngine(),
		dbSyncer:           dbSyncer,
		userReportsChecker: userReportsChecker,
		config:             config,
	}

	log.Info().
		Int("checkers", len(threatCheckers)).
		Msg("[Engine] Threat Analysis Engine initialized")

	return engine
}

// Start inicia el engine (sincronización de DBs en background)
func (e *Engine) Start(ctx context.Context) {
	log.Info().Msg("[Engine] Starting engine...")

	if e.dbSyncer != nil && e.config.EnableDBSync {
		e.dbSyncer.Start(ctx)
		log.Info().Msg("[Engine] DB synchronization started")
	}
}

// Stop detiene el engine
func (e *Engine) Stop() {
	log.Info().Msg("[Engine] Stopping engine...")

	if e.dbSyncer != nil {
		e.dbSyncer.Stop()
	}
}

// Check verifica una URL (método legacy para compatibilidad)
func (e *Engine) Check(ctx context.Context, rawURL string) *URLCheckResponse {
	log.Debug().Str("url", rawURL).Msg("[Engine] Check request received")
	return e.orchestrator.Check(ctx, rawURL)
}

// Analyze es el punto de entrada unificado para analizar URLs, emails o teléfonos
func (e *Engine) Analyze(ctx context.Context, req *AnalysisRequest) *AnalysisResponse {
	startTime := time.Now()

	log.Info().
		Str("input", req.Input).
		Str("type", string(req.Type)).
		Msg("[Engine] Analyze request received")

	// 1. Normalizar input
	indicators, err := e.normalizer.NormalizeInput(ctx, req.Input, req.Type)
	if err != nil {
		log.Error().Err(err).Msg("[Engine] Normalization failed")
		return e.buildErrorAnalysisResponse(req, err.Error(), startTime)
	}

	log.Debug().
		Str("normalized", indicators.Normalized).
		Str("hash", indicators.Hash).
		Msg("[Engine] Input normalized")

	// TODO: 2. Check cache (Redis) - pendiente de implementar
	// if cached, err := e.cache.Get(ctx, indicators.Hash); err == nil {
	//     cached.CacheHit = true
	//     return cached
	// }

	// 3. Búsqueda paralela en motores (filtrada por tipo)
	results := e.orchestrator.CheckWithType(ctx, indicators)

	log.Debug().
		Int("checker_results", len(results)).
		Msg("[Engine] Checker results received")

	// 4. Correlación heurística
	heuristicResult := e.heuristics.Analyze(ctx, indicators, req.Context)
	if heuristicResult.Score > 0 {
		results = append(results, e.heuristics.ToCheckResult(heuristicResult))
		log.Debug().
			Int("heuristic_score", heuristicResult.Score).
			Strs("heuristic_flags", heuristicResult.Flags).
			Msg("[Engine] Heuristic analysis added")
	}

	// 5. Agregar resultados y calcular score
	score, level, reasons := e.aggregateAnalysisResults(results, heuristicResult)

	// 6. Construir respuesta
	response := &AnalysisResponse{
		Input:             req.Input,
		Type:              req.Type,
		NormalizedInput:   indicators.Normalized,
		RiskScore:         score,
		RiskLevel:         string(level),
		Threats:           e.buildThreatDetails(results),
		Reasons:           reasons,
		RecommendedAction: GetActionForLevel(level),
		Sources:           e.buildSourceResults(results),
		CacheHit:          false,
		ResponseTimeMs:    time.Since(startTime).Milliseconds(),
		CheckedAt:         time.Now().UTC(),
	}

	// TODO: 7. Cachear en Redis
	// e.cache.Set(ctx, indicators.Hash, response, 24*time.Hour)

	log.Info().
		Str("input", req.Input).
		Int("risk_score", response.RiskScore).
		Str("risk_level", response.RiskLevel).
		Int64("response_ms", response.ResponseTimeMs).
		Msg("[Engine] Analysis completed")

	return response
}

// aggregateAnalysisResults agrega resultados de checkers y heurísticas
func (e *Engine) aggregateAnalysisResults(results []*checkers.CheckResult, heuristic *correlation.HeuristicResult) (int, RiskLevel, []string) {
	var reasons []string
	var totalScore float64
	var totalWeight float64
	threatsFound := 0

	// PRIMERO: Verificar si algún checker marcó el dominio como whitelisted (SAFE)
	for _, result := range results {
		if result.RawData != nil {
			if isSafe, ok := result.RawData["is_safe"].(bool); ok && isSafe {
				// Dominio está en whitelist - retornar como seguro
				var safeReasons []string
				if rawReasons, ok := result.RawData["reasons"].([]string); ok {
					safeReasons = rawReasons
				}
				if len(safeReasons) == 0 {
					if brand, ok := result.RawData["brand"].(string); ok && brand != "" {
						safeReasons = []string{fmt.Sprintf("Dominio oficial verificado de %s", brand)}
					} else {
						safeReasons = []string{"Dominio verificado como legítimo"}
					}
				}

				log.Info().
					Strs("reasons", safeReasons).
					Msg("[Engine] Domain is WHITELISTED - returning safe")

				return 0, RiskLevelSafe, safeReasons
			}
		}
	}

	// Pesos por fuente (LocalDB tiene mayor peso)
	weights := map[string]float64{
		"localdb":      0.30, // DB local - máxima prioridad
		"urlhaus":      0.15,
		"phishtank":    0.15,
		"webrisk":      0.15,
		"urlscan":      0.10,
		"user_reports": 0.10, // Reportes de usuarios - peso bajo (crowdsourced)
		"heuristics":   0.15,
	}

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		weight := weights[result.Source]
		if weight == 0 {
			weight = 0.1 // Default para fuentes desconocidas
		}
		totalWeight += weight

		if result.Found {
			threatsFound++
			contribution := weight * result.Confidence * 100
			totalScore += contribution

			// Añadir razones del resultado
			if rawReasons, ok := result.RawData["reasons"].([]string); ok {
				reasons = append(reasons, rawReasons...)
			}
		}
	}

	// Añadir razones de heurísticas (evitando duplicados)
	if heuristic != nil && len(heuristic.Reasons) > 0 {
		seen := make(map[string]bool)
		for _, r := range reasons {
			seen[r] = true
		}
		for _, r := range heuristic.Reasons {
			if !seen[r] {
				reasons = append(reasons, r)
				seen[r] = true
			}
		}
	}

	// Calcular score final
	var finalScore int
	if totalWeight > 0 && threatsFound > 0 {
		finalScore = int(totalScore / totalWeight)

		// Boost por múltiples fuentes
		if threatsFound >= 2 {
			boost := 1.0 + (float64(threatsFound-1) * 0.1)
			finalScore = int(float64(finalScore) * boost)
		}

		if finalScore > 100 {
			finalScore = 100
		}
	} else if totalWeight == 0 {
		finalScore = 50 // Incertidumbre
		reasons = append(reasons, ReasonsES["partial_check"])
	} else {
		finalScore = 0
		if len(reasons) == 0 {
			reasons = append(reasons, ReasonsES["no_threats_found"])
		}
	}

	level := GetRiskLevel(finalScore)
	return finalScore, level, reasons
}

// buildThreatDetails construye los detalles de amenazas desde los resultados
func (e *Engine) buildThreatDetails(results []*checkers.CheckResult) []ThreatDetail {
	var threats []ThreatDetail

	for _, result := range results {
		if result.Found && result.Error == nil {
			threats = append(threats, ThreatDetail{
				Source:     result.Source,
				Type:       result.ThreatType,
				Confidence: result.Confidence,
				Tags:       result.Tags,
			})
		}
	}

	return threats
}

// buildSourceResults construye los resultados por fuente
func (e *Engine) buildSourceResults(results []*checkers.CheckResult) []SourceResult {
	var sources []SourceResult

	weights := map[string]float64{
		"localdb":      0.30,
		"urlhaus":      0.15,
		"phishtank":    0.15,
		"webrisk":      0.15,
		"urlscan":      0.10,
		"user_reports": 0.10,
		"heuristics":   0.15,
	}

	for _, result := range results {
		sr := SourceResult{
			Name:    result.Source,
			Found:   result.Found,
			Latency: result.Latency.String(),
			Weight:  weights[result.Source],
		}
		if result.Error != nil {
			sr.Error = result.Error.Error()
		}
		sources = append(sources, sr)
	}

	return sources
}

// buildErrorAnalysisResponse construye respuesta de error
func (e *Engine) buildErrorAnalysisResponse(req *AnalysisRequest, errorMsg string, startTime time.Time) *AnalysisResponse {
	return &AnalysisResponse{
		Input:             req.Input,
		Type:              req.Type,
		NormalizedInput:   req.Input,
		RiskScore:         50,
		RiskLevel:         string(RiskLevelWarning),
		Threats:           []ThreatDetail{},
		Reasons:           []string{"No se pudo analizar: " + errorMsg},
		RecommendedAction: ActionCaution,
		Sources:           []SourceResult{},
		CacheHit:          false,
		ResponseTimeMs:    time.Since(startTime).Milliseconds(),
		CheckedAt:         time.Now().UTC(),
	}
}

// GetStatus retorna el estado del engine
func (e *Engine) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"checkers": e.orchestrator.GetCheckerStatus(),
	}

	if e.dbSyncer != nil {
		status["databases"] = e.dbSyncer.GetStatus()
	}

	return status
}

// ForceDBSync fuerza sincronización de DBs
func (e *Engine) ForceDBSync(ctx context.Context, dbName string) error {
	if e.dbSyncer != nil {
		return e.dbSyncer.ForceSync(ctx, dbName)
	}
	return nil
}

// ReportURLRequest estructura para reportar una URL
type ReportURLRequest struct {
	URL           string `json:"url"`
	UserID        string `json:"user_id"`
	ThreatType    string `json:"threat_type"`    // phishing, malware, scam, spam, other
	Description   string `json:"description"`    // Descripción opcional
	ReportContext string `json:"report_context"` // chat, manual, browser_extension
	UserIP        string `json:"user_ip"`
	UserAgent     string `json:"user_agent"`
}

// ReportURLResponse respuesta al reportar una URL
type ReportURLResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	URLScore    int    `json:"url_score"`     // Score actual de la URL tras el reporte
	IsNewReport bool   `json:"is_new_report"` // Si es el primer reporte de esta URL
}

// ReportURL permite a un usuario reportar una URL como sospechosa
func (e *Engine) ReportURL(ctx context.Context, req *ReportURLRequest) *ReportURLResponse {
	log.Info().
		Str("url", req.URL).
		Str("user_id", req.UserID).
		Str("threat_type", req.ThreatType).
		Msg("[Engine] Report URL request received")

	if e.userReportsChecker == nil || !e.userReportsChecker.IsEnabled() {
		log.Warn().Msg("[Engine] UserReports checker not enabled")
		return &ReportURLResponse{
			Success:  false,
			Message:  "Servicio de reportes no disponible",
			URLScore: 0,
		}
	}

	// Normalizar y extraer dominio
	indicators, err := e.normalizer.NormalizeInput(ctx, req.URL, checkers.InputTypeURL)
	if err != nil {
		return &ReportURLResponse{
			Success:  false,
			Message:  "URL inválida: " + err.Error(),
			URLScore: 0,
		}
	}

	// Llamar al checker para registrar el reporte
	success, message, score, err := e.userReportsChecker.ReportURL(
		ctx,
		indicators.Normalized,
		indicators.Domain,
		req.UserID,
		req.ThreatType,
		req.Description,
		req.ReportContext,
		req.UserIP,
		req.UserAgent,
	)

	if err != nil {
		log.Error().Err(err).Msg("[Engine] Error reporting URL")
		return &ReportURLResponse{
			Success:  false,
			Message:  "Error al procesar reporte",
			URLScore: 0,
		}
	}

	log.Info().
		Bool("success", success).
		Int("score", score).
		Str("url", indicators.Normalized).
		Msg("[Engine] URL report processed")

	return &ReportURLResponse{
		Success:  success,
		Message:  message,
		URLScore: score,
	}
}

// GetUserReportsStats retorna estadísticas del sistema de reportes
func (e *Engine) GetUserReportsStats(ctx context.Context) (map[string]interface{}, error) {
	if e.userReportsChecker == nil || !e.userReportsChecker.IsEnabled() {
		return nil, fmt.Errorf("user reports checker not enabled")
	}
	return e.userReportsChecker.GetStats(ctx)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
