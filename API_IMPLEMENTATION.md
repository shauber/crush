# Crush API Implementation Plan

## Goal
Convert Crush CLI to expose its core functionality via HTTP API while maintaining CLI compatibility. No user management, no authentication - just a simple API server that replicates CLI behavior.

## Architecture Principles
1. **Keep CLI intact** - All existing functionality continues to work
2. **Minimal changes** - Reuse existing services (session, message, agent coordinator)
3. **Additive approach** - New code in separate packages
4. **OpenBSD compatible** - No complex dependencies, chroot-friendly

---

## Phase 1: Basic API Server

### 1. Add Server Command

Create `internal/cmd/server.go`:
- Add `server` subcommand to root command
- Reuse existing `setupApp()` logic
- Start HTTP server with existing app instance

**Key Points:**
- Use existing `app.App` struct - no modifications needed
- Reuse `setupApp()` from root.go
- Simple HTTP server, no fancy frameworks initially

### 2. Create API Package Structure

```
internal/api/
├── server.go      # HTTP server setup
├── handlers.go    # Request handlers
└── streaming.go   # SSE streaming
```

**Dependencies:**
- Standard library `net/http` only
- Maybe `chi` router (already used? check go.mod)
- No authentication libraries
- No database migrations (use existing schema)

### 3. Core Endpoints

```
POST   /sessions              # Create new session
GET    /sessions              # List all sessions
GET    /sessions/:id          # Get session details
DELETE /sessions/:id          # Delete session

POST   /sessions/:id/messages # Send message (returns immediately)
GET    /sessions/:id/messages # List messages
GET    /sessions/:id/stream   # SSE stream for real-time updates

POST   /sessions/:id/cancel   # Cancel running request
GET    /sessions/:id/status   # Check if session is busy
```

### 4. Implementation Details

**server.go:**
- HTTP server listening on configured port
- Graceful shutdown handling
- Simple router (chi or stdlib mux)
- CORS headers (for local development)

**handlers.go:**
- Map HTTP requests to existing service methods
- JSON request/response
- Error handling and status codes
- Direct calls to `app.Sessions`, `app.Messages`, `app.AgentCoordinator`

**streaming.go:**
- SSE (Server-Sent Events) for streaming responses
- Subscribe to existing pubsub events
- Stream message updates in real-time
- Handle client disconnection

### 5. Request/Response Types

```go
// CreateSessionRequest
{
  "title": "Debug login issue"
}

// SessionResponse
{
  "id": "uuid",
  "title": "Debug login issue",
  "message_count": 0,
  "created_at": 1234567890
}

// SendMessageRequest
{
  "content": "List all Go files in this project"
}

// MessageResponse
{
  "id": "uuid",
  "session_id": "uuid",
  "role": "assistant",
  "content": "Found 42 Go files...",
  "created_at": 1234567890
}

// SSE Event Format
data: {"type":"message_update","message_id":"uuid","content":"partial response..."}
data: {"type":"complete","message_id":"uuid"}
```

---

## Phase 2: Streaming Implementation

### How It Works

**Existing Architecture:**
- `app.Messages` already has pubsub via `Subscribe(ctx)`
- Messages are published as they're updated
- Agent coordinator streams token-by-token updates

**Our Implementation:**
1. Client calls `GET /sessions/:id/stream`
2. Server subscribes to message events via `app.Messages.Subscribe(ctx)`
3. Server streams events as SSE
4. Client receives real-time updates

**SSE Format:**
```
event: message_created
data: {"id":"uuid","role":"user","content":"..."}

event: message_updated
data: {"id":"uuid","content":"partial assistant response"}

event: message_updated  
data: {"id":"uuid","content":"partial assistant response continues"}

event: message_complete
data: {"id":"uuid"}
```

---

## Implementation Steps

### Step 1: Add Server Command (30 min)

File: `internal/cmd/server.go`

```go
package cmd

import (
    "fmt"
    "github.com/charmbracelet/crush/internal/api"
    "github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
    Use:   "server",
    Short: "Run Crush as HTTP API server",
    RunE: func(cmd *cobra.Command, args []string) error {
        app, err := setupApp(cmd)
        if err != nil {
            return err
        }
        defer app.Shutdown()

        port, _ := cmd.Flags().GetInt("port")
        return api.Serve(app, port)
    },
}

func init() {
    rootCmd.AddCommand(serverCmd)
    serverCmd.Flags().IntP("port", "p", 8080, "HTTP port")
}
```

### Step 2: Create API Server (1 hour)

File: `internal/api/server.go`

```go
package api

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    "github.com/charmbracelet/crush/internal/app"
)

func Serve(app *app.App, port int) error {
    mux := http.NewServeMux()
    
    // Create handlers with app instance
    h := &handlers{app: app}
    
    // Register routes
    mux.HandleFunc("POST /sessions", h.createSession)
    mux.HandleFunc("GET /sessions", h.listSessions)
    mux.HandleFunc("GET /sessions/{id}", h.getSession)
    mux.HandleFunc("DELETE /sessions/{id}", h.deleteSession)
    
    mux.HandleFunc("POST /sessions/{id}/messages", h.sendMessage)
    mux.HandleFunc("GET /sessions/{id}/messages", h.listMessages)
    mux.HandleFunc("GET /sessions/{id}/stream", h.streamSession)
    
    mux.HandleFunc("POST /sessions/{id}/cancel", h.cancelSession)
    mux.HandleFunc("GET /sessions/{id}/status", h.sessionStatus)
    
    // Wrap with CORS and logging
    handler := corsMiddleware(loggingMiddleware(mux))
    
    addr := fmt.Sprintf(":%d", port)
    slog.Info("Starting API server", "addr", addr)
    
    server := &http.Server{
        Addr:         addr,
        Handler:      handler,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }
    
    return server.ListenAndServe()
}

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
    })
}
```

### Step 3: Implement Handlers (2 hours)

File: `internal/api/handlers.go`

```go
package api

import (
    "encoding/json"
    "net/http"

    "github.com/charmbracelet/crush/internal/app"
    "github.com/charmbracelet/crush/internal/message"
)

type handlers struct {
    app *app.App
}

type createSessionRequest struct {
    Title string `json:"title"`
}

type sendMessageRequest struct {
    Content string `json:"content"`
}

func (h *handlers) createSession(w http.ResponseWriter, r *http.Request) {
    var req createSessionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    title := req.Title
    if title == "" {
        title = "New Session"
    }
    
    session, err := h.app.Sessions.Create(r.Context(), title)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    respondJSON(w, session)
}

func (h *handlers) listSessions(w http.ResponseWriter, r *http.Request) {
    sessions, err := h.app.Sessions.List(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    respondJSON(w, sessions)
}

func (h *handlers) getSession(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    
    session, err := h.app.Sessions.Get(r.Context(), id)
    if err != nil {
        http.Error(w, "Session not found", http.StatusNotFound)
        return
    }
    
    respondJSON(w, session)
}

func (h *handlers) deleteSession(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    
    if err := h.app.Sessions.Delete(r.Context(), id); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) sendMessage(w http.ResponseWriter, r *http.Request) {
    sessionID := r.PathValue("id")
    
    var req sendMessageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Create user message
    msg, err := h.app.Messages.Create(r.Context(), sessionID, message.CreateMessageParams{
        Role:  message.User,
        Parts: []message.ContentPart{message.TextPart{Text: req.Content}},
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Start agent processing (async)
    go func() {
        _, err := h.app.AgentCoordinator.Run(context.Background(), sessionID, req.Content)
        if err != nil {
            // Error will be visible in session stream
            slog.Error("agent run failed", "error", err, "session", sessionID)
        }
    }()
    
    respondJSON(w, msg)
}

func (h *handlers) listMessages(w http.ResponseWriter, r *http.Request) {
    sessionID := r.PathValue("id")
    
    messages, err := h.app.Messages.List(r.Context(), sessionID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    respondJSON(w, messages)
}

func (h *handlers) cancelSession(w http.ResponseWriter, r *http.Request) {
    sessionID := r.PathValue("id")
    
    h.app.AgentCoordinator.Cancel(sessionID)
    
    w.WriteHeader(http.StatusOK)
}

func (h *handlers) sessionStatus(w http.ResponseWriter, r *http.Request) {
    sessionID := r.PathValue("id")
    
    status := map[string]interface{}{
        "busy":           h.app.AgentCoordinator.IsSessionBusy(sessionID),
        "queued_prompts": h.app.AgentCoordinator.QueuedPrompts(sessionID),
    }
    
    respondJSON(w, status)
}

func respondJSON(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}
```

### Step 4: Implement SSE Streaming (1.5 hours)

File: `internal/api/streaming.go`

```go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/charmbracelet/crush/internal/pubsub"
)

func (h *handlers) streamSession(w http.ResponseWriter, r *http.Request) {
    sessionID := r.PathValue("id")
    
    // Verify session exists
    _, err := h.app.Sessions.Get(r.Context(), sessionID)
    if err != nil {
        http.Error(w, "Session not found", http.StatusNotFound)
        return
    }
    
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("X-Accel-Buffering", "no")
    
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }
    
    // Subscribe to message events
    messageEvents := h.app.Messages.Subscribe(r.Context())
    
    // Send initial connection event
    fmt.Fprintf(w, "event: connected\ndata: {\"session_id\":\"%s\"}\n\n", sessionID)
    flusher.Flush()
    
    // Stream events
    for {
        select {
        case <-r.Context().Done():
            return
            
        case event := <-messageEvents:
            msg := event.Payload
            
            // Only send events for this session
            if msg.SessionID != sessionID {
                continue
            }
            
            eventType := "message_updated"
            switch event.Type {
            case pubsub.CreatedEvent:
                eventType = "message_created"
            case pubsub.DeletedEvent:
                eventType = "message_deleted"
            }
            
            data, _ := json.Marshal(map[string]interface{}{
                "type":       eventType,
                "message_id": msg.ID,
                "role":       msg.Role,
                "content":    msg.Content().String(),
                "created_at": msg.CreatedAt,
            })
            
            fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
            flusher.Flush()
        }
    }
}
```

---

## Testing the API

### Start the Server

```bash
# In your crush directory
go run . server --port 8080
```

### Test with curl

```bash
# Create session
curl -X POST http://localhost:8080/sessions \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Session"}'

# Returns: {"id":"abc-123","title":"Test Session",...}

# Send message
curl -X POST http://localhost:8080/sessions/abc-123/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"List all Go files"}'

# Stream session (in another terminal)
curl -N http://localhost:8080/sessions/abc-123/stream

# List messages
curl http://localhost:8080/sessions/abc-123/messages

# Check status
curl http://localhost:8080/sessions/abc-123/status

# Cancel running request
curl -X POST http://localhost:8080/sessions/abc-123/cancel
```

---

## OpenBSD Considerations

### Building
```bash
# On OpenBSD
go build -o crush-server .
```

### Running in Chroot
```bash
# Create chroot structure
mkdir -p /var/crush/{bin,tmp,db}

# Copy binary
cp crush-server /var/crush/bin/

# Copy required libraries (check with ldd)
# Copy any required configs

# Run in chroot
chroot /var/crush /bin/crush-server server --port 8080
```

### Dependencies
- SQLite is statically linked via `ncruces/go-sqlite3`
- No external dependencies needed in chroot
- Config can be passed via environment variables

---

## What We're NOT Doing (Yet)

- ❌ User authentication
- ❌ User management
- ❌ Workspace isolation
- ❌ Per-user MCP configurations
- ❌ Custom prompt templates
- ❌ Multi-mode sessions (code/chat)
- ❌ Web UI
- ❌ WebSocket (using SSE instead)

---

## Files to Create/Modify

### New Files
1. `internal/cmd/server.go` - Server subcommand
2. `internal/api/server.go` - HTTP server setup
3. `internal/api/handlers.go` - Request handlers
4. `internal/api/streaming.go` - SSE implementation

### Modified Files
1. `internal/cmd/root.go` - Add serverCmd to init()
2. `go.mod` - Possibly add chi router (or use stdlib)

### No Changes Needed
- All existing internal packages work as-is
- Database schema unchanged
- Agent coordinator unchanged
- Session/message services unchanged

---

## Expected Work Time

- Step 1 (Server command): 30 minutes
- Step 2 (API server setup): 1 hour
- Step 3 (Handlers): 2 hours
- Step 4 (SSE streaming): 1.5 hours
- Testing & fixes: 1 hour

**Total: ~6 hours of focused work**

---

## Success Criteria

1. ✅ `crush server` starts HTTP server
2. ✅ Can create sessions via POST /sessions
3. ✅ Can send messages via POST /sessions/:id/messages
4. ✅ Messages stream via SSE at GET /sessions/:id/stream
5. ✅ All agent features work (tools, LSP, MCP)
6. ✅ `crush` CLI still works normally
7. ✅ Builds and runs on OpenBSD
8. ✅ Works in chroot environment

---

## Implementation Order

Start with this exact sequence:

1. Create `internal/cmd/server.go` with minimal serverCmd
2. Create `internal/api/server.go` with basic HTTP server and CORS
3. Implement session CRUD in `handlers.go` (create, list, get, delete)
4. Implement message handlers (send, list)
5. Implement SSE streaming in `streaming.go`
6. Add status and cancel endpoints
7. Test everything with curl
8. Verify OpenBSD build

That's it. Keep it simple, reuse existing code, add minimal new code.
