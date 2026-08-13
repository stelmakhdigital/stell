package editor_test

import (
	"strings"
	"testing"

	"github.com/budaev/stell/tui/components/editor"
	"github.com/budaev/stell/tui/theme"
)

func TestEditorValueSubmit(t *testing.T) {
	ed := editor.New(theme.Default())
	ed.SetValue("  fix the loop  ")
	if got := strings.TrimSpace(ed.Value()); got != "fix the loop" {
		t.Fatalf("value=%q", got)
	}
	ed.Reset()
	if ed.Value() != "" {
		t.Fatalf("expected empty after reset, got %q", ed.Value())
	}
}

func TestEditorFocus(t *testing.T) {
	ed := editor.New(theme.Default())
	if !ed.Focused() {
		t.Fatal("expected focused by default")
	}
	ed.Blur()
	if ed.Focused() {
		t.Fatal("expected blurred")
	}
	ed.Focus()
	if !ed.Focused() {
		t.Fatal("expected focused")
	}
}
