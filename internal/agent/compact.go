package agent

import (
	"strings"
	"unicode/utf8"

	"github.com/budaev/stell/internal/llm"
)

// EstimateTokens approximates tokens as runes/4 (tiktoken-compatible fallback).
func EstimateTokens(messages []llm.Message) int {
	n := 0
	for _, m := range messages {
		n += utf8.RuneCountInString(m.Content)
		n += utf8.RuneCountInString(m.Name)
		n += utf8.RuneCountInString(string(m.Role))
		for _, tc := range m.ToolCalls {
			n += utf8.RuneCountInString(tc.Function.Name)
			n += utf8.RuneCountInString(tc.Function.Arguments)
		}
	}
	if n == 0 {
		return 0
	}
	return (n + 3) / 4
}

// CompactMessages replaces middle turns with a summary that keeps decisions,
// constraints, and entity names. System + last keepLast messages are retained.
func CompactMessages(messages []llm.Message, keepLast int) []llm.Message {
	if keepLast < 2 {
		keepLast = 4
	}
	if len(messages) <= keepLast+1 {
		return messages
	}
	sys := messages[0]
	tail := messages[len(messages)-keepLast:]
	dropped := messages[1 : len(messages)-keepLast]
	summary := extractiveSummary(dropped)
	out := make([]llm.Message, 0, 2+len(tail))
	out = append(out, sys)
	out = append(out, llm.Message{
		Role:    llm.RoleUser,
		Content: "[compacted context]\n" + summary,
	})
	out = append(out, tail...)
	return out
}

func extractiveSummary(msgs []llm.Message) string {
	var decisions, constraints, entities []string
	seenEnt := map[string]struct{}{}
	for _, m := range msgs {
		for _, line := range strings.Split(m.Content, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			low := strings.ToLower(line)
			switch {
			case strings.Contains(low, "decision") || strings.Contains(low, "решен"):
				decisions = append(decisions, line)
			case strings.Contains(low, "constraint") || strings.Contains(low, "ограничен") || strings.Contains(low, "must "):
				constraints = append(constraints, line)
			}
			for _, tok := range strings.Fields(line) {
				if looksEntity(tok) {
					if _, ok := seenEnt[tok]; ok {
						continue
					}
					seenEnt[tok] = struct{}{}
					entities = append(entities, tok)
				}
			}
		}
	}
	var b strings.Builder
	b.WriteString("Summary: older turns compacted; noise and resolved errors dropped.\n")
	b.WriteString("Key decisions:\n")
	writeList(&b, decisions, 12)
	b.WriteString("Constraints:\n")
	writeList(&b, constraints, 12)
	b.WriteString("Key entities:\n")
	writeList(&b, entities, 24)
	return b.String()
}

func writeList(b *strings.Builder, items []string, max int) {
	if len(items) == 0 {
		b.WriteString("- (none)\n")
		return
	}
	if len(items) > max {
		items = items[:max]
	}
	for _, it := range items {
		b.WriteString("- ")
		b.WriteString(it)
		b.WriteByte('\n')
	}
}

func looksEntity(tok string) bool {
	if len(tok) < 3 || len(tok) > 64 {
		return false
	}
	if strings.ContainsAny(tok, "/.\\") {
		return true
	}
	upper := 0
	for _, r := range tok {
		if r >= 'A' && r <= 'Z' {
			upper++
		}
	}
	return upper >= 2
}
