package observability_test

import (
	"strings"
	"testing"

	"github.com/budaev/stell/pkg/observability"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRedact(t *testing.T) {
	in := `Authorization: Bearer SECRETTOKEN api_key=abc123 token: xyz`
	out := observability.Redact(in)
	if strings.Contains(out, "SECRETTOKEN") || strings.Contains(out, "abc123") || strings.Contains(out, "xyz") {
		t.Fatalf("not redacted: %s", out)
	}
	if !strings.Contains(out, "***") {
		t.Fatalf("expected masks: %s", out)
	}
}

func TestMetricsToolCalls(t *testing.T) {
	m := observability.NewMetrics()
	m.ToolCallsTotal.WithLabelValues("read_file", "ok").Inc()
	m.ToolCallsTotal.WithLabelValues("read_file", "ok").Inc()
	if got := testutil.ToFloat64(m.ToolCallsTotal.WithLabelValues("read_file", "ok")); got != 2 {
		t.Fatalf("got %v", got)
	}
}
