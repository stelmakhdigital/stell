package sandbox_test

import (
	"strings"
	"testing"

	"github.com/budaev/stell/runtime/sandbox"
)

func TestProductionPolicyHardened(t *testing.T) {
	p := sandbox.ProductionPolicy()
	if p.Network != "none" {
		t.Fatalf("network=%s", p.Network)
	}
	args := p.RunArgs("/tmp/ws", "echo hi")
	joined := strings.Join(args, " ")
	for _, want := range []string{"--network none", "--cap-drop ALL", "--user 65534:65534", "--read-only", "--pids-limit"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %s", want, joined)
		}
	}
}
