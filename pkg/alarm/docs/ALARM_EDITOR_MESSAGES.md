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
{{alarm_name}}
Station: {{station}}
Time: {{timestamp}}
```

### 2. Custom Messages Per Delivery Method

**Use Case**: Fine-tune messages for specific delivery channels (e.g., shorter messages for SMS, detailed HTML for email).

**Benefits**:

**Available Sections**:

### 3. Variable Dropdown

**Use Case**: Insert weather data and alarm metadata into messages without memorizing variable names.

**Available Variables**:

#### Alarm Metadata

#### Temperature

#### Atmospheric

#### Wind

#### Light & UV

#### Precipitation

#### Lightning

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
 - **Email**:  - Recipient addresses (comma-separated)
 - Subject line with variable support
 - Body with variable support
 - **SMS**:
 - Phone numbers (comma-separated, international format)
 - Message (keep short for SMS limits)

## Examples

### Example 1: Simple Temperature Alarm

**Default Message**:
```
 High Temperature Alert
Station: {{station}}
Current Temp: {{temperature_f}}°F ({{temperature_c}}°C)
Time: {{timestamp}}
```

**Result** (all channels receive):
```
 High Temperature Alert
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
Warning: {{alarm_name}}: {{temperature_f}}°F @ {{timestamp}}
```

**Email Subject**:
```
Temperature: Weather Alert: {{alarm_name}}
```

**Email Body**:
```
Weather Alert Notification
========================

Alarm: {{alarm_name}}
Station: {{station}}
Timestamp: {{timestamp}}

Current Conditions:

This is an automated alert from your Tempest weather station.
```

### Example 3: Lightning Detection

**Default Message**:
```
 LIGHTNING DETECTED
Station: {{station}}
Strike Count: {{lightning_count}}
Average Distance: {{lightning_distance}} km
Time: {{timestamp}}

Take shelter immediately!
```

### Example 4: Multi-Condition Weather Event

**Default Message**:
```
️ Severe Weather Event
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
 "template": "ALARM: {{alarm_name}}\\nTemp: {{temperature_f}}°F"
 },
 {
 "type": "email",
 "email": {
 "to": ["user@example.com"],
 "subject": "Warning: {{alarm_name}}",
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

### 2. Variable Usage

### 3. Formatting

### 4. Testing

### 5. Maintenance

## Troubleshooting

### Variables Not Expanding

### Messages Not Saving

### SMS Messages Cut Off

### Email Not Formatting

## Related Documentation


## Version History

 - Default message template
 - Custom messages per delivery method
 - Variable dropdown with all sensor fields
 - Dynamic UI based on selected delivery methods
 - Email and SMS recipient configuration
