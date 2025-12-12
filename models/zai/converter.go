package zai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// RequestConverter handles conversion from PonchoFramework to GLM format
type RequestConverter struct {
	logger interfaces.Logger
}

// NewRequestConverter creates a new request converter
func NewRequestConverter(logger interfaces.Logger) *RequestConverter {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}
	return &RequestConverter{
		logger: logger,
	}
}

// ConvertRequest converts PonchoModelRequest to GLMRequest
func (rc *RequestConverter) ConvertRequest(req *interfaces.PonchoModelRequest) (*GLMRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	glmReq := &GLMRequest{
		Model:  req.Model,
		Stream: req.Stream,
	}

	// Convert messages
	messages, err := rc.convertMessages(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}
	glmReq.Messages = messages

	// Convert parameters
	if req.Temperature != nil {
		glmReq.Temperature = req.Temperature
	}

	if req.MaxTokens != nil {
		glmReq.MaxTokens = req.MaxTokens
	}

	// Convert tools
	if req.Tools != nil && len(req.Tools) > 0 {
		tools, err := rc.convertTools(req.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tools: %w", err)
		}
		glmReq.Tools = tools
		glmReq.ToolChoice = GLMToolChoiceAuto
	}

	return glmReq, nil
}

// convertMessages converts PonchoMessage array to GLMMessage array
func (rc *RequestConverter) convertMessages(messages []*interfaces.PonchoMessage) ([]GLMMessage, error) {
	glmMessages := make([]GLMMessage, len(messages))

	for i, msg := range messages {
		glmMsg, err := rc.convertMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message %d: %w", i, err)
		}
		glmMessages[i] = glmMsg
	}

	return glmMessages, nil
}

// convertMessage converts PonchoMessage to GLMMessage
func (rc *RequestConverter) convertMessage(msg *interfaces.PonchoMessage) (GLMMessage, error) {
	glmMsg := GLMMessage{
		Role: string(msg.Role),
	}

	// Handle name
	if msg.Name != nil {
		glmMsg.Name = msg.Name
	}

	// Convert content parts
	content, err := rc.convertContentParts(msg.Content)
	if err != nil {
		return GLMMessage{}, fmt.Errorf("failed to convert content: %w", err)
	}
	glmMsg.Content = content

	// Convert tool calls
	if len(msg.Content) > 0 {
		for _, part := range msg.Content {
			if part.Type == interfaces.PonchoContentTypeTool && part.Tool != nil {
				toolCall := GLMToolCall{
					ID:   part.Tool.ID,
					Type: "function",
					Function: GLMFunctionCall{
						Name:      part.Tool.Name,
						Arguments: rc.mapToJSONString(part.Tool.Args),
					},
				}
				glmMsg.ToolCalls = append(glmMsg.ToolCalls, toolCall)
			}
		}
	}

	return glmMsg, nil
}

// convertContentParts converts PonchoContentPart array to GLM content format
func (rc *RequestConverter) convertContentParts(parts []*interfaces.PonchoContentPart) (interface{}, error) {
	// Check if this is a multimodal message (has images/media)
	hasMultimodal := false
	for _, part := range parts {
		if part.Type == interfaces.PonchoContentTypeMedia {
			hasMultimodal = true
			break
		}
	}

	if hasMultimodal {
		return rc.convertMultimodalContent(parts)
	} else {
		return rc.convertTextContent(parts)
	}
}

// convertMultimodalContent converts content parts to GLM multimodal format
func (rc *RequestConverter) convertMultimodalContent(parts []*interfaces.PonchoContentPart) ([]GLMContentPart, error) {
	var glmParts []GLMContentPart

	for _, part := range parts {
		switch part.Type {
		case interfaces.PonchoContentTypeText:
			glmParts = append(glmParts, GLMContentPart{
				Type: GLMContentTypeText,
				Text: part.Text,
			})

		case interfaces.PonchoContentTypeMedia:
			if part.Media != nil {
				// Convert media to GLM image_url format
				imageURL, err := rc.convertMediaToImageURL(part.Media)
				if err != nil {
					return nil, fmt.Errorf("failed to convert media: %w", err)
				}
				glmParts = append(glmParts, GLMContentPart{
					Type:     GLMContentTypeImageURL,
					ImageURL: imageURL,
				})
			}

		case interfaces.PonchoContentTypeTool:
			// Tool calls are handled separately in convertMessage
			continue

		default:
			rc.logger.Warn("Unsupported content type in multimodal message", "type", part.Type)
		}
	}

	return glmParts, nil
}

// convertTextContent converts text-only content parts to string
func (rc *RequestConverter) convertTextContent(parts []*interfaces.PonchoContentPart) (string, error) {
	var textBuilder strings.Builder

	for _, part := range parts {
		if part.Type == interfaces.PonchoContentTypeText {
			textBuilder.WriteString(part.Text)
		} else if part.Type == interfaces.PonchoContentTypeTool {
			// Tool calls are handled separately in convertMessage
			continue
		} else {
			rc.logger.Warn("Unsupported content type in text message", "type", part.Type)
		}
	}

	return textBuilder.String(), nil
}

// convertMediaToImageURL converts PonchoMediaPart to GLMImageURL
func (rc *RequestConverter) convertMediaToImageURL(media *interfaces.PonchoMediaPart) (*GLMImageURL, error) {
	if media.URL != "" {
		// Handle direct URL
		return &GLMImageURL{
			URL: media.URL,
		}, nil
	} else if media.URL == "" {
		// Handle base64 data - need to check if it's already base64 encoded
		if media.MimeType == "" {
			return nil, fmt.Errorf("mime type is required for base64 media data")
		}

		// For now, assume URL is provided. In a full implementation,
		// we'd need base64 data in the media part
		return nil, fmt.Errorf("base64 media data not yet supported in converter")
	}

	return nil, fmt.Errorf("media must have either URL or data")
}

// convertTools converts PonchoToolDef array to GLMTool array
func (rc *RequestConverter) convertTools(tools []*interfaces.PonchoToolDef) ([]GLMTool, error) {
	glmTools := make([]GLMTool, len(tools))

	for i, tool := range tools {
		glmTool := GLMTool{
			Type: "function",
			Function: GLMToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		}
		glmTools[i] = glmTool
	}

	return glmTools, nil
}

// ResponseConverter handles conversion from GLM to PonchoFramework format
type ResponseConverter struct {
	logger interfaces.Logger
}

// NewResponseConverter creates a new response converter
func NewResponseConverter(logger interfaces.Logger) *ResponseConverter {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}
	return &ResponseConverter{
		logger: logger,
	}
}

// ConvertResponse converts GLMResponse to PonchoModelResponse
func (rc *ResponseConverter) ConvertResponse(resp *GLMResponse) (*interfaces.PonchoModelResponse, error) {
	if resp == nil {
		return nil, fmt.Errorf("response cannot be nil")
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]

	// Convert message
	message, err := rc.convertGLMMessage(&choice.Message)
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
	finishReason := rc.convertFinishReason(choice.FinishReason)

	// Create response
	ponchoResp := &interfaces.PonchoModelResponse{
		Message:      message,
		Usage:        usage,
		FinishReason: finishReason,
		Metadata:     make(map[string]interface{}),
	}

	// Add metadata
	ponchoResp.Metadata["id"] = resp.ID
	ponchoResp.Metadata["object"] = resp.Object
	ponchoResp.Metadata["created"] = resp.Created
	ponchoResp.Metadata["model"] = resp.Model

	return ponchoResp, nil
}

// convertGLMMessage converts GLMMessage to PonchoMessage
func (rc *ResponseConverter) convertGLMMessage(msg *GLMMessage) (*interfaces.PonchoMessage, error) {
	ponchoMsg := &interfaces.PonchoMessage{
		Role: interfaces.PonchoRole(msg.Role),
	}

	// Handle name
	if msg.Name != nil {
		ponchoMsg.Name = msg.Name
	}

	// Convert content and tool calls
	var contentParts []*interfaces.PonchoContentPart

	// Handle different content types
	switch content := msg.Content.(type) {
	case string:
		// Text-only content
		if content != "" {
			contentParts = append(contentParts, &interfaces.PonchoContentPart{
				Type: interfaces.PonchoContentTypeText,
				Text: content,
			})
		}

	case []interface{}:
		// Multimodal content
		parts, err := rc.convertGLMContentParts(content)
		if err != nil {
			return nil, fmt.Errorf("failed to convert content parts: %w", err)
		}
		contentParts = append(contentParts, parts...)

	default:
		rc.logger.Warn("Unknown content type in GLM message", "type", fmt.Sprintf("%T", content))
	}

	// Add tool calls
	for _, toolCall := range msg.ToolCalls {
		args, err := rc.jsonStringToMap(toolCall.Function.Arguments)
		if err != nil {
			rc.logger.Warn("Failed to parse tool call arguments", "error", err, "arguments", toolCall.Function.Arguments)
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

// convertGLMContentParts converts GLM content parts to PonchoContentPart array
func (rc *ResponseConverter) convertGLMContentParts(parts []interface{}) ([]*interfaces.PonchoContentPart, error) {
	var ponchoParts []*interfaces.PonchoContentPart

	for _, part := range parts {
		partMap, ok := part.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("content part is not an object")
		}

		partType, hasType := partMap["type"].(string)
		if !hasType {
			return nil, fmt.Errorf("content part missing type")
		}

		switch partType {
		case GLMContentTypeText:
			if text, hasText := partMap["text"].(string); hasText {
				ponchoParts = append(ponchoParts, &interfaces.PonchoContentPart{
					Type: interfaces.PonchoContentTypeText,
					Text: text,
				})
			}

		case GLMContentTypeImageURL:
			if imageURL, hasImageURL := partMap["image_url"].(map[string]interface{}); hasImageURL {
				if url, hasURL := imageURL["url"].(string); hasURL {
					ponchoParts = append(ponchoParts, &interfaces.PonchoContentPart{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL: url,
						},
					})
				}
			}

		default:
			rc.logger.Warn("Unsupported GLM content type", "type", partType)
		}
	}

	return ponchoParts, nil
}

// convertFinishReason converts GLM finish reason to PonchoFinishReason
func (rc *ResponseConverter) convertFinishReason(reason string) interfaces.PonchoFinishReason {
	switch reason {
	case GLMFinishReasonStop:
		return interfaces.PonchoFinishReasonStop
	case GLMFinishReasonLength:
		return interfaces.PonchoFinishReasonLength
	case GLMFinishReasonTool:
		return interfaces.PonchoFinishReasonTool
	case GLMFinishReasonError:
		return interfaces.PonchoFinishReasonError
	default:
		return interfaces.PonchoFinishReasonStop
	}
}

// Helper methods for JSON conversion
func (rc *RequestConverter) mapToJSONString(data map[string]interface{}) string {
	if len(data) == 0 {
		return "{}"
	}

	if bytes, err := json.Marshal(data); err == nil {
		return string(bytes)
	}
	return "{}"
}

func (rc *ResponseConverter) jsonStringToMap(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" || jsonStr == "{}" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}
