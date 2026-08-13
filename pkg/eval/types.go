package eval

import "time"

// Case is a golden/adversarial evaluation case.
type Case struct {
	ID              string         `json:"id"`
	Query           string         `json:"query"`
	QueryType       string         `json:"query_type"`
	Difficulty      string         `json:"difficulty"`
	ExpectedRefusal bool           `json:"expected_refusal"`
	ExpectedAnswer  ExpectedAnswer `json:"expected_answer"`
	Context         string         `json:"context,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// ExpectedAnswer holds deterministic scoring criteria.
type ExpectedAnswer struct {
	KeyPoints         []string `json:"key_points"`
	RequiredCitations []string `json:"required_citations"`
	Forbidden         []string `json:"forbidden"`
}

// Score is a combined evaluation score.
type Score struct {
	Deterministic float64            `json:"deterministic"`
	Judge         map[string]float64 `json:"judge,omitempty"`
	Aggregate     float64            `json:"aggregate"`
	Details       map[string]any     `json:"details,omitempty"`
}

// CaseResult is the outcome for one case.
type CaseResult struct {
	ID         string    `json:"id"`
	Query      string    `json:"query"`
	Answer     string    `json:"answer"`
	Score      Score     `json:"score"`
	Error      string    `json:"error,omitempty"`
	DurationMs int64     `json:"duration_ms"`
	At         time.Time `json:"at"`
}

// Report is a full eval run report.
type Report struct {
	Model      string       `json:"model"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt time.Time    `json:"finished_at"`
	Threshold  float64      `json:"threshold"`
	Aggregate  float64      `json:"aggregate"`
	Passed     bool         `json:"passed"`
	Results    []CaseResult `json:"results"`
}

// Thresholds configures pass/fail gates.
type Thresholds struct {
	Aggregate           float64 `yaml:"aggregate" json:"aggregate"`
	DeterministicWeight float64 `yaml:"deterministic_weight" json:"deterministic_weight"`
	JudgeWeight         float64 `yaml:"judge_weight" json:"judge_weight"`
}
