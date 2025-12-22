# Configuration Management System

## Scene: Running crush with --anthropic-friendly flag vs web experience

### CLI Reality
**Developer runs**: `echo "fix this bug" | crush --provider anthropic`
- Instantly knows their provider choice
- Can see "Using anthropic/claude-3.5-sonnet"
- If fails: "API key missing"

**Developer opens web**: Tries to send message
- Stuck wondering "which provider am I using?"
- Can't change provider from web
- No feedback on rate limits or issues

### User Stories

#### Epic: Web users can see and change their provider settings
```
As a web developer
I want to see my current provider configuration
So that I know which model will handle my request

As a web developer  
I want to change my provider without restarting the server
So that I can adapt based on rate limits or available credits

As a web developer
I want to test a new provider configuration
So that I can validate it works before using in production
```

#### Scenario 1: Viewing current setup
**When**: Developer opens web interface
**Then**: See provider list with:
- ✅ Connected providers (green dot)
- ❌ Missing API keys (red badge)
- 🟡 Available but not configured (grayed out)
- Current model in use with pricing info

#### Scenario 2: Adding new provider
**Given**: Developer has new Anthropic API key
**When**: They add provider via web form
**Then**: API validates key, shows model list, sets as active
**And**: Change persists across server restarts

#### Scenario 3: Runtime provider switch
**Given**: Hitting OpenAI rate limits
**When**: Switch provider in web settings
**Then**: Next message uses new provider
**And**: Session shows provider change in history

### API Design - Configuration

#### Get Configuration
```http
GET /config
{
  "providers": {
    "openai": {
      "enabled": true,
      "models": ["gpt-4", "gpt-3.5-turbo"],
      "current_model": "gpt-4",
      "has_api_key": true,
      "last_tested": "2024-01-15T10:30:00Z"
    },
    "anthropic": {
      "enabled": false,
      "models": ["claude-3.5-sonnet", "claude-3-opus"],
      "error": "API key not configured"
    }
  },
  "tools": {
    "bash": {
      "timeout": 30,
      "banned_commands": ["sudo", "rm -rf"],
      "allowed_directories": ["/work", "/tmp"]
    }
  }
}
```

#### Update Provider
```http
PUT /config/providers/anthropic
{
  "api_key": "sk-...",
  "model": "claude-3.5-sonnet"
}

Response:
{
  "success": true,
  "test_result": {
    "status": "ok",
    "models": ["claude-3.5-sonnet", "claude-3-opus"]
  }
}
```

### Development Items
- [ ] Add config CRUD endpoints
- [ ] Provider validation/test endpoints 
- [ ] Real-time provider status updates via SSE
- [ ] API key encryption/storage
- [ ] Config validation rules
- [ ] Hot-reload for config changes

### Priority: HIGH - Basic usability blocker