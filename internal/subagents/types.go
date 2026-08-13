package subagents

import (
	"errors"
	"time"
)

// Kind identifies a specialized child agent.
type Kind string

const (
	KindExplore Kind = "explore"
	KindPlan    Kind = "plan"
	KindBash    Kind = "bash"
	KindGeneral Kind = "general"
)

// MaxParallel is the hard cap on concurrent child agents.
const MaxParallel = 7

// ErrLimit is returned by TrySpawn when 7 slots are already taken.
// Spawn waits instead of returning this error.
var ErrLimit = errors.New("subagent limit reached (max 7 parallel)")

// Request describes a child-agent task.
type Request struct {
	Kind    Kind
	Task    string
	Timeout time.Duration
}

// Result is the child-agent outcome.
type Result struct {
	Kind      Kind
	FinalText string
	Turns     int
	Err       error
}
