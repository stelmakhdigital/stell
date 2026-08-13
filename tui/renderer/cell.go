package renderer

import "github.com/mattn/go-runewidth"

// Cell is one terminal cell.
type Cell struct {
	Rune  rune
	Style string // comparable style key (fg hex or empty)
}

// Equal reports whether two cells paint the same.
func (c Cell) Equal(o Cell) bool {
	cr, or := c.Rune, o.Rune
	if cr == 0 {
		cr = ' '
	}
	if or == 0 {
		or = ' '
	}
	return cr == or && c.Style == o.Style
}

// Buffer is a width×height cell grid.
type Buffer struct {
	cells  [][]Cell
	width  int
	height int
}

// NewBuffer allocates an empty buffer filled with spaces.
func NewBuffer(width, height int) *Buffer {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	cells := make([][]Cell, height)
	for y := range cells {
		row := make([]Cell, width)
		for x := range row {
			row[x] = Cell{Rune: ' '}
		}
		cells[y] = row
	}
	return &Buffer{cells: cells, width: width, height: height}
}

func (b *Buffer) Width() int  { return b.width }
func (b *Buffer) Height() int { return b.height }

// Set writes a cell if in bounds.
func (b *Buffer) Set(x, y int, c Cell) {
	if b == nil || y < 0 || y >= b.height || x < 0 || x >= b.width {
		return
	}
	if c.Rune == 0 {
		c.Rune = ' '
	}
	b.cells[y][x] = c
}

// Get returns a cell or a space.
func (b *Buffer) Get(x, y int) Cell {
	if b == nil || y < 0 || y >= b.height || x < 0 || x >= b.width {
		return Cell{Rune: ' '}
	}
	return b.cells[y][x]
}

// FillFromLines paints plain (ANSI-stripped) lines into the buffer.
func (b *Buffer) FillFromLines(lines []string) {
	for y := 0; y < b.height; y++ {
		line := ""
		if y < len(lines) {
			line = StripANSI(lines[y])
		}
		x := 0
		for _, r := range line {
			if x >= b.width {
				break
			}
			w := runewidth.RuneWidth(r)
			if w <= 0 {
				w = 1
			}
			b.Set(x, y, Cell{Rune: r})
			x++
			for i := 1; i < w && x < b.width; i++ {
				b.Set(x, y, Cell{Rune: ' '})
				x++
			}
		}
	}
}

// String returns the buffer as newline-joined rows (no ANSI).
func (b *Buffer) String() string {
	if b == nil {
		return ""
	}
	out := make([]byte, 0, b.height*(b.width+1))
	for y := 0; y < b.height; y++ {
		if y > 0 {
			out = append(out, '\n')
		}
		for x := 0; x < b.width; x++ {
			r := b.cells[y][x].Rune
			if r == 0 {
				r = ' '
			}
			out = append(out, string(r)...)
		}
	}
	return string(out)
}
