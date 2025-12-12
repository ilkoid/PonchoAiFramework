package core

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns string representation of log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// DefaultLogger is the default implementation of Logger interface
type DefaultLogger struct {
	level  LogLevel
	output io.Writer
	mutex  sync.Mutex
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		level:  LogLevelInfo,
		output: os.Stdout,
	}
}

// NewDefaultLoggerWithOutput creates a new default logger with custom output
func NewDefaultLoggerWithOutput(output io.Writer) *DefaultLogger {
	return &DefaultLogger{
		level:  LogLevelInfo,
		output: output,
	}
}

// NewDefaultLoggerWithLevel creates a new default logger with custom level
func NewDefaultLoggerWithLevel(level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		level:  level,
		output: os.Stdout,
	}
}

// SetLevel sets the logging level
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// SetOutput sets the output writer
func (l *DefaultLogger) SetOutput(output io.Writer) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.output = output
}

// GetLevel returns the current logging level
func (l *DefaultLogger) GetLevel() LogLevel {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.level
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, fields ...interface{}) {
	l.log(LogLevelDebug, msg, fields...)
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	l.log(LogLevelInfo, msg, fields...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, fields ...interface{}) {
	l.log(LogLevelWarn, msg, fields...)
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	l.log(LogLevelError, msg, fields...)
}

// log is the internal logging method
func (l *DefaultLogger) log(level LogLevel, msg string, fields ...interface{}) {
	if level < l.level {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(l.output, "[%s] %s: %s", timestamp, level.String(), msg)

	if len(fields) > 0 {
		fmt.Fprintf(l.output, " %v", fields)
	}

	fmt.Fprintln(l.output)
}

// JSONLogger is a JSON-formatted logger implementation
type JSONLogger struct {
	level  LogLevel
	output io.Writer
	mutex  sync.Mutex
}

// NewJSONLogger creates a new JSON logger
func NewJSONLogger() *JSONLogger {
	return &JSONLogger{
		level:  LogLevelInfo,
		output: os.Stdout,
	}
}

// NewJSONLoggerWithOutput creates a new JSON logger with custom output
func NewJSONLoggerWithOutput(output io.Writer) *JSONLogger {
	return &JSONLogger{
		level:  LogLevelInfo,
		output: output,
	}
}

// SetLevel sets the logging level
func (l *JSONLogger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// SetOutput sets the output writer
func (l *JSONLogger) SetOutput(output io.Writer) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.output = output
}

// Debug logs a debug message in JSON format
func (l *JSONLogger) Debug(msg string, fields ...interface{}) {
	l.log(LogLevelDebug, msg, fields...)
}

// Info logs an info message in JSON format
func (l *JSONLogger) Info(msg string, fields ...interface{}) {
	l.log(LogLevelInfo, msg, fields...)
}

// Warn logs a warning message in JSON format
func (l *JSONLogger) Warn(msg string, fields ...interface{}) {
	l.log(LogLevelWarn, msg, fields...)
}

// Error logs an error message in JSON format
func (l *JSONLogger) Error(msg string, fields ...interface{}) {
	l.log(LogLevelError, msg, fields...)
}

// log is the internal logging method for JSON format
func (l *JSONLogger) log(level LogLevel, msg string, fields ...interface{}) {
	if level < l.level {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level.String(),
		"message":   msg,
	}

	if len(fields) > 0 {
		// If odd number of fields, treat last one as message
		if len(fields)%2 == 1 {
			logEntry["message"] = fmt.Sprintf("%s %v", msg, fields[len(fields)-1])
			fields = fields[:len(fields)-1]
		}

		// Process field pairs
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key := fmt.Sprintf("%v", fields[i])
				logEntry[key] = fields[i+1]
			}
		}
	}

	// Simple JSON marshaling (avoiding json.Marshal for simplicity)
	jsonStr := fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s"`,
		logEntry["timestamp"], logEntry["level"], logEntry["message"])

	// Add fields if any
	for key, value := range logEntry {
		if key != "timestamp" && key != "level" && key != "message" {
			jsonStr += fmt.Sprintf(`,"%s":"%v"`, key, value)
		}
	}

	jsonStr += "}"
	fmt.Fprintln(l.output, jsonStr)
}

// MultiLogger is a logger that writes to multiple loggers
type MultiLogger struct {
	loggers []Logger
	mutex   sync.RWMutex
}

// NewMultiLogger creates a new multi-logger
func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{
		loggers: loggers,
	}
}

// AddLogger adds a logger to the multi-logger
func (l *MultiLogger) AddLogger(logger Logger) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.loggers = append(l.loggers, logger)
}

// RemoveLogger removes a logger from the multi-logger
func (l *MultiLogger) RemoveLogger(logger Logger) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for i, existingLogger := range l.loggers {
		if existingLogger == logger {
			l.loggers = append(l.loggers[:i], l.loggers[i+1:]...)
			break
		}
	}
}

// Debug logs a debug message to all loggers
func (l *MultiLogger) Debug(msg string, fields ...interface{}) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for _, logger := range l.loggers {
		logger.Debug(msg, fields...)
	}
}

// Info logs an info message to all loggers
func (l *MultiLogger) Info(msg string, fields ...interface{}) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for _, logger := range l.loggers {
		logger.Info(msg, fields...)
	}
}

// Warn logs a warning message to all loggers
func (l *MultiLogger) Warn(msg string, fields ...interface{}) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for _, logger := range l.loggers {
		logger.Warn(msg, fields...)
	}
}

// Error logs an error message to all loggers
func (l *MultiLogger) Error(msg string, fields ...interface{}) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for _, logger := range l.loggers {
		logger.Error(msg, fields...)
	}
}

// Helper functions for creating loggers from configuration

// CreateLoggerFromConfig creates a logger based on configuration
func CreateLoggerFromConfig(config *interfaces.LoggingConfig) Logger {
	if config == nil {
		return NewDefaultLogger()
	}

	// Parse log level
	level := parseLogLevel(config.Level)

	// Create base logger
	var logger Logger
	switch config.Format {
	case "json":
		logger = NewJSONLogger()
	default:
		logger = NewDefaultLogger()
	}

	// Set log level
	if defaultLogger, ok := logger.(*DefaultLogger); ok {
		defaultLogger.SetLevel(level)
	} else if jsonLogger, ok := logger.(*JSONLogger); ok {
		jsonLogger.SetLevel(level)
	}

	// Set output if file is specified
	if config.File != "" {
		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			if defaultLogger, ok := logger.(*DefaultLogger); ok {
				defaultLogger.SetOutput(file)
			} else if jsonLogger, ok := logger.(*JSONLogger); ok {
				jsonLogger.SetOutput(file)
			}
		}
	}

	return logger
}

// parseLogLevel parses string log level to LogLevel
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// NoOpLogger is a logger that does nothing (for testing)
type NoOpLogger struct{}

// NewNoOpLogger creates a no-op logger
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// Debug does nothing
func (l *NoOpLogger) Debug(msg string, fields ...interface{}) {}

// Info does nothing
func (l *NoOpLogger) Info(msg string, fields ...interface{}) {}

// Warn does nothing
func (l *NoOpLogger) Warn(msg string, fields ...interface{}) {}

// Error does nothing
func (l *NoOpLogger) Error(msg string, fields ...interface{}) {}
