package core

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger()
	if logger == nil {
		t.Error("NewDefaultLogger() returned nil")
	}

	// Проверяем уровень по умолчанию
	if logger.GetLevel() != LogLevelInfo {
		t.Errorf("Expected default level LogLevelInfo, got %v", logger.GetLevel())
	}
}

func TestNewDefaultLoggerWithOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLoggerWithOutput(&buf)
	if logger == nil {
		t.Error("NewDefaultLoggerWithOutput() returned nil")
	}

	logger.Info("test message")
	if buf.Len() == 0 {
		t.Error("Logger didn't write to custom output")
	}
}

func TestNewDefaultLoggerWithLevel(t *testing.T) {
	logger := NewDefaultLoggerWithLevel(LogLevelDebug)
	if logger == nil {
		t.Error("NewDefaultLoggerWithLevel() returned nil")
	}

	if logger.GetLevel() != LogLevelDebug {
		t.Errorf("Expected LogLevelDebug, got %v", logger.GetLevel())
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLoggerWithOutput(&buf)
	logger.SetLevel(LogLevelWarn)

	// Тестируем, что только warn и error сообщения выводятся
	logger.Debug("debug message") // Не должно выводиться
	logger.Info("info message")   // Не должно выводиться
	logger.Warn("warn message")   // Должно выводиться
	logger.Error("error message") // Должно выводиться

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not be logged at Warn level")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should not be logged at Warn level")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be logged at Warn level")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should be logged at Warn level")
	}
}

func TestLoggerFormats(t *testing.T) {
	tests := []struct {
		name     string
		logger   Logger
		expected string
	}{
		{
			name:     "default logger",
			logger:   NewDefaultLogger(),
			expected: "INFO",
		},
		{
			name:     "JSON logger",
			logger:   NewJSONLogger(),
			expected: `"level":"INFO"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			// Устанавливаем output в зависимости от типа логгера
			switch l := tt.logger.(type) {
			case *DefaultLogger:
				l.SetOutput(&buf)
			case *JSONLogger:
				l.SetOutput(&buf)
			}

			tt.logger.Info("test message")

			if !strings.Contains(buf.String(), tt.expected) {
				t.Errorf("Expected output to contain %s, got: %s", tt.expected, buf.String())
			}
		})
	}
}

func TestLoggerSetLevel(t *testing.T) {
	logger := NewDefaultLogger()

	// Меняем уровень на Debug
	logger.SetLevel(LogLevelDebug)

	if logger.GetLevel() != LogLevelDebug {
		t.Errorf("Expected LogLevelDebug, got %v", logger.GetLevel())
	}
}

func TestLoggerSetOutput(t *testing.T) {
	logger := NewDefaultLogger()

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Info("test message")

	if buf.Len() == 0 {
		t.Error("Message should be written to custom output")
	}
}

func TestJSONLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLoggerWithOutput(&buf)

	logger.Info("test message", "key1", "value1", "key2", "value2")

	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Error("JSON output should contain level")
	}
	if !strings.Contains(output, `"message":"test message"`) {
		t.Error("JSON output should contain message")
	}
	if !strings.Contains(output, `"key1":"value1"`) {
		t.Error("JSON output should contain key1 field")
	}
	if !strings.Contains(output, `"key2":"value2"`) {
		t.Error("JSON output should contain key2 field")
	}
}

func TestMultiLogger(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger1 := NewDefaultLoggerWithOutput(&buf1)
	logger2 := NewDefaultLoggerWithOutput(&buf2)

	multiLogger := NewMultiLogger(logger1, logger2)

	multiLogger.Info("test message")

	if buf1.Len() == 0 {
		t.Error("First logger should receive the message")
	}
	if buf2.Len() == 0 {
		t.Error("Second logger should receive the message")
	}
	if !strings.Contains(buf1.String(), "test message") {
		t.Error("First logger output should contain the message")
	}
	if !strings.Contains(buf2.String(), "test message") {
		t.Error("Second logger output should contain the message")
	}
}

func TestMultiLoggerAddRemove(t *testing.T) {
	var buf bytes.Buffer
	logger1 := NewDefaultLogger()
	logger2 := NewDefaultLoggerWithOutput(&buf)

	multiLogger := NewMultiLogger(logger1)

	// Добавляем второй логгер
	multiLogger.AddLogger(logger2)
	multiLogger.Info("test message")

	if buf.Len() == 0 {
		t.Error("Added logger should receive the message")
	}

	// Удаляем второй логгер
	buf.Reset()
	multiLogger.RemoveLogger(logger2)
	multiLogger.Info("another message")

	if buf.Len() != 0 {
		t.Error("Removed logger should not receive the message")
	}
}

func TestNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()
	if logger == nil {
		t.Error("NewNoOpLogger() returned nil")
	}

	// NoOpLogger не должен ничего выводить и не должен паниковать
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	// Если мы дошли сюда, значит все хорошо
}

func TestCreateLoggerFromConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *interfaces.LoggingConfig
		expect string
	}{
		{
			name:   "nil config",
			config: nil,
			expect: "INFO", // Default logger
		},
		{
			name: "text format",
			config: &interfaces.LoggingConfig{
				Level:  "info",
				Format: "text",
			},
			expect: "INFO",
		},
		{
			name: "json format",
			config: &interfaces.LoggingConfig{
				Level:  "info",
				Format: "json",
			},
			expect: `"level":"INFO"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := CreateLoggerFromConfig(tt.config)

			// Устанавливаем output для теста
			switch l := logger.(type) {
			case *DefaultLogger:
				l.SetOutput(&buf)
			case *JSONLogger:
				l.SetOutput(&buf)
			}

			logger.Info("test message")

			if !strings.Contains(buf.String(), tt.expect) {
				t.Errorf("Expected output to contain %s, got: %s", tt.expect, buf.String())
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"info", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"warning", LogLevelWarn},
		{"error", LogLevelError},
		{"invalid", LogLevelInfo}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Бенчмарки для производительности
func BenchmarkDefaultLoggerInfo(b *testing.B) {
	logger := NewDefaultLogger()
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark test message")
	}
}

func BenchmarkJSONLoggerInfo(b *testing.B) {
	logger := NewJSONLogger()
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark test message")
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	logger := NewDefaultLogger()
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark test message", "key1", "value1", "key2", "value2")
	}
}
