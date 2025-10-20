# Alarm Debug Logging

This document shows the enhanced logging output for alarm system debugging.

## Log Level: INFO

At `--loglevel info`, you'll see alarm names as they're loaded:

```
INFO: Alarm manager initialized with 3 alarms
INFO: Loaded alarm: Hot outside
INFO: Loaded alarm: Lightning Nearby
INFO: Loaded alarm: Lux Change
INFO: 3 of 3 alarms are enabled
```

When an alarm triggers:

```
INFO: ALARM: Lux Change (condition: *lux)
INFO: Sent console notification for alarm Lux Change
```

## Log Level: DEBUG

At `--loglevel debug`, you'll see comprehensive details:

### 1. Startup - Pretty-Formatted JSON

When the alarm manager initializes, all alarms are output as pretty JSON:

```
INFO: Alarm manager initialized with 3 alarms
INFO: Loaded alarm: Hot outside
DEBUG: Condition: temp > 85
DEBUG: Description: Set when temp is > 85F
DEBUG: Cooldown: 1800s
DEBUG: Channels: 2
INFO: Loaded alarm: Lightning Nearby
DEBUG: Condition: *lightning_count
DEBUG: Description: Let me know when lightning strikes nearby
DEBUG: Cooldown: 1800s
DEBUG: Channels: 2
INFO: Loaded alarm: Lux Change
DEBUG: Condition: *lux
DEBUG: Description: This alarm should alert on LUX change
DEBUG: Cooldown: 1800s
DEBUG: Channels: 1
INFO: 3 of 3 alarms are enabled
DEBUG: Alarm configuration JSON:
[
 {
 "name": "Hot outside",
 "description": "Set when temp is > 85F",
 "tags": [
 "hot"
 ],
 "enabled": true,
 "condition": "temp > 85",
 "cooldown": 1800,
 "channels": [
 {
 "type": "console",
 "template": "ALARM: Hot outside\nCondition: {{.Condition}}\nValue: {{.Value}}\nTime: {{.Time}}"
 },
 {
 "type": "syslog",
 "template": "ALARM: Hot outside\nCondition: {{.Condition}}\nValue: {{.Value}}\nTime: {{.Time}}"
 }
 ]
 },
 {
 "name": "Lightning Nearby",
 "description": "Let me know when lightning strikes nearby",
 "enabled": true,
 "condition": "*lightning_count",
 "cooldown": 1800,
 "channels": [
 {
 "type": "console",
 "template": "ALARM: Lightning Nearby\nCondition: {{.Condition}}\nValue: {{.Value}}\nTime: {{.Time}}"
 },
 {
 "type": "syslog",
 "template": "ALARM: Lightning Nearby\nCondition: {{.Condition}}\nValue: {{.Value}}\nTime: {{.Time}}"
 }
 ]
 },
 {
 "name": "Lux Change",
 "description": "This alarm should alert on LUX change",
 "tags": [
 "light"
 ],
 "enabled": true,
 "condition": "*lux",
 "cooldown": 1800,
 "channels": [
 {
 "type": "console",
 "template": "ALARM: {{alarm_name}}\nStation: {{station}}\nTime: {{timestamp}}\nLUX: {{lux}}"
 }
 ]
 }
]
```

### 2. Each Observation - Alarm Evaluation

For each weather observation received, debug logs show:

```
DEBUG: Evaluating alarm: 'Lux Change'
DEBUG: Condition: *lux
DEBUG: Current observation: temp=31.6°C, humidity=51%, pressure=1013.25, wind=0.9m/s, lux=36439, uv=2
DEBUG: No previous value for lux, establishing baseline: 36439.00
DEBUG: Result: false
```

On the next observation with a change:

```
DEBUG: Evaluating alarm: 'Lux Change'
DEBUG: Condition: *lux
DEBUG: Current observation: temp=31.6°C, humidity=51%, pressure=1013.25, wind=1.0m/s, lux=54951, uv=2
DEBUG: Change detected in lux: 36439.00 -> 54951.00
DEBUG: Result: true
INFO: ALARM: Lux Change (condition: *lux)
INFO: Sent console notification for alarm Lux Change
```

### 3. Comparison Operations

For regular comparison alarms (not change detection), each comparison is logged:

```
DEBUG: Evaluating alarm: 'Hot outside'
DEBUG: Condition: temp > 85
DEBUG: Current observation: temp=31.6°C, humidity=51%, pressure=1013.25, wind=0.9m/s, lux=36439, uv=2
DEBUG: Comparison: 31.60 > 85.00 = false
DEBUG: Result: false
```

When the condition is met:

```
DEBUG: Evaluating alarm: 'Hot outside'
DEBUG: Condition: temp > 85
DEBUG: Current observation: temp=29.4°C, humidity=48%, pressure=1013.50, wind=1.2m/s, lux=42000, uv=3
DEBUG: Comparison: 85.52 > 85.00 = true
DEBUG: Result: true
INFO: ALARM: Hot outside (condition: temp > 85)
INFO: Sent console notification for alarm Hot outside
```

### 4. Compound Conditions

For complex conditions with multiple comparisons:

```
DEBUG: Evaluating alarm: 'Hot and Humid'
DEBUG: Condition: temperature > 30 && humidity > 70
DEBUG: Current observation: temp=32.1°C, humidity=75%, pressure=1012.80, wind=0.5m/s, lux=28000, uv=4
DEBUG: Comparison: 32.10 > 30.00 = true
DEBUG: Comparison: 75.00 > 70.00 = true
DEBUG: Result: true
INFO: ALARM: Hot and Humid (condition: temperature > 30 && humidity > 70)
```

### 5. Cooldown Period

When an alarm is in cooldown:

```
DEBUG: Alarm Lux Change in cooldown, skipping (last fired: 2025-10-10 16:05:30 -0700 PDT)
```

### 6. Disabled Alarms

Disabled alarms are mentioned during evaluation:

```
DEBUG: Skipping disabled alarm: Test Alarm
```

## Summary

**--loglevel info:**
- Shows alarm names when loaded
- Shows when alarms trigger with 'ALARM' prefix
- Clean, essential information

**--loglevel debug:**
- Pretty-formatted JSON of entire alarm configuration at startup
- Detailed evaluation for each alarm on each observation
- Current observation values for context
- All comparison operations with results
- Change detection state (baseline establishment, changes detected)
- Cooldown status
- Full visibility into alarm system operation

## Usage Examples

**Production (minimal logging):**
```bash
./tempest-homekit-go --token <token> --station <station> --alarms tempest-alarms.json --loglevel info
```

**Development/Troubleshooting:**
```bash
./tempest-homekit-go --token <token> --station <station> --alarms tempest-alarms.json --loglevel debug
```

**Debug with log filtering (only alarm-related logs):**
```bash
./tempest-homekit-go --token <token> --station <station> --alarms tempest-alarms.json --loglevel debug --logfilter "alarm"
```
