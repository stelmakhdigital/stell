package tools_test

import (
	"testing"

	"github.com/budaev/stell/internal/tools"
	"github.com/budaev/stell/internal/tools/builtin"
)

func TestAllowlistHidesTools(t *testing.T) {
	reg := tools.NewRegistry()
	builtin.RegisterStubs(reg)
	reg.SetAllowlist([]string{"read_file"})
	names := map[string]bool{}
	for _, n := range reg.Names() {
		names[n] = true
	}
	if !names["read_file"] {
		t.Fatal("read_file should be visible")
	}
	if names["bash"] {
		t.Fatal("bash should be hidden from the model")
	}
	if _, err := reg.Get("bash"); err == nil {
		t.Fatal("expected bash not found")
	}
}
