// Package udp provides UDP broadcast listener for local Tempest weather station data.
// This enables offline weather monitoring when internet connectivity is unavailable.
package udp

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/weather"
)

const (
	// UDPPort is the port that Tempest hubs broadcast on
	UDPPort = 50222
)

// MessageType represents the type of UDP broadcast message
type MessageType string

const (
	TypeObservationST  MessageType = "obs_st"     // Tempest observation
	TypeObservationAir MessageType = "obs_air"    // AIR device observation
	TypeObservationSky MessageType = "obs_sky"    // SKY device observation
	TypeRapidWind      MessageType = "rapid_wind" // Rapid wind update
	TypeRainStart      MessageType = "evt_precip" // Rain start event
	TypeLightning      MessageType = "evt_strike" // Lightning strike event
	TypeDeviceStatus   MessageType = "device_status"
	TypeHubStatus      MessageType = "hub_status"
)

// UDPMessage represents the generic structure of all UDP broadcast messages
type UDPMessage struct {
	SerialNumber     string          `json:"serial_number"`
	Type             MessageType     `json:"type"`
	HubSN            string          `json:"hub_sn"`
	FirmwareRevision int             `json:"firmware_revision,omitempty"`
	Obs              [][]interface{} `json:"obs,omitempty"`
	Ob               []interface{}   `json:"ob,omitempty"`  // For rapid_wind
	Evt              []interface{}   `json:"evt,omitempty"` // For events
	// Device status fields
	Timestamp    int64   `json:"timestamp,omitempty"`
	Uptime       int     `json:"uptime,omitempty"`
	Voltage      float64 `json:"voltage,omitempty"`
	RSSI         int     `json:"rssi,omitempty"`
	HubRSSI      int     `json:"hub_rssi,omitempty"`
	SensorStatus int     `json:"sensor_status,omitempty"`
	Debug        int     `json:"debug,omitempty"`
	// Hub status fields
	ResetFlags string        `json:"reset_flags,omitempty"`
	Seq        int           `json:"seq,omitempty"`
	RadioStats []interface{} `json:"radio_stats,omitempty"`
	MqttStats  []interface{} `json:"mqtt_stats,omitempty"`
}

// UDPListener listens for UDP broadcasts from a Tempest weather station
type UDPListener struct {
	conn            *net.UDPConn
	observations    []weather.Observation
	maxHistorySize  int
	latestObs       *weather.Observation
	mu              sync.RWMutex
	packetCount     int64
	lastPacketTime  time.Time
	stationIP       string
	serialNumber    string
	deviceStatus    *DeviceStatus
	hubStatus       *HubStatus
	observationChan chan weather.Observation
	stopChan        chan struct{}
	running         bool
}

// DeviceStatus holds device status information
type DeviceStatus struct {
	Timestamp    time.Time
	Uptime       int
	Voltage      float64
	RSSI         int
	HubRSSI      int
	SensorStatus int
	Debug        int
}

// HubStatus holds hub status information
type HubStatus struct {
	Timestamp   time.Time
	FirmwareRev string
	Uptime      int
	RSSI        int
	ResetFlags  string
	Seq         int
}

// NewUDPListener creates a new UDP listener
func NewUDPListener(maxHistorySize int) *UDPListener {
	if maxHistorySize < 10 {
		maxHistorySize = 1000 // default if invalid
	}

	// Validate history size to prevent excessive memory allocation
	if maxHistorySize > 100000 {
		logger.Warn("History size %d is very large, capping at 100000 to prevent memory issues", maxHistorySize)
		maxHistorySize = 100000
	}

	return &UDPListener{
		observations:    make([]weather.Observation, 0, maxHistorySize),
		maxHistorySize:  maxHistorySize,
		observationChan: make(chan weather.Observation, 100),
		stopChan:        make(chan struct{}),
	}
}

// Start begins listening for UDP broadcasts
func (l *UDPListener) Start() error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return fmt.Errorf("UDP listener already running")
	}
	l.running = true
	l.mu.Unlock()

	addr := net.UDPAddr{
		Port: UDPPort,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		l.mu.Lock()
		l.running = false
		l.mu.Unlock()
		return fmt.Errorf("failed to start UDP listener on port %d: %v", UDPPort, err)
	}

	l.conn = conn
	logger.Info("UDP listener started on port %d", UDPPort)

	// Start listening in a goroutine
	go l.listen()

	return nil
}

// Stop stops the UDP listener
func (l *UDPListener) Stop() error {
	l.mu.Lock()
	if !l.running {
		l.mu.Unlock()
		return nil
	}
	l.running = false
	l.mu.Unlock()

	close(l.stopChan)
	if l.conn != nil {
		return l.conn.Close()
	}
	return nil
}

// listen is the main listening loop
func (l *UDPListener) listen() {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-l.stopChan:
			logger.Info("UDP listener stopped")
			return
		default:
			// Set read deadline to allow checking stopChan periodically
			l.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, remoteAddr, err := l.conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Timeout is expected, continue
					continue
				}
				logger.Error("UDP read error: %v", err)
				continue
			}

			// Update packet statistics
			l.mu.Lock()
			l.packetCount++
			l.lastPacketTime = time.Now()
			if l.stationIP == "" && remoteAddr != nil {
				l.stationIP = remoteAddr.IP.String()
				logger.Info("Detected Tempest station at IP: %s", l.stationIP)
			}
			l.mu.Unlock()

			// Process the message
			l.processMessage(buffer[:n])
		}
	}
}

// processMessage parses and processes a UDP message
func (l *UDPListener) processMessage(data []byte) {
	// Log raw packet data when debug logging is enabled
	logger.Debug("UDP Packet received (%d bytes): %s", len(data), string(data))

	var msg UDPMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Debug("Failed to parse UDP message: %v", err)
		return
	}

	logger.Debug("Parsed UDP message - Type: %s, Serial: %s, Hub: %s", msg.Type, msg.SerialNumber, msg.HubSN)

	// Update serial number if not set
	if l.serialNumber == "" && msg.SerialNumber != "" {
		l.mu.Lock()
		l.serialNumber = msg.SerialNumber
		l.mu.Unlock()
		logger.Info("Detected Tempest device serial: %s", msg.SerialNumber)
	}

	switch msg.Type {
	case TypeObservationST:
		l.processObservationST(msg)
	case TypeObservationAir:
		l.processObservationAir(msg)
	case TypeObservationSky:
		l.processObservationSky(msg)
	case TypeRapidWind:
		l.processRapidWind(msg)
	case TypeDeviceStatus:
		l.processDeviceStatus(msg)
	case TypeHubStatus:
		l.processHubStatus(msg)
	case TypeRainStart:
		timestamp := int64(msg.Evt[0].(float64))
		logger.Debug("UDP evt_precip - Rain start event at timestamp=%d (%v)", timestamp, time.Unix(timestamp, 0))
	case TypeLightning:
		if len(msg.Evt) >= 3 {
			timestamp := int64(msg.Evt[0].(float64))
			distance := msg.Evt[1].(float64)
			energy := msg.Evt[2].(float64)
			logger.Debug("UDP evt_strike - Lightning strike at timestamp=%d, distance=%.1fkm, energy=%.0f", timestamp, distance, energy)
		}
	default:
		logger.Debug("Unknown UDP message type: %s", msg.Type)
	}
}

// processObservationST processes a Tempest (ST) observation message
func (l *UDPListener) processObservationST(msg UDPMessage) {
	if len(msg.Obs) == 0 || len(msg.Obs[0]) < 18 {
		logger.Debug("Invalid Tempest observation data")
		return
	}

	obs := msg.Obs[0]

	// Parse observation according to Tempest UDP format
	// [0]=timestamp, [1]=wind_lull, [2]=wind_avg, [3]=wind_gust, [4]=wind_dir,
	// [5]=wind_sample_interval, [6]=pressure, [7]=temp, [8]=humidity, [9]=lux,
	// [10]=uv, [11]=solar_rad, [12]=rain_1min, [13]=precip_type, [14]=lightning_dist,
	// [15]=lightning_count, [16]=battery, [17]=report_interval

	observation := weather.Observation{
		Timestamp:            int64(obs[0].(float64)),
		WindLull:             obs[1].(float64),
		WindAvg:              obs[2].(float64),
		WindGust:             obs[3].(float64),
		WindDirection:        obs[4].(float64),
		StationPressure:      obs[6].(float64),
		AirTemperature:       obs[7].(float64),
		RelativeHumidity:     obs[8].(float64),
		Illuminance:          obs[9].(float64),
		UV:                   int(obs[10].(float64)),
		SolarRadiation:       obs[11].(float64),
		RainAccumulated:      obs[12].(float64),
		PrecipitationType:    int(obs[13].(float64)),
		LightningStrikeAvg:   obs[14].(float64),
		LightningStrikeCount: int(obs[15].(float64)),
		Battery:              obs[16].(float64),
		ReportInterval:       int(obs[17].(float64)),
	}

	logger.Debug("UDP obs_st - Timestamp=%d, Temp=%.1f°C, Humidity=%.0f%%, Pressure=%.1fmb, Wind=%.1f/%.1f/%.1fm/s@%.0f°, Lux=%.0f, UV=%d, Rain=%.2fmm, Lightning=%d@%.0fkm, Battery=%.2fV",
		observation.Timestamp, observation.AirTemperature, observation.RelativeHumidity, observation.StationPressure,
		observation.WindLull, observation.WindAvg, observation.WindGust, observation.WindDirection,
		observation.Illuminance, observation.UV, observation.RainAccumulated,
		observation.LightningStrikeCount, observation.LightningStrikeAvg, observation.Battery)

	l.addObservation(observation)
}

// processObservationAir processes an AIR device observation
func (l *UDPListener) processObservationAir(msg UDPMessage) {
	if len(msg.Obs) == 0 || len(msg.Obs[0]) < 8 {
		return
	}

	obs := msg.Obs[0]
	// AIR format: [0]=timestamp, [1]=pressure, [2]=temp, [3]=humidity,
	// [4]=lightning_count, [5]=lightning_dist, [6]=battery, [7]=report_interval

	observation := weather.Observation{
		Timestamp:            int64(obs[0].(float64)),
		StationPressure:      obs[1].(float64),
		AirTemperature:       obs[2].(float64),
		RelativeHumidity:     obs[3].(float64),
		LightningStrikeCount: int(obs[4].(float64)),
		LightningStrikeAvg:   obs[5].(float64),
		Battery:              obs[6].(float64),
		ReportInterval:       int(obs[7].(float64)),
	}

	logger.Debug("UDP obs_air - Timestamp=%d, Temp=%.1f°C, Humidity=%.0f%%, Pressure=%.1fmb, Lightning=%d@%.0fkm, Battery=%.2fV",
		observation.Timestamp, observation.AirTemperature, observation.RelativeHumidity,
		observation.StationPressure, observation.LightningStrikeCount, observation.LightningStrikeAvg, observation.Battery)

	l.addObservation(observation)
}

// processObservationSky processes a SKY device observation
func (l *UDPListener) processObservationSky(msg UDPMessage) {
	if len(msg.Obs) == 0 || len(msg.Obs[0]) < 14 {
		return
	}

	obs := msg.Obs[0]
	// SKY format: [0]=timestamp, [1]=lux, [2]=uv, [3]=rain_1min, [4]=wind_lull,
	// [5]=wind_avg, [6]=wind_gust, [7]=wind_dir, [8]=battery, [9]=report_interval,
	// [10]=solar_rad, [11]=local_rain (null), [12]=precip_type, [13]=wind_sample_interval

	observation := weather.Observation{
		Timestamp:         int64(obs[0].(float64)),
		Illuminance:       obs[1].(float64),
		UV:                int(obs[2].(float64)),
		RainAccumulated:   obs[3].(float64),
		WindLull:          obs[4].(float64),
		WindAvg:           obs[5].(float64),
		WindGust:          obs[6].(float64),
		WindDirection:     obs[7].(float64),
		Battery:           obs[8].(float64),
		ReportInterval:    int(obs[9].(float64)),
		SolarRadiation:    obs[10].(float64),
		PrecipitationType: int(obs[12].(float64)),
	}

	logger.Debug("UDP obs_sky - Timestamp=%d, Wind=%.1f/%.1f/%.1fm/s@%.0f°, Lux=%.0f, UV=%d, Solar=%.0fW/m², Rain=%.2fmm, Battery=%.2fV",
		observation.Timestamp, observation.WindLull, observation.WindAvg, observation.WindGust, observation.WindDirection,
		observation.Illuminance, observation.UV, observation.SolarRadiation, observation.RainAccumulated, observation.Battery)

	l.addObservation(observation)
}

// processRapidWind processes rapid wind updates (every 3 seconds)
func (l *UDPListener) processRapidWind(msg UDPMessage) {
	if len(msg.Ob) < 3 {
		return
	}

	// Rapid wind: [0]=timestamp, [1]=wind_speed, [2]=wind_direction
	timestamp := int64(msg.Ob[0].(float64))
	windSpeed := msg.Ob[1].(float64)
	windDir := int(msg.Ob[2].(float64))
	logger.Debug("UDP rapid_wind - Timestamp=%d, Speed=%.1fm/s, Direction=%d°", timestamp, windSpeed, windDir)

	// We could update wind in real-time here, but for now just log it
	// The full observation will be processed when obs_st arrives
}

// processDeviceStatus processes device status messages
func (l *UDPListener) processDeviceStatus(msg UDPMessage) {
	status := &DeviceStatus{
		Timestamp:    time.Unix(msg.Timestamp, 0),
		Uptime:       msg.Uptime,
		Voltage:      msg.Voltage,
		RSSI:         msg.RSSI,
		HubRSSI:      msg.HubRSSI,
		SensorStatus: msg.SensorStatus,
		Debug:        msg.Debug,
	}

	l.mu.Lock()
	l.deviceStatus = status
	l.mu.Unlock()

	logger.Debug("UDP device_status - Serial=%s, Timestamp=%d, Battery=%.2fV, Uptime=%ds, RSSI=%ddBm, Hub RSSI=%ddBm, Sensor Status=0x%X",
		msg.SerialNumber, msg.Timestamp, status.Voltage, status.Uptime, status.RSSI, status.HubRSSI, status.SensorStatus)
}

// processHubStatus processes hub status messages
func (l *UDPListener) processHubStatus(msg UDPMessage) {
	status := &HubStatus{
		Timestamp:   time.Unix(msg.Timestamp, 0),
		FirmwareRev: fmt.Sprintf("%d", msg.FirmwareRevision),
		Uptime:      msg.Uptime,
		RSSI:        msg.RSSI,
		ResetFlags:  msg.ResetFlags,
		Seq:         msg.Seq,
	}

	l.mu.Lock()
	l.hubStatus = status
	l.mu.Unlock()

	logger.Debug("UDP hub_status - Serial=%s, Timestamp=%d, Firmware=%s, Uptime=%ds, RSSI=%ddBm, ResetFlags=%s, Seq=%d",
		msg.SerialNumber, msg.Timestamp, status.FirmwareRev, status.Uptime, status.RSSI, status.ResetFlags, status.Seq)
}

// addObservation adds an observation to the history and notifies observers
func (l *UDPListener) addObservation(obs weather.Observation) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Add to history (circular buffer)
	if len(l.observations) >= l.maxHistorySize {
		// Remove oldest observation
		l.observations = l.observations[1:]
	}
	l.observations = append(l.observations, obs)

	// Update latest
	l.latestObs = &obs

	// Send to channel (non-blocking)
	select {
	case l.observationChan <- obs:
	default:
		logger.Debug("Observation channel full, skipping")
	}
}

// GetLatestObservation returns the most recent observation
func (l *UDPListener) GetLatestObservation() *weather.Observation {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.latestObs == nil {
		return nil
	}
	obs := *l.latestObs
	return &obs
}

// GetObservations returns all stored observations
func (l *UDPListener) GetObservations() []weather.Observation {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]weather.Observation, len(l.observations))
	copy(result, l.observations)
	return result
}

// GetStats returns UDP listener statistics
func (l *UDPListener) GetStats() (packetCount int64, lastPacket time.Time, stationIP, serialNumber string) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.packetCount, l.lastPacketTime, l.stationIP, l.serialNumber
}

// GetDeviceStatus returns the latest device status
func (l *UDPListener) GetDeviceStatus() *DeviceStatus {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.deviceStatus == nil {
		return nil
	}
	status := *l.deviceStatus
	return &status
}

// GetHubStatus returns the latest hub status
func (l *UDPListener) GetHubStatus() *HubStatus {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.hubStatus == nil {
		return nil
	}
	status := *l.hubStatus
	return &status
}

// ObservationChannel returns the channel for receiving new observations
func (l *UDPListener) ObservationChannel() <-chan weather.Observation {
	return l.observationChan
}

// IsReceivingData returns true if we've received data recently (within last 5 minutes)
func (l *UDPListener) IsReceivingData() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.lastPacketTime.IsZero() {
		return false
	}

	return time.Since(l.lastPacketTime) < 5*time.Minute
}
