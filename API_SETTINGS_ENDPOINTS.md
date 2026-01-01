# Settings API Endpoints

This document describes the settings management endpoints, which provide a user-friendly interface to configuration focused on commonly-adjusted preferences.

## Overview

The settings API provides a simplified, user-facing view of configuration. Unlike the `/config` endpoints which expose the full technical configuration, `/settings` endpoints focus on user preferences organized into intuitive categories:

- **UI**: Visual preferences (theme, compact mode, diff view)
- **Editor**: Context and skill paths
- **Models**: Default model selections
- **General**: Debug flags and metrics

## Relationship to Config API

Settings are a subset/view of the configuration:
- `GET /settings` returns user-friendly settings (subset of `/config`)
- `PUT /settings` updates underlying config fields
- Changes via `/settings` are reflected in `/config` and vice versa
- Settings provide simpler keys (e.g., `ui.theme` vs `options.tui.diff_mode`)

## Endpoints

### GET /settings

Returns all user-facing settings organized by category.

#### Response

**Status Code:** 200 OK

**Content-Type:** application/json

```json
{
  "ui": {
    "theme": "dark",
    "syntax_highlighting": true,
    "show_cost": true,
    "compact_mode": false,
    "diff_mode": "unified"
  },
  "editor": {
    "context_paths": [".cursorrules", "CRUSH.md"],
    "skills_paths": ["~/.config/crush/skills"]
  },
  "models": {
    "default_large": "openai/gpt-4o",
    "default_small": "anthropic/claude-3-haiku"
  },
  "general": {
    "debug": false,
    "debug_lsp": false,
    "disable_auto_summarize": false,
    "disable_metrics": false
  }
}
```

#### Examples

```bash
curl http://localhost:8080/settings
```

```bash
# Get just UI settings using jq
curl -s http://localhost:8080/settings | jq '.ui'
```

---

### GET /settings/schema

Returns schema metadata for all settings.

#### Response

**Status Code:** 200 OK

**Content-Type:** application/json

```json
{
  "settings": [
    {
      "key": "ui.theme",
      "type": "string",
      "description": "Color theme for the user interface",
      "default": "dark",
      "enum": ["dark", "light"],
      "category": "ui",
      "example": "dark"
    },
    {
      "key": "ui.compact_mode",
      "type": "boolean",
      "description": "Use compact UI layout with reduced spacing",
      "default": false,
      "category": "ui"
    }
    // ... more settings
  ]
}
```

#### Examples

```bash
curl http://localhost:8080/settings/schema
```

---

### GET /settings/{key}/schema

Returns schema metadata for a specific setting.

#### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `key` | string | Yes | Setting key (e.g., `ui.theme`) |

#### Response

**Status Code:** 200 OK

**Content-Type:** application/json

```json
{
  "key": "ui.theme",
  "type": "string",
  "description": "Color theme for the user interface",
  "default": "dark",
  "enum": ["dark", "light"],
  "category": "ui",
  "example": "dark"
}
```

#### Error Responses

**404 Not Found** - Setting key not found

```json
{
  "error": "Setting not found: invalid.key"
}
```

#### Examples

```bash
curl http://localhost:8080/settings/ui.theme/schema
```

```bash
curl http://localhost:8080/settings/general.debug/schema
```

---

### PUT /settings

Updates multiple settings atomically.

#### Request Body

**Content-Type:** application/json

All fields are optional. Only include the sections/settings you want to update.

```json
{
  "ui": {
    "compact_mode": true,
    "diff_mode": "split",
    "theme": "dark",
    "syntax_highlighting": true,
    "show_cost": true
  },
  "editor": {
    "context_paths": [".cursorrules", "CRUSH.md"],
    "skills_paths": ["~/.config/crush/skills"]
  },
  "models": {
    "default_large": "openai/gpt-4o",
    "default_small": "anthropic/claude-3-haiku"
  },
  "general": {
    "debug": false,
    "debug_lsp": false,
    "disable_auto_summarize": false,
    "disable_metrics": false
  }
}
```

#### Response

**Status Code:** 200 OK

**Content-Type:** application/json

Returns the full updated settings (same format as `GET /settings`).

#### Examples

**Update UI Preferences:**

```bash
curl -X PUT http://localhost:8080/settings \
  -H "Content-Type: application/json" \
  -d '{
    "ui": {
      "compact_mode": true,
      "diff_mode": "split"
    }
  }'
```

**Update Model Defaults:**

```bash
curl -X PUT http://localhost:8080/settings \
  -H "Content-Type: application/json" \
  -d '{
    "models": {
      "default_large": "anthropic/claude-3.5-sonnet"
    }
  }'
```

**Update Multiple Sections:**

```bash
curl -X PUT http://localhost:8080/settings \
  -H "Content-Type: application/json" \
  -d '{
    "ui": {
      "compact_mode": true
    },
    "general": {
      "debug": true
    },
    "editor": {
      "context_paths": [".cursorrules", "CRUSH.md", "docs/AI.md"]
    }
  }'
```

---

### PUT /settings/{key}

Updates a single setting value.

#### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `key` | string | Yes | Setting key (e.g., `ui.compact_mode`) |

#### Request Body

**Content-Type:** application/json

```json
{
  "value": true
}
```

The `value` type must match the setting's schema type.

#### Response

**Status Code:** 200 OK

**Content-Type:** application/json

```json
{
  "message": "Setting updated successfully",
  "key": "ui.compact_mode",
  "value": true
}
```

#### Error Responses

**400 Bad Request** - Invalid value

```json
{
  "error": "Validation failed: invalid value for ui.theme: must be 'dark' or 'light'"
}
```

**500 Internal Server Error** - Failed to persist

```json
{
  "error": "Failed to apply setting update: failed to persist config"
}
```

#### Examples

**Update Boolean Setting:**

```bash
curl -X PUT http://localhost:8080/settings/ui.compact_mode \
  -H "Content-Type: application/json" \
  -d '{"value": true}'
```

**Update String Setting:**

```bash
curl -X PUT http://localhost:8080/settings/ui.diff_mode \
  -H "Content-Type: application/json" \
  -d '{"value": "split"}'
```

**Update Array Setting:**

```bash
curl -X PUT http://localhost:8080/settings/editor.context_paths \
  -H "Content-Type: application/json" \
  -d '{"value": [".cursorrules", "CRUSH.md"]}'
```

**Update Model Setting:**

```bash
curl -X PUT http://localhost:8080/settings/models.default_large \
  -H "Content-Type: application/json" \
  -d '{"value": "openai/gpt-4o"}'
```

---

## Available Settings

### UI Settings

| Key | Type | Default | Options | Description |
|-----|------|---------|---------|-------------|
| `ui.theme` | string | `"dark"` | `"dark"`, `"light"` | Color theme |
| `ui.syntax_highlighting` | boolean | `true` | - | Enable syntax highlighting |
| `ui.show_cost` | boolean | `true` | - | Display API call costs |
| `ui.compact_mode` | boolean | `false` | - | Compact layout |
| `ui.diff_mode` | string | `"unified"` | `"unified"`, `"split"` | Diff view mode |

### Editor Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `editor.context_paths` | array | `[]` | Paths to context files for AI |
| `editor.skills_paths` | array | `[]` | Paths to Agent Skills directories |

### Model Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `models.default_large` | string | `""` | Default large model (format: `provider/model`) |
| `models.default_small` | string | `""` | Default small model (format: `provider/model`) |

### General Settings

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `general.debug` | boolean | `false` | Enable debug logging |
| `general.debug_lsp` | boolean | `false` | Enable LSP debug logging |
| `general.disable_auto_summarize` | boolean | `false` | Disable conversation summarization |
| `general.disable_metrics` | boolean | `false` | Disable usage metrics |

---

## CLI Parity

### Equivalent CLI Commands

Settings endpoints mirror CLI configuration commands:

| CLI Command | API Endpoint | Notes |
|-------------|--------------|-------|
| `crush config get` | `GET /settings` | Settings are subset of config |
| `crush config set compact_mode true` | `PUT /settings/ui.compact_mode` | Key mapping differs |
| `crush config set diff_mode split` | `PUT /settings/ui.diff_mode` | Idempotent |
| `crush config set debug true` | `PUT /settings/general.debug` | Persisted to config |

### Differences from CLI

1. **Key Naming**: Settings use simplified keys (`ui.theme` vs `options.tui.diff_mode`)
2. **Idempotency**: PUT operations are idempotent (safe to retry)
3. **Validation**: Server-side validation with detailed error messages
4. **Response**: API returns updated settings, CLI shows confirmation
5. **Persistence**: Both persist to same underlying config file

---

## Persistence

All settings updates are persisted to the configuration file:

- **Location**: `~/.config/crush/crush.json` (or configured data directory)
- **Atomic Writes**: Changes applied atomically
- **Durability**: Survives application restarts
- **Consistency**: Settings and config views always synchronized

**Note**: Some UI-only settings (theme, syntax_highlighting, show_cost) are client-side preferences not persisted to the config file. These would typically be stored in browser localStorage or similar.

---

## Change Events

Settings updates trigger real-time events:

**Event:** `setting_updated` or `settings_updated`
- Emitted via SSE after successful updates
- Contains the updated setting key/value or full settings
- Allows clients to stay synchronized
- *Prepared for Phase C implementation*

---

## Validation

All setting updates are validated:

1. **Type Checking**: Value must match declared type
2. **Enum Validation**: String settings with options validate against allowed values
3. **Format Validation**: Model settings validate `provider/model` format
4. **Schema Compliance**: All updates checked against setting schemas

**Validation failures return 400 Bad Request with detailed error messages.**

---

## Best Practices

### When to Use Settings vs Config

**Use `/settings` when:**
- Building a user-facing preferences UI
- Working with end-user configurations
- Want simplified, categorized settings
- Need clear schema documentation

**Use `/config` when:**
- Need full technical configuration
- Working with providers, tools, LSP/MCP
- Require detailed control over all options
- Building admin/advanced configuration UI

### Bulk vs Single Updates

**Use `PUT /settings` when:**
- Applying multiple related settings
- Restoring saved preferences
- Importing settings

**Use `PUT /settings/{key}` when:**
- Building toggle/switch UI elements
- Updating one preference at a time
- Want minimal request payload

### Error Handling

Check response status codes and parse error messages:

```javascript
const response = await fetch('/settings/ui.theme', {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ value: 'purple' })
});

if (!response.ok) {
  const error = await response.json();
  console.error('Validation failed:', error.error);
  // "Validation failed: invalid value for ui.theme: must be 'dark' or 'light'"
}
```

---

## Schema Discovery

Use schema endpoints to build dynamic UIs:

```javascript
// Get all settings with their schemas
const schemasResp = await fetch('/settings/schema');
const { settings } = await schemasResp.json();

// Build UI dynamically
settings.forEach(setting => {
  if (setting.type === 'boolean') {
    // Render toggle switch
  } else if (setting.enum) {
    // Render dropdown with enum options
  } else if (setting.type === 'array') {
    // Render list editor
  }
});
```

---

## Examples

### Complete Settings Management Flow

```bash
# 1. Get all settings
curl http://localhost:8080/settings

# 2. Get schema for specific setting
curl http://localhost:8080/settings/ui.compact_mode/schema

# 3. Update single setting
curl -X PUT http://localhost:8080/settings/ui.compact_mode \
  -H "Content-Type: application/json" \
  -d '{"value": true}'

# 4. Update multiple settings
curl -X PUT http://localhost:8080/settings \
  -H "Content-Type: application/json" \
  -d '{
    "ui": {
      "compact_mode": true,
      "diff_mode": "split"
    },
    "general": {
      "debug": false
    }
  }'

# 5. Verify changes
curl http://localhost:8080/settings | jq '.ui'
```

### Building a Settings UI

```javascript
// Fetch current settings
const settings = await fetch('/settings').then(r => r.json());

// Fetch schemas for validation
const schemas = await fetch('/settings/schema')
  .then(r => r.json())
  .then(data => data.settings);

// Update a setting
async function updateSetting(key, value) {
  const response = await fetch(`/settings/${key}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ value })
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error);
  }
  
  return response.json();
}

// Example: Toggle compact mode
await updateSetting('ui.compact_mode', true);
```

---

## Related Endpoints

- `GET /config` - Full technical configuration
- `PUT /config` - Update full configuration
- `PUT /config/field` - Update single config field
- `GET /providers` - List AI providers
- `GET /models` - Get model selections

---

## Implementation Details

- **DTO Layer**: Settings converted from internal config structures
- **Validation**: Reuses config validation with setting-specific rules
- **Persistence**: Uses underlying config persistence layer
- **Schema Generation**: Static schemas built from metadata
- **Performance**: In-memory operations with lazy persistence
