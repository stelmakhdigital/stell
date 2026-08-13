# Coding Agent (Go)

Minimal MVP core: agent loop, Event Bus, tool stubs, LLM providers (Ollama/vLLM/OpenAI-compatible).

## Quickstart

```bash
go test ./...
go run ./cmd/agent run -config configs/agent.yaml "Explain what you can do"
```

Requires an LLM endpoint (Ollama at `http://127.0.0.1:11434/v1` by default).

## Observability

- JSON logs via zap (`logging.level` in config)
- Prometheus: `http://127.0.0.1:9090/metrics` (see `metrics.addr`)
- OpenTelemetry spans for turn / tool (in-process provider; no remote exporter in MVP)

## Eval

```bash
make eval-smoke          # deterministic scoring with fixed answer
make eval                # full agent+LLM (needs Ollama/vLLM)
go run ./cmd/eval --judge --limit 5   # optional LLM-as-Judge
```

## TUI

Docs: [`tui/README.md`](tui/README.md). Themes: [`docs/tui/themes.md`](docs/tui/themes.md). Components: [`docs/tui/components.md`](docs/tui/components.md).

```bash
go run ./tui                          # needs LLM (Ollama by default)
go run ./tui -ui-only                 # layout preview without agent
go run ./tui -theme tui/theme/config.yaml
```

Keys: `ctrl+s` send, `enter` newline, `tab` focus chat/editor, `esc` cancel, `ctrl+q` quit.
