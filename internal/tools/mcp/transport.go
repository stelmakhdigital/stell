package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// ToolDef is a discovered MCP tool.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// Transport is a JSON-RPC 2.0 caller (stdio or in-memory).
type Transport interface {
	Call(ctx context.Context, method string, params any, result any) error
	Close() error
}

// MemoryTransport is an in-process MCP server for tests.
type MemoryTransport struct {
	mu     sync.Mutex
	Tools  []ToolDef
	CallFn func(name string, args map[string]any) (string, error)
	closed bool
}

// Call implements Transport.
func (t *MemoryTransport) Call(_ context.Context, method string, params any, result any) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return fmt.Errorf("mcp transport closed")
	}
	switch method {
	case "initialize":
		return jsonCopy(map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo":      map[string]any{"name": "memory", "version": "1.0"},
			"capabilities":    map[string]any{"tools": map[string]any{}},
		}, result)
	case "tools/list":
		return jsonCopy(map[string]any{"tools": t.Tools}, result)
	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := jsonCopy(params, &p); err != nil {
			return err
		}
		fn := t.CallFn
		if fn == nil {
			fn = func(string, map[string]any) (string, error) { return "ok", nil }
		}
		text, err := fn(p.Name, p.Arguments)
		if err != nil {
			return jsonCopy(map[string]any{
				"isError": true,
				"content": []map[string]any{{"type": "text", "text": err.Error()}},
			}, result)
		}
		return jsonCopy(map[string]any{
			"content": []map[string]any{{"type": "text", "text": text}},
		}, result)
	default:
		return fmt.Errorf("unknown mcp method %s", method)
	}
}

// Close marks the transport closed.
func (t *MemoryTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}

func jsonCopy(in, out any) error {
	if out == nil {
		return nil
	}
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// StdioTransport speaks MCP JSON-RPC with Content-Length framing.
type StdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	mu     sync.Mutex
	nextID atomic.Int64
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// DialStdio starts a stdio MCP server process.
func DialStdio(ctx context.Context, command string, args ...string) (*StdioTransport, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &StdioTransport{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}, nil
}

// Call sends a JSON-RPC request and waits for the matching response.
func (t *StdioTransport) Call(ctx context.Context, method string, params any, result any) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	id := t.nextID.Add(1)
	payload, err := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: id, Method: method, Params: params})
	if err != nil {
		return err
	}
	frame := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(payload), payload)
	if _, err := io.WriteString(t.stdin, frame); err != nil {
		return err
	}
	msg, err := readFrame(t.stdout)
	if err != nil {
		return err
	}
	var resp rpcResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("mcp rpc: %s", resp.Error.Message)
	}
	if result != nil && len(resp.Result) > 0 {
		return json.Unmarshal(resp.Result, result)
	}
	return ctx.Err()
}

// Close stops the process.
func (t *StdioTransport) Close() error {
	_ = t.stdin.Close()
	if t.cmd.Process != nil {
		_ = t.cmd.Process.Kill()
	}
	return t.cmd.Wait()
}

func readFrame(r *bufio.Reader) ([]byte, error) {
	headers := map[string]string{}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		k, v, ok := strings.Cut(line, ":")
		if ok {
			headers[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
		}
	}
	n, err := strconv.Atoi(headers["content-length"])
	if err != nil {
		return nil, fmt.Errorf("mcp frame: missing Content-Length")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf), nil
}
