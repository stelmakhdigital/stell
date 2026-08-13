package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RuntimeConfig configures the Hands process.
type RuntimeConfig struct {
	Addr           string        `yaml:"addr"`
	MaxOutputBytes int           `yaml:"max_output_bytes"`
	HMACKey        string        `yaml:"hmac_key"`
	Production     bool          `yaml:"production"`
	RequireHMAC    bool          `yaml:"require_hmac"`
	Sandbox        SandboxConfig `yaml:"sandbox"`
}

// SandboxConfig configures Docker sandbox.
type SandboxConfig struct {
	Image        string `yaml:"image"`
	Network      string `yaml:"network"`
	Memory       string `yaml:"memory"`
	CPUs         string `yaml:"cpus"`
	User         string `yaml:"user"`
	ReadOnlyRoot bool   `yaml:"read_only_root"`
	PidsLimit    string `yaml:"pids_limit"`
}

// DefaultRuntime returns Hands defaults.
func DefaultRuntime() RuntimeConfig {
	return RuntimeConfig{
		Addr:           "127.0.0.1:8081",
		MaxOutputBytes: 64 * 1024,
		Sandbox: SandboxConfig{
			Image:   "alpine:3.20",
			Network: "none",
			Memory:  "256m",
			CPUs:    "1",
		},
	}
}

// LoadRuntime loads runtime YAML.
func LoadRuntime(path string) (RuntimeConfig, error) {
	cfg := DefaultRuntime()
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
		return cfg, fmt.Errorf("parse runtime config: %w", err)
	}
	return cfg, nil
}
