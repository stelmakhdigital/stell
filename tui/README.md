# TUI

Thin Event Bus client (Bubble Tea v2). No tool execution, no `runtime/` imports.

## Quickstart

```bash
go test ./tui/...
make tui                          # needs LLM (Ollama by default)
go run ./tui -ui-only             # layout preview
go run ./tui -theme tui/theme/config.yaml
go run ./tui -theme tui/theme/examples/latte.yaml
```

Flags: `-config`, `-theme`, `-model`, `-provider`, `-runtime`, `-runtime-url`, `-workspace`, `-ui-only`.

## Requirements

- Go 1.22+ (module is 1.26)
- Terminal with alt-screen (xterm, kitty, iTerm, WezTerm, ghostty, …)
- Optional: Kitty/iTerm for key-release + inline images
- Optional LLM for a live agent (`configs/stell.yaml`)

## Keybindings

| Key | Action |
|---|---|
| `ctrl+s` | Send prompt |
| `enter` | Newline in editor (or accept autocomplete) |
| `tab` | Autocomplete, or toggle chat/editor focus |
| `shift+tab` / `↑` `↓` | Autocomplete list |
| `esc` | Close autocomplete; else cancel running agent |
| `ctrl+c` | Cancel run, or quit if idle |
| `ctrl+q` | Quit |

Slash commands: `/plan`, `/code`, `/help`, `/clear`, `/compact` (`configs/tui_commands.yaml`). File complete: `@path` or `internal/`.

## Feature flags

| Env | Default | Effect |
|---|---|---|
| `STELL_TUI_ASYNC` | on | `0` — sync markdown |
| `STELL_TUI_FULL_REDRAW` | off | force full frames in the cell renderer |
| `STELL_TUI_KITTY` | on | `0` — do not request key-release |
| `STELL_TUI_IMAGES` | on | `0` — placeholders only |
| `STELL_TUI_PPROF` | off | `1` — `127.0.0.1:6060` |

## Docs

- [Themes](../docs/tui/themes.md)
- [Component API](../docs/tui/components.md)
- [Performance](../docs/tui-performance.md)
- Plans: [PLANS/TUI](../PLANS/TUI/README.md)
