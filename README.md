# Tempest HomeKit Go

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org) [![Test Coverage](https://img.shields.io/badge/coverage-59.1%25-yellow?style=flat)](./coverage.out) [![Build](https://github.com/BEHRConsulting/tempest-homekit-go/actions/workflows/ci.yml/badge.svg)](https://github.com/BEHRConsulting/tempest-homekit-go/actions) [![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

Tempest HomeKit Go is a Go service that reads WeatherFlow Tempest station data and exposes sensors to Apple HomeKit while providing a modern web dashboard. It supports UDP stream mode, historical preloading, alarm notifications, and optional device status scraping.

Table of Contents
- [Quick Start](#quick-start)
- [Features](#features)
- [Configuration](#configuration)
- [Roadmap](#roadmap)
- [Contributing](#contributing)

<!-- Version history moved to VERSIONS.md -->

## Research Methodology: Vibe Programming

### Definition of Vibe Programming

**Vibe Programming** represents a novel software development methodology that leverages the intuitive, context-aware capabilities of Large Language Models (LLMs) to enable rapid prototyping and iterative development through natural language interaction. This approach emphasizes:

- **Contextual Understanding**: LLMs maintain awareness of project architecture, requirements, and existing codebase
- **Iterative Refinement**: Continuous feedback loops between developer intent and AI-generated implementations
- **Natural Language Specifications**: Requirements expressed in conversational terms rather than formal specifications
- **Emergent Architecture**: System design evolving organically through AI-assisted exploration of possibilities
- **Real-time Problem Solving**: Immediate debugging and enhancement through conversational programming

### Technical Implementation Environment

**Primary Development Tools:**
- **IDE**: Visual Studio Code on macOS (Apple Silicon)
- **AI Assistants**: 
  - **Claude Sonnet 4.5** - Primary architectural design and complex problem resolution
  - **GitHub Copilot with Grok Code Fast 1 (preview)** - Code completion and rapid prototyping
  - **GPT-5 mini** - Auxiliary AI assistant used for targeted code edits and documentation updates
- **Version Control**: Git with GitHub integration
- **Platform**: macOS development environment with cross-platform deployment

### Vibe Programming Validation Results

This project successfully demonstrates the efficacy of vibe programming techniques in producing production-ready software with:
- **Rapid Development Cycle**: Complete application developed through iterative AI assistance
- **Adaptive Problem Solving**: Real-time debugging and feature enhancement through conversational programming
- **Quality Assurance**: 60.3% test coverage achieved through AI-assisted test generation
- **Professional Standards**: Production-ready deployment with cross-platform service management

## Important Sensor Notes

⚠️ **HomeKit Sensor Compliance**: Due to HomeKit's limited native sensor types, the **Pressure** and **UV Index** sensors use the standard HomeKit **Light Sensor** service for compliance. In the Home app, these will appear as "Light Sensor" with units showing as "lux" - **please ignore the "lux" unit** for these sensors as they represent atmospheric pressure (mb) and UV index values respectively. This is a HomeKit limitation, not an application issue.

🏠 **Web Console Only Mode**: This application can be run with HomeKit services completely disabled by using the `--disable-homekit` flag. In this mode, only the web dashboard will be available, providing a lightweight weather monitoring solution without HomeKit integration.

## Contributors

- **Kent** - Principal Investigator, Vibe Programming methodology implementation

## Public release notes

This repository is prepared for public GitHub release as a Vibe Programming research project. Key discovery keywords included in this repository are: `vibe`, `macOS`, `HomeKit`, `tempest`, `weather`, `TempestWX`, and `WeatherFlow` to improve discoverability.

- Project status: Work in progress (stable) — feature-complete for the core functionality described in this README, actively maintained and under continued test/coverage improvements.
- Authors: Kent and contributors listed above. This project was developed using Vibe Programming techniques with AI-assisted development tools.

If you use or contribute to this project, please follow the contributing guidelines and include references to the Vibe Programming methodology in PR descriptions when changes are research-related.

## Acknowledgments

We acknowledge the human contributors and AI assistants who supported this project:

- Human contributors: Kent
- AI assistants: Claude Sonnet 3.5, GitHub Copilot (Grok Code Fast 1 preview), GPT-5 mini

### Vibe Programming Methodology Validation

This project represents a controlled experiment in AI-assisted software development, demonstrating the practical application of conversational programming techniques in production software development.

For a concise history of versions and notable changes, see `VERSIONS.md`.

- **Real-time Weather Monitoring**: Continuously polls WeatherFlow Tempest station data every 60 seconds
- **HomeKit Integration**: Individual HomeKit accessories for each weather sensor
- **Multiple Sensor Support**: Temperature, Humidity, Wind Speed, Wind Direction, Rain Accumulation, UV Index, Pressure, and Ambient Light
- **Modern Web Dashboard**: Interactive web interface with real-time updates, unit conversions, and professional styling
  - **External JavaScript Architecture**: Clean separation of concerns with all JavaScript externalized to `script.js`
  - **Interactive Chart Pop-out System**: Advanced data visualization with expandable chart windows
    - **80% Screen Coverage**: Pop-out windows automatically sized to 80% of screen dimensions
    - **Resizable & Draggable**: Native browser window controls for optimal user experience
    - **Complete Historical Data**: Each pop-out displays full 1000+ point datasets with proper legends
    - **Professional Styling**: Gradient backgrounds with clean chart containers and interactive controls
    - **Multi-chart Support**: Temperature, humidity, wind, rain, pressure, light, and UV charts
  - **Pressure Analysis System**: Advanced pressure forecasting with trend analysis and weather predictions
  - **Interactive Info Icons**: Clickable info icons (ℹ️) with detailed tooltips for pressure calculations and sensor explanations
  - **Consistent Positioning**: All tooltips positioned with top-left corner aligned with bottom-right of info icons
  - **Rain Info Icon Fix**: Resolved JavaScript issue where unit updates removed the rain info icon
  - **Proper Event Handling**: Enhanced event propagation control to prevent unit toggle interference
  - **UV Index Display**: Complete UV exposure categories using NCBI reference data with EPA color coding
  - **Interactive Tooltips**: Information tooltips for all sensors with standardized positioning
  - **Accessories Status**: Real-time display of enabled/disabled sensor status in HomeKit bridge card
- **Cross-platform Support**: Runs on macOS, Linux, and Windows with automated service installation
- **TempestWX Device Status Scraping** (Optional):
  - **Headless Browser Integration**: Uses Chrome/Chromium to scrape detailed device status from TempestWX
  - **15-Minute Periodic Updates**: Background scraping with automatic caching
  - **Comprehensive Device Data**: Battery voltage, uptime, signal strength, firmware versions, serial numbers
  - **Multiple Fallback Layers**: Headless browser → HTTP scraping → API fallback for reliability
  - **Data Source Transparency**: Clear indication of data source (web-scraped, http-scraped, api, fallback)
  - **Enable with `--use-web-status` flag**: Optional enhancement for users who want detailed device monitoring
- **UDP Stream Feature** (Offline Mode):
  - **Local Network Monitoring**: Listen for UDP broadcasts from Tempest hub on port 50222
  - **Offline Operation**: Enables weather monitoring during internet outages without API access
  - **Real-time Updates**: Process observation messages as they're broadcast every minute
  - **No API Token Required**: Works entirely on local network without WeatherFlow cloud services
  - **Multiple Message Types**: Supports obs_st (Tempest), obs_air, obs_sky, rapid_wind, device_status, hub_status
  - **Enable with `--udp-stream` flag**: Monitor Tempest station locally without internet connectivity
  - **Full Offline Mode with `--disable-internet`**: Disables all internet access for complete offline operation
- **Flexible Configuration**: Command-line flags and environment variables for easy deployment
- **Enhanced Debug Logging**: Multi-level logging with emoji indicators, calculated values, API calls/responses, and comprehensive DOM debugging

## Roadmap

Planned enhancements and strategic priorities for upcoming releases. Items are grouped by priority and include brief implementation notes and suggested CLI/environment configuration where relevant.

High priority
- Alarms and Notifications
  - Description: Rule-based alerting system that triggers notifications when weather conditions meet configured criteria (thresholds, combinations, or event patterns).
  - Channels: Console logging, Syslog, OSLog (macOS unified logging), Email, SMS, and an internal EventManager (webhook/HTTP). Support templated messages with runtime value interpolation.
  - CLI/Env: `--alarms @alarms.json` or `--alarms '{...json...}'`. File-watcher to auto-reload alarm configuration.
  - Notes: Provide a secure secrets mechanism for SMTP/SMS credentials (use `.env` or platform secrets). Include an interactive web-based alarm editor behind `--alarms-edit` for ease of configuration.

- Multi-station Monitoring
  - Description: Allow monitoring of multiple Tempest stations from a single instance. Each station has its own data source, history buffer, and HomeKit grouping.
  - CLI/Env: `--stations config.json` or multiple `--station-url` entries. Provide station tagging and per-station sensor enable/disable.
  - Notes: Architect with per-station goroutines, scoped caches, and aggregated UI views.

Medium priority
- Container & Serverless Deployment
  - Docker: Provide a lightweight Docker image with recommended environment variables and a `Dockerfile` producing a small, reproducible artifact.
  - AWS Lambda (Serverless): Provide a design for a Lambda-friendly mode (stateless handler + external storage for history); `--lambda-mode` for event-driven execution. Note: HomeKit functionality is not supported in serverless mode — this mode targets data ingestion, web dashboard, and notifications.
  - Notes: Include example `docker-compose.yml` for local multi-container setups (web + optional SMTP/SMS mock) and CI build steps for publishing images.

- Notification Integrations
  - Email: SMTP with optional OAUTH2 support for providers like Microsoft 365 / Google Workspace.
  - SMS: Pluggable providers (Twilio, Amazon SNS); supply a provider interface and sample provider configs in `.env.example`.
  - Syslog & EventManager: Local syslog emission and webhook-based event forwarding for enterprise integrations.

Longer-term / stretch goals
- Alarm editor: Interactive web UI to author and validate alarms with templates and live preview.
- Advanced rules engine: Support for boolean logic, time windows, and rate-limiting to prevent notification storms.
- Multi-tenant UI: Role-based access controls and per-tenant station configuration for managed deployments.

Contributing / Implementation notes
- Priorities: Focus test coverage and robust unit tests for `pkg/service` (multi-station orchestration) and `pkg/weather` (data parsing & UDP ingestion) before adding large integration features.
- CLI flags and env variables should be documented in `REQUIREMENTS.md` and `README.md` as features are implemented. Use `.env.example` for credential scaffolding.
- Backward compatibility: Add feature flags (`--enable-alarms`, `--multi-station`) to avoid surprising default behavior.

## Alarms and Notifications

The alarm system enables rule-based weather alerting with multiple notification channels. Configure alarms to trigger when weather conditions meet specific criteria (temperature thresholds, lightning proximity, rain events, etc.).

**Supported Notification Channels:**
- **Console**: Log messages to stdout (always visible regardless of log level)
- **Syslog**: Local or remote syslog server
- **OSLog**: macOS unified logging system (os_log API via CGO)
- **Email**: SMTP or Microsoft 365 OAuth2
- **SMS**: ✅ **AWS SNS** (fully implemented) | ⏳ Twilio (coming soon)
- **EventLog**: System event log (Windows) or syslog (Unix)

**Features:**
- Flexible condition syntax: `temperature > 85`, `humidity > 80 && temperature > 35`, `lux > 10000 && lux < 50000`
- **Change detection operators**: `*field` (any change), `>field` (increase), `<field` (decrease)
  - Example: `*lightning_count` triggers on any lightning strike
  - Example: `>rain_rate` triggers when rain intensifies
  - Example: `<lightning_distance` triggers when lightning gets closer
- Template-based messages with runtime value interpolation (`{{temperature}}`, `{{timestamp}}`, etc.)
- Cooldown periods to prevent notification storms
- Cross-platform file watching for live configuration reloads
- Per-alarm tags for easy filtering and organization
- **Web console alarm status card**: View alarm status, last triggered times, and configuration directly in the dashboard

**Quick Start:**
```bash
# Run with alarm configuration
./tempest-homekit-go --token "your-token" --alarms @alarms.json

# Test email configuration before deploying
./tempest-homekit-go --email-test --alarms @alarms.json

# Edit alarm configuration (standalone editor mode)
./tempest-homekit-go --alarms-edit @alarms.json --alarms-edit-port 8081
```

**Example Alarm Configuration Files:**
- `alarms.example.json` - Complete alarm examples (works with any provider)
- `alarms-ms365.example.json` - Same alarms, shows MS365 setup instructions
- `alarms-aws.example.json` - Same alarms, shows AWS SNS setup instructions

**Important:** All email/SMS credentials are configured in `.env` file only - NOT in alarm JSON files! The alarm JSON files contain only alarm rules. Configure your provider credentials in `.env` (see `.env.example` for details), then use any of the alarm example files above.

### Testing Email Configuration

Before deploying alarms in production, test your email configuration:

```bash
./tempest-homekit-go --test-email user@example.com --alarms @alarms.json
```

The email test will:
1. Validate email provider configuration (Microsoft 365 OAuth2 or SMTP)
2. Check all required credentials from `.env` file
3. Send a test email to the specified address with:
   - Application name and version
   - Timestamp and command line options
   - Email configuration details
   - Current weather data from your station
5. Provide troubleshooting guidance if delivery fails

### Testing SMS Configuration

Before deploying SMS alarms in production, test your SMS configuration:

```bash
./tempest-homekit-go --test-sms +15555551234 --alarms @alarms.json
```

The SMS test will:
1. Validate SMS provider configuration (Twilio or AWS SNS)
2. Check all required credentials from `.env` file
3. Send a test SMS to the specified phone number with:
   - Application name and version
   - Timestamp
   - SMS provider information
5. Provide troubleshooting guidance if delivery fails

**Microsoft 365 Setup:**
For complete Microsoft 365 OAuth2 setup instructions, see the detailed comments in `.env.example`. You'll need:
- Azure AD app registration
- Client ID, Client Secret, and Tenant ID
- Mail.Send API permission with admin consent
- From address (`MS365_FROM_ADDRESS`)

**SMTP Setup:**
For generic SMTP providers (Gmail, SendGrid, Mailgun, etc.), configure:
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`
- `SMTP_FROM_ADDRESS`, `SMTP_USE_TLS=true`

See `.env.example` for provider-specific examples with standard ports and TLS settings.

**AWS SNS SMS Setup:**
For AWS SNS configuration (recommended for production SMS), see detailed setup instructions in `.env.example`. Quick setup:
1. Create IAM user with `sns:Publish` permission (principle of least privilege)
2. Configure credentials in `.env`: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`
3. Run the automated setup script: `./scripts/setup-aws-sns.sh`
4. The script will:
   - Verify AWS CLI credentials (uses your admin credentials from `~/.aws/`)
   - Configure production SMS settings (type, spending limits)
   - Optionally create SNS topics with subscriptions
   - Update `.env` with Topic ARN automatically
   - Send test SMS for verification

**Important**: The AWS credentials in `.env` are for the **application runtime user** (limited permissions). The setup script uses your **admin AWS CLI credentials** from `~/.aws/credentials` or `aws configure`.

**Twilio SMS Setup:**
For Twilio SMS configuration (great for development and moderate volume), see detailed setup instructions in `.env.example`. Quick setup:
1. Sign up for Twilio: https://www.twilio.com/try-twilio (get $15 trial credit)
2. Get credentials from Twilio Console: https://console.twilio.com/
   - Account SID (starts with "AC")
   - Auth Token (click "Show" to reveal)
3. Purchase a phone number with SMS capability from the Twilio Console
4. Configure in `.env`: `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_FROM_NUMBER` (E.164 format: +1XXXXXXXXXX)
5. Test configuration: `./tempest-homekit-go --test-sms +15555551234 --alarms @alarms.json`

**Note**: Trial accounts can only send to verified phone numbers. To verify a number:
1. Go to Twilio Console → Phone Numbers → Verified Caller IDs
2. Click "+" to add a new number
3. Enter the number and verify via SMS or call

**Pricing**: ~$0.0079 per message (US), ~$1/month for phone number. Upgrade to paid account for unrestricted sending.

### Using the Alarm Editor

The alarm editor provides a modern web interface for managing alarm configurations:

1. **Start the editor:**
   ```bash
   ./tempest-homekit-go --alarms-edit @alarms.json
   ```

2. **Open your browser to** `http://localhost:8081` (or custom port with `--alarms-edit-port`)

3. **Editor features:**
   - **Search & filter**: Find alarms by name or tag
   - **Create alarms**: Click "New Alarm" button to add alarms
   - **Edit alarms**: Click "Edit" on any alarm card
   - **Delete alarms**: Click "Delete" on any alarm card
   - **Visual status**: Green dot = enabled, red dot = disabled
   - **Live validation**: Conditions are validated before saving
   - **Auto-save**: Changes saved immediately to JSON file

4. **Alarm form fields:**
   - **Name**: Unique identifier (required)
   - **Description**: Optional description
   - **Condition**: Expression like `temperature > 85` or `humidity > 80 && temperature > 35` (required)
   - **Tags**: Comma-separated tags for organization
   - **Cooldown**: Seconds before alarm can fire again (default: 1800)
   - **Enabled**: Toggle alarm on/off

The editor operates independently from the main service and saves changes directly to your alarm configuration file. If the main service is running with `--alarms`, it will automatically detect and reload the configuration when changes are saved.

### Web Console Alarm Status

When running the main service with alarms enabled, the web dashboard (`http://localhost:8080`) automatically displays an **Alarm Status** card showing:

- **System Status**: Active/Not Configured indicator
- **Configuration File**: Name of the alarm configuration file being monitored
- **Last Read**: Timestamp when the configuration was last loaded (updates only on file changes)
- **Alarm Counts**: Number of enabled alarms vs total alarms
- **Active Alarms List**: Details for each enabled alarm:
  - Alarm name and condition
  - Last triggered timestamp (or "Never")
  - Delivery channels (console, syslog, oslog, email, SMS, eventlog)

The alarm status refreshes automatically every 10 seconds, providing real-time visibility into your alarm system without needing to open the alarm editor or check log files.

**Example:**
```bash
# Start service with alarms - web console will show alarm status card
./tempest-homekit-go --token "your-token" --alarms @tempest-alarms.json

# Open http://localhost:8080 to view dashboard with alarm status
```

## Quick Start

### Prerequisites
- Go 1.24.2 or later
- WeatherFlow Tempest station with API access
- Apple device with HomeKit support
- Google Chrome (optional, for detailed device status via `--use-web-status`)

### Build and Run
```bash
git clone https://github.com/BEHRConsulting/tempest-homekit-go.git
cd tempest-homekit-go
go build
./tempest-homekit-go --token "your-api-token"
```

### Test with Generated Weather
```bash
# Traditional approach
./tempest-homekit-go --use-generated-weather

# New flexible station URL approach
./tempest-homekit-go --station-url http://localhost:8080/api/generate-weather

# Using environment variable (equivalent to above)
STATION_URL=http://localhost:8080/api/generate-weather ./tempest-homekit-go

# With historical data preloading (preloads up to HISTORY_POINTS observations)
./tempest-homekit-go --use-generated-weather --read-history  # preloads up to HISTORY_POINTS observations
```

### Cross-Platform Build (All Platforms)
```bash
./scripts/build.sh
```

### Install as System Service
```bash
sudo ./scripts/install-service.sh --token "your-api-token"
```

## Installation

### Option 1: Build from Source
```bash
git clone https://github.com/BEHRConsulting/tempest-homekit-go.git
cd tempest-homekit-go
go mod tidy
go build -o tempest-homekit-go
```

### Option 2: Platform-Specific Build (Current Platform Only)
```bash
./scripts/build.sh
```
This builds only for your current platform (macOS binaries on macOS, Linux on Linux, etc.).

### Option 3: Cross-Platform Build (All Platforms)
```bash
./scripts/build-cross-platform.sh
```
This builds optimized binaries for Linux, macOS, and Windows from any platform.

### Option 3: Install as Service
For production deployment, install as a system service:
```bash
# Linux (systemd)
sudo ./scripts/install-service.sh --token "your-api-token"

# macOS (launchd)
sudo ./scripts/install-service.sh --token "your-api-token"

# Windows (NSSM)
./scripts/install-service.sh --token "your-api-token"
```

### Dependencies
- `github.com/brutella/hap` - Modern HomeKit Accessory Protocol implementation (v0.0.32)
- `github.com/chromedp/chromedp` - Headless browser automation for TempestWX status scraping
- Custom weather services with unique UUIDs to prevent temperature conversion issues

## Usage

### Basic Usage
If you are using the WeatherFlow Tempest API (default behavior), provide your API token with `--token` or the `TEMPEST_TOKEN` environment variable. If you instead use a custom station URL via `--station-url` or enable generated weather with `--use-generated-weather`, a WeatherFlow API token is not required.

```bash
# WeatherFlow API (requires token)
./tempest-homekit-go --token "your-weatherflow-token"

# Custom station URL (no WeatherFlow token required)
./tempest-homekit-go --station-url http://localhost:8080/api/generate-weather

# Generated weather (no WeatherFlow token required)
./tempest-homekit-go --use-generated-weather
```

### Configuration Options

#### Command-Line Flags (alphabetical order)
- `--alarms`: Alarm configuration: @filename.json or inline JSON string (default: none). Env: ALARMS
- `--alarms-edit`: Run alarm editor for specified config file: @filename.json (default: none)
- `--alarms-edit-port`: Port for alarm editor web UI (default: 8081). Env: ALARMS_EDIT_PORT
- `--cleardb`: Clear HomeKit database and reset device pairing
- `--disable-homekit`: Disable HomeKit services and run web console only
- `--elevation`: Station elevation in meters (default: auto-detect, valid range: -430m to 8848m)
- `--env`: Custom environment file to load (default: ".env"). Env: ENV_FILE
  - Overrides the default `.env` file location
  - Useful for multiple configurations or deployment environments
  - Example: `./tempest-homekit-go --env /etc/tempest/production.env`
- `--loglevel`: Logging level - debug, info, warn/warning, error (default: "error")
- `--logfilter`: Filter log messages to only show those containing this string (case-insensitive) - useful for targeted debugging
- `--pin`: HomeKit pairing PIN (default: "00102003")  
- `--sensors`: Sensors to enable - 'all', 'min' (temp,lux,humidity), or comma-delimited list with aliases supported:
  - **Temperature**: `temp` or `temperature`
  - **Light**: `lux` or `light`
  - **UV**: `uv` or `uvi`
  - **Other sensors**: `humidity`, `wind`, `rain`, `pressure`, `lightning`
  - (default: "temp,lux,humidity")
- `--station`: Tempest station name (default: "Chino Hills")
-- `--station-url`: Custom station URL for weather data (e.g., `http://localhost:8080/api/generate-weather`). Overrides Tempest API
-- `--token`: WeatherFlow API access token (required when using the WeatherFlow API as the data source)
- `--units`: Units system - imperial, metric, or sae (default: "imperial")
- `--units-pressure`: Pressure units - inHg or mb (default: "inHg")
- `--udp-stream`: Enable UDP broadcast listener for local station monitoring (port 50222)
- `--disable-internet`: **Offline Mode** - Disables all internet connectivity for complete offline operation
  - **Alias**: `--no-internet` (backward compatibility)
  - **Requires**: `--udp-stream` or `--use-generated-weather` (must have a local data source)
  - **Incompatible with**: `--use-web-status`, `--read-history` (both require internet access)
  - **Use Case**: Internet outages, air-gapped systems, privacy-focused deployments, testing without network
  - **Limitations**: No forecast data, no historical preloading, no web scraping
- `--disable-webconsole`: **HomeKit Only Mode** - Disables the web dashboard server
  - **Incompatible with**: `--disable-homekit` (cannot disable both HomeKit and web console)
  - **Use Case**: Minimal resource usage, HomeKit-only deployments, reduced attack surface
- `--use-generated-weather`: Use simulated weather data for testing (automatically sets station-url)
- `--use-web-status`: Enable headless browser scraping of TempestWX status page every 15 minutes (requires Chrome, incompatible with `--disable-internet`)
- `--version`: Show version information and exit
- `--web-port`: Web dashboard port (default: "8080")

#### Environment Variables
- `TEMPEST_TOKEN`: WeatherFlow API token
- `TEMPEST_STATION_NAME`: Station name
- `STATION_URL`: Custom station URL for weather data (overrides Tempest API)
- `UDP_STREAM`: Enable UDP broadcast listener (true/false)
- `DISABLE_INTERNET`: Disable all internet access (true/false, requires UDP_STREAM=true or USE_GENERATED_WEATHER=true)
- `DISABLE_WEBCONSOLE`: Disable web dashboard server (true/false)
- `HOMEKIT_PIN`: HomeKit PIN
- `LOG_LEVEL`: Logging level
- `SENSORS`: Sensors to enable (default: "temp,lux,humidity")
- `UNITS`: Units system - imperial, metric, or sae (default: "imperial")
- `UNITS_PRESSURE`: Pressure units - inHg or mb (default: "inHg")
- `WEB_PORT`: Web dashboard port

### Example with Full Configuration
```bash
./tempest-homekit-go \
  --token "your-api-token" \
  --station "Your Station Name" \
  --pin "12345678" \
  --web-port 8080 \
  --loglevel info \
  --sensors "temp,humidity,lux,uv,pressure" \
  --elevation 150 \
  --use-web-status
```

### Sensor Configuration Examples
```bash
# Using sensor aliases (recommended for readability)
./tempest-homekit-go --token "your-token" --sensors "temperature,light,uvi"

# Traditional sensor names (also supported)
./tempest-homekit-go --token "your-token" --sensors "temp,lux,uv"

# Mixed aliases and traditional names
./tempest-homekit-go --token "your-token" --sensors "temperature,humidity,light,wind"

# All available sensors
./tempest-homekit-go --token "your-token" --sensors "all"

# Minimal sensor set
./tempest-homekit-go --token "your-token" --sensors "min"
```

### Offline Mode Examples
```bash
# Valid: Full offline mode with UDP stream (real station)
./tempest-homekit-go --token "your-token" --udp-stream --disable-internet

# Valid: Full offline mode with generated weather (testing/simulation)
./tempest-homekit-go --disable-internet --use-generated-weather

# Valid: Offline with custom sensors
./tempest-homekit-go --token "your-token" --udp-stream --disable-internet --sensors "temp,humidity,pressure"

# Invalid: Missing data source
./tempest-homekit-go --token "your-token" --disable-internet
# ERROR: --disable-internet mode requires --udp-stream or --use-generated-weather (need a local data source)

# Invalid: Can't use web scraping in offline mode
./tempest-homekit-go --token "your-token" --udp-stream --disable-internet --use-web-status
# ERROR: --use-web-status cannot be used with --disable-internet (requires internet access)

# Invalid: Can't preload history in offline mode
./tempest-homekit-go --token "your-token" --udp-stream --disable-internet --read-history
# ERROR: --read-history cannot be used with --disable-internet (requires WeatherFlow API access)
```

### HomeKit Only Mode Examples
```bash
# Valid: HomeKit only, no web console
./tempest-homekit-go --token "your-token" --disable-webconsole

# Valid: Offline HomeKit with UDP stream, no web console
./tempest-homekit-go --token "your-token" --udp-stream --disable-internet --disable-webconsole

# Invalid: Can't disable both HomeKit and web console
./tempest-homekit-go --token "your-token" --disable-homekit --disable-webconsole
# ERROR: cannot disable both HomeKit (--disable-homekit) and web console (--disable-webconsole) - at least one service must be enabled
```

### Validation Examples
```bash
# Invalid elevation (too high) - shows helpful error message
./tempest-homekit-go --token "your-token" --elevation 10000
# Error: elevation must be between -430m and 8848m (Earth's surface range)

# Invalid sensor name - shows available options
./tempest-homekit-go --token "your-token" --sensors "invalid-sensor"
# Error: invalid sensor 'invalid-sensor'. Available: temp/temperature, lux/light, uv/uvi, humidity, wind, rain, pressure, lightning

# Missing required token - shows usage
./tempest-homekit-go --sensors "temp"
# Error: WeatherFlow API token is required. Use --token flag or TEMPEST_TOKEN environment variable
```

### Web Console Only (No HomeKit)
```bash
# Run web dashboard only without HomeKit services
./tempest-homekit-go \
  --token "your-api-token" \
  --disable-homekit \
  --web-port 8080 \
  --loglevel info
```

### TempestWX Device Status Scraping

Enable detailed device status monitoring with the `--use-web-status` flag:

```bash
# Basic usage with device status scraping
./tempest-homekit-go --token "your-token" --use-web-status

# With full configuration
./tempest-homekit-go --token "your-token" --use-web-status --loglevel debug
```

**Requirements:**
- Google Chrome or Chromium installed
- Internet access to https://tempestwx.com

**What it provides:**
- **Battery Status**: Real battery voltage (e.g., "2.69V") and condition (Good/Fair/Poor)
- **Device Uptime**: How long your Tempest device has been running
- **Hub Uptime**: How long your Tempest hub has been running  
- **Signal Strength**: Wi-Fi signal strength for hub, device signal strength
- **Firmware Versions**: Current firmware for both hub and device
- **Serial Numbers**: Hardware serial numbers for troubleshooting
- **Last Activity**: Timestamps of last status updates and observations

**Status API Response with Web Scraping:**
```json
{
  "stationStatus": {
    "batteryVoltage": "2.69V",
    "batteryStatus": "Good",
    "deviceUptime": "128d 6h 19m 29s",
    "hubUptime": "63d 15h 55m 1s",
    "hubWiFiSignal": "Strong (-42)",
    "deviceSignal": "Good (-65)",
    "hubSerialNumber": "HB-00168934",
    "deviceSerialNumber": "ST-00163375",
    "hubFirmware": "v329",
    "deviceFirmware": "v179",
    "dataSource": "web-scraped",
    "lastScraped": "2025-09-18T03:15:30Z",
    "scrapingEnabled": true
  }
}
```

**Without `--use-web-status` (default):**
Basic status with API-only data:
```json
{
  "stationStatus": {
    "batteryVoltage": "--",
    "dataSource": "api",
    "scrapingEnabled": false
  }
}
```

**How it works:**
1. **Headless Browser**: Launches Chrome to load the TempestWX status page
2. **JavaScript Execution**: Waits for JavaScript to populate the device status data
3. **Data Extraction**: Parses the loaded content to extract device information
4. **15-Minute Updates**: Automatically refreshes data every 15 minutes
5. **Graceful Fallbacks**: Falls back to HTTP scraping, then API-only if issues occur

### UDP Stream (Offline Mode)

The UDP stream feature enables local monitoring of your Tempest station without requiring internet connectivity. This is particularly useful during internet outages when you still need weather data for HomeKit automations.

**Use Case: Internet Outage Resilience**

When your internet connection goes down, the WeatherFlow API becomes unavailable. With UDP streaming enabled, your Tempest hub broadcasts weather observations on your local network, allowing continuous monitoring without cloud access.

#### Operation Modes

**1. Hybrid Mode (UDP + Internet)**
```bash
# UDP for real-time observations, API for forecast/history (recommended for most users)
./tempest-homekit-go --udp-stream --token "your-token"

# Add historical data preloading from API
./tempest-homekit-go --udp-stream --read-history --token "your-token"  # preloads up to HISTORY_POINTS observations

# Include web scraping for detailed device status
./tempest-homekit-go --udp-stream --use-web-status --token "your-token"
```
- ✅ Real-time UDP observations every 60 seconds
- ✅ Forecast data from WeatherFlow API
- ✅ Historical data preloading available
- ✅ Optional web scraping for battery/RSSI status

**2. Full Offline Mode (UDP Only)**
```bash
# Complete offline operation - no internet access at all
./tempest-homekit-go --udp-stream --disable-internet

# Offline mode with custom sensors and debug logging
./tempest-homekit-go --udp-stream --disable-internet --sensors "temp,humidity,lux,wind" --loglevel debug
```
- ✅ Real-time UDP observations only
- ❌ No forecast data
- ❌ No historical data preloading (`--read-history` not allowed)
- ❌ No web scraping (`--use-web-status` not allowed)
- ✅ Zero internet dependency - works during complete outages
- ℹ️ API token (`--token`) still required but not used for network calls

**3. HomeKit Only Mode (No Web Console)**
```bash
# HomeKit accessories only, disable web dashboard
./tempest-homekit-go --token "your-token" --disable-webconsole

# HomeKit only with offline mode
./tempest-homekit-go --token "your-token" --udp-stream --disable-internet --disable-webconsole
```
- ✅ HomeKit accessories enabled
- ❌ Web dashboard disabled (port not opened)
- ✅ Reduced resource usage
- ✅ Minimal attack surface

#### Configuration Validation

The `--disable-internet` flag enforces strict validation to prevent conflicting configurations:

| Configuration | Result | Reason |
|--------------|--------|--------|
| `--disable-internet` alone | ❌ **ERROR** | Requires local data source |
| `--disable-internet --udp-stream` | ✅ **Valid** | Pure offline mode with real station |
| `--disable-internet --use-generated-weather` | ✅ **Valid** | Pure offline mode with simulated data |
| `--disable-internet --use-web-status` | ❌ **ERROR** | Web scraping requires internet access |
| `--disable-internet --read-history` | ❌ **ERROR** | Historical data requires WeatherFlow API |
| `--disable-internet --udp-stream --read-history` | ❌ **ERROR** | History requires API calls |
| `--disable-homekit --disable-webconsole` | ❌ **ERROR** | At least one service must be enabled |

**Error Messages:**
```
ERROR: --disable-internet mode requires --udp-stream or --use-generated-weather (need a local data source)
ERROR: --use-web-status cannot be used with --disable-internet (requires internet access)
ERROR: --read-history cannot be used with --disable-internet (requires WeatherFlow API access)
```

#### Network Requirements

**Hardware:**
- Tempest hub and monitoring device must be on the same local network
- UDP port 50222 must be accessible (no firewall blocking)
- Hub broadcasts observations every 60 seconds

**What it provides:**
- **Real-time Observations**: Temperature, humidity, wind, pressure, UV, rain, lightning data
- **Device Status**: Battery voltage, RSSI, sensor status
- **Hub Status**: Firmware version, uptime, reset flags
- **No Internet Required**: Complete offline operation with `--disable-internet` flag

**Network Topology:**
- Both devices on same subnet (hub broadcasts to 255.255.255.255)
- No special router configuration needed for standard LAN setups

**Status API Response with UDP Stream:**
```json
{
  "udpStatus": {
    "enabled": true,
    "receivingData": true,
    "packetCount": 147,
    "stationIP": "192.168.1.50",
    "serialNumber": "ST-00163375",
    "lastPacketTime": "2025-01-20T15:30:45Z"
  }
}
```

**Limitations:**
- No forecast data in full offline mode (`--disable-internet`)
- Historical data limited to observations received since startup
- Requires hub on local network (won't work remotely)

## Testing

The application includes comprehensive testing flags for validating configurations and troubleshooting issues before deployment.

### Testing Flags

#### API and Data Source Testing

**Test WeatherFlow API Endpoints** (`--test-api`)
```bash
./tempest-homekit-go --test-api
```
Tests all WeatherFlow API endpoints:
- Station discovery and details
- Current observations
- Historical data retrieval
- Performance metrics

**Test UDP Broadcast Listener** (`--test-udp [seconds]`)
```bash
# Listen for 120 seconds (default)
./tempest-homekit-go --test-udp

# Listen for custom duration
./tempest-homekit-go --test-udp 30
```
Tests UDP broadcast reception:
- Listens on port 50222 for Tempest station broadcasts
- Displays real-time packet statistics every 5 seconds
- Shows latest observation data when complete
- Helps diagnose network/firewall issues

#### Notification Delivery Testing

**Test Email Delivery** (`--test-email <email>`)
```bash
./tempest-homekit-go --test-email user@example.com --alarms @alarms.json
```
Tests email notification delivery:
- Auto-detects provider (Microsoft 365 OAuth2 or SMTP)
- Validates credentials from environment variables
- Sends test email with weather data
- Uses real delivery path (factory pattern)

**Test SMS Delivery** (`--test-sms <phone>`)
```bash
./tempest-homekit-go --test-sms +15555551234 --alarms @alarms.json
```
Tests SMS notification delivery:
- Auto-detects provider (Twilio or AWS SNS)
- Validates credentials from environment variables
- Sends test SMS with weather data
- Uses real delivery path (factory pattern)

**Test Console Notifications** (`--test-console`)
```bash
./tempest-homekit-go --test-console --alarms @alarms.json
```
Tests console/stdout notification delivery.

**Test Syslog Notifications** (`--test-syslog`)
```bash
./tempest-homekit-go --test-syslog --alarms @alarms.json
```
Tests syslog notification delivery (local or remote).

**Test OSLog Notifications** (`--test-oslog`) - macOS only
```bash
./tempest-homekit-go --test-oslog --alarms @alarms.json
```
Tests macOS unified logging system integration.

**Test Event Log Notifications** (`--test-eventlog`) - Windows only
```bash
./tempest-homekit-go --test-eventlog --alarms @alarms.json
```
Tests Windows Event Log integration.

#### Service Testing

**Test HomeKit Bridge** (`--test-homekit`)
```bash
./tempest-homekit-go --test-homekit
```
Tests HomeKit bridge configuration:
- Displays sensor configuration
- Shows pairing instructions
- Validates PIN and station settings
- Dry-run mode (doesn't start actual bridge)

**Test Web Status Scraping** (`--test-web-status`)
```bash
./tempest-homekit-go --test-web-status
```
Tests web status scraping capability:
- Validates Chrome/Chromium availability
- Provides setup guidance
- Placeholder for future headless browser implementation

**Test Specific Alarm** (`--test-alarm <name>`)
```bash
./tempest-homekit-go --test-alarm "high-temperature" --alarms @alarms.json --station "Test"
```
Tests a specific alarm trigger:
- Validates alarm exists and is enabled
- Sends test observation to trigger alarm
- Tests entire notification delivery pipeline
- Shows notification results for all channels

### Testing Best Practices

1. **Test in order**: Start with `--test-api` to validate connectivity
2. **Test notifications**: Use test flags before deploying alarm configurations
3. **Use factory pattern**: All notification tests use the real delivery path
4. **Check credentials**: Test flags validate environment variables are correct
5. **Validate network**: Use `--test-udp` to diagnose UDP broadcast issues

## HomeKit Setup

1. Start the application with your WeatherFlow API token
2. On your iOS device, open the Home app
3. Tap the "+" icon to add an accessory
4. Select "Don't have a code or can't scan?"
5. Choose the "Tempest Bridge"
6. Enter the PIN (default: 00102003)

The following sensors will appear as separate HomeKit accessories:
- **Temperature Sensor**: Air temperature in Celsius (uses standard HomeKit temperature characteristic)
- **Humidity Sensor**: Relative humidity as percentage (uses standard HomeKit humidity characteristic)  
- **Light Sensor**: Ambient light level in lux (uses built-in HomeKit Light Sensor service)
- **Pressure Sensor**: Atmospheric pressure in mb (uses Light Sensor service for compliance - ignore "lux" unit label)
- **UV Index Sensor**: UV index value (uses Light Sensor service for compliance - ignore "lux" unit label)
- **Custom Wind Speed Sensor**: Wind speed in miles per hour (custom service prevents unit conversion)
- **Custom Wind Gust Sensor**: Wind gust speed in miles per hour (custom service)
- **Custom Wind Direction Sensor**: Wind direction in cardinal format with degrees (custom service)
- **Custom Rain Sensor**: Rain accumulation in inches (custom service)
- **Custom Lightning Count Sensor**: Lightning strike count (custom service)
- **Custom Lightning Distance Sensor**: Lightning strike distance (custom service)
- **Custom Precipitation Type Sensor**: Precipitation type indicator (custom service)

**Important**: The **Pressure** and **UV Index** sensors use HomeKit's standard Light Sensor service for maximum compatibility. In the Home app, they will appear as "Light Sensor" with "lux" units, but display the correct pressure (mb) and UV index values. Please ignore the "lux" unit label for these sensors - this is a HomeKit platform limitation, not an application issue.

⚠️ **HomeKit Compliance Warning**: As of Home.app v10.0, all sensors labeled as "(custom service)" above will return an "Out of Compliance" error when attempting to add the accessory to the Home app. Only the standard HomeKit services (Temperature, Humidity, Light, Pressure, UV Index) will successfully pair. This is due to Apple's stricter compliance enforcement in recent Home app versions.

## Web Dashboard

Access the modern web dashboard at `http://localhost:8080` (or your configured port).

### Dashboard Features
- **External JavaScript Architecture**: Clean separation with all ~800+ lines of JavaScript moved to external `script.js` file
- **Real-time Updates**: Weather data refreshes every 10 seconds with comprehensive error handling
- **Pressure Analysis System**: Advanced atmospheric pressure monitoring with:
  - **Trend Analysis**: Rising, Falling, or Stable pressure trends
  - **Weather Forecasting**: Predictions based on pressure patterns (Fair, Cloudy, Stormy)
  - **Interactive Info Icon**: Click the ℹ️ icon for detailed pressure calculation explanations
- **Interactive Unit Conversion**: Click any sensor card to toggle units:
  - 🌡️ **Temperature**: Celsius (°C) ↔ Fahrenheit (°F)
  - 🌬️ **Wind Speed**: Miles per hour (mph) ↔ Kilometers per hour (kph)
  - 🌧️ **Rain**: Inches (in) ↔ Millimeters (mm)
  - 🌀 **Pressure**: Millibars (mb) ↔ Inches of Mercury (inHg)
- **UV Index Monitor**: 🌞 Complete UV exposure assessment with NCBI reference categories:
  - **Minimal (0-2)**: Low risk exposure with EPA green color coding
  - **Low (3-4)**: Moderate risk with yellow coding  
  - **Moderate (5-6)**: High risk with orange coding
  - **High (7-9)**: Very high risk with red coding
  - **Very High (10+)**: Extreme risk with violet coding
- **Enhanced Information System**: ℹ️ Detailed sensor tooltips with proper event propagation handling
- **Accessories Status**: Real-time HomeKit sensor status showing enabled/disabled state with priority sorting
- **Wind Direction Display**: Shows cardinal direction + degrees (e.g., "WSW (241°)")
- **Unit Persistence**: Preferences saved in browser localStorage
- **Modern Design**: Responsive interface with weather-themed styling and cache-busting script loading
- **All Sensors**: Complete weather data display with comprehensive DOM debugging
- **HomeKit Status**: Bridge status, accessory count, and pairing PIN
- **Connection Status**: Real-time Tempest station connection status
- **Mobile Friendly**: Works perfectly on all devices with enhanced event listener management

### API Endpoints
- `GET /`: Main dashboard HTML with external JavaScript
- `GET /pkg/web/static/script.js`: External JavaScript file with cache-busting timestamps
- `GET /api/weather`: JSON weather data with pressure analysis
- `GET /api/status`: Service and HomeKit status with optional TempestWX device status

## Architecture

```
tempest-homekit-go/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── scripts/
│   ├── build.sh              # Platform-specific build script
│   ├── build-cross-platform.sh # Cross-platform build script
│   ├── install-service.sh    # Service installation script
│   ├── remove-service.sh     # Service removal script
│   └── README.md             # Scripts documentation
├── pkg/
│   ├── config/               # Configuration management
│   │   └── config.go
│   ├── weather/              # WeatherFlow API client
│   │   ├── client.go         # API client and TempestWX scraping
│   │   └── status_manager.go # Periodic status scraping manager
│   ├── homekit/              # HomeKit accessory setup
│   │   ├── modern_setup.go   # Modern HAP library implementation
│   │   └── custom_characteristics.go # Custom weather characteristics
│   ├── web/                  # Web dashboard server
│   │   ├── server.go         # HTTP server with static file serving
│   │   └── static/           # Static web assets
│   │       ├── script.js     # External JavaScript (~800+ lines)
│   │       ├── styles.css    # CSS styling
│   │       └── date-fns.min.js # Date manipulation library
│   └── service/              # Main service orchestration
│       └── service.go
└── README.md
```

## API Integration

### WeatherFlow Tempest API
- **Stations Endpoint**: `GET /swd/rest/stations?token={token}`
- **Observations Endpoint**: `GET /swd/rest/observations/station/{station_id}?token={token}`

### Supported Weather Metrics
- ✅ **Air Temperature**: In Fahrenheit/Celsius
- ✅ **Relative Humidity**: As percentage
- ✅ **Wind Speed**: Average wind speed in mph/kph
- ✅ **Wind Direction**: Degrees with cardinal conversion
- ✅ **Rain Accumulation**: Total precipitation in inches/mm
- ✅ **Air Pressure**: Atmospheric pressure in mb/inHg
- ✅ **UV Index**: UV exposure level (0-15)
- ✅ **Ambient Light**: Illuminance in lux

## Logging

### Log Levels
- **error**: Only errors and critical messages
- **info**: Basic operational messages + sensor data summary
- **debug**: Detailed sensor data + complete API JSON responses

### Example Log Output (Info Level)
```
2025-09-21 10:30:00 Starting service with config: WebPort=8080, LogLevel=info
2025-09-21 10:30:00 Starting Tempest HomeKit service...
2025-09-21 10:30:00 Found station: Chino Hills (ID: 178915)
2025-09-21 10:30:00 INFO: HomeKit server started successfully with PIN: 00102003
2025-09-21 10:30:00 INFO: Starting web dashboard on port 8080
2025-09-21 10:30:00 Starting web server on port 8080
2025-09-21 10:30:00 INFO: Successfully read weather data from Tempest API - Station: Chino Hills
2025-09-21 10:30:00 INFO: Sensor data - Temp: 22.7°C, Humidity: 77%, Wind: 0.3 mph (238°), Rain: 0.000 in, Light: 1 lux
```

### Example Log Output (Debug Level)
```
2025-09-21 10:30:00 service.go:25: Starting Tempest HomeKit service...
2025-09-21 10:30:00 service.go:29: DEBUG: Fetching stations from WeatherFlow API
2025-09-21 10:30:00 modern_setup.go:39: DEBUG: Creating new weather system with hap library
2025-09-21 10:30:00 modern_setup.go:89: DEBUG: Created temperature sensor accessory
2025-09-21 10:30:00 modern_setup.go:169: DEBUG: Created UV Index sensor accessory using light sensor service with UV range
2025-09-21 10:30:00 service.go:284: DEBUG: HomeKit - UV Index: 0
2025-09-21 10:30:00 service.go:304: DEBUG: Updating UV Index: 0.000
```

### Example Log Output (Error Level - Default)
```
2025-09-21 10:30:00 Starting service with config: WebPort=8080, LogLevel=error
2025-09-21 10:30:00 Starting Tempest HomeKit service...
2025-09-21 10:30:00 Found station: Chino Hills (ID: 178915)
2025-09-21 10:30:00 Starting web server on port 8080
```

## Service Management

### Linux (systemd)
```bash
# Install
sudo ./scripts/install-service.sh --token "your-token"

# Check status
sudo systemctl status tempest-homekit-go

# View logs
sudo journalctl -u tempest-homekit-go -f

# Remove
sudo ./scripts/remove-service.sh
```

### macOS (launchd)
```bash
# Install
sudo ./scripts/install-service.sh --token "your-token"

# Check status
sudo launchctl list | grep tempest

# View logs
log show --predicate 'process == "tempest-homekit-go"' --last 1h

# Remove
sudo ./scripts/remove-service.sh
```

### Windows (NSSM)
```bash
# Install
./scripts/install-service.sh --token "your-token"

# Check status
sc query tempest-homekit-go

# View logs (via Event Viewer)
# Remove
./scripts/remove-service.sh
```

## Configuration

### Environment Variables (.env File)

The application supports configuration via environment variables, which can be stored in a `.env` file for convenience. This is particularly useful for persistent configuration without specifying command-line flags every time.

#### Quick Setup

1. **Copy the example file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` with your values:**
   ```bash
   nano .env  # or use your preferred editor
   ```

3. **Run without flags:**
   ```bash
   ./tempest-homekit-go  # Will automatically load .env settings
   ```

#### Available Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TEMPEST_TOKEN` | *(see below)* | WeatherFlow API token |
| `TEMPEST_STATION_NAME` | *(required)* | Your station name from WeatherFlow |
| `HOMEKIT_PIN` | `00102003` | HomeKit pairing PIN |
| `SENSORS` | `temp,lux,humidity,uv` | Enabled sensors |
| `WEB_PORT` | `8080` | Web console port |
| `UNITS` | `imperial` | Unit system (imperial/metric/sae) |
| `UNITS_PRESSURE` | `inHg` | Pressure units (inHg/mb/hpa) |
| `HISTORY_POINTS` | `1000` | Data points to store (min 10) |
| `CHART_HISTORY_HOURS` | `24` | Hours to display in charts (0=all) |
| `UDP_STREAM` | `false` | Enable UDP mode (true/false) |
| `DISABLE_INTERNET` | `false` | Offline mode (true/false) |
| `LOG_LEVEL` | `error` | Logging level (error/warn/warning/info/debug) |
| `LOG_FILTER` | *(empty)* | Filter log messages |

**Note:** Command-line flags always override environment variables.

#### Example .env Configurations

**API Mode (Cloud Data):**
```bash
# Get your token from: https://tempestwx.com/settings/tokens
TEMPEST_TOKEN=your-actual-token-here
TEMPEST_STATION_NAME=My Station Name
SENSORS=temp,humidity,pressure,wind
LOG_LEVEL=info
```

⚠️ **Security Note**: Never commit your `.env` file with real credentials! See [SECURITY.md](SECURITY.md) for details.

**UDP Mode (Local Offline):**
```bash
UDP_STREAM=true
DISABLE_INTERNET=true
HISTORY_POINTS=500
CHART_HISTORY_HOURS=12
LOG_LEVEL=debug
```

**Minimal Memory:**
```bash
HISTORY_POINTS=100
CHART_HISTORY_HOURS=6
SENSORS=temp,humidity
```

### WeatherFlow API Token
1. Visit [tempestwx.com](https://tempestwx.com)
2. Go to Settings → Data Authorizations
3. Create a new personal access token
4. Use with `--token` flag or `TEMPEST_TOKEN` environment variable

### Station Discovery
The application automatically finds your station by name. Ensure your station name in WeatherFlow matches the `--station` parameter.

## Troubleshooting

### HomeKit Re-pairing (Database Reset)

When you make changes to HomeKit accessories (such as modifying sensor types, names, or configurations), you may need to reset the HomeKit database and re-pair the bridge with your Home app. This ensures the changes take effect properly.

#### Using the Built-in --cleardb Command (Recommended)

The easiest way to reset HomeKit pairing is using the built-in `--cleardb` command:

```bash
# Stop the current service if running
pkill -f tempest-homekit-go

# Clear the database and reset pairing
./tempest-homekit-go --cleardb

# Restart the service normally
./tempest-homekit-go --token "your-api-token"
```

#### Manual Database Reset

If you prefer to do it manually:

1. **Stop the Application**
   ```bash
   # If running as a service
   sudo systemctl stop tempest-homekit-go  # Linux
   sudo launchctl stop tempest-homekit-go  # macOS
   sc stop tempest-homekit-go              # Windows
   
   # Or kill the process directly
   pkill -f tempest-homekit-go
   ```

2. **Delete the HomeKit Database**
   ```bash
   # Navigate to the application directory
   cd /path/to/tempest-homekit-go
   
   # Remove the database directory (this contains all pairing information)
   rm -rf ./db/
   
   # Verify the directory is empty
   ls -la ./db/
   ```

3. **Restart the Application**
   ```bash
   # Start the application again
   ./tempest-homekit-go --token "your-api-token"
   
   # Or restart the service
   sudo systemctl start tempest-homekit-go  # Linux
   sudo launchctl start tempest-homekit-go  # macOS
   sc start tempest-homekit-go              # Windows
   ```

4. **Re-pair in Home App**
   - Open the Home app on your iOS device
   - The "Tempest HomeKit Bridge" should appear as a new, unpaired accessory
   - Tap the "+" icon to add an accessory
   - Select "Don't have a code or can't scan?"
   - Choose the "Tempest HomeKit Bridge"
   - Enter the PIN (default: `00102003`)

5. **Verify the Changes**
   - Check that all accessories appear correctly
   - Confirm sensor types and names are as expected
   - Test that sensors are no longer grouped incorrectly

#### Important Notes:
- **Data Loss**: This will remove all HomeKit pairing information and automation rules
- **Re-setup Required**: You'll need to re-add any scenes, automations, or accessory groupings
- **Safe Operation**: The weather data collection continues normally; only HomeKit pairing is affected
- **Backup First**: Consider noting any important automation rules before resetting

#### Alternative: Clear Specific Database Files
If you want to be more selective, you can remove specific database files instead of the entire directory:
```bash
# Remove only pairing information (keeps other HomeKit data)
rm -f ./db/pairings.json

# Remove accessory cache (forces rediscovery)
rm -f ./db/accessories.json
```

### Common Issues
- **"Station not found"**: Verify station name matches exactly (case-sensitive)
- **"API request failed"**: Check internet connection and API token validity
- **HomeKit pairing fails**: Ensure PIN is correct and no other devices are pairing
- **Web dashboard not loading**: Check if port 8080 is available
- **Sensors showing wrong values/types**: Reset HomeKit database and re-pair (see above)

### Debug Mode
Enable detailed logging for troubleshooting:
```bash
./tempest-homekit-go --loglevel debug --token "your-token"
```

Filter logs to show only specific messages (case-insensitive):
```bash
# Show only UDP-related messages
./tempest-homekit-go --loglevel debug --logfilter "udp" --udp-stream

# Show only forecast-related messages
./tempest-homekit-go --loglevel info --logfilter "forecast" --token "your-token"

# Show only observation parsing messages
./tempest-homekit-go --loglevel debug --logfilter "parsed" --udp-stream
```

### Service Issues
```bash
# Check service status
./scripts/install-service.sh --status

# Restart service
./scripts/remove-service.sh
./scripts/install-service.sh --token "your-token"
```

## Development

### GoDoc Server
Browse the complete Go documentation and API references locally:
```bash
# Start GoDoc server on port 6060 (opens browser automatically)
./scripts/start-godoc.sh

# Start on custom port without opening browser
./scripts/start-godoc.sh --port 8080 --no-browser

# View help
./scripts/start-godoc.sh --help
```

Then visit `http://localhost:6060` to browse:
- Package documentation for all modules (`pkg/config`, `pkg/weather`, etc.)
- Function and type definitions with examples
- Cross-referenced source code
- Standard library documentation

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run verbose tests
go test -v ./...

# Run specific package tests
go test ./pkg/config/...
go test ./pkg/weather/...
go test ./pkg/web/...
go test ./pkg/service/...
```

# Test Coverage Overview
- coverage report saved at `./coverage.out` (generated by the most recent coverage run)
- Current per-package snapshot from the latest test runs:
  - pkg/config: 79.9%
  - pkg/generator: ~86%
  - pkg/homekit: 84.5%
  - pkg/logger: 91.3%
  - pkg/service: 49.1%  (highest-leverage area to add tests)
  - pkg/udp: 51.3%
  - pkg/weather: 60.8%
  - pkg/web: 65.0%

**Current Aggregate Coverage**: 60.3% (see `./coverage.out`)

Note: Per-package numbers above were collected from individual `go test` outputs during iterative runs. The aggregate coverage uses a single `coverage.out` collected with `go test -coverprofile=coverage.out ./...` and is the authoritative project-wide percentage. The current project goal is to raise overall coverage to >= 70% by adding targeted tests (priority: `pkg/service`, then `pkg/weather`).

### Package Coverage Breakdown

| Package | Coverage |
|---------|----------|
| pkg/alarm | 61.9% |
| pkg/alarm/editor | 28.8% |
| pkg/config | 80.2% |
| pkg/generator | 87.8% |
| pkg/homekit | 84.5% |
| pkg/logger | 91.3% |
| pkg/service | 47.7% |
| pkg/udp | 51.3% |
| pkg/weather | 60.8% |
| pkg/web | 65.0% |
| **Overall** | **60.3%** |

### Testing Architecture
The project includes unit tests and integration-style tests that use small, isolated test doubles and local httptest servers where appropriate. Test patterns used across the repo:
- Table-driven tests for parsing and validation logic
- Fake implementations for interface dependencies (e.g., `weather.DataSource`, UDP listeners)
- `httptest` servers for HTTP client/server interactions
- Package-level injection points (for example, a `DataSourceFactory` variable in `pkg/service`) to make orchestration testable without starting long-lived goroutines

When running coverage locally, use `go test -coverprofile=coverage.out ./...` and then `go tool cover -func=coverage.out` to inspect the aggregate and per-file percentages.

### Building for Development
```bash
go build -o tempest-homekit-go
```

### Code Quality
- Comprehensive error handling and recovery
- Unit test coverage for all packages with table-driven tests
- Modular design for maintainability
- Follows Go best practices and conventions
- HTTP testing with `httptest.ResponseRecorder`
- Mock data creation for realistic test scenarios

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- **WeatherFlow** for the Tempest weather station and API
- **Apple** for the HomeKit platform
- **hc library** for HomeKit Go implementation
- **Community** for feedback and contributions

## References

This project was developed using various technologies, libraries, and tools. Below is a comprehensive list of key components and resources that contributed to the development:

### Core Technologies
- **Go Programming Language** (v1.24.2+) - Primary programming language
- **HomeKit Accessory Protocol** - Apple's smart home communication protocol
- **WeatherFlow Tempest API** - Weather data source and API integration

### Go Libraries and Dependencies
- **`github.com/brutella/hap`** - HomeKit Accessory Protocol implementation for Go
- **Standard Library Packages**:
  - `net/http` - Web server implementation
  - `encoding/json` - JSON data handling
  - `sync` - Concurrent programming primitives
  - `time` - Time and date operations
  - `log` - Logging functionality
  - `os` - Operating system interface
  - `flag` - Command-line flag parsing

### Web Technologies (Embedded Dashboard)
- **HTML5** - Dashboard structure and markup
- **CSS3** - Responsive styling and animations
- **JavaScript (ES6+)** - Interactive functionality and real-time updates
- **Chart.js** (v4.4.0) - Interactive charts and data visualization
- **date-fns** (v2.30.0) - Date and time manipulation in JavaScript
- **Chart.js Date-Fns Adapter** (v2.0.1) - Time-based chart integration

### Development Tools and AI Assistance
- **GitHub Copilot** - AI-powered code suggestions and development assistance
- **Visual Studio Code** - Primary development environment
- **Go Modules** - Dependency management
- **Git** - Version control system

### Platform-Specific Tools
- **systemd** (Linux) - Service management
- **launchd** (macOS) - Service management
- **NSSM** (Windows) - Non-Sucking Service Manager for Windows services

### Build and Deployment
- **Cross-compilation** - Go's built-in cross-platform compilation
- **Shell scripting** - Bash scripts for automated builds and deployment
- **Platform detection** - Runtime OS and architecture detection

### External Resources and Documentation
- **WeatherFlow API Documentation** - Weather data integration reference
- **Apple HomeKit Developer Documentation** - HomeKit protocol implementation guide
- **Go Documentation** - Standard library and language reference
- **MDN Web Docs** - JavaScript, HTML, and CSS reference

### Development Practices
- **Test-Driven Development** - Unit testing approach
- **Modular Architecture** - Clean code organization
- **Error Handling** - Comprehensive error management
- **Logging** - Multi-level logging system
- **Configuration Management** - Flexible configuration via flags and environment variables

## Additional Documentation

### Alarm System Documentation
- **[ALARM_LOGGING.md](ALARM_LOGGING.md)** - ⭐ Alarm logging behavior (always visible regardless of log level)
- **[ALARM_COOLDOWN_STATUS.md](ALARM_COOLDOWN_STATUS.md)** - Real-time cooldown status display in web console
- **[OSLOG_NOTIFIER.md](OSLOG_NOTIFIER.md)** - macOS unified logging integration for alarms
- **[CHANGE_DETECTION_OPERATORS.md](CHANGE_DETECTION_OPERATORS.md)** - Complete technical reference for change detection operators (*field, >field, <field)
- **[CHANGE_DETECTION_QUICKREF.md](CHANGE_DETECTION_QUICKREF.md)** - Quick reference guide with examples
- **[CHANGE_DETECTION_VISUAL.md](CHANGE_DETECTION_VISUAL.md)** - Visual diagrams and state transition timelines
- **[CHANGE_DETECTION_SUMMARY.md](CHANGE_DETECTION_SUMMARY.md)** - Implementation summary and architecture
- **[ALARM_EDITOR_MESSAGES.md](ALARM_EDITOR_MESSAGES.md)** - Message configuration with variable templates
- **[ALARM_EDITOR_CHANNEL_FIX.md](ALARM_EDITOR_CHANNEL_FIX.md)** - Documentation of alarm channel save fix
- **[WEB_ALARM_STATUS_CARD.md](WEB_ALARM_STATUS_CARD.md)** - Web console alarm status card implementation
- **[examples/alarms-with-change-detection.json](examples/alarms-with-change-detection.json)** - Ready-to-use alarm configurations

### Package Documentation
Each package includes detailed README files:
- **[pkg/alarm/README.md](pkg/alarm/README.md)** - Alarm package documentation
- **[pkg/alarm/editor/README.md](pkg/alarm/editor/README.md)** - Alarm editor documentation
- **[pkg/config/README.md](pkg/config/README.md)** - Configuration package documentation
- **[pkg/weather/README.md](pkg/weather/README.md)** - Weather data source documentation
- **[pkg/web/README.md](pkg/web/README.md)** - Web dashboard documentation
- **[pkg/service/README.md](pkg/service/README.md)** - Service orchestration documentation

---

**Status**: ✅ **COMPLETE** - All planned features implemented and tested
- ✅ Weather monitoring with 11 HomeKit sensors (Temperature + 10 custom weather sensors)
- ✅ Complete HomeKit integration with compliance optimization
- ✅ Modern web dashboard with real-time updates and interactive features
- ✅ UV Index monitoring with NCBI reference data and EPA color coding
- ✅ Information tooltips system with standardized positioning
- ✅ HomeKit accessories status monitoring with enabled/disabled indicators
- ✅ Interactive unit conversions with localStorage persistence
- ✅ Cross-platform build and deployment with automated service management
- ✅ Professional styling and enhanced user experience
- ✅ Comprehensive logging and error handling
- ✅ Database management with --cleardb command
- ✅ Production-ready with graceful error recovery
- ✅ Weather monitoring with 6 metrics (Temperature, Humidity, Wind Speed, Wind Direction, Rain, Light)
- ✅ Complete HomeKit integration with individual sensors
- ✅ Modern web dashboard with real-time updates
- ✅ Interactive unit conversions with persistence
- ✅ Cross-platform build and deployment
- ✅ Service management for all platforms
- ✅ Comprehensive logging and error handling
- ✅ Database management with --cleardb command
- ✅ Production-ready with graceful error handling