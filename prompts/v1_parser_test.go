package prompts

import (
	"strings"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

func TestV1Parser_Parse(t *testing.T) {
	parser := NewV1Parser()

	// Test content from sketch_description.prompt
	content := `{{role "config"}}
model: zai-vision/glm-4.6v-flash
config:
  temperature: 0.1
  maxOutputTokens: 2000
{{role "system"}}
You are a precise fashion sketch analyzer. You are NOT an assistant. You do not speak. Act as a strict JSON generator. Your only task is to convert visual information from clothing sketches into structured JSON. 
The image is a sketch (not a photo), so focus on garment design and construction details, not mood or atmosphere.
HARD FORMAT RULES:
- Output ONLY a single valid JSON object.
- Do NOT add any text before or after the JSON.
- Do NOT use Markdown, code blocks, or HTML.
- The answer MUST start with '{' and end with '}'.
- Keys must be in Russian (snake_case) ONLY. Example: "тип_изделия", "цвет",
- Values must be in Russian ONLY.
- NEVER use English keys under any circumstances.
- If you use English keys, the entire analysis will be rejected.
- Use double quotes for all keys and string values.
- Do NOT use unescaped quotes inside values.
- If something is unclear on the sketch, use values like "неопределено" or "не видно" in Russian.
{{role "user"}}
ЗАДАЧА:
Проанализируй эскиз одежды и опиши все элементы конструкции, извлеки все важные характеристики товара.
ИЗОБРАЖЕНИЕ ДЛЯ АНАЛИЗА:
{{media url=photoUrl}}

Это эскиз одежды (а не фотография). Тебе нужно:
- Найти и перечислить все видимые элементы одежды на эскизе.
- Выделить важные конструктивные детали, которые характеризуют именно этот эскиз.
- Описывать только то, что реально видно или логично вытекает из рисунка.
- Не придумывать атмосферу, эмоции, сюжет и фон.`

	// Parse the content
	result, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check parsed data
	if result.Config == "" {
		t.Error("Config section should not be empty")
	}

	if result.System == "" {
		t.Error("System section should not be empty")
	}

	if result.User == "" {
		t.Error("User section should not be empty")
	}

	// Check media variables
	if len(result.Media) == 0 {
		t.Error("Media variables should be extracted")
	}

	if photoUrl, exists := result.Media["photoUrl"]; !exists {
		t.Error("photoUrl variable should be extracted")
	} else if photoUrl != "photoUrl" {
		t.Errorf("Expected photoUrl to be 'photoUrl', got '%s'", photoUrl)
	}

	// Check variables
	if len(result.Variables) == 0 {
		t.Error("Variables should be extracted")
	}

	if _, exists := result.Variables["photoUrl"]; !exists {
		t.Error("photoUrl should be in variables")
	}
}

func TestV1Parser_ToPromptTemplate(t *testing.T) {
	parser := NewV1Parser()

	// Create test data
	data := &V1PromptData{
		Config: "model: test-model\ntemperature: 0.7",
		System: "You are a helpful assistant.",
		User:   "Please analyze this image: {{media url=imageUrl}}",
		Media: map[string]string{
			"imageUrl": "https://example.com/image.jpg",
		},
		Variables: map[string]interface{}{
			"imageUrl": "https://example.com/image.jpg",
		},
	}

	// Convert to PromptTemplate
	template := parser.ToPromptTemplate(data, "test-template")

	// Check basic fields
	if template.Name != "test-template" {
		t.Errorf("Expected name 'test-template', got '%s'", template.Name)
	}

	if template.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", template.Version)
	}

	if template.Category != "v1" {
		t.Errorf("Expected category 'v1', got '%s'", template.Category)
	}

	// Check parts
	expectedParts := 4 // config, system, user, media
	if len(template.Parts) != expectedParts {
		t.Errorf("Expected %d parts, got %d", expectedParts, len(template.Parts))
	}

	// Check variables
	expectedVars := 1 // imageUrl
	if len(template.Variables) != expectedVars {
		t.Errorf("Expected %d variables, got %d", expectedVars, len(template.Variables))
	}

	// Find media part
	var mediaPart *interfaces.PromptPart
	for _, part := range template.Parts {
		if part.Type == interfaces.PromptPartTypeMedia {
			mediaPart = part
			break
		}
	}

	if mediaPart == nil {
		t.Error("Media part should be present")
	} else if mediaPart.Media == nil {
		t.Error("Media part should have media data")
	} else if mediaPart.Media.URL != "https://example.com/image.jpg" {
		t.Errorf("Expected media URL 'https://example.com/image.jpg', got '%s'", mediaPart.Media.URL)
	}
}

func TestV1Parser_ValidateFormat(t *testing.T) {
	parser := NewV1Parser()

	// Valid format
	validContent := `{{role "system"}}Hello{{role "user"}}World`
	if err := parser.ValidateFormat(validContent); err != nil {
		t.Errorf("Valid format should pass validation: %v", err)
	}

	// Invalid format - no role delimiters
	invalidContent1 := "Hello World"
	if err := parser.ValidateFormat(invalidContent1); err == nil {
		t.Error("Invalid format should fail validation")
	}

	// Invalid format - invalid role
	invalidContent2 := `{{role "invalid"}}Hello`
	if err := parser.ValidateFormat(invalidContent2); err == nil {
		t.Error("Invalid role should fail validation")
	}
}

func TestV1Parser_ExtractMediaVars(t *testing.T) {
	parser := NewV1Parser()

	content := "Please analyze this image: {{media url=photoUrl}} and this one: {{media url=thumbnailUrl}}"
	processedContent, mediaVars := parser.extractMediaVars(content)

	// Check media variables
	expectedVars := 2
	if len(mediaVars) != expectedVars {
		t.Errorf("Expected %d media variables, got %d", expectedVars, len(mediaVars))
	}

	if photoUrl, exists := mediaVars["photoUrl"]; !exists {
		t.Error("photoUrl should be extracted")
	} else if photoUrl != "photoUrl" {
		t.Errorf("Expected photoUrl to be 'photoUrl', got '%s'", photoUrl)
	}

	if thumbnailUrl, exists := mediaVars["thumbnailUrl"]; !exists {
		t.Error("thumbnailUrl should be extracted")
	} else if thumbnailUrl != "thumbnailUrl" {
		t.Errorf("Expected thumbnailUrl to be 'thumbnailUrl', got '%s'", thumbnailUrl)
	}

	// Check that content is unchanged (since we're just extracting variables)
	if processedContent != content {
		t.Error("Processed content should match original content")
	}
}

func TestV1Parser_Integration(t *testing.T) {
	// Integration test with the actual sketch_description.prompt file
	parser := NewV1Parser()
	
	// Test parsing and conversion back to template
	content := `{{role "config"}}
model: zai-vision/glm-4.6v-flash
config:
  temperature: 0.1
  maxOutputTokens: 2000
{{role "system"}}
You are a precise fashion sketch analyzer.
{{role "user"}}
Analyze this sketch: {{media url=photoUrl}}`

	// Parse
	data, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Convert to template
	template := parser.ToPromptTemplate(data, "sketch-description")

	// Verify template structure
	if len(template.Parts) < 3 {
		t.Errorf("Expected at least 3 parts, got %d", len(template.Parts))
	}

	// Find system part (config is also stored as system type in our implementation)
	var systemPart *interfaces.PromptPart
	for _, part := range template.Parts {
		if part.Type == interfaces.PromptPartTypeSystem && part.Content != "" && part.Media == nil {
			// Skip config part, look for actual system content
			if strings.Contains(part.Content, "You are a precise fashion sketch analyzer") {
				systemPart = part
				break
			}
		}
	}

	if systemPart == nil {
		t.Error("System part should be present")
	} else {
		// Check that it contains the expected content (it might include config too)
		if !strings.Contains(systemPart.Content, "You are a precise fashion sketch analyzer") {
			t.Errorf("Expected system content to contain 'You are a precise fashion sketch analyzer.', got '%s'", systemPart.Content)
		}
	}

	// Find user part
	var userPart *interfaces.PromptPart
	for _, part := range template.Parts {
		if part.Type == interfaces.PromptPartTypeUser {
			userPart = part
			break
		}
	}

	if userPart == nil {
		t.Error("User part should be present")
	} else if userPart.Content != "Analyze this sketch: {{media url=photoUrl}}" {
		t.Errorf("Expected user content 'Analyze this sketch: {{media url=photoUrl}}', got '%s'", userPart.Content)
	}
}