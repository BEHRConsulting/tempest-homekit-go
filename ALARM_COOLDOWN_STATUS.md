# Alarm Cooldown Status Display

**Date**: October 10, 2025  
**Feature**: Web console alarm cooldown status  
**Status**: ✅ **IMPLEMENTED**

---

## Overview

The web console now displays real-time cooldown status for each alarm, showing whether an alarm is ready to fire or currently in its cooldown period. This provides immediate visibility into alarm availability and helps users understand why an alarm may not be triggering despite meeting its condition.

---

## Features

### API Enhancements

#### New Alarm Type Methods (`pkg/alarm/types.go`)

```go
// GetCooldownRemaining returns the remaining cooldown time in seconds (0 if can fire)
func (a *Alarm) GetCooldownRemaining() int

// IsInCooldown returns true if the alarm is currently in cooldown
func (a *Alarm) IsInCooldown() bool
```

#### Updated API Response (`pkg/web/server.go`)

The `/api/alarm-status` endpoint now includes cooldown information for each alarm:

```json
{
  "enabled": true,
  "configPath": "tempest-alarms.json",
  "totalAlarms": 4,
  "enabledAlarms": 4,
  "alarms": [
    {
      "name": "Test Cooldown",
      "description": "Test alarm with 120 second cooldown",
      "enabled": true,
      "condition": "*wind_speed",
      "cooldown": 120,
      "cooldownRemaining": 104,
      "inCooldown": true,
      "lastTriggered": "2025-10-10 21:03:19",
      "channels": ["console"]
    }
  ]
}
```

**New Fields:**
- `cooldownRemaining` (int): Seconds remaining in cooldown (0 if ready to fire)
- `inCooldown` (bool): True if alarm is currently in cooldown period

### Web Console Display

The alarm status card in the web dashboard now shows:

#### When Alarm is Ready (Not in Cooldown)
```
✓ Ready (cooldown: 120s)
```
- Green text color
- Shows configured cooldown duration
- Indicates alarm can fire immediately

#### When Alarm is in Cooldown
```
⏳ Cooldown: 1m 44s remaining
```
- Orange/warning text color
- Shows time remaining in human-readable format (minutes and seconds)
- Updates in real-time every 10 seconds

---

## Implementation Details

### Backend Changes

**File: `pkg/alarm/types.go`**

Added methods to calculate and check cooldown status:

```go
func (a *Alarm) GetCooldownRemaining() int {
    if !a.Enabled || a.Cooldown == 0 {
        return 0
    }
    elapsed := time.Since(a.lastFired)
    cooldownDuration := time.Duration(a.Cooldown) * time.Second
    if elapsed >= cooldownDuration {
        return 0
    }
    remaining := cooldownDuration - elapsed
    return int(remaining.Seconds())
}

func (a *Alarm) IsInCooldown() bool {
    return a.GetCooldownRemaining() > 0
}
```

**File: `pkg/web/server.go`**

Updated `AlarmStatus` struct:
```go
type AlarmStatus struct {
    Name              string   `json:"name"`
    Description       string   `json:"description"`
    Enabled           bool     `json:"enabled"`
    Condition         string   `json:"condition"`
    Tags              []string `json:"tags"`
    Channels          []string `json:"channels"`
    LastTriggered     string   `json:"lastTriggered"`
    Cooldown          int      `json:"cooldown"`
    CooldownRemaining int      `json:"cooldownRemaining"` // NEW
    InCooldown        bool     `json:"inCooldown"`        // NEW
}
```

Updated API handler to populate cooldown status:
```go
cooldownRemaining := alm.GetCooldownRemaining()
inCooldown := alm.IsInCooldown()

alarmStatuses = append(alarmStatuses, AlarmStatus{
    // ... other fields ...
    CooldownRemaining: cooldownRemaining,
    InCooldown:        inCooldown,
})
```

### Frontend Changes

**File: `pkg/web/static/script.js`**

Enhanced `updateAlarmStatus()` function to display cooldown information:

```javascript
// Add cooldown status if applicable
const cooldown = document.createElement('div');
cooldown.className = 'alarm-item-cooldown';
if (alarm.inCooldown) {
    const minutes = Math.floor(alarm.cooldownRemaining / 60);
    const seconds = alarm.cooldownRemaining % 60;
    const timeStr = minutes > 0 ? `${minutes}m ${seconds}s` : `${seconds}s`;
    cooldown.textContent = `⏳ Cooldown: ${timeStr} remaining`;
    cooldown.style.color = 'var(--warning-color, #ff9800)';
} else {
    cooldown.textContent = `✓ Ready (cooldown: ${alarm.cooldown}s)`;
    cooldown.style.color = 'var(--success-color, #4caf50)';
}
```

---

## User Benefits

1. **Visibility**: Users can see at a glance which alarms are ready to fire
2. **Troubleshooting**: If an alarm isn't triggering, users can check if it's in cooldown
3. **Real-time Updates**: Cooldown countdown refreshes every 10 seconds automatically
4. **No Config Changes**: Existing alarm configurations work without modification

---

## Example Scenarios

### Scenario 1: Alarm Just Triggered

An alarm with a 30-minute (1800s) cooldown just fired:

**API Response:**
```json
{
  "name": "Hot outside",
  "cooldown": 1800,
  "cooldownRemaining": 1795,
  "inCooldown": true,
  "lastTriggered": "2025-10-10 21:00:00"
}
```

**Web Display:**
```
🔔 Hot outside
Condition: temp > 85
Last: 2025-10-10 21:00:00
Channels: console, syslog
⏳ Cooldown: 29m 55s remaining
```

### Scenario 2: Alarm Ready to Fire

An alarm that hasn't fired recently or has completed its cooldown:

**API Response:**
```json
{
  "name": "Wind Change",
  "cooldown": 10,
  "cooldownRemaining": 0,
  "inCooldown": false,
  "lastTriggered": "2025-10-10 20:55:00"
}
```

**Web Display:**
```
🔔 Wind Change
Condition: *wind_speed
Last: 2025-10-10 20:55:00
Channels: console, oslog
✓ Ready (cooldown: 10s)
```

### Scenario 3: Alarm Never Triggered

A newly configured alarm that hasn't fired yet:

**API Response:**
```json
{
  "name": "Lightning Nearby",
  "cooldown": 1800,
  "cooldownRemaining": 0,
  "inCooldown": false,
  "lastTriggered": "Never"
}
```

**Web Display:**
```
🔔 Lightning Nearby
Condition: *lightning_count
Last: Never
Channels: console, syslog
✓ Ready (cooldown: 1800s)
```

---

## Testing

### Manual Testing

```bash
# Start app with alarms
./tempest-homekit-go --loglevel warning --alarms @tempest-alarms.json

# Open web console
open http://localhost:8080

# Monitor alarm status card - cooldown status updates every 10 seconds
```

### API Testing

```bash
# Check cooldown status via API
curl -s http://localhost:8080/api/alarm-status | python3 -m json.tool

# Monitor specific alarm cooldown
curl -s http://localhost:8080/api/alarm-status | \
  python3 -c "import sys, json; \
  data=json.load(sys.stdin); \
  alarm=[a for a in data['alarms'] if a['name']=='Wind Change'][0]; \
  print(f\"Cooldown: {alarm['cooldownRemaining']}s remaining, In cooldown: {alarm['inCooldown']}\")"
```

### Test with Long Cooldown

Create a test alarm with a long cooldown to observe the countdown:

```json
{
  "alarms": [
    {
      "name": "Test Cooldown",
      "description": "Test alarm with 2 minute cooldown",
      "enabled": true,
      "condition": "*wind_speed",
      "cooldown": 120,
      "channels": [
        {
          "type": "console",
          "template": "Test alarm triggered"
        }
      ]
    }
  ]
}
```

Monitor the cooldown countdown:
```bash
# After alarm triggers
while true; do
  curl -s http://localhost:8080/api/alarm-status | \
    python3 -c "import sys, json; \
    data=json.load(sys.stdin); \
    alarm=[a for a in data['alarms'] if a['name']=='Test Cooldown'][0]; \
    print(f\"Remaining: {alarm['cooldownRemaining']}s\")"
  sleep 5
done
```

---

## Visual Design

### Ready State (Green)
```
✓ Ready (cooldown: 10s)
```
- Color: `var(--success-color, #4caf50)` (green)
- Icon: ✓ (checkmark)
- Message: Shows configured cooldown duration

### Cooldown State (Orange)
```
⏳ Cooldown: 1m 44s remaining
```
- Color: `var(--warning-color, #ff9800)` (orange)
- Icon: ⏳ (hourglass)
- Message: Shows time remaining with automatic formatting

### Time Formatting
- **Less than 1 minute**: `"45s"`
- **1+ minutes**: `"2m 15s"`
- **Hours not shown**: Cooldowns are typically short (< 60 minutes)

---

## Related Documentation

- [Alarm System Overview](pkg/alarm/README.md)
- [Web Console Alarm Status Card](WEB_ALARM_STATUS_CARD.md)
- [Alarm Configuration Guide](ALARM_EDITOR_MESSAGES.md)
- [Change Detection Operators](CHANGE_DETECTION_OPERATORS.md)

---

## Future Enhancements

Potential future improvements:

1. **Visual Progress Bar**: Show cooldown progress graphically
2. **Cooldown Notifications**: Alert when alarm exits cooldown
3. **Manual Cooldown Reset**: Admin button to clear cooldown
4. **Cooldown History**: Track cooldown patterns over time
5. **Predicted Next Fire**: Estimate when alarm will next trigger

---

## Conclusion

The cooldown status display provides critical visibility into alarm availability, helping users:
- ✅ Understand why alarms may not be triggering
- ✅ Monitor alarm readiness in real-time
- ✅ Troubleshoot alarm configuration issues
- ✅ Plan around cooldown periods

The feature integrates seamlessly with the existing alarm system and requires no configuration changes to existing alarm definitions.
