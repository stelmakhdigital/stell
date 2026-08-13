package sessionstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/budaev/stell/internal/domain"
	"github.com/budaev/stell/internal/sessionstore"
)

func TestTwoStoresShareDir(t *testing.T) {
	dir := t.TempDir()
	a, err := sessionstore.NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	b, err := sessionstore.NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	sess := domain.NewSession("s1", "agent-1", "hello")
	sess.Turns = []domain.Turn{{ID: "t1", CreatedAt: time.Now().UTC(), ModelOutput: "hi"}}
	if err := a.Save(context.Background(), sess); err != nil {
		t.Fatal(err)
	}
	got, err := b.Get(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Task != "hello" || len(got.Turns) != 1 {
		t.Fatalf("got %+v", got)
	}
}
