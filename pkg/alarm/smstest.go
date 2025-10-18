package alarm

import (
	"fmt"
	"log"
	"os"
	"strings"

	"tempest-homekit-go/pkg/weather"
)

// TestSMSConfiguration tests the SMS configuration by sending a test SMS
func TestSMSConfiguration(alarmsJSON, stationName string) error {
	// Load alarm config to get provider configuration
	config, err := LoadAlarmConfig(alarmsJSON)
	if err != nil {
		return fmt.Errorf("failed to load alarm configuration: %w", err)
	}

	// Check if SMS is configured
	if config.SMS == nil {
		return fmt.Errorf("no SMS configuration found - set TWILIO_* or AWS_* environment variables in .env")
	}

	provider := config.SMS.Provider
	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println("SMS CONFIGURATION TEST")
	fmt.Println("================================================================")
	fmt.Println()
	fmt.Printf("Provider: %s\n", provider)
	fmt.Println()

	// Display provider-specific configuration
	switch provider {
	case "twilio":
		fmt.Println("Twilio Configuration:")
		fmt.Printf("  Account SID: %s\n", maskString(config.SMS.AccountSID))
		fmt.Printf("  Auth Token: %s\n", maskString(config.SMS.AuthToken))
		fmt.Printf("  From Number: %s\n", config.SMS.FromNumber)
		fmt.Println()
		if config.SMS.AccountSID == "" || config.SMS.AuthToken == "" || config.SMS.FromNumber == "" {
			fmt.Println("⚠️  WARNING: Missing Twilio credentials")
			fmt.Println("   Required environment variables:")
			fmt.Println("   - TWILIO_ACCOUNT_SID")
			fmt.Println("   - TWILIO_AUTH_TOKEN")
			fmt.Println("   - TWILIO_FROM_NUMBER")
			return fmt.Errorf("incomplete Twilio configuration")
		}
	case "aws_sns", "sns", "aws":
		fmt.Println("AWS SNS Configuration:")
		fmt.Printf("  Access Key ID: %s\n", maskString(config.SMS.AWSAccessKey))
		fmt.Printf("  Secret Key: %s\n", maskString(config.SMS.AWSSecretKey))
		fmt.Printf("  Region: %s\n", config.SMS.AWSRegion)
		if config.SMS.AWSSNSTopicARN != "" {
			fmt.Printf("  Topic ARN: %s\n", config.SMS.AWSSNSTopicARN)
		}
		fmt.Println()
		if config.SMS.AWSAccessKey == "" || config.SMS.AWSSecretKey == "" || config.SMS.AWSRegion == "" {
			fmt.Println("⚠️  WARNING: Missing AWS SNS credentials")
			fmt.Println("   Required environment variables:")
			fmt.Println("   - AWS_ACCESS_KEY_ID")
			fmt.Println("   - AWS_SECRET_ACCESS_KEY")
			fmt.Println("   - AWS_REGION")
			return fmt.Errorf("incomplete AWS SNS configuration")
		}
	default:
		return fmt.Errorf("unsupported SMS provider: %s", provider)
	}

	// Recipient phone comes from command line parameter
	recipientNumber := os.Getenv("TEST_SMS_RECIPIENT")
	if recipientNumber == "" {
		return fmt.Errorf("no recipient phone number provided")
	}

	// Validate E.164 format
	if !strings.HasPrefix(recipientNumber, "+") {
		return fmt.Errorf("phone number must start with + (E.164 format required)")
	}
	fmt.Println("================================================================")

	fmt.Println()
	fmt.Println("Sending test SMS...")
	fmt.Println()

	// Create SMS notifier using factory (same path as real alarms)
	factory := NewNotifierFactory(config)
	smsNotifier, err := factory.GetNotifier("sms")
	if err != nil {
		return fmt.Errorf("failed to create SMS notifier: %w", err)
	}

	testAlarm := &Alarm{
		Name:        "Test SMS",
		Description: "Test SMS to verify SMS settings",
		Enabled:     true,
	}

	testChannel := &Channel{
		Type: "sms",
		SMS: &SMSConfig{
			To:      []string{recipientNumber},
			Message: fmt.Sprintf("Test SMS from %s\n\nProvider: %s\nTime: {{timestamp}}\nStation: %s\n\nIf you received this, your SMS configuration is working correctly!", appVersion, provider, stationName),
		},
	}

	// Create test observation
	testObs := &weather.Observation{
		Timestamp:        int64(0), // Will use current time in template
		AirTemperature:   20.0,
		RelativeHumidity: 50.0,
		WindAvg:          5.0,
		StationPressure:  1013.25,
	}

	// Send test SMS
	err = smsNotifier.Send(testAlarm, testChannel, testObs, stationName)
	if err != nil {
		return fmt.Errorf("failed to send test SMS: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Test SMS sent successfully!")
	fmt.Println()
	fmt.Printf("Check the phone %s for the test message\n", recipientNumber)
	fmt.Println()

	switch provider {
	case "twilio":
		fmt.Println("If you don't see the SMS:")
		fmt.Println("  1. Check if your Twilio account is in trial mode")
		fmt.Println("  2. Verify the recipient number is verified (trial accounts only)")
		fmt.Println("  3. Check Twilio console for delivery status")
		fmt.Println("  4. Ensure your Twilio number has SMS capabilities")
		fmt.Println("  5. Check for sufficient Twilio account balance")
	case "aws_sns", "sns", "aws":
		fmt.Println("If you don't see the SMS:")
		fmt.Println("  1. Check if your AWS account is in SNS sandbox mode")
		fmt.Println("  2. Verify the recipient number is verified (sandbox accounts only)")
		fmt.Println("  3. Check AWS SNS console for delivery status")
		fmt.Println("  4. Ensure your region supports SMS (not all regions do)")
		fmt.Println("  5. Check AWS spending limits and account status")
		fmt.Println("  6. Run ./scripts/setup-aws-sns.sh to configure production settings")
	}

	fmt.Println()
	fmt.Println("================================================================")

	return nil
}

// RunSMSTest is a convenience function that wraps TestSMSConfiguration and exits
func RunSMSTest(alarmsJSON, stationName string) {
	err := TestSMSConfiguration(alarmsJSON, stationName)
	if err != nil {
		log.Fatalf("SMS test failed: %v", err)
	}
	os.Exit(0)
}

// maskString masks a string for display (shows first 4 and last 4 characters)
func maskString(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "***" + s[len(s)-2:]
	}
	return s[:4] + "****" + s[len(s)-4:]
}
