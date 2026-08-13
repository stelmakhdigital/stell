package app_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/tui/app"
	"github.com/budaev/stell/tui/components"
	"github.com/budaev/stell/tui/components/chat"
	"github.com/budaev/stell/tui/events"
	"github.com/budaev/stell/tui/renderer"
	"github.com/budaev/stell/tui/theme"
)

func TestBusEventsAppearInChat(t *testing.T) {
	m := app.New(app.Config{Theme: theme.Default()})
	updated, _ := m.Update(events.BusMsg{
		Type: eventbus.EventToolCall,
		Data: map[string]any{"tool": "grep"},
	})
	model := updated.(app.Model)
	// Drive layout so chat has size.
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(app.Model)

	found := false
	for _, msg := range collectChat(model) {
		if msg.Role == chat.RoleTool && msg.Text == "→ grep" {
			found = true
		}
	}
	if !found {
		t.Fatalf("tool line missing: %+v", collectChat(model))
	}
}

func TestSubmitWithoutController(t *testing.T) {
	m := app.New(app.Config{Theme: theme.Default()})
	m.EditorSet("hello world")
	updated, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	model := updated.(app.Model)
	msgs := collectChat(model)
	if len(msgs) == 0 {
		t.Fatal("expected chat messages after submit")
	}
}

func TestViewRecordsDiffStats(t *testing.T) {
	m := app.New(app.Config{Theme: theme.Default()})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model := updated.(app.Model)
	_ = model.View()
	full, patch0 := model.DiffStats()
	if full == 0 {
		t.Fatal("first view should record a full frame")
	}
	model.EditorSet("hello")
	_ = model.View()
	_, patch1 := model.DiffStats()
	if patch1 <= patch0 {
		t.Fatal("second view should record a patch")
	}
}

func TestKeyReleaseDoesNotPanic(t *testing.T) {
	m := app.New(app.Config{Theme: theme.Default()})
	updated, _ := m.Update(tea.KeyReleaseMsg{Code: 'x'})
	_ = updated.(app.Model)
	updated, _ = m.Update(tea.KeyboardEnhancementsMsg{})
	_ = updated.(app.Model)
}

func TestRegistryRendersBelowChat(t *testing.T) {
	reg := components.NewRegistry()
	reg.Register("banner", components.SlotBelowChat, stubLine{s: "EXTRA-SLOT"})
	m := app.New(app.Config{Theme: theme.Default(), Registry: reg})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model := updated.(app.Model)
	v := model.View()
	if !strings.Contains(v.Content, "EXTRA-SLOT") {
		t.Fatalf("registry missing: %q", v.Content)
	}
}

func TestViewHasHorizontalPadding(t *testing.T) {
	m := app.New(app.Config{Theme: theme.Default()})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model := updated.(app.Model)
	plain := renderer.StripANSI(model.View().Content)
	checked := 0
	for _, line := range strings.Split(plain, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			t.Fatalf("missing left pad: %q", line)
		}
		if len([]rune(line)) > 78 {
			t.Fatalf("line hugs right edge (%d): %q", len([]rune(line)), line)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no content lines")
	}
}

type stubLine struct{ s string }

func (s stubLine) Render(int) []string { return []string{s.s} }
func (s stubLine) Update(tea.Msg) (components.Component, tea.Cmd) {
	return s, nil
}
func (s stubLine) Focus()                {}
func (s stubLine) Blur()                 {}
func (s stubLine) WantsKeyRelease() bool { return false }
func (s stubLine) Invalidate()           {}

// helpers via exported test hooks — see export_test.go
func collectChat(m app.Model) []chat.Message {
	return m.ChatMessages()
}
