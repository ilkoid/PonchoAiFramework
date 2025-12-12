package deepseek

import (
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// convertStreamChunk converts DeepSeekStreamResponse to PonchoStreamChunk
func (m *DeepSeekModel) convertStreamChunk(streamResp *DeepSeekStreamResponse) (*interfaces.PonchoStreamChunk, error) {
	if len(streamResp.Choices) == 0 {
		return nil, nil // Skip chunks with no choices
	}

	choice := streamResp.Choices[0]

	// Convert delta message
	var delta *interfaces.PonchoMessage
	if choice.Delta.Role != "" || choice.Delta.Content != "" || len(choice.Delta.ToolCalls) > 0 {
		var err error
		delta, err = m.convertStreamDelta(&choice.Delta)
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
		finishReason = m.convertFinishReason(*choice.FinishReason)
	}

	// Determine if stream is done
	done := choice.FinishReason != nil && *choice.FinishReason != ""

	// Create stream chunk
	chunk := m.PrepareStreamChunk(delta, usage, finishReason, done)

	// Add metadata
	chunk.Metadata["id"] = streamResp.ID
	chunk.Metadata["object"] = streamResp.Object
	chunk.Metadata["created"] = streamResp.Created
	chunk.Metadata["model"] = streamResp.Model
	chunk.Metadata["system_fingerprint"] = streamResp.SystemFingerprint

	return chunk, nil
}

// convertStreamDelta converts DeepSeekStreamDelta to PonchoMessage
func (m *DeepSeekModel) convertStreamDelta(delta *DeepSeekStreamDelta) (*interfaces.PonchoMessage, error) {
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
		args, err := m.jsonStringToMap(toolCall.Function.Arguments)
		if err != nil {
			m.GetLogger().Warn("Failed to parse stream tool call arguments", "error", err, "arguments", toolCall.Function.Arguments)
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

	// Add reasoning content if present
	if delta.ReasoningContent != "" {
		contentParts = append(contentParts, &interfaces.PonchoContentPart{
			Type: interfaces.PonchoContentTypeText,
			Text: delta.ReasoningContent,
		})
	}

	ponchoMsg.Content = contentParts
	return ponchoMsg, nil
}
