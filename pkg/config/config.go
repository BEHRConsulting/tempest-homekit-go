// Package config provides configuration management for the Tempest HomeKit service.
// It handles command-line flags, environment variables, and HomeKit database operations.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration parameters for the Tempest HomeKit service.
type Config struct {
	Token               string
	StationName         string
	Pin                 string
	LogLevel            string
	LogFilter           string // Filter log messages to only show those containing this string
	WebPort             string
	ClearDB             bool
	DisableHomeKit      bool // Disable HomeKit services and run web console only
	DisableWebConsole   bool // Disable web server (HomeKit only mode)
	DisableAlarms       bool // Disable alarm initialization and processing
	Sensors             string
	ReadHistory         bool
	TestAPI             bool
	TestAPILocal        bool    // Test local web API endpoints and exit
	TestEmail           string  // Send test email to this address and exit
	TestSMS             string  // Send test SMS to this phone number and exit
	TestWebhook         string  // Send test webhook to this URL and exit
	TestConsole         bool    // Send test console notification and exit
	TestSyslog          bool    // Send test syslog notification and exit
	TestOSLog           bool    // Send test oslog notification and exit
	TestEventLog        bool    // Send test eventlog notification and exit
	TestUDP             int     // Listen for UDP broadcasts for N seconds and display received data (default: 120)
	TestHomeKit         bool    // Test HomeKit bridge setup and pairing without starting service
	TestWebStatus       bool    // Test web status scraping and exit
	TestAlarm           string  // Trigger a specific alarm by name for testing
	UseWebStatus        bool    // Enable headless browser scraping of TempestWX status
	UseGeneratedWeather bool    // Use generated weather data for testing instead of Tempest API
	UDPStream           bool    // Listen for UDP broadcasts from local Tempest station
	DisableInternet     bool    // Disable all internet access (no API, no status scraping)
	StationURL          string  // Custom station URL for weather data (overrides Tempest API)
	Elevation           float64 // elevation in meters
	Units               string  // Units system: imperial, metric, or sae
	UnitsPressure       string  // Pressure units: inHg or mb
	HistoryPoints       int     // Number of data points to store in history (default: 1000, min: 10)
	ChartHistoryHours   int     // Number of hours of history to display in charts (default: 24, 0 = all)
	Version             bool    // Show version and exit
	// GeneratedWeatherPath is the URL path portion used for the built-in generated
	// weather endpoint. Default: "/api/generate-weather". This can be overridden
	// via the GENERATE_WEATHER_PATH environment variable or the --generate-path flag.
	GeneratedWeatherPath string

	// Alarm configuration
	Alarms         string // Alarm configuration: @filename.json or inline JSON
	AlarmsEdit     string // Alarm editor mode: @filename.json to edit
	AlarmsEditPort string // Port for alarm editor (default: 8081)

	// Webhook listener
	WebhookListener    bool   // Enable webhook listener server (default port: 8082)
	WebhookListenPort  string // Port for webhook listener server (default: 8082)
	WebhookListenerSet bool   // Track if webhook-listener flag was explicitly set
	WebhookPortSet     bool   // Track if webhook-listener-port flag was explicitly set

	// Environment file
	EnvFile string // Custom environment file (default: .env)

	// Status console options
	Status          bool   // Enable curses-based status console
	StatusRefresh   int    // Status refresh interval in seconds (default: 5)
	StatusTimeout   int    // Status timeout in seconds (0 = never, default: 0)
	StatusTheme     string // Color theme name (default: "dark-ocean")
	StatusThemeList bool   // List available themes and exit
}

// customUsage prints a well-formatted help message with grouped flags and examples
func customUsage() {
	fmt.Fprintf(os.Stderr, `Tempest HomeKit Bridge - HomeKit integration for WeatherFlow Tempest weather stations

USAGE:
  tempest-homekit-go [OPTIONS]

DATA SOURCE OPTIONS:
  --token <string>              WeatherFlow API token (required for API mode)
                                Env: TEMPEST_TOKEN
  --station <string>            Tempest station name (required for API mode)
                                Env: TEMPEST_STATION_NAME
  --udp-stream                  Listen for UDP broadcasts from local station (port 50222)
                                Enables offline operation during internet outages
                                Env: UDP_STREAM=true
  --use-generated-weather       Use simulated weather data for testing
                                Automatically sets --station-url to local generator
  --station-url <url>           Custom station URL (e.g., http://localhost:8080/api/generate-weather)
                                Overrides Tempest API, enables custom data sources
                                Env: STATION_URL
	--read-history                Preload historical observations from Tempest API up to HISTORY_POINTS
  --disable-internet            Disable all internet access (offline mode)
                                Requires: --udp-stream or --use-generated-weather
                                Incompatible with: --use-web-status, --read-history
                                Env: DISABLE_INTERNET=true

HOMEKIT OPTIONS:
  --pin <string>                HomeKit PIN for device pairing (default: "00102003")
                                Env: HOMEKIT_PIN
  --sensors <list>              Sensors to enable (default: "temp,lux,humidity,uv")
                                Options: all, min, or custom list
                                Available: temp/temperature, humidity, lux/light, wind,
                                          rain, pressure, uv/uvi, lightning
                                Env: SENSORS
  --disable-homekit             Run web console only (no HomeKit services)
  --cleardb                     Clear HomeKit database and reset device pairing

WEB CONSOLE OPTIONS:
  --web-port <port>             Web dashboard port (default: "8080")
                                Env: WEB_PORT
  --disable-webconsole          Disable web server (HomeKit only mode)
  --use-web-status              Enable Chrome-based scraping of TempestWX status page
                                Updates every 15 minutes, incompatible with --disable-internet

CONFIGURATION OPTIONS:
  --env <filename>              Custom environment file to load (default: ".env")
                                Overrides default .env file location
                                Env: ENV_FILE
  --elevation <value>           Station elevation (e.g., 903ft, 275m)
                                Auto-detected from station coordinates if not provided
  --units <system>              Units system: imperial (default), metric, or sae
                                Env: UNITS
  --units-pressure <unit>       Pressure units: inHg (default) or mb
                                Env: UNITS_PRESSURE
  --history <points>            Number of data points to store in history (default: 1000, min: 10)
                                Env: HISTORY_POINTS
  --chart-history <hours>       Hours of data to display in charts (default: 24, 0=all)
                                Env: CHART_HISTORY_HOURS

STATUS CONSOLE OPTIONS:
  --status                      Enable curses-based status console (TUI mode)
                                Displays real-time console logs, station status,
                                alarm status, and HomeKit sensors
                                Env: STATUS=true
  --status-refresh <seconds>    Status refresh interval in seconds (default: 5)
                                How often to refresh station/alarm/HomeKit data
                                Env: STATUS_REFRESH
  --status-timeout <seconds>    Auto-exit after N seconds (0 = never, default: 0)
                                Useful for automated testing or temporary monitoring
                                Env: STATUS_TIMEOUT
  --status-theme <name>         Color theme for status console (default: "dark-ocean")
                                Use --status-theme-list to see available themes
                                Env: STATUS_THEME
  --status-theme-list           List all available color themes and exit
                                Shows theme names and descriptions for both
                                light and dark background terminals

ALARM OPTIONS:
  --alarms <config>             Enable alarm system with configuration
                                Format: @filename.json or inline JSON string
                                Env: ALARMS
  --disable-alarms              Disable alarm initialization and processing
                                Useful for testing or reducing resource usage
  --alarms-edit <file>          Run alarm editor for config file (standalone mode)
                                Format: @filename.json
                                Opens web UI at http://localhost:<port>
                                Env: ALARMS_EDIT
  --alarms-edit-port <port>     Port for alarm editor web UI (default: 8081)
                                Env: ALARMS_EDIT_PORT
  --webhook-listener            Start webhook listener server (standalone mode)
                                Uses default port 8082
                                Env: WEBHOOK_LISTENER=true
  --webhook-listener-port <port> Port for webhook listener server (default: 8082)
                                Env: WEBHOOK_LISTEN_PORT

LOGGING & DEBUG OPTIONS:
  --loglevel <level>            Log level: error (default), warn/warning, info, debug
                                Env: LOG_LEVEL
  --logfilter <string>          Filter log messages (case-insensitive substring match)
                                Env: LOG_FILTER

TESTING OPTIONS:
  --test-api                    Test WeatherFlow API endpoints and exit
  --test-api-local              Test local web server API endpoints and exit
                                Uses port 8084 by default (override with --web-port)
  --test-email <email>          Send test email to specified address and exit
  --test-sms <phone>            Send test SMS to specified phone number and exit
  --test-webhook <url>          Send test webhook to specified URL and exit
  --test-console                Send test console notification and exit
  --test-syslog                 Send test syslog notification and exit
  --test-oslog                  Send test oslog notification and exit (macOS only)
  --test-eventlog               Send test eventlog notification and exit (Windows only)
  --test-udp [seconds]          Listen for UDP broadcasts for N seconds (default: 120) and exit
  --test-homekit                Test HomeKit bridge setup and pairing info, then exit
  --test-web-status             Test web status scraping from TempestWX and exit
  --test-alarm <name>           Trigger a specific alarm by name for testing and exit

OTHER OPTIONS:
  --version                     Show version information and exit
  --help                        Show this help message

EXAMPLES:
  # Basic HomeKit bridge with API
  tempest-homekit-go --token "your-token" --station "My Station"

  # Offline mode with UDP stream (survives internet outages)
  tempest-homekit-go --token "your-token" --station "My Station" --udp-stream --disable-internet

  # Testing with simulated weather data
  tempest-homekit-go --use-generated-weather --station "Generated"

  # Hybrid mode: UDP for real-time + API for forecast
  tempest-homekit-go --token "your-token" --station "My Station" --udp-stream --read-history

  # Web console only (no HomeKit)
  tempest-homekit-go --token "your-token" --station "My Station" --disable-homekit

  # HomeKit only (no web console)
  tempest-homekit-go --token "your-token" --station "My Station" --disable-webconsole

  # Custom sensors with debug logging
  tempest-homekit-go --token "your-token" --station "My Station" --sensors "temp,humidity,pressure" --loglevel debug

  # Filter logs for UDP messages only
  tempest-homekit-go --token "your-token" --station "My Station" --udp-stream --loglevel debug --logfilter "UDP"

  # Run with alarm notifications
  tempest-homekit-go --token "your-token" --station "My Station" --alarms @alarms.json

  # Edit alarm configuration (standalone)
  tempest-homekit-go --alarms-edit @alarms.json --alarms-edit-port 8081

  # Start webhook listener (standalone)
  tempest-homekit-go --webhook-listener

ENVIRONMENT VARIABLES:
  All flags can also be set via environment variables (see individual flag descriptions above).
  Command-line flags take precedence over environment variables.

For more information, visit: https://github.com/BEHRConsulting/tempest-homekit-go
`)
}

// LoadConfig initializes and returns a new Config struct with values from
// environment variables, command-line flags, and sensible defaults.
func LoadConfig() *Config {
	cfg := &Config{
		Token:                getEnvOrDefault("TEMPEST_TOKEN", ""),
		StationName:          getEnvOrDefault("TEMPEST_STATION_NAME", ""),
		Pin:                  getEnvOrDefault("HOMEKIT_PIN", "00102003"),
		LogLevel:             getEnvOrDefault("LOG_LEVEL", "error"),
		LogFilter:            getEnvOrDefault("LOG_FILTER", ""),
		WebPort:              getEnvOrDefault("WEB_PORT", "8080"),
		Sensors:              getEnvOrDefault("SENSORS", "temp,lux,humidity,uv"),
		ReadHistory:          getEnvOrDefault("READ_HISTORY", "") == "true",
		StationURL:           getEnvOrDefault("STATION_URL", ""),
		UDPStream:            getEnvOrDefault("UDP_STREAM", "") == "true",
		DisableInternet:      getEnvOrDefault("DISABLE_INTERNET", "") == "true",
		Elevation:            275.2, // 903ft default elevation in meters
		Units:                getEnvOrDefault("UNITS", "imperial"),
		UnitsPressure:        getEnvOrDefault("UNITS_PRESSURE", "inHg"),
		HistoryPoints:        parseIntEnv("HISTORY_POINTS", 1000),
		ChartHistoryHours:    parseIntEnv("CHART_HISTORY_HOURS", 24),
		GeneratedWeatherPath: getEnvOrDefault("GENERATE_WEATHER_PATH", "/api/generate-weather"),
		Alarms:               getEnvOrDefault("ALARMS", ""),
		AlarmsEdit:           getEnvOrDefault("ALARMS_EDIT", ""),
		AlarmsEditPort:       getEnvOrDefault("ALARMS_EDIT_PORT", "8081"),
		WebhookListener:      getEnvOrDefault("WEBHOOK_LISTENER", "") == "true",
		WebhookListenPort:    getEnvOrDefault("WEBHOOK_LISTEN_PORT", "8082"),
		EnvFile:              getEnvOrDefault("ENV_FILE", ".env"),
		Status:               getEnvOrDefault("STATUS", "") == "true",
		StatusRefresh:        parseIntEnv("STATUS_REFRESH", 5),
		StatusTimeout:        parseIntEnv("STATUS_TIMEOUT", 0),
		StatusTheme:          getEnvOrDefault("STATUS_THEME", "dark-ocean"),
	}

	// Set custom usage function
	flag.Usage = customUsage

	var elevationStr string
	var elevationProvided bool
	flag.StringVar(&cfg.Token, "token", cfg.Token, "WeatherFlow API token")
	flag.StringVar(&cfg.StationName, "station", cfg.StationName, "Tempest station name")
	flag.StringVar(&cfg.Pin, "pin", cfg.Pin, "HomeKit PIN")
	flag.StringVar(&cfg.LogLevel, "loglevel", cfg.LogLevel, "Log level (debug, info, error)")
	flag.StringVar(&cfg.LogFilter, "logfilter", cfg.LogFilter, "Filter log messages to only show those containing this string (case-insensitive)")
	flag.StringVar(&cfg.WebPort, "web-port", cfg.WebPort, "Web dashboard port")
	flag.StringVar(&cfg.Sensors, "sensors", cfg.Sensors, "Sensors to enable: 'all', 'min' (temp,humidity,lux), or comma-delimited list (temp/temperature,humidity,lux/light,wind,rain,pressure,uv/uvi,lightning)")
	flag.StringVar(&elevationStr, "elevation", "", "Station elevation (e.g., 903ft, 275m). If not provided, elevation will be auto-detected from coordinates")
	flag.BoolVar(&cfg.ClearDB, "cleardb", false, "Clear HomeKit database and reset device pairing")
	flag.BoolVar(&cfg.DisableHomeKit, "disable-homekit", false, "Disable HomeKit services and run web console only")
	flag.BoolVar(&cfg.DisableAlarms, "disable-alarms", false, "Disable alarm initialization and processing")
	flag.BoolVar(&cfg.ReadHistory, "read-history", cfg.ReadHistory, "Preload historical observations from Tempest API up to HISTORY_POINTS")
	flag.BoolVar(&cfg.TestAPI, "test-api", false, "Test WeatherFlow API endpoints and data points")
	flag.BoolVar(&cfg.TestAPILocal, "test-api-local", false, "Test local web server API endpoints and exit")
	flag.StringVar(&cfg.TestEmail, "test-email", "", "Send a test email to the specified address and exit")
	flag.StringVar(&cfg.TestSMS, "test-sms", "", "Send a test SMS to the specified phone number (E.164 format) and exit")
	flag.StringVar(&cfg.TestWebhook, "test-webhook", "", "Send a test webhook to the specified URL and exit")
	flag.BoolVar(&cfg.TestConsole, "test-console", false, "Send a test console notification and exit")
	flag.BoolVar(&cfg.TestSyslog, "test-syslog", false, "Send a test syslog notification and exit")
	flag.BoolVar(&cfg.TestOSLog, "test-oslog", false, "Send a test oslog notification and exit (macOS only)")
	flag.BoolVar(&cfg.TestEventLog, "test-eventlog", false, "Send a test eventlog notification and exit (Windows only)")
	flag.IntVar(&cfg.TestUDP, "test-udp", 0, "Listen for UDP broadcasts for N seconds (default: 120) and exit")
	flag.BoolVar(&cfg.TestHomeKit, "test-homekit", false, "Test HomeKit bridge setup and pairing info, then exit")
	flag.BoolVar(&cfg.TestWebStatus, "test-web-status", false, "Test web status scraping from TempestWX and exit")
	flag.StringVar(&cfg.TestAlarm, "test-alarm", "", "Trigger a specific alarm by name for testing and exit")
	flag.BoolVar(&cfg.UseWebStatus, "use-web-status", false, "Enable headless browser scraping of TempestWX status page every 15 minutes")
	flag.StringVar(&cfg.StationURL, "station-url", cfg.StationURL, "Custom station URL for weather data (e.g., http://localhost:8080/api/generate-weather). Overrides Tempest API. Can also be set via STATION_URL environment variable")
	flag.BoolVar(&cfg.UseGeneratedWeather, "use-generated-weather", false, "Use generated weather data for UI testing instead of Tempest API")
	flag.BoolVar(&cfg.UDPStream, "udp-stream", cfg.UDPStream, "Listen for UDP broadcasts from local Tempest station (port 50222) for offline operation. Can also be set via UDP_STREAM environment variable")
	flag.BoolVar(&cfg.DisableInternet, "disable-internet", cfg.DisableInternet, "Disable all internet access (no WeatherFlow API calls, no status scraping). Requires --udp-stream or --use-generated-weather. Can also be set via DISABLE_INTERNET environment variable")
	flag.BoolVar(&cfg.DisableWebConsole, "disable-webconsole", false, "Disable web server (HomeKit only mode)")
	flag.StringVar(&cfg.Units, "units", cfg.Units, "Units system: imperial (default), metric, or sae. Can also be set via UNITS environment variable")
	flag.StringVar(&cfg.UnitsPressure, "units-pressure", cfg.UnitsPressure, "Pressure units: inHg (default) or mb. Can also be set via UNITS_PRESSURE environment variable")
	flag.IntVar(&cfg.HistoryPoints, "history", cfg.HistoryPoints, "Number of data points to store in history (default: 1000, min: 10). Can also be set via HISTORY_POINTS environment variable")
	flag.IntVar(&cfg.ChartHistoryHours, "chart-history", cfg.ChartHistoryHours, "Number of hours of data to display in charts (default: 24, 0=all). Can also be set via CHART_HISTORY_HOURS environment variable")
	flag.StringVar(&cfg.GeneratedWeatherPath, "generate-path", cfg.GeneratedWeatherPath, "Path for generated weather endpoint (default: /api/generate-weather). Can also be set via GENERATE_WEATHER_PATH environment variable")
	flag.StringVar(&cfg.Alarms, "alarms", cfg.Alarms, "Alarm configuration: @filename.json or inline JSON string")
	flag.StringVar(&cfg.AlarmsEdit, "alarms-edit", cfg.AlarmsEdit, "Run alarm editor for specified config file: @filename.json")
	flag.StringVar(&cfg.AlarmsEditPort, "alarms-edit-port", cfg.AlarmsEditPort, "Port for alarm editor web UI (default: 8081)")
	flag.BoolVar(&cfg.WebhookListener, "webhook-listener", cfg.WebhookListener, "Start webhook listener server (default port: 8082)")
	flag.StringVar(&cfg.WebhookListenPort, "webhook-listener-port", cfg.WebhookListenPort, "Port for webhook listener server (default: 8082)")
	flag.StringVar(&cfg.EnvFile, "env", cfg.EnvFile, "Custom environment file to load (default: .env)")
	flag.BoolVar(&cfg.Status, "status", cfg.Status, "Enable curses-based status console (TUI mode)")
	flag.IntVar(&cfg.StatusRefresh, "status-refresh", cfg.StatusRefresh, "Status refresh interval in seconds (default: 5)")
	flag.IntVar(&cfg.StatusTimeout, "status-timeout", cfg.StatusTimeout, "Auto-exit after N seconds (0 = never, default: 0)")
	flag.StringVar(&cfg.StatusTheme, "status-theme", cfg.StatusTheme, "Color theme for status console (default: dark-ocean)")
	flag.BoolVar(&cfg.StatusThemeList, "status-theme-list", false, "List all available color themes and exit")
	flag.BoolVar(&cfg.Version, "version", false, "Show version information and exit")

	// Parse flags but check if elevation was actually provided
	flag.Parse()

	// Handle station URL configuration
	if cfg.StationURL != "" {
		// If station URL is provided, automatically enable generated weather mode
		cfg.UseGeneratedWeather = true
	} else if cfg.UseGeneratedWeather {
		// If use-generated-weather is set but no custom URL, use default local endpoint
		cfg.StationURL = fmt.Sprintf("http://localhost:%s%s", cfg.WebPort, cfg.GeneratedWeatherPath)
	}

	// Validate command line arguments
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n\n", err)
		flag.Usage()
		os.Exit(2)
	}

	// Check if elevation was provided by user
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "elevation" {
			elevationProvided = true
		}
		if f.Name == "webhook-listener" {
			cfg.WebhookListenerSet = true
		}
		if f.Name == "webhook-listener-port" {
			cfg.WebhookPortSet = true
		}
	})

	// Handle elevation configuration - auto lookup by default
	if !elevationProvided || strings.ToLower(elevationStr) == "auto" {
		// Skip station elevation lookup if using generated weather - elevation will be set later from generated location
		if !cfg.UseGeneratedWeather {
			if elevation, err := lookupStationElevation(cfg.Token, cfg.StationName); err != nil {
				log.Printf("Warning: Failed to lookup elevation automatically: %v", err)
				log.Printf("INFO: Using fallback elevation 903ft (275.2m)")
			} else {
				cfg.Elevation = elevation
				// Don't log here - will be logged later in main.go after logger is set up
			}
		}
		// For generated weather, elevation will be set by the service from the generated location
	} else {
		// Parse manually provided elevation with units
		if elevation, err := parseElevation(elevationStr); err != nil {
			log.Printf("Warning: Invalid elevation format '%s', using fallback 903ft (275.2m): %v", elevationStr, err)
		} else {
			cfg.Elevation = elevation
			log.Printf("INFO: Using specified elevation: %.1f meters (%.0f feet)", elevation, elevation*3.28084)
		}
	}

	return cfg
}

// validateConfig validates command line arguments and returns an error if invalid
func validateConfig(cfg *Config) error {
	// Ensure sensible defaults for fields when Config structs are created programmatically
	// Some tests construct Config with empty values and expect sensible defaults to be applied.
	if strings.TrimSpace(cfg.Units) == "" {
		cfg.Units = "imperial"
	}
	if strings.TrimSpace(cfg.UnitsPressure) == "" {
		cfg.UnitsPressure = "inHg"
	}
	// Default history/chart values when zero-valued Config is used in tests or programmatically
	if cfg.HistoryPoints == 0 {
		cfg.HistoryPoints = 1000
	}
	if cfg.ChartHistoryHours == 0 {
		cfg.ChartHistoryHours = 24
	}
	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "warning", "error"}
	validLevel := false
	for _, level := range validLogLevels {
		if cfg.LogLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level '%s'. Valid options: debug, info, warn/warning, error", cfg.LogLevel)
	}

	// Validate sensor configuration by testing parsing
	if cfg.Sensors != "" {
		// Test if sensor config is valid by attempting to parse it
		// This will help catch invalid sensor names early
		validSensorNames := []string{"temp", "temperature", "humidity", "lux", "light", "wind", "rain", "pressure", "uv", "uvi", "lightning"}
		validPresets := []string{"all", "min"}

		// Check if it's a preset
		isPreset := false
		for _, preset := range validPresets {
			if cfg.Sensors == preset {
				isPreset = true
				break
			}
		}

		if !isPreset {
			// Parse comma-separated sensor list
			sensors := strings.Split(strings.ToLower(cfg.Sensors), ",")
			for _, sensor := range sensors {
				sensor = strings.TrimSpace(sensor)
				if sensor == "" {
					continue
				}
				valid := false
				for _, validName := range validSensorNames {
					if sensor == validName {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("invalid sensor '%s'. Valid sensors: %s. Valid presets: %s",
						sensor, strings.Join(validSensorNames, ", "), strings.Join(validPresets, ", "))
				}
			}
		}
	}

	// Validate web port is numeric
	if _, err := strconv.Atoi(cfg.WebPort); err != nil {
		return fmt.Errorf("invalid web port '%s'. Port must be a number", cfg.WebPort)
	}

	// Validate webhook listen port is numeric
	if cfg.WebhookListenPort != "" {
		if _, err := strconv.Atoi(cfg.WebhookListenPort); err != nil {
			return fmt.Errorf("invalid webhook listen port '%s'. Port must be a number", cfg.WebhookListenPort)
		}
	}

	// Validate HomeKit PIN format (8 digits)
	if len(cfg.Pin) != 8 {
		return fmt.Errorf("invalid HomeKit PIN '%s'. PIN must be exactly 8 digits", cfg.Pin)
	}
	if _, err := strconv.Atoi(cfg.Pin); err != nil {
		return fmt.Errorf("invalid HomeKit PIN '%s'. PIN must contain only digits", cfg.Pin)
	}

	// Validate required fields for WeatherFlow API mode
	// The WeatherFlow API token is required only when using the WeatherFlow API as the
	// data source. If a custom station URL is provided via --station-url, the
	// --use-generated-weather flag is set, or --udp-stream is enabled, a WeatherFlow token is not necessary.
	// Also skip token requirement for alarm editor mode.
	usingWeatherFlowAPI := cfg.StationURL == "" && !cfg.UseGeneratedWeather && !cfg.UDPStream && cfg.AlarmsEdit == ""

	if usingWeatherFlowAPI {
		if cfg.Token == "" {
			return fmt.Errorf("WeatherFlow API token is required when using the WeatherFlow API as the data source. Set via --token flag or TEMPEST_TOKEN environment variable, or use --station-url/--use-generated-weather/--udp-stream for token-less modes")
		}
		if cfg.StationName == "" {
			return fmt.Errorf("both --token and --station are required when using the WeatherFlow API. Set station via --station flag or TEMPEST_STATION_NAME environment variable")
		}
	}

	// Validate DisableInternet mode requires a local data source (UDP or Generated Weather)
	if cfg.DisableInternet && !cfg.UDPStream && !cfg.UseGeneratedWeather {
		return fmt.Errorf("--disable-internet mode requires --udp-stream or --use-generated-weather (need a local data source)")
	}

	// Validate DisableInternet mode is incompatible with internet-dependent features
	if cfg.DisableInternet {
		if cfg.UseWebStatus {
			return fmt.Errorf("--use-web-status cannot be used with --disable-internet (requires internet access)")
		}
		if cfg.ReadHistory {
			return fmt.Errorf("--read-history cannot be used with --disable-internet (requires WeatherFlow API access)")
		}
	}

	// Validate DisableHomeKit and DisableWebConsole are mutually exclusive
	if cfg.DisableHomeKit && cfg.DisableWebConsole {
		return fmt.Errorf("--disable-homekit and --disable-webconsole cannot be used together (would disable everything)")
	}

	// Station name is required for non-alarm-editor modes (already checked above for API mode)
	if cfg.StationName == "" && cfg.AlarmsEdit == "" && !usingWeatherFlowAPI {
		return fmt.Errorf("station name is required. Set via --station flag or TEMPEST_STATION_NAME environment variable")
	}

	// Validate units
	validUnits := []string{"imperial", "metric", "sae"}
	validUnit := false
	for _, unit := range validUnits {
		if cfg.Units == unit {
			validUnit = true
			break
		}
	}
	if !validUnit {
		return fmt.Errorf("invalid units '%s'. Valid options: imperial, metric, sae", cfg.Units)
	}

	// Validate pressure units
	validPressureUnits := []string{"inHg", "mb"}
	validPressureUnit := false
	for _, unit := range validPressureUnits {
		if cfg.UnitsPressure == unit {
			validPressureUnit = true
			break
		}
	}
	if !validPressureUnit {
		return fmt.Errorf("invalid pressure units '%s'. Valid options: inHg, mb", cfg.UnitsPressure)
	}

	// Validate history points
	if cfg.HistoryPoints < 10 {
		return fmt.Errorf("history points must be at least 10 (got %d)", cfg.HistoryPoints)
	}
	// Validate chart history hours (0 means all, so only check if positive)
	if cfg.ChartHistoryHours < 0 {
		return fmt.Errorf("chart history hours must be 0 (all data) or positive (got %d)", cfg.ChartHistoryHours)
	}

	return nil
}

// ClearDatabase removes all files in the HomeKit database directory
func ClearDatabase(dbPath string) error {
	log.Printf("INFO: Clearing HomeKit database at: %s", dbPath)

	// Check if directory exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("INFO: Database directory does not exist: %s", dbPath)
		return nil
	}

	// Remove all files in the directory
	files, err := filepath.Glob(filepath.Join(dbPath, "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			log.Printf("Warning: Failed to remove %s: %v", file, err)
		} else {
			log.Printf("INFO: Removed: %s", filepath.Base(file))
		}
	}

	log.Printf("INFO: HomeKit database cleared successfully")
	return nil
}

// SensorConfig represents which sensors should be enabled
type SensorConfig struct {
	Temperature bool
	Humidity    bool
	Light       bool
	Wind        bool
	Rain        bool
	Pressure    bool
	UV          bool
	Lightning   bool
}

// ParseSensorConfig parses the sensor configuration string and returns a SensorConfig
// with appropriate sensor types enabled based on the input string.
// Supported values: "all", "min", or comma-separated sensor names.
func ParseSensorConfig(sensorsFlag string) SensorConfig {
	switch strings.ToLower(sensorsFlag) {
	case "all":
		return SensorConfig{
			Temperature: true,
			Humidity:    true,
			Light:       true,
			Wind:        true,
			Rain:        true,
			Pressure:    true,
			UV:          true,
			Lightning:   true,
		}
	case "min":
		return SensorConfig{
			Temperature: true,
			Humidity:    true,
			Light:       true,
			// Minimal sensors: temperature, humidity, and light for basic weather monitoring
		}
	default:
		// Parse comma-delimited sensor list
		sensors := strings.Split(strings.ToLower(sensorsFlag), ",")
		config := SensorConfig{}
		for _, sensor := range sensors {
			sensor = strings.TrimSpace(sensor)
			switch sensor {
			case "temp", "temperature":
				config.Temperature = true
			case "humidity":
				config.Humidity = true
			case "light", "lux":
				config.Light = true
			case "wind":
				config.Wind = true
			case "rain":
				config.Rain = true
			case "pressure":
				config.Pressure = true
			case "uv", "uvi":
				config.UV = true
			case "lightning":
				config.Lightning = true
			}
		}
		return config
	}
}

// StationLocation represents station coordinates from WeatherFlow API
type StationLocation struct {
	StationID int     `json:"station_id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Elevation float64 `json:"elevation,omitempty"` // May be provided directly
}

// ElevationResponse represents response from elevation API
type ElevationResponse struct {
	Results []struct {
		Elevation float64 `json:"elevation"`
	} `json:"results"`
}

// lookupStationElevation attempts to get elevation from station coordinates
func lookupStationElevation(token, stationName string) (float64, error) {
	// First try to get station coordinates from WeatherFlow API
	lat, lon, err := getStationCoordinates(token, stationName)
	if err != nil {
		return 0, fmt.Errorf("failed to get station coordinates: %v", err)
	}

	// Then lookup elevation from coordinates
	elevation, err := getElevationFromCoordinates(lat, lon)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup elevation for coordinates (%.4f, %.4f): %v", lat, lon, err)
	}

	return elevation, nil
}

// getStationCoordinates fetches station coordinates from WeatherFlow API
func getStationCoordinates(token, stationName string) (lat, lon float64, err error) {
	// First try to get actual station coordinates from WeatherFlow API
	if coords, err := fetchWeatherFlowStationCoords(token, stationName); err == nil {
		return coords[0], coords[1], nil
	}

	// Fallback to known coordinates for common locations
	knownLocations := map[string][2]float64{
		"Chino Hills": {33.9898, -117.7326},
		"Los Angeles": {34.0522, -118.2437},
		"San Diego":   {32.7157, -117.1611},
		"Phoenix":     {33.4484, -112.0740},
		"Denver":      {39.7392, -104.9903},
		"Seattle":     {47.6062, -122.3321},
		"Portland":    {45.5152, -122.6784},
		"Austin":      {30.2672, -97.7431},
		"Dallas":      {32.7767, -96.7970},
		"Miami":       {25.7617, -80.1918},
	}

	if coords, found := knownLocations[stationName]; found {
		return coords[0], coords[1], nil
	}

	return 0, 0, fmt.Errorf("coordinates not available for station '%s' (consider adding coordinates to known locations)", stationName)
}

// fetchWeatherFlowStationCoords attempts to get coordinates from WeatherFlow API
func fetchWeatherFlowStationCoords(_token, _stationName string) (coords [2]float64, err error) {
	// Explicitly ignore unused parameters to satisfy linter
	_ = _token
	_ = _stationName

	// This would query the WeatherFlow API stations endpoint for detailed station info
	// The API might have an endpoint like: /stations/:station_id/details that includes lat/lon
	// For now, we return an error to fall back to known locations

	// TODO: Implement actual WeatherFlow API call when station details endpoint is available
	// Example implementation would be:
	/*
		url := fmt.Sprintf("https://swd.weatherflow.com/swd/rest/stations/%s/details?token=%s", stationID, token)
		resp, err := http.Get(url)
		if err != nil {
			return coords, err
		}
		defer resp.Body.Close()

		var stationDetails StationDetailsResponse
		if err := json.NewDecoder(resp.Body).Decode(&stationDetails); err != nil {
			return coords, err
		}

		if len(stationDetails.Stations) > 0 {
			station := stationDetails.Stations[0]
			coords[0] = station.Latitude
			coords[1] = station.Longitude
			return coords, nil
		}
	*/

	return coords, fmt.Errorf("WeatherFlow station coordinates API not implemented")
}

// getElevationFromCoordinates uses Open Elevation API to get elevation
func getElevationFromCoordinates(lat, lon float64) (float64, error) {
	// Use Open Elevation API (free, no API key required)
	url := fmt.Sprintf("https://api.open-elevation.com/api/v1/lookup?locations=%.4f,%.4f", lat, lon)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("elevation API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("elevation API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read elevation API response: %v", err)
	}

	var elevResp ElevationResponse
	if err := json.Unmarshal(body, &elevResp); err != nil {
		return 0, fmt.Errorf("failed to parse elevation API response: %v", err)
	}

	if len(elevResp.Results) == 0 {
		return 0, fmt.Errorf("no elevation data returned")
	}

	return elevResp.Results[0].Elevation, nil
}

// parseElevation parses elevation string with units (e.g., "903ft", "275m") and returns meters
func parseElevation(elevationStr string) (float64, error) {
	elevationStr = strings.TrimSpace(strings.ToLower(elevationStr))

	var meters float64
	var err error

	if strings.HasSuffix(elevationStr, "ft") {
		// Parse feet and convert to meters
		valueStr := strings.TrimSuffix(elevationStr, "ft")
		feet, parseErr := strconv.ParseFloat(valueStr, 64)
		if parseErr != nil {
			return 0, parseErr
		}
		meters = feet * 0.3048 // 1 foot = 0.3048 meters
	} else if strings.HasSuffix(elevationStr, "m") {
		// Parse meters directly
		valueStr := strings.TrimSuffix(elevationStr, "m")
		meters, err = strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return 0, err
		}
	} else {
		// Try to parse as number without unit, assume meters
		meters, err = strconv.ParseFloat(elevationStr, 64)
		if err != nil {
			return 0, err
		}
	}

	// Validate elevation range: -1411ft to 29029ft (-430m to 8848m)
	// Dead Sea area is the lowest at -430m, Mount Everest is highest at 8848m
	// Add small tolerance for floating point precision
	const minElevationMeters = -430.1 // -1411 feet with tolerance
	const maxElevationMeters = 8848.1 // 29029 feet with tolerance

	if meters < minElevationMeters {
		return 0, fmt.Errorf("elevation %.1fm is below Earth's lowest point (%.1fm, Dead Sea area)", meters, minElevationMeters)
	}
	if meters > maxElevationMeters {
		return 0, fmt.Errorf("elevation %.1fm is above Earth's highest point (%.1fm, Mount Everest)", meters, maxElevationMeters)
	}

	return meters, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseIntEnv parses an integer from environment variable or returns default
func parseIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
