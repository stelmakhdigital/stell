package app

import "github.com/budaev/agent/tui/components/chat"

// EditorSet is a test helper.
func (m *Model) EditorSet(s string) {
	m.editor.SetValue(s)
}

// ChatMessages is a test helper.
func (m Model) ChatMessages() []chat.Message {
	return m.chat.Messages()
}

// DiffStats is a test helper.
func (m Model) DiffStats() (full, patch int) {
	if m.diff == nil {
		return 0, 0
	}
	return m.diff.BytesFull, m.diff.BytesPatch
}
