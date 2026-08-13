package components_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/agent/tui/components"
)

type stub struct {
	n    int
	line string
}

func (s *stub) Render(int) []string { return []string{s.line} }
func (s *stub) Update(tea.Msg) (components.Component, tea.Cmd) {
	s.n++
	return s, nil
}
func (s *stub) Focus()                {}
func (s *stub) Blur()                 {}
func (s *stub) WantsKeyRelease() bool { return false }
func (s *stub) Invalidate()           {}

func TestRegistryRenderAndReplace(t *testing.T) {
	r := components.NewRegistry()
	r.Register("a", components.SlotBelowChat, &stub{line: "hello"})
	r.Register("a", components.SlotBelowChat, &stub{line: "world"})
	lines := r.RenderSlot(components.SlotBelowChat, 40)
	if len(lines) != 1 || lines[0] != "world" {
		t.Fatalf("%q", lines)
	}
	if r.Len() != 1 {
		t.Fatalf("len=%d", r.Len())
	}
	r.UpdateAll(tea.KeyPressMsg{})
}

func TestRegistryNilSafe(t *testing.T) {
	var r *components.Registry
	if lines := r.RenderSlot(components.SlotBelowChat, 10); lines != nil {
		t.Fatalf("%v", lines)
	}
	if cmd := r.UpdateAll(nil); cmd != nil {
		t.Fatal(cmd)
	}
}
