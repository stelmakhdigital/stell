package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"sync/atomic"
)

// CacheAware wraps a Provider and tracks stable-prefix prompt-cache hits.
// If the inner provider ignores cache headers this is still a local no-op metric layer.
type CacheAware struct {
	Inner  Provider
	Hits   atomic.Int64
	Misses atomic.Int64

	mu      sync.Mutex
	seen    map[string]struct{}
	enabled bool
}

// NewCacheAware wraps inner. enabled=false still hashes but never counts hits (no-op).
func NewCacheAware(inner Provider, enabled bool) *CacheAware {
	return &CacheAware{
		Inner:   inner,
		enabled: enabled,
		seen:    make(map[string]struct{}),
	}
}

func (c *CacheAware) Name() string {
	if c.Inner == nil {
		return "cache"
	}
	return c.Inner.Name()
}

// Generate records cache hit/miss for the system-prefix key, then delegates.
func (c *CacheAware) Generate(ctx context.Context, req Request) (Response, error) {
	key := PrefixKey(req.Messages)
	req.CacheKey = key
	if c.enabled {
		c.mu.Lock()
		_, hit := c.seen[key]
		if hit {
			c.Hits.Add(1)
		} else {
			c.seen[key] = struct{}{}
			c.Misses.Add(1)
		}
		c.mu.Unlock()
	}
	if c.Inner == nil {
		return Response{}, context.Canceled
	}
	return c.Inner.Generate(ctx, req)
}

// PrefixKey hashes concatenated system-role contents (stable project/system layers).
func PrefixKey(messages []Message) string {
	h := sha256.New()
	for _, m := range messages {
		if m.Role != RoleSystem {
			continue
		}
		_, _ = h.Write([]byte(m.Content))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// HitRate is hits/(hits+misses).
func (c *CacheAware) HitRate() float64 {
	h := float64(c.Hits.Load())
	m := float64(c.Misses.Load())
	if h+m == 0 {
		return 0
	}
	return h / (h + m)
}
