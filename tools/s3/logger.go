package s3

import (
	"log"
	"os"
)

// Logger defines the logging interface
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// DefaultLogger provides a simple console logger implementation
type DefaultLogger struct {
	prefix string
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() Logger {
	prefix := "[S3] "

	return &DefaultLogger{prefix: prefix}
}

// Info logs an informational message
func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	log.SetPrefix(l.prefix)
	log.Printf("[INFO] "+msg, args...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	log.SetPrefix(l.prefix)
	log.Printf("[WARN] "+msg, args...)
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	log.SetPrefix(l.prefix)
	log.Printf("[ERROR] "+msg, args...)
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	if os.Getenv("LOG_LEVEL") == "debug" {
		log.SetPrefix(l.prefix)
		log.Printf("[DEBUG] "+msg, args...)
	}
}