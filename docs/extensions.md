# Extensions

Adoption order: **hooks → skills → MCP → sub-agents**. `.so` plugins are not used.

## Hooks

Priority chain on the Event Bus. A hook may `Modify` event data or `Block` (the tool is not executed; the model receives a refusal).

Priorities: HITL (10) → guardrail (20) → logging (100). See `docs/hooks.md`.

## Skills

Catalog `skills/**/SKILL.md` with YAML frontmatter. Only the `name: description` index is injected into the system prompt. The body is read on `LoadSkill` or when a trigger matches the task.

## MCP

stdio JSON-RPC client (Content-Length). Tool names: `namespace:tool@version`.

Config: `configs/mcp.yaml`. Disconnect removes tools from the registry.

## Sub-agents

Orchestrator: max **7** parallel child agents.

- `Spawn` — the 8th caller **waits** for a free slot (or context cancel).
- `TrySpawn` — the 8th caller gets `ErrLimit` immediately.

Explore/Plan: isolated Event Bus, only `read_file` / `grep` / `glob`.
