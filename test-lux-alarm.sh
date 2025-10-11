#!/bin/bash

# Test script to verify "Lux Change" alarm triggers correctly
# This simulates the lux changes the user observed

echo "=========================================="
echo "Testing Lux Change Alarm with *lux operator"
echo "=========================================="
echo ""
echo "This test will run for about 30 seconds to simulate lux changes."
echo "Expected behavior: Alarm should trigger after the first lux value is established."
echo ""
echo "Starting application with --loglevel info..."
echo ""

# Run with timeout to auto-terminate after 30 seconds
timeout 30s ./tempest-homekit-go --loglevel info --alarms @tempest-alarms.json 2>&1 | tee /tmp/lux-test-output.txt

echo ""
echo "=========================================="
echo "Test completed. Checking results..."
echo "=========================================="
echo ""

# Check if alarm was triggered
if grep -q "🚨 Alarm triggered: Lux Change" /tmp/lux-test-output.txt; then
    echo "✅ SUCCESS: Lux Change alarm triggered as expected!"
    echo ""
    echo "Alarm trigger details:"
    grep "🚨 Alarm triggered: Lux Change" /tmp/lux-test-output.txt
    echo ""
    echo "Console notifications:"
    grep -A 4 "🚨 ALARM: Lux Change" /tmp/lux-test-output.txt | head -5
else
    echo "❌ FAILURE: Lux Change alarm did not trigger"
    echo ""
    echo "Last 20 lines of output:"
    tail -20 /tmp/lux-test-output.txt
fi

echo ""
echo "Full output saved to: /tmp/lux-test-output.txt"
