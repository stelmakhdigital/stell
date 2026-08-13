package render_test

import (
	"os"
	"strings"
	"testing"

	"github.com/budaev/stell/tui/render"
	"github.com/budaev/stell/tui/renderer"
	"github.com/budaev/stell/tui/theme"
)

func TestMarkdownHeadingsListsFences(t *testing.T) {
	src, err := os.ReadFile("testdata/basic.md")
	if err != nil {
		t.Fatal(err)
	}
	out := render.Markdown(string(src), 80, theme.Default())
	plain := renderer.StripANSI(out)
	if !strings.Contains(plain, "# Title") {
		t.Fatalf("heading missing: %q", plain)
	}
	if !strings.Contains(plain, "•") && !strings.Contains(plain, "item one") {
		t.Fatalf("list missing: %q", plain)
	}
	if !strings.Contains(plain, "Hello") {
		t.Fatalf("go fence body missing: %q", plain)
	}
	if !strings.Contains(plain, "plain fence") {
		t.Fatalf("unknown lexer should stay plain: %q", plain)
	}
}

func TestHighlightUnknownPlain(t *testing.T) {
	got := render.HighlightCode("foo bar", "not-a-lang")
	if got != "foo bar" {
		t.Fatalf("%q", got)
	}
}

func TestHighlightGoDiffersFromPlain(t *testing.T) {
	code := "func Hello() {}"
	hi := render.HighlightCode(code, "go")
	if hi == "" {
		t.Fatal("empty highlight")
	}
	if renderer.StripANSI(hi) == "" {
		t.Fatal("stripped empty")
	}
}

func TestMarkdownTruncatesLongDumps(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 250; i++ {
		b.WriteString("- item\n")
	}
	out := render.Markdown(b.String(), 80, theme.Default())
	n := strings.Count(out, "\n") + 1
	if n > 200 {
		t.Fatalf("expected truncation to 200 lines, got %d", n)
	}
	if !strings.Contains(out, "truncated") {
		t.Fatalf("missing truncation marker")
	}
}
