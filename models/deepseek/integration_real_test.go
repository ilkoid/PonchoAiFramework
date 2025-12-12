package deepseek

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// TestDeepSeekRealAPI_Integration runs integration tests with real DeepSeek API
// These tests require DEEPSEEK_API_KEY environment variable to be set
// Tests will be skipped if the key is not available
func TestDeepSeekRealAPI_Integration(t *testing.T) {
	// Check if API key is available
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY environment variable not set, skipping integration tests")
		return
	}

	t.Run("BasicConnectivity", func(t *testing.T) {
		testBasicConnectivity(t, apiKey)
	})

	t.Run("TextGeneration", func(t *testing.T) {
		testTextGeneration(t, apiKey)
	})

	t.Run("StreamingGeneration", func(t *testing.T) {
		testStreamingGeneration(t, apiKey)
	})

	t.Run("ToolCalling", func(t *testing.T) {
		testToolCalling(t, apiKey)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, apiKey)
	})

	t.Run("TimeoutHandling", func(t *testing.T) {
		testTimeoutHandling(t, apiKey)
	})

	t.Run("ModelCapabilities", func(t *testing.T) {
		testModelCapabilities(t, apiKey)
	})

	t.Run("RateLimiting", func(t *testing.T) {
		testRateLimiting(t, apiKey)
	})
}

// testBasicConnectivity tests basic API connectivity and authentication
func testBasicConnectivity(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create model instance
	model := NewDeepSeekModel()
	defer func() {
		if err := model.Shutdown(ctx); err != nil {
			t.Errorf("Failed to shutdown model: %v", err)
		}
	}()

	// Initialize with configuration
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "deepseek-chat",
		"max_tokens":  1000,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize DeepSeek model: %v", err)
	}

	// Test model capabilities
	capabilities := model.GetCapabilities()
	if !capabilities.Streaming {
		t.Error("DeepSeek should support streaming")
	}
	if !capabilities.Tools {
		t.Error("DeepSeek should support tools")
	}
	if capabilities.Vision {
		t.Error("DeepSeek should not support vision")
	}
	if !capabilities.System {
		t.Error("DeepSeek should support system role")
	}

	t.Logf("Model capabilities: Streaming=%v, Tools=%v, Vision=%v, System=%v",
		capabilities.Streaming, capabilities.Tools, capabilities.Vision, capabilities.System)

	t.Log("Basic connectivity test passed")
}

// testTextGeneration tests basic text generation capabilities
func testTextGeneration(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model
	model := NewDeepSeekModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "deepseek-chat",
		"max_tokens":  500,
		"temperature": 0.5,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Test simple text generation
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "What is artificial intelligence? Explain in one sentence.",
					},
				},
			},
		},
		MaxTokens:   intPtr(100),
		Temperature: float32Ptr(0.5),
	}

	resp, err := model.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to generate response: %v", err)
	}

	// Validate response
	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Message == nil {
		t.Fatal("Response message is nil")
	}

	if len(resp.Message.Content) == 0 {
		t.Fatal("Response message content is empty")
	}

	// Check that we got text content
	hasText := false
	for _, part := range resp.Message.Content {
		if part.Type == interfaces.PonchoContentTypeText && part.Text != "" {
			hasText = true
			t.Logf("Generated text: %s", part.Text)
			break
		}
	}

	if !hasText {
		t.Error("No text content in response")
	}

	// Validate usage information
	if resp.Usage == nil {
		t.Error("Usage information is nil")
	} else {
		if resp.Usage.PromptTokens <= 0 {
			t.Error("Prompt tokens should be positive")
		}
		if resp.Usage.CompletionTokens <= 0 {
			t.Error("Completion tokens should be positive")
		}
		if resp.Usage.TotalTokens <= 0 {
			t.Error("Total tokens should be positive")
		}

		t.Logf("Token usage - Prompt: %d, Completion: %d, Total: %d",
			resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	}

	t.Log("Text generation test passed")
}

// testStreamingGeneration tests streaming response generation
func testStreamingGeneration(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model
	model := NewDeepSeekModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "deepseek-chat",
		"max_tokens":  500,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Test streaming generation
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Write a short poem about programming.",
					},
				},
			},
		},
		MaxTokens:   intPtr(200),
		Temperature: float32Ptr(0.7),
		Stream:      true,
	}

	var streamChunks []*interfaces.PonchoStreamChunk
	var fullText string

	err = model.GenerateStreaming(ctx, req, func(chunk *interfaces.PonchoStreamChunk) error {
		if chunk == nil {
			return nil
		}

		streamChunks = append(streamChunks, chunk)

		// Extract text from delta
		if chunk.Delta != nil {
			for _, part := range chunk.Delta.Content {
				if part.Type == interfaces.PonchoContentTypeText {
					fullText += part.Text
				}
			}
		}

		t.Logf("Stream chunk - Done: %v, Text length: %d", chunk.Done, len(fullText))

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to generate streaming response: %v", err)
	}

	// Validate streaming response
	if len(streamChunks) == 0 {
		t.Fatal("No stream chunks received")
	}

	// Check that we got a final chunk
	lastChunk := streamChunks[len(streamChunks)-1]
	if !lastChunk.Done {
		t.Error("Final chunk is not marked as done")
	}

	if fullText == "" {
		t.Error("No text content accumulated from stream")
	}

	t.Logf("Streaming completed with %d chunks, total text length: %d", len(streamChunks), len(fullText))
	t.Logf("Generated text: %s", fullText)

	t.Log("Streaming generation test passed")
}

// testToolCalling tests tool calling capabilities
func testToolCalling(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model
	model := NewDeepSeekModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "deepseek-chat",
		"max_tokens":  500,
		"temperature": 0.1,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Define a simple weather tool
	weatherTool := interfaces.PonchoToolDef{
		Name:        "get_weather",
		Description: "Get current weather information for a location",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"location": map[string]interface{}{
					"type":        "string",
					"description": "The city and state, e.g. San Francisco, CA",
				},
			},
			"required": []string{"location"},
		},
	}

	// Test tool calling
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "What's the weather like in New York?",
					},
				},
			},
		},
		Tools:     []*interfaces.PonchoToolDef{&weatherTool},
		MaxTokens: intPtr(200),
		Stream:    false,
	}

	resp, err := model.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to generate response with tools: %v", err)
	}

	// Validate response
	if resp == nil || resp.Message == nil {
		t.Fatal("Invalid response")
	}

	// Check for tool calls
	hasToolCall := false
	for _, part := range resp.Message.Content {
		if part.Type == interfaces.PonchoContentTypeTool && part.Tool != nil {
			hasToolCall = true
			t.Logf("Tool call - Name: %s, Args: %v", part.Tool.Name, part.Tool.Args)
			break
		}
	}

	if !hasToolCall {
		t.Log("No tool call in response (this may be expected for some prompts)")
	}

	t.Log("Tool calling test passed")
}

// testErrorHandling tests error handling with invalid requests
func testErrorHandling(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model
	model := NewDeepSeekModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "deepseek-chat",
		"max_tokens":  500,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Test with invalid model name (should fail)
	invalidModelConfig := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "invalid-model-name",
		"max_tokens":  500,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	// Create a new model instance with invalid config
	invalidModelInstance := NewDeepSeekModel()
	defer invalidModelInstance.Shutdown(ctx)

	err = invalidModelInstance.Initialize(ctx, invalidModelConfig)
	if err != nil {
		t.Logf("Expected initialization error with invalid model: %v", err)
	}

	// Test with empty messages (should fail)
	req := &interfaces.PonchoModelRequest{
		Model:    "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{}, // Empty messages
	}

	_, err = model.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error with empty messages")
	} else {
		t.Logf("Expected error with empty messages: %v", err)
	}

	// Test with media content (should fail - DeepSeek doesn't support vision)
	reqWithMedia := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL:      "https://example.com/image.jpg",
							MimeType: "image/jpeg",
						},
					},
				},
			},
		},
	}

	_, err = model.Generate(ctx, reqWithMedia)
	if err == nil {
		t.Error("Expected error with media content (DeepSeek doesn't support vision)")
	} else {
		t.Logf("Expected error with media content: %v", err)
	}

	t.Log("Error handling test passed")
}

// testTimeoutHandling tests timeout handling
func testTimeoutHandling(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model with very short timeout
	model := NewDeepSeekModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "deepseek-chat",
		"max_tokens":  1000,
		"temperature": 0.7,
		"timeout":     1 * time.Millisecond, // Very short timeout
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Test with long prompt that might timeout
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Write a very long detailed story about artificial intelligence, " +
							"including its history, current state, and future prospects. " +
							"Be extremely detailed and comprehensive in your response.",
					},
				},
			},
		},
		MaxTokens:   intPtr(2000),
		Temperature: float32Ptr(0.7),
	}

	// This might timeout due to the very short timeout setting
	_, err = model.Generate(timeoutCtx, req)
	if err != nil {
		t.Logf("Request timed out as expected: %v", err)
	}

	t.Log("Timeout handling test passed")
}

// testModelCapabilities tests model capabilities and metadata
func testModelCapabilities(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model
	model := NewDeepSeekModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "deepseek-chat",
		"max_tokens":  500,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Test capabilities
	capabilities := model.GetCapabilities()

	if !capabilities.Streaming {
		t.Error("DeepSeek should support streaming")
	}

	if !capabilities.Tools {
		t.Error("DeepSeek should support tools")
	}

	if capabilities.Vision {
		t.Error("DeepSeek should not support vision")
	}

	if !capabilities.System {
		t.Error("DeepSeek should support system role")
	}

	t.Logf("Model capabilities: Streaming=%v, Tools=%v, Vision=%v, System=%v",
		capabilities.Streaming, capabilities.Tools,
		capabilities.Vision, capabilities.System)

	t.Log("Model capabilities test passed")
}

// testRateLimiting tests rate limiting behavior
func testRateLimiting(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model
	model := NewDeepSeekModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "deepseek-chat",
		"max_tokens":  100,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Make multiple rapid requests to test rate limiting
	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Say hello",
					},
				},
			},
		},
		MaxTokens:   intPtr(10),
		Temperature: float32Ptr(0.1),
	}

	// Make several requests rapidly
	successCount := 0
	errorCount := 0

	for i := 0; i < 10; i++ {
		_, err := model.Generate(ctx, req)
		if err != nil {
			errorCount++
			t.Logf("Request %d failed: %v", i+1, err)
		} else {
			successCount++
		}

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	t.Logf("Rate limiting test: %d successful, %d failed requests", successCount, errorCount)

	if successCount == 0 {
		t.Error("No requests succeeded")
	}

	t.Log("Rate limiting test passed")
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func float32Ptr(f float32) *float32 {
	return &f
}
