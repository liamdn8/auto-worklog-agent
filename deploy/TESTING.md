# Testing the Agent

## Embedded Window Watcher

The agent now has **built-in window watching** - no need for aw-watcher-window! It directly monitors your active window and detects IDE activity.

## Quick Test (Without Window Detection Tools)

To test without installing xdotool/window detection tools, use test mode:

```bash
# Start ActivityWatch server
docker compose up -d

# Run agent in test mode (simulates activity every 10 seconds)
./bin/awagent --test
```

Test mode will:
- Auto-discover Git repositories in your home directory
- Simulate IDE activity for each repository every 10 seconds
- Create work sessions and flush them to ActivityWatch
- **Use new bucket naming**: `user_repo_branch` (e.g., `lamdn8_auto-worklog-agent_main`)
- Log all activity to console

Check the results:
```bash
# View buckets created (notice user_repo_branch format)
curl http://localhost:5600/api/0/buckets/ | jq

# View events for a specific repo
curl 'http://localhost:5600/api/0/buckets/lamdn8_auto-worklog-agent_main/events?limit=10' | jq

# Open ActivityWatch web UI
open http://localhost:5600
```

## Running with Real Window Detection

For production use with automatic IDE detection:

1. **Install window detection tools (Linux):**
   ```bash
   sudo apt-get install xdotool  # Ubuntu/Debian
   sudo yum install xdotool      # RHEL/CentOS  
   ```

2. **Run the agent normally:**
   ```bash
   ./bin/awagent
   ```

3. **The agent will:**
   - Poll active window every 1 second
   - Match window titles to discovered Git repositories  
   - Track sessions per repository and branch
   - Create buckets as `user_repo_branch` (e.g., `lamdn8_myproject_feature-branch`)
   - Flush sessions after 30 minutes of inactivity

## Bucket Naming Format

Buckets are now named as: **`user.repository.branch`**

Examples:
- `lamdn8_auto-worklog-agent_main`
- `lamdn8_mc-tool_web-tool`
- `johndoe.myproject.feature-login`

This makes it easy for Jenkins to query sessions by user, project, or branch!

## Troubleshooting

### "could not get active window"

This means window detection tools are not installed. Either:

1. **Install xdotool (Linux):**
   ```bash
   sudo apt-get install xdotool
   ```

2. **Or run in test mode:**
   ```bash
   ./bin/awagent --test
   ```

### No repositories found

The agent scans `$HOME` by default with depth 5. To customize:

```json
{
  "git": {
    "roots": ["$HOME/projects", "$HOME/work"],
    "maxDepth": 0
  }
}
```

Save as `~/.config/awagent/config.json`

### Check what repositories were discovered

```bash
# Run with verbose logging
./bin/awagent -v --test
```

Look for: `Repository scan complete: indexed N repositories`

## Understanding Bucket Names

The new bucket naming format `user_repo_branch` makes it easy to:
- Query all work by a specific user: `curl http://localhost:5600/api/0/buckets/ | jq 'to_entries | map(select(.key | startswith("lamdn8")))'`
- Track work on specific branches: Filter buckets ending with `.main` or `.feature-*`
- Aggregate work per project: Filter buckets containing specific repo names
