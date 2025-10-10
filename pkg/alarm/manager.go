package alarm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/weather"

	"github.com/fsnotify/fsnotify"
)

// Manager manages alarm evaluation and notifications
type Manager struct {
	config          *AlarmConfig
	configPath      string
	lastLoadTime    time.Time
	evaluator       *Evaluator
	notifierFactory *NotifierFactory
	watcher         *fsnotify.Watcher
	stationName     string
	mu              sync.RWMutex
	stopChan        chan struct{}
}

// NewManager creates a new alarm manager
func NewManager(configInput string, stationName string) (*Manager, error) {
	config, err := LoadAlarmConfig(configInput)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		config:          config,
		evaluator:       NewEvaluator(),
		notifierFactory: NewNotifierFactory(config),
		stationName:     stationName,
		stopChan:        make(chan struct{}),
		lastLoadTime:    time.Now(),
	}

	// If config is from file, set up file watching
	if strings.HasPrefix(configInput, "@") {
		m.configPath = strings.TrimPrefix(configInput, "@")
		if err := m.setupFileWatcher(); err != nil {
			logger.Error("Failed to set up file watcher: %v", err)
			// Non-fatal: continue without file watching
		}
	}

	logger.Info("Alarm manager initialized with %d alarms", len(config.Alarms))

	// Log active alarms
	enabledCount := 0
	for _, alarm := range config.Alarms {
		if alarm.Enabled {
			enabledCount++
			logger.Info("Active alarm: %s - %s (cooldown: %ds, channels: %d)",
				alarm.Name, alarm.Condition, alarm.Cooldown, len(alarm.Channels))
		}
	}
	logger.Info("%d of %d alarms are enabled", enabledCount, len(config.Alarms))

	return m, nil
}

// setupFileWatcher sets up cross-platform file watching for alarm config
func (m *Manager) setupFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	m.watcher = watcher

	// Watch the config file directory (not the file directly, for better compatibility)
	configDir := filepath.Dir(m.configPath)
	if err := watcher.Add(configDir); err != nil {
		watcher.Close()
		return fmt.Errorf("failed to watch config directory: %w", err)
	}

	// Start watching in background
	go m.watchConfigFile()

	logger.Info("Watching alarm config file for changes: %s", m.configPath)
	return nil
}

// watchConfigFile monitors for config file changes
func (m *Manager) watchConfigFile() {
	configFileName := filepath.Base(m.configPath)

	for {
		select {
		case <-m.stopChan:
			return
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			// Check if this event is for our config file
			if filepath.Base(event.Name) != configFileName {
				continue
			}

			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				logger.Info("Alarm config file changed, reloading: %s", m.configPath)
				if err := m.reloadConfig(); err != nil {
					logger.Error("Failed to reload alarm config: %v", err)
				} else {
					logger.Info("Alarm config reloaded successfully")
				}
			}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			logger.Error("File watcher error: %v", err)
		}
	}
}

// reloadConfig reloads the alarm configuration from file
func (m *Manager) reloadConfig() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var newConfig AlarmConfig
	if err := json.Unmarshal(data, &newConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	m.mu.Lock()
	m.config = &newConfig
	m.notifierFactory = NewNotifierFactory(&newConfig)
	m.lastLoadTime = time.Now()
	m.mu.Unlock()

	return nil
}

// ProcessObservation evaluates all alarms against a new weather observation
func (m *Manager) ProcessObservation(obs *weather.Observation) {
	if obs == nil {
		return
	}

	m.mu.RLock()
	alarms := make([]Alarm, len(m.config.Alarms))
	copy(alarms, m.config.Alarms)
	m.mu.RUnlock()

	for i := range alarms {
		alarm := &alarms[i]

		if !alarm.Enabled {
			logger.Debug("Skipping disabled alarm: %s", alarm.Name)
			continue
		}

		if !alarm.CanFire() {
			logger.Debug("Alarm %s in cooldown, skipping (last fired: %v)", alarm.Name, alarm.lastFired)
			continue
		}

		logger.Debug("Testing alarm: %s - %s", alarm.Name, alarm.Condition)

		// Evaluate condition (pass alarm for change detection support)
		triggered, err := m.evaluator.EvaluateWithAlarm(alarm.Condition, obs, alarm)
		if err != nil {
			logger.Error("Failed to evaluate alarm %s: %v", alarm.Name, err)
			continue
		}

		if triggered {
			logger.Info("Alarm triggered: %s (condition: %s)", alarm.Name, alarm.Condition)
			m.sendNotifications(alarm, obs)
			alarm.MarkFired()

			// Update the alarm in the config to persist the lastFired time
			m.mu.Lock()
			for j := range m.config.Alarms {
				if m.config.Alarms[j].Name == alarm.Name {
					m.config.Alarms[j].lastFired = alarm.lastFired
					break
				}
			}
			m.mu.Unlock()
		}
	}
}

// sendNotifications sends notifications through all configured channels for an alarm
func (m *Manager) sendNotifications(alarm *Alarm, obs *weather.Observation) {
	for i := range alarm.Channels {
		channel := &alarm.Channels[i]

		notifier, err := m.notifierFactory.GetNotifier(channel.Type)
		if err != nil {
			logger.Error("Failed to get notifier for %s: %v", channel.Type, err)
			continue
		}

		if err := notifier.Send(alarm, channel, obs, m.stationName); err != nil {
			logger.Error("Failed to send %s notification for alarm %s: %v",
				channel.Type, alarm.Name, err)
		} else {
			logger.Info("Sent %s notification for alarm %s", channel.Type, alarm.Name)
		}
	}
}

// Stop stops the alarm manager and file watcher
func (m *Manager) Stop() {
	close(m.stopChan)
	if m.watcher != nil {
		m.watcher.Close()
	}
	logger.Info("Alarm manager stopped")
}

// GetConfig returns a copy of the current alarm configuration
func (m *Manager) GetConfig() *AlarmConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modifications
	configCopy := *m.config
	return &configCopy
}

// GetAlarmCount returns the number of configured alarms
func (m *Manager) GetAlarmCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.config.Alarms)
}

// GetEnabledAlarmCount returns the number of enabled alarms
func (m *Manager) GetEnabledAlarmCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, alarm := range m.config.Alarms {
		if alarm.Enabled {
			count++
		}
	}
	return count
}

// GetConfigPath returns the alarm configuration file path
func (m *Manager) GetConfigPath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.configPath == "" {
		return "Inline configuration"
	}
	return m.configPath
}

// GetLastLoadTime returns when the configuration was last loaded
func (m *Manager) GetLastLoadTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastLoadTime
}
