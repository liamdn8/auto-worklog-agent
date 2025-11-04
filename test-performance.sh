#!/bin/bash
# Performance Test for awagent
# Validates low resource usage during normal operation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Performance thresholds
MAX_MEMORY_MB=30      # Maximum memory usage in MB
MAX_CPU_PERCENT=5     # Maximum CPU usage in %
TEST_DURATION=300     # Run for 5 minutes (300 seconds)
SAMPLE_INTERVAL=5     # Sample every 5 seconds

# Output files
RESULTS_FILE="performance_results.txt"
CSV_FILE="performance_data.csv"

echo "========================================"
echo "  awagent Performance Test"
echo "========================================"
echo ""
echo "Test Parameters:"
echo "  Duration: ${TEST_DURATION}s ($(($TEST_DURATION / 60)) minutes)"
echo "  Sample Interval: ${SAMPLE_INTERVAL}s"
echo "  Max Memory Threshold: ${MAX_MEMORY_MB} MB"
echo "  Max CPU Threshold: ${MAX_CPU_PERCENT}%"
echo ""

# Check if awagent exists
if [ ! -f "./awagent" ]; then
    echo -e "${RED}Error: awagent binary not found${NC}"
    echo "Please run this test from the directory containing awagent"
    exit 1
fi

# Check if ActivityWatch is running
if ! curl -s http://localhost:5600/api/0/info > /dev/null 2>&1; then
    echo -e "${YELLOW}Warning: ActivityWatch server not detected${NC}"
    echo "Starting test anyway, but some features may not work"
    echo ""
fi

# Clean up previous results
rm -f "$RESULTS_FILE" "$CSV_FILE"

# Create CSV header
echo "Timestamp,CPU%,Memory_MB,Memory_RSS_KB,Threads,FDs" > "$CSV_FILE"

# Start awagent in background
echo -e "${BLUE}Starting awagent...${NC}"
./awagent --verbose > awagent_test.log 2>&1 &
AGENT_PID=$!

# Wait for it to start
sleep 3

# Check if it's still running
if ! kill -0 $AGENT_PID 2>/dev/null; then
    echo -e "${RED}Error: awagent failed to start${NC}"
    echo "Check awagent_test.log for details"
    exit 1
fi

echo -e "${GREEN}awagent started (PID: $AGENT_PID)${NC}"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Stopping awagent...${NC}"
    kill $AGENT_PID 2>/dev/null || true
    wait $AGENT_PID 2>/dev/null || true
    echo "Cleanup complete"
}
trap cleanup EXIT INT TERM

# Arrays to store metrics
declare -a cpu_samples
declare -a mem_samples
declare -a rss_samples

# Monitor loop
echo -e "${BLUE}Monitoring performance...${NC}"
echo "Time | CPU% | Memory MB | RSS KB | Threads | FDs"
echo "-----+------+-----------+--------+---------+-----"

START_TIME=$(date +%s)
SAMPLE_COUNT=0
VIOLATIONS_CPU=0
VIOLATIONS_MEM=0

while [ $(($(date +%s) - START_TIME)) -lt $TEST_DURATION ]; do
    # Check if process is still running
    if ! kill -0 $AGENT_PID 2>/dev/null; then
        echo -e "${RED}Error: awagent died during test${NC}"
        exit 1
    fi
    
    # Get process stats
    if [ -f "/proc/$AGENT_PID/stat" ]; then
        # CPU calculation
        # Get total CPU time and system uptime
        read -r pid comm state ppid pgrp session tty_nr tpgid flags minflt cminflt majflt cmajflt \
                utime stime cutime cstime priority nice num_threads itrealvalue starttime vsize rss \
                < /proc/$AGENT_PID/stat
        
        # Get system stats
        read -r cpu user nice system idle iowait irq softirq steal guest guest_nice \
                < /proc/stat
        
        # Calculate CPU % (simplified)
        total_time=$((utime + stime))
        
        # Sleep and recalculate
        sleep $SAMPLE_INTERVAL
        
        if ! kill -0 $AGENT_PID 2>/dev/null; then
            break
        fi
        
        read -r pid2 comm2 state2 ppid2 pgrp2 session2 tty_nr2 tpgid2 flags2 minflt2 cminflt2 majflt2 cmajflt2 \
                utime2 stime2 cutime2 cstime2 priority2 nice2 num_threads2 itrealvalue2 starttime2 vsize2 rss2 \
                < /proc/$AGENT_PID/stat 2>/dev/null || break
        
        read -r cpu2 user2 nice2 system2 idle2 iowait2 irq2 softirq2 steal2 guest2 guest_nice2 \
                < /proc/stat
        
        # Calculate CPU usage
        total_time2=$((utime2 + stime2))
        total_delta=$((total_time2 - total_time))
        
        system_time=$((user2 + nice2 + system2 + idle2 + iowait2 + irq2 + softirq2 + steal2))
        system_time_before=$((user + nice + system + idle + iowait + irq + softirq + steal))
        system_delta=$((system_time - system_time_before))
        
        if [ $system_delta -gt 0 ]; then
            cpu_percent=$(awk "BEGIN {printf \"%.2f\", ($total_delta / $system_delta) * 100}")
        else
            cpu_percent="0.00"
        fi
        
        # Memory in MB (RSS)
        mem_mb=$(awk "BEGIN {printf \"%.2f\", $rss2 * 4096 / 1024 / 1024}")
        rss_kb=$((rss2 * 4))
        
        # Thread count
        threads=$num_threads2
        
        # File descriptors
        fds=$(ls -1 /proc/$AGENT_PID/fd 2>/dev/null | wc -l)
        
        # Store samples
        cpu_samples+=($cpu_percent)
        mem_samples+=($mem_mb)
        rss_samples+=($rss_kb)
        
        # Check thresholds
        cpu_int=${cpu_percent%.*}
        mem_int=${mem_mb%.*}
        
        if [ "${cpu_int:-0}" -gt "$MAX_CPU_PERCENT" ]; then
            VIOLATIONS_CPU=$((VIOLATIONS_CPU + 1))
            echo -ne "${RED}"
        elif [ "${mem_int:-0}" -gt "$MAX_MEMORY_MB" ]; then
            VIOLATIONS_MEM=$((VIOLATIONS_MEM + 1))
            echo -ne "${RED}"
        else
            echo -ne "${GREEN}"
        fi
        
        # Display current stats
        elapsed=$(($(date +%s) - START_TIME))
        printf "%4ds | %5s | %9s | %6s | %7s | %3s${NC}\n" \
               "$elapsed" "$cpu_percent" "$mem_mb" "$rss_kb" "$threads" "$fds"
        
        # Save to CSV
        echo "$(date +%s),$cpu_percent,$mem_mb,$rss_kb,$threads,$fds" >> "$CSV_FILE"
        
        SAMPLE_COUNT=$((SAMPLE_COUNT + 1))
    else
        sleep $SAMPLE_INTERVAL
    fi
done

echo ""
echo "========================================"
echo "  Performance Test Results"
echo "========================================"
echo ""

# Calculate statistics
if [ ${#cpu_samples[@]} -gt 0 ]; then
    # Average CPU
    cpu_sum=0
    cpu_max=0
    for val in "${cpu_samples[@]}"; do
        cpu_sum=$(awk "BEGIN {print $cpu_sum + $val}")
        cpu_max=$(awk "BEGIN {if ($val > $cpu_max) print $val; else print $cpu_max}" cpu_max=$cpu_max)
    done
    cpu_avg=$(awk "BEGIN {printf \"%.2f\", $cpu_sum / ${#cpu_samples[@]}}")
    
    # Average Memory
    mem_sum=0
    mem_max=0
    for val in "${mem_samples[@]}"; do
        mem_sum=$(awk "BEGIN {print $mem_sum + $val}")
        mem_max=$(awk "BEGIN {if ($val > $mem_max) print $val; else print $mem_max}" mem_max=$mem_max)
    done
    mem_avg=$(awk "BEGIN {printf \"%.2f\", $mem_sum / ${#mem_samples[@]}}")
    
    echo "Samples Collected: $SAMPLE_COUNT"
    echo "Test Duration: ${TEST_DURATION}s ($(($TEST_DURATION / 60)) minutes)"
    echo ""
    echo "CPU Usage:"
    echo "  Average: ${cpu_avg}%"
    echo "  Maximum: ${cpu_max}%"
    echo "  Threshold: ${MAX_CPU_PERCENT}%"
    if [ "$VIOLATIONS_CPU" -gt 0 ]; then
        echo -e "  ${RED}Violations: $VIOLATIONS_CPU${NC}"
    else
        echo -e "  ${GREEN}Violations: 0${NC}"
    fi
    echo ""
    echo "Memory Usage:"
    echo "  Average: ${mem_avg} MB"
    echo "  Maximum: ${mem_max} MB"
    echo "  Threshold: ${MAX_MEMORY_MB} MB"
    if [ "$VIOLATIONS_MEM" -gt 0 ]; then
        echo -e "  ${RED}Violations: $VIOLATIONS_MEM${NC}"
    else
        echo -e "  ${GREEN}Violations: 0${NC}"
    fi
    echo ""
    
    # Overall result
    echo "========================================"
    if [ "$VIOLATIONS_CPU" -eq 0 ] && [ "$VIOLATIONS_MEM" -eq 0 ]; then
        echo -e "${GREEN}✓ PASSED - Low resource usage confirmed${NC}"
        EXIT_CODE=0
    else
        echo -e "${RED}✗ FAILED - Resource usage exceeded thresholds${NC}"
        EXIT_CODE=1
    fi
    echo "========================================"
    echo ""
    
    # Save detailed results
    {
        echo "awagent Performance Test Results"
        echo "================================"
        echo "Date: $(date)"
        echo "Duration: ${TEST_DURATION}s"
        echo "Samples: $SAMPLE_COUNT"
        echo ""
        echo "CPU Usage:"
        echo "  Average: ${cpu_avg}%"
        echo "  Maximum: ${cpu_max}%"
        echo "  Threshold: ${MAX_CPU_PERCENT}%"
        echo "  Violations: $VIOLATIONS_CPU"
        echo ""
        echo "Memory Usage:"
        echo "  Average: ${mem_avg} MB"
        echo "  Maximum: ${mem_max} MB"
        echo "  Threshold: ${MAX_MEMORY_MB} MB"
        echo "  Violations: $VIOLATIONS_MEM"
        echo ""
        if [ "$VIOLATIONS_CPU" -eq 0 ] && [ "$VIOLATIONS_MEM" -eq 0 ]; then
            echo "Result: PASSED"
        else
            echo "Result: FAILED"
        fi
    } > "$RESULTS_FILE"
    
    echo "Detailed results saved to: $RESULTS_FILE"
    echo "Performance data saved to: $CSV_FILE"
    echo "Agent logs saved to: awagent_test.log"
    echo ""
    
    exit $EXIT_CODE
else
    echo -e "${RED}Error: No samples collected${NC}"
    exit 1
fi
