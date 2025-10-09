package udp

import (
	"encoding/json"
	"testing"
	"time"

	"tempest-homekit-go/pkg/weather"
)

func TestNewUDPListenerAndProcessObservationST(t *testing.T) {
	l := NewUDPListener(50)
	if l == nil {
		t.Fatal("NewUDPListener returned nil")
	}

	// Build a minimal obs_st message with required 18 items
	obs := make([]interface{}, 18)
	obs[0] = float64(time.Now().Unix()) // timestamp
	obs[1] = 0.1                        // wind_lull
	obs[2] = 1.2                        // wind_avg
	obs[3] = 2.3                        // wind_gust
	obs[4] = 180.0                      // wind_dir
	obs[6] = 1013.25                    // pressure
	obs[7] = 20.0                       // temp
	obs[8] = 50.0                       // humidity
	obs[9] = 100.0                      // lux
	obs[10] = 2.0                       // uv
	obs[11] = 300.0                     // solar_rad
	obs[12] = 0.0                       // rain
	obs[13] = 0.0                       // precip_type
	obs[14] = 0.0                       // lightning dist
	obs[15] = 0.0                       // lightning count
	obs[16] = 3.7                       // battery
	obs[17] = 60.0                      // report interval

	msg := UDPMessage{
		SerialNumber: "SN-TEST",
		Type:         TypeObservationST,
		Obs:          [][]interface{}{obs},
	}

	data, _ := json.Marshal(msg)

	// Process message directly
	l.processMessage(data)

	// Validate latest observation
	latest := l.GetLatestObservation()
	if latest == nil {
		t.Fatal("expected latest observation after processing message")
	}
	if latest.AirTemperature != 20.0 {
		t.Fatalf("unexpected temperature: %v", latest.AirTemperature)
	}

	// Observations list should include this entry
	obsList := l.GetObservations()
	if len(obsList) == 0 {
		t.Fatalf("expected non-empty observations list")
	}

	// Stats should reflect zero packets (we didn't run the listen loop) but device serial should be set
	_, _, _, sn := l.GetStats()
	if sn == "" && l.serialNumber == "" {
		t.Fatalf("expected serial number to be set by processing, got empty")
	}

	// Device and hub status initially nil
	if l.GetDeviceStatus() != nil {
		t.Fatalf("expected nil device status")
	}
	if l.GetHubStatus() != nil {
		t.Fatalf("expected nil hub status")
	}

	// Ensure IsReceivingData false because lastPacketTime is zero
	if l.IsReceivingData() {
		t.Fatalf("expected IsReceivingData to be false")
	}

	// Add observation directly via addObservation to test buffering
	l.addObservation(weather.Observation{AirTemperature: 21.0, Timestamp: time.Now().Unix()})
	if l.GetLatestObservation().AirTemperature != 21.0 {
		t.Fatalf("expected latest temperature 21.0, got %v", l.GetLatestObservation().AirTemperature)
	}
}
