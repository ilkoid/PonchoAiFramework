package deepseek

import (
	"context"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/common"
)

// Basic test to verify the model can be created and initialized
func TestBasicFunctionality(t *testing.T) {
	// Test model creation
	model := NewDeepSeekModel()
	if model == nil {
		t.Fatal("Failed to create DeepSeek model")
	}

	// Test basic properties
	if model.Name() != "deepseek-chat" {
		t.Errorf("Expected name 'deepseek-chat', got '%s'", model.Name())
	}

	if model.Provider() != string(common.ProviderDeepSeek) {
		t.Errorf("Expected provider '%s', got '%s'", common.ProviderDeepSeek, model.Provider())
	}

	if !model.SupportsStreaming() {
		t.Error("Model should support streaming")
	}

	if !model.SupportsTools() {
		t.Error("Model should support tools")
	}

	if model.SupportsVision() {
		t.Error("Model should not support vision")
	}

	if !model.SupportsSystemRole() {
		t.Error("Model should support system role")
	}

	// Test initialization
	config := map[string]interface{}{
		"api_key":    "test-key",
		"model_name": "deepseek-chat",
	}

	ctx := context.Background()
	err := model.Initialize(ctx, config)
	if err != nil {
		t.Errorf("Failed to initialize model: %v", err)
	}

	if !model.isInitialized() {
		t.Error("Model should be initialized after Initialize()")
	}

	// Test shutdown
	err = model.Shutdown(ctx)
	if err != nil {
		t.Errorf("Failed to shutdown model: %v", err)
	}
}

// Test request validation
func TestRequestValidation(t *testing.T) {
	model := NewDeepSeekModel()

	// Valid request
	validReq := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello"},
				},
			},
		},
	}

	err := model.ValidateRequest(validReq)
	if err != nil {
		t.Errorf("Valid request should not produce error: %v", err)
	}

	// Invalid request (nil)
	err = model.ValidateRequest(nil)
	if err == nil {
		t.Error("Nil request should produce error")
	}

	// Invalid request (no messages)
	invalidReq := &interfaces.PonchoModelRequest{
		Model:    "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{},
	}

	err = model.ValidateRequest(invalidReq)
	if err == nil {
		t.Error("Request with no messages should produce error")
	}
}
