package prompts

import (
	"strings"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

func TestV1Parser_RealFile(t *testing.T) {
	// Test with actual sketch_description.prompt file content
	filePath := "../examples/test_data/prompts/sketch_description.prompt"
	
	// Read file content
	content, err := readFileContent(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}

	// Create V1 integration
	integration := NewV1Integration(nil) // No logger for this test

	// Parse the content
	template, err := integration.ParseAndConvert(content, "sketch_description_real")
	if err != nil {
		t.Fatalf("Failed to parse real V1 file: %v", err)
	}

	// Verify the parsed template
	t.Logf("Template name: %s", template.Name)
	t.Logf("Template version: %s", template.Version)
	t.Logf("Number of parts: %d", len(template.Parts))
	t.Logf("Number of variables: %d", len(template.Variables))

	// Check for expected content
	var configPart *interfaces.PromptPart
	var systemPart *interfaces.PromptPart
	var userPart *interfaces.PromptPart
	var mediaPart *interfaces.PromptPart

	for i, part := range template.Parts {
		t.Logf("Part %d - Type: %s, Content: %s", i, part.Type, part.Content)
		
		switch part.Type {
		case interfaces.PromptPartTypeSystem:
			if strings.Contains(part.Content, "temperature") {
				configPart = part
			} else if strings.Contains(part.Content, "You are a precise fashion sketch analyzer") {
				systemPart = part
			}
		case interfaces.PromptPartTypeUser:
			userPart = part
		case interfaces.PromptPartTypeMedia:
			mediaPart = part
		}
	}

	// Validate parts
	if configPart == nil {
		t.Error("Config part should be present")
	} else {
		t.Logf("Found config part: %s", configPart.Content[:50]+"...")
	}

	if systemPart == nil {
		t.Error("System part should be present")
	} else {
		t.Logf("Found system part: %s", systemPart.Content[:50]+"...")
	}

	if userPart == nil {
		t.Error("User part should be present")
	} else {
		t.Logf("Found user part: %s", userPart.Content[:50]+"...")
	}

	if mediaPart == nil {
		t.Error("Media part should be present")
	} else {
		t.Logf("Found media part with URL: %s", mediaPart.Media.URL)
	}

	// Check variables
	photoUrlVar := false
	for _, variable := range template.Variables {
		t.Logf("Variable: %s (type: %s, required: %v)", variable.Name, variable.Type, variable.Required)
		if variable.Name == "photoUrl" {
			photoUrlVar = true
		}
	}

	if !photoUrlVar {
		t.Error("photoUrl variable should be present")
	}

	// Test format validation
	if err := integration.v1Parser.ValidateFormat(content); err != nil {
		t.Errorf("Format validation failed: %v", err)
	}

	t.Logf("✅ Successfully parsed real sketch_description.prompt file!")
	t.Logf("✅ V1 parser is working correctly with actual project files!")
}
