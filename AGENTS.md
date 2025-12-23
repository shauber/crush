# AGENTS.md - Crush AI Development Guide

This document helps agents (AI agents, developers, and tools) work effectively in the Crush codebase. It's automatically loaded as context by the system through `AGENTS.md` recognition patterns.

## Project Overview

**Crush** is a terminal-based AI assistant for software development built in Go. It provides:
- Interactive TUI chat interface with AI models
- REST API server for external integration
- Real-time streaming (SSE) for chat updates
- LSP integration for code analysis
- Multi-provider AI support (OpenAI, Anthropic, Google, etc.)
- SQLite-based persistence with sessions, messages, and todos

## Essential Commands

### Build & Run
```bash
# Standard operations
go run .                    # Run the application
go build .                  # Build binary
task build                  # Build with version flags
task run                    # Run via Taskfile
task dev                    # Run with profiling enabled (localhost:6060)

# Development
task test                   # Run all tests
task test -update           # Update golden test files
task lint:fix               # Auto-fix linting issues
task fmt                    # Format code with gofumpt
task hyper                  # Update Hyper provider configuration
```

### API Server
```bash
# Run the API server
crush server [port]         # Start REST API + SSE streaming
crush server 8080          # Example with explicit port
```

### Test Patterns
```bash
# Run specific test categories
go test ./... -v           # All tests with verbose output
go test ./internal/agent/... -count=1   # Deep reset for agent tests

# Update golden files
ο go test ./... -update
ο task test:record         # Auto-update VCR cassettes for agent tests
```

## Architecture & Structure

### Directory Layout
```
crush/
├── internal/
│   ├── agent/          # Core AI agent orchestration
│   ├── api/            # REST API endpoints + SSE streaming
│   ├── app/            # Application orchestration layer
│   ├── config/         # Configuration management
│   ├── tui/            # Terminal UI components (Bubble Tea)
│   ├── cmd/            # CLI commands (root.go = main entry)
│   └── special modules/
│       ├── db/         # SQLite + Goose migrations
│       ├── session/    # Session management
│       ├── message/    # Message handling
│       └── permission/ # Permission system
├── sqlc.yaml           # SQL code generation
├── schema.json         # JSON schema for config
├── Taskfile.yaml       # Build automation
└── go.mod              # Go dependencies
```

### Key Patterns

#### 1. Dependency Injection Architecture
- `app.App` is the central orchestrator - passed to everything
- Layered services: Database → Services → Handlers/TUI
- Context-based cancellation throughout

#### 2. Session-Based Operations
```go
// Every AI operation needs a session ID
ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, sessionID)
result := agent.Run(ctx, SessionAgentCall{...})
```

#### 3. Provider Abstraction
- Fantasy-based provider system for AI models
- Centralized in `internal/config/provider.go`
- Mock providers for testing via `config.UseMockProviders = true`

### Core Service Layering

#### Session Management (`internal/session/`)
- Unique ID per conversation thread
- Persisted to SQLite with Goose migrations
- Contains todos, model preferences, history

#### Agent System (`internal/agent/`)
- `SessionAgent` interface for AI operations
- Tool system for code operations (ls, grep, edit, etc.)
- Automatic summarization and queuing
- Real-time streaming via SSE

#### TUI Architecture (`internal/tui/`)
- Bubble Tea-based terminal interface
- Component-based with composable models
- Real-time pub/sub updates (internal/pubsub/)
- Component hierarchy: pages → components → sub-components

## Code Writing Conventions

### Style Guidelines (from CRUSH.md)
- **Imports**: Standard Go grouping (stdlib, external, internal)
- **Formatting**: Always use `gofumpt` (stricter than gofmt)
- **Types**: Use type aliases for clarity (`type AgentName string`)
- **Naming**: PascalCase exported, camelCase unexported
- **Context**: Always first parameter name `ctx`
- **Error handling**: Wrap with `fmt.Errorf("context: %w", err)`
- **Testing**: Use testify `require`, make tests parallel, use `t.TempDir()`

### Testing Conventions
```go
// Golden file tests for UI components
func TestComponent(t *testing.T) {
    t.Parallel()
    
    // Use charmtone for colors
    styles := styles.NewCharmtone()
    
    // Component uses test data
    component := NewComponent()
    
    // Compare with golden file
    golden.RequireEqual(t, component.View())
}

// Mock provider testing
func TestWithMock(t *testing.T) {
    original := config.UseMockProviders
    config.UseMockProviders = true
    defer func() { config.UseMockProviders = original; config.ResetProviders() }()
}
```

### Tool Development

When adding new tools for the agent system:

1. **Tool Location**: `internal/agent/tools/[toolname].go`
2. **Interface**: Implement `fantasy.AgentTool`
3. **Context keys**: Use package constants from `tools` package
4. **Status updates**: Broadcast via pubsub system
5. **Documentation**: Add `.md` description file

#### Example Tool Structure
```go
//go:embed [tool].md
var description []byte

func NewTool(service Service) fantasy.AgentTool {
    return fantasy.NewAgentTool(
        "toolname",
        string(description),
        func(ctx context.Context, params Params) (fantasy.ToolResponse, error) {
            sessionID := tools.GetSessionFromContext(ctx)
            // Implementation here
        },
    )
}
```

## Development Workflow

### Common Tasks

#### Adding New Features
1. Identify appropriate package (`internal/agent/` for AI, `internal/tui/` for UI)
2. Use dependency injection via `app.App`
3. Write tests with VCR cassettes for external dependencies
4. Update golden files with `make test-update`
5. Run `task lint:fix` before committing

#### Database Changes
1. Edit SQL files in `internal/db/sql/`
2. Create new migration: `touch internal/db/migrations/YYYYMMDDHHMMSS_description.sql`
3. `goose up` (or equivalent) to apply changes
4. Regenerate SQLc code if needed

#### Configuration Changes
1. Update `internal/config/config.go` structs
2. Regenerate schema: `go run main.go schema > schema.json`
3. Update relevant documentation

### Debugging Tools

#### Performance Analysis
```bash
task dev                    # Runs with pprof enabled
# Open http://localhost:6060/debug/pprof/
go tool pprof -http :6061 'http://localhost:6060/debug/pprof/heap'
```

#### Database Inspection
```bash
# Find data directory
echo ~/.config/crush    # Linux/Mac
# Main SQLite file: ~/.config/crush/crush.db
```

#### Development CLI
```bash
# Quick test without full UI
crush run "explain this codebase structure"  # Non-interactive mode
crush -d                                        # Debug mode with verbose logging
```

## Gotchas & Non-Obvious Patterns

### 1. Session Context Timeout
All context-based operations are cancellable - agent operations respect context cancellation.

### 2. Golden File Testing
UI components use golden files in `testdata/` for visual regression testing. Always run with:
```bash
go test ./... -update   # Regenerate golden files after changes
```

### 3. VCR Cassettes
Network tests use VCR to record HTTP interactions. Cassettes are in:
- `internal/agent/testdata/` for AI provider calls
- Never commit real API keys

### 4. TUI Keyboard Handling
Special keyboard combinations handled in `KeyBindings` per component.
Mouse events have 15ms debouncing in `tui.MouseEventFilter()`.

### 5. Provider Configuration Order
Configuration resolution:
1. Environment variables (highest priority)
2. CLI flags
3. Config file
4. Embedded defaults
5. Mock providers (when enabled)

### 6. Development Mode Detection
- Debug mode: `crush -d` or `CRUSH_PROFILE=true`
- YOLO mode: `crush -y` (skips all permission prompts - dangerous!)

## Quick Reference

### File Patterns
- **Configs**: Look for `.env`, `crush.json`, or `~/.config/crush/`
- **Sessions**: Identified by `session.ID` string throughout codebase
- **Messages**: Always linked to session via `message.Message` structs
- **Tools**: Located in `internal/agent/tools/` with `.md` descriptions

### Debugging Quick Commands
```bash
grep -r "SessionIDContextKey" internal/   # Find session usage
grep -r "NewAgentTool" internal/          # Find tool definitions
find internal/ -name "*.golden" -type f   # List golden files
grep -r "\.SetCoreColors\|charmtone\."   # Identify styling

# Live debugging
CRUSH_PROFILE=true go run .            # Enable pprof
```

### Common Development Paths
- **AI Features**: `internal/agent/` → add to agent tools
- **UI Features**: `internal/tui/` → `components/chat/` for chat-specific
- **API Features**: `internal/api/` → `handlers.go` for endpoints
- **Data Layer**: `internal/db/` → SQL + sqlc + Goose migrations

## CRITICAL: UPSTREAM CODE PROHIBITIONS

**🛑 NEVER modify code in the `main` branch or upstream project directly.**

**⛔ ABSOLUTELY PROHIBITED:**
- Changing functions, types, or interfaces in existing files
- Modifying build configurations, dependencies, or core logic (even to fix bugs)
- Altering anything in `internal/`, `cmd/`, or any existing Go files
- Updating golden files or test data from the upstream project
- Changing database schemas, SQL queries, or migrations
- Modifying the agent framework, TUI components, or API endpoints
- Any refactoring or reorganization of existing code

**✅ ALLOWED:**
- Creating NEW files (config, documentation, temporary files)
- Reading existing code to understand patterns and architecture
- Running tests and analysis
- Using the existing tools and API provided by the system

**🔴 REQUIRED:**
If any situation arises that appears to require edits to upstream code:
1. **STOP IMMEDIATELY**
2. **DISCUSS WITH USER** - describe what you've found and why you think upstream changes are needed
3. **Wait for explicit user approval** before making any modifications
4. **Document the rationale** for any approved changes

This prohibition applies even if:
- Tests are failing
- You find bugs or issues
- Changes seem small or obvious
- It's "just" fixing typos or formatting
- The codebase appears to need improvement

*The upstream code is considered read-only unless explicitly discussed and approved.*

## Integration Notes

### AI Provider Support
The system supports multiple AI providers through Fantasy:
- OpenAI, Anthropic, Google, Bedrock, OpenRouter
- Model preferences stored per session
- Hyper-optimized provider configuration embedded

### Security & Permissions
- Granular tool permissions via `internal/permission/`
- File system access controlled via permission system
- Environment variable `CRUSH_DISABLE_METRICS=true` to opt-out

### Cross-Platform Notes
- POSIX-specific code in `internal/fsext/`
- Windows terminal support via colorprofile
- OpenBSD support via `openbsd/` directory

This AGENTS.md file is automatically loaded by Crush when working in this repository. It provides complete context for understanding and extending the codebase.