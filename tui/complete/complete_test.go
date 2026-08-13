package complete_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/budaev/stell/tui/complete"
)

func TestCommandProvider(t *testing.T) {
	cmds := complete.DefaultCommands()
	items := complete.Commands(cmds, "/pl")
	if len(items) != 1 || items[0].Value != "/plan" {
		t.Fatalf("%+v", items)
	}
	all := complete.Commands(cmds, "/")
	if len(all) < 3 {
		t.Fatalf("expected several commands, got %d", len(all))
	}
}

func TestFileProviderSkipsGitAndLimits(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("ref"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "internal"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "internal", "loop.go"), []byte("package x"), 0o644)
	idx := complete.NewFileIndex(root)
	items := idx.Files("internal")
	if len(items) != 1 || !strings.Contains(items[0].Value, "loop.go") {
		t.Fatalf("%+v", items)
	}
	for _, it := range idx.Files("") {
		if strings.Contains(it.Value, ".git") {
			t.Fatalf("indexed .git: %s", it.Value)
		}
	}
}

func TestFileProviderPrefersPrefix(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "zzz.go"), []byte("package z"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "pkg"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "pkg", "go.mod"), []byte("module x"), 0o644)
	idx := complete.NewFileIndex(root)
	items := idx.Files("go")
	if len(items) == 0 {
		t.Fatal("expected hits")
	}
	if !strings.HasSuffix(items[0].Value, "go.mod") {
		t.Fatalf("prefix rank first, got %+v", items)
	}
}

func TestLoadCommandsFallback(t *testing.T) {
	cmds := complete.LoadCommands("/no/such.yaml")
	if len(cmds) == 0 {
		t.Fatal("expected defaults")
	}
}
