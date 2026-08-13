package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/budaev/stell/tui/render"
)

func TestInlineImagesXtermPlaceholder(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "shot.png")
	png := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	}
	if err := os.WriteFile(p, png, 0o644); err != nil {
		t.Fatal(err)
	}
	getenv := func(k string) string {
		if k == "TERM" {
			return "xterm-256color"
		}
		return ""
	}
	out := render.InlineImagesEnv("see "+p, 40, getenv, dir)
	if !strings.Contains(out, "[image: shot.png]") {
		t.Fatalf("xterm placeholder: %q", out)
	}
	if strings.Contains(out, "\x1b_G") || strings.Contains(out, "1337") {
		t.Fatal("xterm must not emit graphics protocol")
	}
}

func TestDetectImageProtocolKitty(t *testing.T) {
	getenv := func(k string) string {
		if k == "KITTY_WINDOW_ID" {
			return "1"
		}
		return ""
	}
	if render.DetectImageProtocol(getenv) != render.ProtocolKitty {
		t.Fatal("expected kitty")
	}
}

func TestDetectImageProtocolDisabled(t *testing.T) {
	getenv := func(k string) string {
		if k == "STELL_TUI_IMAGES" {
			return "0"
		}
		if k == "KITTY_WINDOW_ID" {
			return "1"
		}
		return ""
	}
	if render.DetectImageProtocol(getenv) != render.ProtocolNone {
		t.Fatal("flag must disable")
	}
}

func TestInlineImagesMissingFile(t *testing.T) {
	getenv := func(string) string { return "" }
	out := render.InlineImagesEnv("![x](/no/such.png)", 20, getenv, t.TempDir())
	if !strings.Contains(out, "[image:") {
		t.Fatalf("%q", out)
	}
}

func TestInlineImagesRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "secret.png")
	_ = os.WriteFile(outside, []byte("not-an-image"), 0o644)
	t.Cleanup(func() { _ = os.Remove(outside) })
	getenv := func(k string) string {
		if k == "KITTY_WINDOW_ID" {
			return "1"
		}
		return ""
	}
	out := render.InlineImagesEnv("![x](../secret.png)", 20, getenv, root)
	if strings.Contains(out, "\x1b_G") {
		t.Fatal("must not encode files outside workspace")
	}
}
