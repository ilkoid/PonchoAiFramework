package zai

import (
	"encoding/json"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// StreamConverter handles streaming response conversion
type StreamConverter struct {
	logger interfaces.Logger
}

// NewStreamConverter creates a new stream converter
func NewStreamConverter(logger interfaces.Logger) *StreamConverter {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}
	return &StreamConverter{
		logger: logger,
	}
}

// convertStreamChunk converts GLMStreamResponse to PonchoStreamChunk
func (sc *StreamConverter) ConvertStreamChunk(streamResp *GLMStreamResponse) (*interfaces.PonchoStreamChunk, error) {
	if len(streamResp.Choices) == 0 {
		return nil, nil // Skip chunks with no choices
	}

	choice := streamResp.Choices[0]

	// Convert delta message
	var delta *interfaces.PonchoMessage
	if choice.Delta.Role != "" || choice.Delta.Content != "" || len(choice.Delta.ToolCalls) > 0 {
		var err error
		delta, err = sc.convertStreamDelta(&choice.Delta)
		if err != nil {
			return nil, err
		}
	}

	// Convert usage (may be nil in most chunks)
	var usage *interfaces.PonchoUsage
	if streamResp.Usage != nil {
		usage = &interfaces.PonchoUsage{
			PromptTokens:     streamResp.Usage.PromptTokens,
			CompletionTokens: streamResp.Usage.CompletionTokens,
			TotalTokens:      streamResp.Usage.TotalTokens,
		}
	}

	// Convert finish reason
	var finishReason interfaces.PonchoFinishReason
	if choice.FinishReason != nil {
		finishReason = sc.convertFinishReason(*choice.FinishReason)
	}

	// Determine if stream is done
	done := choice.FinishReason != nil && *choice.FinishReason != ""

	// Create stream chunk
	chunk := &interfaces.PonchoStreamChunk{
		Delta:        delta,
		Usage:        usage,
		FinishReason: finishReason,
		Done:         done,
		Metadata:     make(map[string]interface{}),
	}

	// Add metadata
	chunk.Metadata["id"] = streamResp.ID
	chunk.Metadata["object"] = streamResp.Object
	chunk.Metadata["created"] = streamResp.Created
	chunk.Metadata["model"] = streamResp.Model

	return chunk, nil
}

// convertStreamDelta converts GLMStreamDelta to PonchoMessage
func (sc *StreamConverter) convertStreamDelta(delta *GLMStreamDelta) (*interfaces.PonchoMessage, error) {
	ponchoMsg := &interfaces.PonchoMessage{}

	// Set role if provided
	if delta.Role != "" {
		ponchoMsg.Role = interfaces.PonchoRole(delta.Role)
	}

	// Convert content and tool calls
	var contentParts []*interfaces.PonchoContentPart

	// Add text content
	if delta.Content != "" {
		contentParts = append(contentParts, &interfaces.PonchoContentPart{
			Type: interfaces.PonchoContentTypeText,
			Text: delta.Content,
		})
	}

	// Add tool calls
	for _, toolCall := range delta.ToolCalls {
		args, err := sc.jsonStringToMap(toolCall.Function.Arguments)
		if err != nil {
			sc.logger.Warn("Failed to parse stream tool call arguments", "error", err, "arguments", toolCall.Function.Arguments)
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

// convertFinishReason converts GLM finish reason to PonchoFinishReason
func (sc *StreamConverter) convertFinishReason(reason string) interfaces.PonchoFinishReason {
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
func (sc *StreamConverter) mapToJSONString(data map[string]interface{}) string {
	if len(data) == 0 {
		return "{}"
	}

	// Simple JSON marshaling - in production, you might want better error handling
	if bytes, err := json.Marshal(data); err == nil {
		return string(bytes)
	}
	return "{}"
}

func (sc *StreamConverter) jsonStringToMap(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" || jsonStr == "{}" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}
