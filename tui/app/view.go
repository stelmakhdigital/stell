package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/tui/components"
	"github.com/budaev/stell/tui/input"
)

// View implements tea.Model.
func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	if m.quitting {
		v.SetContent("")
		return v
	}
	if m.width == 0 {
		m.width = 80
		m.height = 24
	}
	m.layout()

	parts := []string{
		components.JoinLines(m.header.Render(m.width)),
		components.JoinLines(m.chat.Render(m.width)),
	}
	if extra := m.extras.RenderSlot(components.SlotBelowChat, m.width); len(extra) > 0 {
		parts = append(parts, components.JoinLines(extra))
	}
	if spin := m.spinner.Render(m.width); len(spin) > 0 {
		parts = append(parts, components.JoinLines(spin))
	}
	parts = append(parts, components.JoinLines(m.editor.Render(m.width)))
	if extra := m.extras.RenderSlot(components.SlotAboveFooter, m.width); len(extra) > 0 {
		parts = append(parts, components.JoinLines(extra))
	}
	parts = append(parts, components.JoinLines(m.footer.Render(m.width)))
	content := strings.Join(parts, "\n")
	// Bubble Tea paints the frame; Observe records cell-diff stats / STELL_TUI_FULL_REDRAW.
	if m.diff != nil {
		_, _ = m.diff.Observe(m.width, m.height, content)
	}
	v.SetContent(content)
	input.Apply(&v)
	return v
}

// ContentString is the rendered UI without tea.View (tests).
func (m Model) ContentString() string {
	return strings.Join(m.header.Render(m.width), "\n")
}
