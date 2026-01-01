package api

import (
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/csync"
	"github.com/stretchr/testify/require"
)

func TestToConfigResponse(t *testing.T) {
	t.Parallel()

	t.Run("redacts sensitive data", func(t *testing.T) {
		cfg := &config.Config{
			Providers: csync.NewMap[string, config.ProviderConfig](),
			Options:   &config.Options{},
		}

		// Add a provider with an API key
		cfg.Providers.Set("openai", config.ProviderConfig{
			ID:     "openai",
			Name:   "OpenAI",
			APIKey: "sk-1234567890",
		})

		resp := ToConfigResponse(cfg, false)

		require.Len(t, resp.Providers, 1)
		require.Equal(t, "encrypted", resp.Providers[0].APIKey, "API key should be redacted")
	})

	t.Run("handles empty API key", func(t *testing.T) {
		cfg := &config.Config{
			Providers: csync.NewMap[string, config.ProviderConfig](),
			Options:   &config.Options{},
		}

		cfg.Providers.Set("anthropic", config.ProviderConfig{
			ID:     "anthropic",
			Name:   "Anthropic",
			APIKey: "",
		})

		resp := ToConfigResponse(cfg, false)

		require.Len(t, resp.Providers, 1)
		require.Equal(t, "not_configured", resp.Providers[0].APIKey, "Empty API key should be marked as not configured")
	})

	t.Run("includes schema when requested", func(t *testing.T) {
		cfg := &config.Config{
			Providers: csync.NewMap[string, config.ProviderConfig](),
			Options:   &config.Options{},
		}

		resp := ToConfigResponse(cfg, true)

		require.NotNil(t, resp.Schema, "Schema should be included when requested")
		require.NotEmpty(t, resp.Schema.Providers, "Provider schema should be present")
		require.NotEmpty(t, resp.Schema.Tools, "Tool schema should be present")
		require.NotEmpty(t, resp.Schema.UI, "UI schema should be present")
	})

	t.Run("excludes schema when not requested", func(t *testing.T) {
		cfg := &config.Config{
			Providers: csync.NewMap[string, config.ProviderConfig](),
			Options:   &config.Options{},
		}

		resp := ToConfigResponse(cfg, false)

		require.Nil(t, resp.Schema, "Schema should be nil when not requested")
	})

	t.Run("converts models correctly", func(t *testing.T) {
		cfg := &config.Config{
			Providers: csync.NewMap[string, config.ProviderConfig](),
			Models: map[config.SelectedModelType]config.SelectedModel{
				config.SelectedModelTypeLarge: {
					Provider: "openai",
					Model:    "gpt-4o",
				},
				config.SelectedModelTypeSmall: {
					Provider: "anthropic",
					Model:    "claude-3-haiku",
				},
			},
			Options: &config.Options{},
		}

		resp := ToConfigResponse(cfg, false)

		require.NotNil(t, resp.Models.Large)
		require.Equal(t, "openai", resp.Models.Large.Provider)
		require.Equal(t, "gpt-4o", resp.Models.Large.Model)

		require.NotNil(t, resp.Models.Small)
		require.Equal(t, "anthropic", resp.Models.Small.Provider)
		require.Equal(t, "claude-3-haiku", resp.Models.Small.Model)
	})

	t.Run("includes TUI options", func(t *testing.T) {
		cfg := &config.Config{
			Providers: csync.NewMap[string, config.ProviderConfig](),
			Options: &config.Options{
				TUI: &config.TUIOptions{
					CompactMode: true,
					DiffMode:    "split",
				},
			},
		}

		resp := ToConfigResponse(cfg, false)

		require.True(t, resp.UI.CompactMode)
		require.Equal(t, "split", resp.UI.DiffMode)
	})
}

func TestBuildSchemaDTO(t *testing.T) {
	t.Parallel()

	schema := buildSchemaDTO()

	t.Run("provider schema fields", func(t *testing.T) {
		require.Contains(t, schema.Providers, "id")
		require.Contains(t, schema.Providers, "api_key")

		apiKeyField := schema.Providers["api_key"]
		require.True(t, apiKeyField.Redacted, "API key field should be marked as redacted")
		require.Equal(t, "string", apiKeyField.Type)
	})

	t.Run("UI schema fields", func(t *testing.T) {
		require.Contains(t, schema.UI, "theme")

		themeField := schema.UI["theme"]
		require.Equal(t, "string", themeField.Type)
		require.Contains(t, themeField.Enum, "dark")
		require.Contains(t, themeField.Enum, "light")
	})

	t.Run("tools schema fields", func(t *testing.T) {
		require.Contains(t, schema.Tools, "bash.timeout")
		require.Contains(t, schema.Tools, "ls.max_depth")

		timeoutField := schema.Tools["bash.timeout"]
		require.Equal(t, "number", timeoutField.Type)
		require.NotNil(t, timeoutField.Minimum)
	})

	t.Run("models schema fields", func(t *testing.T) {
		require.Contains(t, schema.Models, "model")
		require.Contains(t, schema.Models, "provider")

		modelField := schema.Models["model"]
		require.True(t, modelField.Required, "Model field should be required")

		tempField := schema.Models["temperature"]
		require.NotNil(t, tempField.Minimum)
		require.NotNil(t, tempField.Maximum)
	})
}
