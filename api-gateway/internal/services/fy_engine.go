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

type FyEngineClient struct {
	baseURL    string
	httpClient *http.Client
}

// FyChatRequest request al chat de Fy
type FyChatRequest struct {
	UserID  string          `json:"user_id"`
	Message string          `json:"message"`
	Context []ContextMessage `json:"context"`
}

type ContextMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// FyChatResponse respuesta del chat de Fy
type FyChatResponse struct {
	Response          string `json:"response"`
	Mood              string `json:"mood"`
	PIIDetected       bool   `json:"pii_detected"`
	Intent            string `json:"intent"`
	AnalysisPerformed bool   `json:"analysis_performed"`
	Error             string `json:"error,omitempty"`
}

func NewFyEngineClient(baseURL string, timeout time.Duration) *FyEngineClient {
	return &FyEngineClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Chat envía un mensaje al chat de Fy
func (c *FyEngineClient) Chat(ctx context.Context, userID, message string, conversationContext []ContextMessage) (*FyChatResponse, error) {
	reqBody := FyChatRequest{
		UserID:  userID,
		Message: message,
		Context: conversationContext,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	log.Debug().
		Str("user_id", userID).
		Str("message", truncate(message, 50)).
		Msg("[FyEngine] Sending chat request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("[FyEngine] Request failed")
		return nil, fmt.Errorf("fy-engine returned status %d", resp.StatusCode)
	}

	var fyResp FyChatResponse
	if err := json.Unmarshal(body, &fyResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Debug().
		Str("mood", fyResp.Mood).
		Str("intent", fyResp.Intent).
		Bool("analysis", fyResp.AnalysisPerformed).
		Msg("[FyEngine] Chat response received")

	return &fyResp, nil
}

// Health verifica si fy-engine está disponible
func (c *FyEngineClient) Health(ctx context.Context) bool {
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
