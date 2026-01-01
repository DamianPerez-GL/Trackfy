package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// FyAnalysisClient cliente para comunicarse con fy-analysis
type FyAnalysisClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewFyAnalysisClient crea un nuevo cliente de fy-analysis
func NewFyAnalysisClient(baseURL string, timeout time.Duration) *FyAnalysisClient {
	return &FyAnalysisClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ReportURLRequest petición para reportar una URL
type ReportURLRequest struct {
	URL         string `json:"url"`
	UserID      string `json:"user_id"`
	ThreatType  string `json:"threat_type"`
	Description string `json:"description,omitempty"`
	Context     string `json:"context,omitempty"`
}

// ReportURLResponse respuesta del reporte de URL
type ReportURLResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	URLScore    int    `json:"url_score"`
	IsNewReport bool   `json:"is_new_report,omitempty"`
	Error       string `json:"error,omitempty"`
}

// ReportURL envía un reporte de URL a fy-analysis
func (c *FyAnalysisClient) ReportURL(ctx context.Context, req *ReportURLRequest, userIP, userAgent string) (*ReportURLResponse, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/reports", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Forwarded-For", userIP)
	httpReq.Header.Set("User-Agent", userAgent)

	log.Debug().
		Str("url", req.URL).
		Str("user_id", req.UserID).
		Str("threat_type", req.ThreatType).
		Msg("[FyAnalysis] Sending report request")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Error().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("[FyAnalysis] Report request failed")
		return nil, fmt.Errorf("fy-analysis returned status %d", resp.StatusCode)
	}

	var reportResp ReportURLResponse
	if err := json.Unmarshal(body, &reportResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Debug().
		Bool("success", reportResp.Success).
		Int("url_score", reportResp.URLScore).
		Msg("[FyAnalysis] Report response received")

	return &reportResp, nil
}

// GetReportsStats obtiene estadísticas del sistema de reportes
func (c *FyAnalysisClient) GetReportsStats(ctx context.Context) (map[string]interface{}, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/reports/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fy-analysis returned status %d", resp.StatusCode)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return stats, nil
}

// Health verifica si fy-analysis está disponible
func (c *FyAnalysisClient) Health(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
