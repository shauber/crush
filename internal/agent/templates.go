package agent

import (
	"fmt"

	"github.com/charmbracelet/crush/internal/agent/prompt"
)

// TemplateInfo describes an available prompt template.
type TemplateInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AvailableTemplates returns metadata about all available prompt templates.
func AvailableTemplates() []TemplateInfo {
	return []TemplateInfo{
		{
			ID:          "coder",
			Name:        "Coder",
			Description: "Main development agent for code analysis, editing, and execution",
		},
		{
			ID:          "task",
			Name:        "Task",
			Description: "Specialized agent for executing specific tasks",
		},
		{
			ID:          "tachikoma",
			Name:        "Tachikoma",
			Description: "Adorable but deadly security expert - cybersecurity specialist with APT threat focus",
		},
	}
}

// GetPromptByID returns a prompt builder for the given template ID.
func GetPromptByID(id string, opts ...prompt.Option) (*prompt.Prompt, error) {
	switch id {
	case "coder":
		return coderPrompt(opts...)
	case "task":
		return taskPrompt(opts...)
	case "tachikoma":
		return tachikomaPrompt(opts...)
	default:
		return nil, fmt.Errorf("unknown template ID: %s (available: coder, task, tachikoma)", id)
	}
}
