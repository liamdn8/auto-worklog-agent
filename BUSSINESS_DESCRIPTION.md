# User Story

I need to create an agent running on Linux using Golang to support automatic development tracking. The application should watch development activity in IDEs including IntelliJ IDEA and VS Code.

# Assumptions

- I deployed an ActivityWatch server listening for work events
- The tracking application sends data via API to ActivityWatch server
- The agent runs continuously on a workstation
- The agent auto-discovers development workspaces without per-project setup

# Requirements

## 1. Work Session Tracking
A work session is identified by:
- Local Git user
- Git repository/project
- Git branch
- **Commit history** (all commits made during the session)

## 2. Worklog Automation Support
The work session data must support easy processing for worklog automation:
- **Session ends**: After idle for 30 minutes OR when explicitly flushed
- **Data format**: Rich JSON structure with all metadata
- **Commit tracking**: All commits captured with hash, message, author, timestamp
- **Issue detection**: Issue keys can be extracted from branch names and commit messages
- **Processing ready**: External tools (Jenkins, Python scripts, etc.) can query ActivityWatch API to:
  - Extract issue keys from commits
  - Calculate time spent per issue
  - Automatically post worklogs to Jira

## 3. Easy Installation
- Plug-and-play style application for Linux
- Single static binary (no dependencies)
- Works across all Linux distributions (Ubuntu, CentOS, Fedora, etc.)
- No Python, Node.js, or other runtime required

## 4. Auto-Discovery
- Automatically discover Git repositories in configured directories
- No per-project setup needed
- Continuously monitor for new repositories
- Reference implementations studied:
  - https://github.com/ActivityWatch/aw-watcher-window
  - https://github.com/ActivityWatch/aw-watcher-input
  - https://github.com/ActivityWatch/aw-watcher-afk

# Configuration

## 1. Server Endpoint
- Support configuration file (`config.json`)
- Support command-line arguments to override configuration
- Default: `http://localhost:5600`

## 2. Repository Scanning
- Rescan interval: 5 minutes (fixed)
- Scan depth: Configurable (default: 5 levels, 0 = unlimited)
- Root directories: Configurable list (e.g., `~/projects`, `~/repos`)

## 3. Session Management
- Idle timeout: 30 minutes
- Flush interval: 15 seconds
- Window poll interval: 1 second

# Implementation Status ✅

## Completed Features
- ✅ **Static binary build** - Single 8MB binary, no dependencies
- ✅ **Window detection** - Embedded window watcher (no aw-watcher-window needed)
- ✅ **Git auto-discovery** - Scans configured directories for repositories
- ✅ **Session tracking** - Tracks user, repo, branch, duration, events
- ✅ **Commit tracking** - Captures all commits made during sessions
- ✅ **ActivityWatch integration** - Creates buckets, publishes events
- ✅ **Cross-platform** - Works on all Linux distributions
- ✅ **Configuration** - File + CLI arguments support
- ✅ **Test mode** - Validation without window detection

## Data Structure
Each event includes:
```json
{
  "timestamp": "2025-11-04T10:00:00Z",
  "duration": 900.5,
  "data": {
    "gitUser": "username",
    "gitEmail": "user@example.com",
    "app": "code",
    "repoName": "project-name",
    "repoPath": "/path/to/repo",
    "branch": "feature/PROJ-123-new-feature",
    "remote": "git@github.com:user/repo.git",
    "eventCount": 42,
    "commits": [
      {
        "hash": "a1b2c3d4...",
        "message": "PROJ-123: Add new feature",
        "author": "username <user@example.com>",
        "timestamp": "2025-11-04T10:15:00Z"
      }
    ]
  }
}
```

## Future Integration
A separate tool (Jenkins pipeline, Python script, etc.) will:
1. Query ActivityWatch API for events
2. Extract issue keys from branch names and commit messages (regex: `[A-Z][A-Z0-9]+-\d+`)
3. Group sessions by issue key
4. Calculate total time per issue
5. Round to Jira-compatible intervals (15 minutes)
6. Post worklogs to Jira via REST API

This design maintains clean separation: **awagent = tracking**, **future-tool = Jira sync**.

## Installer behavior

- `install.sh <name>`: optional first argument to explicitly set the machine name written into `config.json` (`activityWatch.machine`).
- If no argument is provided, the installer will use `git config --global user.name` when set.
- If git global name is not set, the installer falls back to the machine's primary non-loopback IP address, and then to the hostname as a final fallback.

The `app` label was added to each event so consumers can differentiate which IDE/application produced the activity (for example: `code`, `idea`, `vim`, etc.).