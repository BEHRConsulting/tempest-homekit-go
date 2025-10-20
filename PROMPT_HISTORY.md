PROMPT HISTORY — Tempest HomeKit Go
===================================

This document captures paraphrased prompts used during development that guided the AI-assisted edits, features, and documentation of the project. It is intended as a living log and should be updated whenever a new high-level prompt is used to make changes to the codebase.

Keep this file current: add a short paraphrase of the prompt, date, and high-level outcome for each new prompt.

TODO: Maintain this file; add new prompts and short outcomes when used. (Task: Maintain prompt history)

Prompts (paraphrased)
---------------------

- [2025-09-10] "Make popout charts deterministic and match small-card visuals exactly (per-dataset styles, units) and add headless tests verifying parity." — Outcome: compact config encoding for popout charts and headless tests.

- [2025-09-12] "Harden headless tests to avoid CDN timing flakiness by injecting vendored Chart.js and exposing in-page test hooks." — Outcome: vendored dependencies and in-page hooks for robust chromedp tests.

- [2025-09-18] "The alarm 'Lux Change' is not triggering after these observations; here's the logs — find and fix the bug." — Outcome: fixed state persistence bug in ProcessObservation().

- [2025-09-25] "Remove provider credentials from example alarm JSON files; store provider config in .env and update docs." — Outcome: canonical `docs/ALARM_CONFIGURATION.md` (rules-only), `.env` usage clarified, and docs updated accordingly.

- [2025-10-10] "Normalize documentation style: remove emoji and unicode icons from docs, trim trailing whitespace, and run a link-check to ensure internal cross-references are valid." — Outcome: normalization script added and many Markdown files cleaned; internal md links validated.

- [2025-08-22] "Implement AWS SNS SMS notifications with environment-first credential handling and setup script." — Outcome: AWS SNS implementation, `scripts/setup-aws-sns.sh`, docs updated.

- [2025-08-29] "Add Microsoft 365 OAuth2 email support, with Azure AD setup docs and fallback to SMTP when credentials are missing." — Outcome: O365 email implementation, docs, and example configs.

- [2025-09-05] "Add change-detection operators to alarms (*field, >field, <field) with cooldown and independent per-field state." — Outcome: change-detection feature implemented with tests and editor support.

- [2025-10-20] "Provide a concise WORK_LOG.md capturing major design decisions, testing, and AI-assisted actions." — Outcome: `WORK_LOG.md` created and maintained.


History
-------
- 2025-10-20 — Initial prompt history file created to track paraphrased prompts and outcomes.


If you want, I can:
- Expand each entry with exact timestamps and the original prompt text (if available).
- Auto-update this file whenever I run a new AI-assist prompt; I can append an entry with date and paraphrase on each change.

- [2025-10-20T16:59:17Z] "Test append: created helper script to append prompts" — Outcome: Verified script writes a timestamped entry
