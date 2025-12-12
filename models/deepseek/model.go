package deepseek

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// DeepSeekModel represents a DeepSeek model implementation
type DeepSeekModel struct {
	*base.PonchoBaseModel
	client *DeepSeekClient
}

// NewDeepSeekModel creates a new DeepSeek model instance
func NewDeepSeekModel() *DeepSeekModel {
	capabilities := interfaces.ModelCapabilities{
		Streaming: true,
		Tools:     true,
		Vision:    false, // DeepSeek doesn't support vision
		System:    true,
	}

	baseModel := base.NewPonchoBaseModel("deepseek-chat", "deepseek", capabilities)

	return &DeepSeekModel{
		PonchoBaseModel: baseModel,
		client:          nil, // Will be initialized in Initialize
	}
}

// Initialize initializes the DeepSeek model with configuration
func (m *DeepSeekModel) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Initialize base model first
	if err := m.PonchoBaseModel.Initialize(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize base model: %w", err)
	}

	// Extract DeepSeek-specific configuration
	deepseekConfig, err := m.extractConfig(config)
	if err != nil {
		return fmt.Errorf("failed to extract DeepSeek config: %w", err)
	}

	// Create DeepSeek client
	m.client = NewDeepSeekClient(deepseekConfig, m.GetLogger())

	m.GetLogger().Info("DeepSeek model initialized",
		"model", deepseekConfig.Model,
		"base_url", deepseekConfig.BaseURL,
		"max_tokens", deepseekConfig.MaxTokens,
		"temperature", deepseekConfig.Temperature)

	return nil
}

// Generate generates a response using DeepSeek API
func (m *DeepSeekModel) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	// Validate request
	if err := m.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert Poncho request to DeepSeek request
	deepseekReq, err := m.convertRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Call DeepSeek API
	deepseekResp, err := m.client.CreateChatCompletion(ctx, deepseekReq)
	if err != nil {
		return nil, fmt.Errorf("DeepSeek API call failed: %w", err)
	}

	// Convert DeepSeek response to Poncho response
	ponchoResp, err := m.convertResponse(deepseekResp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	m.GetLogger().Debug("Generated response",
		"model", deepseekResp.Model,
		"finish_reason", deepseekResp.Choices[0].FinishReason,
		"prompt_tokens", deepseekResp.Usage.PromptTokens,
		"completion_tokens", deepseekResp.Usage.CompletionTokens)

	return ponchoResp, nil
}

// GenerateStreaming generates a streaming response using DeepSeek API
func (m *DeepSeekModel) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	// Validate request
	if err := m.ValidateRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// Check if streaming is supported
	if !m.SupportsStreaming() {
		return fmt.Errorf("model '%s' does not support streaming", m.Name())
	}

	// Convert Poncho request to DeepSeek request
	deepseekReq, err := m.convertRequest(req)
	if err != nil {
		return fmt.Errorf("failed to convert request: %w", err)
	}

	// Call DeepSeek streaming API
	return m.client.CreateChatCompletionStream(ctx, deepseekReq, func(streamResp *DeepSeekStreamResponse) error {
		// Convert streaming response to Poncho format
		chunk, err := m.convertStreamChunk(streamResp)
		if err != nil {
			return fmt.Errorf("failed to convert stream chunk: %w", err)
		}

		// Call the callback
		return callback(chunk)
	})
}

// Shutdown shuts down the DeepSeek model and cleans up resources
func (m *DeepSeekModel) Shutdown(ctx context.Context) error {
	// Close client if it exists
	if m.client != nil {
		if err := m.client.Close(); err != nil {
			m.GetLogger().Warn("Failed to close DeepSeek client", "error", err)
		}
		m.client = nil
	}

	// Shutdown base model
	return m.PonchoBaseModel.Shutdown(ctx)
}

// extractConfig extracts DeepSeek configuration from generic config map
func (m *DeepSeekModel) extractConfig(config map[string]interface{}) (*DeepSeekConfig, error) {
	deepseekConfig := &DeepSeekConfig{
		BaseURL:     DeepSeekDefaultBaseURL,
		Model:       DeepSeekDefaultModel,
		MaxTokens:   m.MaxTokens(),
		Temperature: m.DefaultTemperature(),
		Timeout:     30 * time.Second,
	}

	// Extract API key (required)
	if apiKey, ok := config["api_key"].(string); ok && apiKey != "" {
		deepseekConfig.APIKey = apiKey
	} else {
		return nil, fmt.Errorf("api_key is required for DeepSeek model")
	}

	// Extract optional fields
	if baseURL, ok := config["base_url"].(string); ok && baseURL != "" {
		deepseekConfig.BaseURL = baseURL
	}

	if modelName, ok := config["model_name"].(string); ok && modelName != "" {
		deepseekConfig.Model = modelName
	}

	if maxTokens, ok := config["max_tokens"].(int); ok && maxTokens > 0 {
		deepseekConfig.MaxTokens = maxTokens
	}

	if temperature, ok := config["temperature"].(float32); ok {
		deepseekConfig.Temperature = temperature
	} else if temperature, ok := config["temperature"].(float64); ok {
		deepseekConfig.Temperature = float32(temperature)
	}

	if timeoutStr, ok := config["timeout"].(string); ok && timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			deepseekConfig.Timeout = timeout
		}
	}

	// Extract custom parameters
	if customParams, ok := config["custom_params"].(map[string]interface{}); ok {
		if topP, ok := customParams["top_p"].(float32); ok {
			deepseekConfig.TopP = &topP
		} else if topP, ok := customParams["top_p"].(float64); ok {
			tp := float32(topP)
			deepseekConfig.TopP = &tp
		}

		if freqPenalty, ok := customParams["frequency_penalty"].(float32); ok {
			deepseekConfig.FrequencyPenalty = &freqPenalty
		} else if freqPenalty, ok := customParams["frequency_penalty"].(float64); ok {
			fp := float32(freqPenalty)
			deepseekConfig.FrequencyPenalty = &fp
		}

		if presPenalty, ok := customParams["presence_penalty"].(float32); ok {
			deepseekConfig.PresencePenalty = &presPenalty
		} else if presPenalty, ok := customParams["presence_penalty"].(float64); ok {
			pp := float32(presPenalty)
			deepseekConfig.PresencePenalty = &pp
		}

		if stop, ok := customParams["stop"]; ok {
			deepseekConfig.Stop = stop
		}

		// Response format
		if respFormat, ok := customParams["response_format"].(map[string]interface{}); ok {
			if formatType, ok := respFormat["type"].(string); ok {
				deepseekConfig.ResponseFormat = &DeepSeekResponseFormat{
					Type: formatType,
				}
			}
		}

		// Thinking mode
		if thinking, ok := customParams["thinking"].(map[string]interface{}); ok {
			if thinkingType, ok := thinking["type"].(string); ok {
				deepseekConfig.Thinking = &DeepSeekThinking{
					Type: thinkingType,
				}
			}
		}

		if logProbs, ok := customParams["logprobs"].(bool); ok {
			deepseekConfig.LogProbs = logProbs
		}

		if topLogProbs, ok := customParams["top_logprobs"].(int); ok {
			deepseekConfig.TopLogProbs = &topLogProbs
		}
	}

	return deepseekConfig, nil
}

// convertRequest converts PonchoModelRequest to DeepSeekRequest
func (m *DeepSeekModel) convertRequest(req *interfaces.PonchoModelRequest) (*DeepSeekRequest, error) {
	deepseekReq := m.client.PrepareRequest()

	// Convert messages
	messages := make([]DeepSeekMessage, len(req.Messages))
	for i, msg := range req.Messages {
		deepseekMsg, err := m.convertMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message %d: %w", i, err)
		}
		messages[i] = deepseekMsg
	}
	deepseekReq.Messages = messages

	// Override request-specific parameters
	if req.Model != "" {
		deepseekReq.Model = req.Model
	}

	if req.Temperature != nil {
		deepseekReq.Temperature = req.Temperature
	}

	if req.MaxTokens != nil {
		deepseekReq.MaxTokens = req.MaxTokens
	}

	deepseekReq.Stream = req.Stream

	// Convert tools
	if req.Tools != nil && len(req.Tools) > 0 {
		tools := make([]DeepSeekTool, len(req.Tools))
		for i, tool := range req.Tools {
			deepseekTool := DeepSeekTool{
				Type: "function",
				Function: DeepSeekToolFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			}
			tools[i] = deepseekTool
		}
		deepseekReq.Tools = tools
		deepseekReq.ToolChoice = DeepSeekToolChoiceAuto
	}

	return deepseekReq, nil
}

// convertMessage converts PonchoMessage to DeepSeekMessage
func (m *DeepSeekModel) convertMessage(msg *interfaces.PonchoMessage) (DeepSeekMessage, error) {
	deepseekMsg := DeepSeekMessage{
		Role: string(msg.Role),
	}

	// Handle message name
	if msg.Name != nil {
		deepseekMsg.Name = msg.Name
	}

	// Convert content parts
	var contentBuilder strings.Builder
	var toolCalls []DeepSeekToolCall

	for _, part := range msg.Content {
		switch part.Type {
		case interfaces.PonchoContentTypeText:
			contentBuilder.WriteString(part.Text)
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
				toolCalls = append(toolCalls, toolCall)
			}
		case interfaces.PonchoContentTypeMedia:
			// DeepSeek doesn't support vision, so we skip media content
			m.GetLogger().Warn("Media content skipped - DeepSeek doesn't support vision")
		}
	}

	deepseekMsg.Content = contentBuilder.String()
	deepseekMsg.ToolCalls = toolCalls

	return deepseekMsg, nil
}

// convertResponse converts DeepSeekResponse to PonchoModelResponse
func (m *DeepSeekModel) convertResponse(resp *DeepSeekResponse) (*interfaces.PonchoModelResponse, error) {
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]

	// Convert message
	ponchoMsg, err := m.convertDeepSeekMessage(&choice.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message: %w", err)
	}

	// Convert usage
	usage := &interfaces.PonchoUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}

	// Convert finish reason
	finishReason := m.convertFinishReason(choice.FinishReason)

	// Create response
	ponchoResp := m.PrepareResponse(ponchoMsg, usage, finishReason)

	// Add metadata
	ponchoResp.Metadata["id"] = resp.ID
	ponchoResp.Metadata["object"] = resp.Object
	ponchoResp.Metadata["created"] = resp.Created
	ponchoResp.Metadata["model"] = resp.Model
	ponchoResp.Metadata["system_fingerprint"] = resp.SystemFingerprint

	return ponchoResp, nil
}

// convertDeepSeekMessage converts DeepSeekMessage to PonchoMessage
func (m *DeepSeekModel) convertDeepSeekMessage(msg *DeepSeekMessage) (*interfaces.PonchoMessage, error) {
	ponchoMsg := &interfaces.PonchoMessage{
		Role: interfaces.PonchoRole(msg.Role),
	}

	// Handle name
	if msg.Name != nil {
		ponchoMsg.Name = msg.Name
	}

	// Convert content and tool calls
	var contentParts []*interfaces.PonchoContentPart

	// Add text content
	if msg.Content != "" {
		contentParts = append(contentParts, &interfaces.PonchoContentPart{
			Type: interfaces.PonchoContentTypeText,
			Text: msg.Content,
		})
	}

	// Add tool calls
	for _, toolCall := range msg.ToolCalls {
		args, err := m.jsonStringToMap(toolCall.Function.Arguments)
		if err != nil {
			m.GetLogger().Warn("Failed to parse tool call arguments", "error", err, "arguments", toolCall.Function.Arguments)
			args = make(map[string]interface{})
		}

		contentParts = append(contentParts, &interfaces.PonchoContentPart{
			Type: interfaces.PonchoContentTypeTool,
			Tool: &interfaces.PonchoToolPart{
				ID:   toolCall.ID,
				Name: toolCall.Function.Name,
				Args: args,
			},
		})
	}

	ponchoMsg.Content = contentParts
	return ponchoMsg, nil
}

// convertFinishReason converts DeepSeek finish reason to PonchoFinishReason
func (m *DeepSeekModel) convertFinishReason(reason string) interfaces.PonchoFinishReason {
	switch reason {
	case DeepSeekFinishReasonStop:
		return interfaces.PonchoFinishReasonStop
	case DeepSeekFinishReasonLength:
		return interfaces.PonchoFinishReasonLength
	case DeepSeekFinishReasonTool:
		return interfaces.PonchoFinishReasonTool
	case DeepSeekFinishReasonError:
		return interfaces.PonchoFinishReasonError
	default:
		return interfaces.PonchoFinishReasonStop
	}
}

// Helper methods for JSON conversion
func (m *DeepSeekModel) mapToJSONString(data map[string]interface{}) string {
	if len(data) == 0 {
		return "{}"
	}

	// Simple JSON marshaling - in production, you might want better error handling
	if bytes, err := json.Marshal(data); err == nil {
		return string(bytes)
	}
	return "{}"
}

func (m *DeepSeekModel) jsonStringToMap(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" || jsonStr == "{}" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}
