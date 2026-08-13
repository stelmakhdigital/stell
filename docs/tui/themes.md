# TUI themes

YAML palettes loaded with `-theme`. Missing file → built-in Catppuccin-inspired dark theme.

## Schema

```yaml
name: default
tagline: Let's build something great again
colors:
  primary: "#89B4FA"
  secondary: "#A6E3A1"
  accent: "#F5C2E7"
  background: "#1E1E2E"
  foreground: "#CDD6F4"
  muted: "#6C7086"
  warning: "#F9E2AF"
  error: "#F38BA8"
  border: "#45475A"
```

`background` luminance picks Chroma style (`catppuccin-mocha` vs `catppuccin-latte`).

## Examples

| File | Notes |
|---|---|
| [`tui/theme/config.yaml`](../../tui/theme/config.yaml) | Default dark |
| [`tui/theme/examples/latte.yaml`](../../tui/theme/examples/latte.yaml) | Light |
| [`tui/theme/examples/high-contrast.yaml`](../../tui/theme/examples/high-contrast.yaml) | Stronger contrast |

```bash
go run ./tui -ui-only -theme tui/theme/examples/latte.yaml
```

Do not hardcode hex in component logic — use `theme.Theme` helpers (`Primary()`, `Muted()`, …).
