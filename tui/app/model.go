package app

import (
	"github.com/budaev/agent/tui/commands"
	"github.com/budaev/agent/tui/components"
	"github.com/budaev/agent/tui/components/chat"
	"github.com/budaev/agent/tui/components/editor"
	"github.com/budaev/agent/tui/components/footer"
	"github.com/budaev/agent/tui/components/header"
	"github.com/budaev/agent/tui/components/spinner"
	"github.com/budaev/agent/tui/events"
	"github.com/budaev/agent/tui/renderer"
	"github.com/budaev/agent/tui/theme"
)

const (
	focusEditor = "editor"
	focusChat   = "chat"
)

// Model is the Bubble Tea root model.
type Model struct {
	theme      theme.Theme
	header     *header.Model
	chat       *chat.Model
	editor     *editor.Model
	footer     *footer.Model
	spinner    *spinner.Model
	controller commands.Controller
	events     *events.Client
	focus      string
	width      int
	height     int
	tokens     int
	quitting   bool
	diff       *renderer.Renderer
	async      *Async
	extras     *components.Registry
	keyRelease bool
}

// Config is used to construct the TUI.
type Config struct {
	Theme        theme.Theme
	Controller   commands.Controller
	Events       *events.Client
	ModelName    string
	Runtime      string
	Workspace    string
	CommandsPath string
	Registry     *components.Registry
}
