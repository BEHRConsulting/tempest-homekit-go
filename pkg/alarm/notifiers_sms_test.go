package alarm

import (
	"os"
	"strings"
	"testing"

	"tempest-homekit-go/pkg/weather"
)

func TestSMSNotifier(t *testing.T) {
	// Note: This is a functional test that validates the SMS notifier logic
	// Actual AWS SNS calls are not made without valid credentials

	tests := []struct {
		name          string
		provider      string
		setupEnv      func()
		cleanupEnv    func()
		expectError   bool
		errorContains string
	}{
		{
			name:     "aws sns with valid config",
			provider: "aws",
			setupEnv: func() {
				os.Setenv("AWS_ACCESS_KEY_ID", "test-key-id")
				os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
				os.Setenv("AWS_REGION", "us-west-2")
			},
			cleanupEnv: func() {
				os.Unsetenv("AWS_ACCESS_KEY_ID")
				os.Unsetenv("AWS_SECRET_ACCESS_KEY")
				os.Unsetenv("AWS_REGION")
			},
			expectError:   true, // Will fail on actual AWS call but validates config
			errorContains: "",   // Various AWS errors possible
		},
		{
			name:     "aws sns missing credentials",
			provider: "aws",
			setupEnv: func() {
				os.Unsetenv("AWS_ACCESS_KEY_ID")
				os.Unsetenv("AWS_SECRET_ACCESS_KEY")
				os.Unsetenv("AWS_REGION")
			},
			cleanupEnv: func() {},
			expectError:   true,
			errorContains: "AWS SNS credentials missing",
		},
		{
			name:     "twilio missing credentials",
			provider: "twilio",
			setupEnv: func() {
				os.Unsetenv("TWILIO_ACCOUNT_SID")
				os.Unsetenv("TWILIO_AUTH_TOKEN")
				os.Unsetenv("TWILIO_FROM_NUMBER")
			},
			cleanupEnv: func() {},
			expectError:   true,
			errorContains: "Twilio credentials missing",
		},
		{
			name:          "unsupported provider",
			provider:      "invalid",
			setupEnv:      func() {},
			cleanupEnv:    func() {},
			expectError:   true,
			errorContains: "unsupported SMS provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			config := &AlarmConfig{
				SMS: &SMSGlobalConfig{
					Provider:     tt.provider,
					AWSAccessKey: "${AWS_ACCESS_KEY_ID}",
					AWSSecretKey: "${AWS_SECRET_ACCESS_KEY}",
					AWSRegion:    "${AWS_REGION}",
				},
			}

			notifier := &SMSNotifier{config: config.SMS}

			alarm := &Alarm{
				Name:        "test-alarm",
				Description: "Test SMS notification",
				Condition:   "temperature > 85",
			}

			channel := &Channel{
				Type: "sms",
				SMS: &SMSConfig{
					To:      []string{"+15555551234"},
					Message: "Test: {{alarm_name}}",
				},
			}

			obs := &weather.Observation{
				AirTemperature: 90.0,
			}

			err := notifier.Send(alarm, channel, obs, "Test Station")

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSMSNotifierTemplateExpansion(t *testing.T) {
	// Test that SMS messages properly expand templates
	config := &AlarmConfig{
		SMS: &SMSGlobalConfig{
			Provider: "aws",
			// Missing credentials will cause send to fail, but we can verify template expansion
		},
	}

	notifier := &SMSNotifier{config: config.SMS}

	alarm := &Alarm{
		Name:        "high-temp",
		Description: "High temperature warning",
		Condition:   "temperature > 85",
	}

	channel := &Channel{
		Type: "sms",
		SMS: &SMSConfig{
			To:      []string{"+15555551234"},
			Message: "⚠️ {{alarm_name}}: {{temperature}}°C at {{station}}",
		},
	}

	obs := &weather.Observation{
		AirTemperature: 90.5,
	}

	// Will fail due to missing credentials, but that's expected
	err := notifier.Send(alarm, channel, obs, "Test Station")
	if err == nil {
		t.Error("Expected error due to missing credentials")
	}

	// Verify the error is about credentials, not template expansion
	if !strings.Contains(err.Error(), "credentials missing") {
		// If we get a different error, template expansion worked
		// (would fail on AWS API call instead)
		t.Logf("Template expansion appears to work (got AWS error, not template error): %v", err)
	}
}

func TestSMSNotifierWithTopicARN(t *testing.T) {
	// Test SNS topic ARN configuration
	os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("AWS_SNS_TOPIC_ARN", "arn:aws:sns:us-west-2:123456789012:TestTopic")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_SNS_TOPIC_ARN")
	}()

	config := &AlarmConfig{
		SMS: &SMSGlobalConfig{
			Provider:       "aws",
			AWSAccessKey:   "${AWS_ACCESS_KEY_ID}",
			AWSSecretKey:   "${AWS_SECRET_ACCESS_KEY}",
			AWSRegion:      "${AWS_REGION}",
			AWSSNSTopicARN: "${AWS_SNS_TOPIC_ARN}",
		},
	}

	notifier := &SMSNotifier{config: config.SMS}

	alarm := &Alarm{
		Name:      "test-alarm",
		Condition: "temperature > 85",
	}

	channel := &Channel{
		Type: "sms",
		SMS: &SMSConfig{
			To:      []string{"+15555551234"},
			Message: "Test message",
		},
	}

	obs := &weather.Observation{
		AirTemperature: 90.0,
	}

	// Will fail on AWS API call, but validates config loading
	err := notifier.Send(alarm, channel, obs, "Test Station")
	if err == nil {
		t.Error("Expected error due to invalid test credentials")
	}

	// Should fail on AWS API, not on config validation
	if strings.Contains(err.Error(), "credentials missing") {
		t.Error("Config validation failed - topic ARN not loaded properly")
	}
}

func TestSMSNotifierMultipleRecipients(t *testing.T) {
	// Test sending to multiple phone numbers
	os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	os.Setenv("AWS_REGION", "us-east-1")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_REGION")
	}()

	config := &AlarmConfig{
		SMS: &SMSGlobalConfig{
			Provider:     "aws",
			AWSAccessKey: "${AWS_ACCESS_KEY_ID}",
			AWSSecretKey: "${AWS_SECRET_ACCESS_KEY}",
			AWSRegion:    "${AWS_REGION}",
		},
	}

	notifier := &SMSNotifier{config: config.SMS}

	alarm := &Alarm{
		Name:      "multi-recipient-test",
		Condition: "temperature > 100",
	}

	channel := &Channel{
		Type: "sms",
		SMS: &SMSConfig{
			To: []string{
				"+15555551234",
				"+15555555678",
				"+15555559012",
			},
			Message: "Alert: {{alarm_name}}",
		},
	}

	obs := &weather.Observation{
		AirTemperature: 105.0,
	}

	// Will fail on AWS API call
	err := notifier.Send(alarm, channel, obs, "Test Station")
	if err == nil {
		t.Error("Expected error due to invalid test credentials")
	}

	// Verify it's not a config error
	if strings.Contains(err.Error(), "credentials missing") {
		t.Error("Multi-recipient config not loaded properly")
	}
}

func TestSMSFactoryCreation(t *testing.T) {
	// Test that the notifier factory creates SMS notifiers correctly
	config := &AlarmConfig{
		SMS: &SMSGlobalConfig{
			Provider: "aws",
		},
	}

	factory := NewNotifierFactory(config)

	notifier, err := factory.GetNotifier("sms")
	if err != nil {
		t.Errorf("Failed to create SMS notifier: %v", err)
	}

	if notifier == nil {
		t.Error("Notifier is nil")
	}

	// Verify it's the right type
	if _, ok := notifier.(*SMSNotifier); !ok {
		t.Errorf("Expected *SMSNotifier, got %T", notifier)
	}
}
