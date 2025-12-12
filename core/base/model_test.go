package base

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MockModel реализация PonchoModel для тестов
type MockModel struct {
	*PonchoBaseModel
	initialized bool
}

func NewMockModel(name, provider string) *MockModel {
	capabilities := interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    false,
		System:    true,
	}

	return &MockModel{
		PonchoBaseModel: NewPonchoBaseModel(name, provider, capabilities),
		initialized:     false,
	}
}

func (m *MockModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	err := m.PonchoBaseModel.Initialize(ctx, config)
	if err == nil {
		m.initialized = true
	}
	return err
}

func (m *MockModel) Shutdown(ctx context.Context) error {
	err := m.PonchoBaseModel.Shutdown(ctx)
	if err == nil {
		m.initialized = false
	}
	return err
}

func (m *MockModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model not initialized")
	}

	// Валидируем запрос используя базовый метод
	if err := m.ValidateRequest(req); err != nil {
		return nil, err
	}

	return &interfaces.PonchoModelResponse{
		Message: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "Mock response",
				},
			},
		},
		Usage: m.PrepareUsage(10, 5),
		Metadata: map[string]interface{}{
			"model":     m.Name(),
			"provider":  m.Provider(),
			"processed": true,
		},
	}, nil
}

func (m *MockModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	if !m.initialized {
		return fmt.Errorf("model not initialized")
	}

	if !m.SupportsStreaming() {
		return fmt.Errorf("streaming not supported")
	}

	// Валидируем запрос используя базовый метод
	if err := m.ValidateRequest(req); err != nil {
		return err
	}

	// Симулируем стриминг с несколькими чанками
	chunks := []string{"Hello", " ", "world", "!"}

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

			err := callback(m.PrepareStreamChunk(delta, nil, "", false))
			if err != nil {
				return err
			}
			time.Sleep(2 * time.Millisecond) // Симуляция задержки

			m.GetLogger().Debug("Stream chunk sent", "chunk", chunk, "index", i)
		}
	}

	// Финальный чанк
	finalDelta := &interfaces.PonchoMessage{
		Role:    interfaces.PonchoRoleAssistant,
		Content: []*interfaces.PonchoContentPart{},
	}

	return callback(m.PrepareStreamChunk(finalDelta, m.PrepareUsage(10, 5), interfaces.PonchoFinishReasonStop, true))
}

func TestPonchoBaseModel(t *testing.T) {
	model := NewMockModel("test-model", "test-provider")

	// Тест базовых методов
	if model.Name() != "test-model" {
		t.Errorf("Expected name 'test-model', got '%s'", model.Name())
	}

	if model.Provider() != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got '%s'", model.Provider())
	}

	if model.MaxTokens() != 4000 { // Default from PonchoBaseModel
		t.Errorf("Expected max tokens 4000, got %d", model.MaxTokens())
	}

	if model.DefaultTemperature() != 0.7 { // Default from PonchoBaseModel
		t.Errorf("Expected temperature 0.7, got %f", model.DefaultTemperature())
	}

	// Тест capabilities
	if !model.SupportsStreaming() {
		t.Error("Expected streaming support")
	}

	if !model.SupportsTools() {
		t.Error("Expected tools support")
	}

	if model.SupportsVision() {
		t.Error("Expected no vision support")
	}

	if !model.SupportsSystemRole() {
		t.Error("Expected system role support")
	}
}

func TestModelLifecycle(t *testing.T) {
	ctx := context.Background()
	model := NewMockModel("test-model", "test-provider")

	// Тест инициализации
	err := model.Initialize(ctx, map[string]interface{}{
		"api_key": "test-key",
		"timeout": 30,
	})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if !model.initialized {
		t.Error("Model should be initialized")
	}

	// Тест генерации
	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	resp, err := model.Generate(ctx, req)
	if err != nil {
		t.Errorf("Generate failed: %v", err)
	}

	if resp == nil {
		t.Error("Response should not be nil")
	}

	if resp.Message == nil {
		t.Error("Response message should not be nil")
	}

	if resp.Usage == nil {
		t.Error("Response usage should not be nil")
	}

	// Тест стриминга
	streamCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		if chunk == nil {
			t.Error("Chunk should not be nil")
		}
		return nil
	}

	err = model.GenerateStreaming(ctx, req, streamCallback)
	if err != nil {
		t.Errorf("GenerateStreaming failed: %v", err)
	}

	// Тест shutdown
	err = model.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if model.initialized {
		t.Error("Model should not be initialized after shutdown")
	}
}

func TestModelNotInitialized(t *testing.T) {
	ctx := context.Background()
	model := NewMockModel("test-model", "test-provider")

	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello",
					},
				},
			},
		},
	}

	// Тест генерации без инициализации
	_, err := model.Generate(ctx, req)
	if err == nil {
		t.Error("Generate should fail when model is not initialized")
	}

	// Тест стриминга без инициализации
	streamCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		return nil
	}

	err = model.GenerateStreaming(ctx, req, streamCallback)
	if err == nil {
		t.Error("GenerateStreaming should fail when model is not initialized")
	}
}

func TestModelStreamingNotSupported(t *testing.T) {
	ctx := context.Background()
	model := NewMockModel("test-model", "test-provider")

	// Отключаем поддержку стриминга
	model.SetCapabilities(interfaces.ModelCapabilities{
		Streaming: false,
		Tools:     true,
		Vision:    false,
		System:    true,
	})

	err := model.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello",
					},
				},
			},
		},
	}

	streamCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		return nil
	}

	err = model.GenerateStreaming(ctx, req, streamCallback)
	if err == nil {
		t.Error("GenerateStreaming should fail when streaming is not supported")
	}
}

func TestModelWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	model := NewMockModel("test-model", "test-provider")

	err := model.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello",
					},
				},
			},
		},
	}

	streamCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		return nil
	}

	// Должно вернуть ошибку из-за отмены контекста
	err = model.GenerateStreaming(ctx, req, streamCallback)
	if err == nil {
		t.Error("GenerateStreaming should fail due to context cancellation")
	}
}

func TestModelConfiguration(t *testing.T) {
	ctx := context.Background()
	model := NewMockModel("test-model", "test-provider")

	// Тест с конфигурацией
	config := map[string]interface{}{
		"api_key":     "test-key",
		"max_tokens":  2000,
		"temperature": float32(0.5),
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Проверяем, что конфигурация применилась
	if model.MaxTokens() != 2000 {
		t.Errorf("Expected max tokens 2000, got %d", model.MaxTokens())
	}

	if model.DefaultTemperature() != 0.5 {
		t.Errorf("Expected temperature 0.5, got %f", model.DefaultTemperature())
	}

	// Тест получения конфигурации
	value, exists := model.GetConfig("api_key")
	if !exists {
		t.Error("api_key should exist in config")
	}

	if value != "test-key" {
		t.Errorf("Expected api_key 'test-key', got %v", value)
	}

	// Тест установки конфигурации
	model.SetConfig("new_key", "new_value")
	value, exists = model.GetConfig("new_key")
	if !exists {
		t.Error("new_key should exist in config")
	}

	if value != "new_value" {
		t.Errorf("Expected new_key 'new_value', got %v", value)
	}

	// Тест получения всей конфигурации
	allConfig := model.GetAllConfig()
	if len(allConfig) == 0 {
		t.Error("All config should not be empty")
	}
}

func TestModelValidation(t *testing.T) {
	ctx := context.Background()
	model := NewMockModel("test-model", "test-provider")

	err := model.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Тест валидации пустого запроса
	err = model.ValidateRequest(nil)
	if err == nil {
		t.Error("ValidateRequest should fail for nil request")
	}

	// Тест валидации запроса без сообщений
	req := &interfaces.PonchoModelRequest{
		Model:    "test-model",
		Messages: []*interfaces.PonchoMessage{},
	}

	err = model.ValidateRequest(req)
	if err == nil {
		t.Error("ValidateRequest should fail for empty messages")
	}

	// Тест валидации запроса с невалидными токенами
	req = &interfaces.PonchoModelRequest{
		Model: "test-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Hello",
					},
				},
			},
		},
		MaxTokens: new(int), // Указатель на 0
	}

	err = model.ValidateRequest(req)
	if err == nil {
		t.Error("ValidateRequest should fail for max_tokens = 0")
	}

	// Тест валидации запроса с превышением лимита токенов
	maxTokens := model.MaxTokens() + 1000
	req.MaxTokens = &maxTokens

	err = model.ValidateRequest(req)
	if err == nil {
		t.Error("ValidateRequest should fail for max_tokens > model limit")
	}
}

// Бенчмарки
func BenchmarkModelGenerate(b *testing.B) {
	ctx := context.Background()
	model := NewMockModel("benchmark-model", "benchmark-provider")

	err := model.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		b.Fatalf("Initialize failed: %v", err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "benchmark-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Benchmark test message",
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.Generate(ctx, req)
		if err != nil {
			b.Errorf("Generate failed: %v", err)
		}
	}
}

func BenchmarkModelGenerateStreaming(b *testing.B) {
	ctx := context.Background()
	model := NewMockModel("benchmark-model", "benchmark-provider")

	err := model.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		b.Fatalf("Initialize failed: %v", err)
	}

	req := &interfaces.PonchoModelRequest{
		Model: "benchmark-model",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Benchmark streaming test",
					},
				},
			},
		},
	}

	streamCallback := func(chunk *interfaces.PonchoStreamChunk) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := model.GenerateStreaming(ctx, req, streamCallback)
		if err != nil {
			b.Errorf("GenerateStreaming failed: %v", err)
		}
	}
}
