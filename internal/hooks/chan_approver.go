package hooks

import "context"

// ChanApprover waits for an external decision (Web/API HITL).
type ChanApprover struct {
	C chan string
}

// NewChanApprover creates a buffered decision channel.
func NewChanApprover() *ChanApprover {
	return &ChanApprover{C: make(chan string, 1)}
}

func (a *ChanApprover) Decide(ctx context.Context, _ string, _ map[string]any) (string, error) {
	if a == nil || a.C == nil {
		return DecisionDeny, nil
	}
	select {
	case d := <-a.C:
		if d == "" {
			return DecisionDeny, nil
		}
		return d, nil
	case <-ctx.Done():
		return DecisionDeny, ctx.Err()
	}
}
