# Config File Watching Feature

## Overview

The alarm editor now prominently displays the **active JSON configuration file** that is being watched and edited. This provides clear visibility into which file is being monitored for changes.

## 🎨 Visual Display

### Header Display

The alarm editor header shows the config file path in a styled box:

```
┌─────────────────────────────────────────────────────────┐
│ ⚡ Tempest Alarm Editor                                 │
│ Create and manage weather alarms with real-time         │
│ monitoring                                               │
│                                                          │
│ ┌─────────────────────────────────────────────────────┐│
│ │ 📁 Watching: /path/to/alarms.json                   ││
│ └─────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────┘
```

### Styling Features

- **File icon (📁):** Visual indicator for file path
- **Monospace font:** Easy to read full file paths
- **Light background:** Stands out from header
- **Rounded corners:** Modern, polished appearance
- **Full path display:** Shows complete file system path

## 📂 File Watching Behavior

### Automatic Reload

The alarm manager watches the config file for changes:

1. **File Modified:** Any changes to the JSON file
2. **Manager Detects:** File system notification received
3. **Config Reloaded:** New configuration loaded automatically
4. **Alarms Updated:** Changes take effect immediately

### What Triggers Reload

✅ **Saves from editor:** Changes made in web UI  
✅ **External edits:** Manual edits with text editor  
✅ **File replacement:** Replacing the file entirely  
✅ **Backup restoration:** Restoring from backup  

### Platform Support

| Platform | File Watching | Notes |
|----------|--------------|-------|
| **Linux** | ✅ inotify | Native file watching |
| **macOS** | ✅ FSEvents | Native file watching |
| **Windows** | ✅ ReadDirectoryChangesW | Native file watching |

## 🔍 Configuration Path Sources

### From Command Line

```bash
# Specify config file with @ prefix
./tempest-homekit-go --alarm-config @/path/to/alarms.json

# Full example
./tempest-homekit-go \
  --token your-token \
  --station-id 12345 \
  --alarm-config @/home/user/alarms.json
```

### From Environment

```bash
# Set via environment variable
export ALARM_CONFIG=@/path/to/alarms.json
./tempest-homekit-go
```

### Default Location

If no path specified:
- Linux/macOS: `~/.config/tempest-homekit/alarms.json`
- Windows: `%APPDATA%\tempest-homekit\alarms.json`

## 📊 Use Cases

### Use Case 1: Development

**Scenario:** Testing alarm configurations

```
📁 Watching: /Users/developer/projects/weather/test-alarms.json
```

**Benefits:**
- See exactly which test file is active
- Quick verification you're editing the right file
- Easy to spot wrong file path

### Use Case 2: Production

**Scenario:** Running production alarms

```
📁 Watching: /etc/tempest-homekit/production-alarms.json
```

**Benefits:**
- Confirm production config is loaded
- Document which file to backup
- Clear for operations teams

### Use Case 3: Multiple Stations

**Scenario:** Managing multiple weather stations

```
Station 1: 📁 Watching: /configs/station-backyard.json
Station 2: 📁 Watching: /configs/station-front.json
Station 3: 📁 Watching: /configs/station-roof.json
```

**Benefits:**
- Differentiate between station configs
- Avoid cross-station confusion
- Easy multi-station management

### Use Case 4: Shared Configuration

**Scenario:** Team managing alarms

```
📁 Watching: /shared/weather/team-alarms.json
```

**Benefits:**
- Team knows which file to edit
- No confusion about file locations
- Centralized configuration

## 🛠️ Technical Implementation

### Server-Side

The alarm editor server tracks the config path:

```go
type Server struct {
    configPath string  // Full path to alarm config file
    port       string
}

func NewServer(configPath, port string) (*Server, error) {
    // Remove @ prefix if present
    path := strings.TrimPrefix(configPath, "@")
    
    return &Server{
        configPath: path,
        port:       port,
    }, nil
}
```

### Template Data

Config path is passed to the HTML template:

```go
data := map[string]interface{}{
    "ConfigPath": s.configPath,
}
```

### HTML Rendering

The path is displayed in the header:

```html
<div class="config-path-display">
    <span class="label">📁 Watching:</span>
    <span class="path">{{.ConfigPath}}</span>
</div>
```

### CSS Styling

```css
.config-path-display {
    background: rgba(255,255,255,0.15);
    padding: 12px 16px;
    border-radius: 8px;
    margin-top: 15px;
    font-family: 'Courier New', monospace;
    font-size: 13px;
    display: flex;
    align-items: center;
    gap: 10px;
    border: 1px solid rgba(255,255,255,0.2);
}
```

## 🔄 File Watching Details

### Manager Setup

```go
func (m *Manager) setupFileWatcher() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    
    // Watch the directory (not file directly)
    configDir := filepath.Dir(m.configPath)
    watcher.Add(configDir)
    
    // Start watching in background
    go m.watchConfigFile()
    
    logger.Info("Watching alarm config file: %s", m.configPath)
    return nil
}
```

### Event Handling

```go
func (m *Manager) watchConfigFile() {
    for {
        select {
        case event := <-m.watcher.Events:
            if event.Name == m.configPath {
                if event.Op&fsnotify.Write == fsnotify.Write {
                    m.reloadConfig()
                }
            }
        }
    }
}
```

## 📝 Logging

### Startup Logging

When the alarm system starts:

```
2025-10-09 12:00:00 INFO: Watching alarm config file for changes: /path/to/alarms.json
2025-10-09 12:00:00 INFO: Alarm manager initialized with 5 alarms
2025-10-09 12:00:00 INFO: 4 of 5 alarms are enabled
```

### Change Detection

When config file changes:

```
2025-10-09 12:15:30 INFO: Config file changed, reloading: /path/to/alarms.json
2025-10-09 12:15:30 INFO: Alarm manager initialized with 6 alarms
2025-10-09 12:15:30 INFO: 5 of 6 alarms are enabled
```

### Editor Activity

When editing via web UI:

```
2025-10-09 12:30:00 INFO: Editing: /path/to/alarms.json
2025-10-09 12:30:15 INFO: Saved alarm configuration to: /path/to/alarms.json
```

## 🎯 Benefits

### For Users

1. **Clarity:** Know exactly which file is being monitored
2. **Verification:** Confirm correct file is loaded
3. **Documentation:** Easy to document file locations
4. **Troubleshooting:** Quick path reference for debugging

### For Administrators

1. **Operations:** Clear file path for backups/restores
2. **Multiple Instances:** Differentiate multiple stations
3. **Security:** Verify file permissions on correct file
4. **Auditing:** Document which files are in use

### For Developers

1. **Development:** Switch between test/prod configs easily
2. **Debugging:** Know which file to inspect
3. **Testing:** Verify correct file is loaded in tests
4. **Configuration:** Easy to document requirements

## 🔒 Security Considerations

### File Path Display

**Consideration:** Full path is visible in browser

**Mitigations:**
- Editor should only be accessible on trusted networks
- Use firewall rules to restrict access
- Consider using basepath-relative paths if needed
- Audit log access to editor

### File Permissions

**Recommendations:**
```bash
# Config file should be readable by service
chmod 644 /path/to/alarms.json

# Directory should be writable for auto-save
chmod 755 /path/to/config-dir/

# Owner should be service user
chown tempest:tempest /path/to/alarms.json
```

## 🐛 Troubleshooting

### Path Not Showing

**Problem:** Config path displays as empty  
**Check:** Verify server was initialized with config path  
**Solution:** Pass config path when creating alarm editor server

### Wrong Path Displayed

**Problem:** Shows unexpected file path  
**Check:** Command line args or environment variables  
**Solution:** Verify --alarm-config flag is set correctly

### File Not Updating

**Problem:** Changes to file not reflected in alarms  
**Check:** File watcher logs for errors  
**Solution:** 
1. Verify file permissions
2. Check file watcher is active
3. Restart service if needed

### Path Too Long in UI

**Problem:** Very long paths wrapping awkwardly  
**Workaround:** Use shorter paths or symlinks  
**Future:** Add truncation with tooltip

## 💡 Tips & Tricks

### Use Descriptive Paths

✅ Good:
```
/etc/tempest/backyard-station-alarms.json
/configs/production/weather-alarms.json
/home/user/weather/dev-alarms.json
```

❌ Avoid:
```
/tmp/a.json
/x/y/z/config.json
/home/user/Desktop/untitled.json
```

### Document File Locations

Keep a record of config paths:
```bash
# Production
PROD_ALARMS=/etc/tempest-homekit/production-alarms.json

# Staging
STAGE_ALARMS=/etc/tempest-homekit/staging-alarms.json

# Development
DEV_ALARMS=~/dev/tempest/test-alarms.json
```

### Use Symlinks for Flexibility

```bash
# Create standard location
ln -s /actual/path/to/alarms.json /etc/tempest/alarms.json

# Always reference symlink
--alarm-config @/etc/tempest/alarms.json
```

### Backup Strategy

```bash
# Automated backup with timestamp
cp /path/to/alarms.json \
   /path/to/backups/alarms-$(date +%Y%m%d-%H%M%S).json

# Keep last 30 backups
find /path/to/backups/ -name "alarms-*.json" \
     -mtime +30 -delete
```

## 📚 Related Features

- **File Watching:** Automatic config reload on changes
- **JSON Viewer:** View full config in editor UI
- **Alarm Validation:** Real-time validation before save
- **Tag Management:** Global tag system across all alarms

## 🎉 Summary

### What You See

```
📁 Watching: /path/to/your/alarms.json
```

### What It Tells You

✅ Exact file path being monitored  
✅ File that will be updated on save  
✅ File to backup for disaster recovery  
✅ File that affects active alarms  

### What You Get

✅ **Visibility:** Always know which file is active  
✅ **Confidence:** Verify correct configuration loaded  
✅ **Documentation:** Clear reference for operations  
✅ **Troubleshooting:** Quick path reference for debugging  

The config file path display makes alarm management transparent and professional! 🚀
