# Alarm Package

The alarm package provides rule-based weather alerting with multiple notification channels for the Tempest HomeKit Bridge.

## Overview

The alarm system monitors weather observations and triggers notifications when configured conditions are met. It supports flexible condition syntax, template-based messages, and multiple notification channels.

## Components

### Types (`types.go`)
- **AlarmConfig**: Complete alarm configuration with global settings and alarm definitions
- **Alarm**: Individual alarm rule with condition, channels, tags, and cooldown
- **Channel**: Notification channel configuration (console, syslog, oslog, email, SMS, eventlog)
- **EmailGlobalConfig**: Global email settings (SMTP, Microsoft 365)
- **SMSGlobalConfig**: Global SMS settings (Twilio, AWS SNS)
- **SyslogConfig**: Syslog configuration

### Evaluator (`evaluator.go`)
Parses and evaluates alarm conditions against weather observations.

**Supported operators:**
- Comparison: `>`, `<`, `>=`, `<=`, `==`, `!=`
- Logical: `&&` (and), `||` (or)

**Supported fields:**
- `temperature`, `temp`: Air temperature (°C)
- `humidity`: Relative humidity (%)
- `pressure`: Station pressure (inHg or mb)
- `wind_speed`, `wind`: Wind speed (m/s)
- `wind_gust`: Wind gust (m/s)
- `lux`, `light`: Illuminance (lux)
- `uv`, `uv_index`: UV index
- `rain_rate`, `rain_accumulated`: Rain rate/accumulation
- `lightning_count`: Lightning strike count
- `lightning_distance`: Lightning distance (miles)

**Example conditions:**
```
temperature > 85
humidity > 80 && temperature > 35
lux > 10000 && lux < 50000
lightning_distance < 2
rain_rate > 0
```

### Notifiers (`notifiers.go`)
Implements notification channels with template expansion.

**Available channels:**
- **Console**: Logs to stdout via logger (always visible regardless of log level)
- **Syslog**: Local or remote syslog
- **OSLog**: macOS unified logging system (os_log API via CGO, macOS only)
- **EventLog**: System event log (Windows) or syslog (Unix)
- **Email**: SMTP (with TLS support) or **Microsoft 365 OAuth2** - **SMS**: Twilio or AWS SNS (placeholder implementation)

**Template variables:**
- `{{temperature}}`, `{{temperature_f}}`, `{{temperature_c}}`
- `{{humidity}}`, `{{pressure}}`, `{{wind_speed}}`, `{{wind_gust}}`
- `{{lux}}`, `{{uv}}`, `{{rain_rate}}`, `{{rain_daily}}`
- `{{lightning_count}}`, `{{lightning_distance}}`
- `{{timestamp}}`, `{{station}}`, `{{alarm_name}}`

### Manager (`manager.go`)
Orchestrates alarm evaluation and notification delivery.

**Features:**
- Loads configuration from file or inline JSON
- Cross-platform file watching (macOS, Windows, Linux)
- Automatic configuration reloading on file changes
- Per-alarm cooldown management
- Thread-safe configuration access

## Usage

### Basic Configuration

#### SMTP Email
```json
{
 "email": {
 "provider": "smtp",
 "smtp_host": "smtp.example.com",
 "smtp_port": 587,
 "username": "alerts@example.com",
 "password": "${SMTP_PASSWORD}",
 "from_address": "alerts@example.com",
 "use_tls": true
 },
 "alarms": [
 {
 "name": "high-temperature",
 "description": "Alert when temperature exceeds 85°F",
 "tags": ["temperature", "heat"],
 "enabled": true,
 "condition": "temperature > 85",
 "cooldown": 1800,
 "channels": [
 {
 "type": "console",
 "template": "Tempest-Alarm [high-temperature]: Temperature is {{temperature}}°F at {{timestamp}}"
 },
 {
 "type": "email",
 "email": {
 "to": ["admin@example.com"],
 "subject": "High Temperature Alert: {{temperature}}°F",
 "body": "Temperature: {{temperature}}°F\\nTime: {{timestamp}}"
 }
 }
 ]
 }
 ]
}
```

### Programmatic Usage

```go
import "tempest-homekit-go/pkg/alarm"

// Initialize manager
manager, err := alarm.NewManager("@alarms.json", "Station Name")
if err != nil {
 log.Fatal(err)
}
defer manager.Stop()

// Process observations
for obs := range observationChannel {
 manager.ProcessObservation(&obs)
}
```

#### Microsoft 365 Email (OAuth2)
```json
{
 "email": {
 "provider": "microsoft365",
 "use_oauth2": true,
 "client_id": "${MS365_CLIENT_ID}",
 "client_secret": "${MS365_CLIENT_SECRET}",
 "tenant_id": "${MS365_TENANT_ID}",
 "from_address": "alerts@yourdomain.com",
 "from_name": "Weather Alerts"
 },
 "alarms": [...]
}
```

See [EMAIL_O365_IMPLEMENTATION.md](docs/EMAIL_O365_IMPLEMENTATION.md) for complete setup instructions.

## Environment Variables

Configure notification providers via environment variables:

```bash
# SMTP Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=alerts@example.com
SMTP_PASSWORD=your-password
SMTP_USE_TLS=true

# Microsoft 365 Email (OAuth2)
MS365_CLIENT_ID=your-client-id
MS365_CLIENT_SECRET=your-client-secret
MS365_TENANT_ID=your-tenant-id

# Twilio SMS
TWILIO_ACCOUNT_SID=your-account-sid
TWILIO_AUTH_TOKEN=your-auth-token
TWILIO_FROM_NUMBER=+15555551234

# AWS SNS
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
AWS_SNS_TOPIC_ARN=arn:aws:sns:us-east-1:123456789012:topic
```

## Testing

### Unit Tests

```bash
# Run all alarm tests
go test ./pkg/alarm/...

# Run with coverage
go test -cover ./pkg/alarm/...

# Run with verbose output
go test -v ./pkg/alarm/...

# Run specific test
go test -run TestEvaluator ./pkg/alarm/...
```

### Notification Testing

Test your notification configurations before deploying:

#### Email Testing
```bash
./tempest-homekit-go --test-email user@example.com --alarms @alarms.json
```

#### SMS Testing
```bash
./tempest-homekit-go --test-sms +15555551234 --alarms @alarms.json
```

#### Console Testing
```bash
./tempest-homekit-go --test-console --alarms @alarms.json
```

#### Syslog Testing
```bash
./tempest-homekit-go --test-syslog --alarms @alarms.json
```

#### OSLog Testing (macOS)
```bash
./tempest-homekit-go --test-oslog --alarms @alarms.json
```

#### Event Log Testing (Windows)
```bash
./tempest-homekit-go --test-eventlog --alarms @alarms.json
```

#### Alarm Trigger Testing
```bash
./tempest-homekit-go --test-alarm "high-temperature" --alarms @alarms.json --station "Test"
```

**What it tests:**
- Provider detection (Microsoft 365 OAuth2 vs SMTP)
- Credential validation from environment variables
- Email delivery with test message
- Template expansion with real weather data

**Test email includes:**
- Application name and version (v1.7.0)
- Timestamp with timezone
- Command line options used
- Email provider and configuration
- Current weather observation data

**Provider Priority:**
1. Microsoft 365 OAuth2 (if `MS365_CLIENT_ID`, `MS365_CLIENT_SECRET`, `MS365_TENANT_ID` set)
2. SMTP (if `SMTP_HOST` set)
3. Error if neither configured

**Microsoft 365 Requirements:**
- `MS365_CLIENT_ID` - Azure AD application client ID
- `MS365_CLIENT_SECRET` - Azure AD client secret (value, not ID)
- `MS365_TENANT_ID` - Azure AD tenant ID
- `MS365_FROM_ADDRESS` - Sender email address
- Mail.Send API permission with admin consent

**SMTP Requirements:**
- `SMTP_HOST` - SMTP server hostname
- `SMTP_PORT` - SMTP port (usually 587)
- `SMTP_USERNAME` - SMTP authentication username
- `SMTP_PASSWORD` - SMTP authentication password
- `SMTP_FROM_ADDRESS` or `SMTP_USERNAME` - Sender email
- `SMTP_USE_TLS=true` - Enable TLS encryption

See `.env.example` for complete setup instructions with provider-specific examples.

## Future Enhancements

- **Alarm Editor**: Interactive web UI for alarm management (--alarms-edit mode)
- **SMS Integration**: Full go-sms-sender implementation for Twilio/AWS SNS
- **Generic SMTP**: Complete generic SMTP implementation with enhanced TLS support
- **Advanced Conditions**: Support for time windows, rate limiting, aggregations
- **Notification History**: Track and display notification history in web UI
- **Alarm Templates**: Pre-built alarm configurations for common scenarios

## See Also

- `/examples/alarms.example.json`: Complete example configuration
- `/.env.example`: Environment variable template
- [EMAIL_O365_IMPLEMENTATION.md](docs/EMAIL_O365_IMPLEMENTATION.md): Microsoft 365 setup guide
- [alarms-o365-email.json](docs/examples/alarms-o365-email.json): Example O365 alarm configuration
- Main documentation: `/README.md#alarms-and-notifications`
