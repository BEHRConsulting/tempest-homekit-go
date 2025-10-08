// Package weather provides status management for Tempest weather station data.
// The StatusManager handles periodic web scraping when enabled via CLI flag.
package weather

import (
	"fmt"
	"sync"
	"tempest-homekit-go/pkg/logger"
	"time"
)

// StatusManager handles periodic scraping and caching of station status
type StatusManager struct {
	stationID      int
	logLevel       string
	useWebScraping bool
	cachedStatus   *StationStatus
	mutex          sync.RWMutex
	stopChan       chan bool
	scrapingActive bool
}

// NewStatusManager creates a new status manager
func NewStatusManager(stationID int, logLevel string, useWebScraping bool) *StatusManager {
	manager := &StatusManager{
		stationID:      stationID,
		logLevel:       logLevel,
		useWebScraping: useWebScraping,
		stopChan:       make(chan bool),
	}

	// Initialize with fallback status
	manager.cachedStatus = manager.createFallbackStatus()

	return manager
}

// Start begins the periodic status scraping if web scraping is enabled
func (sm *StatusManager) Start() {
	if !sm.useWebScraping {
		if sm.logLevel == "debug" {
			logger.Debug("Web status scraping disabled, using API fallback only")
		}
		return
	}

	if sm.logLevel == "debug" {
		logger.Debug("Starting status manager with 15-minute web scraping interval")
	}

	sm.scrapingActive = true

	// Do initial scrape
	go sm.performScrape()

	// Start periodic scraping
	go sm.periodicScraping()
}

// Stop stops the periodic scraping
func (sm *StatusManager) Stop() {
	if sm.scrapingActive {
		sm.stopChan <- true
		sm.scrapingActive = false
		if sm.logLevel == "debug" {
			logger.Debug("Status manager stopped")
		}
	}
}

// GetStatus returns the current cached status
func (sm *StatusManager) GetStatus() *StationStatus {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Return a copy to avoid concurrent modification
	statusCopy := *sm.cachedStatus
	return &statusCopy
}

// periodicScraping runs the scraping loop every 15 minutes
func (sm *StatusManager) periodicScraping() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go sm.performScrape()
		case <-sm.stopChan:
			return
		}
	}
}

// performScrape attempts to scrape status data
func (sm *StatusManager) performScrape() {
	if sm.logLevel == "debug" {
		logger.Debug("Performing status scrape for station %d", sm.stationID)
	}

	var status *StationStatus
	var err error

	if sm.useWebScraping {
		// Try headless browser scraping first
		status, err = GetStationStatusWithBrowser(sm.stationID, sm.logLevel)
		if err != nil {
			if sm.logLevel == "debug" {
				logger.Debug("Browser scraping failed: %v", err)
			}
			// Fall back to regular HTTP scraping
			status, err = GetStationStatus(sm.stationID, sm.logLevel)
			if err == nil && sm.hasUsefulData(status) {
				status.DataSource = "api, web status page"
				status.LastScraped = time.Now().UTC().Format(time.RFC3339)
				status.ScrapingEnabled = true
				if sm.logLevel == "debug" {
					logger.Debug("HTTP scraping succeeded with useful data")
				}
			} else if err == nil && !sm.hasUsefulData(status) {
				if sm.logLevel == "debug" {
					logger.Debug("HTTP scraping succeeded but no useful data found")
				}
			}
		} else if sm.hasUsefulData(status) {
			// Browser scraping succeeded and got useful data
			status.DataSource = "api, web status page"
			status.LastScraped = time.Now().UTC().Format(time.RFC3339)
			status.ScrapingEnabled = true
			if sm.logLevel == "debug" {
				logger.Debug("Browser scraping succeeded with useful data")
			}
		} else if status != nil && !sm.hasUsefulData(status) {
			if sm.logLevel == "debug" {
				logger.Debug("Browser scraping succeeded but no useful data found - Battery: %s, DeviceUptime: %s, HubUptime: %s",
					status.BatteryVoltage, status.DeviceUptime, status.HubUptime)
			}
		}
	}

	// If scraping failed or disabled, or got no useful data, create fallback status
	if status == nil || err != nil || !sm.hasUsefulData(status) {
		status = sm.createFallbackStatus()
		if sm.logLevel == "debug" {
			logger.Debug("Using fallback status (scraping failed or no useful data)")
		}
	}

	// Update cached status
	sm.mutex.Lock()
	sm.cachedStatus = status
	sm.mutex.Unlock()

	if sm.logLevel == "debug" {
		logger.Debug("Status updated - Source: %s, Battery: %s, DeviceUptime: %s, LastScraped: %s",
			status.DataSource, status.BatteryVoltage, status.DeviceUptime, status.LastScraped)
	}
}

// hasUsefulData checks if the status contains any useful scraped data
func (sm *StatusManager) hasUsefulData(status *StationStatus) bool {
	if status == nil {
		return false
	}

	// Consider it successful if we got any of these key data points
	hasData := (status.BatteryVoltage != "" && status.BatteryVoltage != "--") ||
		(status.DeviceUptime != "" && status.DeviceUptime != "--") ||
		(status.HubUptime != "" && status.HubUptime != "--") ||
		(status.DeviceNetworkStatus != "" && status.DeviceNetworkStatus != "--") ||
		(status.HubNetworkStatus != "" && status.HubNetworkStatus != "--") ||
		(status.DeviceSerialNumber != "" && status.DeviceSerialNumber != "--") ||
		(status.HubSerialNumber != "" && status.HubSerialNumber != "--")

	if sm.logLevel == "debug" {
		logger.Debug("hasUsefulData check - Battery: '%s', DeviceUptime: '%s', HubUptime: '%s', DeviceNetwork: '%s', HubNetwork: '%s', DeviceSerial: '%s', HubSerial: '%s' -> %t",
			status.BatteryVoltage, status.DeviceUptime, status.HubUptime,
			status.DeviceNetworkStatus, status.HubNetworkStatus,
			status.DeviceSerialNumber, status.HubSerialNumber, hasData)
	}

	return hasData
}

// createFallbackStatus creates a status with fallback values and appropriate metadata
// If latestObs is provided, it will use the battery voltage from the observation
func (sm *StatusManager) createFallbackStatus() *StationStatus {
	dataSource := "api"
	if sm.useWebScraping {
		dataSource = "api" // Even when scraping is enabled but fails, we still have API data
	}

	status := &StationStatus{
		BatteryVoltage:      "--",
		BatteryStatus:       "--",
		DeviceUptime:        "--",
		HubUptime:           "--",
		DeviceNetworkStatus: "--",
		HubNetworkStatus:    "--",
		DeviceSignal:        "--",
		HubWiFiSignal:       "--",
		SensorStatus:        "--",
		DeviceLastObs:       "--",
		DeviceSerialNumber:  "--",
		DeviceFirmware:      "--",
		HubLastStatus:       "--",
		HubSerialNumber:     "--",
		HubFirmware:         "--",
		DataSource:          dataSource,
		LastScraped:         time.Now().UTC().Format(time.RFC3339),
		ScrapingEnabled:     sm.useWebScraping,
	}

	return status
}

// UpdateBatteryFromObservation updates the cached status with battery data from the latest observation
func (sm *StatusManager) UpdateBatteryFromObservation(obs *Observation) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.cachedStatus == nil {
		return
	}

	// Only update battery data if we're using API-only data (no scraping or scraping failed)
	if sm.cachedStatus.DataSource == "api" || sm.cachedStatus.DataSource == "fallback" {
		if obs != nil && obs.Battery > 0 {
			sm.cachedStatus.BatteryVoltage = fmt.Sprintf("%.1fV", obs.Battery)
			sm.cachedStatus.BatteryStatus = "Good" // Assume good status if we have battery data
			if sm.logLevel == "debug" {
				logger.Debug("Updated battery data from observation: %s", sm.cachedStatus.BatteryVoltage)
			}
		}
	}
}
