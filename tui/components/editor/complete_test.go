package editor_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/tui/complete"
	"github.com/budaev/stell/tui/components/editor"
	"github.com/budaev/stell/tui/theme"
)

func TestSlashOpensCommands(t *testing.T) {
	ed := editor.New(theme.Default())
	ed.SetValue("/pl")
	ed.RefreshComplete()
	if !ed.PopupVisible() {
		t.Fatal("expected popup for /pl")
	}
	if !ed.AcceptCompletion() {
		t.Fatal("expected accept")
	}
	if !strings.Contains(ed.Value(), "/plan") {
		t.Fatalf("tab should insert /plan, got %q", ed.Value())
	}
}

func TestEscClosesPopup(t *testing.T) {
	ed := editor.New(theme.Default())
	ed.SetValue("/pl")
	ed.RefreshComplete()
	if !ed.PopupVisible() {
		t.Fatal("expected popup")
	}
	comp, _ := ed.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	ed = comp.(*editor.Model)
	if ed.PopupVisible() {
		t.Fatal("esc should close popup")
	}
}

func TestFileCompleteFromWorkspace(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "loop.go"), []byte("package x"), 0o644); err != nil {
		t.Fatal(err)
	}
	ed := editor.NewWithComplete(theme.Default(), complete.DefaultCommands(), complete.NewFileIndex(root))
	ed.SetValue("@loop")
	ed.RefreshComplete()
	if !ed.PopupVisible() {
		t.Fatal("expected file popup")
	}
	if !ed.AcceptCompletion() {
		t.Fatal("expected accept")
	}
	if !strings.Contains(ed.Value(), "loop.go") {
		t.Fatalf("got %q", ed.Value())
	}
}
