package hooks

import (
	"context"
	"fmt"
	"strings"

	"github.com/budaev/agent/internal/eventbus"
)

var dangerousBash = []string{
	"rm -rf /",
	"mkfs",
	"dd if=",
	":(){ :|:& };:",
	"shutdown",
	"reboot",
	"curl | sh",
	"wget | sh",
}

// GuardrailHook can Block dangerous tool_call events.
type GuardrailHook struct{}

func (h *GuardrailHook) Name() string  { return "guardrail" }
func (h *GuardrailHook) Priority() int { return PriorityGuardrail }

func (h *GuardrailHook) Handle(_ context.Context, event *eventbus.Event) (*eventbus.EventResult, error) {
	if event == nil || event.Type != eventbus.EventToolCall {
		return nil, nil
	}
	tool, _ := event.Data["tool"].(string)
	if tool != "bash" {
		return nil, nil
	}
	args, _ := event.Data["args"].(map[string]any)
	cmd, _ := args["command"].(string)
	low := strings.ToLower(cmd)
	for _, d := range dangerousBash {
		if strings.Contains(low, d) {
			return &eventbus.EventResult{
				Block: true,
				Error: fmt.Errorf("guardrail blocked dangerous bash"),
			}, nil
		}
	}
	return nil, nil
}
