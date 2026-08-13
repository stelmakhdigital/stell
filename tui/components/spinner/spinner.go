package spinner

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/budaev/agent/tui/components"
	"github.com/budaev/agent/tui/theme"
)

// Model wraps bubbles spinner with visibility.
type Model struct {
	theme   theme.Theme
	spinner spinner.Model
	visible bool
	message string
}

// New creates a hidden spinner.
func New(t theme.Theme) *Model {
	s := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	s.Style = lipgloss.NewStyle().Foreground(t.Accent())
	return &Model{
		theme:   t,
		spinner: s,
		message: "agent working",
	}
}

// SetVisible toggles the spinner.
func (m *Model) SetVisible(v bool) {
	m.visible = v
}

// Visible reports whether the spinner is shown.
func (m *Model) Visible() bool { return m.visible }

// SetMessage sets the status text next to the glyph.
func (m *Model) SetMessage(msg string) {
	m.message = msg
}

// Render implements components.Component.
func (m *Model) Render(width int) []string {
	if !m.visible {
		return nil
	}
	line := m.spinner.View() + " " + lipgloss.NewStyle().Foreground(m.theme.Warning()).Render(m.message)
	if width > 0 {
		line = lipgloss.NewStyle().MaxWidth(width).Render(line)
	}
	return []string{line}
}

// Update implements components.Component.
func (m *Model) Update(msg tea.Msg) (components.Component, tea.Cmd) {
	if !m.visible {
		return m, nil
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// Tick returns the start-tick command.
func (m *Model) Tick() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Focus()                {}
func (m *Model) Blur()                 {}
func (m *Model) WantsKeyRelease() bool { return false }
func (m *Model) Invalidate()           {}
