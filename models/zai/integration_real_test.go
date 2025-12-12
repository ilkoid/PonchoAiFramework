package zai

import (
	"context"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// TestZAIRealAPI_Integration runs integration tests with real Z.AI API
// These tests require ZAI_API_KEY environment variable to be set
// Tests will be skipped if key is not available
func TestZAIRealAPI_Integration(t *testing.T) {
	// Check if API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration tests")
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

	t.Run("VisionAnalysis", func(t *testing.T) {
		testVisionAnalysis(t, apiKey)
	})

	t.Run("FashionVisionAnalysis", func(t *testing.T) {
		testFashionVisionAnalysis(t, apiKey)
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
	model := NewZAIModel()
	defer func() {
		if err := model.Shutdown(ctx); err != nil {
			t.Errorf("Failed to shutdown model: %v", err)
		}
	}()

	// Initialize with configuration
	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
		"max_tokens":  1000,
		"temperature": 0.5,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize Z.AI model: %v", err)
	}

	// Test model capabilities
	capabilities := model.GetCapabilities()
	if !capabilities.Streaming {
		t.Error("Z.AI should support streaming")
	}
	if !capabilities.Tools {
		t.Error("Z.AI should support tools")
	}
	if !capabilities.Vision {
		t.Error("Z.AI should support vision")
	}
	if !capabilities.System {
		t.Error("Z.AI should support system role")
	}

	t.Logf("Model capabilities: Streaming=%v, Tools=%v, Vision=%v, System=%v",
		capabilities.Streaming, capabilities.Tools, capabilities.Vision, capabilities.System)

	t.Log("Basic connectivity test passed")
}

// testTextGeneration tests basic text generation capabilities
func testTextGeneration(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model
	model := NewZAIModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
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
		Model: "glm-4.6",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "What is machine learning? Explain in one sentence.",
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
	model := NewZAIModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
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
		Model: "glm-4.6",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Write a short haiku about technology.",
					},
				},
			},
		},
		MaxTokens:   intPtr(100),
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

// testVisionAnalysis tests basic vision analysis capabilities
func testVisionAnalysis(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize vision model
	model := NewZAIVisionModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6v-flash",
		"max_tokens":  1000,
		"temperature": 0.3,
		"timeout":     60 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize Z.AI vision model: %v", err)
	}

	// Create a simple test image (1x1 pixel red PNG in base64)
	// This is a minimal valid PNG image for testing
	testImageBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="

	// Test vision analysis
	req := &interfaces.PonchoModelRequest{
		Model: "glm-4.6v",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "What do you see in this image?",
					},
					{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL:      "data:image/png;base64," + testImageBase64,
							MimeType: "image/png",
						},
					},
				},
			},
		},
		MaxTokens:   intPtr(200),
		Temperature: float32Ptr(0.3),
	}

	resp, err := model.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to generate vision response: %v", err)
	}

	// Validate response
	if resp == nil || resp.Message == nil {
		t.Fatal("Invalid vision response")
	}

	// Check that we got text content
	hasText := false
	for _, part := range resp.Message.Content {
		if part.Type == interfaces.PonchoContentTypeText && part.Text != "" {
			hasText = true
			t.Logf("Vision analysis result: %s", part.Text)
			break
		}
	}

	if !hasText {
		t.Error("No text content in vision response")
	}

	t.Log("Vision analysis test passed")
}

// testFashionVisionAnalysis tests fashion-specific vision analysis
func testFashionVisionAnalysis(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize vision model
	model := NewZAIVisionModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6v",
		"max_tokens":  1000,
		"temperature": 0.3,
		"timeout":     60 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize Z.AI vision model: %v", err)
	}

	// Create vision processor
	visionConfig := &VisionConfig{
		MaxImageSize:     5 * 1024 * 1024, // 5MB
		SupportedFormats: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
		DefaultQuality:   ZAIVisionQualityAuto,
		DefaultDetail:    ZAIVisionDetailAuto,
		Timeout:          30 * time.Second,
		EnableCaching:    false,
		CacheTTL:         time.Hour,
	}

	visionProcessor := NewVisionProcessor(model, model.GetLogger(), visionConfig)

	// Test fashion analysis with a simple test image
	testImageBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="

	analysis, err := visionProcessor.AnalyzeFashionImageFromBase64(ctx, testImageBase64, "image/png")
	if err != nil {
		t.Fatalf("Failed to analyze fashion image: %v", err)
	}

	// Validate fashion analysis
	if analysis == nil {
		t.Fatal("Fashion analysis is nil")
	}

	if analysis.Description == "" {
		t.Error("Fashion analysis description is empty")
	}

	if analysis.Confidence <= 0 {
		t.Error("Fashion analysis confidence should be positive")
	}

	t.Logf("Fashion analysis - Description: %s", analysis.Description)
	t.Logf("Fashion analysis - Confidence: %f", analysis.Confidence)
	t.Logf("Fashion analysis - Clothing items: %d", len(analysis.ClothingItems))
	t.Logf("Fashion analysis - Colors: %v", analysis.Colors)
	t.Logf("Fashion analysis - Style: %s", analysis.Style)

	t.Log("Fashion vision analysis test passed")
}

// testToolCalling tests tool calling capabilities
func testToolCalling(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model
	model := NewZAIModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
		"max_tokens":  500,
		"temperature": 0.1,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Define a simple fashion tool
	fashionTool := interfaces.PonchoToolDef{
		Name:        "analyze_fashion_item",
		Description: "Analyze a fashion item and provide detailed information",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"item_description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the fashion item to analyze",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Category of the fashion item (e.g., shirt, pants, dress)",
				},
			},
			"required": []string{"item_description"},
		},
	}

	// Test tool calling
	req := &interfaces.PonchoModelRequest{
		Model: "glm-4.6",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Analyze this blue cotton t-shirt with round neck",
					},
				},
			},
		},
		Tools:     []*interfaces.PonchoToolDef{&fashionTool},
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
	model := NewZAIModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
		"max_tokens":  500,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	err := model.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Failed to initialize model: %v", err)
	}

	// Test with empty messages (should fail)
	req := &interfaces.PonchoModelRequest{
		Model:    "glm-4.6",
		Messages: []*interfaces.PonchoMessage{}, // Empty messages
	}

	_, err = model.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error with empty messages")
	} else {
		t.Logf("Expected error with empty messages: %v", err)
	}

	// Test with invalid API key (should fail)
	invalidModelConfig := map[string]interface{}{
		"api_key":     "invalid-api-key",
		"model_name":  "glm-4.6",
		"max_tokens":  500,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	invalidModelInstance := NewZAIModel()
	defer invalidModelInstance.Shutdown(ctx)

	err = invalidModelInstance.Initialize(ctx, invalidModelConfig)
	if err != nil {
		t.Logf("Expected initialization error with invalid API key: %v", err)
	}

	// Test with very large image (should fail)
	visionModel := NewZAIVisionModel()
	defer visionModel.Shutdown(ctx)

	visionConfig := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6v-flash",
		"max_tokens":  500,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	err = visionModel.Initialize(ctx, visionConfig)
	if err != nil {
		t.Fatalf("Failed to initialize vision model: %v", err)
	}

	// Create a large base64 image (this will be oversized)
	largeImageData := make([]byte, 20*1024*1024) // 20MB of data
	for i := range largeImageData {
		largeImageData[i] = byte(i % 256)
	}
	largeImageBase64 := base64.StdEncoding.EncodeToString(largeImageData)

	reqWithLargeImage := &interfaces.PonchoModelRequest{
		Model: "glm-4.6v",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "What do you see in this image?",
					},
					{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL:      "data:image/png;base64," + largeImageBase64,
							MimeType: "image/png",
						},
					},
				},
			},
		},
		MaxTokens: intPtr(100),
	}

	_, err = visionModel.Generate(ctx, reqWithLargeImage)
	if err != nil {
		t.Logf("Expected error with large image: %v", err)
	}

	t.Log("Error handling test passed")
}

// testTimeoutHandling tests timeout handling
func testTimeoutHandling(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model with very short timeout
	model := NewZAIModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
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
		Model: "glm-4.6",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Write a very long detailed essay about artificial intelligence, " +
							"including its history, current state, future prospects, " +
							"ethical considerations, and impact on society. " +
							"Be extremely detailed and comprehensive in your response.",
					},
				},
			},
		},
		MaxTokens:   intPtr(2000),
		Temperature: float32Ptr(0.7),
	}

	// This might timeout due to very short timeout setting
	_, err = model.Generate(timeoutCtx, req)
	if err != nil {
		t.Logf("Request timed out as expected: %v", err)
	}

	t.Log("Timeout handling test passed")
}

// testModelCapabilities tests model capabilities and metadata
func testModelCapabilities(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Test text model
	textModel := NewZAIModel()
	defer textModel.Shutdown(ctx)

	textConfig := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
		"max_tokens":  500,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	err := textModel.Initialize(ctx, textConfig)
	if err != nil {
		t.Fatalf("Failed to initialize text model: %v", err)
	}

	// Test vision model
	visionModel := NewZAIVisionModel()
	defer visionModel.Shutdown(ctx)

	visionConfig := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6v",
		"max_tokens":  500,
		"temperature": 0.7,
		"timeout":     30 * time.Second,
	}

	err = visionModel.Initialize(ctx, visionConfig)
	if err != nil {
		t.Fatalf("Failed to initialize vision model: %v", err)
	}

	// Test text model capabilities
	textCapabilities := textModel.GetCapabilities()
	if !textCapabilities.Streaming {
		t.Error("Z.AI text model should support streaming")
	}
	if !textCapabilities.Tools {
		t.Error("Z.AI text model should support tools")
	}
	if textCapabilities.Vision {
		t.Error("Z.AI text model should not support vision")
	}
	if !textCapabilities.System {
		t.Error("Z.AI text model should support system role")
	}

	// Test vision model capabilities
	visionCapabilities := visionModel.GetCapabilities()
	if !visionCapabilities.Streaming {
		t.Error("Z.AI vision model should support streaming")
	}
	if !visionCapabilities.Tools {
		t.Error("Z.AI vision model should support tools")
	}
	if !visionCapabilities.Vision {
		t.Error("Z.AI vision model should support vision")
	}
	if !visionCapabilities.System {
		t.Error("Z.AI vision model should support system role")
	}

	t.Logf("Text model capabilities: Streaming=%v, Tools=%v, Vision=%v, System=%v",
		textCapabilities.Streaming, textCapabilities.Tools, textCapabilities.Vision, textCapabilities.System)

	t.Logf("Vision model capabilities: Streaming=%v, Tools=%v, Vision=%v, System=%v",
		visionCapabilities.Streaming, visionCapabilities.Tools, visionCapabilities.Vision, visionCapabilities.System)

	t.Log("Model capabilities test passed")
}

// testRateLimiting tests rate limiting behavior
func testRateLimiting(t *testing.T, apiKey string) {
	ctx := context.Background()

	// Create and initialize model
	model := NewZAIModel()
	defer model.Shutdown(ctx)

	config := map[string]interface{}{
		"api_key":     apiKey,
		"model_name":  "glm-4.6",
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
		Model: "glm-4.6",
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

// Helper functions are already defined in vision.go
