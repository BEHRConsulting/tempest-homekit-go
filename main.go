// Package main implements a HomeKit bridge for WeatherFlow Tempest weather stations.
// It provides a HomeKit-compatible interface to access weather data from Tempest stations,
// along with a web dashboard for monitoring and configuration.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/service"
	"tempest-homekit-go/pkg/weather"
)

func main() {
	cfg := config.LoadConfig()

	// Set up logging first (before any other operations that might log)
	logger.SetLogLevel(cfg.LogLevel)

	// Note: For generated weather, elevation will be logged by the service once location is selected

	// Handle version flag
	if cfg.Version {
		fmt.Println("tempest-homekit-go v1.5.0")
		fmt.Println("Built with Go 1.24.2")
		fmt.Println("HomeKit integration for WeatherFlow Tempest weather stations")
		os.Exit(0)
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
	err := service.StartService(cfg, "1.5.0")
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}

	logger.Info("Service started successfully, waiting for interrupt signal...") // Wait for interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	logger.Info("Received signal %v, shutting down...", sig)
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
