// Package logger provides centralized logging with level filtering for the Tempest HomeKit application.
package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Global log level for filtering
var currentLogLevel string = "error"
var logFilter string = "" // Filter string for log messages

// Log level constants
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// SetLogLevel configures the global log level and output settings
// Accepts both 'warn' and 'warning' (normalized to 'warn' internally)
func SetLogLevel(level string) {
	// Normalize 'warning' to 'warn' for consistency
	if level == "warning" {
		level = "warn"
	}
	currentLogLevel = level
	switch level {
	case "debug":
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	case "info":
		log.SetFlags(log.LstdFlags)
	case "warn":
		log.SetFlags(log.LstdFlags)
	case "error":
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
	default:
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
	}
}

// SetLogFilter configures the global log filter string
// Only messages containing this string (case-insensitive) will be output
func SetLogFilter(filter string) {
	logFilter = strings.ToLower(filter)
}

// shouldLog checks if a message should be logged based on the filter
func shouldLog(message string) bool {
	if logFilter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(message), logFilter)
}

// Debug prints debug messages only if log level is debug
func Debug(format string, v ...interface{}) {
	if currentLogLevel == LogLevelDebug {
		message := fmt.Sprintf(format, v...)
		if shouldLog(message) {
			log.Printf("DEBUG: %s", message)
		}
	}
}

// Info prints info and debug messages only if log level is debug or info
func Info(format string, v ...interface{}) {
	if currentLogLevel == LogLevelDebug || currentLogLevel == LogLevelInfo {
		message := fmt.Sprintf(format, v...)
		if shouldLog(message) {
			log.Printf("INFO: %s", message)
		}
	}
}

// Warn prints warning messages if log level is debug, info, warn, or error
func Warn(format string, v ...interface{}) {
	if currentLogLevel == LogLevelDebug || currentLogLevel == LogLevelInfo || currentLogLevel == LogLevelWarn || currentLogLevel == LogLevelError {
		message := fmt.Sprintf(format, v...)
		if shouldLog(message) {
			log.Printf("WARN: %s", message)
		}
	}
}

// Error always prints error messages
func Error(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if shouldLog(message) {
		log.Printf("ERROR: %s", message)
	}
}

// Alarm always prints alarm notifications, bypassing log level filtering
// Alarms are critical events that should always be visible
func Alarm(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if shouldLog(message) {
		log.Printf("🚨 ALARM: %s", message)
	}
}
