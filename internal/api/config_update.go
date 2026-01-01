package api

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/crush/internal/config"
)

// ConfigUpdateRequest represents a partial configuration update.
// Only non-nil fields will be updated.
type ConfigUpdateRequest struct {
	Models    *ModelsUpdateRequest    `json:"models,omitempty"`
	Providers *ProvidersUpdateRequest `json:"providers,omitempty"`
	Tools     *ToolsUpdateRequest     `json:"tools,omitempty"`
	UI        *UIUpdateRequest        `json:"ui,omitempty"`
	Options   *OptionsUpdateRequest   `json:"options,omitempty"`
}

// ModelsUpdateRequest represents model configuration updates.
type ModelsUpdateRequest struct {
	Large *ModelSelectionDTO `json:"large,omitempty"`
	Small *ModelSelectionDTO `json:"small,omitempty"`
}

// ProvidersUpdateRequest represents provider configuration updates.
type ProvidersUpdateRequest struct {
	Updates []ProviderUpdateDTO `json:"updates,omitempty"`
}

// ProviderUpdateDTO represents a single provider update.
type ProviderUpdateDTO struct {
	ID      string  `json:"id"`
	Enabled *bool   `json:"enabled,omitempty"`
	BaseURL *string `json:"base_url,omitempty"`
	APIKey  *string `json:"api_key,omitempty"` // Can be set but never returned
}

// ToolsUpdateRequest represents tool configuration updates.
type ToolsUpdateRequest struct {
	Ls            *LsToolUpdateDTO `json:"ls,omitempty"`
	DisabledTools *[]string        `json:"disabled_tools,omitempty"`
}

// LsToolUpdateDTO represents ls tool updates.
type LsToolUpdateDTO struct {
	MaxDepth *int `json:"max_depth,omitempty"`
	MaxItems *int `json:"max_items,omitempty"`
}

// UIUpdateRequest represents UI preference updates.
type UIUpdateRequest struct {
	Theme       *string `json:"theme,omitempty"`
	CompactMode *bool   `json:"compact_mode,omitempty"`
	DiffMode    *string `json:"diff_mode,omitempty"`
}

// OptionsUpdateRequest represents general options updates.
type OptionsUpdateRequest struct {
	ContextPaths         *[]string `json:"context_paths,omitempty"`
	SkillsPaths          *[]string `json:"skills_paths,omitempty"`
	DisableAutoSummarize *bool     `json:"disable_auto_summarize,omitempty"`
	DisableMetrics       *bool     `json:"disable_metrics,omitempty"`
	Debug                *bool     `json:"debug,omitempty"`
}

// FieldUpdateRequest represents a single field update using dot notation.
type FieldUpdateRequest struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// ValidateConfigUpdate validates a configuration update request.
func ValidateConfigUpdate(req *ConfigUpdateRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate models if present
	if req.Models != nil {
		if err := validateModelsUpdate(req.Models); err != nil {
			return fmt.Errorf("models validation failed: %w", err)
		}
	}

	// Validate providers if present
	if req.Providers != nil {
		if err := validateProvidersUpdate(req.Providers); err != nil {
			return fmt.Errorf("providers validation failed: %w", err)
		}
	}

	// Validate tools if present
	if req.Tools != nil {
		if err := validateToolsUpdate(req.Tools); err != nil {
			return fmt.Errorf("tools validation failed: %w", err)
		}
	}

	// Validate UI if present
	if req.UI != nil {
		if err := validateUIUpdate(req.UI); err != nil {
			return fmt.Errorf("ui validation failed: %w", err)
		}
	}

	// Validate options if present
	if req.Options != nil {
		if err := validateOptionsUpdate(req.Options); err != nil {
			return fmt.Errorf("options validation failed: %w", err)
		}
	}

	return nil
}

func validateModelsUpdate(req *ModelsUpdateRequest) error {
	if req.Large != nil {
		if req.Large.Model == "" {
			return fmt.Errorf("large model: model field is required")
		}
		if req.Large.Provider == "" {
			return fmt.Errorf("large model: provider field is required")
		}
		if err := validateModelSelection(req.Large); err != nil {
			return fmt.Errorf("large model: %w", err)
		}
	}

	if req.Small != nil {
		if req.Small.Model == "" {
			return fmt.Errorf("small model: model field is required")
		}
		if req.Small.Provider == "" {
			return fmt.Errorf("small model: provider field is required")
		}
		if err := validateModelSelection(req.Small); err != nil {
			return fmt.Errorf("small model: %w", err)
		}
	}

	return nil
}

func validateModelSelection(model *ModelSelectionDTO) error {
	if model.Temperature != nil {
		if *model.Temperature < 0 || *model.Temperature > 1 {
			return fmt.Errorf("temperature must be between 0 and 1")
		}
	}

	if model.MaxTokens < 0 {
		return fmt.Errorf("max_tokens must be non-negative")
	}

	if model.MaxTokens > 200000 {
		return fmt.Errorf("max_tokens must not exceed 200000")
	}

	if model.ReasoningEffort != "" {
		validEfforts := []string{"low", "medium", "high"}
		valid := false
		for _, effort := range validEfforts {
			if model.ReasoningEffort == effort {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("reasoning_effort must be one of: low, medium, high")
		}
	}

	return nil
}

func validateProvidersUpdate(req *ProvidersUpdateRequest) error {
	if req.Updates == nil || len(req.Updates) == 0 {
		return fmt.Errorf("at least one provider update is required")
	}

	for i, update := range req.Updates {
		if update.ID == "" {
			return fmt.Errorf("provider %d: id is required", i)
		}
	}

	return nil
}

func validateToolsUpdate(req *ToolsUpdateRequest) error {
	if req.Ls != nil {
		if req.Ls.MaxDepth != nil && *req.Ls.MaxDepth < 0 {
			return fmt.Errorf("ls.max_depth must be non-negative")
		}
		if req.Ls.MaxItems != nil && *req.Ls.MaxItems < 0 {
			return fmt.Errorf("ls.max_items must be non-negative")
		}
	}

	return nil
}

func validateUIUpdate(req *UIUpdateRequest) error {
	if req.Theme != nil {
		validThemes := []string{"dark", "light"}
		valid := false
		for _, theme := range validThemes {
			if *req.Theme == theme {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("theme must be one of: dark, light")
		}
	}

	if req.DiffMode != nil {
		validModes := []string{"unified", "split"}
		valid := false
		for _, mode := range validModes {
			if *req.DiffMode == mode {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("diff_mode must be one of: unified, split")
		}
	}

	return nil
}

func validateOptionsUpdate(req *OptionsUpdateRequest) error {
	// All options fields are valid as-is
	// Could add specific validation for paths if needed
	return nil
}

// ValidateFieldUpdate validates a field update request.
func ValidateFieldUpdate(req *FieldUpdateRequest) error {
	if req.Path == "" {
		return fmt.Errorf("path is required")
	}

	// Validate that the path is a known field
	if !isValidFieldPath(req.Path) {
		return fmt.Errorf("unknown field path: %s", req.Path)
	}

	// Type validation based on path
	if err := validateFieldValue(req.Path, req.Value); err != nil {
		return fmt.Errorf("invalid value for %s: %w", req.Path, err)
	}

	return nil
}

func isValidFieldPath(path string) bool {
	validPaths := []string{
		// Models
		"models.large.model",
		"models.large.provider",
		"models.large.max_tokens",
		"models.large.temperature",
		"models.large.reasoning_effort",
		"models.large.think",
		"models.small.model",
		"models.small.provider",
		"models.small.max_tokens",
		"models.small.temperature",
		"models.small.reasoning_effort",
		"models.small.think",
		// Tools
		"tools.ls.max_depth",
		"tools.ls.max_items",
		"tools.disabled_tools",
		// UI
		"ui.theme",
		"ui.compact_mode",
		"ui.diff_mode",
		// Options
		"options.context_paths",
		"options.skills_paths",
		"options.disable_auto_summarize",
		"options.disable_metrics",
		"options.debug",
	}

	for _, valid := range validPaths {
		if path == valid {
			return true
		}
	}

	return false
}

func validateFieldValue(path string, value interface{}) error {
	// Type checking based on path
	switch path {
	case "models.large.temperature", "models.small.temperature":
		if v, ok := value.(float64); ok {
			if v < 0 || v > 1 {
				return fmt.Errorf("must be between 0 and 1")
			}
		} else {
			return fmt.Errorf("must be a number")
		}

	case "models.large.max_tokens", "models.small.max_tokens":
		if v, ok := value.(float64); ok {
			if v < 0 || v > 200000 {
				return fmt.Errorf("must be between 0 and 200000")
			}
		} else {
			return fmt.Errorf("must be a number")
		}

	case "tools.ls.max_depth", "tools.ls.max_items":
		if v, ok := value.(float64); ok {
			if v < 0 {
				return fmt.Errorf("must be non-negative")
			}
		} else {
			return fmt.Errorf("must be a number")
		}

	case "ui.theme":
		if v, ok := value.(string); ok {
			if v != "dark" && v != "light" {
				return fmt.Errorf("must be 'dark' or 'light'")
			}
		} else {
			return fmt.Errorf("must be a string")
		}

	case "ui.diff_mode":
		if v, ok := value.(string); ok {
			if v != "unified" && v != "split" {
				return fmt.Errorf("must be 'unified' or 'split'")
			}
		} else {
			return fmt.Errorf("must be a string")
		}

	case "ui.compact_mode", "options.disable_auto_summarize", "options.disable_metrics", "options.debug", "models.large.think", "models.small.think":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("must be a boolean")
		}

	case "models.large.model", "models.small.model", "models.large.provider", "models.small.provider", "models.large.reasoning_effort", "models.small.reasoning_effort":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("must be a string")
		}

	case "options.context_paths", "options.skills_paths", "tools.disabled_tools":
		// Check if it's an array
		switch v := value.(type) {
		case []interface{}:
			// Validate each element is a string
			for _, item := range v {
				if _, ok := item.(string); !ok {
					return fmt.Errorf("array elements must be strings")
				}
			}
		case []string:
			// Already valid
		default:
			return fmt.Errorf("must be an array of strings")
		}
	}

	return nil
}

// ApplyConfigUpdate applies a configuration update to the config.
func ApplyConfigUpdate(cfg *config.Config, req *ConfigUpdateRequest) error {
	// Apply models updates
	if req.Models != nil {
		if err := applyModelsUpdate(cfg, req.Models); err != nil {
			return fmt.Errorf("failed to apply models update: %w", err)
		}
	}

	// Apply providers updates
	if req.Providers != nil {
		if err := applyProvidersUpdate(cfg, req.Providers); err != nil {
			return fmt.Errorf("failed to apply providers update: %w", err)
		}
	}

	// Apply tools updates
	if req.Tools != nil {
		if err := applyToolsUpdate(cfg, req.Tools); err != nil {
			return fmt.Errorf("failed to apply tools update: %w", err)
		}
	}

	// Apply UI updates
	if req.UI != nil {
		if err := applyUIUpdate(cfg, req.UI); err != nil {
			return fmt.Errorf("failed to apply UI update: %w", err)
		}
	}

	// Apply options updates
	if req.Options != nil {
		if err := applyOptionsUpdate(cfg, req.Options); err != nil {
			return fmt.Errorf("failed to apply options update: %w", err)
		}
	}

	return nil
}

func applyModelsUpdate(cfg *config.Config, req *ModelsUpdateRequest) error {
	if req.Large != nil {
		model := config.SelectedModel{
			Provider:        req.Large.Provider,
			Model:           req.Large.Model,
			MaxTokens:       req.Large.MaxTokens,
			Temperature:     req.Large.Temperature,
			ReasoningEffort: req.Large.ReasoningEffort,
			Think:           req.Large.Think,
		}
		if err := cfg.UpdatePreferredModel(config.SelectedModelTypeLarge, model); err != nil {
			return fmt.Errorf("failed to update large model: %w", err)
		}
	}

	if req.Small != nil {
		model := config.SelectedModel{
			Provider:        req.Small.Provider,
			Model:           req.Small.Model,
			MaxTokens:       req.Small.MaxTokens,
			Temperature:     req.Small.Temperature,
			ReasoningEffort: req.Small.ReasoningEffort,
			Think:           req.Small.Think,
		}
		if err := cfg.UpdatePreferredModel(config.SelectedModelTypeSmall, model); err != nil {
			return fmt.Errorf("failed to update small model: %w", err)
		}
	}

	return nil
}

func applyProvidersUpdate(cfg *config.Config, req *ProvidersUpdateRequest) error {
	for _, update := range req.Updates {
		provider, exists := cfg.Providers.Get(update.ID)
		if !exists {
			return fmt.Errorf("provider %s not found", update.ID)
		}

		if update.Enabled != nil {
			provider.Disable = !*update.Enabled
		}

		if update.BaseURL != nil {
			provider.BaseURL = *update.BaseURL
		}

		if update.APIKey != nil {
			if err := cfg.SetProviderAPIKey(update.ID, *update.APIKey); err != nil {
				return fmt.Errorf("failed to set API key for %s: %w", update.ID, err)
			}
			// Refresh the provider after API key update
			provider, _ = cfg.Providers.Get(update.ID)
		}

		cfg.Providers.Set(update.ID, provider)

		// Persist provider updates
		if update.Enabled != nil {
			if err := cfg.SetConfigField(fmt.Sprintf("providers.%s.disable", update.ID), !*update.Enabled); err != nil {
				return fmt.Errorf("failed to persist provider.disable: %w", err)
			}
		}
		if update.BaseURL != nil {
			if err := cfg.SetConfigField(fmt.Sprintf("providers.%s.base_url", update.ID), *update.BaseURL); err != nil {
				return fmt.Errorf("failed to persist provider.base_url: %w", err)
			}
		}
	}

	return nil
}

func applyToolsUpdate(cfg *config.Config, req *ToolsUpdateRequest) error {
	if req.Ls != nil {
		if req.Ls.MaxDepth != nil {
			cfg.Tools.Ls.MaxDepth = req.Ls.MaxDepth
			if err := cfg.SetConfigField("tools.ls.max_depth", *req.Ls.MaxDepth); err != nil {
				return fmt.Errorf("failed to persist tools.ls.max_depth: %w", err)
			}
		}
		if req.Ls.MaxItems != nil {
			cfg.Tools.Ls.MaxItems = req.Ls.MaxItems
			if err := cfg.SetConfigField("tools.ls.max_items", *req.Ls.MaxItems); err != nil {
				return fmt.Errorf("failed to persist tools.ls.max_items: %w", err)
			}
		}
	}

	if req.DisabledTools != nil {
		if cfg.Options == nil {
			cfg.Options = &config.Options{}
		}
		cfg.Options.DisabledTools = *req.DisabledTools
		if err := cfg.SetConfigField("options.disabled_tools", *req.DisabledTools); err != nil {
			return fmt.Errorf("failed to persist options.disabled_tools: %w", err)
		}
	}

	return nil
}

func applyUIUpdate(cfg *config.Config, req *UIUpdateRequest) error {
	if cfg.Options == nil {
		cfg.Options = &config.Options{}
	}
	if cfg.Options.TUI == nil {
		cfg.Options.TUI = &config.TUIOptions{}
	}

	if req.CompactMode != nil {
		if err := cfg.SetCompactMode(*req.CompactMode); err != nil {
			return fmt.Errorf("failed to set compact mode: %w", err)
		}
	}

	if req.DiffMode != nil {
		cfg.Options.TUI.DiffMode = *req.DiffMode
		if err := cfg.SetConfigField("options.tui.diff_mode", *req.DiffMode); err != nil {
			return fmt.Errorf("failed to persist ui.diff_mode: %w", err)
		}
	}

	// Note: theme is not currently persisted in config struct
	// Would need to be added to config.TUIOptions if needed

	return nil
}

func applyOptionsUpdate(cfg *config.Config, req *OptionsUpdateRequest) error {
	if cfg.Options == nil {
		cfg.Options = &config.Options{}
	}

	if req.ContextPaths != nil {
		cfg.Options.ContextPaths = *req.ContextPaths
		if err := cfg.SetConfigField("options.context_paths", *req.ContextPaths); err != nil {
			return fmt.Errorf("failed to persist options.context_paths: %w", err)
		}
	}

	if req.SkillsPaths != nil {
		cfg.Options.SkillsPaths = *req.SkillsPaths
		if err := cfg.SetConfigField("options.skills_paths", *req.SkillsPaths); err != nil {
			return fmt.Errorf("failed to persist options.skills_paths: %w", err)
		}
	}

	if req.DisableAutoSummarize != nil {
		cfg.Options.DisableAutoSummarize = *req.DisableAutoSummarize
		if err := cfg.SetConfigField("options.disable_auto_summarize", *req.DisableAutoSummarize); err != nil {
			return fmt.Errorf("failed to persist options.disable_auto_summarize: %w", err)
		}
	}

	if req.DisableMetrics != nil {
		cfg.Options.DisableMetrics = *req.DisableMetrics
		if err := cfg.SetConfigField("options.disable_metrics", *req.DisableMetrics); err != nil {
			return fmt.Errorf("failed to persist options.disable_metrics: %w", err)
		}
	}

	if req.Debug != nil {
		cfg.Options.Debug = *req.Debug
		if err := cfg.SetConfigField("options.debug", *req.Debug); err != nil {
			return fmt.Errorf("failed to persist options.debug: %w", err)
		}
	}

	return nil
}

// ApplyFieldUpdate applies a single field update to the config.
func ApplyFieldUpdate(cfg *config.Config, req *FieldUpdateRequest) error {
	// Convert the generic value to the correct type and update
	switch req.Path {
	case "models.large.model", "models.large.provider":
		return updateModelField(cfg, config.SelectedModelTypeLarge, req)
	case "models.small.model", "models.small.provider":
		return updateModelField(cfg, config.SelectedModelTypeSmall, req)
	case "tools.ls.max_depth":
		v := int(req.Value.(float64))
		cfg.Tools.Ls.MaxDepth = &v
		return cfg.SetConfigField(req.Path, v)
	case "tools.ls.max_items":
		v := int(req.Value.(float64))
		cfg.Tools.Ls.MaxItems = &v
		return cfg.SetConfigField(req.Path, v)
	case "ui.compact_mode":
		return cfg.SetCompactMode(req.Value.(bool))
	case "ui.diff_mode":
		if cfg.Options == nil || cfg.Options.TUI == nil {
			return fmt.Errorf("TUI options not initialized")
		}
		cfg.Options.TUI.DiffMode = req.Value.(string)
		return cfg.SetConfigField("options.tui.diff_mode", req.Value)
	default:
		// Generic persistence using SetConfigField
		return cfg.SetConfigField(req.Path, req.Value)
	}
}

func updateModelField(cfg *config.Config, modelType config.SelectedModelType, req *FieldUpdateRequest) error {
	model, ok := cfg.Models[modelType]
	if !ok {
		return fmt.Errorf("model type %s not configured", modelType)
	}

	// Parse the field name from path (e.g., "models.large.model" -> "model")
	var value interface{}
	if v, ok := req.Value.(string); ok {
		value = v
	} else if v, ok := req.Value.(float64); ok {
		value = v
	} else if v, ok := req.Value.(bool); ok {
		value = v
	} else {
		// Try to marshal/unmarshal through JSON to handle the value
		data, err := json.Marshal(req.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		if err := json.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("failed to unmarshal value: %w", err)
		}
	}

	// Update the specific field and persist
	return cfg.UpdatePreferredModel(modelType, model)
}
