package editor

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"tempest-homekit-go/pkg/alarm"
	"tempest-homekit-go/pkg/logger"
	"tempest-homekit-go/pkg/weather"
)

// Server provides the alarm editor web UI
type Server struct {
	configPath   string
	port         string
	version      string
	config       *alarm.AlarmConfig
	lastLoadTime time.Time
}

// NewServer creates a new alarm editor server
func NewServer(configPath, port, version string) (*Server, error) {
	// Remove @ prefix if present
	path := strings.TrimPrefix(configPath, "@")

	s := &Server{
		configPath: path,
		port:       port,
		version:    version,
	}

	// Load existing config or create new one
	if err := s.loadConfig(); err != nil {
		// If file doesn't exist, create a minimal config
		if os.IsNotExist(err) {
			s.config = &alarm.AlarmConfig{
				Alarms: []alarm.Alarm{},
			}
			logger.Info("Creating new alarm configuration file: %s", path)
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return s, nil
}

// loadConfig loads the alarm configuration from file
func (s *Server) loadConfig() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return err
	}

	var config alarm.AlarmConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	s.config = &config
	s.lastLoadTime = time.Now()
	return nil
}

// saveConfig saves the alarm configuration to file
func (s *Server) saveConfig() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logger.Info("Saved alarm configuration to: %s", s.configPath)
	return nil
}

// Start starts the alarm editor web server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Main editor page
	mux.HandleFunc("/", s.handleIndex)

	// Static files
	mux.HandleFunc("/alarm-editor/static/", s.handleStaticFiles)

	// API endpoints
	mux.HandleFunc("/api/config", s.handleGetConfig)
	mux.HandleFunc("/api/config/save", s.handleSaveConfig)
	mux.HandleFunc("/api/alarms", s.handleListAlarms)
	mux.HandleFunc("/api/alarms/create", s.handleCreateAlarm)
	mux.HandleFunc("/api/alarms/update", s.handleUpdateAlarm)
	mux.HandleFunc("/api/alarms/delete", s.handleDeleteAlarm)
	mux.HandleFunc("/api/tags", s.handleGetTags)
	mux.HandleFunc("/api/validate", s.handleValidate)
	mux.HandleFunc("/api/fields", s.handleGetFields)
	mux.HandleFunc("/api/env-defaults", s.handleGetEnvDefaults)

	addr := ":" + s.port
	logger.Info("Starting Alarm Editor on http://localhost%s", addr)
	logger.Info("Editing: %s", s.configPath)
	logger.Info("Press Ctrl+C to stop")

	return http.ListenAndServe(addr, mux)
}

// handleStaticFiles serves static CSS and JS files
func (s *Server) handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Extract filename from URL path
	filename := strings.TrimPrefix(r.URL.Path, "/alarm-editor/static/")

	logger.Debug("Static file request: %s (path: %s)", filename, r.URL.Path)

	// Serve the file from the physical directory
	filePath := "./pkg/alarm/editor/static/" + filename

	// Set appropriate content type
	switch {
	case strings.HasSuffix(filename, ".css"):
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	case strings.HasSuffix(filename, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	http.ServeFile(w, r, filePath)
}

// handleIndex serves the main editor HTML page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(indexHTML)
	if err != nil {
		logger.Error("Failed to parse HTML template: %v", err)
		http.Error(w, "Internal Server Error: Failed to parse template", http.StatusInternalServerError)
		return
	}

	lastLoad := "never"
	if !s.lastLoadTime.IsZero() {
		lastLoad = s.lastLoadTime.Format("2006-01-02 15:04:05")
	}
	data := map[string]interface{}{
		"ConfigPath": s.configPath,
		"Port":       s.port,
		"Version":    s.version,
		"LastLoad":   lastLoad,
	}
	if err := tmpl.Execute(w, data); err != nil {
		logger.Error("Failed to execute template: %v", err)
		http.Error(w, "Internal Server Error: Failed to render page", http.StatusInternalServerError)
		return
	}
}

// handleGetConfig returns the full alarm configuration
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

// handleSaveConfig saves the entire configuration
func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config alarm.AlarmConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid configuration: %v", err), http.StatusBadRequest)
		return
	}

	s.config = &config
	if err := s.saveConfig(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleListAlarms returns all alarms with optional filtering
func (s *Server) handleListAlarms(w http.ResponseWriter, r *http.Request) {
	nameFilter := r.URL.Query().Get("name")
	tagFilter := r.URL.Query().Get("tag")

	alarms := s.config.Alarms
	filtered := []alarm.Alarm{}

	for _, a := range alarms {
		// Apply name filter (case-insensitive substring match)
		if nameFilter != "" && !strings.Contains(strings.ToLower(a.Name), strings.ToLower(nameFilter)) {
			continue
		}

		// Apply tag filter
		if tagFilter != "" {
			hasTag := false
			for _, tag := range a.Tags {
				if tag == tagFilter {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		filtered = append(filtered, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

// handleCreateAlarm creates a new alarm
func (s *Server) handleCreateAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newAlarm alarm.Alarm
	if err := json.NewDecoder(r.Body).Decode(&newAlarm); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Check for duplicate name
	for _, a := range s.config.Alarms {
		if a.Name == newAlarm.Name {
			http.Error(w, fmt.Sprintf("Alarm with name '%s' already exists", newAlarm.Name), http.StatusConflict)
			return
		}
	}

	// Validate channels
	for i, ch := range newAlarm.Channels {
		if err := ch.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("Channel %d validation failed: %v", i, err), http.StatusBadRequest)
			return
		}
	}

	s.config.Alarms = append(s.config.Alarms, newAlarm)

	if err := s.saveConfig(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleUpdateAlarm updates an existing alarm
func (s *Server) handleUpdateAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updatedAlarm alarm.Alarm
	if err := json.NewDecoder(r.Body).Decode(&updatedAlarm); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Get original name from query parameter (for name changes)
	oldName := r.URL.Query().Get("oldName")
	if oldName == "" {
		oldName = updatedAlarm.Name // Default to current name if not changing
	}

	// Check for duplicate name (if name is changing)
	if oldName != updatedAlarm.Name {
		for _, a := range s.config.Alarms {
			if a.Name == updatedAlarm.Name {
				http.Error(w, fmt.Sprintf("Alarm with name '%s' already exists", updatedAlarm.Name), http.StatusConflict)
				return
			}
		}
	}

	// Find and update the alarm by old name
	found := false
	for i, a := range s.config.Alarms {
		if a.Name == oldName {
			// Validate channels
			for j, ch := range updatedAlarm.Channels {
				if err := ch.Validate(); err != nil {
					http.Error(w, fmt.Sprintf("Channel %d validation failed: %v", j, err), http.StatusBadRequest)
					return
				}
			}

			s.config.Alarms[i] = updatedAlarm
			found = true
			break
		}
	}

	if !found {
		http.Error(w, fmt.Sprintf("Alarm '%s' not found", oldName), http.StatusNotFound)
		return
	}

	if err := s.saveConfig(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleDeleteAlarm deletes an alarm by name
func (s *Server) handleDeleteAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name parameter required", http.StatusBadRequest)
		return
	}

	// Find and remove the alarm
	newAlarms := []alarm.Alarm{}
	found := false
	for _, a := range s.config.Alarms {
		if a.Name == name {
			found = true
			continue
		}
		newAlarms = append(newAlarms, a)
	}

	if !found {
		http.Error(w, fmt.Sprintf("Alarm '%s' not found", name), http.StatusNotFound)
		return
	}

	s.config.Alarms = newAlarms

	if err := s.saveConfig(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleGetTags returns all unique tags from all alarms
func (s *Server) handleGetTags(w http.ResponseWriter, r *http.Request) {
	tagSet := make(map[string]bool)
	for _, a := range s.config.Alarms {
		for _, tag := range a.Tags {
			tagSet[tag] = true
		}
	}

	tags := []string{}
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

// handleValidate validates an alarm condition
func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Condition string `json:"condition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create a test observation with reasonable values
	testObs := &weather.Observation{
		AirTemperature:   20.0,
		RelativeHumidity: 50.0,
		StationPressure:  1013.25,
		WindAvg:          5.0,
		Illuminance:      10000,
		UV:               3,
	}

	// Validate condition format before evaluation
	condition := strings.TrimSpace(req.Condition)

	var validationErr error

	if condition == "" {
		validationErr = fmt.Errorf("condition cannot be empty")
	} else if strings.HasSuffix(condition, "&&") || strings.HasSuffix(condition, "||") {
		validationErr = fmt.Errorf("condition ends with incomplete operator (&&/||)")
	} else if strings.HasPrefix(condition, "&&") || strings.HasPrefix(condition, "||") {
		validationErr = fmt.Errorf("condition starts with incomplete operator (&&/||)")
	} else if strings.Contains(condition, "&&") && strings.Contains(condition, "|| ") {
		// Check for empty parts in compound conditions
		parts := strings.FieldsFunc(condition, func(r rune) bool {
			return r == '&' || r == '|'
		})
		for _, part := range parts {
			if strings.TrimSpace(part) == "" {
				validationErr = fmt.Errorf("logical operator (&&/||) requires expressions on both sides")
				break
			}
		}
	}

	response := map[string]interface{}{}

	if validationErr != nil {
		response["valid"] = false
		response["error"] = validationErr.Error()
	} else {
		// Create a dummy alarm context for change detection validation
		dummyAlarm := &alarm.Alarm{
			Name: "validation-test",
		}

		evaluator := alarm.NewEvaluator()
		_, err := evaluator.EvaluateWithAlarm(condition, testObs, dummyAlarm)

		response["valid"] = err == nil

		if err != nil {
			response["error"] = err.Error()
		} else {
			// Add paraphrase for valid conditions
			response["paraphrase"] = evaluator.Paraphrase(condition)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetFields returns available fields for conditions
func (s *Server) handleGetFields(w http.ResponseWriter, r *http.Request) {
	evaluator := alarm.NewEvaluator()
	fields := evaluator.GetAvailableFields()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fields)
}

// handleGetEnvDefaults returns default values from environment variables
func (s *Server) handleGetEnvDefaults(w http.ResponseWriter, r *http.Request) {
	defaults := map[string]string{
		"emailTo": os.Getenv("MS365_TO_ADDRESS"),
		"smsTo":   os.Getenv("SMS_TO_NUMBER"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(defaults)
}
