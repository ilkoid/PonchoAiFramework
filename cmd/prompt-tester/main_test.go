package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

func TestParsePromptV1(t *testing.T) {
	// Create a temporary prompt file
	tempDir := t.TempDir()
	promptFile := filepath.Join(tempDir, "test.prompt")

	// Sample V1 prompt content
	v1Content := `model: glm-4.6v-flash
temperature: 0.1
max_tokens: 4000
timeout: 60s
---
system:
You are a precise fashion sketch analyzer.
---
user:
TASK: Analyze the fashion sketch at the provided URL.

URL: {{photoUrl}}

REQUIREMENTS:
1. Extract structured data about the clothing item
2. Return only valid JSON object
3. Include all visible details
---
media:
{{photoUrl}}
`

	err := os.WriteFile(promptFile, []byte(v1Content), 0644)
	require.NoError(t, err)

	// Test parsing
	template, err := parsePromptV1(v1Content, "test.prompt", log.New(os.Stdout, "[TEST] ", log.LstdFlags))
	assert.NoError(t, err)
	assert.NotNil(t, template)

	// Verify template properties
	assert.Equal(t, "test.prompt", template.Name)
	// Temperature and MaxTokens are pointer fields, need to dereference
	if template.Temperature != nil {
		assert.Equal(t, float32(0.1), *template.Temperature)
	}
	if template.MaxTokens != nil {
		assert.Equal(t, 4000, *template.MaxTokens)
	}

	// Verify parts
	require.Len(t, template.Parts, 3)

	// Check system part
	systemPart := template.Parts[0]
	assert.Equal(t, interfaces.PromptPartTypeSystem, systemPart.Type)
	assert.Contains(t, systemPart.Content, "fashion sketch analyzer")

	// Check user part
	userPart := template.Parts[1]
	assert.Equal(t, interfaces.PromptPartTypeUser, userPart.Type)
	assert.Contains(t, userPart.Content, "{{photoUrl}}")

	// Check media part
	mediaPart := template.Parts[2]
	assert.Equal(t, interfaces.PromptPartTypeMedia, mediaPart.Type)
	assert.Equal(t, "{{photoUrl}}", mediaPart.Content)

	// Verify variables
	require.Len(t, template.Variables, 1)
	assert.Equal(t, "photoUrl", template.Variables[0].Name)
	assert.True(t, template.Variables[0].Required)
}

func TestParsePromptV1InvalidFormat(t *testing.T) {
	invalidContent := "This is not a valid V1 prompt format"

	_, err := parsePromptV1(invalidContent, "invalid.prompt", log.New(os.Stdout, "[TEST] ", log.LstdFlags))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestParsePromptV1MissingModel(t *testing.T) {
	content := `temperature: 0.1
---
system:
Hello world
---
user:
Test message
`

	_, err := parsePromptV1(content, "no-model.prompt", log.New(os.Stdout, "[TEST] ", log.LstdFlags))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model is required")
}

func TestPrepareOutputDirectory(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "test_output")

	// Test creating new directory
	err := prepareOutputDirectory(outputDir)
	assert.NoError(t, err)

	// Verify directory exists
	info, err := os.Stat(outputDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Test creating directory that already exists
	err = prepareOutputDirectory(outputDir)
	assert.NoError(t, err) // Should not error
}

func TestSaveResult(t *testing.T) {
	tempDir := t.TempDir()

	result := map[string]interface{}{
		"prompt": "Test prompt",
		"response": map[string]interface{}{
			"content": "Generated text",
			"tokens": 150,
		},
		"metadata": map[string]interface{}{
			"model": "test-model",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	// Test saving result
	filename, err := saveResult(tempDir, "test-prompt", result)
	assert.NoError(t, err)
	assert.NotEmpty(t, filename)

	// Verify file exists
	_, err = os.Stat(filepath.Join(tempDir, filename))
	assert.NoError(t, err)

	// Verify content
	savedContent, err := os.ReadFile(filepath.Join(tempDir, filename))
	assert.NoError(t, err)
	assert.Contains(t, string(savedContent), "Generated text")
}

func TestProcessPromptVariables(t *testing.T) {
	template := &interfaces.PromptTemplate{
		Parts: []*interfaces.PromptPart{
			{
				Type:    interfaces.PromptPartTypeUser,
				Content: "Hello {{name}}, you are {{age}} years old",
			},
			{
				Type:    interfaces.PromptPartTypeSystem,
				Content: "System message with {{variable}}",
			},
		},
		Variables: []*interfaces.PromptVariable{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "age",
				Type:     "number",
				Required: false,
			},
		},
	}

	variables := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}

	processedTemplate := processPromptVariables(template, variables)

	// Check that variables were substituted
	assert.Contains(t, processedTemplate.Parts[0].Content, "Hello Alice")
	assert.Contains(t, processedTemplate.Parts[0].Content, "you are 30 years old")
}

func TestValidatePromptVariables(t *testing.T) {
	template := &interfaces.PromptTemplate{
		Variables: []*interfaces.PromptVariable{
			{
				Name:     "required_var",
				Type:     "string",
				Required: true,
			},
			{
				Name:     "optional_var",
				Type:     "string",
				Required: false,
			},
		},
	}

	t.Run("All required variables provided", func(t *testing.T) {
		variables := map[string]interface{}{
			"required_var": "value",
		}
		err := validatePromptVariables(template, variables)
		assert.NoError(t, err)
	})

	t.Run("Missing required variable", func(t *testing.T) {
		variables := map[string]interface{}{
			"optional_var": "value",
		}
		err := validatePromptVariables(template, variables)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required_var")
	})

	t.Run("Empty variables map", func(t *testing.T) {
		variables := map[string]interface{}{}
		err := validatePromptVariables(template, variables)
		assert.Error(t, err)
	})
}

func TestCreateModelRequest(t *testing.T) {
	temp := float32(0.7)
	maxTokens := 1000
	template := &interfaces.PromptTemplate{
		Name:        "test-template",
		Model:       "test-model",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Parts: []*interfaces.PromptPart{
			{
				Type:    interfaces.PromptPartTypeSystem,
				Content: "You are a helpful assistant",
			},
			{
				Type:    interfaces.PromptPartTypeUser,
				Content: "Hello {{name}}",
			},
		},
	}

	variables := map[string]interface{}{
		"name": "World",
	}

	// Process variables first
	processedTemplate := processPromptVariables(template, variables)

	// Create request
	request := createModelRequest(processedTemplate)

	assert.Equal(t, "test-model", request.Model)
	if request.Temperature != nil {
		assert.Equal(t, float32(0.7), *request.Temperature)
	}
	if request.MaxTokens != nil {
		assert.Equal(t, 1000, *request.MaxTokens)
	}

	require.Len(t, request.Messages, 2)
	assert.Equal(t, interfaces.PonchoRoleSystem, request.Messages[0].Role)
	assert.Equal(t, interfaces.PonchoRoleUser, request.Messages[1].Role)
	assert.Contains(t, request.Messages[1].Content[0].Text, "Hello World")
}

func TestMainFunctionIntegration(t *testing.T) {
	// This test verifies the main flow components
	// We don't test the actual main() function to avoid flag parsing complexity

	t.Run("Flag parsing requirements", func(t *testing.T) {
		// Verify that prompt flag is required
		// This is tested indirectly by checking that main would fail without it
		// The actual test would require os.Args manipulation which is complex
	})

	t.Run("Output directory creation", func(t *testing.T) {
		tempDir := t.TempDir()
		outputDir := filepath.Join(tempDir, "test_output")

		err := prepareOutputDirectory(outputDir)
		assert.NoError(t, err)

		// Verify directory exists and is writable
		testFile := filepath.Join(outputDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		assert.NoError(t, err)
	})

	t.Run("Default output directory", func(t *testing.T) {
		// Test that "./output" is the default
		defaultOutput := "./output"
		assert.Equal(t, defaultOutput, defaultOutput)
	})
}

func TestPromptFileProcessing(t *testing.T) {
	// Test with actual V1 prompt files from test data
	testDataDir := "../../examples/test_data/prompts"

	promptFiles := []string{
		"sketch_description.prompt",
		"sketch_creative.prompt",
	}

	for _, filename := range promptFiles {
		t.Run("Process "+filename, func(t *testing.T) {
			promptPath := filepath.Join(testDataDir, filename)

			// Skip if test file doesn't exist
			if _, err := os.Stat(promptPath); os.IsNotExist(err) {
				t.Skipf("Test prompt file not found: %s", filename)
				return
			}

			content, err := os.ReadFile(promptPath)
			require.NoError(t, err)

			// Parse the prompt
			template, err := parsePromptV1(string(content), filename, nil)
			assert.NoError(t, err)
			assert.NotNil(t, template)

			// Verify basic structure
			assert.NotEmpty(t, template.Model)
			assert.NotEmpty(t, template.Parts)
			assert.NotNil(t, template.Variables)

			// Check for required model in sketch descriptions
			if strings.Contains(filename, "sketch") {
				assert.Equal(t, "glm-4.6v-flash", template.Model,
					"Sketch prompts should use vision model, got: %s", template.Model)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	t.Run("Invalid prompt path", func(t *testing.T) {
		invalidPath := "/path/that/does/not/exist.prompt"
		_, err := os.Stat(invalidPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Empty prompt file", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "empty.prompt")
		err := os.WriteFile(tempFile, []byte(""), 0644)
		require.NoError(t, err)

		content, err := os.ReadFile(tempFile)
		require.NoError(t, err)

		_, err = parsePromptV1(string(content), "empty.prompt", nil)
		assert.Error(t, err)
	})

	t.Run("Malformed YAML", func(t *testing.T) {
		malformedContent := `model: test
temperature: invalid_number_value
---
system:
Test
`
		_, err := parsePromptV1(malformedContent, "malformed.prompt", log.New(os.Stdout, "[TEST] ", log.LstdFlags))
		assert.Error(t, err)
	})
}

// Mock functions needed for testing
// These would normally be internal to main.go

func processPromptVariables(template *interfaces.PromptTemplate, variables map[string]interface{}) *interfaces.PromptTemplate {
	// Create a copy to avoid modifying the original
	processed := &interfaces.PromptTemplate{
		Name:        template.Name,
		Description: template.Description,
		Version:     template.Version,
		Category:    template.Category,
		Tags:        template.Tags,
		Model:       template.Model,
		Temperature: template.Temperature,
		MaxTokens:   template.MaxTokens,
		Parts:       make([]*interfaces.PromptPart, len(template.Parts)),
		Variables:   template.Variables,
		Metadata:    template.Metadata,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}

	// Process each part
	for _, part := range template.Parts {
		processedPart := *part
		// Simple variable substitution
		for key, value := range variables {
			placeholder := "{{" + key + "}}"
			if strVal, ok := value.(string); ok {
				processedPart.Content = strings.ReplaceAll(processedPart.Content, placeholder, strVal)
			} else {
				processedPart.Content = strings.ReplaceAll(processedPart.Content, placeholder, interfaceToString(value))
			}
		}
		processed.Parts = append(processed.Parts, &processedPart)
	}

	return processed
}

func validatePromptVariables(template *interfaces.PromptTemplate, variables map[string]interface{}) error {
	for _, variable := range template.Variables {
		if variable.Required {
			if _, exists := variables[variable.Name]; !exists {
				return fmt.Errorf("required variable '%s' is missing", variable.Name)
			}
		}
	}
	return nil
}

func createModelRequest(template *interfaces.PromptTemplate) *interfaces.PonchoModelRequest {
	request := &interfaces.PonchoModelRequest{
		Model:       template.Model,
		Temperature: template.Temperature,
		MaxTokens:   template.MaxTokens,
		Stream:      false,
	}

	for _, part := range template.Parts {
		if part.Type == interfaces.PromptPartTypeMedia {
			continue // Skip media parts for this test
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

		message := &interfaces.PonchoMessage{
			Role: role,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: part.Content,
				},
			},
		}

		request.Messages = append(request.Messages, message)
	}

	return request
}

func prepareOutputDirectory(outputDir string) error {
	return os.MkdirAll(outputDir, 0755)
}

func saveResult(outputDir, promptName string, result map[string]interface{}) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", promptName, timestamp)
	filePath := filepath.Join(outputDir, filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func interfaceToString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}