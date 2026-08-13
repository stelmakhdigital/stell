package audit_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/budaev/agent/pkg/audit"
)

func TestAppendRedactsSecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	s := audit.NewStore(path)
	args := map[string]any{"api_key": "super-secret-value", "path": "a.go"}
	if err := s.Append(audit.Record{
		Tool:         "read_file",
		Kind:         "tool_call",
		ArgsRedacted: audit.RedactArgs(args),
		ArgsHash:     audit.HashArgs(args),
	}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "super-secret-value") {
		t.Fatalf("secret leaked: %s", data)
	}
	var rec audit.Record
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatal(err)
	}
	if rec.ArgsHash == "" {
		t.Fatal("missing hash")
	}
	if rec.ArgsRedacted["api_key"] != "***" {
		t.Fatalf("api_key=%v", rec.ArgsRedacted["api_key"])
	}
}
