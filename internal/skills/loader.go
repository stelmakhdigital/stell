package skills

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type frontmatter struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Triggers    Triggers `yaml:"triggers"`
}

// ParseSKILL parses SKILL.md with optional YAML frontmatter.
func ParseSKILL(path string, data []byte) (*Skill, error) {
	s := &Skill{Path: path}
	rest := data
	if bytes.HasPrefix(rest, []byte("---")) {
		rest = rest[3:]
		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}
		idx := bytes.Index(rest, []byte("\n---"))
		if idx < 0 {
			return nil, fmt.Errorf("unclosed frontmatter in %s", path)
		}
		var fm frontmatter
		if err := yaml.Unmarshal(rest[:idx], &fm); err != nil {
			return nil, fmt.Errorf("parse frontmatter %s: %w", path, err)
		}
		s.Name = fm.Name
		s.Version = fm.Version
		s.Description = fm.Description
		s.Triggers = fm.Triggers
		rest = rest[idx+4:]
		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}
	}
	if s.Name == "" {
		s.Name = strings.TrimSuffix(baseName(path), ".md")
	}
	s.body = string(rest)
	s.bodyLoaded = true
	return s, nil
}

// LoadBody reads the skill file and fills Body. Index metadata is kept.
func LoadBody(s *Skill) error {
	if s == nil {
		return fmt.Errorf("nil skill")
	}
	if s.bodyLoaded {
		return nil
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return err
	}
	parsed, err := ParseSKILL(s.Path, data)
	if err != nil {
		return err
	}
	s.body = parsed.body
	s.bodyLoaded = true
	if s.Name == "" {
		s.Name = parsed.Name
	}
	if s.Description == "" {
		s.Description = parsed.Description
	}
	if s.Version == "" {
		s.Version = parsed.Version
	}
	if len(s.Triggers.Keywords) == 0 && len(s.Triggers.Files) == 0 {
		s.Triggers = parsed.Triggers
	}
	return nil
}

func baseName(path string) string {
	i := strings.LastIndexAny(path, "/\\")
	if i < 0 {
		return path
	}
	return path[i+1:]
}
