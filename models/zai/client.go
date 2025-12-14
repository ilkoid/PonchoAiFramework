// Package zai implements Z.AI GLM API client for PonchoFramework
//
// This package provides integration with Z.AI's GLM-4.6 and GLM-4.6V models,
// supporting both text generation and vision analysis capabilities. The client
// handles authentication, request validation, and provides comprehensive
// configuration management for fashion industry applications.
//
// Key Features:
// - Multimodal support (text + vision)
// - Fashion-specific vision analysis
// - Streaming responses with SSE processing
// - Tool calling support
// - Comprehensive error handling and retry logic
// - Configurable vision parameters (quality, detail, image size)
//
// Usage:
//   client, err := NewZAIClient(config, logger)
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer client.Close()
//
//   // Use with ZAIModel for generation
//   model := NewZAIModel()
//   model.Initialize(ctx, configMap)
package zai

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/common"
)

// ZAIClient represents a client for Z.AI GLM API
//
// The client provides comprehensive access to Z.AI's GLM models with support
// for text generation, vision analysis, and tool calling. It includes built-in
// authentication, request validation, and configuration management specifically
// optimized for fashion industry use cases.
type ZAIClient struct {
	httpClient   *common.HTTPClient
	config       *common.CommonModelConfig
	logger       interfaces.Logger
	apiKey       string
	baseURL      string
	visionConfig *ZAIVisionConfig
}

// NewZAIClient creates a new Z.AI client
func NewZAIClient(config *common.CommonModelConfig, logger interfaces.Logger) (*ZAIClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create HTTP client
	httpConfig := &common.DefaultHTTPConfig
	httpConfig.Timeout = config.Timeout
	httpConfig.UserAgent = "PonchoFramework-ZAI/1.0"

	retryConfig := common.DefaultRetryConfig
	retryConfig.MaxAttempts = 3
	retryConfig.BaseDelay = 1 * time.Second
	retryConfig.MaxDelay = 30 * time.Second

	httpClient, err := common.NewHTTPClient(httpConfig, retryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create vision config
	visionConfig := &ZAIVisionConfig{
		MaxImageSize:   ZAIVisionMaxImageSize,
		SupportedTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
		Quality:        ZAIVisionQualityAuto,
		Detail:         ZAIVisionDetailAuto,
	}

	client := &ZAIClient{
		httpClient:   httpClient,
		config:       config,
		logger:       logger,
		apiKey:       config.APIKey,
		baseURL:      config.BaseURL,
		visionConfig: visionConfig,
	}

	// Use default base URL if not provided
	if client.baseURL == "" {
		client.baseURL = common.ZAIDefaultBaseURL
	}

	// Try to get API key from environment if not provided
	if client.apiKey == "" {
		client.apiKey = os.Getenv("ZAI_API_KEY")
		if client.apiKey == "" {
			return nil, fmt.Errorf("ZAI_API_KEY environment variable is required")
		}
	}

	logger.Info("Z.AI client created",
		"model", config.Model,
		"base_url", client.baseURL,
		"timeout", config.Timeout,
		"vision_enabled", config.ModelType == common.ModelTypeVision || config.ModelType == common.ModelTypeMultimodal)

	return client, nil
}

// validateConfig validates Z.AI configuration
func validateConfig(config *common.CommonModelConfig) error {
	if config.Model == "" {
		return fmt.Errorf("model name is required")
	}

	if config.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// API key can be empty if provided via environment variable
	return nil
}

// Close closes Z.AI client and cleans up resources
func (c *ZAIClient) Close() error {
	if c.httpClient != nil {
		return c.httpClient.Close()
	}
	return nil
}

// GetConfig returns client configuration
func (c *ZAIClient) GetConfig() *common.CommonModelConfig {
	return c.config
}

// GetLogger returns client logger
func (c *ZAIClient) GetLogger() interfaces.Logger {
	return c.logger
}

// GetVisionConfig returns vision configuration
func (c *ZAIClient) GetVisionConfig() *ZAIVisionConfig {
	return c.visionConfig
}

// UpdateConfig updates client configuration
func (c *ZAIClient) UpdateConfig(config *common.CommonModelConfig) error {
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	c.config = config
	if config.APIKey != "" {
		c.apiKey = config.APIKey
	}
	if config.BaseURL != "" {
		c.baseURL = config.BaseURL
	}

	c.logger.Info("Z.AI client configuration updated",
		"model", config.Model,
		"base_url", c.baseURL)

	return nil
}

// UpdateVisionConfig updates vision configuration
func (c *ZAIClient) UpdateVisionConfig(config *ZAIVisionConfig) {
	if config != nil {
		c.visionConfig = config
		c.logger.Info("Z.AI vision configuration updated",
			"max_image_size", config.MaxImageSize,
			"quality", config.Quality,
			"detail", config.Detail)
	}
}

// PrepareHeaders prepares HTTP headers for Z.AI API requests
func (c *ZAIClient) PrepareHeaders() map[string]string {
	headers := make(map[string]string)
	headers["Content-Type"] = common.MIMETypeJSON
	headers["Accept"] = common.MIMETypeJSON
	headers["Authorization"] = "Bearer " + c.apiKey
	headers["User-Agent"] = "PonchoFramework-ZAI/1.0"
	return headers
}

// BuildURL builds full URL for API endpoints
func (c *ZAIClient) BuildURL(endpoint string) string {
	return c.baseURL + endpoint
}

// ValidateRequest validates a request before sending to Z.AI API
func (c *ZAIClient) ValidateRequest(req *interfaces.PonchoModelRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Messages == nil || len(req.Messages) == 0 {
		return fmt.Errorf("request must contain at least one message")
	}

	// Validate max_tokens
	if req.MaxTokens != nil && *req.MaxTokens > c.config.MaxTokens {
		return fmt.Errorf("max_tokens (%d) exceeds model maximum (%d)", *req.MaxTokens, c.config.MaxTokens)
	}

	// Validate media content for vision models
	isVisionModel := c.config.ModelType == common.ModelTypeVision || c.config.ModelType == common.ModelTypeMultimodal
	hasMedia := false

	for i, msg := range req.Messages {
		for j, part := range msg.Content {
			if part.Type == interfaces.PonchoContentTypeMedia {
				hasMedia = true
				if !isVisionModel {
					return fmt.Errorf("media content not supported for non-vision model (message %d, part %d)", i, j)
				}

				// Validate media part
				if part.Media == nil {
					return fmt.Errorf("media part cannot be nil (message %d, part %d)", i, j)
				}

				if part.Media.URL == "" {
					return fmt.Errorf("media URL cannot be empty (message %d, part %d)", i, j)
				}
			}
		}
	}

	// Check if vision model is used with media content
	if isVisionModel && !hasMedia {
		c.logger.Warn("Vision model used without media content",
			"model", c.config.Model,
			"messages_count", len(req.Messages))
	}

	return nil
}

// GetModelCapabilities returns capabilities of Z.AI model
func (c *ZAIClient) GetModelCapabilities() *common.ModelCapabilities {
	capabilities := &common.ModelCapabilities{
		SupportsStreaming: true,
		SupportsTools:     true,
		SupportsSystem:    true,
		SupportedTypes:    []common.ContentType{common.ContentTypeText},
		MaxInputTokens:    c.config.MaxTokens,
		MaxOutputTokens:   c.config.MaxTokens,
	}

	// Add vision support for vision models
	if c.config.ModelType == common.ModelTypeVision || c.config.ModelType == common.ModelTypeMultimodal {
		capabilities.SupportsVision = true
		capabilities.SupportedTypes = append(capabilities.SupportedTypes, common.ContentTypeImageURL, common.ContentTypeMedia)
	}

	return capabilities
}

// GetModelMetadata returns metadata about Z.AI model
func (c *ZAIClient) GetModelMetadata() *common.ModelMetadata {
	metadata := &common.ModelMetadata{
		Provider:        common.ProviderZAI,
		Model:           c.config.Model,
		Capabilities:    *c.GetModelCapabilities(),
		Version:         "1.0",
		Description:     "Z.AI GLM - Advanced multimodal language model with vision capabilities",
		CostPer1KTokens: 0.002, // Example cost
	}

	// Set model type based on model name
	if c.config.Model == ZAIVisionModel {
		metadata.ModelType = common.ModelTypeVision
	} else if c.config.ModelType == common.ModelTypeMultimodal {
		metadata.ModelType = common.ModelTypeMultimodal
	} else {
		metadata.ModelType = common.ModelTypeText
	}

	return metadata
}

// IsHealthy checks if Z.AI client is healthy
func (c *ZAIClient) IsHealthy(ctx context.Context) error {
	// Simple health check - try to validate API key
	if c.apiKey == "" {
		return fmt.Errorf("API key is not configured")
	}

	if c.baseURL == "" {
		return fmt.Errorf("base URL is not configured")
	}

	return nil
}

// GetRateLimitInfo returns current rate limit information
func (c *ZAIClient) GetRateLimitInfo() *common.RateLimitInfo {
	// Z.AI rate limits (example values - should be updated based on actual API docs)
	return &common.RateLimitInfo{
		RequestsPerMinute: 120,
		TokensPerMinute:   200000,
	}
}

// PrepareRequestMetrics creates metrics for a request
func (c *ZAIClient) PrepareRequestMetrics(requestID string, startTime time.Time) *common.RequestMetrics {
	return &common.RequestMetrics{
		RequestID: requestID,
		Provider:  common.ProviderZAI,
		Model:     c.config.Model,
		StartTime: startTime,
		Success:   false, // Will be updated after request completes
	}
}

// LogRequest logs a request to Z.AI API
func (c *ZAIClient) LogRequest(req *interfaces.PonchoModelRequest, requestID string) {
	mediaCount := 0
	for _, msg := range req.Messages {
		for _, part := range msg.Content {
			if part.Type == interfaces.PonchoContentTypeMedia {
				mediaCount++
			}
		}
	}

	c.logger.Debug("Z.AI API request",
		"request_id", requestID,
		"model", c.config.Model,
		"messages_count", len(req.Messages),
		"media_count", mediaCount,
		"max_tokens", req.MaxTokens,
		"temperature", req.Temperature,
		"stream", req.Stream,
		"tools_count", len(req.Tools))
}

// LogResponse logs a response from Z.AI API
func (c *ZAIClient) LogResponse(resp *interfaces.PonchoModelResponse, requestID string, duration time.Duration) {
	if resp != nil && resp.Usage != nil {
		c.logger.Debug("Z.AI API response",
			"request_id", requestID,
			"duration_ms", duration.Milliseconds(),
			"prompt_tokens", resp.Usage.PromptTokens,
			"completion_tokens", resp.Usage.CompletionTokens,
			"total_tokens", resp.Usage.TotalTokens,
			"finish_reason", resp.FinishReason)
	} else {
		c.logger.Debug("Z.AI API response",
			"request_id", requestID,
			"duration_ms", duration.Milliseconds(),
			"finish_reason", resp.FinishReason)
	}
}

// LogError logs an error from Z.AI API
func (c *ZAIClient) LogError(err error, requestID string, duration time.Duration) {
	c.logger.Error("Z.AI API error",
		"request_id", requestID,
		"duration_ms", duration.Milliseconds(),
		"error", err.Error())

	if modelErr, ok := err.(*common.ModelError); ok {
		c.logger.Error("Z.AI API model error details",
			"request_id", requestID,
			"error_code", modelErr.Code,
			"error_provider", modelErr.Provider,
			"error_model", modelErr.Model,
			"error_retryable", modelErr.Retryable,
			"error_status_code", modelErr.StatusCode)
	}
}
