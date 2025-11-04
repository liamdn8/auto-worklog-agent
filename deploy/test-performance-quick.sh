#!/bin/bash
# Quick Performance Test (30 seconds)
# Validates awagent runs with low resource usage

set -e

echo "=== Quick Performance Test (30s) ==="
echo ""

if [ ! -f "./awagent" ]; then
    echo "Error: awagent not found"
    exit 1
fi

# Start awagent
./awagent > /dev/null 2>&1 &
PID=$!
sleep 2

# Check if running
if ! kill -0 $PID 2>/dev/null; then
    echo "Error: awagent failed to start"
    exit 1
fi

echo "Monitoring awagent (PID: $PID) for 30 seconds..."
echo ""
echo "Time | CPU% | Memory MB"
echo "-----+------+----------"

# Monitor for 30 seconds
for i in {1..6}; do
    sleep 5
    
    if ! kill -0 $PID 2>/dev/null; then
        echo "Error: awagent crashed"
        exit 1
    fi
    
    # Get stats from /proc
    if [ -f "/proc/$PID/stat" ]; then
        # Read RSS (resident set size)
        rss=$(awk '{print $24}' /proc/$PID/stat)
        mem_mb=$(awk "BEGIN {printf \"%.1f\", $rss * 4096 / 1024 / 1024}")
        
        # Simple CPU calculation
        cpu=$(ps -p $PID -o %cpu --no-headers | tr -d ' ')
        
        printf "%3ds | %4s | %8s\n" $((i*5)) "$cpu" "$mem_mb"
    fi
done

echo ""

# Get final stats
if [ -f "/proc/$PID/stat" ]; then
    rss=$(awk '{print $24}' /proc/$PID/stat)
    mem_mb=$(awk "BEGIN {printf \"%.1f\", $rss * 4096 / 1024 / 1024}")
    cpu=$(ps -p $PID -o %cpu --no-headers | tr -d ' ')
    
    echo "Final Stats:"
    echo "  CPU: ${cpu}%"
    echo "  Memory: ${mem_mb} MB"
    echo ""
    
    # Check thresholds
    mem_int=${mem_mb%.*}
    cpu_int=${cpu%.*}
    
    if [ "${mem_int:-0}" -lt 30 ] && [ "${cpu_int:-0}" -lt 5 ]; then
        echo "✓ PASSED - Low resource usage confirmed"
        RESULT=0
    else
        echo "✗ WARNING - Higher than expected resource usage"
        echo "  (Expected: <5% CPU, <30 MB memory)"
        RESULT=1
    fi
fi

# Cleanup
kill $PID 2>/dev/null || true
wait $PID 2>/dev/null || true

exit $RESULT
