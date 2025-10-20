WORK LOG — Tempest HomeKit Go
================================

Overview
--------
This document captures a concise, developer-focused work log that explains what went into creating and iterating on the Tempest HomeKit Go application. It records high-level design decisions, notable code changes, testing and CI activities, contributor acknowledgements (human + AI), estimated time spent on major tasks, and lessons learned.

This file was generated and curated by the development team and with assistance from an automated coding assistant.

Contributors
------------
- Primary human developer: (You) — feature design, tests, JS/CSS tweaks, service wiring, release decisions.
- AI assistant (Copilot): generated code suggestions, test hardening, automated edits, documentation updates, CI workflow drafts and diagnostic tests.

Timeline & Estimated Effort
---------------------------
These are approximate times aggregated across the development sessions that produced the current repository state.
- Project scaffold and initial feature set (service + web UI + HomeKit): ~40-60 hours
- Popout/dedicated chart feature (design + implementation + tests): ~6-10 hours
- Headless (chromedp) test hardening and diagnostic tests: ~4-6 hours
- Unit tests and integration fixes: ~3-5 hours
- Release version bump, tag, and release workflow drafting: ~2-4 hours
- Documentation, README changes, acknowledgements, and housekeeping: ~2-3 hours
- Iterative debugging and CI troubleshooting: ~2-4 hours
- Alarm system development (editor, JSON validation, debug logging): ~8-10 hours
- Alarm change detection state persistence fix: ~2 hours

Major Design Changes
--------------------
- Web UI architecture:
 - Centralized frontend logic placed in `pkg/web/static/script.js` and static popout template in `pkg/web/static/chart.html` to keep HTML minimal.
 - Charts use vendored Chart.js and date adapters for deterministic rendering.
- Deterministic Popout Configuration:
 - Chart popouts now receive a compact encoded `config` payload containing per-dataset metadata and `incomingUnits`. This ensures the popout graph closely matches small-card visuals (colors, dashes, fills).
- Headless Tests and Determinism:
 - Headless chromedp tests were hardened to prefer in-page hooks (e.g., `window.__dashboardReady`, `window.__lastStatusRaw`) and to inject vendored Chart.js and local scripts to avoid CDN flakiness.
- Unit and Label Handling:
 - Unit label helpers (`prettyUnitLabel`) were restored and harmonized between the dashboard and popout so both show consistent units.
- Release Automation:
 - Drafted a GitHub Actions `release.yml` to build cross-platform binaries and upload assets when an annotated tag is pushed. (Push of workflow file may require additional permissions.)
- Accessibility / UI polish:
 - Tempest Station link label was truncated to 15 characters for card layout with full URL in `title` and `aria-label` for hover and screen readers.
- Alarm System Architecture:
 - Implemented comprehensive alarm system with change-detection operators (`*field`, `>field`, `<field`)
 - Warning log level added (warn/warning aliases) between info and error
 - Alarm editor web UI with template variable system (18 variables including `alarm_description`)
 - JSON validation with line/column error reporting and helpful hints for missing @ prefix
 - Enhanced debug logging with pretty JSON output and detailed evaluation traces
 - **Critical Fix**: Changed `ProcessObservation()` to work with original alarms instead of copies, preserving `previousValue` state between calls. This fixed change-detection operators that were resetting state on every observation.

Best & Worst Prompts (AI-assisted development)
----------------------------------------------
- Best prompts (helpful):
 - "Make popout charts deterministic and match small-card visuals exactly (per-dataset styles, units) and add headless tests verifying parity." — This prompt led to compact, testable config encoding and robust headless tests.
 - "Harden headless tests to avoid CDN timing flakiness by injecting vendored Chart.js and exposing in-page test hooks." — This improved reliability in CI.
 - "The alarm 'Lux Change' is not triggering after these observations..." with full log output — Provided concrete reproduction case that led to discovering the state persistence bug in `ProcessObservation()`.

- Worst/ambiguous prompts (costly):
 - Broad requests like "Add release automation" without specifying how to handle GitHub token permissions led to local-only workflow drafts and a failed push due to token scope restrictions. Lesson: explicitly mention token policy or request a PR instead of direct pushes.
 - Vague UI change requests without specifying exact truncation length or accessibility expectations required follow-up decisions.

Notable Files and Where Changes Happened
----------------------------------------
- `pkg/web/static/script.js` — Centralized frontend; chart creation, popout config encoding, dashboard updates, and station card rendering.
- `pkg/web/static/chart.html` — Dedicated popout page; unit detection and tick formatting.
- `pkg/web/ui_headless_test.go` and `pkg/web/popout_diagnostics_test.go` — Headless tests and diagnostic runner to capture console logs and in-page errors.
- `main.go` — Version bump and startup changes.
- `README.md`, `REQUIREMENTS.md`, `CODE_REVIEW.md` — Documentation updates and acknowledgements.
- `.github/workflows/release.yml` — Drafted release workflow (local). Needs PR or push with workflow scope.

Testing & CI
------------
- Unit tests: `go test ./...` run locally and pass in the developer environment.
- Headless tests: chromedp-based UI tests were added and hardened to avoid CDN/time-related flakiness. One diagnostic test captures `window.__popoutError` and console logs.
- CI notes: Adding workflow changes requires a token with `workflow` permission or merging via a PR. The release workflow was drafted locally and the tag `v1.4.1` was created and pushed.

Accessibility & UX
------------------
- Popout and dashboard charts now share consistent unit labeling.
- Station link truncation added with `title` and `aria-label` so the full URL is accessible to screen readers.

Lessons Learned
---------------
- Vendor critical frontend dependencies (Chart.js/date adapter) for deterministic headless testing.
- Expose small test-only hooks in-page (`window.__dashboardReady`, `window.__lastStatusRaw`) to make headless tests less brittle.
- Keep UI text truncation consistent and add accessible attributes rather than relying on CSS-only truncation.
- When automating pushes to GitHub workflows, ensure the push token has `workflow` permissions or create a PR for manual merge.

Time Tracking (approximate)
---------------------------
- See "Timeline & Estimated Effort" above for task-level estimates. Keep these as a living estimate; they can be refined by tracking commits and timestamps if desired.

Risks & Open Items
------------------
- Release workflow push: requires a PAT with `workflow` scope or a PR merge by a maintainer with sufficient permissions.
- CI environment must have headless Chromium available for chromedp tests; vendoring Chart.js reduces but does not eliminate flakiness.
- More conservative assertions could be added to headless diagnostic tests to turn diagnostics into strict parity checks.

Next recommended actions
------------------------
1. Commit and push the `WORK_LOG.md` and the recent UI changes. If you want the release workflow applied, either push with a token that has `workflow` scope or open a PR so it can be merged by a maintainer.
2. Optionally, add a `CHANGELOG.md` with the `v1.4.1` notes and link to the release workflow.
3. Consider adding more strict headless assertions to ensure exact visual parity (colors, dataset counts) if you need CI-based enforcement.

Appendix: Example commit/PR message
----------------------------------
Title: "UI: truncate Tempest Station link + add WORK_LOG.md; tests: headless fixes"

Body:
- Truncate station link label to 15 chars; full URL in title/aria-label
- Add `WORK_LOG.md` capturing project history and AI-assisted work
- Headless test hardening and popout parity tests
- Drafted `release.yml`; needs PR or PAT with workflow scope to enable

---

If you'd like, I can now:
- Commit & push these changes and open a PR, or
- Expand the `WORK_LOG.md` with more granular timestamps (per-commit breakdown) and link to the exact commit SHAs and diffs.

Tell me how you'd like to proceed (push/PR, or more detail in the work log).