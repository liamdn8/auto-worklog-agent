# Testing Overview

This document provides an overview of all available tests for the awagent project.

## Available Tests

### 1. Quick Performance Test (30 seconds)

**File:** `test-performance-quick.sh`

**Purpose:** Rapid validation of low resource usage

**Usage:**
```bash
./test-performance-quick.sh
```

**Duration:** 30 seconds

**What it tests:**
- CPU usage stays <5%
- Memory usage stays <30 MB
- Agent doesn't crash

**Best for:** Quick validation after changes

---

### 2. Comprehensive Performance Test (5 minutes)

**File:** `test-performance.sh`

**Purpose:** Detailed performance analysis with metrics

**Usage:**
```bash
./test-performance.sh
```

**Duration:** 5 minutes (configurable)

**What it tests:**
- CPU usage over time
- Memory usage over time
- Thread count
- File descriptor count
- Generates detailed reports and CSV data

**Output files:**
- `performance_results.txt` - Summary
- `performance_data.csv` - Time-series data
- `awagent_test.log` - Agent logs

**Best for:** Pre-release validation, regression testing

---

### 3. Memory Leak Test (10 minutes)

**File:** `test-memory-leak.sh`

**Purpose:** Validates memory remains stable over time

**Usage:**
```bash
./test-memory-leak.sh
```

**Duration:** 10 minutes

**What it tests:**
- Memory growth over time
- Detects leaks (>10 MB growth)
- Can run without ActivityWatch server (uses --test mode)

**Best for:** Long-term stability testing

---

### 4. Functional Tests (Manual)

**File:** `TESTING.md`

**Purpose:** Validate core functionality

**Tests:**
- Repository discovery
- Git metadata extraction
- Session tracking
- Bucket creation
- Event publishing
- ActivityWatch integration

**Best for:** Feature validation

---

### 5. Build Tests

**Files:** `build.sh`, `install.sh`

**Purpose:** Validate build and installation

**Tests:**
- Binary builds successfully
- Static linking works
- Systemd service installs
- Cross-platform compatibility

**Best for:** Deployment validation

---

## Test Matrix

### Before Release Checklist

Run these tests before releasing:

```bash
# 1. Build the binary
./build.sh

# 2. Quick smoke test
./test-performance-quick.sh

# 3. Comprehensive performance test
./test-performance.sh

# 4. Memory leak test
./test-memory-leak.sh

# 5. Manual functional tests
# Follow TESTING.md

# 6. Cross-platform build test
GOOS=linux go build ./cmd/awagent
GOOS=darwin go build ./cmd/awagent
GOOS=windows go build ./cmd/awagent
```

### Performance Benchmarks

**Expected values:**

| Metric | Target | Maximum | Actual (your system) |
|--------|--------|---------|----------------------|
| CPU (avg) | <1% | <5% | ___ % |
| CPU (peak) | <2% | <10% | ___ % |
| Memory (avg) | 15-20 MB | 30 MB | ___ MB |
| Memory (peak) | 20-25 MB | 50 MB | ___ MB |
| Threads | ~10 | 20 | ___ |
| File Descriptors | ~15 | 50 | ___ |

### Stress Tests

#### High Repository Count

```bash
# Create 50 test repositories
for i in {1..50}; do
    mkdir -p /tmp/test-repos/repo$i
    cd /tmp/test-repos/repo$i
    git init
    git config user.email "test@test.com"
    git config user.name "Test"
    touch README.md
    git add . && git commit -m "Initial"
done

# Configure to scan them
cat > config.json << EOF
{
  "roots": ["/tmp/test-repos"],
  "maxDepth": 2,
  "rescanIntervalMinutes": 1,
  "pulseTime": 30
}
EOF

# Run performance test
./test-performance.sh
```

#### Rapid Window Switching

```bash
# Test window detection overhead
./awagent &
AGENT_PID=$!

# Manually switch windows rapidly for 60 seconds
# Or automate with xdotool:
for i in {1..30}; do
    xdotool key alt+Tab
    sleep 2
done

# Check performance
ps -p $AGENT_PID -o %cpu,%mem

kill $AGENT_PID
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build
        run: |
          CGO_ENABLED=0 go build -o awagent ./cmd/awagent
          
      - name: Install Dependencies
        run: sudo apt-get install -y xdotool
      
      - name: Quick Performance Test
        run: ./test-performance-quick.sh
      
      - name: Memory Leak Test
        run: ./test-memory-leak.sh
```

## Local Development Workflow

### During Development

```bash
# 1. Make changes to code

# 2. Build
go build -o awagent ./cmd/awagent

# 3. Quick test
./test-performance-quick.sh

# 4. If performance looks good, test manually
./awagent --verbose
```

### Before Committing

```bash
# 1. Run full test suite
./test-performance.sh

# 2. Check for memory leaks
./test-memory-leak.sh

# 3. Manual functional test
# Follow TESTING.md

# 4. Commit if all tests pass
git commit -m "Your changes"
```

## Troubleshooting Tests

### Test Script Fails to Start Agent

```bash
# Check if binary exists
ls -lh awagent

# Make executable
chmod +x awagent

# Try running manually
./awagent --help
```

### High Resource Usage During Tests

```bash
# Check what's consuming resources
top -p $(pgrep awagent)

# View agent logs
tail -f awagent_test.log

# Try test mode (no window detection)
./awagent --test
```

### Tests Timeout

```bash
# Check if ActivityWatch is responding
curl http://localhost:5600/api/0/info

# Reduce test duration
# Edit test-performance.sh:
TEST_DURATION=60  # 1 minute instead of 5
```

## Documentation

- **PERFORMANCE.md** - Detailed performance testing guide
- **TESTING.md** - Functional testing guide  
- **README.md** - General usage and setup
- **DEPLOYMENT.md** - Deployment instructions

## Summary

**Quick validation:**
```bash
./test-performance-quick.sh  # 30 seconds
```

**Full validation:**
```bash
./test-performance.sh        # 5 minutes
./test-memory-leak.sh        # 10 minutes
```

**Manual testing:**
```bash
# See TESTING.md for step-by-step guide
```

All tests should pass before releasing or deploying!
