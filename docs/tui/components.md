# TUI component API

TUI stays a thin Event Bus client. Extra UI is **compile-time**: register components, do not import `runtime/` or run tools.

## Interface

```go
type Component interface {
    Render(width int) []string
    Update(msg tea.Msg) (Component, tea.Cmd)
    Focus()
    Blur()
    WantsKeyRelease() bool
    Invalidate()
}
```

Defined in [`tui/components/component.go`](../../tui/components/component.go).

## Registry (no app core edits)

Slots: `below-chat`, `above-footer`.

```go
package main

import (
    "github.com/budaev/agent/tui/app"
    "github.com/budaev/agent/tui/components"
    "github.com/budaev/agent/tui/theme"
    tea "charm.land/bubbletea/v2"
)

type banner struct{}

func (banner) Render(int) []string { return []string{"custom banner"} }
func (banner) Update(tea.Msg) (components.Component, tea.Cmd) {
    return banner{}, nil
}
func (banner) Focus()                {}
func (banner) Blur()                 {}
func (banner) WantsKeyRelease() bool { return false }
func (banner) Invalidate()           {}

func main() {
    reg := components.NewRegistry()
    reg.Register("banner", components.SlotBelowChat, banner{})
    _ = app.Run(app.Config{
        Theme:    theme.Default(),
        Registry: reg,
    })
}
```

`Register` with the same name replaces the previous extra.

## Built-in components

| Package | Role |
|---|---|
| `tui/components/header` | Logo, tagline, status |
| `tui/components/chat` | Viewport, markdown, images |
| `tui/components/editor` | Prompt + autocomplete |
| `tui/components/footer` | cwd / git / runtime / tokens |
| `tui/components/spinner` | Run indicator |

## Kitty key-release

Request is best-effort (`tui/input`). xterm: `KeyboardEnhancementsMsg` has no event-types; releases are ignored. Editor reports `WantsKeyRelease() == true` so press handling stays on `KeyPressMsg` only.
