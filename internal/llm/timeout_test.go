package llm_test

import (
	"context"
	"testing"
	"time"

	"github.com/budaev/agent/internal/llm"
)

func TestGenerateDeadline(t *testing.T) {
	p := llm.NewFake(llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "x"}})
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)
	_, err := p.Generate(ctx, llm.Request{})
	if err == nil {
		t.Fatal("expected context error, not hang")
	}
}
