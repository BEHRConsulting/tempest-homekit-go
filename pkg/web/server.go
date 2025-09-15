package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"tempest-homekit-go/pkg/weather"
)

type WebServer struct {
	port           string
	server         *http.Server
	weatherData    *weather.Observation
	homekitStatus  map[string]interface{}
	dataHistory    []weather.Observation
	maxHistorySize int
	stationName    string
	startTime      time.Time
	mu             sync.RWMutex
}

type WeatherResponse struct {
	Temperature   float64 `json:"temperature"`
	Humidity      float64 `json:"humidity"`
	WindSpeed     float64 `json:"windSpeed"`
	WindGust      float64 `json:"windGust"`
	WindDirection float64 `json:"windDirection"`
	RainAccum     float64 `json:"rainAccum"`
	Pressure      float64 `json:"pressure"`
	Illuminance   float64 `json:"illuminance"`
	UV            float64 `json:"uv"`
	Battery       float64 `json:"battery"`
	LastUpdate    string  `json:"lastUpdate"`
}

type StatusResponse struct {
	Connected   bool                   `json:"connected"`
	LastUpdate  string                 `json:"lastUpdate"`
	Uptime      string                 `json:"uptime"`
	StationName string                 `json:"stationName,omitempty"`
	HomeKit     map[string]interface{} `json:"homekit"`
	DataHistory []WeatherResponse      `json:"dataHistory"`
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

func (ws *WebServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := ws.getDashboardHTML()
	w.Write([]byte(tmpl))
}

func (ws *WebServer) handleWeatherAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if ws.weatherData == nil {
		http.Error(w, "No weather data available", http.StatusServiceUnavailable)
		return
	}

	response := WeatherResponse{
		Temperature:   ws.weatherData.AirTemperature,
		Humidity:      ws.weatherData.RelativeHumidity,
		WindSpeed:     ws.weatherData.WindAvg,
		WindGust:      ws.weatherData.WindGust,
		WindDirection: ws.weatherData.WindDirection,
		RainAccum:     ws.weatherData.RainAccumulated,
		Pressure:      ws.weatherData.StationPressure,
		Illuminance:   ws.weatherData.Illuminance,
		UV:            ws.weatherData.UV,
		Battery:       ws.weatherData.Battery,
		LastUpdate:    time.Unix(ws.weatherData.Timestamp, 0).Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

func (ws *WebServer) handleStatusAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ws.mu.RLock()
	defer ws.mu.RUnlock()

	connected := ws.weatherData != nil
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
			Temperature:   obs.AirTemperature,
			Humidity:      obs.RelativeHumidity,
			WindSpeed:     obs.WindAvg,
			WindGust:      obs.WindGust,
			WindDirection: obs.WindDirection,
			RainAccum:     obs.RainAccumulated,
			Pressure:      obs.StationPressure,
			Illuminance:   obs.Illuminance,
			UV:            obs.UV,
			Battery:       obs.Battery,
			LastUpdate:    time.Unix(obs.Timestamp, 0).Format(time.RFC3339),
		}
	}

	response := StatusResponse{
		Connected:   connected,
		LastUpdate:  lastUpdate,
		Uptime:      uptimeStr,
		HomeKit:     ws.homekitStatus,
		DataHistory: history,
	}

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

        .accessory-icon {
            margin-right: 8px;
            font-size: 1rem;
        }

        .accessory-name {
            color: #555;
            font-weight: 500;
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
            bottom: 125%;
            left: 50%;
            margin-left: -150px;
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
            top: 25px;
            right: -150px;
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
                    <span class="card-title">Rain</span>
                </div>
                <div class="card-value" id="rain">--</div>
                <div class="card-unit" id="rain-unit" onclick="toggleUnit('rain')">in</div>
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
                <div class="card-unit" id="pressure-unit" onclick="toggleUnit('pressure')">mb</div>
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
        </div>

        <!-- Information Cards -->
        <div class="grid">
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
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0"></script>
    <script src="https://cdn.jsdelivr.net/npm/date-fns@2.30.0/index.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns@2.0.1/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
    <script>
        let units = {
            temperature: localStorage.getItem('temperature-unit') || 'celsius',
            wind: localStorage.getItem('wind-unit') || 'mph',
            rain: localStorage.getItem('rain-unit') || 'inches',
            pressure: localStorage.getItem('pressure-unit') || 'mb'
        };

        let weatherData = null;

        const charts = {};
        const maxDataPoints = 1000; // As specified in requirements

        function initCharts() {
            const ctxTemp = document.getElementById('temperature-chart').getContext('2d');
            const ctxHumidity = document.getElementById('humidity-chart').getContext('2d');
            const ctxWind = document.getElementById('wind-chart').getContext('2d');
            const ctxRain = document.getElementById('rain-chart').getContext('2d');
            const ctxPressure = document.getElementById('pressure-chart').getContext('2d');
            const ctxLight = document.getElementById('light-chart').getContext('2d');

            const chartConfig = {
                type: 'line',
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: { display: false }
                    },
                    scales: {
                        x: {
                            display: true,
                            type: 'time',
                            time: {
                                displayFormats: {
                                    minute: 'HH:mm',
                                    hour: 'HH:mm',
                                    day: 'MMM DD'
                                },
                                tooltipFormat: 'MMM DD, HH:mm:ss'
                            },
                            grid: {
                                display: true,
                                color: 'rgba(0,0,0,0.1)'
                            },
                            ticks: {
                                maxTicksLimit: 6,
                                color: '#666',
                                font: {
                                    size: 10
                                }
                            },
                            title: {
                                display: true,
                                text: 'Time',
                                color: '#666',
                                font: {
                                    size: 12
                                }
                            }
                        },
                        y: {
                            display: true,
                            grid: {
                                display: true,
                                color: 'rgba(0,0,0,0.1)'
                            },
                            ticks: {
                                maxTicksLimit: 5,
                                color: '#666',
                                font: {
                                    size: 10
                                },
                                callback: function(value) {
                                    return value.toFixed(1);
                                }
                            },
                            title: {
                                display: true,
                                text: 'Value',
                                color: '#666',
                                font: {
                                    size: 12
                                }
                            }
                        }
                    },
                    elements: {
                        point: { radius: 0 },
                        line: { borderWidth: 2 }
                    },
                    interaction: {
                        intersect: false,
                        mode: 'index'
                    }
                }
            };

            charts.temperature = new Chart(ctxTemp, {
                ...chartConfig,
                data: {
                    datasets: [{
                        data: [],
                        borderColor: '#ff6384',
                        backgroundColor: 'rgba(255, 99, 132, 0.1)',
                        fill: false,
                        tension: 0.4,
                        label: 'Temperature'
                    }, {
                        data: [],
                        borderColor: '#ff6384',
                        backgroundColor: 'rgba(255, 99, 132, 0.2)',
                        borderDash: [5, 5],
                        borderWidth: 2,
                        fill: false,
                        pointRadius: 0,
                        tension: 0,
                        label: 'Average'
                    }]
                }
            });

            charts.humidity = new Chart(ctxHumidity, {
                ...chartConfig,
                data: {
                    datasets: [{
                        data: [],
                        borderColor: '#36a2eb',
                        backgroundColor: 'rgba(54, 162, 235, 0.1)',
                        fill: false,
                        tension: 0.4,
                        label: 'Humidity'
                    }, {
                        data: [],
                        borderColor: '#36a2eb',
                        backgroundColor: 'rgba(54, 162, 235, 0.2)',
                        borderDash: [5, 5],
                        borderWidth: 2,
                        fill: false,
                        pointRadius: 0,
                        tension: 0,
                        label: 'Average'
                    }]
                }
            });

            charts.wind = new Chart(ctxWind, {
                ...chartConfig,
                data: {
                    datasets: [{
                        data: [],
                        borderColor: '#4bc0c0',
                        backgroundColor: 'rgba(75, 192, 192, 0.1)',
                        fill: false,
                        tension: 0.4,
                        label: 'Wind'
                    }, {
                        data: [],
                        borderColor: '#4bc0c0',
                        backgroundColor: 'rgba(75, 192, 192, 0.2)',
                        borderDash: [5, 5],
                        borderWidth: 2,
                        fill: false,
                        pointRadius: 0,
                        tension: 0,
                        label: 'Average'
                    }]
                }
            });

            charts.rain = new Chart(ctxRain, {
                ...chartConfig,
                data: {
                    datasets: [{
                        data: [],
                        borderColor: '#9966ff',
                        backgroundColor: 'rgba(153, 102, 255, 0.1)',
                        fill: false,
                        tension: 0.4,
                        label: 'Rain'
                    }, {
                        data: [],
                        borderColor: '#9966ff',
                        backgroundColor: 'rgba(153, 102, 255, 0.2)',
                        borderDash: [5, 5],
                        borderWidth: 2,
                        fill: false,
                        pointRadius: 0,
                        tension: 0,
                        label: 'Average'
                    }]
                }
            });

            charts.pressure = new Chart(ctxPressure, {
                ...chartConfig,
                data: {
                    datasets: [{
                        data: [],
                        borderColor: '#ff9f40',
                        backgroundColor: 'rgba(255, 159, 64, 0.1)',
                        fill: false,
                        tension: 0.4,
                        label: 'Pressure'
                    }, {
                        data: [],
                        borderColor: '#ff9f40',
                        backgroundColor: 'rgba(255, 159, 64, 0.2)',
                        borderDash: [5, 5],
                        borderWidth: 2,
                        fill: false,
                        pointRadius: 0,
                        tension: 0,
                        label: 'Average'
                    }]
                }
            });

            charts.light = new Chart(ctxLight, {
                ...chartConfig,
                data: {
                    datasets: [{
                        data: [],
                        borderColor: '#ffcd56',
                        backgroundColor: 'rgba(255, 205, 86, 0.1)',
                        fill: false,
                        tension: 0.4,
                        label: 'Light'
                    }, {
                        data: [],
                        borderColor: '#ffcd56',
                        backgroundColor: 'rgba(255, 205, 86, 0.2)',
                        borderDash: [5, 5],
                        borderWidth: 2,
                        fill: false,
                        pointRadius: 0,
                        tension: 0,
                        label: 'Average'
                    }]
                }
            });
        }

        function updateUnits() {
            document.getElementById('temperature-unit').textContent = units.temperature === 'celsius' ? '°C' : '°F';
            document.getElementById('wind-unit').textContent = units.wind === 'mph' ? 'mph' : 'kph';
            document.getElementById('rain-unit').textContent = units.rain === 'inches' ? 'in' : 'mm';
            document.getElementById('pressure-unit').textContent = units.pressure === 'mb' ? 'mb' : 'inHg';
        }

        function toggleUnit(sensor) {
            if (sensor === 'temperature') {
                units.temperature = units.temperature === 'celsius' ? 'fahrenheit' : 'celsius';
                localStorage.setItem('temperature-unit', units.temperature);
            } else if (sensor === 'wind') {
                units.wind = units.wind === 'mph' ? 'kph' : 'mph';
                localStorage.setItem('wind-unit', units.wind);
            } else if (sensor === 'rain') {
                units.rain = units.rain === 'inches' ? 'mm' : 'inches';
                localStorage.setItem('rain-unit', units.rain);
            } else if (sensor === 'pressure') {
                units.pressure = units.pressure === 'mb' ? 'inHg' : 'mb';
                localStorage.setItem('pressure-unit', units.pressure);
            }
            updateUnits();
            updateDisplay();
            recalculateAverages();
        }

        function degreesToDirection(degrees) {
            const directions = ['N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE', 'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'];
            const index = Math.round(degrees / 22.5) % 16;
            return directions[index];
        }

        function updateArrow(direction) {
            const arrows = {
                'N': '↑', 'NNE': '↗', 'NE': '↗', 'ENE': '↗',
                'E': '→', 'ESE': '↘', 'SE': '↘', 'SSE': '↘',
                'S': '↓', 'SSW': '↙', 'SW': '↙', 'WSW': '↙',
                'W': '←', 'WNW': '↖', 'NW': '↖', 'NNW': '↖'
            };
            return arrows[direction] || '↑';
        }

        function celsiusToFahrenheit(celsius) {
            return (celsius * 9/5) + 32;
        }

        function fahrenheitToCelsius(fahrenheit) {
            return (fahrenheit - 32) * 5/9;
        }

        function mphToKph(mph) {
            return mph * 1.60934;
        }

        function kphToMph(kph) {
            return kph / 1.60934;
        }

        function inchesToMm(inches) {
            return inches * 25.4;
        }

        function mmToInches(mm) {
            return mm / 25.4;
        }

        function mbToInHg(mb) {
            return mb * 0.02953;
        }

        function inHgToMb(inHg) {
            return inHg / 0.02953;
        }

        function calculateHeatIndex(tempC, humidity) {
            // Convert temperature to Fahrenheit for calculation
            const tempF = (tempC * 9/5) + 32;
            
            // If conditions don't warrant heat index calculation, return the temperature
            if (tempF < 80 || humidity < 40) {
                return tempC; // Return original temperature in Celsius
            }
            
            // Heat Index calculation using the NOAA formula
            // Constants for the formula
            const c1 = -42.379;
            const c2 = 2.04901523;
            const c3 = 10.14333127;
            const c4 = -0.22475541;
            const c5 = -0.00683783;
            const c6 = -0.05481717;
            const c7 = 0.00122874;
            const c8 = 0.00085282;
            const c9 = -0.00000199;
            
            // Calculate heat index in Fahrenheit
            const heatIndexF = c1 + (c2 * tempF) + (c3 * humidity) + (c4 * tempF * humidity) +
                             (c5 * tempF * tempF) + (c6 * humidity * humidity) +
                             (c7 * tempF * tempF * humidity) + (c8 * tempF * humidity * humidity) +
                             (c9 * tempF * tempF * humidity * humidity);
            
            // Convert back to Celsius
            return (heatIndexF - 32) * 5/9;
        }

        function calculateAverage(data) {
            if (data.length === 0) return 0;
            const sum = data.reduce((acc, point) => acc + point.y, 0);
            return sum / data.length;
        }

        function updateAverageLine(chart, data, averageValue) {
            if (data.length === 0) {
                chart.data.datasets[1].data = [];
                return;
            }

            // Create average line points spanning the entire time range
            const firstPoint = data[0];
            const lastPoint = data[data.length - 1];

            chart.data.datasets[1].data = [
                { x: firstPoint.x, y: averageValue },
                { x: lastPoint.x, y: averageValue }
            ];
        }

        function getLuxDescription(lux) {
            if (lux <= 0.0001) return "Moonless, overcast night sky (starlight)";
            if (lux <= 0.002) return "Moonless clear night sky with airglow";
            if (lux <= 0.01) return "Quarter moon on a clear night";
            if (lux <= 0.3) return "Full moon on a clear night";
            if (lux <= 3.4) return "Dark limit of civil twilight";
            if (lux <= 50) return "Public areas with dark surroundings";
            if (lux <= 80) return "Family living room lights";
            if (lux <= 100) return "Office building hallway/toilet lighting";
            if (lux <= 150) return "Very dark overcast day";
            if (lux <= 400) return "Train station platforms";
            if (lux <= 500) return "Office lighting";
            if (lux <= 1000) return "Sunrise or sunset on a clear day";
            if (lux <= 25000) return "Overcast day / Full daylight (not direct sun)";
            if (lux <= 100000) return "Direct sunlight";
            return "Extremely bright conditions";
        }

        function toggleLuxTooltip() {
            const tooltip = document.getElementById('lux-tooltip');
            tooltip.classList.toggle('show');
        }

        function closeLuxTooltip() {
            const tooltip = document.getElementById('lux-tooltip');
            tooltip.classList.remove('show');
        }

        function handleLuxTooltipClickOutside(event) {
            const tooltip = document.getElementById('lux-tooltip');
            const context = document.getElementById('lux-context');
            const infoIcon = document.getElementById('lux-info-icon');

            // If tooltip is visible and click is outside the tooltip and info icon
            if (tooltip.classList.contains('show') &&
                !tooltip.contains(event.target) &&
                !infoIcon.contains(event.target)) {
                closeLuxTooltip();
            }
        }

        function toggleHeatIndexTooltip() {
            const tooltip = document.getElementById('heat-index-tooltip');
            tooltip.classList.toggle('show');
        }

        function closeHeatIndexTooltip() {
            const tooltip = document.getElementById('heat-index-tooltip');
            tooltip.classList.remove('show');
        }

        function handleHeatIndexTooltipClickOutside(event) {
            const tooltip = document.getElementById('heat-index-tooltip');
            const context = document.getElementById('heat-index-context');
            const infoIcon = document.getElementById('heat-index-info-icon');

            // If tooltip is visible and click is outside the tooltip and info icon
            if (tooltip.classList.contains('show') &&
                !tooltip.contains(event.target) &&
                !infoIcon.contains(event.target)) {
                closeHeatIndexTooltip();
            }
        }

        function updateDisplay() {
            if (!weatherData) return;

            let temp = weatherData.temperature;
            if (units.temperature === 'fahrenheit') {
                temp = celsiusToFahrenheit(temp);
            }
            document.getElementById('temperature').textContent = temp.toFixed(1);

            document.getElementById('humidity').textContent = weatherData.humidity.toFixed(1);
            
            // Calculate and display heat index
            const heatIndexC = calculateHeatIndex(weatherData.temperature, weatherData.humidity);
            let heatIndexDisplay = heatIndexC;
            if (units.temperature === 'fahrenheit') {
                heatIndexDisplay = celsiusToFahrenheit(heatIndexC);
            }
            const tempUnit = units.temperature === 'celsius' ? '°C' : '°F';
            document.getElementById('heat-index').textContent = heatIndexDisplay.toFixed(1) + tempUnit;

            let windSpeed = weatherData.windSpeed;
            if (units.wind === 'kph') {
                windSpeed = mphToKph(windSpeed);
            }
            document.getElementById('wind-speed').textContent = windSpeed.toFixed(1);

            const direction = degreesToDirection(weatherData.windDirection);
            document.getElementById('wind-direction').textContent = direction + ' (' + weatherData.windDirection.toFixed(0) + '°)';
            document.getElementById('wind-arrow').textContent = updateArrow(direction);

            let rain = weatherData.rainAccum;
            if (units.rain === 'mm') {
                rain = inchesToMm(rain);
            }
            document.getElementById('rain').textContent = rain.toFixed(3);

            let pressure = weatherData.pressure;
            if (units.pressure === 'inHg') {
                pressure = mbToInHg(pressure);
            }
            document.getElementById('pressure').textContent = pressure.toFixed(1);

            document.getElementById('illuminance').textContent = weatherData.illuminance.toFixed(0);
            document.getElementById('lux-description').textContent = getLuxDescription(weatherData.illuminance);

            document.getElementById('last-update').textContent = new Date(weatherData.lastUpdate).toLocaleString();
        }

        function updateCharts() {
            if (!weatherData) return;

            // Add current data to charts
            const now = new Date(weatherData.lastUpdate);

            // Temperature chart
            let tempValue = weatherData.temperature;
            if (units.temperature === 'fahrenheit') {
                tempValue = celsiusToFahrenheit(tempValue);
            }
            charts.temperature.data.datasets[0].data.push({ x: now, y: tempValue });
            if (charts.temperature.data.datasets[0].data.length > maxDataPoints) {
                charts.temperature.data.datasets[0].data.shift();
            }
            const tempAvg = calculateAverage(charts.temperature.data.datasets[0].data);
            updateAverageLine(charts.temperature, charts.temperature.data.datasets[0].data, tempAvg);
            charts.temperature.options.scales.y.title = {
                display: true,
                text: units.temperature === 'celsius' ? '°C' : '°F'
            };
            charts.temperature.update();

            // Humidity chart
            charts.humidity.data.datasets[0].data.push({ x: now, y: weatherData.humidity });
            if (charts.humidity.data.datasets[0].data.length > maxDataPoints) {
                charts.humidity.data.datasets[0].data.shift();
            }
            const humidityAvg = calculateAverage(charts.humidity.data.datasets[0].data);
            updateAverageLine(charts.humidity, charts.humidity.data.datasets[0].data, humidityAvg);
            charts.humidity.options.scales.y.title = {
                display: true,
                text: '%'
            };
            charts.humidity.update();

            // Wind chart
            let windValue = weatherData.windSpeed;
            if (units.wind === 'kph') {
                windValue = mphToKph(windValue);
            }
            charts.wind.data.datasets[0].data.push({ x: now, y: windValue });
            if (charts.wind.data.datasets[0].data.length > maxDataPoints) {
                charts.wind.data.datasets[0].data.shift();
            }
            const windAvg = calculateAverage(charts.wind.data.datasets[0].data);
            updateAverageLine(charts.wind, charts.wind.data.datasets[0].data, windAvg);
            charts.wind.options.scales.y.title = {
                display: true,
                text: units.wind === 'mph' ? 'mph' : 'kph'
            };
            charts.wind.update();

            // Rain chart
            let rainValue = weatherData.rainAccum;
            if (units.rain === 'mm') {
                rainValue = inchesToMm(rainValue);
            }
            charts.rain.data.datasets[0].data.push({ x: now, y: rainValue });
            if (charts.rain.data.datasets[0].data.length > maxDataPoints) {
                charts.rain.data.datasets[0].data.shift();
            }
            const rainAvg = calculateAverage(charts.rain.data.datasets[0].data);
            updateAverageLine(charts.rain, charts.rain.data.datasets[0].data, rainAvg);
            charts.rain.options.scales.y.title = {
                display: true,
                text: units.rain === 'inches' ? 'in' : 'mm'
            };
            charts.rain.update();

            // Pressure chart
            let pressureValue = weatherData.pressure;
            if (units.pressure === 'inHg') {
                pressureValue = mbToInHg(pressureValue);
            }
            charts.pressure.data.datasets[0].data.push({ x: now, y: pressureValue });
            if (charts.pressure.data.datasets[0].data.length > maxDataPoints) {
                charts.pressure.data.datasets[0].data.shift();
            }
            const pressureAvg = calculateAverage(charts.pressure.data.datasets[0].data);
            updateAverageLine(charts.pressure, charts.pressure.data.datasets[0].data, pressureAvg);
            charts.pressure.options.scales.y.title = {
                display: true,
                text: units.pressure === 'mb' ? 'mb' : 'inHg'
            };
            charts.pressure.update();

            // Light chart
            charts.light.data.datasets[0].data.push({ x: now, y: weatherData.illuminance });
            if (charts.light.data.datasets[0].data.length > maxDataPoints) {
                charts.light.data.datasets[0].data.shift();
            }
            const lightAvg = calculateAverage(charts.light.data.datasets[0].data);
            updateAverageLine(charts.light, charts.light.data.datasets[0].data, lightAvg);
            charts.light.options.scales.y.title = {
                display: true,
                text: 'lux'
            };
            charts.light.update();
        }

        function recalculateAverages() {
            // Recalculate temperature data and average
            if (charts.temperature.data.datasets[0].data.length > 0) {
                charts.temperature.data.datasets[0].data.forEach(point => {
                    if (units.temperature === 'fahrenheit') {
                        point.y = celsiusToFahrenheit(point.y);
                    } else {
                        point.y = fahrenheitToCelsius(point.y);
                    }
                });
                const tempAvg = calculateAverage(charts.temperature.data.datasets[0].data);
                updateAverageLine(charts.temperature, charts.temperature.data.datasets[0].data, tempAvg);
                charts.temperature.update();
            }

            // Recalculate wind data and average
            if (charts.wind.data.datasets[0].data.length > 0) {
                charts.wind.data.datasets[0].data.forEach(point => {
                    if (units.wind === 'kph') {
                        point.y = mphToKph(point.y);
                    } else {
                        point.y = kphToMph(point.y);
                    }
                });
                const windAvg = calculateAverage(charts.wind.data.datasets[0].data);
                updateAverageLine(charts.wind, charts.wind.data.datasets[0].data, windAvg);
                charts.wind.update();
            }

            // Recalculate rain data and average
            if (charts.rain.data.datasets[0].data.length > 0) {
                charts.rain.data.datasets[0].data.forEach(point => {
                    if (units.rain === 'mm') {
                        point.y = inchesToMm(point.y);
                    } else {
                        point.y = mmToInches(point.y);
                    }
                });
                const rainAvg = calculateAverage(charts.rain.data.datasets[0].data);
                updateAverageLine(charts.rain, charts.rain.data.datasets[0].data, rainAvg);
                charts.rain.update();
            }

            // Recalculate pressure data and average
            if (charts.pressure.data.datasets[0].data.length > 0) {
                charts.pressure.data.datasets[0].data.forEach(point => {
                    if (units.pressure === 'inHg') {
                        point.y = mbToInHg(point.y);
                    } else {
                        point.y = inHgToMb(point.y);
                    }
                });
                const pressureAvg = calculateAverage(charts.pressure.data.datasets[0].data);
                updateAverageLine(charts.pressure, charts.pressure.data.datasets[0].data, pressureAvg);
                charts.pressure.update();
            }
        }

        async function fetchWeather() {
            try {
                const response = await fetch('/api/weather');
                if (response.ok) {
                    weatherData = await response.json();
                    updateDisplay();
                    updateCharts();
                    document.getElementById('status').textContent = 'Connected to Tempest station';
                    document.getElementById('status').style.background = 'rgba(40, 167, 69, 0.1)';
                } else {
                    throw new Error('Weather API error');
                }
            } catch (error) {
                console.error('Error fetching weather:', error);
                document.getElementById('status').textContent = 'Disconnected from weather station';
                document.getElementById('status').style.background = 'rgba(220, 53, 69, 0.1)';
            }
        }

        async function fetchStatus() {
            try {
                const response = await fetch('/api/status');
                if (response.ok) {
                    const status = await response.json();
                    updateStatusDisplay(status);
                }
            } catch (error) {
                console.error('Error fetching status:', error);
            }
        }

        function updateStatusDisplay(status) {
            // Update Tempest status
            const tempestStatus = document.getElementById('tempest-status');
            const tempestStation = document.getElementById('tempest-station');
            const tempestLastUpdate = document.getElementById('tempest-last-update');
            const tempestUptime = document.getElementById('tempest-uptime');

            tempestStatus.textContent = status.connected ? 'Connected' : 'Disconnected';
            tempestStatus.style.color = status.connected ? '#28a745' : '#dc3545';
            tempestStation.textContent = status.stationName || '--';
            tempestLastUpdate.textContent = status.lastUpdate ? new Date(status.lastUpdate).toLocaleString() : '--';
            tempestUptime.textContent = status.uptime || '--';

            // Update HomeKit status
            const homekitStatus = document.getElementById('homekit-status');
            const homekitAccessories = document.getElementById('homekit-accessories');
            const homekitBridge = document.getElementById('homekit-bridge');
            const homekitPin = document.getElementById('homekit-pin');

            const hk = status.homekit || {};
            homekitStatus.textContent = hk.bridge ? 'Active' : 'Inactive';
            homekitStatus.style.color = hk.bridge ? '#28a745' : '#dc3545';
            homekitAccessories.textContent = hk.accessories || '--';
            homekitBridge.textContent = hk.name || '--';
            homekitPin.textContent = hk.pin || '--';

            // Update accessories list
            updateAccessoriesList(hk.accessoryNames || []);
        }

        function updateAccessoriesList(accessoryNames) {
            const accessoriesList = document.getElementById('accessories-list');
            accessoriesList.innerHTML = '';

            if (!accessoryNames || accessoryNames.length === 0) {
                accessoriesList.innerHTML = '<div class="accessory-item">No accessories available</div>';
                return;
            }

            // Define icons for different accessory types
            const accessoryIcons = {
                'Temperature': '🌡️',
                'Humidity': '💧',
                'Wind Speed': '🌬️',
                'Wind Direction': '🧭',
                'Rain': '🌧️',
                'Light': '☀️'
            };

            accessoryNames.forEach(name => {
                const accessoryDiv = document.createElement('div');
                accessoryDiv.className = 'accessory-item';

                // Get appropriate icon or use default
                let icon = '🔧'; // default icon
                for (const [key, value] of Object.entries(accessoryIcons)) {
                    if (name.includes(key)) {
                        icon = value;
                        break;
                    }
                }

                accessoryDiv.innerHTML = '<span class="accessory-icon">' + icon + '</span><span class="accessory-name">' + name + '</span>';

                accessoriesList.appendChild(accessoryDiv);
            });
        }

        function toggleAccessoriesExpansion() {
            const expandedDiv = document.getElementById('accessories-expanded');
            const expandIcon = document.getElementById('accessories-expand-icon');

            if (expandedDiv.style.display === 'none' || expandedDiv.style.display === '') {
                expandedDiv.style.display = 'block';
                expandIcon.textContent = '▼';
                expandIcon.style.transform = 'rotate(0deg)';
            } else {
                expandedDiv.style.display = 'none';
                expandIcon.textContent = '▶';
                expandIcon.style.transform = 'rotate(0deg)';
            }
        }

        // Initialize
        updateUnits();
        initCharts();

        // Add click event listener for accessories expansion
        document.getElementById('accessories-row').addEventListener('click', toggleAccessoriesExpansion);

        // Add click event listener for lux info icon
        document.getElementById('lux-info-icon').addEventListener('click', toggleLuxTooltip);

        // Add click event listener for lux tooltip close button
        document.getElementById('lux-tooltip-close').addEventListener('click', closeLuxTooltip);

        // Add click event listener for closing tooltip when clicking outside
        document.addEventListener('click', handleLuxTooltipClickOutside);

        // Add click event listener for heat index info icon
        document.getElementById('heat-index-info-icon').addEventListener('click', toggleHeatIndexTooltip);

        // Add click event listener for heat index tooltip close button
        document.getElementById('heat-index-tooltip-close').addEventListener('click', closeHeatIndexTooltip);

        // Add click event listener for closing heat index tooltip when clicking outside
        document.addEventListener('click', handleHeatIndexTooltipClickOutside);

        // Fetch data immediately and then every 10 seconds
        fetchWeather();
        fetchStatus();
        setInterval(() => {
            fetchWeather();
            fetchStatus();
        }, 10000);
    </script>
</body>
</html>`
}

func (ws *WebServer) Stop() error {
	if ws.server != nil {
		return ws.server.Close()
	}
	return nil
}
