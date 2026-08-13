package hooks

import (
	"context"

	"github.com/budaev/stell/internal/eventbus"
)

// Hook is a prioritized event middleware.
type Hook interface {
	Name() string
	Priority() int // lower runs first
	Handle(ctx context.Context, event *eventbus.Event) (*eventbus.EventResult, error)
}
