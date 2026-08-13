package llm_test

import (
	"context"
	"testing"

	"github.com/budaev/agent/internal/llm"
)

func TestCacheHitRateGrowsOnRepeatPrefix(t *testing.T) {
	inner := llm.NewFake(
		llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "a"}},
		llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "b"}},
		llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "c"}},
	)
	c := llm.NewCacheAware(inner, true)
	sys := []llm.Message{{Role: llm.RoleSystem, Content: "stable system"}, {Role: llm.RoleUser, Content: "q1"}}
	_, err := c.Generate(context.Background(), llm.Request{Messages: sys})
	if err != nil {
		t.Fatal(err)
	}
	if c.Hits.Load() != 0 || c.Misses.Load() != 1 {
		t.Fatalf("first: hits=%d misses=%d", c.Hits.Load(), c.Misses.Load())
	}
	sys[1].Content = "q2"
	_, err = c.Generate(context.Background(), llm.Request{Messages: sys})
	if err != nil {
		t.Fatal(err)
	}
	if c.Hits.Load() != 1 {
		t.Fatalf("expected hit, hits=%d rate=%v", c.Hits.Load(), c.HitRate())
	}
	if c.HitRate() <= 0 {
		t.Fatal("hit rate should grow")
	}
}
