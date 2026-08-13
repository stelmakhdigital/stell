package subagents

import "context"

// Explore is a read-only research child agent.
func (o *Orchestrator) Explore(ctx context.Context, task string) (Result, error) {
	return o.Spawn(ctx, Request{Kind: KindExplore, Task: task})
}

// Plan is a read-only planning child agent.
func (o *Orchestrator) Plan(ctx context.Context, task string) (Result, error) {
	return o.Spawn(ctx, Request{Kind: KindPlan, Task: task})
}
