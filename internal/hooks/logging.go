package hooks

import (
	"context"

	"github.com/budaev/agent/internal/eventbus"
	"github.com/budaev/agent/pkg/observability"
	"go.uber.org/zap"
)

// LoggingHook records events; it never blocks.
type LoggingHook struct {
	log     *zap.Logger
	metrics *observability.Metrics
}

// NewLoggingHook creates a low-priority logging hook.
func NewLoggingHook(log *zap.Logger, metrics *observability.Metrics) *LoggingHook {
	if log == nil {
		log = zap.NewNop()
	}
	return &LoggingHook{log: log, metrics: metrics}
}

func (h *LoggingHook) Name() string  { return "logging" }
func (h *LoggingHook) Priority() int { return PriorityLogging }

func (h *LoggingHook) Handle(_ context.Context, e *eventbus.Event) (*eventbus.EventResult, error) {
	switch e.Type {
	case eventbus.EventToolCall:
		tool, _ := e.Data["tool"].(string)
		h.log.Info("tool_call",
			zap.String("session_id", e.SessionID),
			zap.String("tool", tool),
			zap.Any("args", observability.RedactMap(asMap(e.Data["args"]))),
		)
	case eventbus.EventToolResult:
		tool, _ := e.Data["tool"].(string)
		errFlag, _ := e.Data["error"].(bool)
		status := "ok"
		if errFlag {
			status = "error"
		}
		if h.metrics != nil {
			h.metrics.ToolCallsTotal.WithLabelValues(tool, status).Inc()
		}
	case eventbus.EventAgentError:
		if h.metrics != nil {
			h.metrics.AgentErrorsTotal.Inc()
		}
	case eventbus.EventGuardrailBlock:
		if h.metrics != nil && h.metrics.GuardrailBlocksTotal != nil {
			kind, _ := e.Data["kind"].(string)
			h.metrics.GuardrailBlocksTotal.WithLabelValues(kind).Inc()
		}
	}
	return nil, nil
}
