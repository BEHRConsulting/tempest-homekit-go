PROMPT HISTORY — docs/development (annotated with inferred commit dates)
=====================================================================

This document paraphrases the development prompts taken from `docs/development/prompt.txt` and
`docs/development/prompt-refine.txt`. Where repository commits clearly map to a feature, an inferred
date and commit SHA are listed to help trace when the work was introduced.

Guidance
--------
- If no explicit timestamp exists in the source prompt files, entries are marked with "(no timestamp provided)".
- Where a commit clearly implements or touches the feature, I've annotated the paraphrase with the
  relevant commit SHA(s) and date(s) (inferred mapping). This is not an authoritative provenance record
  but a best-effort mapping to help maintain history.
- To append new paraphrased prompts with exact timestamps, use:

```bash
python3 scripts/append_prompt_history.py "Paraphrase text" "Short outcome"
```

Original prompt (from `prompt.txt`) — (no timestamp provided)
------------------------------------------------------------
- Create a Go service that monitors a Tempest weather station and updates Apple HomeKit with sensors
  (temp, humidity, rain, wind, pressure). Provide examples including data key/token placeholders.
- Add `--loglevel` (default: error) and log sensor values when debug is enabled.
- Provide a modern web dashboard with unit toggles (C/F, mph/kph, in/mm) and wind direction.
- Add unit tests, graceful error handling, modular well-documented code, `REQUIREMENTS.md`, `CODE_REVIEW.md`,
  and a GitHub-standard `README.md`.

Refinements and feature prompts (paraphrased, timestamped where inferred)
-------------------------------------------------------------------------

UI / Dashboard
- Add missing HomeKit and Tempest info blocks below sensors on the web console. (no timestamp provided)
- Expand accessories row to show all accessories inline when clicked. (no timestamp provided; UI tests added 2025-10-18 `8f9b065`)
- Add context tables for lux values (use Wikipedia lux table) to the lux card. (no timestamp provided)
- Provide context for pressure and humidity cards; show "Feels like" temperature in humidity box. (no timestamp provided)
- Add a UV Index sensor card and use UV table color/risk text for context. (no timestamp provided)
- Show 5-year average temperature for the same day in the temperature card. (no timestamp provided)
- Adjust pressure to sea-level with a toggle to view raw/adjusted. (no timestamp provided)
- Add precipitation_type to the rain card and show local day rain accumulation. (no timestamp provided)
- Move all JS into `script.js`; keep HTML minimal; include API/response details when debug. (no timestamp provided)
- Allow dashboard grid to expand past 3 columns for larger screens. (no timestamp provided)
- Add a pressure trend line using saved history and add tooltips explaining the calculation. (no timestamp provided)
- Add a "Tempest Forecast" card populated from the Tempest API `/better_forecast`. (no timestamp provided)
- Chart time format improvements for wind and UV graphs. (inferred: commit `520c7d6` 2025-10-13; earlier chart work `c5171b7` 2025-10-02)
- Add battery level to the Tempest Station card. (no timestamp provided)

Data and history
- Add `--read-history` flag to preload historical observations (HISTORY_POINTS). Show progress/status while reading. (inferred: initial support `5a75cea` 2025-10-03; chart/history work `2c08bbd` 2025-09-29)
- Add `--use-generated-weather` to generate synthetic data for UI testing and history generation. (no timestamp provided)
- Add `--udp-stream` / `--station-url` to ingest Tempest UDP broadcasts; show source IP and packet count. (inferred commits: `0780d9f` 2025-10-03, `51b7137` 2025-10-03, `b598ae2` 2025-10-17)

HomeKit sensors and behavior
- Re-implement HomeKit sensors under "Tempest Weather" with numeric types and appropriate ranges. (no timestamp provided)
- Fix HomeKit humidity and lux sensors where needed. (no timestamp provided)

Alarms and notifications
- Add an Alarms subsystem with `--alarms @file.json` and `--alarms "json-string"`, supporting delivery
  methods (console, email, SMS, syslog, eventlog) with a flexible JSON schema and per-alarm templates.
  (inferred commits: initial `bbee062` 2025-10-09; editor/enhancements `aefb75c` 2025-10-11; docs reorg `75274f2` 2025-10-13)
- Add an alarm editor (`--alarms-edit` and `--alarms-edit-port`) as a standalone when used; re-read on file change. (inferred: same alarm commits)
- Delivery backends: console, OSLog(syslog on macOS), email (O365/SMTP), SMS (AWS SNS, Twilio planned), webhook/csv/json files. Add setup scripts and `.env.example`. (inferred: AWS SNS update `04fdccd` 2025-10-15; other email/alarm work early-mid Oct)
- Add `--webhook-listener` and `--test-webhook` for local webhook testing; pretty-print payloads. (inferred: webhook delivery commit `b2e5391` 2025-10-19)

Configuration, CLI, and env
- Add `--env <filename>` to override `.env`. (no timestamp provided)
- Use `.env` / `.env.example` for provider credentials and examples. (no timestamp provided)
- Make `--token`/`--station` required only when using WeatherFlow API stations; document in README. (no timestamp provided)

Testing, logging, and developer experience
- Add unit tests and target >70% coverage; run coverage and document flags in README. (no timestamp provided; unit-test-related commits: `3306dc6` 2025-10-08)
- When `--loglevel debug` set, include calculated values and API requests/responses; add `--logfilter <string>`. (no timestamp provided)
- Add `--history <value>` and `--chart-history <hours>` for stored history size and chart window. (inferred: support commit `5a75cea` 2025-10-03)

UX, docs, and release
- Provide multiple UI styles with a footer style dropdown; propose theme options. (no timestamp provided)
- Update README, REQUIREMENTS, and CODE_REVIEW; add authorship, roadmap, and LLM/Copilot usage notes. (no timestamp provided; docs reorg `75274f2` 2025-10-13)
- Prepare repo for public release: audit docs, keyword optimization, and add disclaimers. (no timestamp provided)

Other / advanced
- Add popout charts: `/chart/<weather-type>` dedicated pages that mirror card visuals and history. (inferred commits: `a3a63e4` 2025-09-22, `c5171b7` 2025-10-02, `520c7d6` 2025-10-13)
- Add DB delivery (Maria/MySQL) with secure config and test tooling. (no timestamp provided)
- Add per-alarm scheduling (daily/hourly/sunrise/sunset/days-of-week). (no timestamp provided)
- Add CSV/JSON file delivery backends with FIFO and temp-file fallback. (no timestamp provided)
- Request recurring documentation audits every 3–4 major changes or monthly. (no timestamp provided)

Appendix: automation
- Use `scripts/append_prompt_history.py` to append new paraphrased prompts with a UTC timestamp.

```bash
python3 scripts/append_prompt_history.py "Paraphrase of prompt" "Short outcome"
```

If you want more precise provenance I can:
- Scan commit diffs for each feature and attach the exact commit that introduced the code change.
- Append commit SHAs/PRs next to every paraphrase and reorder entries strictly by commit date.
PROMPT HISTORY — docs/development
=================================

This file paraphrases the development prompts that guided the implementation, taken from `docs/development/prompt.txt` (initial prompt) and `docs/development/prompt-refine.txt` (feature refinements).

Notes
-----
- Source files: `docs/development/prompt.txt`, `docs/development/prompt-refine.txt`.
- Many prompts in the source files are checklist items and do not include explicit timestamps; where no timestamp is present entries are marked "(no timestamp provided)".
- This document is intended to be updated whenever a new high-level prompt/refinement is applied. See `scripts/append_prompt_history.py` to append timestamped paraphrases automatically.

TODO: Keep this file current. Add new paraphrased prompts and an outcome when making AI-assisted edits.

Original prompt (from `prompt.txt`) — (no timestamp provided)
------------------------------------------------------------
- Request a Go service that monitors a Tempest weather station and updates Apple HomeKit, exposing sensors for temperature, humidity, rain, wind, and pressure. Include the data key and token in examples.
- Make logging configurable via `--loglevel` (default "error", support "info" and "debug"). When debug, log sensor values on each read.
- Provide a modern web dashboard showing station and HomeKit info; allow unit toggles (C/F, mph/kph, in/mm) and include wind direction.
- Add comprehensive unit tests and keep panic-free, graceful error handling with stderr output for errors.
- Keep code modular and well-documented. Generate `REQUIREMENTS.md` and a `CODE_REVIEW.md` from a code review. Ensure `README.md` follows GitHub standards and update `REQUIREMENTS.md` so it can be used to regenerate the app.

Refinements and feature prompts (from `prompt-refine.txt`) — (no timestamps provided unless indicated)
---------------------------------------------------------------------------------
UI / Dashboard
- Add missing HomeKit and Tempest info blocks below sensors on the web console. (no timestamp provided)
- Expand accessories row to show all accessories inline when clicked. (no timestamp provided)
- Add context tables for lux values (use Wikipedia lux table) to the lux card. (no timestamp provided)
- Provide context tables for pressure and humidity cards; show "Feels like" temperature in humidity box. (no timestamp provided)
- Add a UV Index sensor card and use the UV index table (media color + risk text) for context. (no timestamp provided)
- Show 5-year average temperature for the same day in the temperature card. (no timestamp provided)
- Adjust pressure to sea-level when displayed; allow toggle to show raw/adjusted. (no timestamp provided)
- Add precipitation_type to the rain card and show local day rain accumulation. (no timestamp provided)
- Move all JavaScript to `script.js`, keep HTML clean; in debug include calculated values, API calls and responses. (no timestamp provided)
- Allow dashboard grid to expand past 3 columns for large screens. (no timestamp provided)
- Add a pressure trend line using saved history and add tooltips explaining the calculation. (no timestamp provided)
- Add a "Tempest Forecast" card populated by `/better_forecast` from the Tempest API. (no timestamp provided)
-- Add chart time format improvements for wind and UV graphs. (inferred: 2025-10-13 — chart-related commits around this date)
-- Add `--read-history` flag to preload historical observations (up to HISTORY_POINTS, e.g., 200 points at 5-minute intervals) to reduce API rate limiting. Show progress/status while reading. (inferred: ~2025-09-29 to 2025-10-02 — related commits around late Sep / early Oct)
-- Add `--udp-stream` / `--station-url` support to ingest data directly from the Tempest UDP broadcast; record packet count, IP, and show data source as udp-stream. If no UDP packets, show "No UDP Packets"; log UDP packets in debug mode. (inferred: 2025-10-03 → 2025-10-17 — multiple commits referenced udp-stream and UDP status integration)
-- Add an Alarms feature with `--alarms @file.json` or `--alarms "json-string"`, supporting delivery methods: console, email, SMS, syslog, eventlog. Design a flexible JSON schema with global provider config and per-alarm templates. Include tag-based filtering. (inferred: 2025-10-11 → 2025-10-13 — commits added alarm editor and reorganized alarm docs)
-- Implement delivery backends: console, syslog/OSLog (macOS), email (Microsoft 365/O365 Exchange and SMTP fallback), SMS (AWS SNS, Twilio planned), and webhook/csv/json file outputs. Add setup scripts and `.env.example` entries for provider credentials. (inferred: AWS SNS updates around 2025-10-15; related email/alarm commits in early to mid October)
-- Add popout charts feature: dedicated `/chart/<weather-type>` page with parity in units, colors, history, and lines. (inferred: 2025-09-22 → 2025-10-13 — several chart and popout commits during this period)
- Add `--udp-stream` / `--station-url` support to ingest data directly from the Tempest UDP broadcast; record packet count, IP, and show data source as udp-stream. If no UDP packets, show "No UDP Packets"; log UDP packets in debug mode. (no timestamp provided)

HomeKit sensors and behavior
- Re-implement HomeKit sensors grouped under "Tempest Weather"; ensure sensors report numeric values and use appropriate types and ranges (wind average/gust, wind direction, temperature, humidity, lux, uv, rain, precipitation type, lightning count/distance, etc.). (no timestamp provided)
- Fix HomeKit humidity and lux sensors where needed. (no timestamp provided)

Alarms and notifications
- Add an Alarms feature with `--alarms @file.json` or `--alarms "json-string"`, supporting delivery methods: console, email, SMS, syslog, eventlog. Design a flexible JSON schema with global provider config and per-alarm templates. Include tag-based filtering. (no timestamp provided)
- Add an alarm editor (`--alarms-edit @file.json` and `--alarms-edit-port`) as a standalone editor that can filter by name/tags, view the active alarm JSON, and persists changes across platforms; the app should re-read alarms on file changes. (no timestamp provided)
- Support per-delivery defaults and templates; add variables for app-info, alarm-info, and sensor-info; allow HTML email option. Validate conditions on save and provide a paraphrase. (no timestamp provided)
- Implement delivery backends: console, syslog/OSLog (macOS), email (Microsoft 365/O365 Exchange and SMTP fallback), SMS (AWS SNS, Twilio planned), and webhook/csv/json file outputs. Add setup scripts and `.env.example` entries for provider credentials. (no timestamp provided)
- Add `--webhook-listener <port>` and `--test-webhook` for local webhook testing; pretty-print received JSON in the console. (no timestamp provided)

Configuration, CLI, and env
- Add `--env <filename>` to override default `.env`. (no timestamp provided)
- Use `.env` and `.env.example` for configuration; expand them to include email/SMS provider parameters. (no timestamp provided)
- Make `--token` and `--station` required only for WeatherFlow API stations; document usage in README. (no timestamp provided)

Testing, logging, and developer experience
- Add unit tests with a target >70% coverage; run coverage, and document test flags in README. (no timestamp provided)
- When `--loglevel debug` is set, include calculated values and API requests/responses in logs; add `--logfilter <string>` to filter logs to messages containing the string. (no timestamp provided)
- Add `--history <value>` and `--chart-history <hours>` to configure stored history length and chart display windows; attempt allocation and report failures. (no timestamp provided)

UX, docs, and release
- Provide multiple UI styles via CSS and a footer dropdown to switch styles; propose a few clean styles. (no timestamp provided)
- Update README.md, REQUIREMENTS.md, and CODE_REVIEW.md to reflect features, authorship, and the use of LLMs/Copilot in development; add a roadmap section (Alarms, Email/SMS/Syslog/EventLog, multi-station, container support). (no timestamp provided)
- Prepare the codebase for public GitHub release: clean docs, keyword optimization (vibe, macOS, HomeKit, tempest, weather), and explicit disclaimers that it is a work in progress. (no timestamp provided)

Other / advanced
- Add popout charts feature: dedicated `/chart/<weather-type>` page with parity in units, colors, history, and lines. (no timestamp provided)
- Add DB delivery method (Maria/MySQL) with secure connection configuration and test tools; include schema and usage docs. (no timestamp provided)
- Add per-alarm scheduling (daily/hourly/sunrise/sunset/dow) for delivery windows. (no timestamp provided)
- Add webhook delivery method and a local webhook listener mode for testing. (no timestamp provided)
- Add CSV/JSON file delivery backends with FIFO and fallback temp file behavior. (no timestamp provided)
- Request recurring documentation audits every 3–4 major changes or monthly to ensure `REQUIREMENTS.md`, `README.md`, and `CODE_REVIEW.md` match the codebase. (no timestamp provided)

Appendix: automation
- Use `scripts/append_prompt_history.py` to append new paraphrased prompts with a UTC timestamp. Example:

```bash
python3 scripts/append_prompt_history.py "Paraphrase of prompt" "Short outcome"
```

If you'd like, I will now generate `docs/development/PROMPT_HISTORY.md` (this file) and then mark the todo complete. If you want timestamps added based on commit dates, I can scan commits and attach inferred dates to each paraphrase.
