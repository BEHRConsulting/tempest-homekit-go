#!/bin/bash

echo "=========================================="
echo "Testing Alarm Console Output with Warning Log Level"
echo "=========================================="
echo ""

# Kill any existing instances
pkill -f tempest-homekit-go 2>/dev/null
sleep 2

echo "1. Starting app with --loglevel warning (should suppress INFO/DEBUG logs)"
echo "   But ALARM messages should still appear!"
echo ""

# Start app in background and capture output
./tempest-homekit-go --loglevel warning --alarms @tempest-alarms.json > /tmp/alarm-test.log 2>&1 &
APP_PID=$!

echo "   App started with PID: $APP_PID"
sleep 5

echo ""
echo "2. Checking if app is running..."
if ps -p $APP_PID > /dev/null; then
    echo "   ✅ App is running"
else
    echo "   ❌ App failed to start"
    exit 1
fi

echo ""
echo "3. Checking alarm status via API..."
ALARM_STATUS=$(curl -s http://localhost:8080/api/alarm-status | python3 -c "import sys, json; d=json.load(sys.stdin); print(f'{d[\"enabled\"]}:{d[\"totalAlarms\"]}:{d[\"enabledAlarms\"]}')" 2>/dev/null)

if [ "$ALARM_STATUS" = "True:4:4" ]; then
    echo "   ✅ Alarms are configured and enabled (4 of 4)"
else
    echo "   ⚠️  Unexpected alarm status: $ALARM_STATUS"
fi

echo ""
echo "4. Waiting 60 seconds for weather data and potential alarm triggers..."
echo "   (Wind Change alarm has 10s cooldown and triggers on any wind speed change)"
sleep 65

echo ""
echo "5. Checking log output for ALARM messages..."
echo "=========================================="

if grep -q "🚨 ALARM:" /tmp/alarm-test.log; then
    echo "✅ SUCCESS: ALARM messages found in output with warning log level!"
    echo ""
    echo "Alarm messages:"
    grep "🚨 ALARM:" /tmp/alarm-test.log
else
    echo "⚠️  No alarms triggered yet. Checking for INFO/DEBUG logs..."
    if grep -q "INFO:" /tmp/alarm-test.log; then
        echo "❌ FAIL: INFO logs still appearing (should be suppressed at warning level)"
    else
        echo "✅ INFO logs properly suppressed at warning level"
    fi
    
    if grep -q "DEBUG:" /tmp/alarm-test.log; then
        echo "❌ FAIL: DEBUG logs still appearing (should be suppressed at warning level)"
    else
        echo "✅ DEBUG logs properly suppressed at warning level"
    fi
    
    echo ""
    echo "No alarms triggered during test period (weather may be stable)"
fi

echo ""
echo "=========================================="
echo "6. Cleaning up..."
kill $APP_PID 2>/dev/null
sleep 1

echo "✅ Test complete"
echo ""
echo "To manually test alarm output:"
echo "  ./tempest-homekit-go --loglevel warning --alarms @tempest-alarms.json"
echo ""
echo "Log file saved to: /tmp/alarm-test.log"
