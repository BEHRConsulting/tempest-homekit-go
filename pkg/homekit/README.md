# homekit/ Package

The `homekit` package handles Apple HomeKit integration for the Tempest HomeKit Go application. It creates and manages HomeKit accessories for weather sensors using the modern `brutella/hap` library.

## Files

### `modern_setup.go`
**HomeKit Bridge and Accessory Setup**

**Core Functions:**
- `SetupHomeKit(cfg *config.Config) (*HomeKitService, error)` - Initializes HomeKit bridge and accessories
- `UpdateAllSensors(data *weather.Observation)` - Updates all sensor values with weather data
- `StartHomeKitServer(ctx context.Context)` - Starts the HomeKit server with graceful shutdown
- `StopHomeKitServer()` - Gracefully stops the HomeKit server

**HomeKit Architecture:**
```go
type HomeKitService struct {
    Bridge      *accessory.Bridge
    Server      *hap.Server
    Accessories []hap.Accessory
    // Individual sensor accessories
    Temperature  *accessory.TemperatureSensor
    WindSpeed    *WeatherAccessory
    WindGust     *WeatherAccessory
    WindDirection *WeatherAccessory
    Humidity     *WeatherAccessory
    Rain         *WeatherAccessory
    UV           *WeatherAccessory
    Lightning    *WeatherAccessory
    Light        *WeatherAccessory
}
```

**Key Features:**
- **Modern hap Library**: Uses `github.com/brutella/hap` v0.0.32 for HomeKit protocol
- **Bridge Architecture**: Single bridge managing multiple weather sensor accessories
- **Standard Temperature Sensor**: Uses HomeKit's native temperature sensor for air temperature
- **Custom Weather Sensors**: 10 custom accessories with unique service UUIDs to prevent unit conversion issues
- **Graceful Lifecycle**: Context-based server management with proper startup/shutdown

### `custom_characteristics.go`
**Custom Weather Sensor Definitions**

**Custom Service UUIDs:**
- Wind Speed: `F101-0001-1000-8000-0026BB765291`
- Wind Gust: `F111-0001-1000-8000-0026BB765291`
- Wind Direction: `F121-0001-1000-8000-0026BB765291`
- Rain: `F131-0001-1000-8000-0026BB765291`
- UV Index: `F141-0001-1000-8000-0026BB765291`
- Lightning Count: `F151-0001-1000-8000-0026BB765291`
- Lightning Distance: `F161-0001-1000-8000-0026BB765291`
- Precipitation Type: `F171-0001-1000-8000-0026BB765291`
- Humidity: `F181-0001-1000-8000-0026BB765291`
- Ambient Light: `F191-0001-1000-8000-0026BB765291`

**Custom Characteristics:**
- All weather sensors use custom float characteristics
- Prevents HomeKit's automatic temperature unit conversion
- Maintains accurate weather data display in HomeKit apps
- Supports proper value ranges and units for each sensor type

## HomeKit Accessories

### Standard HomeKit Sensor
1. **Temperature Sensor** - Uses native HomeKit temperature characteristic
   - Displays in Celsius (HomeKit standard)
   - Compatible with all HomeKit apps
   - Automatic temperature conversion by iOS

### Custom Weather Sensors (10 Accessories)
1. **Wind Speed** - Average wind speed (mph/kph)
2. **Wind Gust** - Peak wind gust speed (mph/kph)
3. **Wind Direction** - Wind direction (0-360 degrees)
4. **Rain** - Rain accumulation (inches/mm)
5. **UV Index** - UV exposure index (0-11+)
6. **Lightning Count** - Lightning strike count
7. **Lightning Distance** - Lightning strike distance (km/miles)
8. **Precipitation Type** - Type of precipitation (integer code)
9. **Humidity** - Relative humidity (percentage)
10. **Ambient Light** - Light level (lux)

## Usage Examples

### Initialize HomeKit Service
```go
import "tempest-homekit-go/pkg/homekit"
import "tempest-homekit-go/pkg/config"

cfg := config.LoadConfig()
hkService, err := homekit.SetupHomeKit(cfg)
if err != nil {
    log.Fatal("Failed to setup HomeKit:", err)
}

// Start the HomeKit server
ctx := context.Background()
go hkService.StartHomeKitServer(ctx)
```

### Update Weather Data
```go
// Update all sensors with new weather data
err := hkService.UpdateAllSensors(weatherData)
if err != nil {
    log.Printf("Failed to update sensors: %v", err)
}
```

### Graceful Shutdown
```go
// Stop the HomeKit server gracefully
hkService.StopHomeKitServer()
```

## HomeKit Pairing Process

1. **Start Application**: Run with WeatherFlow API token
2. **Find Bridge**: Open iOS Home app → Add Accessory
3. **Manual Entry**: Select "Don't have a code or can't scan?"
4. **Select Bridge**: Choose "Tempest HomeKit Bridge"
5. **Enter PIN**: Use configured PIN (default: 00102003)
6. **Complete Setup**: All 11 sensors will appear as individual accessories

## Database Management

### HomeKit Database Location
- **Path**: `./db/` directory in application root
- **Contents**: Pairing information, accessory cache, encryption keys
- **Persistence**: Maintains pairing across application restarts

### Database Reset
```bash
# Using built-in command
./tempest-homekit-go --cleardb

# Or manual deletion
rm -rf ./db/
```

## Custom Service Architecture

### Why Custom Services?
HomeKit automatically converts temperature values between Celsius and Fahrenheit, which causes issues for weather sensors that report other types of data. Custom services with unique UUIDs prevent this automatic conversion.

### Service UUID Pattern
- **Base UUID**: `F1XX-0001-1000-8000-0026BB765291`
- **XX Values**: Unique identifier for each sensor type (01, 11, 21, etc.)
- **Characteristic**: Custom float characteristic for each service
- **Benefits**: No automatic unit conversion, accurate weather data display

## Troubleshooting

### Common Issues
- **Pairing Fails**: Ensure PIN is correct and no other devices are pairing
- **Sensors Not Updating**: Check WeatherFlow API connectivity
- **Accessories Missing**: Try database reset with `--cleardb` flag
- **Connection Issues**: Verify local network connectivity

### Debug Information
Enable debug logging to see HomeKit operations:
```bash
./tempest-homekit-go --loglevel debug --token "your-token"
```

## Dependencies

- **brutella/hap**: Modern HomeKit Accessory Protocol implementation
- **Standard Library**: Context, sync, log packages for lifecycle management
- **Internal Packages**: config, weather packages for data integration