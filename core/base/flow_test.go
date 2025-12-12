package base

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MockFlow реализация PonchoFlow для тестов
type MockFlow struct {
	*PonchoBaseFlow
	initialized bool
}

func NewMockFlow(name, description string) *MockFlow {
	return &MockFlow{
		PonchoBaseFlow: NewPonchoBaseFlow(name, description, "1.0.0", ""),
		initialized:    false,
	}
}

func (f *MockFlow) Initialize(ctx context.Context, config map[string]interface{}) error {
	err := f.PonchoBaseFlow.Initialize(ctx, config)
	if err == nil {
		f.initialized = true
	}
	return err
}

func (f *MockFlow) Shutdown(ctx context.Context) error {
	err := f.PonchoBaseFlow.Shutdown(ctx)
	if err == nil {
		f.initialized = false
	}
	return err
}

func (f *MockFlow) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	if !f.initialized {
		return nil, fmt.Errorf("flow not initialized")
	}

	// Валидируем ввод используя базовый метод
	if err := f.Validate(input); err != nil {
		return nil, err
	}

	// Простая mock реализация - возвращаем input с префиксом
	if str, ok := input.(string); ok {
		return fmt.Sprintf("flow_processed: %s", str), nil
	}

	if inputMap, ok := input.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		for k, v := range inputMap {
			result[k] = fmt.Sprintf("flow_processed_%v", v)
		}
		return result, nil
	}

	return fmt.Sprintf("flow_processed: %v", input), nil
}

func (f *MockFlow) ExecuteStreaming(ctx context.Context, input interface{}, callback interfaces.PonchoStreamCallback) error {
	if !f.initialized {
		return fmt.Errorf("flow not initialized")
	}

	// Валидируем ввод используя базовый метод
	if err := f.Validate(input); err != nil {
		return err
	}

	// Симулируем стриминг с несколькими чанками
	chunks := []string{"Flow", " ", "processing", " ", "result"}

	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			delta := &interfaces.PonchoMessage{
				Role: interfaces.PonchoRoleAssistant,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: chunk,
					},
				},
			}

			streamChunk := &interfaces.PonchoStreamChunk{
				Delta:    delta,
				Done:     false,
				Metadata: make(map[string]interface{}),
			}
			err := callback(streamChunk)
			if err != nil {
				return err
			}
			time.Sleep(2 * time.Millisecond) // Симуляция задержки

			f.GetLogger().Debug("Flow stream chunk sent", "chunk", chunk, "index", i)
		}
	}

	// Финальный чанк
	finalDelta := &interfaces.PonchoMessage{
		Role:    interfaces.PonchoRoleAssistant,
		Content: []*interfaces.PonchoContentPart{},
	}

	finalChunk := &interfaces.PonchoStreamChunk{
		Delta:        finalDelta,
		Usage:        &interfaces.PonchoUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		FinishReason: interfaces.PonchoFinishReasonStop,
		Done:         true,
		Metadata:     make(map[string]interface{}),
	}
	return callback(finalChunk)
}

func (f *MockFlow) Validate(input interface{}) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	return nil
}

func (f *MockFlow) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": "Input string to process",
	}
}

func (f *MockFlow) OutputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": "Processed string output",
	}
}

func TestPonchoBaseFlow(t *testing.T) {
	flow := NewMockFlow("test-flow", "Test flow for unit testing")

	// Тест базовых методов
	if flow.Name() != "test-flow" {
		t.Errorf("Expected name 'test-flow', got '%s'", flow.Name())
	}

	if flow.Description() != "Test flow for unit testing" {
		t.Errorf("Expected description 'Test flow for unit testing', got '%s'", flow.Description())
	}

	if flow.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", flow.Version())
	}

	// Тест метаданных по умолчанию
	if flow.Category() != "" {
		t.Errorf("Expected empty category, got '%s'", flow.Category())
	}

	if len(flow.Tags()) != 0 {
		t.Errorf("Expected no tags, got %v", flow.Tags())
	}

	if len(flow.Dependencies()) != 0 {
		t.Errorf("Expected no dependencies, got %v", flow.Dependencies())
	}
}

func TestFlowLifecycle(t *testing.T) {
	ctx := context.Background()
	flow := NewMockFlow("test-flow", "Test flow")

	// Тест инициализации
	err := flow.Initialize(ctx, map[string]interface{}{
		"timeout": 30,
		"retries": 3,
	})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if !flow.initialized {
		t.Error("Flow should be initialized")
	}

	// Тест выполнения
	result, err := flow.Execute(ctx, "test input")
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if result == nil {
		t.Error("Result should not be nil")
	}

	expected := "flow_processed: test input"
	if result != expected {
		t.Errorf("Expected result '%s', got '%v'", expected, result)
	}

	// Тест с map input
	mapInput := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	result, err = flow.Execute(ctx, mapInput)
	if err != nil {
		t.Errorf("Execute failed with map input: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Result should be a map")
	}

	if resultMap["key1"] != "flow_processed_value1" {
		t.Errorf("Expected 'flow_processed_value1', got '%v'", resultMap["key1"])
	}

	// Тест стриминга
	streamCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		if chunk == nil {
			t.Error("Chunk should not be nil")
		}
		return nil
	}

	err = flow.ExecuteStreaming(ctx, "test input", streamCallback)
	if err != nil {
		t.Errorf("ExecuteStreaming failed: %v", err)
	}

	// Тест shutdown
	err = flow.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if flow.initialized {
		t.Error("Flow should not be initialized after shutdown")
	}
}

func TestFlowNotInitialized(t *testing.T) {
	ctx := context.Background()
	flow := NewMockFlow("test-flow", "Test flow")

	// Тест выполнения без инициализации
	_, err := flow.Execute(ctx, "test input")
	if err == nil {
		t.Error("Execute should fail when flow is not initialized")
	}

	// Тест стриминга без инициализации
	streamCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		return nil
	}

	err = flow.ExecuteStreaming(ctx, "test input", streamCallback)
	if err == nil {
		t.Error("ExecuteStreaming should fail when flow is not initialized")
	}
}

func TestFlowValidation(t *testing.T) {
	ctx := context.Background()
	flow := NewMockFlow("test-flow", "Test flow")

	err := flow.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Тест валидации nil input
	err = flow.Validate(nil)
	if err == nil {
		t.Error("Validate should fail for nil input")
	}

	// Тест валидации валидного input
	err = flow.Validate("valid input")
	if err != nil {
		t.Errorf("Validate should pass for valid input: %v", err)
	}

	// Тест выполнения с невалидным input
	_, err = flow.Execute(ctx, nil)
	if err == nil {
		t.Error("Execute should fail for nil input")
	}
}

func TestFlowSchemas(t *testing.T) {
	flow := NewMockFlow("test-flow", "Test flow")

	// Тест input schema
	inputSchema := flow.InputSchema()
	if inputSchema == nil {
		t.Error("Input schema should not be nil")
	}

	if inputSchema["type"] != "string" {
		t.Errorf("Expected input schema type 'string', got '%v'", inputSchema["type"])
	}

	// Тест output schema
	outputSchema := flow.OutputSchema()
	if outputSchema == nil {
		t.Error("Output schema should not be nil")
	}

	if outputSchema["type"] != "string" {
		t.Errorf("Expected output schema type 'string', got '%v'", outputSchema["type"])
	}
}

func TestFlowMetadata(t *testing.T) {
	flow := NewMockFlow("test-flow", "Test flow")

	// Тест установки метаданных
	flow.SetCategory("test-category")
	if flow.Category() != "test-category" {
		t.Errorf("Expected category 'test-category', got '%s'", flow.Category())
	}

	flow.ClearTags()
	flow.AddTag("tag1")
	flow.AddTag("tag2")
	flow.AddTag("tag3")

	tags := flow.Tags()
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}

	if tags[0] != "tag1" || tags[1] != "tag2" || tags[2] != "tag3" {
		t.Errorf("Expected tags [tag1, tag2, tag3], got %v", tags)
	}

	flow.ClearDependencies()
	flow.AddDependency("dep1")
	flow.AddDependency("dep2")

	deps := flow.Dependencies()
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	if deps[0] != "dep1" || deps[1] != "dep2" {
		t.Errorf("Expected dependencies [dep1, dep2], got %v", deps)
	}
}

func TestFlowConfiguration(t *testing.T) {
	ctx := context.Background()
	flow := NewMockFlow("test-flow", "Test flow")

	// Тест с конфигурацией
	config := map[string]interface{}{
		"timeout":    60,
		"retries":    5,
		"custom_key": "custom_value",
	}

	err := flow.Initialize(ctx, config)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Тест получения конфигурации
	value, exists := flow.GetConfig("timeout")
	if !exists {
		t.Error("timeout should exist in config")
	}

	if value != 60 {
		t.Errorf("Expected timeout 60, got %v", value)
	}

	value, exists = flow.GetConfig("custom_key")
	if !exists {
		t.Error("custom_key should exist in config")
	}

	if value != "custom_value" {
		t.Errorf("Expected custom_key 'custom_value', got %v", value)
	}

	// Тест установки конфигурации
	flow.SetConfig("new_key", "new_value")
	value, exists = flow.GetConfig("new_key")
	if !exists {
		t.Error("new_key should exist in config")
	}

	if value != "new_value" {
		t.Errorf("Expected new_key 'new_value', got %v", value)
	}

	// Тест получения всей конфигурации
	allConfig := flow.GetAllConfig()
	if len(allConfig) == 0 {
		t.Error("All config should not be empty")
	}

	if len(allConfig) < 3 {
		t.Errorf("Expected at least 3 config items, got %d", len(allConfig))
	}
}

func TestFlowLogger(t *testing.T) {
	ctx := context.Background()
	flow := NewMockFlow("test-flow", "Test flow")

	// Тест установки логгера
	mockLogger := interfaces.NewDefaultLogger()
	flow.SetLogger(mockLogger)

	if flow.GetLogger() != mockLogger {
		t.Error("Logger should be set correctly")
	}

	// Тест инициализации с новым логгером
	err := flow.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Логгер должен использоваться для логирования
	flow.GetLogger().Info("Test message")
}

func TestFlowTagOperations(t *testing.T) {
	flow := NewMockFlow("test-flow", "Test flow")

	// Тест добавления тегов
	flow.AddTag("tag1")
	flow.AddTag("tag2")

	if !flow.HasTag("tag1") {
		t.Error("Should have tag1")
	}

	if !flow.HasTag("tag2") {
		t.Error("Should have tag2")
	}

	if flow.HasTag("nonexistent") {
		t.Error("Should not have nonexistent tag")
	}

	// Тест подсчета тегов
	if flow.GetTagCount() != 2 {
		t.Errorf("Expected 2 tags, got %d", flow.GetTagCount())
	}

	// Тест удаления тега
	flow.RemoveTag("tag1")

	if flow.HasTag("tag1") {
		t.Error("Should not have tag1 after removal")
	}

	if flow.GetTagCount() != 1 {
		t.Errorf("Expected 1 tag after removal, got %d", flow.GetTagCount())
	}

	// Тест очистки тегов
	flow.ClearTags()

	if flow.GetTagCount() != 0 {
		t.Errorf("Expected 0 tags after clear, got %d", flow.GetTagCount())
	}
}

func TestFlowDependencyOperations(t *testing.T) {
	flow := NewMockFlow("test-flow", "Test flow")

	// Тест добавления зависимостей
	flow.AddDependency("dep1")
	flow.AddDependency("dep2")

	if !flow.HasDependency("dep1") {
		t.Error("Should have dep1 dependency")
	}

	if !flow.HasDependency("dep2") {
		t.Error("Should have dep2 dependency")
	}

	if flow.HasDependency("nonexistent") {
		t.Error("Should not have nonexistent dependency")
	}

	// Тест подсчета зависимостей
	if flow.GetDependencyCount() != 2 {
		t.Errorf("Expected 2 dependencies, got %d", flow.GetDependencyCount())
	}

	// Тест удаления зависимости
	flow.RemoveDependency("dep1")

	if flow.HasDependency("dep1") {
		t.Error("Should not have dep1 dependency after removal")
	}

	if flow.GetDependencyCount() != 1 {
		t.Errorf("Expected 1 dependency after removal, got %d", flow.GetDependencyCount())
	}

	// Тест очистки зависимостей
	flow.ClearDependencies()

	if flow.GetDependencyCount() != 0 {
		t.Errorf("Expected 0 dependencies after clear, got %d", flow.GetDependencyCount())
	}
}

func TestFlowContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	flow := NewMockFlow("test-flow", "test-flow")

	err := flow.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	streamCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		return nil
	}

	// Должно вернуть ошибку из-за отмены контекста
	err = flow.ExecuteStreaming(ctx, "test input", streamCallback)
	if err == nil {
		t.Error("ExecuteStreaming should fail due to context cancellation")
	}
}

// Бенчмарки
func BenchmarkFlowExecute(b *testing.B) {
	ctx := context.Background()
	flow := NewMockFlow("benchmark-flow", "Benchmark flow")

	err := flow.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		b.Fatalf("Initialize failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := flow.Execute(ctx, "benchmark test input")
		if err != nil {
			b.Errorf("Execute failed: %v", err)
		}
	}
}

func BenchmarkFlowExecuteStreaming(b *testing.B) {
	ctx := context.Background()
	flow := NewMockFlow("benchmark-flow", "Benchmark flow")

	err := flow.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		b.Fatalf("Initialize failed: %v", err)
	}

	streamCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := flow.ExecuteStreaming(ctx, "benchmark test input", streamCallback)
		if err != nil {
			b.Errorf("ExecuteStreaming failed: %v", err)
		}
	}
}
