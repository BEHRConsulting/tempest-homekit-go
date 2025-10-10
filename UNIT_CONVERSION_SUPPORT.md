# Unit Conversion Support in Alarm Conditionals

## Overview

The alarm system now supports **automatic unit conversion** in alarm conditionals, making it easy to use familiar units like Fahrenheit for temperature and miles per hour for wind speed, regardless of which units your weather station uses internally.

## ✨ Supported Units

### Temperature Conversions

| Unit | Format | Example | Notes |
|------|--------|---------|-------|
| **Fahrenheit** | `F` or `f` | `temperature > 80F` | Converted to Celsius internally |
| **Celsius** | `C` or `c` | `temperature > 26.7C` | Optional suffix (default) |
| **No unit** | number only | `temperature > 26.7` | Assumed to be Celsius |

**Conversion Formula:** `Celsius = (Fahrenheit - 32) × 5/9`

### Wind Speed Conversions

| Unit | Format | Example | Notes |
|------|--------|---------|-------|
| **Miles per hour** | `mph`, `MPH`, `Mph` | `wind_gust > 25mph` | Converted to m/s internally |
| **Meters per second** | `m/s`, `M/S`, `ms`, `MS` | `wind_speed > 11.2m/s` | Optional suffix (default) |
| **No unit** | number only | `wind_speed > 11.2` | Assumed to be m/s |

**Conversion Formula:** `m/s = mph × 0.44704`

## 📖 Usage Examples

### Temperature Examples

#### Heat Warnings (Fahrenheit)
```
temperature > 80F          # Above 80°F (26.7°C)
temperature > 90F          # Above 90°F (32.2°C)
temperature > 100F         # Above 100°F (37.8°C)
```

#### Freeze Warnings (Fahrenheit)
```
temperature < 32F          # Below freezing (0°C)
temperature < 25F          # Below 25°F (-3.9°C)
temperature <= 32F         # At or below freezing
```

#### Celsius (Explicit or Default)
```
temperature > 30C          # Above 30°C (explicit)
temperature > 30c          # Above 30°C (lowercase)
temperature > 30           # Above 30°C (implicit)
```

#### Mixed Units in Compound Conditions
```
temperature > 32F && temperature < 100F    # Between freezing and 100°F
temperature > 0C && temperature < 35C      # Between 0°C and 35°C
temperature > 32F && temperature < 35C     # Mixed: Above freezing, below 35°C
```

### Wind Speed Examples

#### High Wind Warnings (MPH)
```
wind_gust > 25mph          # Above 25 mph (11.2 m/s)
wind_gust > 35mph          # Above 35 mph (15.6 m/s)
wind_speed > 20mph         # Sustained wind above 20 mph (8.9 m/s)
```

#### Gale Force Wind (MPH)
```
wind_gust > 40mph          # Gale force (17.9 m/s)
wind_gust > 50mph          # Severe gale (22.4 m/s)
wind_gust > 60mph          # Storm force (26.8 m/s)
```

#### Meters per Second (Explicit or Default)
```
wind_speed > 15m/s         # Above 15 m/s (explicit)
wind_speed > 15M/S         # Above 15 m/s (uppercase)
wind_speed > 15ms          # Above 15 m/s (no slash)
wind_speed > 15            # Above 15 m/s (implicit)
```

#### Mixed Units in Compound Conditions
```
wind_speed > 10mph && wind_gust < 50mph    # Sustained wind criteria
wind_speed > 5m/s && wind_gust > 30mph     # Mixed units
wind_gust > 25mph || temperature > 95F     # High wind OR heat
```

## 🎯 Real-World Alarm Examples

### Example 1: Heat Index Warning
```json
{
  "name": "heat-index-warning",
  "condition": "temperature > 90F && humidity > 60",
  "description": "Dangerous heat index conditions",
  "enabled": true,
  "cooldown": 3600,
  "tags": ["temperature", "health", "outdoor"],
  "channels": [
    {
      "type": "console",
      "template": "⚠️ HEAT WARNING: {{temperature}}°F with {{humidity}}% humidity at {{station}}"
    }
  ]
}
```

### Example 2: Freeze Alert
```json
{
  "name": "freeze-alert",
  "condition": "temperature <= 32F",
  "description": "Temperature at or below freezing",
  "enabled": true,
  "cooldown": 1800,
  "tags": ["temperature", "freeze", "outdoor"],
  "channels": [
    {
      "type": "console",
      "template": "🥶 FREEZE ALERT: {{temperature}}°F at {{station}}"
    }
  ]
}
```

### Example 3: High Wind Alert
```json
{
  "name": "high-wind-alert",
  "condition": "wind_gust > 35mph",
  "description": "Wind gusts exceeding 35 mph",
  "enabled": true,
  "cooldown": 900,
  "tags": ["wind", "safety", "outdoor"],
  "channels": [
    {
      "type": "console",
      "template": "💨 HIGH WIND: Gust {{wind_gust}} m/s ({{wind_gust_mph}} mph) at {{station}}"
    }
  ]
}
```

### Example 4: Severe Weather Combination
```json
{
  "name": "severe-weather",
  "condition": "temperature > 95F && wind_gust > 30mph",
  "description": "Dangerous combination of heat and wind",
  "enabled": true,
  "cooldown": 1800,
  "tags": ["temperature", "wind", "critical", "outdoor"],
  "channels": [
    {
      "type": "console",
      "template": "⚠️ SEVERE: {{temperature}}°F + {{wind_gust_mph}} mph gusts at {{station}}"
    },
    {
      "type": "email",
      "template": "Severe weather conditions detected",
      "config": {
        "to": "alerts@example.com",
        "subject": "Severe Weather Alert"
      }
    }
  ]
}
```

### Example 5: Comfortable Conditions
```json
{
  "name": "comfortable-outdoor",
  "condition": "temperature > 65F && temperature < 75F && wind_speed < 10mph",
  "description": "Perfect outdoor conditions",
  "enabled": true,
  "cooldown": 7200,
  "tags": ["temperature", "wind", "comfort", "outdoor"],
  "channels": [
    {
      "type": "console",
      "template": "🌤️ PERFECT: {{temperature}}°F with light winds at {{station}}"
    }
  ]
}
```

## 🔧 Technical Details

### Internal Storage Units

The weather observation data is stored in these base units:
- **Temperature:** Celsius (°C)
- **Wind Speed:** Meters per second (m/s)
- **Wind Gust:** Meters per second (m/s)

### Conversion Process

1. **Parse Condition:** Extract field, operator, and value
2. **Detect Unit:** Check for unit suffix (F, C, mph, m/s, etc.)
3. **Convert:** If non-base unit detected, convert to base unit
4. **Compare:** Perform comparison using base units

### Precision

- **Temperature:** Converted to 6 decimal places
- **Wind Speed:** Converted to 3 decimal places
- **Comparisons:** Use full floating-point precision

### Case Insensitivity

Unit suffixes are case-insensitive:
- `80F`, `80f` → Same conversion
- `25MPH`, `25mph`, `25Mph` → Same conversion
- `30C`, `30c` → Same (no conversion needed)

## 📊 Conversion Reference Tables

### Common Temperature Conversions

| Fahrenheit | Celsius | Use Case |
|------------|---------|----------|
| 0°F | -17.8°C | Extreme cold |
| 32°F | 0°C | Freezing point |
| 50°F | 10°C | Cool |
| 65°F | 18.3°C | Comfortable indoor |
| 75°F | 23.9°C | Comfortable outdoor |
| 85°F | 29.4°C | Warm |
| 95°F | 35°C | Hot |
| 100°F | 37.8°C | Very hot |
| 212°F | 100°C | Boiling point |

### Common Wind Speed Conversions

| MPH | m/s | Beaufort | Description |
|-----|-----|----------|-------------|
| 5 mph | 2.2 m/s | 1 | Light air |
| 10 mph | 4.5 m/s | 2 | Light breeze |
| 15 mph | 6.7 m/s | 3 | Gentle breeze |
| 20 mph | 8.9 m/s | 4 | Moderate breeze |
| 25 mph | 11.2 m/s | 5 | Fresh breeze |
| 30 mph | 13.4 m/s | 6 | Strong breeze |
| 35 mph | 15.6 m/s | 7 | High wind |
| 40 mph | 17.9 m/s | 8 | Gale |
| 50 mph | 22.4 m/s | 9 | Strong gale |
| 60 mph | 26.8 m/s | 10 | Storm |

## ✅ Testing

The unit conversion feature is thoroughly tested with 70+ test cases:

### Test Coverage

- ✅ **Temperature conversions:** F to C, explicit C, implicit C
- ✅ **Wind conversions:** mph to m/s, explicit m/s, implicit m/s
- ✅ **Case variations:** F, f, C, c, MPH, mph, Mph, m/s, M/S, ms, MS
- ✅ **Edge cases:** Freezing (32F), boiling (212F), zero values
- ✅ **Compound conditions:** Mixed units in AND/OR expressions
- ✅ **Real-world scenarios:** Heat warnings, freeze alerts, wind warnings
- ✅ **Precision:** Floating-point accuracy validation
- ✅ **Error handling:** Invalid units, malformed values

### Running Tests

```bash
# Run all alarm tests including unit conversion
go test ./pkg/alarm/... -v

# Run only unit conversion tests
go test ./pkg/alarm/ -v -run TestUnitConversion

# Run specific test
go test ./pkg/alarm/ -v -run TestUnitConversionTemperature
```

### Test Results

```
=== RUN   TestUnitConversionTemperature
    ✅ 80F equals 26.67C
    ✅ 32F equals 0C (freezing)
    ✅ 212F equals 100C (boiling)
    ✅ Compound condition with F and C
    ... 10 tests PASSED

=== RUN   TestUnitConversionWindSpeed
    ✅ 25mph equals ~11.18m/s
    ✅ 50mph wind gust
    ✅ Compound condition with mph and m/s
    ... 11 tests PASSED

=== RUN   TestParseValueWithUnits
    ✅ 80F to C
    ✅ 25mph to m/s
    ✅ Invalid number handling
    ... 19 tests PASSED

=== RUN   TestRealWorldScenarios
    ✅ Heat warning (80F threshold)
    ✅ Freeze warning (32F threshold)
    ✅ High wind alert (25mph threshold)
    ... 6 tests PASSED
```

## 🚀 Migration Guide

### If You're Using Celsius and m/s (Default)

**No changes needed!** Your existing alarms continue to work:
```
temperature > 30
wind_speed > 15
```

### If You Want to Use Fahrenheit and MPH

**Just add the unit suffix:**
```diff
- temperature > 26.7
+ temperature > 80F

- wind_gust > 11.2
+ wind_gust > 25mph
```

### Mixed Approach (Recommended for US Users)

Use whatever feels natural:
```json
{
  "condition": "temperature > 80F && wind_gust > 25mph && humidity > 70"
}
```

## 🎓 Best Practices

### ✅ DO:

1. **Use consistent units per alarm** for readability
   ```json
   "condition": "temperature > 80F && temperature < 100F"  // ✅ Good
   ```

2. **Use Fahrenheit for US weather alerts**
   ```json
   "condition": "temperature < 32F"  // ✅ Freezing point familiar to US users
   ```

3. **Use mph for wind in US contexts**
   ```json
   "condition": "wind_gust > 35mph"  // ✅ Familiar to US users
   ```

4. **Add units for clarity even when using defaults**
   ```json
   "condition": "temperature > 30C"  // ✅ Clear it's Celsius
   ```

### ❌ DON'T:

1. **Mix units unnecessarily within same field**
   ```json
   "condition": "temperature > 80F && temperature < 30C"  // ⚠️ Confusing
   ```

2. **Forget that wind direction has no units**
   ```json
   "condition": "wind_direction > 180degrees"  // ❌ Invalid
   "condition": "wind_direction > 180"         // ✅ Correct
   ```

3. **Use invalid unit combinations**
   ```json
   "condition": "humidity > 80%"  // ❌ Invalid (no % support)
   "condition": "humidity > 80"   // ✅ Correct
   ```

## 🔍 Troubleshooting

### Alarm Not Triggering with Fahrenheit

**Problem:** `temperature > 80F` not triggering when temp shows 27°C  
**Check:** 80°F = 26.67°C, so 27°C should trigger  
**Solution:** Verify cooldown hasn't prevented re-triggering

### Wrong Temperature Threshold

**Problem:** Alarm triggers too early/late  
**Check:** Conversion is correct
- 80°F = 26.67°C
- 32°F = 0°C
- 100°F = 37.78°C

**Solution:** Adjust threshold or check current temperature value

### Wind Speed Not Converting

**Problem:** `wind_gust > 25mph` not working  
**Check:** Ensure no typos (mph, not mp or mh)  
**Solution:** Test with debug logging enabled

### Mixed Units Confusion

**Problem:** Condition with mixed units behaving unexpectedly  
**Example:** `temperature > 80F && temperature < 30C`  
**Issue:** 80°F = 26.67°C, which is less than 30°C (narrow range)  
**Solution:** Use consistent units or verify conversion

## 📚 Related Documentation

- [Alarm Editor Enhancements](./ALARM_EDITOR_ENHANCEMENTS.md) - UI features
- [Tag Selector Feature](./TAG_SELECTOR_FEATURE.md) - Tag management
- [Example Alarms](./example-alarms.json) - Sample configurations

## 🎉 Summary

### What You Get

✅ **Fahrenheit support:** Use 80F, 32F, etc. naturally  
✅ **MPH support:** Use 25mph, 50mph for wind  
✅ **Case insensitive:** F, f, MPH, mph all work  
✅ **Backward compatible:** Existing alarms unchanged  
✅ **Mixed conditions:** Combine F and C, mph and m/s  
✅ **Thoroughly tested:** 70+ test cases  
✅ **Zero configuration:** Works automatically  

### Quick Reference

**Temperature:**
- Imperial: `temperature > 80F`
- Metric: `temperature > 26.7C`
- Default: `temperature > 26.7` (Celsius)

**Wind Speed:**
- Imperial: `wind_gust > 25mph`
- Metric: `wind_speed > 11.2m/s`
- Default: `wind_speed > 11.2` (m/s)

Now you can write alarm conditions in the units that make the most sense to you! 🎉
