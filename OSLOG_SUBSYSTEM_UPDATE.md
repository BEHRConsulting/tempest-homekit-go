# OSLog Subsystem Update

**Date**: October 10, 2025  
**Change**: Updated OSLog subsystem identifier  
**Status**: ✅ **COMPLETE**

---

## Change Summary

Updated the OSLog subsystem identifier from `com.behr.tempest-homekit` to `com.bci.tempest-homekit`.

---

## Files Modified

### Code Changes

**File: `pkg/alarm/notifier_oslog_darwin.go`**
```go
// OLD
subsystem := C.CString("com.behr.tempest-homekit")

// NEW
subsystem := C.CString("com.bci.tempest-homekit")
```

### Documentation Updates

**File: `OSLOG_NOTIFIER.md`**

Updated all references to the subsystem:
- Overview section
- Configuration examples  
- Viewing commands (`log stream`, `log show`)
- Testing instructions
- Export examples

**Updated Commands:**
```bash
# Stream logs (NEW subsystem)
log stream --predicate 'subsystem == "com.bci.tempest-homekit"'

# Show recent logs
log show --predicate 'subsystem == "com.bci.tempest-homekit"' --last 1h --info

# Filter by category
log show --predicate 'subsystem == "com.bci.tempest-homekit" AND category == "alarm"' --last 1h

# Export to file
log show --predicate 'subsystem == "com.bci.tempest-homekit"' --last 24h > alarms.log

Note: The `--read-history`/`READ_HISTORY` option controls whether the service preloads historical observations on startup (it preloads up to `HISTORY_POINTS` observations). The `--chart-history`/`CHART_HISTORY_HOURS` setting controls the time range shown on charts (default 24 hours).
```

---

## Verification

### Build Status
```bash
$ go build
✅ Success - no errors
```

### Alarm Configuration
```
Wind Change alarm configured with oslog channel:
  channels=['console', 'oslog']
```

### Usage Instructions

To view OSLog messages with the new subsystem:

```bash
# Real-time streaming
log stream --predicate 'subsystem == "com.bci.tempest-homekit"'

# Or in Console.app
1. Open Console.app
2. Filter by subsystem: com.bci.tempest-homekit
3. Watch for alarm notifications
```

---

## Backward Compatibility

⚠️ **Note**: This is a breaking change for monitoring scripts.

**Impact:**
- Any existing scripts using the old subsystem will need to be updated
- Logs written before this change used `com.behr.tempest-homekit`
- Logs written after this change use `com.bci.tempest-homekit`

**Migration:**
```bash
# OLD (no longer works)
log stream --predicate 'subsystem == "com.behr.tempest-homekit"'

# NEW (use this going forward)
log stream --predicate 'subsystem == "com.bci.tempest-homekit"'
```

**Viewing Historical Logs:**
```bash
# Old logs (before update)
log show --predicate 'subsystem == "com.behr.tempest-homekit"' --last 7d

# New logs (after update)
log show --predicate 'subsystem == "com.bci.tempest-homekit"' --last 7d

# View both
log show --predicate 'subsystem CONTAINS "tempest-homekit"' --last 7d
```

---

## Testing

### Manual Test

1. Start application with OSLog alarm:
   ```bash
   ./tempest-homekit-go --loglevel warning --alarms @tempest-alarms.json
   ```

2. In another terminal, monitor logs:
   ```bash
   log stream --predicate 'subsystem == "com.bci.tempest-homekit"'
   ```

3. Wait for Wind Change alarm to trigger

4. Verify message appears with new subsystem

---

## Related Documentation

- **[OSLOG_NOTIFIER.md](OSLOG_NOTIFIER.md)** - Complete OSLog documentation (updated)
- **[ALARM_LOGGING.md](ALARM_LOGGING.md)** - Alarm logging behavior
- **[tempest-alarms.json](tempest-alarms.json)** - Alarm configuration file

---

## Summary

✅ Subsystem changed: `com.behr.tempest-homekit` → `com.bci.tempest-homekit`  
✅ Code updated in `notifier_oslog_darwin.go`  
✅ All documentation updated with new commands  
✅ Build successful  
✅ Ready for deployment

**Action Required**: Update any monitoring scripts or Console.app filters to use the new subsystem identifier.
