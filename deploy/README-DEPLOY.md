# 🚀 Auto Worklog Agent - Ready to Deploy!

## ✅ This Binary Works on ALL Linux Distributions!

**Statically-linked** - No dependencies, no libraries, no Python needed!

- ✅ Ubuntu / Debian ✅ Fedora / RHEL ✅ Arch Linux
- ✅ CentOS ✅ openSUSE ✅ Alpine ✅ Any Linux x86_64

## Quick Start (Just 3 Steps!)

### 1. Copy to Your Linux Desktop

Transfer the `deploy/` folder to your Linux computer (USB, scp, etc.)

### 2. Run Setup

```bash
cd deploy
chmod +x quick-setup.sh awagent install.sh
./quick-setup.sh
```

### 3. Start Tracking!

```bash
./awagent --verbose
```

**Done!** View your activity at http://localhost:5600 🎉

## What You Get

- ✅ **One binary file** - Works everywhere (12MB, statically linked)
- ✅ **Auto-discovery** - Finds all Git repos automatically  
- ✅ **Smart window tracking** - Tries 5 different methods
- ✅ **Per-repo buckets** - Format: `user_repo_branch`
- ✅ **Zero config** - Works out of the box

## Requirements

**Must have:**
- Linux x86_64 (any distro)
- ActivityWatch server running
- One of: `xdotool`, `xprop`, `wmctrl` (usually pre-installed)

**That's it!** No Python, no pip, no libraries!

## Files in This Package

```
awagent              12MB  Static binary (works on all Linux)
quick-setup.sh       2KB   Automated setup script
install.sh           2KB   Systemd service installer  
config.sample.json   400B  Sample configuration
DEPLOYMENT.md        5KB   Detailed deployment guide
README.md            6KB   Main documentation
TESTING.md           3KB   Testing guide
```

## Installation Methods

### Method 1: Quick Setup (Recommended)

```bash
./quick-setup.sh
./awagent --verbose
```

### Method 2: Manual

```bash
# 1. Install window tool (if needed)
sudo apt-get install xdotool  # Ubuntu/Debian
sudo dnf install xdotool      # Fedora/RHEL
sudo pacman -S xdotool        # Arch

# 2. Create config
mkdir -p ~/.config/awagent
cp config.sample.json ~/.config/awagent/config.json
nano ~/.config/awagent/config.json  # Edit paths

# 3. Run
./awagent --verbose
```

### Method 3: As Systemd Service

```bash
./install.sh
systemctl --user status awagent
```

## Verify It Works

The binary is completely standalone:

```bash
# It's a static binary
file awagent
# Output: statically linked

# No external dependencies
ldd awagent  
# Output: not a dynamic executable

# Just run it!
./awagent --help
```

## Troubleshooting

**"No window detection method available"**
```bash
# Install xdotool (recommended)
sudo apt-get install xdotool
# Or xprop (lighter)
sudo apt-get install x11-utils
```

**"Connection refused"**
```bash
# Start ActivityWatch
docker run -d --name activitywatch \
  -p 5600:5600 \
  -v ~/.local/share/activitywatch:/data \
  activitywatch/activitywatch:latest
```

**"Permission denied"**
```bash
chmod +x awagent
```

## Configuration

Edit `~/.config/awagent/config.json`:

```json
{
  "roots": ["/home/YOUR_USERNAME"],
  "maxDepth": 5,
  "rescanIntervalMinutes": 5,
  "pulseTime": 30
}
```

## Command Line Options

```bash
./awagent --verbose              # Show detailed logs
./awagent --test                 # Simulate activity (for testing)
./awagent --aw-url http://...    # Custom ActivityWatch server
./awagent --config /path/to.json # Custom config file
./awagent --machine my-laptop    # Custom machine ID
```

## How Window Detection Works

The agent automatically tries these methods in order:

1. **xdotool** - Most reliable, commonly installed
2. **xprop** - Usually pre-installed on X11 systems  
3. **wmctrl** - Alternative window manager tool
4. **gdbus** - Works on GNOME desktops
5. **qdbus** - Works on KDE Plasma

**You only need ONE of these!** Most systems already have `xprop`.

## Features

### Auto-Discovery
- Scans configured directories for Git repositories
- No per-project setup needed
- Rescans every 5 minutes for new repos

### Smart Buckets
- One bucket per repo + branch combination
- Example: `lamdn8_myproject_feature-branch`
- Easy to track work across branches

### Session Management
- 30-minute idle timeout
- Auto-flush when switching repos
- Heartbeat-based tracking

## Performance

- **Memory**: ~10-20 MB
- **CPU**: <1%
- **Network**: Minimal (only heartbeats to ActivityWatch)
- **Disk**: 12 MB binary

## Supported Platforms

**Primary:** Linux x86_64 (all distributions)

**Also includes:** macOS and Windows implementations (in the same binary)

## What's Different from aw-watcher-window?

| Feature | aw-watcher-window | awagent |
|---------|-------------------|---------|
| Dependencies | Python + python-xlib | None (static binary) |
| Installation | pip install | Just copy |
| Size | Multiple files | Single 12MB file |
| Tracking | Window only | Window + Git integration |
| Buckets | One per app | One per repo+branch |
| Auto-discovery | No | Yes |
| Works on all Linux | Needs python-xlib | Yes (uses system tools) |

## Next Steps

1. ✅ Run `./quick-setup.sh`
2. ✅ Start agent: `./awagent --verbose`  
3. ✅ Open dashboard: http://localhost:5600
4. ✅ Install as service: `./install.sh` (optional)

## Support & Documentation

- **Quick Start**: This file
- **Detailed Guide**: DEPLOYMENT.md
- **Full Docs**: README.md
- **Testing**: TESTING.md

## View Logs

```bash
# If running as service
journalctl --user -u awagent -f

# If running manually
./awagent --verbose
```

---

**Ready to deploy!** Just copy this folder to your Linux desktop and run `./quick-setup.sh` 🚀
