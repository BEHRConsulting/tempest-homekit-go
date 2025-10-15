package alarm

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/syslog"
	"net/smtp"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/weather"
)

var (
	// appStartTime tracks when the application started
	appStartTime = time.Now()
)

// Notifier interface for sending notifications
type Notifier interface {
	Send(alarm *Alarm, channel *Channel, obs *weather.Observation, stationName string) error
}

// NotifierFactory creates notifiers for different channel types
type NotifierFactory struct {
	config *AlarmConfig
}

// NewNotifierFactory creates a new notifier factory
func NewNotifierFactory(config *AlarmConfig) *NotifierFactory {
	return &NotifierFactory{config: config}
}

// GetNotifier returns a notifier for the given channel type
func (f *NotifierFactory) GetNotifier(channelType string) (Notifier, error) {
	switch channelType {
	case "console":
		return &ConsoleNotifier{}, nil
	case "syslog":
		return &SyslogNotifier{config: f.config.Syslog}, nil
	case "oslog":
		return &OSLogNotifier{}, nil
	case "eventlog":
		return &EventLogNotifier{}, nil
	case "email":
		return &EmailNotifier{config: f.config.Email}, nil
	case "sms":
		return &SMSNotifier{config: f.config.SMS}, nil
	default:
		return nil, fmt.Errorf("unsupported notifier type: %s", channelType)
	}
}

// ConsoleNotifier sends notifications to console/stdout
type ConsoleNotifier struct{}

func (n *ConsoleNotifier) Send(alarm *Alarm, channel *Channel, obs *weather.Observation, stationName string) error {
	message := expandTemplate(channel.Template, alarm, obs, stationName)
	logger.Alarm("%s", message)
	return nil
}

// SyslogNotifier sends notifications to syslog
type SyslogNotifier struct {
	config *SyslogConfig
}

func (n *SyslogNotifier) Send(alarm *Alarm, channel *Channel, obs *weather.Observation, stationName string) error {
	message := expandTemplate(channel.Template, alarm, obs, stationName)

	var priority syslog.Priority
	if n.config != nil {
		switch strings.ToLower(n.config.Priority) {
		case "error":
			priority = syslog.LOG_ERR
		case "warning":
			priority = syslog.LOG_WARNING
		case "info":
			priority = syslog.LOG_INFO
		default:
			priority = syslog.LOG_WARNING
		}
	} else {
		priority = syslog.LOG_WARNING
	}

	var writer *syslog.Writer
	var err error

	if n.config != nil && n.config.Network != "" && n.config.Address != "" {
		writer, err = syslog.Dial(n.config.Network, n.config.Address, priority, n.config.Tag)
	} else {
		writer, err = syslog.New(priority, "tempest-weather")
	}

	if err != nil {
		return fmt.Errorf("failed to connect to syslog: %w", err)
	}
	defer writer.Close()

	return writer.Warning(message)
}

// EventLogNotifier sends notifications to system event log (Windows) or syslog (Unix)
type EventLogNotifier struct{}

func (n *EventLogNotifier) Send(alarm *Alarm, channel *Channel, obs *weather.Observation, stationName string) error {
	message := expandTemplate(channel.Template, alarm, obs, stationName)

	if runtime.GOOS == "windows" {
		// On Windows, use event log (simplified - would need golang.org/x/sys/windows for full implementation)
		logger.Info("[EventLog] %s", message)
		return nil
	}

	// On Unix systems, fall back to syslog
	writer, err := syslog.New(syslog.LOG_WARNING, "tempest-weather")
	if err != nil {
		return fmt.Errorf("failed to connect to syslog: %w", err)
	}
	defer writer.Close()

	return writer.Warning(message)
}

// EmailNotifier sends email notifications
type EmailNotifier struct {
	config *EmailGlobalConfig
}

func (n *EmailNotifier) Send(alarm *Alarm, channel *Channel, obs *weather.Observation, stationName string) error {
	if channel.Email == nil {
		return fmt.Errorf("email configuration missing for channel")
	}

	if n.config == nil {
		return fmt.Errorf("global email configuration not set")
	}

	// Expand templates - use channel.Template if email.Body is empty
	subject := expandTemplate(channel.Email.Subject, alarm, obs, stationName)
	bodyTemplate := channel.Email.Body
	if bodyTemplate == "" {
		bodyTemplate = channel.Template
	}
	body := expandTemplate(bodyTemplate, alarm, obs, stationName)

	// Prepend recipient information to body for better context
	toList := strings.Join(channel.Email.To, ", ")
	body = fmt.Sprintf("To: %s\n\n%s", toList, body)

	// Build email message
	from := n.config.FromAddress
	if n.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", n.config.FromName, n.config.FromAddress)
	}

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(channel.Email.To, ", ")))
	if len(channel.Email.CC) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(channel.Email.CC, ", ")))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))

	// Set content type based on Html flag
	if channel.Email.Html {
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Prepare recipients list
	recipients := append([]string{}, channel.Email.To...)
	recipients = append(recipients, channel.Email.CC...)
	recipients = append(recipients, channel.Email.BCC...)

	// Send email based on provider
	switch n.config.Provider {
	case "smtp":
		return n.sendSMTP(recipients, []byte(msg.String()))
	case "microsoft365", "o365", "exchange":
		if n.config.UseOAuth2 {
			return n.sendMicrosoft365(channel.Email, subject, body)
		}
		// Fall back to SMTP for M365 without OAuth2
		logger.Info("Microsoft 365 OAuth2 not configured, using SMTP for Exchange")
		return n.sendSMTP(recipients, []byte(msg.String()))
	default:
		return fmt.Errorf("unsupported email provider: %s", n.config.Provider)
	}
}

func (n *EmailNotifier) sendSMTP(to []string, msg []byte) error {
	// Get credentials from config (with environment variable expansion)
	username := os.ExpandEnv(n.config.Username)
	password := os.ExpandEnv(n.config.Password)

	auth := smtp.PlainAuth("", username, password, n.config.SMTPHost)

	addr := fmt.Sprintf("%s:%d", n.config.SMTPHost, n.config.SMTPPort)

	if n.config.UseTLS {
		// Determine if we should use implicit TLS (port 465) or STARTTLS (port 587)
		useImplicitTLS := n.config.SMTPPort == 465

		tlsConfig := &tls.Config{
			ServerName: n.config.SMTPHost,
		}

		if useImplicitTLS {
			// Implicit TLS: Connect with TLS from the start (port 465)
			conn, err := tls.Dial("tcp", addr, tlsConfig)
			if err != nil {
				return fmt.Errorf("failed to dial TLS: %w", err)
			}
			defer conn.Close()

			client, err := smtp.NewClient(conn, n.config.SMTPHost)
			if err != nil {
				return fmt.Errorf("failed to create SMTP client: %w", err)
			}
			defer client.Close()

			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP auth failed: %w", err)
			}

			if err = client.Mail(n.config.FromAddress); err != nil {
				return fmt.Errorf("SMTP MAIL failed: %w", err)
			}

			for _, addr := range to {
				if err = client.Rcpt(addr); err != nil {
					return fmt.Errorf("SMTP RCPT failed for %s: %w", addr, err)
				}
			}

			w, err := client.Data()
			if err != nil {
				return fmt.Errorf("SMTP DATA failed: %w", err)
			}

			_, err = w.Write(msg)
			if err != nil {
				return fmt.Errorf("failed to write message: %w", err)
			}

			err = w.Close()
			if err != nil {
				return fmt.Errorf("failed to close data writer: %w", err)
			}

			return client.Quit()
		} else {
			// STARTTLS: Connect plain, then upgrade to TLS (port 587)
			client, err := smtp.Dial(addr)
			if err != nil {
				return fmt.Errorf("failed to dial SMTP: %w", err)
			}
			defer client.Close()

			// Send STARTTLS command
			if err = client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("STARTTLS failed: %w", err)
			}

			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP auth failed: %w", err)
			}

			if err = client.Mail(n.config.FromAddress); err != nil {
				return fmt.Errorf("SMTP MAIL failed: %w", err)
			}

			for _, addr := range to {
				if err = client.Rcpt(addr); err != nil {
					return fmt.Errorf("SMTP RCPT failed for %s: %w", addr, err)
				}
			}

			w, err := client.Data()
			if err != nil {
				return fmt.Errorf("SMTP DATA failed: %w", err)
			}

			_, err = w.Write(msg)
			if err != nil {
				return fmt.Errorf("failed to write message: %w", err)
			}

			err = w.Close()
			if err != nil {
				return fmt.Errorf("failed to close data writer: %w", err)
			}

			return client.Quit()
		}
	}

	// Non-TLS SMTP
	return smtp.SendMail(addr, auth, n.config.FromAddress, to, msg)
}

func (n *EmailNotifier) sendMicrosoft365(emailConfig *EmailConfig, subject, body string) error {
	// Get credentials from environment (expand environment variables)
	clientID := os.ExpandEnv(n.config.ClientID)
	clientSecret := os.ExpandEnv(n.config.ClientSecret)
	tenantID := os.ExpandEnv(n.config.TenantID)
	fromAddress := os.ExpandEnv(n.config.FromAddress)

	if clientID == "" || clientSecret == "" || tenantID == "" {
		return fmt.Errorf("Microsoft 365 OAuth2 credentials missing (CLIENT_ID, CLIENT_SECRET, TENANT_ID required)")
	}

	if fromAddress == "" {
		return fmt.Errorf("FROM_ADDRESS is required for Microsoft 365 email")
	}

	logger.Debug("Sending email via Microsoft 365 Graph API")
	logger.Debug("  Tenant ID: %s", tenantID)
	logger.Debug("  Client ID: %s", clientID)
	logger.Debug("  From: %s", fromAddress)
	logger.Debug("  To: %v", emailConfig.To)

	// Create client credentials
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return fmt.Errorf("failed to create Azure credentials: %w", err)
	}

	// Create Graph client
	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return fmt.Errorf("failed to create Graph client: %w", err)
	}

	// Build message
	message := models.NewMessage()
	message.SetSubject(&subject)

	// Set body
	bodyContent := models.NewItemBody()
	contentType := models.TEXT_BODYTYPE
	if emailConfig.Html {
		contentType = models.HTML_BODYTYPE
	}
	bodyContent.SetContentType(&contentType)
	bodyContent.SetContent(&body)
	message.SetBody(bodyContent)

	// Set recipients
	toRecipients := make([]models.Recipientable, 0, len(emailConfig.To))
	for _, addr := range emailConfig.To {
		recipient := models.NewRecipient()
		emailAddr := models.NewEmailAddress()
		emailAddr.SetAddress(&addr)
		recipient.SetEmailAddress(emailAddr)
		toRecipients = append(toRecipients, recipient)
	}
	message.SetToRecipients(toRecipients)

	// Set CC if provided
	if len(emailConfig.CC) > 0 {
		ccRecipients := make([]models.Recipientable, 0, len(emailConfig.CC))
		for _, addr := range emailConfig.CC {
			recipient := models.NewRecipient()
			emailAddr := models.NewEmailAddress()
			emailAddr.SetAddress(&addr)
			recipient.SetEmailAddress(emailAddr)
			ccRecipients = append(ccRecipients, recipient)
		}
		message.SetCcRecipients(ccRecipients)
	}

	// Set BCC if provided
	if len(emailConfig.BCC) > 0 {
		bccRecipients := make([]models.Recipientable, 0, len(emailConfig.BCC))
		for _, addr := range emailConfig.BCC {
			recipient := models.NewRecipient()
			emailAddr := models.NewEmailAddress()
			emailAddr.SetAddress(&addr)
			recipient.SetEmailAddress(emailAddr)
			bccRecipients = append(bccRecipients, recipient)
		}
		message.SetBccRecipients(bccRecipients)
	}

	// Set from address
	fromRecipient := models.NewRecipient()
	fromEmailAddr := models.NewEmailAddress()
	fromEmailAddr.SetAddress(&fromAddress)
	if n.config.FromName != "" {
		fromEmailAddr.SetName(&n.config.FromName)
	}
	fromRecipient.SetEmailAddress(fromEmailAddr)
	message.SetFrom(fromRecipient)

	// Create send mail request body
	sendMailBody := users.NewItemSendMailPostRequestBody()
	sendMailBody.SetMessage(message)
	saveToSentItems := false
	sendMailBody.SetSaveToSentItems(&saveToSentItems)

	// Send the email using the from address as the user principal
	// Note: The from address must be a valid user principal name (UPN) in the tenant
	ctx := context.Background()

	// Extract the user principal - ensure it's properly formatted
	userPrincipal := fromAddress
	logger.Debug("  Sending as user principal: %s", userPrincipal)

	err = client.Users().ByUserId(userPrincipal).SendMail().Post(ctx, sendMailBody, nil)
	if err != nil {
		return fmt.Errorf("failed to send email via Microsoft Graph API (user: %s): %w", userPrincipal, err)
	}

	logger.Info("Email sent successfully via Microsoft 365 to %v", emailConfig.To)
	return nil
}

// SMSNotifier sends SMS notifications
type SMSNotifier struct {
	config *SMSGlobalConfig
}

func (n *SMSNotifier) Send(alarm *Alarm, channel *Channel, obs *weather.Observation, stationName string) error {
	if channel.SMS == nil {
		return fmt.Errorf("SMS configuration missing for channel")
	}

	if n.config == nil {
		return fmt.Errorf("global SMS configuration not set")
	}

	message := expandTemplate(channel.SMS.Message, alarm, obs, stationName)

	// Note: go-sms-sender integration would go here
	// For now, log the SMS that would be sent
	logger.Info("SMS Notification [%s provider]: To=%v, Message=%s",
		n.config.Provider, channel.SMS.To, message)

	// TODO: Implement actual SMS sending using go-sms-sender
	// Example for Twilio:
	// client := sms.NewClient(n.config.AccountSID, n.config.AuthToken)
	// for _, to := range channel.SMS.To {
	//     err := client.Send(n.config.FromNumber, to, message)
	//     if err != nil {
	//         return err
	//     }
	// }

	return nil
}

// formatAppInfo returns formatted application information
func formatAppInfo(isHTML bool) string {
	uptime := time.Since(appStartTime)
	days := int(uptime.Hours() / 24)
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60

	uptimeStr := ""
	if days > 0 {
		uptimeStr = fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
	} else if hours > 0 {
		uptimeStr = fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else {
		uptimeStr = fmt.Sprintf("%d minutes", minutes)
	}

	if isHTML {
		return fmt.Sprintf(`<div style="font-size: 11px; color: #666; font-family: monospace;">
			<strong>Tempest HomeKit Bridge</strong> %s | Uptime: %s | Go %s
		</div>`, appVersion, uptimeStr, runtime.Version())
	}

	return fmt.Sprintf("Tempest HomeKit Bridge %s | Uptime: %s | Go %s",
		appVersion, uptimeStr, runtime.Version())
}

// formatAlarmInfo returns formatted alarm information
func formatAlarmInfo(alarm *Alarm, isHTML bool) string {
	enabledStr := "enabled"
	if !alarm.Enabled {
		enabledStr = "disabled"
	}

	cooldownStr := fmt.Sprintf("%d seconds", alarm.Cooldown)
	if alarm.Cooldown >= 3600 {
		cooldownStr = fmt.Sprintf("%d hours", alarm.Cooldown/3600)
	} else if alarm.Cooldown >= 60 {
		cooldownStr = fmt.Sprintf("%d minutes", alarm.Cooldown/60)
	}

	tagsStr := "none"
	if len(alarm.Tags) > 0 {
		tagsStr = strings.Join(alarm.Tags, ", ")
	}

	if isHTML {
		return fmt.Sprintf(`<table style="border-collapse: collapse; width: 100%%;">
			<tr><td style="padding: 5px; border: 1px solid #ddd; font-weight: bold;">Alarm:</td><td style="padding: 5px; border: 1px solid #ddd;">%s</td></tr>
			<tr><td style="padding: 5px; border: 1px solid #ddd; font-weight: bold;">Description:</td><td style="padding: 5px; border: 1px solid #ddd;">%s</td></tr>
			<tr><td style="padding: 5px; border: 1px solid #ddd; font-weight: bold;">Condition:</td><td style="padding: 5px; border: 1px solid #ddd;">%s</td></tr>
			<tr><td style="padding: 5px; border: 1px solid #ddd; font-weight: bold;">Status:</td><td style="padding: 5px; border: 1px solid #ddd;">%s</td></tr>
			<tr><td style="padding: 5px; border: 1px solid #ddd; font-weight: bold;">Cooldown:</td><td style="padding: 5px; border: 1px solid #ddd;">%s</td></tr>
			<tr><td style="padding: 5px; border: 1px solid #ddd; font-weight: bold;">Tags:</td><td style="padding: 5px; border: 1px solid #ddd;">%s</td></tr>
		</table>`,
			alarm.Name, alarm.Description, alarm.Condition, enabledStr, cooldownStr, tagsStr)
	}

	return fmt.Sprintf("Alarm: %s\nDescription: %s\nCondition: %s\nStatus: %s\nCooldown: %s\nTags: %s",
		alarm.Name, alarm.Description, alarm.Condition, enabledStr, cooldownStr, tagsStr)
}

// formatSensorInfo returns formatted sensor information
func formatSensorInfo(obs *weather.Observation, isHTML bool) string {
	return formatSensorInfoWithAlarm(obs, nil, isHTML)
}

func formatSensorInfoWithAlarm(obs *weather.Observation, alarm *Alarm, isHTML bool) string {
	tempF := obs.AirTemperature*9/5 + 32
	windSpeedMph := obs.WindAvg * 2.23694
	windGustMph := obs.WindGust * 2.23694
	rainDaily := obs.RainDailyTotal / 25.4 // Convert mm to inches

	// Wind direction cardinal
	dir := obs.WindDirection
	cardinal := "N"
	switch {
	case dir >= 337.5 || dir < 22.5:
		cardinal = "N"
	case dir >= 22.5 && dir < 67.5:
		cardinal = "NE"
	case dir >= 67.5 && dir < 112.5:
		cardinal = "E"
	case dir >= 112.5 && dir < 157.5:
		cardinal = "SE"
	case dir >= 157.5 && dir < 202.5:
		cardinal = "S"
	case dir >= 202.5 && dir < 247.5:
		cardinal = "SW"
	case dir >= 247.5 && dir < 292.5:
		cardinal = "W"
	case dir >= 292.5 && dir < 337.5:
		cardinal = "NW"
	}

	// Helper to check if value changed
	hasChanged := func(key string, current float64, threshold float64) bool {
		if alarm == nil {
			return false
		}
		var prev float64
		var ok bool
		if prev, ok = alarm.GetTriggerValue(key); !ok {
			if prev, ok = alarm.GetPreviousValue(key); !ok {
				return false
			}
		}
		// Check if difference exceeds threshold
		diff := current - prev
		if diff < 0 {
			diff = -diff
		}
		return diff > threshold
	}

	// Helper to get previous value with proper formatting
	getPrevValue := func(key string, current float64, format string) string {
		if alarm == nil {
			return "N/A"
		}
		if prev, ok := alarm.GetTriggerValue(key); ok {
			return fmt.Sprintf(format, prev)
		}
		if prev, ok := alarm.GetPreviousValue(key); ok {
			return fmt.Sprintf(format, prev)
		}
		return "N/A"
	}
	
	// Special handler for illuminance which needs number formatting
	getPrevLux := func() string {
		if alarm == nil {
			return "N/A"
		}
		if prev, ok := alarm.GetTriggerValue("lux"); ok {
			return formatNumber(prev)
		}
		if prev, ok := alarm.GetPreviousValue("lux"); ok {
			return formatNumber(prev)
		}
		return "N/A"
	}
	
	// Helper to get row style based on whether value changed
	getRowStyle := func(changed bool) string {
		if changed {
			return ` style="background: #fff3cd;"`
		}
		return ""
	}

	if isHTML {
		return fmt.Sprintf(`<table style="border-collapse: collapse; width: 100%%;">
			<tr style="background: #f0f0f0;"><th style="padding: 5px; border: 1px solid #ddd;">Sensor</th><th style="padding: 5px; border: 1px solid #ddd;">Current</th><th style="padding: 5px; border: 1px solid #ddd;">Last</th></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Temperature:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%.1f°F (%.1f°C)</td><td style="padding: 5px; border: 1px solid #ddd;">%s°C</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Humidity:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%.0f%%</td><td style="padding: 5px; border: 1px solid #ddd;">%s%%</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Pressure:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%.2f mb</td><td style="padding: 5px; border: 1px solid #ddd;">%s mb</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Wind Speed:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%.1f mph (%.1f m/s)</td><td style="padding: 5px; border: 1px solid #ddd;">%s m/s</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Wind Gust:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%.1f mph (%.1f m/s)</td><td style="padding: 5px; border: 1px solid #ddd;">%s m/s</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Wind Direction:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%.0f° (%s)</td><td style="padding: 5px; border: 1px solid #ddd;">%s°</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>UV Index:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%d</td><td style="padding: 5px; border: 1px solid #ddd;">%s</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Illuminance:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%s lux</td><td style="padding: 5px; border: 1px solid #ddd;">%s lux</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Rain Rate:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%.2f mm/hr</td><td style="padding: 5px; border: 1px solid #ddd;">%s mm/hr</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Daily Rain:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%.2f in (%.1f mm)</td><td style="padding: 5px; border: 1px solid #ddd;">%s mm</td></tr>
			<tr%s><td style="padding: 5px; border: 1px solid #ddd;"><strong>Lightning:</strong></td><td style="padding: 5px; border: 1px solid #ddd;">%d strikes</td><td style="padding: 5px; border: 1px solid #ddd;">%s strikes</td></tr>
		</table>`,
			getRowStyle(hasChanged("temperature", obs.AirTemperature, 0.1)),
			tempF, obs.AirTemperature, getPrevValue("temperature", obs.AirTemperature, "%.1f"),
			getRowStyle(hasChanged("humidity", obs.RelativeHumidity, 1.0)),
			obs.RelativeHumidity, getPrevValue("humidity", obs.RelativeHumidity, "%.0f"),
			getRowStyle(hasChanged("pressure", obs.StationPressure, 0.1)),
			obs.StationPressure, getPrevValue("pressure", obs.StationPressure, "%.2f"),
			getRowStyle(hasChanged("wind_speed", obs.WindAvg, 0.1)),
			windSpeedMph, obs.WindAvg, getPrevValue("wind_speed", obs.WindAvg, "%.1f"),
			getRowStyle(hasChanged("wind_gust", obs.WindGust, 0.1)),
			windGustMph, obs.WindGust, getPrevValue("wind_gust", obs.WindGust, "%.1f"),
			getRowStyle(hasChanged("wind_direction", obs.WindDirection, 5.0)),
			obs.WindDirection, cardinal, getPrevValue("wind_direction", obs.WindDirection, "%.0f"),
			getRowStyle(hasChanged("uv", float64(obs.UV), 0.5)),
			obs.UV, getPrevValue("uv", float64(obs.UV), "%.0f"),
			getRowStyle(hasChanged("lux", obs.Illuminance, 100.0)),
			formatNumber(obs.Illuminance), getPrevLux(),
			getRowStyle(hasChanged("rain_rate", obs.RainAccumulated, 0.01)),
			obs.RainAccumulated, getPrevValue("rain_rate", obs.RainAccumulated, "%.2f"),
			getRowStyle(hasChanged("rain_daily", obs.RainDailyTotal, 0.1)),
			rainDaily, obs.RainDailyTotal, getPrevValue("rain_daily", obs.RainDailyTotal, "%.1f"),
			getRowStyle(hasChanged("lightning_count", float64(obs.LightningStrikeCount), 0.5)),
			obs.LightningStrikeCount, getPrevValue("lightning_count", float64(obs.LightningStrikeCount), "%.0f"))
	}

	return fmt.Sprintf(`Temperature: %.1f°F (%.1f°C) [Last: %s°C]
Humidity: %.0f%% [Last: %s%%]
Pressure: %.2f mb [Last: %s mb]
Wind Speed: %.1f mph (%.1f m/s) [Last: %s m/s]
Wind Gust: %.1f mph (%.1f m/s) [Last: %s m/s]
Wind Direction: %.0f° (%s) [Last: %s°]
UV Index: %d [Last: %s]
Illuminance: %s lux [Last: %s lux]
Rain Rate: %.2f mm/hr [Last: %s mm/hr]
Daily Rain: %.2f in (%.1f mm) [Last: %s mm]
Lightning: %d strikes [Last: %s strikes]`,
		tempF, obs.AirTemperature, getPrevValue("temperature", obs.AirTemperature, "%.1f"),
		obs.RelativeHumidity, getPrevValue("humidity", obs.RelativeHumidity, "%.0f"),
		obs.StationPressure, getPrevValue("pressure", obs.StationPressure, "%.2f"),
		windSpeedMph, obs.WindAvg, getPrevValue("wind_speed", obs.WindAvg, "%.1f"),
		windGustMph, obs.WindGust, getPrevValue("wind_gust", obs.WindGust, "%.1f"),
		obs.WindDirection, cardinal, getPrevValue("wind_direction", obs.WindDirection, "%.0f"),
		obs.UV, getPrevValue("uv", float64(obs.UV), "%.0f"),
		formatNumber(obs.Illuminance), getPrevLux(),
		obs.RainAccumulated, getPrevValue("rain_rate", obs.RainAccumulated, "%.2f"),
		rainDaily, obs.RainDailyTotal, getPrevValue("rain_daily", obs.RainDailyTotal, "%.1f"),
		obs.LightningStrikeCount, getPrevValue("lightning_count", float64(obs.LightningStrikeCount), "%.0f"))
}

// formatNumber formats a number with thousands separator
func formatNumber(n float64) string {
	s := fmt.Sprintf("%.0f", n)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(c)
	}
	return result.String()
}

// expandTemplate replaces template variables with actual values
func expandTemplate(template string, alarm *Alarm, obs *weather.Observation, stationName string) string {
	result := template

	// Detect if this is an HTML template
	isHTML := strings.Contains(template, "<html>") || strings.Contains(template, "<table>") ||
		strings.Contains(template, "<div") || strings.Contains(template, "<h1>") ||
		strings.Contains(template, "<h2>") || strings.Contains(template, "<p>")

	// Replace observation values (current)
	replacements := map[string]string{
		"{{temperature}}":        fmt.Sprintf("%.1f", obs.AirTemperature),
		"{{temperature_f}}":      fmt.Sprintf("%.1f", obs.AirTemperature*9/5+32),
		"{{temperature_c}}":      fmt.Sprintf("%.1f", obs.AirTemperature),
		"{{humidity}}":           fmt.Sprintf("%.0f", obs.RelativeHumidity),
		"{{pressure}}":           fmt.Sprintf("%.2f", obs.StationPressure),
		"{{wind_speed}}":         fmt.Sprintf("%.1f", obs.WindAvg),
		"{{wind_gust}}":          fmt.Sprintf("%.1f", obs.WindGust),
		"{{wind_direction}}":     fmt.Sprintf("%.0f", obs.WindDirection),
		"{{lux}}":                fmt.Sprintf("%.0f", obs.Illuminance),
		"{{uv}}":                 fmt.Sprintf("%d", obs.UV),
		"{{rain_rate}}":          fmt.Sprintf("%.2f", obs.RainAccumulated),
		"{{rain_daily}}":         fmt.Sprintf("%.2f", obs.RainAccumulated),
		"{{lightning_count}}":    fmt.Sprintf("%d", obs.LightningStrikeCount),
		"{{lightning_distance}}": fmt.Sprintf("%.1f", obs.LightningStrikeAvg),
		"{{timestamp}}":          time.Unix(obs.Timestamp, 0).Format("2006-01-02 15:04:05 MST"),
		"{{station}}":            stationName,
		"{{alarm_name}}":         alarm.Name,
		"{{alarm_description}}":  alarm.Description,
		"{{alarm_condition}}":    alarm.Condition,
		// New composite variables
		"{{app_info}}":    formatAppInfo(isHTML),
		"{{alarm_info}}":  formatAlarmInfo(alarm, isHTML),
		"{{sensor_info}}": formatSensorInfoWithAlarm(obs, alarm, isHTML),
	}

	// Add previous values for change detection comparisons
	// These show the value that was compared against to trigger the alarm
	// Use trigger context if available (more accurate), otherwise fall back to previousValue
	if lastTemp, ok := alarm.GetTriggerValue("temperature"); ok {
		replacements["{{last_temperature}}"] = fmt.Sprintf("%.1f", lastTemp)
	} else if lastTemp, ok := alarm.GetPreviousValue("temperature"); ok {
		replacements["{{last_temperature}}"] = fmt.Sprintf("%.1f", lastTemp)
	} else {
		replacements["{{last_temperature}}"] = "N/A"
	}
	if lastHumidity, ok := alarm.GetTriggerValue("humidity"); ok {
		replacements["{{last_humidity}}"] = fmt.Sprintf("%.0f", lastHumidity)
	} else if lastHumidity, ok := alarm.GetPreviousValue("humidity"); ok {
		replacements["{{last_humidity}}"] = fmt.Sprintf("%.0f", lastHumidity)
	} else {
		replacements["{{last_humidity}}"] = "N/A"
	}
	if lastPressure, ok := alarm.GetTriggerValue("pressure"); ok {
		replacements["{{last_pressure}}"] = fmt.Sprintf("%.2f", lastPressure)
	} else if lastPressure, ok := alarm.GetPreviousValue("pressure"); ok {
		replacements["{{last_pressure}}"] = fmt.Sprintf("%.2f", lastPressure)
	} else {
		replacements["{{last_pressure}}"] = "N/A"
	}
	if lastWindSpeed, ok := alarm.GetTriggerValue("wind_speed"); ok {
		replacements["{{last_wind_speed}}"] = fmt.Sprintf("%.1f", lastWindSpeed)
	} else if lastWindSpeed, ok := alarm.GetPreviousValue("wind_speed"); ok {
		replacements["{{last_wind_speed}}"] = fmt.Sprintf("%.1f", lastWindSpeed)
	} else {
		replacements["{{last_wind_speed}}"] = "N/A"
	}
	if lastWindGust, ok := alarm.GetTriggerValue("wind_gust"); ok {
		replacements["{{last_wind_gust}}"] = fmt.Sprintf("%.1f", lastWindGust)
	} else if lastWindGust, ok := alarm.GetPreviousValue("wind_gust"); ok {
		replacements["{{last_wind_gust}}"] = fmt.Sprintf("%.1f", lastWindGust)
	} else {
		replacements["{{last_wind_gust}}"] = "N/A"
	}
	if lastWindDir, ok := alarm.GetTriggerValue("wind_direction"); ok {
		replacements["{{last_wind_direction}}"] = fmt.Sprintf("%.0f", lastWindDir)
	} else if lastWindDir, ok := alarm.GetPreviousValue("wind_direction"); ok {
		replacements["{{last_wind_direction}}"] = fmt.Sprintf("%.0f", lastWindDir)
	} else {
		replacements["{{last_wind_direction}}"] = "N/A"
	}
	if lastLux, ok := alarm.GetTriggerValue("lux"); ok {
		replacements["{{last_lux}}"] = fmt.Sprintf("%.0f", lastLux)
	} else if lastLux, ok := alarm.GetPreviousValue("lux"); ok {
		replacements["{{last_lux}}"] = fmt.Sprintf("%.0f", lastLux)
	} else {
		replacements["{{last_lux}}"] = "N/A"
	}
	if lastUV, ok := alarm.GetTriggerValue("uv"); ok {
		replacements["{{last_uv}}"] = fmt.Sprintf("%d", int(lastUV))
	} else if lastUV, ok := alarm.GetPreviousValue("uv"); ok {
		replacements["{{last_uv}}"] = fmt.Sprintf("%d", int(lastUV))
	} else {
		replacements["{{last_uv}}"] = "N/A"
	}
	if lastRainRate, ok := alarm.GetTriggerValue("rain_rate"); ok {
		replacements["{{last_rain_rate}}"] = fmt.Sprintf("%.2f", lastRainRate)
	} else if lastRainRate, ok := alarm.GetPreviousValue("rain_rate"); ok {
		replacements["{{last_rain_rate}}"] = fmt.Sprintf("%.2f", lastRainRate)
	} else {
		replacements["{{last_rain_rate}}"] = "N/A"
	}
	if lastRainDaily, ok := alarm.GetTriggerValue("rain_daily"); ok {
		replacements["{{last_rain_daily}}"] = fmt.Sprintf("%.2f", lastRainDaily)
	} else if lastRainDaily, ok := alarm.GetPreviousValue("rain_daily"); ok {
		replacements["{{last_rain_daily}}"] = fmt.Sprintf("%.2f", lastRainDaily)
	} else {
		replacements["{{last_rain_daily}}"] = "N/A"
	}
	if lastLightning, ok := alarm.GetTriggerValue("lightning_count"); ok {
		replacements["{{last_lightning_count}}"] = fmt.Sprintf("%d", int(lastLightning))
	} else if lastLightning, ok := alarm.GetPreviousValue("lightning_count"); ok {
		replacements["{{last_lightning_count}}"] = fmt.Sprintf("%d", int(lastLightning))
	} else {
		replacements["{{last_lightning_count}}"] = "N/A"
	}
	if lastLightningDist, ok := alarm.GetTriggerValue("lightning_distance"); ok {
		replacements["{{last_lightning_distance}}"] = fmt.Sprintf("%.1f", lastLightningDist)
	} else if lastLightningDist, ok := alarm.GetPreviousValue("lightning_distance"); ok {
		replacements["{{last_lightning_distance}}"] = fmt.Sprintf("%.1f", lastLightningDist)
	} else {
		replacements["{{last_lightning_distance}}"] = "N/A"
	}

	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}
