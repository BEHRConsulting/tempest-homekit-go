# udp/ Package

The `udp` package provides UDP broadcast listener functionality for receiving real-time weather data directly from a Tempest weather station on the local network. This enables offline weather monitoring when internet connectivity is unavailable.

## Purpose

The primary use case for UDP monitoring is **offline operation**. When your internet connection goes down, the Tempest station continues broadcasting weather data over your local network via UDP on port 50222. This package allows the application to:

- Continue collecting weather data during internet outages
- Operate completely offline without requiring WeatherFlow API access
- Reduce API call overhead by using local data
- Provide lower latency updates (UDP broadcasts occur every minute or faster for rapid wind)

## Features

- **UDP Broadcast Reception**: Listens on port 50222 for Tempest hub broadcasts
- **Multiple Message Types**: Supports obs_st (Tempest observations), obs_air, obs_sky, rapid_wind, device_status, hub_status, lightning events, and rain events
- **Circular Buffer History**: Maintains up to 1000 observations in memory
- **Real-time Statistics**: Tracks packet count, station IP, serial number, and last packet time
- **Thread-safe**: All operations are protected by mutex locks for concurrent access
- **Auto-detection**: Automatically detects station IP and serial number from broadcasts

## Message Types Supported

### Observation Messages
- `obs_st`: Tempest device observations (all sensors in one message)
- `obs_air`: AIR device observations (temperature, pressure, humidity, lightning)
- `obs_sky`: SKY device observations (wind, rain, light, UV)
- `rapid_wind`: High-frequency wind updates (every 3 seconds)

### Event Messages
- `evt_precip`: Rain start events
- `evt_strike`: Lightning strike events with distance and energy

### Status Messages
- `device_status`: Device health, battery, RSSI, sensor status
- `hub_status`: Hub firmware, uptime, network status

## Usage

### Basic Usage

```go
import "tempest-homekit-go/pkg/udp"

// Create a new UDP listener
listener := udp.NewUDPListener()

// Start listening for broadcasts
if err := listener.Start(); err != nil {
    log.Fatal(err)
}
defer listener.Stop()

// Subscribe to observations
obsChan := listener.ObservationChannel()
for obs := range obsChan {
    fmt.Printf("Temperature: %.1f°C\n", obs.AirTemperature)
    fmt.Printf("Humidity: %.0f%%\n", obs.RelativeHumidity)
    // Process observation...
}
```

### Get Latest Data

```go
// Get most recent observation
if obs := listener.GetLatestObservation(); obs != nil {
    fmt.Printf("Latest temperature: %.1f°C\n", obs.AirTemperature)
}

// Get all historical observations
observations := listener.GetObservations()
fmt.Printf("Have %d observations in history\n", len(observations))
```

### Check Connection Status

```go
// Check if receiving data
if listener.IsReceivingData() {
    fmt.Println("Receiving UDP broadcasts")
} else {
    fmt.Println("No UDP packets detected")
}

// Get statistics
packetCount, lastTime, stationIP, serial := listener.GetStats()
fmt.Printf("Packets: %d, IP: %s, Serial: %s\n", packetCount, stationIP, serial)
```

### Device Status

```go
// Get device status (battery, RSSI, etc.)
if status := listener.GetDeviceStatus(); status != nil {
    fmt.Printf("Battery: %.2fV\n", status.Voltage)
    fmt.Printf("Signal: %ddBm\n", status.RSSI)
}

// Get hub status
if status := listener.GetHubStatus(); status != nil {
    fmt.Printf("Hub firmware: %s\n", status.FirmwareRev)
    fmt.Printf("Hub uptime: %ds\n", status.Uptime)
}
```

## Configuration Flags

The UDP listener is activated through command-line flags:

```bash
# Use UDP stream for offline operation
./tempest-homekit-go --udp-stream

# Use UDP with custom station URL
./tempest-homekit-go --station-url http://localhost:8080/api/udp-stream

# Disable all internet access (no API calls, no status scraping)
./tempest-homekit-go --udp-stream --no-internet
```

## Network Requirements

- **Port**: UDP port 50222 must be accessible
- **Network**: Application must be on same local network as Tempest hub
- **Firewall**: Ensure firewall allows UDP traffic on port 50222
- **No Internet Required**: Can operate completely offline

## Data Format

UDP messages are JSON-formatted. Example Tempest observation:

```json
{
  "serial_number": "ST-00000512",
  "type": "obs_st",
  "hub_sn": "HB-00013030",
  "obs": [[1588948614,0.18,0.22,0.27,144,6,1017.57,22.37,50.26,328,0.03,3,0.000000,0,0,0,2.410,1]],
  "firmware_revision": 129
}
```

The `obs` array contains:
- [0]: timestamp (Unix epoch)
- [1]: wind lull (m/s)
- [2]: wind average (m/s)
- [3]: wind gust (m/s)
- [4]: wind direction (degrees)
- [5]: wind sample interval (seconds)
- [6]: station pressure (mb)
- [7]: air temperature (°C)
- [8]: relative humidity (%)
- [9]: illuminance (lux)
- [10]: UV index
- [11]: solar radiation (W/m²)
- [12]: rain accumulated (mm)
- [13]: precipitation type (0=none, 1=rain, 2=hail)
- [14]: lightning average distance (km)
- [15]: lightning strike count
- [16]: battery (volts)
- [17]: report interval (minutes)

## Implementation Details

### Thread Safety
All public methods are thread-safe using `sync.RWMutex` for concurrent access from multiple goroutines.

### Circular Buffer
Observations are stored in a circular buffer with a maximum size of 1000 entries. When the buffer is full, the oldest observation is removed to make room for new data.

### Channel-based Updates
New observations are sent to a buffered channel (capacity 100) for real-time processing. If the channel is full, the observation is added to history but the channel send is skipped (non-blocking).

### Timeout Handling
The UDP listener uses 1-second read timeouts to allow periodic checking of the stop channel while still being responsive to incoming packets.

## Troubleshooting

### No UDP Packets Detected
- Verify Tempest station is powered on and connected to network
- Check that application is on same network as station
- Ensure firewall allows UDP on port 50222
- Confirm hub firmware version supports UDP broadcasts (v17+)

### Intermittent Data
- Normal behavior - Tempest sends observations every 1 minute
- Rapid wind updates occur every 3 seconds
- Check device status for sensor failures

### Old Data
- Check `lastPacketTime` - should be within last 2-3 minutes
- If old, station may be offline or network issue
- Device may be in low-power mode during night

## WeatherFlow Documentation

For complete UDP protocol documentation, see:
https://apidocs.tempestwx.com/reference/tempest-udp-broadcast
