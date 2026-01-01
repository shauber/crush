package api

import (
	"github.com/charmbracelet/crush/internal/config"
)

// SettingsResponse represents user-facing settings (subset of config focused on preferences).
type SettingsResponse struct {
	UI      UISettingsDTO      `json:"ui"`
	Editor  EditorSettingsDTO  `json:"editor"`
	Models  ModelsSettingsDTO  `json:"models"`
	General GeneralSettingsDTO `json:"general"`
}

// UISettingsDTO represents UI-related settings.
type UISettingsDTO struct {
	Theme              string `json:"theme"`
	SyntaxHighlighting bool   `json:"syntax_highlighting"`
	ShowCost           bool   `json:"show_cost"`
	CompactMode        bool   `json:"compact_mode"`
	DiffMode           string `json:"diff_mode"`
}

// EditorSettingsDTO represents editor-related settings.
type EditorSettingsDTO struct {
	ContextPaths []string `json:"context_paths"`
	SkillsPaths  []string `json:"skills_paths"`
}

// ModelsSettingsDTO represents model preference settings.
type ModelsSettingsDTO struct {
	DefaultLarge string `json:"default_large"` // Format: "provider/model"
	DefaultSmall string `json:"default_small"` // Format: "provider/model"
}

// GeneralSettingsDTO represents general application settings.
type GeneralSettingsDTO struct {
	Debug                bool `json:"debug"`
	DebugLSP             bool `json:"debug_lsp"`
	DisableAutoSummarize bool `json:"disable_auto_summarize"`
	DisableMetrics       bool `json:"disable_metrics"`
}

// SettingSchema describes a single setting's metadata.
type SettingSchema struct {
	Key         string      `json:"key"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Category    string      `json:"category"`
	Example     interface{} `json:"example,omitempty"`
}

// SettingUpdateRequest represents a request to update a single setting.
type SettingUpdateRequest struct {
	Value interface{} `json:"value"`
}

// SettingsUpdateRequest represents a request to update multiple settings.
type SettingsUpdateRequest struct {
	UI      *UISettingsDTO      `json:"ui,omitempty"`
	Editor  *EditorSettingsDTO  `json:"editor,omitempty"`
	Models  *ModelsSettingsDTO  `json:"models,omitempty"`
	General *GeneralSettingsDTO `json:"general,omitempty"`
}

// ToSettingsResponse converts config to settings response.
func ToSettingsResponse(cfg *config.Config) SettingsResponse {
	resp := SettingsResponse{
		UI:      toUISettings(cfg),
		Editor:  toEditorSettings(cfg),
		Models:  toModelsSettings(cfg),
		General: toGeneralSettings(cfg),
	}
	return resp
}

func toUISettings(cfg *config.Config) UISettingsDTO {
	dto := UISettingsDTO{
		Theme:              "dark",
		SyntaxHighlighting: true,
		ShowCost:           true,
		CompactMode:        false,
		DiffMode:           "unified",
	}

	if cfg.Options != nil && cfg.Options.TUI != nil {
		dto.CompactMode = cfg.Options.TUI.CompactMode
		if cfg.Options.TUI.DiffMode != "" {
			dto.DiffMode = cfg.Options.TUI.DiffMode
		}
	}

	return dto
}

func toEditorSettings(cfg *config.Config) EditorSettingsDTO {
	dto := EditorSettingsDTO{
		ContextPaths: []string{},
		SkillsPaths:  []string{},
	}

	if cfg.Options != nil {
		if cfg.Options.ContextPaths != nil {
			dto.ContextPaths = cfg.Options.ContextPaths
		}
		if cfg.Options.SkillsPaths != nil {
			dto.SkillsPaths = cfg.Options.SkillsPaths
		}
	}

	return dto
}

func toModelsSettings(cfg *config.Config) ModelsSettingsDTO {
	dto := ModelsSettingsDTO{
		DefaultLarge: "",
		DefaultSmall: "",
	}

	if large, ok := cfg.Models[config.SelectedModelTypeLarge]; ok {
		dto.DefaultLarge = large.Provider + "/" + large.Model
	}

	if small, ok := cfg.Models[config.SelectedModelTypeSmall]; ok {
		dto.DefaultSmall = small.Provider + "/" + small.Model
	}

	return dto
}

func toGeneralSettings(cfg *config.Config) GeneralSettingsDTO {
	dto := GeneralSettingsDTO{
		Debug:                false,
		DebugLSP:             false,
		DisableAutoSummarize: false,
		DisableMetrics:       false,
	}

	if cfg.Options != nil {
		dto.Debug = cfg.Options.Debug
		dto.DebugLSP = cfg.Options.DebugLSP
		dto.DisableAutoSummarize = cfg.Options.DisableAutoSummarize
		dto.DisableMetrics = cfg.Options.DisableMetrics
	}

	return dto
}

// GetAllSettingsSchemas returns schemas for all settings.
func GetAllSettingsSchemas() []SettingSchema {
	return []SettingSchema{
		// UI Settings
		{
			Key:         "ui.theme",
			Type:        "string",
			Description: "Color theme for the user interface",
			Default:     "dark",
			Enum:        []string{"dark", "light"},
			Category:    "ui",
			Example:     "dark",
		},
		{
			Key:         "ui.syntax_highlighting",
			Type:        "boolean",
			Description: "Enable syntax highlighting in code blocks",
			Default:     true,
			Category:    "ui",
		},
		{
			Key:         "ui.show_cost",
			Type:        "boolean",
			Description: "Display estimated cost of API calls",
			Default:     true,
			Category:    "ui",
		},
		{
			Key:         "ui.compact_mode",
			Type:        "boolean",
			Description: "Use compact UI layout with reduced spacing",
			Default:     false,
			Category:    "ui",
		},
		{
			Key:         "ui.diff_mode",
			Type:        "string",
			Description: "Diff view mode for file changes",
			Default:     "unified",
			Enum:        []string{"unified", "split"},
			Category:    "ui",
			Example:     "unified",
		},

		// Editor Settings
		{
			Key:         "editor.context_paths",
			Type:        "array",
			Description: "Paths to files containing context for the AI",
			Default:     []string{},
			Category:    "editor",
			Example:     []string{".cursorrules", "CRUSH.md"},
		},
		{
			Key:         "editor.skills_paths",
			Type:        "array",
			Description: "Paths to directories containing Agent Skills",
			Default:     []string{},
			Category:    "editor",
			Example:     []string{"~/.config/crush/skills"},
		},

		// Model Settings
		{
			Key:         "models.default_large",
			Type:        "string",
			Description: "Default large model (format: provider/model)",
			Default:     "",
			Category:    "models",
			Example:     "openai/gpt-4o",
		},
		{
			Key:         "models.default_small",
			Type:        "string",
			Description: "Default small model (format: provider/model)",
			Default:     "",
			Category:    "models",
			Example:     "anthropic/claude-3-haiku",
		},

		// General Settings
		{
			Key:         "general.debug",
			Type:        "boolean",
			Description: "Enable debug logging",
			Default:     false,
			Category:    "general",
		},
		{
			Key:         "general.debug_lsp",
			Type:        "boolean",
			Description: "Enable debug logging for LSP servers",
			Default:     false,
			Category:    "general",
		},
		{
			Key:         "general.disable_auto_summarize",
			Type:        "boolean",
			Description: "Disable automatic conversation summarization",
			Default:     false,
			Category:    "general",
		},
		{
			Key:         "general.disable_metrics",
			Type:        "boolean",
			Description: "Disable sending usage metrics",
			Default:     false,
			Category:    "general",
		},
	}
}

// GetSettingSchema returns the schema for a specific setting key.
func GetSettingSchema(key string) *SettingSchema {
	schemas := GetAllSettingsSchemas()
	for _, schema := range schemas {
		if schema.Key == key {
			return &schema
		}
	}
	return nil
}

// ValidateSettingValue validates a setting value against its schema.
func ValidateSettingValue(key string, value interface{}) error {
	schema := GetSettingSchema(key)
	if schema == nil {
		return nil // Unknown keys are allowed (forward compatibility)
	}

	// Reuse field validation from config_update.go
	return validateFieldValue(key, value)
}

// ApplySettingsUpdate applies a settings update to the config.
func ApplySettingsUpdate(cfg *config.Config, req *SettingsUpdateRequest) error {
	if req.UI != nil {
		if err := applyUISettings(cfg, req.UI); err != nil {
			return err
		}
	}

	if req.Editor != nil {
		if err := applyEditorSettings(cfg, req.Editor); err != nil {
			return err
		}
	}

	if req.Models != nil {
		if err := applyModelsSettings(cfg, req.Models); err != nil {
			return err
		}
	}

	if req.General != nil {
		if err := applyGeneralSettings(cfg, req.General); err != nil {
			return err
		}
	}

	return nil
}

func applyUISettings(cfg *config.Config, settings *UISettingsDTO) error {
	if cfg.Options == nil {
		cfg.Options = &config.Options{}
	}
	if cfg.Options.TUI == nil {
		cfg.Options.TUI = &config.TUIOptions{}
	}

	// Update compact mode
	if err := cfg.SetCompactMode(settings.CompactMode); err != nil {
		return err
	}

	// Update diff mode
	cfg.Options.TUI.DiffMode = settings.DiffMode
	if err := cfg.SetConfigField("options.tui.diff_mode", settings.DiffMode); err != nil {
		return err
	}

	// Note: theme, syntax_highlighting, show_cost are UI-only and not persisted in config
	// These would typically be stored in a separate UI preferences store

	return nil
}

func applyEditorSettings(cfg *config.Config, settings *EditorSettingsDTO) error {
	if cfg.Options == nil {
		cfg.Options = &config.Options{}
	}

	cfg.Options.ContextPaths = settings.ContextPaths
	if err := cfg.SetConfigField("options.context_paths", settings.ContextPaths); err != nil {
		return err
	}

	cfg.Options.SkillsPaths = settings.SkillsPaths
	if err := cfg.SetConfigField("options.skills_paths", settings.SkillsPaths); err != nil {
		return err
	}

	return nil
}

func applyModelsSettings(cfg *config.Config, settings *ModelsSettingsDTO) error {
	// Parse provider/model format
	if settings.DefaultLarge != "" {
		if err := setModelFromString(cfg, config.SelectedModelTypeLarge, settings.DefaultLarge); err != nil {
			return err
		}
	}

	if settings.DefaultSmall != "" {
		if err := setModelFromString(cfg, config.SelectedModelTypeSmall, settings.DefaultSmall); err != nil {
			return err
		}
	}

	return nil
}

func applyGeneralSettings(cfg *config.Config, settings *GeneralSettingsDTO) error {
	if cfg.Options == nil {
		cfg.Options = &config.Options{}
	}

	cfg.Options.Debug = settings.Debug
	if err := cfg.SetConfigField("options.debug", settings.Debug); err != nil {
		return err
	}

	cfg.Options.DebugLSP = settings.DebugLSP
	if err := cfg.SetConfigField("options.debug_lsp", settings.DebugLSP); err != nil {
		return err
	}

	cfg.Options.DisableAutoSummarize = settings.DisableAutoSummarize
	if err := cfg.SetConfigField("options.disable_auto_summarize", settings.DisableAutoSummarize); err != nil {
		return err
	}

	cfg.Options.DisableMetrics = settings.DisableMetrics
	if err := cfg.SetConfigField("options.disable_metrics", settings.DisableMetrics); err != nil {
		return err
	}

	return nil
}

func setModelFromString(cfg *config.Config, modelType config.SelectedModelType, modelStr string) error {
	// Parse "provider/model" format
	parts := []rune(modelStr)
	slashIdx := -1
	for i, r := range parts {
		if r == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx == -1 {
		return nil // Invalid format, skip
	}

	provider := string(parts[:slashIdx])
	model := string(parts[slashIdx+1:])

	selectedModel := config.SelectedModel{
		Provider: provider,
		Model:    model,
	}

	return cfg.UpdatePreferredModel(modelType, selectedModel)
}

// ApplySettingUpdate applies a single setting update.
func ApplySettingUpdate(cfg *config.Config, key string, value interface{}) error {
	// Map setting keys to config field paths
	switch key {
	case "ui.compact_mode":
		return cfg.SetCompactMode(value.(bool))
	case "ui.diff_mode":
		if cfg.Options == nil || cfg.Options.TUI == nil {
			cfg.Options = &config.Options{TUI: &config.TUIOptions{}}
		}
		cfg.Options.TUI.DiffMode = value.(string)
		return cfg.SetConfigField("options.tui.diff_mode", value)
	case "editor.context_paths":
		if cfg.Options == nil {
			cfg.Options = &config.Options{}
		}
		paths := []string{}
		if arr, ok := value.([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					paths = append(paths, s)
				}
			}
		}
		cfg.Options.ContextPaths = paths
		return cfg.SetConfigField("options.context_paths", paths)
	case "editor.skills_paths":
		if cfg.Options == nil {
			cfg.Options = &config.Options{}
		}
		paths := []string{}
		if arr, ok := value.([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					paths = append(paths, s)
				}
			}
		}
		cfg.Options.SkillsPaths = paths
		return cfg.SetConfigField("options.skills_paths", paths)
	case "general.debug":
		if cfg.Options == nil {
			cfg.Options = &config.Options{}
		}
		cfg.Options.Debug = value.(bool)
		return cfg.SetConfigField("options.debug", value)
	case "general.debug_lsp":
		if cfg.Options == nil {
			cfg.Options = &config.Options{}
		}
		cfg.Options.DebugLSP = value.(bool)
		return cfg.SetConfigField("options.debug_lsp", value)
	case "general.disable_auto_summarize":
		if cfg.Options == nil {
			cfg.Options = &config.Options{}
		}
		cfg.Options.DisableAutoSummarize = value.(bool)
		return cfg.SetConfigField("options.disable_auto_summarize", value)
	case "general.disable_metrics":
		if cfg.Options == nil {
			cfg.Options = &config.Options{}
		}
		cfg.Options.DisableMetrics = value.(bool)
		return cfg.SetConfigField("options.disable_metrics", value)
	case "models.default_large":
		return setModelFromString(cfg, config.SelectedModelTypeLarge, value.(string))
	case "models.default_small":
		return setModelFromString(cfg, config.SelectedModelTypeSmall, value.(string))
	default:
		// UI-only settings (not persisted to config file)
		return nil
	}
}
