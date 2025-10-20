# Enhanced Alarm JSON Validation

## Summary

Enhanced the alarm configuration validation to provide clear, actionable error messages when invalid JSON is provided via the `--alarms` flag.

## Changes Made

### File: `pkg/alarm/types.go`

**Enhanced `LoadAlarmConfig()` function:**

1. **Pre-validation of JSON syntax** - Tests if the string is valid JSON before attempting to parse into AlarmConfig structure
2. **Detailed syntax errors** - Provides line and column numbers for JSON syntax errors
3. **Helpful hints** - Detects common mistakes and provides suggestions
4. **Separate error messages** - Different messages for file vs inline JSON errors

## Error Messages

### 1. Missing @ Prefix (Forgot to specify file)

**Command:**
```bash
./tempest-homekit-go --alarms tempest-alarms.json
```

**Error:**
```
ERROR: Failed to initialize alarm manager: invalid JSON string: invalid character 'e' in literal true (expecting 'r')
Hint: Did you mean to use '@tempest-alarms.json'? File paths must be prefixed with @
```

### 2. Invalid JSON Syntax - Missing Closing Brace

**Command:**
```bash
./tempest-homekit-go --alarms '{"alarms":[{"name":"test"'
```

**Error:**
```
ERROR: Failed to initialize alarm manager: invalid JSON syntax at line 1, column 26: unexpected end of JSON input
Provide valid JSON string or use @filename.json to load from file
```

### 3. Invalid JSON Syntax - Missing Comma

**Command:**
```bash
./tempest-homekit-go --alarms '{"alarms":[{"name":"test" "enabled":true}]}'
```

**Error:**
```
ERROR: Failed to initialize alarm manager: invalid JSON syntax at line 1, column 28: invalid character '"' after object key:value pair
Provide valid JSON string or use @filename.json to load from file
```

### 4. Valid JSON But Wrong Structure

**Command:**
```bash
./tempest-homekit-go --alarms '{"wrong":"field"}'
```

**Error:**
```
ERROR: Failed to initialize alarm manager: invalid alarm config: at least one alarm must be defined
```

### 5. Valid Structure But Missing Required Fields

**Command:**
```bash
./tempest-homekit-go --alarms '{"alarms":[{"name":"test","enabled":true,"channels":[]}]}'
```

**Error:**
```
ERROR: Failed to initialize alarm manager: invalid alarm config: alarm test: condition is required
```

**Command:**
```bash
./tempest-homekit-go --alarms '{"alarms":[{"name":"test","enabled":true,"condition":"temp>85"}]}'
```

**Error:**
```
ERROR: Failed to initialize alarm manager: invalid alarm config: alarm test: at least one channel is required
```

### 6. Correct Usage

**Command:**
```bash
./tempest-homekit-go --alarms @tempest-alarms.json
```

**Success:**
```
INFO: Loaded alarm: Hot outside
INFO: Loaded alarm: Lightning Nearby
INFO: Loaded alarm: Lux Change
INFO: 3 of 3 alarms are enabled
```

## Test Script

Created `test-alarm-validation.sh` to test all error scenarios with automatic timeout:

```bash
#!/bin/bash
# Runs all validation tests with 2-second timeout per test
./test-alarm-validation.sh
```

**Features:**
- Uses `timeout 2s` to prevent hanging
- Tests 7 different error scenarios
- Shows clear output for each test
- Verifies both error cases and success case

## Validation Flow

```
Input String
 ↓
[Starts with @?]
 ↓ ↓
 YES NO
 ↓ ↓
Read File Test JSON Syntax
 ↓ ↓
 ├→ File Error [Valid JSON?]
 ↓ ↓
Parse JSON YES NO
 ↓ ↓ ↓
 ├→ Parse Error Parse Syntax Error
 ↓ JSON with line/col
Validate Config ↓ ↓
 ↓ ↓ [Looks like
 ├→ Validation Validate filename?]
 ↓ Error Config ↓
Success ↓ YES NO
 ↓ ↓ ↓
 Success Hint Generic
 @file error
```

## Error Message Categories

| Category | Detection | Error Message |
|----------|-----------|---------------|
| **Missing @** | No `{` start + ends with `.json` | "Hint: Did you mean to use '@filename'?" |
| **JSON Syntax Error** | json.Unmarshal fails on any JSON | "invalid JSON syntax at line X, column Y: ..." |
| **Structure Error** | Valid JSON, wrong structure | "invalid alarm config: at least one alarm must be defined" |
| **Validation Error** | Valid structure, missing fields | "alarm X: field Y is required" |
| **File Read Error** | File with @ prefix doesn't exist | "failed to read alarm config file: ..." |

## Benefits

1. **Clear Error Location**: Line and column numbers for JSON syntax errors
2. **Actionable Hints**: Suggests using @ prefix when appropriate
3. **Early Detection**: Validates JSON syntax before parsing structure
4. **Better UX**: Users know exactly what's wrong and how to fix it
5. **Separate Contexts**: Different messages for file vs inline JSON

## Usage Examples

### Correct - File Reference
```bash
./tempest-homekit-go --alarms @tempest-alarms.json
./tempest-homekit-go --alarms @/full/path/alarms.json
```

### Correct - Inline JSON
```bash
./tempest-homekit-go --alarms '{"alarms":[{"name":"test","enabled":true,"condition":"temp>85","channels":[{"type":"console","template":"Alert!"}]}]}'
```

### Incorrect - Missing @
```bash
./tempest-homekit-go --alarms tempest-alarms.json
# Error: Hint: Did you mean to use '@tempest-alarms.json'?
```

### Incorrect - Invalid JSON
```bash
./tempest-homekit-go --alarms '{"alarms":['
# Error: invalid JSON syntax at line 1, column 12: unexpected end of JSON input
```

### Incorrect - Wrong Structure
```bash
./tempest-homekit-go --alarms '{}'
# Error: invalid alarm config: at least one alarm must be defined
```

## Testing

Run the test suite:
```bash
./test-alarm-validation.sh
```

Each test runs for maximum 2 seconds and shows the error message for verification.

## Files Modified

1. **pkg/alarm/types.go** - Enhanced `LoadAlarmConfig()` with pre-validation and detailed error messages
2. **test-alarm-validation.sh** - Automated test script with timeout for all error scenarios

## Documentation

- **ALARM_FILE_PATH_FIX.md** - Explains @ prefix requirement
- **ALARM_DEBUG_LOGGING.md** - Shows debug output examples
- **THIS FILE** - Comprehensive validation error message reference
