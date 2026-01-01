package api

import (
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigUpdate(t *testing.T) {
	t.Parallel()

	t.Run("valid models update", func(t *testing.T) {
		temp := 0.7
		req := &ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:       "gpt-4o",
					Provider:    "openai",
					Temperature: &temp,
				},
			},
		}

		err := ValidateConfigUpdate(req)
		require.NoError(t, err)
	})

	t.Run("invalid temperature", func(t *testing.T) {
		temp := 1.5
		req := &ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:       "gpt-4o",
					Provider:    "openai",
					Temperature: &temp,
				},
			},
		}

		err := ValidateConfigUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "temperature")
	})

	t.Run("missing model field", func(t *testing.T) {
		req := &ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Provider: "openai",
				},
			},
		}

		err := ValidateConfigUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "model field is required")
	})

	t.Run("invalid reasoning effort", func(t *testing.T) {
		req := &ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:           "o1-preview",
					Provider:        "openai",
					ReasoningEffort: "invalid",
				},
			},
		}

		err := ValidateConfigUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "reasoning_effort")
	})

	t.Run("valid UI update", func(t *testing.T) {
		theme := "dark"
		compactMode := true
		req := &ConfigUpdateRequest{
			UI: &UIUpdateRequest{
				Theme:       &theme,
				CompactMode: &compactMode,
			},
		}

		err := ValidateConfigUpdate(req)
		require.NoError(t, err)
	})

	t.Run("invalid theme", func(t *testing.T) {
		theme := "purple"
		req := &ConfigUpdateRequest{
			UI: &UIUpdateRequest{
				Theme: &theme,
			},
		}

		err := ValidateConfigUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "theme")
	})

	t.Run("valid tools update", func(t *testing.T) {
		maxDepth := 10
		maxItems := 100
		req := &ConfigUpdateRequest{
			Tools: &ToolsUpdateRequest{
				Ls: &LsToolUpdateDTO{
					MaxDepth: &maxDepth,
					MaxItems: &maxItems,
				},
			},
		}

		err := ValidateConfigUpdate(req)
		require.NoError(t, err)
	})

	t.Run("negative max_depth", func(t *testing.T) {
		maxDepth := -1
		req := &ConfigUpdateRequest{
			Tools: &ToolsUpdateRequest{
				Ls: &LsToolUpdateDTO{
					MaxDepth: &maxDepth,
				},
			},
		}

		err := ValidateConfigUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "max_depth")
	})

	t.Run("valid provider update", func(t *testing.T) {
		enabled := true
		baseURL := "https://api.openai.com/v1"
		req := &ConfigUpdateRequest{
			Providers: &ProvidersUpdateRequest{
				Updates: []ProviderUpdateDTO{
					{
						ID:      "openai",
						Enabled: &enabled,
						BaseURL: &baseURL,
					},
				},
			},
		}

		err := ValidateConfigUpdate(req)
		require.NoError(t, err)
	})

	t.Run("provider update missing ID", func(t *testing.T) {
		enabled := true
		req := &ConfigUpdateRequest{
			Providers: &ProvidersUpdateRequest{
				Updates: []ProviderUpdateDTO{
					{
						Enabled: &enabled,
					},
				},
			},
		}

		err := ValidateConfigUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "id is required")
	})
}

func TestValidateFieldUpdate(t *testing.T) {
	t.Parallel()

	t.Run("valid field path", func(t *testing.T) {
		req := &FieldUpdateRequest{
			Path:  "ui.compact_mode",
			Value: true,
		}

		err := ValidateFieldUpdate(req)
		require.NoError(t, err)
	})

	t.Run("invalid field path", func(t *testing.T) {
		req := &FieldUpdateRequest{
			Path:  "invalid.path",
			Value: true,
		}

		err := ValidateFieldUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown field path")
	})

	t.Run("invalid temperature value", func(t *testing.T) {
		req := &FieldUpdateRequest{
			Path:  "models.large.temperature",
			Value: 2.0,
		}

		err := ValidateFieldUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be between 0 and 1")
	})

	t.Run("invalid theme value", func(t *testing.T) {
		req := &FieldUpdateRequest{
			Path:  "ui.theme",
			Value: "purple",
		}

		err := ValidateFieldUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be 'dark' or 'light'")
	})

	t.Run("wrong type for boolean field", func(t *testing.T) {
		req := &FieldUpdateRequest{
			Path:  "ui.compact_mode",
			Value: "true",
		}

		err := ValidateFieldUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be a boolean")
	})

	t.Run("valid array value", func(t *testing.T) {
		req := &FieldUpdateRequest{
			Path:  "options.context_paths",
			Value: []interface{}{".cursorrules", "CRUSH.md"},
		}

		err := ValidateFieldUpdate(req)
		require.NoError(t, err)
	})

	t.Run("invalid array element type", func(t *testing.T) {
		req := &FieldUpdateRequest{
			Path:  "options.context_paths",
			Value: []interface{}{"valid", 123, "another"},
		}

		err := ValidateFieldUpdate(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be strings")
	})
}

func TestApplyConfigUpdate(t *testing.T) {
	t.Run("apply models update", func(t *testing.T) {
		cfg := &config.Config{
			Models: make(map[config.SelectedModelType]config.SelectedModel),
		}

		// Set up temp directory for config persistence
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		req := &ConfigUpdateRequest{
			Models: &ModelsUpdateRequest{
				Large: &ModelSelectionDTO{
					Model:    "gpt-4o",
					Provider: "openai",
				},
			},
		}

		err = ApplyConfigUpdate(cfg, req)
		require.NoError(t, err)

		model, ok := cfg.Models[config.SelectedModelTypeLarge]
		require.True(t, ok)
		require.Equal(t, "gpt-4o", model.Model)
		require.Equal(t, "openai", model.Provider)
	})

	t.Run("apply tools update", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		maxDepth := 10
		maxItems := 100
		req := &ConfigUpdateRequest{
			Tools: &ToolsUpdateRequest{
				Ls: &LsToolUpdateDTO{
					MaxDepth: &maxDepth,
					MaxItems: &maxItems,
				},
			},
		}

		err = ApplyConfigUpdate(cfg, req)
		require.NoError(t, err)

		depth, items := cfg.Tools.Ls.Limits()
		require.Equal(t, 10, depth)
		require.Equal(t, 100, items)
	})

	t.Run("apply UI update", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		compactMode := true
		diffMode := "split"
		req := &ConfigUpdateRequest{
			UI: &UIUpdateRequest{
				CompactMode: &compactMode,
				DiffMode:    &diffMode,
			},
		}

		err = ApplyConfigUpdate(cfg, req)
		require.NoError(t, err)

		require.True(t, cfg.Options.TUI.CompactMode)
		require.Equal(t, "split", cfg.Options.TUI.DiffMode)
	})

	t.Run("apply options update", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		contextPaths := []string{".cursorrules", "CRUSH.md"}
		disableMetrics := true
		req := &ConfigUpdateRequest{
			Options: &OptionsUpdateRequest{
				ContextPaths:   &contextPaths,
				DisableMetrics: &disableMetrics,
			},
		}

		err = ApplyConfigUpdate(cfg, req)
		require.NoError(t, err)

		require.Equal(t, contextPaths, cfg.Options.ContextPaths)
		require.True(t, cfg.Options.DisableMetrics)
	})
}

func TestApplyFieldUpdate(t *testing.T) {
	t.Run("update ls max_depth", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		req := &FieldUpdateRequest{
			Path:  "tools.ls.max_depth",
			Value: float64(15),
		}

		err = ApplyFieldUpdate(cfg, req)
		require.NoError(t, err)

		depth, _ := cfg.Tools.Ls.Limits()
		require.Equal(t, 15, depth)
	})

	t.Run("update compact mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		req := &FieldUpdateRequest{
			Path:  "ui.compact_mode",
			Value: true,
		}

		err = ApplyFieldUpdate(cfg, req)
		require.NoError(t, err)

		require.True(t, cfg.Options.TUI.CompactMode)
	})

	t.Run("update diff mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg, err := config.Init(tmpDir, tmpDir, false)
		require.NoError(t, err)

		req := &FieldUpdateRequest{
			Path:  "ui.diff_mode",
			Value: "split",
		}

		err = ApplyFieldUpdate(cfg, req)
		require.NoError(t, err)

		require.Equal(t, "split", cfg.Options.TUI.DiffMode)
	})
}
