package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/deepseek"
	"github.com/ilkoid/PonchoAiFramework/models/zai"
	"github.com/ilkoid/PonchoAiFramework/prompts"
)

// Test the data structures
func TestTestResult(t *testing.T) {
	result := TestResult{
		Name:       "test-prompt",
		Type:       "prompt-test",
		Success:    true,
		Duration:   1500 * time.Millisecond,
		Model:      "deepseek-chat",
		TokensUsed: 100,
		Metadata:   map[string]interface{}{"test": "value"},
	}

	assert.Equal(t, "test-prompt", result.Name)
	assert.Equal(t, "prompt-test", result.Type)
	assert.True(t, result.Success)
	assert.Equal(t, "deepseek-chat", result.Model)
	assert.Equal(t, 100, result.TokensUsed)
}

func TestTestSuite(t *testing.T) {
	now := time.Now()
	suite := TestSuite{
		Name:        "integration-tests",
		StartTime:   now,
		EndTime:     now.Add(5 * time.Second),
		Duration:    5 * time.Second,
		TotalTests:  3,
		PassedTests: 2,
		FailedTests: 1,
		Results: []TestResult{
			{
				Name:    "test-1",
				Success: true,
			},
			{
				Name:    "test-2",
				Success: false,
			},
		},
	}

	assert.Equal(t, "integration-tests", suite.Name)
	assert.Equal(t, 3, suite.TotalTests)
	assert.Equal(t, 2, suite.PassedTests)
	assert.Equal(t, 1, suite.FailedTests)
	assert.Len(t, suite.Results, 2)
}

func TestTestSummary(t *testing.T) {
	summary := TestSummary{
		ModelsTested:    []string{"deepseek-chat", "glm-chat"},
		PromptsTested:   []string{"test.prompt", "sketch.prompt"},
		TotalTokens:     500,
		AvgResponseTime: 1200 * time.Millisecond,
	}

	assert.Contains(t, summary.ModelsTested, "deepseek-chat")
	assert.Contains(t, summary.ModelsTested, "glm-chat")
	assert.Equal(t, 500, summary.TotalTokens)
}

func TestCreateTestModel(t *testing.T) {
	t.Run("Create DeepSeek model", func(t *testing.T) {
		apiKey := os.Getenv("DEEPSEEK_API_KEY")
		if apiKey == "" {
			t.Skip("Skipping test: no DEEPSEEK_API_KEY")
		}

		config := map[string]interface{}{
			"api_key":    apiKey,
			"model_name": "deepseek-chat",
			"timeout":    "30s",
		}

		model := createTestModel("deepseek", config)
		assert.NotNil(t, model)

		// Verify model type
		_, ok := model.(*deepseek.DeepSeekModel)
		assert.True(t, ok)
	})

	t.Run("Create ZAI model", func(t *testing.T) {
		apiKey := os.Getenv("ZAI_API_KEY")
		if apiKey == "" {
			t.Skip("Skipping test: no ZAI_API_KEY")
		}

		config := map[string]interface{}{
			"api_key":    apiKey,
			"model_name": "glm-chat",
			"timeout":    "30s",
		}

		model := createTestModel("zai", config)
		assert.NotNil(t, model)

		// Verify model type
		_, ok := model.(*zai.ZAIModel)
		assert.True(t, ok)
	})

	t.Run("Invalid provider", func(t *testing.T) {
		config := map[string]interface{}{
			"api_key": "test-key",
		}

		model := createTestModel("invalid-provider", config)
		assert.Nil(t, model)
	})
}

func TestRunModelTest(t *testing.T) {
	// Skip if no API keys
	if os.Getenv("DEEPSEEK_API_KEY") == "" && os.Getenv("ZAI_API_KEY") == "" {
		t.Skip("Skipping test: no API keys available")
	}

	ctx := context.Background()
	testModel := createTestModel("deepseek", map[string]interface{}{
		"api_key":    os.Getenv("DEEPSEEK_API_KEY"),
		"model_name": "deepseek-chat",
		"timeout":    "30s",
	})

	if testModel == nil {
		t.Skip("Could not create test model")
	}

	test := ModelTest{
		Name:    "basic-generation",
		Enabled: true,
		Request: &interfaces.PonchoModelRequest{
			Model: "deepseek-chat",
			Messages: []*interfaces.PonchoMessage{
				{
					Role: interfaces.PonchoRoleUser,
					Content: []*interfaces.PonchoContentPart{
						{
							Type: interfaces.PonchoContentTypeText,
							Text: "Say hello in one word",
						},
					},
				},
			},
			MaxTokens: func() *int { v := 10; return &v }(),
		},
	}

	result := runModelTest(ctx, test, testModel)

	assert.NotNil(t, result)
	assert.Equal(t, "basic-generation", result.Name)
	assert.Equal(t, "model-test", result.Type)
	assert.Equal(t, "deepseek-chat", result.Model)
	assert.Greater(t, result.Duration.Milliseconds(), int64(0))
}

func TestRunPromptTest(t *testing.T) {
	// Create a test prompt
	promptContent := `model: deepseek-chat
temperature: 0.7
---
system:
You are a helpful assistant.
---
user:
What is {{item}}?
`

	tempFile := writeTempPrompt(t, promptContent)
	defer os.Remove(tempFile)

	ctx := context.Background()
	test := PromptTest{
		Name:      "test-prompt",
		Enabled:   true,
		FilePath:  tempFile,
		Variables: map[string]interface{}{"item": "AI"},
	}

	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("Skipping test: no DEEPSEEK_API_KEY")
	}

	model := createTestModel("deepseek", map[string]interface{}{
		"api_key":    os.Getenv("DEEPSEEK_API_KEY"),
		"model_name": "deepseek-chat",
		"timeout":    "30s",
	})

	result := runPromptTest(ctx, test, model)

	assert.NotNil(t, result)
	assert.Equal(t, "test-prompt", result.Name)
	assert.Equal(t, "prompt-test", result.Type)
}

func TestTestSuitesIntegration(t *testing.T) {
	now := time.Now()
	suite := TestSuite{
		Name:      "test-suite",
		StartTime: now,
		Results: []TestResult{
			{
				Name:       "test-1",
				Success:    true,
				Duration:   100 * time.Millisecond,
				Model:      "model-1",
				TokensUsed: 50,
			},
			{
				Name:       "test-2",
				Success:    false,
				Duration:   200 * time.Millisecond,
				Model:      "model-2",
				TokensUsed: 75,
				Error:      "test error",
			},
		},
	}

	// Test Finalize
	suite.EndTime = now.Add(300 * time.Millisecond)
	suite.Duration = suite.EndTime.Sub(suite.StartTime)
	suite.TotalTests = len(suite.Results)
	suite.PassedTests = 1
	suite.FailedTests = 1

	assert.Equal(t, 2, suite.TotalTests)
	assert.Equal(t, 1, suite.PassedTests)
	assert.Equal(t, 1, suite.FailedTests)

	// Test summary calculation
	suite.Summary = calculateSummary(suite.Results)
	assert.Contains(t, suite.Summary.ModelsTested, "model-1")
	assert.Contains(t, suite.Summary.ModelsTested, "model-2")
	assert.Equal(t, 125, suite.Summary.TotalTokens)
}

func TestPromptParsing(t *testing.T) {
	promptContent := `model: deepseek-chat
temperature: 0.5
max_tokens: 100
---
system:
You are a test assistant.
---
user:
Process: {{input}}
---
media:
{{image_url}}
`

	tempFile := writeTempPrompt(t, promptContent)
	defer os.Remove(tempFile)

	// Parse using V1 parser
	v1Integration := prompts.NewV1Integration(nil)
	template, err := v1Integration.ParseAndConvert(promptContent, "test")

	assert.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "test", template.Name)
}

// Helper functions for testing
func createTestModel(provider string, config map[string]interface{}) interfaces.PonchoModel {
	switch provider {
	case "deepseek":
		return deepseek.NewDeepSeekModel()
	case "zai":
		return zai.NewZAIModel()
	default:
		return nil
	}
}

func runModelTest(ctx context.Context, test ModelTest, model interfaces.PonchoModel) TestResult {
	start := time.Now()
	result := TestResult{
		Name:  test.Name,
		Type:  "model-test",
		Model: test.Request.Model,
	}

	// Initialize model if needed
	if initializer, ok := model.(interface{ Initialize(context.Context, map[string]interface{}) error }); ok {
		err := initializer.Initialize(ctx, map[string]interface{}{})
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Duration = time.Since(start)
			return result
		}
	}

	// Execute request
	resp, err := model.Generate(ctx, test.Request)
	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	result.Success = true
	if resp.Usage != nil {
		result.TokensUsed = resp.Usage.TotalTokens
	}

	return result
}

func runPromptTest(ctx context.Context, test PromptTest, model interfaces.PonchoModel) TestResult {
	start := time.Now()
	result := TestResult{
		Name: test.Name,
		Type: "prompt-test",
	}

	// Parse prompt
	content, err := os.ReadFile(test.FilePath)
	if err != nil {
		result.Success = false
		result.Error = "Failed to read prompt file: " + err.Error()
		result.Duration = time.Since(start)
		return result
	}

	v1Integration := prompts.NewV1Integration(nil)
	template, err := v1Integration.ParseAndConvert(string(content), test.Name)
	if err != nil {
		result.Success = false
		result.Error = "Failed to parse prompt: " + err.Error()
		result.Duration = time.Since(start)
		return result
	}

	result.Model = template.Model

	// Create request from template
	messages := make([]*interfaces.PonchoMessage, 0)
	for _, part := range template.Parts {
		if part.Type == interfaces.PromptPartTypeMedia {
			continue // Skip media for this test
		}

		var role interfaces.PonchoRole
		switch part.Type {
		case interfaces.PromptPartTypeSystem:
			role = interfaces.PonchoRoleSystem
		case interfaces.PromptPartTypeUser:
			role = interfaces.PonchoRoleUser
		default:
			continue
		}

		// Simple variable substitution
		content := part.Content
		for key, value := range test.Variables {
			placeholder := "{{" + key + "}}"
			if strVal, ok := value.(string); ok {
				content = strings.ReplaceAll(content, placeholder, strVal)
			} else {
				content = strings.ReplaceAll(content, placeholder, interfaceToString(value))
			}
		}

		messages = append(messages, &interfaces.PonchoMessage{
			Role: role,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: content,
				},
			},
		})
	}

	request := &interfaces.PonchoModelRequest{
		Model:       template.Model,
		Messages:    messages,
		Temperature: template.Temperature,
		MaxTokens:   template.MaxTokens,
	}

	// Execute request
	resp, err := model.Generate(ctx, request)
	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	result.Success = true
	if resp.Usage != nil {
		result.TokensUsed = resp.Usage.TotalTokens
	}

	return result
}

func calculateSummary(results []TestResult) TestSummary {
	summary := TestSummary{
		ModelsTested:  make([]string, 0),
		PromptsTested: make([]string, 0),
		TotalTokens:   0,
	}

	modelSet := make(map[string]bool)
	promptSet := make(map[string]bool)

	var totalTime time.Duration
	for _, result := range results {
		if result.Model != "" {
			modelSet[result.Model] = true
		}
		if result.Name != "" {
			promptSet[result.Name] = true
		}
		summary.TotalTokens += result.TokensUsed
		totalTime += result.Duration
	}

	for model := range modelSet {
		summary.ModelsTested = append(summary.ModelsTested, model)
	}
	for prompt := range promptSet {
		summary.PromptsTested = append(summary.PromptsTested, prompt)
	}

	if len(results) > 0 {
		summary.AvgResponseTime = totalTime / time.Duration(len(results))
	}

	return summary
}

// Types needed for testing
type ModelTest struct {
	Name    string                      `json:"name"`
	Enabled bool                        `json:"enabled"`
	Request *interfaces.PonchoModelRequest `json:"request"`
}

type PromptTest struct {
	Name      string                 `json:"name"`
	Enabled   bool                   `json:"enabled"`
	FilePath  string                 `json:"file_path"`
	Variables map[string]interface{} `json:"variables"`
}

// Helper functions
func writeTempPrompt(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "test-*.prompt")
	require.NoError(t, err)
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	return tmpFile.Name()
}

func interfaceToString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}