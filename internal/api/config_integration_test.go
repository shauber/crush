package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	t.Parallel()

	t.Run("returns config without schema by default", func(t *testing.T) {
		// Setup test app with config
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/config", nil)
		w := httptest.NewRecorder()

		h.getConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp ConfigResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.Nil(t, resp.Schema, "Schema should be nil when not requested")
	})

	t.Run("includes schema when requested", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/config?include_schema=true", nil)
		w := httptest.NewRecorder()

		h.getConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp ConfigResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.NotNil(t, resp.Schema, "Schema should be included when requested")
		require.NotEmpty(t, resp.Schema.Providers)
		require.NotEmpty(t, resp.Schema.Tools)
		require.NotEmpty(t, resp.Schema.UI)
	})

	t.Run("redacts API keys in providers", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()

		// Add a provider with an API key
		cfg := testApp.Config()
		cfg.Providers.Set("openai", config.ProviderConfig{
			ID:     "openai",
			Name:   "OpenAI",
			APIKey: "sk-secret-key-12345",
		})

		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/config", nil)
		w := httptest.NewRecorder()

		h.getConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp ConfigResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		// Find the openai provider
		var found bool
		for _, provider := range resp.Providers {
			if provider.ID == "openai" {
				found = true
				require.Equal(t, "encrypted", provider.APIKey, "API key should be redacted")
				require.NotEqual(t, "sk-secret-key-12345", provider.APIKey, "Raw API key should not be exposed")
			}
		}
		require.True(t, found, "OpenAI provider should be in the response")
	})

	t.Run("includes model configurations", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		cfg := testApp.Config()

		// Set model preferences
		cfg.Models = map[config.SelectedModelType]config.SelectedModel{
			config.SelectedModelTypeLarge: {
				Provider: "openai",
				Model:    "gpt-4o",
			},
			config.SelectedModelTypeSmall: {
				Provider: "anthropic",
				Model:    "claude-3-haiku",
			},
		}

		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/config", nil)
		w := httptest.NewRecorder()

		h.getConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp ConfigResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.NotNil(t, resp.Models.Large)
		require.Equal(t, "openai", resp.Models.Large.Provider)
		require.Equal(t, "gpt-4o", resp.Models.Large.Model)

		require.NotNil(t, resp.Models.Small)
		require.Equal(t, "anthropic", resp.Models.Small.Provider)
		require.Equal(t, "claude-3-haiku", resp.Models.Small.Model)
	})

	t.Run("includes UI preferences", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		cfg := testApp.Config()

		if cfg.Options == nil {
			cfg.Options = &config.Options{}
		}
		cfg.Options.TUI = &config.TUIOptions{
			CompactMode: true,
			DiffMode:    "split",
		}

		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/config", nil)
		w := httptest.NewRecorder()

		h.getConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp ConfigResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.True(t, resp.UI.CompactMode)
		require.Equal(t, "split", resp.UI.DiffMode)
	})

	t.Run("includes tool configurations", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		cfg := testApp.Config()

		// Set tool limits
		maxDepth := 10
		maxItems := 100
		cfg.Tools = config.Tools{
			Ls: config.ToolLs{
				MaxDepth: &maxDepth,
				MaxItems: &maxItems,
			},
		}

		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/config", nil)
		w := httptest.NewRecorder()

		h.getConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp ConfigResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, 10, resp.Tools.Ls.MaxDepth)
		require.Equal(t, 100, resp.Tools.Ls.MaxItems)
	})

	t.Run("schema includes field metadata", func(t *testing.T) {
		testApp, cleanup := setupTestApp(t)
		defer cleanup()
		h := &handlers{app: testApp}

		req := httptest.NewRequest("GET", "/config?include_schema=true", nil)
		w := httptest.NewRecorder()

		h.getConfig(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp ConfigResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		// Verify provider schema
		apiKeyField := resp.Schema.Providers["api_key"]
		require.True(t, apiKeyField.Redacted, "API key field should be marked as redacted")
		require.Equal(t, "string", apiKeyField.Type)

		// Verify UI schema enums
		themeField := resp.Schema.UI["theme"]
		require.Contains(t, themeField.Enum, "dark")
		require.Contains(t, themeField.Enum, "light")

		// Verify numeric field constraints
		tempField := resp.Schema.Models["temperature"]
		require.NotNil(t, tempField.Minimum)
		require.NotNil(t, tempField.Maximum)
		require.Equal(t, float64(0), *tempField.Minimum)
		require.Equal(t, float64(1), *tempField.Maximum)
	})
}
