package deepseek

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

// DeepSeekClient represents an HTTP client for DeepSeek API
type DeepSeekClient struct {
	httpClient *http.Client
	config     *DeepSeekConfig
	logger     interfaces.Logger
}

// NewDeepSeekClient creates a new DeepSeek HTTP client
func NewDeepSeekClient(config *DeepSeekConfig, logger interfaces.Logger) *DeepSeekClient {
	if config == nil {
		config = &DeepSeekConfig{
			BaseURL:     DeepSeekDefaultBaseURL,
			Model:       DeepSeekDefaultModel,
			MaxTokens:   4000,
			Temperature: 0.7,
			Timeout:     30 * time.Second,
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

	return &DeepSeekClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// CreateChatCompletion creates a chat completion request
func (c *DeepSeekClient) CreateChatCompletion(ctx context.Context, req *DeepSeekRequest) (*DeepSeekResponse, error) {
	// Set default values
	if req.Model == "" {
		req.Model = c.config.Model
	}

	// Prepare request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.config.BaseURL + DeepSeekEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("User-Agent", "PonchoFramework/1.0")

	c.logger.Debug("Sending request to DeepSeek API",
		"url", url,
		"model", req.Model,
		"messages_count", len(req.Messages))

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
		var apiErr DeepSeekError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s (type: %s, code: %s)",
			apiErr.Error.Message, apiErr.Error.Type, apiErr.Error.Code)
	}

	// Parse successful response
	var response DeepSeekResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Debug("Received response from DeepSeek API",
		"id", response.ID,
		"model", response.Model,
		"choices_count", len(response.Choices),
		"prompt_tokens", response.Usage.PromptTokens,
		"completion_tokens", response.Usage.CompletionTokens)

	return &response, nil
}

// CreateChatCompletionStream creates a streaming chat completion request
func (c *DeepSeekClient) CreateChatCompletionStream(ctx context.Context, req *DeepSeekRequest, callback func(*DeepSeekStreamResponse) error) error {
	// Set streaming mode
	req.Stream = true

	// Set default values
	if req.Model == "" {
		req.Model = c.config.Model
	}

	// Prepare request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.config.BaseURL + DeepSeekEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("User-Agent", "PonchoFramework/1.0")

	c.logger.Debug("Sending streaming request to DeepSeek API",
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
		var apiErr DeepSeekError
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
			c.logger.Debug("Stream completed")
			break
		}

		// Parse JSON chunk
		var streamResp DeepSeekStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			c.logger.Warn("Failed to unmarshal stream chunk", "error", err, "data", data)
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

// SetConfig updates the client configuration
func (c *DeepSeekClient) SetConfig(config *DeepSeekConfig) {
	c.config = config

	// Update HTTP client timeout if needed
	if config.Timeout > 0 {
		c.httpClient.Timeout = config.Timeout
	}
}

// GetConfig returns the current client configuration
func (c *DeepSeekClient) GetConfig() *DeepSeekConfig {
	return c.config
}

// Close closes the HTTP client and cleans up resources
func (c *DeepSeekClient) Close() error {
	// Close idle connections
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// ValidateRequest validates a DeepSeek request before sending
func (c *DeepSeekClient) ValidateRequest(req *DeepSeekRequest) error {
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
		if msg.Content == "" && len(msg.ToolCalls) == 0 {
			return fmt.Errorf("message %d must have content or tool calls", i)
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

// PrepareRequest creates a DeepSeek request from configuration
func (c *DeepSeekClient) PrepareRequest() *DeepSeekRequest {
	return &DeepSeekRequest{
		Model:            c.config.Model,
		MaxTokens:        &c.config.MaxTokens,
		Temperature:      &c.config.Temperature,
		TopP:             c.config.TopP,
		FrequencyPenalty: c.config.FrequencyPenalty,
		PresencePenalty:  c.config.PresencePenalty,
		Stop:             c.config.Stop,
		ResponseFormat:   c.config.ResponseFormat,
		Thinking:         c.config.Thinking,
		LogProbs:         c.config.LogProbs,
		TopLogProbs:      c.config.TopLogProbs,
	}
}
