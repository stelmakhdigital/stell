package renderer

import (
	"io"
	"os"
	"strings"
	"sync"
)

const envFullRedraw = "AGENT_TUI_FULL_REDRAW"

// Renderer keeps the previous frame and writes only dirty cells.
type Renderer struct {
	mu         sync.Mutex
	old        *Buffer
	width      int
	height     int
	fullRedraw bool
	BytesFull  int
	BytesPatch int
}

// New creates a renderer. AGENT_TUI_FULL_REDRAW=1 forces full frames.
func New(width, height int) *Renderer {
	r := &Renderer{width: width, height: height}
	r.fullRedraw = os.Getenv(envFullRedraw) == "1" || os.Getenv(envFullRedraw) == "true"
	return r
}

// SetFullRedraw toggles the fallback (tests).
func (r *Renderer) SetFullRedraw(v bool) { r.fullRedraw = v }

// Resize allocates a new buffer and forces the next frame to be a full refresh.
func (r *Renderer) Resize(width, height int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.width, r.height = width, height
	r.old = nil
}

// Observe diffs content against the previous frame and returns the bytes that
// would be written (patch or full). Updates stats. Does not touch stdout.
func (r *Renderer) Observe(width, height int, content string) (out string, patch bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if width != r.width || height != r.height {
		r.width, r.height = width, height
		r.old = nil
	}
	buf := NewBuffer(r.width, r.height)
	buf.FillFromLines(strings.Split(content, "\n"))
	full := FullANSI(buf)
	r.BytesFull += len(full)

	useFull := r.fullRedraw || r.old == nil
	if !useFull {
		d := CalculateDiff(r.old, buf)
		if d.Resize {
			useFull = true
		} else {
			out = RenderDiff(d)
			r.BytesPatch += len(out)
			r.old = buf
			return out, true
		}
	}
	r.BytesPatch += len(full)
	r.old = buf
	return full, false
}

// Write renders lines to w (tests / custom output).
func (r *Renderer) Write(w io.Writer, lines []string) error {
	content := strings.Join(lines, "\n")
	out, _ := r.Observe(r.width, r.height, content)
	_, err := io.WriteString(w, out)
	return err
}
