package hooks_test

import (
	"context"
	"testing"

	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/internal/hooks"
)

type stubHook struct {
	name     string
	priority int
	fn       func(*eventbus.Event) *eventbus.EventResult
	calls    *int
}

func (s stubHook) Name() string  { return s.name }
func (s stubHook) Priority() int { return s.priority }
func (s stubHook) Handle(_ context.Context, e *eventbus.Event) (*eventbus.EventResult, error) {
	if s.calls != nil {
		*s.calls++
	}
	if s.fn == nil {
		return nil, nil
	}
	return s.fn(e), nil
}

func TestChainPriorityAndBlock(t *testing.T) {
	reg := hooks.NewRegistry()
	var order []string
	reg.Register(stubHook{name: "low", priority: 50, fn: func(e *eventbus.Event) *eventbus.EventResult {
		order = append(order, "low")
		return nil
	}})
	reg.Register(stubHook{name: "block", priority: 10, fn: func(e *eventbus.Event) *eventbus.EventResult {
		order = append(order, "block")
		return &eventbus.EventResult{Block: true}
	}})
	res, err := reg.Handle(context.Background(), &eventbus.Event{Type: eventbus.EventToolCall, Data: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !res.Block {
		t.Fatal("expected block")
	}
	if len(order) != 1 || order[0] != "block" {
		t.Fatalf("order=%v", order)
	}
}

func TestChainModify(t *testing.T) {
	reg := hooks.NewRegistry()
	reg.Register(stubHook{name: "mod", priority: 1, fn: func(e *eventbus.Event) *eventbus.EventResult {
		return &eventbus.EventResult{Modified: true, Data: map[string]any{"tool": "grep", "args": e.Data["args"]}}
	}})
	ev := &eventbus.Event{Type: eventbus.EventToolCall, Data: map[string]any{"tool": "bash"}}
	res, err := reg.Handle(context.Background(), ev)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !res.Modified {
		t.Fatal("expected modify")
	}
	if ev.Data["tool"] != "grep" {
		t.Fatalf("tool=%v", ev.Data["tool"])
	}
}
