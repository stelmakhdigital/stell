package eventbus

import (
	"fmt"
	"sync"
	"time"
)

// Bus is a synchronous in-process event bus.
type Bus struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler
}

// New creates an empty bus.
func New() *Bus {
	return &Bus{
		handlers: make(map[EventType][]Handler),
	}
}

// Subscribe registers a handler for an event type.
func (b *Bus) Subscribe(t EventType, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[t] = append(b.handlers[t], h)
}

// Publish delivers an event to all handlers of its type.
// If a handler sets Block, remaining handlers for this publish are skipped.
func (b *Bus) Publish(event *Event) (*EventResult, error) {
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}

	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[event.Type]...)
	b.mu.RUnlock()

	var last *EventResult
	for _, h := range handlers {
		res, err := safeHandle(h, event)
		if err != nil {
			return res, err
		}
		if res == nil {
			continue
		}
		if res.Modified && res.Data != nil {
			event.Data = res.Data
		}
		last = res
		if res.Block {
			return res, nil
		}
	}
	return last, nil
}

// PublishSimple publishes without caring about the result.
func (b *Bus) PublishSimple(t EventType, sessionID, turnID string, data map[string]any) {
	_, _ = b.Publish(&Event{
		Type:      t,
		SessionID: sessionID,
		TurnID:    turnID,
		Data:      data,
		At:        time.Now().UTC(),
	})
}

func safeHandle(h Handler, event *Event) (res *EventResult, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("event handler panic: %v", rec)
		}
	}()
	return h(event)
}
