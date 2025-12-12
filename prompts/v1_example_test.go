package prompts

import (
	"strings"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

func TestV1Integration_RealFile(t *testing.T) {
	// Test with actual sketch_description.prompt content
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

	// Create V1 integration
	integration := NewV1Integration(nil) // No logger for this test

	// Parse the content
	template, err := integration.ParseAndConvert(content, "sketch_description_test")
	if err != nil {
		t.Fatalf("Failed to parse V1 content: %v", err)
	}

	// Verify the parsed template
	if template.Name != "sketch_description_test" {
		t.Errorf("Expected template name 'sketch_description_test', got '%s'", template.Name)
	}

	if template.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", template.Version)
	}

	if len(template.Parts) < 3 {
		t.Errorf("Expected at least 3 parts, got %d", len(template.Parts))
	}

	// Check for config part (should be system type with config content)
	var configPart *interfaces.PromptPart
	var systemPart *interfaces.PromptPart
	var userPart *interfaces.PromptPart
	var mediaPart *interfaces.PromptPart

	for _, part := range template.Parts {
		switch part.Type {
		case interfaces.PromptPartTypeSystem:
			if strings.Contains(part.Content, "temperature") {
				configPart = part
			} else {
				systemPart = part
			}
		case interfaces.PromptPartTypeUser:
			userPart = part
		case interfaces.PromptPartTypeMedia:
			mediaPart = part
		}
	}

	if configPart == nil {
		t.Error("Config part should be present as system type")
	}

	if systemPart == nil {
		t.Error("System part should be present")
	} else if !strings.Contains(systemPart.Content, "You are a precise fashion sketch analyzer") {
		t.Errorf("System part should contain analyzer text, got: %s", systemPart.Content)
	}

	if userPart == nil {
		t.Error("User part should be present")
	} else if !strings.Contains(userPart.Content, "ЗАДАЧА:") {
		t.Errorf("User part should contain task description, got: %s", userPart.Content)
	}

	if mediaPart == nil {
		t.Error("Media part should be present")
	} else if mediaPart.Media == nil {
		t.Error("Media part should have media data")
	} else if mediaPart.Media.URL != "photoUrl" {
		t.Errorf("Expected media URL 'photoUrl', got '%s'", mediaPart.Media.URL)
	}

	// Check variables
	photoUrlVar := false
	for _, variable := range template.Variables {
		if variable.Name == "photoUrl" {
			photoUrlVar = true
			break
		}
	}

	if !photoUrlVar {
		t.Error("photoUrl variable should be present")
	}

	t.Logf("Successfully parsed V1 template with %d parts and %d variables", 
		len(template.Parts), len(template.Variables))
}