---
name: go-testing
version: "1.0"
description: Go test idioms — table-driven, t.Helper, no flakes.
triggers:
  keywords: ["go test", "testing", "table-driven"]
  files: ["*_test.go"]
---

Write table-driven tests. Use t.Helper() in asserts. Do not sleep in tests unless needed.
Compare errors with errors.Is. Test names: TestType_Method_Condition.
