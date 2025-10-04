package web

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"tempest-homekit-go/pkg/generator"
	"tempest-homekit-go/pkg/weather"
)

// TestRegenerateWeatherEndpoint verifies that POST /api/regenerate-weather calls
// the generator's GenerateNewSeason and updates ws.generatedWeather fields.
func TestRegenerateWeatherEndpoint(t *testing.T) {
	// Create a fake generator with deterministic values
	obs := &weather.Observation{Timestamp: 1234567890}
	cfg := &fakeGeneratorConfig{
		LocationName:   "UnitTestVille",
		Season:         generator.Season(1),
		DailyRainTotal: 2.5,
		ClimateZone:    "UnitClimate",
		Observation:    obs,
	}
	fg := newFakeGenerator(cfg)

	// Create generated weather info enabled so handler will operate
	gw := &GeneratedWeatherInfo{Enabled: true, Location: "", Season: "", ClimateZone: ""}

	ws := NewWebServer("0", 10.0, "debug", 0, false, "test", "", gw, fg, "metric", "mb", 1000, 24)

	// Build request and recorder
	req := httptest.NewRequest("POST", "/api/regenerate-weather", strings.NewReader(""))
	rr := httptest.NewRecorder()

	// Call handler
	ws.handleRegenerateWeatherAPI(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected status 200, got %d; body=%s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if success, ok := resp["success"].(bool); !ok || !success {
		t.Fatalf("expected success=true in response, got %v", resp["success"])
	}

	// Verify generatedWeather on server updated
	if ws.generatedWeather == nil {
		t.Fatalf("generatedWeather should not be nil")
	}
	if ws.generatedWeather.Location != "UnitTestVille" {
		t.Fatalf("expected generatedWeather.Location to be UnitTestVille, got %q", ws.generatedWeather.Location)
	}
	if ws.generatedWeather.Season == "" {
		t.Fatalf("expected generatedWeather.Season to be set, got empty string")
	}
	if ws.generatedWeather.ClimateZone != "UnitClimate" {
		t.Fatalf("expected generatedWeather.ClimateZone to be UnitClimate, got %q", ws.generatedWeather.ClimateZone)
	}
}
