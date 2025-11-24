// Package weather provides a client for the WeatherFlow Tempest API.
// It handles authentication, data retrieval, and parsing of weather observations
// and forecast data from WeatherFlow stations.
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/types"

	"github.com/chromedp/chromedp"
)

const (
	BaseURL = "https://swd.weatherflow.com/swd/rest"
)

type Device struct {
	DeviceID     int    `json:"device_id"`
	DeviceType   string `json:"device_type"`
	SerialNumber string `json:"serial_number"`
}

type Station struct {
	StationID   int      `json:"station_id"`
	Name        string   `json:"name"`
	StationName string   `json:"station_name"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Devices     []Device `json:"devices"`
}

type StationsResponse struct {
	Stations []Station `json:"stations"`
}

type StationDetailsResponse struct {
	Stations []Station `json:"stations"`
}

type Observation = types.Observation

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
	Time              int64   `json:"time"`
	Icon              string  `json:"icon"`
	Conditions        string  `json:"conditions"`
	AirTemperature    float64 `json:"air_temperature"`
	AirTempHigh       float64 `json:"air_temp_high"`
	AirTempLow        float64 `json:"air_temp_low"`
	FeelsLike         float64 `json:"feels_like"`
	SeaLevelPressure  float64 `json:"sea_level_pressure"`
	RelativeHumidity  int     `json:"relative_humidity"`
	PrecipProbability int     `json:"precip_probability"`
	PrecipIcon        string  `json:"precip_icon"`
	PrecipType        string  `json:"precip_type"`
	WindAvg           float64 `json:"wind_avg"`
	WindDirection     int     `json:"wind_direction"`
	WindGust          float64 `json:"wind_gust"`
	UV                int     `json:"uv"`
}

// ForecastResponse represents the structure for forecast data from WeatherFlow API
type ForecastResponse struct {
	Status      map[string]interface{} `json:"status"`
	StationID   int                    `json:"station_id"`
	StationName string                 `json:"station_name"`
	Timezone    string                 `json:"timezone"`
	Forecast    struct {
		Daily []ForecastPeriod `json:"daily"`
	} `json:"forecast"`
	CurrentConditions ForecastPeriod `json:"current_conditions"`
}

// GetStations retrieves all weather stations associated with the provided API token.
func GetStations(token string) ([]Station, error) {
	url := fmt.Sprintf("%s/stations?token=%s", BaseURL, token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

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

// GetObservation retrieves the latest weather observation for the specified station.
func GetObservation(stationID int, token string) (*Observation, error) {
	url := fmt.Sprintf("%s/observations/station/%d?token=%s", BaseURL, stationID, token)
	return GetObservationFromURL(url)
}

// GetObservationFromURL fetches weather data from a custom URL (e.g., generated weather endpoint)
func GetObservationFromURL(url string) (*Observation, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

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
		UV:                   getInt(latest["uv"]),
		SolarRadiation:       getFloat64(latest["solar_radiation"]),
		RainAccumulated:      getFloat64(latest["precip"]),                 // Incremental rain since last obs
		RainDailyTotal:       getFloat64(latest["precip_accum_local_day"]), // Total rain since midnight (mm)
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
	defer func() { _ = resp.Body.Close() }()

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

// FindStationByName searches for a station with the given name in the provided stations slice.
// Returns nil if no matching station is found.
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
func GetHistoricalObservations(stationID int, token string, logLevel string) ([]*Observation, error) {
	return GetHistoricalObservationsWithProgress(stationID, token, logLevel, nil, 1000)
}

// GetHistoricalObservationsWithProgress fetches historical weather data with progress reporting
func GetHistoricalObservationsWithProgress(stationID int, token string, logLevel string, progressCallback ProgressCallback, maxPoints int) ([]*Observation, error) {
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

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Collecting historical data for station %d using device %d\n", stationID, deviceID)
		fmt.Printf("DEBUG: Fetching today and yesterday's observations using day_offset parameter...\n")
	}

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

		if logLevel == "debug" {
			fmt.Printf("DEBUG: Fetching observations for %s (day_offset=%d)...\n", dayName, dayOffset)
		}

		resp, err := http.Get(url)
		if err != nil {
			errorCount++
			fmt.Printf("ERROR: API call failed for %s: %v\n", dayName, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			errorCount++
			fmt.Printf("ERROR: API call for %s returned HTTP %d\n", dayName, resp.StatusCode)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
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
			if logLevel == "debug" {
				fmt.Printf("DEBUG: Successfully retrieved %d observations for %s\n", len(observations), dayName)
			}

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

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Collection complete - %d successful calls, %d errors, %d total observations\n",
			successCount, errorCount, len(allObservations))
	}

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

	// Limit to configured maximum points
	if maxPoints > 0 && len(uniqueObs) > maxPoints {
		uniqueObs = uniqueObs[:maxPoints]
	}

	// Calculate actual time span
	var timeSpanHours float64 = 0
	if len(uniqueObs) > 1 {
		oldestTime := time.Unix(uniqueObs[len(uniqueObs)-1].Timestamp, 0)
		newestTime := time.Unix(uniqueObs[0].Timestamp, 0)
		timeSpanHours = newestTime.Sub(oldestTime).Hours()
	}

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Final dataset: %d unique observations spanning %.1f hours of data\n",
			len(uniqueObs), timeSpanHours)
	}

	// Print detailed statistics for verification
	if logLevel == "debug" && len(uniqueObs) > 0 {
		oldestObs := time.Unix(uniqueObs[len(uniqueObs)-1].Timestamp, 0)
		newestObs := time.Unix(uniqueObs[0].Timestamp, 0)
		fmt.Printf("DEBUG: Data range: %s to %s\n",
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

		fmt.Printf("DEBUG: Today: %d observations, Yesterday: %d observations\n", todayCount, yesterdayCount)
	}

	return uniqueObs, nil
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
			Timestamp:            int64(getFloat64(obsArray[0])), // timestamp
			WindLull:             getFloat64(obsArray[1]),        // wind_lull
			WindAvg:              getFloat64(obsArray[2]),        // wind_avg
			WindGust:             getFloat64(obsArray[3]),        // wind_gust
			WindDirection:        getFloat64(obsArray[4]),        // wind_direction
			StationPressure:      getFloat64(obsArray[6]),        // station_pressure (skip [5])
			AirTemperature:       getFloat64(obsArray[7]),        // air_temperature
			RelativeHumidity:     getFloat64(obsArray[8]),        // relative_humidity
			Illuminance:          getFloat64(obsArray[9]),        // illuminance
			UV:                   getInt(obsArray[10]),           // uv
			SolarRadiation:       getFloat64(obsArray[11]),       // solar_radiation
			RainAccumulated:      getFloat64(obsArray[12]),       // rain_accumulated
			PrecipitationType:    getInt(obsArray[13]),           // precipitation_type
			LightningStrikeAvg:   getFloat64(obsArray[14]),       // lightning_strike_avg_distance
			LightningStrikeCount: getInt(obsArray[15]),           // lightning_strike_count
			Battery:              getFloat64(obsArray[16]),       // battery
			ReportInterval:       getInt(obsArray[17]),           // report_interval
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

// package-level no-op reference to avoid static analyzer warnings when the helper
// function is currently unused. It is kept for future use by charting logic.
var _ = filterToOneMinuteIncrements

// GetForecast fetches forecast data from the WeatherFlow better_forecast endpoint
func GetForecast(stationID int, token string) (*ForecastResponse, error) {
	url := fmt.Sprintf("%s/better_forecast?station_id=%d&token=%s", BaseURL, stationID, token)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch forecast data: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

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

// GetAllHistoricalObservations attempts to collect historical observations by
// querying the device observations endpoint with increasing day_offset values
// until no more data is returned or maxDays is reached. maxPoints limits the
// total returned observations (0 = no limit).
func GetAllHistoricalObservations(stationID int, token string, logLevel string, maxDays int, reduceMethod string, binMinutes int, keepRecentHours int, reduceFactor int, maxPoints int) ([]*Observation, error) {
	// Resolve station details and Tempest device ID
	stationDetails, err := GetStationDetails(stationID, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get station details: %v", err)
	}

	deviceID, err := GetTempestDeviceID(stationDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to find Tempest device: %v", err)
	}

	if maxDays <= 0 {
		maxDays = 365 // sane default
	}

	var allObservations []*Observation
	consecutiveEmpty := 0

	for dayOffset := 0; dayOffset < maxDays; dayOffset++ {
		// Build URL
		url := fmt.Sprintf("%s/observations/device/%d?day_offset=%d&token=%s", BaseURL, deviceID, dayOffset, token)

		if logLevel == "debug" {
			fmt.Printf("DEBUG: Fetching day_offset=%d from %s\n", dayOffset, url)
		}

		// Attempt request with retries for 429 (rate limit). Use Retry-After header if provided.
		var resp *http.Response
		var err error
		maxRetries := 6
		attempt := 0
		for {
			resp, err = http.Get(url)
			if err != nil {
				// On transient error, retry a few times
				logger.Warn("API call failed for day_offset=%d: %v", dayOffset, err)
				attempt++
				if attempt > maxRetries {
					consecutiveEmpty++
					break
				}
				// small backoff with jitter
				backoffMs := 150 + rand.Intn(200)
				time.Sleep(time.Duration(backoffMs) * time.Millisecond)
				continue
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				// Read Retry-After header if present
				ra := resp.Header.Get("Retry-After")
				_ = resp.Body.Close()
				waitSec := 0
				if ra != "" {
					if n, perr := strconv.Atoi(ra); perr == nil {
						waitSec = n
					}
				}
				// fallback exponential backoff with jitter
				if waitSec <= 0 {
					base := 1 << attempt // 1,2,4,8...
					jitter := rand.Intn(base + 1)
					waitSec = base + jitter
					if waitSec > 60 {
						waitSec = 60
					}
				}
				logger.Warn("API returned 429 for day_offset=%d (attempt %d/%d) - sleeping %ds before retry", dayOffset, attempt+1, maxRetries, waitSec)
				time.Sleep(time.Duration(waitSec) * time.Second)
				attempt++
				if attempt > maxRetries {
					consecutiveEmpty++
					break
				}
				continue
			}

			if resp.StatusCode != http.StatusOK {
				_ = resp.Body.Close()
				logger.Warn("API call for day_offset=%d returned HTTP %d", dayOffset, resp.StatusCode)
				time.Sleep(150 * time.Millisecond)
				consecutiveEmpty++
				if consecutiveEmpty >= 3 {
					break
				}
				break
			}

			// success
			break
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			fmt.Printf("ERROR: Error reading response for day_offset=%d: %v\n", dayOffset, err)
			time.Sleep(150 * time.Millisecond)
			consecutiveEmpty++
			if consecutiveEmpty >= 3 {
				break
			}
			continue
		}

		var apiResp HistoricalResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			fmt.Printf("ERROR: Error parsing JSON for day_offset=%d: %v\n", dayOffset, err)
			time.Sleep(150 * time.Millisecond)
			consecutiveEmpty++
			if consecutiveEmpty >= 3 {
				break
			}
			continue
		}

		observations := parseDeviceObservations(apiResp.Obs)
		if len(observations) == 0 {
			// No data for this day; increment consecutive empty counter and possibly stop
			consecutiveEmpty++
			if consecutiveEmpty >= 3 {
				if logLevel == "debug" {
					fmt.Printf("DEBUG: %d consecutive empty days encountered, stopping\n", consecutiveEmpty)
				}
				break
			}
		} else {
			allObservations = append(allObservations, observations...)
			consecutiveEmpty = 0
		}

		// Respectful pause between requests
		time.Sleep(150 * time.Millisecond)
	}

	// Sort newest first
	sort.Slice(allObservations, func(i, j int) bool {
		return allObservations[i].Timestamp > allObservations[j].Timestamp
	})

	// Deduplicate by timestamp
	uniqueObs := make([]*Observation, 0, len(allObservations))
	seen := make(map[int64]bool)
	for _, obs := range allObservations {
		if !seen[obs.Timestamp] {
			seen[obs.Timestamp] = true
			uniqueObs = append(uniqueObs, obs)
		}
	}

	// Limit to maxPoints if requested
	if maxPoints > 0 && len(uniqueObs) > maxPoints {
		uniqueObs = uniqueObs[:maxPoints]
	}

	// Reduction pipeline
	if strings.ToLower(reduceMethod) == "timebin" {
		// Keep recent high-resolution points covering the last keepRecentHours
		var recent []*Observation
		var older []*Observation

		if keepRecentHours > 0 && len(uniqueObs) > 0 {
			newestTs := uniqueObs[0].Timestamp
			cutoff := newestTs - int64(keepRecentHours*3600)
			for _, o := range uniqueObs {
				if o.Timestamp >= cutoff {
					recent = append(recent, o)
				} else {
					older = append(older, o)
				}
			}
		} else {
			older = uniqueObs
		}

		// Bin older observations into bins of binMinutes
		binSec := int64(binMinutes * 60)
		if binSec <= 0 {
			binSec = 600 // default 10 minutes
		}

		// Map from bucket -> slice of observations (bucket key = floor(timestamp/binSec))
		buckets := make(map[int64][]*Observation)
		var bucketKeys []int64

		// older is newest-first; for consistent bin assignment, iterate oldest-first
		for i := len(older) - 1; i >= 0; i-- {
			o := older[i]
			key := o.Timestamp / binSec
			if _, ok := buckets[key]; !ok {
				bucketKeys = append(bucketKeys, key)
			}
			buckets[key] = append(buckets[key], o)
		}

		// Aggregate each bucket
		var binned []*Observation
		for _, key := range bucketKeys {
			group := buckets[key]
			if len(group) == 0 {
				continue
			}
			// aggregate per-field
			var tsSum int64 = 0
			var windLullSum float64 = 0
			var windAvgSum float64 = 0
			var windGustMax float64 = 0
			var windDirSum float64 = 0
			var pressureSum float64 = 0
			var tempSum float64 = 0
			var rhSum float64 = 0
			var illumSum float64 = 0
			var solarSum float64 = 0
			var rainAccumSum float64 = 0
			var rainDailySum float64 = 0
			var lightningAvgSum float64 = 0
			var batterySum float64 = 0
			var uvSum = 0
			var precipType = 0
			var lightningCountSum = 0
			var reportIntervalSum = 0
			count := len(group)
			for _, g := range group {
				tsSum += g.Timestamp
				windLullSum += g.WindLull
				windAvgSum += g.WindAvg
				if g.WindGust > windGustMax {
					windGustMax = g.WindGust
				}
				windDirSum += g.WindDirection
				pressureSum += g.StationPressure
				tempSum += g.AirTemperature
				rhSum += g.RelativeHumidity
				illumSum += g.Illuminance
				solarSum += g.SolarRadiation
				rainAccumSum += g.RainAccumulated
				rainDailySum += g.RainDailyTotal
				lightningAvgSum += g.LightningStrikeAvg
				batterySum += g.Battery
				uvSum += g.UV
				precipType = g.PrecipitationType // use last observed precip type in bucket
				lightningCountSum += g.LightningStrikeCount
				reportIntervalSum += g.ReportInterval
			}
			// Use bucket start timestamp for the aggregated point
			aggTs := key * binSec
			avg := &Observation{
				Timestamp:            aggTs,
				WindLull:             windLullSum / float64(count),
				WindAvg:              windAvgSum / float64(count),
				WindGust:             windGustMax,
				WindDirection:        windDirSum / float64(count),
				StationPressure:      pressureSum / float64(count),
				AirTemperature:       tempSum / float64(count),
				RelativeHumidity:     rhSum / float64(count),
				Illuminance:          illumSum / float64(count),
				UV:                   uvSum / int(count),
				SolarRadiation:       solarSum / float64(count),
				RainAccumulated:      rainAccumSum,
				RainDailyTotal:       rainDailySum / float64(count),
				PrecipitationType:    precipType,
				LightningStrikeAvg:   lightningAvgSum / float64(count),
				LightningStrikeCount: lightningCountSum,
				Battery:              batterySum / float64(count),
				ReportInterval:       reportIntervalSum / int(count),
			}
			binned = append(binned, avg)
		}

		// binned is oldest-first; reverse to newest-first
		for i, j := 0, len(binned)-1; i < j; i, j = i+1, j-1 {
			binned[i], binned[j] = binned[j], binned[i]
		}

		// Combine recent (newest-first) + binned (newest-first)
		combined := append(recent, binned...)
		logger.Info("Historical points fetched: %d, after timebin reduction: %d (keepRecent=%dh, bin=%dm)", len(uniqueObs), len(combined), keepRecentHours, binMinutes)
		uniqueObs = combined
	} else if strings.ToLower(reduceMethod) == "factor" {
		// Existing fixed-group averaging behavior
		if reduceFactor > 1 && len(uniqueObs) > 0 {
			reduced := make([]*Observation, 0, (len(uniqueObs)+reduceFactor-1)/reduceFactor)
			for i := 0; i < len(uniqueObs); i += reduceFactor {
				end := i + reduceFactor
				if end > len(uniqueObs) {
					end = len(uniqueObs)
				}
				group := uniqueObs[i:end]
				var tsSum int64 = 0
				var windLullSum float64 = 0
				var windAvgSum float64 = 0
				var windGustSum float64 = 0
				var windDirSum float64 = 0
				var pressureSum float64 = 0
				var tempSum float64 = 0
				var rhSum float64 = 0
				var illumSum float64 = 0
				var solarSum float64 = 0
				var rainAccumSum float64 = 0
				var rainDailySum float64 = 0
				var lightningAvgSum float64 = 0
				var batterySum float64 = 0
				var uvSum = 0
				var precipSum = 0
				var lightningCountSum = 0
				var reportIntervalSum = 0
				count := len(group)
				for _, g := range group {
					tsSum += g.Timestamp
					windLullSum += g.WindLull
					windAvgSum += g.WindAvg
					windGustSum += g.WindGust
					windDirSum += g.WindDirection
					pressureSum += g.StationPressure
					tempSum += g.AirTemperature
					rhSum += g.RelativeHumidity
					illumSum += g.Illuminance
					solarSum += g.SolarRadiation
					rainAccumSum += g.RainAccumulated
					rainDailySum += g.RainDailyTotal
					lightningAvgSum += g.LightningStrikeAvg
					batterySum += g.Battery
					uvSum += g.UV
					precipSum += g.PrecipitationType
					lightningCountSum += g.LightningStrikeCount
					reportIntervalSum += g.ReportInterval
				}
				avg := &Observation{
					Timestamp:            tsSum / int64(count),
					WindLull:             windLullSum / float64(count),
					WindAvg:              windAvgSum / float64(count),
					WindGust:             windGustSum / float64(count),
					WindDirection:        windDirSum / float64(count),
					StationPressure:      pressureSum / float64(count),
					AirTemperature:       tempSum / float64(count),
					RelativeHumidity:     rhSum / float64(count),
					Illuminance:          illumSum / float64(count),
					UV:                   uvSum / int(count),
					SolarRadiation:       solarSum / float64(count),
					RainAccumulated:      rainAccumSum / float64(count),
					RainDailyTotal:       rainDailySum / float64(count),
					PrecipitationType:    precipSum / int(count),
					LightningStrikeAvg:   lightningAvgSum / float64(count),
					LightningStrikeCount: lightningCountSum / int(count),
					Battery:              batterySum / float64(count),
					ReportInterval:       reportIntervalSum / int(count),
				}
				reduced = append(reduced, avg)
			}
			logger.Info("Historical points fetched: %d, reduced (factor=%d): %d", len(uniqueObs), reduceFactor, len(reduced))
			uniqueObs = reduced
		}
	} else {
		// Unknown method: no reduction
		fmt.Printf("INFO: Unknown history reduce method '%s' - skipping reduction\n", reduceMethod)
	}

	return uniqueObs, nil
}

// StationStatus represents the status information from the TempestWX station status page
type StationStatus struct {
	HubNetworkStatus    string `json:"hubNetworkStatus"`
	HubLastStatus       string `json:"hubLastStatus"`
	HubWiFiSignal       string `json:"hubWiFiSignal"`
	HubSerialNumber     string `json:"hubSerialNumber"`
	HubFirmware         string `json:"hubFirmware"`
	HubUptime           string `json:"hubUptime"`
	DeviceNetworkStatus string `json:"deviceNetworkStatus"`
	DeviceLastObs       string `json:"deviceLastObs"`
	DeviceSignal        string `json:"deviceSignal"`
	DeviceSerialNumber  string `json:"deviceSerialNumber"`
	DeviceFirmware      string `json:"deviceFirmware"`
	DeviceUptime        string `json:"deviceUptime"`
	BatteryVoltage      string `json:"batteryVoltage"`
	BatteryStatus       string `json:"batteryStatus"`
	SensorStatus        string `json:"sensorStatus"`
	// Metadata for tracking data source and freshness
	DataSource      string `json:"dataSource"`      // "web-scraped", "api", "fallback"
	LastScraped     string `json:"lastScraped"`     // ISO 8601 timestamp when data was scraped
	ScrapingEnabled bool   `json:"scrapingEnabled"` // Whether web scraping is enabled
}

// GetStationStatus scrapes the TempestWX station status page for detailed device information
func GetStationStatus(stationID int, logLevel string) (*StationStatus, error) {
	url := fmt.Sprintf("https://tempestwx.com/settings/station/%d/status", stationID)

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Fetching station status from %s\n", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		if logLevel == "debug" {
			fmt.Printf("DEBUG: HTTP request failed: %v\n", err)
		}
		return nil, fmt.Errorf("failed to fetch station status: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		if logLevel == "debug" {
			fmt.Printf("DEBUG: HTTP request returned status %d\n", resp.StatusCode)
		}
		return nil, fmt.Errorf("station status request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if logLevel == "debug" {
			fmt.Printf("DEBUG: Failed to read response body: %v\n", err)
		}
		return nil, fmt.Errorf("failed to read station status response: %v", err)
	}

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Retrieved %d bytes of HTML content\n", len(body))
	}

	// Parse the HTML to extract status information
	status, err := parseStationStatusHTML(string(body), logLevel)
	if err != nil {
		if logLevel == "debug" {
			fmt.Printf("DEBUG: HTML parsing failed: %v\n", err)
		}
		return nil, err
	}

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Parsed station status - Battery: %s, Device Uptime: %s\n", status.BatteryVoltage, status.DeviceUptime)
	}
	return status, nil
}

// parseStationStatusHTML parses the HTML content from the TempestWX station status page
func parseStationStatusHTML(html string, logLevel string) (*StationStatus, error) {
	status := &StationStatus{}

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Parsing HTML content (%d bytes)\n", len(html))
		// Look for battery voltage section in the HTML for debugging
		if strings.Contains(html, "Battery Voltage") {
			start := strings.Index(html, "Battery Voltage")
			end := start + 200
			if end > len(html) {
				end = len(html)
			}
			fmt.Printf("DEBUG: Found Battery Voltage section: %s\n", html[start:end])
		}
	}

	// Extract data using the actual HTML structure: <span class="lv-param-label">Label</span>...<span class="lv-value-display">Value</span>

	// Extract Battery Voltage - pattern: "Good (2.69v)" from Battery Voltage row
	// Handle multi-line whitespace between tags
	batteryPattern := regexp.MustCompile(`<span class="lv-param-label">Battery Voltage</span>.*?<span class="lv-value-display"[^>]*>\s*([^<(]*?)\s*\(([0-9.]+)v\)\s*</span>`)
	if match := batteryPattern.FindStringSubmatch(html); len(match) >= 3 {
		status.BatteryStatus = strings.TrimSpace(match[1]) // "Good"
		status.BatteryVoltage = match[2] + "V"             // "2.69V"
		if logLevel == "debug" {
			fmt.Printf("DEBUG: Found battery info - Status: %s, Voltage: %s\n", status.BatteryStatus, status.BatteryVoltage)
		}
	} else {
		// Try alternative pattern with more flexible whitespace handling
		altBatteryPattern := regexp.MustCompile(`(?s)Battery Voltage.*?([A-Za-z]+)\s+\(([0-9.]+)v\)`)
		if match := altBatteryPattern.FindStringSubmatch(html); len(match) >= 3 {
			status.BatteryStatus = strings.TrimSpace(match[1]) // "Good"
			status.BatteryVoltage = match[2] + "V"             // "2.69V"
			if logLevel == "debug" {
				fmt.Printf("DEBUG: Found battery info (alt pattern) - Status: %s, Voltage: %s\n", status.BatteryStatus, status.BatteryVoltage)
			}
		} else {
			if logLevel == "debug" {
				fmt.Printf("DEBUG: Battery voltage pattern not found\n")
			}
		}
	}

	// Extract Uptime values - look for both Hub and Device uptime patterns
	uptimePattern := regexp.MustCompile(`<span class="lv-param-label">Uptime</span>.*?<span class="lv-value-display">([0-9]+d\s+[0-9]+h\s+[0-9]+m\s+[0-9]+s)</span>`)
	uptimeMatches := uptimePattern.FindAllStringSubmatch(html, -1)

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Found %d uptime matches\n", len(uptimeMatches))
	}

	if len(uptimeMatches) >= 2 {
		// First uptime is Hub, second is Device (based on HTML order)
		status.HubUptime = uptimeMatches[0][1]    // "63d 13h 6m 1s"
		status.DeviceUptime = uptimeMatches[1][1] // "128d 3h 30m 29s"
	} else if len(uptimeMatches) == 1 {
		// If only one, assume it's device uptime
		status.DeviceUptime = uptimeMatches[0][1]
	}

	// Extract Network Status (appears twice - Hub and Device)
	networkPattern := regexp.MustCompile(`<span class="lv-param-label">Network Status</span>.*?<span class="lv-value-display"[^>]*>.*?([A-Za-z]+)\s*</span>`)
	networkMatches := networkPattern.FindAllStringSubmatch(html, -1)
	if len(networkMatches) >= 2 {
		status.HubNetworkStatus = strings.TrimSpace(networkMatches[0][1])    // First "Online"
		status.DeviceNetworkStatus = strings.TrimSpace(networkMatches[1][1]) // Second "Online"
	}

	// Extract Wi-Fi Signal (Hub)
	wifiPattern := regexp.MustCompile(`<span class="lv-param-label">Wi-Fi Signal \(RSSI\)</span>.*?<span class="lv-value-display">([^<]+)</span>`)
	if match := wifiPattern.FindStringSubmatch(html); len(match) >= 2 {
		status.HubWiFiSignal = strings.TrimSpace(match[1]) // "Strong (-32)"
	}

	// Extract Device Signal
	deviceSignalPattern := regexp.MustCompile(`<span class="lv-param-label">Device Signal \(RSSI\)</span>.*?<span class="lv-value-display">([^<]+)</span>`)
	if match := deviceSignalPattern.FindStringSubmatch(html); len(match) >= 2 {
		status.DeviceSignal = strings.TrimSpace(match[1]) // "Good (-63)"
	}

	// Extract Serial Numbers
	serialPattern := regexp.MustCompile(`<span class="lv-param-label">Serial Number</span>.*?<span class="lv-value-display">([^<]+)</span>`)
	serialMatches := serialPattern.FindAllStringSubmatch(html, -1)
	if len(serialMatches) >= 2 {
		for _, match := range serialMatches {
			serialNum := strings.TrimSpace(match[1])
			if strings.HasPrefix(serialNum, "HB-") {
				status.HubSerialNumber = serialNum
			} else if strings.HasPrefix(serialNum, "ST-") {
				status.DeviceSerialNumber = serialNum
			}
		}
	}

	// Extract Firmware Revisions
	firmwarePattern := regexp.MustCompile(`<span class="lv-param-label">Firmware Revision</span>.*?<span class="lv-value-display">([^<]+)</span>`)
	firmwareMatches := firmwarePattern.FindAllStringSubmatch(html, -1)
	if len(firmwareMatches) >= 2 {
		status.HubFirmware = "v" + strings.TrimSpace(firmwareMatches[0][1])    // "v329"
		status.DeviceFirmware = "v" + strings.TrimSpace(firmwareMatches[1][1]) // "v179"
	}

	// Extract Last Status Message (Hub)
	lastStatusPattern := regexp.MustCompile(`<span class="lv-param-label">Last Status Message</span>.*?<span class="lv-value-display">([^<]+)</span>`)
	if match := lastStatusPattern.FindStringSubmatch(html); len(match) >= 2 {
		status.HubLastStatus = strings.TrimSpace(match[1]) // "09/17/2025 5:26:08 pm"
	}

	// Extract Last Observation (Device)
	lastObsPattern := regexp.MustCompile(`<span class="lv-param-label">Last Observation</span>.*?<span class="lv-value-display">([^<]+)</span>`)
	if match := lastObsPattern.FindStringSubmatch(html); len(match) >= 2 {
		status.DeviceLastObs = strings.TrimSpace(match[1]) // "09/17/2025 5:25:45 pm"
	}

	// Extract Sensor Status
	sensorStatusPattern := regexp.MustCompile(`<span class="lv-param-label">Sensor Status</span>.*?<span class="lv-value-display"[^>]*>.*?([A-Za-z]+)\s*</span>`)
	if match := sensorStatusPattern.FindStringSubmatch(html); len(match) >= 2 {
		status.SensorStatus = strings.TrimSpace(match[1]) // "Good"
	}

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Final parsed status - Battery: %s, DeviceUptime: %s, HubUptime: %s\n",
			status.BatteryVoltage, status.DeviceUptime, status.HubUptime)
	}

	return status, nil
}

// GetStationStatusWithBrowser uses a headless browser to scrape the TempestWX status page
// This version waits for JavaScript to load the content before parsing
func GetStationStatusWithBrowser(stationID int, logLevel string) (*StationStatus, error) {
	url := fmt.Sprintf("https://tempestwx.com/settings/station/%d/status", stationID)

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Fetching station status with headless browser from %s\n", url)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create headless browser context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	var htmlContent string

	// Run browser tasks
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		// Wait for the diagnostic info div to be populated
		chromedp.WaitVisible(`#diagnostic-info ul.sw-list`, chromedp.ByID),
		// Wait a bit more for JavaScript to finish loading content
		chromedp.Sleep(3*time.Second),
		// Get the HTML content of the diagnostic info section
		chromedp.OuterHTML(`#diagnostic-info`, &htmlContent, chromedp.ByID),
	)

	if err != nil {
		if logLevel == "debug" {
			fmt.Printf("DEBUG: Headless browser failed: %v\n", err)
		}
		return nil, fmt.Errorf("failed to scrape status with browser: %v", err)
	}

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Retrieved %d bytes of HTML content from browser\n", len(htmlContent))
	}

	// Parse the HTML content
	status, err := parseStationStatusHTML(htmlContent, logLevel)
	if err != nil {
		if logLevel == "debug" {
			fmt.Printf("DEBUG: HTML parsing failed: %v\n", err)
		}
		return nil, err
	}

	// Add metadata about the scraping
	status.LastScraped = time.Now().UTC().Format(time.RFC3339)
	status.ScrapingEnabled = true

	if logLevel == "debug" {
		fmt.Printf("DEBUG: Browser-scraped station status - Battery: %s, Device Uptime: %s\n",
			status.BatteryVoltage, status.DeviceUptime)
	}

	return status, nil
}
