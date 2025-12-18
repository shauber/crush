# Crush Sysadmin Mode - Local Build Notes

## Setup Complete ✅

**Fork**: https://github.com/shauber/crush
**Branch**: `feature/sysadmin-mode`
**Commits**: 2 commits pushed

## What Was Done

1. **Forked** `charmbracelet/crush` → `shauber/crush`
2. **Cloned** fork to `~/crush-sysadmin`
3. **Created** feature branch `feature/sysadmin-mode`
4. **Implemented** sysadmin mode:
   - Added `--sysadmin` CLI flag
   - Modified bash tool to skip all command restrictions when enabled
   - Updated config, tool instantiation, tests
5. **Documented** in `SYSADMIN_MODE.md`
6. **Pushed** to your fork

## Local Build

```bash
# Build custom binary
cd ~/crush-sysadmin
go build -o ~/go/bin/crush-sysadmin .

# Test it
crush-sysadmin --sysadmin --help

# Use it (with caution!)
crush-sysadmin --sysadmin
```

## Usage on OpenBSD Router

```bash
# Interactive sysadmin mode
crush-sysadmin --sysadmin

# Non-interactive with full power
crush-sysadmin --sysadmin run "install unbound and configure as DNS server"

# Combined with yolo (no prompts)
crush-sysadmin --sysadmin -y run "setup firewall rules for web server"
```

## Testing Scenarios

### 1. Verify restrictions are removed

```bash
# This should work in sysadmin mode
crush-sysadmin --sysadmin run "check if doas command is available"

# This should fail in normal mode
crush-sysadmin run "try to use doas"
```

### 2. OpenBSD package management

```bash
crush-sysadmin --sysadmin run "update package database with pkg_add -u"
```

### 3. Service configuration

```bash
crush-sysadmin --sysadmin run "enable and start httpd service with rcctl"
```

### 4. Firewall rules

```bash
crush-sysadmin --sysadmin run "show current PF rules with pfctl"
```

## Before Submitting PR

- [ ] Test on OpenBSD (your router)
- [ ] Test on Linux
- [ ] Test on macOS
- [ ] Add unit tests for sysadmin mode
- [ ] Consider warning prompt on first use
- [ ] Add audit logging for privileged commands
- [ ] Review security implications
- [ ] Update upstream docs

## Remotes Configured

```
origin   → git@github.com:shauber/crush.git (your fork)
upstream → git@github.com:charmbracelet/crush.git (upstream)
```

## Sync with Upstream (later)

```bash
cd ~/crush-sysadmin
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
```

## Files Modified

- `internal/config/config.go` - Added SysadminMode field
- `internal/agent/tools/bash.go` - Modified blockFuncs() logic
- `internal/cmd/root.go` - Added --sysadmin flag
- `internal/agent/coordinator.go` - Pass sysadmin mode to tool
- `internal/agent/common_test.go` - Update test instantiation
- `SYSADMIN_MODE.md` - Documentation

## GitHub Repository

https://github.com/shauber/crush/tree/feature/sysadmin-mode

Ready to test! Use `crush-sysadmin --sysadmin` to enable unrestricted mode.
