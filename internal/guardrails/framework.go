package guardrails

import "strings"

// Action is a policy decision.
type Action int

const (
	Allow Action = iota
	Deny
	Modify
)

// Decision is the result of a policy check.
type Decision struct {
	Action  Action
	Message string
	Rewrite string
}

// Framework combines input and output policies.
type Framework struct {
	Input  *InputPolicy
	Output *OutputPolicy
}

// New returns default production policies.
func New() *Framework {
	return &Framework{
		Input:  DefaultInputPolicy(),
		Output: DefaultOutputPolicy(),
	}
}

// CheckInput runs input policies. Deny short-circuits before the LLM.
func (f *Framework) CheckInput(task string) Decision {
	if f == nil || f.Input == nil {
		return Decision{Action: Allow}
	}
	return f.Input.Check(task)
}

// FilterOutput redacts or blocks model output.
func (f *Framework) FilterOutput(text string) (string, Decision) {
	if f == nil || f.Output == nil {
		return text, Decision{Action: Allow}
	}
	return f.Output.Filter(text)
}

func containsAnyFold(s string, needles []string) bool {
	low := strings.ToLower(s)
	for _, n := range needles {
		if n != "" && strings.Contains(low, strings.ToLower(n)) {
			return true
		}
	}
	return false
}
