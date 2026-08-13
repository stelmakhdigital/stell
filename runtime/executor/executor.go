package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/budaev/agent/runtime/sandbox"
)

const defaultMaxOutput = 64 * 1024

// Request is an execution request.
type Request struct {
	Tool      string
	Args      map[string]any
	Workspace string
	TimeoutMs int
}

// Result is an execution result.
type Result struct {
	Content   string
	Truncated bool
	Error     string
	Metadata  map[string]any
}

// Executor runs tools in the Hands process.
type Executor struct {
	Sandbox        *sandbox.Docker
	MaxOutput      int
	DefaultTimeout time.Duration
}

// New creates an executor with default sandbox policy.
func New(sb *sandbox.Docker) *Executor {
	return &Executor{
		Sandbox:        sb,
		MaxOutput:      defaultMaxOutput,
		DefaultTimeout: 30 * time.Second,
	}
}

// Execute dispatches a tool.
func (e *Executor) Execute(ctx context.Context, req Request) (Result, error) {
	if req.Workspace == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Result{}, err
		}
		req.Workspace = wd
	}
	root, err := filepath.Abs(req.Workspace)
	if err != nil {
		return Result{}, err
	}

	timeout := e.DefaultTimeout
	if req.TimeoutMs > 0 {
		timeout = time.Duration(req.TimeoutMs) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		content string
		meta    map[string]any
		execErr error
	)

	switch req.Tool {
	case "read_file":
		content, execErr = e.readFile(root, strArg(req.Args, "path"))
	case "write_file":
		content, execErr = e.writeFile(root, strArg(req.Args, "path"), strArg(req.Args, "content"))
	case "grep":
		content, meta, execErr = e.grep(root, strArg(req.Args, "pattern"), strArg(req.Args, "path"))
	case "glob":
		content, meta, execErr = e.glob(root, strArg(req.Args, "pattern"))
	case "bash":
		content, execErr = e.bash(ctx, root, strArg(req.Args, "command"))
	default:
		return Result{Error: fmt.Sprintf("unknown tool %q", req.Tool)}, nil
	}

	res := Result{Metadata: meta}
	if execErr != nil {
		res.Error = execErr.Error()
		res.Content = "error: " + execErr.Error()
		return res, nil
	}
	res.Content, res.Truncated = truncate(content, e.MaxOutput)
	return res, nil
}

func (e *Executor) readFile(root, rel string) (string, error) {
	path, err := resolvePath(root, rel)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (e *Executor) writeFile(root, rel, content string) (string, error) {
	path, err := resolvePath(root, rel)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), rel), nil
}

func (e *Executor) grep(root, pattern, sub string) (string, map[string]any, error) {
	if pattern == "" {
		return "", nil, fmt.Errorf("pattern is required")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", nil, fmt.Errorf("invalid pattern: %w", err)
	}
	searchRoot := root
	if sub != "" {
		searchRoot, err = resolvePath(root, sub)
		if err != nil {
			return "", nil, err
		}
	}

	const maxMatches = 100
	var lines []string
	matches := 0
	err = filepath.WalkDir(searchRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "bin" {
				return filepath.SkipDir
			}
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		for i, line := range strings.Split(string(data), "\n") {
			if re.MatchString(line) {
				lines = append(lines, fmt.Sprintf("%s:%d:%s", rel, i+1, line))
				matches++
				if matches >= maxMatches {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", nil, err
	}
	return strings.Join(lines, "\n"), map[string]any{"matches": matches}, nil
}

func (e *Executor) glob(root, pattern string) (string, map[string]any, error) {
	if pattern == "" {
		return "", nil, fmt.Errorf("pattern is required")
	}
	matches, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil {
		return "", nil, err
	}
	const max = 200
	var rels []string
	for i, m := range matches {
		if i >= max {
			break
		}
		rel, err := filepath.Rel(root, m)
		if err != nil {
			rel = m
		}
		rels = append(rels, rel)
	}
	return strings.Join(rels, "\n"), map[string]any{"count": len(rels)}, nil
}

func (e *Executor) bash(ctx context.Context, root, command string) (string, error) {
	if command == "" {
		return "", fmt.Errorf("command is required")
	}
	if e.Sandbox == nil {
		return "", fmt.Errorf("docker sandbox is not configured")
	}
	if !e.Sandbox.Available(ctx) {
		return "", fmt.Errorf("docker is not available")
	}
	return e.Sandbox.Run(ctx, root, command)
}

func resolvePath(root, rel string) (string, error) {
	if rel == "" {
		return "", fmt.Errorf("path is required")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	clean := filepath.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes workspace")
	}
	full := filepath.Join(root, clean)
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	sep := string(os.PathSeparator)
	if fullAbs != rootAbs && !strings.HasPrefix(fullAbs, rootAbs+sep) {
		return "", fmt.Errorf("path escapes workspace")
	}
	return fullAbs, nil
}

func strArg(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	v, ok := args[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func truncate(s string, max int) (string, bool) {
	if max <= 0 || len(s) <= max {
		return s, false
	}
	return s[:max] + "\n...[truncated]", true
}
