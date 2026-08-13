package skills

import (
	"path/filepath"
	"strings"
)

// Skill is a lazy-loaded markdown knowledge pack.
type Skill struct {
	Name        string
	Version     string
	Description string
	Triggers    Triggers
	Path        string
	body        string
	bodyLoaded  bool
}

// Triggers decide when a skill is relevant.
type Triggers struct {
	Keywords []string `yaml:"keywords"`
	Files    []string `yaml:"files"`
}

// IndexEntry is the prompt-facing summary (no body).
type IndexEntry struct {
	Name        string
	Description string
}

// Body returns the markdown body. Empty until LoadBody.
func (s *Skill) Body() string {
	return s.body
}

func (s *Skill) matches(query string, files []string) bool {
	q := strings.ToLower(query)
	for _, kw := range s.Triggers.Keywords {
		if kw != "" && strings.Contains(q, strings.ToLower(kw)) {
			return true
		}
	}
	for _, pat := range s.Triggers.Files {
		for _, f := range files {
			ok, err := filepath.Match(pat, filepath.Base(f))
			if err == nil && ok {
				return true
			}
		}
	}
	return false
}
