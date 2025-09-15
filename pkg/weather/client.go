package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

const (
	BaseURL = "https://swd.weatherflow.com/swd/rest"
)

type Device struct {
	DeviceID   int    `json:"device_id"`
	DeviceType string `json:"device_type"`
	SerialNumber string `json:"serial_number"`
}

type Station struct {
	StationID   int      `json:"station_id"`
	Name        string   `json:"name"`
	StationName string   `json:"station_name"`
	Devices     []Device `json:"devices"`
}

type StationsResponse struct {
	Stations []Station `json:"stations"`
}

type StationDetailsResponse struct {
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

// HistoricalResponse represents the structure for historical data from WeatherFlow API
type HistoricalResponse struct {
	Status       map[string]interface{} `json:"status"`
	StationID    int                    `json:"station_id"`
	StationName  string                 `json:"station_name"`
	StationUnits map[string]string      `json:"station_units"`
	Obs          [][]interface{}        `json:"obs"` // Device API returns array of arrays, not array of maps
}

// ForecastPeriod represents a single forecast period from the better_forecast API
type ForecastPeriod struct {
	Time             int64   `json:"time"`
	Icon             string  `json:"icon"`
	Conditions       string  `json:"conditions"`
	AirTemperature   float64 `json:"air_temperature"`
	FeelsLike        float64 `json:"feels_like"`
	SeaLevelPressure float64 `json:"sea_level_pressure"`
	RelativeHumidity int     `json:"relative_humidity"`
	PrecipProbability int    `json:"precip_probability"`
	PrecipIcon       string  `json:"precip_icon"`
	PrecipType       string  `json:"precip_type"`
	WindAvg          float64 `json:"wind_avg"`
	WindDirection    int     `json:"wind_direction"`
	WindGust         float64 `json:"wind_gust"`
	UV               int     `json:"uv"`
}

// ForecastResponse represents the structure for forecast data from WeatherFlow API
type ForecastResponse struct {
	Status       map[string]interface{} `json:"status"`
	StationID    int                    `json:"station_id"`
	StationName  string                 `json:"station_name"`
	Timezone     string                 `json:"timezone"`
	Forecast     struct {
		Daily []ForecastPeriod `json:"daily"`
	} `json:"forecast"`
	CurrentConditions ForecastPeriod `json:"current_conditions"`
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

func GetStationDetails(stationID int, token string) (*Station, error) {
	url := fmt.Sprintf("%s/stations/%d?token=%s", BaseURL, stationID, token)
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

	var stationResp StationDetailsResponse
	err = json.Unmarshal(body, &stationResp)
	if err != nil {
		return nil, err
	}

	if len(stationResp.Stations) == 0 {
		return nil, fmt.Errorf("no station details found")
	}

	return &stationResp.Stations[0], nil
}

func FindStationByName(stations []Station, name string) *Station {
	for _, s := range stations {
		if s.Name == name || s.StationName == name {
			return &s
		}
	}
	return nil
}

// GetTempestDeviceID returns the first Tempest device ID from a station
func GetTempestDeviceID(station *Station) (int, error) {
	for _, device := range station.Devices {
		if device.DeviceType == "ST" { // ST = Tempest
			return device.DeviceID, nil
		}
	}
	return 0, fmt.Errorf("no Tempest device found in station")
}

// ProgressCallback is a function type for reporting progress during historical data loading
type ProgressCallback func(currentStep, totalSteps int, description string)

// GetHistoricalObservations fetches historical weather data using the device-based endpoint with day_offset
func GetHistoricalObservations(stationID int, token string) ([]*Observation, error) {
	return GetHistoricalObservationsWithProgress(stationID, token, nil)
}

// GetHistoricalObservationsWithProgress fetches historical weather data with progress reporting
func GetHistoricalObservationsWithProgress(stationID int, token string, progressCallback ProgressCallback) ([]*Observation, error) {
	// First get station details to find the Tempest device ID
	stationDetails, err := GetStationDetails(stationID, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get station details: %v", err)
	}

	deviceID, err := GetTempestDeviceID(stationDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to find Tempest device: %v", err)
	}

	var allObservations []*Observation

	fmt.Printf("INFO: Collecting historical data for station %d using device %d\n", stationID, deviceID)
	fmt.Printf("INFO: Fetching today and yesterday's observations using day_offset parameter...\n")

	successCount := 0
	errorCount := 0
	totalSteps := 2 // Today and yesterday

	// Report initial progress
	if progressCallback != nil {
		progressCallback(0, totalSteps, "Starting historical data collection...")
	}

	// Get observations for today (day_offset=0) and yesterday (day_offset=1)
	for dayOffset := 0; dayOffset <= 1; dayOffset++ {
		currentStep := dayOffset + 1
		dayName := "today"
		if dayOffset == 1 {
			dayName = "yesterday"
		}

		// Report progress before fetching
		if progressCallback != nil {
			progressCallback(currentStep-1, totalSteps, fmt.Sprintf("Fetching %s's observations...", dayName))
		}

		url := fmt.Sprintf("%s/observations/device/%d?day_offset=%d&token=%s",
			BaseURL, deviceID, dayOffset, token)

		fmt.Printf("INFO: Fetching observations for %s (day_offset=%d)...\n", dayName, dayOffset)

		resp, err := http.Get(url)
		if err != nil {
			errorCount++
			fmt.Printf("ERROR: API call failed for %s: %v\n", dayName, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			errorCount++
			fmt.Printf("ERROR: API call for %s returned HTTP %d\n", dayName, resp.StatusCode)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			fmt.Printf("ERROR: Error reading response for %s: %v\n", dayName, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		var apiResp HistoricalResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			errorCount++
			fmt.Printf("ERROR: Error parsing JSON for %s: %v\n", dayName, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		observations := parseDeviceObservations(apiResp.Obs)
		if len(observations) > 0 {
			allObservations = append(allObservations, observations...)
			successCount++
			fmt.Printf("INFO: Successfully retrieved %d observations for %s\n", len(observations), dayName)
			
			// Report progress after successful fetch
			if progressCallback != nil {
				progressCallback(currentStep, totalSteps, fmt.Sprintf("Processed %d observations for %s", len(observations), dayName))
			}
		} else {
			fmt.Printf("WARN: No observations found for %s\n", dayName)
		}

		// Rate limiting: brief pause between requests to be respectful
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("INFO: Collection complete - %d successful calls, %d errors, %d total observations\n",
		successCount, errorCount, len(allObservations))

	// Sort observations by timestamp (newest first)
	sort.Slice(allObservations, func(i, j int) bool {
		return allObservations[i].Timestamp > allObservations[j].Timestamp
	})

	// Remove duplicates if any (based on timestamp)
	uniqueObs := make([]*Observation, 0, len(allObservations))
	seen := make(map[int64]bool)

	for _, obs := range allObservations {
		if !seen[obs.Timestamp] {
			seen[obs.Timestamp] = true
			uniqueObs = append(uniqueObs, obs)
		}
	}

	// Limit to 1000 points maximum for web dashboard
	if len(uniqueObs) > 1000 {
		uniqueObs = uniqueObs[:1000]
	}

	// Calculate actual time span
	var timeSpanHours float64 = 0
	if len(uniqueObs) > 1 {
		oldestTime := time.Unix(uniqueObs[len(uniqueObs)-1].Timestamp, 0)
		newestTime := time.Unix(uniqueObs[0].Timestamp, 0)
		timeSpanHours = newestTime.Sub(oldestTime).Hours()
	}

	fmt.Printf("INFO: Final dataset: %d unique observations spanning %.1f hours of data\n",
		len(uniqueObs), timeSpanHours)

	// Print detailed statistics for verification
	if len(uniqueObs) > 0 {
		oldestObs := time.Unix(uniqueObs[len(uniqueObs)-1].Timestamp, 0)
		newestObs := time.Unix(uniqueObs[0].Timestamp, 0)
		fmt.Printf("INFO: Data range: %s to %s\n",
			oldestObs.Format("2006-01-02 15:04:05"),
			newestObs.Format("2006-01-02 15:04:05"))
		
		// Count observations by day for verification
		todayCount := 0
		yesterdayCount := 0
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		yesterday := today.Add(-24 * time.Hour)
		
		for _, obs := range uniqueObs {
			obsTime := time.Unix(obs.Timestamp, 0).In(now.Location())
			obsDay := time.Date(obsTime.Year(), obsTime.Month(), obsTime.Day(), 0, 0, 0, 0, now.Location())
			
			if obsDay.Equal(today) {
				todayCount++
			} else if obsDay.Equal(yesterday) {
				yesterdayCount++
			}
		}
		
		fmt.Printf("INFO: Today: %d observations, Yesterday: %d observations\n", todayCount, yesterdayCount)
	}

	return uniqueObs, nil
}

// parseObservations converts raw observation maps to Observation structs (for station API)
func parseObservations(obsData []map[string]interface{}) []*Observation {
	var observations []*Observation

	for _, obsMap := range obsData {
		obs := &Observation{
			Timestamp:            int64(getFloat64(obsMap["timestamp"])),
			WindLull:             getFloat64(obsMap["wind_lull"]),
			WindAvg:              getFloat64(obsMap["wind_avg"]),
			WindGust:             getFloat64(obsMap["wind_gust"]),
			WindDirection:        getFloat64(obsMap["wind_direction"]),
			StationPressure:      getFloat64(obsMap["station_pressure"]),
			AirTemperature:       getFloat64(obsMap["air_temperature"]),
			RelativeHumidity:     getFloat64(obsMap["relative_humidity"]),
			Illuminance:          getFloat64(obsMap["brightness"]), // API uses "brightness" instead of "illuminance"
			UV:                   getFloat64(obsMap["uv"]),
			SolarRadiation:       getFloat64(obsMap["solar_radiation"]),
			RainAccumulated:      getFloat64(obsMap["precip"]),                     // API uses "precip" instead of "rain_accumulated"
			PrecipitationType:    getInt(obsMap["precip_analysis_type_yesterday"]), // Note: might need adjustment
			LightningStrikeAvg:   getFloat64(obsMap["lightning_strike_last_distance"]),
			LightningStrikeCount: getInt(obsMap["lightning_strike_count"]),
			Battery:              0, // Not provided in historical data
			ReportInterval:       0, // Not provided in historical data
		}
		observations = append(observations, obs)
	}

	return observations
}

// parseDeviceObservations converts device API observations (arrays) to Observation structs
// Device API returns observations as arrays. Based on API testing, the structure is:
// [0]: timestamp, [1]: wind_lull, [2]: wind_avg, [3]: wind_gust, [4]: wind_direction, [5]: ?, 
// [6]: station_pressure, [7]: air_temperature, [8]: relative_humidity, [9]: illuminance, 
// [10]: uv, [11]: solar_radiation, [12]: rain_accumulated, [13]: precipitation_type, 
// [14]: lightning_strike_avg_distance, [15]: lightning_strike_count, [16]: battery, [17]: report_interval
func parseDeviceObservations(obsData [][]interface{}) []*Observation {
	var observations []*Observation

	for _, obsArray := range obsData {
		// Ensure we have enough elements in the array
		if len(obsArray) < 18 {
			continue // Skip incomplete observations
		}

		obs := &Observation{
			Timestamp:            int64(getFloat64(obsArray[0])),  // timestamp
			WindLull:             getFloat64(obsArray[1]),         // wind_lull
			WindAvg:              getFloat64(obsArray[2]),         // wind_avg
			WindGust:             getFloat64(obsArray[3]),         // wind_gust
			WindDirection:        getFloat64(obsArray[4]),         // wind_direction
			StationPressure:      getFloat64(obsArray[6]),         // station_pressure (skip [5])
			AirTemperature:       getFloat64(obsArray[7]),         // air_temperature
			RelativeHumidity:     getFloat64(obsArray[8]),         // relative_humidity
			Illuminance:          getFloat64(obsArray[9]),         // illuminance
			UV:                   getFloat64(obsArray[10]),        // uv
			SolarRadiation:       getFloat64(obsArray[11]),        // solar_radiation
			RainAccumulated:      getFloat64(obsArray[12]),        // rain_accumulated
			PrecipitationType:    getInt(obsArray[13]),            // precipitation_type
			LightningStrikeAvg:   getFloat64(obsArray[14]),        // lightning_strike_avg_distance
			LightningStrikeCount: getInt(obsArray[15]),            // lightning_strike_count
			Battery:              getFloat64(obsArray[16]),        // battery
			ReportInterval:       getInt(obsArray[17]),            // report_interval
		}
		observations = append(observations, obs)
	}

	return observations
}

// filterToOneMinuteIncrements filters observations to get approximately count points
// spaced one minute apart, working backwards from the most recent observation
func filterToOneMinuteIncrements(observations []*Observation, maxCount int) []*Observation {
	if len(observations) == 0 {
		return observations
	}

	// Sort observations by timestamp (newest first) to ensure proper ordering
	sort.Slice(observations, func(i, j int) bool {
		return observations[i].Timestamp > observations[j].Timestamp
	})

	var filtered []*Observation
	var lastTimestamp int64 = 0
	const oneMinute = 60 // seconds

	// Work through observations and select ones that are approximately 1 minute apart
	for _, obs := range observations {
		if lastTimestamp == 0 || obs.Timestamp <= lastTimestamp-oneMinute {
			filtered = append(filtered, obs)
			lastTimestamp = obs.Timestamp

			// Stop when we have enough observations
			if len(filtered) >= maxCount {
				break
			}
		}
	}

	// Reverse the slice to have oldest first (chronological order for charts)
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	return filtered
}

// GetForecast fetches forecast data from the WeatherFlow better_forecast endpoint
func GetForecast(stationID int, token string) (*ForecastResponse, error) {
	url := fmt.Sprintf("%s/better_forecast?station_id=%d&token=%s", BaseURL, stationID, token)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch forecast data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("forecast API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read forecast response: %v", err)
	}

	var forecastResp ForecastResponse
	if err := json.Unmarshal(body, &forecastResp); err != nil {
		return nil, fmt.Errorf("failed to parse forecast JSON: %v", err)
	}

	return &forecastResp, nil
}
