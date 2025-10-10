# Alarm Editor Package

The `editor` package provides a web-based user interface for creating, editing, and managing alarm configurations for the Tempest HomeKit Go application.

## Features

- **Web-based Editor**: Modern, responsive UI for alarm management
- **Search & Filter**: Filter alarms by name or tags
- **CRUD Operations**: Create, read, update, and delete alarms
- **Tag Management**: Organize alarms with tags
- **Live Validation**: Validates alarm conditions in real-time
- **Auto-save**: Saves configuration changes to JSON file
- **Visual Status**: Color-coded status indicators for enabled/disabled alarms

## Usage

Start the alarm editor using the `--alarms-edit` flag:

```bash
# Edit existing alarm file
./tempest-homekit-go --alarms-edit @alarms.json

# Specify custom port (default: 8081)
./tempest-homekit-go --alarms-edit @alarms.json --alarms-edit-port 8082
```

Then open your browser to:
```
http://localhost:8081
```

## API Endpoints

The editor provides the following REST API endpoints:

- `GET /` - Main editor UI
- `GET /api/config` - Get full alarm configuration
- `POST /api/config/save` - Save entire configuration
- `GET /api/alarms` - List alarms (supports `?name=` and `?tag=` filters)
- `POST /api/alarms/create` - Create new alarm
- `POST /api/alarms/update` - Update existing alarm
- `POST /api/alarms/delete?name=<name>` - Delete alarm
- `GET /api/tags` - Get all unique tags
- `POST /api/validate` - Validate alarm condition
- `GET /api/fields` - Get available fields for conditions

## UI Features

### Toolbar
- **Search Box**: Filter alarms by name (case-insensitive)
- **Tag Filter**: Filter alarms by tag
- **New Alarm**: Create a new alarm
- **Save All**: Download configuration as JSON file

### Alarm Cards
Each alarm is displayed as a card showing:
- Status indicator (green = enabled, red = disabled)
- Alarm name and description
- Condition expression
- Tags
- Notification channels
- Edit and Delete buttons

### Alarm Form
The alarm editor modal includes:
- **Name**: Unique alarm identifier (required)
- **Description**: Optional description
- **Condition**: Expression to evaluate (required)
- **Tags**: Comma-separated tags for organization
- **Cooldown**: Time in seconds before alarm can fire again (default: 1800)
- **Enabled**: Toggle alarm on/off

## Architecture

The editor is structured as:

```
pkg/alarm/editor/
├── server.go    # HTTP server and API handlers
└── html.go      # Embedded HTML template
```

### Key Components

**Server**
- Loads and saves alarm configuration from JSON file
- Provides HTTP handlers for all operations
- Validates alarms before saving

**HTML Template**
- Modern, gradient-styled UI
- Responsive design (works on mobile/desktop)
- JavaScript SPA for dynamic interactions
- No external dependencies (standalone HTML)

## Development

The editor is designed to be:
- **Self-contained**: All HTML, CSS, and JavaScript embedded in Go
- **Zero dependencies**: No external web frameworks required
- **Simple**: Single-page application with vanilla JavaScript
- **Fast**: Direct JSON file I/O, no database required

## Example Workflow

1. Start editor: `./tempest-homekit-go --alarms-edit @alarms.json`
2. Open browser to `http://localhost:8081`
3. Click "New Alarm" to create an alarm
4. Fill in the form:
   - Name: "High Temperature"
   - Condition: `temperature > 85`
   - Tags: `temperature, heat`
   - Cooldown: `3600`
5. Click "Save Alarm"
6. Changes are automatically saved to `alarms.json`

## Notes

- The editor operates on the alarm configuration file in real-time
- Changes are saved immediately to disk
- The alarm manager will automatically reload the configuration when file changes are detected (if running with `--alarms` flag)
- The editor runs independently and doesn't require the main service to be running
