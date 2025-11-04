# Data Format

## Event Structure

awagent publishes events to ActivityWatch with the following structure:

## Complete Event Example

```json
{
  "id": 42,
  "timestamp": "2025-11-04T10:00:00.000000+00:00",
  "duration": 1800.5,
  "data": {
    "gitUser": "john",
    "gitEmail": "john@example.com",
    "repoName": "my-project",
    "repoPath": "/home/john/projects/my-project",
    "branch": "feature/PROJ-123-new-feature",
    "remote": "git@github.com:company/my-project.git",
    "eventCount": 180,
    "commits": [
      {
        "hash": "a1b2c3d4e5f6789012345678901234567890abcd",
        "message": "PROJ-123: Implement user authentication",
        "author": "John Doe <john@example.com>",
        "timestamp": "2025-11-04T10:15:00+00:00"
      },
      {
        "hash": "bcdef1234567890abcdef1234567890abcdef123",
        "message": "PROJ-123: Add login form validation",
        "author": "John Doe <john@example.com>",
        "timestamp": "2025-11-04T10:30:00+00:00"
      }
    ]
  }
}
```

## Field Descriptions

### Top-Level Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Event ID (assigned by ActivityWatch) |
| `timestamp` | string | ISO 8601 timestamp when session started |
| `duration` | float | Session duration in seconds |
| `data` | object | Event metadata (see below) |

### Data Object Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `gitUser` | string | Yes | Git user.name from repository config |
| `gitEmail` | string | Yes | Git user.email from repository config |
| `repoName` | string | Yes | Repository directory name |
| `repoPath` | string | Yes | Absolute path to repository |
| `branch` | string | Yes | Current Git branch |
| `remote` | string | No | Git remote URL (if configured) |
| `eventCount` | integer | Yes | Number of activity detections in session |
| `commits` | array | No | Array of commits made during session |

### Commit Object Fields

| Field | Type | Description |
|-------|------|-------------|
| `hash` | string | Full 40-character commit SHA-1 hash |
| `message` | string | Commit message (first line) |
| `author` | string | Commit author in format "Name <email>" |
| `timestamp` | string | ISO 8601 timestamp of commit |

## Bucket Naming

Buckets are named using the pattern: `{gitUser}_{repoName}_{branch}`

### Examples

| Git User | Repository | Branch | Bucket ID |
|----------|-----------|--------|-----------|
| john | my-project | main | `john_my-project_main` |
| alice | web-app | feature/PROJ-123 | `alice_web-app_feature-proj-123` |
| bob.smith | api-server | hotfix/bug-456 | `bob-smith_api-server_hotfix-bug-456` |

### Sanitization Rules

Characters are sanitized according to these rules:
- **Allowed**: a-z, A-Z, 0-9, underscore (_), hyphen (-)
- **Replaced with hyphen**: All other characters
- **Trimmed**: Leading/trailing hyphens
- **Lowercase**: All converted to lowercase

Examples:
- `Bob Smith` → `bob-smith`
- `feature/PROJ-123` → `feature-proj-123`
- `my.project` → `my-project`

## Data Scenarios

### Scenario 1: New Session (No Commits)

When a session starts and no commits are made:

```json
{
  "timestamp": "2025-11-04T09:00:00Z",
  "duration": 600.0,
  "data": {
    "gitUser": "alice",
    "gitEmail": "alice@company.com",
    "repoName": "frontend",
    "repoPath": "/home/alice/projects/frontend",
    "branch": "develop",
    "remote": "git@github.com:company/frontend.git",
    "eventCount": 60
  }
}
```

Note: `commits` field is absent (not an empty array).

### Scenario 2: Session with Single Commit

```json
{
  "timestamp": "2025-11-04T10:00:00Z",
  "duration": 900.0,
  "data": {
    "gitUser": "bob",
    "gitEmail": "bob@company.com",
    "repoName": "backend",
    "repoPath": "/home/bob/work/backend",
    "branch": "feature/ISSUE-789",
    "remote": "https://github.com/company/backend.git",
    "eventCount": 90,
    "commits": [
      {
        "hash": "abc123...",
        "message": "ISSUE-789: Add API endpoint",
        "author": "Bob <bob@company.com>",
        "timestamp": "2025-11-04T10:10:00Z"
      }
    ]
  }
}
```

### Scenario 3: Session with Multiple Commits, Multiple Issues

```json
{
  "timestamp": "2025-11-04T14:00:00Z",
  "duration": 3600.0,
  "data": {
    "gitUser": "charlie",
    "gitEmail": "charlie@company.com",
    "repoName": "monorepo",
    "repoPath": "/home/charlie/monorepo",
    "branch": "develop",
    "remote": "git@gitlab.com:company/monorepo.git",
    "eventCount": 360,
    "commits": [
      {
        "hash": "111aaa...",
        "message": "TASK-100: Update dependencies",
        "author": "Charlie <charlie@company.com>",
        "timestamp": "2025-11-04T14:15:00Z"
      },
      {
        "hash": "222bbb...",
        "message": "TASK-101: Fix security vulnerability",
        "author": "Charlie <charlie@company.com>",
        "timestamp": "2025-11-04T14:45:00Z"
      },
      {
        "hash": "333ccc...",
        "message": "TASK-100: Update changelog",
        "author": "Charlie <charlie@company.com>",
        "timestamp": "2025-11-04T15:00:00Z"
      }
    ]
  }
}
```

Issues detected: TASK-100 (2 commits), TASK-101 (1 commit)

## Querying Events

### Get All Buckets

```bash
curl -s http://localhost:5600/api/0/buckets/ | jq 'keys'
```

Response:
```json
[
  "alice_frontend_develop",
  "bob_backend_feature-issue-789",
  "charlie_monorepo_develop"
]
```

### Get Events from Bucket

```bash
curl -s "http://localhost:5600/api/0/buckets/alice_frontend_develop/events" | jq '.'
```

### Get Events in Date Range

```bash
curl -s "http://localhost:5600/api/0/buckets/alice_frontend_develop/events?start=2025-11-04T00:00:00&end=2025-11-05T00:00:00" | jq '.'
```

### Get Latest Event

```bash
curl -s "http://localhost:5600/api/0/buckets/alice_frontend_develop/events?limit=1" | jq '.[0]'
```

## Processing Examples

### Extract All Issue Keys

```bash
curl -s "http://localhost:5600/api/0/buckets/BUCKET_ID/events" \
  | jq -r '.[].data.commits[]?.message' \
  | grep -oE '[A-Z][A-Z0-9]+-[0-9]+' \
  | sort -u
```

### Count Commits per Issue

```bash
curl -s "http://localhost:5600/api/0/buckets/BUCKET_ID/events" \
  | jq -r '.[].data.commits[]?.message' \
  | grep -oE '[A-Z][A-Z0-9]+-[0-9]+' \
  | sort | uniq -c
```

### Calculate Time per Day

```bash
curl -s "http://localhost:5600/api/0/buckets/BUCKET_ID/events" \
  | jq -r '.[] | "\(.timestamp | split("T")[0]): \(.duration)s"'
```

### Get Commits Timeline

```bash
curl -s "http://localhost:5600/api/0/buckets/BUCKET_ID/events" \
  | jq -r '.[].data.commits[]? | "\(.timestamp) \(.hash[0:8]) \(.message)"' \
  | sort
```

## Data Processing for Jira

### Step 1: Query Events

```python
import requests
from datetime import datetime, timedelta

# Query last 24 hours
end = datetime.now()
start = end - timedelta(days=1)

events = requests.get(
    f'http://localhost:5600/api/0/buckets/{bucket_id}/events',
    params={
        'start': start.isoformat(),
        'end': end.isoformat()
    }
).json()
```

### Step 2: Extract Issue Keys

```python
import re

issue_pattern = re.compile(r'[A-Z][A-Z0-9]+-\d+')

issues = {}
for event in events:
    commits = event.get('data', {}).get('commits', [])
    for commit in commits:
        issue_keys = issue_pattern.findall(commit['message'])
        for issue in issue_keys:
            if issue not in issues:
                issues[issue] = {
                    'time': 0,
                    'commits': []
                }
            issues[issue]['time'] += event['duration']
            issues[issue]['commits'].append(commit)
```

### Step 3: Round Time to Jira Format

```python
def round_to_jira_interval(seconds):
    """Round to nearest 15-minute interval, minimum 15 min"""
    minutes = seconds / 60
    rounded = max(15, round(minutes / 15) * 15)
    return rounded * 60  # Return as seconds
```

### Step 4: Post to Jira

```python
for issue_key, data in issues.items():
    time_seconds = round_to_jira_interval(data['time'])
    
    response = requests.post(
        f'https://your-jira.atlassian.net/rest/api/3/issue/{issue_key}/worklog',
        headers={'Authorization': 'Bearer YOUR_TOKEN'},
        json={
            'timeSpentSeconds': time_seconds,
            'comment': {
                'content': [
                    {
                        'type': 'paragraph',
                        'content': [
                            {
                                'type': 'text',
                                'text': f"Auto-logged from awagent: {len(data['commits'])} commits"
                            }
                        ]
                    }
                ]
            },
            'started': datetime.utcnow().isoformat() + '+0000'
        }
    )
```

## Schema Validation

### JSON Schema (for reference)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["timestamp", "duration", "data"],
  "properties": {
    "id": {"type": "integer"},
    "timestamp": {"type": "string", "format": "date-time"},
    "duration": {"type": "number"},
    "data": {
      "type": "object",
      "required": ["gitUser", "gitEmail", "repoName", "repoPath", "branch", "eventCount"],
      "properties": {
        "gitUser": {"type": "string"},
        "gitEmail": {"type": "string"},
        "repoName": {"type": "string"},
        "repoPath": {"type": "string"},
        "branch": {"type": "string"},
        "remote": {"type": "string"},
        "eventCount": {"type": "integer"},
        "commits": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["hash", "message", "author", "timestamp"],
            "properties": {
              "hash": {"type": "string", "minLength": 40, "maxLength": 40},
              "message": {"type": "string"},
              "author": {"type": "string"},
              "timestamp": {"type": "string", "format": "date-time"}
            }
          }
        }
      }
    }
  }
}
```

## Next Steps

- [Integration Guide](integration.md) - Build Jira sync tool
- [Usage Guide](usage.md) - Generate real data
- [Troubleshooting](troubleshooting.md) - Debug data issues
