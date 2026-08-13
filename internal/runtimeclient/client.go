package runtimeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/budaev/agent/pkg/hmacauth"
	"github.com/google/uuid"
)

// ExecuteRequest is sent from Brain to Hands.
type ExecuteRequest struct {
	Tool      string         `json:"tool"`
	Args      map[string]any `json:"args"`
	Workspace string         `json:"workspace"`
	TimeoutMs int            `json:"timeout_ms,omitempty"`
}

// ExecuteResponse is returned by Hands.
type ExecuteResponse struct {
	Content   string         `json:"content"`
	Truncated bool           `json:"truncated,omitempty"`
	Error     string         `json:"error,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Client talks to the Hands runtime.
type Client interface {
	Execute(ctx context.Context, req ExecuteRequest) (ExecuteResponse, error)
}

// HTTPClient is a Brain-side client over JSON HTTP.
type HTTPClient struct {
	baseURL string
	client  *http.Client
	hmacKey []byte
}

// NewHTTP creates an HTTP runtime client.
func NewHTTP(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 130 * time.Second},
	}
}

// WithHMACKey enables HMAC signing of /v1/execute requests.
func (c *HTTPClient) WithHMACKey(key string) *HTTPClient {
	if key != "" {
		c.hmacKey = []byte(key)
	}
	return c
}

// Execute POSTs /v1/execute.
func (c *HTTPClient) Execute(ctx context.Context, req ExecuteRequest) (ExecuteResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return ExecuteResponse{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/execute", bytes.NewReader(payload))
	if err != nil {
		return ExecuteResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if len(c.hmacKey) > 0 {
		hmacauth.SignRequest(httpReq, c.hmacKey, payload, time.Now(), uuid.NewString())
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return ExecuteResponse{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ExecuteResponse{}, err
	}
	if resp.StatusCode >= 300 {
		return ExecuteResponse{}, fmt.Errorf("runtime status %d: %s", resp.StatusCode, string(body))
	}
	var out ExecuteResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return ExecuteResponse{}, err
	}
	return out, nil
}
