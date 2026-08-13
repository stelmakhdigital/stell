package subagents_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/budaev/agent/internal/agent"
	"github.com/budaev/agent/internal/llm"
	"github.com/budaev/agent/internal/subagents"
)

type gateProvider struct {
	entered *sync.WaitGroup
	block   <-chan struct{}
}

func (p *gateProvider) Name() string { return "gate" }

func (p *gateProvider) Generate(ctx context.Context, _ llm.Request) (llm.Response, error) {
	p.entered.Done()
	select {
	case <-p.block:
		return llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "ok"}}, nil
	case <-ctx.Done():
		return llm.Response{}, ctx.Err()
	}
}

func TestTrySpawnLimit(t *testing.T) {
	var entered sync.WaitGroup
	entered.Add(subagents.MaxParallel)
	block := make(chan struct{})
	p := &gateProvider{entered: &entered, block: block}
	o := subagents.NewOrchestrator(func(subagents.Kind) *agent.Agent {
		return agent.New(agent.WithProvider(p))
	})

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < subagents.MaxParallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = o.Spawn(ctx, subagents.Request{Kind: subagents.KindExplore, Task: "x", Timeout: time.Minute})
		}()
	}
	done := make(chan struct{})
	go func() {
		entered.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for 7 slots")
	}
	_, err := o.TrySpawn(ctx, subagents.Request{Kind: subagents.KindPlan, Task: "8"})
	if !errors.Is(err, subagents.ErrLimit) {
		t.Fatalf("expected ErrLimit, got %v", err)
	}
	close(block)
	wg.Wait()
}

func TestSpawnExplore(t *testing.T) {
	provider := llm.NewFake(llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "ok"}})
	o := subagents.NewOrchestrator(func(subagents.Kind) *agent.Agent {
		return agent.New(agent.WithProvider(provider))
	})
	res, err := o.Explore(context.Background(), "task")
	if err != nil {
		t.Fatal(err)
	}
	if res.FinalText != "ok" {
		t.Fatalf("text=%q", res.FinalText)
	}
}
