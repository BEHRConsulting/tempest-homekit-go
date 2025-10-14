# Enhanced Alarm Config Reload Logging

## Enhancement

When the alarm configuration file is modified and automatically reloaded by the file watcher, the system now displays the same detailed logging information as the initial load.

## Before

Previously, reloading only showed minimal feedback:

```
INFO: Alarm config file changed, reloading: tempest-alarms.json
INFO: Alarm config reloaded successfully
```

Users couldn't see which alarms were loaded or their enabled/disabled status without restarting the application.

## After

Now, reloading shows complete alarm information:

```
INFO: Alarm config file changed, reloading: tempest-alarms.json
INFO: Alarm manager initialized with 3 alarms
INFO: Loaded alarm: Hot outside
INFO: Loaded alarm: Lightning Nearby
INFO: Loaded alarm: Lux Change
INFO: 3 of 3 alarms are enabled
INFO: Alarm config reloaded successfully
```

### With Debug Level

When running with `--loglevel debug`, additional details are shown for each alarm:

```
DEBUG:   Condition: temp > 85
DEBUG:   Description: Set when temp is > 85F
DEBUG:   Cooldown: 1800s
DEBUG:   Channels: 2
```

Plus the full JSON configuration of all alarms.

## Benefits

1. **Visibility**: Immediately see which alarms were reloaded and their status
2. **Verification**: Confirm changes took effect without checking the file
3. **Troubleshooting**: Identify if an alarm is disabled or misconfigured
4. **Consistency**: Same logging format as initial load for familiar UX
5. **Confidence**: Clear confirmation that file watching is working

## Usage

### Normal Operation

```bash
# Start with alarm file watching enabled
./tempest-homekit-go --alarms @tempest-alarms.json --loglevel info

# Edit tempest-alarms.json in another terminal/editor
# Save the file
# Watch the console for automatic reload with detailed output
```

### Debug Mode

```bash
# Start with debug logging for maximum detail
./tempest-homekit-go --alarms @tempest-alarms.json --loglevel debug

# Edit and save alarm config
# See full JSON output and per-alarm debug details
```

## File Watching Behavior

### Triggers Reload
- **Write operations**: Saving file in editor (vim, nano, VS Code, etc.)
- **Create operations**: Overwriting file with `mv` or `cp`
- **Modification**: Changes via `sed -i`, scripts, or applications

### Does NOT Trigger Reload
- **Touch only**: `touch filename` without actual modification (no content change)
- **Permissions**: `chmod` changes
- **Ownership**: `chown` changes

**Note**: `touch` may not trigger reload on all systems because it only updates timestamps without generating a WRITE event. To test reload, make an actual file modification.

## Example Scenarios

### Scenario 1: Adding a New Alarm

**Edit**: Add "Rain Alert" alarm to `tempest-alarms.json` and save

**Console Output**:
```
INFO: Alarm config file changed, reloading: tempest-alarms.json
INFO: Alarm manager initialized with 4 alarms
INFO: Loaded alarm: Hot outside
INFO: Loaded alarm: Lightning Nearby
INFO: Loaded alarm: Lux Change
INFO: Loaded alarm: Rain Alert
INFO: 4 of 4 alarms are enabled
INFO: Alarm config reloaded successfully
```

### Scenario 2: Disabling an Alarm

**Edit**: Change `"enabled": true` to `"enabled": false` for "Lux Change"

**Console Output**:
```
INFO: Alarm config file changed, reloading: tempest-alarms.json
INFO: Alarm manager initialized with 3 alarms
INFO: Loaded alarm: Hot outside
INFO: Loaded alarm: Lightning Nearby
INFO: Loaded alarm (disabled): Lux Change
INFO: 2 of 3 alarms are enabled
INFO: Alarm config reloaded successfully
```

### Scenario 3: Invalid Configuration

**Edit**: Introduce JSON syntax error or validation error

**Console Output**:
```
INFO: Alarm config file changed, reloading: tempest-alarms.json
ERROR: Failed to reload alarm config: invalid JSON syntax at line 15, column 5: unexpected end of JSON input
```

The previous valid configuration remains active. The system is resilient to bad reloads.

## Testing

### Automated Test

```bash
# Run the reload logging test
./test-alarm-reload-modified.sh
```

This test:
1. Starts the application with alarm file watching
2. Waits for initialization
3. Modifies the alarm config file
4. Verifies detailed reload logging appears
5. Restores original configuration
6. Shows before/after output comparison

### Manual Testing

```bash
# Terminal 1: Start application
./tempest-homekit-go --alarms @tempest-alarms.json --loglevel info

# Terminal 2: Modify alarm config
vim tempest-alarms.json  # Make a change and save

# Terminal 1: Watch for automatic reload with detailed output
```

## Implementation Details

### File Modified: `pkg/alarm/manager.go`

**Function**: `reloadConfig()`

**Enhancement**: Added same logging logic as `NewManager()`:
- Count and log all alarms
- Log each alarm name (with enabled/disabled status)
- Log debug details per alarm (condition, description, cooldown, channels)
- Log enabled alarm count summary
- Output pretty-formatted JSON at debug level

**Behavior**: After successfully parsing and validating the new configuration, the detailed alarm information is logged before returning.

**Thread Safety**: Logging happens after releasing the mutex, so it doesn't block alarm evaluation.

## Benefits Over Initial Implementation

1. **Immediate Feedback**: No need to restart to see what changed
2. **Audit Trail**: Clear record of configuration changes in logs
3. **Error Detection**: Quickly spot if an alarm didn't load as expected
4. **Development UX**: Faster iteration when developing/testing alarms
5. **Production Monitoring**: Log aggregation systems can track alarm changes

## Related Features

- **File Watching**: Automatic detection of config file changes (fsnotify)
- **Validation**: Configuration errors prevent reload (original config remains)
- **Debug Logging**: Enhanced details at debug level
- **Alarm Editor**: Web UI for alarm configuration (`--alarms-edit`)

## Cross-Platform Support

File watching works on:
- ✅ macOS (FSEvents via fsnotify)
- ✅ Linux (inotify via fsnotify)
- ✅ Windows (ReadDirectoryChangesW via fsnotify)

The logging enhancement works identically on all platforms.
