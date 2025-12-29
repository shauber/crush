## Updated Plan: Auto-Enabled Server Mode Streaming

### Core Adjustment: Server Mode Autodetection
**Replace env flag with runtime detection:**
```go
// streaming/config.go
func AutoEnableForServerMode(app *app.App) {
    // Detect if running in server mode via cmd analysis or runtime context
    if isServerMode() {
        registerStreaming(app)
    }
}

func isServerMode() bool {
    // Detect via process name or via presence of running API components
    return os.Args[0] != "" && contains(os.Args[0], "server")
}
```

### Streamlined Integration Points:
1. **Hook into server startup**: `/internal/api/server.go:24` - right after `sseHandler := streaming.NewSSEHandler()`
2. **Leverage existing dispatcher**: Use the same `SSEHandler` instance created for the server
3. **Tool registration injection**: Wrap tools at agent level during server initialization

### Reduced Surface Area:
- **Single integration file**: `streaming/server_mode.go`
- **One-time wrap during server init**: No config flags needed
- **Passive wrapper**: Tools stream automatically when server starts, silent otherwise
