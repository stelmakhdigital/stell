package domain

import "time"

// AgentID identifies an agent instance.
type AgentID string

// SessionID identifies a session.
type SessionID string

// TurnID identifies a turn within a session.
type TurnID string

// ToolCallID identifies a tool invocation.
type ToolCallID string

// ToolName is a registered tool name.
type ToolName string

// Agent is the root aggregate.
type Agent struct {
	ID        AgentID
	Name      string
	Session   *Session
	CreatedAt time.Time
}

// Session holds conversation state for one task.
type Session struct {
	ID        SessionID
	AgentID   AgentID
	Task      string
	Turns     []Turn
	Memory    Memory
	CreatedAt time.Time
	EndedAt   *time.Time
}

// Turn is one model interaction cycle.
type Turn struct {
	ID          TurnID
	SessionID   SessionID
	Depth       int
	ToolCalls   []ToolCall
	ModelOutput string
	CreatedAt   time.Time
}

// ToolCall records a single tool invocation.
type ToolCall struct {
	ID        ToolCallID
	TurnID    TurnID
	Tool      ToolName
	Args      map[string]any
	Result    *ToolResult
	Error     string
	CreatedAt time.Time
}

// ToolResult is the outcome of a tool execution.
type ToolResult struct {
	Content  string
	Metadata map[string]any
}

// Memory holds session and persistent memory snapshots.
type Memory struct {
	SessionMemory    string
	PersistentMemory string
}

// NewSession creates a session for a task.
func NewSession(id SessionID, agentID AgentID, task string) *Session {
	return &Session{
		ID:        id,
		AgentID:   agentID,
		Task:      task,
		Turns:     make([]Turn, 0),
		CreatedAt: time.Now().UTC(),
	}
}

// AddTurn appends a turn to the session.
func (s *Session) AddTurn(t Turn) {
	s.Turns = append(s.Turns, t)
}

// End marks the session as finished.
func (s *Session) End() {
	now := time.Now().UTC()
	s.EndedAt = &now
}
