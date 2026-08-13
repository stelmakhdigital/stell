package footer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/budaev/stell/tui/components"
	"github.com/budaev/stell/tui/theme"
)

// Model is a Starship-style status line.
type Model struct {
	theme      theme.Theme
	cwd        string
	gitBranch  string
	gitDirty   bool
	runtime    string
	tokensUsed int
	costUSD    float64
	help       string
	width      int
}

// New creates a footer with cwd/git snapshots.
func New(t theme.Theme) *Model {
	cwd, _ := os.Getwd()
	branch, dirty := detectGit(cwd)
	return &Model{
		theme:     t,
		cwd:       cwd,
		gitBranch: branch,
		gitDirty:  dirty,
		runtime:   "Go",
		help:      "ctrl+s send · esc cancel · tab focus · ctrl+q quit",
		width:     80,
	}
}

// SetMetrics updates token/cost display.
func (m *Model) SetMetrics(tokens int, cost float64) {
	m.tokensUsed = tokens
	m.costUSD = cost
}

// SetRuntime sets the runtime label (e.g. local/http).
func (m *Model) SetRuntime(name string) {
	if name != "" {
		m.runtime = name
	}
}

// Render implements components.Component.
func (m *Model) Render(width int) []string {
	m.width = width
	sep := lipgloss.NewStyle().Foreground(m.theme.Border()).Render(strings.Repeat("─", max(0, width)))

	cwd := lipgloss.NewStyle().Foreground(m.theme.Primary()).Bold(true).Render(shortenPath(m.cwd))
	git := ""
	if m.gitBranch != "" {
		mark := ""
		if m.gitDirty {
			mark = "*"
		}
		git = lipgloss.NewStyle().Foreground(m.theme.Secondary()).Render(m.gitBranch + mark)
	}
	rt := lipgloss.NewStyle().Foreground(m.theme.Accent()).Render(m.runtime)
	tok := lipgloss.NewStyle().Foreground(m.theme.Warning()).Render(fmt.Sprintf("tok %d", m.tokensUsed))
	cost := lipgloss.NewStyle().Foreground(m.theme.Muted()).Render(fmt.Sprintf("$%.4f", m.costUSD))

	parts := []string{cwd}
	if git != "" {
		parts = append(parts, git)
	}
	parts = append(parts, rt, tok, cost)
	line := strings.Join(parts, lipgloss.NewStyle().Foreground(m.theme.Muted()).Render("  ·  "))
	help := lipgloss.NewStyle().Foreground(m.theme.Muted()).Render(m.help)
	return []string{sep, line, help}
}

func (m *Model) Update(tea.Msg) (components.Component, tea.Cmd) { return m, nil }
func (m *Model) Focus()                                         {}
func (m *Model) Blur()                                          {}
func (m *Model) WantsKeyRelease() bool                          { return false }
func (m *Model) Invalidate()                                    {}

func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return filepath.Base(path)
}

func detectGit(dir string) (branch string, dirty bool) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	branch = strings.TrimSpace(string(out))
	st := exec.Command("git", "-C", dir, "status", "--porcelain")
	stOut, err := st.Output()
	if err == nil && len(bytesTrim(stOut)) > 0 {
		dirty = true
	}
	return branch, dirty
}

func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
