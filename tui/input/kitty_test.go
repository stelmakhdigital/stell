package input_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/tui/input"
)

func TestDetectXtermIsFalse(t *testing.T) {
	env := func(k string) string {
		switch k {
		case "TERM":
			return "xterm-256color"
		case "TERM_PROGRAM", "KITTY_WINDOW_ID":
			return ""
		default:
			return ""
		}
	}
	if input.Detect(env) {
		t.Fatal("xterm must not look like kitty")
	}
}

func TestDetectKitty(t *testing.T) {
	env := func(k string) string {
		if k == "KITTY_WINDOW_ID" {
			return "1"
		}
		return ""
	}
	if !input.Detect(env) {
		t.Fatal("expected kitty")
	}
}

func TestKeyReleaseNoPanic(t *testing.T) {
	msg := tea.KeyReleaseMsg{Code: 'a'}
	if !input.IsKeyRelease(msg) {
		t.Fatal("expected release")
	}
	var v tea.View
	input.Apply(&v) // must not panic on any TERM
}

func TestReleaseSupportedDegrades(t *testing.T) {
	var msg tea.KeyboardEnhancementsMsg // Flags==0: xterm
	if input.ReleaseSupported(msg) {
		t.Fatal("xterm flags must not claim key-release")
	}
}
