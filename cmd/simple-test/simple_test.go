package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// TestBasicFunctionality tests basic components without requiring the full framework
func TestBasicFunctionality(t *testing.T) {
	t.Run("Template Creation", func(t *testing.T) {
		template := &interfaces.PromptTemplate{
			Name:        "simple-test",
			Description: "Simple test prompt",
			Version:     "1.0.0",
			Category:    "test",
			Tags:        []string{"test", "simple"},
			Parts: []*interfaces.PromptPart{
				{
					Type:    interfaces.PromptPartTypeSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Type:    interfaces.PromptPartTypeUser,
					Content: "Generate description for {{product}}",
				},
			},
			Variables: []*interfaces.PromptVariable{
				{
					Name:        "product",
					Type:        "string",
					Description: "Product name",
					Required:    true,
				},
			},
		}

		assert.Equal(t, "simple-test", template.Name)
		assert.Len(t, template.Parts, 2)
		assert.Len(t, template.Variables, 1)
	})

	t.Run("Environment Variables", func(t *testing.T) {
		deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
		zaiKey := os.Getenv("ZAI_API_KEY")

		// At least one key should be available for real execution
		if deepseekKey != "" || zaiKey != "" {
			t.Logf("API keys found: deepseek=%t, zai=%t", deepseekKey != "", zaiKey != "")
		} else {
			t.Log("No API keys found - expected in CI environment")
		}
	})

	t.Run("Model Selection Logic", func(t *testing.T) {
		deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
		zaiKey := os.Getenv("ZAI_API_KEY")

		modelName := "deepseek-chat" // default
		if deepseekKey == "" && zaiKey != "" {
			modelName = "glm-chat"
		}

		assert.Contains(t, []string{"deepseek-chat", "glm-chat"}, modelName)
	})
}

// TestProcessVariablesConcept tests the concept of variable processing
func TestProcessVariablesConcept(t *testing.T) {
	// Test placeholder generation
	placeholder := fmt.Sprintf("{{%s}}", "product")
	assert.Equal(t, "{{product}}", placeholder)

	// Test string conversion for different types
	tests := []struct {
		value    interface{}
		expected string
	}{
		{"hello", "hello"},
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("type_%T", tt.value), func(t *testing.T) {
			result := fmt.Sprintf("%v", tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPromptTemplateValidation validates prompt template structures
func TestPromptTemplateValidation(t *testing.T) {
	t.Run("Valid Template", func(t *testing.T) {
		template := &interfaces.PromptTemplate{
			Name:        "valid-template",
			Description: "A valid prompt template",
			Version:     "1.0.0",
			Category:    "test",
			Model:       "deepseek-chat",
			Parts: []*interfaces.PromptPart{
				{
					Type:    interfaces.PromptPartTypeSystem,
					Content: "System instructions",
				},
			},
			Variables: []*interfaces.PromptVariable{
				{
					Name:     "input",
					Type:     "string",
					Required: true,
				},
			},
		}

		assert.NotNil(t, template)
		assert.Equal(t, "valid-template", template.Name)
		assert.NotEmpty(t, template.Model)
		assert.NotEmpty(t, template.Parts)
	})

	t.Run("Template with Variables", func(t *testing.T) {
		template := &interfaces.PromptTemplate{
			Name:  "template-with-vars",
			Model: "deepseek-chat",
			Variables: []*interfaces.PromptVariable{
				{
					Name:        "required_var",
					Type:        "string",
					Required:    true,
					Description: "A required variable",
				},
				{
					Name:        "optional_var",
					Type:        "number",
					Required:    false,
					DefaultValue: 42,
				},
			},
		}

		assert.Len(t, template.Variables, 2)
		assert.True(t, template.Variables[0].Required)
		assert.False(t, template.Variables[1].Required)
		assert.Equal(t, 42, template.Variables[1].DefaultValue)
	})
}

// TestConfiguration tests configuration concepts
func TestConfiguration(t *testing.T) {
	t.Run("Model Configuration", func(t *testing.T) {
		config := map[string]interface{}{
			"provider":      "deepseek",
			"model_name":    "deepseek-chat",
			"api_key":       "test-key",
			"max_tokens":    1000,
			"temperature":   0.7,
			"timeout":       "30s",
			"supports": map[string]bool{
				"streaming": true,
				"tools":     true,
				"vision":    false,
			},
		}

		assert.Equal(t, "deepseek", config["provider"])
		assert.Equal(t, "deepseek-chat", config["model_name"])
		assert.Equal(t, 1000, config["max_tokens"])
		assert.Equal(t, 0.7, config["temperature"])

		supports, ok := config["supports"].(map[string]bool)
		assert.True(t, ok)
		assert.True(t, supports["streaming"])
		assert.False(t, supports["vision"])
	})

	t.Run("Required Environment Variables", func(t *testing.T) {
		requiredVars := []string{"DEEPSEEK_API_KEY", "ZAI_API_KEY"}
		var found []string

		for _, varName := range requiredVars {
			if os.Getenv(varName) != "" {
				found = append(found, varName)
			}
		}

		t.Logf("Found environment variables: %v", found)
		// In CI, we might not have these, so we just log the result
	})
}

// BenchmarkPlaceholderGeneration benchmarks placeholder generation
func BenchmarkPlaceholderGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("{{%s}}", "product")
	}
}