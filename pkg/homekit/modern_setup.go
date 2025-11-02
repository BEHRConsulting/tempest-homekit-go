// Package homekit provides HomeKit accessory setup and management for weather sensors.
// It creates HomeKit-compatible accessories for various weather measurements using the HAP protocol.
package homekit

import (
	"context"
	"os"
	"strings"

	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/logger"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
)

// Custom service for weather sensors that don't interfere with temperature conversion
const TypeWeatherSensor = "F000-0001-1000-8000-0026BB765291"

// WeatherAccessoryModern - Simplified accessory structure using the new hap library
type WeatherAccessoryModern struct {
	AccessoryPtr *accessory.A
	WeatherValue interface{} // Changed to interface{} to support custom characteristics
}

// WeatherSystemModern - New implementation using hap library and custom services
type WeatherSystemModern struct {
	Bridge      *accessory.A
	Server      *hap.Server
	Accessories map[string]*WeatherAccessoryModern
	LogLevel    string
	cancel      context.CancelFunc
}

// NewWeatherSystemModern creates a new weather system using the modern hap library.
// It initializes HomeKit accessories based on the sensor configuration and starts the HAP server.
func NewWeatherSystemModern(pin string, sensorConfig *config.SensorConfig, logLevel string) (*WeatherSystemModern, error) {
	if logLevel == "debug" {
		logger.Debug("Creating new weather system with hap library")
		logger.Debug("Sensor configuration: Temp=%v, Humidity=%v, Light=%v, Wind=%v, Rain=%v, Pressure=%v, UV=%v, Lightning=%v",
			sensorConfig.Temperature, sensorConfig.Humidity, sensorConfig.Light, sensorConfig.Wind,
			sensorConfig.Rain, sensorConfig.Pressure, sensorConfig.UV, sensorConfig.Lightning)
	}

	// Create file storage for HomeKit data
	fs := hap.NewFsStore("./db")

	// Create bridge accessory - this is the main hub
	bridgeInfo := accessory.Info{
		Name:         "Tempest Weather Bridge",
		SerialNumber: "TWB-001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest Bridge v2.0",
		Firmware:     "1.0.0",
	}
	bridge := accessory.NewBridge(bridgeInfo)
	if logLevel == "debug" {
		logger.Debug("Created bridge: %s", bridgeInfo.Name)
	}

	accessories := make(map[string]*WeatherAccessoryModern)
	var hapAccessories []*accessory.A

	// Create standard HomeKit accessories based on sensor configuration
	var accessoryCount int

	// Create separate accessories for each sensor type (more compliant approach)

	// Temperature Sensor Accessory
	if sensorConfig.Temperature {
		tempInfo := accessory.Info{
			Name:         "Temperature Sensor",
			SerialNumber: "TWS-TEMP-001",
			Manufacturer: "WeatherFlow",
			Model:        "Tempest Temperature",
			Firmware:     "1.0.0",
		}
		tempAccessory := accessory.New(tempInfo, accessory.TypeSensor)
		tempService := service.NewTemperatureSensor()
		tempAccessory.AddS(tempService.S)

		hapAccessories = append(hapAccessories, tempAccessory)
		accessories["Air Temperature"] = &WeatherAccessoryModern{
			AccessoryPtr: tempAccessory,
			WeatherValue: tempService.CurrentTemperature.Float,
		}
		accessoryCount++
		if logLevel == "debug" {
			logger.Debug("Created temperature sensor accessory")
		}
	}

	// Humidity Sensor Accessory
	if sensorConfig.Humidity {
		humidityInfo := accessory.Info{
			Name:         "Humidity Sensor",
			SerialNumber: "TWS-HUM-001",
			Manufacturer: "WeatherFlow",
			Model:        "Tempest Humidity",
			Firmware:     "1.0.0",
		}
		humidityAccessory := accessory.New(humidityInfo, accessory.TypeSensor)
		humidityService := service.NewHumiditySensor()
		humidityAccessory.AddS(humidityService.S)

		hapAccessories = append(hapAccessories, humidityAccessory)
		accessories["Relative Humidity"] = &WeatherAccessoryModern{
			AccessoryPtr: humidityAccessory,
			WeatherValue: humidityService.CurrentRelativeHumidity.Float,
		}
		accessoryCount++
		if logLevel == "debug" {
			logger.Debug("Created humidity sensor accessory")
		}
	}

	// Light Sensor Accessory (Lux)
	if sensorConfig.Light {
		lightInfo := accessory.Info{
			Name:         "Light Sensor",
			SerialNumber: "TWS-LUX-001",
			Manufacturer: "WeatherFlow",
			Model:        "Tempest Light",
			Firmware:     "1.0.0",
		}
		lightAccessory := accessory.New(lightInfo, accessory.TypeSensor)
		lightService := service.NewLightSensor()
		lightAccessory.AddS(lightService.S)

		hapAccessories = append(hapAccessories, lightAccessory)
		accessories["Ambient Light"] = &WeatherAccessoryModern{
			AccessoryPtr: lightAccessory,
			WeatherValue: lightService.CurrentAmbientLightLevel.Float,
		}
		accessoryCount++
		if logLevel == "debug" {
			logger.Debug("Created light sensor accessory")
		}
	}

	// UV Sensor Accessory
	if sensorConfig.UV {
		uvInfo := accessory.Info{
			Name:         "UV Index Sensor",
			SerialNumber: "TWS-UV-001",
			Manufacturer: "WeatherFlow",
			Model:        "Tempest UV",
			Firmware:     "1.0.0",
		}
		uvAccessory := accessory.New(uvInfo, accessory.TypeSensor)

		// Use Light Sensor service for UV with proper UV Index range
		uvService := service.NewLightSensor()
		uvService.CurrentAmbientLightLevel.Description = "UV Index"
		uvService.CurrentAmbientLightLevel.Unit = "UV Index"
		uvService.CurrentAmbientLightLevel.SetMinValue(0.0)
		uvService.CurrentAmbientLightLevel.SetMaxValue(15.0)
		uvService.CurrentAmbientLightLevel.SetStepValue(0.1)
		uvService.CurrentAmbientLightLevel.SetValue(0.0)
		uvAccessory.AddS(uvService.S)

		hapAccessories = append(hapAccessories, uvAccessory)
		accessories["UV Index"] = &WeatherAccessoryModern{
			AccessoryPtr: uvAccessory,
			WeatherValue: uvService.CurrentAmbientLightLevel.Float,
		}
		accessoryCount++
		if logLevel == "debug" {
			logger.Debug("Created UV Index sensor accessory using light sensor service with UV range")
		}
	}

	// Pressure Sensor Accessory (using standard light sensor service with custom labels)
	if sensorConfig.Pressure {
		pressureInfo := accessory.Info{
			Name:         "Pressure Sensor",
			SerialNumber: "TWS-PRESS-001",
			Manufacturer: "WeatherFlow",
			Model:        "Tempest Pressure",
			Firmware:     "1.0.0",
		}
		pressureAccessory := accessory.New(pressureInfo, accessory.TypeSensor)

		// Use standard Light Sensor service but customize for pressure
		pressureService := service.NewLightSensor()
		pressureService.CurrentAmbientLightLevel.Description = "Atmospheric Pressure (mb)"
		pressureService.CurrentAmbientLightLevel.Unit = "mb"
		pressureService.CurrentAmbientLightLevel.SetMinValue(700.0)
		pressureService.CurrentAmbientLightLevel.SetMaxValue(1200.0)
		pressureService.CurrentAmbientLightLevel.SetStepValue(0.1) // Decimal precision
		pressureService.CurrentAmbientLightLevel.SetValue(1013.25) // Standard atmospheric pressure

		pressureAccessory.AddS(pressureService.S)

		hapAccessories = append(hapAccessories, pressureAccessory)
		accessories["Atmospheric Pressure"] = &WeatherAccessoryModern{
			AccessoryPtr: pressureAccessory,
			WeatherValue: pressureService.CurrentAmbientLightLevel.Float,
		}
		accessoryCount++
		if logLevel == "debug" {
			logger.Debug("Created atmospheric pressure sensor accessory using standard light sensor service with custom labels")
		}
	}

	// Store all other sensors as null references to maintain API compatibility
	allSensorNames := []string{
		"Wind Speed", "Wind Gust", "Wind Direction", "Rain Accumulation",
		"Lightning Count", "Lightning Distance", "Precipitation Type",
	}
	// Add the configured sensors to null list if not enabled
	if !sensorConfig.Temperature {
		allSensorNames = append(allSensorNames, "Air Temperature")
	}
	if !sensorConfig.Humidity {
		allSensorNames = append(allSensorNames, "Relative Humidity")
	}
	if !sensorConfig.Light {
		allSensorNames = append(allSensorNames, "Ambient Light")
	}

	for _, name := range allSensorNames {
		if _, exists := accessories[name]; !exists {
			accessories[name] = &WeatherAccessoryModern{
				AccessoryPtr: nil, // Will be ignored in updates
				WeatherValue: nil,
			}
		}
	}

	// Create the HAP server with configured accessories
	if logLevel == "debug" {
		logger.Debug("Creating server with %d accessories based on sensor configuration", len(hapAccessories))
	}
	server, err := hap.NewServer(fs, bridge.A, hapAccessories...)
	if err != nil {
		return nil, err
	}

	// Set the PIN for pairing
	server.Pin = pin

	if logLevel == "debug" {
		logger.Debug("Weather system created successfully with PIN: %s", pin)
		logger.Debug("HomeKit compliance: %d accessories created based on sensor configuration", accessoryCount)
		logger.Debug("Sensors enabled: Temp=%v, Humidity=%v, Light=%v, UV=%v, Pressure=%v", sensorConfig.Temperature, sensorConfig.Humidity, sensorConfig.Light, sensorConfig.UV, sensorConfig.Pressure)
	}

	return &WeatherSystemModern{
		Bridge:      bridge.A,
		Server:      server,
		Accessories: accessories,
		LogLevel:    logLevel,
	}, nil
}

// Start the weather system with graceful shutdown
func (ws *WeatherSystemModern) Start() error {
	if ws.LogLevel == "debug" {
		logger.Debug("Starting weather system server")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	ws.cancel = cancel

	// Start the server in background
	go func() {
		if ws.LogLevel == "debug" {
			logger.Debug("HomeKit server starting with PIN: %s", ws.Server.Pin)
		}
		if err := ws.Server.ListenAndServe(ctx); err != nil {
			logger.Error("HomeKit server error: %v", err)
		}
	}()

	return nil
}

// Stop the weather system gracefully
func (ws *WeatherSystemModern) Stop() {
	if ws.LogLevel == "debug" {
		logger.Debug("Stopping weather system server")
	}
	if ws.cancel != nil {
		ws.cancel()
	}
}

// UpdateSensor updates a specific sensor value
func (ws *WeatherSystemModern) UpdateSensor(sensorName string, value float64) {
	if accessory, exists := ws.Accessories[sensorName]; exists {
		// Check if this sensor has a valid characteristic (some are intentionally nil for compatibility)
		if accessory.WeatherValue != nil {
			if ws.LogLevel == "debug" {
				logger.Debug("Updating %s: %.3f", sensorName, value)
			}
			// Handle different characteristic types
			switch v := accessory.WeatherValue.(type) {
			case *characteristic.Float:
				v.SetValue(value)
			default:
				logger.Warn("Unsupported characteristic type for sensor %s", sensorName)
			}
		} else {
			if ws.LogLevel == "debug" {
				logger.Debug("Skipping %s (not included in minimal setup)", sensorName)
			}
		}
	} else {
		logger.Warn("Sensor %s not found", sensorName)
	}
} // GetAvailableSensors returns the list of available sensor names
func (ws *WeatherSystemModern) GetAvailableSensors() []string {
	sensors := make([]string, 0, len(ws.Accessories))
	for name := range ws.Accessories {
		sensors = append(sensors, name)
	}
	return sensors
}

// countPairedDevices counts the number of paired devices by reading pairing files from the database
func countPairedDevices() int {
	dbDir := "./db"
	entries, err := os.ReadDir(dbDir)
	if err != nil {
		logger.Warn("Failed to read database directory for paired devices count: %v", err)
		return 0
	}

	count := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".pairing") {
			count++
		}
	}

	if count > 0 {
		logger.Debug("Found %d paired device(s) in database", count)
	}

	return count
}

// GetDetailedInfo returns detailed HomeKit bridge information
func (ws *WeatherSystemModern) GetDetailedInfo() map[string]interface{} {
	if ws.Bridge == nil || ws.Server == nil {
		return map[string]interface{}{
			"bridge": false,
		}
	}

	info := map[string]interface{}{
		"bridge":         true,
		"name":           ws.Bridge.Info.Name.Value(),
		"bridgeId":       ws.Bridge.Info.SerialNumber.Value(),
		"category":       "Bridge",
		"pin":            ws.Server.Pin,
		"setupCode":      "X-" + ws.Server.Pin,
		"port":           "51826", // Standard HAP port
		"hapVersion":     "1.1",   // HAP protocol version
		"accessories":    len(ws.Accessories),
		"accessoryNames": ws.GetAvailableSensors(),
		"manufacturer":   ws.Bridge.Info.Manufacturer.Value(),
		"model":          ws.Bridge.Info.Model.Value(),
		"firmware":       ws.Bridge.Info.FirmwareRevision.Value(),
	}

	// Get paired devices count by reading database files
	pairedCount := countPairedDevices()
	info["pairedDevices"] = pairedCount
	info["reachability"] = true
	info["lastRequest"] = "Active"

	// Configuration number increments with accessory changes
	info["configNumber"] = len(ws.Accessories)

	return info
}
