# JSON Viewer Feature

## Overview

The alarm editor now includes comprehensive JSON viewing capabilities, allowing you to inspect the raw JSON configuration for the entire alarm system or individual alarms.

## Features

### 1. View Full Configuration JSON

**Location:** Toolbar (top of page)  
**Button:** 📄 View Full JSON

Displays the complete alarm configuration in a modal window with:
- Syntax-highlighted JSON (dark theme)
- Pretty-printed formatting (2-space indentation)
- Scrollable view for large configurations
- Copy to clipboard functionality

**Use Cases:**
- Export entire alarm configuration
- Backup current settings
- Share configuration with team members
- Debug configuration structure
- Migrate to another system

### 2. View Individual Alarm JSON

**Location:** Each alarm card  
**Button:** 📄 JSON (small button in alarm actions)

Displays the JSON for a single alarm including:
- Alarm metadata (name, description, tags)
- Condition configuration
- All delivery channels and their settings
- Cooldown and enabled status

**Use Cases:**
- Inspect alarm configuration details
- Copy specific alarm for reuse
- Troubleshoot channel configurations
- Verify email/SMS settings
- Document alarm setup

### 3. Copy to Clipboard

Both JSON views include a "📋 Copy to Clipboard" button that:
- Copies the formatted JSON to your clipboard
- Shows success notification
- Falls back to legacy copy method for older browsers
- Maintains JSON formatting for paste operations

## User Interface

### JSON Viewer Modal

**Styling:**
- Dark code editor theme (GitHub Dark-inspired)
- Monospace font (Courier New/Consolas/Monaco)
- 900px wide modal for comfortable viewing
- 600px max height with vertical scrolling
- Syntax highlighting (coming in enhanced version)

**Controls:**
- **Close** button - Dismiss the modal
- **Copy to Clipboard** button - Copy JSON to clipboard
- Click outside modal to close (standard modal behavior)

## Usage Examples

### Example 1: Export Full Configuration

1. Click "📄 View Full JSON" in the toolbar
2. Click "📋 Copy to Clipboard"
3. Paste into a text editor
4. Save as `alarms-backup.json`

Result: Complete backup of all alarms.

### Example 2: Copy Alarm for Modification

1. Find the alarm you want to duplicate
2. Click "📄 JSON" on that alarm card
3. Click "📋 Copy to Clipboard"
4. Paste into text editor
5. Modify the JSON (change name, adjust condition)
6. Use the JSON to create a new alarm via file edit

Result: Quickly create similar alarms.

### Example 3: Inspect Email Configuration

1. Find alarm with email delivery
2. Click "📄 JSON" on the alarm card
3. Scroll to "channels" array
4. Find the email channel object
5. Verify `to`, `subject`, `body` settings

Result: Confirm email settings without editing.

### Example 4: Share Configuration with Team

1. Configure alarms in the editor
2. Click "📄 View Full JSON"
3. Click "📋 Copy to Clipboard"
4. Paste into Slack/email/documentation
5. Team members can review before deployment

Result: Collaborative alarm configuration.

## JSON Structure Examples

### Full Configuration JSON
```json
{
  "alarms": [
    {
      "name": "high-temperature",
      "description": "Alert when temperature exceeds 85°F",
      "condition": "temperature > 29.4",
      "tags": ["temperature", "heat"],
      "enabled": true,
      "cooldown": 1800,
      "channels": [
        {
          "type": "console",
          "template": "🌡️ HIGH TEMP: {{temperature}}°C at {{station}}"
        }
      ]
    },
    {
      "name": "heavy-rain",
      "condition": "rain_rate > 5",
      "tags": ["rain"],
      "enabled": true,
      "cooldown": 600,
      "channels": [
        {
          "type": "console",
          "template": "🌧️ HEAVY RAIN: {{rain_rate}} mm/hr"
        },
        {
          "type": "syslog",
          "template": "Heavy rain detected: {{rain_rate}} mm/hr"
        }
      ]
    }
  ]
}
```

### Individual Alarm JSON
```json
{
  "name": "heat-and-humidity",
  "description": "Combined heat and humidity warning",
  "condition": "temperature > 30 && humidity > 70",
  "tags": ["temperature", "humidity", "comfort"],
  "enabled": true,
  "cooldown": 3600,
  "channels": [
    {
      "type": "console",
      "template": "🥵 WARNING: {{temperature}}°C / {{humidity}}%"
    },
    {
      "type": "email",
      "template": "Heat and humidity alert",
      "email": {
        "to": ["admin@example.com"],
        "subject": "Heat & Humidity Alert - {{station}}",
        "body": "Current conditions:\nTemp: {{temperature}}°C\nHumidity: {{humidity}}%\nStation: {{station}}\nTime: {{timestamp}}"
      }
    }
  ]
}
```

## Technical Details

### JavaScript API

```javascript
// Show full configuration
function showFullJSON() {
  const config = { alarms: alarms };
  displayJSON(config, 'Full Configuration JSON');
}

// Show single alarm
function showAlarmJSON(name) {
  const alarm = alarms.find(a => a.name === name);
  displayJSON(alarm, 'Alarm: ' + name);
}

// Generic JSON display
function displayJSON(data, title) {
  document.getElementById('jsonModalTitle').textContent = title;
  const jsonString = JSON.stringify(data, null, 2);
  document.getElementById('jsonContent').textContent = jsonString;
  document.getElementById('jsonModal').classList.add('active');
}

// Copy to clipboard with fallback
async function copyJSON() {
  const jsonText = document.getElementById('jsonContent').textContent;
  try {
    await navigator.clipboard.writeText(jsonText);
    showNotification('JSON copied to clipboard!', 'success');
  } catch (err) {
    // Legacy fallback using execCommand
    // ... fallback code ...
  }
}
```

### CSS Styling

```css
.json-viewer {
  background: #282c34;
  color: #abb2bf;
  padding: 20px;
  border-radius: 8px;
  font-family: 'Courier New', Consolas, Monaco, monospace;
  font-size: 13px;
  overflow-x: auto;
  max-height: 600px;
  line-height: 1.6;
  white-space: pre;
}

.modal-content.wide {
  max-width: 900px;
}
```

## Browser Compatibility

### Clipboard API
- **Modern browsers:** Uses `navigator.clipboard.writeText()` (Chrome 66+, Firefox 63+, Safari 13.1+)
- **Legacy browsers:** Falls back to `document.execCommand('copy')`
- **All browsers:** Success/error notifications provided

### Modal Display
- Uses CSS flexbox for centering
- Responsive design adapts to screen size
- Scrollable content for long JSON
- Click outside to close (standard UX)

## Best Practices

### When to Use JSON View

✅ **Good Use Cases:**
- Backing up configurations before major changes
- Documenting alarm setup in tickets/wikis
- Debugging complex channel configurations
- Sharing alarms with other team members
- Learning the JSON structure
- Migrating between environments

❌ **Not Recommended:**
- Editing JSON directly in the viewer (read-only)
- Replacing the visual editor for simple changes
- Storing sensitive credentials (use environment variables)

### Security Considerations

⚠️ **Important:** The JSON viewer displays all configuration including:
- Email addresses in channel configurations
- SMS phone numbers
- Template strings
- Any custom settings

**Recommendations:**
- Don't share JSON with untrusted parties
- Redact sensitive information before posting publicly
- Use environment variables for credentials (not in JSON)
- Review JSON before copying to shared locations

### Performance

The JSON viewer is optimized for:
- ✅ Configurations with up to 100 alarms
- ✅ Individual alarms with multiple channels
- ✅ Large template strings
- ⚠️ Very large configurations (>500 alarms) may be slow to render

For very large configurations, consider:
- Viewing individual alarms instead of full config
- Filtering alarms before viewing
- Using the API endpoints directly

## Troubleshooting

### JSON Not Displaying

**Problem:** Modal shows but JSON content is empty  
**Solution:** Refresh the page to reload alarm data

**Problem:** "undefined" appears in JSON  
**Solution:** Some alarms may have missing fields (normal), just means that field isn't set

### Copy to Clipboard Fails

**Problem:** "Failed to copy JSON" notification  
**Solution:** 
1. Check browser permissions for clipboard access
2. Try manually selecting text and using Ctrl+C / Cmd+C
3. Update your browser to latest version

### Modal Won't Close

**Problem:** Can't dismiss the JSON viewer  
**Solution:**
1. Press ESC key
2. Click the "Close" button
3. Click outside the modal (on the dark overlay)
4. Refresh the page if stuck

## Future Enhancements

Planned improvements for JSON viewer:
- [ ] Syntax highlighting with colors for keys, strings, numbers, booleans
- [ ] Expand/collapse JSON nodes for large configurations
- [ ] Download as file button (direct file save)
- [ ] Side-by-side diff view when editing
- [ ] JSON validation with error highlighting
- [ ] Import JSON feature (paste to create/update alarms)
- [ ] Export filtered alarms (by tag or search)

## API Endpoints

The JSON viewer uses existing API endpoints:

```
GET /api/config
Returns: Full alarm configuration with all alarms

GET /api/alarms/list  
Returns: Array of all alarms (same data, different format)
```

No additional API endpoints are required for the JSON viewer.
