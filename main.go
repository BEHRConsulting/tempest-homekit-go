// Package main implements a HomeKit bridge for WeatherFlow Tempest weather stations.
// It provides a HomeKit-compatible interface to access weather data from Tempest stations,
// along with a web dashboard for monitoring and configuration.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tempest-homekit-go/pkg/alarm"
	"tempest-homekit-go/pkg/alarm/editor"
	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/service"
	"tempest-homekit-go/pkg/udp"
	"tempest-homekit-go/pkg/weather"

	"github.com/joho/godotenv"
)

func main() {
	// Parse --env flag early to determine which environment file to load
	envFile := ".env"
	for i, arg := range os.Args {
		if arg == "--env" && i+1 < len(os.Args) {
			envFile = os.Args[i+1]
			break
		}
	}

	// Load environment file (silently ignore if not present)
	if err := godotenv.Load(envFile); err != nil && envFile != ".env" {
		// If a custom env file was specified but couldn't be loaded, show error
		log.Printf("Warning: Could not load environment file '%s': %v", envFile, err)
	}

	cfg := config.LoadConfig()

	// Set up logging first (before any other operations that might log)
	logger.SetLogLevel(cfg.LogLevel)
	if cfg.LogFilter != "" {
		logger.SetLogFilter(cfg.LogFilter)
		logger.Info("Log filter enabled: only messages containing '%s' will be shown", cfg.LogFilter)
	}

	// Note: For generated weather, elevation will be logged by the service once location is selected

	// Handle version flag
	if cfg.Version {
		fmt.Println("tempest-homekit-go v1.8.0")
		fmt.Println("Built with Go 1.24.2")
		fmt.Println("HomeKit integration for WeatherFlow Tempest weather stations")
		os.Exit(0)
	}

	// Handle alarm editor mode
	if cfg.AlarmsEdit != "" {
		logger.Info("Alarm editor mode detected, starting alarm editor...")
		editorServer, err := editor.NewServer(cfg.AlarmsEdit, cfg.AlarmsEditPort)
		if err != nil {
			log.Fatalf("Failed to create alarm editor: %v", err)
		}
		if err := editorServer.Start(); err != nil {
			log.Fatalf("Failed to start alarm editor: %v", err)
		}
		return
	}

	// Handle email testing if requested
	if cfg.TestEmail != "" {
		// Validate email address doesn't look like a flag
		if strings.HasPrefix(cfg.TestEmail, "-") {
			log.Fatalf("Invalid email address: %s. Usage: --test-email user@example.com", cfg.TestEmail)
		}
		logger.Info("TestEmail flag detected, sending test email to %s...", cfg.TestEmail)
		runEmailTest(cfg)
		return
	}

	// Handle SMS testing if requested
	if cfg.TestSMS != "" {
		// Validate phone number doesn't look like a flag
		if strings.HasPrefix(cfg.TestSMS, "-") && !strings.HasPrefix(cfg.TestSMS, "+") {
			log.Fatalf("Invalid phone number: %s. Usage: --test-sms +15555551234", cfg.TestSMS)
		}
		logger.Info("TestSMS flag detected, sending test SMS to %s...", cfg.TestSMS)
		runSMSTest(cfg)
		return
	}

	// Handle console testing if requested
	if cfg.TestConsole {
		logger.Info("TestConsole flag detected, sending test console notification...")
		runConsoleTest(cfg)
		return
	}

	// Handle syslog testing if requested
	if cfg.TestSyslog {
		logger.Info("TestSyslog flag detected, sending test syslog notification...")
		runSyslogTest(cfg)
		return
	}

	// Handle oslog testing if requested
	if cfg.TestOSLog {
		logger.Info("TestOSLog flag detected, sending test oslog notification...")
		runOSLogTest(cfg)
		return
	}

	// Handle eventlog testing if requested
	if cfg.TestEventLog {
		logger.Info("TestEventLog flag detected, sending test eventlog notification...")
		runEventLogTest(cfg)
		return
	}

	// Handle UDP testing if requested
	if cfg.TestUDP != 0 || (len(os.Args) > 1 && contains(os.Args, "--test-udp")) {
		seconds := cfg.TestUDP
		if seconds == 0 {
			seconds = 120 // Default to 120 seconds
		}
		logger.Info("TestUDP flag detected, listening for UDP broadcasts for %d seconds...", seconds)
		runUDPTest(cfg, seconds)
		return
	}

	// Handle HomeKit testing if requested
	if cfg.TestHomeKit {
		logger.Info("TestHomeKit flag detected, testing HomeKit bridge setup...")
		runHomeKitTest(cfg)
		return
	}

	// Handle web status testing if requested
	if cfg.TestWebStatus {
		logger.Info("TestWebStatus flag detected, testing web status scraping...")
		runWebStatusTest(cfg)
		return
	}

	// Handle alarm testing if requested
	if cfg.TestAlarm != "" {
		logger.Info("TestAlarm flag detected, triggering alarm '%s'...", cfg.TestAlarm)
		runAlarmTest(cfg)
		return
	}

	// Handle API testing if requested
	if cfg.TestAPI {
		logger.Info("TestAPI flag detected, running API endpoint tests...")
		runAPITests(cfg)
		return
	}

	// Handle database clearing if requested
	if cfg.ClearDB {
		logger.Info("ClearDB flag detected, clearing HomeKit database...")
		if err := config.ClearDatabase("./db"); err != nil {
			log.Fatalf("Failed to clear database: %v", err)
		}
		logger.Info("Database cleared successfully. Please restart the application without --cleardb flag.")
		return
	}

	logger.Info("Starting service with config: WebPort=%s, LogLevel=%s", cfg.WebPort, cfg.LogLevel)
	err := service.StartService(cfg, "1.8.0")
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}

	logger.Info("Service started successfully, waiting for interrupt signal...") // Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	logger.Info("Received signal %v, shutting down...", sig)
}

// runEmailTest sends a test email using the configured email settings
func runEmailTest(cfg *config.Config) {
	fmt.Println("=== Email Configuration Test ===")
	fmt.Println()

	if cfg.Alarms == "" {
		log.Fatal("❌ No alarm configuration specified. Use --alarms flag or ALARMS environment variable.")
	}

	// Set recipient via environment variable for test function
	os.Setenv("TEST_EMAIL_RECIPIENT", cfg.TestEmail)

	// Use alarm package's email test function
	alarm.RunEmailTest(cfg.Alarms, cfg.StationName)
}

// runSMSTest sends a test SMS using the configured SMS settings
func runSMSTest(cfg *config.Config) {
	fmt.Println("=== SMS Configuration Test ===")
	fmt.Println()

	if cfg.Alarms == "" {
		log.Fatal("❌ No alarm configuration specified. Use --alarms flag or ALARMS environment variable.")
	}

	// Set recipient via environment variable for test function
	os.Setenv("TEST_SMS_RECIPIENT", cfg.TestSMS)

	// Use alarm package's SMS test function
	alarm.RunSMSTest(cfg.Alarms, cfg.StationName)
}

// runConsoleTest sends a test console notification
func runConsoleTest(cfg *config.Config) {
	fmt.Println("=== Console Notification Test ===")
	fmt.Println()

	if cfg.Alarms == "" {
		log.Fatal("❌ No alarm configuration specified. Use --alarms flag or ALARMS environment variable.")
	}

	// Use alarm package's console test function
	alarm.RunConsoleTest(cfg.Alarms, cfg.StationName)
}

// runSyslogTest sends a test syslog notification
func runSyslogTest(cfg *config.Config) {
	fmt.Println("=== Syslog Notification Test ===")
	fmt.Println()

	if cfg.Alarms == "" {
		log.Fatal("❌ No alarm configuration specified. Use --alarms flag or ALARMS environment variable.")
	}

	// Use alarm package's syslog test function
	alarm.RunSyslogTest(cfg.Alarms, cfg.StationName)
}

// runOSLogTest sends a test oslog notification
func runOSLogTest(cfg *config.Config) {
	fmt.Println("=== OSLog Notification Test ===")
	fmt.Println()

	if cfg.Alarms == "" {
		log.Fatal("❌ No alarm configuration specified. Use --alarms flag or ALARMS environment variable.")
	}

	// Use alarm package's oslog test function
	alarm.RunOSLogTest(cfg.Alarms, cfg.StationName)
}

// runEventLogTest sends a test eventlog notification
func runEventLogTest(cfg *config.Config) {
	fmt.Println("=== EventLog Notification Test ===")
	fmt.Println()

	if cfg.Alarms == "" {
		log.Fatal("❌ No alarm configuration specified. Use --alarms flag or ALARMS environment variable.")
	}

	// Use alarm package's eventlog test function
	alarm.RunEventLogTest(cfg.Alarms, cfg.StationName)
}

// runUDPTest listens for UDP broadcasts from a local Tempest station
func runUDPTest(_ *config.Config, seconds int) {
	fmt.Printf("=== UDP Broadcast Listener Test (%d seconds) ===\n\n", seconds)

	udpListener := udp.NewUDPListener(100)

	// Set up packet callback for real-time pretty printing
	udpListener.SetPacketCallback(func(data []byte) {
		fmt.Println(udp.PrettyPrintMessage(data))
	})

	fmt.Println("📡 Starting UDP listener on port 50222...")
	if err := udpListener.Start(); err != nil {
		log.Fatalf("❌ Failed to start UDP listener: %v", err)
	}
	defer udpListener.Stop()

	fmt.Println("✅ UDP listener started successfully")
	fmt.Printf("⏱️  Listening for %d seconds...\n\n", seconds)
	fmt.Println("Waiting for UDP broadcasts from Tempest station...")
	fmt.Println("(Make sure your station is on the same network and broadcasting)")
	fmt.Println("\n--- Live Packet Stream ---")
	fmt.Println()

	// Create ticker for periodic stats
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Create timeout timer
	timeout := time.After(time.Duration(seconds) * time.Second)

	startTime := time.Now()
	lastPacketCount := int64(0)

	for {
		select {
		case <-timeout:
			elapsed := time.Since(startTime)
			fmt.Printf("\n⏱️  Test completed after %v\n", elapsed)

			// Get final stats
			packetCount, lastPacket, stationIP, serialNumber := udpListener.GetStats()
			fmt.Println("\n=== Final Statistics ===")
			fmt.Printf("Total packets received: %d\n", packetCount)
			if packetCount > 0 {
				fmt.Printf("Station IP: %s\n", stationIP)
				fmt.Printf("Serial Number: %s\n", serialNumber)
				fmt.Printf("Last packet: %v\n", lastPacket.Format("2006-01-02 15:04:05"))

				// Get latest observation
				if obs := udpListener.GetLatestObservation(); obs != nil {
					fmt.Println("\n=== Latest Observation ===")
					fmt.Printf("Temperature: %.1f°C (%.1f°F)\n", obs.AirTemperature, obs.AirTemperature*9/5+32)
					fmt.Printf("Humidity: %.0f%%\n", obs.RelativeHumidity)
					fmt.Printf("Pressure: %.2f mb\n", obs.StationPressure)
					fmt.Printf("Wind Speed: %.1f m/s\n", obs.WindAvg)
					fmt.Printf("Wind Gust: %.1f m/s\n", obs.WindGust)
					fmt.Printf("Wind Direction: %.0f°\n", obs.WindDirection)
					fmt.Printf("UV Index: %d\n", obs.UV)
					fmt.Printf("Light: %.0f lux\n", obs.Illuminance)
					fmt.Printf("Rain Rate: %.3f in\n", obs.RainAccumulated)
					if obs.LightningStrikeCount > 0 {
						fmt.Printf("Lightning: %d strikes, avg %.1f km away\n", obs.LightningStrikeCount, obs.LightningStrikeAvg)
					}
				}

				fmt.Println("\n✅ UDP broadcast test completed successfully!")
			} else {
				fmt.Println("\n⚠️  No packets received. Possible issues:")
				fmt.Println("  - Tempest station not on same network")
				fmt.Println("  - Firewall blocking UDP port 50222")
				fmt.Println("  - Station not broadcasting (check station settings)")
			}
			os.Exit(0)
			return

		case <-ticker.C:
			// Show periodic statistics
			packetCount, _, stationIP, serialNumber := udpListener.GetStats()
			if packetCount > lastPacketCount {
				newPackets := packetCount - lastPacketCount
				elapsed := time.Since(startTime).Truncate(time.Second)
				fmt.Printf("\n📊 [%v elapsed] Total: %d packets | New: %d | Station: %s | Serial: %s\n\n", elapsed, packetCount, newPackets, stationIP, serialNumber)
				lastPacketCount = packetCount
			} else if packetCount == 0 {
				elapsed := time.Since(startTime).Truncate(time.Second)
				fmt.Printf("⏳ [%v elapsed] Still waiting for packets...\n", elapsed)
			}
		}
	}
}

// runHomeKitTest tests the HomeKit bridge setup
func runHomeKitTest(cfg *config.Config) {
	fmt.Println("=== HomeKit Bridge Test ===")
	fmt.Println()

	fmt.Println("📋 HomeKit Configuration:")
	fmt.Printf("  PIN: %s\n", cfg.Pin)
	fmt.Printf("  Station: %s\n", cfg.StationName)
	fmt.Printf("  Sensors: %s\n", cfg.Sensors)
	fmt.Println()

	// Parse sensor config
	sensorConfig := config.ParseSensorConfig(cfg.Sensors)
	fmt.Println("✅ Sensor Configuration:")
	fmt.Printf("  Temperature: %v\n", sensorConfig.Temperature)
	fmt.Printf("  Humidity: %v\n", sensorConfig.Humidity)
	fmt.Printf("  Light: %v\n", sensorConfig.Light)
	fmt.Printf("  Wind: %v\n", sensorConfig.Wind)
	fmt.Printf("  Rain: %v\n", sensorConfig.Rain)
	fmt.Printf("  Pressure: %v\n", sensorConfig.Pressure)
	fmt.Printf("  UV: %v\n", sensorConfig.UV)
	fmt.Printf("  Lightning: %v\n", sensorConfig.Lightning)
	fmt.Println()

	fmt.Println("🏠 HomeKit Bridge would be created with:")
	fmt.Printf("  Name: Tempest - %s\n", cfg.StationName)
	fmt.Printf("  Manufacturer: WeatherFlow\n")
	fmt.Printf("  Model: Tempest Weather System\n")
	fmt.Printf("  Serial: Tempest-%s\n", cfg.StationName)
	fmt.Println()

	fmt.Println("📱 To pair with HomeKit:")
	fmt.Println("  1. Open Home app on iOS/macOS")
	fmt.Println("  2. Tap '+' to add accessory")
	fmt.Println("  3. Select 'More Options'")
	fmt.Printf("  4. Select 'Tempest - %s'\n", cfg.StationName)
	fmt.Printf("  5. Enter PIN: %s\n", cfg.Pin)
	fmt.Println()

	fmt.Println("✅ HomeKit configuration test completed successfully!")
	fmt.Println("   (Bridge was not actually started - this is a dry run)")
	os.Exit(0)
}

// runWebStatusTest tests web status scraping
func runWebStatusTest(cfg *config.Config) {
	fmt.Println("=== Web Status Scraping Test ===")
	fmt.Println()

	if cfg.Token == "" || cfg.StationName == "" {
		log.Fatal("❌ Token and station name are required for web status testing")
	}

	fmt.Printf("Testing status scraping for station: %s\n\n", cfg.StationName)

	// Note: This would require implementing a scraper test function
	// For now, provide guidance
	fmt.Println("⚠️  Web status scraping test not yet implemented")
	fmt.Println()
	fmt.Println("To test web status scraping:")
	fmt.Println("  1. Ensure Chrome/Chromium is installed")
	fmt.Println("  2. Run the application with --use-web-status flag")
	fmt.Println("  3. Check logs for status scraping activity")
	fmt.Println("  4. Visit http://localhost:8080/api/status")
	fmt.Println()
	fmt.Println("Note: This feature requires headless browser support")
	fmt.Println("See pkg/web/ui_headless_test.go for implementation details")
	os.Exit(0)
}

// runAlarmTest triggers a specific alarm for testing
func runAlarmTest(cfg *config.Config) {
	fmt.Printf("=== Alarm Trigger Test: %s ===\n\n", cfg.TestAlarm)

	if cfg.Alarms == "" {
		log.Fatal("❌ No alarm configuration specified. Use --alarms flag or ALARMS environment variable.")
	}

	// Load alarm configuration
	alarmConfig, err := alarm.LoadAlarmConfig(cfg.Alarms)
	if err != nil {
		log.Fatalf("❌ Failed to load alarm config: %v", err)
	}

	// Find the alarm by name
	var targetAlarm *alarm.Alarm
	for i := range alarmConfig.Alarms {
		if alarmConfig.Alarms[i].Name == cfg.TestAlarm {
			targetAlarm = &alarmConfig.Alarms[i]
			break
		}
	}

	if targetAlarm == nil {
		log.Fatalf("❌ Alarm '%s' not found in configuration", cfg.TestAlarm)
	}

	fmt.Printf("Found alarm: %s\n", targetAlarm.Name)
	fmt.Printf("Description: %s\n", targetAlarm.Description)
	fmt.Printf("Condition: %s\n", targetAlarm.Condition)
	fmt.Printf("Enabled: %v\n", targetAlarm.Enabled)
	fmt.Printf("Channels: %d\n", len(targetAlarm.Channels))
	fmt.Println()

	if !targetAlarm.Enabled {
		log.Fatalf("❌ Alarm '%s' is disabled in configuration", cfg.TestAlarm)
	}

	// Create a test observation that will trigger the alarm
	fmt.Println("Creating test observation to trigger alarm...")
	testObs := weather.Observation{
		Timestamp:            time.Now().Unix(),
		AirTemperature:       25.0, // Default values
		RelativeHumidity:     60.0,
		StationPressure:      1013.25,
		WindAvg:              5.0,
		WindGust:             10.0,
		WindDirection:        180.0,
		UV:                   5,
		Illuminance:          50000.0,
		RainAccumulated:      0.0,
		RainDailyTotal:       0.0,
		LightningStrikeCount: 0,
		LightningStrikeAvg:   0.0,
	}

	// Create alarm manager
	manager, err := alarm.NewManager(cfg.Alarms, cfg.StationName)
	if err != nil {
		log.Fatalf("❌ Failed to create alarm manager: %v", err)
	}

	fmt.Println("Triggering alarm by sending test observation...")
	fmt.Println()

	// Force the alarm to fire by temporarily setting condition to always true
	// This is a test, so we modify the condition
	originalCondition := targetAlarm.Condition
	targetAlarm.Condition = "temperature > 0" // Always true condition

	// Send the observation
	manager.ProcessObservation(&testObs)

	// Restore original condition
	targetAlarm.Condition = originalCondition

	fmt.Println()
	fmt.Println("✅ Alarm test completed!")
	fmt.Println("   Check above output for notification delivery results")
	os.Exit(0)
}

// contains checks if a string slice contains a specific string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// runAPITests performs comprehensive testing of all WeatherFlow API endpoints
// to verify connectivity and data availability before starting the main service.
func runAPITests(cfg *config.Config) {
	fmt.Println("=== WeatherFlow API Endpoint Tests ===")

	// Test 1: Get Stations
	fmt.Println("\n1. Testing Stations API...")
	stations, err := weather.GetStations(cfg.Token)
	if err != nil {
		log.Fatalf("❌ Failed to get stations: %v", err)
	}
	fmt.Printf("✅ Found %d stations\n", len(stations))
	for _, station := range stations {
		fmt.Printf("   - ID: %d, Name: '%s', StationName: '%s'\n",
			station.StationID, station.Name, station.StationName)
	}

	// Test 2: Find and get station details
	fmt.Printf("\n2. Testing Station Details API for '%s'...\n", cfg.StationName)
	station := weather.FindStationByName(stations, cfg.StationName)
	if station == nil {
		log.Fatalf("❌ Station '%s' not found", cfg.StationName)
	}
	fmt.Printf("✅ Found station: %s (ID: %d)\n", station.Name, station.StationID)

	stationDetails, err := weather.GetStationDetails(station.StationID, cfg.Token)
	if err != nil {
		log.Fatalf("❌ Failed to get station details: %v", err)
	}
	fmt.Printf("✅ Station has %d devices\n", len(stationDetails.Devices))
	for i, device := range stationDetails.Devices {
		fmt.Printf("   Device %d: ID=%d, Type=%s, Serial=%s\n",
			i+1, device.DeviceID, device.DeviceType, device.SerialNumber)
	}

	// Test 3: Get Tempest device ID
	fmt.Println("\n3. Testing Tempest Device Discovery...")
	deviceID, err := weather.GetTempestDeviceID(stationDetails)
	if err != nil {
		log.Fatalf("❌ Failed to find Tempest device: %v", err)
	}
	fmt.Printf("✅ Tempest Device ID: %d\n", deviceID)

	// Test 4: Get current observation
	fmt.Println("\n4. Testing Current Observation API...")
	obs, err := weather.GetObservation(station.StationID, cfg.Token)
	if err != nil {
		log.Fatalf("❌ Failed to get current observation: %v", err)
	}
	obsTime := time.Unix(obs.Timestamp, 0)
	fmt.Printf("✅ Current observation retrieved\n")
	fmt.Printf("   - Time: %s\n", obsTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("   - Temperature: %.1f°C\n", obs.AirTemperature)
	fmt.Printf("   - Humidity: %.1f%%\n", obs.RelativeHumidity)
	fmt.Printf("   - Rain: %.3f in\n", obs.RainAccumulated)

	// Test 5: Get historical observations using day_offset
	fmt.Println("\n5. Testing Historical Observations API (day_offset)...")
	startTime := time.Now()
	observations, err := weather.GetHistoricalObservations(station.StationID, cfg.Token, cfg.LogLevel)
	if err != nil {
		log.Fatalf("❌ Failed to get historical observations: %v", err)
	}
	elapsed := time.Since(startTime)

	fmt.Printf("✅ Historical data retrieved in %.2f seconds\n", elapsed.Seconds())
	fmt.Printf("   - Total observations: %d\n", len(observations))

	if len(observations) > 0 {
		oldestObs := time.Unix(observations[len(observations)-1].Timestamp, 0)
		newestObs := time.Unix(observations[0].Timestamp, 0)
		timeSpan := newestObs.Sub(oldestObs)

		fmt.Printf("   - Time span: %.1f hours\n", timeSpan.Hours())
		fmt.Printf("   - Oldest: %s\n", oldestObs.Format("2006-01-02 15:04:05"))
		fmt.Printf("   - Newest: %s\n", newestObs.Format("2006-01-02 15:04:05"))

		// Count observations by day
		todayCount := 0
		yesterdayCount := 0
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		yesterday := today.Add(-24 * time.Hour)

		for _, obs := range observations {
			obsTime := time.Unix(obs.Timestamp, 0).In(now.Location())
			obsDay := time.Date(obsTime.Year(), obsTime.Month(), obsTime.Day(), 0, 0, 0, 0, now.Location())

			if obsDay.Equal(today) {
				todayCount++
			} else if obsDay.Equal(yesterday) {
				yesterdayCount++
			}
		}

		fmt.Printf("   - Today: %d observations\n", todayCount)
		fmt.Printf("   - Yesterday: %d observations\n", yesterdayCount)

		// Show sample observations
		fmt.Printf("   - Sample observations:\n")
		sampleCount := 3
		if len(observations) < sampleCount {
			sampleCount = len(observations)
		}

		for i := 0; i < sampleCount; i++ {
			obs := observations[i]
			obsTime := time.Unix(obs.Timestamp, 0)
			fmt.Printf("     %d. %s: Temp=%.1f°C, Rain=%.3fin\n",
				i+1, obsTime.Format("15:04:05"), obs.AirTemperature, obs.RainAccumulated)
		}
	}

	fmt.Println("\n🎉 All API endpoint tests completed successfully!")
	fmt.Println("\n=== Summary ===")
	fmt.Printf("- Stations API: ✅ Working\n")
	fmt.Printf("- Station Details API: ✅ Working\n")
	fmt.Printf("- Device Discovery: ✅ Working\n")
	fmt.Printf("- Current Observations: ✅ Working\n")
	fmt.Printf("- Historical Observations (day_offset): ✅ Working\n")
	fmt.Printf("- Data Points Retrieved: %d observations\n", len(observations))
	fmt.Printf("- API Performance: %.2f seconds for historical data\n", elapsed.Seconds())
}
