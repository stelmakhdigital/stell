package header

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/budaev/agent/tui/components"
	"github.com/budaev/agent/tui/theme"
)

var logoFrames = []string{">_", "=>", ">>", "*>"}

// TickMsg advances the logo animation.
type TickMsg time.Time

// Model is the top status bar.
type Model struct {
	theme       theme.Theme
	logoFrame   int
	agentStatus string
	modelName   string
	running     bool
	width       int
}

// New creates a header.
func New(t theme.Theme, modelName string) *Model {
	if modelName == "" {
		modelName = "local"
	}
	return &Model{
		theme:       t,
		agentStatus: "idle",
		modelName:   modelName,
		width:       80,
	}
}

// SetStatus updates the idle/running label.
func (m *Model) SetStatus(status string, running bool) {
	m.agentStatus = status
	m.running = running
}

// SetModelName updates the displayed model.
func (m *Model) SetModelName(name string) {
	m.modelName = name
}

// Render implements components.Component.
func (m *Model) Render(width int) []string {
	m.width = width
	logo := lipgloss.NewStyle().Foreground(m.theme.Accent()).Bold(true).Render(m.logo())
	tagline := lipgloss.NewStyle().Foreground(m.theme.Accent()).Render(m.theme.Tagline)
	title := lipgloss.JoinHorizontal(lipgloss.Left, logo, "  ", tagline)

	statusColor := m.theme.Secondary()
	if m.running {
		statusColor = m.theme.Warning()
	}
	status := lipgloss.NewStyle().Foreground(statusColor).Render(fmt.Sprintf("[%s]", m.agentStatus))
	model := lipgloss.NewStyle().Foreground(m.theme.Muted()).Render("model: " + m.modelName)
	bar := lipgloss.JoinHorizontal(lipgloss.Left, status, "  ", model)

	line := lipgloss.NewStyle().Foreground(m.theme.Border()).Render(strings.Repeat("─", max(0, width)))
	return []string{title, bar, line}
}

func (m *Model) logo() string {
	if len(logoFrames) == 0 {
		return ">"
	}
	return logoFrames[m.logoFrame%len(logoFrames)]
}

// TickCmd schedules the next logo frame.
func TickCmd() tea.Cmd {
	return tea.Tick(220*time.Millisecond, func(t time.Time) tea.Msg { return TickMsg(t) })
}

// Update implements components.Component.
func (m *Model) Update(msg tea.Msg) (components.Component, tea.Cmd) {
	if _, ok := msg.(TickMsg); ok {
		m.logoFrame++
		return m, TickCmd()
	}
	return m, nil
}

func (m *Model) Focus()                {}
func (m *Model) Blur()                 {}
func (m *Model) WantsKeyRelease() bool { return false }
func (m *Model) Invalidate()           {}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
