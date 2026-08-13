package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompat is an OpenAI-compatible chat completions client.
type OpenAICompat struct {
	name    string
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewOpenAICompat creates a provider hitting baseURL (e.g. http://localhost:11434/v1).
func NewOpenAICompat(name, baseURL, apiKey string) *OpenAICompat {
	return &OpenAICompat{
		name:    name,
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				// Local inference servers often drop keep-alive sockets → broken pipe.
				DisableKeepAlives: true,
			},
		},
	}
}

func (p *OpenAICompat) Name() string { return p.name }

type chatRequest struct {
	Model          string     `json:"model"`
	Messages       []Message  `json:"messages"`
	Tools          []ToolSpec `json:"tools,omitempty"`
	Temperature    float64    `json:"temperature,omitempty"`
	MaxTokens      int        `json:"max_tokens,omitempty"`
	PromptCacheKey string     `json:"prompt_cache_key,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// Generate calls /chat/completions.
func (p *OpenAICompat) Generate(ctx context.Context, req Request) (Response, error) {
	body := chatRequest{
		Model:          req.Model,
		Messages:       req.Messages,
		Tools:          req.Tools,
		Temperature:    req.Temperature,
		MaxTokens:      req.MaxTokens,
		PromptCacheKey: req.CacheKey,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return Response{}, err
	}

	url := p.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.ContentLength = int64(len(payload))
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(payload)), nil
	}
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	if req.CacheKey != "" {
		httpReq.Header.Set("X-Prompt-Cache-Key", req.CacheKey)
	}

	start := time.Now()
	resp, err := p.do(httpReq)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}
	if resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("llm %s: status %d: %s", p.name, resp.StatusCode, truncate(string(raw), 512))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Response{}, fmt.Errorf("decode llm response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return Response{}, fmt.Errorf("llm %s: empty choices", p.name)
	}

	choice := parsed.Choices[0]
	return Response{
		Message:      choice.Message,
		TokensUsed:   parsed.Usage.TotalTokens,
		LatencyMs:    time.Since(start).Milliseconds(),
		FinishReason: choice.FinishReason,
	}, nil
}

func (p *OpenAICompat) do(req *http.Request) (*http.Response, error) {
	resp, err := p.client.Do(req)
	if err != nil && isRetryableNetErr(err) && req.GetBody != nil {
		body, rewindErr := req.GetBody()
		if rewindErr == nil {
			req.Body = body
			resp, err = p.client.Do(req)
		}
	}
	return resp, err
}

func isRetryableNetErr(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "connection reset") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "unexpected eof")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
