# Auto Worklog Agent

Go-based agent that automatically tracks development activity across Git repositories on your workstation. The agent has an **embedded window watcher** (no need for external aw-watcher-window), monitors IDE activity, correlates it with Git repositories, and publishes work sessions to an ActivityWatch server. Sessions are identified by **Git user, repository name, and branch** for easy downstream automation (e.g., Jenkins aggregation and Jira worklog updates).

## Features
- **Embedded window watching**: Built-in window activity detection (no aw-watcher-window dependency)
- **Auto-discovery**: Automatically scans configured directory trees to find Git repositories (no per-project setup required)
- **IDE integration**: Detects activity in VS Code, IntelliJ IDEA, PyCharm, GoLand, and other IDEs
- **Smart bucket naming**: Buckets named as `user_repo_branch` for easy querying (e.g., `lamdn8_auto-worklog-agent_main`)
- **Smart session management**: Automatically closes sessions after 30 minutes of inactivity (configurable)
- **Flexible configuration**: Supports JSON config files with environment variable expansion, plus CLI argument overrides
- **Depth-controlled scanning**: Configure how deep to scan for repositories (default: 5 levels, use 0 for unlimited)
- **Periodic rescanning**: Automatically discovers new repositories every 5 minutes (configurable)

## Prerequisites
- Go 1.21+
- Docker (for running a local ActivityWatch server)
- **Linux**: `xdotool` or `xprop` for window detection
  ```bash
  sudo apt-get install xdotool  # Ubuntu/Debian
  sudo yum install xdotool      # RHEL/CentOS
  ```
- **macOS**: Built-in AppleScript support
- **Windows**: Built-in PowerShell support
- Git repositories available on the local filesystem

## Quick Start

### Option 1: Run with Docker Compose (Testing)

1. **Start ActivityWatch server:**

   ```bash
   docker compose up -d
   ```

   This launches `aw-server` at `http://localhost:5600` with data persisted in the `aw_data` volume.

2. **Build and run the agent in test mode:**

   ```bash
   ./build.sh
   ./bin/awagent --test
   ```

   Test mode simulates IDE activity without requiring aw-watcher-window. The agent will auto-discover Git repositories in your home directory and create simulated work sessions every 10 seconds.

3. **View the results:**

   ```bash
   # Check buckets created
   curl http://localhost:5600/api/0/buckets/ | jq
   
   # View events
   curl 'http://localhost:5600/api/0/buckets/awagent.auto-worklog-agent/events' | jq
   
   # Or open web UI
   open http://localhost:5600
   ```

See [TESTING.md](TESTING.md) for more testing options.

### Option 2: Run with aw-watcher-window (Production)

For production use with real IDE activity detection:

1. **Start ActivityWatch server:**

   ```bash
   docker compose up -d
   ```

2. **Install and start aw-watcher-window:**

   Follow instructions at https://github.com/ActivityWatch/aw-watcher-window to install the window watcher for your platform. This is required for the agent to detect IDE activity.

2. **Install and start aw-watcher-window:**

   Follow instructions at https://github.com/ActivityWatch/aw-watcher-window to install the window watcher for your platform. This is required for the agent to detect IDE activity.

3. **Build and run the agent:**

   ```bash
   ./build.sh
   ./bin/awagent
   ```

   The agent will auto-discover Git repositories and track your IDE activity.

### Option 3: Install as systemd User Service (Recommended)

1. **Run the install script:**

   ```bash
   ./install.sh
   ```

   This will:
   - Build the binary and copy it to `~/.local/bin/awagent`
   - Create config at `~/.config/awagent/config.json`
   - Install systemd user service (no root required)

2. **Edit the configuration:**

   ```bash
   nano ~/.config/awagent/config.json
   ```

   Configure your repository scan roots and other settings.

3. **Start the service:**

   ```bash
   systemctl --user start awagent
   systemctl --user enable awagent  # Enable at login
   ```

4. **Check status and logs:**

   ```bash
   systemctl --user status awagent
   journalctl --user -u awagent -f
   ```

## Configuration Reference

Create or edit `~/.config/awagent/config.json`:

   ```jsonc
   {
     "activityWatch": {
       "baseURL": "http://localhost:5600",
       "bucketPrefix": "awagent",
       "machine": "developer-workstation"
     },
     "git": {
       "repositories": [],
       "roots": [
         "$HOME/projects",
         "$HOME/dev"
       ],
       "maxDepth": 0,
       "rescanIntervalMin": 5
     },
     "session": {
       "idleTimeoutMinutes": 30,
       "pollInterval": "5s",
       "flushInterval": "15s",
       "pulseTime": "10s"
     }
   }
   ```

**Configuration options:**
- `repositories`: Explicit list of repo paths (optional if using `roots`)
- `roots`: Directory trees to scan for Git repositories
- `maxDepth`: How deep to scan (0 = unlimited, default: 5)
- `rescanIntervalMin`: How often to rescan for new repositories (default: 5 minutes)
- `idleTimeoutMinutes`: Inactivity timeout before closing a session (default: 30)
- `pollInterval`: How often to poll window events (default: 5s)
- `pulseTime`: ActivityWatch heartbeat merge window (default: 10s)

**CLI Overrides:**

   ```bash
   awagent --config ./config.json --aw-url http://custom-server:5600 --machine my-laptop
   ```

## How It Works
- The agent polls the `aw-watcher-window` bucket to detect IDE activity.
- When a window title matches a known IDE and contains a repository name, activity is recorded for that session.
- Sessions are grouped by repository path and branch.
- After the configured idle timeout (default 30 min), sessions are flushed to ActivityWatch as events.
- Every 5 minutes (configurable), the agent rescans configured roots to discover new repositories.
- Events include Git metadata (user, email, remote, branch) for easy downstream processing.

## Next Steps
- Set up aw-watcher-window on your development machine
- Configure Jenkins pipeline to query ActivityWatch API and sync worklogs to Jira
- Implement push detection hooks to flush sessions immediately after `git push` operations

## Development

```bash
go test ./...
```

To rebuild binaries:

```bash
go build -o awagent ./cmd/awagent
```