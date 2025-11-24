# Release Notes — v1.11.0 (2025-11-24)

Summary
-------
v1.11.0 is a usability and alarm-enhancement release that introduces a new interactive status console (terminal UI), a more powerful rules engine for alarms, flexible alarm delivery to local CSV/JSON files, and per-alarm scheduling to limit false positives.

Highlights
----------
- Status Console: New curses-based TUI for real-time monitoring of logs, sensors, station status, HomeKit state, and system info. Supports manual refresh, auto-refresh, theme cycling, and responsive layout.
- Advanced Rules Engine: Boolean logic (`AND`/`OR`), time-window conditions, and notification throttling/rate-limiting to compose complex alarm rules.
- Additional Alarm Delivery Methods: Local CSV and JSON file logging of alarm events with configurable retention and fallback handling.
- Per-Alarm Scheduling System: Per-alarm activation windows (daily/hourly ranges), sunrise/sunset triggers, and day-of-week restrictions.

Notes for operators
-------------------
- See `CHANGELOG.md` and `VERSIONS.md` for full developer-facing details.
- If you use the status console in scripted environments, note that the console is interactive — use `--status-timeout` to auto-exit for automation.
- CSV/JSON alarm logs live alongside other configured notification channels; configure retention and file paths in your environment `.env`.

Upgrade & Release
-----------------
- Tag: `v1.11.0`
- To upgrade: pull the latest `main` and restart the service. Example:

```bash
git fetch origin main && git checkout main && git pull
go build ./...
# restart your service manager (systemd, launchd, etc.)
```

Acknowledgements
----------------
Thanks to contributors and test authors for improving stability and documentation.
