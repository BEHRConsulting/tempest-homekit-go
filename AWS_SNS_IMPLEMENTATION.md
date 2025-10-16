# AWS SNS SMS Notification Implementation

**Version**: 1.8.0  
**Status**: ✅ **COMPLETE** - Production ready  
**Date**: October 15, 2025

## Overview

Complete implementation of AWS SNS SMS notification delivery for the alarm system. This provides production-ready SMS alerting capabilities using Amazon Simple Notification Service with both direct SMS and topic-based broadcasting support.

## Implementation Summary

### What Was Implemented

1. **AWS SDK v2 Integration**
   - Installed `github.com/aws/aws-sdk-go-v2` v1.39.2
   - Installed `github.com/aws/aws-sdk-go-v2/config` v1.31.12
   - Installed `github.com/aws/aws-sdk-go-v2/service/sns` v1.38.5
   - Installed `github.com/aws/aws-sdk-go-v2/credentials` v1.18.16

2. **SMS Notifier Implementation** (`pkg/alarm/notifiers.go`)
   - `sendAWSSNS()` method for AWS SNS integration
   - Support for direct SMS to phone numbers
   - Support for SNS topic publishing (broadcast to subscribers)
   - Environment variable expansion for credentials
   - Comprehensive error handling and logging
   - Multi-recipient support with success tracking

3. **Configuration Architecture**
   - Credentials stored in `.env` file only (not in alarm JSON)
   - Environment-first design with precedence over JSON configs
   - Separate IAM user for application runtime (least privilege)
   - Clean separation: alarm rules in JSON, credentials in `.env`

4. **Setup Automation** (`scripts/setup-aws-sns.sh`)
   - Interactive bash script for production configuration
   - Uses admin AWS CLI credentials from `~/.aws/credentials`
   - Configures SMS type (Transactional/Promotional)
   - Sets spending limits
   - Creates SNS topics with subscriptions
   - Sends test SMS
   - Updates `.env` automatically with Topic ARN
   - Color-coded output for UX
   - Comprehensive error handling

5. **Documentation**
   - Detailed setup instructions in `.env` and `.env.example`
   - Clarified difference between admin and runtime credentials
   - Step-by-step IAM user creation guide
   - Production considerations (sandbox, spending limits, regional capabilities)
   - README.md updated with AWS SNS features

6. **Unit Tests** (`pkg/alarm/notifiers_sms_test.go`)
   - AWS SNS configuration validation tests
   - Template expansion verification
   - Topic ARN configuration tests
   - Multiple recipient handling tests
   - Factory creation tests
   - Missing credentials error handling tests

## Architecture

### Two-Tier Credential System

**Admin Credentials (Setup Time)**:
- Used by `scripts/setup-aws-sns.sh`
- From `~/.aws/credentials` or `aws configure`
- Has full SNS permissions for topic creation, configuration
- Only needed during initial setup

**Application Runtime Credentials (Stored in .env)**:
- Used by running application to send SMS
- Separate IAM user with minimal permissions (`sns:Publish` only)
- Never used for setup operations
- Follows principle of least privilege

### Data Flow

```
Alarm Trigger → SMSNotifier.Send()
              → sendAWSSNS()
              → Load credentials from .env (environment vars)
              → Create AWS config with static credentials
              → Create SNS client
              → For each recipient:
                  If Topic ARN present:
                    → Publish to SNS topic
                  Else:
                    → Send direct SMS to phone number
              → Track success/failure
              → Return error if all sends failed
```

### Configuration Flow

```
.env file → Environment variables → types.go LoadConfigFromEnv()
                                  → SMSGlobalConfig struct
                                  → SMSNotifier
                                  → sendAWSSNS() method
                                  → AWS SDK credentials provider
```

## Usage

### 1. Create Application IAM User

```bash
# In AWS Console:
# 1. IAM > Users > Create User
# 2. User name: tempest-homekit-sns-sender
# 3. Attach custom policy with ONLY sns:Publish:
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["sns:Publish"],
    "Resource": "*"
  }]
}
# 4. Create access key > Application running outside AWS
# 5. Save credentials to .env file
```

### 2. Configure .env File

```bash
# Application runtime credentials (limited permissions)
AWS_ACCESS_KEY_ID=AKIAXXXXXXXXXXXXXXXX
AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
AWS_REGION=us-west-2

# Optional: Topic ARN for broadcasting
AWS_SNS_TOPIC_ARN=arn:aws:sns:us-west-2:123456789012:WeatherAlert
```

### 3. Run Setup Script (Optional)

```bash
# Uses your admin AWS CLI credentials to configure production settings
./scripts/setup-aws-sns.sh

# Script will:
# - Verify AWS CLI access
# - Configure SMS type and spending limits
# - Create SNS topics with subscriptions (optional)
# - Update .env with Topic ARN
# - Send test SMS
```

### 4. Add SMS Channel to Alarms

```json
{
  "alarms": [
    {
      "name": "high-temperature",
      "condition": "temperature > 95",
      "channels": [
        {
          "type": "sms",
          "sms": {
            "to": ["+15555551234", "+15555555678"],
            "message": "⚠️ {{alarm_name}}: {{temperature}}°F at {{station}}"
          }
        }
      ]
    }
  ]
}
```

### 5. Run Application

```bash
./tempest-homekit-go --alarms @tempest-alarms.json
```

## Features

### Direct SMS Mode
- Send SMS directly to phone numbers
- No topic creation required
- Simple configuration
- Good for single recipient

### Topic Broadcasting Mode  
- Create SNS topic with multiple subscribers
- Publish once, deliver to all subscribers
- Easier subscriber management
- Support for SMS, Email, HTTP, etc. subscribers
- Good for multiple recipients

### Template Support
All standard alarm template variables supported:
- `{{alarm_name}}`, `{{alarm_description}}`, `{{alarm_condition}}`
- `{{station}}`, `{{timestamp}}`
- `{{temperature}}`, `{{temperature_f}}`, `{{humidity}}`
- `{{wind_speed}}`, `{{wind_gust}}`, `{{wind_direction}}`
- `{{pressure}}`, `{{uv}}`, `{{lux}}`
- `{{rain_rate}}`, `{{rain_daily}}`
- `{{lightning_count}}`, `{{lightning_distance}}`
- `{{last_temperature}}`, `{{last_humidity}}`, etc.
- `{{sensor_info}}`, `{{alarm_info}}`, `{{app_info}}`

### Error Handling
- Missing credentials → Clear error message
- AWS API errors → Logged with details
- Partial send failures → Track success count
- All sends fail → Return error to alarm system

## Testing

### Run Unit Tests

```bash
# Run all SMS tests
go test ./pkg/alarm/... -run TestSMS

# Run specific test
go test ./pkg/alarm/... -run TestSMSNotifierWithTopicARN

# Run with coverage
go test -cover ./pkg/alarm/...
```

### Test SMS Delivery

```bash
# Using the setup script's test function
./scripts/setup-aws-sns.sh
# Follow prompts to send test SMS

# Or manually test from AWS Console:
# SNS > Topics > Select topic > Publish message
```

## Security Considerations

### Credential Management
- ✅ Never commit `.env` file to git (added to `.gitignore`)
- ✅ Use separate IAM user for application (not personal account)
- ✅ Minimal permissions: `sns:Publish` only
- ✅ Rotate access keys regularly
- ✅ Monitor CloudTrail logs for API usage

### .gitignore Protection
```
.env
.env.*
.env.bak
.env.backup
.env.old
*.env.bak
*.env.backup
*.env.old
!.env.example  # Allow template file
```

### AWS IAM Best Practices
- Create dedicated IAM user per application
- Use IAM policies with least privilege
- Enable MFA for admin accounts (used for setup)
- Use AWS Organizations for multi-account environments
- Regular access key rotation (90 days recommended)

## Production Considerations

### AWS SNS Sandbox Mode
- New accounts start in sandbox mode
- Can only send to verified phone numbers
- Request production access through AWS Support case
- Production mode unlocks unrestricted sending

### Spending Limits
- Default: $1.00/month
- Increase through SNS Console or support case
- Monitor usage in CloudWatch
- Set up billing alerts

### Regional Capabilities
- Not all regions support SMS
- Check SMS pricing by region
- Consider closest region to recipients
- Some countries have restrictions

### SMS Type Selection
- **Transactional**: Time-sensitive, critical messages (higher priority)
- **Promotional**: Marketing, bulk messages (lower cost)
- Choose based on use case (weather alerts = Transactional)

### Cost Optimization
- Use SNS topics for multiple recipients (publish once)
- Direct SMS charges per message per recipient
- Monitor usage with CloudWatch metrics
- Set spending limits to prevent overages

## Troubleshooting

### "AWS SNS credentials missing"
- Check `.env` file exists
- Verify `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` are set
- Check environment variable expansion (no typos in var names)

### "Failed to load AWS config"
- Invalid credentials format
- Network connectivity issues
- AWS region not available

### "Failed to send SMS"
- Phone number format (must start with + and country code)
- Sandbox mode (verify phone number first)
- Spending limit reached
- Invalid Topic ARN
- IAM permission issues

### Check AWS CloudWatch Logs
```bash
# View SNS publish logs
aws logs filter-log-events \
  --log-group-name /aws/sns/us-west-2/123456789012/WeatherAlert \
  --start-time $(date -u -d '1 hour ago' +%s)000
```

## Files Modified

### Core Implementation
- `pkg/alarm/notifiers.go` (~910 lines)
  - Added `sendAWSSNS()` method
  - Added AWS SDK imports
  - Updated `SMSNotifier.Send()` routing

- `pkg/alarm/types.go` (~416 lines)
  - `SMSGlobalConfig` already had AWS fields
  - `LoadConfigFromEnv()` loads AWS config
  - `LoadAlarmConfig()` merges environment config

### Testing
- `pkg/alarm/notifiers_sms_test.go` (NEW, ~250 lines)
  - AWS SNS configuration validation
  - Template expansion tests
  - Topic ARN tests
  - Multiple recipient tests
  - Factory creation tests

### Configuration
- `.env` - Updated with comprehensive AWS SNS docs
- `.env.example` - Updated with comprehensive AWS SNS docs
- `tempest-alarms.json` - Removed provider config (now in .env only)
- `alarms-aws.example.json` - Removed provider config, added clarification

### Scripts
- `scripts/setup-aws-sns.sh` (NEW, ~450 lines)
  - Interactive production setup
  - AWS CLI integration
  - Topic creation
  - Test SMS sending
  - Automatic .env updates

### Documentation
- `README.md` - Added AWS SNS section with quick start
- `REQUIREMENTS.md` - Updated version to 1.8.0, added AWS SNS vars
- `.gitignore` - Enhanced to protect all .env variants

### Dependencies
- `go.mod` - Added AWS SDK v2 packages

## Next Steps

### Twilio Implementation (Coming Next)
- Similar architecture to AWS SNS
- Credentials in `.env` file
- `sendTwilio()` method already stubbed
- Will use Twilio REST API
- Unit tests following same pattern

### Future Enhancements
- SMS message templates with rich formatting
- Delivery confirmation tracking
- SMS rate limiting per alarm
- Cost tracking integration
- Multi-provider failover (try Twilio if AWS fails)

## Related Documentation

- `.env.example` - Comprehensive setup instructions
- `README.md` - User-facing AWS SNS documentation
- `REQUIREMENTS.md` - Technical requirements and version history
- `scripts/setup-aws-sns.sh` - Automated setup script with inline docs
- `pkg/alarm/README.md` - Alarm system architecture

## Completion Checklist

- ✅ AWS SDK v2 packages installed
- ✅ `sendAWSSNS()` method implemented
- ✅ Direct SMS support working
- ✅ Topic publishing support working
- ✅ Configuration architecture clean (.env only)
- ✅ Setup script created and tested
- ✅ Unit tests written and passing
- ✅ Documentation complete (.env, README.md, REQUIREMENTS.md)
- ✅ .gitignore updated for .env protection
- ✅ Version bumped to 1.8.0
- ⏳ Twilio implementation (next)
- ⏳ Documentation audit (after 3-4 major changes)

---

**Status**: Production Ready  
**Version**: 1.8.0  
**Date**: October 15, 2025
