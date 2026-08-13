package skills

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Manager indexes skills without loading bodies until requested.
type Manager struct {
	mu     sync.RWMutex
	skills map[string]*Skill
}

// NewManager creates an empty manager.
func NewManager() *Manager {
	return &Manager{skills: make(map[string]*Skill)}
}

// LoadDir indexes all SKILL.md files under root. Bodies are not stored.
func LoadDir(root string) (*Manager, error) {
	m := NewManager()
	if root == "" {
		return m, nil
	}
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("skills root is not a directory: %s", root)
	}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.EqualFold(d.Name(), "SKILL.md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		parsed, err := ParseSKILL(path, data)
		if err != nil {
			return err
		}
		indexed := &Skill{
			Name:        parsed.Name,
			Version:     parsed.Version,
			Description: parsed.Description,
			Triggers:    parsed.Triggers,
			Path:        path,
		}
		m.skills[indexed.Name] = indexed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Index returns name+description entries (no bodies).
func (m *Manager) Index() []IndexEntry {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]IndexEntry, 0, len(m.skills))
	for _, s := range m.skills {
		out = append(out, IndexEntry{Name: s.Name, Description: s.Description})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get returns a skill by name without loading the body.
func (m *Manager) Get(name string) (*Skill, error) {
	if m == nil {
		return nil, fmt.Errorf("skill not found: %s", name)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.skills[name]
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", name)
	}
	return s, nil
}

// LoadSkill loads and returns the skill with body populated.
func (m *Manager) LoadSkill(name string) (*Skill, error) {
	s, err := m.Get(name)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := LoadBody(s); err != nil {
		return nil, err
	}
	return s, nil
}

// Match returns skills whose triggers hit query/files. Bodies stay unloaded.
func (m *Manager) Match(query string, files []string) []*Skill {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*Skill
	for _, s := range m.skills {
		if s.matches(query, files) {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// PromptIndex formats the lazy skill list for the system prompt.
func (m *Manager) PromptIndex() string {
	idx := m.Index()
	if len(idx) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Available skills (name: description). Load a skill only when relevant:\n")
	for _, e := range idx {
		fmt.Fprintf(&b, "- %s: %s\n", e.Name, e.Description)
	}
	return b.String()
}
