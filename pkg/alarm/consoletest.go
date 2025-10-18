package alarm

import (
	"fmt"
	"log"
	"os"
	"time"

	"tempest-homekit-go/pkg/weather"
)

// TestConsoleConfiguration tests console notification by sending a test message
func TestConsoleConfiguration(alarmsJSON, stationName string) error {
	fmt.Println("Testing console notification output...")
	fmt.Println()

	// Load alarm configuration (uses factory for real delivery path)
	config, err := LoadAlarmConfig(alarmsJSON)
	if err != nil {
		return fmt.Errorf("failed to load alarm configuration: %w", err)
	}

	// Create console notifier using factory
	factory := NewNotifierFactory(config)
	notifier, err := factory.GetNotifier("console")
	if err != nil {
		return fmt.Errorf("failed to create console notifier: %w", err)
	}

	// Create test alarm
	testAlarm := &Alarm{
		Name:        "Console Test",
		Description: "Test console notification output",
		Enabled:     true,
	}

	// Create test channel with template
	testChannel := &Channel{
		Type:     "console",
		Template: fmt.Sprintf("🔔 TEST NOTIFICATION from Tempest HomeKit Go\n\nStation: {{station}}\nTimestamp: {{timestamp}}\nTemperature: {{temperature_c}}°C / {{temperature_f}}°F\nHumidity: {{humidity}}%%\nPressure: {{pressure}} mb\n\nThis is a test of the console notification system."),
	}

	// Create test observation
	testObs := &weather.Observation{
		Timestamp:        time.Now().Unix(),
		AirTemperature:   20.0,
		RelativeHumidity: 50.0,
		WindAvg:          5.0,
		StationPressure:  1013.25,
	}

	// Send test notification
	fmt.Println("Output below should appear in console:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	if err = notifier.Send(testAlarm, testChannel, testObs, stationName); err != nil {
		return fmt.Errorf("failed to send test notification: %w", err)
	}
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println()

	fmt.Println("✅ Console notification test completed successfully!")
	fmt.Println()
	fmt.Println("The notification was printed above between the separator lines.")

	return nil
}

// RunConsoleTest is a convenience function that wraps TestConsoleConfiguration and exits
func RunConsoleTest(alarmsJSON, stationName string) {
	if err := TestConsoleConfiguration(alarmsJSON, stationName); err != nil {
		log.Fatalf("❌ Console test failed: %v", err)
	}
	os.Exit(0)
}
