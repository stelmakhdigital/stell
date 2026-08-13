package builtin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/budaev/agent/internal/runtimeclient"
	"github.com/budaev/agent/internal/tools"
)

// RuntimeTool is a Brain-side tool that delegates to Hands.
type RuntimeTool struct {
	name        string
	description string
	schema      json.RawMessage
	manifest    tools.Manifest
	client      runtimeclient.Client
	workspace   string
}

func (t *RuntimeTool) Name() string            { return t.name }
func (t *RuntimeTool) Description() string     { return t.description }
func (t *RuntimeTool) Schema() json.RawMessage { return t.schema }

func (t *RuntimeTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	if t.client == nil {
		return "", tools.ErrNotImplemented
	}
	timeoutMs := int(t.manifest.Timeout.Milliseconds())
	resp, err := t.client.Execute(ctx, runtimeclient.ExecuteRequest{
		Tool:      t.name,
		Args:      args,
		Workspace: t.workspace,
		TimeoutMs: timeoutMs,
	})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf("%s", resp.Error)
	}
	return resp.Content, nil
}

// RegisterRuntimeTools registers the five MVP tools backed by a RuntimeClient.
func RegisterRuntimeTools(r *tools.Registry, client runtimeclient.Client, workspace string) {
	manifests := tools.DefaultManifests()

	r.Register(&RuntimeTool{
		name: "read_file", description: "Read a file from the workspace by path.",
		schema: objectSchema(map[string]any{
			"path": map[string]any{"type": "string", "description": "Relative path within workspace"},
		}, []string{"path"}),
		manifest: manifests["read_file"], client: client, workspace: workspace,
	})
	r.Register(&RuntimeTool{
		name: "write_file", description: "Write content to a file in the workspace.",
		schema: objectSchema(map[string]any{
			"path":    map[string]any{"type": "string"},
			"content": map[string]any{"type": "string"},
		}, []string{"path", "content"}),
		manifest: manifests["write_file"], client: client, workspace: workspace,
	})
	r.Register(&RuntimeTool{
		name: "bash", description: "Execute a shell command in the Docker sandbox.",
		schema: objectSchema(map[string]any{
			"command": map[string]any{"type": "string"},
		}, []string{"command"}),
		manifest: manifests["bash"], client: client, workspace: workspace,
	})
	r.Register(&RuntimeTool{
		name: "grep", description: "Search file contents with a regex pattern.",
		schema: objectSchema(map[string]any{
			"pattern": map[string]any{"type": "string"},
			"path":    map[string]any{"type": "string", "description": "Optional subdirectory"},
		}, []string{"pattern"}),
		manifest: manifests["grep"], client: client, workspace: workspace,
	})
	r.Register(&RuntimeTool{
		name: "glob", description: "Find files by glob pattern.",
		schema: objectSchema(map[string]any{
			"pattern": map[string]any{"type": "string"},
		}, []string{"pattern"}),
		manifest: manifests["glob"], client: client, workspace: workspace,
	})
}
