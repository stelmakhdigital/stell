package subagents

import (
	"context"
	"fmt"
	"time"

	"github.com/budaev/agent/internal/agent"
	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/internal/llm"
	"github.com/budaev/agent/internal/tools"
)

// Factory builds an isolated child agent for a kind.
type Factory func(kind Kind) *agent.Agent

// Orchestrator limits parallel child agents to MaxParallel.
type Orchestrator struct {
	sem     chan struct{}
	factory Factory
}

// NewOrchestrator creates an orchestrator. factory may be nil until SetFactory.
func NewOrchestrator(factory Factory) *Orchestrator {
	return &Orchestrator{
		sem:     make(chan struct{}, MaxParallel),
		factory: factory,
	}
}

// SetFactory replaces the child-agent factory.
func (o *Orchestrator) SetFactory(f Factory) {
	o.factory = f
}

// Spawn waits for a free slot (8th caller blocks until a slot frees or ctx ends).
func (o *Orchestrator) Spawn(ctx context.Context, req Request) (Result, error) {
	select {
	case o.sem <- struct{}{}:
		defer func() { <-o.sem }()
	case <-ctx.Done():
		return Result{Kind: req.Kind, Err: ctx.Err()}, ctx.Err()
	}
	return o.run(ctx, req)
}

// TrySpawn returns ErrLimit immediately when all 7 slots are busy.
func (o *Orchestrator) TrySpawn(ctx context.Context, req Request) (Result, error) {
	select {
	case o.sem <- struct{}{}:
		defer func() { <-o.sem }()
	default:
		return Result{Kind: req.Kind, Err: ErrLimit}, ErrLimit
	}
	return o.run(ctx, req)
}

func (o *Orchestrator) run(ctx context.Context, req Request) (Result, error) {
	if o.factory == nil {
		return Result{Kind: req.Kind, Err: fmt.Errorf("subagent factory is not configured")}, fmt.Errorf("subagent factory is not configured")
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	child := o.factory(req.Kind)
	if child == nil {
		err := fmt.Errorf("subagent factory returned nil for %s", req.Kind)
		return Result{Kind: req.Kind, Err: err}, err
	}
	res := child.Run(ctx, req.Task)
	out := Result{Kind: req.Kind, FinalText: res.FinalText, Turns: res.Turns, Err: res.Err}
	return out, res.Err
}

// DefaultFactory builds a child with an isolated bus and a filtered registry.
func DefaultFactory(provider llm.Provider, parent *tools.Registry, model string) Factory {
	return func(kind Kind) *agent.Agent {
		reg := tools.NewRegistry()
		if parent != nil {
			for _, name := range toolsForKind(kind) {
				if t, ok := parent.Lookup(name); ok {
					reg.Register(t)
				}
			}
		}
		return agent.New(
			agent.WithProvider(provider),
			agent.WithRegistry(reg),
			agent.WithEventBus(eventbus.New()),
			agent.WithModel(model),
			agent.WithName(string(kind)+"-subagent"),
			agent.WithMaxLoopDepth(20),
		)
	}
}

func toolsForKind(kind Kind) []string {
	switch kind {
	case KindExplore, KindPlan:
		return []string{"read_file", "grep", "glob"}
	case KindBash:
		return []string{"bash", "read_file", "grep", "glob"}
	default:
		return []string{"read_file", "write_file", "bash", "grep", "glob"}
	}
}
