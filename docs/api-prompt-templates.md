# API Design: Prompt Template Switching

## Overview

Design for REST API endpoints to allow switching between existing prompt templates for agents. This enables external tools to dynamically change the system prompt used by agents without requiring code changes or restarts.

## Current Architecture

### Template System
- Templates defined in `internal/agent/templates/`
  - `coder.md.tpl` - Main development agent (default)
  - `task.md.tpl` - Task execution agent
  - `initialize.md.tpl` - Session initialization
  - `summary.md` - Message summarization
  - `title.md` - Title generation
  - Others: `agent_tool.md`, `agentic_fetch.md`, `agentic_fetch_prompt.md.tpl`

- Templates embedded at compile time via `//go:embed` in `internal/agent/prompts.go`
- Built using `prompt.NewPrompt()` with runtime context (working dir, git status, config, etc.)
- Currently hardcoded: `coderPrompt()` is always used in `coordinator.NewCoordinator()` (line 102)

### Agent System
- `Agent` config struct exists (`internal/config/config.go:319-340`) with:
  - `ID`, `Name`, `Description`
  - `Model` (large/small)
  - `AllowedTools`, `AllowedMCP`
  - `ContextPaths` (custom context file overrides)
  - **Missing**: No template/prompt selection field

- Current agents defined as constants:
  - `AgentCoder` = "coder"
  - `AgentTask` = "task"

- `Config.Agents map[string]Agent` exists (line 377) but only `AgentCoder` is used

### Integration Points
1. **Session Level**: Sessions don't track which template they use
2. **Coordinator Level**: `coordinator.NewCoordinator()` creates agent with hardcoded `coderPrompt()`
3. **Agent Level**: `SessionAgent` is built once at coordinator creation, not per-session

## Proposed Design

### 1. Template Registry

Add template registry to expose available templates:

```go
// internal/agent/prompts.go

// TemplateInfo describes an available prompt template
type TemplateInfo struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Category    string `json:"category"` // "agent", "utility", "internal"
}

// AvailableTemplates returns metadata about all available templates
func AvailableTemplates() []TemplateInfo {
    return []TemplateInfo{
        {
            ID:          "coder",
            Name:        "Coder Agent",
            Description: "Main development agent for code analysis, editing, and execution",
            Category:    "agent",
        },
        {
            ID:          "task",
            Name:        "Task Agent",
            Description: "Specialized agent for executing specific tasks",
            Category:    "agent",
        },
        {
            ID:          "initialize",
            Name:        "Initialize Agent",
            Description: "Session initialization and project setup",
            Category:    "utility",
        },
        {
            ID:          "summary",
            Name:        "Summary Generator",
            Description: "Message and conversation summarization",
            Category:    "utility",
        },
        {
            ID:          "title",
            Name:        "Title Generator",
            Description: "Session title generation",
            Category:    "utility",
        },
    }
}

// GetPromptByID returns a prompt builder for the given template ID
func GetPromptByID(id string, opts ...prompt.Option) (*prompt.Prompt, error) {
    switch id {
    case "coder":
        return coderPrompt(opts...)
    case "task":
        return taskPrompt(opts...)
    case "initialize":
        return prompt.NewPrompt("initialize", string(initializePromptTmpl), opts...)
    default:
        return nil, fmt.Errorf("unknown template ID: %s", id)
    }
}
```

### 2. Session Template Tracking

Extend session to track which template it uses:

```sql
-- internal/db/migrations/YYYYMMDDHHMMSS_add_session_template.sql
ALTER TABLE sessions ADD COLUMN prompt_template TEXT DEFAULT 'coder';
```

```go
// internal/session/session.go
type Session struct {
    // ... existing fields ...
    PromptTemplate string `json:"prompt_template"` // New field
}
```

### 3. Configuration Extension

Add template selection to Agent config:

```go
// internal/config/config.go
type Agent struct {
    ID               string              `json:"id,omitempty"`
    Name             string              `json:"name,omitempty"`
    Description      string              `json:"description,omitempty"`
    Disabled         bool                `json:"disabled,omitempty"`
    Model            SelectedModelType   `json:"model" jsonschema:"required"`
    PromptTemplate   string              `json:"prompt_template,omitempty" jsonschema:"description=The prompt template to use for this agent,enum=coder,enum=task,default=coder"` // NEW
    AllowedTools     []string            `json:"allowed_tools,omitempty"`
    AllowedMCP       map[string][]string `json:"allowed_mcp,omitempty"`
    ContextPaths     []string            `json:"context_paths,omitempty"`
}
```

### 4. API Endpoints

#### GET /templates
List all available prompt templates.

**Response:**
```json
{
  "templates": [
    {
      "id": "coder",
      "name": "Coder Agent",
      "description": "Main development agent for code analysis, editing, and execution",
      "category": "agent"
    },
    {
      "id": "task",
      "name": "Task Agent",
      "description": "Specialized agent for executing specific tasks",
      "category": "agent"
    }
  ]
}
```

#### GET /sessions/{id}/template
Get the current template for a session.

**Response:**
```json
{
  "session_id": "abc123",
  "template_id": "coder",
  "template_name": "Coder Agent"
}
```

#### PUT /sessions/{id}/template
Change the template for a session.

**Request:**
```json
{
  "template_id": "task"
}
```

**Response:**
```json
{
  "session_id": "abc123",
  "template_id": "task",
  "template_name": "Task Agent",
  "message": "Template updated successfully. New messages will use this template."
}
```

**Validation:**
- Template ID must exist in registry
- Session must not be currently processing (busy)

**Behavior:**
- Updates session record with new template ID
- Next agent invocation will rebuild agent with new template
- Does NOT rebuild existing agent (lightweight approach)
- Emits SSE event: `template_changed`

#### GET /config/agents
List configured agents (extends existing config endpoints).

**Response:**
```json
{
  "agents": {
    "coder": {
      "id": "coder",
      "name": "Coder",
      "model": "large",
      "prompt_template": "coder",
      "disabled": false
    },
    "task": {
      "id": "task",
      "name": "Task Executor",
      "model": "small",
      "prompt_template": "task",
      "disabled": true
    }
  }
}
```

### 5. Implementation Strategy

#### Phase 1: Registry & Read Operations (Safe)
1. Add `TemplateInfo` struct and `AvailableTemplates()` function
2. Add `GetPromptByID()` helper function
3. Implement `GET /templates` endpoint
4. **No upstream conflicts**: All new code

#### Phase 2: Session Tracking (Requires Discussion)
1. Database migration to add `prompt_template` column
2. Update session struct and queries
3. Implement `GET /sessions/{id}/template` endpoint
4. **Potential conflicts**: Session schema changes

#### Phase 3: Dynamic Switching (Requires Discussion)
1. Modify coordinator to accept template ID parameter
2. Implement `PUT /sessions/{id}/template` endpoint
3. Add SSE event for template changes
4. **Potential conflicts**: Coordinator initialization logic

#### Phase 4: Config Integration (Safe)
1. Add `PromptTemplate` field to Agent config
2. Implement `GET /config/agents` endpoint
3. Allow setting default template via config
4. **No upstream conflicts**: Config extension pattern

## Constraints & Considerations

### Safety
- **READ-ONLY Phase 1** can be implemented immediately (no upstream risk)
- **Phases 2-3** modify core session/agent lifecycle (requires user approval)
- **Phase 4** extends config schema (follows existing patterns)

### Scope Limitations
- Only switch between **existing** templates (no custom template upload)
- Templates remain compile-time embedded (no runtime loading)
- Template content cannot be modified via API
- Each session uses one template at a time

### Performance
- Template switching does NOT rebuild the current agent immediately
- New template takes effect on next `coordinator.Run()` call
- Lightweight operation (just database update + config change)

### Use Cases
- External UI could switch between "focused coder" and "task executor" modes
- Different sessions could use different templates simultaneously
- API clients could optimize for specific workflows (debugging vs. exploration)

### Future Extensions
- Per-message template override (advanced)
- Template inheritance/composition
- Runtime template customization (variables/flags)
- Template versioning

## Alternative Approaches Considered

### 1. Agent Switching (Instead of Template Switching)
- Could expose `coordinator.SetMainAgent(agentID)` API
- More aligned with existing `Config.Agents` map
- **Rejected**: Agent vs. Template distinction is unclear; templates are the actual differentiator

### 2. Provider-Level SystemPromptPrefix
- Already exists: `ProviderConfig.SystemPromptPrefix` (config.go:111)
- Allows per-provider prompt customization
- **Limitation**: Only adds prefix, doesn't replace entire template

### 3. Context Paths Override
- Already exists: `Agent.ContextPaths` (config.go:339)
- Allows custom memory files per agent
- **Limitation**: Augments template, doesn't replace it

## Open Questions

1. Should template switching rebuild the agent immediately or lazily?
   - **Proposed**: Lazy (next Run() call) for simplicity
   
2. Should we track template history per session?
   - **Proposed**: No, just current template
   
3. Should message history include which template generated each response?
   - **Proposed**: Phase 2+ consideration

4. Should config file changes to agent templates hot-reload?
   - **Proposed**: No, requires restart (existing pattern)

## Next Steps

1. Get approval for Phase 1 (safe, read-only registry)
2. Discuss Phases 2-3 (session/coordinator modifications)
3. Implement Phase 1 to validate API design
4. Test with external API client

---

**Status**: Design complete, awaiting approval for implementation
**Author**: AI Assistant
**Date**: 2026-01-11
