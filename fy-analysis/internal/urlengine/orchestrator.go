package urlengine

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trackfy/fy-analysis/internal/checkers"
)

// Orchestrator coordina la verificación paralela con múltiples checkers
type Orchestrator struct {
	checkers   []checkers.ThreatChecker
	timeout    time.Duration
	normalizer *Normalizer
	extractor  *Extractor
	aggregator *Aggregator
}

// NewOrchestrator crea un nuevo orchestrator
func NewOrchestrator(threatCheckers []checkers.ThreatChecker, timeout time.Duration) *Orchestrator {
	return &Orchestrator{
		checkers:   threatCheckers,
		timeout:    timeout,
		normalizer: NewNormalizer(),
		extractor:  NewExtractor(),
		aggregator: NewAggregator(),
	}
}

// Check realiza la verificación completa de una URL
func (o *Orchestrator) Check(ctx context.Context, rawURL string) *URLCheckResponse {
	startTime := time.Now()

	log.Info().
		Str("url", rawURL).
		Int("checkers", len(o.checkers)).
		Dur("timeout", o.timeout).
		Msg("[Orchestrator] Starting URL check")

	// 1. Normalizar URL
	normalized := o.normalizer.Normalize(ctx, rawURL)
	if normalized.Error != nil {
		log.Error().Err(normalized.Error).Msg("[Orchestrator] Normalization failed")
		return o.buildErrorResponse(rawURL, normalized.Error.Error(), startTime)
	}

	log.Debug().
		Str("normalized", normalized.NormalizedURL).
		Bool("is_shortener", normalized.IsShortener).
		Str("expanded", normalized.ExpandedURL).
		Msg("[Orchestrator] URL normalized")

	// 2. Extraer indicadores
	indicators := o.extractor.Extract(normalized)

	log.Debug().
		Str("domain", indicators.Domain).
		Str("ip", indicators.IP).
		Msg("[Orchestrator] Indicators extracted")

	// 3. Ejecutar checkers en paralelo
	results := o.runCheckersParallel(ctx, indicators)

	log.Info().
		Int("results", len(results)).
		Dur("elapsed", time.Since(startTime)).
		Msg("[Orchestrator] All checkers completed")

	// 4. Agregar resultados y calcular score
	response := o.aggregator.Aggregate(normalized, indicators, results, startTime)

	// Log resultado final
	log.Info().
		Str("url", rawURL).
		Int("risk_score", response.RiskScore).
		Str("risk_level", string(response.RiskLevel)).
		Int("threats_found", len(response.Threats)).
		Str("latency", response.Latency).
		Msg("[Orchestrator] Check completed")

	return response
}

// runCheckersParallel ejecuta todos los checkers en paralelo con timeout
func (o *Orchestrator) runCheckersParallel(ctx context.Context, indicators *checkers.Indicators) []*checkers.CheckResult {
	// Crear context con timeout
	checkCtx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	var wg sync.WaitGroup
	resultsChan := make(chan *checkers.CheckResult, len(o.checkers))

	// Filtrar solo checkers habilitados
	enabledCheckers := o.getEnabledCheckers()

	log.Debug().
		Int("enabled", len(enabledCheckers)).
		Int("total", len(o.checkers)).
		Msg("[Orchestrator] Running enabled checkers")

	// Lanzar goroutine por cada checker
	for _, checker := range enabledCheckers {
		wg.Add(1)
		go func(c checkers.ThreatChecker) {
			defer wg.Done()
			o.runSingleChecker(checkCtx, c, indicators, resultsChan)
		}(checker)
	}

	// Esperar a que terminen todos o timeout
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Recolectar resultados
	var results []*checkers.CheckResult
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// runSingleChecker ejecuta un checker individual con manejo de errores
func (o *Orchestrator) runSingleChecker(
	ctx context.Context,
	checker checkers.ThreatChecker,
	indicators *checkers.Indicators,
	resultsChan chan<- *checkers.CheckResult,
) {
	startTime := time.Now()
	checkerName := checker.Name()

	log.Debug().
		Str("checker", checkerName).
		Str("url", indicators.FullURL).
		Msg("[Orchestrator] Starting checker")

	// Ejecutar checker
	result, err := checker.Check(ctx, indicators)

	latency := time.Since(startTime)

	if err != nil {
		log.Warn().
			Err(err).
			Str("checker", checkerName).
			Dur("latency", latency).
			Msg("[Orchestrator] Checker failed")

		// Enviar resultado con error (graceful degradation)
		resultsChan <- &checkers.CheckResult{
			Source:  checkerName,
			Found:   false,
			Error:   err,
			Latency: latency,
		}
		return
	}

	result.Source = checkerName
	result.Latency = latency

	log.Debug().
		Str("checker", checkerName).
		Bool("found", result.Found).
		Str("threat_type", result.ThreatType).
		Float64("confidence", result.Confidence).
		Dur("latency", latency).
		Msg("[Orchestrator] Checker completed")

	resultsChan <- result
}

// getEnabledCheckers retorna solo los checkers habilitados
func (o *Orchestrator) getEnabledCheckers() []checkers.ThreatChecker {
	var enabled []checkers.ThreatChecker
	for _, c := range o.checkers {
		if c.IsEnabled() {
			enabled = append(enabled, c)
		}
	}
	return enabled
}

// getCheckersForType retorna los checkers habilitados que soportan un tipo de input
func (o *Orchestrator) getCheckersForType(inputType checkers.InputType) []checkers.ThreatChecker {
	var compatible []checkers.ThreatChecker
	for _, c := range o.checkers {
		if c.IsEnabled() && checkers.SupportsType(c, inputType) {
			compatible = append(compatible, c)
		}
	}
	return compatible
}

// CheckWithType realiza verificación filtrando por tipo de input
func (o *Orchestrator) CheckWithType(ctx context.Context, indicators *checkers.Indicators) []*checkers.CheckResult {
	// Crear context con timeout
	checkCtx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	var wg sync.WaitGroup

	// Filtrar checkers por tipo
	compatibleCheckers := o.getCheckersForType(indicators.InputType)
	resultsChan := make(chan *checkers.CheckResult, len(compatibleCheckers))

	log.Debug().
		Str("input_type", string(indicators.InputType)).
		Int("compatible_checkers", len(compatibleCheckers)).
		Msg("[Orchestrator] Running type-filtered checkers")

	// Lanzar goroutine por cada checker compatible
	for _, checker := range compatibleCheckers {
		wg.Add(1)
		go func(c checkers.ThreatChecker) {
			defer wg.Done()
			o.runSingleChecker(checkCtx, c, indicators, resultsChan)
		}(checker)
	}

	// Esperar a que terminen todos
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Recolectar resultados
	var results []*checkers.CheckResult
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// buildErrorResponse construye respuesta de error
func (o *Orchestrator) buildErrorResponse(rawURL, errorMsg string, startTime time.Time) *URLCheckResponse {
	return &URLCheckResponse{
		URL:           rawURL,
		NormalizedURL: rawURL,
		RiskScore:     50, // Score medio por incertidumbre
		RiskLevel:     RiskLevelWarning,
		Threats:       []ThreatDetail{},
		Explanation:   "No se pudo verificar la URL: " + errorMsg,
		Action:        "WARN",
		Sources:       []SourceResult{},
		CheckedAt:     time.Now().UTC(),
		Cached:        false,
		Latency:       time.Since(startTime).String(),
	}
}

// AddChecker añade un checker al orchestrator
func (o *Orchestrator) AddChecker(checker checkers.ThreatChecker) {
	o.checkers = append(o.checkers, checker)
	log.Info().
		Str("checker", checker.Name()).
		Float64("weight", checker.Weight()).
		Bool("enabled", checker.IsEnabled()).
		Msg("[Orchestrator] Checker added")
}

// GetCheckerStatus retorna el estado de todos los checkers
func (o *Orchestrator) GetCheckerStatus() []map[string]interface{} {
	var status []map[string]interface{}
	for _, c := range o.checkers {
		status = append(status, map[string]interface{}{
			"name":    c.Name(),
			"weight":  c.Weight(),
			"enabled": c.IsEnabled(),
		})
	}
	return status
}
