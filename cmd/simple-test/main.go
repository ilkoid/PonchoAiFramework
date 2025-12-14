package main

// This is a simple test program demonstrating the PonchoFramework architecture.
// It shows how to initialize the framework, create prompt templates, process variables,
// and execute AI model requests. This file serves as an educational example of the
// entire PonchoFramework ecosystem, including:
// - Framework initialization and lifecycle management
// - Configuration loading from YAML files
// - Model registration and selection
// - Prompt template creation and variable processing
// - API integration patterns with DeepSeek and Z.AI models
// - Response handling and token usage tracking

import (
	"context" // Provides context for request cancellation and timeouts across the framework
	"fmt"     // Standard formatting for output and string manipulation
	"log"     // Simple logging for errors and status messages
	"os"      // Access to environment variables and system operations
	"strings" // String manipulation utilities for variable processing

	"github.com/ilkoid/PonchoAiFramework/core"       // Core framework implementation with PonchoFramework struct
	"github.com/ilkoid/PonchoAiFramework/interfaces" // All interface definitions and type aliases for the framework
)

// main is the entry point of the simple test program.
// It demonstrates the complete lifecycle of using PonchoFramework:
// 1. Environment validation (API keys)
// 2. Framework initialization with configuration loading
// 3. Framework startup and model registration
// 4. Prompt template creation with variables
// 5. Model selection based on available API keys
// 6. Prompt execution and response handling
// 7. Framework shutdown
//
// This function showcases the PonchoFramework's clean architecture where:
// - Models are registered in a central registry
// - Prompts are processed through a template system
// - Requests flow through the framework to appropriate models
// - Responses are standardized across different AI providers
func main() {
	// Environment Validation: Check for required API keys
	// PonchoFramework integrates with multiple AI providers (DeepSeek, Z.AI GLM)
	// Each provider requires its own API key for authentication
	// This ensures the test can run with at least one available model
	deepseekKey := os.Getenv("DEEPSEEK_API_KEY") // DeepSeek API key for text generation
	zaiKey := os.Getenv("ZAI_API_KEY")           // Z.AI API key for vision and multimodal tasks

	// Fatal error if no API keys are available - cannot proceed without model access
	if deepseekKey == "" && zaiKey == "" {
		log.Fatal("Either DEEPSEEK_API_KEY or ZAI_API_KEY environment variable must be set")
	}

	// Framework Initialization: Create the core PonchoFramework instance
	// The framework follows a registry pattern where models, tools, and flows are registered
	// and accessed through a central orchestrator. This provides loose coupling and extensibility.

	// Logger creation: PonchoFramework uses structured logging for observability
	// The logger implements the interfaces.Logger interface and supports multiple formats
	logger := interfaces.NewDefaultLogger() // Creates a default logger with text format and info level

	// Framework instantiation: Passing nil config means load from config.yaml file
	// The config system supports YAML/JSON with environment variable substitution
	// This allows runtime configuration without code changes
	framework := core.NewPonchoFramework(nil, logger) // nil config triggers file-based loading

	// Framework Lifecycle Management: Start/Stop pattern for resource management
	// Start() initializes all registered components, loads configurations, and establishes connections
	// Stop() gracefully shuts down all components and releases resources

	// Context creation: Used throughout the framework for request cancellation and timeouts
	// This enables graceful shutdown and prevents resource leaks in long-running operations
	ctx := context.Background() // Background context for the entire application lifecycle

	// Framework startup: This is where the magic happens
	// - Loads config.yaml and validates configuration
	// - Creates model instances (DeepSeek, Z.AI GLM) based on config
	// - Registers models in the model registry for later retrieval
	// - Initializes HTTP clients with connection pooling and retries
	// - Sets up logging and metrics collection
	if err := framework.Start(ctx); err != nil {
		log.Fatalf("Failed to start framework: %v", err) // Fatal error - cannot continue without framework
	}

	// Deferred shutdown: Ensures framework is properly stopped even if main panics
	// This releases HTTP connections, closes model clients, and cleans up resources
	defer framework.Stop(ctx)

	// Prompt Template Creation: Demonstrates the PonchoFramework's advanced prompt management system
	// Prompt templates are reusable, versioned structures that define AI conversations
	// They support variable substitution, multiple parts (system/user/assistant), and metadata
	// This enables consistent, maintainable AI interactions across the application

	// PromptTemplate structure: The core data structure for prompt definitions
	// - Name: Unique identifier for the prompt
	// - Description: Human-readable explanation of the prompt's purpose
	// - Version: Semantic versioning for prompt evolution
	// - Category: Grouping for organization (e.g., "fashion", "test")
	// - Tags: Flexible labeling for search and filtering
	// - Parts: Array of conversation parts (system instructions, user messages, etc.)
	// - Variables: Defined parameters that can be substituted in the prompt
	// - Model: Specifies which AI model to use (from the model registry)
	// - Metadata: Additional information like author, creation date, etc.
	template := &interfaces.PromptTemplate{
		Name:        "simple-test",                                    // Unique identifier
		Description: "Simple test prompt for product description",    // Purpose explanation
		Version:     "1.0.0",                                         // Semantic version
		Category:    "test",                                          // Organizational category
		Tags:        []string{"test", "simple"},                      // Searchable tags

		// Parts: Define the conversation structure
		// Each part has a type (system/user/assistant) and content
		// System parts set AI behavior, user parts provide input, assistant parts are for few-shot examples
		Parts: []*interfaces.PromptPart{
			{
				Type:    interfaces.PromptPartTypeSystem, // System instructions for AI behavior
				Content: "You are a helpful assistant that generates product descriptions.",
			},
			{
				Type:    interfaces.PromptPartTypeUser, // User input with variable placeholders
				Content: "Generate a description for a {{product}} made of {{material}}.",
			},
		},

		// Variables: Define substitutable parameters in the prompt
		// Each variable has name, type, description, and required flag
		// This enables type-safe, validated variable substitution
		Variables: []*interfaces.PromptVariable{
			{
				Name:        "product",        // Variable name (matches {{product}} in content)
				Type:        "string",         // Data type for validation
				Description: "The product name", // Documentation for developers
				Required:    true,             // Must be provided when using template
			},
			{
				Name:        "material",       // Variable name (matches {{material}} in content)
				Type:        "string",         // Data type for validation
				Description: "The material of the product", // Documentation
				Required:    true,             // Must be provided
			},
		},

		// Model specification: Which AI model to use for this prompt
		// Must match a model registered in the framework's model registry
		// Models are configured in config.yaml with provider-specific settings
		Model: "deepseek-chat", // Default model (will be overridden based on available API keys)

		// Metadata: Additional information about the prompt
		// Used for tracking, auditing, and organizational purposes
		Metadata: &interfaces.PromptMetadata{
			Author: "simple-test", // Creator attribution
		},
	}

	// Variable Preparation: Create the actual values for template substitution
	// Variables are passed as a map[string]interface{} to allow different data types
	// The prompt system validates variables against the template's variable definitions
	// This enables dynamic content generation based on runtime data
	variables := map[string]interface{}{
		"product":  "cotton t-shirt",      // Concrete value for {{product}} placeholder
		"material": "100% organic cotton", // Concrete value for {{material}} placeholder
	}

	// Model Selection Logic: Choose appropriate AI model based on available API keys
	// PonchoFramework supports multiple AI providers with different capabilities:
	// - DeepSeek: Excellent for text generation and reasoning (OpenAI-compatible API)
	// - Z.AI GLM: Specialized for vision analysis and fashion industry tasks
	// This demonstrates provider-agnostic architecture - same interface, different backends
	modelName := "deepseek-chat" // Default to DeepSeek for text generation
	if deepseekKey == "" && zaiKey != "" {
		// If only Z.AI key is available, switch to GLM model
		// GLM models can handle text generation but excel at vision tasks
		modelName = "glm-chat"
		template.Model = "glm-chat" // Update template to use GLM model
	}

	// User Feedback: Display configuration for transparency
	// Shows which model is being used and what variables will be processed
	fmt.Printf("Testing PonchoFramework with model: %s\n", modelName)
	fmt.Printf("Variables: product=%s, material=%s\n\n", variables["product"], variables["material"])

	// Prompt Execution: The core operation - sending the processed prompt to the AI model
	// This demonstrates the framework's Generate() method which:
	// 1. Retrieves the specified model from the model registry
	// 2. Validates the request parameters
	// 3. Makes the API call to the AI provider (DeepSeek or Z.AI)
	// 4. Handles retries, timeouts, and error responses
	// 5. Returns a standardized response regardless of the underlying provider
	response, err := executePromptDirectly(ctx, framework, template, variables, modelName)
	if err != nil {
		log.Fatalf("Failed to execute prompt: %v", err) // Fatal - test cannot continue without response
	}

	// Response Processing and Display: Parse and display the AI model's response
	// PonchoFramework standardizes responses across different providers
	// Response structure includes message content, token usage, and completion metadata
	fmt.Println("=== PONCHOFRAMEWORK TEST RESULT ===")

	// Message Content: Extract and display the generated text
	// Responses can contain multiple content parts (text, images, etc.)
	// This test focuses on text content from the AI assistant
	if response.Message != nil && len(response.Message.Content) > 0 {
		for _, part := range response.Message.Content {
			if part.Type == interfaces.PonchoContentTypeText {
				fmt.Printf("Response: %s\n", part.Text) // Display the generated product description
			}
		}
	}

	// Token Usage Tracking: Display resource consumption metrics
	// Important for cost monitoring and optimization
	// Shows breakdown between prompt tokens (input) and completion tokens (output)
	if response.Usage != nil {
		fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
			response.Usage.TotalTokens,    // Total tokens consumed
			response.Usage.PromptTokens,   // Tokens in the input prompt
			response.Usage.CompletionTokens) // Tokens in the generated response
	}

	// Completion Metadata: Reason why the model stopped generating
	// Common reasons: "stop" (natural completion), "length" (max tokens reached), "content_filter"
	fmt.Printf("Finish reason: %s\n", response.FinishReason)

	// Test Completion: Indicate successful execution
	fmt.Println("=== TEST COMPLETED SUCCESSFULLY ===")
}

// executePromptDirectly executes a prompt template using the framework directly.
// This function demonstrates the complete prompt processing pipeline:
// 1. Template parsing and variable substitution
// 2. Message construction for AI model consumption
// 3. Model request creation with appropriate parameters
// 4. Framework integration for model execution
//
// Parameters:
// - ctx: Context for cancellation and timeouts
// - framework: The PonchoFramework instance with registered models
// - template: The prompt template with parts and variables
// - variables: Runtime values to substitute in the template
// - modelName: Which model to use (must be registered in framework)
//
// Returns: Standardized model response or error
func executePromptDirectly(ctx context.Context, framework interfaces.PonchoFramework, template *interfaces.PromptTemplate, variables map[string]interface{}, modelName string) (*interfaces.PonchoModelResponse, error) {
	// Message Construction: Convert template parts into AI model message format
	// AI models expect conversations as arrays of messages with roles (system/user/assistant)
	// Each message contains content parts (text, images, etc.)
	// This transforms the template structure into the universal message format
	messages := make([]*interfaces.PonchoMessage, 0, len(template.Parts)) // Pre-allocate for efficiency

	// Template Processing Loop: Iterate through each part of the prompt template
	for _, part := range template.Parts {
		// Variable Substitution: Replace {{variable}} placeholders with actual values
		// This enables dynamic prompt generation based on runtime data
		// The processVariables function handles string replacement and type conversion
		content := processVariables(part.Content, variables)

		// Role Mapping: Convert template part types to AI model roles
		// Different AI providers use different role names, but PonchoFramework standardizes them
		// - System: Instructions for AI behavior (not user input)
		// - User: Human input or requests
		// - Assistant: AI responses (used for few-shot examples)
		var role interfaces.PonchoRole
		switch part.Type {
		case interfaces.PromptPartTypeSystem:
			role = interfaces.PonchoRoleSystem // AI behavior instructions
		case interfaces.PromptPartTypeUser:
			role = interfaces.PonchoRoleUser // User input with substituted variables
		default:
			continue // Skip unsupported parts (like media) in this simple test
		}

		// Message Creation: Build the standardized message structure
		// PonchoMessage is the universal format used across all AI providers
		// Content is an array to support multimodal inputs (text + images)
		message := &interfaces.PonchoMessage{
			Role: role, // Who is speaking in the conversation
			Content: []*interfaces.PonchoContentPart{ // Array of content parts
				{
					Type: interfaces.PonchoContentTypeText, // Content type (text, image, etc.)
					Text: content,                           // The actual text content
				},
			},
		}
		messages = append(messages, message) // Add to conversation
	}

	// Model Request Construction: Create the standardized request structure
	// This encapsulates all parameters needed for AI model execution
	// The framework will route this to the appropriate model based on the Model field
	request := &interfaces.PonchoModelRequest{
		Model:       modelName,           // Which model to use (from registry)
		Messages:    messages,            // The conversation to send
		Temperature: template.Temperature, // Creativity control (0.0-1.0, nil uses default)
		MaxTokens:   template.MaxTokens,   // Maximum response length (nil uses default)
		Stream:      false,               // Streaming disabled for this simple test
	}

	// Framework Execution: The core operation - send request to AI model
	// This is where the framework's power shines:
	// - Retrieves the model from the registry
	// - Validates the request
	// - Makes the API call (HTTP to DeepSeek/Z.AI)
	// - Handles retries, rate limiting, and errors
	// - Returns standardized response regardless of provider
	return framework.Generate(ctx, request)
}

// processVariables replaces {{variable}} placeholders with actual values in template content.
// This is the core variable substitution engine used throughout PonchoFramework's prompt system.
//
// How it works:
// 1. Takes template content string with {{variable}} placeholders
// 2. Iterates through provided variables map
// 3. Replaces each {{key}} with the corresponding value
// 4. Handles type conversion (string preferred, fallback to fmt.Sprintf)
//
// Example:
// Input: "Hello {{name}}, you are {{age}} years old"
// Variables: {"name": "Alice", "age": 25}
// Output: "Hello Alice, you are 25 years old"
//
// Parameters:
// - content: Template string with {{variable}} placeholders
// - variables: Map of variable names to values (any type)
//
// Returns: Processed string with variables substituted
//
// This function demonstrates:
// - Template processing patterns used in prompt management
// - Type-safe variable handling with fallbacks
// - String manipulation for dynamic content generation
func processVariables(content string, variables map[string]interface{}) string {
	result := content // Start with original content

	// Variable Substitution Loop: Process each variable in the map
	for key, value := range variables {
		// Placeholder Construction: Create the {{key}} pattern to replace
		// Double braces {{}} are used to avoid conflicts with single braces in text
		placeholder := fmt.Sprintf("{{%s}}", key)

		// Type-Aware Substitution: Handle different value types appropriately
		// Prefer string values for direct substitution
		// Fall back to string conversion for other types (int, float, etc.)
		if strValue, ok := value.(string); ok {
			// String value: Direct replacement
			result = strings.ReplaceAll(result, placeholder, strValue)
		} else {
			// Non-string value: Convert to string using fmt.Sprintf
			// This handles int, float, bool, and other types gracefully
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}

	// Return processed content with all variables substituted
	return result
}