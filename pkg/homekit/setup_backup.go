package homekit

import "tempest-homekit-go/pkg/config"

// WeatherAccessories is maintained for compatibility
// This will delegate to the modern implementation
type WeatherAccessories struct {
	modern *WeatherSystemModern
}

func NewWeatherAccessories() *WeatherAccessories {
	// Don't create the system here - wait for SetupHomeKit with proper PIN
	return &WeatherAccessories{
		modern: nil,
	}
}

// Legacy compatibility methods that delegate to modern implementation
func (wa *WeatherAccessories) UpdateWindAverage(speed float64) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("Wind Speed", speed)
	}
}

func (wa *WeatherAccessories) UpdateWindGust(speed float64) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("Wind Gust", speed)
	}
}

func (wa *WeatherAccessories) UpdateWindDirection(degrees float64) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("Wind Direction", degrees)
	}
}

func (wa *WeatherAccessories) UpdateAirTemperature(temp float64) {
	// Temperature uses regular sensor with conversion, so it remains accurate
	if wa.modern != nil {
		wa.modern.UpdateSensor("Air Temperature", temp)
	}
}

func (wa *WeatherAccessories) UpdateRelativeHumidity(humidity float64) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("Relative Humidity", humidity)
	}
}

func (wa *WeatherAccessories) UpdateLux(lux float64) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("Ambient Light", lux)
	}
}

func (wa *WeatherAccessories) UpdateUV(uv float64) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("UV Index", uv)
	}
}

func (wa *WeatherAccessories) UpdateRain(rain float64) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("Rain Accumulation", rain)
	}
}

func (wa *WeatherAccessories) UpdatePrecipitationType(precipType int) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("Precipitation Type", float64(precipType))
	}
}

func (wa *WeatherAccessories) UpdateLightningCount(count int) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("Lightning Count", float64(count))
	}
}

func (wa *WeatherAccessories) UpdateLightningDistance(distance float64) {
	if wa.modern != nil {
		wa.modern.UpdateSensor("Lightning Distance", distance)
	}
}

func degreesToCompass(degrees float64) string {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	index := int((degrees+11.25)/22.5) % 16
	return directions[index]
}

// SetupHomeKit creates and starts the HomeKit server using modern hap library
func SetupHomeKit(wa *WeatherAccessories, pin string, sensorConfig *config.SensorConfig) (*WeatherSystemModern, error) {
	// Create a new system with the correct PIN and sensor configuration
	modern, err := NewWeatherSystemModern(pin, sensorConfig)
	if err != nil {
		return nil, err
	}

	// Replace the default system with the properly configured one
	wa.modern = modern

	// Start the server
	err = wa.modern.Start()
	if err != nil {
		return nil, err
	}

	return wa.modern, nil
}
