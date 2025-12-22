# Workspace & File System Integration

## Scene: File operations from CLI vs web experience

### CLI Reality
**Developer says**: "List all Python files in this project"
- Sees immediate project file tree
- Can browse and edit files effortlessly
- Context is automatically set to project root

**Web developer tries**: Uploads files manually
- Has to manually specify project root
- No file browser, just text
- Can't see project structure to understand context

### User Stories

#### Epic: Web users can work with their actual project files
```
As a web developer
I want to see my project file tree
So I can understand what we're working with

As a web developer
I want to see file contents from the web
So I can reference code in conversations

As a web developer  
I want to edit files through the web
So I can make changes based on AI suggestions
```

#### Scenario 1: Project discovery
**Given**: Developer opens workspace folder
**When**: Web interface loads
**Then**: Shows project structure with:
- Git-ignored files grayed out
- Language indicators (.py, .js, etc.)
- File size and modified dates
- Quick preview capability

#### Scenario 2: Interactive file operations
**Given**: AI suggests editing a specific file
**When**: User clicks "edit file" button
**Then**: Opens file editor with syntax highlighting
**And**: Suggests changes side-by-side with original
**And**: Shows git diff before saving

#### Scenario 3: Safe bulk operations
**Given**: AI wants to refactor multiple files
**When**: User approves changes
**Then**: Shows preview of all affected files
**And**: Allows selective file saving
**And**: Provides rollback capability

### API Design - File Operations

#### File Tree
```http
GET /workspace/tree?depth=3
{
  "root": "/home/user/project",
  "files": [
    { "path": "app.py", "type": "file", "size": 1200, "lang": "python" },
    {
      "path": "src", 
      "type": "dir", 
      "children": [
        { "path": "src/main.py", "type": "file" }
      ]
    }
  ]
}
```

#### File Content with Safety
```http
GET /files/src/main.py
{
  "content": "#!/usr/bin/env python...",
  "metadata": {
    "size": 1200,
    "language": "python",
    "read_only": false,
    "git_status": "modified"
  },
  "safety": {
    "allowed": true,
    "warnings": ["File is executable"]
  }
}
```

#### Workspace Context
```http
POST /workspace/analyze
{
  "project_type": "python",
  "dependencies": ["requirements.txt", "Poetry"], 
  "git": true,
  "languages": ["python", "javascript"],
  "recommendations": ["requirements.txt", "tests/"]
}
```

### File Safety Layer
- **Path restriction**: Enforce workspace root
- **Git integration**: Show changes, allow diffs
- **Permission validation**: Same rules as CLI
- **Content validation**: File type scanning
- **Rate limiting**: Prevent abuse

### Priority: MEDIUM - Enhances productivity dramatically