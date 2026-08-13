package agent

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/budaev/stell/pkg/observability"
)

const previewMaxRunes = 80

// toolResultPreview is a short redacted line for EventToolResult (TUI / API).
func toolResultPreview(content, errText string) string {
	src := strings.TrimSpace(errText)
	if src == "" {
		src = strings.TrimSpace(content)
	}
	src = observability.Redact(src)
	if src == "" {
		return "ok"
	}
	collapsed := strings.Join(strings.Fields(src), " ")
	if utf8.RuneCountInString(collapsed) <= previewMaxRunes {
		return collapsed
	}
	lines := strings.Count(src, "\n") + 1
	if lines > 1 {
		return strconv.Itoa(lines) + " lines  " + clipRunes(collapsed, 56)
	}
	return clipRunes(collapsed, previewMaxRunes)
}

func clipRunes(s string, max int) string {
	if max < 2 {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max-1]) + "…"
}
