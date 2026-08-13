package commands

import (
	"context"
	"sync"

	"github.com/budaev/stell/internal/agent"
	"github.com/budaev/stell/tui/events"
)

// Controller starts/cancels agent sessions. TUI must not execute tools.
type Controller interface {
	Start(task string)
	Cancel()
	Running() bool
	ModelName() string
}

// AgentController is a thin wrapper over Brain.
type AgentController struct {
	agent *agent.Agent
	out   *events.Client

	mu      sync.Mutex
	cancel  context.CancelFunc
	running bool
}

// NewAgentController binds an agent and event client.
func NewAgentController(a *agent.Agent, out *events.Client) *AgentController {
	return &AgentController{agent: a, out: out}
}

// ModelName returns the LLM model.
func (c *AgentController) ModelName() string {
	if c == nil || c.agent == nil {
		return ""
	}
	return c.agent.ModelName()
}

// Running reports an in-flight session.
func (c *AgentController) Running() bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

// Start runs the agent loop in a goroutine.
func (c *AgentController) Start(task string) {
	if c == nil || c.agent == nil {
		return
	}
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.running = true
	c.mu.Unlock()

	go func() {
		res := c.agent.Run(ctx, task)
		c.mu.Lock()
		c.running = false
		c.cancel = nil
		c.mu.Unlock()
		if c.out != nil {
			c.out.Send(events.DoneMsg{Text: res.FinalText, Err: res.Err, Turns: res.Turns})
		}
	}()
}

// Cancel interrupts the current session.
func (c *AgentController) Cancel() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
	}
}
