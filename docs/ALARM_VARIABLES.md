# Alarm Template Variables

This document describes the template variables available for use in alarm messages.

## New Variables Added

### `{{app_info}}`
**Description:** Displays application information including version, uptime, and build details.

**Example Output:**
```
Tempest HomeKit Bridge v1.7.0
Built: 2025-10-15
Uptime: 2 days, 3 hours, 45 minutes
Go version: 1.24.2
```

**Implementation Required:**
- Version string from `main.go` (currently hardcoded as "v1.7.0")
- Application start time to calculate uptime
- Build timestamp
- Go version from `runtime.Version()`

### `{{alarm_info}}`
**Description:** Displays comprehensive alarm information in a nicely formatted block.

**Example Output:**
```
Alarm: High Temperature Alert
Description: Temperature exceeded safe operating threshold
Condition: temperature > 85F
Status: TRIGGERED
Cooldown: 30 minutes
Tags: critical, temperature, safety
```

**Implementation Required:**
- Alarm name
- Alarm description
- Alarm condition (the expression that triggered)
- Current alarm status
- Cooldown period in human-readable format
- Tag list (comma-separated)

### `{{sensor_info}}`
**Description:** Displays current sensor readings that are relevant to the alarm condition, formatted appropriately for the delivery method.

**Example Output (Plain Text):**
```
Temperature: 87.5°F (30.8°C)
Humidity: 65%
Pressure: 1013.2 mb
Wind Speed: 12.5 mph (5.6 m/s)
Wind Gust: 18.2 mph (8.1 m/s)
Wind Direction: 245° (WSW)
UV Index: 6
Illuminance: 45,230 lux
Rain Rate: 0.0 mm/hr
Daily Rain: 0.5 in (12.7 mm)
Lightning: 0 strikes
```

**Example Output (HTML):**
```html
<table>
<tr><td>Temperature:</td><td>87.5°F (30.8°C)</td></tr>
<tr><td>Humidity:</td><td>65%</td></tr>
<tr><td>Pressure:</td><td>1013.2 mb</td></tr>
...
</table>
```

**Implementation Required:**
- Current sensor readings from weather data
- Unit conversions (C/F for temp, mph/m/s for wind)
- Formatted output based on context (HTML vs plain text)
- Highlighting of sensors involved in the alarm condition

## Existing Variables

All existing variables remain supported:

### Basic Info
- `{{alarm_name}}` - Name of the alarm
- `{{alarm_description}}` - Alarm description text
- `{{alarm_condition}}` - The condition expression
- `{{station}}` - Weather station name
- `{{timestamp}}` - Current date/time

### Current Sensor Values
- `{{temperature}}` - Temperature in °C
- `{{temperature_f}}` - Temperature in °F
- `{{temperature_c}}` - Temperature in °C (alias)
- `{{humidity}}` - Humidity percentage
- `{{pressure}}` - Barometric pressure in mb
- `{{wind_speed}}` - Wind speed in m/s
- `{{wind_gust}}` - Wind gust in m/s
- `{{wind_direction}}` - Wind direction in degrees
- `{{lux}}` - Illuminance in lux
- `{{uv}}` - UV index
- `{{rain_rate}}` - Rain rate in mm/hr
- `{{rain_daily}}` - Daily accumulated rain in mm
- `{{lightning_count}}` - Lightning strike count
- `{{lightning_distance}}` - Lightning distance in km

### Previous Sensor Values
All current sensor variables also have `last_*` versions for previous readings:
- `{{last_temperature}}`
- `{{last_humidity}}`
- `{{last_pressure}}`
- ... and so on

## Email HTML Format Support

The email delivery method now supports an `html` flag in its configuration:

```json
{
  "type": "email",
  "email": {
    "to": ["recipient@example.com"],
    "subject": "Weather Alert",
    "body": "<h1>Alert!</h1><p>{{alarm_info}}</p>",
    "html": true
  }
}
```

When `html: true`, the email body is sent as HTML content-type and can include:
- HTML tags: `<h1>`, `<h2>`, `<p>`, `<strong>`, `<em>`, `<br>`, `<hr>`
- Tables: `<table>`, `<tr>`, `<td>`, `<th>`
- Inline styles: `style="color: red;"`, etc.
- Divs and spans for layout

When `html: false` or omitted, plain text is used.

## Implementation Notes

1. **Template Engine:** The backend needs to implement variable substitution for all variables
2. **Context Awareness:** `{{sensor_info}}` should detect if it's being used in an HTML context and format accordingly
3. **Error Handling:** If a variable cannot be resolved, it should be replaced with a sensible default (e.g., "N/A" or the variable name)
4. **Performance:** Variable resolution should be efficient as it occurs on every alarm trigger

## Backend Changes Required

### In `pkg/alarm/manager.go` or similar:

1. Add application info tracking:
```go
var (
    appVersion   = "v1.7.0"
    appStartTime = time.Now()
    buildTime    string // Set via ldflags during build
)
```

2. Implement variable resolver:
```go
func (m *Manager) resolveVariables(template string, alarm *Alarm, data *weather.Observation) string {
    // Implement substitution for all variables
    // Handle special cases for {{app_info}}, {{alarm_info}}, {{sensor_info}}
}
```

3. Add HTML detection for `{{sensor_info}}`:
```go
func formatSensorInfo(data *weather.Observation, isHTML bool) string {
    if isHTML {
        return formatSensorInfoHTML(data)
    }
    return formatSensorInfoText(data)
}
```

## Testing

Test templates with the new variables:

```
Console: {{alarm_name}} triggered at {{timestamp}}
{{sensor_info}}

Email HTML:
<h2>{{alarm_name}}</h2>
<p>{{alarm_description}}</p>
{{sensor_info}}
{{app_info}}

SMS: {{alarm_name}}: {{sensor_info}}
```
