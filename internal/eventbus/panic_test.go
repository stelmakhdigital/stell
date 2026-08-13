package eventbus_test

import (
	"strings"
	"testing"

	"github.com/budaev/stell/internal/eventbus"
)

func TestHandlerPanicIsRecovered(t *testing.T) {
	bus := eventbus.New()
	bus.Subscribe(eventbus.EventAgentError, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		panic("boom")
	})
	_, err := bus.Publish(&eventbus.Event{Type: eventbus.EventAgentError, Data: map[string]any{}})
	if err == nil || !strings.Contains(err.Error(), "panic") {
		t.Fatalf("expected panic error, got %v", err)
	}
}
