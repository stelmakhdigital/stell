package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/tui/components/chat"
	"github.com/budaev/agent/tui/components/editor"
	"github.com/budaev/agent/tui/components/header"
	"github.com/budaev/agent/tui/components/spinner"
	"github.com/budaev/agent/tui/events"
	"github.com/budaev/agent/tui/input"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.diff != nil {
			m.diff.Resize(msg.Width, msg.Height)
		}
		return m, tea.Batch(m.layoutCmd(), m.extras.UpdateAll(msg))

	case header.TickMsg:
		comp, cmd := m.header.Update(msg)
		m.header = comp.(*header.Model)
		return m, tea.Batch(cmd, m.extras.UpdateAll(msg))

	case RenderDoneMsg:
		m.chat.ApplyDone(msg.ID, msg.Width, msg.Out)
		return m, m.async.Next()

	case tea.KeyboardEnhancementsMsg:
		m.keyRelease = input.ReleaseSupported(msg)
		return m, nil

	case tea.KeyReleaseMsg:
		// Swallow releases unless a focused component asked for them.
		if m.focus == focusEditor && m.editor.WantsKeyRelease() && m.keyRelease {
			return m, nil
		}
		return m, nil

	case events.BusMsg:
		cmd := m.handleBus(msg)
		return m, tea.Batch(cmd, m.listen())

	case events.DoneMsg:
		m.setRunning(false)
		var cmds []tea.Cmd
		if msg.Err != nil {
			cmds = append(cmds, m.appendChat(chat.RoleError, msg.Err.Error()))
		}
		if strings.TrimSpace(msg.Text) != "" {
			cmds = append(cmds, m.appendChat(chat.RoleAssistant, msg.Text))
		}
		cmds = append(cmds, m.listen())
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, m.forward(msg)
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+q":
		m.quitting = true
		if m.controller != nil {
			m.controller.Cancel()
		}
		return m, tea.Quit
	case "ctrl+c":
		if m.isRunning() {
			if m.controller != nil {
				m.controller.Cancel()
			}
			m.appendChat(chat.RoleSystem, "cancelled")
			m.setRunning(false)
			return m, nil
		}
		m.quitting = true
		return m, tea.Quit
	case "esc":
		if m.focus == focusEditor && m.editor.PopupVisible() {
			return m, m.forward(msg)
		}
		if m.isRunning() {
			if m.controller != nil {
				m.controller.Cancel()
			}
			m.appendChat(chat.RoleSystem, "cancelled")
			m.setRunning(false)
		}
		return m, nil
	case "tab":
		if m.focus == focusEditor && m.editor.PopupVisible() {
			return m, m.forward(msg)
		}
		m.toggleFocus()
		return m, nil
	}

	if m.focus == focusEditor && editor.IsSubmitKey(msg) {
		text := strings.TrimSpace(m.editor.Value())
		if text == "" || m.isRunning() {
			return m, nil
		}
		m.editor.Reset()
		cmds := []tea.Cmd{m.appendChat(chat.RoleUser, text)}
		m.setRunning(true)
		if m.controller != nil {
			m.controller.Start(text)
		} else {
			cmds = append(cmds, m.appendChat(chat.RoleSystem, "no agent controller (UI-only mode)"))
			m.setRunning(false)
		}
		if m.spinner.Visible() {
			cmds = append(cmds, m.spinner.Tick())
		}
		return m, tea.Batch(cmds...)
	}

	return m, m.forward(msg)
}

func (m *Model) handleBus(msg events.BusMsg) tea.Cmd {
	if msg.Type == eventbus.EventModelResponse {
		if n := events.TokensFrom(msg.Data); n > 0 {
			m.tokens += n
			m.footer.SetMetrics(m.tokens, 0)
		}
	}
	role, text := events.FormatLine(msg)
	var cmd tea.Cmd
	switch role {
	case "tool":
		cmd = m.appendChat(chat.RoleTool, text)
	case "error":
		cmd = m.appendChat(chat.RoleError, text)
	case "system":
		cmd = m.appendChat(chat.RoleSystem, text)
	}
	if msg.Type == eventbus.EventSessionStart {
		m.setRunning(true)
		return tea.Batch(cmd, m.spinner.Tick())
	}
	if msg.Type == eventbus.EventSessionEnd || msg.Type == eventbus.EventAgentError {
		m.setRunning(false)
	}
	return cmd
}

func (m Model) forward(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	if m.spinner.Visible() {
		comp, cmd := m.spinner.Update(msg)
		m.spinner = comp.(*spinner.Model)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if m.focus == focusEditor {
		comp, cmd := m.editor.Update(msg)
		m.editor = comp.(*editor.Model)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		comp, cmd := m.chat.Update(msg)
		m.chat = comp.(*chat.Model)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if extra := m.extras.UpdateAll(msg); extra != nil {
		cmds = append(cmds, extra)
	}
	return tea.Batch(cmds...)
}

func (m *Model) toggleFocus() {
	if m.focus == focusEditor {
		m.focus = focusChat
		m.editor.Blur()
		m.chat.Focus()
		return
	}
	m.focus = focusEditor
	m.chat.Blur()
	m.editor.Focus()
}

func (m *Model) setRunning(v bool) {
	m.spinner.SetVisible(v)
	if v {
		m.header.SetStatus("running", true)
		m.spinner.SetMessage("agent working")
	} else {
		m.header.SetStatus("idle", false)
	}
}

func (m Model) isRunning() bool {
	if m.controller != nil {
		return m.controller.Running()
	}
	return m.spinner.Visible()
}

func (m Model) listen() tea.Cmd {
	if m.events == nil {
		return nil
	}
	return events.Listen(m.events.Chan())
}

func (m *Model) layout() []chat.HeavyJob {
	headerH := 3
	footerH := 3
	spinnerH := 0
	if m.spinner.Visible() {
		spinnerH = 1
	}
	avail := m.height - headerH - footerH - spinnerH
	if avail < 7 {
		avail = 7
	}
	editorH := 6 + m.editor.PopupHeight()
	chatH := avail - editorH
	if chatH < 3 {
		chatH = 3
		editorH = avail - chatH
	}
	if editorH < 4 {
		editorH = 4
		chatH = avail - editorH
		if chatH < 3 {
			chatH = 3
		}
	}
	jobs := m.chat.SetSize(m.width, chatH)
	m.editor.SetWidth(m.width)
	return jobs
}

func (m *Model) layoutCmd() tea.Cmd {
	return m.submitJobs(m.layout())
}

func (m *Model) appendChat(role chat.Role, text string) tea.Cmd {
	return m.submitJobs(m.chat.Append(role, text))
}

func (m *Model) submitJobs(jobs []chat.HeavyJob) tea.Cmd {
	if len(jobs) == 0 || m.async == nil {
		return nil
	}
	var cmds []tea.Cmd
	for _, j := range jobs {
		cmds = append(cmds, m.async.Submit(RenderJob{
			ID: j.ID, Width: j.Width, Text: j.Text, Theme: m.theme,
		}))
	}
	return tea.Batch(cmds...)
}
