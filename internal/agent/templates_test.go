package agent

import (
	"testing"

	"github.com/charmbracelet/crush/internal/agent/prompt"
)

func TestAvailableTemplates(t *testing.T) {
	templates := AvailableTemplates()
	if len(templates) == 0 {
		t.Fatal("Expected at least one template")
	}

	// Verify expected templates exist
	found := make(map[string]bool)
	for _, tmpl := range templates {
		if tmpl.ID == "" {
			t.Errorf("Template has empty ID")
		}
		if tmpl.Name == "" {
			t.Errorf("Template %s has empty Name", tmpl.ID)
		}
		if tmpl.Description == "" {
			t.Errorf("Template %s has empty Description", tmpl.ID)
		}
		found[tmpl.ID] = true
	}

	// Verify core templates are present
	if !found["coder"] {
		t.Error("Expected 'coder' template to be available")
	}
	if !found["task"] {
		t.Error("Expected 'task' template to be available")
	}
}

func TestGetPromptByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        string
		expectErr bool
	}{
		{
			name:      "valid coder template",
			id:        "coder",
			expectErr: false,
		},
		{
			name:      "valid task template",
			id:        "task",
			expectErr: false,
		},
		{
			name:      "invalid template",
			id:        "invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := GetPromptByID(tt.id, prompt.WithWorkingDir("/tmp"))
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if p == nil {
				t.Error("Expected prompt but got nil")
			}
		})
	}
}
