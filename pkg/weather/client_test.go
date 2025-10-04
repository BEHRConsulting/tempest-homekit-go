package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockStationsResponse creates a mock WeatherFlow API response for stations
func mockStationsResponse() map[string]interface{} {
	return map[string]interface{}{
		"stations": []interface{}{
			map[string]interface{}{
				"station_id":  float64(12345),
				"name":        "Test Station",
				"public_name": "Test Station Public",
				"latitude":    float64(34.0),
				"longitude":   float64(-118.0),
				"timezone":    "America/Los_Angeles",
				"station_meta": map[string]interface{}{
					"elevation": float64(100.0),
				},
			},
			map[string]interface{}{
				"station_id":  float64(67890),
				"name":        "Another Station",
				"public_name": "Another Station Public",
				"latitude":    float64(35.0),
				"longitude":   float64(-119.0),
				"timezone":    "America/Los_Angeles",
				"station_meta": map[string]interface{}{
					"elevation": float64(200.0),
				},
			},
		},
	}
}

// Simple mock for testing - we'll test the actual structures separately

func TestGetObservation_Success(t *testing.T) {
	// Create mock server with observation response in array format
	now := time.Now().Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Match the actual API format: obs is an array of maps
		response := map[string]interface{}{
			"obs": []map[string]interface{}{
				{
					"timestamp":                     float64(now),
					"air_temperature":               22.5,
					"relative_humidity":             65.0,
					"wind_avg":                      2.5,
					"wind_direction":                180.0,
					"station_pressure":              1013.25,
					"battery":                       2.65,
					"illuminance":                   5000.0,
					"uv":                            3.0,
					"wind_gust":                     4.0,
					"wind_lull":                     1.5,
					"solar_radiation":               150.0,
					"rain_accumulated":              0.0,
					"precipitation_type":            0.0,
					"lightning_strike_avg_distance": 0.0,
					"lightning_strike_count":        0.0,
					"report_interval":               60.0,
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test GetObservationFromURL directly
	obs, err := GetObservationFromURL(server.URL)
	if err != nil {
		t.Fatalf("Failed to get observation: %v", err)
	}

	// Verify observation data
	if obs.AirTemperature != 22.5 {
		t.Errorf("Expected temperature 22.5, got %.1f", obs.AirTemperature)
	}
	if obs.RelativeHumidity != 65.0 {
		t.Errorf("Expected humidity 65.0, got %.1f", obs.RelativeHumidity)
	}
	if obs.WindAvg != 2.5 {
		t.Errorf("Expected wind speed 2.5, got %.1f", obs.WindAvg)
	}
	if obs.WindDirection != 180.0 {
		t.Errorf("Expected wind direction 180.0, got %.1f", obs.WindDirection)
	}
	if obs.StationPressure != 1013.25 {
		t.Errorf("Expected pressure 1013.25, got %.1f", obs.StationPressure)
	}
	if obs.Battery != 2.65 {
		t.Errorf("Expected battery 2.65, got %.1f", obs.Battery)
	}
}

func TestGetObservationFromURL_InvalidURL(t *testing.T) {
	_, err := GetObservationFromURL("http://invalid.url.that.does.not.exist.123456789")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestGetObservationFromURL_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	_, err := GetObservationFromURL(server.URL)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetObservationFromURL_EmptyObservations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := &HistoricalResponse{
			Status: map[string]interface{}{
				"status_code":    float64(0),
				"status_message": "SUCCESS",
			},
			Obs: [][]interface{}{}, // Empty observations
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, err := GetObservationFromURL(server.URL)
	if err == nil {
		t.Error("Expected error for empty observations, got nil")
	}
}

func TestGetStations_Success(t *testing.T) {
	mockResp := mockStationsResponse()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify token is passed
		if r.URL.Query().Get("token") == "" {
			t.Error("Expected token parameter in request")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	// We need to test this differently since GetStations uses hardcoded URL
	// For now, let's test the structure
	t.Skip("GetStations uses hardcoded URL, needs refactoring for testability")
}

func TestFindStationByName_Found(t *testing.T) {
	stations := []Station{
		{StationID: 12345, Name: "Test Station", StationName: "Test Station"},
		{StationID: 67890, Name: "Another Station", StationName: "Another"},
	}

	station := FindStationByName(stations, "Test Station")
	if station == nil {
		t.Fatal("Expected to find station, got nil")
	}
	if station.StationID != 12345 {
		t.Errorf("Expected station ID 12345, got %d", station.StationID)
	}
}

func TestFindStationByName_FoundByStationName(t *testing.T) {
	stations := []Station{
		{StationID: 12345, Name: "Test Station", StationName: "TestSt"},
		{StationID: 67890, Name: "Another Station", StationName: "Another"},
	}

	station := FindStationByName(stations, "TestSt")
	if station == nil {
		t.Fatal("Expected to find station by StationName, got nil")
	}
	if station.StationID != 12345 {
		t.Errorf("Expected station ID 12345, got %d", station.StationID)
	}
}

func TestFindStationByName_NotFound(t *testing.T) {
	stations := []Station{
		{StationID: 12345, Name: "Test Station", StationName: "Test Station"},
	}

	station := FindStationByName(stations, "Nonexistent")
	if station != nil {
		t.Error("Expected nil for nonexistent station, got station")
	}
}

func TestFindStationByName_CaseSensitive(t *testing.T) {
	// FindStationByName is case-sensitive
	stations := []Station{
		{StationID: 12345, Name: "Test Station", StationName: "Test Station"},
	}

	station := FindStationByName(stations, "TEST STATION")
	if station != nil {
		t.Error("Expected no match for different case, got station")
	}

	// Exact match should work
	station = FindStationByName(stations, "Test Station")
	if station == nil {
		t.Fatal("Expected exact match, got nil")
	}
}

func TestObservation_DataTypes(t *testing.T) {
	// Test that Observation struct can hold all expected data
	now := time.Now().Unix()
	obs := Observation{
		Timestamp:            now,
		WindLull:             1.5,
		WindAvg:              2.5,
		WindGust:             4.0,
		WindDirection:        180.0,
		StationPressure:      1013.25,
		AirTemperature:       22.5,
		RelativeHumidity:     65.0,
		Illuminance:          5000.0,
		UV:                   3,
		SolarRadiation:       150.0,
		RainAccumulated:      0.5,
		PrecipitationType:    1,
		LightningStrikeAvg:   5.0,
		LightningStrikeCount: 2,
		Battery:              2.65,
		ReportInterval:       60,
	}

	// Verify all fields are correctly stored
	if obs.Timestamp != now {
		t.Errorf("Expected Timestamp %d, got %d", now, obs.Timestamp)
	}
	if obs.WindLull != 1.5 {
		t.Errorf("Expected WindLull 1.5, got %.1f", obs.WindLull)
	}
	if obs.WindAvg != 2.5 {
		t.Errorf("Expected WindAvg 2.5, got %.1f", obs.WindAvg)
	}
	if obs.WindGust != 4.0 {
		t.Errorf("Expected WindGust 4.0, got %.1f", obs.WindGust)
	}
	if obs.WindDirection != 180.0 {
		t.Errorf("Expected WindDirection 180.0, got %.1f", obs.WindDirection)
	}
	if obs.StationPressure != 1013.25 {
		t.Errorf("Expected StationPressure 1013.25, got %.2f", obs.StationPressure)
	}
	if obs.AirTemperature != 22.5 {
		t.Errorf("Expected AirTemperature 22.5, got %.1f", obs.AirTemperature)
	}
	if obs.RelativeHumidity != 65.0 {
		t.Errorf("Expected RelativeHumidity 65.0, got %.1f", obs.RelativeHumidity)
	}
	if obs.Illuminance != 5000.0 {
		t.Errorf("Expected Illuminance 5000.0, got %.1f", obs.Illuminance)
	}
	if obs.UV != 3 {
		t.Errorf("Expected UV 3, got %d", obs.UV)
	}
	if obs.SolarRadiation != 150.0 {
		t.Errorf("Expected SolarRadiation 150.0, got %.1f", obs.SolarRadiation)
	}
	if obs.RainAccumulated != 0.5 {
		t.Errorf("Expected RainAccumulated 0.5, got %.1f", obs.RainAccumulated)
	}
	if obs.PrecipitationType != 1 {
		t.Errorf("Expected PrecipitationType 1, got %d", obs.PrecipitationType)
	}
	if obs.LightningStrikeAvg != 5.0 {
		t.Errorf("Expected LightningStrikeAvg 5.0, got %.1f", obs.LightningStrikeAvg)
	}
	if obs.LightningStrikeCount != 2 {
		t.Errorf("Expected LightningStrikeCount 2, got %d", obs.LightningStrikeCount)
	}
	if obs.Battery != 2.65 {
		t.Errorf("Expected Battery 2.65, got %.2f", obs.Battery)
	}
	if obs.ReportInterval != 60 {
		t.Errorf("Expected ReportInterval 60, got %d", obs.ReportInterval)
	}
}

func TestStation_Structure(t *testing.T) {
	station := Station{
		StationID:   12345,
		Name:        "Test Station",
		StationName: "TestSt",
	}

	if station.StationID != 12345 {
		t.Error("StationID field incorrect")
	}
	if station.Name != "Test Station" {
		t.Error("Name field incorrect")
	}
	if station.StationName != "TestSt" {
		t.Error("StationName field incorrect")
	}
}
