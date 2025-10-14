#!/bin/bash

# Test script to verify enhanced alarm message with description and last values
echo "=========================================="
echo "Testing Enhanced Alarm Messages"
echo "=========================================="
echo ""
echo "Testing: Description on line 2, last_lux variable"
echo "This test will run for 130 seconds to get 2 observations"
echo ""

# Start the application
echo "Starting the application..."
./tempest-homekit-go --loglevel info --alarms @tempest-alarms.json --web-port 8084 > /tmp/enhanced-alarm-test.log 2>&1 &
PID=$!

echo "Application started with PID: $PID"
echo "Waiting 130 seconds for 2 observations..."
sleep 130

# Kill the application
kill $PID 2>/dev/null
wait $PID 2>/dev/null

echo ""
echo "=========================================="
echo "Results:" 
echo "=========================================="
echo ""

# Show alarm trigger
if grep -q "🚨 Alarm triggered: Lux Change" /tmp/enhanced-alarm-test.log; then
    echo "✅ Alarm triggered successfully!"
    echo ""
    echo "Console notification with new format:"
    echo "---"
    grep -A 7 "🚨 ALARM: Lux Change" /tmp/enhanced-alarm-test.log | head -8
    echo "---"
    echo ""
    
    # Check if description appears on line 2
    if grep -A 1 "🚨 ALARM: Lux Change" /tmp/enhanced-alarm-test.log | grep -q "This alarm should alert on LUX change"; then
        echo "✅ Description appears on line 2"
    else
        echo "❌ Description not on line 2"
    fi
    
    # Check if last_lux variable works
    if grep -A 7 "🚨 ALARM: Lux Change" /tmp/enhanced-alarm-test.log | grep -q "Previous LUX:"; then
        echo "✅ last_lux variable present"
        grep -A 7 "🚨 ALARM: Lux Change" /tmp/enhanced-alarm-test.log | grep "Previous LUX:"
    else
        echo "❌ last_lux variable missing"
    fi
else
    echo "❌ Alarm did not trigger"
    echo ""
    echo "Last 20 lines of log:"
    tail -20 /tmp/enhanced-alarm-test.log
fi

echo ""
echo "Full log saved to: /tmp/enhanced-alarm-test.log"
