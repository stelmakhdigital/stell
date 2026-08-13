package hooks_test

import (
	"context"
	"testing"
	"time"

	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/internal/hooks"
	"github.com/budaev/agent/internal/tools"
)

func TestHITLDenyBlocks(t *testing.T) {
	store := tools.NewManifestStore(map[string]tools.Manifest{
		"bash": {Name: "bash", RequiresHITL: true},
	}, false)
	h := hooks.NewHITLHook(store, hooks.StaticApprover{Decision: hooks.DecisionDeny}, time.Second, eventbus.New())
	res, err := h.Handle(context.Background(), &eventbus.Event{
		Type: eventbus.EventToolCall,
		Data: map[string]any{"tool": "bash", "args": map[string]any{"command": "echo"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !res.Block {
		t.Fatal("expected block")
	}
}

func TestHITLTimeoutDenies(t *testing.T) {
	store := tools.NewManifestStore(map[string]tools.Manifest{
		"bash": {Name: "bash", RequiresHITL: true},
	}, false)
	h := hooks.NewHITLHook(store, delayApprover{d: 200 * time.Millisecond}, 20*time.Millisecond, nil)
	res, err := h.Handle(context.Background(), &eventbus.Event{
		Type: eventbus.EventToolCall,
		Data: map[string]any{"tool": "bash"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !res.Block {
		t.Fatal("expected timeout deny")
	}
}

type delayApprover struct{ d time.Duration }

func (a delayApprover) Decide(ctx context.Context, _ string, _ map[string]any) (string, error) {
	select {
	case <-time.After(a.d):
		return hooks.DecisionApprove, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
