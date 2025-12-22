# OAuth & Provider Connectivity

## Scene: "Connecting AI accounts" nightmare vs CLI simplicity

### CLI Reality - The Flow
**Developer opens CLI**: Sees "GitHub Copilot: Authenticate? (y/N)"
- Guided device flow opens browser
- Code displayed for copying
- "Success! GitHub Copilot connected"
- Now can use GitHub models seamlessly

**Web Reality - What's Missing**
- No OAuth flows for providers
- Can't add GitHub Copilot models
- No Anthropic account connection
- Missing API key management
- No provider health status

### User Provider Journeys

#### Story 1: GitHub Copilot Setup
```
Developer: "I want to use GitHub Copilot models"
Web Flow:
1. Click "Add Provider" → "GitHub Copilot"  
2. Redirect to GitHub device flow
3. Enter code: "8-letters-4-numbers"
4. "GitHub Copilot Connected"
5. Models now available: gpt-4-copilot, etc.
```

#### Story 2: Multi-provider Management
```
Developer sees UI:
┌─ Provider Status ─────────────────────────┐
│ ✅ OpenAI               34,512 tokens used │
│ ✅ Anthropic            $12.50 spent      │
│ 🔴 GitHub Copilot       Rate limited      │
│ ❌ Hyper                Not connected     │
└───────────────────────────────────────────┘
```

## OAuth Flow Architecture

### GitHub Copilot OAuth
```http
POST /oauth/github-copilot/start
{
  "redirect_uri": "http://localhost:3000/oauth/callback"
}

Response:
{
  "device_code": "ABCD-EFGH",
  "user_code": "8L4K-9R2T",
  "verification_uri": "https://github.com/login/device",
  "expires_in": 900,
  "interval": 5
}
```

### Status Polling for UX
```http
GET /oauth/github-copilot/status/8L4K-9R2T
{
  "status": "pending|verified|expired",
  "account": "username",  // Once verified
  "models": ["gpt-4-copilot", "gpt-3.5-copilot"],
  "credits": null          // GitHub-style credits
}
```

## Provider Management APIs

### List Connected Providers
```http
GET /providers
{
  "connected": [
    {
      "id": "anthropic",
      "name": "Anthropic",
      "connected": true,
      "models": ["claude-3.5-sonnet", "claude-3-opus"],
      "current_model": "claude-3.5-sonnet",
      "credits": "$12.50",
      "last_used": "2024-01-15T10:30:00Z"
    }
  ],
  "available": [
    {
      "id": "github-copilot",
      "name": "GitHub Copilot", 
      "requires_oauth": true,
      "auth_url": "/oauth/github-copilot/start"
    }
  ]
}
```

### Add New Provider
```http
POST /providers/custom
{
  "type": "hyper",
  "name": "My Hyper Server",
  "endpoint": "https://hyper.company.com/api",
  "api_key": "hsk-key-...",  // Will be encrypted
  "models": [
    {"id": "claude-3.5", "name": "Claude 3.5"}
  ]
}
```

### Test Provider
```http
POST /providers/{id}/test
{
  "test_message": "Hello, testing connection...",
  "expected_response": "string"
}

Response:
{
  "success": true,
  "latency": 245,
  "available_models": 3,
  "estimated_tokens": 1000
}
```

## OAuth Provider Support

### Supported OAuth Flows
- **GitHub Copilot**: Device flow → Browser auth → Instant access
- **Anthropic**: API key input → Direct connection
- **OpenAI**: API key + organization ID
- **Hyper**: Device flow or API key support

### Security & Storage
```javascript
// Encrypted storage
{
  "providers": {
    "github-copilot": {
      "token": "encrypted_jwt_token",
      "refresh": "encrypted_refresh",
      "expires": "2024-03-15T12:00:00Z"
    }
  }
}
```

## Rate Limiting & Credits

### Real-time Usage Tracking
```http
GET /providers/{id}/usage
{
  "daily_tokens": 43829,
  "daily_cost": "$2.15",
  "monthly_budget": "$20.00", 
  "remaining": "$17.85",
  "rate_limit": {
    "limit": 3500,
    "remaining": 2147,
    "reset_time": "2024-01-15T23:59:59Z"
  }
}
```

### Rate Limit Events in SSE
```json
event: provider_rate_limit
data: {
  "provider": "openai",
  "limit": "daily_tokens",
  "current": 48000,
  "max": 50000,
  "advice": "Switch to anthropic/claude-3.5-sonnet"
}
```

## Health Check System

### Provider Health Dashboard
```http
GET /providers/health
{
  "statuses": {
    "openai": {
      "healthy": true,
      "latency": 123,
      "last_check": "2024-01-15T15:30:00Z"
    },
    "anthropic": {
      "healthy": "degraded",
      "reason": "high latency",
      "latency": 890,
      "advice": "Expect delays >3s"
    }
  }
}
```

### Auto-fallback Logic
```javascript
// If OpenAI down, automatically try Anthropic
{
  "primary": "openai",
  "fallback": "anthropic",
  "criteria": {
    "max_latency": 1000,
    "min_availability": 95
  }
}
```

## Connection Troubleshooting

### Diagnostic Events
```http
event: connection_issue
data: {
  "provider": "github-copilot",
  "issue": "token_expired",
  "resolution": "/oauth/github-copilot/refresh"
}

event: api_key_error  
data: {
  "provider": "anthropic",
  "error": "invalid_key_format",
  "help_url": "/docs/providers/anthropic-setup"
}
```

### Priority: HIGH - Essential for provider flexibility