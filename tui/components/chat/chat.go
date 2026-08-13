package chat

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/budaev/stell/tui/components"
	"github.com/budaev/stell/tui/render"
	"github.com/budaev/stell/tui/theme"
)

// Role is a chat line kind.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
	RoleError     Role = "error"
)

// Message is a chat line. ID is assigned on Append.
type Message struct {
	ID   int
	Role Role
	Text string
}

// HeavyJob is markdown that should run off the input loop.
type HeavyJob struct {
	ID    int
	Width int
	Text  string
}

// Model is a scrollable chat viewport.
type Model struct {
	theme     theme.Theme
	viewport  viewport.Model
	messages  []Message
	focused   bool
	width     int
	height    int
	cache     *render.Cache
	nextID    int
	pending   map[pendKey]bool
	workspace string
}

type pendKey struct {
	id, width int
}

// New creates an empty chat view.
func New(t theme.Theme) *Model {
	vp := viewport.New()
	vp.SoftWrap = true
	return &Model{
		theme:    t,
		viewport: vp,
		messages: nil,
		width:    80,
		height:   10,
		cache:    render.NewCache(0, 0),
		pending:  map[pendKey]bool{},
	}
}

// SetWorkspace jails inline-image paths to root.
func (m *Model) SetWorkspace(root string) {
	m.workspace = root
}

// SetSize updates the viewport. Markdown cache is keyed by width; height-only is cheap.
func (m *Model) SetSize(width, height int) []HeavyJob {
	if height < 1 {
		height = 1
	}
	widthChanged := m.width != width
	same := m.width == width && m.height == height
	m.width = width
	m.height = height
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height)
	if same {
		return nil
	}
	var jobs []HeavyJob
	if widthChanged {
		jobs = m.requeueHeavy()
	}
	m.refresh()
	return jobs
}

// Append adds a message and scrolls to bottom. Heavy assistant markdown is returned as jobs.
func (m *Model) Append(role Role, text string) []HeavyJob {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	m.nextID++
	msg := Message{ID: m.nextID, Role: role, Text: text}
	m.messages = append(m.messages, msg)
	jobs := m.ensure(msg)
	m.refresh()
	m.viewport.GotoBottom()
	return jobs
}

// ApplyDone stores an async render and refreshes.
func (m *Model) ApplyDone(id, width int, out string) {
	m.cache.Put(id, width, out)
	delete(m.pending, pendKey{id, width})
	m.refresh()
	m.viewport.GotoBottom()
}

// Messages returns a copy of chat history.
func (m *Model) Messages() []Message {
	out := make([]Message, len(m.messages))
	copy(out, m.messages)
	return out
}

func (m *Model) requeueHeavy() []HeavyJob {
	var jobs []HeavyJob
	for _, msg := range m.messages {
		jobs = append(jobs, m.ensure(msg)...)
	}
	return jobs
}

func (m *Model) ensure(msg Message) []HeavyJob {
	if msg.Role != RoleAssistant {
		return nil
	}
	key := pendKey{msg.ID, m.width}
	if _, ok := m.cache.Get(msg.ID, m.width); ok {
		delete(m.pending, key)
		return nil
	}
	if len(msg.Text) < render.AsyncMarkdownBytes || render.AsyncOff() {
		out := render.Markdown(msg.Text, m.width, m.theme)
		m.cache.Put(msg.ID, m.width, out)
		return nil
	}
	if m.pending[key] {
		return nil
	}
	m.pending[key] = true
	return []HeavyJob{{ID: msg.ID, Width: m.width, Text: msg.Text}}
}

func (m *Model) refresh() {
	var b strings.Builder
	for i, msg := range m.messages {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(m.format(msg))
	}
	m.viewport.SetContent(b.String())
}

func (m *Model) format(msg Message) string {
	var label, color string
	switch msg.Role {
	case RoleUser:
		label, color = "you", m.theme.Colors.Primary
	case RoleAssistant:
		label, color = "agent", m.theme.Colors.Secondary
	case RoleTool:
		label, color = "tool", m.theme.Colors.Accent
	case RoleError:
		label, color = "error", m.theme.Colors.Error
	default:
		label, color = "sys", m.theme.Colors.Muted
	}
	head := lipgloss.NewStyle().Foreground(theme.C(color)).Bold(true).Render(label)
	body := msg.Text
	if msg.Role == RoleAssistant {
		body = m.assistantBody(msg)
	} else {
		body = lipgloss.NewStyle().Foreground(m.theme.Foreground()).Render(msg.Text)
	}
	if img := render.InlineImages(msg.Text, m.width, m.workspace); img != "" {
		body += "\n" + img
	}
	return head + "\n" + body
}

func (m *Model) assistantBody(msg Message) string {
	if cached, ok := m.cache.Get(msg.ID, m.width); ok {
		return cached
	}
	if m.pending[pendKey{msg.ID, m.width}] {
		return lipgloss.NewStyle().Foreground(m.theme.Muted()).Italic(true).Render("rendering…")
	}
	if len(msg.Text) < render.AsyncMarkdownBytes || render.AsyncOff() {
		out := render.Markdown(msg.Text, m.width, m.theme)
		m.cache.Put(msg.ID, m.width, out)
		return out
	}
	return lipgloss.NewStyle().Foreground(m.theme.Muted()).Italic(true).Render("rendering…")
}

// Render implements components.Component.
func (m *Model) Render(width int) []string {
	m.width = width
	m.viewport.SetWidth(width)
	view := m.viewport.View()
	if strings.TrimSpace(view) == "" {
		hint := lipgloss.NewStyle().Foreground(m.theme.Muted()).Italic(true).
			Render("No messages yet. Type a task below and press ctrl+s.")
		return []string{hint}
	}
	return strings.Split(view, "\n")
}

// Update implements components.Component.
func (m *Model) Update(msg tea.Msg) (components.Component, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// Focus focuses the chat for scrolling.
func (m *Model) Focus() { m.focused = true }

// Blur unfocuses the chat.
func (m *Model) Blur() { m.focused = false }

func (m *Model) WantsKeyRelease() bool { return false }
func (m *Model) Invalidate()           { m.refresh() }

// Focused reports chat focus.
func (m *Model) Focused() bool { return m.focused }

// CacheLen is a test helper.
func (m *Model) CacheLen() int { return m.cache.Len() }
