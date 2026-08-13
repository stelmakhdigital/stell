package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the agent configuration.
type Config struct {
	Agent   AgentConfig   `yaml:"agent"`
	LLM     LLMConfig     `yaml:"llm"`
	Logging LoggingConfig `yaml:"logging"`
	Metrics MetricsConfig `yaml:"metrics"`
}

// AgentConfig holds loop settings.
type AgentConfig struct {
	Name              string   `yaml:"name"`
	MaxLoopDepth      int      `yaml:"max_loop_depth"`
	Temperature       float64  `yaml:"temperature"`
	MaxTokens         int      `yaml:"max_tokens"`
	Production        bool     `yaml:"production"`
	ToolAllowlist     []string `yaml:"tool_allowlist"`
	SkillsDir         string   `yaml:"skills_dir"`
	HMACKey           string   `yaml:"hmac_key"`
	AuditPath         string   `yaml:"audit_path"`
	ManifestsDir      string   `yaml:"manifests_dir"`
	MCPConfig         string   `yaml:"mcp_config"`
	TokenBudget       int      `yaml:"token_budget"`
	HITLTimeout       string   `yaml:"hitl_timeout"`
	ContextWindow     int      `yaml:"context_window"`
	CompactRatio      float64  `yaml:"compact_ratio"`
	CompressToolBytes int      `yaml:"compress_tool_bytes"`
	PromptCache       bool     `yaml:"prompt_cache"`
	SessionStore      string   `yaml:"session_store"`
	RuntimeURLs       []string `yaml:"runtime_urls"`
	RuntimeSticky     bool     `yaml:"runtime_sticky"`
	APIAddr           string   `yaml:"api_addr"`
	APIToken          string   `yaml:"api_token"`
}

// LLMConfig selects the provider.
type LLMConfig struct {
	Provider string `yaml:"provider"` // ollama | vllm | openai
	BaseURL  string `yaml:"base_url"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

// LoggingConfig configures logs (used fully in observability phase).
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// MetricsConfig configures metrics endpoint.
type MetricsConfig struct {
	Addr string `yaml:"addr"`
}

// Default returns sensible MVP defaults.
func Default() Config {
	return Config{
		Agent: AgentConfig{
			Name:              "coding-agent",
			MaxLoopDepth:      50,
			Temperature:       0.2,
			MaxTokens:         4096,
			SkillsDir:         "skills",
			ManifestsDir:      "configs/tools",
			MCPConfig:         "configs/mcp.yaml",
			AuditPath:         "eval/results/audit.jsonl",
			ContextWindow:     32000,
			CompactRatio:      0.8,
			CompressToolBytes: 8192,
			PromptCache:       true,
			APIAddr:           "127.0.0.1:8080",
		},
		LLM: LLMConfig{
			Provider: "ollama",
			BaseURL:  "http://127.0.0.1:11434/v1",
			Model:    "llama3",
		},
		Logging: LoggingConfig{Level: "info"},
		Metrics: MetricsConfig{Addr: ":9090"},
	}
}

// Load reads YAML config from path. Missing file returns Default.
func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Agent.MaxLoopDepth <= 0 {
		cfg.Agent.MaxLoopDepth = 50
	}
	return cfg, nil
}
