# API Enhancement Plan - From Basic to CLI Parity

## Overview
Current API provides basic CRUD but lacks the rich interactive features that make the CLI powerful. This plan focuses on **incremental expansion** to enable full CLI-like web experience.

## Phase 1: Real-Time Foundation (Priority 1)
**Goal**: Close the real-time gap between CLI interactive experience and API

### New Endpoints
```
GET    /sessions/{id}/live-stream       # Enhanced SSE with tool events
GET    /sessions/{id}/executions        # Tool call execution history
GET    /agent/status                    # Global agent busy/idle state
POST   /sessions/{id}/cancel-tool       # Cancel specific tool execution
```

### Enhanced SSE Events
```javascript
// Tool start/progress
{
  "type": "tool_start",
  "message_id": "uuid",
  "tool": "bash|edit|download",
  "args": {...},
  "timestamp": 1234567890
}

{
  "type": "tool_progress",
  "message_id": "uuid", 
  "stdout": "partial output...",
  "stderr": "",
  "progress": 0.65,
  "timestamp": 1234567890
}

{
  "type": "tool_complete", 
  "message_id": "uuid",
  "success": true,
  "exit_code": 0,
  "duration": 2.5,
  "result": {"type": "success", "content": "..."}
}
```

**Implementation approach**: Extend existing pubsub to emit tool-specific events rather than just message updates.

## Phase 2: Configuration API (Priority 2)
**Goal**: Enable runtime configuration changes like CLI

### New Endpoints
```
GET    /config               # Full configuration as JSON schema
PUT    /config               # Bulk update configuration
PUT    /config/{key}         # Update specific config field
POST   /config/validate      # Validate config changes
GET    /config/schema        # Get config validation schema
```

### Provider Management
```
GET    /providers            # Detailed provider configs
POST   /providers            # Add new provider  
PUT    /providers/{id}       # Update provider
DELETE /providers/{id}       # Remove provider
POST   /providers/test       # Test provider connectivity
```

## Phase 3: Files & Workspace (Priority 3)
**Goal**: Enable file operations and workspace context like CLI

### File System API
```
GET    /files/tree           # File system tree view
POST   /files/upload         # Upload files (limited size)
POST   /files/create         # Create new file
PUT    /files/{path}         # Write file content
POST   /files/validate       # Validate file operations safety
GET    /workspace/root       # Current working directory
POST   /workspace/change     # Change root directory
```

### Context Discovery
```
GET    /context/projects     # Detect project type/rules
GET    /context/suggestions  # Recommend relevant files
POST   /context/add          # Add files to session context
DELETE /context/remove       # Remove from session context
```

## Phase 4: Session Relationships (Priority 4)
**Goal**: Handle nested tool calls and session forking

### Session Relationships
```
GET    /sessions/{id}/tree   # Parent/child session tree
POST   /sessions/{id}/fork   # Fork session (like CLI fork)
POST   /sessions/{id}/merge  # Merge sessions back together
GET    /sessions/{id}/related # Related child sessions
```

### Session Actions
```
POST   /sessions/{id}/summarize   # Trigger summarization
POST   /sessions/bulk            # Bulk operations (delete, archive)
GET    /sessions/templates       # Session templates
POST   /sessions/from-template   # Create from template
```

## Phase 5: Permissions & Safety (Priority 5)
**Goal**: Handle CLI's permission prompts through API

### Permission API
```
GET    /permissions/pending       # Current pending requests
POST   /permissions/request       # Submit permission request
POST   /permissions/{id}/grant    # Grant permission
POST   /permissions/{id}/deny     # Deny permission
GET    /permissions/policy        # Current safety rules
PUT    /permissions/policy        # Update safety rules
```

### Safety Validation
```
POST   /safety/validate           # Validate specific operation
GET    /safety/banned-commands    # Get banned commands list
POST   /safety/test               # Test safety validation
```

## Phase 6: Advanced Tooling (Priority 6)
**Goal**: MCP/LSP configuration like CLI

### MCP Management
```
GET    /mcp/servers               # List configured servers
POST   /mcp/servers               # Add MCP server
PUT    /mcp/servers/{id}          # Update server config
DELETE /mcp/servers/{id}          # Remove server
POST   /mcp/servers/{id}/start    # Start server
POST   /mcp/servers/{id}/stop     # Stop server
```

### LSP Configuration
```
GET    /lsp/servers               # List configured language servers
POST   /lsp/servers               # Add LSP server
POST   /lsp/refresh               # Refresh LSP connections
GET    /lsp/diagnostics           # Get LSP diagnostics
```

## Implementation Strategy

### Week 1: Core Real-Time
1. **Day 1**: Enhance SSE with tool events (Phase 1)
2. **Day 2**: Add execution tracking endpoints
3. **Day 3**: Implement agent status streaming

### Week 2: Configuration + Files  
1. **Day 1**: Configuration read/write endpoints (Phase 2)
2. **Day 2**: File management safety layer (Phase 3)
3. **Day 3**: Workspace context APIs

### Week 3: Advanced Features
1. **Day 1**: Session relationships (Phase 4)
2. **Day 2**: Permission system (Phase 5)
3. **Day 3**: MCP/LSP Management (Phase 6)

## Technical Architecture

### Existing Patterns to Extend
- **PubSub events**: Use existing `internal/pubsub` infrastructure
- **Session management**: Extend existing `app.Sessions` service
- **Tool coordination**: Leverage `AgentCoordinator` hooks
- **Safety layer**: Reuse existing permission and ban mechanisms

### Data Models Needed
```go
// Tool execution tracking
type ToolExecution struct {
    ID        string    `json:"id"`
    SessionID string    `json:"session_id"`
    Tool      string    `json:"tool"`
    Args      any       `json:"args"`
    Started   time.Time `json:"started"`
    Finished  time.Time `json:"finished,omitempty"`
    Success   bool      `json:"success"`
    Output    string    `json:"output"`
    Error     string    `json:"error,omitempty"`
}

// Permission request
type PermissionRequest struct {
    ID        string    `json:"id"`
    SessionID string    `json:"session_id"`
    Type      string    `json:"type"`
    Resource  string    `json:"resource"`
    Args      any       `json:"args"`
    Created   time.Time `json:"created"`
    Status    string    `json:"status"`
}
```

## Security Considerations

### Input Validation
- All file operations through safety validation layer
- Path restriction to configured workspace
- Rate limiting on tool executions
- Size limits on file uploads

### Authentication
- No user auth (per project scope)
- Usage tracking per session
- Audit logging for permission decisions
- Config secrets encryption at rest

## Deployment Impact

### Minimal Required Changes
- No database schema changes
- Uses existing safety/security layers
- Extends pubsub rather than replacing it
- Backward compatible with current API

### Testing Strategy
- Unit tests for each new endpoint
- Integration tests mimicking CLI workflows
- Security validation test suite
- Performance testing for file operations

## Success Criteria

### Phase 1 Complete When
- [ ] Agent shows live tool execution via SSE
- [ ] Tool output streams in real-time
- [ ] Session busy state visible via API
- [ ] Cancel tool execution working

### Final Complete When
- [ ] All CLI features available via API
- [ ] Web UI can replicate full CLI experience
- [ ] Real-time interactive features working
- [ ] Configuration changes persistent
- [ ] Security validation passes

## Rollback Strategy

Each phase is additive and independent:
- **Phase 1**: Disable new SSE events
- **Phases 2-6**: Remove new endpoints temporarily
- **All phases**: APIs are non-breaking to existing calls

This plan transforms the basic API into a comprehensive CLI-equivalent interface while maintaining security and backward compatibility.