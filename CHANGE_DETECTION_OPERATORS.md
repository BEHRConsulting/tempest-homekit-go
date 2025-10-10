# Change Detection Operators

## Overview

The alarm system now supports **unary change-detection operators** that trigger alarms based on changes in sensor values rather than absolute thresholds. These are particularly useful for event-based monitoring like lightning strikes, rain onset, or approaching storms.

## Operators

### `*field` - Any Change
Triggers whenever the field value changes from its previous value.

**Use Cases:**
- Lightning detection (any new strike)
- Precipitation type changes
- Storm activity monitoring

**Examples:**
```json
{
  "name": "Lightning Detected",
  "condition": "*lightning_count",
  "description": "Alert on any lightning strike"
}
```

### `>field` - Increase Detected
Triggers when the field value increases from its previous value.

**Use Cases:**
- Rain intensity increasing
- Wind strengthening
- Temperature rising rapidly
- UV index climbing

**Examples:**
```json
{
  "name": "Rain Intensifying",
  "condition": ">rain_rate",
  "description": "Alert when rain gets heavier"
}

{
  "name": "Wind Strengthening",
  "condition": ">wind_gust",
  "description": "Alert when wind gusts increase"
}
```

### `<field` - Decrease Detected
Triggers when the field value decreases from its previous value.

**Use Cases:**
- Lightning getting closer (distance decreasing)
- Pressure dropping (storm approaching)
- Temperature falling rapidly

**Examples:**
```json
{
  "name": "Lightning Approaching",
  "condition": "<lightning_distance",
  "description": "Alert when lightning gets closer"
}

{
  "name": "Pressure Dropping",
  "condition": "<pressure",
  "description": "Alert when barometric pressure falls"
}
```

## How It Works

### State Tracking
- Each alarm maintains a **previous value** map for all fields used in change-detection conditions
- On first observation, a baseline is established (no trigger)
- Subsequent observations compare against the baseline
- Previous values are updated after each evaluation

### Baseline Establishment
```
Observation 1: lightning_count = 0  → No trigger (establishing baseline)
Observation 2: lightning_count = 0  → No trigger (no change)
Observation 3: lightning_count = 1  → TRIGGER! (change detected: 0→1)
Observation 4: lightning_count = 1  → No trigger (no change)
Observation 5: lightning_count = 3  → TRIGGER! (change detected: 1→3)
```

### Directional Triggers
```
Rain Rate Example:
Observation 1: rain_rate = 0.0   → No trigger (baseline)
Observation 2: rain_rate = 0.5   → TRIGGER with >rain_rate (increase)
Observation 3: rain_rate = 2.0   → TRIGGER with >rain_rate (increase)
Observation 4: rain_rate = 2.0   → No trigger (no change)
Observation 5: rain_rate = 1.0   → No trigger with >rain_rate (decreased, not increased)
Observation 6: rain_rate = 3.0   → TRIGGER with >rain_rate (increase again)
```

## Compound Conditions

Change-detection operators can be combined with regular threshold operators and logical operators.

### Examples

#### Lightning Close AND Active
```json
{
  "name": "Close Lightning",
  "condition": "*lightning_count && lightning_distance < 10",
  "description": "Alert on any lightning strike within 10km"
}
```

#### Rain Increasing OR Wind Increasing
```json
{
  "name": "Storm Intensifying",
  "condition": ">rain_rate || >wind_gust",
  "description": "Alert when storm conditions worsen"
}
```

#### High Temperature AND Rising Humidity
```json
{
  "name": "Heat Index Rising",
  "condition": "temperature > 30 && >humidity",
  "description": "Alert when it's hot and getting more humid"
}
```

#### Approaching Storm
```json
{
  "name": "Storm Approaching",
  "condition": "<pressure && >wind_speed",
  "description": "Alert when pressure drops and wind increases"
}
```

## All Supported Fields

Change-detection operators work with **all** weather sensor fields:

### Temperature
- `*temperature`, `>temperature`, `<temperature`
- `*temp`, `>temp`, `<temp` (alias)

### Humidity
- `*humidity`, `>humidity`, `<humidity`

### Pressure
- `*pressure`, `>pressure`, `<pressure`

### Wind
- `*wind_speed`, `>wind_speed`, `<wind_speed`
- `*wind`, `>wind`, `<wind` (alias)
- `*wind_gust`, `>wind_gust`, `<wind_gust`
- `*wind_direction`, `>wind_direction`, `<wind_direction`

### Light
- `*lux`, `>lux`, `<lux`
- `*light`, `>light`, `<light` (alias)
- `*uv`, `>uv`, `<uv`
- `*uv_index`, `>uv_index`, `<uv_index` (alias)

### Precipitation
- `*rain_rate`, `>rain_rate`, `<rain_rate`
- `*rain_daily`, `>rain_daily`, `<rain_daily`
- `*precipitation_type`, `>precipitation_type`, `<precipitation_type`

### Lightning
- `*lightning_count`, `>lightning_count`, `<lightning_count`
- `*lightning_distance`, `>lightning_distance`, `<lightning_distance`

## Real-World Use Cases

### 1. Lightning Monitor
```json
{
  "name": "Lightning Activity",
  "condition": "*lightning_count",
  "cooldown": 60,
  "description": "Alert on any lightning strike (1 min cooldown)",
  "channels": [
    {
      "type": "console",
      "template": "⚡ Lightning strike detected! Count: {{lightning_count}}, Distance: {{lightning_distance}}km"
    }
  ]
}
```

### 2. Flash Flood Warning
```json
{
  "name": "Rapid Rain Onset",
  "condition": ">rain_rate && rain_rate > 10",
  "description": "Alert when heavy rain starts or intensifies",
  "channels": [
    {
      "type": "sms",
      "sms": {
        "to": ["+1234567890"],
        "message": "FLASH FLOOD WARNING: Rain rate increasing rapidly to {{rain_rate}}mm/hr"
      }
    }
  ]
}
```

### 3. Storm Approach Warning
```json
{
  "name": "Storm Approaching",
  "condition": "<pressure && pressure < 1000 && >wind_speed",
  "description": "Alert when pressure drops rapidly with increasing wind",
  "channels": [
    {
      "type": "email",
      "email": {
        "to": ["weather@example.com"],
        "subject": "⚠️ Storm Approaching - {{station_name}}",
        "body": "Rapid pressure drop detected with increasing winds.\n\nCurrent conditions:\n- Pressure: {{pressure}} mb (falling)\n- Wind: {{wind_speed}} m/s (increasing)\n- Temperature: {{temperature_f}}°F\n\nTime: {{timestamp}}"
      }
    }
  ]
}
```

### 4. Lightning Getting Closer
```json
{
  "name": "Lightning Approaching",
  "condition": "<lightning_distance && lightning_distance < 20",
  "cooldown": 120,
  "description": "Alert when lightning moves closer and is within 20km",
  "channels": [
    {
      "type": "console",
      "template": "⚠️ Lightning approaching! Distance decreased to {{lightning_distance}}km"
    }
  ]
}
```

### 5. Dangerous Heat Building
```json
{
  "name": "Heat Index Rising",
  "condition": "temperature > 32 && >humidity && humidity > 60",
  "description": "Alert when high temperature combines with rising humidity",
  "channels": [
    {
      "type": "console",
      "template": "🌡️ Dangerous heat building: {{temperature_f}}°F with {{humidity}}% humidity (rising)"
    }
  ]
}
```

### 6. Multi-Condition Storm Alert
```json
{
  "name": "Severe Storm Conditions",
  "condition": "(*lightning_count || >wind_gust) && (>rain_rate || <pressure)",
  "description": "Complex storm detection with multiple change indicators",
  "channels": [
    {
      "type": "syslog",
      "template": "SEVERE: Storm detected - Lightning: {{lightning_count}}, Wind: {{wind_gust}}m/s, Rain: {{rain_rate}}mm/hr, Pressure: {{pressure}}mb"
    }
  ]
}
```

## Implementation Details

### Previous Value Storage
```go
type Alarm struct {
    // ... other fields
    previousValue map[string]float64 // Internal state tracking
}
```

### Evaluation Flow
1. Parse condition for unary operator (`*`, `>`, `<`)
2. Get current field value from observation
3. Look up previous value (if exists)
4. Compare values based on operator:
   - `*`: currentValue != previousValue
   - `>`: currentValue > previousValue
   - `<`: currentValue < previousValue
5. Store current value as new previous value
6. Return trigger result

### Thread Safety
- Each alarm maintains its own state
- State updates are handled per-observation
- Manager uses mutex for config access
- No race conditions in state tracking

## Testing

Comprehensive test coverage in `evaluator_change_detection_test.go`:
- ✅ Any change detection (`*field`)
- ✅ Increase detection (`>field`)
- ✅ Decrease detection (`<field`)
- ✅ Compound conditions with change operators
- ✅ Multiple field tracking
- ✅ Error handling
- ✅ Backward compatibility

### Test Results
```
=== RUN   TestChangeDetectionAnyChange
--- PASS: TestChangeDetectionAnyChange (6 subtests)

=== RUN   TestChangeDetectionIncrease
--- PASS: TestChangeDetectionIncrease (7 subtests)

=== RUN   TestChangeDetectionDecrease
--- PASS: TestChangeDetectionDecrease (7 subtests)

=== RUN   TestChangeDetectionWithCompoundConditions
--- PASS: TestChangeDetectionWithCompoundConditions (4 subtests)

=== RUN   TestChangeDetectionMultipleFields
--- PASS: TestChangeDetectionMultipleFields

=== RUN   TestChangeDetectionErrors
--- PASS: TestChangeDetectionErrors (4 subtests)

=== RUN   TestBackwardCompatibility
--- PASS: TestBackwardCompatibility (2 subtests)
```

## Alarm Editor UI

The alarm editor UI has been updated with help text and examples:

```
Click sensor names above to insert into condition.
Supports units: 80F or 26.7C (temp), 25mph or 11.2m/s (wind).
Change detection: *field (any change), >field (increase), <field (decrease).

Examples:
- temperature > 85F
- *lightning_count (any strike)
- >rain_rate (rain increasing)
- <lightning_distance (lightning closer)
```

## Why These Operators?

The choice of operators follows established patterns:

1. **`*field`** - Universal "wildcard" symbol, commonly used for "any" or "all"
   - Intuitive: `*lightning_count` = "any lightning strike"
   - Not ambiguous with multiplication (no second operand)

2. **`>field` / `<field`** - Reuse of comparison operators in unary context
   - Intuitive: `>temperature` = "temperature going up"
   - Visual: Arrow direction shows trend
   - No ambiguity: Regular comparisons require a value (e.g., `temp > 30`)

## Backward Compatibility

✅ All existing alarm conditions continue to work unchanged
✅ Regular comparison operators (`>`, `<`, `>=`, `<=`, `==`, `!=`) unaffected
✅ Compound conditions (`&&`, `||`) work as before
✅ Old `Evaluate()` method still available
✅ Change detection requires explicit use of unary operators

## Limitations and Considerations

### First Observation
- Change detection requires a baseline
- First observation always establishes baseline (no trigger)
- This is by design to prevent false alarms

### State Persistence
- State is kept in memory during runtime
- State is lost on process restart
- After restart, first observation re-establishes baseline

### Cooldown Interaction
- Cooldown applies to change-detection alarms
- If cooldown is active, changes are still tracked but no notification sent
- Useful to prevent notification spam on rapidly changing values

### Compound Conditions
- Each field in compound condition maintains independent state
- Example: `*lightning_count && temperature > 30`
  - Lightning state tracked independently
  - Temperature compared to threshold (no state)

## Performance

- Minimal memory overhead (one float64 per tracked field)
- No additional API calls or database queries
- State lookup is O(1) hash map operation
- Typical alarm has 1-3 tracked fields
- Memory per alarm: ~40 bytes per tracked field

## Future Enhancements

Potential additions (not yet implemented):

1. **Rate of change operators**
   - `>>field` - Rapid increase (e.g., `>>temperature` for quick temp rise)
   - `<<field` - Rapid decrease

2. **Threshold combinations**
   - `>field:value` - Increase beyond threshold (e.g., `>rain_rate:5` = rain increasing AND > 5mm/hr)

3. **State persistence**
   - Save previous values to disk
   - Restore state on restart

4. **Delta operators**
   - `Δfield > value` - Change magnitude (e.g., `Δtemperature > 5` = temp changed by more than 5°)

## Summary

Change-detection operators provide powerful event-based monitoring:

- ✅ **Simple syntax**: `*field`, `>field`, `<field`
- ✅ **All sensors supported**: Works with every weather field
- ✅ **Compound conditions**: Combine with thresholds and logical operators
- ✅ **Fully tested**: Comprehensive test coverage
- ✅ **Backward compatible**: Existing alarms unaffected
- ✅ **Documented**: Help text in alarm editor
- ✅ **Real-world ready**: Production-tested with lightning, rain, and storm scenarios

Perfect for monitoring:
- ⚡ Lightning strikes
- 🌧️ Rain onset and intensification
- 💨 Wind changes
- 📉 Pressure drops (storm approach)
- 🌡️ Rapid temperature changes
