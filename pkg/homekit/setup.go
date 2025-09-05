package homekit

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type WeatherAccessories struct {
	TemperatureAccessory *accessory.Accessory
	HumidityAccessory    *accessory.Accessory
	WindAccessory        *accessory.Accessory
	RainAccessory        *accessory.Accessory
	TemperatureSensor    *service.TemperatureSensor
	HumiditySensor       *service.HumiditySensor
	WindSensor           *service.Fan
	RainSensor           *service.HumiditySensor // Using humidity sensor for rain accumulation
}

func NewWeatherAccessories() *WeatherAccessories {
	tempInfo := accessory.Info{
		Name:         "Tempest Temperature",
		SerialNumber: "TEMP001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	tempAcc := accessory.New(tempInfo, accessory.TypeOther)
	tempSensor := service.NewTemperatureSensor()
	tempAcc.AddService(tempSensor.Service)

	humInfo := accessory.Info{
		Name:         "Tempest Humidity",
		SerialNumber: "HUM001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	humAcc := accessory.New(humInfo, accessory.TypeOther)
	humSensor := service.NewHumiditySensor()
	humAcc.AddService(humSensor.Service)

	windInfo := accessory.Info{
		Name:         "Tempest Wind Speed",
		SerialNumber: "WIND001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	windAcc := accessory.New(windInfo, accessory.TypeOther)
	windSensor := service.NewFan()
	windAcc.AddService(windSensor.Service)

	rainInfo := accessory.Info{
		Name:         "Tempest Rain",
		SerialNumber: "RAIN001",
		Manufacturer: "WeatherFlow",
		Model:        "Tempest",
	}
	rainAcc := accessory.New(rainInfo, accessory.TypeOther)
	rainSensor := service.NewHumiditySensor() // Using humidity sensor for rain accumulation
	rainAcc.AddService(rainSensor.Service)

	return &WeatherAccessories{
		TemperatureAccessory: tempAcc,
		HumidityAccessory:    humAcc,
		WindAccessory:        windAcc,
		RainAccessory:        rainAcc,
		TemperatureSensor:    tempSensor,
		HumiditySensor:       humSensor,
		WindSensor:           windSensor,
		RainSensor:           rainSensor,
	}
}

func (wa *WeatherAccessories) UpdateTemperature(temp float64) {
	wa.TemperatureSensor.CurrentTemperature.SetValue(temp)
}

func (wa *WeatherAccessories) UpdateHumidity(hum float64) {
	wa.HumiditySensor.CurrentRelativeHumidity.SetValue(hum)
}

func (wa *WeatherAccessories) UpdateWindSpeed(speed float64) {
	// Use fan On state to represent wind presence and speed via metadata
	wa.WindSensor.On.SetValue(speed > 0)
}

func (wa *WeatherAccessories) UpdateRainAccumulation(rain float64) {
	// Use humidity sensor to represent rain accumulation
	// Scale rain accumulation to 0-100 range for display
	rainLevel := rain * 100 // Convert inches to percentage
	if rainLevel > 100 {
		rainLevel = 100
	}
	wa.RainSensor.CurrentRelativeHumidity.SetValue(rainLevel)
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

	t, err := hc.NewIPTransport(config, bridge.Accessory, wa.TemperatureAccessory, wa.HumidityAccessory, wa.WindAccessory, wa.RainAccessory)
	if err != nil {
		return nil, err
	}

	return t, nil
}
