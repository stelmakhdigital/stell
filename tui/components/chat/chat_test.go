package chat_test

import (
	"strings"
	"testing"

	"github.com/budaev/stell/tui/components/chat"
	"github.com/budaev/stell/tui/renderer"
	"github.com/budaev/stell/tui/theme"
)

func TestAssistantRendersMarkdown(t *testing.T) {
	c := chat.New(theme.Default())
	c.SetSize(80, 12)
	c.Append(chat.RoleAssistant, "# Hello\n\n- one")
	plain := renderer.StripANSI(strings.Join(c.Render(80), "\n"))
	if !strings.Contains(plain, "# Hello") {
		t.Fatalf("heading missing: %q", plain)
	}
	if !strings.Contains(plain, "one") {
		t.Fatalf("list missing: %q", plain)
	}
	c.SetSize(40, 12)
	plain = renderer.StripANSI(strings.Join(c.Render(40), "\n"))
	if !strings.Contains(plain, "# Hello") {
		t.Fatalf("after resize: %q", plain)
	}
}

func TestToolLineIsSingleLine(t *testing.T) {
	c := chat.New(theme.Default())
	c.SetSize(80, 12)
	c.Append(chat.RoleTool, "→ glob  **/*.go")
	plain := renderer.StripANSI(strings.Join(c.Render(80), "\n"))
	if !strings.Contains(plain, "tool  → glob  **/*.go") {
		t.Fatalf("expected single-line tool step: %q", plain)
	}
	if strings.Count(plain, "\n") > 1 && strings.Contains(plain, "tool\n") {
		t.Fatalf("tool label should not sit on its own line: %q", plain)
	}
}

func TestHeavyAssistantIsAsync(t *testing.T) {
	c := chat.New(theme.Default())
	c.SetSize(80, 12)
	body := strings.Repeat("paragraph of markdown text that is not tiny.\n\n", 200)
	jobs := c.Append(chat.RoleAssistant, body)
	if len(jobs) == 0 {
		t.Fatal("expected async job for large markdown")
	}
	plain := renderer.StripANSI(strings.Join(c.Render(80), "\n"))
	if !strings.Contains(plain, "rendering") {
		t.Fatalf("placeholder missing: %q", plain)
	}
	c.ApplyDone(jobs[0].ID, jobs[0].Width, "done-body")
	plain = renderer.StripANSI(strings.Join(c.Render(80), "\n"))
	if !strings.Contains(plain, "done-body") {
		t.Fatalf("apply missing: %q", plain)
	}
}
