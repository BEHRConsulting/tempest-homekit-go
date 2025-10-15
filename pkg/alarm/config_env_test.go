package alarm

import (
	"os"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original env vars
	origVars := map[string]string{
		"SMTP_HOST":           os.Getenv("SMTP_HOST"),
		"SMTP_PORT":           os.Getenv("SMTP_PORT"),
		"SMTP_USERNAME":       os.Getenv("SMTP_USERNAME"),
		"SMTP_PASSWORD":       os.Getenv("SMTP_PASSWORD"),
		"SMTP_FROM_ADDRESS":   os.Getenv("SMTP_FROM_ADDRESS"),
		"SMTP_FROM_NAME":      os.Getenv("SMTP_FROM_NAME"),
		"SMTP_USE_TLS":        os.Getenv("SMTP_USE_TLS"),
		"TWILIO_ACCOUNT_SID":  os.Getenv("TWILIO_ACCOUNT_SID"),
		"TWILIO_AUTH_TOKEN":   os.Getenv("TWILIO_AUTH_TOKEN"),
		"TWILIO_FROM_NUMBER":  os.Getenv("TWILIO_FROM_NUMBER"),
		"MS365_CLIENT_ID":     os.Getenv("MS365_CLIENT_ID"),
		"MS365_CLIENT_SECRET": os.Getenv("MS365_CLIENT_SECRET"),
		"MS365_TENANT_ID":     os.Getenv("MS365_TENANT_ID"),
		"MS365_FROM_ADDRESS":  os.Getenv("MS365_FROM_ADDRESS"),
	}
	defer func() {
		// Restore original env vars
		for k, v := range origVars {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	t.Run("SMTP configuration", func(t *testing.T) {
		// Clear all env vars
		for k := range origVars {
			os.Unsetenv(k)
		}

		// Set SMTP env vars
		os.Setenv("SMTP_HOST", "smtp.example.com")
		os.Setenv("SMTP_PORT", "587")
		os.Setenv("SMTP_USERNAME", "user@example.com")
		os.Setenv("SMTP_PASSWORD", "password123")
		os.Setenv("SMTP_FROM_ADDRESS", "alerts@example.com")
		os.Setenv("SMTP_FROM_NAME", "Test Alerts")
		os.Setenv("SMTP_USE_TLS", "true")

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("LoadConfigFromEnv failed: %v", err)
		}

		if config.Email == nil {
			t.Fatal("Email config should not be nil")
		}

		if config.Email.Provider != "smtp" {
			t.Errorf("Expected provider 'smtp', got '%s'", config.Email.Provider)
		}
		if config.Email.SMTPHost != "smtp.example.com" {
			t.Errorf("Expected SMTP host 'smtp.example.com', got '%s'", config.Email.SMTPHost)
		}
		if config.Email.SMTPPort != 587 {
			t.Errorf("Expected SMTP port 587, got %d", config.Email.SMTPPort)
		}
		if config.Email.Username != "user@example.com" {
			t.Errorf("Expected username 'user@example.com', got '%s'", config.Email.Username)
		}
		if config.Email.FromAddress != "alerts@example.com" {
			t.Errorf("Expected from address 'alerts@example.com', got '%s'", config.Email.FromAddress)
		}
		if !config.Email.UseTLS {
			t.Error("Expected UseTLS to be true")
		}
	})

	t.Run("Microsoft 365 configuration", func(t *testing.T) {
		// Clear all env vars
		for k := range origVars {
			os.Unsetenv(k)
		}

		// Set MS365 env vars
		os.Setenv("MS365_CLIENT_ID", "client-id-123")
		os.Setenv("MS365_CLIENT_SECRET", "client-secret-456")
		os.Setenv("MS365_TENANT_ID", "tenant-id-789")
		os.Setenv("MS365_FROM_ADDRESS", "alerts@company.com")

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("LoadConfigFromEnv failed: %v", err)
		}

		if config.Email == nil {
			t.Fatal("Email config should not be nil")
		}

		if config.Email.Provider != "microsoft365" {
			t.Errorf("Expected provider 'microsoft365', got '%s'", config.Email.Provider)
		}
		if !config.Email.UseOAuth2 {
			t.Error("Expected UseOAuth2 to be true")
		}
		if config.Email.ClientID != "client-id-123" {
			t.Errorf("Expected client ID 'client-id-123', got '%s'", config.Email.ClientID)
		}
		if config.Email.FromAddress != "alerts@company.com" {
			t.Errorf("Expected from address 'alerts@company.com', got '%s'", config.Email.FromAddress)
		}
	})

	t.Run("Twilio SMS configuration", func(t *testing.T) {
		// Clear all env vars
		for k := range origVars {
			os.Unsetenv(k)
		}

		// Set Twilio env vars
		os.Setenv("TWILIO_ACCOUNT_SID", "AC1234567890")
		os.Setenv("TWILIO_AUTH_TOKEN", "auth-token-abc")
		os.Setenv("TWILIO_FROM_NUMBER", "+15555551234")

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("LoadConfigFromEnv failed: %v", err)
		}

		if config.SMS == nil {
			t.Fatal("SMS config should not be nil")
		}

		if config.SMS.Provider != "twilio" {
			t.Errorf("Expected provider 'twilio', got '%s'", config.SMS.Provider)
		}
		if config.SMS.AccountSID != "AC1234567890" {
			t.Errorf("Expected account SID 'AC1234567890', got '%s'", config.SMS.AccountSID)
		}
		if config.SMS.FromNumber != "+15555551234" {
			t.Errorf("Expected from number '+15555551234', got '%s'", config.SMS.FromNumber)
		}
	})

	t.Run("No environment configuration", func(t *testing.T) {
		// Clear all env vars
		for k := range origVars {
			os.Unsetenv(k)
		}

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("LoadConfigFromEnv failed: %v", err)
		}

		if config.Email != nil {
			t.Error("Email config should be nil when no env vars set")
		}
		if config.SMS != nil {
			t.Error("SMS config should be nil when no env vars set")
		}
	})

	t.Run("Environment overrides JSON config", func(t *testing.T) {
		// Clear all env vars
		for k := range origVars {
			os.Unsetenv(k)
		}

		// Set env vars
		os.Setenv("SMTP_HOST", "smtp.override.com")
		os.Setenv("SMTP_PORT", "465")
		os.Setenv("SMTP_USERNAME", "override@example.com")
		os.Setenv("SMTP_PASSWORD", "override-pass")
		os.Setenv("SMTP_FROM_ADDRESS", "override@example.com")
		os.Setenv("SMTP_USE_TLS", "false")

		// Load alarm config with JSON that has email config (should be overridden)
		jsonConfig := `{
			"email": {
				"provider": "smtp",
				"smtp_host": "smtp.json.com",
				"smtp_port": 587,
				"username": "json@example.com",
				"password": "json-pass",
				"from_address": "json@example.com",
				"use_tls": true
			},
			"alarms": [
				{
					"name": "test",
					"enabled": true,
					"condition": "temperature > 100",
					"channels": [{"type": "console", "template": "test"}]
				}
			]
		}`

		config, err := LoadAlarmConfig(jsonConfig)
		if err != nil {
			t.Fatalf("LoadAlarmConfig failed: %v", err)
		}

		// Environment should override JSON
		if config.Email.SMTPHost != "smtp.override.com" {
			t.Errorf("Expected env to override SMTP host, got '%s'", config.Email.SMTPHost)
		}
		if config.Email.SMTPPort != 465 {
			t.Errorf("Expected env to override SMTP port, got %d", config.Email.SMTPPort)
		}
		if config.Email.UseTLS {
			t.Error("Expected env to override UseTLS to false")
		}
	})
}
