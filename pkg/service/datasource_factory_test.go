package service

import (
	"testing"
	"time"

	"tempest-homekit-go/pkg/config"
	"tempest-homekit-go/pkg/weather"
)

// mockUDPListener implements the weather.UDPListener interface for testing
type mockUDPListener struct {
	started      bool
	stopped      bool
	observations []weather.Observation
	obsChan      chan weather.Observation
	stopChan     chan struct{}
	packetCount  int64
	lastPacket   time.Time
	stationIP    string
	serialNumber string
}

func newMockUDPListener() *mockUDPListener {
	return &mockUDPListener{
		observations: []weather.Observation{},
		obsChan:      make(chan weather.Observation, 10),
		stopChan:     make(chan struct{}),
		stationIP:    "192.168.1.100",
		serialNumber: "ST-00000001",
	}
}

func (m *mockUDPListener) Start() error {
	m.started = true
	return nil
}

func (m *mockUDPListener) Stop() error {
	m.stopped = true
	if m.stopChan != nil {
		close(m.stopChan)
	}
	return nil
}

func (m *mockUDPListener) GetLatestObservation() *weather.Observation {
	if len(m.observations) == 0 {
		return nil
	}
	return &m.observations[len(m.observations)-1]
}

func (m *mockUDPListener) GetStats() (packetCount int64, lastPacket time.Time, stationIP, serialNumber string) {
	return m.packetCount, m.lastPacket, m.stationIP, m.serialNumber
}

func (m *mockUDPListener) GetObservations() []weather.Observation {
	result := make([]weather.Observation, len(m.observations))
	copy(result, m.observations)
	return result
}

func (m *mockUDPListener) ObservationChannel() <-chan weather.Observation {
	return m.obsChan
}

func (m *mockUDPListener) IsReceivingData() bool {
	return len(m.observations) > 0
}

// mockStation creates a test station
func mockStation() *weather.Station {
	return &weather.Station{
		StationID:   12345,
		Name:        "Test Station",
		StationName: "Test Station",
	}
}

func TestCreateDataSource_UDPMode(t *testing.T) {
	cfg := &config.Config{
		UDPStream:       true,
		DisableInternet: true,
		StationName:     "Test Station",
		Token:           "test-token",
	}

	station := mockStation()
	udpListener := newMockUDPListener()

	dataSource, err := CreateDataSource(cfg, station, udpListener)
	if err != nil {
		t.Fatalf("Failed to create UDP data source: %v", err)
	}

	// Verify it's the correct type
	if dataSource.GetType() != weather.DataSourceUDP {
		t.Errorf("Expected UDP data source, got %s", dataSource.GetType())
	}

	// Verify status
	status := dataSource.GetStatus()
	if status.Type != weather.DataSourceUDP {
		t.Errorf("Expected status type UDP, got %s", status.Type)
	}
	// UDP data source gets station info from broadcasts, not initial config
	if !status.Active {
		t.Log("Note: Status not yet active (no data received, expected in test)")
	}
}

func TestCreateDataSource_CustomURL(t *testing.T) {
	cfg := &config.Config{
		StationURL:  "http://localhost:8080/weather",
		StationName: "Test Station",
		Token:       "test-token",
	}

	station := mockStation()

	dataSource, err := CreateDataSource(cfg, station, nil)
	if err != nil {
		t.Fatalf("Failed to create custom URL data source: %v", err)
	}

	// Custom URL uses API data source type
	if dataSource.GetType() != weather.DataSourceCustomURL {
		t.Errorf("Expected CustomURL data source, got %s", dataSource.GetType())
	}

	// Verify status
	status := dataSource.GetStatus()
	if status.Type != weather.DataSourceCustomURL {
		t.Errorf("Expected status type CustomURL, got %s", status.Type)
	}
	if status.CustomURL != "http://localhost:8080/weather" {
		t.Errorf("Expected URL in status, got %s", status.CustomURL)
	}
}

func TestCreateDataSource_GeneratedWeather(t *testing.T) {
	cfg := &config.Config{
		UseGeneratedWeather: true,
		StationName:         "Test Station",
		Token:               "test-token",
	}

	station := mockStation()

	dataSource, err := CreateDataSource(cfg, station, nil)
	if err != nil {
		t.Fatalf("Failed to create generated weather data source: %v", err)
	}

	// Generated weather uses API data source with localhost URL
	if dataSource.GetType() != weather.DataSourceGenerated {
		t.Errorf("Expected Generated data source, got %s", dataSource.GetType())
	}

	// Verify status
	status := dataSource.GetStatus()
	if status.Type != weather.DataSourceGenerated {
		t.Errorf("Expected status type Generated, got %s", status.Type)
	}
}

func TestCreateDataSource_RealAPI(t *testing.T) {
	cfg := &config.Config{
		StationName: "Test Station",
		Token:       "test-token",
	}

	station := mockStation()

	dataSource, err := CreateDataSource(cfg, station, nil)
	if err != nil {
		t.Fatalf("Failed to create API data source: %v", err)
	}

	// Should default to API
	if dataSource.GetType() != weather.DataSourceAPI {
		t.Errorf("Expected API data source, got %s", dataSource.GetType())
	}

	// Verify status
	status := dataSource.GetStatus()
	if status.Type != weather.DataSourceAPI {
		t.Errorf("Expected status type API, got %s", status.Type)
	}
	if status.StationName != station.Name {
		t.Errorf("Expected station name %s, got %s", station.Name, status.StationName)
	}
}

func TestCreateDataSource_Priority(t *testing.T) {
	// Test that UDP takes priority over everything
	cfg := &config.Config{
		UDPStream:           true,
		UseGeneratedWeather: true,
		StationURL:          "http://custom.url",
		StationName:         "Test Station",
		Token:               "test-token",
	}

	station := mockStation()
	udpListener := newMockUDPListener()

	dataSource, err := CreateDataSource(cfg, station, udpListener)
	if err != nil {
		t.Fatalf("Failed to create data source: %v", err)
	}

	if dataSource.GetType() != weather.DataSourceUDP {
		t.Errorf("Expected UDP to take priority, got %s", dataSource.GetType())
	}
}

func TestCreateDataSource_CustomURLPriorityOverGenerated(t *testing.T) {
	// Test that Custom URL takes priority over Generated
	cfg := &config.Config{
		StationURL:          "http://custom.url",
		UseGeneratedWeather: true,
		StationName:         "Test Station",
		Token:               "test-token",
	}

	station := mockStation()

	dataSource, err := CreateDataSource(cfg, station, nil)
	if err != nil {
		t.Fatalf("Failed to create data source: %v", err)
	}

	if dataSource.GetType() != weather.DataSourceCustomURL {
		t.Errorf("Expected CustomURL to take priority over Generated, got %s", dataSource.GetType())
	}
}

func TestCreateDataSource_UDPWithoutListener(t *testing.T) {
	cfg := &config.Config{
		UDPStream:   true,
		StationName: "Test Station",
		Token:       "test-token",
	}

	station := mockStation()

	_, err := CreateDataSource(cfg, station, nil)
	if err == nil {
		t.Error("Expected error when UDP mode enabled but no listener provided")
	}
}

func TestCreateDataSource_InvalidUDPListenerType(t *testing.T) {
	cfg := &config.Config{
		UDPStream:   true,
		StationName: "Test Station",
		Token:       "test-token",
	}

	station := mockStation()
	invalidListener := "not a listener" // Wrong type

	_, err := CreateDataSource(cfg, station, invalidListener)
	if err == nil {
		t.Error("Expected error when invalid listener type provided")
	}
}

func TestCreateDataSource_NilStation(t *testing.T) {
	cfg := &config.Config{
		StationName: "Test Station",
		Token:       "test-token",
	}

	_, err := CreateDataSource(cfg, nil, nil)
	if err == nil {
		t.Fatal("Expected error when creating API data source with nil station")
	}

	// API data source requires station info
	if err.Error() != "station required for API data source" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestDataSourceStatus_Structure(t *testing.T) {
	cfg := &config.Config{
		StationName: "Test Station",
		Token:       "test-token",
	}

	station := mockStation()

	dataSource, err := CreateDataSource(cfg, station, nil)
	if err != nil {
		t.Fatalf("Failed to create data source: %v", err)
	}

	status := dataSource.GetStatus()

	// Verify all required fields are present
	if status.Type == "" {
		t.Error("Status type is empty")
	}
	if status.StationName == "" {
		t.Error("Status station name is empty")
	}
	if status.LastUpdate.IsZero() {
		t.Log("Note: LastUpdate is zero (no data fetched yet, expected in test)")
	}
}

func TestDataSourceInterfaces(t *testing.T) {
	// Verify all data source types implement the interface correctly
	cfg := &config.Config{
		StationName: "Test Station",
		Token:       "test-token",
	}

	station := mockStation()

	dataSource, err := CreateDataSource(cfg, station, nil)
	if err != nil {
		t.Fatalf("Failed to create data source: %v", err)
	}

	// Test interface methods don't panic
	_ = dataSource.GetType()
	_ = dataSource.GetStatus()
	_ = dataSource.GetForecast() // May return nil initially

	// Start and stop should work
	obsChan, err := dataSource.Start()
	if err != nil {
		t.Fatalf("Failed to start data source: %v", err)
	}

	if obsChan == nil {
		t.Error("Expected non-nil observation channel")
	}

	// Give it a moment
	time.Sleep(100 * time.Millisecond)

	err = dataSource.Stop()
	if err != nil {
		t.Errorf("Failed to stop data source: %v", err)
	}
}

// TestCreateDataSource_StationURLConstruction ensures CreateDataSource constructs
// the generated StationURL the same way LoadConfig would (http://localhost:<port><path>). 
func TestCreateDataSource_StationURLConstruction(t *testing.T) {
	cfg := &config.Config{
		UseGeneratedWeather: true,
		WebPort:             "12345",
		GeneratedWeatherPath: "/api/custom-generate",
		StationName:         "Test Station",
		Token:               "test-token",
	}

	station := mockStation()

	ds, err := CreateDataSource(cfg, station, nil)
	if err != nil {
		t.Fatalf("CreateDataSource failed: %v", err)
	}

	status := ds.GetStatus()

	expected := "http://localhost:12345/api/custom-generate"
	if status.CustomURL != expected {
		t.Fatalf("expected StationURL %s, got %s", expected, status.CustomURL)
	}
}
