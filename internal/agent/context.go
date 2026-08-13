package agent

import (
	"fmt"
	"strings"

	"github.com/budaev/agent/internal/llm"
	"github.com/budaev/agent/internal/skills"
	"github.com/budaev/agent/internal/tools"
)

const defaultSystemPrompt = `You are a coding AI agent. Your job is to help the user solve software development tasks.

Rules:
1. Always follow the user's instructions.
2. If the task is unclear — ask clarifying questions.
3. Use tools to read/write files, run commands, and search the codebase.
4. Before changing files — read them to understand the context.
5. After changes — run tests if they exist.
6. If something does not work — analyze errors and propose fixes.
7. Do not invent facts. If you do not know — say so.
8. Stay safe: do not run dangerous commands without confirmation.

Response format:
- A short explanation of what you are doing.
- Tool calls (if needed).
- The result.

Start working.`

// ContextBuilder builds the layered prompt messages (SYSTEM + TASK minimum).
type ContextBuilder struct {
	SystemPrompt string
	Skills       *skills.Manager
}

// NewContextBuilder creates a builder with the default system prompt.
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{SystemPrompt: defaultSystemPrompt}
}

// Build returns initial messages for a task.
func (b *ContextBuilder) Build(task string, registry *tools.Registry) []llm.Message {
	sys := b.SystemPrompt
	if registry != nil {
		names := registry.Names()
		if len(names) > 0 {
			sys = fmt.Sprintf("%s\n\nAvailable tools: %s", sys, strings.Join(names, ", "))
		}
	}
	if b.Skills != nil {
		if idx := b.Skills.PromptIndex(); idx != "" {
			sys = sys + "\n\n" + idx
		}
		for _, s := range b.Skills.Match(task, nil) {
			loaded, err := b.Skills.LoadSkill(s.Name)
			if err != nil || loaded.Body() == "" {
				continue
			}
			sys = fmt.Sprintf("%s\n\nSkill %s:\n%s", sys, loaded.Name, loaded.Body())
		}
	}
	return []llm.Message{
		{Role: llm.RoleSystem, Content: sys},
		{Role: llm.RoleUser, Content: task},
	}
}

// ToolSpecs converts registry definitions to LLM tool specs.
func ToolSpecs(registry *tools.Registry) []llm.ToolSpec {
	defs := registry.Definitions()
	out := make([]llm.ToolSpec, 0, len(defs))
	for _, d := range defs {
		out = append(out, llm.ToolSpec{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        d.Name,
				Description: d.Description,
				Parameters:  d.Parameters,
			},
		})
	}
	return out
}
