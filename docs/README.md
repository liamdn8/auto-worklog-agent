# awagent Documentation

**awagent** (ActivityWatch Agent) is a lightweight development activity tracker for Linux that automatically monitors your coding sessions and captures Git commit metadata.

## Table of Contents

1. [Overview](overview.md) - What is awagent and how it works
2. [Installation](installation.md) - How to install and deploy
3. [Configuration](configuration.md) - Configuration options and settings
4. [Usage](usage.md) - Running and using awagent
5. [Architecture](architecture.md) - Technical architecture and design
6. [Commit Tracking](commit-tracking.md) - How commit tracking works
7. [Data Format](data-format.md) - Event data structure and schema
8. [Integration](integration.md) - Integrating with Jira and other tools
9. [Troubleshooting](troubleshooting.md) - Common issues and solutions
10. [Development](development.md) - Building and developing awagent

## Quick Start

```bash
# 1. Download binary
wget https://github.com/liamdn8/auto-worklog-agent/releases/latest/download/awagent
chmod +x awagent

# 2. Create config
cat > config.json << EOF
{
  "activityWatch": {
    "baseURL": "http://localhost:5600"
  },
  "git": {
    "roots": ["/home/youruser/projects"]
  }
}
EOF

# 3. Run
./awagent --config config.json
```

## Features

✅ **Automatic Git Discovery** - Scans configured directories for repositories  
✅ **Commit Tracking** - Captures all commits with metadata  
✅ **Window Detection** - Monitors IDE activity (VSCode, IntelliJ, etc.)  
✅ **Static Binary** - 8MB binary, no dependencies  
✅ **Session Management** - Groups activity into work sessions  
✅ **ActivityWatch Integration** - Stores data in ActivityWatch server  

## Key Concepts

### Work Session
A work session represents a continuous period of activity on a Git repository:
- Starts when IDE activity is detected
- Tracked by: user, repository, branch
- Captures all commits made during the session
- Ends after 30 minutes of inactivity

### Bucket
Each unique combination of `user_repository_branch` gets its own bucket in ActivityWatch:
```
lamdn8_auto-worklog-agent_main
lamdn8_my-project_feature-branch
```

### Event
An event is published when a session ends, containing:
- Session duration
- Git metadata (user, email, branch, remote)
- All commits made during the session
- Activity count

## System Requirements

- **OS**: Linux (any distribution)
- **Architecture**: x86_64 (amd64)
- **Dependencies**: None (static binary)
- **ActivityWatch**: Server running (v0.12.0+)
- **Git**: Installed and configured

## License

MIT License - See [LICENSE](../LICENSE) file for details.

## Support

- **Issues**: https://github.com/liamdn8/auto-worklog-agent/issues
- **Documentation**: https://github.com/liamdn8/auto-worklog-agent/tree/main/docs
- **Examples**: See [examples/](../examples/) directory
