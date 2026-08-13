package events_test

import (
	"testing"

	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/tui/events"
)

func TestFormatLineToolCall(t *testing.T) {
	role, text := events.FormatLine(events.BusMsg{
		Type: eventbus.EventToolCall,
		Data: map[string]any{"tool": "read_file"},
	})
	if role != "tool" || text != "→ read_file" {
		t.Fatalf("role=%q text=%q", role, text)
	}
}

func TestTokensFrom(t *testing.T) {
	if n := events.TokensFrom(map[string]any{"tokens": 42.0}); n != 42 {
		t.Fatalf("got %d", n)
	}
}
