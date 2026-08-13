package executor_test

import (
	"context"
	"strings"
	"testing"

	"github.com/budaev/stell/runtime/executor"
)

func TestBashWithoutSandboxRefused(t *testing.T) {
	e := &executor.Executor{}
	res, err := e.Execute(context.Background(), executor.Request{
		Tool:      "bash",
		Args:      map[string]any{"command": "echo hi"},
		Workspace: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error == "" || !strings.Contains(res.Error, "sandbox") {
		t.Fatalf("expected sandbox refuse, got %+v", res)
	}
}
