package renderer

import (
	"strconv"
	"strings"
)

// Change is one dirty cell.
type Change struct {
	X, Y int
	Cell Cell
}

// Diff is the set of cells that differ between two buffers.
type Diff struct {
	Changes []Change
	Resize  bool
}

// CalculateDiff compares old vs new. Mismatched size is a full refresh (Resize).
func CalculateDiff(old, neu *Buffer) Diff {
	if old == nil || neu == nil || old.width != neu.width || old.height != neu.height {
		return Diff{Resize: true}
	}
	var ch []Change
	for y := 0; y < neu.height; y++ {
		rowDirty := false
		for x := 0; x < neu.width; x++ {
			if !old.Get(x, y).Equal(neu.Get(x, y)) {
				rowDirty = true
				break
			}
		}
		if !rowDirty {
			continue
		}
		for x := 0; x < neu.width; x++ {
			nc := neu.Get(x, y)
			if !old.Get(x, y).Equal(nc) {
				ch = append(ch, Change{X: x, Y: y, Cell: nc})
			}
		}
	}
	return Diff{Changes: ch}
}

// RenderDiff emits ANSI CUP + rune for each change. Empty if nothing dirty.
func RenderDiff(d Diff) string {
	if d.Resize || len(d.Changes) == 0 {
		return ""
	}
	var b strings.Builder
	for _, c := range d.Changes {
		b.WriteString("\x1b[")
		b.WriteString(strconv.Itoa(c.Y + 1))
		b.WriteByte(';')
		b.WriteString(strconv.Itoa(c.X + 1))
		b.WriteByte('H')
		r := c.Cell.Rune
		if r == 0 {
			r = ' '
		}
		b.WriteRune(r)
	}
	return b.String()
}

// FullANSI dumps the whole buffer from 1;1 with newlines (no CUP per cell).
func FullANSI(buf *Buffer) string {
	if buf == nil {
		return "\x1b[2J\x1b[H"
	}
	var b strings.Builder
	b.WriteString("\x1b[2J\x1b[H")
	b.WriteString(strings.TrimRight(buf.String(), " \n"))
	return b.String()
}
