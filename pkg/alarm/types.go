package alarm

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// AlarmConfig represents the complete alarm configuration
type AlarmConfig struct {
	// Global email settings
	Email *EmailGlobalConfig `json:"email,omitempty"`
	// Global SMS settings
	SMS *SMSGlobalConfig `json:"sms,omitempty"`
	// Global syslog settings
	Syslog *SyslogConfig `json:"syslog,omitempty"`
	// List of alarm rules
	Alarms []Alarm `json:"alarms"`
}

// EmailGlobalConfig contains global email configuration
type EmailGlobalConfig struct {
	Provider     string `json:"provider"` // "smtp", "microsoft365"
	SMTPHost     string `json:"smtp_host,omitempty"`
	SMTPPort     int    `json:"smtp_port,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	FromAddress  string `json:"from_address"`
	FromName     string `json:"from_name,omitempty"`
	UseTLS       bool   `json:"use_tls"`
	UseOAuth2    bool   `json:"use_oauth2,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	TenantID     string `json:"tenant_id,omitempty"`
}

// SMSGlobalConfig contains global SMS configuration
type SMSGlobalConfig struct {
	Provider       string `json:"provider"` // "twilio", "aws_sns"
	AccountSID     string `json:"account_sid,omitempty"`
	AuthToken      string `json:"auth_token,omitempty"`
	FromNumber     string `json:"from_number,omitempty"`
	AWSAccessKey   string `json:"aws_access_key,omitempty"`
	AWSSecretKey   string `json:"aws_secret_key,omitempty"`
	AWSRegion      string `json:"aws_region,omitempty"`
	AWSSNSTopicARN string `json:"aws_sns_topic_arn,omitempty"`
}

// SyslogConfig contains syslog configuration
type SyslogConfig struct {
	Network  string `json:"network,omitempty"`  // "tcp", "udp", "" for local
	Address  string `json:"address,omitempty"`  // "localhost:514" or empty for local
	Priority string `json:"priority,omitempty"` // "info", "warning", "error"
	Tag      string `json:"tag,omitempty"`
}

// Alarm represents a single alarm rule
type Alarm struct {
	Name           string             `json:"name"`
	Description    string             `json:"description,omitempty"`
	Tags           []string           `json:"tags,omitempty"`
	Enabled        bool               `json:"enabled"`
	Condition      string             `json:"condition"`          // e.g., "temperature > 85", "humidity > 80 && temperature > 35", "*lightning_count"
	Cooldown       int                `json:"cooldown,omitempty"` // Seconds between repeated notifications
	Channels       []Channel          `json:"channels"`
	lastFired      time.Time          // Internal: last trigger time
	previousValue  map[string]float64 // Internal: previous field values for change detection
	triggerContext map[string]float64 // Internal: field values at time of trigger (for notification display)
}

// Channel represents a notification channel configuration
type Channel struct {
	Type     string            `json:"type"`               // "console", "email", "sms", "syslog", "eventlog"
	Template string            `json:"template,omitempty"` // For console, syslog, eventlog, sms
	Email    *EmailConfig      `json:"email,omitempty"`
	SMS      *SMSConfig        `json:"sms,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"` // Additional channel-specific settings
}

// EmailConfig for alarm-specific email settings
type EmailConfig struct {
	To      []string `json:"to"`
	CC      []string `json:"cc,omitempty"`
	BCC     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"` // Template string
}

// SMSConfig for alarm-specific SMS settings
type SMSConfig struct {
	To      []string `json:"to"`      // Phone numbers
	Message string   `json:"message"` // Template string
}

// LoadAlarmConfig loads alarm configuration from file or JSON string
func LoadAlarmConfig(input string) (*AlarmConfig, error) {
	var data []byte
	var err error
	isFile := false

	// Check if input is a file reference (@filename.json)
	if strings.HasPrefix(input, "@") {
		isFile = true
		filename := strings.TrimPrefix(input, "@")
		data, err = os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read alarm config file %s: %w", filename, err)
		}
	} else {
		// Treat as inline JSON string - first validate it's JSON
		data = []byte(input)

		// Test if the string is valid JSON before attempting to parse into AlarmConfig
		var jsonTest interface{}
		if err := json.Unmarshal(data, &jsonTest); err != nil {
			// Detect if they forgot the @ prefix
			if !strings.HasPrefix(input, "{") && (strings.HasSuffix(input, ".json") || strings.Contains(input, "/")) {
				return nil, fmt.Errorf("invalid JSON string: %w\nHint: Did you mean to use '@%s'? File paths must be prefixed with @", err, input)
			}

			// Provide detailed error for invalid JSON
			if syntaxErr, ok := err.(*json.SyntaxError); ok {
				// Calculate line and column
				lines := strings.Split(input, "\n")
				var line, col int
				offset := syntaxErr.Offset
				currentOffset := int64(0)

				for i, l := range lines {
					lineLen := int64(len(l) + 1) // +1 for newline
					if currentOffset+lineLen > offset {
						line = i + 1
						col = int(offset - currentOffset + 1)
						break
					}
					currentOffset += lineLen
				}

				if line == 0 {
					line = 1
					col = int(offset)
				}

				return nil, fmt.Errorf("invalid JSON syntax at line %d, column %d: %v\nProvide valid JSON string or use @filename.json to load from file", line, col, syntaxErr)
			}

			return nil, fmt.Errorf("invalid JSON string: %w\nProvide valid JSON string or use @filename.json to load from file", err)
		}
	}

	var config AlarmConfig
	if err := json.Unmarshal(data, &config); err != nil {
		if isFile {
			return nil, fmt.Errorf("failed to parse alarm config from file: %w", err)
		}
		return nil, fmt.Errorf("failed to parse alarm config from JSON string: %w\nEnsure your JSON matches the AlarmConfig structure", err)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid alarm config: %w", err)
	}

	return &config, nil
}

// Validate checks if the alarm configuration is valid
func (c *AlarmConfig) Validate() error {
	if len(c.Alarms) == 0 {
		return fmt.Errorf("at least one alarm must be defined")
	}

	names := make(map[string]bool)
	for i, alarm := range c.Alarms {
		if alarm.Name == "" {
			return fmt.Errorf("alarm at index %d: name is required", i)
		}
		if names[alarm.Name] {
			return fmt.Errorf("duplicate alarm name: %s", alarm.Name)
		}
		names[alarm.Name] = true

		if alarm.Condition == "" {
			return fmt.Errorf("alarm %s: condition is required", alarm.Name)
		}

		if len(alarm.Channels) == 0 {
			return fmt.Errorf("alarm %s: at least one channel is required", alarm.Name)
		}

		for j, channel := range alarm.Channels {
			if err := channel.Validate(); err != nil {
				return fmt.Errorf("alarm %s, channel %d: %w", alarm.Name, j, err)
			}
		}
	}

	return nil
}

// Validate checks if a channel configuration is valid
func (c *Channel) Validate() error {
	validTypes := map[string]bool{
		"console":  true,
		"email":    true,
		"sms":      true,
		"syslog":   true,
		"oslog":    true,
		"eventlog": true,
	}

	if !validTypes[c.Type] {
		return fmt.Errorf("invalid channel type: %s (must be console, email, sms, syslog, oslog, or eventlog)", c.Type)
	}

	switch c.Type {
	case "console", "syslog", "oslog", "eventlog":
		if c.Template == "" {
			return fmt.Errorf("template is required for %s channel", c.Type)
		}
	case "email":
		if c.Email == nil {
			return fmt.Errorf("email configuration is required for email channel")
		}
		if len(c.Email.To) == 0 {
			return fmt.Errorf("at least one recipient is required for email channel")
		}
		if c.Email.Subject == "" {
			return fmt.Errorf("subject is required for email channel")
		}
		if c.Email.Body == "" {
			return fmt.Errorf("body template is required for email channel")
		}
	case "sms":
		if c.SMS == nil {
			return fmt.Errorf("sms configuration is required for sms channel")
		}
		if len(c.SMS.To) == 0 {
			return fmt.Errorf("at least one phone number is required for sms channel")
		}
		if c.SMS.Message == "" {
			return fmt.Errorf("message template is required for sms channel")
		}
	}

	return nil
}

// CanFire checks if the alarm can fire based on cooldown
func (a *Alarm) CanFire() bool {
	if !a.Enabled {
		return false
	}
	if a.Cooldown == 0 {
		return true
	}
	return time.Since(a.lastFired) >= time.Duration(a.Cooldown)*time.Second
}

// MarkFired updates the last fired timestamp
func (a *Alarm) MarkFired() {
	a.lastFired = time.Now()
}

// GetLastFired returns the last fired timestamp
func (a *Alarm) GetLastFired() time.Time {
	return a.lastFired
}

// GetCooldownRemaining returns the remaining cooldown time in seconds (0 if can fire)
func (a *Alarm) GetCooldownRemaining() int {
	if !a.Enabled || a.Cooldown == 0 {
		return 0
	}
	elapsed := time.Since(a.lastFired)
	cooldownDuration := time.Duration(a.Cooldown) * time.Second
	if elapsed >= cooldownDuration {
		return 0
	}
	remaining := cooldownDuration - elapsed
	return int(remaining.Seconds())
}

// IsInCooldown returns true if the alarm is currently in cooldown
func (a *Alarm) IsInCooldown() bool {
	return a.GetCooldownRemaining() > 0
}

// GetPreviousValue returns the previous value for a field
func (a *Alarm) GetPreviousValue(field string) (float64, bool) {
	if a.previousValue == nil {
		return 0, false
	}
	val, ok := a.previousValue[field]
	return val, ok
}

// SetPreviousValue stores the previous value for a field
func (a *Alarm) SetPreviousValue(field string, value float64) {
	if a.previousValue == nil {
		a.previousValue = make(map[string]float64)
	}
	a.previousValue[field] = value
}

// GetTriggerValue returns the trigger context value for a field
func (a *Alarm) GetTriggerValue(field string) (float64, bool) {
	if a.triggerContext == nil {
		return 0, false
	}
	val, ok := a.triggerContext[field]
	return val, ok
}

// SetTriggerContext stores the field values at time of trigger
func (a *Alarm) SetTriggerContext(values map[string]float64) {
	a.triggerContext = values
}
