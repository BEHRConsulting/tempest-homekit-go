package alarm

import (
	"crypto/tls"
	"fmt"
	"log/syslog"
	"net/smtp"
	"os"
	"runtime"
	"strings"
	"time"

	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/weather"
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
	logger.Info("%s", message)
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

	// Expand templates
	subject := expandTemplate(channel.Email.Subject, alarm, obs, stationName)
	body := expandTemplate(channel.Email.Body, alarm, obs, stationName)

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
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
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
	case "microsoft365":
		// OAuth2 for Microsoft 365 would require additional implementation
		// For now, fall back to SMTP
		logger.Info("Microsoft 365 OAuth2 not yet implemented, using SMTP")
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
		// Use TLS connection
		tlsConfig := &tls.Config{
			ServerName: n.config.SMTPHost,
		}

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
	}

	// Non-TLS SMTP
	return smtp.SendMail(addr, auth, n.config.FromAddress, to, msg)
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

// expandTemplate replaces template variables with actual values
func expandTemplate(template string, alarm *Alarm, obs *weather.Observation, stationName string) string {
	result := template

	// Replace observation values
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
	}

	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}
