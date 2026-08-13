package app

import (
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/tui/render"
	"github.com/budaev/stell/tui/theme"
)

const (
	maxInflight = 2
	maxQueued   = 32
)

type jobKey struct {
	id, width int
}

// RenderJob is markdown work that must not run on the input loop.
type RenderJob struct {
	ID    int
	Width int
	Text  string
	Theme theme.Theme
}

// RenderDoneMsg is delivered when an async markdown job finishes.
type RenderDoneMsg struct {
	ID    int
	Width int
	Out   string
}

// Async queues heavy render jobs with a bounded inflight count.
type Async struct {
	mu       sync.Mutex
	inflight int
	queued   map[jobKey]RenderJob
}

func (j RenderJob) key() jobKey { return jobKey{j.ID, j.Width} }

func (j RenderJob) cmd() tea.Cmd {
	return func() tea.Msg {
		out := render.Markdown(j.Text, j.Width, j.Theme)
		return RenderDoneMsg{ID: j.ID, Width: j.Width, Out: out}
	}
}

// Submit starts the job or queues it if inflight is at the cap.
func (a *Async) Submit(job RenderJob) tea.Cmd {
	if a == nil {
		return job.cmd()
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.inflight >= maxInflight {
		if a.queued == nil {
			a.queued = map[jobKey]RenderJob{}
		}
		k := job.key()
		if _, ok := a.queued[k]; !ok && len(a.queued) >= maxQueued {
			for dk := range a.queued {
				delete(a.queued, dk)
				break
			}
		}
		a.queued[k] = job
		return nil
	}
	a.inflight++
	return job.cmd()
}

// Next starts the next queued job after a completion.
func (a *Async) Next() tea.Cmd {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.inflight > 0 {
		a.inflight--
	}
	for k, job := range a.queued {
		delete(a.queued, k)
		a.inflight++
		return job.cmd()
	}
	return nil
}

// Pending is the queued length (tests).
func (a *Async) Pending() int {
	if a == nil {
		return 0
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.queued)
}
