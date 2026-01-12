# Template Switching Feature - Implementation Summary

**Branch**: `feat/template-switching`  
**Base**: `upstream/main` (commit 0868681f)  
**Status**: ✅ Complete and tested

## Overview

Implemented prompt template switching functionality that allows users to select different agent behaviors via CLI flag or configuration file. This addresses community feature requests for plan mode (#1734) and subagent support (#1807).

## Implementation Details

### 1. Template Registry (`internal/agent/templates.go`)

Created a new module that:
- Enumerates available templates via `AvailableTemplates()`
- Loads templates by ID via `GetPromptByID()`
- Provides metadata (ID, Name, Description) for each template
- Currently supports: `coder`, `task`

**Benefits:**
- Discoverability for future TUI picker
- Type-safe template loading
- Easy to extend with new templates

### 2. Config Schema Extension (`internal/config/config.go`)

Added `PromptTemplate` field to `Agent` struct:
```go
PromptTemplate string `json:"prompt_template,omitempty"`
```

**Benefits:**
- Persistent template selection per agent
- Config-driven agent behavior
- JSON schema aware for validation

### 3. CLI Integration (`internal/cmd/root.go`)

Added `--template` / `-t` flag:
- Works with both interactive and `run` modes
- Overrides config file settings
- Validates against available templates

**Usage examples:**
```bash
crush run --template task "Create a plan"
crush --template task
```

### 4. Coordinator Integration (`internal/agent/coordinator.go`)

Modified `NewCoordinator()` to:
- Read template ID from agent config
- Use `GetPromptByID()` instead of hardcoded `coderPrompt()`
- Default to "coder" for backward compatibility
- Provide clear error messages for invalid templates

### 5. Test Coverage (`internal/agent/templates_test.go`)

Added comprehensive tests:
- Template enumeration validation
- Template loading for valid IDs
- Error handling for invalid templates
- All tests pass ✅

## Usage

### CLI Usage

```bash
# Default (coder template)
crush run "explain this code"

# Use task template
crush run --template task "create implementation plan"

# Interactive with template
crush --template task

# Invalid template (errors gracefully)
crush --template invalid "hello"
# Error: unknown template ID: invalid (available: coder, task)
```

### Config File Usage

```json
{
  "agents": {
    "coder": {
      "prompt_template": "task",
      "model": "large"
    }
  }
}
```

### Help Text

```
FLAGS
  -t --template    Agent prompt template (coder, task)
```

## File Changes

```
 internal/agent/coordinator.go    |  7 +++--
 internal/agent/templates.go      | 42 ++++++++++++++++++++++
 internal/agent/templates_test.go | 84 ++++++++++++++++++++++++++++++++++++
 internal/cmd/root.go             | 18 +++++++++
 internal/config/config.go        |  3 ++
 5 files changed, 151 insertions(+), 3 deletions(-)
```

## Testing

All tests pass:
```bash
go test ./internal/agent -run TestAvailableTemplates  # ✅ PASS
go test ./internal/agent -run TestGetPromptByID       # ✅ PASS
go build .                                             # ✅ Success
```

## Backward Compatibility

✅ **Fully backward compatible:**
- Defaults to "coder" template if unspecified
- Existing configs work without changes
- No breaking changes to APIs or behavior

## Future Extensions

This implementation enables:

1. **Plan/Build Mode** (Issue #1734):
   - Add `planner` template with restricted tools
   - Users switch via `--template planner`
   - No TUI changes required

2. **Custom Templates**:
   - Users can add custom `.md.tpl` files
   - Registry can scan directories
   - Config points to custom templates

3. **TUI Integration**:
   - Template picker dialog using `AvailableTemplates()`
   - Keyboard shortcut to switch (Shift+Tab)
   - Session-level template switching

4. **API Extension** (for our fork):
   - `GET /templates` - List available
   - `PUT /sessions/{id}/template` - Switch template
   - Already has the infrastructure

## Contribution Plan

**Ready for upstream submission:**

1. **What to contribute:**
   - All 5 files in this commit
   - Clean, self-contained feature
   - No dependencies on API code
   - Addresses real community issues

2. **PR Message:**
   > This PR adds prompt template switching via CLI and config file, addressing
   > community requests for plan mode (#1734) and subagent support (#1807).
   > 
   > Provides infrastructure for multiple agent behaviors without requiring TUI
   > changes. Users can switch templates via `--template` flag or config file.
   > 
   > Fully backward compatible - defaults to existing "coder" template.

3. **Discussion points:**
   - Would maintainers prefer TUI integration first?
   - Interest in plan/build mode as follow-up?
   - Naming conventions for templates?

## Technical Notes

**Why this design?**
- Minimal changes to existing code
- Follows existing patterns (Agent config, CLI flags)
- Easy to test and maintain
- Extensible for future needs

**What's NOT included:**
- Session-level template tracking (requires DB migration)
- Dynamic template switching (requires coordinator rebuild)
- TUI integration (no changes to TUI code)
- API endpoints (upstream doesn't have API)

These can be added later without breaking changes.

## Verification

Branch is ready for:
- ✅ Local testing
- ✅ Code review
- ✅ Upstream PR submission
- ✅ Merge into feature branch

**Next steps:**
1. Test with real workflows
2. Get feedback from maintainers
3. Extend with additional templates if accepted
4. Consider TUI integration as follow-up

---

**Implementation Date**: January 11, 2026  
**Commit**: b4995648  
**Branch**: feat/template-switching (based on upstream/main)
