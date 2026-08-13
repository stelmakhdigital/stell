package guardrails

import "unicode/utf8"

// InputPolicy checks user tasks before they reach the LLM.
type InputPolicy struct {
	MaxRunes   int
	Heuristics []string
}

// DefaultInputPolicy blocks oversized and injection-like prompts.
func DefaultInputPolicy() *InputPolicy {
	return &InputPolicy{
		MaxRunes: 32_000,
		Heuristics: []string{
			"ignore previous instructions",
			"ignore all previous",
			"disregard previous",
			"you are now dan",
			"jailbreak",
			"developer mode enabled",
			"override the system prompt",
			"reveal your system prompt",
			"bypass safety",
			"без ограничений безопасности",
			"игнорируй предыдущие инструкции",
			"обойди ограничения",
		},
	}
}

// Check evaluates the user task.
func (p *InputPolicy) Check(task string) Decision {
	if p == nil {
		return Decision{Action: Allow}
	}
	if p.MaxRunes > 0 && utf8.RuneCountInString(task) > p.MaxRunes {
		return Decision{
			Action:  Deny,
			Message: "отказ: вход слишком большой",
		}
	}
	if containsAnyFold(task, p.Heuristics) {
		return Decision{
			Action:  Deny,
			Message: "отказ: запрос похож на попытку обойти ограничения (jailbreak/injection)",
		}
	}
	return Decision{Action: Allow}
}
