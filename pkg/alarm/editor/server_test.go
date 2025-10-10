package editor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"tempest-homekit-go/pkg/alarm"
	"testing"
)

func TestNewServer(t *testing.T) {
	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "alarms_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write a minimal config
	config := alarm.AlarmConfig{
		Alarms: []alarm.Alarm{
			{
				Name:      "test-alarm",
				Condition: "temperature > 85",
				Enabled:   true,
			},
		},
	}
	data, _ := json.Marshal(config)
	tmpfile.Write(data)
	tmpfile.Close()

	// Create server
	server, err := NewServer("@"+tmpfile.Name(), "8081")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server.configPath != tmpfile.Name() {
		t.Errorf("Expected configPath %s, got %s", tmpfile.Name(), server.configPath)
	}

	if server.port != "8081" {
		t.Errorf("Expected port 8081, got %s", server.port)
	}

	if len(server.config.Alarms) != 1 {
		t.Errorf("Expected 1 alarm, got %d", len(server.config.Alarms))
	}
}

func TestHandleGetConfig(t *testing.T) {
	// Create test server
	config := &alarm.AlarmConfig{
		Alarms: []alarm.Alarm{
			{
				Name:      "test-alarm",
				Condition: "temperature > 85",
				Enabled:   true,
			},
		},
	}

	server := &Server{
		configPath: "test.json",
		port:       "8081",
		config:     config,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	server.handleGetConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result alarm.AlarmConfig
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result.Alarms) != 1 {
		t.Errorf("Expected 1 alarm, got %d", len(result.Alarms))
	}

	if result.Alarms[0].Name != "test-alarm" {
		t.Errorf("Expected alarm name 'test-alarm', got '%s'", result.Alarms[0].Name)
	}
}

func TestHandleListAlarms(t *testing.T) {
	config := &alarm.AlarmConfig{
		Alarms: []alarm.Alarm{
			{
				Name:      "high-temp",
				Condition: "temperature > 85",
				Tags:      []string{"temperature", "heat"},
				Enabled:   true,
			},
			{
				Name:      "lightning",
				Condition: "lightning_strike_count > 0",
				Tags:      []string{"lightning"},
				Enabled:   true,
			},
		},
	}

	server := &Server{
		configPath: "test.json",
		port:       "8081",
		config:     config,
	}

	// Test without filter
	req := httptest.NewRequest(http.MethodGet, "/api/alarms", nil)
	w := httptest.NewRecorder()
	server.handleListAlarms(w, req)

	var result []alarm.Alarm
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("Expected 2 alarms, got %d", len(result))
	}

	// Test with name filter
	req = httptest.NewRequest(http.MethodGet, "/api/alarms?name=temp", nil)
	w = httptest.NewRecorder()
	server.handleListAlarms(w, req)

	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Errorf("Expected 1 alarm with name filter, got %d", len(result))
	}
	if result[0].Name != "high-temp" {
		t.Errorf("Expected 'high-temp', got '%s'", result[0].Name)
	}

	// Test with tag filter
	req = httptest.NewRequest(http.MethodGet, "/api/alarms?tag=lightning", nil)
	w = httptest.NewRecorder()
	server.handleListAlarms(w, req)

	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Errorf("Expected 1 alarm with tag filter, got %d", len(result))
	}
	if result[0].Name != "lightning" {
		t.Errorf("Expected 'lightning', got '%s'", result[0].Name)
	}
}

func TestHandleGetTags(t *testing.T) {
	config := &alarm.AlarmConfig{
		Alarms: []alarm.Alarm{
			{
				Name: "alarm1",
				Tags: []string{"temperature", "heat"},
			},
			{
				Name: "alarm2",
				Tags: []string{"temperature", "cold"},
			},
			{
				Name: "alarm3",
				Tags: []string{"lightning"},
			},
		},
	}

	server := &Server{
		configPath: "test.json",
		port:       "8081",
		config:     config,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	w := httptest.NewRecorder()
	server.handleGetTags(w, req)

	var tags []string
	json.NewDecoder(w.Body).Decode(&tags)

	// Check that we get unique tags
	expectedTags := map[string]bool{
		"temperature": false,
		"heat":        false,
		"cold":        false,
		"lightning":   false,
	}

	for _, tag := range tags {
		if _, exists := expectedTags[tag]; exists {
			expectedTags[tag] = true
		}
	}

	for tag, found := range expectedTags {
		if !found {
			t.Errorf("Expected tag '%s' not found", tag)
		}
	}
}

func TestHandleValidate(t *testing.T) {
	server := &Server{
		configPath: "test.json",
		port:       "8081",
		config:     &alarm.AlarmConfig{},
	}

	tests := []struct {
		name      string
		condition string
		wantValid bool
	}{
		{"valid simple", "temperature > 85", true},
		{"valid compound", "temperature > 85 && humidity > 80", true},
		{"invalid syntax", "temperature >> 85", false},
		{"invalid field", "fake_field > 100", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"condition":"` + tt.condition + `"}`
			req := httptest.NewRequest(http.MethodPost, "/api/validate", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleValidate(w, req)

			var result map[string]interface{}
			json.NewDecoder(w.Body).Decode(&result)

			valid, ok := result["valid"].(bool)
			if !ok {
				t.Fatal("Response missing 'valid' field")
			}

			if valid != tt.wantValid {
				t.Errorf("Expected valid=%v, got %v", tt.wantValid, valid)
				if !tt.wantValid {
					t.Logf("Error: %v", result["error"])
				}
			}
		})
	}
}

func TestHandleGetFields(t *testing.T) {
	server := &Server{
		configPath: "test.json",
		port:       "8081",
		config:     &alarm.AlarmConfig{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/fields", nil)
	w := httptest.NewRecorder()
	server.handleGetFields(w, req)

	var fields []string
	json.NewDecoder(w.Body).Decode(&fields)

	// Check for some expected fields
	expectedFields := []string{"temperature", "humidity", "pressure", "wind_speed", "lux"}
	for _, expected := range expectedFields {
		found := false
		for _, field := range fields {
			if field == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected field '%s' not found in result", expected)
		}
	}
}
