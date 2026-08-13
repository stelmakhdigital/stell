package events_test

import (
	"strings"
	"testing"

	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/tui/events"
)

func TestFormatLineToolCall(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		want string
	}{
		{
			name: "name only",
			data: map[string]any{"tool": "read_file"},
			want: "→ read_file",
		},
		{
			name: "glob pattern",
			data: map[string]any{
				"tool": "glob",
				"args": map[string]any{"pattern": "**/*.go"},
			},
			want: "→ glob  **/*.go",
		},
		{
			name: "grep pattern and path",
			data: map[string]any{
				"tool": "grep",
				"args": map[string]any{"pattern": "FormatLine", "path": "tui/"},
			},
			want: "→ grep  FormatLine  tui/",
		},
		{
			name: "write_file skips content",
			data: map[string]any{
				"tool": "write_file",
				"args": map[string]any{"path": "tui/events/messages.go", "content": "secret body"},
			},
			want: "→ write_file  tui/events/messages.go",
		},
		{
			name: "bash command",
			data: map[string]any{
				"tool": "bash",
				"args": map[string]any{"command": "git status --short"},
			},
			want: "→ bash  git status --short",
		},
		{
			name: "unknown tool uses query",
			data: map[string]any{
				"tool": "mcp_search",
				"args": map[string]any{"query": "bubble tea"},
			},
			want: "→ mcp_search  bubble tea",
		},
		{
			name: "redacts token in command",
			data: map[string]any{
				"tool": "bash",
				"args": map[string]any{"command": "curl -H token=supersecret https://example"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, text := events.FormatLine(events.BusMsg{
				Type: eventbus.EventToolCall,
				Data: tt.data,
			})
			if role != "tool" {
				t.Fatalf("role=%q", role)
			}
			if tt.want != "" && text != tt.want {
				t.Fatalf("text=%q want %q", text, tt.want)
			}
			if strings.Contains(text, "supersecret") {
				t.Fatalf("secret leaked: %q", text)
			}
			if strings.Contains(text, "secret body") {
				t.Fatalf("file content leaked: %q", text)
			}
		})
	}
}

func TestFormatLineToolResult(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		want string
	}{
		{
			name: "ok with preview",
			data: map[string]any{"tool": "glob", "preview": "tui/app/update.go\ntui/events/messages.go"},
			want: "← glob  tui/app/update.go tui/events/messages.go",
		},
		{
			name: "error",
			data: map[string]any{"tool": "glob", "error": true, "preview": "pattern is required"},
			want: "← glob  error: pattern is required",
		},
		{
			name: "blocked",
			data: map[string]any{"tool": "bash", "blocked": true, "error": true, "preview": "blocked by policy"},
			want: "← bash  blocked by policy",
		},
		{
			name: "name only",
			data: map[string]any{"tool": "grep"},
			want: "← grep",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, text := events.FormatLine(events.BusMsg{
				Type: eventbus.EventToolResult,
				Data: tt.data,
			})
			if role != "tool" || text != tt.want {
				t.Fatalf("role=%q text=%q want %q", role, text, tt.want)
			}
		})
	}
}

func TestTokensFrom(t *testing.T) {
	if n := events.TokensFrom(map[string]any{"tokens": 42.0}); n != 42 {
		t.Fatalf("got %d", n)
	}
}
