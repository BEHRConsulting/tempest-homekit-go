package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"tempest-homekit-go/pkg/weather"
)

type WebServer struct {
	port                 string
	server               *http.Server
	weatherData          *weather.Observation
	forecastData         *weather.ForecastResponse
	homekitStatus        map[string]interface{}
	dataHistory          []weather.Observation
	maxHistorySize       int
	stationName          string
	startTime            time.Time
	historicalDataLoaded bool
	historicalDataCount  int
	historyLoadingProgress struct {
		isLoading    bool
		currentStep  int
		totalSteps   int
		description  string
	}
	mu                   sync.RWMutex
}

type WeatherResponse struct {
	Temperature          float64 `json:"temperature"`
	Humidity             float64 `json:"humidity"`
	WindSpeed            float64 `json:"windSpeed"`
	WindGust             float64 `json:"windGust"`
	WindDirection        float64 `json:"windDirection"`
	RainAccum            float64 `json:"rainAccum"`
	RainDailyTotal       float64 `json:"rainDailyTotal"`
	PrecipitationType    int     `json:"precipitationType"`
	Pressure             float64 `json:"pressure"`
	PressureCondition    string  `json:"pressure_condition"`
	PressureTrend        string  `json:"pressure_trend"`
	WeatherForecast      string  `json:"weather_forecast"`
	Illuminance          float64 `json:"illuminance"`
	UV                   float64 `json:"uv"`
	Battery              float64 `json:"battery"`
	LightningStrikeAvg   float64 `json:"lightningStrikeAvg"`
	LightningStrikeCount int     `json:"lightningStrikeCount"`
	LastUpdate           string  `json:"lastUpdate"`
}

type StatusResponse struct {
	Connected            bool                   `json:"connected"`
	LastUpdate           string                 `json:"lastUpdate"`
	Uptime               string                 `json:"uptime"`
	StationName          string                 `json:"stationName,omitempty"`
	HomeKit              map[string]interface{} `json:"homekit"`
	DataHistory          []WeatherResponse      `json:"dataHistory"`
	ObservationCount     int                    `json:"observationCount"`
	HistoricalDataLoaded bool                   `json:"historicalDataLoaded"`
	HistoricalDataCount  int                    `json:"historicalDataCount"`
	HistoryLoadingProgress struct {
		IsLoading    bool   `json:"isLoading"`
		CurrentStep  int    `json:"currentStep"`
		TotalSteps   int    `json:"totalSteps"`
		Description  string `json:"description"`
	} `json:"historyLoadingProgress"`
	Forecast             *weather.ForecastResponse `json:"forecast,omitempty"`
}

// Precipitation type helper function
func getPrecipitationTypeDescription(precipType int) string {
	switch precipType {
	case 0:
		return "None"
	case 1:
		return "Rain"
	case 2:
		return "Hail"
	case 3:
		return "Rain + Hail"
	default:
		return "Unknown"
	}
}

// Calculate daily rain accumulation from historical data
func (ws *WebServer) calculateDailyRainAccumulation() float64 {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if len(ws.dataHistory) == 0 {
		log.Printf("DEBUG: No data history available for daily rain calculation")
		return 0.0
	}

	// Get the start of the current local day
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Find observations from today
	var dailyObservations []weather.Observation
	for _, obs := range ws.dataHistory {
		obsTime := time.Unix(obs.Timestamp, 0)
		if obsTime.After(startOfDay) || obsTime.Equal(startOfDay) {
			dailyObservations = append(dailyObservations, obs)
		}
	}

	log.Printf("DEBUG: Daily rain calculation - Total history: %d, Today's observations: %d, Start of day: %s",
		len(ws.dataHistory), len(dailyObservations), startOfDay.Format("2006-01-02 15:04:05"))

	if len(dailyObservations) == 0 {
		log.Printf("DEBUG: No observations found for today")
		return 0.0
	}

	// Sort by timestamp to ensure we process chronologically
	sort.Slice(dailyObservations, func(i, j int) bool {
		return dailyObservations[i].Timestamp < dailyObservations[j].Timestamp
	})

	// Calculate total rain for the day
	// The rain_accumulated field from Tempest represents cumulative rain since station started
	// To get daily total, we find the difference between current and earliest reading today
	if len(dailyObservations) == 1 {
		// Only one observation today, so we can't calculate a difference
		// Return the accumulated value if it seems reasonable for a daily total
		singleValue := dailyObservations[0].RainAccumulated
		log.Printf("DEBUG: Only one observation today, rain value: %.3f", singleValue)
		if singleValue <= 10.0 { // Reasonable daily limit in inches
			return singleValue
		}
		return 0.0
	}

	// Find the earliest and latest readings for today
	earliestToday := dailyObservations[0].RainAccumulated
	latestToday := dailyObservations[len(dailyObservations)-1].RainAccumulated

	log.Printf("DEBUG: Daily rain calculation - Earliest: %.3f, Latest: %.3f, Observations count: %d",
		earliestToday, latestToday, len(dailyObservations))

	// If latest is greater than earliest, we have rain accumulation for the day
	if latestToday >= earliestToday {
		dailyTotal := latestToday - earliestToday
		log.Printf("DEBUG: Daily rain total calculated: %.3f inches", dailyTotal)
		// Sanity check: daily total shouldn't exceed reasonable limits
		if dailyTotal <= 20.0 { // 20 inches would be extreme but possible
			return dailyTotal
		}
		log.Printf("DEBUG: Daily rain total exceeds sanity limit (%.3f > 20.0), returning 0", dailyTotal)
	}

	// If we can't calculate a reliable daily total, return 0
	log.Printf("DEBUG: Cannot calculate reliable daily total, returning 0")
	return 0.0
}

// Pressure analysis functions
func getPressureDescription(pressure float64) string {
	if pressure < 980 {
		return "Low"
	} else if pressure > 1020 {
		return "High"
	}
	return "Normal"
}

func getPressureTrend(dataHistory []weather.Observation) string {
	if len(dataHistory) < 2 {
		return "Stable"
	}

	// Look at last hour of data for trend
	recentData := make([]float64, 0)
	for i := len(dataHistory) - 1; i >= 0 && len(recentData) < 60; i-- {
		recentData = append([]float64{dataHistory[i].StationPressure}, recentData...)
	}

	if len(recentData) < 2 {
		return "Stable"
	}

	pressureChange := recentData[len(recentData)-1] - recentData[0]

	if pressureChange > 1.0 {
		return "Rising"
	} else if pressureChange < -1.0 {
		return "Falling"
	}
	return "Stable"
}

func getPressureWeatherForecast(pressure float64, trend string) string {
	switch trend {
	case "Rising":
		if pressure > 1013 {
			return "Fair Weather"
		} else {
			return "Storm Clearing"
		}
	case "Falling":
		if pressure < 1000 {
			return "Stormy"
		} else if pressure < 1013 {
			return "Unsettled"
		} else {
			return "Change Coming"
		}
	default: // Stable
		if pressure > 1020 {
			return "Fair Weather"
		} else if pressure < 1000 {
			return "Stormy"
		} else {
			return "Settled"
		}
	}
}

func NewWebServer(port string) *WebServer {
	ws := &WebServer{
		port:           port,
		maxHistorySize: 1000,
		dataHistory:    make([]weather.Observation, 0, 1000),
		startTime:      time.Now(),
		homekitStatus: map[string]interface{}{
			"bridge":      false,
			"accessories": 0,
			"pin":         "00102003",
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.handleDashboard)
	mux.HandleFunc("/api/weather", ws.handleWeatherAPI)
	mux.HandleFunc("/api/status", ws.handleStatusAPI)

	ws.server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return ws
}

func (ws *WebServer) Start() error {
	log.Printf("Starting web server on port %s", ws.port)
	return ws.server.ListenAndServe()
}

func (ws *WebServer) UpdateWeather(obs *weather.Observation) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.weatherData = obs

	// Add to history
	ws.dataHistory = append(ws.dataHistory, *obs)
	if len(ws.dataHistory) > ws.maxHistorySize {
		ws.dataHistory = ws.dataHistory[1:]
	}
}

func (ws *WebServer) UpdateHomeKitStatus(status map[string]interface{}) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for k, v := range status {
		ws.homekitStatus[k] = v
	}
}

func (ws *WebServer) SetStationName(name string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.stationName = name
}

func (ws *WebServer) SetHistoricalDataStatus(count int) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.historicalDataLoaded = true
	ws.historicalDataCount = count
}

func (ws *WebServer) SetHistoryLoadingProgress(currentStep, totalSteps int, description string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.historyLoadingProgress.isLoading = true
	ws.historyLoadingProgress.currentStep = currentStep
	ws.historyLoadingProgress.totalSteps = totalSteps
	ws.historyLoadingProgress.description = description
}

func (ws *WebServer) SetHistoryLoadingComplete() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.historyLoadingProgress.isLoading = false
	ws.historyLoadingProgress.currentStep = 0
	ws.historyLoadingProgress.totalSteps = 0
	ws.historyLoadingProgress.description = ""
}

func (ws *WebServer) UpdateForecast(forecast *weather.ForecastResponse) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.forecastData = forecast
}

func (ws *WebServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Handle test-api.html specifically
	if r.URL.Path == "/test-api.html" {
		http.ServeFile(w, r, "./test-api.html")
		return
	}

	// Handle static files - support both /static/ and /pkg/web/static/ paths
	if strings.HasPrefix(r.URL.Path, "/pkg/web/static/") || strings.HasPrefix(r.URL.Path, "/static/") {
		var filename string
		if strings.HasPrefix(r.URL.Path, "/pkg/web/static/") {
			filename = strings.TrimPrefix(r.URL.Path, "/pkg/web/static/")
		} else {
			filename = strings.TrimPrefix(r.URL.Path, "/static/")
		}

		log.Printf("Static file request: %s (path: %s)", filename, r.URL.Path)

		// Serve the file from the physical directory
		filePath := "./pkg/web/static/" + filename

		// Set appropriate content type and cache-busting headers
		switch filename {
		case "styles.css":
			w.Header().Set("Content-Type", "text/css")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		case "script.js":
			w.Header().Set("Content-Type", "application/javascript")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		case "date-fns.min.js":
			w.Header().Set("Content-Type", "application/javascript")
		}

		log.Printf("Serving static file: %s", filePath)

		// Try to serve the file
		http.ServeFile(w, r, filePath)
		return
	}

	// Default to dashboard for root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl := ws.getDashboardHTML()
	w.Write([]byte(tmpl))
}

func (ws *WebServer) handleWeatherAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("API: Weather endpoint called from %s", r.RemoteAddr)

	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if ws.weatherData == nil {
		log.Printf("API: No weather data available")
		http.Error(w, "No weather data available", http.StatusServiceUnavailable)
		return
	}

	// Calculate pressure analysis with debug logging
	pressureCondition := getPressureDescription(ws.weatherData.StationPressure)
	pressureTrend := getPressureTrend(ws.dataHistory)
	weatherForecast := getPressureWeatherForecast(ws.weatherData.StationPressure, pressureTrend)

	// Calculate daily rain accumulation
	dailyRainTotal := ws.calculateDailyRainAccumulation()

	log.Printf("DEBUG: Pressure analysis calculated - Condition: %s, Trend: %s, Forecast: %s, Pressure: %.2f mb",
		pressureCondition, pressureTrend, weatherForecast, ws.weatherData.StationPressure)
	log.Printf("DEBUG: Rain data calculated - Current: %.3f in, Daily Total: %.3f in",
		ws.weatherData.RainAccumulated, dailyRainTotal)

	response := WeatherResponse{
		Temperature:          ws.weatherData.AirTemperature,
		Humidity:             ws.weatherData.RelativeHumidity,
		WindSpeed:            ws.weatherData.WindAvg,
		WindGust:             ws.weatherData.WindGust,
		WindDirection:        ws.weatherData.WindDirection,
		RainAccum:            ws.weatherData.RainAccumulated,
		RainDailyTotal:       dailyRainTotal,
		PrecipitationType:    ws.weatherData.PrecipitationType,
		Pressure:             ws.weatherData.StationPressure,
		PressureCondition:    pressureCondition,
		PressureTrend:        pressureTrend,
		WeatherForecast:      weatherForecast,
		Illuminance:          ws.weatherData.Illuminance,
		UV:                   ws.weatherData.UV,
		Battery:              ws.weatherData.Battery,
		LightningStrikeAvg:   ws.weatherData.LightningStrikeAvg,
		LightningStrikeCount: ws.weatherData.LightningStrikeCount,
		LastUpdate:           time.Unix(ws.weatherData.Timestamp, 0).Format(time.RFC3339),
	}

	log.Printf("DEBUG: Weather API response prepared - Temperature: %.1f°C, Humidity: %.1f%%, UV: %.1f, Illuminance: %.0f lux",
		response.Temperature, response.Humidity, response.UV, response.Illuminance)

	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) handleStatusAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.Printf("API: Status endpoint called from %s", r.RemoteAddr)

	ws.mu.RLock()
	defer ws.mu.RUnlock()

	connected := ws.weatherData != nil
	log.Printf("API: Status check - weatherData exists: %t", connected)
	lastUpdate := ""
	if ws.weatherData != nil {
		lastUpdate = time.Unix(ws.weatherData.Timestamp, 0).Format(time.RFC3339)
	}

	// Calculate uptime
	uptime := time.Since(ws.startTime)
	uptimeStr := fmt.Sprintf("%dh%dm%ds", int(uptime.Hours()), int(uptime.Minutes())%60, int(uptime.Seconds())%60)

	// Convert data history to response format
	history := make([]WeatherResponse, len(ws.dataHistory))
	for i, obs := range ws.dataHistory {
		history[i] = WeatherResponse{
			Temperature:          obs.AirTemperature,
			Humidity:             obs.RelativeHumidity,
			WindSpeed:            obs.WindAvg,
			WindGust:             obs.WindGust,
			WindDirection:        obs.WindDirection,
			RainAccum:            obs.RainAccumulated,
			RainDailyTotal:       0, // Historical data doesn't calculate individual daily totals
			PrecipitationType:    obs.PrecipitationType,
			Pressure:             obs.StationPressure,
			Illuminance:          obs.Illuminance,
			UV:                   obs.UV,
			Battery:              obs.Battery,
			LightningStrikeAvg:   obs.LightningStrikeAvg,
			LightningStrikeCount: obs.LightningStrikeCount,
			LastUpdate:           time.Unix(obs.Timestamp, 0).Format(time.RFC3339),
		}
	}

	response := StatusResponse{
		Connected:            connected,
		LastUpdate:           lastUpdate,
		Uptime:               uptimeStr,
		HomeKit:              ws.homekitStatus,
		DataHistory:          history,
		ObservationCount:     len(ws.dataHistory),
		HistoricalDataLoaded: ws.historicalDataLoaded,
		HistoricalDataCount:  ws.historicalDataCount,
	}
	
	// Add progress information
	response.HistoryLoadingProgress.IsLoading = ws.historyLoadingProgress.isLoading
	response.HistoryLoadingProgress.CurrentStep = ws.historyLoadingProgress.currentStep
	response.HistoryLoadingProgress.TotalSteps = ws.historyLoadingProgress.totalSteps
	response.HistoryLoadingProgress.Description = ws.historyLoadingProgress.description

	// Add forecast data if available
	response.Forecast = ws.forecastData

	// Add station name if available
	if ws.stationName != "" {
		response.StationName = ws.stationName
	}

	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) getDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tempest Weather Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            color: #333;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }

        /* Container responsive adjustments for larger screens */
        @media (min-width: 1200px) {
            .container {
                max-width: 1300px;
            }
        }

        @media (min-width: 1400px) {
            .container {
                max-width: 1500px;
            }
        }

        @media (min-width: 1600px) {
            .container {
                max-width: 1700px;
            }
        }

        @media (min-width: 1800px) {
            .container {
                max-width: 1900px;
            }
        }

        @media (min-width: 2000px) {
            .container {
                max-width: 2100px;
            }
        }

        @media (min-width: 2200px) {
            .container {
                max-width: none;
                padding: 20px 40px;
            }
        }

        .header {
            text-align: center;
            margin-bottom: 30px;
        }

        .header h1 {
            color: white;
            font-size: 2.5rem;
            margin-bottom: 10px;
            text-shadow: 0 2px 4px rgba(0,0,0,0.3);
        }

        .status {
            background: rgba(255,255,255,0.1);
            color: white;
            padding: 10px;
            border-radius: 8px;
            margin-bottom: 20px;
            text-align: center;
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        /* Responsive grid layouts for larger screens */
        @media (min-width: 1024px) {
            .grid {
                grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
                max-width: none;
            }
        }

        @media (min-width: 1200px) {
            .grid {
                grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
            }
        }

        @media (min-width: 1400px) {
            .grid {
                grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            }
        }

        @media (min-width: 1600px) {
            .grid {
                grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            }
        }

        @media (min-width: 1800px) {
            .grid {
                grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
            }
        }

        @media (min-width: 2000px) {
            .grid {
                grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
            }
        }

        @media (min-width: 2200px) {
            .grid {
                grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            }
        }

        /* Force more cards for ultra-wide screens like 2290px */
        @media (min-width: 2250px) {
            .grid {
                grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
                gap: 15px;
            }
            .container {
                max-width: none;
                padding: 20px 30px;
            }
        }

        .card {
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 15px rgba(0,0,0,0.2);
        }

        .card-header {
            display: flex;
            align-items: center;
            margin-bottom: 15px;
        }

        .card-icon {
            font-size: 2rem;
            margin-right: 10px;
        }

        .card-title {
            font-size: 1.2rem;
            font-weight: 600;
            color: #555;
        }

        .card-value {
            font-size: 2.5rem;
            font-weight: bold;
            color: #333;
            margin-bottom: 5px;
        }

        .card-unit {
            font-size: 1rem;
            color: #666;
            cursor: pointer;
            user-select: none;
        }

        .card-unit:hover {
            color: #007bff;
        }

        .lightning-info {
            margin: 10px 0;
            padding: 8px;
            background-color: rgba(255, 193, 7, 0.1);
            border-radius: 6px;
            border-left: 3px solid #ffc107;
        }
        
        .precipitation-type {
            margin: 5px 0;
            padding: 6px 8px;
            background-color: rgba(54, 162, 235, 0.1);
            border-radius: 6px;
            border-left: 3px solid #36a2eb;
            font-size: 0.85rem;
        }
        
        .precipitation-info {
            display: flex;
            align-items: center;
            color: #666;
            font-weight: 500;
        }

        .lightning-strikes, .lightning-distance {
            display: flex;
            align-items: center;
            font-size: 0.9rem;
            color: #666;
            margin-bottom: 4px;
        }

        .lightning-strikes:last-child, .lightning-distance:last-child {
            margin-bottom: 0;
        }

        .lightning-strikes span, .lightning-distance span {
            margin-left: 5px;
            font-weight: 600;
        }

        .wind-direction {
            display: flex;
            align-items: center;
            margin-top: 10px;
        }

        .direction-arrow {
            font-size: 1.5rem;
            margin-right: 5px;
        }

        .chart-container {
            height: 150px;
            margin-top: 15px;
        }

        .card-content {
            padding-top: 15px;
        }

        .info-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
            padding: 4px 0;
        }

        .info-row.clickable {
            cursor: pointer;
            transition: background-color 0.2s;
        }

        .info-row.clickable:hover {
            background-color: rgba(0, 123, 255, 0.1);
            border-radius: 4px;
        }

        .accessories-expanded {
            margin-top: 10px;
            padding: 10px;
            background-color: rgba(0, 123, 255, 0.05);
            border-radius: 6px;
            border-left: 3px solid #007bff;
        }

        .accessory-item {
            display: flex;
            align-items: center;
            padding: 4px 0;
            font-size: 0.85rem;
        }

        .accessory-item.disabled {
            opacity: 0.5;
        }

        .accessory-icon {
            margin-right: 8px;
            font-size: 1rem;
        }

        .accessory-name {
            color: #555;
            font-weight: 500;
        }

        .accessory-name.disabled {
            color: #999;
        }

        .accessory-status {
            margin-left: auto;
            font-size: 0.75rem;
            padding: 2px 6px;
            border-radius: 3px;
            font-weight: 500;
        }

        .accessory-status.enabled {
            background-color: rgba(40, 167, 69, 0.1);
            color: #28a745;
        }

        .accessory-status.disabled {
            background-color: rgba(220, 53, 69, 0.1);
            color: #dc3545;
        }

        .expand-icon {
            margin-left: auto;
            font-size: 0.8rem;
            color: #666;
            transition: transform 0.2s;
        }

        .info-label {
            font-weight: 500;
            color: #666;
            font-size: 0.9rem;
        }

        .info-value {
            font-weight: 600;
            color: #333;
            font-size: 0.9rem;
        }

        .footer {
            text-align: center;
            color: white;
            margin-top: 30px;
            font-size: 0.9rem;
        }

        .lux-context {
            position: relative;
            display: inline-block;
            margin-top: 5px;
        }

        .lux-tooltip {
            visibility: hidden;
            width: 300px;
            background-color: rgba(0, 0, 0, 0.9);
            color: #fff;
            text-align: left;
            border-radius: 6px;
            padding: 10px;
            position: absolute;
            z-index: 1;
            top: 100%;
            left: 100%;
            margin-top: 2px;
            margin-left: 2px;
            opacity: 0;
            transition: opacity 0.3s;
            font-size: 0.8rem;
        }

        .lux-tooltip.show {
            visibility: visible;
            opacity: 1;
        }

        .lux-tooltip-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
            padding-bottom: 5px;
            border-bottom: 1px solid #555;
        }

        .lux-tooltip-close {
            cursor: pointer;
            font-size: 1.2rem;
            color: #ccc;
            user-select: none;
            padding: 2px 6px;
            border-radius: 3px;
            transition: color 0.2s, background-color 0.2s;
        }

        .lux-tooltip-close:hover {
            color: #fff;
            background-color: rgba(255, 255, 255, 0.1);
        }

        .lux-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 5px;
        }

        .lux-table th, .lux-table td {
            border: 1px solid #555;
            padding: 4px 6px;
            text-align: left;
        }

        .lux-table th {
            background-color: #333;
            font-weight: bold;
        }

        .lux-table td:first-child {
            text-align: right;
            font-family: monospace;
        }

        .info-icon {
            cursor: pointer;
            user-select: none;
            margin-left: 5px;
        }

        .lux-description {
            font-size: 0.8rem;
            color: #666;
            margin-top: 5px;
            font-style: italic;
        }

        .heat-index-context {
            position: relative;
            display: inline-block;
        }

        .heat-index-tooltip {
            visibility: hidden;
            width: 350px;
            background-color: rgba(0, 0, 0, 0.9);
            color: #fff;
            text-align: left;
            border-radius: 6px;
            padding: 12px;
            position: absolute;
            z-index: 1;
            top: 100%;
            left: 100%;
            margin-top: 2px;
            margin-left: 2px;
            opacity: 0;
            transition: opacity 0.3s;
            font-size: 0.8rem;
        }

        .heat-index-tooltip.show {
            visibility: visible;
            opacity: 1;
        }

        .heat-index-tooltip-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
            padding-bottom: 5px;
            border-bottom: 1px solid #555;
        }

        .heat-index-tooltip-close {
            cursor: pointer;
            font-size: 1.2rem;
            color: #ccc;
            user-select: none;
            padding: 2px 6px;
            border-radius: 3px;
            transition: color 0.2s, background-color 0.2s;
        }

        .heat-index-tooltip-close:hover {
            color: #fff;
            background-color: rgba(255, 255, 255, 0.1);
        }

        .heat-index-table {
            width: 100%;
            border-collapse: collapse;
            margin: 8px 0;
        }

        .heat-index-table td {
            border: 1px solid #555;
            padding: 4px 6px;
            text-align: left;
            font-size: 0.8rem;
        }

        .heat-index-table td:first-child {
            font-family: monospace;
            font-weight: bold;
        }

        .uv-context {
            position: relative;
            display: inline-block;
            margin-top: 5px;
        }

        .uv-tooltip {
            visibility: hidden;
            width: 400px;
            background-color: rgba(0, 0, 0, 0.9);
            color: #fff;
            text-align: left;
            border-radius: 6px;
            padding: 10px;
            position: absolute;
            z-index: 1;
            top: 100%;
            left: 100%;
            margin-top: 2px;
            margin-left: 2px;
            opacity: 0;
            transition: opacity 0.3s;
            font-size: 0.8rem;
        }

        .uv-tooltip.show {
            visibility: visible;
            opacity: 1;
        }

        .uv-tooltip-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
            padding-bottom: 5px;
            border-bottom: 1px solid #555;
        }

        .uv-tooltip-close {
            cursor: pointer;
            font-size: 1.2rem;
            color: #ccc;
            user-select: none;
            padding: 2px 6px;
            border-radius: 3px;
            transition: color 0.2s, background-color 0.2s;
        }

        .uv-tooltip-close:hover {
            color: #fff;
            background-color: rgba(255, 255, 255, 0.1);
        }

        .uv-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 5px;
        }

        .uv-table th, .uv-table td {
            border: 1px solid #555;
            padding: 4px 6px;
            text-align: left;
        }

        .uv-table th {
            background-color: #333;
            font-weight: bold;
        }

        .uv-table td:first-child {
            text-align: center;
            font-family: monospace;
            font-weight: bold;
        }

        .uv-description {
            font-size: 0.8rem;
            color: #666;
            margin-top: 5px;
            font-style: italic;
        }

        .pressure-context {
            position: relative;
            display: inline-block;
            margin-top: 5px;
        }

        .pressure-tooltip {
            visibility: hidden;
            width: 450px;
            background-color: rgba(0, 0, 0, 0.9);
            color: #fff;
            text-align: left;
            border-radius: 6px;
            padding: 10px;
            position: absolute;
            z-index: 1;
            top: 100%;
            left: 100%;
            margin-top: 2px;
            margin-left: 2px;
            opacity: 0;
            transition: opacity 0.3s;
            font-size: 0.8rem;
        }

        .pressure-tooltip.show {
            visibility: visible;
            opacity: 1;
        }

        .pressure-tooltip-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
            padding-bottom: 5px;
            border-bottom: 1px solid #555;
        }

        .pressure-tooltip-close {
            cursor: pointer;
            font-size: 1.2rem;
            color: #ccc;
            user-select: none;
            padding: 2px 6px;
            border-radius: 3px;
            transition: color 0.2s, background-color 0.2s;
        }

        .pressure-tooltip-close:hover {
            color: #fff;
            background-color: rgba(255, 255, 255, 0.1);
        }

        .pressure-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 5px;
        }

        .pressure-table th, .pressure-table td {
            border: 1px solid #555;
            padding: 4px 6px;
            text-align: left;
        }

        .pressure-table th {
            background-color: #333;
            font-weight: bold;
        }

        .pressure-table td:first-child, .pressure-table td:nth-child(2) {
            text-align: center;
            font-family: monospace;
            font-weight: bold;
        }

        .pressure-description {
            font-size: 0.8rem;
            color: #666;
            margin-top: 5px;
            font-style: italic;
        }

        /* Forecast Card Styles */
        .forecast-current {
            margin-bottom: 15px;
            padding-bottom: 15px;
            border-bottom: 1px solid #ddd;
        }

        .forecast-current-main {
            display: flex;
            align-items: center;
            gap: 15px;
            margin-bottom: 10px;
        }

        .forecast-icon {
            font-size: 2.5rem;
            min-width: 60px;
            text-align: center;
        }

        .forecast-temp-container {
            flex: 1;
        }

        .forecast-temp {
            font-size: 2rem;
            font-weight: bold;
            color: #333;
        }

        .forecast-feels-like {
            font-size: 0.9rem;
            color: #666;
            margin-top: 2px;
        }

        .forecast-conditions {
            font-size: 1.1rem;
            color: #555;
            font-weight: 500;
            text-align: center;
            min-width: 100px;
        }

        .forecast-current-details {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 8px;
        }

        .forecast-detail {
            display: flex;
            justify-content: space-between;
            font-size: 0.9rem;
        }

        .forecast-label {
            color: #666;
        }

        .forecast-value {
            font-weight: 500;
            color: #333;
        }

        .forecast-daily {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .forecast-day {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 8px;
            background-color: #f8f9fa;
            border-radius: 6px;
            font-size: 0.9rem;
        }

        .forecast-day-name {
            font-weight: 500;
            min-width: 60px;
            color: #333;
        }

        .forecast-day-icon {
            font-size: 1.2rem;
            min-width: 30px;
            text-align: center;
        }

        .forecast-day-conditions {
            flex: 1;
            text-align: center;
            color: #555;
            font-size: 0.8rem;
            padding: 0 10px;
        }

        .forecast-day-temps {
            display: flex;
            gap: 5px;
            min-width: 70px;
            justify-content: flex-end;
        }

        .forecast-day-high {
            font-weight: bold;
            color: #333;
        }

        .forecast-day-low {
            color: #666;
        }

        .forecast-day-precip {
            font-size: 0.8rem;
            color: #4a90e2;
            min-width: 30px;
            text-align: right;
        }

        /* Mobile and narrow screen responsive styles for forecast card */
        @media (max-width: 768px) {
            .forecast-current-main {
                flex-direction: column;
                gap: 10px;
                text-align: center;
            }
            
            .forecast-icon {
                font-size: 2rem;
                min-width: auto;
            }
            
            .forecast-temp {
                font-size: 1.8rem;
            }
            
            .forecast-conditions {
                min-width: auto;
                margin-top: 5px;
            }
            
            .forecast-current-details {
                grid-template-columns: 1fr;
                gap: 6px;
            }
            
            .forecast-day {
                flex-wrap: wrap;
                gap: 5px;
                padding: 6px;
            }
            
            .forecast-day-name {
                min-width: 50px;
            }
            
            .forecast-day-conditions {
                font-size: 0.75rem;
                padding: 0 5px;
            }
            
            .forecast-day-temps {
                min-width: 60px;
            }
            
            .forecast-day-precip {
                min-width: 25px;
            }
        }

        @media (max-width: 480px) {
            .forecast-current-main {
                gap: 8px;
            }
            
            .forecast-icon {
                font-size: 1.8rem;
            }
            
            .forecast-temp {
                font-size: 1.6rem;
            }
            
            .forecast-feels-like {
                font-size: 0.8rem;
            }
            
            .forecast-day {
                flex-direction: column;
                align-items: flex-start;
                gap: 4px;
            }
            
            .forecast-day-name,
            .forecast-day-icon,
            .forecast-day-conditions,
            .forecast-day-temps,
            .forecast-day-precip {
                width: 100%;
                min-width: auto;
                text-align: left;
            }
            
            .forecast-day-temps {
                justify-content: flex-start;
            }
            
            .forecast-day-precip {
                text-align: left;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌤️ Tempest Weather Dashboard</h1>
            <div class="status" id="status">
                Connecting to weather station...
            </div>
        </div>

        <div class="grid">
            <div class="card" id="temperature-card">
                <div class="card-header">
                    <span class="card-icon">🌡️</span>
                    <span class="card-title">Temperature</span>
                </div>
                <div class="card-value" id="temperature">--</div>
                <div class="card-unit" id="temperature-unit" onclick="toggleUnit('temperature')">°C</div>
                <div class="chart-container">
                    <canvas id="temperature-chart"></canvas>
                </div>
            </div>

            <div class="card" id="humidity-card">
                <div class="card-header">
                    <span class="card-icon">💧</span>
                    <span class="card-title">Humidity</span>
                </div>
                <div class="card-value" id="humidity">--</div>
                <div class="card-unit">%</div>
                <div class="feels-like-info" style="margin-top: 10px; font-size: 0.9rem; color: #666;">
                    <div style="display: flex; align-items: center; gap: 8px;">
                        <span>Heat Index (feels like):</span>
                        <span id="heat-index" style="font-weight: 600; color: #333;">--</span>
                        <span class="info-icon" id="heat-index-info-icon" title="Click for heat index information">ℹ️</span>
                    </div>
                    <div class="heat-index-context" id="heat-index-context">
                        <div class="heat-index-tooltip" id="heat-index-tooltip">
                            <div class="heat-index-tooltip-header">
                                <strong>Heat Index Calculation:</strong>
                                <span class="heat-index-tooltip-close" id="heat-index-tooltip-close" title="Close">×</span>
                            </div>
                            <div style="margin-top: 8px; font-size: 0.85rem; line-height: 1.4;">
                                <p><strong>What is Heat Index?</strong><br>
                                The heat index combines air temperature and relative humidity to determine the human-perceived equivalent temperature.</p>
                                
                                <p><strong>Calculation:</strong><br>
                                Uses the official NOAA formula with temperature ≥80°F (26.7°C) and humidity ≥40%.</p>
                                
                                <p><strong>Heat Index Categories:</strong></p>
                                <table class="heat-index-table">
                                    <tr><td>80-90°F (27-32°C)</td><td>Caution - Fatigue possible</td></tr>
                                    <tr><td>90-105°F (32-41°C)</td><td>Extreme caution - Heat cramps possible</td></tr>
                                    <tr><td>105-130°F (41-54°C)</td><td>Danger - Heat exhaustion likely</td></tr>
                                    <tr><td>130°F+ (54°C+)</td><td>Extreme danger - Heat stroke imminent</td></tr>
                                </table>
                                
                                <p style="margin-top: 8px; font-style: italic; font-size: 0.8rem;">
                                Note: If conditions don't meet the threshold, actual temperature is displayed.
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="chart-container">
                    <canvas id="humidity-chart"></canvas>
                </div>
            </div>

            <div class="card" id="wind-card">
                <div class="card-header">
                    <span class="card-icon">🌬️</span>
                    <span class="card-title">Wind</span>
                </div>
                <div class="card-value" id="wind-speed">--</div>
                <div class="card-unit" id="wind-unit" onclick="toggleUnit('wind')">mph</div>
                <div class="wind-direction">
                    <span class="direction-arrow" id="wind-arrow">↑</span>
                    <span id="wind-direction">--</span>
                </div>
                <div class="chart-container">
                    <canvas id="wind-chart"></canvas>
                </div>
            </div>

            <div class="card" id="rain-card">
                <div class="card-header">
                    <span class="card-icon">🌧️</span>
                    <span class="card-title">Rain & Lightning</span>
                </div>
                <div class="card-value" id="rain">--</div>
                <div class="card-unit" id="rain-unit" onclick="toggleUnit('rain')">in</div>
                <div class="daily-rain-info" style="margin-top: 8px; padding: 6px 8px; background-color: rgba(54, 162, 235, 0.1); border-radius: 6px; border-left: 3px solid #36a2eb;">
                    <div class="daily-rain-content" style="display: flex; justify-content: space-between; align-items: center; font-size: 0.85rem;">
                        <span style="color: #666; font-weight: 500;">Today Total:</span>
                        <span id="daily-rain-total" style="color: #333; font-weight: 600;">--</span>
                    </div>
                </div>
                <div class="precipitation-type">
                    <div class="precipitation-info">💧 Type: <span id="precipitation-type">--</span></div>
                </div>
                <div class="lightning-info">
                    <div class="lightning-strikes">⚡ <span id="lightning-count">--</span> strikes</div>
                    <div class="lightning-distance">📏 <span id="lightning-distance">--</span> <span id="lightning-distance-unit">km</span></div>
                </div>
                <div class="chart-container">
                    <canvas id="rain-chart"></canvas>
                </div>
            </div>

            <div class="card" id="pressure-card">
                <div class="card-header">
                    <span class="card-icon">📊</span>
                    <span class="card-title">Pressure</span>
                </div>
                <div class="card-value" id="pressure">--</div>
                <div class="card-unit" id="pressure-unit" onclick="toggleUnit('pressure')">mb <span class="info-icon" id="pressure-info-icon" title="Click for pressure interpretation table">ℹ️</span></div>
                <div style="margin-top: 10px; padding: 8px; background-color: rgba(0, 123, 255, 0.05); border-radius: 6px; border-left: 3px solid #007bff;">
                    <div style="display: flex; justify-content: space-between; font-size: 0.9rem;">
                        <span style="color: #666; font-weight: 500;">Condition:</span>
                        <span id="pressure-condition" style="color: #333; font-weight: 600;">--</span>
                    </div>
                    <div style="display: flex; justify-content: space-between; font-size: 0.9rem; margin-top: 4px;">
                        <span style="color: #666; font-weight: 500;">Trend:</span>
                        <span id="pressure-trend" style="color: #333; font-weight: 600;">--</span>
                    </div>
                    <div style="display: flex; justify-content: space-between; font-size: 0.9rem; margin-top: 4px;">
                        <span style="color: #666; font-weight: 500;">Forecast:</span>
                        <span id="pressure-forecast" style="color: #333; font-weight: 600;">--</span>
                    </div>
                </div>
                <div class="pressure-context" id="pressure-context">
                    <div class="pressure-tooltip" id="pressure-tooltip">
                        <div class="pressure-tooltip-header">
                            <strong>Barometric Pressure Interpretation:</strong>
                            <span class="pressure-tooltip-close" id="pressure-tooltip-close" title="Close">×</span>
                        </div>
                        <table class="pressure-table">
                            <thead>
                                <tr>
                                    <th>Pressure (mb)</th>
                                    <th>Pressure (inHg)</th>
                                    <th>Condition</th>
                                    <th>Weather Trend</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr style="background-color: rgba(255, 0, 0, 0.1);"><td>&lt; 980</td><td>&lt; 28.94</td><td><strong>Very Low</strong></td><td>Stormy weather likely</td></tr>
                                <tr style="background-color: rgba(255, 165, 0, 0.1);"><td>980-1000</td><td>28.94-29.53</td><td><strong>Low</strong></td><td>Unsettled weather, possible storms</td></tr>
                                <tr style="background-color: rgba(255, 255, 0, 0.1);"><td>1000-1020</td><td>29.53-30.12</td><td><strong>Normal</strong></td><td>Fair weather conditions</td></tr>
                                <tr style="background-color: rgba(0, 255, 0, 0.1);"><td>1020-1040</td><td>30.12-30.71</td><td><strong>High</strong></td><td>Generally clear and stable</td></tr>
                                <tr style="background-color: rgba(0, 0, 255, 0.1);"><td>&gt; 1040</td><td>&gt; 30.71</td><td><strong>Very High</strong></td><td>Very stable, clear conditions</td></tr>
                            </tbody>
                        </table>
                        <div style="margin-top: 12px; padding: 8px; background-color: rgba(0, 123, 255, 0.1); border-radius: 4px;">
                            <strong>Pressure Trend Analysis & Weather Forecast:</strong>
                            <ul style="margin: 4px 0; padding-left: 16px; font-size: 0.85rem;">
                                <li><strong>Rising Rapidly:</strong> Quick improvement, clearing skies</li>
                                <li><strong>Rising:</strong> Improving weather, fair conditions ahead</li>
                                <li><strong>Steady:</strong> Current weather conditions will continue</li>
                                <li><strong>Falling:</strong> Weather deteriorating, clouds/rain possible</li>
                                <li><strong>Falling Rapidly:</strong> Storm approaching quickly, take precautions</li>
                            </ul>
                            <div style="margin-top: 8px; font-size: 0.8rem; font-style: italic;">
                                <strong>Combined Forecast:</strong> The condition shown combines current pressure with trend analysis for more accurate weather prediction.
                            </div>
                        </div>
                        <p style="margin-top: 8px; font-style: italic; font-size: 0.8rem;">
                        Note: Weather predictions based on pressure require considering local conditions and trends over time.
                        </p>
                    </div>
                </div>
                <div class="chart-container">
                    <canvas id="pressure-chart"></canvas>
                </div>
            </div>

            <div class="card" id="light-card">
                <div class="card-header">
                    <span class="card-icon">☀️</span>
                    <span class="card-title">Light</span>
                </div>
                <div class="card-value" id="illuminance">--</div>
                <div class="card-unit">lux <span class="info-icon" id="lux-info-icon" title="Click for lux reference table">ℹ️</span></div>
                <div class="lux-description" id="lux-description">--</div>
                <div class="lux-context" id="lux-context">
                    <div class="lux-tooltip" id="lux-tooltip">
                        <div class="lux-tooltip-header">
                            <strong>Lux Reference Table:</strong>
                            <span class="lux-tooltip-close" id="lux-tooltip-close" title="Close">×</span>
                        </div>
                        <table class="lux-table">
                            <thead>
                                <tr>
                                    <th>Lux</th>
                                    <th>Condition</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr><td>0.0001</td><td>Moonless, overcast night sky (starlight)</td></tr>
                                <tr><td>0.002</td><td>Moonless clear night sky with airglow</td></tr>
                                <tr><td>0.01</td><td>Quarter moon on a clear night</td></tr>
                                <tr><td>0.05–0.3</td><td>Full moon on a clear night</td></tr>
                                <tr><td>3.4</td><td>Dark limit of civil twilight under a clear sky</td></tr>
                                <tr><td>20–50</td><td>Public areas with dark surroundings</td></tr>
                                <tr><td>50</td><td>Family living room lights</td></tr>
                                <tr><td>80</td><td>Office building hallway/toilet lighting</td></tr>
                                <tr><td>100</td><td>Very dark overcast day</td></tr>
                                <tr><td>150</td><td>Train station platforms</td></tr>
                                <tr><td>320–500</td><td>Office lighting</td></tr>
                                <tr><td>400</td><td>Sunrise or sunset on a clear day</td></tr>
                                <tr><td>1000</td><td>Overcast day; typical TV studio lighting</td></tr>
                                <tr><td>10,000–25,000</td><td>Full daylight (not direct sun)</td></tr>
                                <tr><td>32,000–100,000</td><td>Direct sunlight</td></tr>
                            </tbody>
                        </table>
                    </div>
                </div>
                <div class="chart-container">
                    <canvas id="light-chart"></canvas>
                </div>
            </div>

            <div class="card" id="uv-card">
                <div class="card-header">
                    <span class="card-icon">🌞</span>
                    <span class="card-title">UV Index</span>
                </div>
                <div class="card-value" id="uv-index">--</div>
                <div class="card-unit">UVI <span class="info-icon" id="uv-info-icon" title="Click for UV Index exposure categories">ℹ️</span></div>
                <div class="uv-description" id="uv-description">--</div>
                <div class="uv-context" id="uv-context">
                    <div class="uv-tooltip" id="uv-tooltip">
                        <div class="uv-tooltip-header">
                            <strong>UV Index Exposure Categories:</strong>
                            <span class="uv-tooltip-close" id="uv-tooltip-close" title="Close">×</span>
                        </div>
                        <table class="uv-table">
                            <thead>
                                <tr>
                                    <th>UV Index</th>
                                    <th>Category</th>
                                    <th>Protection Advice</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr style="background-color: rgba(0, 255, 0, 0.1);"><td>0–2</td><td><strong>Low</strong></td><td>Low danger from the sun's UV rays for the average person.</td></tr>
                                <tr style="background-color: rgba(255, 255, 0, 0.1);"><td>3–5</td><td><strong>Moderate</strong></td><td>Moderate risk of harm from unprotected sun exposure.</td></tr>
                                <tr style="background-color: rgba(255, 165, 0, 0.1);"><td>6–7</td><td><strong>High</strong></td><td>High risk of harm from unprotected sun exposure. Protection against skin and eye damage is needed.</td></tr>
                                <tr style="background-color: rgba(255, 0, 0, 0.1);"><td>8–10</td><td><strong>Very High</strong></td><td>Very high risk of harm from unprotected sun exposure. Take extra precautions because unprotected skin and eyes will be damaged and can burn quickly.</td></tr>
                                <tr style="background-color: rgba(128, 0, 128, 0.1);"><td>11+</td><td><strong>Extreme</strong></td><td>Extreme risk of harm from unprotected sun exposure. Take all precautions because unprotected skin and eyes can burn in minutes.</td></tr>
                            </tbody>
                        </table>
                        <p style="margin-top: 8px; font-style: italic; font-size: 0.8rem;">
                        Source: U.S. Environmental Protection Agency UV Index scale
                        </p>
                    </div>
                </div>
                <div class="chart-container">
                    <canvas id="uv-chart"></canvas>
                </div>
            </div>
        </div>

        <!-- Information Cards -->
        <div class="grid">
            <!-- Tempest Forecast Card -->
            <div class="card" id="forecast-card">
                <div class="card-header">
                    <span class="card-icon">📅</span>
                    <span class="card-title">Tempest Forecast</span>
                </div>
                <div class="card-content">
                    <div class="forecast-current">
                        <div class="forecast-current-main">
                            <div class="forecast-icon" id="forecast-current-icon">--</div>
                            <div class="forecast-temp-container">
                                <div class="forecast-temp" id="forecast-current-temp">--°</div>
                                <div class="forecast-feels-like">Feels like <span id="forecast-current-feels-like">--°</span></div>
                            </div>
                            <div class="forecast-conditions" id="forecast-current-conditions">--</div>
                        </div>
                        <div class="forecast-current-details">
                            <div class="forecast-detail">
                                <span class="forecast-label">Humidity:</span>
                                <span class="forecast-value" id="forecast-current-humidity">--%</span>
                            </div>
                            <div class="forecast-detail">
                                <span class="forecast-label">Wind:</span>
                                <span class="forecast-value" id="forecast-current-wind">-- mph</span>
                            </div>
                            <div class="forecast-detail">
                                <span class="forecast-label">Pressure:</span>
                                <span class="forecast-value" id="forecast-current-pressure">-- mb</span>
                            </div>
                            <div class="forecast-detail">
                                <span class="forecast-label">Precip:</span>
                                <span class="forecast-value" id="forecast-current-precip">--%</span>
                            </div>
                        </div>
                    </div>
                    <div class="forecast-daily" id="forecast-daily-container">
                        <!-- Daily forecast items will be populated by JavaScript -->
                    </div>
                </div>
            </div>

            <div class="card" id="tempest-card">
                <div class="card-header">
                    <span class="card-icon">🌤️</span>
                    <span class="card-title">Tempest Station</span>
                </div>
                <div class="card-content">
                    <div class="info-row">
                        <span class="info-label">Status:</span>
                        <span class="info-value" id="tempest-status">Disconnected</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Station:</span>
                        <span class="info-value" id="tempest-station">--</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Last Update:</span>
                        <span class="info-value" id="tempest-last-update">--</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Uptime:</span>
                        <span class="info-value" id="tempest-uptime">--</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Data Points:</span>
                        <span class="info-value" id="tempest-data-count">--</span>
                    </div>
                    <div class="info-row" id="tempest-historical-row" style="display: none;">
                        <span class="info-label">Historical:</span>
                        <span class="info-value" id="tempest-historical-count">--</span>
                    </div>
                </div>
            </div>

            <div class="card" id="homekit-card">
                <div class="card-header">
                    <span class="card-icon">🏠</span>
                    <span class="card-title">HomeKit Bridge</span>
                </div>
                <div class="card-content">
                    <div class="info-row">
                        <span class="info-label">Status:</span>
                        <span class="info-value" id="homekit-status">Inactive</span>
                    </div>
                    <div class="info-row clickable" id="accessories-row">
                        <span class="info-label">Accessories:</span>
                        <span class="info-value" id="homekit-accessories">--</span>
                        <span class="expand-icon" id="accessories-expand-icon">▶</span>
                    </div>
                    <div class="accessories-expanded" id="accessories-expanded" style="display: none;">
                        <div id="accessories-list">
                            <!-- Accessories will be populated here -->
                        </div>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Bridge:</span>
                        <span class="info-value" id="homekit-bridge">--</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">PIN:</span>
                        <span class="info-value" id="homekit-pin">--</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="footer">
            <p>Last updated: <span id="last-update">--</span></p>
            <p>Tempest HomeKit Service</p>
        </div>
    <!-- External JavaScript Libraries -->
    <script src="https://unpkg.com/chart.js@4.4.4/dist/chart.umd.js"></script>
    <script src="https://unpkg.com/chartjs-adapter-date-fns@3.0.0/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
    
    <!-- Main Application Script -->
    <script src="pkg/web/static/script.js?v=` + fmt.Sprintf("%d", time.Now().UnixNano()) + `"></script>
</body>
</html>`
}

func (ws *WebServer) Stop() error {
	if ws.server != nil {
		return ws.server.Close()
	}
	return nil
}
