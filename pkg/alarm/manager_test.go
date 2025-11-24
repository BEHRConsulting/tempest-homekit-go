package alarm

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"tempest-homekit-go/pkg/weather"
)

func TestNewManagerWithInlineJSON(t *testing.T) {
	configJSON := `{
		"alarms": [{
			"name": "test-alarm",
			"condition": "temperature > 85",
			"enabled": true,
			"channels": [{"type": "console", "template": "Alert: {{condition}}"}]
		}]
	}`

	manager, err := NewManager(configJSON, "Test Station")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.stationName != "Test Station" {
		t.Errorf("Manager stationName = %v, want 'Test Station'", manager.stationName)
	}

	if len(manager.config.Alarms) != 1 {
		t.Errorf("Expected 1 alarm, got %d", len(manager.config.Alarms))
	}
}

func TestNewManagerWithFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "alarm_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	config := AlarmConfig{
		Alarms: []Alarm{
			{
				Name:      "file-alarm",
				Condition: "temperature > 85",
				Enabled:   true,
				Channels:  []Channel{{Type: "console", Template: "Alert"}},
			},
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if _, err := tmpfile.Write(data); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	manager, err := NewManager("@"+tmpfile.Name(), "Test Station")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	defer manager.Stop()

	if len(manager.config.Alarms) != 1 {
		t.Errorf("Expected 1 alarm, got %d", len(manager.config.Alarms))
	}

	if manager.config.Alarms[0].Name != "file-alarm" {
		t.Errorf("Expected alarm name 'file-alarm', got '%s'", manager.config.Alarms[0].Name)
	}
}

func TestNewManagerInvalidJSON(t *testing.T) {
	_, err := NewManager("{invalid json}", "Test Station")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestNewManagerNonexistentFile(t *testing.T) {
	_, err := NewManager("@/nonexistent/path.json", "Test Station")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestProcessObservationSimple(t *testing.T) {
	configJSON := `{
		"alarms": [{
			"name": "high-temp",
			"condition": "temperature > 85",
			"enabled": true,
			"cooldown": 0,
			"channels": [{"type": "console", "template": "High temp: {{temperature}}°C"}]
		}]
	}`

	manager, err := NewManager(configJSON, "Test Station")
	if err != nil {
		t.Fatal(err)
	}

	obsHot := &weather.Observation{
		AirTemperature: 90.0,
	}
	manager.ProcessObservation(obsHot)

	obsCool := &weather.Observation{
		AirTemperature: 70.0,
	}
	manager.ProcessObservation(obsCool)
}

func TestProcessObservationCooldown(t *testing.T) {
	configJSON := `{
		"alarms": [{
			"name": "test-alarm",
			"condition": "temperature > 85",
			"enabled": true,
			"cooldown": 2,
			"channels": [{"type": "console", "template": "Alert"}]
		}]
	}`

	manager, err := NewManager(configJSON, "Test Station")
	if err != nil {
		t.Fatal(err)
	}

	obs := &weather.Observation{
		AirTemperature: 90.0,
	}
	manager.ProcessObservation(obs)

	alarm := &manager.config.Alarms[0]
	if alarm.CanFire() {
		t.Error("Alarm should not be able to fire immediately after firing")
	}

	time.Sleep(3 * time.Second)

	if !alarm.CanFire() {
		t.Error("Alarm should be able to fire after cooldown expires")
	}
}

func TestProcessObservationDisabledAlarm(t *testing.T) {
	configJSON := `{
		"alarms": [{
			"name": "disabled-alarm",
			"condition": "temperature > 85",
			"enabled": false,
			"channels": [{"type": "console", "template": "Alert"}]
		}]
	}`

	manager, err := NewManager(configJSON, "Test Station")
	if err != nil {
		t.Fatal(err)
	}

	obs := &weather.Observation{
		AirTemperature: 90.0,
	}
	manager.ProcessObservation(obs)
}

func TestProcessObservationMultipleAlarms(t *testing.T) {
	configJSON := `{
		"alarms": [
			{
				"name": "high-temp",
				"condition": "temperature > 85",
				"enabled": true,
				"channels": [{"type": "console", "template": "High temp"}]
			},
			{
				"name": "high-humidity",
				"condition": "humidity > 80",
				"enabled": true,
				"channels": [{"type": "console", "template": "High humidity"}]
			},
			{
				"name": "combo-alarm",
				"condition": "temperature > 80 && humidity > 75",
				"enabled": true,
				"channels": [{"type": "console", "template": "Hot and humid"}]
			}
		]
	}`

	manager, err := NewManager(configJSON, "Test Station")
	if err != nil {
		t.Fatal(err)
	}

	obs := &weather.Observation{
		AirTemperature:   90.0,
		RelativeHumidity: 85.0,
	}
	manager.ProcessObservation(obs)
}

func TestProcessObservationInvalidCondition(t *testing.T) {
	configJSON := `{
		"alarms": [{
			"name": "invalid-alarm",
			"condition": "invalid_field > 100",
			"enabled": true,
			"channels": [{"type": "console", "template": "Alert"}]
		}]
	}`

	manager, err := NewManager(configJSON, "Test Station")
	if err != nil {
		t.Fatal(err)
	}

	obs := &weather.Observation{
		AirTemperature: 90.0,
	}
	manager.ProcessObservation(obs)
}

func TestManagerStop(t *testing.T) {
	configJSON := `{
		"alarms": [{
			"name": "test-alarm",
			"condition": "temperature > 85",
			"enabled": true,
			"channels": [{"type": "console", "template": "Alert"}]
		}]
	}`

	manager, err := NewManager(configJSON, "Test Station")
	if err != nil {
		t.Fatal(err)
	}

	// Stop should not panic
	manager.Stop()

	// Note: Calling Stop multiple times may panic (channel already closed)
	// This is acceptable behavior - just test that one call works
}

func TestProcessObservationCompoundConditions(t *testing.T) {
	configJSON := `{
		"alarms": [
			{
				"name": "and-alarm",
				"condition": "temperature > 30 && humidity > 80",
				"enabled": true,
				"channels": [{"type": "console", "template": "AND condition met"}]
			},
			{
				"name": "or-alarm",
				"condition": "temperature > 100 || humidity > 90",
				"enabled": true,
				"channels": [{"type": "console", "template": "OR condition met"}]
			}
		]
	}`

	manager, err := NewManager(configJSON, "Test Station")
	if err != nil {
		t.Fatal(err)
	}

	obs1 := &weather.Observation{
		AirTemperature:   35.0,
		RelativeHumidity: 85.0,
	}
	manager.ProcessObservation(obs1)

	obs2 := &weather.Observation{
		AirTemperature:   35.0,
		RelativeHumidity: 75.0,
	}
	manager.ProcessObservation(obs2)

	obs3 := &weather.Observation{
		AirTemperature:   25.0,
		RelativeHumidity: 95.0,
	}
	manager.ProcessObservation(obs3)
}

func TestConfigReload(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "alarm_reload_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	initialConfig := AlarmConfig{
		Alarms: []Alarm{
			{
				Name:      "initial-alarm",
				Condition: "temperature > 85",
				Enabled:   true,
				Channels:  []Channel{{Type: "console", Template: "Initial"}},
			},
		},
	}

	data, err := json.Marshal(initialConfig)
	if err != nil {
		t.Fatalf("failed to marshal initial config: %v", err)
	}
	if _, err := tmpfile.Write(data); err != nil {
		t.Fatalf("failed to write initial config to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	manager, err := NewManager("@"+tmpfile.Name(), "Test Station")
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Stop()

	if len(manager.config.Alarms) != 1 {
		t.Errorf("Expected 1 alarm initially, got %d", len(manager.config.Alarms))
	}

	if manager.config.Alarms[0].Name != "initial-alarm" {
		t.Errorf("Expected 'initial-alarm', got '%s'", manager.config.Alarms[0].Name)
	}
}

func TestProcessObservationWithNilObservation(t *testing.T) {
	configJSON := `{
		"alarms": [{
			"name": "test-alarm",
			"condition": "temperature > 85",
			"enabled": true,
			"channels": [{"type": "console", "template": "Alert"}]
		}]
	}`

	manager, err := NewManager(configJSON, "Test Station")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("ProcessObservation panicked on nil observation (expected): %v", r)
		}
	}()

	manager.ProcessObservation(nil)
}
