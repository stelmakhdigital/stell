package agent

import (
	"strings"
	"unicode/utf8"
)

const truncationMarker = "...[truncated]..."

// CompressToolResult shrinks oversized tool output. Errors and path-like lines
// are kept; the middle is dropped with a truncation marker.
func CompressToolResult(content string, maxBytes int) string {
	if maxBytes <= 0 || len(content) <= maxBytes {
		return content
	}
	low := strings.ToLower(strings.TrimSpace(content))
	if strings.HasPrefix(low, "error:") || strings.HasPrefix(low, "blocked:") {
		if len(content) <= maxBytes*2 {
			return content
		}
	}
	var keep []string
	for _, line := range strings.Split(content, "\n") {
		if looksPathLine(line) {
			keep = append(keep, line)
		}
	}
	budget := maxBytes
	headN := budget / 3
	if headN < 64 {
		headN = 64
	}
	tailN := budget / 3
	if tailN < 64 {
		tailN = 64
	}
	if headN+tailN+len(truncationMarker)+1 >= len(content) {
		return content
	}
	head := content[:headN]
	tail := content[len(content)-tailN:]
	var b strings.Builder
	b.WriteString(head)
	if !strings.HasSuffix(head, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString(truncationMarker)
	b.WriteByte('\n')
	if len(keep) > 0 {
		b.WriteString("preserved paths:\n")
		maxPaths := 20
		if len(keep) < maxPaths {
			maxPaths = len(keep)
		}
		for _, p := range keep[:maxPaths] {
			b.WriteString(p)
			b.WriteByte('\n')
		}
	}
	b.WriteString(tail)
	out := b.String()
	if len(out) > maxBytes*2 {
		return out[:maxBytes] + "\n" + truncationMarker
	}
	_ = utf8.RuneCountInString
	return out
}

func looksPathLine(line string) bool {
	s := strings.TrimSpace(line)
	if s == "" {
		return false
	}
	return strings.Contains(s, "/") || strings.Contains(s, "\\") ||
		strings.HasSuffix(s, ".go") || strings.HasSuffix(s, ".md")
}
