package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultModel = "gemini-2.5-flash"

type Client struct {
	apiKey     string
	model      string
	http       *http.Client
	limiter    <-chan time.Time
	maxRetries int
	retryDelay time.Duration
}

func NewClient(apiKey string) *Client {
	return NewClientWithLimits(apiKey, 100, 500*time.Millisecond)
}

func NewClientWithLimits(apiKey string, requestsPerSecond int, retryDelay time.Duration) *Client {
	interval := time.Second / time.Duration(requestsPerSecond)
	limiter := time.NewTicker(interval).C

	return &Client{
		apiKey:     apiKey,
		model:      defaultModel,
		http:       &http.Client{Timeout: 30 * time.Second},
		limiter:    limiter,
		maxRetries: 3,
		retryDelay: retryDelay,
	}
}

type generateRequest struct {
	Contents         []content         `json:"contents"`
	GenerationConfig *generationConfig `json:"generationConfig,omitempty"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	Temperature float64 `json:"temperature"`
}

type generateResponse struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Rate limit: wait for next available slot
		select {
		case <-c.limiter:
		case <-ctx.Done():
			return "", ctx.Err()
		}

		result, err := c.generate(ctx, prompt)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !isRetryable(err) {
			return "", err
		}

		if attempt < c.maxRetries {
			backoff := c.exponentialBackoff(attempt)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

// make a single API call without retry.
func (c *Client) generate(ctx context.Context, prompt string) (string, error) {
	reqBody := generateRequest{
		Contents: []content{
			{Parts: []part{{Text: prompt}}},
		},
		GenerationConfig: &generationConfig{Temperature: 0},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("encoding request: %w", err)
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.model, c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var parsed generateResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", fmt.Errorf("gemini error (%d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return "", fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, raw)
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content")
	}

	return parsed.Candidates[0].Content.Parts[0].Text, nil
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "network error") ||
		contains(errStr, "429") ||
		contains(errStr, "500") ||
		contains(errStr, "502") ||
		contains(errStr, "503") ||
		contains(errStr, "504")
}

func (c *Client) exponentialBackoff(attempt int) time.Duration {
	base := c.retryDelay
	for i := 0; i < attempt; i++ {
		base *= 2
	}
	return base
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
