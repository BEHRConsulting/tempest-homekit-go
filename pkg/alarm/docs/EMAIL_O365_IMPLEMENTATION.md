# Microsoft 365 / O365 Exchange Email Implementation

## Overview

The alarm system now supports sending email notifications through Microsoft 365 / Office 365 Exchange using the Microsoft Graph API with OAuth2 authentication. This provides a more secure and modern alternative to SMTP for organizations using Microsoft 365.

## Features

- **OAuth2 Authentication**: Uses Azure AD application credentials (Client ID, Client Secret, Tenant ID)
- **Microsoft Graph API**: Leverages the official Microsoft Graph SDK for Go
- **Full Email Support**: Including To, CC, BCC recipients
- **From Address Customization**: Supports custom display names
- **Automatic Fallback**: Falls back to SMTP if OAuth2 is not configured

## Configuration

### Azure AD App Registration

1. **Register an Application in Azure Portal**:
   - Navigate to Azure Active Directory → App registrations → New registration
   - Name: "Tempest Weather Alerts" (or your preference)
   - Supported account types: Accounts in this organizational directory only
   - Redirect URI: Not required for this use case

2. **Configure API Permissions**:
   - Go to API permissions → Add a permission → Microsoft Graph → Application permissions
   - Add: `Mail.Send` permission
   - Click "Grant admin consent" for your organization

3. **Create Client Secret**:
   - Go to Certificates & secrets → New client secret
   - Description: "Tempest Email Notifier"
   - Expiration: Choose appropriate duration
   - **Save the secret value** - you won't be able to see it again!

4. **Get Required IDs**:
   - **Client ID**: Found on the app's Overview page
   - **Tenant ID**: Found on the app's Overview page
   - **Client Secret**: The value you saved in step 3

### Environment Variables

Add these to your `.env` file:

```bash
# Microsoft 365 OAuth2 Configuration
MS365_CLIENT_ID=your-client-id-here
MS365_CLIENT_SECRET=your-client-secret-here
MS365_TENANT_ID=your-tenant-id-here
```

The application will expand environment variables like `${MS365_CLIENT_ID}` in the alarm configuration.

### Alarm Configuration

In your alarm JSON configuration file, set up the global email configuration:

```json
{
  "email": {
    "provider": "microsoft365",
    "use_oauth2": true,
    "client_id": "${MS365_CLIENT_ID}",
    "client_secret": "${MS365_CLIENT_SECRET}",
    "tenant_id": "${MS365_TENANT_ID}",
    "from_address": "alerts@yourdomain.com",
    "from_name": "Tempest Weather Alerts"
  },
  "alarms": [
    {
      "name": "High Temperature Alert",
      "enabled": true,
      "condition": "temperature > 85",
      "channels": [
        {
          "type": "email",
          "email": {
            "to": ["admin@yourdomain.com"],
            "cc": ["team@yourdomain.com"],
            "subject": "⚠️ High Temperature: {{temperature_f}}°F",
            "body": "Temperature has exceeded threshold.\n\nCurrent: {{temperature_f}}°F\nStation: {{station}}\nTime: {{timestamp}}"
          }
        }
      ]
    }
  ]
}
```

## Supported Providers

The `provider` field supports the following values for Microsoft 365:
- `"microsoft365"` - Recommended
- `"o365"` - Alias
- `"exchange"` - Alias

## SMTP Fallback

If `use_oauth2` is set to `false` or OAuth2 credentials are missing, the system will automatically fall back to SMTP authentication. This allows gradual migration or testing:

```json
{
  "email": {
    "provider": "microsoft365",
    "use_oauth2": false,
    "smtp_host": "smtp.office365.com",
    "smtp_port": 587,
    "username": "${SMTP_USERNAME}",
    "password": "${SMTP_PASSWORD}",
    "from_address": "alerts@yourdomain.com",
    "use_tls": true
  }
}
```

## Template Variables

Email subjects and bodies support template variables:

### Current Values
- `{{temperature}}` - Temperature in Celsius
- `{{temperature_f}}` - Temperature in Fahrenheit
- `{{temperature_c}}` - Temperature in Celsius (explicit)
- `{{humidity}}` - Relative humidity percentage
- `{{pressure}}` - Station pressure
- `{{wind_speed}}` - Average wind speed
- `{{wind_gust}}` - Wind gust speed
- `{{wind_direction}}` - Wind direction in degrees
- `{{lux}}` - Illuminance in lux
- `{{uv}}` - UV index
- `{{rain_rate}}` - Rain rate
- `{{rain_daily}}` - Daily rain accumulation
- `{{lightning_count}}` - Lightning strike count
- `{{lightning_distance}}` - Average lightning distance
- `{{timestamp}}` - Observation timestamp
- `{{station}}` - Station name
- `{{alarm_name}}` - Name of the triggered alarm
- `{{alarm_description}}` - Description of the alarm

### Previous Values (for change detection)
- `{{last_temperature}}` - Previous temperature
- `{{last_humidity}}` - Previous humidity
- `{{last_pressure}}` - Previous pressure
- `{{last_wind_speed}}` - Previous wind speed
- `{{last_wind_gust}}` - Previous wind gust
- `{{last_wind_direction}}` - Previous wind direction
- `{{last_lux}}` - Previous lux value
- `{{last_uv}}` - Previous UV index
- `{{last_rain_rate}}` - Previous rain rate
- `{{last_rain_daily}}` - Previous daily rain
- `{{last_lightning_count}}` - Previous lightning count
- `{{last_lightning_distance}}` - Previous lightning distance

## Testing

1. **Test Configuration**:
   ```bash
   ./tempest-homekit-go --alarms @alarms.json --loglevel debug
   ```

2. **Check Logs**: Look for these messages:
   ```
   DEBUG: Sending email via Microsoft 365 Graph API
   DEBUG:   Tenant ID: your-tenant-id
   DEBUG:   Client ID: your-client-id
   DEBUG:   From: alerts@yourdomain.com
   DEBUG:   To: [admin@yourdomain.com]
   INFO: Email sent successfully via Microsoft 365 to [admin@yourdomain.com]
   ```

3. **Common Issues**:
   - **Missing Permissions**: Ensure `Mail.Send` application permission is granted
   - **Wrong Tenant**: Verify the tenant ID matches your organization
   - **Expired Secret**: Client secrets have expiration dates
   - **Invalid From Address**: The from address must be a valid mailbox in your organization

## Security Considerations

1. **Client Secret Protection**:
   - Store client secrets in `.env` file
   - Add `.env` to `.gitignore`
   - Never commit secrets to version control
   - Rotate secrets periodically

2. **Least Privilege**:
   - Only grant `Mail.Send` permission
   - Use application-specific credentials, not user accounts
   - Monitor app usage in Azure AD logs

3. **Audit Trail**:
   - All emails are logged with INFO level
   - Failed sends are logged with ERROR level
   - Azure AD tracks all Graph API calls

## Troubleshooting

### Error: "Microsoft 365 OAuth2 credentials missing"
- Solution: Ensure `MS365_CLIENT_ID`, `MS365_CLIENT_SECRET`, and `MS365_TENANT_ID` are set in `.env`

### Error: "failed to create Azure credentials"
- Solution: Verify tenant ID, client ID, and client secret are correct
- Check: Azure Portal → App registrations → Your app → Overview

### Error: "failed to send email via Microsoft Graph API"
- Common causes:
  - Missing or insufficient API permissions
  - Admin consent not granted
  - Invalid from address (not a valid mailbox)
  - Client secret expired

### Email Not Received
- Check spam/junk folders
- Verify from address is a real mailbox in your organization
- Check recipient addresses are correct
- Review Azure AD sign-in logs for the app

## Dependencies

This feature requires the following Go modules:
- `github.com/Azure/azure-sdk-for-go/sdk/azidentity` - Azure authentication
- `github.com/microsoftgraph/msgraph-sdk-go` - Microsoft Graph SDK
- `golang.org/x/oauth2` - OAuth2 support

These are automatically installed with `go mod tidy`.

## Migration from SMTP

To migrate from SMTP to OAuth2:

1. Keep existing SMTP configuration as fallback
2. Add OAuth2 configuration
3. Set `use_oauth2: true`
4. Test with a single alarm
5. Monitor for any errors
6. Once stable, remove SMTP configuration

Example migration config:
```json
{
  "email": {
    "provider": "microsoft365",
    "use_oauth2": true,
    "client_id": "${MS365_CLIENT_ID}",
    "client_secret": "${MS365_CLIENT_SECRET}",
    "tenant_id": "${MS365_TENANT_ID}",
    "smtp_host": "smtp.office365.com",
    "smtp_port": 587,
    "username": "${SMTP_USERNAME}",
    "password": "${SMTP_PASSWORD}",
    "from_address": "alerts@yourdomain.com",
    "use_tls": true
  }
}
```

## References

- [Microsoft Graph API - Send Mail](https://learn.microsoft.com/en-us/graph/api/user-sendmail)
- [Azure AD App Registration](https://learn.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app)
- [Graph SDK for Go](https://github.com/microsoftgraph/msgraph-sdk-go)
