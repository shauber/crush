package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSettings(t *testing.T) {
	t.Run("returns all settings", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/settings", nil)
		w := httptest.NewRecorder()

		h.getSettings(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp SettingsResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		// Check that all sections are present
		require.NotNil(t, resp.UI)
		require.NotNil(t, resp.Editor)
		require.NotNil(t, resp.Models)
		require.NotNil(t, resp.General)
	})
}

func TestGetSettingsSchema(t *testing.T) {
	t.Run("returns all schemas", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/settings/schema", nil)
		w := httptest.NewRecorder()

		h.getSettingsSchema(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		settings, ok := resp["settings"].([]interface{})
		require.True(t, ok)
		require.NotEmpty(t, settings)
	})

	t.Run("returns schema for specific setting", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/settings/ui.theme/schema", nil)
		req.SetPathValue("key", "ui.theme")
		w := httptest.NewRecorder()

		h.getSettingsSchema(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp SettingSchema
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, "ui.theme", resp.Key)
		require.Equal(t, "string", resp.Type)
		require.Contains(t, resp.Enum, "dark")
		require.Contains(t, resp.Enum, "light")
	})

	t.Run("returns 404 for unknown setting", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/settings/unknown.key/schema", nil)
		req.SetPathValue("key", "unknown.key")
		w := httptest.NewRecorder()

		h.getSettingsSchema(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestUpdateSettings(t *testing.T) {
	t.Run("update UI settings", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := SettingsUpdateRequest{
			UI: &UISettingsDTO{
				CompactMode: true,
				DiffMode:    "split",
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateSettings(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp SettingsResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.True(t, resp.UI.CompactMode)
		require.Equal(t, "split", resp.UI.DiffMode)
	})

	t.Run("update editor settings", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		contextPaths := []string{".cursorrules", "CRUSH.md"}
		reqBody := SettingsUpdateRequest{
			Editor: &EditorSettingsDTO{
				ContextPaths: contextPaths,
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateSettings(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp SettingsResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, contextPaths, resp.Editor.ContextPaths)
	})

	t.Run("update models settings", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := SettingsUpdateRequest{
			Models: &ModelsSettingsDTO{
				DefaultLarge: "openai/gpt-4o",
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateSettings(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp SettingsResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, "openai/gpt-4o", resp.Models.DefaultLarge)
	})

	t.Run("update general settings", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := SettingsUpdateRequest{
			General: &GeneralSettingsDTO{
				Debug:          true,
				DisableMetrics: true,
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateSettings(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp SettingsResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.True(t, resp.General.Debug)
		require.True(t, resp.General.DisableMetrics)
	})

	t.Run("update multiple sections", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := SettingsUpdateRequest{
			UI: &UISettingsDTO{
				CompactMode: true,
			},
			General: &GeneralSettingsDTO{
				Debug: true,
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
		w := httptest.NewRecorder()

		h.updateSettings(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp SettingsResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.True(t, resp.UI.CompactMode)
		require.True(t, resp.General.Debug)
	})

	t.Run("malformed JSON request", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("PUT", "/settings", bytes.NewReader([]byte("{invalid")))
		w := httptest.NewRecorder()

		h.updateSettings(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateSetting(t *testing.T) {
	t.Run("update single setting - compact_mode", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := SettingUpdateRequest{
			Value: true,
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings/ui.compact_mode", bytes.NewReader(body))
		req.SetPathValue("key", "ui.compact_mode")
		w := httptest.NewRecorder()

		h.updateSetting(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, "Setting updated successfully", resp["message"])
		require.Equal(t, "ui.compact_mode", resp["key"])
		require.Equal(t, true, resp["value"])

		// Verify the setting was updated
		cfg := testApp.Config()
		require.True(t, cfg.Options.TUI.CompactMode)
	})

	t.Run("update single setting - diff_mode", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := SettingUpdateRequest{
			Value: "split",
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings/ui.diff_mode", bytes.NewReader(body))
		req.SetPathValue("key", "ui.diff_mode")
		w := httptest.NewRecorder()

		h.updateSetting(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify the setting was updated
		cfg := testApp.Config()
		require.Equal(t, "split", cfg.Options.TUI.DiffMode)
	})

	t.Run("update single setting - debug", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := SettingUpdateRequest{
			Value: true,
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings/general.debug", bytes.NewReader(body))
		req.SetPathValue("key", "general.debug")
		w := httptest.NewRecorder()

		h.updateSetting(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify the setting was updated
		cfg := testApp.Config()
		require.True(t, cfg.Options.Debug)
	})

	t.Run("update single setting - context_paths", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		paths := []string{".cursorrules", "CRUSH.md"}
		reqBody := SettingUpdateRequest{
			Value: paths,
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings/editor.context_paths", bytes.NewReader(body))
		req.SetPathValue("key", "editor.context_paths")
		w := httptest.NewRecorder()

		h.updateSetting(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify the setting was updated
		cfg := testApp.Config()
		require.Len(t, cfg.Options.ContextPaths, 2)
	})

	t.Run("validation error - invalid value type", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := SettingUpdateRequest{
			Value: "not a boolean",
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings/ui.compact_mode", bytes.NewReader(body))
		req.SetPathValue("key", "ui.compact_mode")
		w := httptest.NewRecorder()

		h.updateSetting(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "must be a boolean")
	})

	t.Run("validation error - invalid theme value", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		reqBody := SettingUpdateRequest{
			Value: "purple",
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/settings/ui.theme", bytes.NewReader(body))
		req.SetPathValue("key", "ui.theme")
		w := httptest.NewRecorder()

		h.updateSetting(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "must be 'dark' or 'light'")
	})

	t.Run("malformed JSON request", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("PUT", "/settings/ui.compact_mode", bytes.NewReader([]byte("{invalid")))
		req.SetPathValue("key", "ui.compact_mode")
		w := httptest.NewRecorder()

		h.updateSetting(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}
