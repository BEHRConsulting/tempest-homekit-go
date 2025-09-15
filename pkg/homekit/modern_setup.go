package homekit

import (
	"context"
	"log"

	"tempest-homekit-go/pkg/config"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
)

// Custom service for weather sensors that don't interfere with temperature conversion
const TypeWeatherSensor = "F000-0001-1000-8000-0026BB765291"

// WeatherService - Custom service for weather data without temperature conversion issues
type WeatherService struct {
	*service.S
	WeatherValue *characteristic.Float
}

func NewWeatherService(serviceType, characteristicType string) *WeatherService {
	s := service.New(serviceType)

	// Create a custom float characteristic that won't be treated as temperature
	weatherValue := characteristic.NewFloat(characteristicType)
	weatherValue.Format = characteristic.FormatFloat
	weatherValue.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
	weatherValue.SetValue(0.0)
	s.AddC(weatherValue.C)

	return &WeatherService{
		S:            s,
		WeatherValue: weatherValue,
	}
}

// WeatherAccessoryModern - Simplified accessory structure using the new hap library
type WeatherAccessoryModern struct {
	AccessoryPtr *accessory.A
	WeatherValue *characteristic.Float
}

// WeatherSystemModern - New implementation using hap library and custom services
type WeatherSystemModern struct {
	Bridge      *accessory.A
	Server      *hap.Server
	Accessories map[string]*WeatherAccessoryModern
	LogLevel    string
	cancel      context.CancelFunc
}

// NewWeatherSystemModern creates a new weather system using the modern hap library
func NewWeatherSystemModern(pin string, sensorConfig *config.SensorConfig, logLevel string) (*WeatherSystemModern, error) {
	if logLevel == "debug" {
		log.Printf("DEBUG: Creating new weather system with hap library")
		log.Printf("DEBUG: Sensor configuration: Temp=%v, Humidity=%v, Light=%v, Wind=%v, Rain=%v, Pressure=%v, UV=%v, Lightning=%v",
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
		log.Printf("DEBUG: Created bridge: %s", bridgeInfo.Name)
	}

	accessories := make(map[string]*WeatherAccessoryModern)
	var hapAccessories []*accessory.A

	// Create standard HomeKit accessories based on sensor configuration
	var accessoryCount int

	// 1. Temperature Sensor (if enabled)
	if sensorConfig.Temperature {
		tempInfo := accessory.Info{
			Name:         "Air Temperature",
			SerialNumber: "AT-001",
			Manufacturer: "WeatherFlow",
			Model:        "Tempest",
			Firmware:     "1.0.0",
		}
		tempAccessory := accessory.NewTemperatureSensor(tempInfo)
		accessories["Air Temperature"] = &WeatherAccessoryModern{
			AccessoryPtr: tempAccessory.A,
			WeatherValue: tempAccessory.TempSensor.CurrentTemperature.Float,
		}
		hapAccessories = append(hapAccessories, tempAccessory.A)
		accessoryCount++
		if logLevel == "debug" {
			log.Printf("DEBUG: Created standard temperature sensor")
		}
	}

	// 2. Humidity Sensor (if enabled) - using custom service since hap doesn't have HumiditySensor
	if sensorConfig.Humidity {
		humidityInfo := accessory.Info{
			Name:         "Relative Humidity",
			SerialNumber: "RH-001",
			Manufacturer: "WeatherFlow",
			Model:        "Tempest",
			Firmware:     "1.0.0",
		}
		humidityAccessory := accessory.New(humidityInfo, accessory.TypeSensor)

		// Add custom humidity service
		humidityService := service.New("F180-0001-1000-8000-0026BB765291")
		humidityChar := characteristic.NewFloat("F181-0001-1000-8000-0026BB765291")
		humidityChar.Format = characteristic.FormatFloat
		humidityChar.Unit = "percentage"
		humidityChar.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
		humidityChar.SetMinValue(0.0)
		humidityChar.SetMaxValue(100.0)
		humidityChar.SetStepValue(0.1)
		humidityChar.SetValue(0.0)
		humidityService.AddC(humidityChar.C)
		humidityAccessory.AddS(humidityService)

		accessories["Relative Humidity"] = &WeatherAccessoryModern{
			AccessoryPtr: humidityAccessory,
			WeatherValue: humidityChar,
		}
		hapAccessories = append(hapAccessories, humidityAccessory)
		accessoryCount++
		if logLevel == "debug" {
			log.Printf("DEBUG: Created humidity sensor")
		}
	}

	// 3. Light Sensor (if enabled) - using custom service since hap doesn't have LightSensor
	if sensorConfig.Light {
		lightInfo := accessory.Info{
			Name:         "Ambient Light",
			SerialNumber: "AL-001",
			Manufacturer: "WeatherFlow",
			Model:        "Tempest",
			Firmware:     "1.0.0",
		}
		lightAccessory := accessory.New(lightInfo, accessory.TypeSensor)

		// Add custom light service
		lightService := service.New("F190-0001-1000-8000-0026BB765291")
		lightChar := characteristic.NewFloat("F191-0001-1000-8000-0026BB765291")
		lightChar.Format = characteristic.FormatFloat
		lightChar.Unit = "lux"
		lightChar.Permissions = []string{characteristic.PermissionRead, characteristic.PermissionEvents}
		lightChar.SetMinValue(0.0001)
		lightChar.SetMaxValue(100000.0)
		lightChar.SetStepValue(0.1)
		lightChar.SetValue(0.0)
		lightService.AddC(lightChar.C)
		lightAccessory.AddS(lightService)

		accessories["Ambient Light"] = &WeatherAccessoryModern{
			AccessoryPtr: lightAccessory,
			WeatherValue: lightChar,
		}
		hapAccessories = append(hapAccessories, lightAccessory)
		accessoryCount++
		if logLevel == "debug" {
			log.Printf("DEBUG: Created light sensor")
		}
	}

	// Store all other sensors as null references to maintain API compatibility
	allSensorNames := []string{
		"Wind Speed", "Wind Gust", "Wind Direction", "Rain Accumulation", "UV Index",
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
		log.Printf("DEBUG: Creating server with %d accessories based on sensor configuration", len(hapAccessories))
	}
	server, err := hap.NewServer(fs, bridge.A, hapAccessories...)
	if err != nil {
		return nil, err
	}

	// Set the PIN for pairing
	server.Pin = pin

	if logLevel == "debug" {
		log.Printf("DEBUG: Weather system created successfully with PIN: %s", pin)
		log.Printf("DEBUG: HomeKit compliance: %d accessories created based on sensor configuration", accessoryCount)
		log.Printf("DEBUG: Sensors enabled: Temp=%v, Humidity=%v, Light=%v", sensorConfig.Temperature, sensorConfig.Humidity, sensorConfig.Light)
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
		log.Printf("DEBUG: Starting weather system server")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	ws.cancel = cancel

	// Start the server in background
	go func() {
		if ws.LogLevel == "debug" {
			log.Printf("DEBUG: HomeKit server starting with PIN: %s", ws.Server.Pin)
		}
		if err := ws.Server.ListenAndServe(ctx); err != nil {
			log.Printf("HomeKit server error: %v", err)
		}
	}()

	return nil
}

// Stop the weather system gracefully
func (ws *WeatherSystemModern) Stop() {
	if ws.LogLevel == "debug" {
		log.Printf("DEBUG: Stopping weather system server")
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
				log.Printf("DEBUG: Updating %s: %.3f", sensorName, value)
			}
			accessory.WeatherValue.SetValue(value)
		} else {
			if ws.LogLevel == "debug" {
				log.Printf("DEBUG: Skipping %s (not included in minimal setup)", sensorName)
			}
		}
	} else {
		log.Printf("WARNING: Sensor %s not found", sensorName)
	}
}

// GetAvailableSensors returns the list of available sensor names
func (ws *WeatherSystemModern) GetAvailableSensors() []string {
	sensors := make([]string, 0, len(ws.Accessories))
	for name := range ws.Accessories {
		sensors = append(sensors, name)
	}
	return sensors
}
