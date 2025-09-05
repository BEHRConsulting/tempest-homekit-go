package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	BaseURL = "https://swd.weatherflow.com/swd/rest"
)

type Station struct {
	StationID   int    `json:"station_id"`
	Name        string `json:"name"`
	StationName string `json:"station_name"`
}

type StationsResponse struct {
	Stations []Station `json:"stations"`
}

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
	UV                   float64 `json:"uv"`
	SolarRadiation       float64 `json:"solar_radiation"`
	RainAccumulated      float64 `json:"rain_accumulated"`
	PrecipitationType    int     `json:"precipitation_type"`
	LightningStrikeAvg   float64 `json:"lightning_strike_avg_distance"`
	LightningStrikeCount int     `json:"lightning_strike_count"`
	Battery              float64 `json:"battery"`
	ReportInterval       int     `json:"report_interval"`
}

type ObservationResponse struct {
	Obs []map[string]interface{} `json:"obs"`
}

func GetStations(token string) ([]Station, error) {
	url := fmt.Sprintf("%s/stations?token=%s", BaseURL, token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stationsResp StationsResponse
	err = json.Unmarshal(body, &stationsResp)
	if err != nil {
		return nil, err
	}

	return stationsResp.Stations, nil
}

func GetObservation(stationID int, token string) (*Observation, error) {
	url := fmt.Sprintf("%s/observations/station/%d?token=%s", BaseURL, stationID, token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var obsResp ObservationResponse
	err = json.Unmarshal(body, &obsResp)
	if err != nil {
		return nil, err
	}

	if len(obsResp.Obs) == 0 {
		return nil, fmt.Errorf("no observations found")
	}

	latest := obsResp.Obs[0] // latest is first

	obs := &Observation{
		Timestamp:            int64(getFloat64(latest["timestamp"])),
		WindLull:             getFloat64(latest["wind_lull"]),
		WindAvg:              getFloat64(latest["wind_avg"]),
		WindGust:             getFloat64(latest["wind_gust"]),
		WindDirection:        getFloat64(latest["wind_direction"]),
		StationPressure:      getFloat64(latest["station_pressure"]),
		AirTemperature:       getFloat64(latest["air_temperature"]),
		RelativeHumidity:     getFloat64(latest["relative_humidity"]),
		Illuminance:          getFloat64(latest["brightness"]), // API uses "brightness" instead of "illuminance"
		UV:                   getFloat64(latest["uv"]),
		SolarRadiation:       getFloat64(latest["solar_radiation"]),
		RainAccumulated:      getFloat64(latest["precip"]), // API uses "precip" instead of "rain_accumulated"
		PrecipitationType:    getInt(latest["precipitation_type"]),
		LightningStrikeAvg:   getFloat64(latest["lightning_strike_avg"]),
		LightningStrikeCount: getInt(latest["lightning_strike_count"]),
		Battery:              getFloat64(latest["battery"]),
		ReportInterval:       getInt(latest["report_interval"]),
	}

	return obs, nil
}

func getFloat64(value interface{}) float64 {
	if value == nil {
		return 0.0
	}
	if f, ok := value.(float64); ok {
		return f
	}
	return 0.0
}

func getInt(value interface{}) int {
	if value == nil {
		return 0
	}
	if f, ok := value.(float64); ok {
		return int(f)
	}
	return 0
}

func FindStationByName(stations []Station, name string) *Station {
	for _, s := range stations {
		if s.Name == name || s.StationName == name {
			return &s
		}
	}
	return nil
}
