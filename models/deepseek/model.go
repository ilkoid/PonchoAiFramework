package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/common"
)

// DeepSeekModel represents a DeepSeek model implementation
type DeepSeekModel struct {
	*base.PonchoBaseModel
	client *DeepSeekClient
}

// NewDeepSeekModel creates a new DeepSeek model instance
func NewDeepSeekModel() *DeepSeekModel {
	baseModel := base.NewPonchoBaseModel("deepseek-chat", string(common.ProviderDeepSeek), interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    false,
		System:    true,
	})

	return &DeepSeekModel{
		PonchoBaseModel: baseModel,
	}
}

// Initialize initializes DeepSeek model with configuration
func (m *DeepSeekModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Convert config to CommonModelConfig
	commonConfig, err := m.convertConfig(config)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Create DeepSeek client
	client, err := NewDeepSeekClient(commonConfig, m.GetLogger())
	if err != nil {
		return fmt.Errorf("failed to create DeepSeek client: %w", err)
	}

	m.client = client

	// Initialize base model
	if err := m.PonchoBaseModel.Initialize(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize base model: %w", err)
	}

	m.GetLogger().Info("DeepSeek model initialized",
		"name", m.Name(),
		"model", commonConfig.Model,
		"max_tokens", commonConfig.MaxTokens,
		"temperature", commonConfig.Temperature)

	return nil
}

// Generate generates a response using DeepSeek API
func (m *DeepSeekModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	if !m.isInitialized() {
		return nil, fmt.Errorf("model '%s' is not initialized", m.Name())
	}

	// Validate request
	if err := m.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate DeepSeek-specific requirements
	if err := m.client.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Log request
	requestID := m.generateRequestID()
	startTime := time.Now()
	metrics := m.client.PrepareRequestMetrics(requestID, startTime)
	m.client.LogRequest(req, requestID)

	// Convert to DeepSeek request format
	deepseekReq, err := m.convertRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Make API call
	deepseekResp, err := m.createChatCompletion(ctx, deepseekReq)
	duration := time.Since(startTime)

	if err != nil {
		metrics.Success = false
		metrics.Error = &common.ErrorInfo{
			Code:      common.ErrCodeUnknown,
			Type:      "api_error",
			Message:   err.Error(),
			Provider:  common.ProviderDeepSeek,
			Retryable: true,
			Timestamp: time.Now(),
			RequestID: requestID,
		}
		m.client.LogError(err, requestID, duration)
		return nil, fmt.Errorf("DeepSeek API call failed: %w", err)
	}

	// Convert response
	resp, err := m.convertResponse(deepseekResp)
	if err != nil {
		metrics.Success = false
		metrics.Error = &common.ErrorInfo{
			Code:      common.ErrCodeParsingError,
			Type:      "response_conversion",
			Message:   err.Error(),
			Provider:  common.ProviderDeepSeek,
			Retryable: false,
			Timestamp: time.Now(),
			RequestID: requestID,
		}
		m.client.LogError(err, requestID, duration)
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	// Update metrics
	metrics.Success = true
	metrics.EndTime = time.Now()
	metrics.Duration = duration
	if resp.Usage != nil {
		metrics.TokenUsage = common.TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}
	m.client.LogResponse(resp, requestID, duration)

	m.GetLogger().Debug("DeepSeek generation completed",
		"request_id", requestID,
		"duration_ms", duration.Milliseconds(),
		"prompt_tokens", resp.Usage.PromptTokens,
		"completion_tokens", resp.Usage.CompletionTokens)

	return resp, nil
}

// GenerateStreaming generates a streaming response using DeepSeek API
func (m *DeepSeekModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	if !m.isInitialized() {
		return fmt.Errorf("model '%s' is not initialized", m.Name())
	}

	if !m.SupportsStreaming() {
		return fmt.Errorf("model '%s' does not support streaming", m.Name())
	}

	// Validate request
	if err := m.ValidateRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// Validate DeepSeek-specific requirements
	if err := m.client.ValidateRequest(req); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}

	// Log request
	requestID := m.generateRequestID()
	startTime := time.Now()
	metrics := m.client.PrepareRequestMetrics(requestID, startTime)
	m.client.LogRequest(req, requestID)

	// Convert to DeepSeek request format
	deepseekReq, err := m.convertRequest(req)
	if err != nil {
		return fmt.Errorf("failed to convert request: %w", err)
	}

	// Make streaming API call
	err = m.createChatCompletionStream(ctx, deepseekReq, func(streamResp *DeepSeekStreamResponse) error {
		// Convert stream chunk
		chunk, err := m.convertStreamChunk(streamResp)
		if err != nil {
			m.GetLogger().Error("Failed to convert stream chunk",
				"request_id", requestID,
				"error", err.Error())
			return err
		}

		// Call callback
		if err := callback(chunk); err != nil {
			m.GetLogger().Error("Stream callback error",
				"request_id", requestID,
				"error", err.Error())
			return err
		}

		return nil
	})

	duration := time.Since(startTime)

	if err != nil {
		metrics.Success = false
		metrics.Error = &common.ErrorInfo{
			Code:      common.ErrCodeUnknown,
			Type:      "stream_error",
			Message:   err.Error(),
			Provider:  common.ProviderDeepSeek,
			Retryable: true,
			Timestamp: time.Now(),
			RequestID: requestID,
		}
		m.client.LogError(err, requestID, duration)
		return fmt.Errorf("DeepSeek streaming API call failed: %w", err)
	}

	// Update metrics
	metrics.Success = true
	metrics.EndTime = time.Now()
	metrics.Duration = duration
	m.client.LogResponse(nil, requestID, duration) // No final response for streaming

	m.GetLogger().Debug("DeepSeek streaming generation completed",
		"request_id", requestID,
		"duration_ms", duration.Milliseconds())

	return nil
}

// Shutdown shuts down DeepSeek model
func (m *DeepSeekModel) Shutdown(ctx context.Context) error {
	if m.client != nil {
		if err := m.client.Close(); err != nil {
			m.GetLogger().Error("Failed to close DeepSeek client", "error", err.Error())
			return fmt.Errorf("failed to close DeepSeek client: %w", err)
		}
	}

	return m.PonchoBaseModel.Shutdown(ctx)
}

// Helper methods

// convertConfig converts generic config to CommonModelConfig
func (m *DeepSeekModel) convertConfig(config map[string]interface{}) (*common.CommonModelConfig, error) {
	commonConfig := &common.CommonModelConfig{
		Provider:    common.ProviderDeepSeek,
		Model:       common.DeepSeekDefaultModel,
		MaxTokens:   4000,
		Temperature: 0.7,
		Timeout:     30 * time.Second,
	}

	// Extract configuration values
	if apiKey, ok := config["api_key"].(string); ok {
		commonConfig.APIKey = apiKey
	}

	if model, ok := config["model_name"].(string); ok {
		commonConfig.Model = model
	}

	if maxTokens, ok := config["max_tokens"].(int); ok {
		commonConfig.MaxTokens = maxTokens
	}

	if temperature, ok := config["temperature"].(float64); ok {
		commonConfig.Temperature = float32(temperature)
	}

	if temperature, ok := config["temperature"].(float32); ok {
		commonConfig.Temperature = temperature
	}

	if baseURL, ok := config["base_url"].(string); ok {
		commonConfig.BaseURL = baseURL
	}

	// Validate required fields
	if commonConfig.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	return commonConfig, nil
}

// convertRequest converts Poncho request to DeepSeek request format
func (m *DeepSeekModel) convertRequest(req *interfaces.PonchoModelRequest) (*DeepSeekRequest, error) {
	deepseekReq := &DeepSeekRequest{
		Model:       m.client.GetConfig().Model,
		Messages:    make([]DeepSeekMessage, 0),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      req.Stream,
	}

	// Convert tools
	if len(req.Tools) > 0 {
		deepseekReq.Tools = make([]DeepSeekTool, 0)
		for _, tool := range req.Tools {
			deepseekTool := DeepSeekTool{
				Type: "function",
				Function: DeepSeekToolFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			}
			deepseekReq.Tools = append(deepseekReq.Tools, deepseekTool)
		}
	}

	// Convert messages
	for _, msg := range req.Messages {
		deepseekMsg := DeepSeekMessage{
			Role:    string(msg.Role),
			Content: "",
		}

		if msg.Name != nil {
			deepseekMsg.Name = msg.Name
		}

		// Convert content parts
		for _, part := range msg.Content {
			switch part.Type {
			case interfaces.PonchoContentTypeText:
				deepseekMsg.Content += part.Text
			case interfaces.PonchoContentTypeTool:
				if part.Tool != nil {
					toolCall := DeepSeekToolCall{
						ID:   part.Tool.ID,
						Type: "function",
						Function: DeepSeekFunctionCall{
							Name:      part.Tool.Name,
							Arguments: m.mapToJSONString(part.Tool.Args),
						},
					}
					deepseekMsg.ToolCalls = append(deepseekMsg.ToolCalls, toolCall)
				}
			default:
				return nil, fmt.Errorf("unsupported content type: %s", part.Type)
			}
		}

		deepseekReq.Messages = append(deepseekReq.Messages, deepseekMsg)
	}

	return deepseekReq, nil
}

// convertResponse converts DeepSeek response to Poncho response
func (m *DeepSeekModel) convertResponse(deepseekResp *DeepSeekResponse) (*interfaces.PonchoModelResponse, error) {
	if deepseekResp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	resp := &interfaces.PonchoModelResponse{
		FinishReason: interfaces.PonchoFinishReasonStop,
		Metadata:     make(map[string]interface{}),
	}

	// Convert message
	if len(deepseekResp.Choices) > 0 {
		choice := deepseekResp.Choices[0]
		resp.Message = &interfaces.PonchoMessage{
			Role:    interfaces.PonchoRoleAssistant,
			Content: make([]*interfaces.PonchoContentPart, 0),
		}

		// Convert content
		if choice.Message.Content != "" {
			resp.Message.Content = append(resp.Message.Content, &interfaces.PonchoContentPart{
				Type: interfaces.PonchoContentTypeText,
				Text: choice.Message.Content,
			})
		}

		// Convert tool calls
		if len(choice.Message.ToolCalls) > 0 {
			for _, toolCall := range choice.Message.ToolCalls {
				args := make(map[string]interface{})
				if toolCall.Function.Arguments != "" {
					json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				}

				resp.Message.Content = append(resp.Message.Content, &interfaces.PonchoContentPart{
					Type: interfaces.PonchoContentTypeTool,
					Tool: &interfaces.PonchoToolPart{
						ID:   toolCall.ID,
						Name: toolCall.Function.Name,
						Args: args,
					},
				})
			}
		}

		// Convert finish reason
		resp.FinishReason = common.ToPonchoFinishReason(common.FinishReason(choice.FinishReason))
	}

	// Convert usage
	resp.Usage = &interfaces.PonchoUsage{
		PromptTokens:     deepseekResp.Usage.PromptTokens,
		CompletionTokens: deepseekResp.Usage.CompletionTokens,
		TotalTokens:      deepseekResp.Usage.TotalTokens,
	}

	return resp, nil
}

// convertStreamChunk converts DeepSeek stream response to Poncho stream chunk
func (m *DeepSeekModel) convertStreamChunk(streamResp *DeepSeekStreamResponse) (*interfaces.PonchoStreamChunk, error) {
	if streamResp == nil {
		return nil, fmt.Errorf("stream response is nil")
	}

	chunk := &interfaces.PonchoStreamChunk{
		Done:     len(streamResp.Choices) > 0 && streamResp.Choices[0].FinishReason != nil && *streamResp.Choices[0].FinishReason != "",
		Metadata: make(map[string]interface{}),
	}

	// Add metadata
	chunk.Metadata["id"] = streamResp.ID
	chunk.Metadata["object"] = streamResp.Object
	chunk.Metadata["created"] = streamResp.Created
	chunk.Metadata["model"] = streamResp.Model

	// Convert delta
	if len(streamResp.Choices) > 0 {
		choice := streamResp.Choices[0]

		chunk.Delta = &interfaces.PonchoMessage{
			Role:    interfaces.PonchoRoleAssistant,
			Content: make([]*interfaces.PonchoContentPart, 0),
		}

		// Convert content delta
		if choice.Delta.Content != "" {
			chunk.Delta.Content = append(chunk.Delta.Content, &interfaces.PonchoContentPart{
				Type: interfaces.PonchoContentTypeText,
				Text: choice.Delta.Content,
			})
		}

		// Convert tool call deltas
		if len(choice.Delta.ToolCalls) > 0 {
			for _, toolCall := range choice.Delta.ToolCalls {
				args := make(map[string]interface{})
				if toolCall.Function.Arguments != "" {
					json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				}

				chunk.Delta.Content = append(chunk.Delta.Content, &interfaces.PonchoContentPart{
					Type: interfaces.PonchoContentTypeTool,
					Tool: &interfaces.PonchoToolPart{
						ID:   toolCall.ID,
						Name: toolCall.Function.Name,
						Args: args,
					},
				})
			}
		}

		// Convert finish reason
		if choice.FinishReason != nil {
			chunk.FinishReason = common.ToPonchoFinishReason(common.FinishReason(*choice.FinishReason))
		}
	}

	// Convert usage
	if streamResp.Usage != nil {
		chunk.Usage = &interfaces.PonchoUsage{
			PromptTokens:     streamResp.Usage.PromptTokens,
			CompletionTokens: streamResp.Usage.CompletionTokens,
			TotalTokens:      streamResp.Usage.TotalTokens,
		}
	}

	return chunk, nil
}

// generateRequestID generates a unique request ID
func (m *DeepSeekModel) generateRequestID() string {
	return fmt.Sprintf("deepseek_%d", time.Now().UnixNano())
}

// isInitialized checks if model is properly initialized
func (m *DeepSeekModel) isInitialized() bool {
	return m.client != nil
}

// mapToJSONString converts map to JSON string
func (m *DeepSeekModel) mapToJSONString(args map[string]interface{}) string {
	if args == nil {
		return "{}"
	}
	data, _ := json.Marshal(args)
	return string(data)
}

// createChatCompletion makes a non-streaming chat completion API call
func (m *DeepSeekModel) createChatCompletion(ctx context.Context, req *DeepSeekRequest) (*DeepSeekResponse, error) {
	// Build URL
	url := m.client.BuildURL(common.DeepSeekEndpoint)

	// Prepare headers
	headers := m.client.PrepareHeaders()

	// Create request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, common.WrapError(err, common.ErrCodeParsingError, "Failed to marshal request", string(common.ProviderDeepSeek), req.Model)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, common.WrapError(err, common.ErrorCodeNetworkError, "Failed to create request", string(common.ProviderDeepSeek), req.Model)
	}

	// Set headers
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// Make API call
	resp, err := m.client.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, common.WrapError(err, common.ErrorCodeNetworkError, "Failed to make API request", string(common.ProviderDeepSeek), req.Model)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, common.ErrorFromHTTPStatus(resp.StatusCode, string(common.ProviderDeepSeek), req.Model)
	}

	// Parse response
	var deepseekResp DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
		return nil, common.WrapError(err, common.ErrCodeParsingError, "Failed to parse response", string(common.ProviderDeepSeek), req.Model)
	}

	return &deepseekResp, nil
}

// createChatCompletionStream makes a streaming chat completion API call
func (m *DeepSeekModel) createChatCompletionStream(ctx context.Context, req *DeepSeekRequest, callback func(*DeepSeekStreamResponse) error) error {
	// Set stream flag
	req.Stream = true

	// Build URL
	url := m.client.BuildURL(common.DeepSeekEndpoint)

	// Prepare headers
	headers := m.client.PrepareHeaders()
	headers["Accept"] = common.MIMETypeEventStream

	// Create request body
	body, err := json.Marshal(req)
	if err != nil {
		return common.WrapError(err, common.ErrCodeParsingError, "Failed to marshal request", string(common.ProviderDeepSeek), req.Model)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return common.WrapError(err, common.ErrorCodeNetworkError, "Failed to create request", string(common.ProviderDeepSeek), req.Model)
	}

	// Set headers
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// Make API call
	resp, err := m.client.httpClient.Do(ctx, httpReq)
	if err != nil {
		return common.WrapError(err, common.ErrorCodeNetworkError, "Failed to make streaming request", string(common.ProviderDeepSeek), req.Model)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return common.ErrorFromHTTPStatus(resp.StatusCode, string(common.ProviderDeepSeek), req.Model)
	}

	// Process stream
	return m.processStream(ctx, resp.Body, callback)
}

// processStream processes SSE stream from DeepSeek API
func (m *DeepSeekModel) processStream(ctx context.Context, body io.ReadCloser, callback func(*DeepSeekStreamResponse) error) error {
	return ProcessSSEStream(ctx, body, callback)
}
