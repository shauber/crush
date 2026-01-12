package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/crush/internal/agent"
	"github.com/stretchr/testify/require"
)

func TestListTemplates(t *testing.T) {
	t.Parallel()

	app, cleanup := setupTestApp(t)
	defer cleanup()

	h := &handlers{app: app}

	req := httptest.NewRequest("GET", "/templates", nil)
	w := httptest.NewRecorder()

	h.listTemplates(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	templates, ok := response["templates"].([]interface{})
	require.True(t, ok, "Expected templates to be an array")
	require.NotEmpty(t, templates, "Expected at least one template")

	// Verify template structure
	firstTemplate := templates[0].(map[string]interface{})
	require.Contains(t, firstTemplate, "id")
	require.Contains(t, firstTemplate, "name")
	require.Contains(t, firstTemplate, "description")

	// Verify expected templates exist
	expectedTemplates := agent.AvailableTemplates()
	require.Len(t, templates, len(expectedTemplates))

	// Verify we have coder and task templates
	templateIDs := make(map[string]bool)
	for _, tmpl := range templates {
		tmplMap := tmpl.(map[string]interface{})
		templateIDs[tmplMap["id"].(string)] = true
	}
	require.True(t, templateIDs["coder"], "Expected 'coder' template")
	require.True(t, templateIDs["task"], "Expected 'task' template")
}
