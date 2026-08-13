package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/budaev/stell/internal/domain"
	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/internal/guardrails"
	"github.com/budaev/stell/internal/llm"
	"github.com/budaev/stell/pkg/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// RunResult is the outcome of an agent session.
type RunResult struct {
	SessionID domain.SessionID
	FinalText string
	Turns     int
	Err       error
}

// Run executes the agent loop for a user task.
func (a *Agent) Run(ctx context.Context, task string) RunResult {
	if a.provider == nil {
		return RunResult{Err: fmt.Errorf("llm provider is not configured")}
	}

	sessionID := domain.SessionID(uuid.NewString())
	session := domain.NewSession(sessionID, a.id, task)

	a.bus.PublishSimple(eventbus.EventSessionStart, string(sessionID), "", map[string]any{
		"task": task,
	})

	if a.guardrails != nil {
		if d := a.guardrails.CheckInput(task); d.Action == guardrails.Deny {
			a.bus.PublishSimple(eventbus.EventGuardrailBlock, string(sessionID), "", map[string]any{
				"kind":    "input",
				"message": d.Message,
			})
			a.bus.PublishSimple(eventbus.EventSessionEnd, string(sessionID), "", map[string]any{
				"turns": 0,
				"error": false,
			})
			return RunResult{SessionID: sessionID, FinalText: d.Message}
		}
	}

	messages := a.contextBuilder.Build(task, a.registry)
	toolSpecs := ToolSpecs(a.registry)

	var finalText string
	var runErr error
	usedTokens := 0

	for depth := 1; depth <= a.maxLoopDepth; depth++ {
		if err := ctx.Err(); err != nil {
			runErr = domain.ErrSessionCancelled
			a.bus.PublishSimple(eventbus.EventAgentError, string(sessionID), "", map[string]any{
				"error": err.Error(),
			})
			break
		}

		turnID := domain.TurnID(uuid.NewString())
		turnCtx, turnSpan := observability.Tracer("agent").Start(ctx, "turn")
		turnSpan.SetAttributes(
			attribute.String("session_id", string(sessionID)),
			attribute.String("turn_id", string(turnID)),
			attribute.Int("depth", depth),
			attribute.String("trace_id", observability.TraceIDFromContext(turnCtx)),
		)

		turn := domain.Turn{
			ID:        turnID,
			SessionID: sessionID,
			Depth:     depth,
			CreatedAt: time.Now().UTC(),
		}

		a.bus.PublishSimple(eventbus.EventTurnStart, string(sessionID), string(turnID), map[string]any{
			"depth":    depth,
			"trace_id": observability.TraceIDFromContext(turnCtx),
		})
		a.bus.PublishSimple(eventbus.EventModelRequest, string(sessionID), string(turnID), map[string]any{
			"model": a.model,
		})

		if a.compactor != nil {
			est := EstimateTokens(messages)
			if a.compactor.ShouldCompact(est) {
				messages = CompactMessages(messages, 4)
				after := EstimateTokens(messages)
				a.compactor.RecordFire()
				a.bus.PublishSimple(eventbus.EventContextCompact, string(sessionID), string(turnID), map[string]any{
					"tokens_before": est,
					"tokens_after":  after,
					"fired":         a.compactor.Fired,
				})
			}
		}

		resp, err := a.provider.Generate(turnCtx, llm.Request{
			Model:       a.model,
			Messages:    messages,
			Tools:       toolSpecs,
			Temperature: a.temperature,
			MaxTokens:   a.maxTokens,
		})
		if err != nil {
			runErr = err
			turnSpan.RecordError(err)
			turnSpan.SetStatus(codes.Error, err.Error())
			a.bus.PublishSimple(eventbus.EventAgentError, string(sessionID), string(turnID), map[string]any{
				"error": err.Error(),
			})
			a.bus.PublishSimple(eventbus.EventTurnEnd, string(sessionID), string(turnID), map[string]any{
				"error": true,
			})
			turnSpan.End()
			break
		}

		a.bus.PublishSimple(eventbus.EventModelResponse, string(sessionID), string(turnID), map[string]any{
			"tokens":        resp.TokensUsed,
			"latency_ms":    resp.LatencyMs,
			"finish_reason": resp.FinishReason,
		})
		usedTokens += resp.TokensUsed
		if a.tokenBudget > 0 && usedTokens > a.tokenBudget {
			runErr = fmt.Errorf("token budget exceeded (%d > %d)", usedTokens, a.tokenBudget)
			a.bus.PublishSimple(eventbus.EventAgentError, string(sessionID), string(turnID), map[string]any{
				"error": runErr.Error(),
			})
			turnSpan.End()
			break
		}

		assistant := resp.Message
		if assistant.Role == "" {
			assistant.Role = llm.RoleAssistant
		}
		messages = append(messages, assistant)
		turn.ModelOutput = assistant.Content

		if len(assistant.ToolCalls) == 0 {
			finalText = assistant.Content
			if a.guardrails != nil {
				filtered, d := a.guardrails.FilterOutput(finalText)
				finalText = filtered
				if d.Action == guardrails.Deny {
					a.bus.PublishSimple(eventbus.EventGuardrailBlock, string(sessionID), string(turnID), map[string]any{
						"kind":    "output",
						"message": d.Message,
					})
				}
			}
			session.AddTurn(turn)
			a.bus.PublishSimple(eventbus.EventTurnEnd, string(sessionID), string(turnID), map[string]any{
				"final": true,
			})
			turnSpan.End()
			break
		}

		for _, tc := range assistant.ToolCalls {
			args := map[string]any{}
			if tc.Function.Arguments != "" {
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
			}

			callID := domain.ToolCallID(tc.ID)
			if callID == "" {
				callID = domain.ToolCallID(uuid.NewString())
			}

			name := tc.Function.Name
			content, errText, blocked := a.invokeTool(turnCtx, string(sessionID), string(turnID), name, args)

			result := &domain.ToolResult{Content: content}
			turn.ToolCalls = append(turn.ToolCalls, domain.ToolCall{
				ID:        callID,
				TurnID:    turnID,
				Tool:      domain.ToolName(name),
				Args:      args,
				Result:    result,
				Error:     errText,
				CreatedAt: time.Now().UTC(),
			})

			a.bus.PublishSimple(eventbus.EventToolResult, string(sessionID), string(turnID), map[string]any{
				"tool":    name,
				"error":   errText != "",
				"blocked": blocked,
				"preview": toolResultPreview(content, errText),
			})

			messages = append(messages, llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: string(callID),
				Name:       name,
				Content:    content,
			})
		}

		session.AddTurn(turn)
		a.bus.PublishSimple(eventbus.EventTurnEnd, string(sessionID), string(turnID), map[string]any{
			"tool_calls": len(turn.ToolCalls),
		})
		turnSpan.End()

		if depth == a.maxLoopDepth {
			runErr = domain.ErrMaxDepthExceeded
			a.bus.PublishSimple(eventbus.EventAgentError, string(sessionID), string(turnID), map[string]any{
				"error": runErr.Error(),
				"depth": depth,
			})
		}
	}

	session.End()
	_ = a.sessions.Save(ctx, session)
	a.bus.PublishSimple(eventbus.EventSessionEnd, string(sessionID), "", map[string]any{
		"turns": len(session.Turns),
		"error": runErr != nil,
	})

	return RunResult{
		SessionID: sessionID,
		FinalText: finalText,
		Turns:     len(session.Turns),
		Err:       runErr,
	}
}

func (a *Agent) invokeTool(ctx context.Context, sessionID, turnID, name string, args map[string]any) (content, errText string, blocked bool) {
	toolCtx, toolSpan := observability.Tracer("agent").Start(ctx, "tool:"+name)
	defer toolSpan.End()

	if m, ok := a.manifests.Get(name); ok {
		if m.BlockedInModelLoop {
			msg := "blocked: tool is not available in the model loop"
			a.bus.PublishSimple(eventbus.EventToolCall, sessionID, turnID, map[string]any{
				"tool": name, "args": args, "blocked": true,
			})
			return msg, msg, true
		}
	} else if a.production || (a.manifests != nil && a.manifests.Production()) {
		msg := "blocked: missing tool manifest"
		a.bus.PublishSimple(eventbus.EventGuardrailBlock, sessionID, turnID, map[string]any{
			"kind": "manifest", "tool": name,
		})
		return msg, msg, true
	}

	ev := &eventbus.Event{
		Type:      eventbus.EventToolCall,
		SessionID: sessionID,
		TurnID:    turnID,
		Data:      map[string]any{"tool": name, "args": args},
	}
	res, _ := a.bus.Publish(ev)
	if res != nil && res.Modified && ev.Data != nil {
		if t, ok := ev.Data["tool"].(string); ok && t != "" {
			name = t
		}
		if next, ok := ev.Data["args"].(map[string]any); ok {
			args = next
		}
	}
	if res != nil && res.Block {
		msg := "blocked by policy"
		if res.Error != nil {
			msg = "blocked: " + res.Error.Error()
		}
		a.bus.PublishSimple(eventbus.EventGuardrailBlock, sessionID, turnID, map[string]any{
			"kind": "tool", "tool": name, "message": msg,
		})
		return msg, msg, true
	}

	out, execErr := a.registry.Execute(toolCtx, name, args)
	if execErr != nil {
		errText = execErr.Error()
		toolSpan.RecordError(execErr)
		toolSpan.SetStatus(codes.Error, errText)
		return "error: " + errText, errText, false
	}
	if a.compressToolBytes > 0 {
		out = CompressToolResult(out, a.compressToolBytes)
	}
	return out, "", false
}
