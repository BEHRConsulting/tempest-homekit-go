# Provider Configuration Validation

## Overview

The alarm system now validates that all necessary environment variables are configured for the delivery methods used in your alarm rules. This validation runs automatically when:

1. The application starts and loads alarm configuration
2. The alarm configuration file is reloaded (via file watcher or manual reload)

## Behavior

When an alarm is configured to use email or SMS delivery, the system checks:

- **Email Delivery**: Are the necessary SMTP or MS365 OAuth2 credentials configured?
- **SMS Delivery**: Are the necessary Twilio or AWS SNS credentials configured?

### Log Level

All validation messages are logged at **INFO** level. This means:
- They appear when log level is set to `info` or `debug`
- They do not appear at `warn` or `error` levels
- They are informational only and do not prevent the application from starting

### Validation Scope

- Only **enabled** alarms are checked
- Only **email** and **sms** delivery methods trigger validation
- **console**, **syslog**, **oslog**, and **eventlog** channels do not require configuration

## Validation Messages

### No Provider Configured

If an alarm uses email but no email provider is set up:

```
⚠️  Email delivery is configured in alarms, but no email provider is configured.
    Set either SMTP_* or MS365_* environment variables in .env file.
    For SMTP: SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, SMTP_FROM_ADDRESS
    For MS365: MS365_CLIENT_ID, MS365_CLIENT_SECRET, MS365_TENANT_ID, MS365_FROM_ADDRESS
```

If an alarm uses SMS but no SMS provider is set up:

```
⚠️  SMS delivery is configured in alarms, but no SMS provider is configured.
    Set either Twilio or AWS SNS environment variables in .env file.
    For Twilio: TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_FROM_NUMBER
    For AWS SNS: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION
```

### Missing Required Variables

If a provider is partially configured (e.g., only `SMTP_HOST` is set):

```
⚠️  SMTP email is configured but missing required environment variables: SMTP_USERNAME, SMTP_PASSWORD, SMTP_FROM_ADDRESS
```

```
⚠️  Twilio SMS is configured but missing required environment variables: TWILIO_AUTH_TOKEN, TWILIO_FROM_NUMBER
```

## Examples

### Example 1: SMTP Email Setup

**Alarm configuration:**
```json
{
  "alarms": [{
    "name": "High Temperature Alert",
    "enabled": true,
    "condition": "temperature > 35C",
    "channels": [{
      "type": "email",
      "email": {
        "to": ["alerts@example.com"],
        "subject": "Temperature Alert",
        "body": "Temperature exceeded threshold"
      }
    }]
  }]
}
```

**Missing variables in .env:**
- No email configuration at all

**Validation output (INFO level):**
```
INFO: ⚠️  Email delivery is configured in alarms, but no email provider is configured.
INFO:     Set either SMTP_* or MS365_* environment variables in .env file.
INFO:     For SMTP: SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, SMTP_FROM_ADDRESS
INFO:     For MS365: MS365_CLIENT_ID, MS365_CLIENT_SECRET, MS365_TENANT_ID, MS365_FROM_ADDRESS
```

**Fix:** Add to .env:
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=alerts@example.com
SMTP_PASSWORD=your_app_password
SMTP_FROM_ADDRESS=alerts@example.com
SMTP_USE_TLS=true
```

### Example 2: Partial Twilio Configuration

**Alarm configuration:**
```json
{
  "alarms": [{
    "name": "Storm Warning",
    "enabled": true,
    "condition": "*lightning_count && lightning_distance < 5",
    "channels": [{
      "type": "sms",
      "sms": {
        "to": ["+15555551234"],
        "message": "Lightning detected nearby!"
      }
    }]
  }]
}
```

**Partial configuration in .env:**
```bash
TWILIO_ACCOUNT_SID=AC1234567890abcdef
# Missing: TWILIO_AUTH_TOKEN and TWILIO_FROM_NUMBER
```

**Validation output (INFO level):**
```
INFO: ⚠️  Twilio SMS is configured but missing required environment variables: TWILIO_AUTH_TOKEN, TWILIO_FROM_NUMBER
```

**Fix:** Add to .env:
```bash
TWILIO_ACCOUNT_SID=AC1234567890abcdef
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_FROM_NUMBER=+15555550000
```

## Testing

Validation logic is tested in `pkg/alarm/manager_validation_test.go` with comprehensive test cases:

1. ✅ No delivery methods used (console only) - no warnings
2. ✅ Email used but no provider configured - warning shown
3. ✅ SMTP partially configured - lists missing variables
4. ✅ SMS used but no provider configured - warning shown
5. ✅ Twilio partially configured - lists missing variables
6. ✅ Disabled alarms are ignored - no validation
7. ✅ Fully configured providers - no warnings

Run tests:
```bash
go test -v ./pkg/alarm/ -run TestValidateConfigProviders
```

## Implementation Details

### Function Location

Validation is implemented in `pkg/alarm/manager.go`:

```go
func validateConfigProviders(config *AlarmConfig)
```

This function:
1. Scans all enabled alarms to detect email/SMS usage
2. Checks if providers are configured (from environment variables)
3. For configured providers, validates all required fields are present
4. Logs INFO-level warnings with specific missing variable names
5. Provides setup instructions for each provider type

### Integration Points

The validation function is called from:

1. **`NewManager()`** - During initial alarm manager creation
2. **`reloadConfig()`** - When configuration file changes are detected

This ensures validation runs:
- At application startup
- Whenever alarms are modified and reloaded

### Log Output

All validation uses the custom logger package (`pkg/logger`):
```go
logger.Info("⚠️  Email delivery is configured...")
```

The logger respects the global log level setting. To see validation warnings:
```bash
# In .env file
LOG_LEVEL=info
```

Or via command line:
```bash
./tempest-homekit-go --log-level=info ...
```

## Related Documentation

- [Alarm Configuration Guide](ALARM_CONFIGURATION.md) - Complete environment variable setup
- [Alarm Validation](ALARM_VALIDATION.md) - Condition validation and paraphrase features
- [Architecture: Data Sources](../ARCHITECTURE_DATA_SOURCES.md) - Overall system architecture

## Best Practices

1. **Always use environment variables** for credentials - never hardcode in JSON
2. **Set log level to `info`** during initial setup to see validation messages
3. **Test alarm delivery** after configuration to ensure credentials work
4. **Keep .env file secure** - never commit to version control (already in .gitignore)
5. **Use disabled alarms** during testing to avoid validation warnings

## Troubleshooting

### Not seeing validation messages?

Check your log level:
```bash
# .env file
LOG_LEVEL=info  # Must be 'info' or 'debug'
```

### Validation passes but delivery fails?

The validation only checks if environment variables are **set**, not if they are **correct**. You may need to:
- Verify credentials are valid
- Check network connectivity
- Review provider-specific logs
- Test with a simple alarm first

### Want to suppress validation temporarily?

Set log level to `warn` or `error`:
```bash
LOG_LEVEL=warn
```

Or disable the alarm temporarily while testing:
```json
{
  "name": "Test Alarm",
  "enabled": false,  // Won't be validated
  ...
}
```
