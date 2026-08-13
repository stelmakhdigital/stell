package render_test

import (
	"strings"
	"testing"

	"github.com/budaev/stell/tui/render"
)

func TestCacheLRUAndByteBound(t *testing.T) {
	c := render.NewCache(2, 64)
	c.Put(1, 80, strings.Repeat("a", 20))
	c.Put(2, 80, strings.Repeat("b", 20))
	if c.Len() != 2 {
		t.Fatalf("len=%d", c.Len())
	}
	c.Put(3, 80, strings.Repeat("c", 20))
	if c.Len() != 2 {
		t.Fatalf("evict entries: len=%d", c.Len())
	}
	if _, ok := c.Get(1, 80); ok {
		t.Fatal("id 1 should be evicted")
	}
	c.Put(4, 80, strings.Repeat("d", 80))
	if c.Bytes() > 64 {
		t.Fatalf("bytes=%d over budget", c.Bytes())
	}
}

func TestCacheHitByIDAndWidth(t *testing.T) {
	c := render.NewCache(8, 1024)
	c.Put(1, 80, "hello")
	got, ok := c.Get(1, 80)
	if !ok || got != "hello" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
	if _, ok := c.Get(1, 40); ok {
		t.Fatal("different width must miss")
	}
}
