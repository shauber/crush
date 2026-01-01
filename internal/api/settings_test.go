package api

import (
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/csync"
	"github.com/stretchr/testify/require"
)

func TestToSettingsResponse(t *testing.T) {
	t.Parallel()

	t.Run("converts config to settings", func(t *testing.T) {
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
			Options: &config.Options{
				TUI: &config.TUIOptions{
					CompactMode: true,
					DiffMode:    "split",
				},
				ContextPaths:         []string{".cursorrules"},
				Debug:                true,
				DisableAutoSummarize: false,
				DisableMetrics:       true,
			},
		}

		resp := ToSettingsResponse(cfg)

		// Check UI settings
		require.True(t, resp.UI.CompactMode)
		require.Equal(t, "split", resp.UI.DiffMode)

		// Check editor settings
		require.Equal(t, []string{".cursorrules"}, resp.Editor.ContextPaths)

		// Check model settings
		require.Equal(t, "openai/gpt-4o", resp.Models.DefaultLarge)
		require.Equal(t, "anthropic/claude-3-haiku", resp.Models.DefaultSmall)

		// Check general settings
		require.True(t, resp.General.Debug)
		require.False(t, resp.General.DisableAutoSummarize)
		require.True(t, resp.General.DisableMetrics)
	})

	t.Run("handles nil options", func(t *testing.T) {
		cfg := &config.Config{
			Providers: csync.NewMap[string, config.ProviderConfig](),
			Models:    make(map[config.SelectedModelType]config.SelectedModel),
			Options:   nil,
		}

		resp := ToSettingsResponse(cfg)

		// Should use defaults
		require.False(t, resp.UI.CompactMode)
		require.Equal(t, "unified", resp.UI.DiffMode)
		require.Empty(t, resp.Editor.ContextPaths)
		require.False(t, resp.General.Debug)
	})
}

func TestGetAllSettingsSchemas(t *testing.T) {
	t.Parallel()

	schemas := GetAllSettingsSchemas()

	require.NotEmpty(t, schemas)
	require.GreaterOrEqual(t, len(schemas), 10, "Should have at least 10 settings")

	// Check for key schemas
	foundTheme := false
	foundCompactMode := false
	foundDebug := false

	for _, schema := range schemas {
		require.NotEmpty(t, schema.Key)
		require.NotEmpty(t, schema.Type)
		require.NotEmpty(t, schema.Description)
		require.NotEmpty(t, schema.Category)

		switch schema.Key {
		case "ui.theme":
			foundTheme = true
			require.Equal(t, "string", schema.Type)
			require.Contains(t, schema.Enum, "dark")
			require.Contains(t, schema.Enum, "light")
		case "ui.compact_mode":
			foundCompactMode = true
			require.Equal(t, "boolean", schema.Type)
		case "general.debug":
			foundDebug = true
			require.Equal(t, "boolean", schema.Type)
		}
	}

	require.True(t, foundTheme, "Should have theme schema")
	require.True(t, foundCompactMode, "Should have compact_mode schema")
	require.True(t, foundDebug, "Should have debug schema")
}

func TestGetSettingSchema(t *testing.T) {
	t.Parallel()

	t.Run("returns schema for valid key", func(t *testing.T) {
		schema := GetSettingSchema("ui.theme")
		require.NotNil(t, schema)
		require.Equal(t, "ui.theme", schema.Key)
		require.Equal(t, "string", schema.Type)
	})

	t.Run("returns nil for invalid key", func(t *testing.T) {
		schema := GetSettingSchema("invalid.key")
		require.Nil(t, schema)
	})
}

func TestApplySettingsUpdate(t *testing.T) {
	t.Run("update UI settings", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		req := &SettingsUpdateRequest{
			UI: &UISettingsDTO{
				CompactMode: true,
				DiffMode:    "split",
			},
		}

		err = ApplySettingsUpdate(cfg, req)
		require.NoError(t, err)

		require.True(t, cfg.Options.TUI.CompactMode)
		require.Equal(t, "split", cfg.Options.TUI.DiffMode)
	})

	t.Run("update editor settings", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		contextPaths := []string{".cursorrules", "CRUSH.md"}
		skillsPaths := []string{"~/.config/crush/skills"}

		req := &SettingsUpdateRequest{
			Editor: &EditorSettingsDTO{
				ContextPaths: contextPaths,
				SkillsPaths:  skillsPaths,
			},
		}

		err = ApplySettingsUpdate(cfg, req)
		require.NoError(t, err)

		require.Equal(t, contextPaths, cfg.Options.ContextPaths)
		require.Equal(t, skillsPaths, cfg.Options.SkillsPaths)
	})

	t.Run("update models settings", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		req := &SettingsUpdateRequest{
			Models: &ModelsSettingsDTO{
				DefaultLarge: "openai/gpt-4o",
				DefaultSmall: "anthropic/claude-3-haiku",
			},
		}

		err = ApplySettingsUpdate(cfg, req)
		require.NoError(t, err)

		large, ok := cfg.Models[config.SelectedModelTypeLarge]
		require.True(t, ok)
		require.Equal(t, "openai", large.Provider)
		require.Equal(t, "gpt-4o", large.Model)

		small, ok := cfg.Models[config.SelectedModelTypeSmall]
		require.True(t, ok)
		require.Equal(t, "anthropic", small.Provider)
		require.Equal(t, "claude-3-haiku", small.Model)
	})

	t.Run("update general settings", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		req := &SettingsUpdateRequest{
			General: &GeneralSettingsDTO{
				Debug:                true,
				DebugLSP:             false,
				DisableAutoSummarize: true,
				DisableMetrics:       false,
			},
		}

		err = ApplySettingsUpdate(cfg, req)
		require.NoError(t, err)

		require.True(t, cfg.Options.Debug)
		require.False(t, cfg.Options.DebugLSP)
		require.True(t, cfg.Options.DisableAutoSummarize)
		require.False(t, cfg.Options.DisableMetrics)
	})

	t.Run("update multiple sections", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		req := &SettingsUpdateRequest{
			UI: &UISettingsDTO{
				CompactMode: true,
			},
			General: &GeneralSettingsDTO{
				Debug: true,
			},
		}

		err = ApplySettingsUpdate(cfg, req)
		require.NoError(t, err)

		require.True(t, cfg.Options.TUI.CompactMode)
		require.True(t, cfg.Options.Debug)
	})
}

func TestApplySettingUpdate(t *testing.T) {
	t.Run("update compact mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		err = ApplySettingUpdate(cfg, "ui.compact_mode", true)
		require.NoError(t, err)

		require.True(t, cfg.Options.TUI.CompactMode)
	})

	t.Run("update diff mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		err = ApplySettingUpdate(cfg, "ui.diff_mode", "split")
		require.NoError(t, err)

		require.Equal(t, "split", cfg.Options.TUI.DiffMode)
	})

	t.Run("update context paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		paths := []interface{}{".cursorrules", "CRUSH.md"}
		err = ApplySettingUpdate(cfg, "editor.context_paths", paths)
		require.NoError(t, err)

		require.Len(t, cfg.Options.ContextPaths, 2)
		require.Contains(t, cfg.Options.ContextPaths, ".cursorrules")
		require.Contains(t, cfg.Options.ContextPaths, "CRUSH.md")
	})

	t.Run("update debug flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		err = ApplySettingUpdate(cfg, "general.debug", true)
		require.NoError(t, err)

		require.True(t, cfg.Options.Debug)
	})

	t.Run("update model settings", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		err = ApplySettingUpdate(cfg, "models.default_large", "openai/gpt-4o")
		require.NoError(t, err)

		large, ok := cfg.Models[config.SelectedModelTypeLarge]
		require.True(t, ok)
		require.Equal(t, "openai", large.Provider)
		require.Equal(t, "gpt-4o", large.Model)
	})
}

func TestSetModelFromString(t *testing.T) {
	t.Run("valid provider/model format", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		err = setModelFromString(cfg, config.SelectedModelTypeLarge, "openai/gpt-4o")
		require.NoError(t, err)

		model, ok := cfg.Models[config.SelectedModelTypeLarge]
		require.True(t, ok)
		require.Equal(t, "openai", model.Provider)
		require.Equal(t, "gpt-4o", model.Model)
	})

	t.Run("invalid format without slash", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		err = setModelFromString(cfg, config.SelectedModelTypeLarge, "invalid")
		require.NoError(t, err) // Should not error, just skip
	})
}
