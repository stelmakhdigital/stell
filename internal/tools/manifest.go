package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Manifest describes tool execution policy.
type Manifest struct {
	Name               string        `yaml:"name"`
	Timeout            time.Duration `yaml:"timeout"`
	MaxOutputBytes     int           `yaml:"max_output_bytes"`
	WorkspaceOnly      bool          `yaml:"workspace_only"`
	RequiresSandbox    bool          `yaml:"requires_sandbox"`
	RequiresHITL       bool          `yaml:"requires_hitl"`
	BlockedInModelLoop bool          `yaml:"blocked_in_model_loop"`
}

type manifestFile struct {
	Name               string `yaml:"name"`
	Timeout            string `yaml:"timeout"`
	MaxOutputBytes     int    `yaml:"max_output_bytes"`
	WorkspaceOnly      bool   `yaml:"workspace_only"`
	RequiresSandbox    bool   `yaml:"requires_sandbox"`
	RequiresHITL       bool   `yaml:"requires_hitl"`
	BlockedInModelLoop bool   `yaml:"blocked_in_model_loop"`
}

// ManifestStore looks up manifests. Production mode fails closed on miss.
type ManifestStore struct {
	mu         sync.RWMutex
	byName     map[string]Manifest
	production bool
}

// NewManifestStore wraps a map.
func NewManifestStore(m map[string]Manifest, production bool) *ManifestStore {
	if m == nil {
		m = map[string]Manifest{}
	}
	return &ManifestStore{byName: m, production: production}
}

// DefaultManifests returns MVP manifests for builtin tools.
func DefaultManifests() map[string]Manifest {
	return map[string]Manifest{
		"read_file": {
			Name:           "read_file",
			Timeout:        10 * time.Second,
			MaxOutputBytes: 64 * 1024,
			WorkspaceOnly:  true,
		},
		"write_file": {
			Name:           "write_file",
			Timeout:        10 * time.Second,
			MaxOutputBytes: 64 * 1024,
			WorkspaceOnly:  true,
		},
		"bash": {
			Name:            "bash",
			Timeout:         30 * time.Second,
			MaxOutputBytes:  64 * 1024,
			WorkspaceOnly:   true,
			RequiresSandbox: true,
		},
		"grep": {
			Name:           "grep",
			Timeout:        20 * time.Second,
			MaxOutputBytes: 64 * 1024,
			WorkspaceOnly:  true,
		},
		"glob": {
			Name:           "glob",
			Timeout:        10 * time.Second,
			MaxOutputBytes: 64 * 1024,
			WorkspaceOnly:  true,
		},
	}
}

// Get returns a manifest. In production a miss is ok=false (fail closed).
func (s *ManifestStore) Get(name string) (Manifest, bool) {
	if s == nil {
		return Manifest{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.byName[name]
	return m, ok
}

// Put adds or replaces a manifest.
func (s *ManifestStore) Put(m Manifest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byName == nil {
		s.byName = map[string]Manifest{}
	}
	s.byName[m.Name] = m
}

// Production reports fail-closed mode.
func (s *ManifestStore) Production() bool {
	if s == nil {
		return false
	}
	return s.production
}

// LoadManifestDir reads configs/tools/*.yaml and merges with defaults.
func LoadManifestDir(dir string, production bool) (*ManifestStore, error) {
	merged := DefaultManifests()
	if dir != "" {
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext != ".yaml" && ext != ".yml" {
				continue
			}
			m, err := loadManifestFile(filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, err
			}
			merged[m.Name] = m
		}
	}
	return NewManifestStore(merged, production), nil
}

func loadManifestFile(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var raw manifestFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Manifest{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if raw.Name == "" {
		return Manifest{}, fmt.Errorf("%s: name is required", path)
	}
	timeout := 10 * time.Second
	if raw.Timeout != "" {
		d, err := time.ParseDuration(raw.Timeout)
		if err != nil {
			return Manifest{}, fmt.Errorf("%s: timeout: %w", path, err)
		}
		timeout = d
	}
	return Manifest{
		Name:               raw.Name,
		Timeout:            timeout,
		MaxOutputBytes:     raw.MaxOutputBytes,
		WorkspaceOnly:      raw.WorkspaceOnly,
		RequiresSandbox:    raw.RequiresSandbox,
		RequiresHITL:       raw.RequiresHITL,
		BlockedInModelLoop: raw.BlockedInModelLoop,
	}, nil
}
