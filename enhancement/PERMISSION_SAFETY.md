# Permission & Safety System

## Scene: When AI requests touch the system

### CLI Reality - The Scary Moment
**Developer runs**: "Can I sudo apt update your system?"
- CLI shows: "💀 DANGER: Blocking sudo command"
- Developer types "y" to allow specific actions
- Has whitelist of safe commands
- Can configure "always allow" settings

**Web Reality - Security Nightmare**
- No way to interactive permission approvals
- AI could potentially run dangerous commands
- No user feedback on safety decisions
- Stuck with all-or-nothing permissions

### Security Story Mapping

#### Story 1: File Write Permission
```
AI: "I need to update requirements.txt"
Web UI: Shows permission popup
User: [See diff first] [Approve] [Deny] [Always allow]
Web UI: Applies file lock, shows progress
Security: Validates against project boundaries
```

#### Story 2: Network Access
```
AI: "I need to download malware scanner"
Web UI: Shows request details
User: [See URL] [View docs] [Approve] [Sandbox] [Deny]
Security: Validates against allowlist
```

## Permission Flow API

### Request Permission
```http
POST /permissions/request
{
  "session_id": "uuid",
  "type": "file_write",
  "resource": "/app/config.py",
  "action": "append",
  "details": {
    "proposed_lines": 5,
    "lines_before": 12,
    "preview": "\nDATABASE_URL = ...\n"
  }
}

Response:
{
  "request_id": "perm-123", 
  "status": "pending",
  "expires_at": "2024-01-15T15:30:00Z",
  "level": "medium"
}
```

### Permission Decision
```http
POST /permissions/perm-123/respond
{
  "decision": "grant",  // "deny", "once", "always"
  "reason": "File belongs to project, diff looks good"
}
```

### Real-time Permission Events
```http
GET /sessions/{id}/live-stream

event: permission_request
data: {
  "request_id": "perm-123",
  "type": "file_write",
  "resource": "/app/models.py",
  "preview": {
    "diff": "+234 lines",
    "summary": "Add new model code"
  },
  "context": {
    "session_title": "Fix database issue",
    "confidence": 0.85
  }
}
```

## Safety Levels System

### Permission Categories
- **🟢 GREEN**: Safe commands (ls, cat, grep)
- **🟡 YELLOW**: File modifications within project
- **🟠 ORANGE**: Network requests, system commands  
- **🔴 RED**: Sudo, rm -rf/, system-level changes

### Granular Controls
```json
{
  "safety_rules": {
    "file_write": {
      "allowed_extensions": [".py", ".js", ".txt"],
      "max_file_size": "1MB",
      "required_patterns": ["test_", ".py"],
      "blacklisted_patterns": ["config.py", "secrets.json"]
    },
    "network": {
      "allowed_domains": ["pypi.org", "github.com"],
      "max_response_size": "10MB"
    },
    "execute": {
      "allowed_commands": ["python", "node", "npm", "pip"],
      "banned_commands": ["sudo", "rm -rf", "systemctl"],
      "max_duration": 30
    }
  }
}
```

### Advanced Permission APIs

#### Batch Permissions
```http
POST /permissions/bulk
{
  "session_id": "uuid",
  "requests": [
    { "type": "file_write", "resource": "models.py", "details": "add validation"},
    { "type": "file_write", "resource": "tests/test_models.py"}
  ]
}
```

#### Context-aware permissions
```
Project Structure Analysis:
├── src/            # Allow write
├── tests/          # Allow write  
├── .env            # Block - secrets
├── /etc/hosts      # RED - system file
```

### Permission Breakdown UI
Shows to user:
- **Context**: "Your project's models.py file"
- **Risk Level**: "Low - adding validation logic"
- **Rollback**: "Git revert available with ID: abc123"
- **Duration**: "Temporary permission, expires in 5min"

### Safety Workflows

#### 1. Request → Approve → Execute → Confirm
```
Permission Request:
├── Automatic Analysis (project context)
├── Risk Assessment (green/yellow/orange/red)
├── User Decision (approve/deny/customize)
├── Execute with Credits
├── Confirmation & Logging
└── Rollback available
```

#### 2. Escalation System
```
User Request → Auto-approve (green) → Manual approval (yellow) → Admin required (red)
```

### Priority: CRITICAL - Security foundation for tool usage