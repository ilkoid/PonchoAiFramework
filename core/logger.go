package core

// This file provides compatibility layer for core package to use interfaces logger
// All logger implementations have been moved to interfaces/interfaces.go to eliminate duplication
// This file now provides type aliases and helper functions for backward compatibility

import (
	"os"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// Compatibility functions that delegate to interfaces package
func NewDefaultLogger() interfaces.Logger {
	return interfaces.NewDefaultLogger()
}

func NewDefaultLoggerWithOutput(output interface{}) interfaces.Logger {
	if writer, ok := output.(interface{ Write([]byte) (int, error) }); ok {
		return interfaces.NewDefaultLoggerWithOutput(writer)
	}
	return interfaces.NewDefaultLogger()
}

func NewDefaultLoggerWithLevel(level interfaces.LogLevel) interfaces.Logger {
	return interfaces.NewDefaultLoggerWithLevel(level)
}

func NewJSONLogger() interfaces.Logger {
	return interfaces.NewJSONLogger()
}

func NewJSONLoggerWithOutput(output interface{}) interfaces.Logger {
	if writer, ok := output.(interface{ Write([]byte) (int, error) }); ok {
		return interfaces.NewJSONLoggerWithOutput(writer)
	}
	return interfaces.NewJSONLogger()
}

func NewMultiLogger(loggers ...interfaces.Logger) interfaces.Logger {
	return interfaces.NewMultiLogger(loggers...)
}

func NewNoOpLogger() interfaces.Logger {
	return interfaces.NewNoOpLogger()
}

// Helper functions for creating loggers from configuration

// CreateLoggerFromConfig creates a logger based on configuration
func CreateLoggerFromConfig(config *interfaces.LoggingConfig) interfaces.Logger {
	if config == nil {
		return NewDefaultLogger()
	}

	// Parse log level
	level := parseLogLevel(config.Level)

	// Create base logger
	var logger interfaces.Logger
	switch config.Format {
	case "json":
		logger = NewJSONLogger()
	default:
		logger = NewDefaultLogger()
	}

	// Set log level using type assertion
	if dl, ok := logger.(interface{ SetLevel(interfaces.LogLevel) }); ok {
		dl.SetLevel(level)
	}

	// Set output if file is specified
	if config.File != "" {
		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			if sl, ok := logger.(interface{ SetOutput(interface{ Write([]byte) (int, error) }) }); ok {
				sl.SetOutput(file)
			}
		}
	}

	return logger
}

// parseLogLevel parses string log level to LogLevel
func parseLogLevel(level string) interfaces.LogLevel {
	switch level {
	case "debug":
		return interfaces.LogLevelDebug
	case "info":
		return interfaces.LogLevelInfo
	case "warn", "warning":
		return interfaces.LogLevelWarn
	case "error":
		return interfaces.LogLevelError
	default:
		return interfaces.LogLevelInfo
	}
}