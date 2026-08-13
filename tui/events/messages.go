package events

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/pkg/observability"
)

const detailMaxRunes = 64

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
		return "tool", joinTool("→", tool, toolCallDetail(tool, msg.Data))
	case eventbus.EventToolResult:
		tool, _ := msg.Data["tool"].(string)
		return "tool", joinTool("←", tool, toolResultDetail(msg.Data))
	default:
		return "", ""
	}
}

func joinTool(arrow, tool, detail string) string {
	line := arrow + " " + tool
	if detail == "" {
		return line
	}
	return line + "  " + detail
}

func toolCallDetail(tool string, data map[string]any) string {
	args, _ := data["args"].(map[string]any)
	if len(args) == 0 {
		return ""
	}
	args = observability.RedactMap(args)
	switch tool {
	case "read_file", "write_file":
		return clip(strArg(args, "path"))
	case "bash":
		return clip(collapseWS(strArg(args, "command")))
	case "glob":
		return clip(strArg(args, "pattern"))
	case "grep":
		p := strArg(args, "pattern")
		if path := strArg(args, "path"); path != "" {
			return clip(p + "  " + path)
		}
		return clip(p)
	default:
		return clip(genericArgs(args))
	}
}

func toolResultDetail(data map[string]any) string {
	preview := clip(collapseWS(observability.Redact(strAny(data["preview"]))))
	blocked, _ := data["blocked"].(bool)
	errFlag, _ := data["error"].(bool)
	switch {
	case blocked:
		if preview == "" {
			return "blocked"
		}
		if strings.HasPrefix(preview, "blocked") {
			return preview
		}
		return "blocked: " + preview
	case errFlag:
		if preview == "" {
			return "error"
		}
		return "error: " + preview
	default:
		return preview
	}
}

func genericArgs(args map[string]any) string {
	skip := map[string]bool{"content": true, "body": true, "data": true, "text": true}
	preferred := []string{"path", "pattern", "command", "query", "url", "name", "file", "target"}
	var parts []string
	seen := map[string]bool{}
	for _, k := range preferred {
		if skip[k] {
			continue
		}
		s := strArg(args, k)
		if s == "" {
			continue
		}
		parts = append(parts, s)
		seen[k] = true
	}
	if len(parts) > 0 {
		return strings.Join(parts, "  ")
	}
	keys := make([]string, 0, len(args))
	for k := range args {
		if skip[k] || seen[k] {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s, ok := args[k].(string)
		if !ok || s == "" || utf8.RuneCountInString(s) > 40 {
			continue
		}
		parts = append(parts, k+"="+s)
	}
	return strings.Join(parts, "  ")
}

func strArg(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	return strAny(args[key])
}

func strAny(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		s = strings.TrimSpace(fmt.Sprint(v))
	}
	return strings.TrimSpace(s)
}

func collapseWS(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func clip(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) <= detailMaxRunes {
		return s
	}
	return string([]rune(s)[:detailMaxRunes-1]) + "…"
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
