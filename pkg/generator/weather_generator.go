// Package generator provides synthetic weather data generation for UI testing.
// It creates realistic weather patterns based on different seasons and locations.
package generator

import (
	"math"
	"math/rand"
	"time"

	"tempest-homekit-go/pkg/types"
)

// Season represents different weather seasons
type Season int

const (
	Spring Season = iota
	Summer
	Fall
	Winter
)

func (s Season) String() string {
	seasons := []string{"Spring", "Summer", "Fall", "Winter"}
	return seasons[s]
}

// Location represents different climate locations
type Location struct {
	Name        string
	Latitude    float64
	Longitude   float64
	Elevation   float64
	ClimateZone string
}

// WeatherGenerator generates synthetic weather data
type WeatherGenerator struct {
	Location               Location
	Season                 Season
	CurrentTime            time.Time
	BaseTemperature        float64 // Celsius
	BasePressure           float64 // mb
	BaseHumidity           float64 // %
	current                *types.Observation
	history                []*types.Observation
	rng                    *rand.Rand
	cumulativeRain         float64 // Total accumulated rain since station start (like real Tempest)
	dailyRainTotal         float64 // Total rain for the current day (resets at midnight)
	lastDayCheck           int     // Day of year for checking when to reset daily total
	isGeneratingHistorical bool    // Flag to prevent historical generation from affecting daily totals
}

// Predefined locations with different climates
var Locations = []Location{
	{
		Name: "Miami, FL", Latitude: 25.7617, Longitude: -80.1918, Elevation: 2.0,
		ClimateZone: "Tropical",
	},
	{
		Name: "Denver, CO", Latitude: 39.7392, Longitude: -104.9903, Elevation: 1609.0,
		ClimateZone: "Continental",
	},
	{
		Name: "Seattle, WA", Latitude: 47.6062, Longitude: -122.3321, Elevation: 56.0,
		ClimateZone: "Oceanic",
	},
	{
		Name: "Phoenix, AZ", Latitude: 33.4484, Longitude: -112.0740, Elevation: 331.0,
		ClimateZone: "Desert",
	},
	{
		Name: "Minneapolis, MN", Latitude: 44.9778, Longitude: -93.2650, Elevation: 264.0,
		ClimateZone: "Continental",
	},
	{
		Name: "San Diego, CA", Latitude: 32.7157, Longitude: -117.1611, Elevation: 19.0,
		ClimateZone: "Mediterranean",
	},
	{
		Name: "Anchorage, AK", Latitude: 61.2181, Longitude: -149.9003, Elevation: 35.0,
		ClimateZone: "Subarctic",
	},
	{
		Name: "New Orleans, LA", Latitude: 29.9511, Longitude: -90.0715, Elevation: -0.5,
		ClimateZone: "Subtropical",
	},
}

// NewWeatherGenerator creates a new weather generator with random location and season
func NewWeatherGenerator() *WeatherGenerator {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Randomly select location and season
	location := Locations[rng.Intn(len(Locations))]
	season := Season(rng.Intn(4))

	wg := &WeatherGenerator{
		Location:    location,
		Season:      season,
		CurrentTime: time.Now(),
		rng:         rng,
	}

	wg.initializeBaseValues()
	return wg
}

// NewWeatherGeneratorWithParams creates a weather generator with specific location and season
func NewWeatherGeneratorWithParams(location Location, season Season) *WeatherGenerator {
	wg := &WeatherGenerator{
		Location:    location,
		Season:      season,
		CurrentTime: time.Now(),
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	wg.initializeBaseValues()
	return wg
}

// initializeBaseValues sets realistic base values based on location and season
func (wg *WeatherGenerator) initializeBaseValues() {
	// Set base temperature based on location and season
	wg.BaseTemperature = wg.getSeasonalTemperature()

	// Set base pressure (adjusted to sea level, then we'll adjust for elevation)
	wg.BasePressure = 1013.25 + wg.rng.Float64()*40 - 20 // 993-1033 mb range

	// Set base humidity based on climate zone
	wg.BaseHumidity = wg.getClimateHumidity()

	// Initialize cumulative rain and daily total
	wg.cumulativeRain = 1.5 + wg.rng.Float64()*8.0 // Start with some pre-existing accumulation (1.5-9.5 inches)
	wg.dailyRainTotal = 0.0                        // Start daily total at 0
	wg.lastDayCheck = wg.CurrentTime.YearDay()     // Track current day
}

// getSeasonalTemperature returns realistic temperatures for location and season
func (wg *WeatherGenerator) getSeasonalTemperature() float64 {
	baseTemp := 15.0 // Default 15°C (59°F)

	// Adjust for latitude (rough approximation)
	latAdjust := (45.0 - math.Abs(wg.Location.Latitude)) * 0.3
	baseTemp += latAdjust

	// Seasonal adjustments
	switch wg.Season {
	case Spring:
		baseTemp += wg.rng.Float64()*8 + 3 // 3-11°C adjustment
	case Summer:
		baseTemp += wg.rng.Float64()*10 + 8 // 8-18°C adjustment
	case Fall:
		baseTemp += wg.rng.Float64() * 6 // 0-6°C adjustment
	case Winter:
		baseTemp -= wg.rng.Float64()*10 + 2 // -12 to -2°C adjustment
	}

	// Climate zone adjustments (more moderate)
	switch {
	case wg.Location.ClimateZone == "Tropical":
		baseTemp += 5 // Reduced from 8
	case wg.Location.ClimateZone == "Desert":
		switch wg.Season {
		case Summer:
			baseTemp += 8 // Reduced from 15
		case Winter:
			baseTemp += 2 // Reduced from 5
		}
	case wg.Location.ClimateZone == "Subarctic":
		baseTemp -= 15 // Reduced from 20
	case wg.Location.ClimateZone == "Mediterranean":
		baseTemp += 2 // Reduced from 3
	}

	// Ensure temperatures stay within reasonable bounds
	baseTemp = math.Max(-25, math.Min(42, baseTemp)) // -13°F to 107.6°F range

	return baseTemp
}

// getClimateHumidity returns realistic humidity for the climate zone
func (wg *WeatherGenerator) getClimateHumidity() float64 {
	switch wg.Location.ClimateZone {
	case "Tropical", "Subtropical":
		return 70 + wg.rng.Float64()*25 // 70-95%
	case "Desert":
		return 10 + wg.rng.Float64()*30 // 10-40%
	case "Oceanic":
		return 60 + wg.rng.Float64()*30 // 60-90%
	case "Continental":
		return 40 + wg.rng.Float64()*40 // 40-80%
	case "Mediterranean":
		return 45 + wg.rng.Float64()*35 // 45-80%
	case "Subarctic":
		return 60 + wg.rng.Float64()*25 // 60-85%
	default:
		return 50 + wg.rng.Float64()*30 // 50-80%
	}
}

// SetCurrentWeatherMode ensures the generator is in current weather mode (not historical)
func (wg *WeatherGenerator) SetCurrentWeatherMode() {
	wg.isGeneratingHistorical = false
}

// GenerateObservation creates a single realistic weather observation
func (wg *WeatherGenerator) GenerateObservation() *types.Observation {
	// Use CurrentTime if set (for historical data), otherwise use current time
	observationTime := wg.CurrentTime
	if observationTime.IsZero() {
		observationTime = time.Now()
	}

	// Generate temperature with daily variation
	hourOfDay := float64(observationTime.Hour())
	tempVariation := math.Sin((hourOfDay-6)*math.Pi/12) * 4                      // Reduced from 8 to 4 degrees variation, peak at 2 PM, minimum at 6 AM
	temperature := wg.BaseTemperature + tempVariation + (wg.rng.Float64()-0.5)*2 // Reduced random variation from 4 to 2

	// Generate humidity (inversely related to temperature)
	humidity := wg.BaseHumidity - tempVariation*2 + (wg.rng.Float64()-0.5)*15
	humidity = math.Max(10, math.Min(98, humidity))

	// Generate pressure with realistic variation
	pressure := wg.BasePressure + (wg.rng.Float64()-0.5)*10
	// Adjust for elevation
	pressure = pressure * math.Pow(1-0.0065*wg.Location.Elevation/288.15, 5.255)

	// Generate wind based on season and location
	windSpeed := wg.generateWind()
	windDirection := wg.rng.Float64() * 360
	windGust := windSpeed * (1.2 + wg.rng.Float64()*0.8)

	// Generate illuminance based on time of day
	illuminance := wg.generateIlluminance(observationTime)

	// Generate UV based on time of day and season
	uv := wg.generateUV(observationTime)

	// Generate rain based on season and climate
	rain := wg.generateRain()

	// Generate solar radiation
	solar := wg.generateSolar(observationTime)

	obs := &types.Observation{
		Timestamp:            observationTime.Unix(),
		WindLull:             math.Max(0, windSpeed-wg.rng.Float64()*2),
		WindAvg:              windSpeed,
		WindGust:             windGust,
		WindDirection:        windDirection,
		StationPressure:      pressure,
		AirTemperature:       temperature,
		RelativeHumidity:     humidity,
		Illuminance:          illuminance,
		UV:                   int(uv),
		SolarRadiation:       solar,
		RainAccumulated:      rain,
		RainDailyTotal:       wg.dailyRainTotal,
		PrecipitationType:    wg.generatePrecipitationType(temperature, rain),
		LightningStrikeAvg:   wg.generateLightning(),
		LightningStrikeCount: wg.generateLightningCount(),
		Battery:              3.6 + wg.rng.Float64()*0.3, // 3.6-3.9V
		ReportInterval:       60,
	}

	wg.current = obs
	return obs
}

// generateWind creates realistic wind patterns
func (wg *WeatherGenerator) generateWind() float64 {
	baseWind := 2.0 + wg.rng.Float64()*8 // 2-10 mph base

	// Seasonal adjustments
	switch wg.Season {
	case Spring:
		baseWind *= 1.3 // Windier in spring
	case Summer:
		baseWind *= 0.8 // Calmer in summer
	case Fall:
		baseWind *= 1.2 // Moderate wind in fall
	case Winter:
		baseWind *= 1.5 // Windier in winter
	}

	// Climate adjustments
	switch wg.Location.ClimateZone {
	case "Oceanic":
		baseWind *= 1.4
	case "Desert":
		baseWind *= 0.9
	case "Continental":
		baseWind *= 1.1
	}

	return baseWind
}

// generateIlluminance creates realistic light levels
func (wg *WeatherGenerator) generateIlluminance(t time.Time) float64 {
	hour := t.Hour()

	// Night time - low light levels
	if hour < 6 || hour > 20 {
		return wg.rng.Float64() * 10 // 0-10 lux at night (moon/stars/artificial light)
	}

	// Dawn/dusk
	if hour == 6 || hour == 7 || hour == 19 || hour == 20 {
		return 50 + wg.rng.Float64()*200 // 50-250 lux
	}

	// Daytime - varies by season and weather
	baseLux := 10000.0 // Clear day baseline

	// Seasonal adjustments
	switch wg.Season {
	case Summer:
		baseLux *= 1.2
	case Winter:
		baseLux *= 0.6
	case Spring, Fall:
		baseLux *= 0.9
	}

	// Add some cloud variation
	cloudFactor := 0.3 + wg.rng.Float64()*0.7 // 30-100% of clear sky

	return baseLux*cloudFactor + wg.rng.Float64()*5000
}

// generateUV creates realistic UV index
func (wg *WeatherGenerator) generateUV(t time.Time) float64 {
	hour := t.Hour()

	// No UV at night
	if hour < 8 || hour > 18 {
		return 0
	}

	// Peak UV around noon
	uvFactor := math.Sin((float64(hour) - 6) * math.Pi / 12)
	if uvFactor < 0 {
		uvFactor = 0
	}

	maxUV := 3.0 // Base maximum UV

	// Seasonal adjustments
	switch wg.Season {
	case Summer:
		maxUV = 9.0
	case Spring, Fall:
		maxUV = 6.0
	case Winter:
		maxUV = 3.0
	}

	// Latitude adjustments
	if math.Abs(wg.Location.Latitude) < 30 {
		maxUV *= 1.3 // Higher UV near equator
	} else if math.Abs(wg.Location.Latitude) > 50 {
		maxUV *= 0.7 // Lower UV at high latitudes
	}

	return math.Max(0, maxUV*uvFactor+(wg.rng.Float64()-0.5)*2)
}

// GetDailyRainTotal returns the daily rain total for generated weather
func (wg *WeatherGenerator) GetDailyRainTotal() float64 {
	return wg.dailyRainTotal
}

// generateRain creates realistic precipitation and updates cumulative total
func (wg *WeatherGenerator) generateRain() float64 {
	// Only update daily totals if not generating historical data
	if !wg.isGeneratingHistorical {
		// Check if it's a new day and reset daily total if needed
		currentDay := time.Now().YearDay()
		if currentDay != wg.lastDayCheck {
			wg.dailyRainTotal = 0.0 // Reset daily total at midnight
			wg.lastDayCheck = currentDay
		}
	}

	// Base probability of rain
	rainChance := 0.1 // 10% base chance

	// Seasonal adjustments
	switch wg.Season {
	case Spring:
		rainChance = 0.25
	case Summer:
		rainChance = 0.15
	case Fall:
		rainChance = 0.2
	case Winter:
		rainChance = 0.3
	}

	// Climate adjustments
	switch wg.Location.ClimateZone {
	case "Tropical", "Oceanic":
		rainChance *= 2
	case "Desert":
		rainChance *= 0.2
	case "Subtropical":
		rainChance *= 1.5
	}

	var incrementalRain float64
	if wg.rng.Float64() < rainChance {
		// Light to moderate rain (per minute/observation)
		incrementalRain = wg.rng.Float64() * 2.54 // 0-2.54 mm per observation (equivalent to 0-0.1 inches)
		wg.cumulativeRain += incrementalRain

		// Only add to daily total if not generating historical data
		if !wg.isGeneratingHistorical {
			wg.dailyRainTotal += incrementalRain
		}
	}

	return incrementalRain
}

// generateSolar creates realistic solar radiation
func (wg *WeatherGenerator) generateSolar(t time.Time) float64 {
	hour := t.Hour()

	if hour < 6 || hour > 19 {
		return 0
	}

	// Peak solar around noon
	solarFactor := math.Sin((float64(hour) - 6) * math.Pi / 13)
	if solarFactor < 0 {
		solarFactor = 0
	}

	maxSolar := 800.0 // W/m²

	// Seasonal adjustments
	switch wg.Season {
	case Summer:
		maxSolar = 1000.0
	case Winter:
		maxSolar = 400.0
	case Spring, Fall:
		maxSolar = 700.0
	}

	return maxSolar * solarFactor * (0.5 + wg.rng.Float64()*0.5)
}

// generatePrecipitationType determines precipitation type
func (wg *WeatherGenerator) generatePrecipitationType(temp, rain float64) int {
	if rain == 0 {
		return 0 // None
	}

	if temp < 0 {
		return 3 // Snow
	} else if temp < 2 {
		return 2 // Ice pellets
	} else {
		return 1 // Rain
	}
}

// generateLightning creates lightning distance
func (wg *WeatherGenerator) generateLightning() float64 {
	// Lightning is rare
	if wg.rng.Float64() < 0.05 { // 5% chance
		return 5 + wg.rng.Float64()*40 // 5-45 km
	}
	return 0
}

// generateLightningCount creates lightning count
func (wg *WeatherGenerator) generateLightningCount() int {
	if wg.rng.Float64() < 0.05 { // 5% chance
		return int(wg.rng.Float64() * 5) // 0-4 strikes
	}
	return 0
}

// GenerateHistoricalData creates a series of historical observations
func (wg *WeatherGenerator) GenerateHistoricalData(count int) []*types.Observation {
	observations := make([]*types.Observation, count)

	// Save the current state to restore later (historical generation should not affect current day)
	originalDailyTotal := wg.dailyRainTotal
	originalCumulativeRain := wg.cumulativeRain
	originalTime := wg.CurrentTime

	// Start from 24 hours ago and work forward
	startTime := time.Now().Add(-24 * time.Hour)
	interval := 24 * time.Hour / time.Duration(count)

	// Set a flag to prevent rain generation from affecting daily totals during historical generation
	wg.isGeneratingHistorical = true

	for i := 0; i < count; i++ {
		// Set the current time for this observation
		wg.CurrentTime = startTime.Add(time.Duration(i) * interval)

		// Generate observation for this time
		obs := wg.GenerateObservation()
		obs.Timestamp = wg.CurrentTime.Unix()
		observations[i] = obs

		// Add some continuity - slight drift in base values
		wg.BaseTemperature += (wg.rng.Float64() - 0.5) * 0.2
		wg.BasePressure += (wg.rng.Float64() - 0.5) * 0.5
		wg.BaseHumidity += (wg.rng.Float64() - 0.5) * 1.0

		// Keep values in reasonable ranges
		wg.BaseTemperature = math.Max(-20, math.Min(50, wg.BaseTemperature))
		wg.BasePressure = math.Max(980, math.Min(1040, wg.BasePressure))
		wg.BaseHumidity = math.Max(20, math.Min(95, wg.BaseHumidity))
	}

	// Clear the historical generation flag
	wg.isGeneratingHistorical = false

	// Restore the original state (historical generation should not corrupt current day)
	wg.dailyRainTotal = originalDailyTotal
	wg.cumulativeRain = originalCumulativeRain
	wg.CurrentTime = originalTime

	wg.history = observations
	return observations
}

// GetLocation returns the current location
func (wg *WeatherGenerator) GetLocation() Location {
	return wg.Location
}

// GetSeason returns the current season
func (wg *WeatherGenerator) GetSeason() Season {
	return wg.Season
}

// Regenerate creates a new random location and season combination
func (wg *WeatherGenerator) Regenerate() {
	// Select new random location and season
	wg.Location = Locations[wg.rng.Intn(len(Locations))]
	wg.Season = Season(wg.rng.Intn(4))

	// Reinitialize base values
	wg.initializeBaseValues()

	// Clear history to force regeneration
	wg.history = nil
}

// GenerateNewSeason generates a new random location and season (alias for Regenerate)
func (wg *WeatherGenerator) GenerateNewSeason() {
	wg.Regenerate()
}
