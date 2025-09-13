package homekit

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

// Custom service types for weather sensors
type WindSensor struct {
	*service.Service
	WindSpeed *characteristic.CurrentTemperature
}

type WindDirectionSensor struct {
	*service.Service
	WindDirection *characteristic.CurrentTemperature
}

type RainSensor struct {
	*service.Service
	RainAccumulation *characteristic.CurrentTemperature
}

// Constructor functions for custom service types
func NewWindSensor() *WindSensor {
	svc := service.New(service.TypeThermostat) // Use thermostat service type as base for wind speed
	windSpeed := characteristic.NewCurrentTemperature()
	windSpeed.SetValue(0.0)
	windSpeed.SetMinValue(0.0)
	windSpeed.SetMaxValue(500.0)
	windSpeed.SetStepValue(0.1)

	svc.AddCharacteristic(windSpeed.Characteristic)

	return &WindSensor{
		Service:   svc,
		WindSpeed: windSpeed,
	}
}

func NewWindDirectionSensor() *WindDirectionSensor {
	svc := service.New(service.TypeThermostat) // Use thermostat service type as base for wind direction
	windDirection := characteristic.NewCurrentTemperature()
	windDirection.SetValue(0.0)
	windDirection.SetMinValue(0.0)
	windDirection.SetMaxValue(360.0)
	windDirection.SetStepValue(1.0)

	svc.AddCharacteristic(windDirection.Characteristic)

	return &WindDirectionSensor{
		Service:       svc,
		WindDirection: windDirection,
	}
}

func NewRainSensor() *RainSensor {
	svc := service.New(service.TypeIrrigationSystem) // Use irrigation system service type for rain
	rainAccumulation := characteristic.NewCurrentTemperature()
	rainAccumulation.SetValue(0.0)
	rainAccumulation.SetMinValue(0.0)
	rainAccumulation.SetMaxValue(100.0)
	rainAccumulation.SetStepValue(0.1)

	svc.AddCharacteristic(rainAccumulation.Characteristic)

	return &RainSensor{
		Service:          svc,
		RainAccumulation: rainAccumulation,
	}
}

type WeatherAccessories struct {
	TemperatureAccessory   *accessory.Accessory
	HumidityAccessory      *accessory.Accessory
	WindAccessory          *accessory.Accessory
	WindDirectionAccessory *accessory.Accessory
	RainAccessory          *accessory.Accessory
	LightAccessory         *accessory.Accessory
	TemperatureSensor      *service.TemperatureSensor
	HumiditySensor         *service.HumiditySensor
	WindSensor             *WindSensor
	WindDirectionSensor    *WindDirectionSensor
	RainSensor             *RainSensor
	LightSensor            *service.LightSensor
}

func NewWeatherAccessories() *WeatherAccessories {
	tempInfo := accessory.Info{
		Name:         "Temperature Sensor",
		SerialNumber: "TEMP001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	// Use Thermometer accessory type for temperature
	tempAcc := accessory.New(tempInfo, accessory.TypeThermostat)
	tempSensor := service.NewTemperatureSensor()
	nameChar := characteristic.NewName()
	nameChar.SetValue("Temperature Sensor")
	tempSensor.Service.AddCharacteristic(nameChar.Characteristic)
	tempAcc.AddService(tempSensor.Service)

	humInfo := accessory.Info{
		Name:         "Humidity Sensor",
		SerialNumber: "HUM001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	humAcc := accessory.New(humInfo, accessory.TypeSensor)
	humSensor := service.NewHumiditySensor()
	humNameChar := characteristic.NewName()
	humNameChar.SetValue("Humidity Sensor")
	humSensor.Service.AddCharacteristic(humNameChar.Characteristic)
	humAcc.AddService(humSensor.Service)

	windInfo := accessory.Info{
		Name:         "Wind Speed Sensor",
		SerialNumber: "WIND001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	// Use Sensor accessory type for wind speed
	windAcc := accessory.New(windInfo, accessory.TypeSensor)
	windSensor := NewWindSensor()
	windNameChar := characteristic.NewName()
	windNameChar.SetValue("Wind Speed Sensor")
	windSensor.Service.AddCharacteristic(windNameChar.Characteristic)
	windAcc.AddService(windSensor.Service)

	windDirInfo := accessory.Info{
		Name:         "Wind Direction Sensor",
		SerialNumber: "WINDDIR001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	// Use Sensor accessory type for wind direction
	windDirAcc := accessory.New(windDirInfo, accessory.TypeSensor)
	windDirSensor := NewWindDirectionSensor()
	windDirNameChar := characteristic.NewName()
	windDirNameChar.SetValue("Wind Direction Sensor")
	windDirSensor.Service.AddCharacteristic(windDirNameChar.Characteristic)
	windDirAcc.AddService(windDirSensor.Service)

	rainInfo := accessory.Info{
		Name:         "Rain Sensor",
		SerialNumber: "RAIN001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	// Use Sprinklers accessory type for rain
	rainAcc := accessory.New(rainInfo, accessory.TypeSprinklers)
	rainSensor := NewRainSensor()
	rainNameChar := characteristic.NewName()
	rainNameChar.SetValue("Rain Sensor")
	rainSensor.Service.AddCharacteristic(rainNameChar.Characteristic)
	rainAcc.AddService(rainSensor.Service)

	lightInfo := accessory.Info{
		Name:         "Light Sensor",
		SerialNumber: "LIGHT001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	lightAcc := accessory.New(lightInfo, accessory.TypeSensor)
	lightSensor := service.NewLightSensor()
	lightNameChar := characteristic.NewName()
	lightNameChar.SetValue("Light Sensor")
	lightSensor.Service.AddCharacteristic(lightNameChar.Characteristic)
	lightAcc.AddService(lightSensor.Service)

	return &WeatherAccessories{
		TemperatureAccessory:   tempAcc,
		HumidityAccessory:      humAcc,
		WindAccessory:          windAcc,
		WindDirectionAccessory: windDirAcc,
		RainAccessory:          rainAcc,
		LightAccessory:         lightAcc,
		TemperatureSensor:      tempSensor,
		HumiditySensor:         humSensor,
		WindSensor:             windSensor,
		WindDirectionSensor:    windDirSensor,
		RainSensor:             rainSensor,
		LightSensor:            lightSensor,
	}
}

func (wa *WeatherAccessories) UpdateTemperature(temp float64) {
	wa.TemperatureSensor.CurrentTemperature.SetValue(temp)
}

func (wa *WeatherAccessories) UpdateHumidity(hum float64) {
	wa.HumiditySensor.CurrentRelativeHumidity.SetValue(hum)
}

func (wa *WeatherAccessories) UpdateWindSpeed(speed float64) {
	// Use custom wind speed sensor with temperature characteristic for display
	wa.WindSensor.WindSpeed.SetValue(speed)
}

func (wa *WeatherAccessories) UpdateWindDirection(direction float64) {
	// Use custom wind direction sensor with temperature characteristic for display
	wa.WindDirectionSensor.WindDirection.SetValue(direction)
}

func (wa *WeatherAccessories) UpdateRainAccumulation(rain float64) {
	// Use custom rain sensor - scale rain accumulation to temperature range
	// Map 0-1000 inches to 0-100 temperature units
	tempValue := (rain / 1000.0) * 100.0
	if tempValue > 100.0 {
		tempValue = 100.0
	}
	wa.RainSensor.RainAccumulation.SetValue(tempValue)
}

func (wa *WeatherAccessories) UpdateIlluminance(illuminance float64) {
	// Update light sensor with illuminance data
	wa.LightSensor.CurrentAmbientLightLevel.SetValue(illuminance)
}

func degreesToCompass(degrees float64) string {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	index := int((degrees+11.25)/22.5) % 16
	return directions[index]
}

func SetupHomeKit(wa *WeatherAccessories, pin string) (hc.Transport, error) {
	config := hc.Config{
		Pin:         pin,
		StoragePath: "./db",
	}

	bridgeInfo := accessory.Info{
		Name:         "Tempest HomeKit Bridge",
		SerialNumber: "BRIDGE001",
		Manufacturer: "Custom",
		Model:        "Bridge",
	}
	bridge := accessory.NewBridge(bridgeInfo)

	t, err := hc.NewIPTransport(config, bridge.Accessory, wa.TemperatureAccessory, wa.HumidityAccessory, wa.WindAccessory, wa.WindDirectionAccessory, wa.RainAccessory, wa.LightAccessory)
	if err != nil {
		return nil, err
	}

	return t, nil
}
