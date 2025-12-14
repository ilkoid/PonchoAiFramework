// Package zai provides Z.AI GLM model implementation for PonchoFramework
//
// This package implements the PonchoModel interface for Z.AI's GLM-4.6 and GLM-4.6V
// models, providing unified access to text generation, vision analysis, and tool
// calling capabilities. The implementation supports both streaming and non-streaming
// responses with comprehensive error handling and metrics collection.
//
// Key Features:
// - Multimodal model support (text + vision)
// - Fashion-specific vision analysis
// - Streaming responses with SSE processing
// - Tool calling and function execution
// - Comprehensive request/response conversion
// - Performance metrics and monitoring
// - Graceful error handling and recovery
//
// Model Types:
// - GLM-4.6: General purpose multimodal model
// - GLM-4.6V: Vision-optimized model for fashion analysis
//
// Usage:
//   model := NewZAIModel()
//   err := model.Initialize(ctx, configMap)
//   if err != nil {
//       log.Fatal(err)
//   }
//
//   resp, err := model.Generate(ctx, request)
//   // or streaming:
//   err := model.GenerateStreaming(ctx, request, callback)
package zai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/common"
)

// ZAIModel represents a Z.AI GLM model implementation
//
// This struct implements the PonchoModel interface for Z.AI's GLM models,
// providing unified access to text generation, vision analysis, and tool calling
// capabilities. It supports both streaming and non-streaming responses with
// comprehensive error handling and performance monitoring.
type ZAIModel struct {
	*base.PonchoBaseModel
	client *ZAIClient
}

// NewZAIModel creates a new Z.AI model instance
func NewZAIModel() *ZAIModel {
	baseModel := base.NewPonchoBaseModel("glm-4.6", string(common.ProviderZAI), interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    true,
		System:    true,
	})

	return &ZAIModel{
		PonchoBaseModel: baseModel,
	}
}

// NewZAIVisionModel creates a new Z.AI vision model instance
func NewZAIVisionModel() *ZAIModel {
	baseModel := base.NewPonchoBaseModel("glm-4.6v", string(common.ProviderZAI), interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    true,
		System:    true,
	})

	return &ZAIModel{
		PonchoBaseModel: baseModel,
	}
}

// Initialize initializes Z.AI model with configuration
func (m *ZAIModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Convert config to CommonModelConfig
	commonConfig, err := m.convertConfig(config)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Create Z.AI client
	client, err := NewZAIClient(commonConfig, m.GetLogger())
	if err != nil {
		return fmt.Errorf("failed to create Z.AI client: %w", err)
	}

	m.client = client

	// Initialize base model
	if err := m.PonchoBaseModel.Initialize(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize base model: %w", err)
	}

	m.GetLogger().Info("Z.AI model initialized",
		"name", m.Name(),
		"model", commonConfig.Model,
		"model_type", commonConfig.ModelType,
		"max_tokens", commonConfig.MaxTokens,
		"temperature", commonConfig.Temperature)

	return nil
}

// Generate generates a response using Z.AI API
func (m *ZAIModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	if !m.isInitialized() {
		return nil, fmt.Errorf("model '%s' is not initialized", m.Name())
	}

	// Validate request
	if err := m.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate Z.AI-specific requirements
	if err := m.client.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Log request
	requestID := m.generateRequestID()
	startTime := time.Now()
	metrics := m.client.PrepareRequestMetrics(requestID, startTime)
	m.client.LogRequest(req, requestID)

	// Convert to Z.AI request format
	zaiReq, err := m.convertRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Make API call
	zaiResp, err := m.createChatCompletion(ctx, zaiReq)
	duration := time.Since(startTime)

	if err != nil {
		metrics.Success = false
		metrics.Error = &common.ErrorInfo{
			Code:      common.ErrCodeUnknown,
			Type:      "api_error",
			Message:   err.Error(),
			Provider:  common.ProviderZAI,
			Retryable: true,
			Timestamp: time.Now(),
			RequestID: requestID,
		}
		m.client.LogError(err, requestID, duration)
		return nil, fmt.Errorf("Z.AI API call failed: %w", err)
	}

	// Convert response
	resp, err := m.convertResponse(zaiResp)
	if err != nil {
		metrics.Success = false
		metrics.Error = &common.ErrorInfo{
			Code:      common.ErrCodeParsingError,
			Type:      "response_conversion",
			Message:   err.Error(),
			Provider:  common.ProviderZAI,
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

	m.GetLogger().Debug("Z.AI generation completed",
		"request_id", requestID,
		"duration_ms", duration.Milliseconds(),
		"prompt_tokens", resp.Usage.PromptTokens,
		"completion_tokens", resp.Usage.CompletionTokens)

	return resp, nil
}

// GenerateStreaming generates a streaming response using Z.AI API
func (m *ZAIModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
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

	// Validate Z.AI-specific requirements
	if err := m.client.ValidateRequest(req); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}

	// Log request
	requestID := m.generateRequestID()
	startTime := time.Now()
	metrics := m.client.PrepareRequestMetrics(requestID, startTime)
	m.client.LogRequest(req, requestID)

	// Convert to Z.AI request format
	zaiReq, err := m.convertRequest(req)
	if err != nil {
		return fmt.Errorf("failed to convert request: %w", err)
	}

	// Make streaming API call
	err = m.createChatCompletionStream(ctx, zaiReq, func(streamResp *ZAIStreamResponse) error {
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
			Provider:  common.ProviderZAI,
			Retryable: true,
			Timestamp: time.Now(),
			RequestID: requestID,
		}
		m.client.LogError(err, requestID, duration)
		return fmt.Errorf("Z.AI streaming API call failed: %w", err)
	}

	// Update metrics
	metrics.Success = true
	metrics.EndTime = time.Now()
	metrics.Duration = duration
	m.client.LogResponse(nil, requestID, duration) // No final response for streaming

	m.GetLogger().Debug("Z.AI streaming generation completed",
		"request_id", requestID,
		"duration_ms", duration.Milliseconds())

	return nil
}

// Shutdown shuts down Z.AI model
func (m *ZAIModel) Shutdown(ctx context.Context) error {
	if m.client != nil {
		if err := m.client.Close(); err != nil {
			m.GetLogger().Error("Failed to close Z.AI client", "error", err.Error())
			return fmt.Errorf("failed to close Z.AI client: %w", err)
		}
	}

	return m.PonchoBaseModel.Shutdown(ctx)
}

// Helper methods

// convertConfig converts generic config to CommonModelConfig
func (m *ZAIModel) convertConfig(config map[string]interface{}) (*common.CommonModelConfig, error) {
	commonConfig := &common.CommonModelConfig{
		Provider:    common.ProviderZAI,
		Model:       common.ZAIDefaultModel,
		ModelType:   common.ModelTypeMultimodal,
		MaxTokens:   2000,
		Temperature: 0.5,
		Timeout:     60 * time.Second,
	}

	// Extract configuration values
	if apiKey, ok := config["api_key"].(string); ok {
		commonConfig.APIKey = apiKey
	} else {
		// Try environment variable
		commonConfig.APIKey = os.Getenv("ZAI_API_KEY")
	}

	if model, ok := config["model_name"].(string); ok {
		commonConfig.Model = model
		// Set model type based on model name
		if model == ZAIVisionModel {
			commonConfig.ModelType = common.ModelTypeVision
		}
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
		return nil, fmt.Errorf("api_key is required (set via config or ZAI_API_KEY environment variable)")
	}

	return commonConfig, nil
}

// convertRequest converts Poncho request to Z.AI request format
func (m *ZAIModel) convertRequest(req *interfaces.PonchoModelRequest) (*ZAIRequest, error) {
	zaiReq := &ZAIRequest{
		Model:       m.client.GetConfig().Model,
		Messages:    make([]ZAIMessage, 0),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      req.Stream,
	}

	// Convert tools
	if len(req.Tools) > 0 {
		zaiReq.Tools = make([]ZAITool, 0)
		for _, tool := range req.Tools {
			zaiTool := ZAITool{
				Type: "function",
				Function: ZAIToolFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			}
			zaiReq.Tools = append(zaiReq.Tools, zaiTool)
		}
	}

	// Convert messages
	for _, msg := range req.Messages {
		zaiMsg, err := m.convertMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}
		zaiReq.Messages = append(zaiReq.Messages, zaiMsg)
	}

	return zaiReq, nil
}

// convertMessage converts Poncho message to Z.AI message format
func (m *ZAIModel) convertMessage(msg *interfaces.PonchoMessage) (ZAIMessage, error) {
	zaiMsg := ZAIMessage{
		Role:    string(msg.Role),
		Content: "",
	}

	if msg.Name != nil {
		zaiMsg.Name = msg.Name
	}

	// Check if we need multimodal content
	hasMedia := false
	for _, part := range msg.Content {
		if part.Type == interfaces.PonchoContentTypeMedia {
			hasMedia = true
			break
		}
	}

	if hasMedia {
		// Convert to array of content parts
		contentParts := make([]ZAIContentPart, 0)
		for _, part := range msg.Content {
			switch part.Type {
			case interfaces.PonchoContentTypeText:
				contentParts = append(contentParts, ZAIContentPart{
					Type: ZAIContentTypeText,
					Text: part.Text,
				})
			case interfaces.PonchoContentTypeMedia:
				imageURL, err := m.convertMediaToImageURL(part.Media)
				if err != nil {
					return ZAIMessage{}, fmt.Errorf("failed to convert media: %w", err)
				}
				contentParts = append(contentParts, ZAIContentPart{
					Type:     ZAIContentTypeImageURL,
					ImageURL: imageURL,
				})
			case interfaces.PonchoContentTypeTool:
				if part.Tool != nil {
					toolCall := ZAIToolCall{
						ID:   part.Tool.ID,
						Type: "function",
						Function: ZAIFunctionCall{
							Name:      part.Tool.Name,
							Arguments: m.mapToJSONString(part.Tool.Args),
						},
					}
					zaiMsg.ToolCalls = append(zaiMsg.ToolCalls, toolCall)
				}
			default:
				return ZAIMessage{}, fmt.Errorf("unsupported content type: %s", part.Type)
			}
		}
		zaiMsg.Content = contentParts
	} else {
		// Convert to simple string content
		var content string
		for _, part := range msg.Content {
			if part.Type == interfaces.PonchoContentTypeText {
				content += part.Text
			}
		}
		zaiMsg.Content = content
	}

	return zaiMsg, nil
}

// convertMediaToImageURL converts PonchoMediaPart to ZAIImageURL
func (m *ZAIModel) convertMediaToImageURL(media *interfaces.PonchoMediaPart) (*ZAIImageURL, error) {
	if media == nil {
		return nil, fmt.Errorf("media cannot be nil")
	}

	if media.URL == "" {
		return nil, fmt.Errorf("media URL cannot be empty")
	}

	return &ZAIImageURL{
		URL:    media.URL,
		Detail: ZAIVisionDetailAuto,
	}, nil
}

// convertResponse converts Z.AI response to Poncho response
func (m *ZAIModel) convertResponse(zaiResp *ZAIResponse) (*interfaces.PonchoModelResponse, error) {
	if zaiResp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	resp := &interfaces.PonchoModelResponse{
		FinishReason: interfaces.PonchoFinishReasonStop,
		Metadata:     make(map[string]interface{}),
	}

	// Convert message
	if len(zaiResp.Choices) > 0 {
		choice := zaiResp.Choices[0]
		ponchoMsg, err := m.convertZAIMessage(&choice.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}
		resp.Message = ponchoMsg

		// Convert finish reason
		resp.FinishReason = common.ToPonchoFinishReason(common.FinishReason(choice.FinishReason))
	}

	// Convert usage
	resp.Usage = &interfaces.PonchoUsage{
		PromptTokens:     zaiResp.Usage.PromptTokens,
		CompletionTokens: zaiResp.Usage.CompletionTokens,
		TotalTokens:      zaiResp.Usage.TotalTokens,
	}

	return resp, nil
}

// convertZAIMessage converts Z.AI message to Poncho message
func (m *ZAIModel) convertZAIMessage(zaiMsg *ZAIMessage) (*interfaces.PonchoMessage, error) {
	ponchoMsg := &interfaces.PonchoMessage{
		Role:    interfaces.PonchoRole(zaiMsg.Role),
		Content: make([]*interfaces.PonchoContentPart, 0),
	}

	if zaiMsg.Name != nil {
		ponchoMsg.Name = zaiMsg.Name
	}

	// Convert content
	switch content := zaiMsg.Content.(type) {
	case string:
		if content != "" {
			ponchoMsg.Content = append(ponchoMsg.Content, &interfaces.PonchoContentPart{
				Type: interfaces.PonchoContentTypeText,
				Text: content,
			})
		}
	case []interface{}:
		// Multimodal content
		for _, part := range content {
			if partMap, ok := part.(map[string]interface{}); ok {
				contentPart, err := m.convertZAIContentPart(partMap)
				if err != nil {
					return nil, fmt.Errorf("failed to convert content part: %w", err)
				}
				ponchoMsg.Content = append(ponchoMsg.Content, contentPart)
			}
		}
	}

	// Convert tool calls
	for _, toolCall := range zaiMsg.ToolCalls {
		args := make(map[string]interface{})
		if toolCall.Function.Arguments != "" {
			json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
		}

		ponchoMsg.Content = append(ponchoMsg.Content, &interfaces.PonchoContentPart{
			Type: interfaces.PonchoContentTypeTool,
			Tool: &interfaces.PonchoToolPart{
				ID:   toolCall.ID,
				Name: toolCall.Function.Name,
				Args: args,
			},
		})
	}

	return ponchoMsg, nil
}

// convertZAIContentPart converts Z.AI content part to Poncho content part
func (m *ZAIModel) convertZAIContentPart(zaiPart map[string]interface{}) (*interfaces.PonchoContentPart, error) {
	partType, ok := zaiPart["type"].(string)
	if !ok {
		return nil, fmt.Errorf("content part missing type")
	}

	switch partType {
	case ZAIContentTypeText:
		text, ok := zaiPart["text"].(string)
		if !ok {
			return nil, fmt.Errorf("text content part missing text")
		}
		return &interfaces.PonchoContentPart{
			Type: interfaces.PonchoContentTypeText,
			Text: text,
		}, nil
	case ZAIContentTypeImageURL:
		imageURLMap, ok := zaiPart["image_url"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("image_url content part missing image_url")
		}
		url, ok := imageURLMap["url"].(string)
		if !ok {
			return nil, fmt.Errorf("image_url missing url")
		}
		return &interfaces.PonchoContentPart{
			Type: interfaces.PonchoContentTypeMedia,
			Media: &interfaces.PonchoMediaPart{
				URL:      url,
				MimeType: "image/jpeg", // Default, should be detected
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported content part type: %s", partType)
	}
}

// convertStreamChunk converts Z.AI stream response to Poncho stream chunk
func (m *ZAIModel) convertStreamChunk(streamResp *ZAIStreamResponse) (*interfaces.PonchoStreamChunk, error) {
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
func (m *ZAIModel) generateRequestID() string {
	return fmt.Sprintf("zai_%d", time.Now().UnixNano())
}

// isInitialized checks if model is properly initialized
func (m *ZAIModel) isInitialized() bool {
	return m.client != nil
}

// mapToJSONString converts map to JSON string
func (m *ZAIModel) mapToJSONString(args map[string]interface{}) string {
	if args == nil {
		return "{}"
	}
	data, _ := json.Marshal(args)
	return string(data)
}

// createChatCompletion makes a non-streaming chat completion API call
func (m *ZAIModel) createChatCompletion(ctx context.Context, req *ZAIRequest) (*ZAIResponse, error) {
	// Build URL
	url := m.client.BuildURL(common.ZAIEndpoint)

	// Prepare headers
	headers := m.client.PrepareHeaders()

	// Create request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, common.WrapError(err, common.ErrCodeParsingError, "Failed to marshal request", string(common.ProviderZAI), req.Model)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, common.WrapError(err, common.ErrorCodeNetworkError, "Failed to create request", string(common.ProviderZAI), req.Model)
	}

	// Set headers
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// Make API call
	resp, err := m.client.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, common.WrapError(err, common.ErrorCodeNetworkError, "Failed to make API request", string(common.ProviderZAI), req.Model)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, common.ErrorFromHTTPStatus(resp.StatusCode, string(common.ProviderZAI), req.Model)
	}

	// Parse response
	var zaiResp ZAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&zaiResp); err != nil {
		return nil, common.WrapError(err, common.ErrCodeParsingError, "Failed to parse response", string(common.ProviderZAI), req.Model)
	}

	return &zaiResp, nil
}

// createChatCompletionStream makes a streaming chat completion API call
func (m *ZAIModel) createChatCompletionStream(ctx context.Context, req *ZAIRequest, callback func(*ZAIStreamResponse) error) error {
	// Set stream flag
	req.Stream = true

	// Build URL
	url := m.client.BuildURL(common.ZAIEndpoint)

	// Prepare headers
	headers := m.client.PrepareHeaders()
	headers["Accept"] = common.MIMETypeEventStream

	// Create request body
	body, err := json.Marshal(req)
	if err != nil {
		return common.WrapError(err, common.ErrCodeParsingError, "Failed to marshal request", string(common.ProviderZAI), req.Model)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return common.WrapError(err, common.ErrorCodeNetworkError, "Failed to create request", string(common.ProviderZAI), req.Model)
	}

	// Set headers
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// Make API call
	resp, err := m.client.httpClient.Do(ctx, httpReq)
	if err != nil {
		return common.WrapError(err, common.ErrorCodeNetworkError, "Failed to make streaming request", string(common.ProviderZAI), req.Model)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return common.ErrorFromHTTPStatus(resp.StatusCode, string(common.ProviderZAI), req.Model)
	}

	// Process stream
	return m.processStream(ctx, resp.Body, callback)
}

// processStream processes SSE stream from Z.AI API
func (m *ZAIModel) processStream(ctx context.Context, body io.ReadCloser, callback func(*ZAIStreamResponse) error) error {
	return ProcessSSEStream(ctx, body, callback)
}
