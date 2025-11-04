#!/bin/bash
# Demo: Show commit tracking in real-time in ActivityWatch Web UI

set -e

echo "=== ActivityWatch Commit Tracking Demo ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Save current state
CURRENT_BRANCH=$(git branch --show-current)
TEST_BRANCH="demo/TICKET-555-feature-demo"

# Cleanup
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"
    git checkout "$CURRENT_BRANCH" 2>/dev/null || true
    git branch -D "$TEST_BRANCH" 2>/dev/null || true
    rm -f demo-*.txt awagent-demo.log
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}
trap cleanup EXIT

echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  Step 1: Open ActivityWatch Web UI                        ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "Open in your browser: http://localhost:9600"
echo ""
echo "Press ENTER when ready to continue..."
read

# Create test branch
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  Step 2: Creating demo branch                             ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
git checkout -b "$TEST_BRANCH" 2>/dev/null || git checkout "$TEST_BRANCH"
echo -e "${GREEN}✓ Branch: $TEST_BRANCH${NC}"
echo ""
sleep 2

# Start awagent
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  Step 3: Starting awagent (test mode)                     ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
./awagent --aw-url http://localhost:9600 --test -v > awagent-demo.log 2>&1 &
AWAGENT_PID=$!
echo -e "${GREEN}✓ awagent running (PID: $AWAGENT_PID)${NC}"
echo ""
echo "Waiting for initial session to start (12 seconds)..."
sleep 12

# Create commits with delays
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  Step 4: Creating commits (watch ActivityWatch UI!)       ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${YELLOW}Creating commit 1/4...${NC}"
echo "Demo file 1: $(date)" > demo-1.txt
git add demo-1.txt
git commit -m "TICKET-555: Implement user authentication"
echo -e "${GREEN}✓ Committed: TICKET-555: Implement user authentication${NC}"
sleep 4

echo -e "${YELLOW}Creating commit 2/4...${NC}"
echo "Demo file 2: $(date)" > demo-2.txt
git add demo-2.txt
git commit -m "TICKET-555: Add login form validation"
echo -e "${GREEN}✓ Committed: TICKET-555: Add login form validation${NC}"
sleep 4

echo -e "${YELLOW}Creating commit 3/4...${NC}"
echo "Demo file 3: $(date)" > demo-3.txt
git add demo-3.txt
git commit -m "TICKET-666: Fix password reset bug"
echo -e "${GREEN}✓ Committed: TICKET-666: Fix password reset bug${NC}"
sleep 4

echo -e "${YELLOW}Creating commit 4/4...${NC}"
echo "Demo file 4: $(date)" > demo-4.txt
git add demo-4.txt
git commit -m "TICKET-555: Add session management"
echo -e "${GREEN}✓ Committed: TICKET-555: Add session management${NC}"
echo ""

echo -e "${YELLOW}Waiting for flush cycle (18 seconds)...${NC}"
echo "Watch the ActivityWatch UI - data will appear shortly!"
sleep 18

# Stop awagent
echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  Step 5: Stopping awagent                                 ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
kill $AWAGENT_PID 2>/dev/null || true
wait $AWAGENT_PID 2>/dev/null || true
echo -e "${GREEN}✓ awagent stopped${NC}"
echo ""

# Show results
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  Step 6: Results                                          ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Calculate bucket ID
GIT_USER=$(git config user.name | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9_-]/-/g' | sed 's/^-*//;s/-*$//')
BUCKET_ID="${GIT_USER}_auto-worklog-agent_demo-ticket-555-feature-demo"

echo -e "${BLUE}Bucket ID:${NC} $BUCKET_ID"
echo ""

# Show commits in git
echo -e "${BLUE}Git commits created:${NC}"
git log --oneline -4
echo ""

# Query ActivityWatch
echo -e "${BLUE}ActivityWatch event data:${NC}"
EVENT_DATA=$(curl -s "http://localhost:9600/api/0/buckets/${BUCKET_ID}/events?limit=1")

echo "$EVENT_DATA" | jq -C '.'
echo ""

# Extract and show commit details
echo -e "${BLUE}Commits captured in ActivityWatch:${NC}"
echo "$EVENT_DATA" | jq -r '.[0].data.commits[] | "  [\(.timestamp | split("T")[1] | split("+")[0])] \(.hash[0:8]) - \(.message)"'
echo ""

# Issue summary
echo -e "${BLUE}Issues detected:${NC}"
echo "$EVENT_DATA" | jq -r '.[0].data.commits[].message' | grep -oE '[A-Z]+-[0-9]+' | sort | uniq -c
echo ""

# Time calculation
DURATION=$(echo "$EVENT_DATA" | jq -r '.[0].duration')
MINUTES=$(echo "$DURATION / 60" | bc)
echo -e "${BLUE}Time spent:${NC} ${DURATION}s (${MINUTES} minutes)"
echo ""

echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Demo Complete! ✓                                         ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "View the full data in ActivityWatch UI:"
echo "  1. Go to: http://localhost:9600"
echo "  2. Click 'Raw Data' → Select bucket: $BUCKET_ID"
echo "  3. Inspect the 'commits' field in the event data"
echo ""
echo "Next steps:"
echo "  • Build a Jira worklog sync tool to process this data"
echo "  • Extract issue keys from commit messages"
echo "  • Calculate time per issue"
echo "  • Auto-post worklogs to Jira"
echo ""
