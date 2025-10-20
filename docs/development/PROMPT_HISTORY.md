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
