package weather

import (
	"testing"
	"time"
)

func TestFindStationByName(t *testing.T) {
	stations := []Station{
		{StationID: 1, Name: "station1", StationName: "Station 1"},
		{StationID: 2, Name: "tempest-homekit", StationName: "Tempest HomeKit"},
	}

	station := FindStationByName(stations, "tempest-homekit")
	if station == nil {
		t.Fatal("Station not found")
	}
	if station.StationID != 2 {
		t.Errorf("Expected ID 2, got %d", station.StationID)
	}
}

func TestFindStationByNameNotFound(t *testing.T) {
	stations := []Station{
		{StationID: 1, Name: "station1"},
	}

	station := FindStationByName(stations, "notfound")
	if station != nil {
		t.Error("Expected nil, got station")
	}
}

func TestFindStationByStationName(t *testing.T) {
	stations := []Station{
		{StationID: 1, Name: "station1", StationName: "Station 1"},
		{StationID: 2, Name: "tempest-homekit", StationName: "Tempest HomeKit"},
	}

	// Should find by StationName field
	station := FindStationByName(stations, "Station 1")
	if station == nil {
		t.Fatal("Station not found by StationName")
	}
	if station.StationID != 1 {
		t.Errorf("Expected ID 1, got %d", station.StationID)
	}
}

func TestGetTempestDeviceID(t *testing.T) {
	station := &Station{
		StationID: 123,
		Devices: []Device{
			{DeviceID: 1001, DeviceType: "HB"},
			{DeviceID: 1002, DeviceType: "ST"}, // Tempest device
			{DeviceID: 1003, DeviceType: "SK"},
		},
	}

	deviceID, err := GetTempestDeviceID(station)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if deviceID != 1002 {
		t.Errorf("Expected device ID 1002, got %d", deviceID)
	}
}

func TestGetTempestDeviceIDNotFound(t *testing.T) {
	station := &Station{
		StationID: 123,
		Devices: []Device{
			{DeviceID: 1001, DeviceType: "HB"},
			{DeviceID: 1003, DeviceType: "SK"},
		},
	}

	_, err := GetTempestDeviceID(station)
	if err == nil {
		t.Error("Expected error when Tempest device not found")
	}
}

func TestGetTempestDeviceIDEmptyDevices(t *testing.T) {
	station := &Station{
		StationID: 123,
		Devices:   []Device{},
	}

	_, err := GetTempestDeviceID(station)
	if err == nil {
		t.Error("Expected error when no devices")
	}
}

func TestGetFloat64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
	}{
		{float64(123.45), 123.45},
		{nil, 0.0},      // Should return 0 for nil
		{"123.45", 0.0}, // Should return 0 for string
		{int(123), 0.0}, // Should return 0 for int (JSON only uses float64)
		{true, 0.0},     // Should return 0 for bool
	}

	for _, test := range tests {
		result := getFloat64(test.input)
		if result != test.expected {
			t.Errorf("getFloat64(%v) = %f, expected %f", test.input, result, test.expected)
		}
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected int
	}{
		{float64(123.0), 123},
		{float64(123.9), 123}, // Should truncate
		{nil, 0},              // Should return 0 for nil
		{"123", 0},            // Should return 0 for string
		{int(123), 0},         // Should return 0 for int (JSON only uses float64)
		{true, 0},             // Should return 0 for bool
	}

	for _, test := range tests {
		result := getInt(test.input)
		if result != test.expected {
			t.Errorf("getInt(%v) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestFilterToOneMinuteIncrements(t *testing.T) {
	baseTime := time.Now()
	observations := []*Observation{
		{Timestamp: baseTime.Unix()},                        // 0 seconds
		{Timestamp: baseTime.Add(10 * time.Second).Unix()},  // 10 seconds
		{Timestamp: baseTime.Add(30 * time.Second).Unix()},  // 30 seconds
		{Timestamp: baseTime.Add(60 * time.Second).Unix()},  // 1 minute
		{Timestamp: baseTime.Add(70 * time.Second).Unix()},  // 1 minute 10 seconds
		{Timestamp: baseTime.Add(120 * time.Second).Unix()}, // 2 minutes
		{Timestamp: baseTime.Add(180 * time.Second).Unix()}, // 3 minutes
	}

	filtered := filterToOneMinuteIncrements(observations, 1000)

	// Should keep observations at roughly 1-minute intervals
	if len(filtered) < 3 {
		t.Errorf("Expected at least 3 filtered observations, got %d", len(filtered))
	}

	// After filtering and sorting, the first observation should be one of the earlier ones
	// (the function reverses to chronological order)
	if len(filtered) > 0 {
		if filtered[0].Timestamp > baseTime.Add(1*time.Minute).Unix() {
			t.Error("Expected filtered observations to be in chronological order with earlier timestamps first")
		}
	}
}

func TestFilterToOneMinuteIncrementsMaxCount(t *testing.T) {
	baseTime := time.Now()
	var observations []*Observation

	// Create 1000 observations
	for i := 0; i < 1000; i++ {
		observations = append(observations, &Observation{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute).Unix(),
		})
	}

	maxCount := 100
	filtered := filterToOneMinuteIncrements(observations, maxCount)

	if len(filtered) > maxCount {
		t.Errorf("Expected max %d observations, got %d", maxCount, len(filtered))
	}
}
