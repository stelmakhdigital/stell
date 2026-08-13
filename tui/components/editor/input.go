package editor

import tea "charm.land/bubbletea/v2"

// SubmitMsg is emitted by the app when the user sends the editor contents.
type SubmitMsg struct {
	Text string
}

// NewlineKey is enter (handled by textarea). Submit is ctrl+s at the app layer.
func IsSubmitKey(msg tea.KeyPressMsg) bool {
	s := msg.String()
	return s == "ctrl+s" || s == "ctrl+enter"
}
