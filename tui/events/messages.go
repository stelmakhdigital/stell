package events

import (
	tea "charm.land/bubbletea/v2"
	"github.com/budaev/agent/internal/eventbus"
)

// BusMsg is a tea message wrapping an agent Event Bus event.
type BusMsg struct {
	Type      eventbus.EventType
	SessionID string
	TurnID    string
	Data      map[string]any
}

// DoneMsg is sent when Agent.Run returns.
type DoneMsg struct {
	Text  string
	Err   error
	Turns int
}

// Listen waits for the next UI message from the event channel.
func Listen(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// FormatLine maps an event to a chat line (role, text). Empty text means skip.
func FormatLine(msg BusMsg) (role, text string) {
	switch msg.Type {
	case eventbus.EventToolCall:
		tool, _ := msg.Data["tool"].(string)
		return "tool", "→ " + tool
	case eventbus.EventToolResult:
		tool, _ := msg.Data["tool"].(string)
		errFlag, _ := msg.Data["error"].(bool)
		if errFlag {
			return "tool", "← " + tool + " (error)"
		}
		return "tool", "← " + tool
	default:
		return "", ""
	}
}

// TokensFrom extracts token count from model_response data.
func TokensFrom(data map[string]any) int {
	if data == nil {
		return 0
	}
	switch v := data["tokens"].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}
