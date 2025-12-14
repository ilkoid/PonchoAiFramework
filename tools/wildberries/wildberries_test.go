package wildberries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MockModelRegistry for testing
type MockModelRegistry struct {
	mock.Mock
}

func (m *MockModelRegistry) Register(name string, model interfaces.PonchoModel) error {
	args := m.Called(name, model)
	return args.Error(0)
}

func (m *MockModelRegistry) Get(name string) (interfaces.PonchoModel, error) {
	args := m.Called(name)
	return args.Get(0).(interfaces.PonchoModel), args.Error(1)
}

func (m *MockModelRegistry) List() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockModelRegistry) ListByCategory(category string) []string {
	args := m.Called(category)
	return args.Get(0).([]string)
}

func (m *MockModelRegistry) Unregister(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockModelRegistry) Clear() error {
	args := m.Called()
	return args.Error(0)
}

// MockLogger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(5, 100*time.Millisecond)
	ctx := context.Background()

	// Should be able to consume all tokens
	for i := 0; i < 5; i++ {
		err := rl.WaitForToken(ctx)
		assert.NoError(t, err)
	}

	// Next token should require waiting
	start := time.Now()
	err := rl.WaitForToken(ctx)
	assert.NoError(t, err)

	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 90*time.Millisecond) // Allow some tolerance
}

func TestNewWBClient(t *testing.T) {
	logger := &MockLogger{}

	// Test with custom base URL
	client := NewWBClient("test-api-key", "https://test.wildberries.ru", logger)
	assert.NotNil(t, client)
	assert.Equal(t, "test-api-key", client.apiKey)
	assert.Equal(t, "https://test.wildberries.ru", client.baseURL)
	assert.NotNil(t, client.rateLimiter)

	// Test with default base URL
	client2 := NewWBClient("test-api-key", "", logger)
	assert.Equal(t, "https://content-api.wildberries.ru", client2.baseURL)
}

func TestWBError(t *testing.T) {
	err := &WBError{
		Code:    429,
		Message: "Rate limit exceeded",
		Details: "Wait 2 seconds before retrying",
	}

	expected := "WB API Error 429: Rate limit exceeded - Wait 2 seconds before retrying"
	assert.Equal(t, expected, err.Error())

	// Test error without details
	err2 := &WBError{
		Code:    404,
		Message: "Not found",
	}

	expected2 := "WB API Error 404: Not found"
	assert.Equal(t, expected2, err2.Error())
}

func TestFashionAnalysis(t *testing.T) {
	analysis := &FashionAnalysis{
		Type:        "платье",
		Style:       "летний",
		Season:      "лето",
		Gender:      "женский",
		Material:    "хлопок",
		Color:       "синий",
		Name:        "Летнее хлопковое платье",
		Description: "Удобное летнее платье из натурального хлопка",
		Tags:        []string{"платье", "лето", "хлопок"},
		Confidence:  0.95,
		SubjectID:   105,
		Subject:     "Платья",
	}

	assert.Equal(t, "платье", analysis.Type)
	assert.Equal(t, "летний", analysis.Style)
	assert.Equal(t, 105, analysis.SubjectID)
	assert.Equal(t, "Платья", analysis.Subject)
	assert.Len(t, analysis.Tags, 3)
	assert.Equal(t, 0.95, analysis.Confidence)
}

func TestGetSubjectsOptions(t *testing.T) {
	opts := &GetSubjectsOptions{
		Locale:   "ru",
		Name:     "платье",
		Limit:    100,
		Offset:   0,
		ParentID: 1,
	}

	assert.Equal(t, "ru", opts.Locale)
	assert.Equal(t, "платье", opts.Name)
	assert.Equal(t, 100, opts.Limit)
	assert.Equal(t, 0, opts.Offset)
	assert.Equal(t, 1, opts.ParentID)
}

func TestFashionAnalyzerMapCharcType(t *testing.T) {
	fa := &FashionAnalyzer{}

	tests := []struct {
		charcType int
		expected  string
	}{
		{1, "string"},
		{2, "number"},
		{3, "boolean"},
		{4, "list"},
		{5, "text"},
		{99, "string"}, // Unknown type defaults to string
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("type_%d", tt.charcType), func(t *testing.T) {
			result := fa.mapCharcType(tt.charcType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFashionAnalyzerCalculateMatchScore(t *testing.T) {
	fa := &FashionAnalyzer{}

	tests := []struct {
		text     string
		keywords []string
		expected float64
	}{
		{
			text:     "платье женское летнее",
			keywords: []string{"платье", "женское"},
			expected: 3.0, // 2 exact matches + 1 partial match
		},
		{
			text:     "кроссовки спортивные",
			keywords: []string{"платье", "туфли"},
			expected: 0.0,
		},
		{
			text:     "джинсы",
			keywords: []string{"джинсы", "брюки"},
			expected: 1.5, // джинсы matches exactly + брюки is partial
		},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			score := fa.calculateMatchScore(tt.text, tt.keywords)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestWBToolCreation(t *testing.T) {
	modelRegistry := &MockModelRegistry{}
	logger := &MockLogger{}

	tool := NewWBTool(modelRegistry, logger)
	assert.NotNil(t, tool)
	assert.Equal(t, "wildberries", tool.Name())
	assert.Equal(t, "Wildberries marketplace integration tool", tool.Description())
	assert.Equal(t, "1.0.0", tool.Version())
	assert.Equal(t, "marketplace", tool.Category())

	tags := tool.Tags()
	assert.Contains(t, tags, "marketplace")
	assert.Contains(t, tags, "wildberries")
	assert.Contains(t, tags, "fashion")
}

func TestWBToolInitialization(t *testing.T) {
	modelRegistry := &MockModelRegistry{}
	logger := &MockLogger{}

	// Setup mock expectations
	logger.On("Info", "Initializing Wildberries tool", mock.Anything).Return()
	logger.On("Info", "Wildberries tool initialized successfully", mock.Anything).Maybe().Return()

	tool := NewWBTool(modelRegistry, logger)

	// Test missing API key
	ctx := context.Background()
	err := tool.Initialize(ctx, map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_key is required")

	// Test with valid config but without actual API connection
	// This will fail at connection test, which is expected
	config := map[string]interface{}{
		"api_key": "test-key",
	}
	logger.On("Info", "Initializing Wildberries tool", mock.Anything).Return()
	err = tool.Initialize(ctx, config)
	assert.Error(t, err) // Should fail connection test with test key

	// Verify all expectations were met
	logger.AssertExpectations(t)
}

func TestFashionProduct(t *testing.T) {
	now := time.Now()
	product := &FashionProduct{
		Name:        "Тестовое платье",
		Description: "Описание тестового платья",
		Brand:       "TestBrand",
		SubjectID:   105,
		Characteristics: map[string]interface{}{
			"color":  "синий",
			"size":   "M",
			"material": "хлопок",
		},
		Images: []ProductImage{
			{
				URL:   "https://example.com/image1.jpg",
				Name:  "Основное фото",
				Main:  true,
				Index: 0,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, "Тестовое платье", product.Name)
	assert.Equal(t, 105, product.SubjectID)
	assert.Len(t, product.Images, 1)
	assert.True(t, product.Images[0].Main)
	assert.Equal(t, "синий", product.Characteristics["color"])
}

func TestProductImage(t *testing.T) {
	image := ProductImage{
		URL:   "https://example.com/test.jpg",
		Name:  "Test image",
		Main:  true,
		Index: 1,
	}

	assert.Equal(t, "https://example.com/test.jpg", image.URL)
	assert.Equal(t, "Test image", image.Name)
	assert.True(t, image.Main)
	assert.Equal(t, 1, image.Index)
}