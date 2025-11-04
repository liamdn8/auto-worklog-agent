#!/bin/bash
# Quick demo: Create real data to view in browser RIGHT NOW

set -e

echo "=== Quick Browser Demo ==="
echo ""

CURRENT_BRANCH=$(git branch --show-current)
TEST_BRANCH="demo/QUICK-999-browser-test"

cleanup() {
    git checkout "$CURRENT_BRANCH" 2>/dev/null || true
    git branch -D "$TEST_BRANCH" 2>/dev/null || true
    rm -f quick-*.txt
}
trap cleanup EXIT

# Create branch
echo "1. Creating test branch..."
git checkout -b "$TEST_BRANCH" 2>/dev/null || git checkout "$TEST_BRANCH"

# Create 3 commits
echo "2. Creating 3 commits..."
echo "Quick test 1" > quick-1.txt
git add quick-1.txt
git commit -m "QUICK-999: First commit for browser demo" -q

echo "Quick test 2" > quick-2.txt
git add quick-2.txt
git commit -m "QUICK-999: Second commit with more work" -q

echo "Quick test 3" > quick-3.txt
git add quick-3.txt
git commit -m "QUICK-888: Bug fix in related feature" -q

echo "✓ Commits created"
echo ""

# Show commits
echo "3. Commits:"
git log --oneline -3
echo ""

# Rebuild and run
echo "4. Building and running awagent..."
go build -o awagent cmd/awagent/main.go 2>&1 | grep -v "warning:" || true

timeout 35 ./awagent --aw-url http://localhost:9600 --test 2>&1 | grep -E "(Session started|Activity detected|Publishing|Session published)" &
AWAGENT_PID=$!

echo "   Waiting for data collection (35 seconds)..."
wait $AWAGENT_PID 2>/dev/null || true

echo ""
echo "5. ✓ Data sent to ActivityWatch!"
echo ""

# Get bucket ID
GIT_USER=$(git config user.name | tr '[:upper:]' '[:lower:]' | sed 's/[^a-zA-Z0-9_-]/-/g' | sed 's/^-*//;s/-*$//')
BUCKET_ID="${GIT_USER}_auto-worklog-agent_demo-quick-999-browser-test"

echo "6. View in browser:"
echo ""
echo "   URL: http://localhost:9600"
echo "   Bucket: $BUCKET_ID"
echo ""
echo "   Or query directly:"
echo "   curl -s \"http://localhost:9600/api/0/buckets/${BUCKET_ID}/events?limit=1\" | jq '.'"
echo ""

# Show data preview
echo "7. Preview of captured data:"
sleep 2
curl -s "http://localhost:9600/api/0/buckets/${BUCKET_ID}/events?limit=1" 2>/dev/null | jq '.[0] | {
  timestamp,
  duration,
  branch: .data.branch,
  commits: (.data.commits | length),
  issues: [.data.commits[].message | scan("QUICK-[0-9]+")] | unique,
  commitDetails: .data.commits
}' || echo "Data not yet available - refresh browser"

echo ""
echo "=== Done! Check the browser now ===" 
echo ""
