package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/budaev/stell/internal/tools"
)

// ServerConfig describes a stdio MCP server.
type ServerConfig struct {
	Name    string   `yaml:"name"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Version string   `yaml:"version"`
}

// Config is configs/mcp.yaml.
type Config struct {
	Servers []ServerConfig `yaml:"servers"`
}

type conn struct {
	cfg   ServerConfig
	tr    Transport
	tools []string
}

// Client manages MCP connections and registry registration.
type Client struct {
	mu       sync.Mutex
	conns    map[string]*conn
	registry *tools.Registry
}

// NewClient binds to a tool registry.
func NewClient(reg *tools.Registry) *Client {
	return &Client{conns: make(map[string]*conn), registry: reg}
}

// Connect initializes a server over the given transport and registers tools.
func (c *Client) Connect(ctx context.Context, cfg ServerConfig, tr Transport) error {
	if cfg.Name == "" {
		return fmt.Errorf("mcp server name is required")
	}
	if cfg.Version == "" {
		cfg.Version = "1.0"
	}
	if err := tr.Call(ctx, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "agent", "version": "0.1"},
	}, nil); err != nil {
		return fmt.Errorf("mcp initialize %s: %w", cfg.Name, err)
	}
	var listed struct {
		Tools []ToolDef `json:"tools"`
	}
	if err := tr.Call(ctx, "tools/list", map[string]any{}, &listed); err != nil {
		return fmt.Errorf("mcp tools/list %s: %w", cfg.Name, err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if prev, ok := c.conns[cfg.Name]; ok {
		c.unregisterLocked(prev)
		_ = prev.tr.Close()
	}
	cn := &conn{cfg: cfg, tr: tr}
	for _, def := range listed.Tools {
		full := FormatName(cfg.Name, def.Name, cfg.Version)
		schema := def.InputSchema
		if len(schema) == 0 {
			schema = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		c.registry.Register(&mcpTool{
			name:        full,
			description: def.Description,
			schema:      schema,
			remoteName:  def.Name,
			tr:          tr,
		})
		cn.tools = append(cn.tools, full)
	}
	c.conns[cfg.Name] = cn
	return nil
}

// Disconnect closes a server and unregisters its tools.
func (c *Client) Disconnect(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cn, ok := c.conns[name]
	if !ok {
		return fmt.Errorf("mcp server %q is not connected", name)
	}
	c.unregisterLocked(cn)
	delete(c.conns, name)
	return cn.tr.Close()
}

func (c *Client) unregisterLocked(cn *conn) {
	if c.registry == nil {
		return
	}
	for _, n := range cn.tools {
		c.registry.Unregister(n)
	}
}

// Connected reports whether a named server is connected.
func (c *Client) Connected(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.conns[name]
	return ok
}

type mcpTool struct {
	name        string
	description string
	schema      json.RawMessage
	remoteName  string
	tr          Transport
}

func (t *mcpTool) Name() string            { return t.name }
func (t *mcpTool) Description() string     { return t.description }
func (t *mcpTool) Schema() json.RawMessage { return t.schema }

func (t *mcpTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	var out struct {
		IsError bool `json:"isError"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := t.tr.Call(ctx, "tools/call", map[string]any{
		"name":      t.remoteName,
		"arguments": args,
	}, &out); err != nil {
		return "", err
	}
	var b strings.Builder
	for _, part := range out.Content {
		if part.Text == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(part.Text)
	}
	text := b.String()
	if out.IsError {
		return "", fmt.Errorf("%s", text)
	}
	return text, nil
}
