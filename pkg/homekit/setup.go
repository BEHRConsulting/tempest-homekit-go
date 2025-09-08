package homekit

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

// Custom service types for weather sensors
type WindSensor struct {
	*service.TemperatureSensor
}

type WindDirectionSensor struct {
	*service.TemperatureSensor
}

type RainSensor struct {
	*service.TemperatureSensor
}

// Constructor functions for custom service types
func NewWindSensor() *WindSensor {
	return &WindSensor{
		TemperatureSensor: service.NewTemperatureSensor(),
	}
}

func NewWindDirectionSensor() *WindDirectionSensor {
	return &WindDirectionSensor{
		TemperatureSensor: service.NewTemperatureSensor(),
	}
}

func NewRainSensor() *RainSensor {
	return &RainSensor{
		TemperatureSensor: service.NewTemperatureSensor(),
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
	tempAcc := accessory.New(tempInfo, accessory.TypeSensor)
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
	windAcc := accessory.New(windInfo, accessory.TypeSensor) // Using sensor accessory type for wind speed
	windSensor := NewWindSensor()                            // Use custom wind sensor
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
	windDirAcc := accessory.New(windDirInfo, accessory.TypeSensor) // Using sensor accessory type for wind direction
	windDirSensor := NewWindDirectionSensor()                      // Use custom wind direction sensor
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
	rainAcc := accessory.New(rainInfo, accessory.TypeSensor) // Using sensor accessory type for rain
	rainSensor := NewRainSensor()                            // Using custom rain sensor
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
	// Use temperature sensor to display wind speed as a numeric value
	// Wind speed ranges from 0.00 to 500.00 mph
	wa.WindSensor.CurrentTemperature.SetValue(speed)
}

func (wa *WeatherAccessories) UpdateWindDirection(direction float64) {
	// Use temperature sensor to display wind direction as degrees (0-360)
	wa.WindDirectionSensor.CurrentTemperature.SetValue(direction)
}

func (wa *WeatherAccessories) UpdateRainAccumulation(rain float64) {
	// Use temperature sensor to display rain accumulation as inches (0.00 to 1000.00)
	wa.RainSensor.CurrentTemperature.SetValue(rain)
}

func (wa *WeatherAccessories) UpdateIlluminance(illuminance float64) {
	// Update light sensor with illuminance data
	wa.LightSensor.CurrentAmbientLightLevel.SetValue(illuminance)
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
