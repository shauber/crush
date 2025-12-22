# Crush Web UI - Architectural Plan

## Current State Assessment
The existing API in `internal/api/` provides comprehensive functionality for a web UI. Based on analysis of the codebase, here are the key features available:

### ✅ Existing API Capabilities
- **Session Management**: Create, list, get, delete sessions
- **Real-time Chat**: SSE streaming with live updates
- **Message Handling**: Send messages, list messages, async processing
- **Model Control**: Switch between providers/models (OpenAI, Anthropic, etc.)
- **Health Monitoring**: Provider connectivity status
- **CORS**: Ready for browser access

### ✅ CLI Features Available for Web UI
- **Code Tools**: File browsing (ls, glob), content viewing (view), editing (edit, multiedit)
- **Search**: grep, sourcegraph, web search (web_search, web_fetch, agentic_fetch)
- **System Commands**: bash execution with safety controls
- **MCP Integration**: Available through agent coordinator
- **Session History**: Persistent across sessions
- **Model Switching**: Large/small model patterns

## Web UI Feature Categories

### 1. Core Chat Experience
- **Real-time Messaging** 📱
  - Live message streaming via SSE
  - Typing indicators between turn-based responses
  - Markdown/ANSI rendering of responses
  - Message history with scrollable timeline
  - Message threading/references (existing session model supports)

- **Session Management** 🗂️
  - Session list with titles/dates
  - Create new sessions from sidebar
  - Delete/archive sessions
  - Fork existing sessions (builds on parent session feature)
  - Session search/filter

- **Message Composition** ✍️
  - Multi-line text editor with syntax highlighting
  - Attach images/prompt files via existing attachment system
  - Auto-focus and keyboard shortcuts
  - Save draft messages

### 2. AI Toolkit Integration
- **Code Intelligence** 💻
  - File tree browser (ls, glob → UI sidebar)
  - File content viewer with syntax highlighting
  - Multi-file editor with change tracking
  - Code search across project (grep, sourcegraph)
  - Git integration awareness (git commands through bash)

- **System Tools** 🔧
  - Terminal access when needed (safe bash execution)
  - Process monitoring (job kill/output tools)
  - File operations (download, view/write)
  - Web content fetching with safety controls

- **Multi-Provider Support** 🏪
  - Model picker with provider switching
  - Cost tracking per session
  - Provider health monitoring
  - Custom model configurations

### 3. UI Architecture Decisions

#### Frontend Framework Choices
- **Minimal React setup** (avoid complex frameworks)
- **Vanilla TypeScript** with Web Components for simplicity
- **Tailwind CSS** for consistency with Charmbracelet design
- **Server-Sent Events** for real-time (already implemented, no WebSocket needed)

#### Component Structure
```
src/
├── components/
│   ├── Chat/
│   │   ├── ChatWindow.tsx (messages, streaming)
│   │   ├── MessageInput.tsx (text editor, attachments)
│   │   └── Message.tsx (rendering, markdown)
│   ├── Session/
│   │   ├── SessionList.tsx (create, list, delete)
│   │   └── SessionView.tsx (session metadata)
│   ├── Tools/
│   │   ├── FileTree.tsx (ls, glob integration)
│   │   ├── Editor.tsx (view/edit files)
│   │   └── Search.tsx (grep, sourcegraph)
│   └── Shared/
       ├── ModelPicker.tsx (provider switching)
       └── HealthMonitor.tsx (status indicators)
```

#### State Management
- **Simple client-side state** (avoid Redux)
- **Session state** managed via API calls
- **Real-time updates** via SSE events
- **File state** cached locally for performance

### 4. Web-Specific Features Needed

#### API Extensions (if needed)
- **File Operations**: Add dedicated file management endpoints
- **Workspace Detection**: Automatically detect project root
- **Preview Generation**: Add endpoint for rendered markdown preview
- **Configuration**: Add endpoint for client-side settings

#### Security Considerations
- **Tool Restrictions**: Keep existing CLI safety (banned commands, etc.)
- **File Access**: Same sandboxing as CLI via chroot/config
- **Rate Limiting**: Add basic rate limiting per session
- **Input Validation**: Extend existing validation to web context

### 5. Development Phases

#### Phase 1: Basic Chat (2-3 hours)
- [ ] Simple React setup with TypeScript
- [ ] Session list and creation
- [ ] Basic message sending/receiving via SSE
- [ ] Message list with markdown rendering

#### Phase 2: Tool Integration (3-4 hours)
- [ ] File tree browser component
- [ ] File viewer/editor integration
- [ ] Search functionality via grep/sourcegraph
- [ ] Terminal access when needed

#### Phase 3: Advanced Features (4-5 hours)
- [ ] Model picker with provider switching
- [ ] Session management (delete, fork)
- [ ] Cost tracking display
- [ ] Advanced tools menu

#### Phase 4: Polish & Deployment (2-3 hours)
- [ ] Responsive design
- [ ] Dark/light themes
- [ ] Loading states and error handling
- [ ] Build optimization and deployment

### 6. Open Questions & Considerations

#### Tool Access Ambiguities
- **Security validation**: How much of CLI tool access should be exposed?
- **File system limits**: Same restrictions as CLI or web-specific limits?
- **Interactive prompts**: How to handle CLI prompts in web context?

#### UI/UX Decisions
- **Split vs unified view**: File browser in sidebar vs new tab?
- **Editor integration**: Inline vs dedicated editor view?
- **Command suggestions**: How to present available tools to users?

#### Deployment Options
- **Single binary**: Bundle web assets with Go binary
- **Static serving**: Serve from OpenBSD chroot
- **Reverse proxy**: Use nginx for static files and API routing

### 7. Technical Implementation Notes

#### Session Synchronization
- WebSocket alternative: Already have SSE via `/sessions/:id/stream`
- Message ordering: Use timestamps from existing message model
- Conflict resolution: Last-writer-wins for concurrent sessions

#### File System Integration
- **Root path**: Use current working directory as project root
- **Relative paths**: All operations relative to project root
- **Navigation**: Tree view with breadcrumb support
- **Editing**: Line-based editing to match CLI experience

#### Performance Optimizations
- **Lazy loading**: Messages load on demand
- **Caching**: Client-side cache for file contents
- **Debouncing**: Rate limit FS operations
- **Streaming**: Progressive rendering for large responses

## Next Steps
1. Create basic HTML/TypeScript structure
2. Implement chat functionality first
3. Add file browser and tools
4. Polish UI and add advanced features
5. Test with OpenBSD deployment

This plan focuses on exposing CLI functionality through a web interface rather than adding new capabilities, maintaining the same safety and security model while providing a modern browser-based experience.