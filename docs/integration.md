# Integration Guide

## Overview

This guide shows how to build tools that process awagent data for Jira worklog automation, reporting, and analytics.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  awagent (Tracking)                                     │
│  • Monitors IDE activity                                │
│  • Captures Git commits                                 │
│  • Publishes to ActivityWatch                           │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│  ActivityWatch (Storage)                                │
│  • SQLite database                                      │
│  • REST API                                             │
│  • Event storage                                        │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│  Integration Tool (Processing)                          │
│  • Query events                                         │
│  • Extract issue keys                                   │
│  • Calculate time                                       │
│  • Post to Jira                                         │
└─────────────────────────────────────────────────────────┘
```

## Jira Worklog Sync Tool

### Complete Python Example

```python
#!/usr/bin/env python3
"""
Jira Worklog Sync - Processes awagent data and posts to Jira
"""

import requests
import re
from datetime import datetime, timedelta
from collections import defaultdict

# Configuration
AW_URL = "http://localhost:5600"
JIRA_URL = "https://your-company.atlassian.net"
JIRA_TOKEN = "your_api_token"
JIRA_EMAIL = "your.email@company.com"

# Issue key pattern
ISSUE_PATTERN = re.compile(r'[A-Z][A-Z0-9]+-\d+')

def get_buckets():
    """Get all awagent buckets"""
    response = requests.get(f"{AW_URL}/api/0/buckets/")
    response.raise_for_status()
    buckets = response.json()
    
    # Filter for awagent buckets (contain underscores)
    return [b for b in buckets.keys() if '_' in b]

def get_events(bucket_id, start_date, end_date):
    """Get events from bucket in date range"""
    params = {
        'start': start_date.isoformat(),
        'end': end_date.isoformat()
    }
    response = requests.get(
        f"{AW_URL}/api/0/buckets/{bucket_id}/events",
        params=params
    )
    response.raise_for_status()
    return response.json()

def extract_issues(events):
    """Extract issue keys and time from events"""
    issues = defaultdict(lambda: {
        'time': 0,
        'commits': [],
        'sessions': []
    })
    
    for event in events:
        commits = event.get('data', {}).get('commits', [])
        
        # Extract issue keys from commits
        event_issues = set()
        for commit in commits:
            keys = ISSUE_PATTERN.findall(commit['message'])
            event_issues.update(keys)
            for key in keys:
                issues[key]['commits'].append(commit)
        
        # Also check branch name
        branch = event.get('data', {}).get('branch', '')
        branch_issues = ISSUE_PATTERN.findall(branch)
        event_issues.update(branch_issues)
        
        # Distribute time across issues
        if event_issues:
            time_per_issue = event['duration'] / len(event_issues)
            for issue in event_issues:
                issues[issue]['time'] += time_per_issue
                issues[issue]['sessions'].append({
                    'timestamp': event['timestamp'],
                    'duration': time_per_issue,
                    'branch': branch
                })
    
    return dict(issues)

def round_to_jira_interval(seconds):
    """Round to nearest 15-minute interval (minimum 15 min)"""
    minutes = seconds / 60
    rounded_minutes = max(15, round(minutes / 15) * 15)
    return rounded_minutes * 60

def post_jira_worklog(issue_key, time_seconds, comment, started_time):
    """Post worklog to Jira"""
    url = f"{JIRA_URL}/rest/api/3/issue/{issue_key}/worklog"
    
    headers = {
        "Authorization": f"Basic {JIRA_EMAIL}:{JIRA_TOKEN}",
        "Content-Type": "application/json"
    }
    
    data = {
        "timeSpentSeconds": int(time_seconds),
        "comment": {
            "content": [
                {
                    "type": "paragraph",
                    "content": [
                        {
                            "type": "text",
                            "text": comment
                        }
                    ]
                }
            ],
            "type": "doc",
            "version": 1
        },
        "started": started_time
    }
    
    response = requests.post(url, json=data, headers=headers)
    response.raise_for_status()
    return response.json()

def main():
    # Process last 24 hours
    end_time = datetime.now()
    start_time = end_time - timedelta(days=1)
    
    print(f"Processing worklogs from {start_time} to {end_time}")
    print()
    
    # Get all buckets
    buckets = get_buckets()
    print(f"Found {len(buckets)} buckets")
    
    # Aggregate issues across all buckets
    all_issues = defaultdict(lambda: {'time': 0, 'commits': [], 'sessions': []})
    
    for bucket in buckets:
        print(f"Processing bucket: {bucket}")
        events = get_events(bucket, start_time, end_time)
        
        if not events:
            continue
        
        bucket_issues = extract_issues(events)
        
        # Merge into all_issues
        for issue, data in bucket_issues.items():
            all_issues[issue]['time'] += data['time']
            all_issues[issue]['commits'].extend(data['commits'])
            all_issues[issue]['sessions'].extend(data['sessions'])
    
    print()
    print(f"Found {len(all_issues)} issues")
    print()
    
    # Post to Jira
    for issue_key, data in all_issues.items():
        time_seconds = round_to_jira_interval(data['time'])
        minutes = int(time_seconds / 60)
        
        commit_count = len(set(c['hash'] for c in data['commits']))
        comment = f"Auto-logged from awagent: {commit_count} commits, " \
                  f"{len(data['sessions'])} sessions"
        
        print(f"{issue_key}: {minutes} minutes ({commit_count} commits)")
        
        try:
            result = post_jira_worklog(
                issue_key,
                time_seconds,
                comment,
                start_time.isoformat() + '+0000'
            )
            print(f"  ✓ Posted worklog ID: {result['id']}")
        except Exception as e:
            print(f"  ✗ Error: {e}")
    
    print()
    print("Done!")

if __name__ == "__main__":
    main()
```

### Usage

```bash
# Install dependencies
pip install requests

# Set environment variables
export JIRA_URL="https://your-company.atlassian.net"
export JIRA_EMAIL="your.email@company.com"
export JIRA_TOKEN="your_api_token"

# Run sync
python jira-sync.py
```

### Jenkins Pipeline Example

```groovy
pipeline {
    agent any
    
    triggers {
        // Run daily at 6 PM
        cron('0 18 * * *')
    }
    
    environment {
        AW_URL = 'http://activitywatch:5600'
        JIRA_URL = credentials('jira-url')
        JIRA_TOKEN = credentials('jira-api-token')
        JIRA_EMAIL = credentials('jira-email')
    }
    
    stages {
        stage('Sync Worklogs') {
            steps {
                sh '''
                    python3 jira-sync.py \
                        --aw-url "${AW_URL}" \
                        --jira-url "${JIRA_URL}" \
                        --jira-token "${JIRA_TOKEN}" \
                        --jira-email "${JIRA_EMAIL}" \
                        --days 1
                '''
            }
        }
    }
    
    post {
        success {
            echo 'Worklogs synced successfully'
        }
        failure {
            emailext (
                subject: "Jira Worklog Sync Failed",
                body: "Check Jenkins for details",
                to: "team@company.com"
            )
        }
    }
}
```

## Alternative: Go Integration Tool

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "regexp"
    "time"
)

type Event struct {
    Timestamp string                 `json:"timestamp"`
    Duration  float64                `json:"duration"`
    Data      map[string]interface{} `json:"data"`
}

type Commit struct {
    Hash      string `json:"hash"`
    Message   string `json:"message"`
    Author    string `json:"author"`
    Timestamp string `json:"timestamp"`
}

var issuePattern = regexp.MustCompile(`[A-Z][A-Z0-9]+-\d+`)

func getEvents(awURL, bucket string) ([]Event, error) {
    resp, err := http.Get(fmt.Sprintf("%s/api/0/buckets/%s/events", awURL, bucket))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var events []Event
    if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
        return nil, err
    }
    
    return events, nil
}

func extractIssues(events []Event) map[string]float64 {
    issues := make(map[string]float64)
    
    for _, event := range events {
        commits, ok := event.Data["commits"].([]interface{})
        if !ok {
            continue
        }
        
        eventIssues := make(map[string]bool)
        
        for _, c := range commits {
            commit := c.(map[string]interface{})
            message := commit["message"].(string)
            
            for _, issue := range issuePattern.FindAllString(message, -1) {
                eventIssues[issue] = true
            }
        }
        
        // Distribute time
        if len(eventIssues) > 0 {
            timePerIssue := event.Duration / float64(len(eventIssues))
            for issue := range eventIssues {
                issues[issue] += timePerIssue
            }
        }
    }
    
    return issues
}

func main() {
    awURL := "http://localhost:5600"
    bucket := "user_repo_branch"
    
    events, err := getEvents(awURL, bucket)
    if err != nil {
        panic(err)
    }
    
    issues := extractIssues(events)
    
    for issue, seconds := range issues {
        minutes := int(seconds / 60)
        fmt.Printf("%s: %d minutes\n", issue, minutes)
    }
}
```

## Reporting and Analytics

### Daily Summary Report

```python
def generate_daily_report(date):
    """Generate daily activity report"""
    buckets = get_buckets()
    
    report = {
        'date': date.isoformat(),
        'total_time': 0,
        'repositories': [],
        'issues': {}
    }
    
    for bucket in buckets:
        events = get_events(bucket, date, date + timedelta(days=1))
        
        repo_time = sum(e['duration'] for e in events)
        report['total_time'] += repo_time
        
        report['repositories'].append({
            'name': bucket,
            'time': repo_time,
            'sessions': len(events)
        })
        
        issues = extract_issues(events)
        for issue, data in issues.items():
            if issue not in report['issues']:
                report['issues'][issue] = {'time': 0, 'commits': 0}
            report['issues'][issue]['time'] += data['time']
            report['issues'][issue]['commits'] += len(data['commits'])
    
    return report
```

### Weekly Time Tracking

```python
def weekly_summary(start_date):
    """Generate weekly time tracking summary"""
    summary = defaultdict(lambda: defaultdict(float))
    
    for day in range(7):
        date = start_date + timedelta(days=day)
        report = generate_daily_report(date)
        
        for issue, data in report['issues'].items():
            summary[issue]['total'] += data['time']
            summary[issue][date.strftime('%A')] = data['time']
    
    return summary
```

### Export to CSV

```python
import csv

def export_to_csv(issues, filename):
    """Export issues and time to CSV"""
    with open(filename, 'w', newline='') as f:
        writer = csv.writer(f)
        writer.writerow(['Issue', 'Time (minutes)', 'Commits', 'Sessions'])
        
        for issue, data in sorted(issues.items()):
            writer.writerow([
                issue,
                int(data['time'] / 60),
                len(data['commits']),
                len(data['sessions'])
            ])
```

## API Query Examples

### Get all work for specific issue

```bash
# Query all buckets for issue PROJ-123
for bucket in $(curl -s http://localhost:5600/api/0/buckets/ | jq -r 'keys[]'); do
    echo "Bucket: $bucket"
    curl -s "http://localhost:5600/api/0/buckets/$bucket/events" \
        | jq --arg issue "PROJ-123" '
            .[] | select(.data.commits[]?.message | contains($issue))
        '
done
```

### Calculate total time per developer

```bash
curl -s http://localhost:5600/api/0/buckets/ | jq -r 'keys[]' | while read bucket; do
    curl -s "http://localhost:5600/api/0/buckets/$bucket/events" \
        | jq -r --arg bucket "$bucket" '
            .[] | "\(.data.gitUser):\(.duration):\($bucket)"
        '
done | awk -F: '{time[$1]+=$2} END {for(u in time) print u, int(time[u]/60), "minutes"}'
```

## Best Practices

### 1. Incremental Processing
- Process only new events since last sync
- Store last sync timestamp
- Avoid reprocessing old data

### 2. Error Handling
- Retry failed Jira API calls
- Log failures for manual review
- Don't fail entire sync on single error

### 3. Duplicate Prevention
- Check if worklog already exists
- Use comment field to store awagent event ID
- Skip if already processed

### 4. Time Attribution
- Define rules for multi-issue sessions
- Consider primary issue vs. all issues
- Document attribution logic

### 5. Validation
- Verify issue exists in Jira before posting
- Check user permissions
- Validate time is reasonable

## Next Steps

- [Data Format](data-format.md) - Understanding event structure
- [Commit Tracking](commit-tracking.md) - How commits are captured
- [Troubleshooting](troubleshooting.md) - Debug integration issues
