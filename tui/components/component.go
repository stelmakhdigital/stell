package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Component is the Pi-style UI building block.
type Component interface {
	Render(width int) []string
	Update(msg tea.Msg) (Component, tea.Cmd)
	Focus()
	Blur()
	WantsKeyRelease() bool
	Invalidate()
}

// Focusable is a component that owns a cursor.
type Focusable interface {
	Component
	CursorPosition() (x, y int)
	SetCursorPosition(x, y int)
}

// JoinLines concatenates rendered lines.
func JoinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
