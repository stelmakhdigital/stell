package hooks

import (
	"context"
	"sort"
	"sync"

	"github.com/budaev/stell/internal/eventbus"
)

// Registry stores hooks and can attach them to an Event Bus.
type Registry struct {
	mu    sync.RWMutex
	hooks []Hook
}

// NewRegistry creates an empty hook registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a hook.
func (r *Registry) Register(h Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks = append(r.hooks, h)
	sort.Slice(r.hooks, func(i, j int) bool {
		return r.hooks[i].Priority() < r.hooks[j].Priority()
	})
}

// Hooks returns a snapshot in priority order.
func (r *Registry) Hooks() []Hook {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Hook, len(r.hooks))
	copy(out, r.hooks)
	return out
}

// Attach subscribes a chain of all hooks to the given event types.
func (r *Registry) Attach(bus *eventbus.Bus, types ...eventbus.EventType) {
	chain := NewChain(r)
	for _, t := range types {
		t := t
		bus.Subscribe(t, func(event *eventbus.Event) (*eventbus.EventResult, error) {
			return chain.Handle(context.Background(), event)
		})
	}
}

// Handle runs the chain for a single event (tests / direct use).
func (r *Registry) Handle(ctx context.Context, event *eventbus.Event) (*eventbus.EventResult, error) {
	return NewChain(r).Handle(ctx, event)
}
