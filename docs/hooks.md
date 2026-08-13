# Hooks

`hooks.Hook` interface: `Name`, `Priority` (lower runs earlier), `Handle`.

`Registry` sorts hooks and `Attach` subscribes the chain to event types (usually `tool_call`).

`Chain` stops on the first `Block`. `Modify` replaces `event.Data` for later hooks and for the loop.

The loop calls `Bus.Publish` (not `PublishSimple`) on `tool_call` and **does not execute** the tool on `Block`.

Built-in hooks:

- `HITLHook` — `requires_hitl` in the manifest; timeout → deny
- `GuardrailHook` — dangerous bash
- `LoggingHook` — logs/metrics, never blocks
