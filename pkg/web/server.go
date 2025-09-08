package web

import (
	"encoding/json"
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
	HomeKit     map[string]interface{} `json:"homekit"`
	DataHistory []WeatherResponse      `json:"dataHistory"`
}

func NewWebServer(port string) *WebServer {
	ws := &WebServer{
		port:           port,
		maxHistorySize: 1000,
		dataHistory:    make([]weather.Observation, 0, 1000),
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
		Uptime:      "N/A", // TODO: implement uptime tracking
		HomeKit:     ws.homekitStatus,
		DataHistory: history,
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

        .footer {
            text-align: center;
            color: white;
            margin-top: 30px;
            font-size: 0.9rem;
        }

        @media (max-width: 768px) {
            .container {
                padding: 10px;
            }

            .header h1 {
                font-size: 2rem;
            }

            .grid {
                grid-template-columns: 1fr;
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
                <div class="card-unit">lux</div>
                <div class="chart-container">
                    <canvas id="light-chart"></canvas>
                </div>
            </div>
        </div>

        <div class="footer">
            <p>Last updated: <span id="last-update">--</span></p>
            <p>Tempest HomeKit Service</p>
        </div>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/date-fns@2.29.3/index.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns@2.0.1/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
    <script>
        let units = {
            temperature: localStorage.getItem('temperature-unit') || 'celsius',
            wind: localStorage.getItem('wind-unit') || 'mph',
            rain: localStorage.getItem('rain-unit') || 'inches',
            pressure: localStorage.getItem('pressure-unit') || 'mb'
        };

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

        let weatherData = null;

        function updateDisplay() {
            if (!weatherData) return;

            let temp = weatherData.temperature;
            if (units.temperature === 'fahrenheit') {
                temp = celsiusToFahrenheit(temp);
            }
            document.getElementById('temperature').textContent = temp.toFixed(1);

            document.getElementById('humidity').textContent = weatherData.humidity.toFixed(1);

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
                    // Update HomeKit status if needed
                }
            } catch (error) {
                console.error('Error fetching status:', error);
            }
        }

        // Initialize
        updateUnits();
        initCharts();

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
