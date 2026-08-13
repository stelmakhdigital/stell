package agent_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/budaev/stell/internal/agent"
	"github.com/budaev/stell/internal/domain"
	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/internal/llm"
	"github.com/budaev/stell/internal/tools"
	"github.com/budaev/stell/internal/tools/builtin"
)

func TestRunFinalAnswer(t *testing.T) {
	provider := llm.NewFake(llm.Response{
		Message:      llm.Message{Role: llm.RoleAssistant, Content: "hello"},
		FinishReason: "stop",
	})
	bus := eventbus.New()
	var events []eventbus.EventType
	var mu sync.Mutex
	for _, et := range []eventbus.EventType{
		eventbus.EventSessionStart,
		eventbus.EventTurnStart,
		eventbus.EventModelRequest,
		eventbus.EventModelResponse,
		eventbus.EventTurnEnd,
		eventbus.EventSessionEnd,
	} {
		et := et
		bus.Subscribe(et, func(e *eventbus.Event) (*eventbus.EventResult, error) {
			mu.Lock()
			events = append(events, e.Type)
			mu.Unlock()
			return nil, nil
		})
	}

	reg := tools.NewRegistry()
	builtin.RegisterStubs(reg)

	a := agent.New(
		agent.WithProvider(provider),
		agent.WithEventBus(bus),
		agent.WithRegistry(reg),
		agent.WithModel("fake"),
	)

	res := a.Run(context.Background(), "say hi")
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if res.FinalText != "hello" {
		t.Fatalf("final=%q", res.FinalText)
	}
	if res.Turns != 1 {
		t.Fatalf("turns=%d", res.Turns)
	}
	if provider.Calls() != 1 {
		t.Fatalf("provider calls=%d", provider.Calls())
	}
}

func TestRunToolThenFinal(t *testing.T) {
	provider := llm.NewFake(
		llm.Response{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{{
					ID:   "call-1",
					Type: "function",
					Function: llm.FunctionCall{
						Name:      "read_file",
						Arguments: `{"path":"README.md"}`,
					},
				}},
			},
			FinishReason: "tool_calls",
		},
		llm.Response{
			Message:      llm.Message{Role: llm.RoleAssistant, Content: "file missing (stub)"},
			FinishReason: "stop",
		},
	)

	reg := tools.NewRegistry()
	builtin.RegisterStubs(reg)

	a := agent.New(
		agent.WithProvider(provider),
		agent.WithRegistry(reg),
	)

	res := a.Run(context.Background(), "read README")
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if res.FinalText != "file missing (stub)" {
		t.Fatalf("final=%q", res.FinalText)
	}
	if res.Turns != 2 {
		t.Fatalf("turns=%d", res.Turns)
	}
}

func TestToolResultEventHasPreview(t *testing.T) {
	provider := llm.NewFake(
		llm.Response{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{{
					ID:   "call-1",
					Type: "function",
					Function: llm.FunctionCall{
						Name:      "read_file",
						Arguments: `{"path":"README.md"}`,
					},
				}},
			},
			FinishReason: "tool_calls",
		},
		llm.Response{
			Message:      llm.Message{Role: llm.RoleAssistant, Content: "done"},
			FinishReason: "stop",
		},
	)
	var preview, tool string
	bus := eventbus.New()
	bus.Subscribe(eventbus.EventToolResult, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		tool, _ = e.Data["tool"].(string)
		preview, _ = e.Data["preview"].(string)
		return nil, nil
	})
	reg := tools.NewRegistry()
	builtin.RegisterStubs(reg)
	a := agent.New(
		agent.WithProvider(provider),
		agent.WithRegistry(reg),
		agent.WithEventBus(bus),
	)
	res := a.Run(context.Background(), "read README")
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if tool != "read_file" || preview == "" {
		t.Fatalf("tool=%q preview=%q", tool, preview)
	}
}

func TestMaxDepth(t *testing.T) {
	// Always request a tool call → hit max depth.
	mk := func() llm.Response {
		return llm.Response{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{{
					ID:   "c",
					Type: "function",
					Function: llm.FunctionCall{
						Name:      "grep",
						Arguments: `{"pattern":"x"}`,
					},
				}},
			},
			FinishReason: "tool_calls",
		}
	}
	responses := make([]llm.Response, 0, 3)
	for i := 0; i < 3; i++ {
		responses = append(responses, mk())
	}
	provider := llm.NewFake(responses...)

	reg := tools.NewRegistry()
	builtin.RegisterStubs(reg)

	sawError := false
	bus := eventbus.New()
	bus.Subscribe(eventbus.EventAgentError, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		sawError = true
		return nil, nil
	})

	a := agent.New(
		agent.WithProvider(provider),
		agent.WithRegistry(reg),
		agent.WithEventBus(bus),
		agent.WithMaxLoopDepth(3),
	)

	res := a.Run(context.Background(), "loop")
	if !errors.Is(res.Err, domain.ErrMaxDepthExceeded) {
		t.Fatalf("expected max depth, got %v", res.Err)
	}
	if !sawError {
		t.Fatal("expected agent_error event")
	}
}

func TestToolErrorDoesNotAbort(t *testing.T) {
	provider := llm.NewFake(
		llm.Response{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{{
					ID:       "c1",
					Type:     "function",
					Function: llm.FunctionCall{Name: "bash", Arguments: `{"command":"echo"}`},
				}},
			},
		},
		llm.Response{
			Message: llm.Message{Role: llm.RoleAssistant, Content: "recovered"},
		},
	)
	reg := tools.NewRegistry()
	builtin.RegisterStubs(reg)

	a := agent.New(agent.WithProvider(provider), agent.WithRegistry(reg))
	res := a.Run(context.Background(), "run bash")
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if res.FinalText != "recovered" {
		t.Fatalf("final=%q", res.FinalText)
	}
}
