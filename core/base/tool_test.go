package base

import (
	"context"
	"fmt"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MockTool реализация PonchoTool для тестов
type MockTool struct {
	*PonchoBaseTool
	initialized bool
}

func NewMockTool(name, description string) *MockTool {
	return &MockTool{
		PonchoBaseTool: NewPonchoBaseTool(name, description, "1.0.0", ""),
		initialized:    false,
	}
}

func (t *MockTool) Initialize(ctx context.Context, config map[string]interface{}) error {
	err := t.PonchoBaseTool.Initialize(ctx, config)
	if err == nil {
		t.initialized = true
	}
	return err
}

func (t *MockTool) Shutdown(ctx context.Context) error {
	err := t.PonchoBaseTool.Shutdown(ctx)
	if err == nil {
		t.initialized = false
	}
	return err
}

func (t *MockTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	if !t.initialized {
		return nil, fmt.Errorf("tool not initialized")
	}

	// Валидируем ввод используя базовый метод
	if err := t.Validate(input); err != nil {
		return nil, err
	}

	// Простая mock реализация - возвращаем input с префиксом
	if str, ok := input.(string); ok {
		return fmt.Sprintf("processed: %s", str), nil
	}

	if inputMap, ok := input.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		for k, v := range inputMap {
			result[k] = fmt.Sprintf("processed_%v", v)
		}
		return result, nil
	}

	return fmt.Sprintf("processed: %v", input), nil
}

func (t *MockTool) Validate(input interface{}) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	return nil
}

func (t *MockTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": "Input string to process",
	}
}

func (t *MockTool) OutputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": "Processed string output",
	}
}

func TestPonchoBaseTool(t *testing.T) {
	tool := NewMockTool("test-tool", "Test tool for unit testing")

	// Тест базовых методов
	if tool.Name() != "test-tool" {
		t.Errorf("Expected name 'test-tool', got '%s'", tool.Name())
	}

	if tool.Description() != "Test tool for unit testing" {
		t.Errorf("Expected description 'Test tool for unit testing', got '%s'", tool.Description())
	}

	if tool.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", tool.Version())
	}

	// Тест метаданных по умолчанию
	if tool.Category() != "" {
		t.Errorf("Expected empty category, got '%s'", tool.Category())
	}

	if len(tool.Tags()) != 0 {
		t.Errorf("Expected no tags, got %v", tool.Tags())
	}

	if len(tool.Dependencies()) != 0 {
		t.Errorf("Expected no dependencies, got %v", tool.Dependencies())
	}
}

func TestToolLifecycle(t *testing.T) {
	ctx := context.Background()
	tool := NewMockTool("test-tool", "Test tool")

	// Тест инициализации
	err := tool.Initialize(ctx, map[string]interface{}{
		"timeout": 30,
		"retries": 3,
	})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if !tool.initialized {
		t.Error("Tool should be initialized")
	}

	// Тест выполнения
	result, err := tool.Execute(ctx, "test input")
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if result == nil {
		t.Error("Result should not be nil")
	}

	expected := "processed: test input"
	if result != expected {
		t.Errorf("Expected result '%s', got '%v'", expected, result)
	}

	// Тест с map input
	mapInput := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	result, err = tool.Execute(ctx, mapInput)
	if err != nil {
		t.Errorf("Execute failed with map input: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Result should be a map")
	}

	if resultMap["key1"] != "processed_value1" {
		t.Errorf("Expected 'processed_value1', got '%v'", resultMap["key1"])
	}

	// Тест shutdown
	err = tool.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if tool.initialized {
		t.Error("Tool should not be initialized after shutdown")
	}
}

func TestToolNotInitialized(t *testing.T) {
	ctx := context.Background()
	tool := NewMockTool("test-tool", "Test tool")

	// Тест выполнения без инициализации
	_, err := tool.Execute(ctx, "test input")
	if err == nil {
		t.Error("Execute should fail when tool is not initialized")
	}
}

func TestToolValidation(t *testing.T) {
	ctx := context.Background()
	tool := NewMockTool("test-tool", "Test tool")

	err := tool.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Тест валидации nil input
	err = tool.Validate(nil)
	if err == nil {
		t.Error("Validate should fail for nil input")
	}

	// Тест валидации валидного input
	err = tool.Validate("valid input")
	if err != nil {
		t.Errorf("Validate should pass for valid input: %v", err)
	}

	// Тест выполнения с невалидным input
	_, err = tool.Execute(ctx, nil)
	if err == nil {
		t.Error("Execute should fail for nil input")
	}
}

func TestToolSchemas(t *testing.T) {
	tool := NewMockTool("test-tool", "Test tool")

	// Тест input schema
	inputSchema := tool.InputSchema()
	if inputSchema == nil {
		t.Error("Input schema should not be nil")
	}

	if inputSchema["type"] != "string" {
		t.Errorf("Expected input schema type 'string', got '%v'", inputSchema["type"])
	}

	// Тест output schema
	outputSchema := tool.OutputSchema()
	if outputSchema == nil {
		t.Error("Output schema should not be nil")
	}

	if outputSchema["type"] != "string" {
		t.Errorf("Expected output schema type 'string', got '%v'", outputSchema["type"])
	}
}

func TestToolMetadata(t *testing.T) {
	tool := NewMockTool("test-tool", "Test tool")

	// Тест установки метаданных
	tool.SetCategory("test-category")
	if tool.Category() != "test-category" {
		t.Errorf("Expected category 'test-category', got '%s'", tool.Category())
	}

	tool.ClearTags()
	tool.AddTag("tag1")
	tool.AddTag("tag2")
	tool.AddTag("tag3")
	tags := tool.Tags()
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}

	if tags[0] != "tag1" || tags[1] != "tag2" || tags[2] != "tag3" {
		t.Errorf("Expected tags [tag1, tag2, tag3], got %v", tags)
	}

	tool.ClearDependencies()
	tool.AddDependency("dep1")
	tool.AddDependency("dep2")
	deps := tool.Dependencies()
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	if deps[0] != "dep1" || deps[1] != "dep2" {
		t.Errorf("Expected dependencies [dep1, dep2], got %v", deps)
	}
}

func TestToolConfiguration(t *testing.T) {
	ctx := context.Background()
	tool := NewMockTool("test-tool", "Test tool")

	// Тест с конфигурацией
	config := map[string]interface{}{
		"timeout":    60,
		"retries":    5,
		"custom_key": "custom_value",
	}

	err := tool.Initialize(ctx, config)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Тест получения конфигурации
	value, exists := tool.GetConfig("timeout")
	if !exists {
		t.Error("timeout should exist in config")
	}

	if value != 60 {
		t.Errorf("Expected timeout 60, got %v", value)
	}

	value, exists = tool.GetConfig("custom_key")
	if !exists {
		t.Error("custom_key should exist in config")
	}

	if value != "custom_value" {
		t.Errorf("Expected custom_key 'custom_value', got %v", value)
	}

	// Тест установки конфигурации
	tool.SetConfig("new_key", "new_value")
	value, exists = tool.GetConfig("new_key")
	if !exists {
		t.Error("new_key should exist in config")
	}

	if value != "new_value" {
		t.Errorf("Expected new_key 'new_value', got %v", value)
	}

	// Тест получения всей конфигурации
	allConfig := tool.GetAllConfig()
	if len(allConfig) == 0 {
		t.Error("All config should not be empty")
	}

	if len(allConfig) < 3 {
		t.Errorf("Expected at least 3 config items, got %d", len(allConfig))
	}
}

func TestToolLogger(t *testing.T) {
	ctx := context.Background()
	tool := NewMockTool("test-tool", "Test tool")

	// Тест установки логгера
	mockLogger := interfaces.NewDefaultLogger()
	tool.SetLogger(mockLogger)

	if tool.GetLogger() != mockLogger {
		t.Error("Logger should be set correctly")
	}

	// Тест инициализации с новым логгером
	err := tool.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Логгер должен использоваться для логирования
	tool.GetLogger().Info("Test message")
}

// Бенчмарки
func BenchmarkToolExecute(b *testing.B) {
	ctx := context.Background()
	tool := NewMockTool("benchmark-tool", "Benchmark tool")

	err := tool.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		b.Fatalf("Initialize failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tool.Execute(ctx, "benchmark test input")
		if err != nil {
			b.Errorf("Execute failed: %v", err)
		}
	}
}

func BenchmarkToolValidate(b *testing.B) {
	tool := NewMockTool("benchmark-tool", "Benchmark tool")

	input := "benchmark test input"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := tool.Validate(input)
		if err != nil {
			b.Errorf("Validate failed: %v", err)
		}
	}
}
