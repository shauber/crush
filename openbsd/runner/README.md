# OpenBSD Chroot Runner for Crush

This directory contains scripts for running Crush in an isolated chroot environment on OpenBSD, configured to work with Vultr (Kimi) or mock providers safely.

## Quick Start

```bash
cd crush/openbsd/runner
./setup-chroot.sh start        # Setup and start server immediately
./setup-chroot.sh              # Just setup
```

## Features

- **Dynamic port selection**: Uses 8080 if available, falls back to next free port
- **Multi-user support**: Works with any user's Vultr API key from their profile
- **Isolated testing**: Complete filesystem isolation in `/tmp/crush-test`
- **Shared libraries**: Uses bind mounts for efficiency (no duplication)
- **Backup system**: Original `crush.json` backed up as `crush.json.backup`
- **Error handling**: Comprehensive validation and cleanup

## Configuration

### Environment Variables

the script checks these environment variables for Vultr API keys (in order):
- `VULTR_API_KEY` (preferred)
- `VULTR_KEY` 
- `VULTR_TOKEN`

Add to your shell profile (`~/.profile`, `~/.kshrc`, `~/.bashrc`):
```bash
export VULTR_API_KEY="your-api-key-here"
```

### Without Vultr Key

If no API key is found, the script automatically configures a mock provider for testing purposes.

## Usage

### Setup and Start
```bash
./setup-chroot.sh start
```

### Manual Control
```bash
# Setup only
./setup-chroot.sh

# Start server
sudo /sbin/chroot /tmp/crush-test /bin/sh /run-crush.sh

# Or use wrapper script
/tmp/crush-test/../launch-crush-chroot.sh
```

### Cleanup
```bash
/tmp/crush-test/cleanup.sh
```

### Force Port
```bash
# Will pick next available if 8080 is in use
PORT=8082 ./setup-chroot.sh
```

## Files

| File | Purpose |
|------|---------|
| `setup-chroot.sh` | Main script - creates chroot and configures crush |
| `run-crush.sh` | Runs inside chroot, starts server |
| `cleanup.sh` | Removes chroot and unmounts |
| `launch-crush-chroot.sh` | Wrapper for easy launching |

## Technical Details

### Chroot Structure
```
/tmp/crush-test/
├── usr/
│   ├── lib/          (bind mount)
│   ├── local/lib/   (bind mount)
│   └── local/bin/   (bind mount)
├── home/
│   └── runner/
│       └── crush/   (full project)
├── run-crush.sh    (server launcher)
└── cleanup.sh      (cleanup script)
```

### Port Selection Algorithm
1. Checks if 8080 is available via `netstat`
2. If taken, increments to 8081, then 8082, etc.
3. Stops at 65535 if no ports available
4. Updates `crush.json` with selected port

### Safety Features
- **Port isolation**: Different port than main instance
- **File isolation**: Chroot prevents filesystem conflicts
- **Process isolation**: Runs in separate namespace
- **Cleanup**: Automatic mount/umount on exit
- **Validation**: Checks for missing dependencies

## Troubleshooting

### "No chroot utility found"
```bash
doas pkg_add chroot-utils
```

### "Mount failed" warnings
Are expected if directories aren't standardized - script continues with fallbacks.

### "Go not found in chroot"
```bash
# Ensure go is in system path
which go  # Should return /usr/local/bin/go on OpenBSD
```

### Port conflicts
The script automatically handles this - check the output startup message to see actual port used.

### Missing VULTR key
You can still test the interface fully with the mock provider - responses will simulate API calls.

## Integration

For project CI/testing, you can integrate this runner:

```bash
# In your CI pipeline
cd openbsd/runner
./setup-chroot.sh start &
sleep 10  # Wait for server startup
# Run tests against http://localhost:$PORT
test-endpoint "http://localhost:$(cat /tmp/crush-test/port.txt)"
/tmp/crush-test/cleanup.sh
```