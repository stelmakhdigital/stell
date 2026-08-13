package renderer_test

import (
	"strings"
	"testing"

	"github.com/budaev/stell/tui/renderer"
)

func TestDiffDirtyCellsOnly(t *testing.T) {
	a := renderer.NewBuffer(8, 2)
	a.FillFromLines([]string{"hello   ", "world   "})
	b := renderer.NewBuffer(8, 2)
	b.FillFromLines([]string{"hello!  ", "world   "})
	d := renderer.CalculateDiff(a, b)
	if d.Resize {
		t.Fatal("unexpected resize")
	}
	if len(d.Changes) != 1 {
		t.Fatalf("changes=%d %+v", len(d.Changes), d.Changes)
	}
	if d.Changes[0].X != 5 || d.Changes[0].Y != 0 || d.Changes[0].Cell.Rune != '!' {
		t.Fatalf("%+v", d.Changes[0])
	}
	patch := renderer.RenderDiff(d)
	if !strings.Contains(patch, "!") || !strings.Contains(patch, "\x1b[") {
		t.Fatalf("patch=%q", patch)
	}
}

func TestResizeForcesFull(t *testing.T) {
	a := renderer.NewBuffer(4, 1)
	b := renderer.NewBuffer(8, 2)
	d := renderer.CalculateDiff(a, b)
	if !d.Resize {
		t.Fatal("expected resize")
	}
	r := renderer.New(4, 1)
	_, patch := r.Observe(4, 1, "abcd")
	if patch {
		t.Fatal("first frame must be full")
	}
	out, patch := r.Observe(10, 5, "abcd")
	if patch {
		t.Fatal("resize must full-refresh")
	}
	if !strings.Contains(out, "\x1b[2J") {
		t.Fatalf("expected clear screen, got %q", out[:min(20, len(out))])
	}
}

func TestObservePatchSmallerThanFull(t *testing.T) {
	r := renderer.New(40, 8)
	frame1 := strings.Repeat("chat line one\n", 6) + "stable footer"
	r.Observe(40, 8, frame1)
	frame2 := strings.Repeat("chat line one\n", 5) + "chat line TWO\n" + "stable footer"
	patch, isPatch := r.Observe(40, 8, frame2)
	full := renderer.FullANSI(func() *renderer.Buffer {
		buf := renderer.NewBuffer(40, 8)
		buf.FillFromLines(strings.Split(frame2, "\n"))
		return buf
	}())
	if !isPatch {
		t.Fatal("expected patch on small chat update")
	}
	if len(patch) >= len(full) {
		t.Fatalf("patch %d should be < full %d", len(patch), len(full))
	}
}

func TestFullRedrawFlag(t *testing.T) {
	r := renderer.New(20, 4)
	r.SetFullRedraw(true)
	_, patch := r.Observe(20, 4, "aaaa")
	if patch {
		t.Fatal("first frame must be full")
	}
	out, patch := r.Observe(20, 4, "aaab")
	if patch {
		t.Fatal("STELL_TUI_FULL_REDRAW must skip patches")
	}
	if !strings.Contains(out, "\x1b[2J") {
		t.Fatalf("expected full clear, got %q", out[:min(24, len(out))])
	}
}

func BenchmarkChatUpdateBytes(b *testing.B) {
	base := strings.Repeat("agent thinking about the task\n", 10)
	r := renderer.New(80, 24)
	r.Observe(80, 24, base+"old")
	b.ReportAllocs()
	var n int
	for i := 0; i < b.N; i++ {
		out, patch := r.Observe(80, 24, base+"new")
		n = len(out)
		if !patch && i > 0 {
			b.Fatal("expected patch after first frame")
		}
		// restore previous so the next iteration still diffs a 3-char change
		r.Observe(80, 24, base+"old")
	}
	b.SetBytes(int64(n))
}
