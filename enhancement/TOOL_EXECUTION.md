# Real-Time Tool Execution

## Scene: "Live coding with AI" vs file-only API

### CLI Reality - The Magic Moment
**Developer says**: "Run pytest on this file"
- Sees real-time test output scrolling
- Watch AI create new test cases live
- Edits files and immediately sees test results
- Can stop long-running operations

**Web Reality - The Dead End**
- File updated in web, but no way to run tests
- No feedback if changes actually work
- Edit, save, manually test, repeat... friction

### Problem Statement
The current API treats tools as "black boxes" - you get results, not experiences. CLI gives you **live streaming, progress, interaction** - web needs the same.

## Real-Time Tool Events - User Experience

### Story A: Code Generation Flow
```
Alice: "Create a web server"
Assistant: [starts typing]
Alice sees: "Creating server.py... [Writing imports 1/5]
[Adding routes 2/5] [Setting up server 3/5]"
[File created successfully, 25 lines]
[Would you like to test it? y/n]
```

### Story B: Database Migration  
```
Bob: "update database schema"
Assistant: [creates migrations]
Bob sees: "Running migration... [Step 1/4]
[Applying migration to table 'users'] [Success]
[Rolling back... something went wrong]

Would you like to:
• View the error log
• Edit the migration file
• Try again with different parameters
```

## Live Execution API Design

### SSE Streaming Events
```http
GET /sessions/{id}/live-stream

// Tool starting
event: tool_start
data: {
  "tool": "bash",
  "command": "python test.py",
  "estimated_duration": 5.2,
  "session_id": "uuid"
}

// Progressive progress
event: tool_progress  
data: {
  "stdout": "Tests collected: 15\n",
  "stderr": "",
  "percent_complete": 0.2,
  "current_action": "Test 3 of 15: test_user_creation"
}

// Tool completion
event: tool_complete
data: {
  "success": true,
  "duration": 4.1,
  "exit_code": 0,
  "summary": "15 tests passed, 0 failed",
  "files_created": ["test_results.xml"]
}
```

### Interactive Tool APIs

#### Start Script Execution
```http
POST /sessions/{id}/execute
{
  "tool": "bash",
  "command": "pytest test_file.py -v",
  "async": true,
  "timeout": 30
}

{
  "execution_id": "exec-123",
  "status": "starting",
  "poll_url": "/executions/exec-123",
  "stream_url": "/sessions/uuid/live-stream"
}
```

#### Interactive Filesystem
```http
// File creation with live preview
POST /sessions/{id}/files/preview-create
{
  "filename": "server.py",
  "content": [
    "from flask import Flask",
    "app = Flask(__name__)",
    "@app.route('/')\n    return 'Hello World'"
  ],
  "commit_message": "Add basic Flask server"
}

// Shows diff before committing
{
  "proposed_changes": [...],
  "conflicts": [],
  "affected_tests": ["test_server.py"]
}
```

### Progressive Interactions

#### Multi-step Tools
User wants to: "Install package and run tests"
```
Step 1: pip install flask -> [Running... 2/15kB downloaded]
Step 2: python -m pytest -> [Collecting tests...]
Step 3: All tests pass ✅
```

#### Stoppable Operations
User sees: "Training ML model... [Epoch 15/100] [ETA: 2min]"
```http
POST /executions/exec-123/stop
Response: {"stopped": true, "saved_checkpoints": ["model_14.pth"]}
```

#### Retry with Parameters
Tool fails, UI shows:
```http
POST /sessions/{id}/retry-last
{
  "tool": "pip-install", 
  "new_params": {"package": "flask==2.3.3"},
  "use_cached_errors": false
}
```

## Technical Flow

### 1. Tool Execution Pipeline
```
User Request → Session Management → Tool Dispatcher → Live SSE Stream → Result Notification
```

### 2. Safety & Progress Integration
```
Command Start → Safety Check → Permission Check → Execution Monitor → Live Updates
```

### Key Events Streamed
- **tool_start**: Tool initialization (with args)
- **progress_update**: Chunked output every 200ms
- **permission_needed**: Interactive prompts
- **file_changes**: Previews of file modifications
- **network_activity**: Downloads, API calls
- **completion**: Success/failure with artifacts

### Connection Management
- **Heartbeat**: Every 5s to maintain connection
- **Reconnection**: Resume from last known state
- **State buffering**: Store execution state for reconnection
- **Cleanup**: Auto-cleanup on disconnection

### Error Handling Showcase
User tries: "pip install invalid-package"
Shows: 
- Real-time error streaming
- Safe CommandArray validation
- Retry suggestion: "Did you mean 'requests' instead?"
- Rollback option for partial installs

### Priority: CRITICAL - Makes or breaks "live coding" experience