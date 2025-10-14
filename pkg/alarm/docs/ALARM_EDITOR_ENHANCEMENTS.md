# Alarm Editor Enhancements

## New Features (v1.6.0)

### 1. Quick-Insert Sensor Fields

The alarm editor now includes clickable sensor field buttons that make writing conditions faster and error-free.

**Available Sensor Fields:**
- `temperature` - Air temperature (°C)
- `humidity` - Relative humidity (%)
- `pressure` - Station pressure (hPa)
- `wind_speed` - Average wind speed (m/s)
- `wind_gust` - Wind gust speed (m/s)
- `wind_direction` - Wind direction (degrees)
- `lux` - Light intensity (lux)
- `uv` - UV index
- `rain_rate` - Rain rate (mm/hr)
- `rain_daily` - Daily accumulated rain (mm)
- `lightning_count` - Lightning strike count
- `lightning_distance` - Lightning distance (km)

**How to Use:**
1. Click on any sensor field button above the condition text area
2. The field name is inserted at your cursor position
3. Add comparison operators (>, <, >=, <=, ==, !=) and values
4. Combine multiple conditions with `&&` (AND) or `||` (OR)

**Example Workflow:**
1. Click `temperature`
2. Type ` > 30 && `
3. Click `humidity`
4. Type ` > 70`
5. Result: `temperature > 30 && humidity > 70`

### 2. Delivery Method Checkboxes

Select multiple notification channels with intuitive checkboxes. No more manual JSON editing!

**Available Delivery Methods:**
- 📟 **Console** - Log to console (default, always recommended)
- 📋 **Syslog** - System log integration (Unix/Linux)
- 📊 **Event Log** - Windows Event Log integration
- ✉️ **Email** - Email notifications (requires SMTP configuration)
- 📱 **SMS** - SMS notifications (requires Twilio configuration)

**Default Behavior:**
- New alarms default to Console delivery only
- At least one delivery method must be selected
- Multiple methods can be selected simultaneously

**Template Generation:**
All delivery methods use a default template:
```
ALARM: {{alarm_name}} - {{condition}} (Station: {{station}}, Temp: {{temperature}}°C, Humidity: {{humidity}}%)
```

For Email and SMS channels, additional configuration is required in the JSON (the editor preserves existing configurations when editing).

### 3. Enhanced Debug Logging

Comprehensive logging helps troubleshoot alarm behavior and monitor system activity.

**Debug Level Logs (--log-level=debug):**
- Every alarm evaluation attempt with current sensor values
- Condition parsing for compound (AND/OR) conditions
- Individual part evaluation in compound conditions
- Disabled alarm skip notifications
- Cooldown period status checks

**Example Debug Output:**
```
DEBUG: Testing alarm: high-temperature - temperature > 29.4
DEBUG: Evaluating condition: temperature > 29.4 (temp=30.2, humidity=65, pressure=1013.25)
DEBUG: AND condition passed: temperature > 29.4
INFO: Alarm triggered: high-temperature (condition: temperature > 29.4)
INFO: Sent console notification for alarm high-temperature
```

**Info Level Logs (default):**
- Alarm manager initialization with total alarm count
- Active alarm summary at startup with details:
  - Alarm name and condition
  - Cooldown period
  - Number of delivery channels
- Count of enabled vs total alarms
- Alarm trigger events with condition
- Notification delivery confirmations per channel
- Config file reload events

**Example Info Output:**
```
INFO: Alarm manager initialized with 5 alarms
INFO: Active alarm: high-temperature - temperature > 29.4 (cooldown: 1800s, channels: 1)
INFO: Active alarm: heavy-rain - rain_rate > 5 (cooldown: 600s, channels: 2)
INFO: Active alarm: high-wind - wind_gust > 20 (cooldown: 900s, channels: 1)
INFO: Active alarm: heat-and-humidity - temperature > 30 && humidity > 70 (cooldown: 3600s, channels: 1)
INFO: Active alarm: uv-warning - uv > 8 (cooldown: 7200s, channels: 1)
INFO: 5 of 5 alarms are enabled
```

**Error Level Logs:**
- Alarm evaluation failures (invalid conditions, unknown fields)
- Notifier creation failures
- Notification delivery failures
- Config file reload errors

### 4. Improved User Experience

**Form Enhancements:**
- Sensor field buttons use primary color scheme (purple/blue gradient)
- Hover effects for better interactivity
- Delivery method checkboxes in responsive grid layout
- Visual feedback on hover for all clickable elements
- Icons for each delivery method (emoji-based, universal)

**Validation:**
- Ensures at least one delivery method is selected before save
- Shows error notification if no delivery method selected
- Preserves existing email/SMS configurations when editing

## Usage Examples

### Creating a Simple Temperature Alarm

1. Click "New Alarm" button
2. Enter name: `hot-weather`
3. Click the `temperature` button
4. Type: ` > 35`
5. Select delivery: ✓ Console, ✓ Syslog
6. Click "Save Alarm"

Result: Alarm triggers when temperature exceeds 35°C, logs to both console and syslog.

### Creating a Compound Condition

1. Click "New Alarm" button
2. Enter name: `dangerous-conditions`
3. Click `temperature` → type ` > 38 && `
4. Click `humidity` → type ` < 20`
5. Select delivery: ✓ Console, ✓ Email
6. Click "Save Alarm"

Result: Alarm triggers when it's both very hot AND very dry (fire weather conditions).

### Debugging Alarm Behavior

Run with debug logging enabled:
```bash
./tempest-homekit-go --log-level=debug --alarms=@alarms.json
```

Watch for:
- `DEBUG: Testing alarm: <name>` - Shows which alarms are being evaluated
- `DEBUG: Evaluating condition:` - Shows the condition and current sensor values
- `DEBUG: Skipping disabled alarm:` - Confirms disabled alarms aren't evaluated
- `DEBUG: Alarm <name> in cooldown` - Shows cooldown is working
- `INFO: Alarm triggered:` - Confirms successful trigger

## Configuration Files

### Example: Multi-Delivery Alarm

```json
{
  "name": "severe-weather",
  "description": "Critical weather conditions",
  "condition": "wind_gust > 30 || rain_rate > 20",
  "tags": ["severe", "emergency"],
  "enabled": true,
  "cooldown": 300,
  "channels": [
    {
      "type": "console",
      "template": "⚠️ SEVERE WEATHER: {{condition}}"
    },
    {
      "type": "syslog",
      "template": "SEVERE WEATHER ALERT: {{condition}} at {{station}}"
    },
    {
      "type": "email",
      "template": "Severe weather detected at {{station}}: {{condition}}",
      "email": {
        "to": ["admin@example.com"],
        "subject": "URGENT: Severe Weather Alert",
        "body": "Severe weather conditions detected.\n\nStation: {{station}}\nCondition: {{condition}}\nWind Gust: {{wind_gust}} m/s\nRain Rate: {{rain_rate}} mm/hr\nTime: {{timestamp}}"
      }
    }
  ]
}
```

When you edit this alarm in the UI:
- ✓ Console checkbox will be checked
- ✓ Syslog checkbox will be checked  
- ✓ Email checkbox will be checked
- The existing email configuration is preserved

## Best Practices

1. **Always enable Console delivery** - It's free, instant, and helps with debugging
2. **Use appropriate cooldowns** - Prevents notification spam (default: 1800s = 30 minutes)
3. **Test conditions with debug logging** - Verify your conditions work as expected
4. **Use descriptive alarm names** - Makes logs easier to read
5. **Add tags for organization** - Filter alarms by tag in the editor
6. **Start simple** - Test single conditions before creating compound conditions
7. **Monitor initial deployment** - Watch logs to ensure alarms work as expected

## Troubleshooting

**Alarm not triggering:**
- Enable debug logging to see evaluation attempts
- Check if alarm is enabled
- Verify condition syntax (use sensor field buttons to avoid typos)
- Check cooldown period (alarm may be in cooldown)

**No delivery method checkbox appears selected when editing:**
- Check if channels array is empty in JSON
- Verify channel types match: console, syslog, eventlog, email, sms

**Email/SMS not working:**
- Check that global configuration includes SMTP/Twilio settings
- Verify the email/sms config block in the channel
- Check error logs for delivery failures

## API Reference

The editor automatically handles channel creation when you check delivery methods:

- **Console**: Creates `{type: "console", template: "..."}`
- **Syslog**: Creates `{type: "syslog", template: "..."}`
- **Event Log**: Creates `{type: "eventlog", template: "..."}`
- **Email**: Creates with email config (or preserves existing)
- **SMS**: Creates with SMS config (or preserves existing)

All use the same default template with customizable sensor variables.
