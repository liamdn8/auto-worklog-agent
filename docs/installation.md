# Installation Guide

## Prerequisites

### Required
- **Linux OS** - Any distribution (Ubuntu, CentOS, Fedora, Debian, Arch, etc.)
- **Architecture** - x86_64 (amd64)
- **Git** - Installed and configured with user.name and user.email
- **ActivityWatch Server** - Running and accessible

### Optional
- **Docker** - For running ActivityWatch server locally
- **Window Manager Tools** - xdotool, xprop, wmctrl (usually pre-installed)

## Step 1: Install ActivityWatch Server

### Option A: Using Docker (Recommended)

```bash
# Create docker-compose.yaml
cat > docker-compose.yaml << 'EOF'
version: '3.8'

services:
  aw-server:
    image: activitywatch/aw-server:v0.13.2
    container_name: aw-server
    ports:
      - "5600:5600"
    volumes:
      - aw-data:/root/.local/share/activitywatch
    environment:
      - AW_CORS_ORIGINS=*
      - AW_HOST=0.0.0.0
      - AW_PORT=5600
    restart: unless-stopped

volumes:
  aw-data:
EOF

# Start server
docker-compose up -d

# Verify
curl http://localhost:5600/api/0/info
```

### Option B: Native Installation

Download from: https://activitywatch.net/downloads/

```bash
# Extract and run
tar -xzf activitywatch-*.tar.gz
cd activitywatch
./aw-qt
```

## Step 2: Install awagent

### Option A: Download Binary (Recommended)

```bash
# Download latest release
wget https://github.com/liamdn8/auto-worklog-agent/releases/latest/download/awagent

# Make executable
chmod +x awagent

# Move to system path (optional)
sudo mv awagent /usr/local/bin/
```

### Option B: Build from Source

```bash
# Prerequisites: Go 1.21+
go version

# Clone repository
git clone https://github.com/liamdn8/auto-worklog-agent.git
cd auto-worklog-agent

# Build static binary
./build.sh

# Binary is in: ./deploy/awagent
```

## Step 3: Create Configuration

```bash
# Create config directory
mkdir -p ~/.config/awagent

# Create config file
cat > ~/.config/awagent/config.json << 'EOF'
{
  "activityWatch": {
    "baseURL": "http://localhost:5600",
    "clientName": "awagent",
    "hostname": ""
  },
  "git": {
    "roots": [
      "/home/youruser/projects",
      "/home/youruser/workspace"
    ],
    "repositories": [],
    "maxDepth": 5,
    "rescanIntervalMin": 5
  },
  "session": {
    "pollIntervalSeconds": 1,
    "flushIntervalSeconds": 15,
    "idleTimeoutMinutes": 30
  }
}
EOF

# Replace 'youruser' with your actual username
sed -i "s/youruser/$USER/g" ~/.config/awagent/config.json
```

## Step 4: Verify Git Configuration

```bash
# Check Git is configured
git config --global user.name
git config --global user.email

# If not set:
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

## Step 5: Test Installation

### Basic Test
```bash
# Run in test mode (simulates IDE activity)
./awagent --config ~/.config/awagent/config.json --test -v

# You should see:
# - Repository scan complete: indexed X repositories
# - Session started repo=...
# - Activity detected repo=...
```

### Real Usage Test
```bash
# Run normally (will monitor actual windows)
./awagent --config ~/.config/awagent/config.json -v

# Open VSCode or IntelliJ with a Git repository
# Check logs for: "Activity detected repo=..."
```

### Verify in ActivityWatch
```bash
# List buckets
curl http://localhost:5600/api/0/buckets/ | jq 'keys'

# Should see buckets like:
# ["youruser_project-name_branch", ...]

# Check events
curl http://localhost:5600/api/0/buckets/BUCKET_ID/events | jq '.'
```

## Step 6: Run as System Service (Optional)

### Using systemd

```bash
# Create service file
sudo tee /etc/systemd/system/awagent.service << 'EOF'
[Unit]
Description=ActivityWatch Development Tracker
After=network.target

[Service]
Type=simple
User=youruser
ExecStart=/usr/local/bin/awagent --config /home/youruser/.config/awagent/config.json
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Replace 'youruser' with your username
sudo sed -i "s/youruser/$USER/g" /etc/systemd/system/awagent.service

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable awagent
sudo systemctl start awagent

# Check status
sudo systemctl status awagent

# View logs
sudo journalctl -u awagent -f
```

### Using cron (startup on login)

```bash
# Add to crontab
(crontab -l 2>/dev/null; echo "@reboot /usr/local/bin/awagent --config ~/.config/awagent/config.json >> ~/awagent.log 2>&1") | crontab -
```

### Using supervisord

```bash
# Install supervisor
sudo apt-get install supervisor  # Ubuntu/Debian
# or
sudo yum install supervisor      # CentOS/RHEL

# Create config
sudo tee /etc/supervisor/conf.d/awagent.conf << EOF
[program:awagent]
command=/usr/local/bin/awagent --config /home/$USER/.config/awagent/config.json
user=$USER
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/awagent.log
EOF

# Start
sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl start awagent
```

## Step 7: Verify Everything Works

### Create Test Commits

```bash
# Go to a Git repository
cd ~/projects/some-project

# Make a test commit
echo "test" > test.txt
git add test.txt
git commit -m "TEST-123: Test commit for awagent"

# Wait 15-30 seconds

# Check ActivityWatch for the commit
USER=$(git config user.name | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9_-]/-/g')
REPO=$(basename $(pwd))
BRANCH=$(git branch --show-current | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9_-]/-/g')
BUCKET="${USER}_${REPO}_${BRANCH}"

curl -s "http://localhost:5600/api/0/buckets/${BUCKET}/events?limit=1" | jq '.[0].data.commits'

# Should see your test commit!
```

## Troubleshooting

### Issue: "No repositories found"

**Solution:** Check your Git roots in config.json
```bash
# Verify config
cat ~/.config/awagent/config.json | jq '.git.roots'

# Check if paths exist
ls -la /home/youruser/projects
```

### Issue: "Window detection failed"

**Solution:** Install window manager tools
```bash
# Ubuntu/Debian
sudo apt-get install xdotool wmctrl

# CentOS/RHEL
sudo yum install xdotool wmctrl

# Fedora
sudo dnf install xdotool wmctrl
```

### Issue: "Cannot connect to ActivityWatch"

**Solution:** Check if server is running
```bash
# Test connection
curl http://localhost:5600/api/0/info

# If using Docker
docker ps | grep aw-server

# Check Docker logs
docker logs aw-server
```

### Issue: "Permission denied"

**Solution:** Ensure binary is executable
```bash
chmod +x awagent
```

### Issue: "Sessions not appearing"

**Solution:** Check logs
```bash
# Run with verbose logging
./awagent --config ~/.config/awagent/config.json -v

# Or if running as service
sudo journalctl -u awagent -f
```

## Next Steps

- [Configuration Guide](configuration.md) - Customize settings
- [Usage Guide](usage.md) - Daily usage patterns
- [Troubleshooting](troubleshooting.md) - Common issues

## Uninstallation

```bash
# Stop service
sudo systemctl stop awagent
sudo systemctl disable awagent

# Remove files
sudo rm /etc/systemd/system/awagent.service
sudo rm /usr/local/bin/awagent
rm -rf ~/.config/awagent

# Remove ActivityWatch data (optional)
docker-compose down -v  # If using Docker
```
