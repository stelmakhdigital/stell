package executor_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/budaev/agent/runtime/executor"
	"github.com/budaev/agent/runtime/sandbox"
)

func TestReadWriteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	exec := executor.New(sandbox.NewDocker(sandbox.DefaultPolicy()))

	res, err := exec.Execute(context.Background(), executor.Request{
		Tool: "write_file", Workspace: dir,
		Args: map[string]any{"path": "a.txt", "content": "hello"},
	})
	if err != nil || res.Error != "" {
		t.Fatalf("write: err=%v res=%+v", err, res)
	}

	res, err = exec.Execute(context.Background(), executor.Request{
		Tool: "read_file", Workspace: dir,
		Args: map[string]any{"path": "a.txt"},
	})
	if err != nil || res.Error != "" {
		t.Fatalf("read: err=%v res=%+v", err, res)
	}
	if res.Content != "hello" {
		t.Fatalf("content=%q", res.Content)
	}
}

func TestPathEscape(t *testing.T) {
	dir := t.TempDir()
	exec := executor.New(nil)
	res, err := exec.Execute(context.Background(), executor.Request{
		Tool: "read_file", Workspace: dir,
		Args: map[string]any{"path": "../secret"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error == "" || !strings.Contains(res.Error, "escape") {
		t.Fatalf("expected escape error, got %+v", res)
	}
}

func TestGrepAndGlob(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "foo.go"), []byte("package foo\nfunc Hello() {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "bar.txt"), []byte("nope\n"), 0o644)

	exec := executor.New(nil)
	gres, err := exec.Execute(context.Background(), executor.Request{
		Tool: "grep", Workspace: dir,
		Args: map[string]any{"pattern": "Hello"},
	})
	if err != nil || gres.Error != "" {
		t.Fatalf("grep: %+v %v", gres, err)
	}
	if !strings.Contains(gres.Content, "foo.go") {
		t.Fatalf("grep content=%q", gres.Content)
	}

	globRes, err := exec.Execute(context.Background(), executor.Request{
		Tool: "glob", Workspace: dir,
		Args: map[string]any{"pattern": "*.go"},
	})
	if err != nil || globRes.Error != "" {
		t.Fatalf("glob: %+v %v", globRes, err)
	}
	if !strings.Contains(globRes.Content, "foo.go") {
		t.Fatalf("glob content=%q", globRes.Content)
	}
}

func TestBashDocker(t *testing.T) {
	sb := sandbox.NewDocker(sandbox.DefaultPolicy())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if !sb.Available(ctx) {
		t.Skip("docker not available")
	}

	dir := t.TempDir()
	exec := executor.New(sb)
	res, err := exec.Execute(context.Background(), executor.Request{
		Tool: "bash", Workspace: dir,
		Args:      map[string]any{"command": "echo hi-from-sandbox"},
		TimeoutMs: 20000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" {
		t.Fatalf("bash error: %s / %s", res.Error, res.Content)
	}
	if !strings.Contains(res.Content, "hi-from-sandbox") {
		t.Fatalf("content=%q", res.Content)
	}
}

func TestTruncate(t *testing.T) {
	dir := t.TempDir()
	big := strings.Repeat("x", 1000)
	_ = os.WriteFile(filepath.Join(dir, "big.txt"), []byte(big), 0o644)
	exec := executor.New(nil)
	exec.MaxOutput = 100
	res, err := exec.Execute(context.Background(), executor.Request{
		Tool: "read_file", Workspace: dir,
		Args: map[string]any{"path": "big.txt"},
	})
	if err != nil || res.Error != "" {
		t.Fatalf("%+v %v", res, err)
	}
	if !res.Truncated || !strings.Contains(res.Content, "truncated") {
		t.Fatalf("expected truncation: %+v", res)
	}
}
