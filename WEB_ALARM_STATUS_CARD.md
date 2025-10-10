# Web Console Alarm Status Card

## Feature Summary

**Date**: October 9, 2025  
**Purpose**: Display alarm status and configuration in the web console dashboard

## Overview

Added a new "Alarm Status" card to the web console dashboard that displays:
- Alarm system status (enabled/disabled)
- Configuration file information  
- Last configuration read timestamp
- Count of total and enabled alarms
- Detailed list of active alarms with their conditions, last triggered times, and delivery channels

## Implementation Details

### Backend Changes

#### 1. Web Server (`pkg/web/server.go`)

**Added AlarmManagerInterface**:
```go
type AlarmManagerInterface interface {
    GetConfig() *alarm.AlarmConfig
    GetAlarmCount() int
    GetEnabledAlarmCount() int
}
```

**Added to WebServer struct**:
```go
alarmManager AlarmManagerInterface // alarm manager for status display
```

**New Methods**:
- `SetAlarmManager(manager AlarmManagerInterface)` - Connects alarm manager to web server
- `handleAlarmStatusAPI(w http.ResponseWriter, r *http.Request)` - API endpoint for alarm status

**New API Endpoint**:
- `GET /api/alarm-status` - Returns JSON with alarm status and details

**Response Structure**:
```json
{
  "enabled": true,
  "configPath": "Configured",
  "lastReadTime": "2025-10-09 22:13:20",
  "totalAlarms": 2,
  "enabledAlarms": 2,
  "alarms": [
    {
      "name": "Hot outside",
      "description": "Set when temp is > 85F",
      "enabled": true,
      "condition": "temp > 85",
      "tags": ["hot"],
      "channels": ["console", "syslog"],
      "lastTriggered": "Never",
      "cooldown": 1800
    }
  ]
}
```

#### 2. Alarm Types (`pkg/alarm/types.go`)

**Added Method**:
```go
func (a *Alarm) GetLastFired() time.Time
```

Returns the timestamp when an alarm was last triggered. Used by the web console to display "Last Triggered" information.

#### 3. Service Integration (`pkg/service/service.go`)

**Connected alarm manager to web server**:
```go
if alarmManager != nil && webServer != nil {
    webServer.SetAlarmManager(alarmManager)
}
```

This enables the web console to access alarm status without creating dependencies between packages.

### Frontend Changes

#### 1. HTML Card (`pkg/web/server.go` - getDashboardHTML)

**Added Alarm Status Card**:
```html
<div class="card" id="alarm-card">
    <div class="card-header">
        <span class="card-icon">🚨</span>
        <span class="card-title">Alarm Status</span>
    </div>
    <div class="alarm-status-content">
        <div class="alarm-info-row">
            <span class="alarm-label">Status:</span>
            <span class="alarm-value" id="alarm-status">Loading...</span>
        </div>
        <!-- More info rows -->
        <div class="alarm-list" id="alarm-list">
            <div class="alarm-list-header">Active Alarms:</div>
            <!-- Alarm items populated by JavaScript -->
        </div>
    </div>
</div>
```

**Card Placement**: Spans 2 grid columns, positioned before the footer section

#### 2. JavaScript (`pkg/web/static/script.js`)

**New Functions**:

**`fetchAlarmStatus()`**:
- Fetches alarm status from `/api/alarm-status`
- Called on page load and every 10 seconds
- Handles errors gracefully with fallback display

**`updateAlarmStatus(data)`**:
- Updates all alarm status display elements
- Shows status indicator (✅ Active / ⚠️ Not Configured)
- Displays config path and last read time
- Shows alarm counts (enabled/total)
- Populates alarm list with details

**Display Logic**:
- Only shows **enabled** alarms in the list
- Displays "No active alarms configured" when empty
- Each alarm item shows:
  - 🔔 Alarm name
  - Condition expression
  - Last triggered timestamp
  - Delivery channels (comma-separated)

**Polling Integration**:
```javascript
setInterval(() => {
    fetchWeather();
    fetchStatus();
    fetchAlarmStatus(); // Added
}, 10000);
```

#### 3. CSS Styling (`pkg/web/static/styles.css`)

**Card Layout**:
```css
#alarm-card {
    grid-column: span 2; /* Spans 2 columns */
}
```

**Components**:
- `.alarm-status-content` - Main container with flexbox layout
- `.alarm-info-row` - Two-column layout for label/value pairs
- `.alarm-list` - Scrollable container (max 400px height)
- `.alarm-item` - Individual alarm card with hover effects
- `.alarm-item-name` - Bold alarm title with icon
- `.alarm-item-details` - Stacked detail fields

**Visual Features**:
- Border-left accent color matching theme
- Hover animation (translateX with shadow)
- Responsive design (single column on mobile)
- Scrollable list for many alarms
- Semi-transparent backgrounds

## Features

### Status Display

**When Alarms Enabled**:
- ✅ **Status**: Active (green color)
- **Config**: "Configured"
- **Last Read**: Timestamp of last config read
- **Alarm Counts**: "2 / 2 enabled" format

**When Alarms Disabled**:
- ⚠️ **Status**: Not Configured (orange color)
- **Config**: "Not configured"
- **Last Read**: N/A
- **Alarm Counts**: 0 / 0 enabled

### Alarm List

**Shows for Each Alarm**:
1. **Name** with 🔔 icon
2. **Condition** - The expression being evaluated
3. **Last Triggered** - Timestamp or "Never"
4. **Channels** - Comma-separated list of delivery methods

**Example Display**:
```
🔔 Hot outside
Condition: temp > 85
Last: Never
Channels: console, syslog
```

**Interactive Features**:
- Hover effect with subtle animation
- Scrollable list when many alarms
- Real-time updates every 10 seconds

## Usage

### Viewing Alarm Status

1. **Start the service with alarms**:
   ```bash
   ./tempest-homekit-go --token "your-token" --alarms="@tempest-alarms.json"
   ```

2. **Open web console**:
   ```
   http://localhost:8080
   ```

3. **Alarm Status Card**:
   - Automatically displays at bottom of dashboard
   - Refreshes every 10 seconds
   - Shows current status and alarm details

### Monitoring Alarms

**Check Status**: Green ✅ = Active, Orange ⚠️ = Not Configured

**View Last Triggered**:
- See when each alarm last fired
- "Never" if alarm hasn't triggered yet
- Timestamp format: `2025-10-09 22:13:45`

**Track Configuration**:
- Last Read timestamp shows when config was loaded
- Updates automatically when config file changes
- File watcher reloads configuration on save

## API Reference

### GET /api/alarm-status

**Response** (200 OK):
```json
{
  "enabled": true,
  "configPath": "Configured",
  "lastReadTime": "2025-10-09 22:13:20",
  "totalAlarms": 2,
  "enabledAlarms": 2,
  "alarms": [
    {
      "name": "string",
      "description": "string",
      "enabled": true,
      "condition": "string",
      "tags": ["string"],
      "channels": ["string"],
      "lastTriggered": "string",
      "cooldown": 1800
    }
  ]
}
```

**Fields**:
- `enabled` (bool) - Whether alarm system is active
- `configPath` (string) - Config file path (masked for security)
- `lastReadTime` (string) - Timestamp of last config read
- `totalAlarms` (int) - Total number of configured alarms
- `enabledAlarms` (int) - Number of enabled alarms
- `alarms` (array) - List of alarm details

**Alarm Object**:
- `name` - Unique alarm identifier
- `description` - Human-readable description
- `enabled` - Whether alarm is active
- `condition` - Evaluation expression
- `tags` - Categorization tags
- `channels` - Delivery method types
- `lastTriggered` - Last trigger timestamp or "Never"
- `cooldown` - Seconds between repeated notifications

## Security Considerations

**Config Path**: The actual file path is not exposed in the API response. It always returns "Configured" or "Not configured" to avoid leaking filesystem information.

**Read-Only**: The web API only provides status information. Alarm configuration must be edited through:
- Direct file editing
- Alarm editor mode (`--alarms-edit`)

## Testing

### Manual Testing

1. **Start with alarms**:
   ```bash
   ./tempest-homekit-go --use-generated-weather --disable-homekit \
     --alarms="@tempest-alarms.json"
   ```

2. **Verify API**:
   ```bash
   curl http://localhost:8080/api/alarm-status | jq .
   ```

3. **Check web console**:
   - Open http://localhost:8080
   - Scroll to bottom
   - Verify alarm card displays

### Expected Results

**With Alarms**:
- Status: ✅ Active
- Alarm list populated
- Counts accurate
- Last triggered times shown

**Without Alarms**:
- Status: ⚠️ Not Configured
- Empty alarm list
- Counts show 0
- Message: "No active alarms configured"

### Trigger Testing

1. **Create test alarm**:
   ```json
   {
     "name": "Test",
     "enabled": true,
     "condition": "temperature > 0",
     "cooldown": 60,
     "channels": [{"type": "console", "template": "Test"}]
   }
   ```

2. **Wait for observation cycle** (every 60 seconds with generated weather)

3. **Check "Last Triggered"** in web console - should show timestamp

## Responsive Design

**Desktop** (> 768px):
- Card spans 2 columns
- Full-width alarm list
- Side-by-side info rows

**Mobile** (≤ 768px):
- Card spans 1 column (full width)
- Stacked layout
- Scrollable alarm list

## Performance

**API Call Frequency**: Every 10 seconds (same as other status data)

**Data Size**: Minimal
- ~100-200 bytes base response
- ~100-150 bytes per alarm
- Efficient for dozens of alarms

**Rendering**: Fast
- Simple DOM updates
- No complex charts/visualizations
- Minimal CSS transforms

## Future Enhancements

Potential improvements:
1. **Alarm History** - Chart showing trigger frequency over time
2. **Quick Actions** - Enable/disable alarms from web console
3. **Alarm Editor Link** - Direct link to alarm editor mode
4. **Status Indicators** - Color coding based on trigger frequency
5. **Filtering** - Show/hide by tags or enabled status
6. **Sorting** - By name, last triggered, condition
7. **Export** - Download alarm configuration as JSON

## Related Files

### Backend
- `pkg/web/server.go` - Web server and API endpoints
- `pkg/alarm/types.go` - Alarm data structures
- `pkg/alarm/manager.go` - Alarm management logic
- `pkg/service/service.go` - Service integration

### Frontend  
- `pkg/web/static/script.js` - JavaScript fetch and display
- `pkg/web/static/styles.css` - Card styling

### Documentation
- `README.md` - Main documentation (should be updated)
- `ALARM_EDITOR_CHANNEL_FIX.md` - Related alarm fixes

## Troubleshooting

### Card Not Appearing

**Check**:
1. Alarms configured? `--alarms="@file.json"`
2. Service started? `ps aux | grep tempest-homekit-go`
3. Port correct? Default is 8080
4. Browser cache cleared?

### API Returns "enabled: false"

**Causes**:
- No `--alarms` flag provided
- Alarm config file not found
- Alarm config parse error

**Solution**: Check logs for alarm initialization errors

### "Last Triggered" Always Shows "Never"

**Reasons**:
- Alarm condition never met
- Alarm disabled
- Service just started
- Cooldown period active

**Check**: Review alarm condition and current weather values

### Card Display Issues

**CSS Not Loading**:
- Clear browser cache
- Verify `/pkg/web/static/styles.css` exists
- Check browser console for 404 errors

**JavaScript Errors**:
- Open browser developer console
- Look for fetch errors or JavaScript exceptions
- Verify API endpoint returns JSON

## Conclusion

The Alarm Status card provides real-time visibility into the alarm system directly from the web console. It integrates seamlessly with the existing dashboard, uses the same polling mechanism, and maintains consistency with the UI design. The implementation is secure (no config paths exposed), efficient (minimal data transfer), and extensible (easy to add more features).
