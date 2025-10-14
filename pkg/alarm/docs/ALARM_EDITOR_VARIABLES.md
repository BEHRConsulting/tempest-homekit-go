# Enhanced Alarm Editor Variables

## Summary

Enhanced the alarm editor and template system to provide better alarm messages with description prominently displayed and previous sensor values for comparison.

## Changes Made

### 1. Default Message Template

**Old Format:**
```
🚨 ALARM: {{alarm_name}}
Station: {{station}}
Time: {{timestamp}}
Description: {{alarm_description}}
```

**New Format (Description Prominent):**
```
🚨 ALARM: {{alarm_name}}
{{alarm_description}}

Station: {{station}}
Time: {{timestamp}}
```

The description now appears on line 2 in its own line, making it immediately visible.

### 2. Added Previous Value Variables

Added 12 new template variables showing the sensor values that were compared to trigger the alarm:

| Variable | Description | Format | Example |
|----------|-------------|--------|---------|
| `{{last_temperature}}` | Previous temperature in °C | `%.1f` | `31.2` |
| `{{last_humidity}}` | Previous humidity percentage | `%.0f` | `51` |
| `{{last_pressure}}` | Previous pressure in mb | `%.2f` | `978.50` |
| `{{last_wind_speed}}` | Previous wind speed in m/s | `%.1f` | `0.8` |
| `{{last_wind_gust}}` | Previous wind gust in m/s | `%.1f` | `1.2` |
| `{{last_wind_direction}}` | Previous wind direction in degrees | `%.0f` | `180` |
| `{{last_lux}}` | Previous light level in lux | `%.0f` | `5997` |
| `{{last_uv}}` | Previous UV index | `%d` | `2` |
| `{{last_rain_rate}}` | Previous rain rate in mm | `%.2f` | `0.00` |
| `{{last_rain_daily}}` | Previous daily rain in mm | `%.2f` | `0.00` |
| `{{last_lightning_count}}` | Previous lightning strike count | `%d` | `0` |
| `{{last_lightning_distance}}` | Previous lightning distance in km | `%.1f` | `10.5` |

**Behavior:**
- Shows `N/A` if no previous value exists (first observation)
- Shows the actual previous value that was compared against

### 3. Clarified Current vs Previous in UI

All dropdown menus now indicate whether a variable shows current or previous values:

**Before:**
```
{{temperature}} - Temperature °C
{{lux}} - Light Lux
```

**After:**
```
{{temperature}} - Temperature °C (current)
{{lux}} - Light Lux (current)
{{last_temperature}} - Temperature °C (previous)
{{last_lux}} - Light Lux (previous)
```

## Use Cases

### Change Detection Alarms

Perfect for alarms using `*field`, `>field`, or `<field` operators:

**Example - Lux Change:**
```
🚨 ALARM: {{alarm_name}}
{{alarm_description}}

Current LUX: {{lux}}
Previous LUX: {{last_lux}}
Change: {{lux}} lux (was {{last_lux}} lux)

Station: {{station}}
Time: {{timestamp}}
```

**Output:**
```
🚨 ALARM: Lux Change
This alarm should alert on LUX change

Current LUX: 5950
Previous LUX: 5997
Change: 5950 lux (was 5997 lux)

Station: Chino Hills
Time: 2025-10-10 17:15:54 PDT
```

### Temperature Increase Detection

```
🚨 ALARM: {{alarm_name}}
{{alarm_description}}

Temperature increased from {{last_temperature}}°C to {{temperature}}°C
Increase: {{temperature}} - {{last_temperature}} = [calculate manually] °C

Time: {{timestamp}}
```

### Rain Detection

```
🚨 ALARM: Rain Started
It's raining! Bring in your laundry.

Current rain rate: {{rain_rate}} mm/hr
Previous rain rate: {{last_rain_rate}} mm/hr

Station: {{station}}
```

## Implementation Details

### Files Modified

**1. pkg/alarm/editor/html.go**
- Updated default message template (5 lines, description on line 2)
- Added `last_*` variables to all 6 dropdown menus:
  - Default message dropdown
  - Console message dropdown
  - Syslog message dropdown
  - Event log message dropdown
  - Email message dropdown
  - SMS message dropdown
- Added "(current)" and "(previous)" labels to clarify variable purpose

**2. pkg/alarm/notifiers.go** - `expandTemplate()` function
- Added 12 new variable expansions for `{{last_*}}` placeholders
- Uses `alarm.GetPreviousValue(field)` to retrieve previous values
- Returns "N/A" if no previous value exists
- Maintains same formatting as current values for consistency

**3. pkg/alarm/evaluator.go** - `evaluateChangeDetection()` function
- Fixed timing issue where previous value was updated too early
- Now updates `previousValue` AFTER comparison but BEFORE notification
- Ensures template expansion sees the correct previous value
- Removed `defer` statement to control exact timing

## Total Variables Available

The alarm template system now supports **30 variables**:

**Alarm Info (2):**
- `{{alarm_name}}` - Alarm name
- `{{alarm_description}}` - Alarm description

**Context (2):**
- `{{station}}` - Station name
- `{{timestamp}}` - Current time formatted

**Current Sensor Values (13):**
- `{{temperature}}`, `{{temperature_f}}`, `{{temperature_c}}`
- `{{humidity}}`
- `{{pressure}}`
- `{{wind_speed}}`, `{{wind_gust}}`, `{{wind_direction}}`
- `{{lux}}`
- `{{uv}}`
- `{{rain_rate}}`, `{{rain_daily}}`
- `{{lightning_count}}`, `{{lightning_distance}}`

**Previous Sensor Values (12):**
- `{{last_temperature}}`
- `{{last_humidity}}`
- `{{last_pressure}}`
- `{{last_wind_speed}}`, `{{last_wind_gust}}`, `{{last_wind_direction}}`
- `{{last_lux}}`
- `{{last_uv}}`
- `{{last_rain_rate}}`, `{{last_rain_daily}}`
- `{{last_lightning_count}}`, `{{last_lightning_distance}}`

## Testing

### Automated Test
```bash
./test-enhanced-alarm-message.sh
```

Verifies:
- ✅ Description appears on line 2
- ✅ `last_lux` variable is present
- ✅ Previous value shows correctly

### Manual Testing with Alarm Editor

1. Start alarm editor:
   ```bash
   ./tempest-homekit-go --alarms-edit @tempest-alarms.json --alarms-edit-port 8081
   ```

2. Open browser to http://localhost:8081

3. Create or edit an alarm

4. Use "Insert Variable" dropdown to see all 30 variables

5. Select any `last_*` variable to add to template

6. Save and test with actual application

## Benefits

1. **Better Context**: See both current and previous values in notifications
2. **Debugging**: Understand exactly what triggered the alarm
3. **Change Magnitude**: Calculate how much a value changed
4. **User Friendly**: Description prominently displayed on line 2
5. **Consistent Format**: Previous values use same formatting as current values
6. **Graceful Fallback**: Shows "N/A" if no previous value exists

## Backwards Compatibility

✅ **Fully compatible** - Existing alarm configurations continue to work:
- Old variables still work
- Old message templates unchanged
- New variables are optional
- No breaking changes

## Example Configurations

### Minimal (Quick Setup)
```json
{
  "type": "console",
  "template": "🚨 {{alarm_name}}: {{alarm_description}}"
}
```

### Standard (Recommended)
```json
{
  "type": "console",
  "template": "🚨 ALARM: {{alarm_name}}\n{{alarm_description}}\n\nStation: {{station}}\nTime: {{timestamp}}"
}
```

### Detailed (Change Detection)
```json
{
  "type": "console",
  "template": "🚨 ALARM: {{alarm_name}}\n{{alarm_description}}\n\nCurrent: {{lux}} lux\nPrevious: {{last_lux}} lux\n\nStation: {{station}}\nTime: {{timestamp}}"
}
```

### Comprehensive (All Data)
```json
{
  "type": "console",
  "template": "🚨 ALARM: {{alarm_name}}\n{{alarm_description}}\n\nCurrent Conditions:\n  Temp: {{temperature}}°C ({{temperature_f}}°F)\n  Humidity: {{humidity}}%\n  Light: {{lux}} lux\n\nPrevious Conditions:\n  Temp: {{last_temperature}}°C\n  Humidity: {{last_humidity}}%\n  Light: {{last_lux}} lux\n\nStation: {{station}}\nTime: {{timestamp}}"
}
```

## Documentation Updated

- ALARM_EDITOR_VARIABLES.md (this file)
- Updated inline help in alarm editor UI
- Variable dropdown descriptions
- Test scripts with examples
