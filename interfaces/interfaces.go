// Package interfaces provides core interface definitions for PonchoFramework
//
// Key responsibilities:
// - Define unified interfaces for models, tools, and flows
// - Establish contracts for all framework components
// - Enable extensibility and modularity
//
// Core interfaces:
// - PonchoModel: AI model abstraction with streaming and capabilities
// - PonchoTool: Tool execution interface with schema validation
// - PonchoFlow: Workflow orchestration interface with dependencies
// - PonchoFramework: Main framework orchestrator
//
// Implementation notes:
// - Uses dependency injection pattern
// - Supports streaming and batch operations
// - Thread-safe design for concurrent access
// - Registry pattern for component management
package interfaces

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// PonchoModel defines the interface for all AI models in the framework
type PonchoModel interface {
	// Core generation methods
	Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
	GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error

	// Capability checks
	SupportsStreaming() bool
	SupportsTools() bool
	SupportsVision() bool
	SupportsSystemRole() bool

	// Metadata
	Name() string
	Provider() string
	MaxTokens() int
	DefaultTemperature() float32

	// Lifecycle
	Initialize(ctx context.Context, config map[string]interface{}) error
	Shutdown(ctx context.Context) error
}

// PonchoTool defines the interface for all tools in the framework
type PonchoTool interface {
	// Identity
	Name() string
	Description() string
	Version() string

	// Core execution
	Execute(ctx context.Context, input interface{}) (interface{}, error)

	// Schema validation
	InputSchema() map[string]interface{}
	OutputSchema() map[string]interface{}
	Validate(input interface{}) error

	// Metadata
	Category() string
	Tags() []string
	Dependencies() []string

	// Lifecycle
	Initialize(ctx context.Context, config map[string]interface{}) error
	Shutdown(ctx context.Context) error
}

// PonchoFlow defines the interface for all workflows in the framework
type PonchoFlow interface {
	// Identity
	Name() string
	Description() string
	Version() string

	// Core execution
	Execute(ctx context.Context, input interface{}) (interface{}, error)
	ExecuteStreaming(ctx context.Context, input interface{}, callback PonchoStreamCallback) error

	// Schema validation
	InputSchema() map[string]interface{}
	OutputSchema() map[string]interface{}

	// Metadata
	Category() string
	Tags() []string
	Dependencies() []string

	// Lifecycle
	Initialize(ctx context.Context, config map[string]interface{}) error
	Shutdown(ctx context.Context) error
}

// PonchoStreamCallback is the callback function for streaming responses
type PonchoStreamCallback func(chunk *PonchoStreamChunk) error

// Registry interfaces

// PonchoModelRegistry defines the interface for model registry
type PonchoModelRegistry interface {
	Register(name string, model PonchoModel) error
	Get(name string) (PonchoModel, error)
	List() []string
	Unregister(name string) error
	Clear() error
}

// PonchoToolRegistry defines the interface for tool registry
type PonchoToolRegistry interface {
	Register(name string, tool PonchoTool) error
	Get(name string) (PonchoTool, error)
	List() []string
	ListByCategory(category string) []string
	Unregister(name string) error
	Clear() error
}

// PonchoFlowRegistry defines the interface for flow registry
type PonchoFlowRegistry interface {
	Register(name string, flow PonchoFlow) error
	Get(name string) (PonchoFlow, error)
	List() []string
	ListByCategory(category string) []string
	Unregister(name string) error
	Clear() error
	ValidateDependencies(flow PonchoFlow, modelRegistry PonchoModelRegistry, toolRegistry PonchoToolRegistry) error
}

// PonchoFramework defines the main framework interface
type PonchoFramework interface {
	// Lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Component registration
	RegisterModel(name string, model PonchoModel) error
	RegisterTool(name string, tool PonchoTool) error
	RegisterFlow(name string, flow PonchoFlow) error

	// Core operations
	Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
	GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error
	ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error)
	ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error)
	ExecuteFlowStreaming(ctx context.Context, flowName string, input interface{}, callback PonchoStreamCallback) error

	// Registry access
	GetModelRegistry() PonchoModelRegistry
	GetToolRegistry() PonchoToolRegistry
	GetFlowRegistry() PonchoFlowRegistry

	// Configuration
	GetConfig() *PonchoFrameworkConfig
	ReloadConfig(ctx context.Context) error

	// Health and status
	Health(ctx context.Context) (*PonchoHealthStatus, error)
	Metrics(ctx context.Context) (*PonchoMetrics, error)
}

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// LogLevel represents logging level
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

// DefaultLogger is the thread-safe implementation of Logger interface
type DefaultLogger struct {
	level  LogLevel
	output io.Writer
	mutex  sync.Mutex
}

// NewDefaultLogger creates a new thread-safe default logger
func NewDefaultLogger() Logger {
	return &DefaultLogger{
		level:  LogLevelInfo,
		output: os.Stdout,
	}
}

// NewDefaultLoggerWithOutput creates a new default logger with custom output
func NewDefaultLoggerWithOutput(output io.Writer) Logger {
	return &DefaultLogger{
		level:  LogLevelInfo,
		output: output,
	}
}

// NewDefaultLoggerWithLevel creates a new default logger with custom level
func NewDefaultLoggerWithLevel(level LogLevel) Logger {
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
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if level < l.level {
		return
	}

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
func NewJSONLogger() Logger {
	return &JSONLogger{
		level:  LogLevelInfo,
		output: os.Stdout,
	}
}

// NewJSONLoggerWithOutput creates a new JSON logger with custom output
func NewJSONLoggerWithOutput(output io.Writer) Logger {
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
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if level < l.level {
		return
	}

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
func NewMultiLogger(loggers ...Logger) Logger {
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

// NoOpLogger is a logger that does nothing (for testing)
type NoOpLogger struct{}

// NewNoOpLogger creates a no-op logger
func NewNoOpLogger() Logger {
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

// Validator interface for schema validation
type Validator interface {
	Validate(data interface{}, schema map[string]interface{}) error
	SetSchema(schema map[string]interface{}) error
	GetSchema() map[string]interface{}
}

// ConfigLoader interface for configuration loading
type ConfigLoader interface {
	Load(path string) (*PonchoFrameworkConfig, error)
	LoadFromString(content string) (*PonchoFrameworkConfig, error)
	Validate(config *PonchoFrameworkConfig) error
}

// MetricsCollector interface for metrics collection
type MetricsCollector interface {
	RecordGeneration(model string, duration int64, tokens int, success bool)
	RecordToolExecution(tool string, duration int64, success bool)
	RecordFlowExecution(flow string, duration int64, success bool)
	RecordError(component string, errorType string)
	GetMetrics() *PonchoMetrics
	Reset()
}