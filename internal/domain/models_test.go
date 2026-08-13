package domain_test

import (
	"context"
	"testing"

	"github.com/budaev/agent/internal/domain"
)

func TestMemorySessionRepository(t *testing.T) {
	repo := domain.NewMemorySessionRepository()
	s := domain.NewSession("s1", "a1", "task")
	s.AddTurn(domain.Turn{ID: "t1", SessionID: "s1", Depth: 1})

	if err := repo.Save(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Task != "task" || len(got.Turns) != 1 {
		t.Fatalf("unexpected session: %+v", got)
	}
	if _, err := repo.Get(context.Background(), "missing"); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
