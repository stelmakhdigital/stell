package input

import (
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
)

const envKitty = "AGENT_TUI_KITTY"

// Enabled reports whether Kitty keyboard enhancements may be requested.
// AGENT_TUI_KITTY=0 disables; otherwise request and let the terminal degrade.
func Enabled() bool {
	v := os.Getenv(envKitty)
	return v != "0" && v != "false"
}

// Detect reports a Kitty-family keyboard terminal from env (best effort).
func Detect(getenv func(string) string) bool {
	if getenv == nil {
		getenv = os.Getenv
	}
	if getenv("KITTY_WINDOW_ID") != "" {
		return true
	}
	term := strings.ToLower(getenv("TERM"))
	if strings.Contains(term, "kitty") || strings.Contains(term, "ghostty") {
		return true
	}
	switch getenv("TERM_PROGRAM") {
	case "WezTerm", "ghostty", "kitty":
		return true
	}
	return false
}

// Apply requests key-release reporting on the view when enabled.
func Apply(v *tea.View) {
	if v == nil || !Enabled() {
		return
	}
	v.KeyboardEnhancements.ReportEventTypes = true
}

// IsKeyRelease reports a Kitty key-release event.
func IsKeyRelease(msg tea.Msg) bool {
	_, ok := msg.(tea.KeyReleaseMsg)
	return ok
}

// ReleaseSupported is true after KeyboardEnhancementsMsg from a capable terminal.
func ReleaseSupported(msg tea.KeyboardEnhancementsMsg) bool {
	return msg.SupportsEventTypes()
}
