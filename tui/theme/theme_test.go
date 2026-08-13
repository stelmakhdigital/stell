package theme_test

import (
	"testing"

	"github.com/budaev/stell/tui/theme"
)

func TestParseTheme(t *testing.T) {
	raw := []byte(`
name: custom
tagline: Hello
colors:
  accent: "#FF00AA"
  primary: "#111111"
`)
	th, err := theme.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if th.Name != "custom" {
		t.Fatalf("name=%q", th.Name)
	}
	if th.Tagline != "Hello" {
		t.Fatalf("tagline=%q", th.Tagline)
	}
	if th.Colors.Accent != "#FF00AA" {
		t.Fatalf("accent=%q", th.Colors.Accent)
	}
}

func TestLoadMissingUsesDefault(t *testing.T) {
	th, err := theme.Load("does-not-exist.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if th.Name != "default" {
		t.Fatalf("expected default, got %q", th.Name)
	}
	if th.Colors.Accent == "" {
		t.Fatal("expected default accent")
	}
}
