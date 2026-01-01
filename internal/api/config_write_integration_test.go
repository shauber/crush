package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/stretchr/testify/require"
)

func TestUpdateConfig(t *testing.T) {
	t.Run("update models configuration", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		temp := 0.8
		reqBody := ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:       "gpt-4o",
					Provider:    "openai",
					Temperature: &temp,
				},
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp ConfigResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.NotNil(t, resp.Models.Large)
		require.Equal(t, "gpt-4o", resp.Models.Large.Model)
		require.Equal(t, "openai", resp.Models.Large.Provider)
	})

	t.Run("update tools configuration", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		maxDepth := 15
		maxItems := 200
		reqBody := ConfigUpdateRequest{
			Tools: &ToolsUpdateRequest{
				Ls: &LsToolUpdateDTO{
					MaxDepth: &maxDepth,
					MaxItems: &maxItems,
				},
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify the config was updated
		cfg := testApp.Config()
		depth, items := cfg.Tools.Ls.Limits()
		require.Equal(t, 15, depth)
		require.Equal(t, 200, items)
	})

	t.Run("update UI preferences", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		compactMode := true
		diffMode := "split"
		reqBody := ConfigUpdateRequest{
			UI: &UIUpdateRequest{
				CompactMode: &compactMode,
				DiffMode:    &diffMode,
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp ConfigResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.True(t, resp.UI.CompactMode)
		require.Equal(t, "split", resp.UI.DiffMode)
	})

	t.Run("validation error for invalid temperature", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		temp := 2.0
		reqBody := ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:       "gpt-4o",
					Provider:    "openai",
					Temperature: &temp,
				},
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfig(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "temperature")
	})

	t.Run("validation error for invalid theme", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		theme := "purple"
		reqBody := ConfigUpdateRequest{
			UI: &UIUpdateRequest{
				Theme: &theme,
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfig(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "theme")
	})

	t.Run("malformed JSON request", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("PUT", "/config", bytes.NewReader([]byte("{invalid json")))
		w := httptest.NewRecorder()

		h.updateConfig(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("multiple updates in single request", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		temp := 0.7
		maxDepth := 10
		compactMode := true
		contextPaths := []string{".cursorrules", "CRUSH.md"}

		reqBody := ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:       "gpt-4o",
					Provider:    "openai",
					Temperature: &temp,
				},
			},
			Tools: &ToolsUpdateRequest{
				Ls: &LsToolUpdateDTO{
					MaxDepth: &maxDepth,
				},
			},
			UI: &UIUpdateRequest{
				CompactMode: &compactMode,
			},
			Options: &OptionsUpdateRequest{
				ContextPaths: &contextPaths,
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify all updates were applied
		cfg := testApp.Config()

		model, ok := cfg.Models[config.SelectedModelTypeLarge]
		require.True(t, ok)
		require.Equal(t, "gpt-4o", model.Model)

		depth, _ := cfg.Tools.Ls.Limits()
		require.Equal(t, 10, depth)

		require.True(t, cfg.Options.TUI.CompactMode)
		require.Equal(t, contextPaths, cfg.Options.ContextPaths)
	})
}

func TestUpdateConfigField(t *testing.T) {
	t.Run("update single field - compact_mode", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := FieldUpdateRequest{
			Path:  "ui.compact_mode",
			Value: true,
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config/field", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfigField(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, "Field updated successfully", resp["message"])
		require.Equal(t, "ui.compact_mode", resp["path"])

		// Verify the field was updated
		cfg := testApp.Config()
		require.True(t, cfg.Options.TUI.CompactMode)
	})

	t.Run("update single field - ls max_depth", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := FieldUpdateRequest{
			Path:  "tools.ls.max_depth",
			Value: float64(20),
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config/field", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfigField(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify the field was updated
		cfg := testApp.Config()
		depth, _ := cfg.Tools.Ls.Limits()
		require.Equal(t, 20, depth)
	})

	t.Run("update single field - diff_mode", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := FieldUpdateRequest{
			Path:  "ui.diff_mode",
			Value: "split",
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config/field", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfigField(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify the field was updated
		cfg := testApp.Config()
		require.Equal(t, "split", cfg.Options.TUI.DiffMode)
	})

	t.Run("validation error - invalid field path", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := FieldUpdateRequest{
			Path:  "invalid.path.here",
			Value: true,
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config/field", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfigField(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "unknown field path")
	})

	t.Run("validation error - invalid value type", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := FieldUpdateRequest{
			Path:  "ui.compact_mode",
			Value: "not a boolean",
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config/field", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfigField(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "must be a boolean")
	})

	t.Run("validation error - out of range value", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := FieldUpdateRequest{
			Path:  "models.large.temperature",
			Value: 5.0,
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/config/field", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateConfigField(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "must be between 0 and 1")
	})

	t.Run("malformed JSON request", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("PUT", "/config/field", bytes.NewReader([]byte("{bad json")))
		w := httptest.NewRecorder()

		h.updateConfigField(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}
