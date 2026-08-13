package editor

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/budaev/agent/tui/complete"
	"github.com/budaev/agent/tui/components"
	"github.com/budaev/agent/tui/theme"
)

const defaultHeight = 4

// Model is a multi-line prompt box with an accent rail.
type Model struct {
	theme    theme.Theme
	textarea textarea.Model
	focused  bool
	width    int
	height   int
	ac       *Autocomplete
}

// New creates an editor focused by default.
func New(t theme.Theme) *Model {
	return NewWithComplete(t, complete.DefaultCommands(), nil)
}

// NewWithComplete wires slash/file providers.
func NewWithComplete(t theme.Theme, cmds []complete.Command, files *complete.FileIndex) *Model {
	ta := textarea.New()
	ta.Placeholder = "Describe a task…  ctrl+s send · enter newline · tab focus"
	ta.CharLimit = 20000
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.SetHeight(defaultHeight)

	styles := textarea.DefaultDarkStyles()
	styles.Focused.Base = lipgloss.NewStyle()
	styles.Blurred.Base = lipgloss.NewStyle()
	styles.Focused.CursorLine = lipgloss.NewStyle()
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(t.Muted())
	ta.SetStyles(styles)
	_ = ta.Focus()

	return &Model{
		theme:    t,
		textarea: ta,
		focused:  true,
		width:    80,
		height:   defaultHeight,
		ac:       newAutocomplete(t, cmds, files),
	}
}

// SetWidth resizes the inner textarea.
func (m *Model) SetWidth(width int) {
	m.width = width
	inner := width - 4
	if inner < 10 {
		inner = 10
	}
	m.textarea.SetWidth(inner)
}

// Render implements components.Component.
func (m *Model) Render(width int) []string {
	m.SetWidth(width)
	borderColor := m.theme.Accent()
	if !m.focused {
		borderColor = m.theme.Border()
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(max(10, width-2)).
		Render(m.textarea.View())
	lines := strings.Split(box, "\n")
	if m.ac != nil {
		lines = append(lines, m.ac.render(width)...)
	}
	return lines
}

// Update implements components.Component.
func (m *Model) Update(msg tea.Msg) (components.Component, tea.Cmd) {
	if _, ok := msg.(filesReadyMsg); ok {
		return m, m.ac.update(msg, m.textarea.Value())
	}
	if !m.focused {
		return m, nil
	}
	if kp, ok := msg.(tea.KeyPressMsg); ok && m.ac.Visible() {
		switch kp.String() {
		case "tab", "enter":
			if next, ok := m.ac.accept(m.textarea.Value()); ok {
				m.textarea.SetValue(next)
			}
			return m, nil
		case "up", "down", "esc", "shift+tab":
			return m, m.ac.update(msg, m.textarea.Value())
		}
	}
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	acCmd := m.ac.update(msg, m.textarea.Value())
	return m, tea.Batch(cmd, acCmd)
}

// Focus focuses the textarea.
func (m *Model) Focus() {
	m.focused = true
	_ = m.textarea.Focus()
}

// Blur unfocuses the textarea.
func (m *Model) Blur() {
	m.focused = false
	m.textarea.Blur()
}

func (m *Model) WantsKeyRelease() bool { return true }
func (m *Model) Invalidate()           {}

// Focused reports whether the editor owns keyboard input.
func (m *Model) Focused() bool { return m.focused }

// Value returns the current text.
func (m *Model) Value() string { return m.textarea.Value() }

// SetValue replaces the editor contents.
func (m *Model) SetValue(v string) { m.textarea.SetValue(v) }

// Reset clears the editor.
func (m *Model) Reset() {
	m.textarea.Reset()
	if m.ac != nil {
		m.ac.hide()
	}
}

// PopupVisible reports an open completion list.
func (m *Model) PopupVisible() bool { return m.ac.Visible() }

// PopupHeight is extra rows used by the completion popup.
func (m *Model) PopupHeight() int {
	if m.ac == nil {
		return 0
	}
	return m.ac.height()
}

// RefreshComplete runs providers immediately (tests / after SetValue).
func (m *Model) RefreshComplete() {
	if m.ac == nil {
		return
	}
	if m.ac.files != nil && !m.ac.files.Fresh() {
		m.ac.files.Rebuild()
	}
	m.ac.refresh(m.textarea.Value())
}

// AcceptCompletion inserts the selected item.
func (m *Model) AcceptCompletion() bool {
	next, ok := m.ac.accept(m.textarea.Value())
	if ok {
		m.textarea.SetValue(next)
	}
	return ok
}
