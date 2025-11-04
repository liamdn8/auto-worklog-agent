#!/bin/bash
# Test simulation: Create commits and run awagent to capture them

set -e

echo "=== Commit Tracking Test Simulation ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Create test commits in current repo
echo -e "${BLUE}Step 1: Creating test commits...${NC}"
echo ""

# Save current branch
CURRENT_BRANCH=$(git branch --show-current)

# Create a test branch
TEST_BRANCH="test/PROJ-123-commit-tracking-demo"
echo "Creating test branch: $TEST_BRANCH"
git checkout -b "$TEST_BRANCH" 2>/dev/null || git checkout "$TEST_BRANCH"

# Create first commit
echo "Test commit 1: $(date)" > test-commit-1.txt
git add test-commit-1.txt
git commit -m "PROJ-123: Add initial test file for commit tracking demo"

# Create second commit
echo "Test commit 2: $(date)" > test-commit-2.txt
git add test-commit-2.txt
git commit -m "PROJ-123: Add second test file to validate multiple commits"

# Create third commit
echo "Test commit 3: $(date)" > test-commit-3.txt
git add test-commit-3.txt
git commit -m "PROJ-456: Fix bug in related feature (different issue)"

echo -e "${GREEN}✓ Created 3 test commits${NC}"
echo ""

# Step 2: Show commits
echo -e "${BLUE}Step 2: Test commits created:${NC}"
git log --oneline -3
echo ""

# Step 3: Rebuild binary
echo -e "${BLUE}Step 3: Rebuilding awagent with commit tracking...${NC}"
go build -o awagent cmd/awagent/main.go
echo -e "${GREEN}✓ Build complete${NC}"
echo ""

# Step 4: Run test mode
echo -e "${BLUE}Step 4: Running awagent in test mode for 35 seconds...${NC}"
echo "This will create sessions with commit data"
echo ""

timeout 35 ./awagent --aw-url http://localhost:9600 --test -v || true

echo ""
echo -e "${GREEN}✓ Test run complete${NC}"
echo ""

# Step 5: Query ActivityWatch for commit data
echo -e "${BLUE}Step 5: Querying ActivityWatch for commit data...${NC}"
echo ""

# Get current user
GIT_USER=$(git config user.name | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9_-]/-/g' | sed 's/^-*//;s/-*$//')
REPO_NAME="auto-worklog-agent"
BRANCH_NAME=$(echo "$TEST_BRANCH" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9_-]/-/g' | sed 's/^-*//;s/-*$//')
BUCKET_ID="${GIT_USER}_${REPO_NAME}_${BRANCH_NAME}"

echo "Expected bucket: $BUCKET_ID"
echo ""

# Query the bucket
echo -e "${YELLOW}Events with commits:${NC}"
curl -s "http://localhost:9600/api/0/buckets/${BUCKET_ID}/events?limit=5" | jq '.[] | {
  timestamp: .timestamp,
  duration: .duration,
  eventCount: .data.eventCount,
  commits: .data.commits | length,
  commitDetails: .data.commits
}'

echo ""
echo -e "${GREEN}=== Test Complete ===${NC}"
echo ""
echo "To clean up:"
echo "  git checkout $CURRENT_BRANCH"
echo "  git branch -D $TEST_BRANCH"
echo "  rm -f test-commit-*.txt"
echo ""
echo "To view all data in ActivityWatch Web UI:"
echo "  Open: http://localhost:9600"
