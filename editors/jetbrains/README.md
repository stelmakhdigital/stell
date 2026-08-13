# JetBrains plugin (MVP)

Shared protocol: HTTP `api/openapi.yaml` / `pkg/protocol`. Same as VS Code and Web.

Kotlin client (call from an action):

```kotlin
// editors/jetbrains/src/main/kotlin/com/stell/Client.kt
```

Wire `CreateSession` + poll `GetSession` or SSE. Auth: `Authorization: Bearer <token>` — not HMAC.

Open files via `VirtualFile` under the trusted project root only.
