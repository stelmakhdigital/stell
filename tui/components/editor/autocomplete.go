package editor

import (
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/budaev/agent/tui/complete"
	"github.com/budaev/agent/tui/theme"
	"github.com/mattn/go-runewidth"
)

const debounce = 80 * time.Millisecond

type completeTickMsg struct{ gen int }
type filesReadyMsg struct{}

// Autocomplete is the slash/file popup attached to the editor.
type Autocomplete struct {
	cmds    []complete.Command
	files   *complete.FileIndex
	items   []complete.Item
	sel     int
	visible bool
	token   string
	start   int
	gen     int
	theme   theme.Theme
}

func newAutocomplete(th theme.Theme, cmds []complete.Command, files *complete.FileIndex) *Autocomplete {
	return &Autocomplete{cmds: cmds, files: files, theme: th}
}

// Visible reports whether the popup is shown.
func (a *Autocomplete) Visible() bool {
	return a != nil && a.visible && len(a.items) > 0
}

func (a *Autocomplete) height() int {
	if !a.Visible() {
		return 0
	}
	n := len(a.items)
	if n > 8 {
		n = 8
	}
	return n
}

func (a *Autocomplete) windowStart() int {
	n := a.height()
	if a.sel >= n {
		return a.sel - n + 1
	}
	return 0
}

func (a *Autocomplete) render(width int) []string {
	if !a.Visible() {
		return nil
	}
	n := a.height()
	start := a.windowStart()
	lines := make([]string, 0, n)
	for i := 0; i < n; i++ {
		idx := start + i
		if idx >= len(a.items) {
			break
		}
		it := a.items[idx]
		st := lipgloss.NewStyle().Foreground(a.theme.Muted())
		prefix := "  "
		if idx == a.sel {
			st = lipgloss.NewStyle().Foreground(a.theme.Primary()).Bold(true)
			prefix = "▸ "
		}
		label := clipWidth(it.Label, width-4)
		lines = append(lines, st.Render(prefix+label))
	}
	return lines
}

func (a *Autocomplete) update(msg tea.Msg, value string) tea.Cmd {
	if a == nil {
		return nil
	}
	switch msg := msg.(type) {
	case filesReadyMsg:
		a.refresh(value)
		return nil
	case completeTickMsg:
		if msg.gen != a.gen {
			return nil
		}
		a.refresh(value)
		return nil
	case tea.KeyPressMsg:
		if !a.Visible() {
			break
		}
		switch msg.String() {
		case "up", "shift+tab":
			if a.sel > 0 {
				a.sel--
			}
			return nil
		case "down":
			if a.sel < len(a.items)-1 {
				a.sel++
			}
			return nil
		case "esc":
			a.hide()
			return nil
		}
	}
	tok, _, ok := trailingToken(value)
	if ok && strings.HasPrefix(tok, "/") {
		a.refresh(value)
		return nil
	}
	if ok && isFileToken(tok) {
		if a.files != nil && a.files.Fresh() {
			a.refresh(value)
			return nil
		}
		a.gen++
		gen := a.gen
		return tea.Batch(a.loadFilesCmd(), tea.Tick(debounce, func(time.Time) tea.Msg {
			return completeTickMsg{gen: gen}
		}))
	}
	a.gen++
	gen := a.gen
	return tea.Tick(debounce, func(time.Time) tea.Msg { return completeTickMsg{gen: gen} })
}

func (a *Autocomplete) loadFilesCmd() tea.Cmd {
	if a.files == nil || !a.files.StartRebuild() {
		return nil
	}
	return func() tea.Msg {
		a.files.Rebuild()
		return filesReadyMsg{}
	}
}

func (a *Autocomplete) refresh(value string) {
	tok, start, ok := trailingToken(value)
	if !ok {
		a.hide()
		return
	}
	var items []complete.Item
	switch {
	case strings.HasPrefix(tok, "/"):
		items = complete.Commands(a.cmds, tok)
	case isFileToken(tok):
		if a.files != nil {
			items = a.files.Match(strings.TrimPrefix(tok, "@"))
		}
	default:
		a.hide()
		return
	}
	prev := a.token
	a.token, a.start, a.items = tok, start, items
	if prev != tok || a.sel >= len(items) {
		a.sel = 0
	}
	a.visible = len(items) > 0
}

func (a *Autocomplete) hide() {
	a.visible = false
	a.items = nil
	a.sel = 0
}

func (a *Autocomplete) accept(value string) (string, bool) {
	if !a.Visible() {
		return value, false
	}
	if a.sel < 0 || a.sel >= len(a.items) {
		return value, false
	}
	it := a.items[a.sel]
	prefix := value[:a.start]
	a.hide()
	return prefix + it.Value + " ", true
}

func trailingToken(value string) (tok string, start int, ok bool) {
	if value == "" {
		return "", 0, false
	}
	start = strings.LastIndexAny(value, " \n\t") + 1
	tok = value[start:]
	if tok == "" {
		return "", start, false
	}
	return tok, start, true
}

func isFileToken(tok string) bool {
	return strings.HasPrefix(tok, "@") || strings.Contains(tok, "/") || strings.Contains(tok, ".")
}

func clipWidth(s string, max int) string {
	if max <= 1 || runewidth.StringWidth(s) <= max {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if !utf8.ValidRune(r) {
			continue
		}
		w := runewidth.RuneWidth(r)
		if w < 1 {
			w = 1
		}
		if n+w > max-1 {
			b.WriteRune('…')
			break
		}
		b.WriteRune(r)
		n += w
	}
	return b.String()
}
