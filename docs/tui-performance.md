# TUI performance

Budgets (see `tui/render/cache.go`):

| Metric | Budget |
|---|---|
| Header/spinner animation | 30 FPS (`render.TargetFPS`) |
| Input latency (key → next `View`) | p99 ≤ 50 ms (`render.InputLatencyBudgetMS`) |
| Sync markdown on the Update loop | < 4 KiB (`render.AsyncMarkdownBytes`) |
| Render cache | 64 entries / 4 MiB |

## What is off the input loop

Assistant markdown larger than 4 KiB is queued (`tui/app/async.go`, max 2 inflight). The chat shows `rendering…` until `RenderDoneMsg`. Set `AGENT_TUI_ASYNC=0` to force sync (debug).

## Profiling

```bash
AGENT_TUI_PPROF=1 go run ./tui -ui-only
# other terminal
./scripts/tui_profile.sh /tmp/tui-cpu.pprof 10
```

Optional: `AGENT_TUI_PPROF_ADDR=127.0.0.1:6060` (localhost only).

Heap: `curl -s http://127.0.0.1:6060/debug/pprof/heap > /tmp/tui-heap.pprof`

Trace: `curl -s "http://127.0.0.1:6060/debug/pprof/trace?seconds=5" > /tmp/tui.trace`

## Cache

Keyed by **message id + width**. Width change re-renders; height-only does not. LRU eviction keeps memory bounded under a long session with streaming tool logs + large assistant replies.
