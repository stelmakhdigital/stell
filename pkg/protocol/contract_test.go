package protocol_test

import (
	"os"
	"strings"
	"testing"

	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/pkg/protocol"
	"gopkg.in/yaml.v3"
)

func TestFromBus(t *testing.T) {
	ev := protocol.FromBus(&eventbus.Event{Type: eventbus.EventToolCall, SessionID: "s", Data: map[string]any{"tool": "bash"}})
	if ev.Type != "tool_call" || ev.Data["tool"] != "bash" {
		t.Fatalf("%+v", ev)
	}
}

func TestOpenAPIContract(t *testing.T) {
	data, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	paths, _ := doc["paths"].(map[string]any)
	for _, p := range []string{"/v1/sessions", "/v1/sessions/{id}/events", "/v1/sessions/{id}/cancel"} {
		if _, ok := paths[p]; !ok {
			t.Fatalf("missing path %s", p)
		}
	}
	raw := string(data)
	if !strings.Contains(raw, "bearerAuth") {
		t.Fatal("expected bearer auth")
	}
}

func TestProtoExists(t *testing.T) {
	data, err := os.ReadFile("../../api/proto/agent.proto")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "service AgentAPI") {
		t.Fatal("missing AgentAPI")
	}
}
