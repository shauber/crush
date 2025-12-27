# GPT-Image-1.5 Image Generation Implementation Plan

## Current State Analysis
- **No image generation support** in Crush codebase
- Only **image viewing/attachment** capabilities exist
- Support for **DALLE-3** but **no GPT-Image-1.5** integration

## Implementation Overview

### 1. Core Image Generation Tool
**Location**: `internal/agent/tools/image.go`

```go
type ImageGenerationParams struct {
    Prompt    string `json:"prompt"`
    Model     string `json:"model"` // "gpt-image-1.5"
    Size      string `json:"size"`
    Quality   string `json:"quality"
    Style     string `json:"style"
}

type ImageGenerationResult struct {
    URL         string `json:"url"`
    RevisedPrompt string `json:"revised_prompt"
    SavedPath   string `json:"saved_path,omitempty"
}
```

### 2. API Endpoints
- **POST** `/sessions/{id}/images` - Generate image
- **GET** `/sessions/{id}/images` - List generated images
- **GET** `/sessions/{id}/images/{imageId}` - Download/view specific image

### 3. Provider Configuration
Update `internal/config/provider.go` to support GPT-Image-1.5 models:

```go
"openai": {
    Models: []string{"gpt-image-1.5", "dall-e-3"},
    Capabilities: []string{"text", "images", "image_generation"}
}
```

### 4. Storage Schema
**SQLite Extension**: Add `images` table
- Session association with foreign keys
- Local file storage path references
- Metadata (prompt, model, timestamp)

### 5. Technical Considerations

#### API Compatibility
- OpenAI `/v1/images/generations` endpoint
- GPT-Image-1.5 response format
- Rate limiting handling
- API quota management

#### Security & Validation
- Prompt sanitization
- File size limits
- File type validation (PNG/JPEG)
- Input safety filters

#### Asynchronous Processing
- SSE streaming for status updates
- Background processing with job queues
- Caching mechanism for duplicate requests

#### Development Complexity: Medium-High
- **Est. LOCs**: 800-1200 across 3-4 files
- **Testing**: Mock server for API testing
- **Dependencies**: Image processing libraries
- **Error handling**: Network failures, quotas, validation

### 6. Storage & Deployment
- **Local storage** scratch directory configuration
- **API key management** for GPT-Image-1.5
- **Session cleanup** automation for temp images
- **Cross-platform** file path handling

### 7. Integration Points
- **TUI**: Inline image preview, progress indicators
- **SSE**: Real-time generation status
- **Session**: Link images to chat sessions
- **Tools**: Agent tool registration

## Next Steps
1. Create `ImageGenerationTool` in `internal/agent/tools/image.go`
2. Update provider configuration for GPT-Image-1.5
3. Add database schema migration for images table
4. Implement API endpoints
5. Add TUI integration for preview/upload
6. **Testing**: Mock server implementation
7. **Documentation**: User guides and API docs