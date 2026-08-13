package events

import (
	tea "charm.land/bubbletea/v2"
	"github.com/budaev/stell/internal/eventbus"
)

// Client bridges Event Bus → tea messages via a channel.
// Handlers only enqueue; the TUI loop consumes with Listen (no Program.Send).
type Client struct {
	bus *eventbus.Bus
	ch  chan tea.Msg
}

// NewClient subscribes to MVP agent events.
func NewClient(bus *eventbus.Bus) *Client {
	c := &Client{
		bus: bus,
		ch:  make(chan tea.Msg, 64),
	}
	if bus == nil {
		return c
	}
	for _, t := range []eventbus.EventType{
		eventbus.EventSessionStart,
		eventbus.EventSessionEnd,
		eventbus.EventTurnStart,
		eventbus.EventTurnEnd,
		eventbus.EventToolCall,
		eventbus.EventToolResult,
		eventbus.EventModelRequest,
		eventbus.EventModelResponse,
		eventbus.EventAgentError,
	} {
		t := t
		bus.Subscribe(t, func(e *eventbus.Event) (*eventbus.EventResult, error) {
			c.enqueue(BusMsg{
				Type:      e.Type,
				SessionID: e.SessionID,
				TurnID:    e.TurnID,
				Data:      e.Data,
			})
			return nil, nil
		})
	}
	return c
}

// Chan is the outbound message queue.
func (c *Client) Chan() <-chan tea.Msg { return c.ch }

// Send injects a tea message (controller done/errors).
func (c *Client) Send(msg tea.Msg) {
	c.enqueue(msg)
}

func (c *Client) enqueue(msg tea.Msg) {
	if c == nil || c.ch == nil {
		return
	}
	select {
	case c.ch <- msg:
	default:
		// drop if UI is too slow; avoid blocking Brain
	}
}
