# Weather Generator Package

This package provides synthetic weather data generation for UI testing purposes. It creates realistic weather patterns based on different seasons and geographic locations.

## Features

- **Multiple Climate Zones**: Tropical, Continental, Oceanic, Desert, Mediterranean, Subarctic, Subtropical
- **Seasonal Variation**: Spring, Summer, Fall, Winter patterns
- **Realistic Data**: Temperature, humidity, pressure, wind, UV, illuminance, precipitation
- **Historical Generation**: Can generate 1000+ data points for historical data
- **Location-Specific**: 8 predefined locations with different climates

## Usage

```go
// Create generator with random location and season
generator := NewWeatherGenerator()

// Generate a single observation
obs := generator.GenerateObservation()

// Generate historical data
history := generator.GenerateHistoricalData(1000)

// Get current location and season
location := generator.GetLocation()
season := generator.GetSeason()

// Regenerate with new random location and season
generator.Regenerate()
```

## Locations

1. Miami, FL (Tropical)
2. Denver, CO (Continental) 
3. Seattle, WA (Oceanic)
4. Phoenix, AZ (Desert)
5. Minneapolis, MN (Continental)
6. San Diego, CA (Mediterranean)
7. Anchorage, AK (Subarctic)
8. New Orleans, LA (Subtropical)

## Data Generation

The generator creates realistic data by:
- Adjusting base values for climate zone and season
- Adding daily temperature variations (peak afternoon, minimum dawn)
- Creating weather-appropriate humidity levels
- Generating time-of-day appropriate illuminance and UV levels
- Including realistic wind patterns and precipitation
- Maintaining data continuity across time series