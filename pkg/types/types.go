// Package types defines common types used across the application.
// This package helps break import cycles by providing shared type definitions.
package types

// Observation represents a weather observation from a Tempest station
type Observation struct {
	Timestamp            int64   `json:"timestamp"`
	WindLull             float64 `json:"wind_lull"`
	WindAvg              float64 `json:"wind_avg"`
	WindGust             float64 `json:"wind_gust"`
	WindDirection        float64 `json:"wind_direction"`
	StationPressure      float64 `json:"station_pressure"`
	AirTemperature       float64 `json:"air_temperature"`
	RelativeHumidity     float64 `json:"relative_humidity"`
	Illuminance          float64 `json:"illuminance"`
	UV                   int     `json:"uv"`
	SolarRadiation       float64 `json:"solar_radiation"`
	RainAccumulated      float64 `json:"rain_accumulated"` // Incremental rain since last obs (from "precip" field)
	RainDailyTotal       float64 `json:"rain_daily_total"` // Total rain since midnight (from "precip_accum_local_day" field)
	PrecipitationType    int     `json:"precipitation_type"`
	LightningStrikeAvg   float64 `json:"lightning_strike_avg_distance"`
	LightningStrikeCount int     `json:"lightning_strike_count"`
	Battery              float64 `json:"battery"`
	ReportInterval       int     `json:"report_interval"`
}
