package app_test

import (
	"strings"
	"testing"

	"github.com/budaev/stell/tui/app"
	"github.com/budaev/stell/tui/theme"
)

func TestAsyncQueuesBeyondInflight(t *testing.T) {
	var q app.Async
	th := theme.Default()
	var started int
	for i := 1; i <= 5; i++ {
		cmd := q.Submit(app.RenderJob{ID: i, Width: 80, Text: strings.Repeat("x", 16), Theme: th})
		if cmd != nil {
			started++
		}
	}
	if started != 2 {
		t.Fatalf("inflight started=%d want 2", started)
	}
	if q.Pending() != 3 {
		t.Fatalf("pending=%d", q.Pending())
	}
	cmd := q.Next()
	if cmd == nil {
		t.Fatal("expected next job")
	}
	if q.Pending() != 2 {
		t.Fatalf("pending after next=%d", q.Pending())
	}
}
