package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/tui/complete"
	"github.com/budaev/stell/tui/components"
	"github.com/budaev/stell/tui/components/chat"
	"github.com/budaev/stell/tui/components/editor"
	"github.com/budaev/stell/tui/components/footer"
	"github.com/budaev/stell/tui/components/header"
	"github.com/budaev/stell/tui/components/spinner"
	"github.com/budaev/stell/tui/events"
	"github.com/budaev/stell/tui/renderer"
)

// New creates the root TUI model.
func New(cfg Config) Model {
	modelName := cfg.ModelName
	if cfg.Controller != nil && modelName == "" {
		modelName = cfg.Controller.ModelName()
	}
	m := Model{
		theme:      cfg.Theme,
		header:     header.New(cfg.Theme, modelName),
		chat:       chat.New(cfg.Theme),
		editor:     editor.NewWithComplete(cfg.Theme, complete.LoadCommands(cfg.CommandsPath), complete.NewFileIndex(cfg.Workspace)),
		footer:     footer.New(cfg.Theme),
		spinner:    spinner.New(cfg.Theme),
		controller: cfg.Controller,
		events:     cfg.Events,
		focus:      focusEditor,
		width:      80,
		height:     24,
		diff:       renderer.New(80, 24),
		async:      &Async{},
		extras:     cfg.Registry,
	}
	if m.extras == nil {
		m.extras = components.NewRegistry()
	}
	m.footer.SetRuntime(cfg.Runtime)
	m.chat.SetWorkspace(cfg.Workspace)
	m.editor.Focus()
	m.chat.Blur()
	return m
}

// Init starts animation and event listening.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{header.TickCmd()}
	if m.events != nil {
		cmds = append(cmds, events.Listen(m.events.Chan()))
	}
	return tea.Batch(cmds...)
}
