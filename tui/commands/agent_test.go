package commands_test

import (
	"context"
	"testing"
	"time"

	"github.com/budaev/agent/internal/agent"
	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/internal/llm"
	"github.com/budaev/agent/internal/tools"
	"github.com/budaev/agent/internal/tools/builtin"
	"github.com/budaev/agent/tui/commands"
	"github.com/budaev/agent/tui/events"
)

func TestControllerStartAndDone(t *testing.T) {
	provider := llm.NewFake(llm.Response{
		Message:      llm.Message{Role: llm.RoleAssistant, Content: "ok from fake"},
		FinishReason: "stop",
	})
	bus := eventbus.New()
	reg := tools.NewRegistry()
	builtin.RegisterStubs(reg)
	a := agent.New(
		agent.WithProvider(provider),
		agent.WithEventBus(bus),
		agent.WithRegistry(reg),
	)
	client := events.NewClient(bus)
	ctrl := commands.NewAgentController(a, client)
	ctrl.Start("hi")

	deadline := time.After(2 * time.Second)
	for {
		select {
		case msg := <-client.Chan():
			if done, ok := msg.(events.DoneMsg); ok {
				if done.Text != "ok from fake" {
					t.Fatalf("text=%q err=%v", done.Text, done.Err)
				}
				if ctrl.Running() {
					t.Fatal("still running after done")
				}
				return
			}
		case <-deadline:
			t.Fatal("timeout waiting for DoneMsg")
		}
	}
}

func TestControllerCancel(t *testing.T) {
	started := make(chan struct{})
	provider := &blockProvider{started: started}
	bus := eventbus.New()
	a := agent.New(agent.WithProvider(provider), agent.WithEventBus(bus))
	client := events.NewClient(bus)
	ctrl := commands.NewAgentController(a, client)
	ctrl.Start("loop")
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("provider never started")
	}
	ctrl.Cancel()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case msg := <-client.Chan():
			if done, ok := msg.(events.DoneMsg); ok {
				if done.Err == nil {
					t.Fatal("expected cancel error")
				}
				return
			}
		case <-deadline:
			t.Fatal("timeout waiting for cancel")
		}
	}
}

type blockProvider struct {
	started chan struct{}
}

func (p *blockProvider) Name() string { return "block" }

func (p *blockProvider) Generate(ctx context.Context, _ llm.Request) (llm.Response, error) {
	select {
	case <-p.started:
	default:
		close(p.started)
	}
	<-ctx.Done()
	return llm.Response{}, ctx.Err()
}

