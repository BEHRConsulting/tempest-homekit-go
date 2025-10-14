#!/bin/bash
# Test script for alarm JSON validation with timeout

echo "Testing Alarm JSON Validation"
echo "=============================="
echo ""

# Test 1: Missing @ prefix (filename without @)
echo "Test 1: Missing @ prefix (should show hint)"
timeout 2s ./tempest-homekit-go --loglevel error --alarms tempest-alarms.json 2>&1 | grep -A3 "ERROR.*alarm" | head -5
echo ""

# Test 2: Invalid JSON syntax (missing closing brace)
echo "Test 2: Invalid JSON - missing closing brace"
timeout 2s ./tempest-homekit-go --loglevel error --alarms '{"alarms":[{"name":"test"' 2>&1 | grep -A3 "ERROR.*alarm" | head -5
echo ""

# Test 3: Invalid JSON with typo (missing comma)
echo "Test 3: Invalid JSON - missing comma"
timeout 2s ./tempest-homekit-go --loglevel error --alarms '{"alarms":[{"name":"test" "enabled":true}]}' 2>&1 | grep -A3 "ERROR.*alarm" | head -5
echo ""

# Test 4: Valid JSON but wrong structure
echo "Test 4: Valid JSON but wrong structure (missing alarms array)"
timeout 2s ./tempest-homekit-go --loglevel error --alarms '{"wrong":"field"}' 2>&1 | grep -A3 "ERROR.*alarm" | head -5
echo ""

# Test 5: Valid JSON with correct structure but invalid alarm (missing condition)
echo "Test 5: Valid structure but missing required field (condition)"
timeout 2s ./tempest-homekit-go --loglevel error --alarms '{"alarms":[{"name":"test","enabled":true,"channels":[]}]} ' 2>&1 | grep -A3 "ERROR.*alarm" | head -5
echo ""

# Test 6: Valid JSON with correct structure but invalid alarm (missing channels)
echo "Test 6: Valid structure but missing required field (channels)"
timeout 2s ./tempest-homekit-go --loglevel error --alarms '{"alarms":[{"name":"test","enabled":true,"condition":"temp>85"}]}' 2>&1 | grep -A3 "ERROR.*alarm" | head -5
echo ""

# Test 7: Correct usage with @ prefix
echo "Test 7: Correct usage with @ prefix (should succeed)"
timeout 2s ./tempest-homekit-go --loglevel info --alarms @tempest-alarms.json 2>&1 | grep -E "(Loaded alarm|alarms are enabled)" | head -5
echo ""

echo "=============================="
echo "All tests completed"
