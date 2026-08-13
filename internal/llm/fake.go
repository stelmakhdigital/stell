package llm

import (
	"context"
	"sync"
)

// FakeProvider is a scripted provider for tests.
type FakeProvider struct {
	mu        sync.Mutex
	responses []Response
	calls     int
	err       error
}

// NewFake creates a fake provider that returns responses in order.
func NewFake(responses ...Response) *FakeProvider {
	return &FakeProvider{responses: responses}
}

// WithError makes the next Generate call fail.
func (f *FakeProvider) WithError(err error) *FakeProvider {
	f.err = err
	return f
}

func (f *FakeProvider) Name() string { return "fake" }

// Calls returns how many Generate invocations happened.
func (f *FakeProvider) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

// Generate returns the next scripted response.
func (f *FakeProvider) Generate(ctx context.Context, _ Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil {
		return Response{}, f.err
	}
	if len(f.responses) == 0 {
		return Response{Message: Message{Role: RoleAssistant, Content: "done"}, FinishReason: "stop"}, nil
	}
	resp := f.responses[0]
	f.responses = f.responses[1:]
	return resp, nil
}
