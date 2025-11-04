# Commit Tracking Feature

## Overview

The `awagent` now captures all Git commits made during a work session and includes them in the ActivityWatch event data. This provides complete visibility into what code changes were made during each work session.

## Test Results ✅

**Test execution:** Live commit tracking test
**Date:** 2025-11-04

### Test Scenario
1. Started `awagent` in test mode
2. Created 3 commits while agent was running:
   - `PROJ-789: Add first live test commit`
   - `PROJ-789: Add second live test commit`
   - `PROJ-999: Different issue commit`
3. Session flushed to ActivityWatch after 15 seconds

### Result
✅ **All 3 commits captured successfully!**

```json
{
  "id": 9,
  "timestamp": "2025-11-04T08:55:31.135000+00:00",
  "duration": 19.999358,
  "data": {
    "branch": "test/PROJ-789-live-commit-test",
    "commits": [
      {
        "hash": "76557c9eafc4b8097df76eb537ecca255c99117e",
        "message": "PROJ-789: Add first live test commit",
        "author": "lamdn8 <lamdn8@gmail.com>",
        "timestamp": "2025-11-04T15:55:34+07:00"
      },
      {
        "hash": "4a26305cd3b7e92a48fe9e0f44841985abf57d3c",
        "message": "PROJ-789: Add second live test commit",
        "author": "lamdn8 <lamdn8@gmail.com>",
        "timestamp": "2025-11-04T15:55:38+07:00"
      },
      {
        "hash": "b584ae38b040ab9c3e53cc81eff513a51f1f1db5",
        "message": "PROJ-999: Different issue commit",
        "author": "lamdn8 <lamdn8@gmail.com>",
        "timestamp": "2025-11-04T15:55:41+07:00"
      }
    ],
    "eventCount": 3,
    "gitEmail": "lamdn8@gmail.com",
    "gitUser": "lamdn8",
    "remote": "git@github.com:liamdn8/auto-worklog-agent.git",
    "repoName": "auto-worklog-agent",
    "repoPath": "/home/liamdn/auto-worklog-agent"
  }
}
```

## Data Flow

```
┌──────────────────────────────────────────────────────────────┐
│ 1. Session Start                                             │
│ - User opens VSCode with repo                                │
│ - awagent detects IDE activity                               │
│ - Creates new session                                        │
│ - Captures current HEAD commit hash: 6c6ef034                │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────────┐
│ 2. User Makes Commits                                        │
│ - Commit 1: PROJ-789: Add first test                         │
│ - Commit 2: PROJ-789: Add second test                        │
│ - Commit 3: PROJ-999: Fix bug                                │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────────┐
│ 3. Activity Detection (every 10s in test mode)               │
│ - awagent polls for activity                                 │
│ - Runs: git log <startCommit>..HEAD                          │
│ - Parses commits with: hash, message, author, timestamp      │
│ - Updates session.Commits array                              │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────────┐
│ 4. Session Flush (every 15s, or on idle/shutdown)            │
│ - Calculate session duration                                 │
│ - Build event payload with all session data                  │
│ - Include commits array in event.data                        │
│ - POST to ActivityWatch API                                  │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────────┐
│ 5. ActivityWatch Storage                                     │
│ - Stores event in SQLite database                            │
│ - data column contains full JSON (including commits)         │
│ - No schema validation - accepts any JSON                    │
└──────────────────────────────────────────────────────────────┘
```

## Event Data Structure

### Session WITHOUT Commits
```json
{
  "timestamp": "2025-11-04T10:00:00Z",
  "duration": 900.5,
  "data": {
    "gitUser": "lamdn8",
    "gitEmail": "lamdn8@gmail.com",
    "repoName": "auto-worklog-agent",
    "repoPath": "/home/liamdn/auto-worklog-agent",
    "branch": "main",
    "remote": "git@github.com:liamdn8/auto-worklog-agent.git",
    "eventCount": 42
  }
}
```

### Session WITH Commits
```json
{
  "timestamp": "2025-11-04T10:00:00Z",
  "duration": 900.5,
  "data": {
    "gitUser": "lamdn8",
    "gitEmail": "lamdn8@gmail.com",
    "repoName": "auto-worklog-agent",
    "repoPath": "/home/liamdn/auto-worklog-agent",
    "branch": "feature/PROJ-123-new-feature",
    "remote": "git@github.com:liamdn8/auto-worklog-agent.git",
    "eventCount": 42,
    "commits": [
      {
        "hash": "a1b2c3d4e5f6...",
        "message": "PROJ-123: Add new feature",
        "author": "lamdn8 <lamdn8@gmail.com>",
        "timestamp": "2025-11-04T10:15:00Z"
      },
      {
        "hash": "f6e5d4c3b2a1...",
        "message": "PROJ-123: Fix typo in feature",
        "author": "lamdn8 <lamdn8@gmail.com>",
        "timestamp": "2025-11-04T10:30:00Z"
      }
    ]
  }
}
```

## Implementation Details

### New Code Added

**1. internal/gitinfo/gitinfo.go**
- `Commit` struct: Represents a single commit
- `GetCommitsSince(repoPath, startHash)`: Retrieves commits from startHash to HEAD
- `GetCurrentCommitHash(repoPath)`: Gets current HEAD hash

**2. internal/session/session.go**
- Added `StartCommit` field: Stores commit hash when session starts
- Added `Commits` field: Array of commits made during session

**3. internal/agent/tracker.go**
- `recordEvent()`: Captures start commit hash on new session
- `recordEvent()`: Updates commits array on each activity detection
- `publishSession()`: Includes commits in event data when publishing

### Git Commands Used

```bash
# Get current HEAD commit hash
git rev-parse HEAD

# Get commits from startHash to HEAD (chronological order)
git log --reverse --pretty=format:%H|%an <%ae>|%aI|%s <startHash>..HEAD
```

### Format String Breakdown
- `%H` - Full commit hash
- `%an` - Author name
- `%ae` - Author email
- `%aI` - Author date (ISO 8601 format)
- `%s` - Commit subject (first line of message)

## Query Examples

### Get All Events with Commits
```bash
curl -s "http://localhost:9600/api/0/buckets/lamdn8_auto-worklog-agent_main/events" \
  | jq '.[] | select(.data.commits != null)'
```

### Extract Issue Keys from Commits
```bash
curl -s "http://localhost:9600/api/0/buckets/lamdn8_auto-worklog-agent_main/events" \
  | jq -r '.[] | .data.commits[]? | .message' \
  | grep -oE '[A-Z][A-Z0-9]+-[0-9]+' \
  | sort -u
```

### Count Commits by Issue
```bash
curl -s "http://localhost:9600/api/0/buckets/lamdn8_auto-worklog-agent_main/events" \
  | jq -r '.[] | .data.commits[]? | .message' \
  | grep -oE '[A-Z][A-Z0-9]+-[0-9]+' \
  | sort | uniq -c
```

### Get Time Spent per Issue
```bash
curl -s "http://localhost:9600/api/0/buckets/lamdn8_auto-worklog-agent_main/events" \
  | jq -r '.[] | select(.data.commits != null) | 
    .data.commits[] | .message as $msg | 
    ($msg | match("[A-Z][A-Z0-9]+-[0-9]+").string) as $issue | 
    "\($issue),\(.duration)"'
```

## Future Processing Tool (Example)

The commit data is ready for a future Jira worklog sync tool:

```python
import requests
import re
from datetime import datetime, timedelta

# 1. Query ActivityWatch for events
events = requests.get(
    'http://localhost:9600/api/0/buckets/lamdn8_auto-worklog-agent_main/events',
    params={'start': '2025-11-04', 'end': '2025-11-05'}
).json()

# 2. Extract issue keys and calculate time
issue_time = {}
for event in events:
    commits = event.get('data', {}).get('commits', [])
    for commit in commits:
        # Extract issue keys from commit message
        issues = re.findall(r'[A-Z][A-Z0-9]+-\d+', commit['message'])
        for issue in issues:
            if issue not in issue_time:
                issue_time[issue] = 0
            # Add duration (in seconds)
            issue_time[issue] += event['duration']

# 3. Round to 15-minute intervals
for issue, seconds in issue_time.items():
    minutes = seconds / 60
    rounded_minutes = round(minutes / 15) * 15  # Round to nearest 15min
    print(f"{issue}: {rounded_minutes} minutes")

# 4. Post to Jira (example)
for issue, seconds in issue_time.items():
    minutes = seconds / 60
    rounded_minutes = max(15, round(minutes / 15) * 15)  # Min 15min
    
    requests.post(
        f'https://your-jira.atlassian.net/rest/api/3/issue/{issue}/worklog',
        headers={'Authorization': 'Bearer YOUR_TOKEN'},
        json={
            'timeSpentSeconds': rounded_minutes * 60,
            'comment': 'Auto-logged from awagent',
            'started': datetime.utcnow().isoformat() + '+0000'
        }
    )
```

## Benefits

✅ **Complete Work History**: See all commits made during each work session
✅ **Multi-Issue Support**: Extract multiple issue keys from different commits
✅ **Time Attribution**: Understand how time was spent across different issues
✅ **Audit Trail**: Full commit metadata (hash, message, author, timestamp)
✅ **No Jira Coupling**: awagent stays simple, Jira sync is separate tool
✅ **Flexible Processing**: Any future tool can process the commit data

## Testing

### Run Live Test
```bash
cd /home/liamdn/auto-worklog-agent
./test-commits-live.sh
```

This will:
1. Create a test branch
2. Start awagent in test mode
3. Create 3 commits while running
4. Show the captured commit data
5. Clean up automatically

### Manual Test
```bash
# 1. Start awagent
./awagent --aw-url http://localhost:9600 --test -v

# 2. In another terminal, make commits
echo "test" > test.txt
git add test.txt
git commit -m "PROJ-123: Test commit"

# 3. Wait 15 seconds for flush

# 4. Query ActivityWatch
curl -s "http://localhost:9600/api/0/buckets/YOUR_BUCKET/events?limit=1" | jq '.[-1].data.commits'
```

## View in ActivityWatch UI

1. Open: http://localhost:9600
2. Navigate to "Raw Data" view
3. Select your bucket (e.g., `lamdn8_auto-worklog-agent_main`)
4. Inspect events - commits will be in the `data.commits` field

## Performance Impact

- **Minimal overhead**: Git log is fast (< 10ms for most repos)
- **Only on activity**: Commits fetched during activity detection, not continuously
- **Cached by Git**: Git log results are cached by Git itself
- **No network calls**: All local Git operations

## Conclusion

The commit tracking feature is now fully functional and tested. All Git commits made during a work session are automatically captured and stored in ActivityWatch, ready for future processing by Jira worklog sync tools or other analytics.
