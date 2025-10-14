Tests and integration scripts
=============================

This folder contains convenience shell scripts used for local integration and manual testing of
the `tempest-homekit-go` binary. These are not part of the automated Go unit tests but are
helpful for reproducing real-world behaviors (file reloads, alarms, long-poll observations).

Usage
-----

1. Build the binary in the repo root:

```bash
go build
```

2. Run any test script from the repository root (scripts assume the built binary is `./tempest-homekit-go`):

```bash
./tests/test-env.sh
./tests/test-alarm-console.sh
```

Notes
-----
- Many scripts are time-based (long sleeps) because they simulate real polling intervals. Expect
  some scripts to run several minutes.
- Scripts write logs to `/tmp/*.log` (see the script headers). Adjust paths as needed.
- These scripts are intended for local/manual testing only. For CI use, convert them into
  deterministic Go tests or shorten polling intervals to avoid long-running jobs.

Scripts
-------
- `test-env.sh` — quick check that `.env` values are loaded (uses `HISTORY_POINTS`).
- `test-alarm-console.sh` — starts the app with `--loglevel warning` and verifies alarm console output.
- `test-alarm-validation.sh` — runs several quick checks that invalid alarm JSON shows helpful errors.
- `test-alarm-reload.sh` — verifies the file-watcher reload path when the alarm file is touched.
- `test-alarm-reload-modified.sh` — modifies the alarm file content and ensures the reload occurs.
- `test-lux-alarm.sh` — short run that verifies the Lux Change alarm triggers in a single observation window.
- `test-lux-change-alarm.sh` — long run (≈130s) to allow two observations and trigger lux-change logic.
- `test-wind-previous-value.sh` — long run to verify previous/current wind values are handled correctly.
- `test-enhanced-alarm-message.sh` — long run to exercise enhanced alarm message formatting.

Contributing
------------
If you convert any script into a unit/integration test under `pkg/` or `tests/` (Go), please
add a short README entry and, where appropriate, reduce or mock real-time waits to keep CI fast.
