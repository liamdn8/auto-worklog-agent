# Performance Testing Guide

## Overview

This guide provides tools and procedures to validate that `awagent` runs with minimal resource usage.

## Performance Requirements

**Target Specifications:**
- **CPU Usage**: < 5% average, < 10% peak
- **Memory Usage**: < 30 MB average, < 50 MB peak
- **File Descriptors**: < 50
- **Threads**: < 20
- **Network**: Minimal (only heartbeats to ActivityWatch)

## Quick Test (30 seconds)

For a quick validation:

```bash
./test-performance-quick.sh
```

**What it does:**
- Runs awagent for 30 seconds
- Samples CPU and memory every 5 seconds
- Reports pass/fail based on thresholds

**Expected output:**
```
=== Quick Performance Test (30s) ===

Monitoring awagent (PID: 12345) for 30 seconds...

Time | CPU% | Memory MB
-----+------+----------
  5s |  0.3 |     12.4
 10s |  0.5 |     14.2
 15s |  0.2 |     15.1
 20s |  0.4 |     15.8
 25s |  0.3 |     16.2
 30s |  0.2 |     16.5

Final Stats:
  CPU: 0.2%
  Memory: 16.5 MB

✓ PASSED - Low resource usage confirmed
```

## Comprehensive Test (5 minutes)

For detailed performance analysis:

```bash
./test-performance.sh
```

**What it does:**
- Runs awagent for 5 minutes
- Samples every 5 seconds (60 samples total)
- Monitors CPU, memory, threads, file descriptors
- Generates detailed reports and CSV data
- Counts threshold violations

**Output files:**
- `performance_results.txt` - Summary results
- `performance_data.csv` - Time-series data
- `awagent_test.log` - Agent logs

**Expected output:**
```
========================================
  awagent Performance Test
========================================

Test Parameters:
  Duration: 300s (5 minutes)
  Sample Interval: 5s
  Max Memory Threshold: 30 MB
  Max CPU Threshold: 5%

Starting awagent...
awagent started (PID: 12345)

Monitoring performance...
Time | CPU% | Memory MB | RSS KB | Threads | FDs
-----+------+-----------+--------+---------+-----
  5s |  0.30 |     12.45 |  12752 |      10 |  15
 10s |  0.25 |     14.23 |  14571 |      10 |  15
...

========================================
  Performance Test Results
========================================

Samples Collected: 60
Test Duration: 300s (5 minutes)

CPU Usage:
  Average: 0.32%
  Maximum: 1.20%
  Threshold: 5%
  Violations: 0

Memory Usage:
  Average: 16.45 MB
  Maximum: 18.32 MB
  Threshold: 30 MB
  Violations: 0

========================================
✓ PASSED - Low resource usage confirmed
========================================
```

## Custom Duration Test

Run with custom duration:

```bash
# Edit test-performance.sh and change:
TEST_DURATION=600  # 10 minutes
TEST_DURATION=1800 # 30 minutes
TEST_DURATION=3600 # 1 hour
```

## Analyzing Results

### View Performance Data

```bash
# View summary
cat performance_results.txt

# View time-series data
cat performance_data.csv

# Plot with gnuplot (if available)
gnuplot << EOF
set datafile separator ','
set xlabel 'Time (s)'
set ylabel 'CPU %'
set y2label 'Memory (MB)'
set ytics nomirror
set y2tics
plot 'performance_data.csv' using 1:2 with lines title 'CPU%' axes x1y1, \
     'performance_data.csv' using 1:3 with lines title 'Memory MB' axes x1y2
pause -1
EOF
```

### Import to Spreadsheet

The `performance_data.csv` file can be imported into Excel, Google Sheets, or LibreOffice Calc for analysis.

**Columns:**
- Timestamp (Unix epoch)
- CPU% (percentage)
- Memory_MB (megabytes)
- Memory_RSS_KB (kilobytes)
- Threads (count)
- FDs (file descriptor count)

## Stress Testing

### High Repository Count

Test with many repositories:

```bash
# Create test repos
for i in {1..50}; do
    mkdir -p /tmp/test-repos/repo$i
    cd /tmp/test-repos/repo$i
    git init
    git config user.email "test@test.com"
    git config user.name "Test"
    touch README.md
    git add .
    git commit -m "Initial"
done

# Update config to scan /tmp/test-repos
{
  "roots": ["/tmp/test-repos"],
  "maxDepth": 2,
  "rescanIntervalMinutes": 1,
  "pulseTime": 30
}

# Run performance test
./test-performance.sh
```

### Rapid Window Switching

Test window detection overhead:

```bash
# Run test while rapidly switching windows
# Manually switch between applications every 2-3 seconds
# Or use xdotool to automate:

./awagent &
AGENT_PID=$!

# Switch windows rapidly for 60 seconds
for i in {1..30}; do
    xdotool key alt+Tab
    sleep 2
done

# Check if CPU usage spiked
ps -p $AGENT_PID -o %cpu,%mem

kill $AGENT_PID
```

## Benchmarking Different Scenarios

### Scenario 1: Idle Desktop

```bash
# Leave computer idle, no window switching
./test-performance.sh
# Expected: Very low CPU (<0.5%), stable memory
```

### Scenario 2: Active Development

```bash
# Actively code/switch windows during test
./test-performance.sh
# Expected: Slightly higher CPU (<2%), stable memory
```

### Scenario 3: Many Repositories

```bash
# Configure 20+ repos, frequent rescans
# Set rescanIntervalMinutes: 1
./test-performance.sh
# Expected: Higher CPU during scans, but still <5% average
```

## Troubleshooting High Resource Usage

### High CPU Usage

**Possible causes:**
1. Window detection tool (xdotool/xprop) is slow
2. Too frequent window polling (default: 1 second)
3. Network issues with ActivityWatch server

**Solutions:**
```bash
# Try different window detection tool
sudo apt-get install xdotool  # Usually faster than xprop

# Check network latency to ActivityWatch
curl -w "@-" -o /dev/null -s http://localhost:5600/api/0/info << 'EOF'
    time_total: %{time_total}s\n
EOF

# Review logs for errors
./awagent --verbose | grep -i error
```

### High Memory Usage

**Possible causes:**
1. Too many repositories discovered
2. Memory leak (report as bug!)
3. Large event buffers

**Solutions:**
```bash
# Limit repository depth
{
  "maxDepth": 3,  # Reduce from 5
  ...
}

# Monitor for memory leaks
watch -n 5 'ps -p $(pgrep awagent) -o rss,vsz,%mem'

# If memory grows continuously over hours, it's a leak
```

### High Thread/FD Count

**Possible causes:**
1. Goroutine leak (bug)
2. File descriptor leak (bug)

**Solutions:**
```bash
# Monitor thread count
watch -n 5 'ps -p $(pgrep awagent) -o nlwp'

# Monitor file descriptors
watch -n 5 'ls -1 /proc/$(pgrep awagent)/fd | wc -l'

# If either grows continuously, report as bug
```

## Continuous Monitoring

### Using top

```bash
top -p $(pgrep awagent)
```

### Using htop

```bash
htop -p $(pgrep awagent)
```

### Using systemd (if installed as service)

```bash
systemctl --user status awagent
```

### Long-term Monitoring Script

```bash
#!/bin/bash
# Monitor awagent continuously
LOGFILE="awagent_monitor_$(date +%Y%m%d_%H%M%S).log"

echo "Timestamp,CPU%,Memory_MB,Threads,FDs" > "$LOGFILE"

while true; do
    PID=$(pgrep awagent)
    if [ -n "$PID" ]; then
        CPU=$(ps -p $PID -o %cpu --no-headers | tr -d ' ')
        RSS=$(awk '{print $24}' /proc/$PID/stat 2>/dev/null)
        MEM=$(awk "BEGIN {printf \"%.1f\", $RSS * 4096 / 1024 / 1024}" 2>/dev/null)
        THREADS=$(ps -p $PID -o nlwp --no-headers | tr -d ' ')
        FDS=$(ls -1 /proc/$PID/fd 2>/dev/null | wc -l)
        
        echo "$(date +%s),$CPU,$MEM,$THREADS,$FDS" >> "$LOGFILE"
    fi
    sleep 60  # Sample every minute
done
```

## Performance Comparison

Compare against aw-watcher-window:

```bash
# Test awagent
./test-performance.sh
mv performance_results.txt awagent_results.txt

# Test aw-watcher-window
./activitywatch/aw-watcher-window/aw-watcher-window &
PID=$!
sleep 300
ps -p $PID -o %cpu,%mem,rss,vsz
kill $PID

# Compare results
```

## Expected Performance Profile

**Startup (first 10 seconds):**
- CPU: 1-3% (repository scanning)
- Memory: 8-12 MB

**Idle (no window changes):**
- CPU: <0.5%
- Memory: 15-20 MB (stable)

**Active (frequent window switching):**
- CPU: 1-2%
- Memory: 15-20 MB (stable)

**Repository rescan:**
- CPU: 2-5% (brief spike)
- Memory: 15-25 MB (temporary increase)

## CI/CD Integration

Add to your CI pipeline:

```yaml
# .github/workflows/performance.yml
name: Performance Test

on: [push, pull_request]

jobs:
  performance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Build
        run: go build -o awagent ./cmd/awagent
      - name: Performance Test
        run: |
          sudo apt-get install -y xdotool
          ./test-performance-quick.sh
```

## Reporting Performance Issues

If you find performance issues, report with:

1. **System info:** `uname -a`
2. **Test results:** Attach `performance_results.txt` and `performance_data.csv`
3. **Configuration:** Your `config.json`
4. **Repository count:** How many repos were scanned
5. **Logs:** `awagent_test.log` (if relevant)

## Summary

- **Quick test**: `./test-performance-quick.sh` (30 seconds)
- **Full test**: `./test-performance.sh` (5 minutes)
- **Target**: <5% CPU, <30 MB memory
- **Monitor**: Use provided scripts or system tools
- **Report**: Any sustained high usage as a bug

The awagent is designed to be a lightweight background process that doesn't interfere with your development work!
