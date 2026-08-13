package builtin

import (
	"context"
	"encoding/json"

	"github.com/budaev/stell/internal/tools"
)

type stubTool struct {
	name        string
	description string
	schema      json.RawMessage
}

func (s *stubTool) Name() string            { return s.name }
func (s *stubTool) Description() string     { return s.description }
func (s *stubTool) Schema() json.RawMessage { return s.schema }
func (s *stubTool) Execute(context.Context, map[string]any) (string, error) {
	return "", tools.ErrNotImplemented
}

func objectSchema(props map[string]any, required []string) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{
		"type":                 "object",
		"properties":           props,
		"required":             required,
		"additionalProperties": false,
	})
	return raw
}

// RegisterStubs registers the five MVP tools as stubs (Brain-side schemas only).
func RegisterStubs(r *tools.Registry) {
	r.Register(&stubTool{
		name:        "read_file",
		description: "Read a file from the workspace by path.",
		schema: objectSchema(map[string]any{
			"path": map[string]any{"type": "string", "description": "Relative path within workspace"},
		}, []string{"path"}),
	})
	r.Register(&stubTool{
		name:        "write_file",
		description: "Write content to a file in the workspace.",
		schema: objectSchema(map[string]any{
			"path":    map[string]any{"type": "string"},
			"content": map[string]any{"type": "string"},
		}, []string{"path", "content"}),
	})
	r.Register(&stubTool{
		name:        "bash",
		description: "Execute a shell command in the Docker sandbox.",
		schema: objectSchema(map[string]any{
			"command": map[string]any{"type": "string"},
		}, []string{"command"}),
	})
	r.Register(&stubTool{
		name:        "grep",
		description: "Search file contents with a regex pattern.",
		schema: objectSchema(map[string]any{
			"pattern": map[string]any{"type": "string"},
			"path":    map[string]any{"type": "string", "description": "Optional subdirectory"},
		}, []string{"pattern"}),
	})
	r.Register(&stubTool{
		name:        "glob",
		description: "Find files by glob pattern.",
		schema: objectSchema(map[string]any{
			"pattern": map[string]any{"type": "string"},
		}, []string{"pattern"}),
	})
}
