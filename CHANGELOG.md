# Changelog

All notable changes to this project will be documented in this file.

The format is based on "Keep a Changelog" and this project adheres to Semantic Versioning.

## [Unreleased]
- Ongoing test and coverage improvements

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
