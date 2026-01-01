# Configuration API Endpoints

This document describes the configuration management endpoints exposed by the Crush API server.

## Overview

The configuration API provides read and write access to the application's configuration, including providers, tools, UI preferences, and model settings. All sensitive data (e.g., API keys) is automatically redacted when reading configuration.

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

---

### PUT /config

Updates multiple configuration sections atomically. Only the provided fields will be updated; all others remain unchanged (partial update).

#### Request Body

**Content-Type:** application/json

```json
{
  "models": {
    "large": {
      "model": "gpt-4o",
      "provider": "openai",
      "temperature": 0.7,
      "max_tokens": 4096
    },
    "small": {
      "model": "claude-3-haiku",
      "provider": "anthropic"
    }
  },
  "tools": {
    "ls": {
      "max_depth": 10,
      "max_items": 100
    },
    "disabled_tools": ["sourcegraph"]
  },
  "ui": {
    "compact_mode": true,
    "diff_mode": "split"
  },
  "options": {
    "context_paths": [".cursorrules", "CRUSH.md"],
    "disable_metrics": false
  },
  "providers": {
    "updates": [
      {
        "id": "openai",
        "enabled": true,
        "base_url": "https://api.openai.com/v1"
      }
    ]
  }
}
```

**Note:** All fields are optional. Only include the sections you want to update.

#### Response

**Status Code:** 200 OK

**Content-Type:** application/json

Returns the full updated configuration (same format as `GET /config`).

#### Validation Rules

| Field | Type | Constraints |
|-------|------|-------------|
| `models.*.model` | string | Required when updating model |
| `models.*.provider` | string | Required when updating model |
| `models.*.temperature` | number | 0.0 - 1.0 |
| `models.*.max_tokens` | number | 0 - 200000 |
| `models.*.reasoning_effort` | string | "low", "medium", "high" |
| `tools.ls.max_depth` | number | >= 0 |
| `tools.ls.max_items` | number | >= 0 |
| `ui.theme` | string | "dark", "light" |
| `ui.diff_mode` | string | "unified", "split" |
| `ui.compact_mode` | boolean | - |
| `providers.*.id` | string | Required for provider updates |

#### Error Responses

**400 Bad Request** - Validation failed

```json
{
  "error": "Validation failed: temperature must be between 0 and 1"
}
```

**500 Internal Server Error** - Failed to apply updates

```json
{
  "error": "Failed to apply config update: provider not found"
}
```

#### Examples

**Update Model Configuration:**

```bash
curl -X PUT http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{
    "models": {
      "large": {
        "model": "gpt-4o",
        "provider": "openai",
        "temperature": 0.8
      }
    }
  }'
```

**Update Multiple Sections:**

```bash
curl -X PUT http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{
    "models": {
      "large": {
        "model": "claude-3.5-sonnet",
        "provider": "anthropic"
      }
    },
    "ui": {
      "compact_mode": true,
      "diff_mode": "split"
    },
    "tools": {
      "ls": {
        "max_depth": 15
      }
    }
  }'
```

**Update UI Preferences Only:**

```bash
curl -X PUT http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{
    "ui": {
      "compact_mode": false,
      "diff_mode": "unified"
    }
  }'
```

---

### PUT /config/field

Updates a single configuration field using dot notation. More efficient than `PUT /config` for single-field updates.

#### Request Body

**Content-Type:** application/json

```json
{
  "path": "ui.compact_mode",
  "value": true
}
```

#### Supported Field Paths

**Models:**
- `models.large.model`
- `models.large.provider`
- `models.large.max_tokens`
- `models.large.temperature`
- `models.large.reasoning_effort`
- `models.large.think`
- `models.small.*` (same fields)

**Tools:**
- `tools.ls.max_depth`
- `tools.ls.max_items`
- `tools.disabled_tools`

**UI:**
- `ui.theme`
- `ui.compact_mode`
- `ui.diff_mode`

**Options:**
- `options.context_paths`
- `options.skills_paths`
- `options.disable_auto_summarize`
- `options.disable_metrics`
- `options.debug`

#### Response

**Status Code:** 200 OK

**Content-Type:** application/json

```json
{
  "message": "Field updated successfully",
  "path": "ui.compact_mode",
  "value": true
}
```

#### Error Responses

**400 Bad Request** - Invalid path or value

```json
{
  "error": "Validation failed: unknown field path: invalid.path"
}
```

```json
{
  "error": "Validation failed: invalid value for ui.theme: must be 'dark' or 'light'"
}
```

**500 Internal Server Error** - Failed to persist update

```json
{
  "error": "Failed to apply field update: failed to persist config"
}
```

#### Examples

**Update Compact Mode:**

```bash
curl -X PUT http://localhost:8080/config/field \
  -H "Content-Type: application/json" \
  -d '{
    "path": "ui.compact_mode",
    "value": true
  }'
```

**Update Ls Max Depth:**

```bash
curl -X PUT http://localhost:8080/config/field \
  -H "Content-Type: application/json" \
  -d '{
    "path": "tools.ls.max_depth",
    "value": 20
  }'
```

**Update Context Paths:**

```bash
curl -X PUT http://localhost:8080/config/field \
  -H "Content-Type: application/json" \
  -d '{
    "path": "options.context_paths",
    "value": [".cursorrules", "CRUSH.md", "docs/AI.md"]
  }'
```

**Update Model Temperature:**

```bash
curl -X PUT http://localhost:8080/config/field \
  -H "Content-Type: application/json" \
  -d '{
    "path": "models.large.temperature",
    "value": 0.9
  }'
```

---

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
- `POST /models` - Set model (alternative to PUT /config for models)
- `GET /settings` (Phase A Day 3) - Settings-specific endpoint
- `PUT /settings/{key}` (Phase A Day 3) - Update individual settings

## Persistence

All configuration updates are automatically persisted to disk in the application's data directory. Changes survive application restarts.

**Persistence Details:**
- Configuration stored in: `~/.config/crush/crush.json` (or configured data directory)
- Atomic writes prevent corruption
- Changes are applied in-memory first, then persisted
- Failed persistence rolls back in-memory changes

## Change Events

Configuration updates trigger events for real-time synchronization:

**Event:** `config_updated`
- Emitted via SSE after successful config updates
- Contains the updated configuration section
- Allows web clients to stay synchronized
- *Note: SSE event emission is prepared for Phase C implementation*

## Validation

All configuration updates are validated before being applied:

1. **Type Checking**: Values must match expected types (string, number, boolean, array)
2. **Range Validation**: Numeric values must be within allowed ranges
3. **Enum Validation**: String fields with limited options are checked against allowed values
4. **Required Fields**: Required fields must be present when updating objects
5. **Consistency**: Related fields are validated together (e.g., model + provider)

**Validation Errors Return 400 Bad Request** with detailed error messages.

## Best Practices

### Bulk vs. Field Updates

**Use `PUT /config` when:**
- Updating multiple related fields
- Applying a configuration template
- Making coordinated changes across sections
- Want to receive the full updated config in response

**Use `PUT /config/field` when:**
- Updating a single specific field
- Building a settings UI with individual toggles
- Want minimal request/response payload
- Only care about confirming that specific field update

### Partial Updates

Both `PUT /config` and `PUT /config/field` perform partial updates:
- Only specified fields are changed
- Unspecified fields retain their current values
- null values are generally ignored (field-specific behavior)

### Error Handling

Always check HTTP status codes:
- **200 OK**: Update successful
- **400 Bad Request**: Validation failed, check error message
- **500 Internal Server Error**: Server-side issue, may need retry

Parse error messages for specific validation failures:
```bash
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "Validation failed: models validation failed: large model: temperature must be between 0 and 1"
}
```

## Related Endpoints

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
