# API Coverage Gaps - CLI vs API

## Missing Endpoints Needed for Full CLI Parity

### 1. Configuration Management
```
GET    /config             - Get complete config structure
PUT    /config             - Update multiple config fields
PUT    /config/field       - Update specific field (e.g., theme, model defaults)
POST   /providers          - Add new AI provider
PUT    /providers/{id}     - Update provider settings
DELETE /providers/{id}     - Remove provider
```

### 2. OAuth & Authentication
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

### TUI-Readable Config Structure
```json
{
  "providers": {
    "anthropic": {
      "enabled": true,
      "api_key": "encrypted",
      "models": ["claude-3.5-sonnet", "claude-3-opus"],
      "default": "claude-3.5-sonnet"
    }
  },
  "tools": {
    "bash": {
      "banned_commands": ["sudo", "rm -rf"],
      "timeout": 30
    },
    "mcp_servers": [...],
    "lsp_config": {...}
  },
  "ui": {
    "theme": "dark|light",
    "syntax_highlighting": true,
    "show_cost": true
  }
}
```

## Priority Implementation Order

1. **Phase A** (Critical): Config/get, basic settings
2. **Phase B** (Tools): File picker, workspace context
3. **Phase C** (Real-time): Enhanced SSE events, agent status
4. **Phase D** (Auth): OAuth flows for providers
5. **Phase E** (Visual): Rendering/Preview endpoints
6. **Phase F** (Advanced): MCP/LSP management

This provides a comprehensive roadmap for API parity with the CLI experience.