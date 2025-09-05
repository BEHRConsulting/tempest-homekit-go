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
	port          string
	server        *http.Server
	weatherMux    sync.RWMutex
	lastObs       *weather.Observation
	lastUpdate    time.Time
	homekitMux    sync.RWMutex
	homekitStatus map[string]interface{}
}

type DashboardData struct {
	Temperature float64
	Humidity    float64
	WindSpeed   float64
	RainAccum   float64
	LastUpdate  time.Time
	IsConnected bool
	StationName string
}

func NewWebServer(port string) *WebServer {
	ws := &WebServer{
		port: port,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.handleDashboard)
	mux.HandleFunc("/api/weather", ws.handleWeatherAPI)
	mux.HandleFunc("/api/status", ws.handleStatusAPI)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("pkg/web/static/"))))

	ws.server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return ws
}

func (ws *WebServer) Start() error {
	log.Printf("Starting web dashboard on port %s", ws.port)
	return ws.server.ListenAndServe()
}

func (ws *WebServer) Stop() error {
	return ws.server.Close()
}

func (ws *WebServer) UpdateHomeKitStatus(status map[string]interface{}) {
	ws.homekitMux.Lock()
	defer ws.homekitMux.Unlock()
	ws.homekitStatus = status
}

func (ws *WebServer) UpdateWeather(obs *weather.Observation) {
	ws.weatherMux.Lock()
	defer ws.weatherMux.Unlock()
	ws.lastObs = obs
	ws.lastUpdate = time.Now()
}

func (ws *WebServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
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
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
        }

        .header {
            text-align: center;
            margin-bottom: 40px;
            color: white;
        }

        .header h1 {
            font-size: 3rem;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }

        .header p {
            font-size: 1.2rem;
            opacity: 0.9;
        }

        .dashboard {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .card {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 20px;
            padding: 30px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.2);
            transition: transform 0.3s ease;
            cursor: pointer;
        }

        .card:hover {
            transform: translateY(-5px);
        }

        .card-header {
            display: flex;
            align-items: center;
            margin-bottom: 20px;
        }

        .card-icon {
            width: 50px;
            height: 50px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 15px;
            font-size: 24px;
        }

        .temperature .card-icon { background: linear-gradient(45deg, #ff6b6b, #ee5a24); }
        .humidity .card-icon { background: linear-gradient(45deg, #4ecdc4, #44a08d); }
        .wind .card-icon { background: linear-gradient(45deg, #74b9ff, #0984e3); }
        .rain .card-icon { background: linear-gradient(45deg, #a29bfe, #6c5ce7); }

        .card-title {
            font-size: 1.2rem;
            font-weight: 600;
            color: #2d3436;
        }

        .card-value {
            font-size: 3rem;
            font-weight: 700;
            color: #2d3436;
            margin-bottom: 5px;
        }

        .card-unit {
            font-size: 1.2rem;
            color: #636e72;
            font-weight: 500;
            cursor: pointer;
            transition: color 0.2s ease;
        }

        .card-unit:hover {
            color: #0984e3;
        }

        .status-section {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .status-card {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 15px;
            padding: 20px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
        }

        .status-card h3 {
            font-size: 1.2rem;
            margin-bottom: 15px;
            color: #2d3436;
        }

        .status-item {
            display: flex;
            justify-content: space-between;
            margin-bottom: 8px;
            padding: 5px 0;
        }

        .status-label {
            font-weight: 500;
            color: #636e72;
        }

        .status-value {
            font-weight: 600;
            color: #2d3436;
        }

        .wind-direction {
            font-size: 1rem;
            color: #636e72;
            font-weight: 500;
            margin-top: 5px;
            text-align: center;
        }

        .status-disconnected {
            color: #d63031;
        }

        .last-update {
            font-size: 0.9rem;
            color: #636e72;
            margin-top: 10px;
        }

        @media (max-width: 768px) {
            .header h1 {
                font-size: 2rem;
            }

            .dashboard {
                grid-template-columns: 1fr;
            }

            .status-section {
                grid-template-columns: 1fr;
            }

            .card {
                padding: 20px;
            }

            .card-value {
                font-size: 2.5rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌤️ Tempest Weather</h1>
            <p>Real-time weather monitoring dashboard</p>
        </div>

        <div class="dashboard">
            <div class="card temperature" onclick="toggleTemperatureUnit()">
                <div class="card-header">
                    <div class="card-icon">🌡️</div>
                    <div class="card-title">Temperature</div>
                </div>
                <div class="card-value" id="temperature">--</div>
                <div class="card-unit" id="temp-unit">°C</div>
            </div>

            <div class="card humidity">
                <div class="card-header">
                    <div class="card-icon">💧</div>
                    <div class="card-title">Humidity</div>
                </div>
                <div class="card-value" id="humidity">--</div>
                <div class="card-unit">%</div>
            </div>

            <div class="card wind" onclick="toggleWindUnit()">
                <div class="card-header">
                    <div class="card-icon">🌬️</div>
                    <div class="card-title">Wind Speed</div>
                </div>
                <div class="card-value" id="wind">--</div>
                <div class="card-unit" id="wind-unit">mph</div>
                <div class="wind-direction" id="wind-direction">--</div>
            </div>

            <div class="card rain" onclick="toggleRainUnit()">
                <div class="card-header">
                    <div class="card-icon">🌧️</div>
                    <div class="card-title">Rain Today</div>
                </div>
                <div class="card-value" id="rain">--</div>
                <div class="card-unit" id="rain-unit">in</div>
            </div>
        </div>

        <div class="status-section">
            <div class="status-card">
                <h3>🌐 Connection Status</h3>
                <div class="status-item">
                    <span class="status-label">Tempest Station:</span>
                    <span id="connection-status" class="status-disconnected">⏳ Connecting...</span>
                </div>
                <div class="status-item">
                    <span class="status-label">Last Update:</span>
                    <span id="last-update">Never</span>
                </div>
            </div>

            <div class="status-card">
                <h3>🏠 HomeKit Status</h3>
                <div class="status-item">
                    <span class="status-label">Bridge:</span>
                    <span id="homekit-bridge" class="status-disconnected">Unknown</span>
                </div>
                <div class="status-item">
                    <span class="status-label">Accessories:</span>
                    <span id="homekit-accessories">0</span>
                </div>
                <div class="status-item">
                    <span class="status-label">PIN:</span>
                    <span id="homekit-pin">Not set</span>
                </div>
            </div>
        </div>
    </div>

    <script>
        let updateInterval;
        let units = {
            temperature: localStorage.getItem('tempUnit') || 'celsius',
            wind: localStorage.getItem('windUnit') || 'mph',
            rain: localStorage.getItem('rainUnit') || 'inches'
        };

        // Unit conversion functions
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
            return kph * 0.621371;
        }

        function inchesToMm(inches) {
            return inches * 25.4;
        }

        function mmToInches(mm) {
            return mm * 0.0393701;
        }

        function degreesToDirection(degrees) {
            const directions = ['N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE', 'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'];
            const index = Math.round(degrees / 22.5) % 16;
            return directions[index];
        }

        // Toggle functions
        function toggleTemperatureUnit() {
            units.temperature = units.temperature === 'celsius' ? 'fahrenheit' : 'celsius';
            localStorage.setItem('tempUnit', units.temperature);
            updateUnitsDisplay();
            fetchWeatherData();
        }

        function toggleWindUnit() {
            units.wind = units.wind === 'mph' ? 'kph' : 'mph';
            localStorage.setItem('windUnit', units.wind);
            updateUnitsDisplay();
            fetchWeatherData();
        }

        function toggleRainUnit() {
            units.rain = units.rain === 'inches' ? 'mm' : 'inches';
            localStorage.setItem('rainUnit', units.rain);
            updateUnitsDisplay();
            fetchWeatherData();
        }

        function updateUnitsDisplay() {
            document.getElementById('temp-unit').textContent = units.temperature === 'celsius' ? '°C' : '°F';
            document.getElementById('wind-unit').textContent = units.wind === 'mph' ? 'mph' : 'kph';
            document.getElementById('rain-unit').textContent = units.rain === 'inches' ? 'in' : 'mm';
        }

        async function fetchWeatherData() {
            try {
                const response = await fetch('/api/weather');
                const data = await response.json();

                let temp = data.temperature;
                let wind = data.windSpeed;
                let rain = data.rainAccum;

                // Apply unit conversions
                if (units.temperature === 'fahrenheit') {
                    temp = celsiusToFahrenheit(temp);
                }

                if (units.wind === 'kph') {
                    wind = mphToKph(wind);
                }

                if (units.rain === 'mm') {
                    rain = inchesToMm(rain);
                }

                document.getElementById('temperature').textContent = temp.toFixed(1);
                document.getElementById('humidity').textContent = data.humidity.toFixed(0);
                document.getElementById('wind').textContent = wind.toFixed(1);
                document.getElementById('wind-direction').textContent = degreesToDirection(data.windDirection) + ' (' + data.windDirection.toFixed(0) + '°)';
                document.getElementById('rain').textContent = rain.toFixed(units.rain === 'inches' ? 3 : 1);

                document.getElementById('connection-status').className = 'status-connected';
                document.getElementById('connection-status').textContent = '✅ Connected';

                const lastUpdate = new Date(data.lastUpdate);
                document.getElementById('last-update').textContent = lastUpdate.toLocaleTimeString();

            } catch (error) {
                console.error('Failed to fetch weather data:', error);
                document.getElementById('connection-status').className = 'status-disconnected';
                document.getElementById('connection-status').textContent = '❌ Connection lost';
            }
        }

        async function fetchStatusData() {
            try {
                const response = await fetch('/api/status');
                const data = await response.json();

                // Update HomeKit status
                if (data.homekit) {
                    document.getElementById('homekit-bridge').className = data.homekit.bridge ? 'status-connected' : 'status-disconnected';
                    document.getElementById('homekit-bridge').textContent = data.homekit.bridge ? '✅ Active' : '❌ Inactive';
                    document.getElementById('homekit-accessories').textContent = data.homekit.accessories || '0';
                    document.getElementById('homekit-pin').textContent = data.homekit.pin || 'Not set';
                }
            } catch (error) {
                console.error('Failed to fetch status data:', error);
            }
        }

        function startUpdates() {
            updateUnitsDisplay();
            fetchWeatherData();
            fetchStatusData();

            updateInterval = setInterval(() => {
                fetchWeatherData();
                fetchStatusData();
            }, 10000); // Update every 10 seconds
        }

        function stopUpdates() {
            if (updateInterval) {
                clearInterval(updateInterval);
            }
        }

        // Start updates when page loads
        document.addEventListener('DOMContentLoaded', startUpdates);

        // Stop updates when page unloads
        window.addEventListener('beforeunload', stopUpdates);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

func (ws *WebServer) handleWeatherAPI(w http.ResponseWriter, r *http.Request) {
	ws.weatherMux.RLock()
	defer ws.weatherMux.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if ws.lastObs == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "No weather data available"})
		return
	}

	data := map[string]interface{}{
		"temperature":   ws.lastObs.AirTemperature,
		"humidity":      ws.lastObs.RelativeHumidity,
		"windSpeed":     ws.lastObs.WindAvg,
		"windDirection": ws.lastObs.WindDirection,
		"rainAccum":     ws.lastObs.RainAccumulated,
		"lastUpdate":    ws.lastUpdate,
	}

	json.NewEncoder(w).Encode(data)
}

func (ws *WebServer) handleStatusAPI(w http.ResponseWriter, r *http.Request) {
	ws.weatherMux.RLock()
	ws.homekitMux.RLock()
	defer ws.weatherMux.RUnlock()
	defer ws.homekitMux.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	status := map[string]interface{}{
		"connected":  ws.lastObs != nil,
		"lastUpdate": ws.lastUpdate,
		"uptime":     time.Since(time.Now().Add(-time.Hour)).String(), // Placeholder
		"homekit":    ws.homekitStatus,
	}

	json.NewEncoder(w).Encode(status)
}
