package deepseek

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ProcessStreamChunk processes a streaming chunk from DeepSeek API
func ProcessStreamChunk(chunk []byte) (*DeepSeekStreamResponse, error) {
	var streamResp DeepSeekStreamResponse
	if err := json.Unmarshal(chunk, &streamResp); err != nil {
		return nil, fmt.Errorf("failed to parse stream chunk: %w", err)
	}
	return &streamResp, nil
}

// ConvertStreamChunkToPoncho converts DeepSeek stream chunk to Poncho format
func ConvertStreamChunkToPoncho(chunk *DeepSeekStreamResponse) (*interfaces.PonchoStreamChunk, error) {
	if chunk == nil {
		return nil, fmt.Errorf("chunk is nil")
	}

	ponchoChunk := &interfaces.PonchoStreamChunk{
		Done:     false,
		Metadata: make(map[string]interface{}),
	}

	// Add metadata
	ponchoChunk.Metadata["id"] = chunk.ID
	ponchoChunk.Metadata["object"] = chunk.Object
	ponchoChunk.Metadata["created"] = chunk.Created
	ponchoChunk.Metadata["model"] = chunk.Model
	if chunk.SystemFingerprint != "" {
		ponchoChunk.Metadata["system_fingerprint"] = chunk.SystemFingerprint
	}

	// Convert choices
	if len(chunk.Choices) > 0 {
		choice := chunk.Choices[0]

		ponchoChunk.Delta = &interfaces.PonchoMessage{
			Role:    interfaces.PonchoRoleAssistant,
			Content: make([]*interfaces.PonchoContentPart, 0),
		}

		// Convert content delta
		if choice.Delta.Content != "" {
			ponchoChunk.Delta.Content = append(ponchoChunk.Delta.Content, &interfaces.PonchoContentPart{
				Type: interfaces.PonchoContentTypeText,
				Text: choice.Delta.Content,
			})
		}

		// Convert tool call deltas
		if len(choice.Delta.ToolCalls) > 0 {
			for _, toolCall := range choice.Delta.ToolCalls {
				args := make(map[string]interface{})
				if toolCall.Function.Arguments != "" {
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
						// If parsing fails, keep as raw string
						args = map[string]interface{}{
							"raw_arguments": toolCall.Function.Arguments,
						}
					}
				}

				ponchoChunk.Delta.Content = append(ponchoChunk.Delta.Content, &interfaces.PonchoContentPart{
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
			ponchoChunk.Done = *choice.FinishReason != ""
			switch *choice.FinishReason {
			case DeepSeekFinishReasonStop:
				ponchoChunk.FinishReason = interfaces.PonchoFinishReasonStop
			case DeepSeekFinishReasonLength:
				ponchoChunk.FinishReason = interfaces.PonchoFinishReasonLength
			case DeepSeekFinishReasonTool:
				ponchoChunk.FinishReason = interfaces.PonchoFinishReasonTool
			case DeepSeekFinishReasonError:
				ponchoChunk.FinishReason = interfaces.PonchoFinishReasonError
			default:
				ponchoChunk.FinishReason = interfaces.PonchoFinishReasonStop
			}
		}
	}

	// Convert usage
	if chunk.Usage != nil {
		ponchoChunk.Usage = &interfaces.PonchoUsage{
			PromptTokens:     chunk.Usage.PromptTokens,
			CompletionTokens: chunk.Usage.CompletionTokens,
			TotalTokens:      chunk.Usage.TotalTokens,
		}
	}

	return ponchoChunk, nil
}

// IsStreamComplete checks if the stream is complete
func IsStreamComplete(chunk *DeepSeekStreamResponse) bool {
	if chunk == nil || len(chunk.Choices) == 0 {
		return false
	}

	return chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != ""
}

// GetStreamChunkID returns the ID of a stream chunk
func GetStreamChunkID(chunk *DeepSeekStreamResponse) string {
	if chunk == nil {
		return ""
	}
	return chunk.ID
}

// GetStreamChunkModel returns the model of a stream chunk
func GetStreamChunkModel(chunk *DeepSeekStreamResponse) string {
	if chunk == nil {
		return ""
	}
	return chunk.Model
}

// GetStreamChunkCreated returns the creation time of a stream chunk
func GetStreamChunkCreated(chunk *DeepSeekStreamResponse) int64 {
	if chunk == nil {
		return 0
	}
	return chunk.Created
}

// ProcessSSEStream processes Server-Sent Events stream
func ProcessSSEStream(ctx context.Context, body io.ReadCloser, callback func(*DeepSeekStreamResponse) error) error {
	defer body.Close()

	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Remove "data: " prefix
		if strings.HasPrefix(line, "data: ") {
			line = line[6:]
		} else if strings.HasPrefix(line, "data:") {
			line = line[5:]
		}

		// Check for [DONE] marker
		if line == "[DONE]" {
			return nil
		}

		// Skip empty data lines
		if line == "" {
			continue
		}

		// Parse JSON
		var streamResp DeepSeekStreamResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			// Log error but continue processing
			continue
		}

		// Call callback
		if err := callback(&streamResp); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream processing error: %w", err)
	}

	return nil
}
