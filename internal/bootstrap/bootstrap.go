package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/budaev/stell/internal/agent"
	"github.com/budaev/stell/internal/domain"
	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/internal/guardrails"
	"github.com/budaev/stell/internal/hooks"
	"github.com/budaev/stell/internal/llm"
	"github.com/budaev/stell/internal/runtimeclient"
	"github.com/budaev/stell/internal/sessionstore"
	"github.com/budaev/stell/internal/skills"
	"github.com/budaev/stell/internal/subagents"
	"github.com/budaev/stell/internal/tools"
	"github.com/budaev/stell/internal/tools/builtin"
	"github.com/budaev/stell/internal/tools/mcp"
	"github.com/budaev/stell/pkg/audit"
	"github.com/budaev/stell/pkg/config"
	"github.com/budaev/stell/runtime/executor"
	"github.com/budaev/stell/runtime/sandbox"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Options configures Brain + Hands wiring.
type Options struct {
	ConfigPath  string
	Model       string
	Provider    string
	RuntimeMode string
	RuntimeURL  string
	Workspace   string
	Logger      *zap.Logger
	Approver    hooks.Approver
}

// Runtime is a wired agent instance for CLI/TUI entrypoints.
type Runtime struct {
	Agent     *agent.Agent
	Bus       *eventbus.Bus
	Config    config.Config
	Workspace string
	Hooks     *hooks.Registry
	Skills    *skills.Manager
	Subagents *subagents.Orchestrator
	MCP       *mcp.Client
	Manifests *tools.ManifestStore
}

// New builds an agent with tools, event bus, and LLM provider.
func New(opts Options) (*Runtime, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, err
	}
	if opts.Model != "" {
		cfg.LLM.Model = opts.Model
	}
	if opts.Provider != "" {
		cfg.LLM.Provider = opts.Provider
	}

	log := opts.Logger
	if log == nil {
		log = zap.NewNop()
	}

	provider, err := newProvider(cfg)
	if err != nil {
		return nil, err
	}
	provider = llm.NewBulkhead(llm.NewCacheAware(provider, cfg.Agent.PromptCache), 8, 120*time.Second)

	ws := opts.Workspace
	if ws == "" {
		ws, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	ws, err = filepath.Abs(ws)
	if err != nil {
		return nil, err
	}

	mode := opts.RuntimeMode
	if mode == "" {
		mode = "local"
	}
	hmacKey := cfg.Agent.HMACKey
	if hmacKey == "" {
		hmacKey = os.Getenv("STELL_HMAC_KEY")
	}
	client, err := newRuntimeClient(mode, opts.RuntimeURL, hmacKey, cfg.Agent.Production, cfg.Agent.RuntimeURLs, cfg.Agent.RuntimeSticky)
	if err != nil {
		return nil, err
	}

	bus := eventbus.New()
	hooks.RegisterObservabilityHooks(bus, log, nil)

	reg := tools.NewRegistry()
	builtin.RegisterRuntimeTools(reg, client, ws)
	if len(cfg.Agent.ToolAllowlist) > 0 {
		reg.SetAllowlist(cfg.Agent.ToolAllowlist)
	}

	manifests, err := tools.LoadManifestDir(cfg.Agent.ManifestsDir, cfg.Agent.Production)
	if err != nil {
		return nil, err
	}

	skillMgr, err := skills.LoadDir(cfg.Agent.SkillsDir)
	if err != nil {
		return nil, err
	}

	hookReg := hooks.NewRegistry()
	hitlTimeout := 30 * time.Second
	if cfg.Agent.HITLTimeout != "" {
		if d, err := time.ParseDuration(cfg.Agent.HITLTimeout); err == nil {
			hitlTimeout = d
		}
	}
	approver := opts.Approver
	if approver == nil {
		approver = hooks.StaticApprover{Decision: hooks.DecisionDeny}
	}
	hookReg.Register(hooks.NewHITLHook(manifests, approver, hitlTimeout, bus))
	hookReg.Register(&hooks.GuardrailHook{})
	hookReg.Register(hooks.NewLoggingHook(log, nil))

	if cfg.Agent.AuditPath != "" {
		guardrails.AttachAudit(bus, audit.NewStore(cfg.Agent.AuditPath))
	}
	hookReg.Attach(bus, eventbus.EventToolCall)

	mcpClient := mcp.NewClient(reg)
	if err := connectMCP(cfg.Agent.MCPConfig, mcpClient); err != nil {
		log.Warn("mcp connect", zap.Error(err))
	}

	orch := subagents.NewOrchestrator(subagents.DefaultFactory(provider, reg, cfg.LLM.Model))

	var sessions domain.SessionRepository
	if cfg.Agent.SessionStore != "" {
		sessions, err = sessionstore.NewFileStore(cfg.Agent.SessionStore)
		if err != nil {
			return nil, err
		}
	}

	a := agent.New(
		agent.WithEventBus(bus),
		agent.WithRegistry(reg),
		agent.WithProvider(provider),
		agent.WithName(cfg.Agent.Name),
		agent.WithMaxLoopDepth(cfg.Agent.MaxLoopDepth),
		agent.WithModel(cfg.LLM.Model),
		agent.WithTemperature(cfg.Agent.Temperature),
		agent.WithMaxTokens(cfg.Agent.MaxTokens),
		agent.WithTokenBudget(cfg.Agent.TokenBudget),
		agent.WithSkills(skillMgr),
		agent.WithGuardrails(guardrails.New()),
		agent.WithManifests(manifests),
		agent.WithProduction(cfg.Agent.Production),
		agent.WithCompactor(hooks.NewCompactor(cfg.Agent.ContextWindow, cfg.Agent.CompactRatio)),
		agent.WithCompressToolBytes(cfg.Agent.CompressToolBytes),
	)
	if sessions != nil {
		agent.WithSessionRepo(sessions)(a)
	}

	return &Runtime{
		Agent:     a,
		Bus:       bus,
		Config:    cfg,
		Workspace: ws,
		Hooks:     hookReg,
		Skills:    skillMgr,
		Subagents: orch,
		MCP:       mcpClient,
		Manifests: manifests,
	}, nil
}

func connectMCP(path string, client *mcp.Client) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var cfg mcp.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	ctx := context.Background()
	for _, srv := range cfg.Servers {
		if srv.Command == "" {
			continue
		}
		tr, err := mcp.DialStdio(ctx, srv.Command, srv.Args...)
		if err != nil {
			return fmt.Errorf("mcp %s: %w", srv.Name, err)
		}
		if err := client.Connect(ctx, srv, tr); err != nil {
			_ = tr.Close()
			return err
		}
	}
	return nil
}

func newRuntimeClient(mode, url, hmacKey string, production bool, urls []string, sticky bool) (runtimeclient.Client, error) {
	switch strings.ToLower(mode) {
	case "local", "":
		policy := sandbox.DefaultPolicy()
		if production {
			policy = sandbox.ProductionPolicy()
		}
		if production && policy.Network != "none" {
			return nil, fmt.Errorf("production sandbox must use network=none")
		}
		return runtimeclient.NewLocal(executor.New(sandbox.NewDocker(policy))), nil
	case "http":
		if len(urls) == 0 {
			if url == "" {
				url = "http://127.0.0.1:8081"
			}
			if strings.Contains(url, ",") {
				urls = strings.Split(url, ",")
			} else {
				urls = []string{url}
			}
		}
		if len(urls) == 1 {
			return runtimeclient.NewHTTP(strings.TrimSpace(urls[0])).WithHMACKey(hmacKey), nil
		}
		return runtimeclient.NewFailover(urls, hmacKey, sticky)
	default:
		return nil, fmt.Errorf("unknown runtime mode %q", mode)
	}
}

func newProvider(cfg config.Config) (llm.Provider, error) {
	apiKey := cfg.LLM.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("STELL_LLM_API_KEY")
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

// Spawn builds an isolated session agent sharing provider/registry/manifests.
func (rt *Runtime) Spawn(approver hooks.Approver) (*agent.Agent, *eventbus.Bus) {
	if approver == nil {
		approver = hooks.StaticApprover{Decision: hooks.DecisionDeny}
	}
	bus := eventbus.New()
	hooks.RegisterObservabilityHooks(bus, zap.NewNop(), nil)
	hr := hooks.NewRegistry()
	timeout := 30 * time.Second
	if rt.Config.Agent.HITLTimeout != "" {
		if d, err := time.ParseDuration(rt.Config.Agent.HITLTimeout); err == nil {
			timeout = d
		}
	}
	hr.Register(hooks.NewHITLHook(rt.Manifests, approver, timeout, bus))
	hr.Register(&hooks.GuardrailHook{})
	hr.Attach(bus, eventbus.EventToolCall)
	src := rt.Agent
	a := agent.New(
		agent.WithEventBus(bus),
		agent.WithRegistry(src.Registry()),
		agent.WithProvider(src.Provider()),
		agent.WithName(rt.Config.Agent.Name),
		agent.WithMaxLoopDepth(rt.Config.Agent.MaxLoopDepth),
		agent.WithModel(src.ModelName()),
		agent.WithTemperature(rt.Config.Agent.Temperature),
		agent.WithMaxTokens(rt.Config.Agent.MaxTokens),
		agent.WithTokenBudget(rt.Config.Agent.TokenBudget),
		agent.WithSkills(rt.Skills),
		agent.WithGuardrails(src.Guardrails()),
		agent.WithManifests(src.Manifests()),
		agent.WithProduction(rt.Config.Agent.Production),
		agent.WithCompactor(hooks.NewCompactor(rt.Config.Agent.ContextWindow, rt.Config.Agent.CompactRatio)),
		agent.WithCompressToolBytes(rt.Config.Agent.CompressToolBytes),
		agent.WithSessionRepo(src.Sessions()),
	)
	return a, bus
}
