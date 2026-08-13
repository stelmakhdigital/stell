package tools

import (
	"context"
	"fmt"
	"sync"
)

// Registry stores allowlisted tools.
type Registry struct {
	mu        sync.RWMutex
	tools     map[string]Tool
	allowlist map[string]struct{}
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool. Duplicate names overwrite.
func (r *Registry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[t.Name()] = t
}

// Unregister removes a tool by name.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
}

// SetAllowlist restricts Definitions/Names/Execute to the given names.
// Empty or nil allowlist means all registered tools are visible (dev mode).
func (r *Registry) SetAllowlist(names []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(names) == 0 {
		r.allowlist = nil
		return
	}
	r.allowlist = make(map[string]struct{}, len(names))
	for _, n := range names {
		if n != "" {
			r.allowlist[n] = struct{}{}
		}
	}
}

func (r *Registry) allowedLocked(name string) bool {
	if r.allowlist == nil {
		return true
	}
	_, ok := r.allowlist[name]
	return ok
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.allowedLocked(name) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return t, nil
}

// Lookup returns a registered tool even if it is not on the allowlist.
func (r *Registry) Lookup(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// Definitions returns LLM tool schemas for visible tools.
func (r *Registry) Definitions() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Definition, 0, len(r.tools))
	for _, t := range r.tools {
		if !r.allowedLocked(t.Name()) {
			continue
		}
		out = append(out, Definition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Schema(),
		})
	}
	return out
}

// Names returns registered tool names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for n := range r.tools {
		if !r.allowedLocked(n) {
			continue
		}
		names = append(names, n)
	}
	return names
}

// Execute runs a tool by name.
func (r *Registry) Execute(ctx context.Context, name string, args map[string]any) (string, error) {
	t, err := r.Get(name)
	if err != nil {
		return "", err
	}
	return t.Execute(ctx, args)
}
