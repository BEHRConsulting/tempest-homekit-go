# Changelog

All notable changes to this project will be documented in this file.

The format is based on "Keep a Changelog" and this project adheres to Semantic Versioning.

## [Unreleased]
- Ongoing test and coverage improvements

## [1.8.0] - 2025-10-15
### Added
- **AWS SNS SMS Notifications**: Complete implementation of AWS SNS for SMS alarm delivery
  - Direct SMS to phone numbers or SNS topic broadcasting
  - AWS SDK v2 integration (aws-sdk-go-v2 v1.39.2)
  - Support for cross-account SNS topics with resource-based policies
  - Environment-first configuration (credentials in `.env`, rules in JSON)
  - Interactive setup script: `scripts/setup-aws-sns.sh` for production configuration
  - Automated AWS CLI integration for topic creation, SMS testing, and `.env` updates
  - Two-tier credential system: admin credentials for setup, runtime credentials for sending
  - Complete documentation in `.env`, `.env.example`, and `AWS_SNS_IMPLEMENTATION.md`
  - Unit test suite: `pkg/alarm/notifiers_sms_test.go`
- **Alarm Name Editing**: Users can now edit alarm names in the alarm editor
  - Removed read-only restriction on alarm name field
  - Server-side validation to prevent duplicate names
  - Automatic tracking of original name for updates
- Enhanced `.gitignore` protection for all `.env` file variants and backups

### Changed
- Version bumped to 1.8.0
- Updated `REQUIREMENTS.md` with AWS SNS environment variables
- Updated `README.md` with AWS SNS quick start and setup instructions
- Updated `WORK_LOG.md` with AWS SNS implementation details

### Fixed
- Cross-account SNS topic access with resource-based policy configuration
- Alarm editor now properly handles alarm name changes without conflicts

## [1.7.0] - 2025-10-10
### Added
- **Microsoft 365 OAuth2 Email Integration**: Complete implementation for alarm notifications
  - OAuth2 authentication using Azure AD app credentials
  - Microsoft Graph API integration for email delivery
  - Client credentials flow for server-to-server communication
  - Support for HTML email rendering with proper content type handling
- **Alarm System Enhancements**: Enhanced notification capabilities
  - Previous sensor values tracking for all sensors (not just change detection fields)
  - Row highlighting in sensor info tables for changed values
  - Smart change detection thresholds (0.1°C for temp, 1% humidity, etc.)
  - Template variables: `{{last_temperature}}`, `{{last_humidity}}`, etc.
  - Composite variables: `{{app_info}}`, `{{alarm_info}}`, `{{sensor_info}}`
- **Comprehensive Testing Infrastructure**: 11 test flags for validation and troubleshooting
  - `--test-email <email>`: Email delivery testing with provider auto-detection
  - `--test-sms <phone>`: SMS delivery testing with provider auto-detection
  - `--test-console`: Console notification testing
  - `--test-syslog`: Syslog notification testing
  - `--test-oslog`: macOS unified logging testing
  - `--test-eventlog`: Windows Event Log testing
  - `--test-udp [seconds]`: UDP broadcast listener testing (default: 120s)
  - `--test-homekit`: HomeKit bridge configuration testing
  - `--test-web-status`: Web status scraping testing
  - `--test-alarm <name>`: Specific alarm trigger testing
  - All tests use factory pattern for real delivery path validation
- **Test Suite**: 98+ unit tests covering all test flags
  - `pkg/config/config_test_flags_test.go`: Flag parsing and configuration tests
  - `pkg/config/config_test_validation_test.go`: Parameter validation tests
  - `main_test.go`: Integration and handler validation tests
- Dependencies: Azure SDK for Go, Microsoft Graph SDK for Go

### Changed
- `formatSensorInfoWithAlarm()`: Now displays both current and previous sensor values
- `ProcessObservation()`: Stores ALL sensor values after evaluation for notification display
- Email notifier: Fixed HTML email rendering (was defaulting to TEXT_BODYTYPE)

### Fixed
- HTML emails now render properly in Microsoft 365 (set HTML_BODYTYPE when Html flag is true)
- Previous values now tracked for all sensors, not just change detection operators
- Changed sensor rows properly highlighted in yellow background

## [1.6.0] - 2025-10-08
### Added
- **Alarm System**: Rule-based weather alerting with multiple notification channels
  - Console logging, Syslog, Email (SMTP/Microsoft 365), SMS (Twilio, AWS SNS)
  - Configurable alarm conditions with operators (>, <, >=, <=, ==, !=, &&, ||)
  - Template-based messages with runtime value interpolation
  - Cross-platform file watching for live configuration reloads
  - Alarm cooldown periods to prevent notification storms
  - CLI flags: `--alarms @filename.json`, `--alarms-edit @filename.json`, `--alarms-edit-port`
- **Alarm Editor**: Interactive web UI for managing alarm configurations
  - Modern, responsive interface with search and filter capabilities
  - Create, edit, delete alarms with live validation
  - Tag-based organization and filtering
  - Visual status indicators for enabled/disabled alarms
  - Auto-save to JSON configuration file
  - Standalone mode accessible at `http://localhost:8081`
- Example alarm configuration: `alarms.example.json`
- Environment variables for SMTP, SMS providers (Twilio, AWS SNS), and Syslog

### Changed
- Version bumped to 1.6.0
- Updated `.env.example` with alarm provider credentials

## [1.5.0] - YYYY-MM-DD
### Added
- UDP stream support for local Tempest hub (offline mode)
- Web dashboard and HomeKit integration improvements
- Test additions and coverage improvements across multiple packages

### Changed
- Public release preparation and documentation updates

## [1.4.1] - YYYY-MM-DD
### Fixed
- Fixed rain accumulation bug

### Changed
- Unified data pipeline to avoid dual data paths
- Improved chart rendering and deterministic popout charts

## [1.3.0] - YYYY-MM-DD
### Added
- Comprehensive command-line validation and helpful error messages
- Sensor name aliases (temp/temperature, lux/light, uv/uvi)
- Elevation validation (Earth-realistic range)
- Initial large-scale unit test additions

### Changed
- Improved logging compliance and prefixing
- UV value rounding and improved sensor configuration



[Unreleased]: #
[1.5.0]: #
[1.4.1]: #
[1.3.0]: #
