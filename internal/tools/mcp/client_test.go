package mcp_test

import (
	"context"
	"testing"

	"github.com/budaev/agent/internal/tools"
	"github.com/budaev/agent/internal/tools/mcp"
)

func TestConnectRegisterAndDisconnect(t *testing.T) {
	reg := tools.NewRegistry()
	client := mcp.NewClient(reg)
	tr := &mcp.MemoryTransport{
		Tools: []mcp.ToolDef{{Name: "echo", Description: "echo"}},
		CallFn: func(name string, args map[string]any) (string, error) {
			return "pong", nil
		},
	}
	ctx := context.Background()
	if err := client.Connect(ctx, mcp.ServerConfig{Name: "demo", Version: "2.0"}, tr); err != nil {
		t.Fatal(err)
	}
	full := mcp.FormatName("demo", "echo", "2.0")
	if _, err := reg.Get(full); err != nil {
		t.Fatal(err)
	}
	out, err := reg.Execute(ctx, full, map[string]any{"x": 1})
	if err != nil {
		t.Fatal(err)
	}
	if out != "pong" {
		t.Fatalf("out=%q", out)
	}
	if err := client.Disconnect("demo"); err != nil {
		t.Fatal(err)
	}
	if _, err := reg.Get(full); err == nil {
		t.Fatal("expected tool removed after disconnect")
	}
}

func TestParseName(t *testing.T) {
	ns, tool, ver, err := mcp.ParseName("demo:echo@1.0")
	if err != nil || ns != "demo" || tool != "echo" || ver != "1.0" {
		t.Fatalf("%s %s %s %v", ns, tool, ver, err)
	}
}
