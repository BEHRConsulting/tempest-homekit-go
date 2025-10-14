# Quick Reference: Alarm Units

## 🌡️ Temperature

| Write This | Means | Example Alarm |
|------------|-------|---------------|
| `> 80F` | Above 80°F (26.7°C) | Heat warning |
| `< 32F` | Below 32°F (0°C) | Freeze alert |
| `> 100F` | Above 100°F (37.8°C) | Extreme heat |
| `> 30C` | Above 30°C | Metric heat |
| `> 25` | Above 25°C (default) | Simple threshold |

## 💨 Wind Speed

| Write This | Means | Example Alarm |
|------------|-------|---------------|
| `> 25mph` | Above 25 mph (11.2 m/s) | Fresh breeze |
| `> 35mph` | Above 35 mph (15.6 m/s) | High wind |
| `> 50mph` | Above 50 mph (22.4 m/s) | Gale force |
| `> 15m/s` | Above 15 m/s | Metric wind |
| `> 10` | Above 10 m/s (default) | Simple threshold |

## 📋 Common Alarm Conditions

### US Users (Fahrenheit + MPH)
```
temperature > 80F                    ← Heat warning
temperature < 32F                    ← Freeze alert
temperature > 90F && humidity > 70   ← Heat index warning
wind_gust > 35mph                    ← High wind alert
wind_gust > 50mph                    ← Gale warning
temperature > 95F && wind_gust > 30mph  ← Severe weather
```

### International Users (Celsius + m/s)
```
temperature > 26.7C                  ← Heat warning
temperature < 0C                     ← Freeze alert
temperature > 32C && humidity > 70   ← Heat index warning
wind_gust > 15.6m/s                  ← High wind alert
wind_gust > 22.4m/s                  ← Gale warning
temperature > 35C && wind_gust > 13m/s  ← Severe weather
```

## 🔢 Conversion Table

### Temperature (°F ↔ °C)
```
32°F  = 0°C    (Freezing)
50°F  = 10°C   (Cool)
68°F  = 20°C   (Room temp)
80°F  = 26.7°C (Warm)
90°F  = 32.2°C (Hot)
100°F = 37.8°C (Very hot)
```

### Wind Speed (mph ↔ m/s)
```
10 mph = 4.5 m/s   (Light breeze)
20 mph = 8.9 m/s   (Moderate breeze)
25 mph = 11.2 m/s  (Fresh breeze)
35 mph = 15.6 m/s  (High wind)
50 mph = 22.4 m/s  (Gale)
60 mph = 26.8 m/s  (Storm)
```

## ✍️ Syntax Examples

### Valid Temperature Conditions
```
✅ temperature > 80F
✅ temperature > 80f     (lowercase)
✅ temperature > 26.7C
✅ temperature > 26.7c   (lowercase)
✅ temperature > 26.7    (default Celsius)
✅ temp > 80F            (alias)
```

### Valid Wind Conditions
```
✅ wind_gust > 25mph
✅ wind_gust > 25MPH     (uppercase)
✅ wind_gust > 25Mph     (mixed case)
✅ wind_speed > 11.2m/s
✅ wind_speed > 11.2M/S  (uppercase)
✅ wind_speed > 11.2ms   (no slash)
✅ wind_speed > 11.2     (default m/s)
✅ wind > 25mph          (alias)
```

### Invalid Conditions
```
❌ temperature > 80 F    (space before unit)
❌ temperature > 80°F    (degree symbol)
❌ wind_speed > 25 mph   (space before unit)
❌ humidity > 80%        (% not supported)
```

## 🎯 Real-World Examples

### 1. Heat Warning
```json
{
  "name": "heat-warning",
  "condition": "temperature > 90F",
  "cooldown": 3600
}
```

### 2. Freeze Alert
```json
{
  "name": "freeze-alert", 
  "condition": "temperature <= 32F",
  "cooldown": 1800
}
```

### 3. High Wind
```json
{
  "name": "high-wind",
  "condition": "wind_gust > 35mph",
  "cooldown": 900
}
```

### 4. Severe Weather
```json
{
  "name": "severe-weather",
  "condition": "temperature > 95F && wind_gust > 30mph",
  "cooldown": 1800
}
```

### 5. Comfortable Weather
```json
{
  "name": "perfect-outdoor",
  "condition": "temperature > 65F && temperature < 75F && wind_speed < 10mph",
  "cooldown": 7200
}
```

## 🔍 Need More Info?

- **Full Documentation:** See `UNIT_CONVERSION_SUPPORT.md`
- **Config File Info:** See `CONFIG_FILE_WATCHING.md`
- **Examples:** See `example-alarms.json`
- **Editor Guide:** See `ALARM_EDITOR_ENHANCEMENTS.md`
