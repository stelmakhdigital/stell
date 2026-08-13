package guardrails

import (
	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/pkg/audit"
)

// AttachAudit writes redacted tool_call / tool_result lines to the DCL store.
func AttachAudit(bus *eventbus.Bus, store *audit.Store) {
	if bus == nil || store == nil {
		return
	}
	bus.Subscribe(eventbus.EventToolCall, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		tool, _ := e.Data["tool"].(string)
		args, _ := e.Data["args"].(map[string]any)
		_ = store.Append(audit.Record{
			At:           e.At,
			SessionID:    e.SessionID,
			TurnID:       e.TurnID,
			Tool:         tool,
			ArgsRedacted: audit.RedactArgs(args),
			ArgsHash:     audit.HashArgs(args),
			Kind:         "tool_call",
		})
		return nil, nil
	})
	bus.Subscribe(eventbus.EventToolResult, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		tool, _ := e.Data["tool"].(string)
		errFlag, _ := e.Data["error"].(bool)
		blocked, _ := e.Data["blocked"].(bool)
		_ = store.Append(audit.Record{
			At:        e.At,
			SessionID: e.SessionID,
			TurnID:    e.TurnID,
			Tool:      tool,
			Error:     errFlag,
			Blocked:   blocked,
			Kind:      "tool_result",
		})
		return nil, nil
	})
}
