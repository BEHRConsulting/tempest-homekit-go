# Alarm Editor Message Configuration Enhancement

## Overview

The alarm editor now includes comprehensive message configuration capabilities, allowing users to customize notification messages for each delivery method with easy-to-use variable insertion.

## Features

### 1. Default Message Template

**Use Case**: Quickly set up alarms with a single message that applies to all selected delivery methods.

**Benefits**:
- Simplifies alarm creation
- Ensures consistent messaging across channels
- Single point of maintenance for common messages

**Default Template**:
```
🚨 ALARM: {{alarm_name}}
Station: {{station}}
Time: {{timestamp}}
```

### 2. Custom Messages Per Delivery Method

**Use Case**: Fine-tune messages for specific delivery channels (e.g., shorter messages for SMS, detailed HTML for email).

**Benefits**:
- Channel-specific formatting
- Optimize message length (e.g., SMS character limits)
- Include different details per channel

**Available Sections**:
- **Console**: Standard output messages
- **Syslog**: System log messages
- **Event Log**: Windows event log / Unix syslog messages
- **Email**: Subject, recipient(s), and body with rich formatting
- **SMS**: Phone number(s) and short message

### 3. Variable Dropdown

**Use Case**: Insert weather data and alarm metadata into messages without memorizing variable names.

**Available Variables**:

#### Alarm Metadata
- `{{alarm_name}}` - Name of the triggered alarm
- `{{station}}` - Weather station name
- `{{timestamp}}` - Current timestamp (format: YYYY-MM-DD HH:MM:SS TZ)

#### Temperature
- `{{temperature}}` - Temperature in Celsius
- `{{temperature_f}}` - Temperature in Fahrenheit
- `{{temperature_c}}` - Temperature in Celsius (explicit)

#### Atmospheric
- `{{humidity}}` - Relative humidity percentage
- `{{pressure}}` - Station pressure in millibars

#### Wind
- `{{wind_speed}}` - Average wind speed (m/s)
- `{{wind_gust}}` - Peak wind gust (m/s)
- `{{wind_direction}}` - Wind direction in degrees

#### Light & UV
- `{{lux}}` - Light intensity in lux
- `{{uv}}` - UV index

#### Precipitation
- `{{rain_rate}}` - Current rain rate (mm)
- `{{rain_daily}}` - Daily rain accumulation (mm)

#### Lightning
- `{{lightning_count}}` - Number of lightning strikes
- `{{lightning_distance}}` - Average lightning strike distance (km)

## User Interface

### Message Configuration Workflow

1. **Select Delivery Methods**
   - Check one or more delivery method checkboxes
   - Relevant message sections appear/disappear dynamically

2. **Choose Message Strategy**
   - **Default Message (Recommended)**: Check "Use Default Message for All Selected Methods"
     - Single message template applies to all channels
     - Simpler to maintain
   - **Custom Messages**: Uncheck to reveal individual message sections
     - Customize each channel independently
     - Email includes separate subject and body fields
     - SMS includes phone number configuration

3. **Insert Variables**
   - Click variable dropdown in any message section
   - Select variable from categorized list
   - Variable inserted at cursor position
   - Dropdown resets automatically

4. **Configure Channel-Specific Settings**
   - **Email**: 
     - Recipient addresses (comma-separated)
     - Subject line with variable support
     - Body with variable support
   - **SMS**:
     - Phone numbers (comma-separated, international format)
     - Message (keep short for SMS limits)

## Examples

### Example 1: Simple Temperature Alarm

**Default Message**:
```
🚨 High Temperature Alert
Station: {{station}}
Current Temp: {{temperature_f}}°F ({{temperature_c}}°C)
Time: {{timestamp}}
```

**Result** (all channels receive):
```
🚨 High Temperature Alert
Station: Backyard Weather
Current Temp: 95.0°F (35.0°C)
Time: 2025-10-10 14:23:15 PDT
```

### Example 2: Custom Messages for Different Channels

**Console Message**:
```
[ALARM] {{alarm_name}} - Temp: {{temperature_f}}°F, Humidity: {{humidity}}% at {{timestamp}}
```

**SMS Message** (short):
```
⚠️ {{alarm_name}}: {{temperature_f}}°F @ {{timestamp}}
```

**Email Subject**:
```
🌡️ Weather Alert: {{alarm_name}}
```

**Email Body**:
```
Weather Alert Notification
========================

Alarm: {{alarm_name}}
Station: {{station}}
Timestamp: {{timestamp}}

Current Conditions:
- Temperature: {{temperature_f}}°F ({{temperature_c}}°C)
- Humidity: {{humidity}}%
- Pressure: {{pressure}} mb
- Wind: {{wind_speed}} m/s from {{wind_direction}}°

This is an automated alert from your Tempest weather station.
```

### Example 3: Lightning Detection

**Default Message**:
```
⚡ LIGHTNING DETECTED
Station: {{station}}
Strike Count: {{lightning_count}}
Average Distance: {{lightning_distance}} km
Time: {{timestamp}}

Take shelter immediately!
```

### Example 4: Multi-Condition Weather Event

**Default Message**:
```
🌧️ Severe Weather Event
{{alarm_name}}

Station: {{station}}
Time: {{timestamp}}

Conditions:
• Temperature: {{temperature_f}}°F
• Wind Speed: {{wind_speed}} m/s (Gust: {{wind_gust}} m/s)
• Rain Rate: {{rain_rate}} mm/hr
• Pressure: {{pressure}} mb

Stay safe and monitor conditions.
```

## Technical Details

### Message Storage

Messages are stored in the alarm configuration JSON as part of each channel:

```json
{
  "name": "High Temperature",
  "condition": "temperature > 85F",
  "enabled": true,
  "channels": [
    {
      "type": "console",
      "template": "🚨 ALARM: {{alarm_name}}\\nTemp: {{temperature_f}}°F"
    },
    {
      "type": "email",
      "email": {
        "to": ["user@example.com"],
        "subject": "⚠️ {{alarm_name}}",
        "body": "Alert at {{timestamp}}\\nTemp: {{temperature_f}}°F"
      }
    },
    {
      "type": "sms",
      "sms": {
        "to": ["+15551234567"],
        "message": "{{alarm_name}}: {{temperature_f}}°F"
      }
    }
  ]
}
```

### Variable Expansion

Variables are expanded at notification time using the `expandTemplate()` function in `pkg/alarm/notifiers.go`. This ensures real-time values are always used.

### Editing Existing Alarms

When editing an alarm:
1. Existing messages are loaded into appropriate fields
2. If all channels use the same message, "Use Default Message" is checked
3. If channels have different messages, custom message sections are shown
4. Email and SMS configurations (recipients, phone numbers) are preserved

## Best Practices

### 1. Message Length
- **Console/Syslog**: No strict limits, but keep under 500 characters for readability
- **Email**: Can be longer, use formatting for clarity
- **SMS**: Keep under 160 characters to avoid multi-part messages

### 2. Variable Usage
- Always include `{{alarm_name}}` for context
- Include `{{timestamp}}` for record-keeping
- Use relevant sensor variables based on alarm condition

### 3. Formatting
- Use emojis for visual impact (🚨 ⚠️ 🌡️ ⚡ 🌧️)
- Use line breaks for readability
- For email, consider using bullet points and headers

### 4. Testing
- Test all delivery methods before deploying
- Verify variable expansion works as expected
- Check SMS character count for international delivery

### 5. Maintenance
- Use default message when possible for easier updates
- Document custom message logic for complex setups
- Review and update messages periodically

## Troubleshooting

### Variables Not Expanding
- Ensure variable names are exact (case-sensitive)
- Check for typos in variable syntax `{{variable_name}}`
- Verify the sensor data is available for the variable

### Messages Not Saving
- Ensure at least one delivery method is selected
- Check that required fields are filled (e.g., email recipients)
- Verify JSON syntax if manually editing config file

### SMS Messages Cut Off
- Keep total message under 160 characters
- Reduce variable usage or use abbreviations
- Test with actual phone numbers to verify delivery

### Email Not Formatting
- Email body supports plain text only (not HTML)
- Use line breaks (`\n`) for structure
- Consider using spacing and ASCII characters for visual separation

## Related Documentation

- **[pkg/alarm/README.md](pkg/alarm/README.md)** - Alarm system architecture
- **[pkg/alarm/editor/README.md](pkg/alarm/editor/README.md)** - Alarm editor overview
- **[CHANGE_DETECTION_OPERATORS.md](CHANGE_DETECTION_OPERATORS.md)** - Change detection in conditions
- **[examples/alarms-with-change-detection.json](examples/alarms-with-change-detection.json)** - Example configurations

## Version History

- **v1.0** (October 2025): Initial implementation
  - Default message template
  - Custom messages per delivery method
  - Variable dropdown with all sensor fields
  - Dynamic UI based on selected delivery methods
  - Email and SMS recipient configuration
