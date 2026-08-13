package llm

import (
	"context"
	"fmt"
	"time"
)

// Bulkhead limits concurrent Generate calls (LLM isolation).
type Bulkhead struct {
	Inner   Provider
	slots   chan struct{}
	Timeout time.Duration
}

// NewBulkhead allows at most n concurrent inner calls.
func NewBulkhead(inner Provider, n int, timeout time.Duration) *Bulkhead {
	if n <= 0 {
		n = 8
	}
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &Bulkhead{Inner: inner, slots: make(chan struct{}, n), Timeout: timeout}
}

func (b *Bulkhead) Name() string {
	if b.Inner == nil {
		return "bulkhead"
	}
	return b.Inner.Name()
}

func (b *Bulkhead) Generate(ctx context.Context, req Request) (Response, error) {
	ctx, cancel := context.WithTimeout(ctx, b.Timeout)
	defer cancel()
	select {
	case b.slots <- struct{}{}:
		defer func() { <-b.slots }()
	case <-ctx.Done():
		return Response{}, fmt.Errorf("llm bulkhead: %w", ctx.Err())
	}
	return b.Inner.Generate(ctx, req)
}
