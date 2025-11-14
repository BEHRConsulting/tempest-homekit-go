# Status Console - Terminal UI Monitor

The Status Console provides a real-time terminal-based interface for monitoring your Tempest weather station without opening a web browser. Built with `tview`, it offers a responsive multi-panel layout with live updates and interactive controls.

## Quick Start

```bash
# Basic usage with default settings
./tempest-homekit-go --status --token "your-token" --station "Your Station"

# Custom refresh and theme
./tempest-homekit-go --status --status-refresh 10 --status-theme dark-cyberpunk \
  --token "your-token" --station "Your Station"

# List all available themes
./tempest-homekit-go --status-theme-list
```

## Display Layout

The status console uses a 7-panel layout that automatically adapts to terminal size:

```
┌────────────────────────────────────────────────────────────────┐
│ Tempest HomeKit v1.9.0                                         │
├──────────────────────────┬─────────────────────────────────────┤
│                          │                                     │
│  Console Logs (3/6)      │  Alarm Status (1/3)                 │
│  [Live log output]       │  [Triggered & cooling alarms]       │
│                          │                                     │
├──────────────────────────┼─────────────────────────────────────┤
│                          │                                     │
│  Tempest Sensors (2/6)   │  HomeKit Status (1/3)               │
│  [Current readings]      │  [Active/disabled + sensors]        │
│                          │                                     │
├──────────────────────────┼─────────────────────────────────────┤
│                          │                                     │
│  Station Status (1/6)    │  System Info (1/3)                  │
│  [Device info]           │  [App metadata]                     │
│                          │                                     │
├──────────────────────────┴─────────────────────────────────────┤
│ Running: 00:05:23 | Refresh: 00:00:02 | Theme: dark-ocean     │
│ q:quit r:refresh t:theme                                       │
└────────────────────────────────────────────────────────────────┘
```

## Panels

### Console Logs
- Real-time application log output
- Color-coded by severity:
  - **ERROR**: Red (critical issues)
  - **WARN**: Yellow (warnings)
  - **INFO**: Green (informational)
  - **DEBUG**: Cyan (detailed debugging)
- Automatic scrolling to latest messages
- ANSI escape sequence stripping for clean display
- Circular buffer (1000 lines)

### Tempest Sensors
All current sensor readings:
- Temperature (°F or °C)
- Relative Humidity (%)
- Wind Speed and Direction (mph/kph, degrees)
- Wind Gust (mph/kph)
- Atmospheric Pressure (mb or inHg)
- UV Index (0-15)
- Light Level (lux)
- Rain Rate (mm/hr or in/hr)
- Daily Rain Accumulation (mm or in)
- Lightning Strike Count
- Lightning Distance (miles or km)

### Station Status
Device and hub information:
- Battery Voltage (e.g., "2.69V") with status (Good/Fair/Low)
- Device Uptime (e.g., "128d 6h 19m")
- Hub Uptime
- Signal Strength (RSSI)
- Firmware Versions (hub and device)
- Serial Numbers
- Data Source (API, web-scraped, UDP, fallback)
- Last update timestamp

### Alarm Status
Active alarm monitoring:
- **Triggered Alarms**: Currently active with trigger time
- **Cooling Down**: Alarms in cooldown period with time remaining
- Alarm Summary: Enabled count / Total count
- Configuration file path
- Last reload timestamp
- "No alarms configured" when alarm system disabled

### HomeKit Status
HomeKit bridge information:
- Status: Active or Disabled
- Published Sensors (when active):
  - Temperature, Humidity, Light, UV, Pressure
- PIN and accessory count (when active)
- "HomeKit services are disabled" when not active

### System Info
Application metadata:
- Application name and version
- Station name
- Units system (imperial/metric/sae)
- Pressure units (inHg/mb)
- Log level (error/warn/info/debug)

### Footer
Real-time counters and controls:
- **Running Time**: Total uptime (hh:mm:ss)
- **Refresh Countdown**: Seconds until next refresh (hh:mm:ss)
- **Current Theme**: Active color theme name
- **Keyboard Shortcuts**: Quick reference

## Keyboard Controls

| Key | Action | Description |
|-----|--------|-------------|
| `q` or `Q` | Quit | Exit status console gracefully |
| `r` or `R` | Refresh | Refresh immediately (reset countdown) |
| `t` or `T` | Theme | Cycle to next color theme |
| `ESC` | Quit | Alternative quit key |
| `Ctrl-C` | Quit | Force quit (signal handler) |

## Configuration

### Command-Line Flags

```bash
--status                    # Enable status console mode
--status-refresh 5          # Refresh interval in seconds (1-3600)
--status-timeout 300        # Auto-exit after 5 minutes (0=never)
--status-theme dark-ocean   # Color theme name
--status-theme-list         # List all themes and exit
```

### Environment Variables

Add to `.env` file:
```bash
STATUS=true
STATUS_REFRESH=10
STATUS_TIMEOUT=0
STATUS_THEME=dark-forest
```

### All Available Options

| Option | Default | Range/Values | Description |
|--------|---------|--------------|-------------|
| `STATUS` | `false` | true/false | Enable status console |
| `STATUS_REFRESH` | `5` | 1-3600 seconds | Update interval |
| `STATUS_TIMEOUT` | `0` | 0-∞ seconds | Auto-exit (0=never) |
| `STATUS_THEME` | `dark-ocean` | See themes below | Color theme |

## Color Themes

### Dark Themes
Optimized for dark terminal backgrounds:

- **dark-ocean** (default) - Deep blue with cyan accents, professional
- **dark-forest** - Forest green with emerald highlights, natural
- **dark-sunset** - Warm amber and orange tones, cozy
- **dark-twilight** - Purple and lavender palette, mystical
- **dark-matrix** - Classic green terminal style, retro
- **dark-cyberpunk** - Neon pink and cyan accents, futuristic

### Light Themes
Optimized for light terminal backgrounds:

- **light-sky** - Sky blue with navy accents, fresh
- **light-garden** - Olive green with earth tones, organic
- **light-autumn** - Rust orange and brown palette, warm
- **light-lavender** - Soft purple and pink tones, gentle
- **light-monochrome** - Clean black and gray, minimalist
- **light-ocean** - Teal and aqua accents, aquatic

### Theme Components

Each theme defines colors for:
- **Title**: Top banner
- **Borders**: Panel outlines
- **Labels**: Field names
- **Values**: Sensor readings
- **Footer**: Bottom status bar
- **Timers**: Countdown displays
- **Log Levels**: ERROR, WARN, INFO, DEBUG
- **Status**: Success, danger, muted states

## Technical Details

### Implementation

**Framework**: `tview` v0.42.0 (Go terminal UI library)
- Built on `tcell` v2.9.0 for terminal control
- Responsive flex-box layout system
- Event-driven architecture

**Concurrency**:
- Context-based goroutine coordination
- Auto-refresh ticker goroutine
- Footer update ticker goroutine
- Synchronized countdown state with `sync.Mutex`
- Clean shutdown with context cancellation

**Data Sources**:
- HTTP API polling to `localhost:8080` (or configured port)
- Endpoints: `/api/weather`, `/api/status`, `/api/alarm-status`
- 500ms timeout per request
- Non-blocking with `app.Draw()` updates

**Log Capture**:
- `io.Pipe` redirects application logs
- Background goroutine reads pipe to buffer
- Thread-safe circular buffer (1000 lines)
- ANSI escape sequence stripping
- Smart log level detection and colorization

### Performance

**Resource Usage**:
- CPU: < 1% on modern systems
- Memory: < 5MB additional overhead
- Network: Minimal (3 HTTP requests per refresh)

**Responsiveness**:
- Non-blocking UI updates
- Immediate keyboard response
- Smooth theme transitions
- Terminal resize handling

### API Integration

The status console polls these internal endpoints:

**GET /api/weather**
```json
{
  "temperature": 24.4,
  "humidity": 66.0,
  "windSpeed": 0.3,
  "windDirection": 241,
  "pressure": 979.7,
  "uv": 2.5,
  "lux": 45000,
  "rainRate": 0.0,
  "rainDaily": 0.0,
  "lightningCount": 0,
  "lightningDistance": 0
}
```

**GET /api/status**
```json
{
  "connected": true,
  "stationStatus": {
    "batteryVoltage": "2.69V",
    "batteryStatus": "Good",
    "deviceUptime": "128d 6h 19m",
    "hubUptime": "63d 15h 55m",
    "dataSource": "web-scraped"
  },
  "homekit": {
    "enabled": true,
    "accessories": ["Temperature", "Humidity", "Light"],
    "pin": "00102003"
  }
}
```

**GET /api/alarm-status**
```json
{
  "enabled": true,
  "configFile": "tempest-alarms.json",
  "lastRead": "2025-11-14T10:30:45Z",
  "alarms": [
    {
      "name": "high-temperature",
      "enabled": true,
      "condition": "temperature > 85",
      "triggered": true,
      "lastTriggered": "2025-11-14T10:25:00Z",
      "cooldownRemaining": 1650
    }
  ]
}
```

## Use Cases

### 1. Quick Status Check
```bash
# Check station for 2 minutes then exit
./tempest-homekit-go --status --status-timeout 120 \
  --token "your-token" --station "Your Station"
```

### 2. Continuous Monitoring
```bash
# Monitor indefinitely with fast refresh
./tempest-homekit-go --status --status-refresh 3 \
  --token "your-token" --station "Your Station"
```

### 3. Debugging Session
```bash
# Debug with verbose logs and light theme
./tempest-homekit-go --status --loglevel debug \
  --status-theme light-monochrome \
  --token "your-token" --station "Your Station"
```

### 4. Alarm Monitoring
```bash
# Monitor alarms with status console
./tempest-homekit-go --status --alarms @alarms.json \
  --token "your-token" --station "Your Station"
```

### 5. Offline Mode
```bash
# UDP stream with status console
./tempest-homekit-go --status --udp-stream --disable-internet \
  --status-theme dark-matrix
```

## Troubleshooting

### Status console not updating

**Symptom**: Panels frozen, no data updates

**Solutions**:
1. Check API token: `--token "your-valid-token"`
2. Verify station name: `--station "Exact Station Name"`
3. Check network: `curl http://localhost:8080/api/weather`
4. Increase refresh: `--status-refresh 10`
5. Check logs: Press `q` to exit and check terminal output

### Colors look wrong

**Symptom**: Unreadable text, poor contrast

**Solutions**:
1. Press `t` to cycle themes until readable
2. Use `--status-theme-list` to preview all themes
3. Match theme to background:
   - Dark background → dark-* themes
   - Light background → light-* themes
4. Try monochrome: `--status-theme light-monochrome`

### Terminal too small

**Symptom**: Layout cramped or overlapping

**Solutions**:
1. Resize terminal to at least 80x24 characters
2. Maximize terminal window
3. Reduce font size to fit more content
4. Use fullscreen terminal mode

### Logs not showing

**Symptom**: Console Logs panel empty

**Solutions**:
1. Adjust log level: `--loglevel info` or `--loglevel debug`
2. Generate activity: Trigger alarms or API errors
3. Check if logs redirect worked: Look for startup message
4. Verify service is running: Check other panels for data

### Can't quit

**Symptom**: Keyboard doesn't work, stuck in console

**Solutions**:
1. Try all quit keys: `q`, `Q`, `ESC`, `Ctrl-C`
2. Terminal force quit: `Cmd-W` (macOS) or close window
3. Kill process: `pkill -f tempest-homekit-go` in another terminal
4. If hung, report bug with reproduction steps

### API timeout errors

**Symptom**: "Request timeout" in Console Logs

**Solutions**:
1. Increase refresh interval: `--status-refresh 10`
2. Check web server port: `--web-port 8080`
3. Verify web server running: `curl http://localhost:8080/api/status`
4. Check for port conflicts: `lsof -i :8080`
5. Reduce network load: Close other applications

## Best Practices

1. **Refresh Interval**: 5-10 seconds for most use cases
2. **Theme Selection**: Match your terminal background
3. **Log Level**: `error` for production, `debug` for troubleshooting
4. **Timeout**: Use for automated monitoring scripts
5. **Combine Features**: Status console + alarms for comprehensive monitoring

## Examples

### Basic Monitoring
```bash
./tempest-homekit-go --status \
  --token "your-token" --station "Your Station"
```

### Production Monitoring
```bash
./tempest-homekit-go --status \
  --status-refresh 10 \
  --status-theme dark-ocean \
  --loglevel error \
  --token "your-token" --station "Your Station"
```

### Debug Session
```bash
./tempest-homekit-go --status \
  --status-refresh 5 \
  --status-theme light-monochrome \
  --loglevel debug \
  --token "your-token" --station "Your Station"
```

### Automated Check
```bash
./tempest-homekit-go --status \
  --status-timeout 300 \
  --status-refresh 15 \
  --token "your-token" --station "Your Station" > status.log 2>&1
```

### Alarm Monitoring
```bash
./tempest-homekit-go --status \
  --status-theme dark-cyberpunk \
  --alarms @alarms.json \
  --loglevel info \
  --token "your-token" --station "Your Station"
```

## See Also

- [README.md](../README.md) - Main documentation
- [REQUIREMENTS.md](../REQUIREMENTS.md) - Technical specifications
- [pkg/status/console.go](../pkg/status/console.go) - Implementation
- [pkg/status/themes.go](../pkg/status/themes.go) - Theme definitions
- [.env.example](../.env.example) - Environment variable configuration
