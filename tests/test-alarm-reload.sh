#!/bin/bash

# Test script to verify alarm config reload shows detailed logging
echo "=========================================="
echo "Testing Alarm Config Reload Logging"
echo "=========================================="
echo ""

# Start the application in background
echo "Starting application with alarm file watching..."
./tempest-homekit-go --loglevel info --alarms @tempest-alarms.json --web-port 8083 > /tmp/reload-test.log 2>&1 &
PID=$!
echo "Application started with PID: $PID"

# Wait for initialization
echo "Waiting 5 seconds for initialization..."
sleep 5

echo ""
echo "Initial load output:"
grep -E "(Alarm manager initialized|Loaded alarm:|alarms are enabled)" /tmp/reload-test.log
echo ""

# Trigger a reload by touching the file
echo "Triggering config reload by touching tempest-alarms.json..."
touch tempest-alarms.json

# Wait for reload to happen
echo "Waiting 3 seconds for file watcher to detect change..."
sleep 3

echo ""
echo "=========================================="
echo "Reload Output:"
echo "=========================================="
grep -A 10 "Alarm config file changed" /tmp/reload-test.log | tail -15

echo ""
echo "=========================================="
echo "Full reload section:"
echo "=========================================="
grep -E "(Alarm config file changed|Alarm config reloaded|Alarm manager initialized|Loaded alarm:|alarms are enabled)" /tmp/reload-test.log | tail -10

# Clean up
echo ""
echo "Stopping application..."
kill $PID 2>/dev/null
wait $PID 2>/dev/null

echo ""
echo "✅ Test complete. Full log saved to: /tmp/reload-test.log"
echo ""
echo "Expected output after reload:"
echo "  - INFO: Alarm config file changed, reloading: tempest-alarms.json"
echo "  - INFO: Alarm manager initialized with 3 alarms"
echo "  - INFO: Loaded alarm: Hot outside"
echo "  - INFO: Loaded alarm: Lightning Nearby"
echo "  - INFO: Loaded alarm: Lux Change"
echo "  - INFO: 3 of 3 alarms are enabled"
echo "  - INFO: Alarm config reloaded successfully"
