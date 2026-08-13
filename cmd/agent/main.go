package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/budaev/agent/internal/agent"
	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/internal/hooks"
	"github.com/budaev/agent/internal/llm"
	"github.com/budaev/agent/internal/runtimeclient"
	"github.com/budaev/agent/internal/tools"
	"github.com/budaev/agent/internal/tools/builtin"
	"github.com/budaev/agent/pkg/config"
	"github.com/budaev/agent/pkg/observability"
	"github.com/budaev/agent/runtime/executor"
	"github.com/budaev/agent/runtime/sandbox"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "run":
		if err := runCmd(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  agent run [flags] "task"

Flags:
  -config string     path to agent.yaml (default configs/agent.yaml)
  -model string      override model name
  -provider string   override llm provider (ollama|vllm|openai)
  -runtime string    hands mode: local|http (default local)
  -runtime-url string hands base URL when -runtime=http
  -workspace string  workspace root for tools (default cwd)
  -metrics-addr string prometheus listen addr (default from config, empty to disable)
`)
}

func runCmd(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	cfgPath := fs.String("config", "configs/agent.yaml", "path to config")
	model := fs.String("model", "", "override model")
	providerName := fs.String("provider", "", "override provider")
	runtimeMode := fs.String("runtime", "local", "hands mode: local|http")
	runtimeURL := fs.String("runtime-url", "http://127.0.0.1:8081", "hands URL for http mode")
	workspace := fs.String("workspace", "", "workspace root")
	metricsAddr := fs.String("metrics-addr", "", "prometheus listen addr override")
	if err := fs.Parse(args); err != nil {
		return err
	}
	task := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if task == "" {
		return fmt.Errorf("task prompt is required")
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return err
	}
	if *model != "" {
		cfg.LLM.Model = *model
	}
	if *providerName != "" {
		cfg.LLM.Provider = *providerName
	}

	log, err := observability.NewLogger(cfg.Logging.Level)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	shutdownTrace, err := observability.InitTracing("agent")
	if err != nil {
		return err
	}
	defer func() { _ = shutdownTrace(context.Background()) }()

	metrics := observability.NewMetrics()
	addr := cfg.Metrics.Addr
	if *metricsAddr != "" {
		addr = *metricsAddr
	}
	if addr != "" {
		mux := http.NewServeMux()
		mux.Handle("/metrics", metrics.Handler())
		go func() {
			log.Info("metrics listening", zap.String("addr", addr))
			if err := http.ListenAndServe(addr, mux); err != nil {
				log.Warn("metrics server stopped", zap.Error(err))
			}
		}()
	}

	provider, err := newProvider(cfg)
	if err != nil {
		return err
	}

	ws := *workspace
	if ws == "" {
		ws, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	ws, err = filepath.Abs(ws)
	if err != nil {
		return err
	}

	client, err := newRuntimeClient(*runtimeMode, *runtimeURL)
	if err != nil {
		return err
	}

	bus := eventbus.New()
	hooks.RegisterObservabilityHooks(bus, log, metrics)

	reg := tools.NewRegistry()
	builtin.RegisterRuntimeTools(reg, client, ws)

	a := agent.New(
		agent.WithEventBus(bus),
		agent.WithRegistry(reg),
		agent.WithProvider(provider),
		agent.WithName(cfg.Agent.Name),
		agent.WithMaxLoopDepth(cfg.Agent.MaxLoopDepth),
		agent.WithModel(cfg.LLM.Model),
		agent.WithTemperature(cfg.Agent.Temperature),
		agent.WithMaxTokens(cfg.Agent.MaxTokens),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	result := a.Run(ctx, task)
	if result.FinalText != "" {
		fmt.Println(result.FinalText)
	}
	if result.Err != nil {
		return result.Err
	}
	fmt.Fprintf(os.Stderr, "session=%s turns=%d workspace=%s\n", result.SessionID, result.Turns, ws)
	return nil
}

func newRuntimeClient(mode, url string) (runtimeclient.Client, error) {
	switch strings.ToLower(mode) {
	case "local", "":
		sb := sandbox.NewDocker(sandbox.DefaultPolicy())
		return runtimeclient.NewLocal(executor.New(sb)), nil
	case "http":
		return runtimeclient.NewHTTP(url), nil
	default:
		return nil, fmt.Errorf("unknown runtime mode %q", mode)
	}
}

func newProvider(cfg config.Config) (llm.Provider, error) {
	apiKey := cfg.LLM.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("AGENT_LLM_API_KEY")
	}
	switch strings.ToLower(cfg.LLM.Provider) {
	case "ollama", "":
		return llm.NewOllama(cfg.LLM.BaseURL, apiKey), nil
	case "vllm":
		return llm.NewVLLM(cfg.LLM.BaseURL, apiKey), nil
	case "openai":
		base := cfg.LLM.BaseURL
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		return llm.NewOpenAICompat("openai", base, apiKey), nil
	default:
		return nil, fmt.Errorf("unknown llm provider %q", cfg.LLM.Provider)
	}
}
