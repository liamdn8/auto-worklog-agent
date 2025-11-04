#!/bin/bash
# Memory Leak Test - Validates memory stays stable over time
# Runs for 10 minutes and checks for memory growth

set -e

DURATION=600  # 10 minutes
SAMPLE_INTERVAL=30  # Every 30 seconds

echo "=== Memory Leak Test ==="
echo "Duration: $((DURATION / 60)) minutes"
echo "Checking for memory growth..."
echo ""

if [ ! -f "./awagent" ]; then
    echo "Error: awagent not found"
    exit 1
fi

# Start awagent
./awagent --test > /dev/null 2>&1 &
PID=$!
sleep 3

if ! kill -0 $PID 2>/dev/null; then
    echo "Error: awagent failed to start"
    exit 1
fi

echo "Monitoring memory for $((DURATION / 60)) minutes..."
echo "Sample | Memory MB | Growth"
echo "-------+-----------+-------"

# First sample
rss=$(awk '{print $24}' /proc/$PID/stat 2>/dev/null)
initial_mem=$(awk "BEGIN {printf \"%.1f\", $rss * 4096 / 1024 / 1024}")
prev_mem=$initial_mem
max_mem=$initial_mem
sample=0

START=$(date +%s)
while [ $(($(date +%s) - START)) -lt $DURATION ]; do
    sleep $SAMPLE_INTERVAL
    
    if ! kill -0 $PID 2>/dev/null; then
        echo "Error: awagent crashed"
        exit 1
    fi
    
    sample=$((sample + 1))
    rss=$(awk '{print $24}' /proc/$PID/stat 2>/dev/null)
    current_mem=$(awk "BEGIN {printf \"%.1f\", $rss * 4096 / 1024 / 1024}")
    growth=$(awk "BEGIN {printf \"%.1f\", $current_mem - $initial_mem}")
    
    # Track max
    max_mem=$(awk "BEGIN {if ($current_mem > $max_mem) print $current_mem; else print $max_mem}" max_mem=$max_mem)
    
    printf "  %2d   | %9s | +%5s\n" "$sample" "$current_mem" "$growth"
done

echo ""
echo "Results:"
echo "  Initial Memory: ${initial_mem} MB"
echo "  Final Memory:   ${current_mem} MB"
echo "  Maximum Memory: ${max_mem} MB"
echo "  Growth:         ${growth} MB"
echo ""

# Check for leak (>10 MB growth is suspicious)
growth_int=${growth%.*}
if [ "${growth_int:-0}" -gt 10 ] || [ "${growth_int:-0}" -lt -10 ]; then
    abs_growth=${growth_int#-}
    if [ "$abs_growth" -gt 10 ]; then
        echo "✗ WARNING - Possible memory leak detected"
        echo "  Memory grew by ${growth} MB over $((DURATION / 60)) minutes"
        RESULT=1
    else
        echo "✓ PASSED - Memory stable"
        RESULT=0
    fi
else
    echo "✓ PASSED - Memory stable (${growth} MB change is normal)"
    RESULT=0
fi

# Cleanup
kill $PID 2>/dev/null || true
wait $PID 2>/dev/null || true

exit $RESULT
