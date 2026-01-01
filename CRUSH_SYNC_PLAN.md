# crush-sync: Implementation Plan & Progress

## Project Overview
Standalone sync tool for Crush conversations - enables cross-instance synchronization and backup.

## Architecture Plan

### Directory Structure
```
crush-sync/
cmd/
├── root.go         // Main CLI setup
├── export.go       // Export conversations
├── import.go       // Import conversations  
sync.go           // Network sync commands
├── backup.go       // Database backup
internal/
├── db/
│   ├── reader.go   // SQLite conversation reading
│   └── writer.go   // SQLite conversation writing
├── export/
│   ├── format.go   // JSON schema definitions
│   └── export.go   // Core export logic
├── import/
│   └── import.go   // Core import logic
├── backup/
│   └── backup.go   // Database backup utilities
└── transport/
    ├── client.go   // Sync client
    └── server.go   // Sync server (optional)
pkg/
├── types/
│   └── conversation.go  // Common type definitions
└── crypto/
    └── encrypt.go       // Optional encryption support
```

### CLI Interface
```bash
crush-sync export [file]          # Export all conversations to JSON
crush-sync import [file]          # Import conversations from JSON
crush-sync sync --server URL      # Bidirectional network sync
crush-sync backup [dir]          # Full database backup
crush-sync stats                 # Show conversation counts
crush-sync list                  # List conversations in sync files
crush-sync verify [file]         # Validate sync file integrity
```

### JSON Schema
```json
{
  "version": "1.0",
  "source": "crush-sync-v0.1.0", 
  "timestamp": 1703123456,
  "conversations": [
    {
      "session": {/* Session object */},
      "messages": [/* Message objects */],
      "files": [/* File objects */]
    }
  ]
}
```

## Implementation Phases

### Phase 1: MVP (Export/Import) ✅
- [ ] Set up Go module structure
- [ ] Create basic CLI framework  
- [ ] Implement conversation export
- [ ] Implement conversation import
- [ ] Basic session/message/file serialization

### Phase 2: Network Transport
- [ ] HTTP client for sync operations
- [ ] Sync server implementation
- [ ] Differential sync (only missing conversations)

### Phase 3: Polish & Safety
- [ ] Database backup before import
- [ ] Progress bars for operations
- [ ] Gzip compression for transport
- [ ] File validation and integrity checks
- [ ] Installation scripts and documentation

## Technical Decisions

### Core Dependencies
- **go-sqlite3**: Direct database reading
- **cobra**: CLI framework
- **go-json**: Fast JSON serialization
- **progressbar**: UI feedback

### Data Handling
- Direct SQLite queries (reuse existing schema)
- UUID collision handling via rename
- Timestamp-based ordering for conflicts
- Gzip compression for >50KB conversations

### Error Handling
- Database backup on every import
- Rollback capability via backup copy
- Validation of conversation integrity
- Graceful degradation for partial failures

## Current Progress
*Last updated: 2025-12-20*

- [x] ✅ Basic module structure created (go.mod, main.go)
- [x] ✅ CLI framework with Cobra setup complete
- [x] ✅ Export command implemented and working
- [x] ✅ Database reader with SQLite integration
- [x] ✅ JSON schema and type definitions
- [ ] 🔄 Import functionality in progress
- [ ] ⏳ Backup strategy for safe imports
- [ ] ⏳ Final import implementation
- [ ] ⏳ Testing and validation

## Usage Examples

```bash
# Install
go install github.com/[you]/crush-sync@latest

# Export conversations to backup file
crush-sync export backup-2024-01-15.json

# Import from another machine
crush-sync import backup-2024-01-15.json

# Sync with central server
crush-sync sync https://sync.crush.ai

# Create daily backup
crush-sync backup /home/user/crush-backups/
```