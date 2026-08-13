package agent

import (
	"github.com/budaev/agent/internal/domain"
	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/internal/guardrails"
	"github.com/budaev/agent/internal/hooks"
	"github.com/budaev/agent/internal/llm"
	"github.com/budaev/agent/internal/skills"
	"github.com/budaev/agent/internal/tools"
)

const defaultMaxDepth = 50

// Agent is the Brain orchestrator.
type Agent struct {
	id                domain.AgentID
	name              string
	bus               *eventbus.Bus
	registry          *tools.Registry
	provider          llm.Provider
	sessions          domain.SessionRepository
	maxLoopDepth      int
	model             string
	temperature       float64
	maxTokens         int
	tokenBudget       int
	contextBuilder    *ContextBuilder
	guardrails        *guardrails.Framework
	manifests         *tools.ManifestStore
	production        bool
	compactor         *hooks.Compactor
	compressToolBytes int
}

// Option configures an Agent.
type Option func(*Agent)

// WithEventBus sets the event bus.
func WithEventBus(bus *eventbus.Bus) Option {
	return func(a *Agent) { a.bus = bus }
}

// WithRegistry sets the tool registry.
func WithRegistry(r *tools.Registry) Option {
	return func(a *Agent) { a.registry = r }
}

// WithProvider sets the LLM provider.
func WithProvider(p llm.Provider) Option {
	return func(a *Agent) { a.provider = p }
}

// WithSessionRepo sets session persistence.
func WithSessionRepo(r domain.SessionRepository) Option {
	return func(a *Agent) { a.sessions = r }
}

// WithMaxLoopDepth sets max agent loop iterations.
func WithMaxLoopDepth(depth int) Option {
	return func(a *Agent) {
		if depth > 0 {
			a.maxLoopDepth = depth
		}
	}
}

// WithModel sets the default model name.
func WithModel(model string) Option {
	return func(a *Agent) { a.model = model }
}

// WithTemperature sets sampling temperature.
func WithTemperature(t float64) Option {
	return func(a *Agent) { a.temperature = t }
}

// WithMaxTokens sets max completion tokens.
func WithMaxTokens(n int) Option {
	return func(a *Agent) { a.maxTokens = n }
}

// WithTokenBudget sets a session-wide token circuit breaker (0 = unlimited).
func WithTokenBudget(n int) Option {
	return func(a *Agent) { a.tokenBudget = n }
}

// WithName sets the agent display name.
func WithName(name string) Option {
	return func(a *Agent) { a.name = name }
}

// WithSkills attaches a skill manager to the prompt builder.
func WithSkills(m *skills.Manager) Option {
	return func(a *Agent) { a.contextBuilder.Skills = m }
}

// WithGuardrails sets input/output policies.
func WithGuardrails(f *guardrails.Framework) Option {
	return func(a *Agent) { a.guardrails = f }
}

// WithManifests sets tool policy manifests.
func WithManifests(s *tools.ManifestStore) Option {
	return func(a *Agent) { a.manifests = s }
}

// WithProduction enables fail-closed tool policy.
func WithProduction(on bool) Option {
	return func(a *Agent) { a.production = on }
}

// WithCompactor sets context compaction policy.
func WithCompactor(c *hooks.Compactor) Option {
	return func(a *Agent) { a.compactor = c }
}

// WithCompressToolBytes sets the tool-result compression threshold (0 = off).
func WithCompressToolBytes(n int) Option {
	return func(a *Agent) { a.compressToolBytes = n }
}

// Provider returns the LLM provider.
func (a *Agent) Provider() llm.Provider { return a.provider }

// Registry returns the tool registry.
func (a *Agent) Registry() *tools.Registry { return a.registry }

// Guardrails returns input/output policies.
func (a *Agent) Guardrails() *guardrails.Framework { return a.guardrails }

// Manifests returns tool manifests.
func (a *Agent) Manifests() *tools.ManifestStore { return a.manifests }

// Sessions returns the session repository.
func (a *Agent) Sessions() domain.SessionRepository { return a.sessions }

// Compactor returns the compaction policy.
func (a *Agent) Compactor() *hooks.Compactor { return a.compactor }

// New creates an Agent with options.
func New(opts ...Option) *Agent {
	a := &Agent{
		id:             domain.AgentID("agent-1"),
		name:           "coding-agent",
		bus:            eventbus.New(),
		registry:       tools.NewRegistry(),
		sessions:       domain.NewMemorySessionRepository(),
		maxLoopDepth:   defaultMaxDepth,
		model:          "llama3",
		temperature:    0.2,
		maxTokens:      4096,
		contextBuilder: NewContextBuilder(),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Bus returns the agent's event bus.
func (a *Agent) Bus() *eventbus.Bus {
	return a.bus
}

// ModelName returns the configured LLM model.
func (a *Agent) ModelName() string {
	return a.model
}
