package interfaces

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// DefaultLogger is a simple implementation of the Logger interface
type DefaultLogger struct {
	level  LogLevel
	format LogFormat
	file   *os.File
}

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogFormat represents the log format
type LogFormat int

const (
	LogFormatText LogFormat = iota
	LogFormatJSON
)

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() Logger {
	return &DefaultLogger{
		level:  LogLevelInfo,
		format: LogFormatText,
	}
}

// NewLogger creates a new logger with specified settings
func NewLogger(level LogLevel, format LogFormat, file *os.File) Logger {
	return &DefaultLogger{
		level:  level,
		format: format,
		file:   file,
	}
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

	var output *os.File
	if l.file != nil {
		output = l.file
	} else {
		output = os.Stdout
	}

	switch l.format {
	case LogFormatJSON:
		l.logJSON(output, level, msg, fields...)
	default:
		l.logText(output, level, msg, fields...)
	}
}

// logText logs in text format
func (l *DefaultLogger) logText(output *os.File, level LogLevel, msg string, fields ...interface{}) {
	levelStr := []string{"DEBUG", "INFO", "WARN", "ERROR"}[level]

	if len(fields) > 0 {
		var fieldStrs []string
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				fieldStrs = append(fieldStrs, fmt.Sprintf("%v=%v", fields[i], fields[i+1]))
			}
		}
		fmt.Fprintf(output, "[%s] %s (%s)\n", levelStr, msg, strings.Join(fieldStrs, ", "))
	} else {
		fmt.Fprintf(output, "[%s] %s\n", levelStr, msg)
	}
}

// logJSON logs in JSON format
func (l *DefaultLogger) logJSON(output *os.File, level LogLevel, msg string, fields ...interface{}) {
	levelStr := []string{"debug", "info", "warn", "error"}[level]

	logEntry := map[string]interface{}{
		"level":   levelStr,
		"message": msg,
	}

	if len(fields) > 0 {
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key := fmt.Sprintf("%v", fields[i])
				logEntry[key] = fields[i+1]
			}
		}
	}

	// Simple JSON marshaling for basic types
	var jsonStr string
	if len(fields) > 0 {
		var parts []string
		parts = append(parts, fmt.Sprintf(`"level":"%s"`, levelStr))
		parts = append(parts, fmt.Sprintf(`"message":"%s"`, msg))

		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key := fmt.Sprintf("%v", fields[i])
				value := fields[i+1]
				parts = append(parts, fmt.Sprintf(`"%v":%v`, key, value))
			}
		}

		jsonStr = "{" + strings.Join(parts, ",") + "}"
	} else {
		jsonStr = fmt.Sprintf(`{"level":"%s","message":"%s"}`, levelStr, msg)
	}

	fmt.Fprintf(output, "%s\n", jsonStr)
}

// MultiLogger logs to multiple loggers
type MultiLogger struct {
	loggers []Logger
}

// NewMultiLogger creates a new multi-logger
func NewMultiLogger(loggers ...Logger) Logger {
	return &MultiLogger{loggers: loggers}
}

// Debug logs a debug message to all loggers
func (m *MultiLogger) Debug(msg string, fields ...interface{}) {
	for _, logger := range m.loggers {
		logger.Debug(msg, fields...)
	}
}

// Info logs an info message to all loggers
func (m *MultiLogger) Info(msg string, fields ...interface{}) {
	for _, logger := range m.loggers {
		logger.Info(msg, fields...)
	}
}

// Warn logs a warning message to all loggers
func (m *MultiLogger) Warn(msg string, fields ...interface{}) {
	for _, logger := range m.loggers {
		logger.Warn(msg, fields...)
	}
}

// Error logs an error message to all loggers
func (m *MultiLogger) Error(msg string, fields ...interface{}) {
	for _, logger := range m.loggers {
		logger.Error(msg, fields...)
	}
}

// StandardLogger adapts standard log to Logger interface
type StandardLogger struct {
	logger *log.Logger
}

// NewStandardLogger creates a new standard logger
func NewStandardLogger(logger *log.Logger) Logger {
	return &StandardLogger{logger: logger}
}

// Debug logs a debug message
func (s *StandardLogger) Debug(msg string, fields ...interface{}) {
	s.logger.Printf("[DEBUG] %s %v", msg, fields)
}

// Info logs an info message
func (s *StandardLogger) Info(msg string, fields ...interface{}) {
	s.logger.Printf("[INFO] %s %v", msg, fields)
}

// Warn logs a warning message
func (s *StandardLogger) Warn(msg string, fields ...interface{}) {
	s.logger.Printf("[WARN] %s %v", msg, fields)
}

// Error logs an error message
func (s *StandardLogger) Error(msg string, fields ...interface{}) {
	s.logger.Printf("[ERROR] %s %v", msg, fields)
}
