package agent_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/budaev/stell/internal/agent"
	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/internal/hooks"
	"github.com/budaev/stell/internal/llm"
	"github.com/budaev/stell/internal/tools"
)

func TestCompactPreservesDecisions(t *testing.T) {
	msgs := []llm.Message{
		{Role: llm.RoleSystem, Content: "sys"},
		{Role: llm.RoleUser, Content: "noise " + strings.Repeat("x", 200)},
		{Role: llm.RoleAssistant, Content: "Decision: use Postgres for sessions."},
		{Role: llm.RoleUser, Content: "Constraint: HMAC must stay enabled."},
		{Role: llm.RoleAssistant, Content: "touched internal/sessionstore/file.go"},
		{Role: llm.RoleUser, Content: "more noise " + strings.Repeat("y", 200)},
		{Role: llm.RoleAssistant, Content: "ok"},
		{Role: llm.RoleUser, Content: "latest"},
	}
	out := agent.CompactMessages(msgs, 2)
	blob := ""
	for _, m := range out {
		blob += m.Content + "\n"
	}
	if !strings.Contains(blob, "Postgres") {
		t.Fatalf("lost decision: %s", blob)
	}
	if !strings.Contains(blob, "HMAC") {
		t.Fatalf("lost constraint: %s", blob)
	}
	if !strings.Contains(blob, "[compacted context]") {
		t.Fatal("missing compact marker")
	}
}

func TestCompressHasTruncationMarker(t *testing.T) {
	big := strings.Repeat("line\n", 2000) + "path/to/foo.go\n" + strings.Repeat("z", 2000)
	out := agent.CompressToolResult(big, 400)
	if !strings.Contains(out, "...[truncated]...") {
		t.Fatalf("missing marker: %s", out[:min(len(out), 80)])
	}
	if !strings.Contains(out, "foo.go") {
		t.Fatal("expected preserved path")
	}
}

func TestLongSessionCompacts(t *testing.T) {
	var fires int
	bus := eventbus.New()
	bus.Subscribe(eventbus.EventContextCompact, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		fires++
		return nil, nil
	})
	payload := strings.Repeat("word ", 80)
	var responses []llm.Response
	for i := 0; i < 6; i++ {
		responses = append(responses, llm.Response{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{{
					ID:       "c",
					Type:     "function",
					Function: llm.FunctionCall{Name: "echo", Arguments: `{}`},
				}},
			},
		})
	}
	responses = append(responses, llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "done"}})
	provider := llm.NewFake(responses...)
	reg := tools.NewRegistry()
	reg.Register(&echoTool{out: payload})
	a := agent.New(
		agent.WithProvider(provider),
		agent.WithRegistry(reg),
		agent.WithEventBus(bus),
		agent.WithCompactor(hooks.NewCompactor(40, 0.8)),
		agent.WithCompressToolBytes(64),
		agent.WithMaxLoopDepth(20),
	)
	res := a.Run(context.Background(), "go")
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if fires < 1 {
		t.Fatalf("expected compact >=1, got %d", fires)
	}
}

type echoTool struct{ out string }

func (t *echoTool) Name() string            { return "echo" }
func (t *echoTool) Description() string     { return "echo" }
func (t *echoTool) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }
func (t *echoTool) Execute(context.Context, map[string]any) (string, error) {
	return t.out, nil
}
