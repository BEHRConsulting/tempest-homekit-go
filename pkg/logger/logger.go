// Package logger provides centralized logging with level filtering for the Tempest HomeKit application.
package logger

import (
	"log"
	"os"
)

// Global log level for filtering
var currentLogLevel string = "error"

// Log level constants
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelError = "error"
)

// SetLogLevel configures the global log level and output settings
func SetLogLevel(level string) {
	currentLogLevel = level
	switch level {
	case "debug":
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	case "info":
		log.SetFlags(log.LstdFlags)
	case "error":
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
	default:
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
	}
}

// Debug prints debug messages only if log level is debug
func Debug(format string, v ...interface{}) {
	if currentLogLevel == LogLevelDebug {
		log.Printf("DEBUG: "+format, v...)
	}
}

// Info prints info and debug messages only if log level is debug or info
func Info(format string, v ...interface{}) {
	if currentLogLevel == LogLevelDebug || currentLogLevel == LogLevelInfo {
		log.Printf("INFO: "+format, v...)
	}
}

// Error always prints error messages
func Error(format string, v ...interface{}) {
	log.Printf("ERROR: "+format, v...)
}
