package hooks

import (
	"context"

	"github.com/budaev/stell/internal/eventbus"
)

// Chain runs registered hooks in priority order.
type Chain struct {
	reg *Registry
}

// NewChain wraps a registry.
func NewChain(reg *Registry) *Chain {
	return &Chain{reg: reg}
}

// Handle executes hooks until one Blocks or errors.
func (c *Chain) Handle(ctx context.Context, event *eventbus.Event) (*eventbus.EventResult, error) {
	var last *eventbus.EventResult
	for _, h := range c.reg.Hooks() {
		res, err := h.Handle(ctx, event)
		if err != nil {
			return res, err
		}
		if res == nil {
			continue
		}
		if res.Modified && res.Data != nil {
			event.Data = res.Data
		}
		last = res
		if res.Block {
			return res, nil
		}
	}
	return last, nil
}
