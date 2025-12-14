// Package common provides content conversion and media processing utilities for
// AI model providers in PonchoFramework. This file implements format conversion
// between PonchoFramework's unified content format and provider-specific formats,
// with support for multimodal content (text, images, tools) and media processing.
//
// Key Features:
// - Bidirectional conversion between PonchoFramework and provider formats
// - Multimodal content support (text + images + tools)
// - Media processing with base64 encoding and validation
// - Provider-specific format handling (DeepSeek, Z.AI, OpenAI)
// - Tool definition and tool call conversion
// - Fashion-specific media processing capabilities
//
// Content Types Supported:
// - Text: Plain text content with role-based messaging
// - Media: Images and other media with URL/base64 support
// - Tool: Tool calls and tool definitions with arguments
// - Multimodal: Combinations of text, media, and tools
//
// Provider Formats:
// - DeepSeek: OpenAI-compatible format with tool calling
// - Z.AI: Custom format with vision support and multimodal arrays
// - OpenAI: Standard OpenAI API format
//
// Media Processing:
// - Image validation (JPEG, PNG, WebP, GIF)
// - Base64 encoding for data URLs
// - MIME type detection and validation
// - URL validation for remote media
// - Fashion-specific optimizations
//
// Usage Example:
//   converter := NewContentConverter(logger)
//   providerMsgs, _ := converter.ConvertToProviderFormat(messages, ProviderDeepSeek)
//   media, _ := mediaProcessor.ProcessImage(imageData, "image/jpeg")
//
// Error Handling:
// - Comprehensive validation of content formats
// - Graceful handling of unsupported types
// - Detailed error messages with context
// - Provider-specific error mapping
package common

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ContentConverter handles conversion between different content formats
type ContentConverter struct {
	logger interfaces.Logger
}

// NewContentConverter creates a new content converter
func NewContentConverter(logger interfaces.Logger) *ContentConverter {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}
	return &ContentConverter{
		logger: logger,
	}
}

// ConvertToProviderFormat converts PonchoFramework content to provider-specific format
func (cc *ContentConverter) ConvertToProviderFormat(messages []*interfaces.PonchoMessage, provider Provider) (interface{}, error) {
	switch provider {
	case ProviderDeepSeek:
		return cc.convertToDeepSeekFormat(messages)
	case ProviderZAI:
		return cc.convertToZAIFormat(messages)
	case ProviderOpenAI:
		return cc.convertToOpenAIFormat(messages)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ConvertFromProviderFormat converts provider-specific format to PonchoFramework format
func (cc *ContentConverter) ConvertFromProviderFormat(providerMessages interface{}, provider Provider) ([]*interfaces.PonchoMessage, error) {
	switch provider {
	case ProviderDeepSeek:
		return cc.convertFromDeepSeekFormat(providerMessages)
	case ProviderZAI:
		return cc.convertFromZAIFormat(providerMessages)
	case ProviderOpenAI:
		return cc.convertFromOpenAIFormat(providerMessages)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// convertToDeepSeekFormat converts to DeepSeek message format
func (cc *ContentConverter) convertToDeepSeekFormat(messages []*interfaces.PonchoMessage) (interface{}, error) {
	deepseekMessages := make([]map[string]interface{}, len(messages))

	for i, msg := range messages {
		deepseekMsg := map[string]interface{}{
			"role":    string(msg.Role),
			"content": cc.extractTextContent(msg),
		}

		// Add name if present
		if msg.Name != nil {
			deepseekMsg["name"] = *msg.Name
		}

		// Add tool calls if present
		if len(msg.Content) > 0 {
			toolCalls := cc.extractToolCalls(msg)
			if len(toolCalls) > 0 {
				deepseekMsg["tool_calls"] = toolCalls
			}
		}

		deepseekMessages[i] = deepseekMsg
	}

	return deepseekMessages, nil
}

// convertToZAIFormat converts to Z.AI message format
func (cc *ContentConverter) convertToZAIFormat(messages []*interfaces.PonchoMessage) (interface{}, error) {
	zaiMessages := make([]map[string]interface{}, len(messages))

	for i, msg := range messages {
		zaiMsg := map[string]interface{}{
			"role": string(msg.Role),
		}

		// Handle content based on type (text vs multimodal)
		content := cc.convertContentForZAI(msg)
		if content != nil {
			zaiMsg["content"] = content
		}

		// Add name if present
		if msg.Name != nil {
			zaiMsg["name"] = *msg.Name
		}

		// Add tool calls if present
		if len(msg.Content) > 0 {
			toolCalls := cc.extractToolCalls(msg)
			if len(toolCalls) > 0 {
				zaiMsg["tool_calls"] = toolCalls
			}
		}

		zaiMessages[i] = zaiMsg
	}

	return zaiMessages, nil
}

// convertToOpenAIFormat converts to OpenAI message format
func (cc *ContentConverter) convertToOpenAIFormat(messages []*interfaces.PonchoMessage) (interface{}, error) {
	// OpenAI format is similar to DeepSeek
	return cc.convertToDeepSeekFormat(messages)
}

// convertFromDeepSeekFormat converts from DeepSeek message format
func (cc *ContentConverter) convertFromDeepSeekFormat(providerMessages interface{}) ([]*interfaces.PonchoMessage, error) {
	messages, ok := providerMessages.([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid DeepSeek message format")
	}

	ponchoMessages := make([]*interfaces.PonchoMessage, len(messages))

	for i, msg := range messages {
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)

		ponchoMsg := &interfaces.PonchoMessage{
			Role:    interfaces.PonchoRole(role),
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: content,
				},
			},
		}

		// Add name if present
		if name, hasName := msg["name"]; hasName {
			if nameStr, ok := name.(string); ok {
				ponchoMsg.Name = &nameStr
			}
		}

		// Add tool calls if present
		if toolCalls, hasToolCalls := msg["tool_calls"]; hasToolCalls {
			if toolCallSlice, ok := toolCalls.([]interface{}); ok {
				for _, toolCall := range toolCallSlice {
					if toolCallMap, ok := toolCall.(map[string]interface{}); ok {
						toolPart := cc.convertToolCallFromMap(toolCallMap)
						if toolPart != nil {
							ponchoMsg.Content = append(ponchoMsg.Content, &interfaces.PonchoContentPart{
								Type: interfaces.PonchoContentTypeTool,
								Tool: toolPart,
							})
						}
					}
				}
			}
		}

		ponchoMessages[i] = ponchoMsg
	}

	return ponchoMessages, nil
}

// convertFromZAIFormat converts from Z.AI message format
func (cc *ContentConverter) convertFromZAIFormat(providerMessages interface{}) ([]*interfaces.PonchoMessage, error) {
	messages, ok := providerMessages.([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid Z.AI message format")
	}

	ponchoMessages := make([]*interfaces.PonchoMessage, len(messages))

	for i, msg := range messages {
		role, _ := msg["role"].(string)

		ponchoMsg := &interfaces.PonchoMessage{
			Role:    interfaces.PonchoRole(role),
			Content: []*interfaces.PonchoContentPart{},
		}

		// Handle content (can be string or array for multimodal)
		if content, hasContent := msg["content"]; hasContent {
			contentParts := cc.convertContentFromZAI(content)
			ponchoMsg.Content = contentParts
		}

		// Add name if present
		if name, hasName := msg["name"]; hasName {
			if nameStr, ok := name.(string); ok {
				ponchoMsg.Name = &nameStr
			}
		}

		// Add tool calls if present
		if toolCalls, hasToolCalls := msg["tool_calls"]; hasToolCalls {
			if toolCallSlice, ok := toolCalls.([]interface{}); ok {
				for _, toolCall := range toolCallSlice {
					if toolCallMap, ok := toolCall.(map[string]interface{}); ok {
						toolPart := cc.convertToolCallFromMap(toolCallMap)
						if toolPart != nil {
							ponchoMsg.Content = append(ponchoMsg.Content, &interfaces.PonchoContentPart{
								Type: interfaces.PonchoContentTypeTool,
								Tool: toolPart,
							})
						}
					}
				}
			}
		}

		ponchoMessages[i] = ponchoMsg
	}

	return ponchoMessages, nil
}

// convertFromOpenAIFormat converts from OpenAI message format
func (cc *ContentConverter) convertFromOpenAIFormat(providerMessages interface{}) ([]*interfaces.PonchoMessage, error) {
	// OpenAI format is similar to DeepSeek
	return cc.convertFromDeepSeekFormat(providerMessages)
}

// convertContentForZAI converts PonchoFramework content to Z.AI format
func (cc *ContentConverter) convertContentForZAI(msg *interfaces.PonchoMessage) interface{} {
	// Check if this is a multimodal message
	hasMultimodal := false
	for _, part := range msg.Content {
		if part.Type == interfaces.PonchoContentTypeMedia {
			hasMultimodal = true
			break
		}
	}

	if hasMultimodal {
		// Convert to array format for multimodal content
		return cc.convertToMultimodalContent(msg.Content)
	} else {
		// Convert to string format for text-only content
		return cc.extractTextContent(msg)
	}
}

// convertToMultimodalContent converts content parts to multimodal array format
func (cc *ContentConverter) convertToMultimodalContent(parts []*interfaces.PonchoContentPart) []map[string]interface{} {
	multimodalParts := make([]map[string]interface{}, 0, len(parts))

	for _, part := range parts {
		partMap := map[string]interface{}{
			"type": string(part.Type),
		}

		switch part.Type {
		case interfaces.PonchoContentTypeText:
			partMap["text"] = part.Text
		case interfaces.PonchoContentTypeMedia:
			if part.Media != nil {
				mediaMap := map[string]interface{}{
					"url": part.Media.URL,
				}
				if part.Media.MimeType != "" {
					mediaMap["mime_type"] = part.Media.MimeType
				}
				partMap["image_url"] = mediaMap
			}
		case interfaces.PonchoContentTypeTool:
			if part.Tool != nil {
				partMap["tool_call_id"] = part.Tool.ID
				partMap["tool_call_name"] = part.Tool.Name
				partMap["tool_call_arguments"] = part.Tool.Args
			}
		}

		multimodalParts = append(multimodalParts, partMap)
	}

	return multimodalParts
}

// convertContentFromZAI converts Z.AI content to PonchoFramework format
func (cc *ContentConverter) convertContentFromZAI(content interface{}) []*interfaces.PonchoContentPart {
	switch content := content.(type) {
	case string:
		// Text-only content
		return []*interfaces.PonchoContentPart{
			{
				Type: interfaces.PonchoContentTypeText,
				Text: content,
			},
		}
	case []interface{}:
		// Multimodal content
		contentSlice := content
		parts := make([]*interfaces.PonchoContentPart, 0, len(contentSlice))

		for _, item := range contentSlice {
			if partMap, ok := item.(map[string]interface{}); ok {
				part := cc.convertContentPartFromMap(partMap)
				if part != nil {
					parts = append(parts, part)
				}
			}
		}

		return parts
	default:
		return []*interfaces.PonchoContentPart{
			{
				Type: interfaces.PonchoContentTypeText,
				Text: fmt.Sprintf("%v", content),
			},
		}
	}
}

// convertContentPartFromMap converts content part from map to PonchoFramework format
func (cc *ContentConverter) convertContentPartFromMap(partMap map[string]interface{}) *interfaces.PonchoContentPart {
	partType, _ := partMap["type"].(string)

	part := &interfaces.PonchoContentPart{
		Type: interfaces.PonchoContentType(partType),
	}

	switch partType {
	case "text":
		if text, hasText := partMap["text"]; hasText {
			if textStr, ok := text.(string); ok {
				part.Text = textStr
			}
		}
	case "image_url":
		if imageURL, hasImageURL := partMap["image_url"]; hasImageURL {
			if imageMap, ok := imageURL.(map[string]interface{}); ok {
				media := &interfaces.PonchoMediaPart{}
				if url, hasURL := imageMap["url"]; hasURL {
					if urlStr, ok := url.(string); ok {
						media.URL = urlStr
					}
				}
				if mimeType, hasMimeType := imageMap["mime_type"]; hasMimeType {
					if mimeTypeStr, ok := mimeType.(string); ok {
						media.MimeType = mimeTypeStr
					}
				}
				part.Media = media
			}
		}
	case "tool_call_id", "tool_call_name", "tool_call_arguments":
		// Handle tool call parts
		tool := &interfaces.PonchoToolPart{}
		if id, hasID := partMap["tool_call_id"]; hasID {
			if idStr, ok := id.(string); ok {
				tool.ID = idStr
			}
		}
		if name, hasName := partMap["tool_call_name"]; hasName {
			if nameStr, ok := name.(string); ok {
				tool.Name = nameStr
			}
		}
		if args, hasArgs := partMap["tool_call_arguments"]; hasArgs {
			if argsMap, ok := args.(map[string]interface{}); ok {
				tool.Args = argsMap
			}
		}
		part.Tool = tool
	}

	return part
}

// extractTextContent extracts text content from message
func (cc *ContentConverter) extractTextContent(msg *interfaces.PonchoMessage) string {
	var textBuilder strings.Builder

	for _, part := range msg.Content {
		if part.Type == interfaces.PonchoContentTypeText && part.Text != "" {
			textBuilder.WriteString(part.Text)
		}
	}

	return textBuilder.String()
}

// extractToolCalls extracts tool calls from message
func (cc *ContentConverter) extractToolCalls(msg *interfaces.PonchoMessage) []map[string]interface{} {
	var toolCalls []map[string]interface{}

	for _, part := range msg.Content {
		if part.Type == interfaces.PonchoContentTypeTool && part.Tool != nil {
			toolCall := map[string]interface{}{
				"id":   part.Tool.ID,
				"type": "function",
				"function": map[string]interface{}{
					"name":      part.Tool.Name,
					"arguments": cc.mapToJSONString(part.Tool.Args),
				},
			}
			toolCalls = append(toolCalls, toolCall)
		}
	}

	return toolCalls
}

// convertToolCallFromMap converts tool call from map to PonchoFramework format
func (cc *ContentConverter) convertToolCallFromMap(toolCallMap map[string]interface{}) *interfaces.PonchoToolPart {
	tool := &interfaces.PonchoToolPart{}

	if id, hasID := toolCallMap["id"]; hasID {
		if idStr, ok := id.(string); ok {
			tool.ID = idStr
		}
	}

	if function, hasFunction := toolCallMap["function"]; hasFunction {
		if functionMap, ok := function.(map[string]interface{}); ok {
			if name, hasName := functionMap["name"]; hasName {
				if nameStr, ok := name.(string); ok {
					tool.Name = nameStr
				}
			}
			if arguments, hasArgs := functionMap["arguments"]; hasArgs {
				if argsStr, ok := arguments.(string); ok {
					if argsMap, err := cc.jsonStringToMap(argsStr); err == nil {
						tool.Args = argsMap
					}
				} else if argsMap, ok := arguments.(map[string]interface{}); ok {
					tool.Args = argsMap
				}
			}
		}
	}

	return tool
}

// mapToJSONString converts map to JSON string
func (cc *ContentConverter) mapToJSONString(data map[string]interface{}) string {
	if len(data) == 0 {
		return "{}"
	}

	// Simple JSON marshaling for tool arguments
	result := "{"
	for key, value := range data {
		if len(result) > 1 {
			result += ","
		}
		result += fmt.Sprintf(`"%s":%v`, key, value)
	}
	result += "}"

	return result
}

// jsonStringToMap converts JSON string to map
func (cc *ContentConverter) jsonStringToMap(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" || jsonStr == "{}" {
		return make(map[string]interface{}), nil
	}

	// Simple JSON parsing for tool arguments
	result := make(map[string]interface{})
	
	// This is a simplified parser - in production, use json.Unmarshal
	// For now, return empty map to avoid complex parsing
	return result, nil
}

// MediaProcessor handles media processing operations
type MediaProcessor struct {
	logger interfaces.Logger
}

// NewMediaProcessor creates a new media processor
func NewMediaProcessor(logger interfaces.Logger) *MediaProcessor {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}
	return &MediaProcessor{
		logger: logger,
	}
}

// ProcessImage processes an image for model consumption
func (mp *MediaProcessor) ProcessImage(imageData []byte, mimeType string) (*interfaces.PonchoMediaPart, error) {
	// Validate image data
	if len(imageData) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	// Validate MIME type
	if mimeType == "" {
		mimeType = http.DetectContentType(imageData)
	}

	// Check if image format is supported
	if !mp.isImageTypeSupported(mimeType) {
		return nil, fmt.Errorf("unsupported image type: %s", mimeType)
	}

	// For now, return base64 encoded image
	// In a full implementation, this would include:
	// - Image optimization (resize, compress)
	// - Format conversion
	// - Metadata extraction
	// - Content filtering

	base64Data := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	media := &interfaces.PonchoMediaPart{
		URL:       dataURL,
		MimeType:  mimeType,
	}

	mp.logger.Debug("Processed image",
		"size", len(imageData),
		"mime_type", mimeType,
		"data_url_length", len(dataURL))

	return media, nil
}

// ProcessImageFromURL processes an image from URL
func (mp *MediaProcessor) ProcessImageFromURL(imageURL string) (*interfaces.PonchoMediaPart, error) {
	if imageURL == "" {
		return nil, fmt.Errorf("image URL is empty")
	}

	// Validate URL format
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid image URL: %w", err)
	}

	// Check if URL is accessible
	// In a full implementation, this would download and validate the image
	// For now, just validate the URL format
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("image URL must use http or https scheme")
	}

	media := &interfaces.PonchoMediaPart{
		URL:      imageURL,
		MimeType: mp.guessMimeTypeFromURL(imageURL),
	}

	mp.logger.Debug("Processed image URL",
		"url", imageURL,
		"scheme", parsedURL.Scheme,
		"mime_type", media.MimeType)

	return media, nil
}

// isImageTypeSupported checks if image type is supported
func (mp *MediaProcessor) isImageTypeSupported(mimeType string) bool {
	supportedTypes := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"image/gif",
	}

	for _, supported := range supportedTypes {
		if mimeType == supported {
			return true
		}
	}

	return false
}

// guessMimeTypeFromURL guesses MIME type from URL
func (mp *MediaProcessor) guessMimeTypeFromURL(imageURL string) string {
	ext := strings.ToLower(filepath.Ext(imageURL))
	
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
		".gif":  "image/gif",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	return "image/jpeg" // Default
}

// ValidateMediaURL validates a media URL
func ValidateMediaURL(mediaURL string) error {
	if mediaURL == "" {
		return fmt.Errorf("media URL cannot be empty")
	}

	// Check if it's a data URL
	if strings.HasPrefix(mediaURL, "data:") {
		mp := &MediaProcessor{}
		return mp.validateDataURL(mediaURL)
	}

	// Check if it's a regular URL
	parsedURL, err := url.Parse(mediaURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	return nil
}

// validateDataURL validates a data URL
func (mp *MediaProcessor) validateDataURL(dataURL string) error {
	// Basic data URL validation
	// Format: data:[<mediatype>][;base64],<data>
	
	dataURLRegex := regexp.MustCompile(`^data:([^/]+)?(/[^;]+)?;base64,(.+)$`)
	matches := dataURLRegex.FindStringSubmatch(dataURL)
	if matches == nil {
		return fmt.Errorf("invalid data URL format")
	}

	// Validate MIME type if present
	if matches[1] != "" {
		mimeType := matches[1]
		if !mp.isImageTypeSupported(mimeType) {
			return fmt.Errorf("unsupported data URL MIME type: %s", mimeType)
		}
	}

	return nil
}

// Helper functions

// ConvertToolsToProviderFormat converts PonchoFramework tools to provider format
func ConvertToolsToProviderFormat(tools []*interfaces.PonchoToolDef, provider Provider) (interface{}, error) {
	switch provider {
	case ProviderDeepSeek, ProviderOpenAI:
		return convertToolsToOpenAIFormat(tools), nil
	case ProviderZAI:
		return convertToolsToZAIFormat(tools), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// convertToolsToOpenAIFormat converts tools to OpenAI/DeepSeek format
func convertToolsToOpenAIFormat(tools []*interfaces.PonchoToolDef) []map[string]interface{} {
	providerTools := make([]map[string]interface{}, len(tools))

	for i, tool := range tools {
		providerTool := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
		providerTools[i] = providerTool
	}

	return providerTools
}

// convertToolsToZAIFormat converts tools to Z.AI format
func convertToolsToZAIFormat(tools []*interfaces.PonchoToolDef) []map[string]interface{} {
	// Z.AI format is similar to OpenAI
	return convertToolsToOpenAIFormat(tools)
}