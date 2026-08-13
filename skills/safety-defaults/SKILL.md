---
name: safety-defaults
version: "1.0"
description: Safety by Default — allowlist, sandbox, HITL for dangerous ops.
triggers:
  keywords: ["allowlist", "sandbox", "HITL", "security"]
  files: []
---

Deny unless allowlisted. bash only in Docker. Dangerous tool_call requires HITL.
Do not weaken the sandbox “to make the demo work”.
