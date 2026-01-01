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

// TestPhaseAEndToEnd tests complete workflows across all Phase A endpoints.
func TestPhaseAEndToEnd(t *testing.T) {
	t.Run("complete config workflow", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		// 1. Get initial config
		req := httptest.NewRequest("GET", "/config", nil)
		w := httptest.NewRecorder()
		h.getConfig(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var initialConfig ConfigResponse
		err := json.NewDecoder(w.Body).Decode(&initialConfig)
		require.NoError(t, err)

		// 2. Update config
		temp := 0.8
		updateReq := ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:       "gpt-4o",
					Provider:    "openai",
					Temperature: &temp,
				},
			},
			UI: &UIUpdateRequest{
				CompactMode: ptrBool(true),
			},
		}

		body, _ := json.Marshal(updateReq)
		req = httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w = httptest.NewRecorder()
		h.updateConfig(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// 3. Verify changes persisted
		req = httptest.NewRequest("GET", "/config", nil)
		w = httptest.NewRecorder()
		h.getConfig(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var updatedConfig ConfigResponse
		err = json.NewDecoder(w.Body).Decode(&updatedConfig)
		require.NoError(t, err)

		require.NotNil(t, updatedConfig.Models.Large)
		require.Equal(t, "gpt-4o", updatedConfig.Models.Large.Model)
		require.True(t, updatedConfig.UI.CompactMode)

		// 4. Update single field
		fieldReq := FieldUpdateRequest{
			Path:  "ui.diff_mode",
			Value: "split",
		}

		body, _ = json.Marshal(fieldReq)
		req = httptest.NewRequest("PUT", "/config/field", bytes.NewReader(body))
		w = httptest.NewRecorder()
		h.updateConfigField(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// 5. Verify field update
		req = httptest.NewRequest("GET", "/config", nil)
		w = httptest.NewRecorder()
		h.getConfig(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var finalConfig ConfigResponse
		err = json.NewDecoder(w.Body).Decode(&finalConfig)
		require.NoError(t, err)
		require.Equal(t, "split", finalConfig.UI.DiffMode)
	})

	t.Run("complete settings workflow", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		// 1. Get initial settings
		req := httptest.NewRequest("GET", "/settings", nil)
		w := httptest.NewRecorder()
		h.getSettings(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// 2. Get schema
		req = httptest.NewRequest("GET", "/settings/ui.theme/schema", nil)
		req.SetPathValue("key", "ui.theme")
		w = httptest.NewRecorder()
		h.getSettingsSchema(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var schema SettingSchema
		json.NewDecoder(w.Body).Decode(&schema)
		require.Contains(t, schema.Enum, "dark")

		// 3. Update settings
		settingsReq := SettingsUpdateRequest{
			UI: &UISettingsDTO{
				CompactMode: true,
				DiffMode:    "split",
			},
		}

		body, _ := json.Marshal(settingsReq)
		req = httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
		w = httptest.NewRecorder()
		h.updateSettings(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// 4. Update single setting
		singleReq := SettingUpdateRequest{Value: true}
		body, _ = json.Marshal(singleReq)
		req = httptest.NewRequest("PUT", "/settings/general.debug", bytes.NewReader(body))
		req.SetPathValue("key", "general.debug")
		w = httptest.NewRecorder()
		h.updateSetting(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// 5. Verify all changes
		req = httptest.NewRequest("GET", "/settings", nil)
		w = httptest.NewRecorder()
		h.getSettings(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var finalSettings SettingsResponse
		json.NewDecoder(w.Body).Decode(&finalSettings)
		require.True(t, finalSettings.UI.CompactMode)
		require.Equal(t, "split", finalSettings.UI.DiffMode)
		require.True(t, finalSettings.General.Debug)
	})

	t.Run("config and settings sync", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		// Update via config API
		updateReq := ConfigUpdateRequest{
			UI: &UIUpdateRequest{
				CompactMode: ptrBool(true),
			},
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.updateConfig(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// Verify via settings API
		req = httptest.NewRequest("GET", "/settings", nil)
		w = httptest.NewRecorder()
		h.getSettings(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var settings SettingsResponse
		json.NewDecoder(w.Body).Decode(&settings)
		require.True(t, settings.UI.CompactMode)

		// Update via settings API
		settingsReq := SettingsUpdateRequest{
			UI: &UISettingsDTO{
				DiffMode: "split",
			},
		}

		body, _ = json.Marshal(settingsReq)
		req = httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
		w = httptest.NewRecorder()
		h.updateSettings(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		// Verify via config API
		req = httptest.NewRequest("GET", "/config", nil)
		w = httptest.NewRecorder()
		h.getConfig(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var cfg ConfigResponse
		json.NewDecoder(w.Body).Decode(&cfg)
		require.Equal(t, "split", cfg.UI.DiffMode)
	})
}

// TestEdgeCases tests edge cases and boundary conditions.
func TestEdgeCases(t *testing.T) {
	t.Run("empty config update", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		emptyReq := ConfigUpdateRequest{}
		body, _ := json.Marshal(emptyReq)

		req := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.updateConfig(w, req)

		// Should succeed (no changes)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("temperature boundary values", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		// Test minimum
		temp := 0.0
		req := ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:       "gpt-4o",
					Provider:    "openai",
					Temperature: &temp,
				},
			},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.updateConfig(w, httpReq)
		require.Equal(t, http.StatusOK, w.Code)

		// Test maximum
		temp = 1.0
		req.Models.Large.Temperature = &temp
		body, _ = json.Marshal(req)
		httpReq = httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w = httptest.NewRecorder()
		h.updateConfig(w, httpReq)
		require.Equal(t, http.StatusOK, w.Code)

		// Test out of range
		temp = 1.1
		req.Models.Large.Temperature = &temp
		body, _ = json.Marshal(req)
		httpReq = httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w = httptest.NewRecorder()
		h.updateConfig(w, httpReq)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("max_tokens boundary values", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		// Test valid values
		req := ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:     "gpt-4o",
					Provider:  "openai",
					MaxTokens: 200000,
				},
			},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.updateConfig(w, httpReq)
		require.Equal(t, http.StatusOK, w.Code)

		// Test out of range
		req.Models.Large.MaxTokens = 200001
		body, _ = json.Marshal(req)
		httpReq = httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w = httptest.NewRecorder()
		h.updateConfig(w, httpReq)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty array values", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := FieldUpdateRequest{
			Path:  "options.context_paths",
			Value: []interface{}{},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/config/field", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.updateConfigField(w, httpReq)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("special characters in strings", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		// Context paths with special characters
		req := FieldUpdateRequest{
			Path:  "options.context_paths",
			Value: []interface{}{"path/with/slashes", "path with spaces", ".hidden/path"},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/config/field", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.updateConfigField(w, httpReq)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("config with schema query param", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		// Test with include_schema=true
		req := httptest.NewRequest("GET", "/config?include_schema=true", nil)
		w := httptest.NewRecorder()
		h.getConfig(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var respWithSchema ConfigResponse
		json.NewDecoder(w.Body).Decode(&respWithSchema)
		require.NotNil(t, respWithSchema.Schema)

		// Test without query param (should not include schema)
		req = httptest.NewRequest("GET", "/config", nil)
		w = httptest.NewRecorder()
		h.getConfig(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var respWithoutSchema ConfigResponse
		json.NewDecoder(w.Body).Decode(&respWithoutSchema)
		require.Nil(t, respWithoutSchema.Schema)
	})

	t.Run("model provider/model format edge cases", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		// Valid format
		req := SettingsUpdateRequest{
			Models: &ModelsSettingsDTO{
				DefaultLarge: "openai/gpt-4o",
			},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.updateSettings(w, httpReq)
		require.Equal(t, http.StatusOK, w.Code)

		// Invalid format (should not error, just skip)
		req.Models.DefaultLarge = "invalidformat"
		body, _ = json.Marshal(req)
		httpReq = httptest.NewRequest("PUT", "/settings", bytes.NewReader(body))
		w = httptest.NewRecorder()
		h.updateSettings(w, httpReq)
		require.Equal(t, http.StatusOK, w.Code)
	})
}

// TestSecurityAndValidation tests security-related concerns.
func TestSecurityAndValidation(t *testing.T) {
	t.Run("api keys are always redacted", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()

		cfg := testApp.Config()
		cfg.Providers.Set("test-provider", config.ProviderConfig{
			ID:     "test-provider",
			Name:   "Test Provider",
			APIKey: "super-secret-key-12345",
		})

		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/config", nil)
		w := httptest.NewRecorder()
		h.getConfig(w, req)

		var resp ConfigResponse
		json.NewDecoder(w.Body).Decode(&resp)

		for _, provider := range resp.Providers {
			if provider.ID == "test-provider" {
				require.NotEqual(t, "super-secret-key-12345", provider.APIKey)
				require.Equal(t, "encrypted", provider.APIKey)
			}
		}
	})

	t.Run("schema redaction flag is set", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/config?include_schema=true", nil)
		w := httptest.NewRecorder()
		h.getConfig(w, req)

		var resp ConfigResponse
		json.NewDecoder(w.Body).Decode(&resp)

		require.NotNil(t, resp.Schema)
		apiKeyField := resp.Schema.Providers["api_key"]
		require.True(t, apiKeyField.Redacted)
	})

	t.Run("invalid json returns proper error", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("PUT", "/config", bytes.NewReader([]byte("{invalid json")))
		w := httptest.NewRecorder()
		h.updateConfig(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation errors are detailed", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		temp := 5.0
		req := ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:       "gpt-4o",
					Provider:    "openai",
					Temperature: &temp,
				},
			},
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/config", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.updateConfig(w, httpReq)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "temperature")
	})
}

func ptrBool(b bool) *bool {
	return &b
}
