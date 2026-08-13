package hooks

import (
	"fmt"
	"strconv"
	"time"

	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/pkg/observability"
	"go.uber.org/zap"
)

// RegisterObservabilityHooks wires logging + metrics to the event bus.
func RegisterObservabilityHooks(bus *eventbus.Bus, log *zap.Logger, metrics *observability.Metrics) {
	if log == nil {
		log = zap.NewNop()
	}

	bus.Subscribe(eventbus.EventSessionStart, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		log.Info("session_start",
			zap.String("session_id", e.SessionID),
			zap.Any("data", observability.RedactMap(e.Data)),
		)
		return nil, nil
	})
	bus.Subscribe(eventbus.EventSessionEnd, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		log.Info("session_end",
			zap.String("session_id", e.SessionID),
			zap.Any("data", observability.RedactMap(e.Data)),
		)
		return nil, nil
	})
	bus.Subscribe(eventbus.EventTurnStart, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		if metrics != nil {
			metrics.TurnsTotal.Inc()
			if d, ok := asFloat(e.Data["depth"]); ok {
				metrics.LoopDepth.Observe(d)
			}
		}
		log.Info("turn_start",
			zap.String("session_id", e.SessionID),
			zap.String("turn_id", e.TurnID),
			zap.Any("data", e.Data),
		)
		return nil, nil
	})
	bus.Subscribe(eventbus.EventGuardrailBlock, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		if metrics != nil && metrics.GuardrailBlocksTotal != nil {
			kind, _ := e.Data["kind"].(string)
			metrics.GuardrailBlocksTotal.WithLabelValues(kind).Inc()
		}
		log.Warn("guardrail_block",
			zap.String("session_id", e.SessionID),
			zap.Any("data", observability.RedactMap(e.Data)),
		)
		return nil, nil
	})
	bus.Subscribe(eventbus.EventHITLRequest, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		if metrics != nil && metrics.HitlRequestsTotal != nil {
			metrics.HitlRequestsTotal.Inc()
		}
		log.Info("hitl_request",
			zap.String("session_id", e.SessionID),
			zap.Any("data", observability.RedactMap(e.Data)),
		)
		return nil, nil
	})
	bus.Subscribe(eventbus.EventHITLDecision, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		if metrics != nil && metrics.HitlDecisionsTotal != nil {
			dec, _ := e.Data["decision"].(string)
			metrics.HitlDecisionsTotal.WithLabelValues(dec).Inc()
		}
		log.Info("hitl_decision",
			zap.String("session_id", e.SessionID),
			zap.Any("data", observability.RedactMap(e.Data)),
		)
		return nil, nil
	})
	bus.Subscribe(eventbus.EventToolResult, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		tool, _ := e.Data["tool"].(string)
		status := "ok"
		if errFlag, _ := e.Data["error"].(bool); errFlag {
			status = "error"
		}
		if metrics != nil {
			metrics.ToolCallsTotal.WithLabelValues(tool, status).Inc()
		}
		log.Info("tool_result",
			zap.String("session_id", e.SessionID),
			zap.String("turn_id", e.TurnID),
			zap.String("tool", tool),
			zap.String("status", status),
		)
		return nil, nil
	})
	bus.Subscribe(eventbus.EventModelResponse, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		if metrics != nil {
			if tokens, ok := asFloat(e.Data["tokens"]); ok && tokens > 0 {
				metrics.LLMTokensTotal.WithLabelValues("total").Add(tokens)
			}
			if ms, ok := asFloat(e.Data["latency_ms"]); ok {
				metrics.LLMLatency.Observe(ms / 1000.0)
			}
		}
		log.Info("model_response",
			zap.String("session_id", e.SessionID),
			zap.String("turn_id", e.TurnID),
			zap.Any("data", e.Data),
		)
		return nil, nil
	})
	bus.Subscribe(eventbus.EventAgentError, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		if metrics != nil {
			metrics.AgentErrorsTotal.Inc()
		}
		msg := fmt.Sprint(e.Data["error"])
		log.Warn("agent_error",
			zap.String("session_id", e.SessionID),
			zap.String("turn_id", e.TurnID),
			zap.String("error", observability.Redact(msg)),
		)
		return nil, nil
	})

	bus.Subscribe(eventbus.EventContextCompact, func(e *eventbus.Event) (*eventbus.EventResult, error) {
		if metrics != nil && metrics.CompactTokensBefore != nil {
			if b, ok := asFloat(e.Data["tokens_before"]); ok {
				metrics.CompactTokensBefore.Add(b)
			}
			if a, ok := asFloat(e.Data["tokens_after"]); ok && metrics.CompactTokensAfter != nil {
				metrics.CompactTokensAfter.Add(a)
			}
		}
		log.Info("context_compact",
			zap.String("session_id", e.SessionID),
			zap.Any("data", e.Data),
		)
		return nil, nil
	})
	_ = time.Second
}

func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case float64:
		return t, true
	case string:
		f, err := strconv.ParseFloat(t, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}
