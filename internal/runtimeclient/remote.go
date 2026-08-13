package runtimeclient

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// Healthy reports whether Hands /healthz is OK. HMAC is not required.
func (c *HTTPClient) Healthy(ctx context.Context) bool {
	if c == nil || c.baseURL == "" {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.baseURL, "/")+"/healthz", nil)
	if err != nil {
		return false
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode < 300
}

// Failover tries runtime replicas in order (or sticky by workspace). HMAC is per-replica.
type Failover struct {
	replicas []*HTTPClient
	sticky   bool
	rr       atomic.Uint32
}

// NewFailover wraps HMAC-enabled HTTP clients. urls must be non-empty.
func NewFailover(urls []string, hmacKey string, sticky bool) (*Failover, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no runtime urls")
	}
	f := &Failover{sticky: sticky}
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		f.replicas = append(f.replicas, NewHTTP(u).WithHMACKey(hmacKey))
	}
	if len(f.replicas) == 0 {
		return nil, fmt.Errorf("no runtime urls")
	}
	return f, nil
}

// Execute retries on the next healthy replica after transport or 5xx errors.
func (f *Failover) Execute(ctx context.Context, req ExecuteRequest) (ExecuteResponse, error) {
	n := len(f.replicas)
	start := int(f.rr.Add(1)-1) % n
	if f.sticky && req.Workspace != "" {
		h := fnv.New32a()
		_, _ = h.Write([]byte(req.Workspace))
		start = int(h.Sum32()) % n
	}
	var lastErr error
	for i := 0; i < n; i++ {
		idx := (start + i) % n
		c := f.replicas[idx]
		hctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		ok := c.Healthy(hctx)
		cancel()
		if !ok && n > 1 {
			lastErr = fmt.Errorf("runtime %s unhealthy", c.baseURL)
			continue
		}
		resp, err := c.Execute(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("all runtime replicas failed")
	}
	return ExecuteResponse{}, lastErr
}
