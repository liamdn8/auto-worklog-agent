# Configuration Guide

## Configuration File

awagent uses a JSON configuration file, typically located at `~/.config/awagent/config.json`.

## Complete Configuration Example

```json
{
  "activityWatch": {
    "baseURL": "http://localhost:5600",
    "clientName": "awagent",
    "hostname": ""
  },
  "git": {
    "roots": [
      "/home/user/projects",
      "/home/user/workspace",
      "/opt/repos"
    ],
    "repositories": [
      "/home/user/special-project"
    ],
    "maxDepth": 5,
    "rescanIntervalMin": 5
  },
  "session": {
    "pollIntervalSeconds": 1,
    "flushIntervalSeconds": 15,
    "idleTimeoutMinutes": 30
  }
}
```

## Configuration Sections

### ActivityWatch Settings

```json
"activityWatch": {
  "baseURL": "http://localhost:5600",
  "clientName": "awagent",
  "hostname": ""
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `baseURL` | string | `http://localhost:5600` | ActivityWatch server URL |
| `clientName` | string | `awagent` | Client identifier in ActivityWatch |
| `hostname` | string | *(auto-detected)* | Machine hostname for bucket naming |

**Examples:**

Remote ActivityWatch server:
```json
"baseURL": "http://192.168.1.100:5600"
```

Custom hostname:
```json
"hostname": "dev-workstation"
```

### Git Repository Settings

```json
"git": {
  "roots": ["/home/user/projects"],
  "repositories": [],
  "maxDepth": 5,
  "rescanIntervalMin": 5
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `roots` | array | `[]` | Root directories to scan for Git repos |
| `repositories` | array | `[]` | Specific repositories to track (optional) |
| `maxDepth` | int | `5` | Maximum directory depth to scan (0 = unlimited) |
| `rescanIntervalMin` | int | `5` | Minutes between repository rescans |

**Examples:**

Multiple scan roots:
```json
"roots": [
  "/home/user/work",
  "/home/user/personal",
  "/opt/company-repos"
]
```

Deep scanning (unlimited depth):
```json
"maxDepth": 0
```

Specific repositories only:
```json
"repositories": [
  "/home/user/critical-project",
  "/opt/legacy-app"
],
"roots": []
```

### Session Settings

```json
"session": {
  "pollIntervalSeconds": 1,
  "flushIntervalSeconds": 15,
  "idleTimeoutMinutes": 30
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `pollIntervalSeconds` | int | `1` | Window polling interval (seconds) |
| `flushIntervalSeconds` | int | `15` | Event flush interval (seconds) |
| `idleTimeoutMinutes` | int | `30` | Idle timeout before ending session (minutes) |

**Performance Tuning:**

Lower CPU usage (slower polling):
```json
"pollIntervalSeconds": 2
```

More frequent updates:
```json
"flushIntervalSeconds": 10
```

Longer sessions:
```json
"idleTimeoutMinutes": 60
```

## Command-Line Arguments

Command-line arguments override configuration file settings:

```bash
awagent [flags]

Flags:
  --config string      Path to config file (default: ./config.json)
  --aw-url string      Override ActivityWatch server URL
  --machine string     Override machine hostname
  --test               Run in test mode (simulate IDE activity)
  -v, --verbose        Enable verbose logging
  -h, --help           Show help
```

### Examples

**Custom config location:**
```bash
awagent --config ~/.config/awagent/config.json
```

**Override server URL:**
```bash
awagent --aw-url http://192.168.1.100:5600
```

**Test mode with verbose logging:**
```bash
awagent --test -v
```

**Override machine name:**
```bash
awagent --machine dev-laptop
```

**Multiple overrides:**
```bash
awagent --config /etc/awagent.json --aw-url http://server:5600 --machine ci-runner -v
```

## Environment Variables

Currently not supported. Use command-line arguments for runtime overrides.

## Configuration Strategies

### Strategy 1: Single User, Local ActivityWatch

```json
{
  "activityWatch": {
    "baseURL": "http://localhost:5600"
  },
  "git": {
    "roots": ["/home/username/projects"]
  }
}
```

### Strategy 2: Multi-User, Shared ActivityWatch

```json
{
  "activityWatch": {
    "baseURL": "http://activitywatch-server.company.com:5600",
    "hostname": "dev-laptop-john"
  },
  "git": {
    "roots": ["/home/john/workspace"]
  }
}
```

### Strategy 3: CI/CD Environment

```json
{
  "activityWatch": {
    "baseURL": "http://aw-server:5600",
    "hostname": "jenkins-agent-1"
  },
  "git": {
    "repositories": ["/workspace/current-build"],
    "roots": []
  },
  "session": {
    "idleTimeoutMinutes": 5,
    "flushIntervalSeconds": 5
  }
}
```

### Strategy 4: Development Workstation

```json
{
  "activityWatch": {
    "baseURL": "http://localhost:5600"
  },
  "git": {
    "roots": [
      "/home/dev/work",
      "/home/dev/opensource"
    ],
    "maxDepth": 3
  },
  "session": {
    "pollIntervalSeconds": 1,
    "flushIntervalSeconds": 15,
    "idleTimeoutMinutes": 30
  }
}
```

## Configuration Validation

### Check if config is valid:

```bash
# Test with verbose logging
awagent --config config.json --test -v

# Should show:
# - Configuration loaded
# - Repository scan results
# - Session tracking
```

### Common Validation Errors:

**Invalid JSON:**
```
Error: invalid character '}' looking for beginning of object key string
```
Fix: Check JSON syntax (use `jq` to validate):
```bash
jq '.' config.json
```

**Invalid URL:**
```
Error: cannot connect to ActivityWatch server
```
Fix: Check baseURL format (must include `http://`)

**Path doesn't exist:**
```
Warning: repository scan: skip root /invalid/path
```
Fix: Ensure paths in `roots` exist

## Default Values

If configuration file is missing or incomplete, defaults are used:

| Setting | Default Value |
|---------|---------------|
| ActivityWatch URL | `http://localhost:5600` |
| Client Name | `awagent` |
| Hostname | System hostname |
| Git Roots | `[]` (empty - must configure!) |
| Max Depth | `5` |
| Rescan Interval | `5` minutes |
| Poll Interval | `1` second |
| Flush Interval | `15` seconds |
| Idle Timeout | `30` minutes |

## Security Considerations

### File Permissions

```bash
# Restrict config file to user only
chmod 600 ~/.config/awagent/config.json
```

### Network Security

If using remote ActivityWatch server:
- Use HTTPS if possible (configure reverse proxy)
- Use firewall rules to restrict access
- Consider VPN for remote access

### Data Privacy

awagent captures:
- ✅ Git metadata (user, email, branch, remote URL)
- ✅ Commit hashes and messages
- ✅ Window titles (IDE names and project names)
- ✅ Activity timestamps and durations

awagent does NOT capture:
- ❌ Code content
- ❌ File contents
- ❌ Passwords or secrets
- ❌ Keystrokes
- ❌ Screenshots

## Advanced Configuration

### Multiple Configurations

Run different instances for different purposes:

```bash
# Personal projects
awagent --config ~/.config/awagent/personal.json &

# Work projects
awagent --config ~/.config/awagent/work.json &
```

### Dynamic Configuration

Generate config programmatically:

```bash
#!/bin/bash
cat > config.json << EOF
{
  "activityWatch": {
    "baseURL": "${AW_SERVER:-http://localhost:5600}"
  },
  "git": {
    "roots": ["${PROJECT_ROOT:-$HOME/projects}"]
  }
}
EOF

awagent --config config.json
```

## Next Steps

- [Usage Guide](usage.md) - Daily usage patterns
- [Troubleshooting](troubleshooting.md) - Common configuration issues
- [Development](development.md) - Modify configuration schema
