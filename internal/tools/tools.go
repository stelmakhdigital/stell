package tools

import (
	"context"
	"encoding/json"
	"errors"
)

var (
	// ErrNotFound is returned when a tool is not registered.
	ErrNotFound = errors.New("tool not found")
	// ErrNotImplemented is returned by stub tools before Hands wiring.
	ErrNotImplemented = errors.New("tool not implemented")
)

// Tool is a callable capability exposed to the model.
type Tool interface {
	Name() string
	Description() string
	Schema() json.RawMessage
	Execute(ctx context.Context, args map[string]any) (string, error)
}

// Definition is the LLM-facing tool descriptor.
type Definition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}
