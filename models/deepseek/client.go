package deepseek

import (
	"context"
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/common"
)

// DeepSeekClient represents a client for DeepSeek API
type DeepSeekClient struct {
	httpClient *common.HTTPClient
	config     *common.CommonModelConfig
	logger     interfaces.Logger
	apiKey     string
	baseURL    string
}

// NewDeepSeekClient creates a new DeepSeek client
func NewDeepSeekClient(config *common.CommonModelConfig, logger interfaces.Logger) (*DeepSeekClient, error) {
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
	httpConfig.UserAgent = "PonchoFramework-DeepSeek/1.0"

	retryConfig := common.DefaultRetryConfig
	retryConfig.MaxAttempts = 3
	retryConfig.BaseDelay = 1 * time.Second
	retryConfig.MaxDelay = 30 * time.Second

	httpClient, err := common.NewHTTPClient(httpConfig, retryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	client := &DeepSeekClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
		apiKey:     config.APIKey,
		baseURL:    config.BaseURL,
	}

	// Use default base URL if not provided
	if client.baseURL == "" {
		client.baseURL = common.DeepSeekDefaultBaseURL
	}

	logger.Info("DeepSeek client created",
		"model", config.Model,
		"base_url", client.baseURL,
		"timeout", config.Timeout)

	return client, nil
}

// validateConfig validates DeepSeek configuration
func validateConfig(config *common.CommonModelConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

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

	return nil
}

// Close closes the DeepSeek client and cleans up resources
func (c *DeepSeekClient) Close() error {
	if c.httpClient != nil {
		return c.httpClient.Close()
	}
	return nil
}

// GetConfig returns the client configuration
func (c *DeepSeekClient) GetConfig() *common.CommonModelConfig {
	return c.config
}

// GetLogger returns the client logger
func (c *DeepSeekClient) GetLogger() interfaces.Logger {
	return c.logger
}

// UpdateConfig updates the client configuration
func (c *DeepSeekClient) UpdateConfig(config *common.CommonModelConfig) error {
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	c.config = config
	c.apiKey = config.APIKey
	if config.BaseURL != "" {
		c.baseURL = config.BaseURL
	}

	c.logger.Info("DeepSeek client configuration updated",
		"model", config.Model,
		"base_url", c.baseURL)

	return nil
}

// PrepareHeaders prepares HTTP headers for DeepSeek API requests
func (c *DeepSeekClient) PrepareHeaders() map[string]string {
	headers := make(map[string]string)
	headers["Content-Type"] = common.MIMETypeJSON
	headers["Accept"] = common.MIMETypeJSON
	headers["Authorization"] = "Bearer " + c.apiKey
	headers["User-Agent"] = "PonchoFramework-DeepSeek/1.0"
	return headers
}

// BuildURL builds the full URL for API endpoints
func (c *DeepSeekClient) BuildURL(endpoint string) string {
	return c.baseURL + endpoint
}

// ValidateRequest validates a request before sending to DeepSeek API
func (c *DeepSeekClient) ValidateRequest(req *interfaces.PonchoModelRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Messages == nil {
		return fmt.Errorf("request must contain at least one message")
	}

	// Validate max_tokens
	if req.MaxTokens != nil && *req.MaxTokens > c.config.MaxTokens {
		return fmt.Errorf("max_tokens (%d) exceeds model maximum (%d)", *req.MaxTokens, c.config.MaxTokens)
	}

	// DeepSeek doesn't support vision
	for i, msg := range req.Messages {
		for j, part := range msg.Content {
			if part.Type == interfaces.PonchoContentTypeMedia {
				return fmt.Errorf("DeepSeek does not support media content (message %d, part %d)", i, j)
			}
		}
	}

	return nil
}

// GetModelCapabilities returns the capabilities of DeepSeek model
func (c *DeepSeekClient) GetModelCapabilities() *common.ModelCapabilities {
	return &common.ModelCapabilities{
		SupportsStreaming:  true,
		SupportsTools:     true,
		SupportsVision:     false,
		SupportsSystem:     true,
		SupportedTypes:     []common.ContentType{common.ContentTypeText},
		MaxInputTokens:     c.config.MaxTokens,
		MaxOutputTokens:    c.config.MaxTokens,
	}
}

// GetModelMetadata returns metadata about the DeepSeek model
func (c *DeepSeekClient) GetModelMetadata() *common.ModelMetadata {
	return &common.ModelMetadata{
		Provider:     common.ProviderDeepSeek,
		Model:        c.config.Model,
		ModelType:    common.ModelTypeText,
		Capabilities: *c.GetModelCapabilities(),
		Version:      "1.0",
		Description:  "DeepSeek Chat - Advanced language model for text generation and reasoning",
		CostPer1KTokens: 0.001, // Example cost
	}
}

// IsHealthy checks if the DeepSeek client is healthy
func (c *DeepSeekClient) IsHealthy(ctx context.Context) error {
	// Simple health check - try to validate API key
	// In practice, you might want to call a specific health endpoint
	if c.apiKey == "" {
		return fmt.Errorf("API key is not configured")
	}

	if c.baseURL == "" {
		return fmt.Errorf("base URL is not configured")
	}

	return nil
}

// GetRateLimitInfo returns current rate limit information
func (c *DeepSeekClient) GetRateLimitInfo() *common.RateLimitInfo {
	// DeepSeek rate limits (example values - should be updated based on actual API docs)
	return &common.RateLimitInfo{
		RequestsPerMinute: 60,
		TokensPerMinute:   100000,
	}
}

// PrepareRequestMetrics creates metrics for a request
func (c *DeepSeekClient) PrepareRequestMetrics(requestID string, startTime time.Time) *common.RequestMetrics {
	return &common.RequestMetrics{
		RequestID:  requestID,
		Provider:   common.ProviderDeepSeek,
		Model:      c.config.Model,
		StartTime:  startTime,
		Success:    false, // Will be updated after request completes
	}
}

// LogRequest logs a request to DeepSeek API
func (c *DeepSeekClient) LogRequest(req *interfaces.PonchoModelRequest, requestID string) {
	c.logger.Debug("DeepSeek API request",
		"request_id", requestID,
		"model", c.config.Model,
		"messages_count", len(req.Messages),
		"max_tokens", req.MaxTokens,
		"temperature", req.Temperature,
		"stream", req.Stream,
		"tools_count", len(req.Tools))
}

// LogResponse logs a response from DeepSeek API
func (c *DeepSeekClient) LogResponse(resp *interfaces.PonchoModelResponse, requestID string, duration time.Duration) {
	if resp != nil && resp.Usage != nil {
		c.logger.Debug("DeepSeek API response",
			"request_id", requestID,
			"duration_ms", duration.Milliseconds(),
			"prompt_tokens", resp.Usage.PromptTokens,
			"completion_tokens", resp.Usage.CompletionTokens,
			"total_tokens", resp.Usage.TotalTokens,
			"finish_reason", resp.FinishReason)
	} else {
		c.logger.Debug("DeepSeek API response",
			"request_id", requestID,
			"duration_ms", duration.Milliseconds(),
			"finish_reason", resp.FinishReason)
	}
}

// LogError logs an error from DeepSeek API
func (c *DeepSeekClient) LogError(err error, requestID string, duration time.Duration) {
	c.logger.Error("DeepSeek API error",
		"request_id", requestID,
		"duration_ms", duration.Milliseconds(),
		"error", err.Error())

	if modelErr, ok := err.(*common.ModelError); ok {
		c.logger.Error("DeepSeek API model error details",
			"request_id", requestID,
			"error_code", modelErr.Code,
			"error_provider", modelErr.Provider,
			"error_model", modelErr.Model,
			"error_retryable", modelErr.Retryable,
			"error_status_code", modelErr.StatusCode)
	}
}