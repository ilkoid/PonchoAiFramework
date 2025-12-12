package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ilkoid/PonchoAiFramework/core"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/deepseek"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY environment variable is required")
	}

	// Create framework configuration
	config := &interfaces.PonchoFrameworkConfig{
		Models: map[string]*interfaces.ModelConfig{
			"deepseek-chat": {
				Provider:     "deepseek",
				ModelName:    "deepseek-chat",
				APIKey:       apiKey,
				BaseURL:      "https://api.deepseek.com",
				MaxTokens:    2000,
				Temperature:  0.7,
				Timeout:      "30s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    false,
					System:    true,
				},
				CustomParams: map[string]interface{}{
					"top_p":             0.9,
					"frequency_penalty":   0.0,
					"presence_penalty":    0.0,
				},
			},
		},
	}

	// Create framework
	framework := core.NewPonchoFramework(config, nil)

	// Start framework
	ctx := context.Background()
	if err := framework.Start(ctx); err != nil {
		log.Fatalf("Failed to start framework: %v", err)
	}
	defer framework.Stop(ctx)

	// Create and register DeepSeek model
	model := deepseek.NewDeepSeekModel()
	if err := framework.RegisterModel("deepseek-chat", model); err != nil {
		log.Fatalf("Failed to register model: %v", err)
	}

	// Example 1: Simple text generation
	fmt.Println("=== Example 1: Simple Text Generation ===")
	simpleRequest := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Hello! Tell me a joke about programming."},
				},
			},
		},
		Temperature: func() *float32 { t := float32(0.8); return &t }(),
		MaxTokens:   func() *int { i := 150; return &i }(),
	}

	response, err := framework.Generate(ctx, simpleRequest)
	if err != nil {
		log.Printf("Generation failed: %v", err)
	} else {
		fmt.Printf("Response: %s\n", response.Message.Content[0].Text)
		fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n", 
			response.Usage.TotalTokens, response.Usage.PromptTokens, response.Usage.CompletionTokens)
	}

	// Example 2: Streaming generation
	fmt.Println("\n=== Example 2: Streaming Generation ===")
	streamingRequest := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "Write a short poem about Go programming."},
				},
			},
		},
		Stream: true,
	}

	fmt.Print("Streaming response: ")
	err = framework.GenerateStreaming(ctx, streamingRequest, func(chunk *interfaces.PonchoStreamChunk) error {
		if chunk.Delta != nil && len(chunk.Delta.Content) > 0 {
			fmt.Print(chunk.Delta.Content[0].Text)
		}
		if chunk.Done {
			fmt.Println("\n[Stream completed]")
		}
		return nil
	})

	if err != nil {
		log.Printf("Streaming generation failed: %v", err)
	}

	// Example 3: Tool calling
	fmt.Println("\n=== Example 3: Tool Calling ===")
	toolRequest := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleSystem,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "You are a helpful assistant with access to tools."},
				},
			},
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "What's the weather like today? Use the weather tool."},
				},
			},
		},
		Tools: []*interfaces.PonchoToolDef{
			{
				Name:        "get_weather",
				Description: "Get current weather information for a location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"celsius", "fahrenheit"},
							"description": "The unit of temperature",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	response, err = framework.Generate(ctx, toolRequest)
	if err != nil {
		log.Printf("Tool generation failed: %v", err)
	} else {
		fmt.Printf("Tool Response: %s\n", response.Message.Content[0].Text)
		
		// Check if tool was called
		for _, part := range response.Message.Content {
			if part.Type == interfaces.PonchoContentTypeTool {
				fmt.Printf("Tool called: %s with args: %v\n", part.Tool.Name, part.Tool.Args)
			}
		}
	}

	// Example 4: System message with context
	fmt.Println("\n=== Example 4: System Message with Context ===")
	systemRequest := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleSystem,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "You are an expert Go programmer. Provide helpful, concise answers about Go programming."},
				},
			},
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: "What's the difference between make and go build?"},
				},
			},
		},
		Temperature: func() *float32 { t := float32(0.3); return &t }(),
	}

	systemResponse, err := framework.Generate(ctx, systemRequest)
	if err != nil {
		log.Printf("System message generation failed: %v", err)
	} else {
		fmt.Printf("System Response: %s\n", systemResponse.Message.Content[0].Text)
	}

	fmt.Println("\n=== All examples completed successfully! ===")
}
