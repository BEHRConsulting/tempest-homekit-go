# Alarm Configuration Architecture

## Overview

The alarm system uses a **clean separation of concerns** between alarm rules and provider credentials:

- **Alarm JSON files** (`alarms.json`) → Contains ONLY alarm rules and notification preferences
- **Environment variables** (`.env` file) → Contains ALL email/SMS provider credentials

This design provides:
- ✅ **Security**: Credentials never stored in version control
- ✅ **Simplicity**: One place to configure providers, many alarm files can reference them
- ✅ **Flexibility**: Switch providers by changing `.env` without modifying alarm rules

## Configuration Flow

```
┌─────────────────┐
│   .env file     │
│                 │
│ SMTP_HOST=...   │
│ SMTP_PASSWORD=..│
│ TWILIO_SID=...  │
└────────┬────────┘
         │
         │ LoadConfigFromEnv()
         ↓
┌─────────────────┐        ┌──────────────────┐
│ Environment     │        │  alarms.json     │
│ Config Object   │ ←merge─┤                  │
│                 │        │  {               │
│ Email:  {...}   │        │    "alarms": [   │
│ SMS:    {...}   │        │      {...}       │
│ Syslog: {...}   │        │    ]             │
└────────┬────────┘        │  }               │
         │                 └──────────────────┘
         │
         ↓
┌─────────────────┐
│ AlarmConfig     │
│                 │
│ • Email config  │
│ • SMS config    │
│ • Alarm rules   │
└─────────────────┘
```

## Configuration Precedence

**Environment variables ALWAYS take precedence over JSON configuration.**

If you have email config in both `.env` and `alarms.json`, the `.env` values are used:

```go
// Environment config overrides JSON config
envConfig, _ := LoadConfigFromEnv()
if envConfig.Email != nil {
    config.Email = envConfig.Email  // Environment wins
}
```

## Supported Providers

### Email Providers

**1. Generic SMTP** (Gmail, Office 365 SMTP, SendGrid, Mailgun, etc.)
```bash
# .env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_ADDRESS=alerts@example.com
SMTP_FROM_NAME=Weather Alerts
SMTP_USE_TLS=true
```

**2. Microsoft 365 OAuth2** (Enterprise, more secure)
```bash
# .env
MS365_CLIENT_ID=your-client-id
MS365_CLIENT_SECRET=your-client-secret
MS365_TENANT_ID=your-tenant-id
MS365_FROM_ADDRESS=alerts@yourdomain.com
SMTP_FROM_NAME=Weather Alerts
```

### SMS Providers

**1. Twilio**
```bash
# .env
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=your-auth-token
TWILIO_FROM_NUMBER=+15555551234
```

**2. Amazon SNS**
```bash
# .env
AWS_ACCESS_KEY_ID=AKIAXXXXXXXX
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
AWS_SNS_TOPIC_ARN=arn:aws:sns:us-east-1:123456789012:topic
```

### Syslog (Optional)

```bash
# .env
SYSLOG_NETWORK=tcp
SYSLOG_ADDRESS=localhost:514
SYSLOG_PRIORITY=warning
SYSLOG_TAG=tempest-weather
```

## Alarm JSON File Structure

Alarm JSON files contain **ONLY** alarm rules, no credentials:

```json
{
  "alarms": [
    {
      "name": "high-temperature",
      "description": "Alert when temperature exceeds 85°F",
      "tags": ["temperature", "heat"],
      "enabled": true,
      "condition": "temperature > 85F",
      "cooldown": 1800,
      "channels": [
        {
          "type": "console",
          "template": "🌡️ Temperature: {{temperature}}°F"
        },
        {
          "type": "email",
          "email": {
            "to": ["admin@example.com"],
            "subject": "High Temperature: {{temperature}}°F",
            "body": "{{alarm_info}}\n\n{{sensor_info}}",
            "html": true
          }
        },
        {
          "type": "sms",
          "sms": {
            "to": ["+15555551234"],
            "message": "High temp: {{temperature}}°F"
          }
        }
      ]
    }
  ]
}
```

## Example Files

The repository includes three example alarm files:

1. **`alarms.example.json`** - Basic examples, works with any provider
2. **`alarms-ms365.example.json`** - Same alarms, with MS365 setup notes
3. **`alarms-aws.example.json`** - Same alarms, with AWS SNS setup notes

All three files have identical alarm rules. The only difference is the setup instructions in the comments showing which environment variables to configure.

## Migration Guide

### If you have old alarm files with embedded credentials:

**Before** (old format - DON'T use):
```json
{
  "email": {
    "provider": "smtp",
    "smtp_host": "smtp.example.com",
    "smtp_port": 587,
    "username": "user@example.com",
    "password": "secret-password",
    "from_address": "alerts@example.com"
  },
  "alarms": [...]
}
```

**After** (new format - DO use):

`.env` file:
```bash
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=user@example.com
SMTP_PASSWORD=secret-password
SMTP_FROM_ADDRESS=alerts@example.com
```

`alarms.json` file:
```json
{
  "alarms": [...]
}
```

### Migration Steps:

1. **Extract credentials from your alarm JSON file**
2. **Add them to your `.env` file** (or set as environment variables)
3. **Remove the `email`, `sms`, and `syslog` sections from your alarm JSON**
4. **Keep only the `alarms` array**
5. **Test**: `./tempest-homekit-go --alarms @alarms.json`

## Security Best Practices

✅ **DO:**
- Store credentials in `.env` file
- Add `.env` to `.gitignore`
- Use `.env.example` as a template (with fake values)
- Use environment-specific files (`.env.production`, `.env.staging`)
- Rotate credentials regularly

❌ **DON'T:**
- Put credentials in alarm JSON files
- Commit `.env` to version control
- Share credentials in documentation
- Use production credentials in examples

## Testing

Test your email configuration:
```bash
./tempest-homekit-go --test-email user@example.com
```

This will:
1. Load email config from `.env`
2. Validate all required credentials
3. Prompt for a test recipient
4. Send a test email

## Troubleshooting

### "No email configuration found"

**Cause**: Neither SMTP nor MS365 credentials are set in environment variables.

**Solution**: Set either SMTP_* or MS365_* variables in your `.env` file.

### "Email config in alarms.json is ignored"

**Expected behavior**: Environment variables always take precedence. If you have email config in your `.env` file, any email config in the JSON is ignored. This is by design for security.

### "Want to use different providers for different alarms"

**Not supported**: The system uses one email provider and one SMS provider globally. All alarms use the same providers. This is intentional to keep configuration simple and secure.

**Workaround**: Run multiple instances of the application with different `.env` files and alarm configurations.

## Code References

- **Configuration Loading**: `pkg/alarm/types.go` - `LoadConfigFromEnv()`
- **Environment Override**: `pkg/alarm/types.go` - `LoadAlarmConfig()`
- **Notifier Factory**: `pkg/alarm/notifiers.go` - `NewNotifierFactory()`
- **Email Test**: `pkg/alarm/emailtest.go` - `TestEmail()`

## Version History

- **v1.7.0+**: Credentials loaded from environment variables by default
- **v1.6.x**: Credentials could be in alarm JSON or environment
- **Earlier**: Only supported credentials in alarm JSON
