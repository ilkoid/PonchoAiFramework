package zai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// GLMClient represents an HTTP client for Z.AI GLM API
type GLMClient struct {
	httpClient *http.Client
	config     *GLMConfig
	logger     interfaces.Logger
}

// NewGLMClient creates a new GLM HTTP client
func NewGLMClient(config *GLMConfig, logger interfaces.Logger) *GLMClient {
	if config == nil {
		config = &GLMConfig{
			BaseURL:     GLMDefaultBaseURL,
			Model:       GLMDefaultModel,
			MaxTokens:   2000,
			Temperature: 0.5,
			Timeout:     60 * time.Second,
		}
	}

	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	// Configure HTTP client with connection pooling and timeouts
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	return &GLMClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// CreateChatCompletion creates a chat completion request
func (c *GLMClient) CreateChatCompletion(ctx context.Context, req *GLMRequest) (*GLMResponse, error) {
	// Set default values
	if req.Model == "" {
		req.Model = c.config.Model
	}

	// Validate request
	if err := c.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Prepare request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.config.BaseURL + GLMEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("User-Agent", "PonchoFramework/1.0")

	c.logger.Debug("Sending request to GLM API",
		"url", url,
		"model", req.Model,
		"messages_count", len(req.Messages),
		"stream", req.Stream,
		"thinking_enabled", req.Thinking != nil && req.Thinking.Type == GLMThinkingEnabled)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		var apiErr GLMError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s (type: %s, code: %s)",
			apiErr.Error.Message, apiErr.Error.Type, apiErr.Error.Code)
	}

	// Parse successful response
	var response GLMResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Debug("Received response from GLM API",
		"id", response.ID,
		"model", response.Model,
		"choices_count", len(response.Choices),
		"prompt_tokens", response.Usage.PromptTokens,
		"completion_tokens", response.Usage.CompletionTokens,
		"finish_reason", response.Choices[0].FinishReason)

	return &response, nil
}

// CreateChatCompletionStream creates a streaming chat completion request
func (c *GLMClient) CreateChatCompletionStream(ctx context.Context, req *GLMRequest, callback func(*GLMStreamResponse) error) error {
	// Set streaming mode
	req.Stream = true

	// Set default values
	if req.Model == "" {
		req.Model = c.config.Model
	}

	// Validate request
	if err := c.ValidateRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// Prepare request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.config.BaseURL + GLMEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers for streaming
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("User-Agent", "PonchoFramework/1.0")

	c.logger.Debug("Sending streaming request to GLM API",
		"url", url,
		"model", req.Model,
		"messages_count", len(req.Messages))

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var apiErr GLMError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("API error: %s (type: %s, code: %s)",
			apiErr.Error.Message, apiErr.Error.Type, apiErr.Error.Code)
	}

	// Process streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Remove "data: " prefix
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for end of stream
		if data == "[DONE]" {
			c.logger.Debug("GLM stream completed")
			break
		}

		// Parse JSON chunk
		var streamResp GLMStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			c.logger.Warn("Failed to unmarshal GLM stream chunk", "error", err, "data", data)
			continue
		}

		// Call callback with chunk
		if err := callback(&streamResp); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream reading error: %w", err)
	}

	return nil
}

// SetConfig updates client configuration
func (c *GLMClient) SetConfig(config *GLMConfig) {
	c.config = config

	// Update HTTP client timeout if needed
	if config.Timeout > 0 {
		c.httpClient.Timeout = config.Timeout
	}
}

// GetConfig returns current client configuration
func (c *GLMClient) GetConfig() *GLMConfig {
	return c.config
}

// Close closes HTTP client and cleans up resources
func (c *GLMClient) Close() error {
	// Close idle connections
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// ValidateRequest validates a GLM request before sending
func (c *GLMClient) ValidateRequest(req *GLMRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Messages == nil || len(req.Messages) == 0 {
		return fmt.Errorf("request must contain at least one message")
	}

	// Validate messages
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d must have a role", i)
		}

		// Validate content based on type
		if msg.Content == nil {
			return fmt.Errorf("message %d must have content", i)
		}

		// If content is a string, check if it's not empty
		if contentStr, ok := msg.Content.(string); ok {
			if contentStr == "" && len(msg.ToolCalls) == 0 {
				return fmt.Errorf("message %d must have non-empty content or tool calls", i)
			}
		}

		// If content is an array (multimodal), validate parts
		if contentParts, ok := msg.Content.([]interface{}); ok {
			if len(contentParts) == 0 && len(msg.ToolCalls) == 0 {
				return fmt.Errorf("message %d must have content parts or tool calls", i)
			}

			// Validate each content part
			for j, part := range contentParts {
				if partMap, ok := part.(map[string]interface{}); ok {
					partType, hasType := partMap["type"].(string)
					if !hasType || partType == "" {
						return fmt.Errorf("message %d part %d must have a type", i, j)
					}

					switch partType {
					case GLMContentTypeText:
						if _, hasText := partMap["text"].(string); !hasText {
							return fmt.Errorf("message %d part %d text type must have text content", i, j)
						}
					case GLMContentTypeImageURL:
						if imageURL, hasURL := partMap["image_url"].(map[string]interface{}); hasURL {
							if _, hasURLField := imageURL["url"].(string); !hasURLField {
								return fmt.Errorf("message %d part %d image_url must have url", i, j)
							}
						} else {
							return fmt.Errorf("message %d part %d image_url type must have image_url object", i, j)
						}
					default:
						return fmt.Errorf("message %d part %d has unsupported type: %s", i, j, partType)
					}
				} else {
					return fmt.Errorf("message %d part %d must be an object", i, j)
				}
			}
		}
	}

	// Validate model
	if req.Model == "" {
		req.Model = c.config.Model
	}

	// Validate max_tokens
	if req.MaxTokens != nil && *req.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	// Validate temperature
	if req.Temperature != nil && (*req.Temperature < 0 || *req.Temperature > 2) {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	// Validate top_p
	if req.TopP != nil && (*req.TopP < 0 || *req.TopP > 1) {
		return fmt.Errorf("top_p must be between 0 and 1")
	}

	return nil
}

// PrepareRequest creates a GLM request from configuration
func (c *GLMClient) PrepareRequest() *GLMRequest {
	return &GLMRequest{
		Model:            c.config.Model,
		MaxTokens:        &c.config.MaxTokens,
		Temperature:      &c.config.Temperature,
		TopP:             c.config.TopP,
		FrequencyPenalty: c.config.FrequencyPenalty,
		PresencePenalty:  c.config.PresencePenalty,
		Stop:             c.config.Stop,
		Thinking:         c.config.Thinking,
	}
}

// IsVisionModel checks if the specified model supports vision
func (c *GLMClient) IsVisionModel(model string) bool {
	return model == GLMVisionModel || strings.Contains(model, "v")
}

// SupportsThinking checks if the model supports thinking mode
func (c *GLMClient) SupportsThinking(model string) bool {
	// GLM-4.6 models support thinking mode
	return strings.Contains(model, "glm-4.6")
}
