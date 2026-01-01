package api

import (
	"github.com/charmbracelet/crush/internal/config"
)

// ConfigResponse represents the full application configuration exposed via the API.
// Sensitive fields (e.g., API keys) are redacted before sending to clients.
type ConfigResponse struct {
	Providers []ProviderDTO `json:"providers,omitempty"`
	Tools     ToolsDTO      `json:"tools,omitempty"`
	UI        UIDTO         `json:"ui,omitempty"`
	Models    ModelsDTO     `json:"models,omitempty"`
	Options   OptionsDTO    `json:"options,omitempty"`
	Schema    *SchemaDTO    `json:"schema,omitempty"` // Only included if ?include_schema=true
}

// SchemaDTO contains metadata describing the configuration schema.
type SchemaDTO struct {
	Providers map[string]FieldSchema `json:"providers,omitempty"`
	Tools     map[string]FieldSchema `json:"tools,omitempty"`
	UI        map[string]FieldSchema `json:"ui,omitempty"`
	Models    map[string]FieldSchema `json:"models,omitempty"`
	Options   map[string]FieldSchema `json:"options,omitempty"`
}

// FieldSchema describes a configuration field's metadata.
type FieldSchema struct {
	Type        string      `json:"type"`                  // e.g., "string", "number", "boolean", "array", "object"
	Description string      `json:"description,omitempty"` // Human-readable description
	Enum        []string    `json:"enum,omitempty"`        // Allowed values for enum fields
	Required    bool        `json:"required,omitempty"`    // Whether the field is required
	Default     interface{} `json:"default,omitempty"`     // Default value
	Minimum     *float64    `json:"minimum,omitempty"`     // For numeric fields
	Maximum     *float64    `json:"maximum,omitempty"`     // For numeric fields
	Example     interface{} `json:"example,omitempty"`     // Example value
	Redacted    bool        `json:"redacted,omitempty"`    // Whether the field is redacted (e.g., API keys)
	ReadOnly    bool        `json:"read_only,omitempty"`   // Whether the field is read-only
	Properties  interface{} `json:"properties,omitempty"`  // Nested schema for object types
}

// ProviderDTO represents a provider configuration with redacted sensitive data.
type ProviderDTO struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	BaseURL      string                 `json:"base_url,omitempty"`
	APIKey       string                 `json:"api_key"` // Always redacted
	Enabled      bool                   `json:"enabled"` // Inverse of Disable field
	Models       []string               `json:"models"`  // List of model IDs
	ExtraHeaders map[string]string      `json:"extra_headers,omitempty"`
	ExtraBody    map[string]interface{} `json:"extra_body,omitempty"`
	HasOAuth     bool                   `json:"has_oauth"` // Whether OAuth is configured
}

// ToolsDTO represents tool configurations.
type ToolsDTO struct {
	Bash          BashToolDTO        `json:"bash,omitempty"`
	LSP           map[string]LSPInfo `json:"lsp,omitempty"`
	MCP           map[string]MCPInfo `json:"mcp,omitempty"`
	DisabledTools []string           `json:"disabled_tools,omitempty"`
	Ls            LsToolDTO          `json:"ls,omitempty"`
}

// BashToolDTO represents bash tool configuration.
type BashToolDTO struct {
	Timeout int `json:"timeout,omitempty"` // Timeout in seconds
}

// LsToolDTO represents ls tool configuration.
type LsToolDTO struct {
	MaxDepth int `json:"max_depth,omitempty"`
	MaxItems int `json:"max_items,omitempty"`
}

// LSPInfo represents LSP server information.
type LSPInfo struct {
	Command     string   `json:"command"`
	Args        []string `json:"args,omitempty"`
	FileTypes   []string `json:"filetypes,omitempty"`
	RootMarkers []string `json:"root_markers,omitempty"`
	Disabled    bool     `json:"disabled"`
}

// MCPInfo represents MCP server information.
type MCPInfo struct {
	Type          string   `json:"type"`
	Command       string   `json:"command,omitempty"`
	Args          []string `json:"args,omitempty"`
	URL           string   `json:"url,omitempty"`
	Disabled      bool     `json:"disabled"`
	DisabledTools []string `json:"disabled_tools,omitempty"`
	Timeout       int      `json:"timeout,omitempty"`
}

// UIDTO represents UI configuration.
type UIDTO struct {
	Theme              string `json:"theme,omitempty"` // "dark" or "light"
	SyntaxHighlighting bool   `json:"syntax_highlighting"`
	ShowCost           bool   `json:"show_cost"`
	CompactMode        bool   `json:"compact_mode"`
	DiffMode           string `json:"diff_mode,omitempty"` // "unified" or "split"
}

// ModelsDTO represents model configuration.
type ModelsDTO struct {
	Large  *ModelSelectionDTO             `json:"large,omitempty"`
	Small  *ModelSelectionDTO             `json:"small,omitempty"`
	Recent map[string][]ModelSelectionDTO `json:"recent,omitempty"`
}

// ModelSelectionDTO represents a selected model.
type ModelSelectionDTO struct {
	Model           string   `json:"model"`
	Provider        string   `json:"provider"`
	MaxTokens       int64    `json:"max_tokens,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	ReasoningEffort string   `json:"reasoning_effort,omitempty"`
	Think           bool     `json:"think,omitempty"`
}

// OptionsDTO represents general application options.
type OptionsDTO struct {
	ContextPaths         []string        `json:"context_paths,omitempty"`
	SkillsPaths          []string        `json:"skills_paths,omitempty"`
	DataDirectory        string          `json:"data_directory,omitempty"`
	Debug                bool            `json:"debug"`
	DebugLSP             bool            `json:"debug_lsp"`
	DisableAutoSummarize bool            `json:"disable_auto_summarize"`
	DisableMetrics       bool            `json:"disable_metrics"`
	Attribution          *AttributionDTO `json:"attribution,omitempty"`
	InitializeAs         string          `json:"initialize_as,omitempty"`
}

// AttributionDTO represents attribution settings.
type AttributionDTO struct {
	TrailerStyle  string `json:"trailer_style,omitempty"`
	GeneratedWith bool   `json:"generated_with"`
}

// ToConfigResponse converts internal config to API response with redaction.
func ToConfigResponse(cfg *config.Config, includeSchema bool) ConfigResponse {
	resp := ConfigResponse{
		Providers: toProviderDTOs(cfg),
		Tools:     toToolsDTO(cfg),
		UI:        toUIDTO(cfg),
		Models:    toModelsDTO(cfg),
		Options:   toOptionsDTO(cfg),
	}

	if includeSchema {
		resp.Schema = buildSchemaDTO()
	}

	return resp
}

func toProviderDTOs(cfg *config.Config) []ProviderDTO {
	var providers []ProviderDTO
	for prov := range cfg.Providers.Seq() {
		models := make([]string, len(prov.Models))
		for i, m := range prov.Models {
			models[i] = m.ID
		}

		// Redact API key
		apiKeyDisplay := "not_configured"
		if prov.APIKey != "" {
			apiKeyDisplay = "encrypted"
		}

		providers = append(providers, ProviderDTO{
			ID:           prov.ID,
			Name:         prov.Name,
			Type:         string(prov.Type),
			BaseURL:      prov.BaseURL,
			APIKey:       apiKeyDisplay,
			Enabled:      !prov.Disable,
			Models:       models,
			ExtraHeaders: prov.ExtraHeaders,
			ExtraBody:    prov.ExtraBody,
			HasOAuth:     prov.OAuthToken != nil,
		})
	}
	return providers
}

func toToolsDTO(cfg *config.Config) ToolsDTO {
	dto := ToolsDTO{
		DisabledTools: []string{},
		LSP:           make(map[string]LSPInfo),
		MCP:           make(map[string]MCPInfo),
	}

	if cfg.Options != nil {
		dto.DisabledTools = cfg.Options.DisabledTools
	}

	// Convert LSP configs
	for _, lsp := range cfg.LSP.Sorted() {
		dto.LSP[lsp.Name] = LSPInfo{
			Command:     lsp.LSP.Command,
			Args:        lsp.LSP.Args,
			FileTypes:   lsp.LSP.FileTypes,
			RootMarkers: lsp.LSP.RootMarkers,
			Disabled:    lsp.LSP.Disabled,
		}
	}

	// Convert MCP configs
	for _, mcp := range cfg.MCP.Sorted() {
		dto.MCP[mcp.Name] = MCPInfo{
			Type:          string(mcp.MCP.Type),
			Command:       mcp.MCP.Command,
			Args:          mcp.MCP.Args,
			URL:           mcp.MCP.URL,
			Disabled:      mcp.MCP.Disabled,
			DisabledTools: mcp.MCP.DisabledTools,
			Timeout:       mcp.MCP.Timeout,
		}
	}

	// Add ls tool config
	depth, items := cfg.Tools.Ls.Limits()
	dto.Ls = LsToolDTO{
		MaxDepth: depth,
		MaxItems: items,
	}

	return dto
}

func toUIDTO(cfg *config.Config) UIDTO {
	dto := UIDTO{
		Theme:              "dark", // Default
		SyntaxHighlighting: true,   // Default
		ShowCost:           true,   // Default
	}

	if cfg.Options != nil && cfg.Options.TUI != nil {
		dto.CompactMode = cfg.Options.TUI.CompactMode
		dto.DiffMode = cfg.Options.TUI.DiffMode
	}

	return dto
}

func toModelsDTO(cfg *config.Config) ModelsDTO {
	dto := ModelsDTO{
		Recent: make(map[string][]ModelSelectionDTO),
	}

	// Convert selected models
	if large, ok := cfg.Models[config.SelectedModelTypeLarge]; ok {
		dto.Large = toModelSelectionDTO(large)
	}
	if small, ok := cfg.Models[config.SelectedModelTypeSmall]; ok {
		dto.Small = toModelSelectionDTO(small)
	}

	// Convert recent models
	for modelType, recents := range cfg.RecentModels {
		dtoRecents := make([]ModelSelectionDTO, len(recents))
		for i, r := range recents {
			dtoRecents[i] = *toModelSelectionDTO(r)
		}
		dto.Recent[string(modelType)] = dtoRecents
	}

	return dto
}

func toModelSelectionDTO(sel config.SelectedModel) *ModelSelectionDTO {
	return &ModelSelectionDTO{
		Model:           sel.Model,
		Provider:        sel.Provider,
		MaxTokens:       sel.MaxTokens,
		Temperature:     sel.Temperature,
		ReasoningEffort: sel.ReasoningEffort,
		Think:           sel.Think,
	}
}

func toOptionsDTO(cfg *config.Config) OptionsDTO {
	dto := OptionsDTO{}

	if cfg.Options != nil {
		dto.ContextPaths = cfg.Options.ContextPaths
		dto.SkillsPaths = cfg.Options.SkillsPaths
		dto.DataDirectory = cfg.Options.DataDirectory
		dto.Debug = cfg.Options.Debug
		dto.DebugLSP = cfg.Options.DebugLSP
		dto.DisableAutoSummarize = cfg.Options.DisableAutoSummarize
		dto.DisableMetrics = cfg.Options.DisableMetrics
		dto.InitializeAs = cfg.Options.InitializeAs

		if cfg.Options.Attribution != nil {
			dto.Attribution = &AttributionDTO{
				TrailerStyle:  string(cfg.Options.Attribution.TrailerStyle),
				GeneratedWith: cfg.Options.Attribution.GeneratedWith,
			}
		}
	}

	return dto
}

// buildSchemaDTO constructs the schema metadata for configuration.
func buildSchemaDTO() *SchemaDTO {
	return &SchemaDTO{
		Providers: map[string]FieldSchema{
			"id": {
				Type:        "string",
				Description: "Unique identifier for the provider",
				Required:    true,
				Example:     "openai",
			},
			"name": {
				Type:        "string",
				Description: "Human-readable name for the provider",
				Example:     "OpenAI",
			},
			"type": {
				Type:        "string",
				Description: "Provider type that determines the API format",
				Enum:        []string{"openai", "openai-compat", "anthropic", "gemini", "azure", "vertexai"},
				Example:     "openai",
			},
			"base_url": {
				Type:        "string",
				Description: "Base URL for the provider's API",
				Example:     "https://api.openai.com/v1",
			},
			"api_key": {
				Type:        "string",
				Description: "API key for authentication (always redacted in responses)",
				Redacted:    true,
				Example:     "$OPENAI_API_KEY",
			},
			"enabled": {
				Type:        "boolean",
				Description: "Whether this provider is enabled",
				Default:     true,
			},
		},
		Tools: map[string]FieldSchema{
			"bash.timeout": {
				Type:        "number",
				Description: "Timeout in seconds for bash commands",
				Default:     float64(30),
				Minimum:     ptrFloat64(0),
			},
			"ls.max_depth": {
				Type:        "number",
				Description: "Maximum depth for the ls tool",
				Default:     float64(0),
				Example:     float64(10),
			},
			"ls.max_items": {
				Type:        "number",
				Description: "Maximum number of items to return for the ls tool",
				Default:     float64(1000),
				Example:     float64(100),
			},
		},
		UI: map[string]FieldSchema{
			"theme": {
				Type:        "string",
				Description: "UI theme",
				Enum:        []string{"dark", "light"},
				Default:     "dark",
			},
			"syntax_highlighting": {
				Type:        "boolean",
				Description: "Enable syntax highlighting in code blocks",
				Default:     true,
			},
			"show_cost": {
				Type:        "boolean",
				Description: "Show estimated cost of API calls",
				Default:     true,
			},
			"compact_mode": {
				Type:        "boolean",
				Description: "Enable compact mode for the TUI interface",
				Default:     false,
			},
			"diff_mode": {
				Type:        "string",
				Description: "Diff mode for the TUI interface",
				Enum:        []string{"unified", "split"},
			},
		},
		Models: map[string]FieldSchema{
			"model": {
				Type:        "string",
				Description: "The model ID as used by the provider API",
				Required:    true,
				Example:     "gpt-4o",
			},
			"provider": {
				Type:        "string",
				Description: "The model provider ID that matches a key in the providers config",
				Required:    true,
				Example:     "openai",
			},
			"max_tokens": {
				Type:        "number",
				Description: "Maximum number of tokens for model responses",
				Maximum:     ptrFloat64(200000),
				Example:     float64(4096),
			},
			"temperature": {
				Type:        "number",
				Description: "Sampling temperature",
				Minimum:     ptrFloat64(0),
				Maximum:     ptrFloat64(1),
				Example:     0.7,
			},
			"reasoning_effort": {
				Type:        "string",
				Description: "Reasoning effort level for OpenAI models that support it",
				Enum:        []string{"low", "medium", "high"},
			},
		},
		Options: map[string]FieldSchema{
			"context_paths": {
				Type:        "array",
				Description: "Paths to files containing context information for the AI",
				Example:     []string{".cursorrules", "CRUSH.md"},
			},
			"data_directory": {
				Type:        "string",
				Description: "Directory for storing application data (relative to working directory)",
				Default:     ".crush",
				Example:     ".crush",
			},
			"debug": {
				Type:        "boolean",
				Description: "Enable debug logging",
				Default:     false,
			},
			"disable_metrics": {
				Type:        "boolean",
				Description: "Disable sending metrics",
				Default:     false,
			},
		},
	}
}

func ptrFloat64(f float64) *float64 {
	return &f
}
