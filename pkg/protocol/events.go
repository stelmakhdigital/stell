package protocol

import (
	"time"

	"github.com/budaev/stell/internal/eventbus"
)

// Event is the shared wire format for TUI / Web / IDE / public API.
type Event struct {
	Type      string         `json:"type"`
	SessionID string         `json:"session_id"`
	TurnID    string         `json:"turn_id,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	At        time.Time      `json:"at"`
}

// FromBus converts an in-process bus event.
func FromBus(e *eventbus.Event) Event {
	if e == nil {
		return Event{}
	}
	return Event{
		Type:      string(e.Type),
		SessionID: e.SessionID,
		TurnID:    e.TurnID,
		Data:      e.Data,
		At:        e.At,
	}
}

// CreateSessionRequest starts a new agent run.
type CreateSessionRequest struct {
	Message   string `json:"message"`
	Workspace string `json:"workspace,omitempty"`
}

// CreateSessionResponse is returned by POST /v1/sessions.
type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
}

// HITLRequest is POST /v1/sessions/{id}/hitl.
type HITLRequest struct {
	Decision string `json:"decision"`
}

// SessionStatus is GET /v1/sessions/{id}.
type SessionStatus struct {
	SessionID string `json:"session_id"`
	Running   bool   `json:"running"`
	FinalText string `json:"final_text,omitempty"`
	Error     string `json:"error,omitempty"`
}
