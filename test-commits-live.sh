#!/bin/bash
# Live test: Start awagent, then create commits while it's running

set -e

echo "=== Live Commit Tracking Test ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Save current branch
CURRENT_BRANCH=$(git branch --show-current)
TEST_BRANCH="test/PROJ-789-live-commit-test"

# Cleanup function
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"
    git checkout "$CURRENT_BRANCH" 2>/dev/null || true
    git branch -D "$TEST_BRANCH" 2>/dev/null || true
    rm -f test-live-*.txt
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

trap cleanup EXIT

# Step 1: Create test branch (no commits yet)
echo -e "${BLUE}Step 1: Creating test branch (no commits yet)...${NC}"
git checkout -b "$TEST_BRANCH" 2>/dev/null || git checkout "$TEST_BRANCH"
echo -e "${GREEN}✓ On branch: $TEST_BRANCH${NC}"
echo ""

# Step 2: Rebuild binary
echo -e "${BLUE}Step 2: Rebuilding awagent...${NC}"
go build -o awagent cmd/awagent/main.go
echo -e "${GREEN}✓ Build complete${NC}"
echo ""

# Step 3: Start awagent in background
echo -e "${BLUE}Step 3: Starting awagent in test mode...${NC}"
./awagent --aw-url http://localhost:9600 --test -v > awagent-test.log 2>&1 &
AWAGENT_PID=$!
echo "awagent PID: $AWAGENT_PID"
echo -e "${GREEN}✓ awagent started${NC}"
echo ""

# Wait for initial session to start
sleep 12

# Step 4: Create commits WHILE awagent is running
echo -e "${BLUE}Step 4: Creating commits while awagent is running...${NC}"
echo ""

sleep 2

echo "Creating commit 1..."
echo "Live test commit 1: $(date)" > test-live-1.txt
git add test-live-1.txt
git commit -m "PROJ-789: Add first live test commit"
echo -e "${GREEN}✓ Commit 1 created${NC}"

sleep 3

echo "Creating commit 2..."
echo "Live test commit 2: $(date)" > test-live-2.txt
git add test-live-2.txt
git commit -m "PROJ-789: Add second live test commit"
echo -e "${GREEN}✓ Commit 2 created${NC}"

sleep 3

echo "Creating commit 3..."
echo "Live test commit 3: $(date)" > test-live-3.txt
git add test-live-3.txt
git commit -m "PROJ-999: Different issue commit"
echo -e "${GREEN}✓ Commit 3 created${NC}"

echo ""
echo -e "${YELLOW}Waiting for next flush cycle (15 seconds)...${NC}"
sleep 18

# Step 5: Stop awagent
echo ""
echo -e "${BLUE}Step 5: Stopping awagent...${NC}"
kill $AWAGENT_PID 2>/dev/null || true
wait $AWAGENT_PID 2>/dev/null || true
echo -e "${GREEN}✓ awagent stopped${NC}"
echo ""

# Step 6: Show log tail
echo -e "${BLUE}Step 6: Last 20 lines from awagent log:${NC}"
tail -20 awagent-test.log
echo ""

# Step 7: Query ActivityWatch
echo -e "${BLUE}Step 7: Querying ActivityWatch for commit data...${NC}"
echo ""

GIT_USER=$(git config user.name | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9_-]/-/g' | sed 's/^-*//;s/-*$//')
REPO_NAME="auto-worklog-agent"
BRANCH_NAME=$(echo "$TEST_BRANCH" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9_-]/-/g' | sed 's/^-*//;s/-*$//')
BUCKET_ID="${GIT_USER}_${REPO_NAME}_${BRANCH_NAME}"

echo "Bucket ID: $BUCKET_ID"
echo ""

# Show commits
echo -e "${YELLOW}Git commits created:${NC}"
git log --oneline -3
echo ""

# Query events
echo -e "${YELLOW}ActivityWatch events:${NC}"
EVENTS=$(curl -s "http://localhost:9600/api/0/buckets/${BUCKET_ID}/events?limit=10")
echo "$EVENTS" | jq '.'
echo ""

# Pretty print commit details
echo -e "${YELLOW}Commit details from ActivityWatch:${NC}"
echo "$EVENTS" | jq -r '.[] | select(.data.commits != null) | "Event \(.id): \(.data.commits | length) commits\n" + (.data.commits // [] | map("  - \(.hash[0:8]) \(.message)") | join("\n"))'

echo ""
echo -e "${GREEN}=== Test Complete ===${NC}"
echo ""
echo "Log file saved: awagent-test.log"
