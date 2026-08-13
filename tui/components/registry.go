package components

import tea "charm.land/bubbletea/v2"

// Slot is a layout region extras can occupy without editing app internals.
type Slot string

const (
	SlotBelowChat   Slot = "below-chat"
	SlotAboveFooter Slot = "above-footer"
)

type registration struct {
	name string
	slot Slot
	comp Component
}

// Registry holds extra components (compile-time UI plugins).
type Registry struct {
	items []registration
}

// NewRegistry returns an empty registry (zero value is also usable).
func NewRegistry() *Registry { return &Registry{} }

// Register adds a named component to a slot. Later registers with the same
// name replace the previous one.
func (r *Registry) Register(name string, slot Slot, c Component) {
	if r == nil || c == nil || name == "" {
		return
	}
	for i, it := range r.items {
		if it.name == name {
			r.items[i] = registration{name: name, slot: slot, comp: c}
			return
		}
	}
	r.items = append(r.items, registration{name: name, slot: slot, comp: c})
}

// RenderSlot concatenates components in a slot.
func (r *Registry) RenderSlot(slot Slot, width int) []string {
	if r == nil {
		return nil
	}
	var lines []string
	for _, it := range r.items {
		if it.slot != slot {
			continue
		}
		lines = append(lines, it.comp.Render(width)...)
	}
	return lines
}

// UpdateAll forwards a message to every extra.
func (r *Registry) UpdateAll(msg tea.Msg) tea.Cmd {
	if r == nil {
		return nil
	}
	var cmds []tea.Cmd
	for i := range r.items {
		next, cmd := r.items[i].comp.Update(msg)
		r.items[i].comp = next
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// Len is the number of extras (tests).
func (r *Registry) Len() int {
	if r == nil {
		return 0
	}
	return len(r.items)
}
