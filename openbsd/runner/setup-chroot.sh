#!/usr/local/bin/ksh
# ksh script to setup chroot and start crush with Vultr/Kimi config
# Integrated into project, works for any user

set -e

CHROOT_DIR="/tmp/crush-test"
PROJECT_SRC="$(dirname "$0")/../.."
PROJECT_DIR="$CHROOT_DIR/home/runner/crush"

echo "=== Crush Chroot + Dynamic Config Setup ==="
echo "Chroot: $CHROOT_DIR"
echo "Project: $PROJECT_SRC"

# Source user's profile  
for profile_file in "$HOME/.profile" "$HOME/.kshrc" "$HOME/.bashrc"; do
    if [ -f "$profile_file" ]; then
        echo "Sourcing profile: $profile_file"
        . "$profile_file"
        break
    fi
done

# Try multiple env var names
VULTR_KEY=""
for key_var in VULTR_API_KEY VULTR_KEY VULTR_TOKEN; do
    eval "key_value=\${$key_var:-}"
    if [ -n "$key_value" ]; then
        VULTR_KEY="$key_value"
        echo "✓ Found $key_var: ***${key_value:0:8}..."
        break
    fi
done

if [ -z "$VULTR_KEY" ]; then
    echo "⚠ No Vultr key found - will use mock providers for testing"
    USE_MOCK=true
else
    USE_MOCK=false
fi

# Find available port for macOS compatibility
find_available_port() {
    local base_port=$1
    local port=$base_port
    
    while /bin/netstat -an | /usr/bin/grep ":${port}" >/dev/null 2>&1; do
        port=$((port + 1))
        if [ $port -gt 65535 ]; then
            echo "No available ports found!"
            exit 1
        fi
    done
    
    echo "$port"
}

PORT=$(find_available_port 8080)
echo "Server will run on port: $PORT"

# Cleanup previous chroot
if [ -d "$CHROOT_DIR" ]; then
    echo "Cleaning previous chroot..."
    chmod -R 755 "$CHROOT_DIR" 2>/dev/null || true
    /bin/rm -rf "$CHROOT_DIR"
fi

# Create chroot structure
echo "Creating chroot structure..."
mkdir -p "$CHROOT_DIR"/usr/{lib,libexec,bin} "$CHROOT_DIR"/{tmp,var/tmp,dev,etc,home/runner}

# Use bind mounts for libraries  
setup_mounts() {
    echo "Setting up bind mounts..."
    
    local directories="/usr/lib /usr/local/lib /usr/local/bin /usr/local/libexec"
    for dir in $directories; do
        if [ -d "$dir" ]; then
            mkdir -p "$CHROOT_DIR"$(dirname "$dir") 2>/dev/null || true
            /sbin/mount -o bind,ro "$dir" "$CHROOT_DIR$dir" 2>/dev/null || echo "Warning: Could not mount $dir"
        fi
    done
}

setup_mounts

# Copy entire project
echo "Copying project files..."
mkdir -p "$PROJECT_DIR"
/bin/cp -R "$PROJECT_SRC"/. "$PROJECT_DIR/"

# Create dynamic config
echo "Creating environment config..."
/bin/cat > "$PROJECT_DIR/chroot-config.json" << EOF
{
  "active_providers": [$([ "$USE_MOCK" = true ] && echo '"mock"' || echo '"vultr"')],
  "providers": {
    $(if [ "$USE_MOCK" = true ]; then cat << 'MOCK'
    "mock": {
      "config": {
        "model": "gpt-3.5-turbo",
        "response": "This is a mock response for testing"
      },
      "name": "Mock Provider",
      "type": "mock", 
      "enabled": true,
      "systems": false
    }
MOCK
    else cat <<VULTR
    "vultr": {
      "config": {
        "api_key": "$VULTR_KEY",
        "model": "kimi",
        "max_tokens": 8000,
        "temperature": 0.7
      },
      "name": "Vultr",
      "type": "vultr",
      "enabled": true,
      "systems": false
    }
VULTR
    fi)
  },
  "settings": {
    "server": {
      "port": $PORT,
      "bind": "0.0.0.0"
    }
  }
}
EOF

# Check if crush.json exists and merge configs
if [ -f "$PROJECT_DIR/crush.json" ]; then
    echo "Backing up original crush.json -> crush.json.backup"
    /bin/cp "$PROJECT_DIR/crush.json" "$PROJECT_DIR/crush.json.backup"
fi

# Use chroot config
/bin/cp "$PROJECT_DIR/chroot-config.json" "$PROJECT_DIR/crush.json"

# Create run script
echo "Creating environment run script..."
/bin/cat > "$CHROOT_DIR/run-crush.sh" << 'EOF'
#!/bin/sh
set -e

export PATH=/usr/local/bin:$PATH
export HOME=/home/runner
export GOPATH=/home/runner/go
cd /home/runner/crush

echo "=== Crush Server in Chroot ==="
echo "Go version: $(go version 2>/dev/null || echo 'Go not found')"
echo "Project loaded from: $(pwd)"
echo "Port: $(grep -o '"port": [0-9]*' crush.json | cut -d' ' -f2 || echo 'unknown')"
echo "Active provider: $(grep -o '"active_providers": \["[^"]*"' crush.json | sed 's/.*\[\?"//g' | sed 's/"//g')"

# Build if needed
if [ ! -f "./crush" ] || [ "./crush" -ot "./go.mod" ]; then
    echo "Building crush..."
    /usr/local/bin/go mod tidy
    /usr/local/bin/go build -o crush . 2>/dev/null || {
        echo "Build failed, trying without cgo..."
        CGO_ENABLED=0 /usr/local/bin/go build -o crush .
    }
    echo "Build complete"
fi

# Start server - explicit port for safety
PORT=$(grep -o '"port": [0-9]*' crush.json | tail -1 | cut -d' ' -f2)
echo "Starting server on port $PORT..."
echo "Access from host: http://localhost:$PORT"
echo "Terminate with: Ctrl+C"

exec ./crush server --port "$PORT" --bind 0.0.0.0
EOF

/bin/chmod +x "$CHROOT_DIR/run-crush.sh"

# Create launch wrapper
echo "Creating launch wrapper..."
/bin/cat > "$CHROOT_DIR/../launch-crush-chroot.sh" << EOF
#!/bin/sh
# Wrapper to start crush chroot server
CHROOT_DIR="/tmp/crush-test"

echo "=== Starting Crush in Chroot ==="
echo "Chroot directory: $CHROOT_DIR"

cd /tmp
if [ -f "\$CHROOT_DIR/cleanup.sh" ]; then
    ./\$CHROOT_DIR/cleanup.sh
fi

if [ -x /sbin/chroot ]; then
    /sbin/chroot "$CHROOT_DIR" /bin/sh /run-crush.sh
elif [ -x /usr/bin/chroot ]; then
    /usr/bin/chroot "$CHROOT_DIR" /bin/sh /run-crush.sh
else
    echo "No chroot utility found. Install with: doas pkg_add chroot-utils"
    exit 1
fi
EOF

chmod +x "$CHROOT_DIR/../launch-crush-chroot.sh"

# Create cleanup script
echo "Creating cleanup script..."
/bin/cat > "$CHROOT_DIR/cleanup.sh" << 'EOF'
#!/bin/sh
echo "=== Cleaning up chroot ==="

# Unmount in reverse order
for mount_point in /usr/local/libexec /usr/local/bin /usr/local/lib /usr/lib /proc; do
    if /sbin/mount | /usr/bin/grep -q "$mount_point"; then
        echo "Unmounting /tmp/crush-test$mount_point..."
        /sbin/umount "/tmp/crush-test$mount_point" 2>/dev/null || true
    fi
done

# Remove chroot
echo "Removing chroot directory..."
/bin/rm -rf "/tmp/crush-test"
echo "Cleanup complete"
EOF

/bin/chmod +x "$CHROOT_DIR/cleanup.sh"

echo "=== Setup Complete ==="
echo "Configuration: $([ "$USE_MOCK" = true ] && echo "Mock Provider" || echo "Vultr (Kimi)")"
echo "Port: $PORT"
echo "Launch: $CHROOT_DIR/../launch-crush-chroot.sh"
echo "Direct: sudo /sbin/chroot $CHROOT_DIR /bin/sh /run-crush.sh"
echo "Cleanup: $CHROOT_DIR/cleanup.sh"

if [ "$1" = "start" ]; then
    echo
    echo "Starting server..."
    exec "$CHROOT_DIR/../launch-crush-chroot.sh"
fi