package app

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/tui/theme"
)

// Run starts the Bubble Tea program.
func Run(cfg Config) error {
	if cfg.Theme.Name == "" {
		cfg.Theme = theme.Default()
	}
	startPprof()
	p := tea.NewProgram(New(cfg))
	_, err := p.Run()
	return err
}

func startPprof() {
	v := os.Getenv("STELL_TUI_PPROF")
	if v != "1" && v != "true" {
		return
	}
	addr := os.Getenv("STELL_TUI_PPROF_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6060"
	}
	go func() {
		_ = http.ListenAndServe(addr, nil)
	}()
}
