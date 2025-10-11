# Fix: Alarm Configuration File Path Issue

## Problem

When running:
```bash
./tempest-homekit-go --loglevel info --alarms tempest-alarms.json
```

Error occurred:
```
ERROR: Failed to initialize alarm manager: failed to parse alarm config: invalid character 'e' in literal true (expecting 'r')
ERROR: Continuing without alarms - fix configuration to enable alarm notifications
```

## Root Cause

The `--alarms` flag requires the `@` prefix when specifying a file path. Without it, the system tries to parse the filename itself as inline JSON, which fails.

## Solution

### ✅ Correct Usage (with @ prefix)

```bash
./tempest-homekit-go --loglevel info --alarms @tempest-alarms.json
```

### ❌ Incorrect Usage (missing @)

```bash
./tempest-homekit-go --loglevel info --alarms tempest-alarms.json
```

## Why the @ Prefix?

The `--alarms` flag supports two modes:

1. **File Mode** (`@filename`): Reads alarm configuration from a file
   ```bash
   --alarms @tempest-alarms.json
   --alarms @/path/to/alarms.json
   ```

2. **Inline Mode** (JSON string): Accepts JSON directly as a string
   ```bash
   --alarms '{"alarms":[{"name":"test","enabled":true,"condition":"temp>85","channels":[{"type":"console"}]}]}'
   ```

The `@` prefix tells the system to read the file rather than treating it as inline JSON.

## Enhanced Error Message

The error message now provides a helpful hint when a filename is detected:

```
ERROR: Failed to initialize alarm manager: failed to parse alarm config: invalid character 'e' in literal true (expecting 'r')
Hint: Did you mean to use '@tempest-alarms.json'? File paths must be prefixed with @
```

## Examples

### Start with alarms from file

```bash
./tempest-homekit-go \
  --token <your-token> \
  --station <your-station> \
  --alarms @tempest-alarms.json \
  --loglevel info
```

### Start with debug logging

```bash
./tempest-homekit-go \
  --token <your-token> \
  --station <your-station> \
  --alarms @tempest-alarms.json \
  --loglevel debug
```

### Use environment variable

```bash
export ALARMS=@tempest-alarms.json
./tempest-homekit-go --token <your-token> --station <your-station> --loglevel info
```

### Edit alarms configuration

```bash
./tempest-homekit-go --alarms-edit @tempest-alarms.json --alarms-edit-port 8081
```

## File Watching

When using the file mode with `@`, the alarm system automatically watches the file for changes and reloads when it's modified. This allows you to edit alarms while the application is running.

## Help Text

The help text shows the correct format:

```
ALARM OPTIONS:
  --alarms <config>             Enable alarm system with configuration
                                Format: @filename.json or inline JSON string
                                Env: ALARMS
```

## Summary

- ✅ Always use `@` prefix for file paths: `--alarms @tempest-alarms.json`
- ✅ File paths can be relative or absolute: `--alarms @/full/path/alarms.json`
- ✅ Environment variable also needs `@`: `ALARMS=@tempest-alarms.json`
- ✅ Inline JSON does NOT use `@`: `--alarms '{"alarms":[...]}'`
- ✅ Error message now provides helpful hint when `@` is missing
