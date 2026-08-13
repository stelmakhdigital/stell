package hooks

// DefaultCompactRatio is the context-window fill ratio that triggers compaction.
const DefaultCompactRatio = 0.8

// Compactor decides when the loop should compact messages.
type Compactor struct {
	Window int
	Ratio  float64
	Fired  int
}

// NewCompactor returns defaults (32k window, 80%).
func NewCompactor(window int, ratio float64) *Compactor {
	if window <= 0 {
		window = 32_000
	}
	if ratio <= 0 || ratio > 1 {
		ratio = DefaultCompactRatio
	}
	return &Compactor{Window: window, Ratio: ratio}
}

// ShouldCompact reports whether tokens fill enough of the window.
func (c *Compactor) ShouldCompact(tokens int) bool {
	if c == nil || c.Window <= 0 {
		return false
	}
	return float64(tokens) >= float64(c.Window)*c.Ratio
}

// RecordFire increments the compact counter (tests / metrics).
func (c *Compactor) RecordFire() {
	if c != nil {
		c.Fired++
	}
}
