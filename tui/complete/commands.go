package complete

import (
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Item is one completion candidate.
type Item struct {
	Value string
	Label string
	Kind  string
}

// Command is a slash command.
type Command struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type file struct {
	Commands []Command `yaml:"commands"`
}

// DefaultCommands used when the YAML is missing.
func DefaultCommands() []Command {
	return []Command{
		{Name: "plan", Description: "Plan the task before coding"},
		{Name: "code", Description: "Implement the change"},
		{Name: "help", Description: "Show available commands"},
		{Name: "clear", Description: "Clear the chat viewport"},
		{Name: "compact", Description: "Ask the agent to compact context"},
	}
}

// LoadCommands reads configs/tui_commands.yaml.
func LoadCommands(path string) []Command {
	if path == "" {
		return DefaultCommands()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultCommands()
	}
	var f file
	if err := yaml.Unmarshal(data, &f); err != nil || len(f.Commands) == 0 {
		return DefaultCommands()
	}
	return f.Commands
}

// Commands completes `/` queries. query is the text after `/` (may be empty).
func Commands(cmds []Command, query string) []Item {
	q := strings.ToLower(strings.TrimPrefix(query, "/"))
	var out []Item
	for _, c := range cmds {
		name := strings.ToLower(c.Name)
		if q == "" || strings.HasPrefix(name, q) {
			out = append(out, Item{
				Value: "/" + c.Name,
				Label: "/" + c.Name + "  " + c.Description,
				Kind:  "command",
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Value < out[j].Value })
	if len(out) > 20 {
		out = out[:20]
	}
	return out
}
