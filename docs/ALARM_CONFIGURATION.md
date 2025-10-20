# Alarm Configuration

This document describes the structure and purpose of alarm JSON files used by the application.

Important: alarm JSON files must contain only alarm rules and per-alarm notification preferences. They must not include provider credentials, secrets, or other runtime configuration. All runtime/provider configuration belongs in the repository root `.env` file or as environment variables. Use `.env.example` as the canonical template for which variables to set.

## Alarm JSON structure

Alarm files contain a single top-level `alarms` array. Each alarm entry describes the rule, metadata, and the channels used for notifications.

Example (rules-only):

```json
{
 "alarms": [
 {
 "name": "high-temperature",
 "description": "Alert when temperature exceeds threshold",
 "tags": ["temperature", "heat"],
 "enabled": true,
 "condition": "temperature > 85F",
 "cooldown": 1800,
 "channels": [
 {
 "type": "console",
 "template": "Temperature: Temperature: {{temperature}}°F"
 },
 {
 "type": "email",
 "email": {
 "to": ["admin@example.com"],
 "subject": "High Temperature: {{temperature}}°F",
 "body": "{{alarm_info}}\\n\\n{{sensor_info}}",
 "html": true
 }
 }
 ]
 }
 ]
}
```

Common fields:
- `name` (string) — unique identifier for the alarm
- `description` (string) — human-friendly description
- `tags` (array) — optional tags useful for grouping
- `enabled` (boolean) — whether the alarm is active
- `condition` (string) — expression evaluated against sensor data
- `cooldown` (seconds) — minimum time between successive notifications
- `channels` (array) — notification channels and their per-alarm templates/recipients

## Where to put provider configuration

- DO store SMTP, MS365, Twilio, AWS SNS, syslog, webhook listener ports, and similar credentials or runtime options in `.env` (or as real environment variables).
- DO NOT store secrets or provider credentials in `alarms.json`.
- Use `.env.example` in the repository as the canonical template for which variables to set.

## Migrating legacy alarm files

If your alarm JSON files currently embed provider settings:

1. Extract credentials from the alarm JSON.
2. Add those values to your `.env` (or environment variables) using `.env.example` as a guide.
3. Remove the provider-specific sections from the alarm JSON so it contains only the `alarms` array.
4. Restart the application and verify alarms trigger correctly.

## Testing and troubleshooting

- Use the application's test helpers to validate provider configuration, e.g. `./tempest-homekit-go --test-email <recipient>`.
- If you see an error stating "No email configuration found", confirm required variables are set in `.env` or the environment.

## References

- `.env.example` — canonical list of runtime/provider variables
- `pkg/alarm/types.go` — configuration loading helpers
- `pkg/alarm/notifiers.go` — notifier factory implementation
- `pkg/alarm/emailtest.go` — helper used by `--test-email`

## Version history

- v1.7.0+: provider credentials are loaded from environment variables by default
 - `cooldown` (seconds) — minimum time between successive notifications
 - `channels` (array) — notification channels and their per-alarm templates/recipients

 ## Where to put provider configuration

 - DO store SMTP, MS365, Twilio, AWS SNS, syslog, webhook listener ports, and similar credentials or runtime options in `.env` (or as real environment variables).
 - DO NOT store secrets or provider credentials in `alarms.json`.
 - Use `.env.example` in the repository as the canonical template for which variables to set.

 ## Migrating legacy alarm files

 If your alarm JSON files currently embed provider settings, migrate as follows:

 1. Extract credentials from the alarm JSON.
 2. Add those values to your `.env` (or environment variables) using `.env.example` as a guide.
 3. Remove the provider-specific sections from the alarm JSON so it contains only the `alarms` array.
 4. Restart the application and verify alarms trigger correctly.

 ## Testing and troubleshooting

 - Use the application's test helpers to validate provider configuration, e.g. `./tempest-homekit-go --test-email <recipient>`.
 - If you see an error stating "No email configuration found", confirm required variables are set in `.env` or the environment.

 ## References

 - `.env.example` — canonical list of runtime/provider variables
 - `pkg/alarm/types.go` — configuration loading helpers
 - `pkg/alarm/notifiers.go` — notifier factory implementation
 - `pkg/alarm/emailtest.go` — helper used by `--test-email`

 ## Version history

 - v1.7.0+: provider credentials are loaded from environment variables by default
 }
 ]
 }
 ```

 Fields you will commonly use:
 - `name` (string) — unique identifier for the alarm
 - `description` (string) — human-friendly description
 - `tags` (array) — optional tags useful for grouping
 - `enabled` (boolean) — whether the alarm is active
 - `condition` (string) — expression evaluated against sensor data
 - `cooldown` (seconds) — minimum time between successive notifications
 - `channels` (array) — notification channels and their per-alarm templates/recipients

 ## Where to put provider configuration

 - DO store SMTP, MS365, Twilio, AWS SNS, syslog, webhook listener ports, and similar credentials or runtime options in `.env` (or as real environment variables).
 - DO NOT store secrets or provider credentials in `alarms.json`.
 - Use `.env.example` in the repository as the canonical template for which variables to set.

 ## Migrating legacy alarm files

 If your alarm JSON files currently embed provider settings, migrate as follows:

 1. Extract credentials from the alarm JSON.
 2. Add those values to your `.env` (or environment variables) using `.env.example` as a guide.
 3. Remove the provider-specific sections from the alarm JSON so it contains only the `alarms` array.
 4. Restart the application and verify alarms trigger correctly.

 ## Testing and troubleshooting

 - Use the application's test helpers to validate provider configuration, e.g. `./tempest-homekit-go --test-email <recipient>`.
 - If you see an error stating "No email configuration found", confirm required variables are set in `.env` or the environment.

 ## References

 - `.env.example` — canonical list of runtime/provider variables
 - `pkg/alarm/types.go` — configuration loading helpers
 - `pkg/alarm/notifiers.go` — notifier factory implementation
 - `pkg/alarm/emailtest.go` — helper used by `--test-email`

 ## Version history

 - v1.7.0+: provider credentials are loaded from environment variables by default
