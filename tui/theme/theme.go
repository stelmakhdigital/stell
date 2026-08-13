package theme

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"gopkg.in/yaml.v3"
)

// Colors is the TUI palette loaded from YAML.
type Colors struct {
	Primary    string `yaml:"primary"`
	Secondary  string `yaml:"secondary"`
	Accent     string `yaml:"accent"`
	Background string `yaml:"background"`
	Foreground string `yaml:"foreground"`
	Muted      string `yaml:"muted"`
	Warning    string `yaml:"warning"`
	Error      string `yaml:"error"`
	Border     string `yaml:"border"`
}

// Theme holds visual settings.
type Theme struct {
	Name    string `yaml:"name"`
	Tagline string `yaml:"tagline"`
	Colors  Colors `yaml:"colors"`
}

// Default returns the built-in Catppuccin-inspired theme.
func Default() Theme {
	return Theme{
		Name:    "default",
		Tagline: "Let's build something great",
		Colors: Colors{
			Primary:    "#89B4FA",
			Secondary:  "#A6E3A1",
			Accent:     "#F5C2E7",
			Background: "#1E1E2E",
			Foreground: "#CDD6F4",
			Muted:      "#6C7086",
			Warning:    "#F9E2AF",
			Error:      "#F38BA8",
			Border:     "#45475A",
		},
	}
}

// Load reads a theme YAML file. Missing file returns Default.
func Load(path string) (Theme, error) {
	t := Default()
	if path == "" {
		return t, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return t, nil
		}
		return t, err
	}
	if err := yaml.Unmarshal(data, &t); err != nil {
		return t, fmt.Errorf("parse theme: %w", err)
	}
	if t.Tagline == "" {
		t.Tagline = Default().Tagline
	}
	return t, nil
}

// Parse unmarshals theme YAML from bytes (tests).
func Parse(data []byte) (Theme, error) {
	t := Default()
	if err := yaml.Unmarshal(data, &t); err != nil {
		return t, err
	}
	return t, nil
}

// C converts a hex/ANSI color string to a terminal color.
func C(s string) color.Color {
	return lipgloss.Color(s)
}

func (t Theme) Accent() color.Color     { return C(t.Colors.Accent) }
func (t Theme) Primary() color.Color    { return C(t.Colors.Primary) }
func (t Theme) Secondary() color.Color  { return C(t.Colors.Secondary) }
func (t Theme) Foreground() color.Color { return C(t.Colors.Foreground) }
func (t Theme) Muted() color.Color      { return C(t.Colors.Muted) }
func (t Theme) Warning() color.Color    { return C(t.Colors.Warning) }
func (t Theme) Error() color.Color      { return C(t.Colors.Error) }
func (t Theme) Border() color.Color     { return C(t.Colors.Border) }

// Dark reports a dark background (luminance < 50%).
func (t Theme) Dark() bool {
	s := strings.TrimPrefix(t.Colors.Background, "#")
	if len(s) < 6 {
		return true
	}
	var r, g, b int
	if _, err := fmt.Sscanf(s[:6], "%02x%02x%02x", &r, &g, &b); err != nil {
		return true
	}
	return (299*r+587*g+114*b)/1000 < 128
}
