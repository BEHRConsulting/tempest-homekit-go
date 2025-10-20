# Change Detection Quick Reference

## Operators

| Operator | Name | Triggers When | Example |
|----------|------|---------------|---------|
| `*field` | Any Change | Value changes at all | `*lightning_count` |
| `>field` | Increase | Value goes up | `>rain_rate` |
| `<field` | Decrease | Value goes down | `<lightning_distance` |

## Common Patterns

### Lightning Monitoring
```
*lightning_count → Any strike
<lightning_distance → Getting closer
*lightning_count && lightning_distance < 10 → Strike within 10km
```

### Rain Detection
```
>rain_rate → Rain starting or increasing
>rain_rate && rain_rate > 5 → Heavy rain onset
```

### Storm Approach
```
<pressure → Pressure dropping
<pressure && pressure < 1000 → Rapid pressure drop
<pressure && >wind_speed → Storm approaching (falling pressure + rising wind)
```

### Wind Monitoring
```
>wind_gust → Wind strengthening
>wind_speed || >wind_gust → Any wind increase
>wind_gust && wind_gust > 15 → Strong wind strengthening
```

### Temperature Changes
```
>temperature → Warming
<temperature → Cooling
>temperature && temperature > 35 → Heat wave
<temperature && temperature < 0 → Freeze approaching
```

### Humidity Changes
```
>humidity → Humidity rising
>humidity && temperature > 30 → Muggy conditions
<humidity && humidity < 30 → Drying out
```

### UV Index
```
>uv → UV increasing
>uv && uv > 8 → Dangerous UV levels
```

## All Supported Fields

### Atmospheric
- `temperature`, `temp` (Celsius)
- `humidity` (%)
- `pressure` (mb)

### Wind
- `wind_speed`, `wind` (m/s)
- `wind_gust` (m/s)
- `wind_direction` (degrees)

### Precipitation
- `rain_rate` (mm/hr)
- `rain_daily` (mm/day)
- `precipitation_type` (0=none, 1=rain, 2=hail, 3=rain+hail)

### Lightning
- `lightning_count` (strikes)
- `lightning_distance` (km)

### Light
- `lux`, `light` (lux)
- `uv`, `uv_index` (index)

## Combining with Thresholds

You can mix change detection with regular comparisons:

```json
"condition": "*lightning_count && lightning_distance < 10"
```
Triggers when: Lightning detected AND within 10km

```json
"condition": ">rain_rate && rain_rate > 5"
```
Triggers when: Rain increasing AND rate > 5mm/hr

```json
"condition": "<pressure && pressure < 1000 && >wind_speed"
```
Triggers when: Pressure falling AND < 1000mb AND wind rising

## Logical Operators

### AND (`&&`) - All conditions must be true
```json
"condition": "*lightning_count && lightning_distance < 10"
```

### OR (`||`) - Any condition can be true
```json
"condition": ">rain_rate || >wind_gust"
```

### Complex
```json
"condition": "(*lightning_count && lightning_distance < 15) || (>wind_gust && wind_gust > 20)"
```

## Unit Support

Change detection works with unit suffixes:

### Temperature
- `temperature > 80F` (Fahrenheit)
- `temperature > 26.7C` (Celsius, explicit)
- `temperature > 26.7` (Celsius, default)

### Wind Speed
- `wind_gust > 25mph` (miles per hour)
- `wind_gust > 11.2m/s` (meters per second, explicit)
- `wind_gust > 11.2` (m/s, default)

## Important Notes

### First Observation
The first observation establishes a baseline and will NOT trigger:
```
Obs 1: lightning_count=0 → No trigger (baseline)
Obs 2: lightning_count=1 → TRIGGER (changed!)
```

### Cooldown
Use cooldown to prevent spam on rapidly changing values:
```json
{
 "condition": "*lightning_count",
 "cooldown": 60
}
```
Will only notify once per minute, even if multiple strikes occur.

### State Tracking
- Each field in change conditions maintains independent state
- State persists during runtime
- State resets on application restart

## Examples

### Simple Examples
```json
{"condition": "*lightning_count"} // Any lightning
{"condition": ">rain_rate"} // Rain increasing
{"condition": "<lightning_distance"} // Lightning closer
{"condition": "<pressure"} // Pressure falling
{"condition": ">wind_gust"} // Wind strengthening
```

### Compound Examples
```json
{"condition": "*lightning_count && lightning_distance < 10"}
{"condition": ">rain_rate || >wind_gust"}
{"condition": "<pressure && pressure < 1000"}
{"condition": "temperature > 32 && >humidity"}
```

### Complex Examples
```json
{
 "condition": "(*lightning_count && lightning_distance < 15) || (>wind_gust && wind_gust > 15)"
}

{
 "condition": "<pressure && pressure < 1000 && (>wind_speed || >rain_rate)"
}
```

## Testing in Alarm Editor

1. Open alarm editor: `./tempest-homekit-go --alarm-editor @alarms.json --port 8081`
2. Create new alarm
3. In condition field, use change operators:
 - `*lightning_count`
 - `>rain_rate`
 - `<lightning_distance`
4. Click sensor names to insert them
5. Add operators before field name
6. Combine with thresholds and logical operators

## Troubleshooting

### "Change not detected"
- First observation establishes baseline (no trigger)
- Check that value actually changed
- Verify field name is correct

### "Too many notifications"
- Add or increase cooldown period
- Use threshold with change operator: `>rain_rate && rain_rate > 1`

### "Operator not working"
- Verify operator is before field name: `*field` not `field*`
- Check for typos in field name
- Ensure no spaces: `*lightning_count` not `* lightning_count`

## See Also

- [Full Documentation](CHANGE_DETECTION_OPERATORS.md) - Detailed explanation
- [Example Config](../../examples/alarms-with-change-detection.json) - Ready-to-use examples
- [Unit Conversion](UNIT_CONVERSION_SUPPORT.md) - Temperature and wind units
