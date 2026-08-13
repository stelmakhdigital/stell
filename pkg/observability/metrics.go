package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds Prometheus collectors for the agent.
type Metrics struct {
	registry             prometheus.Registerer
	gatherer             prometheus.Gatherer
	TurnsTotal           prometheus.Counter
	LoopDepth            prometheus.Histogram
	ToolCallsTotal       *prometheus.CounterVec
	ToolLatency          *prometheus.HistogramVec
	LLMTokensTotal       *prometheus.CounterVec
	LLMLatency           prometheus.Histogram
	AgentErrorsTotal     prometheus.Counter
	GuardrailBlocksTotal *prometheus.CounterVec
	HitlRequestsTotal    prometheus.Counter
	HitlDecisionsTotal   *prometheus.CounterVec
	PromptCacheHits      prometheus.Counter
	PromptCacheMisses    prometheus.Counter
	CompactTokensBefore  prometheus.Counter
	CompactTokensAfter   prometheus.Counter
}

// NewMetrics registers MVP metrics on a dedicated registry.
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	_ = reg.Register(collectors.NewGoCollector())
	return NewMetricsWithRegisterer(reg, reg)
}

// NewMetricsWithRegisterer allows custom registerer/gatherer (tests).
func NewMetricsWithRegisterer(reg prometheus.Registerer, gatherer prometheus.Gatherer) *Metrics {
	m := &Metrics{registry: reg, gatherer: gatherer}
	m.TurnsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "agent_turns_total",
		Help: "Total agent turns",
	})
	m.LoopDepth = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "agent_loop_depth",
		Help:    "Loop depth at turn end",
		Buckets: []float64{1, 2, 3, 5, 8, 13, 21, 34, 50},
	})
	m.ToolCallsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tool_calls_total",
		Help: "Tool invocations by name and status",
	}, []string{"tool", "status"})
	m.ToolLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tool_latency_seconds",
		Help:    "Tool execution latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"tool"})
	m.LLMTokensTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "llm_tokens_total",
		Help: "LLM tokens by direction",
	}, []string{"direction"})
	m.LLMLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "llm_latency_seconds",
		Help:    "LLM call latency",
		Buckets: prometheus.DefBuckets,
	})
	m.AgentErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "agent_errors_total",
		Help: "Agent errors",
	})
	m.GuardrailBlocksTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "guardrail_blocks_total",
		Help: "Guardrail blocks by kind",
	}, []string{"kind"})
	m.HitlRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hitl_requests_total",
		Help: "HITL approval requests",
	})
	m.HitlDecisionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hitl_decisions_total",
		Help: "HITL decisions by outcome",
	}, []string{"decision"})
	m.PromptCacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prompt_cache_hits_total",
		Help: "Prompt cache hits on stable system prefix",
	})
	m.PromptCacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "prompt_cache_misses_total",
		Help: "Prompt cache misses",
	})
	m.CompactTokensBefore = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "compact_tokens_before_total",
		Help: "Estimated tokens before context compaction",
	})
	m.CompactTokensAfter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "compact_tokens_after_total",
		Help: "Estimated tokens after context compaction",
	})
	reg.MustRegister(
		m.TurnsTotal, m.LoopDepth, m.ToolCallsTotal, m.ToolLatency,
		m.LLMTokensTotal, m.LLMLatency, m.AgentErrorsTotal,
		m.GuardrailBlocksTotal, m.HitlRequestsTotal, m.HitlDecisionsTotal,
		m.PromptCacheHits, m.PromptCacheMisses, m.CompactTokensBefore, m.CompactTokensAfter,
	)
	return m
}

// Handler returns the Prometheus HTTP handler.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.gatherer, promhttp.HandlerOpts{})
}
