package eventbus

import "time"

// EventType identifies an event kind.
type EventType string

const (
	EventSessionStart   EventType = "session_start"
	EventSessionEnd     EventType = "session_end"
	EventTurnStart      EventType = "turn_start"
	EventTurnEnd        EventType = "turn_end"
	EventToolCall       EventType = "tool_call"
	EventToolResult     EventType = "tool_result"
	EventModelRequest   EventType = "model_request"
	EventModelResponse  EventType = "model_response"
	EventAgentError     EventType = "agent_error"
	EventGuardrailBlock EventType = "guardrail_block"
	EventHITLRequest    EventType = "hitl_request"
	EventHITLDecision   EventType = "hitl_decision"
	EventContextCompact EventType = "context_compact"
)

// Event is a bus message.
type Event struct {
	Type      EventType
	SessionID string
	TurnID    string
	Data      map[string]any
	At        time.Time
}

// EventResult is returned by synchronous handlers.
type EventResult struct {
	Modified bool
	Data     map[string]any
	Block    bool
	Error    error
}

// Handler processes an event synchronously.
type Handler func(event *Event) (*EventResult, error)
