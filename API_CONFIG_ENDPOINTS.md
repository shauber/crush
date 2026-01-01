# Configuration API Endpoints

This document describes the configuration management endpoints exposed by the Crush API server.

## Overview

The configuration API provides read access to the application's configuration, including providers, tools, UI preferences, and model settings. All sensitive data (e.g., API keys) is automatically redacted before being sent to clients.

## Endpoints

### GET /config

Returns the full application configuration with optional schema metadata.

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `include_schema` | boolean | No | `false` | Include schema metadata describing field types, constraints, and descriptions |

#### Response

**Status Code:** 200 OK

**Content-Type:** application/json

**Response Body:**

```json
{
  "providers": [
    {
      "id": "openai",
      "name": "OpenAI",
      "type": "openai",
      "base_url": "https://api.openai.com/v1",
      "api_key": "encrypted",
      "enabled": true,
      "models": ["gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"],
      "has_oauth": false
    },
    {
      "id": "anthropic",
      "name": "Anthropic",
      "type": "anthropic",
      "base_url": "https://api.anthropic.com/v1",
      "api_key": "not_configured",
      "enabled": false,
      "models": ["claude-3.5-sonnet", "claude-3-opus", "claude-3-haiku"],
      "has_oauth": false
    }
  ],
  "tools": {
    "bash": {
      "timeout": 30
    },
    "ls": {
      "max_depth": 0,
      "max_items": 1000
    },
    "disabled_tools": [],
    "lsp": {
      "gopls": {
        "command": "gopls",
        "args": [],
        "filetypes": ["go", "mod"],
        "root_markers": ["go.mod"],
        "disabled": false
      }
    },
    "mcp": {}
  },
  "ui": {
    "theme": "dark",
    "syntax_highlighting": true,
    "show_cost": true,
    "compact_mode": false,
    "diff_mode": "unified"
  },
  "models": {
    "large": {
      "model": "gpt-4o",
      "provider": "openai",
      "max_tokens": 4096,
      "temperature": 0.7
    },
    "small": {
      "model": "claude-3-haiku",
      "provider": "anthropic"
    },
    "recent": {
      "large": [
        {
          "model": "gpt-4o",
          "provider": "openai"
        },
        {
          "model": "claude-3.5-sonnet",
          "provider": "anthropic"
        }
      ]
    }
  },
  "options": {
    "context_paths": [".cursorrules", "CRUSH.md"],
    "skills_paths": ["~/.config/crush/skills"],
    "data_directory": ".crush",
    "debug": false,
    "debug_lsp": false,
    "disable_auto_summarize": false,
    "disable_metrics": false,
    "attribution": {
      "trailer_style": "assisted-by",
      "generated_with": true
    },
    "initialize_as": "AGENTS.md"
  }
}
```

#### With Schema Metadata

When `?include_schema=true` is specified, the response includes an additional `schema` field:

```json
{
  "providers": [...],
  "tools": {...},
  "ui": {...},
  "models": {...},
  "options": {...},
  "schema": {
    "providers": {
      "id": {
        "type": "string",
        "description": "Unique identifier for the provider",
        "required": true,
        "example": "openai"
      },
      "api_key": {
        "type": "string",
        "description": "API key for authentication (always redacted in responses)",
        "redacted": true,
        "example": "$OPENAI_API_KEY"
      },
      "enabled": {
        "type": "boolean",
        "description": "Whether this provider is enabled",
        "default": true
      }
    },
    "tools": {
      "bash.timeout": {
        "type": "number",
        "description": "Timeout in seconds for bash commands",
        "default": 30,
        "minimum": 0
      },
      "ls.max_depth": {
        "type": "number",
        "description": "Maximum depth for the ls tool",
        "default": 0,
        "example": 10
      }
    },
    "ui": {
      "theme": {
        "type": "string",
        "description": "UI theme",
        "enum": ["dark", "light"],
        "default": "dark"
      },
      "compact_mode": {
        "type": "boolean",
        "description": "Enable compact mode for the TUI interface",
        "default": false
      }
    },
    "models": {
      "model": {
        "type": "string",
        "description": "The model ID as used by the provider API",
        "required": true,
        "example": "gpt-4o"
      },
      "temperature": {
        "type": "number",
        "description": "Sampling temperature",
        "minimum": 0,
        "maximum": 1,
        "example": 0.7
      }
    }
  }
}
```

#### Examples

**Basic Configuration Request:**

```bash
curl http://localhost:8080/config
```

**With Schema Metadata:**

```bash
curl http://localhost:8080/config?include_schema=true
```

**Using jq to extract specific fields:**

```bash
# Get all provider IDs
curl -s http://localhost:8080/config | jq -r '.providers[].id'

# Get large model configuration
curl -s http://localhost:8080/config | jq '.models.large'

# Check if a provider is enabled
curl -s http://localhost:8080/config | jq '.providers[] | select(.id == "openai") | .enabled'
```

## Security Considerations

### API Key Redaction

All API keys in the configuration response are automatically redacted. The `api_key` field will contain one of the following values:

- `"encrypted"` - Provider has a configured API key
- `"not_configured"` - Provider has no API key configured

The actual API key values are **never** sent to clients via the API.

### OAuth Tokens

OAuth tokens are not included in the response. The `has_oauth` field indicates whether a provider uses OAuth authentication, but token values remain server-side only.

## Schema Metadata

When `include_schema=true` is specified, the response includes metadata describing each configuration field:

### Field Schema Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | string | Field data type: `"string"`, `"number"`, `"boolean"`, `"array"`, `"object"` |
| `description` | string | Human-readable description of the field |
| `enum` | array | List of allowed values (for enumerated fields) |
| `required` | boolean | Whether the field is required |
| `default` | any | Default value for the field |
| `minimum` | number | Minimum value (for numeric fields) |
| `maximum` | number | Maximum value (for numeric fields) |
| `example` | any | Example value |
| `redacted` | boolean | Whether the field is redacted in responses (e.g., API keys) |

### Use Cases for Schema Metadata

1. **Dynamic UI Generation**: Build configuration forms based on field metadata
2. **Validation**: Client-side validation of config updates before submission
3. **Documentation**: Auto-generate documentation from schema descriptions
4. **Type Safety**: Ensure correct data types when updating configuration

## Response Structure

### Providers

Each provider object includes:

- `id`: Provider identifier (matches config key)
- `name`: Human-readable provider name
- `type`: Provider type (openai, anthropic, gemini, etc.)
- `base_url`: API endpoint URL
- `api_key`: Redacted API key status
- `enabled`: Whether the provider is active
- `models`: Array of model IDs available from this provider
- `has_oauth`: Whether OAuth is configured

### Tools

Tool configuration includes:

- **bash**: Bash tool settings (timeout)
- **ls**: Directory listing limits (max_depth, max_items)
- **lsp**: Language Server Protocol configurations
- **mcp**: Model Context Protocol server configurations
- **disabled_tools**: List of disabled tool names

### UI

User interface preferences:

- `theme`: Color theme ("dark" or "light")
- `syntax_highlighting`: Enable code syntax highlighting
- `show_cost`: Display API call costs
- `compact_mode`: Use compact UI layout
- `diff_mode`: Diff display mode ("unified" or "split")

### Models

Model configuration for different size categories:

- **large**: Primary model configuration
- **small**: Smaller/faster model configuration
- **recent**: Recently used models by category

Each model configuration includes:

- `model`: Model ID
- `provider`: Provider ID
- `max_tokens`: Maximum output tokens (optional)
- `temperature`: Sampling temperature (optional)
- `reasoning_effort`: Reasoning level for OpenAI o1 models (optional)
- `think`: Enable thinking for Anthropic models (optional)

### Options

General application options:

- `context_paths`: Paths to context files for AI
- `skills_paths`: Paths to skill directories
- `data_directory`: Application data storage directory
- `debug`: Debug logging enabled
- `debug_lsp`: LSP debug logging enabled
- `disable_auto_summarize`: Disable conversation summarization
- `disable_metrics`: Disable usage metrics
- `attribution`: Attribution settings for generated content
- `initialize_as`: Default context file name for new projects

## Error Responses

### 500 Internal Server Error

```json
{
  "error": "Failed to load configuration"
}
```

Occurs when the server cannot read or process the configuration. Check server logs for details.

## Notes

- Configuration is read-only via this endpoint. Write operations will be supported in Phase A Day 2+.
- Schema metadata is cached and generated on-demand for efficiency.
- The configuration structure mirrors the CLI's config file format for consistency.

## Related Endpoints

- `GET /providers` - List providers without full config (lightweight)
- `GET /models` - Get model selection without full config
- `PUT /config` (Phase A Day 2) - Update configuration
- `PUT /config/field` (Phase A Day 2) - Update individual fields

## CLI Parity

This endpoint provides the data necessary to replicate the CLI's configuration viewing functionality. The TUI configuration browser can be implemented using this API by:

1. Fetching config with schema metadata
2. Rendering a browsable tree of configuration sections
3. Displaying field descriptions and current values
4. Marking sensitive fields as redacted

## Implementation Details

- **Serialization**: Config structs converted to DTOs for API responses
- **Redaction**: Centralized in `ToConfigResponse()` function
- **Schema Generation**: Static schema built from struct tags and constants
- **Performance**: No database queries required; config loaded from memory
