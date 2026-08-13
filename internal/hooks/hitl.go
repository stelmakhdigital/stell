package hooks

import (
	"context"
	"fmt"
	"time"

	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/internal/tools"
)

const (
	DecisionApprove = "approve"
	DecisionDeny    = "deny"
)

// Approver decides whether a HITL tool may run.
type Approver interface {
	Decide(ctx context.Context, tool string, args map[string]any) (decision string, err error)
}

// StaticApprover always returns the same decision.
type StaticApprover struct {
	Decision string
}

func (a StaticApprover) Decide(context.Context, string, map[string]any) (string, error) {
	if a.Decision == "" {
		return DecisionDeny, nil
	}
	return a.Decision, nil
}

// HITLHook blocks tools that require human approval.
type HITLHook struct {
	manifests *tools.ManifestStore
	approver  Approver
	timeout   time.Duration
	bus       *eventbus.Bus
}

// NewHITLHook creates a high-priority HITL hook. timeout defaults to 30s; timeout → deny.
func NewHITLHook(store *tools.ManifestStore, approver Approver, timeout time.Duration, bus *eventbus.Bus) *HITLHook {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if approver == nil {
		approver = StaticApprover{Decision: DecisionDeny}
	}
	return &HITLHook{manifests: store, approver: approver, timeout: timeout, bus: bus}
}

func (h *HITLHook) Name() string  { return "hitl" }
func (h *HITLHook) Priority() int { return PriorityHITL }

func (h *HITLHook) Handle(_ context.Context, event *eventbus.Event) (*eventbus.EventResult, error) {
	if event == nil || event.Type != eventbus.EventToolCall {
		return nil, nil
	}
	tool, _ := event.Data["tool"].(string)
	m, ok := h.manifests.Get(tool)
	if !ok || !m.RequiresHITL {
		return nil, nil
	}
	args, _ := event.Data["args"].(map[string]any)
	if h.bus != nil {
		h.bus.PublishSimple(eventbus.EventHITLRequest, event.SessionID, event.TurnID, map[string]any{
			"tool": tool,
			"args": args,
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()
	decision, err := h.approver.Decide(ctx, tool, args)
	if err != nil || ctx.Err() != nil {
		decision = DecisionDeny
		if err == nil {
			err = fmt.Errorf("hitl timeout")
		}
	}
	if h.bus != nil {
		h.bus.PublishSimple(eventbus.EventHITLDecision, event.SessionID, event.TurnID, map[string]any{
			"tool":     tool,
			"decision": decision,
		})
	}
	if decision != DecisionApprove {
		msg := "hitl denied"
		if err != nil {
			msg = err.Error()
		}
		return &eventbus.EventResult{Block: true, Error: fmt.Errorf("%s", msg)}, nil
	}
	return nil, nil
}
