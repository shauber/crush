# API Coverage Gaps - CLI vs API

**Last Updated:** Phase A Complete (Day 4)

## Phase A Implementation Status ✅

### 1. Configuration Management ✅ COMPLETE
```
✅ GET    /config             - Get complete config structure (Day 1)
✅ PUT    /config             - Update multiple config fields (Day 2)
✅ PUT    /config/field       - Update specific field with dot notation (Day 2)
✅ GET    /settings           - User-friendly settings view (Day 3)
✅ GET    /settings/schema    - All settings schemas (Day 3)
✅ GET    /settings/{key}/schema - Per-setting schema (Day 3)
✅ PUT    /settings           - Bulk settings updates (Day 3)
✅ PUT    /settings/{key}     - Single setting updates (Day 3)

⏸️  POST   /providers          - Add new AI provider (Phase D)
⏸️  PUT    /providers/{id}     - Update provider settings (Phase D)
⏸️  DELETE /providers/{id}     - Remove provider (Phase D)
```

**Implemented Features:**
- Complete config read/write with validation
- Schema metadata with type/enum/description
- API key redaction for security
- Settings abstraction layer for user preferences
- Bidirectional sync between config and settings
- Atomic persistence with rollback
- 85+ test cases covering all scenarios

**Documentation:**
- API_CONFIG_ENDPOINTS.md (complete)
- API_SETTINGS_ENDPOINTS.md (complete)

---

## Missing Endpoints Needed for Full CLI Parity

### 2. OAuth & Authentication (Phase D)
```
POST   /oauth/providers/{provider}/start   - Init OAuth flow
POST   /oauth/callback                   - Handle OAuth callbacks
GET    /oauth/accounts                   - List connected accounts
DELETE /oauth/accounts/{id}              - Disconnect account
```

### 3. File & Tool Integration
```
GET    /files/picker                     - File system view (with context)
POST   /files/picker/select              - Select file with context
POST   /files/context                    - Get project context files
GET    /workspace                        - Get current workspace info
POST   /workspace/change                 - Change workspace directory
```

### 4. Tool Call & Session Relationships
```
GET    /sessions/{id}/tool-tree         - Tool call hierarchy/nesting
GET    /sessions/{id}/executions        - Tool execution history
POST   /sessions/{id}/ summarize        - Trigger summarization
GET    /agent/status                    - Real-time agent status
POST   /sessions/{id}/fork             - Fork session (replicate CLI fork)
```

### 5. Permission Management
```
GET    /permissions/pending             - Pending permission requests
POST   /permissions/{id}/grant          - Grant permission
POST   /permissions/{id}/deny           - Deny permission
GET    /permissions/history             - Permission decisions log
```

### 6. MCP & LSP Configuration
```
GET    /mcp/servers                     - List MCP servers
POST   /mcp/servers                     - Add MCP server
PUT    /mcp/servers/{id}                - Update MCP config
DELETE /mcp/servers/{id}                - Remove MCP server
POST   /mcp/servers/{id}/reload         - Reload MCP server
GET    /lsp/servers                     - List language servers
POST   /lsp/servers                     - Configure LSP server
```

### 7. Visual & Context Data
```
GET    /render/markdown                 - Markdown rendering endpoint
POST   /render/markdown                 - Render markdown with syntax
GET    /preview/diff                    - Diff preview generation
POST   /preview/image                   - Image attachment preview
GET    /completions/commands            - Available commands
```

### 8. Settings & Preferences
```
GET    /settings                       - All user preferences
PUT    /settings                       - Bulk update preferences
PUT    /settings/{key}                 - Update specific setting
GET    /settings/{key}/schema         - Get setting metadata
```

## Real-time Event Streaming Extensions

### Currently Missing Event Types
- **agent_status**: Real-time agent busy/idle updates
- **tool_start**: Tool execution started
- **tool_progress**: Tool execution progress
- **permission_request**: Permission dialog triggered
- **config_updated**: Configuration changes
- **oauth_refresh**: OAuth token refresh status
- **session_forked**: Session fork/merge events

### Enhanced SSE Event Format
```javascript
// Agent status
{
  "type": "agent_status",
  "data": {
    "session_id": "uuid",
    "status": "busy|idle",
    "current_tool": "string",
    "progress": 0.75
  }
}

// Tool start
{
  "type": "tool_start", 
  "data": {
    "session_id": "uuid",
    "tool": "bash|edit|download",
    "args": {...},
    "parent_session_id": "uuid" // For nested calls
  }
}

// Permission request
{
  "type": "permission_request",
  "data": {
    "id": "permission-id",
    "type": "file_write|command_exec|url_fetch",
    "context": {...},
    "prompt": "Do you want to allow..."
  }
}
```

## Advanced Session Features

### Missing Session APIs
- **Session relationships**: Track parent/child sessions from tool calls
- **Session templates**: Pre-configured prompt templates
- **Session bundles**: Group related sessions
- **Session sharing**: Export/import functionality
- **Session performance**: Token/speed analysis endpoints

### Contextual API Extensions
- **Project context discovery**: Auto-detect project type/settings
- **Git integration**: Branch/commit context for sessions
- **Environment detection**: Language/runtime detection
- **Dependency analysis**: Scan project dependencies

## Configuration Schema Exposure

### TUI-Readable Config Structure ✅ IMPLEMENTED
```json
{
  "providers": {
    "anthropic": {
      "id": "anthropic",
      "name": "Anthropic",
      "type": "anthropic",
      "enabled": true,
      "api_key": "encrypted",  // Always redacted
      "models": ["claude-3.5-sonnet", "claude-3-opus", "claude-3-haiku"],
      "has_oauth": false
    }
  },
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
        "filetypes": ["go", "mod"],
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
      "temperature": 0.7
    },
    "small": {
      "model": "claude-3-haiku",
      "provider": "anthropic"
    }
  },
  "options": {
    "context_paths": [".cursorrules", "CRUSH.md"],
    "debug": false,
    "disable_metrics": false
  }
}
```

**Note:** Actual implementation in Phase A exceeds this structure with full config exposure.

---

## Priority Implementation Order

1. **Phase A** ✅ COMPLETE - Config/get, basic settings
2. **Phase B** (Tools): File picker, workspace context
3. **Phase C** (Real-time): Enhanced SSE events, agent status
4. **Phase D** (Auth): OAuth flows for providers
5. **Phase E** (Visual): Rendering/Preview endpoints
6. **Phase F** (Advanced): MCP/LSP management

This provides a comprehensive roadmap for API parity with the CLI experience.