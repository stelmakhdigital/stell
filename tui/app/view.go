package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/tui/components"
	"github.com/budaev/stell/tui/input"
)

// framePad is the left/right gutter so text does not hug the terminal edge.
const framePad = 2

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
	inner := m.innerWidth()

	parts := []string{
		components.JoinLines(m.header.Render(inner)),
		components.JoinLines(m.chat.Render(inner)),
	}
	if extra := m.extras.RenderSlot(components.SlotBelowChat, inner); len(extra) > 0 {
		parts = append(parts, components.JoinLines(extra))
	}
	if spin := m.spinner.Render(inner); len(spin) > 0 {
		parts = append(parts, components.JoinLines(spin))
	}
	parts = append(parts, components.JoinLines(m.editor.Render(inner)))
	if extra := m.extras.RenderSlot(components.SlotAboveFooter, inner); len(extra) > 0 {
		parts = append(parts, components.JoinLines(extra))
	}
	parts = append(parts, components.JoinLines(m.footer.Render(inner)))
	content := m.padFrame(strings.Join(parts, "\n"))
	// Bubble Tea paints the frame; Observe records cell-diff stats / STELL_TUI_FULL_REDRAW.
	if m.diff != nil {
		_, _ = m.diff.Observe(m.width, m.height, content)
	}
	v.SetContent(content)
	input.Apply(&v)
	return v
}

func (m Model) innerWidth() int {
	w := m.width - 2*framePad
	if w < 10 {
		if m.width < 1 {
			return 1
		}
		return m.width
	}
	return w
}

func (m Model) padFrame(content string) string {
	pad := (m.width - m.innerWidth()) / 2
	if pad <= 0 {
		return content
	}
	gutter := strings.Repeat(" ", pad)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = gutter + line
	}
	return strings.Join(lines, "\n")
}

// ContentString is the rendered UI without tea.View (tests).
func (m Model) ContentString() string {
	return strings.Join(m.header.Render(m.innerWidth()), "\n")
}
