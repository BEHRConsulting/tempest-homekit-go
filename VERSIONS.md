# Version History

This file tracks released versions and the notable changes implemented in each release. Its purpose is to keep the README concise and centralize the version-by-version changelog.

## v1.5.0 (current)
- Public release preparation and documentation updates
- UDP stream support for local Tempest hub (offline mode)
- Web dashboard and HomeKit integration improvements
- Test additions and coverage improvements across multiple packages

## v1.4.1
- Unified data pipeline to avoid dual data paths
- Fixed rain accumulation bug
- Improved chart rendering and deterministic popout charts

## v1.3.0
- Comprehensive command-line validation and helpful error messages
- Sensor name aliases (temp/temperature, lux/light, uv/uvi)
- Elevation validation (Earth-realistic range)
- Improved logging compliance and prefixing
- UV value rounding and improved sensor configuration
- Initial large-scale unit test additions

## Recent Enhancements (collected)
- Tooltip positioning and UI/UX tweaks
- JavaScript refactor: moved inline JS into `pkg/web/static/script.js` and introduced cache-busting
- Pressure analysis system and interactive info icons
- Headless popout diagnostic tests (chromedp-based) to reduce CI flakiness
- Enhanced debug logging and configurable log filters
- Vibe Programming methodology notes and documentation updates


*Notes:* This is intended as a concise developer-facing version history. For user-facing change logs, consider a dedicated `CHANGELOG.md` following "Keep a Changelog" conventions.
