# Sysadmin Mode - YOLO+

This fork adds a `--sysadmin` mode to Crush that removes **all** bash command restrictions, enabling AI to execute system administration commands.

## ⚠️ WARNING

**USE WITH EXTREME CAUTION**

This mode allows the AI to execute:
- `doas`, `sudo`, `su` (privilege escalation)
- `pkg_add`, `apt`, `yum`, etc. (package installation)
- `rcctl`, `systemctl`, `service` (service management)
- `pfctl`, `iptables`, `firewall-cmd` (firewall configuration)
- `crontab`, `mount`, `fdisk`, etc. (system modification)
- Network tools like `curl`, `wget`, `ssh`, `scp`

Only use this mode when you:
1. Fully trust the AI model
2. Are automating system administration tasks
3. Can monitor and verify all commands
4. Accept full responsibility for any system changes

## Usage

### Enable Sysadmin Mode

```bash
# Interactive mode with sysadmin powers
crush --sysadmin

# Non-interactive with sysadmin mode
crush --sysadmin run "install nginx and configure it"

# Combine with yolo mode (no prompts + no restrictions)
crush --sysadmin -y
```

### Build from Source

```bash
# Clone the fork
git clone https://github.com/shauber/crush.git
cd crush
git checkout feature/sysadmin-mode

# Build
go build -o crush .

# Install (optional)
sudo mv crush /usr/local/bin/
```

### OpenBSD-Specific Usage

Perfect for automating OpenBSD router/server setup:

```bash
# Example: Setup ad-blocker with full system access
crush --sysadmin run "Install and configure unbound with blocklists, setup PF rules"

# Example: Package management
crush --sysadmin run "Update all packages and apply syspatch"

# Example: Service configuration
crush --sysadmin run "Enable and start httpd with TLS"
```

## Implementation Details

### Changes Made

1. **Config** (`internal/config/config.go`)
   - Added `SysadminMode bool` to `Permissions` struct

2. **Bash Tool** (`internal/agent/tools/bash.go`)
   - Modified `blockFuncs(sysadminMode bool)` to return empty blockers when enabled
   - Updated `NewBashTool()` signature to accept sysadmin mode parameter
   - Pass sysadmin mode to all shell execution calls

3. **CLI** (`internal/cmd/root.go`)
   - Added `--sysadmin` / `-s` flag
   - Read flag in `setupApp()` and set `cfg.Permissions.SysadminMode`

4. **Tool Registration** (`internal/agent/coordinator.go`, `internal/agent/common_test.go`)
   - Updated all `NewBashTool()` calls to pass sysadmin mode

### Differences from --yolo

| Feature | --yolo | --sysadmin |
|---------|--------|------------|
| Auto-accept permissions | ✅ | ❌ (requires separate -y) |
| Remove command restrictions | ❌ | ✅ |
| Recommended use | Development | System automation |

Combine both for full automation: `crush --sysadmin -y`

## Future PR Plans

Before submitting upstream PR:
1. Test thoroughly on OpenBSD, Linux, macOS
2. Add warning prompt on first use
3. Consider requiring explicit config file option
4. Add logging of all privileged commands
5. Document security implications
6. Add unit tests for sysadmin mode
7. Consider making certain commands (rm -rf /) still blocked

## Testing

```bash
# Test that restrictions are removed
crush --sysadmin run "test doas command availability"

# Verify normal mode still blocks
crush run "try to use doas"  # Should fail
```

## Contributing

This is a fork for testing. Once stable, will submit PR to upstream:
https://github.com/charmbracelet/crush

## License

Same as upstream Crush (MIT License)
