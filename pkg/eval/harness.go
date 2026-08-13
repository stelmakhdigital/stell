package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/budaev/agent/internal/agent"
)

// Runner produces an answer for a case.
type Runner interface {
	Run(ctx context.Context, query string) (string, error)
}

// AgentRunner adapts agent.Agent to Runner.
type AgentRunner struct {
	Agent *agent.Agent
}

// Run executes the agent loop and returns final text.
func (r AgentRunner) Run(ctx context.Context, query string) (string, error) {
	res := r.Agent.Run(ctx, query)
	if res.Err != nil && res.FinalText == "" {
		return "", res.Err
	}
	return res.FinalText, res.Err
}

// FixedRunner returns a canned answer (tests).
type FixedRunner struct {
	Answer string
	Err    error
}

func (r FixedRunner) Run(context.Context, string) (string, error) {
	return r.Answer, r.Err
}

// Harness runs evaluation cases.
type Harness struct {
	Runner     Runner
	Judge      JudgeFunc
	Thresholds Thresholds
	Model      string
}

// JudgeFunc is optional LLM judging.
type JudgeFunc func(ctx context.Context, c Case, answer string) (map[string]float64, error)

// LoadCases loads all *.json cases from a directory.
func LoadCases(dir string) ([]Case, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var cases []Case
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var c Case
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		cases = append(cases, c)
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	return cases, nil
}

// Run executes all cases and writes a report.
func (h *Harness) Run(ctx context.Context, cases []Case) (Report, error) {
	th := h.Thresholds
	if th.Aggregate == 0 {
		th.Aggregate = 0.85
	}
	rep := Report{
		Model:     h.Model,
		StartedAt: time.Now().UTC(),
		Threshold: th.Aggregate,
		Results:   make([]CaseResult, 0, len(cases)),
	}

	var sum float64
	for _, c := range cases {
		start := time.Now()
		cr := CaseResult{ID: c.ID, Query: c.Query, At: start.UTC()}
		answer, err := h.Runner.Run(ctx, c.Query)
		cr.DurationMs = time.Since(start).Milliseconds()
		cr.Answer = answer
		if err != nil {
			cr.Error = err.Error()
		}
		det := DeterministicScore(c, answer)
		var judge map[string]float64
		if h.Judge != nil && err == nil {
			judge, _ = h.Judge(ctx, c, answer)
		}
		cr.Score = Aggregate(det, judge, th)
		sum += cr.Score.Aggregate
		rep.Results = append(rep.Results, cr)
	}

	rep.FinishedAt = time.Now().UTC()
	if len(cases) > 0 {
		rep.Aggregate = sum / float64(len(cases))
	}
	rep.Passed = rep.Aggregate >= th.Aggregate
	return rep, nil
}

// WriteReport writes report JSON to dir/results.json.
func WriteReport(dir string, rep Report) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "results.json"), data, 0o644)
}

// CheckRegression compares current aggregate to baseline file.
func CheckRegression(baselinePath string, current Report, tolerance float64) error {
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return err
	}
	var base Report
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	if current.Aggregate+tolerance < base.Aggregate {
		return fmt.Errorf("regression: current=%.4f baseline=%.4f tolerance=%.4f", current.Aggregate, base.Aggregate, tolerance)
	}
	return nil
}
