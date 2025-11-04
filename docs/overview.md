# Overview

## What is awagent?

**awagent** (ActivityWatch Agent) is an automated development activity tracker for Linux workstations. It monitors your IDE activity and Git commits, storing structured data in ActivityWatch for later analysis and worklog automation.

## How It Works

```
┌──────────────────────────────────────────────────────────────┐
│                    Developer Workstation                     │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │  VSCode    │  │ IntelliJ   │  │  Terminal  │            │
│  │            │  │    IDEA    │  │            │            │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘            │
│        │               │               │                    │
│        └───────────────┴───────────────┘                    │
│                        │                                    │
│                        ▼                                    │
│              ┌──────────────────┐                           │
│              │    awagent       │◄─── Embedded window       │
│              │                  │     detection (xdotool,   │
│              │  • Detects IDE   │     xprop, wmctrl, etc.)  │
│              │  • Scans repos   │                           │
│              │  • Tracks commits│                           │
│              └────────┬─────────┘                           │
│                       │                                     │
└───────────────────────┼─────────────────────────────────────┘
                        │ HTTP API
                        ▼
┌──────────────────────────────────────────────────────────────┐
│              ActivityWatch Server (Docker)                   │
├──────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────┐ │
│  │ SQLite Database                                        │ │
│  │                                                        │ │
│  │ Buckets:                                               │ │
│  │  • lamdn8_auto-worklog-agent_main                      │ │
│  │  • lamdn8_my-project_feature-PROJ-123                  │ │
│  │                                                        │ │
│  │ Events: {timestamp, duration, data: {...}}            │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  Web UI: http://localhost:5600                              │
└──────────────────────────────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│           Future: Jira Worklog Sync Tool                     │
├──────────────────────────────────────────────────────────────┤
│  1. Query ActivityWatch API                                  │
│  2. Extract issue keys from commits                          │
│  3. Calculate time per issue                                 │
│  4. Post to Jira REST API                                    │
└──────────────────────────────────────────────────────────────┘
```

## Key Features

### 1. Automatic Git Discovery
- Scans configured root directories (e.g., `~/projects`)
- Finds all Git repositories recursively
- Rescans every 5 minutes to detect new projects
- No per-project configuration needed

### 2. Window Detection
- Monitors active window title and application name
- Detects known IDEs: VSCode, IntelliJ IDEA, PyCharm, GoLand, etc.
- Matches window titles to Git repository names
- Works with multiple detection methods (xdotool, xprop, wmctrl, gdbus, qdbus)

### 3. Session Tracking
A **session** is created when IDE activity is detected:
- Identified by: `gitUser_repoName_branch`
- Tracks activity count and duration
- Captures Git metadata (user, email, remote URL)
- Stores all commits made during the session

### 4. Commit Tracking
- Captures commit hash when session starts
- On each activity check, fetches new commits since session start
- Stores complete commit metadata:
  - Full commit hash
  - Commit message (contains issue keys!)
  - Author name and email
  - Commit timestamp

### 5. Event Publishing
Sessions are flushed to ActivityWatch when:
- 30 minutes of inactivity (idle timeout)
- Agent shutdown (graceful or SIGTERM)
- Every 15 seconds if active (heartbeat)

## Workflow Example

### Scenario: Developer works on JIRA-123

```
09:00 - Developer opens VSCode with project "my-app"
        ├─ awagent detects window: "VSCode - my-app"
        ├─ Matches to repo: /home/dev/projects/my-app
        ├─ Creates session: dev_my-app_feature-JIRA-123
        └─ Captures start commit: abc1234

09:15 - Developer makes first commit
        ├─ git commit -m "JIRA-123: Add user authentication"
        └─ awagent detects commit on next check

09:30 - Developer makes second commit
        ├─ git commit -m "JIRA-123: Add login form validation"
        └─ awagent updates session commits array

09:45 - Developer takes coffee break
        └─ No activity for 30 minutes...

10:15 - Session flushed to ActivityWatch
        ├─ Duration: 45 minutes
        ├─ Commits: 2 (both JIRA-123)
        └─ Bucket: dev_my-app_feature-JIRA-123
```

**Result in ActivityWatch:**
```json
{
  "timestamp": "2025-11-04T09:00:00Z",
  "duration": 2700,
  "data": {
    "branch": "feature/JIRA-123-user-auth",
    "commits": [
      {
        "hash": "def5678...",
        "message": "JIRA-123: Add user authentication",
        "author": "Developer <dev@example.com>",
        "timestamp": "2025-11-04T09:15:00Z"
      },
      {
        "hash": "ghi9012...",
        "message": "JIRA-123: Add login form validation",
        "author": "Developer <dev@example.com>",
        "timestamp": "2025-11-04T09:30:00Z"
      }
    ],
    "eventCount": 45,
    "gitUser": "dev",
    "repoName": "my-app"
  }
}
```

## Why awagent?

### Problem
- Manual worklog entry is tedious and error-prone
- Developers forget what they worked on
- Time tracking is inaccurate
- Hard to attribute time to specific JIRA issues

### Solution
- **Automatic tracking** - No manual input required
- **Git-based sessions** - Natural workflow boundaries
- **Commit metadata** - Clear record of work done
- **Issue detection** - Extract JIRA keys from commits/branches
- **Accurate timing** - Precise session durations

### Benefits
✅ No manual time tracking  
✅ Accurate worklog data  
✅ Audit trail of all work  
✅ Easy Jira integration  
✅ Developer-friendly (runs in background)  
✅ Privacy-focused (only Git metadata, no code content)  

## Architecture Principles

### 1. Separation of Concerns
- **awagent**: Pure tracking (Git + window monitoring)
- **Future tool**: Processing and Jira sync
- Clean API boundary via ActivityWatch

### 2. Developer-Friendly
- Single binary, no dependencies
- Minimal configuration
- Auto-discovery (no per-project setup)
- Runs silently in background

### 3. Data-First
- Rich, structured event data
- Complete commit metadata
- Easy to query via API
- Flexible for future use cases

### 4. Production-Ready
- Static binary (no runtime dependencies)
- Cross-platform Linux support
- Graceful shutdown handling
- Error recovery and logging

## What awagent Does NOT Do

❌ **Does NOT** integrate directly with Jira  
❌ **Does NOT** post worklogs automatically  
❌ **Does NOT** require per-project configuration  
❌ **Does NOT** capture code content (only metadata)  
❌ **Does NOT** require aw-watcher-window (embedded detection)  

These are intentional design decisions to keep awagent simple and focused.

## Next Steps

- [Installation Guide](installation.md) - Deploy awagent
- [Configuration](configuration.md) - Configure for your setup
- [Usage](usage.md) - Daily usage patterns
- [Integration](integration.md) - Build Jira sync tool
