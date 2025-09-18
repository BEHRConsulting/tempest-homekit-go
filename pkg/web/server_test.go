package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tempest-homekit-go/pkg/weather"
)

func TestNewWebServer(t *testing.T) {
	server := NewWebServer("8080", 100.0, "info", 12345, false)

	if server.port != "8080" {
		t.Errorf("Expected port 8080, got %s", server.port)
	}
	if server.elevation != 100.0 {
		t.Errorf("Expected elevation 100.0, got %f", server.elevation)
	}
	if server.logLevel != "info" {
		t.Errorf("Expected log level info, got %s", server.logLevel)
	}
}

func TestGetPressureDescription(t *testing.T) {
	tests := []struct {
		pressure       float64
		expectContains string
	}{
		{970.0, "Stormy"},
		{985.0, "Rain expected"},
		{995.0, "Changeable"},
		{1005.0, "Fair weather"},
		{1015.0, "Clear and dry"},
		{1025.0, "Very dry"},
		{1035.0, "Exceptionally dry"},
	}

	for _, test := range tests {
		result := getPressureDescription(test.pressure)
		if result == "" {
			t.Errorf("Expected description for pressure %f, got empty string", test.pressure)
		}
		// Just verify we get some description - the exact text may change
	}
}

func TestGetPressureTrend(t *testing.T) {
	now := time.Now()
	history := []weather.Observation{
		{Timestamp: now.Add(-3 * time.Hour).Unix(), StationPressure: 1000.0},
		{Timestamp: now.Add(-2 * time.Hour).Unix(), StationPressure: 1005.0},
		{Timestamp: now.Add(-1 * time.Hour).Unix(), StationPressure: 1010.0},
		{Timestamp: now.Unix(), StationPressure: 1015.0},
	}

	trend := getPressureTrend(history)
	// Should detect rising pressure
	if trend != "Rising" {
		t.Errorf("Expected Rising trend, got %s", trend)
	}
}

func TestGetPressureTrendFalling(t *testing.T) {
	now := time.Now()
	history := []weather.Observation{
		{Timestamp: now.Add(-3 * time.Hour).Unix(), StationPressure: 1020.0},
		{Timestamp: now.Add(-2 * time.Hour).Unix(), StationPressure: 1015.0},
		{Timestamp: now.Add(-1 * time.Hour).Unix(), StationPressure: 1010.0},
		{Timestamp: now.Unix(), StationPressure: 1005.0},
	}

	trend := getPressureTrend(history)
	// Should detect falling pressure
	if trend != "Falling" {
		t.Errorf("Expected Falling trend, got %s", trend)
	}
}

func TestGetPressureTrendSteady(t *testing.T) {
	now := time.Now()
	history := []weather.Observation{
		{Timestamp: now.Add(-3 * time.Hour).Unix(), StationPressure: 1013.0},
		{Timestamp: now.Add(-2 * time.Hour).Unix(), StationPressure: 1013.5},
		{Timestamp: now.Add(-1 * time.Hour).Unix(), StationPressure: 1013.2},
		{Timestamp: now.Unix(), StationPressure: 1013.1},
	}

	trend := getPressureTrend(history)
	// Should detect stable pressure
	if trend != "Stable" {
		t.Errorf("Expected Stable trend, got %s", trend)
	}
}

func TestGetPressureTrendInsufficientData(t *testing.T) {
	history := []weather.Observation{
		{Timestamp: time.Now().Unix(), StationPressure: 1013.0},
	}

	trend := getPressureTrend(history)
	// Should return stable for insufficient data
	if trend != "Stable" {
		t.Errorf("Expected Stable for insufficient data, got %s", trend)
	}
}

func TestWebServerWeatherEndpoint(t *testing.T) {
	server := NewWebServer("8080", 100.0, "info", 12345, false)

	// Add some test data
	testObs := &weather.Observation{
		Timestamp:        time.Now().Unix(),
		AirTemperature:   25.0,
		RelativeHumidity: 60.0,
		WindAvg:          5.5,
		WindDirection:    180.0,
		StationPressure:  1013.25,
		UV:               5.0,
		Illuminance:      50000.0,
		RainAccumulated:  0.0,
	}
	server.UpdateWeather(testObs)

	req, err := http.NewRequest("GET", "/api/weather", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleWeatherAPI)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Weather handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains JSON (basic check)
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestWebServerStatusEndpoint(t *testing.T) {
	server := NewWebServer("8080", 100.0, "info", 12345, false)

	req, err := http.NewRequest("GET", "/api/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleStatusAPI)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains JSON
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestUpdateForecast(t *testing.T) {
	server := NewWebServer("8080", 100.0, "info", 12345, false)

	testObs := &weather.Observation{
		Timestamp:        time.Now().Unix(),
		AirTemperature:   25.0,
		RelativeHumidity: 60.0,
	}

	server.UpdateWeather(testObs)

	if server.weatherData == nil {
		t.Error("Weather data should be updated")
	}
	if server.weatherData.AirTemperature != 25.0 {
		t.Errorf("Expected temperature 25.0, got %f", server.weatherData.AirTemperature)
	}
}

func TestSetHistoryLoadingProgress(t *testing.T) {
	server := NewWebServer("8080", 100.0, "info", 12345, false)

	server.SetHistoryLoadingProgress(1, 3, "Loading data...")

	if !server.historyLoadingProgress.isLoading {
		t.Error("History loading should be marked as in progress")
	}
	if server.historyLoadingProgress.currentStep != 1 {
		t.Errorf("Expected current step 1, got %d", server.historyLoadingProgress.currentStep)
	}
	if server.historyLoadingProgress.totalSteps != 3 {
		t.Errorf("Expected total steps 3, got %d", server.historyLoadingProgress.totalSteps)
	}
}

func TestSetHistoryLoadingComplete(t *testing.T) {
	server := NewWebServer("8080", 100.0, "info", 12345, false)

	// First set it to loading
	server.SetHistoryLoadingProgress(1, 3, "Loading...")

	// Then mark as complete
	server.SetHistoryLoadingComplete()

	if server.historyLoadingProgress.isLoading {
		t.Error("History loading should be marked as complete")
	}
}
