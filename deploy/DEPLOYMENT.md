# Deployment Guide for Linux Desktop

## Prerequisites

1. **ActivityWatch Server** (running via Docker or native)
2. **One of these window detection tools** (usually pre-installed):
   - `xdotool` (recommended)
   - `xprop` (usually pre-installed on X11)
   - `wmctrl` (alternative)
   - `gdbus` (pre-installed on GNOME)
   - `qdbus` (pre-installed on KDE)
3. **Git repositories** to track

**Note:** The `awagent` binary is **statically linked** and works on **ALL Linux distributions** (Ubuntu, Debian, Fedora, Arch, CentOS, etc.) without any runtime dependencies!

## Installation Steps

### 1. Install Window Detection Tool (if needed)

The agent tries multiple methods automatically. On most systems, `xprop` is already installed. If not:

```bash
# Ubuntu/Debian
sudo apt-get install xdotool

# Fedora/RHEL
sudo dnf install xdotool

# Arch Linux
sudo pacman -S xdotool

# Or use xprop (usually pre-installed)
# Ubuntu/Debian: sudo apt-get install x11-utils
# Fedora/RHEL: sudo dnf install xorg-x11-utils
```

### 2. Deploy the Binary

```bash
# Copy the binary to your Linux desktop
# Option A: Direct copy if on same network
scp awagent user@your-desktop:~/

# Option B: Use USB drive or other transfer method
# Then on your desktop:
chmod +x ~/awagent
```

### 3. Create Configuration File

Create `~/.config/awagent/config.json`:

```json
{
  "roots": ["/home/YOUR_USERNAME"],
  "maxDepth": 5,
  "rescanIntervalMinutes": 5,
  "pulseTime": 30
}
```

Or copy the sample:

```bash
mkdir -p ~/.config/awagent
cp config.sample.json ~/.config/awagent/config.json
# Edit the file to set your home directory
```

### 4. Start ActivityWatch Server

```bash
# If using Docker (recommended):
docker run -d --name activitywatch \
  -p 5600:5600 \
  -v ~/.local/share/activitywatch:/data \
  activitywatch/activitywatch:latest

# Verify it's running:
curl http://localhost:5600/api/0/info
```

### 5. Test the Agent

```bash
# Test with verbose output
./awagent --verbose

# You should see:
# - Repository scan complete: indexed X repositories
# - Embedded window watcher started
# - Active window detection working
```

### 6. Install as Systemd Service (Optional)

```bash
# Create systemd user service
mkdir -p ~/.config/systemd/user

cat > ~/.config/systemd/user/awagent.service << 'EOF'
[Unit]
Description=ActivityWatch Development Tracking Agent
After=network.target

[Service]
Type=simple
ExecStart=%h/awagent
Restart=on-failure
RestartSec=10

[Install]
WantedBy=default.target
EOF

# Enable and start the service
systemctl --user daemon-reload
systemctl --user enable awagent
systemctl --user start awagent

# Check status
systemctl --user status awagent

# View logs
journalctl --user -u awagent -f
```

## Features

### Auto-Discovery
- Automatically discovers all Git repositories under configured root directories
- Scans up to configured depth (default: 5 levels)
- Rescans periodically (default: every 5 minutes) to detect new repositories

### Window Tracking
- Embeds window watcher functionality (no aw-watcher-window needed!)
- Uses python-xlib (same approach as official aw-watcher-window)
- Detects active application and window title

### Bucket Naming
- Format: `{user}.{repository}.{branch}`
- Example: `lamdn8_auto-worklog-agent_main`
- Automatically creates buckets per repo/branch combination

### Session Management
- 30-minute idle timeout
- Automatically flushes sessions when switching repos or going idle
- Heartbeat-based event publishing to ActivityWatch

## Troubleshooting

### No Active Window Detected

```bash
# Check which tools are available
which xdotool xprop wmctrl gdbus qdbus

# Install xdotool (recommended)
sudo apt-get install xdotool  # Ubuntu/Debian
sudo dnf install xdotool      # Fedora/RHEL
sudo pacman -S xdotool        # Arch

# Or install xprop
sudo apt-get install x11-utils  # Ubuntu/Debian

# Check if DISPLAY is set
echo $DISPLAY
```

### ActivityWatch Connection Failed

```bash
# Check if server is running
curl http://localhost:5600/api/0/info

# Check configured URL
./awagent --aw-url http://localhost:5600 --verbose
```

### No Repositories Found

```bash
# Verify your config
cat ~/.config/awagent/config.json

# Test with manual config
./awagent --config /path/to/config.json --verbose
```

## Command Line Options

```bash
# Show help
./awagent --help

# Custom ActivityWatch server URL
./awagent --aw-url http://other-host:5600

# Custom machine identifier
./awagent --machine my-laptop

# Custom config file
./awagent --config /path/to/config.json

# Verbose logging
./awagent --verbose

# Test mode (simulate activity without real window detection)
./awagent --test
```

## Files Included

- `awagent` - Main binary (12MB, statically linked)
- `config.sample.json` - Sample configuration
- `install.sh` - Systemd service installer script
- `README.md` - Full documentation
- `TESTING.md` - Testing guide

## Performance

- **Memory**: ~10-20 MB
- **CPU**: Negligible (<1%)
- **Network**: Minimal (only heartbeat updates to ActivityWatch)
- **Disk I/O**: Low (only during repository scans every 5 minutes)

## Next Steps

1. Install on your Linux desktop
2. Start ActivityWatch server
3. Install python-xlib
4. Run `./awagent --verbose` to test
5. Install as systemd service for automatic startup
6. View your development activity at http://localhost:5600

## Support

For issues or questions:
- Check the logs: `journalctl --user -u awagent -f`
- Run with `--verbose` flag for detailed output
- Verify dependencies are installed
- Check ActivityWatch server is accessible
