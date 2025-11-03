# Alarm Scheduling

The alarm system supports flexible scheduling to control when alarms are active. This allows you to restrict alarms to specific times of day, days of week, or based on sunrise/sunset times.

## Schedule Types

### 1. Always Active (Default)

If no schedule is specified, alarms are active 24/7.

```json
{
  "name": "example-alarm",
  "enabled": true,
  "condition": "temperature > 85",
  "channels": [...]
}
```

Or explicitly:

```json
{
  "name": "example-alarm",
  "enabled": true,
  "condition": "temperature > 85",
  "schedule": {
    "type": "always"
  },
  "channels": [...]
}
```

### 2. Time Range (Daily)

Active during specific hours each day (24-hour format HH:MM).

```json
{
  "name": "daytime-temperature",
  "enabled": true,
  "condition": "temperature > 85",
  "schedule": {
    "type": "daily",
    "start_time": "09:00",
    "end_time": "17:00"
  },
  "channels": [...]
}
```

**Overnight Ranges**: Start and end times can span midnight:

```json
{
  "schedule": {
    "type": "daily",
    "start_time": "22:00",
    "end_time": "06:00"
  }
}
```

This is active from 10 PM until 6 AM the next day.

### 3. Weekly Schedule

Active on specific days of the week (0=Sunday, 1=Monday, ..., 6=Saturday).

```json
{
  "name": "weekday-alert",
  "enabled": true,
  "condition": "temperature > 85",
  "schedule": {
    "type": "weekly",
    "days_of_week": [1, 2, 3, 4, 5]
  },
  "channels": [...]
}
```

**With Time Range**: Combine day restrictions with time ranges:

```json
{
  "schedule": {
    "type": "weekly",
    "days_of_week": [1, 2, 3, 4, 5],
    "start_time": "08:00",
    "end_time": "18:00"
  }
}
```

This is active Monday through Friday, 8 AM to 6 PM.

**Weekend Only**:

```json
{
  "schedule": {
    "type": "weekly",
    "days_of_week": [0, 6]
  }
}
```

### 4. Sunrise/Sunset Based

Active based on sunrise and sunset times for your location.

**After Sunrise (Until Midnight)**:

```json
{
  "schedule": {
    "type": "sun",
    "sun_event": "sunrise",
    "sun_offset": 0
  }
}
```

**Sunrise to Sunset Only**:

```json
{
  "name": "uv-alert-daylight",
  "enabled": true,
  "condition": "uv >= 8",
  "schedule": {
    "type": "sun",
    "sun_event": "sunrise",
    "sun_event_end": "sunset",
    "sun_offset": 0,
    "sun_offset_end": 0
  },
  "channels": [...]
}
```

**With Offsets**: Add offsets in minutes (negative for before, positive for after):

```json
{
  "schedule": {
    "type": "sun",
    "sun_event": "sunrise",
    "sun_event_end": "sunset",
    "sun_offset": -30,
    "sun_offset_end": 30
  }
}
```

This is active from 30 minutes before sunrise until 30 minutes after sunset.

**Location for Sun Calculations**: You can specify latitude/longitude in the schedule:

```json
{
  "schedule": {
    "type": "sun",
    "sun_event": "sunrise",
    "sun_event_end": "sunset",
    "latitude": 34.0522,
    "longitude": -118.2437
  }
}
```

If not specified, the system will use the weather station's location (if available).

## Examples

### Example 1: Business Hours Only

```json
{
  "name": "office-temperature",
  "description": "Alert during business hours",
  "enabled": true,
  "condition": "temperature > 80",
  "schedule": {
    "type": "weekly",
    "days_of_week": [1, 2, 3, 4, 5],
    "start_time": "08:00",
    "end_time": "18:00"
  },
  "cooldown": 3600,
  "channels": [
    {
      "type": "email",
      "email": {
        "to": ["facilities@company.com"],
        "subject": "Office Temperature Alert",
        "body": "Temperature is {{temperature}}°F at {{timestamp}}"
      }
    }
  ]
}
```

### Example 2: UV Alert During Daylight

```json
{
  "name": "high-uv-daylight",
  "description": "UV alert only during sunlight hours",
  "enabled": true,
  "condition": "uv >= 8",
  "schedule": {
    "type": "sun",
    "sun_event": "sunrise",
    "sun_event_end": "sunset",
    "sun_offset": 0,
    "sun_offset_end": 0
  },
  "cooldown": 10800,
  "channels": [
    {
      "type": "sms",
      "sms": {
        "to": ["+15555551234"],
        "message": "⚠️ High UV Index: {{uv}} - Use sun protection!"
      }
    }
  ]
}
```

### Example 3: Night Watch

```json
{
  "name": "overnight-temperature-drop",
  "description": "Alert if temperature drops too low overnight",
  "enabled": true,
  "condition": "temperature < 35",
  "schedule": {
    "type": "daily",
    "start_time": "22:00",
    "end_time": "06:00"
  },
  "cooldown": 1800,
  "channels": [
    {
      "type": "console",
      "template": "Tempest-Alarm [overnight-temperature-drop]: Temperature dropped to {{temperature}}°F at {{timestamp}}"
    }
  ]
}
```

### Example 4: Weekend Storm Watch

```json
{
  "name": "weekend-storm",
  "description": "Lightning alerts on weekends",
  "enabled": true,
  "condition": "lightning_distance < 5",
  "schedule": {
    "type": "weekly",
    "days_of_week": [0, 6]
  },
  "cooldown": 600,
  "channels": [
    {
      "type": "sms",
      "sms": {
        "to": ["+15555551234"],
        "message": "⚡ Weekend Storm Alert: Lightning {{lightning_distance}}mi away"
      }
    }
  ]
}
```

## How It Works

1. **Evaluation**: Before checking alarm conditions, the system evaluates the schedule
2. **Active Check**: If the current time doesn't match the schedule, the alarm is skipped
3. **Debug Logging**: When `--loglevel debug` is set, schedule checks are logged:
   ```
   DEBUG: Alarm high-temperature outside scheduled time: Daily from 09:00 to 17:00
   ```

## Configuration Location

The schedule is set on a per-alarm basis by including a `schedule` object in the alarm's JSON configuration:

```json
{
  "alarms": [
    {
      "name": "my-alarm",
      "enabled": true,
      "condition": "...",
      "schedule": {
        "type": "daily",
        "start_time": "09:00",
        "end_time": "17:00"
      },
      "channels": [...]
    }
  ]
}
```

## Schedule Fields Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | `"always"`, `"time"`, `"daily"`, `"weekly"`, or `"sun"` |
| `start_time` | string | For time/daily/weekly | HH:MM format (24-hour) |
| `end_time` | string | For time/daily/weekly | HH:MM format (24-hour) |
| `days_of_week` | array | For weekly | Array of integers 0-6 (0=Sunday, 6=Saturday) |
| `sun_event` | string | For sun | `"sunrise"` or `"sunset"` |
| `sun_event_end` | string | Optional for sun | `"sunrise"` or `"sunset"` (defines range end) |
| `sun_offset` | integer | Optional for sun | Minutes offset from `sun_event` (negative=before, positive=after) |
| `sun_offset_end` | integer | Optional for sun | Minutes offset from `sun_event_end` |
| `latitude` | float | Optional for sun | Override latitude for sun calculations |
| `longitude` | float | Optional for sun | Override longitude for sun calculations |

## Validation

Schedules are validated when the alarm configuration is loaded:

- Time formats must be `HH:MM` (24-hour)
- Days of week must be 0-6
- Sun events must be `"sunrise"` or `"sunset"`
- Required fields must be present for each schedule type

Invalid schedules will cause the alarm configuration to fail loading with a descriptive error.

## Testing

You can test schedules by temporarily adjusting the system time or by using the `--test-alarm` flag with a specific time (if implemented).

To see schedule evaluation in action:

```bash
./tempest-homekit-go --alarms @alarms.json --loglevel debug
```

This will log schedule checks as they occur.

## Timezone Considerations

- All times use the local system timezone
- Sunrise/sunset calculations use the provided latitude/longitude
- If using sunrise/sunset without lat/lon, the station's location is used
- Times are compared against the current local time, not UTC

## Performance

Schedule evaluation is very fast:
- Time range checks: Simple integer comparison
- Weekly checks: Array lookup + time check
- Sun calculations: Cached per day (recalculated at midnight)

There is no performance impact from using schedules.
