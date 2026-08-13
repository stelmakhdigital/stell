package eventbus_test

import (
	"testing"

	"github.com/budaev/agent/internal/eventbus"
)

func TestPublishCallsHandlersAndBlock(t *testing.T) {
	bus := eventbus.New()
	order := []string{}

	bus.Subscribe(eventbus.EventToolCall, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		order = append(order, "first")
		return &eventbus.EventResult{Block: true}, nil
	})
	bus.Subscribe(eventbus.EventToolCall, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		order = append(order, "second")
		return nil, nil
	})

	res, err := bus.Publish(&eventbus.Event{Type: eventbus.EventToolCall, Data: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !res.Block {
		t.Fatalf("expected block result")
	}
	if len(order) != 1 || order[0] != "first" {
		t.Fatalf("expected only first handler, got %v", order)
	}
}

func TestPublishModify(t *testing.T) {
	bus := eventbus.New()
	bus.Subscribe(eventbus.EventToolCall, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		return &eventbus.EventResult{
			Modified: true,
			Data:     map[string]any{"tool": "changed"},
		}, nil
	})

	ev := &eventbus.Event{Type: eventbus.EventToolCall, Data: map[string]any{"tool": "orig"}}
	_, err := bus.Publish(ev)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Data["tool"] != "changed" {
		t.Fatalf("data not modified: %v", ev.Data)
	}
}
