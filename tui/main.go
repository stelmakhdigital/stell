package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/budaev/agent/internal/bootstrap"
	"github.com/budaev/agent/pkg/observability"
	"github.com/budaev/agent/tui/app"
	"github.com/budaev/agent/tui/commands"
	"github.com/budaev/agent/tui/events"
	"github.com/budaev/agent/tui/theme"
)

func main() {
	cfgPath := flag.String("config", "configs/agent.yaml", "agent config")
	themePath := flag.String("theme", "tui/theme/config.yaml", "TUI theme YAML")
	model := flag.String("model", "", "override LLM model")
	provider := flag.String("provider", "", "override LLM provider")
	runtimeMode := flag.String("runtime", "local", "hands mode: local|http")
	runtimeURL := flag.String("runtime-url", "http://127.0.0.1:8081", "hands URL")
	workspace := flag.String("workspace", "", "workspace root")
	uiOnly := flag.Bool("ui-only", false, "skip agent wiring (layout preview)")
	flag.Parse()

	th, err := theme.Load(*themePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "theme: %v\n", err)
		os.Exit(1)
	}

	cfg := app.Config{Theme: th, Runtime: *runtimeMode}
	if !*uiOnly {
		log, err := observability.NewFileLogger("info", "agent-tui.log")
		if err != nil {
			fmt.Fprintf(os.Stderr, "logger: %v\n", err)
			os.Exit(1)
		}
		rt, err := bootstrap.New(bootstrap.Options{
			ConfigPath:  *cfgPath,
			Model:       *model,
			Provider:    *provider,
			RuntimeMode: *runtimeMode,
			RuntimeURL:  *runtimeURL,
			Workspace:   *workspace,
			Logger:      log,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "bootstrap: %v\n", err)
			os.Exit(1)
		}
		client := events.NewClient(rt.Bus)
		cfg.Events = client
		cfg.Controller = commands.NewAgentController(rt.Agent, client)
		cfg.ModelName = rt.Config.LLM.Model
		if *workspace != "" {
			cfg.Workspace = *workspace
		} else {
			cfg.Workspace = rt.Workspace
		}
	}
	if cfg.Workspace == "" {
		cfg.Workspace, _ = os.Getwd()
	}
	cfg.CommandsPath = "configs/tui_commands.yaml"

	if err := app.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "tui: %v\n", err)
		os.Exit(1)
	}
}
