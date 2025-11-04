# Commit Tracking

## Overview

awagent automatically captures all Git commits made during work sessions, providing complete visibility into code changes and enabling accurate worklog attribution.

## How It Works

### 1. Session Start
When awagent detects IDE activity for a repository:
```
1. Create new session
2. Get current HEAD commit hash
3. Store as session.StartCommit
```

### 2. Activity Detection
Every second (default), awagent checks for activity:
```
1. Poll active window
2. Match window to repository
3. Run: git log <startCommit>..HEAD
4. Parse commits (hash, message, author, timestamp)
5. Update session.Commits array
```

### 3. Session End
When session ends (idle timeout or shutdown):
```
1. Build event with all session data
2. Include commits array in event.data
3. POST to ActivityWatch API
```

## Git Commands Used

### Get Current Commit Hash
```bash
git -C /path/to/repo rev-parse HEAD
```

Output: `a1b2c3d4e5f6789012345678901234567890abcd`

### Get Commits Since Session Start
```bash
git -C /path/to/repo log \
  --reverse \
  --pretty=format:'%H|%an <%ae>|%aI|%s' \
  <startHash>..HEAD
```

Output:
```
a1b2c3d4|John Doe <john@example.com>|2025-11-04T10:15:00+07:00|PROJ-123: Add feature
bcdef123|John Doe <john@example.com>|2025-11-04T10:30:00+07:00|PROJ-123: Fix bug
```

### Format String Breakdown
- `%H` - Full commit hash (40 chars)
- `%an` - Author name
- `%ae` - Author email
- `%aI` - Author date (ISO 8601)
- `%s` - Subject (first line of message)
- `--reverse` - Chronological order (oldest first)

## Data Flow Example

### Timeline
```
09:00 - Open VSCode with repo "my-app" on branch "feature/PROJ-123"
        Session starts, StartCommit = abc1234

09:15 - git commit -m "PROJ-123: Implement login"
        Commit def5678 created

09:20 - Activity detected (window poll)
        git log abc1234..HEAD
        → Finds commit def5678
        → Updates session.Commits = [def5678]

09:30 - git commit -m "PROJ-123: Add validation"
        Commit ghi9012 created

09:35 - Activity detected
        git log abc1234..HEAD
        → Finds commits def5678 + ghi9012
        → Updates session.Commits = [def5678, ghi9012]

10:05 - 30 minutes idle, session ends
        Event published with 2 commits
```

### Resulting Event
```json
{
  "timestamp": "2025-11-04T09:00:00Z",
  "duration": 3900,
  "data": {
    "branch": "feature/PROJ-123",
    "commits": [
      {
        "hash": "def5678...",
        "message": "PROJ-123: Implement login",
        "author": "Developer <dev@example.com>",
        "timestamp": "2025-11-04T09:15:00Z"
      },
      {
        "hash": "ghi9012...",
        "message": "PROJ-123: Add validation",
        "author": "Developer <dev@example.com>",
        "timestamp": "2025-11-04T09:30:00Z"
      }
    ]
  }
}
```

## Use Cases

### 1. Single Issue Workflow
**Branch:** `feature/PROJ-123-user-auth`

**Commits:**
- `PROJ-123: Add login form`
- `PROJ-123: Add authentication API`
- `PROJ-123: Add session management`

**Result:** All time attributed to PROJ-123

### 2. Multi-Issue Workflow
**Branch:** `develop`

**Commits:**
- `PROJ-100: Fix dashboard bug`
- `PROJ-101: Update dependencies`
- `PROJ-100: Add unit tests`

**Result:** Time split between PROJ-100 and PROJ-101

### 3. Branch Name + Commit Message
**Branch:** `feature/PROJ-200-api-refactor`

**Commits:**
- `PROJ-200: Refactor user service`
- `PROJ-200: Update API docs`
- `PROJ-999: Fix unrelated bug`

**Result:** Issue keys from both branch (PROJ-200) and commits (PROJ-999)

### 4. No Commits
**Branch:** `main`

**Commits:** (none)

**Result:** Session tracked, but no commits array in event

## Benefits

### For Developers
✅ **Automatic** - No manual tracking  
✅ **Accurate** - Exact commit metadata  
✅ **Non-intrusive** - Runs in background  
✅ **Audit trail** - Complete work history  

### For Managers
✅ **Visibility** - See what was done  
✅ **Accountability** - Commit-level tracking  
✅ **Reporting** - Easy to generate reports  
✅ **Time attribution** - Link time to work  

### For Automation
✅ **Structured data** - Easy to parse  
✅ **Issue detection** - Extract JIRA keys  
✅ **Flexible** - Support multiple workflows  
✅ **API access** - Query from any tool  

## Edge Cases

### Multiple Commits in Quick Succession
If commits happen faster than poll interval:
```
09:00:00 - Commit A
09:00:05 - Commit B
09:00:10 - Commit C
09:00:15 - Activity poll
```

**Result:** All 3 commits captured on next poll

### Amended Commits
If you amend a commit:
```bash
git commit -m "Initial message"
git commit --amend -m "Updated message"
```

**Result:** Only the amended version is captured (Git history is linear)

### Rebased Commits
If you rebase:
```bash
git rebase main
```

**Result:** New commit hashes after rebase, old commits replaced

### Cherry-Picked Commits
If you cherry-pick:
```bash
git cherry-pick abc123
```

**Result:** New commit hash, treated as separate commit

### No Commits
If no commits during session:

**Result:** Event has no `commits` field (not an empty array)

## Performance

### Impact on Git Operations
- **Minimal**: `git log` is very fast (<10ms for most repos)
- **Cached**: Git caches log results
- **Local**: No network calls
- **Incremental**: Only new commits since session start

### CPU Usage
- Poll interval: 1 second (default)
- Git log per poll: Only when activity detected
- Typical overhead: <1% CPU

### Memory Usage
- Commits stored in memory until session flush
- Typical session: <100 commits
- Memory per commit: ~500 bytes
- Total: <50 KB per session

## Testing

### Manual Test
```bash
# 1. Start awagent
./awagent --test -v

# 2. In another terminal, create commits
cd ~/projects/test-repo
echo "change" >> file.txt
git add file.txt
git commit -m "TEST-123: Test commit"

# 3. Wait for flush (15 seconds)

# 4. Query ActivityWatch
USER=$(git config user.name | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9_-]/-/g')
BUCKET="${USER}_test-repo_main"
curl -s "http://localhost:5600/api/0/buckets/${BUCKET}/events?limit=1" \
  | jq '.[0].data.commits'
```

### Automated Test Script
See: `test-commits-live.sh`

```bash
./test-commits-live.sh
```

This script:
1. Creates test branch
2. Starts awagent
3. Creates 3 commits
4. Verifies commits appear in ActivityWatch
5. Cleans up

## Troubleshooting

### Commits Not Appearing

**Check 1:** Verify git is configured
```bash
git config user.name
git config user.email
```

**Check 2:** Check awagent logs
```bash
./awagent -v
# Look for: "Session started... commit=abc1234"
# Look for: "Publishing session with N commits"
```

**Check 3:** Verify commits exist
```bash
git log --oneline -5
```

**Check 4:** Check session timing
- Commits must happen AFTER session starts
- Commits created BEFORE session starts are not captured

### Wrong Commits Captured

**Issue:** Seeing commits from before session

**Cause:** StartCommit hash incorrect

**Fix:** Session start should capture current HEAD correctly

### Duplicate Commits

**Issue:** Same commit appears multiple times

**Cause:** Multiple sessions on same branch

**Solution:** This is expected - each session captures its commits independently

## Next Steps

- [Data Format](data-format.md) - Event structure details
- [Integration](integration.md) - Process commit data for Jira
- [Usage](usage.md) - Best practices for commit tracking
