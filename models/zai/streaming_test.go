package zai

import (
	"os"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestProcessStreamChunk(t *testing.T) {
	// Valid JSON chunk
	jsonChunk := `{"id":"test-123","object":"chat.completion.chunk","created":1234567890,"model":"glm-4.6","choices":[{"index":0,"delta":{"content":"Hello "},"finish_reason":null}]}`

	chunk, err := ProcessStreamChunk([]byte(jsonChunk))
	assert.NoError(t, err)
	assert.NotNil(t, chunk)
	assert.Equal(t, "test-123", chunk.ID)
	assert.Equal(t, "chat.completion.chunk", chunk.Object)
	assert.Equal(t, int64(1234567890), chunk.Created)
	assert.Equal(t, "glm-4.6", chunk.Model)
	assert.Len(t, chunk.Choices, 1)
	assert.Equal(t, "Hello ", chunk.Choices[0].Delta.Content)
	assert.Nil(t, chunk.Choices[0].FinishReason)
}

func TestProcessStreamChunk_InvalidJSON(t *testing.T) {
	invalidJSON := `{"invalid": json}`

	chunk, err := ProcessStreamChunk([]byte(invalidJSON))
	assert.Error(t, err)
	assert.Nil(t, chunk)
	assert.Contains(t, err.Error(), "failed to parse stream chunk")
}

func TestConvertStreamChunkToPoncho(t *testing.T) {
	// Create a sample Z.AI stream response
	zaiChunk := &ZAIStreamResponse{
		ID:      "test-456",
		Object:  "chat.completion.chunk",
		Created: 1234567890,
		Model:   "glm-4.6",
		Choices: []ZAIStreamChoice{
			{
				Index: 0,
				Delta: ZAIStreamDelta{
					Content: "Hello, world!",
				},
				FinishReason: stringPtr("stop"),
			},
		},
		Usage: &ZAIUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	ponchoChunk, err := ConvertStreamChunkToPoncho(zaiChunk)
	assert.NoError(t, err)
	assert.NotNil(t, ponchoChunk)

	// Check metadata
	assert.Equal(t, "test-456", ponchoChunk.Metadata["id"])
	assert.Equal(t, "chat.completion.chunk", ponchoChunk.Metadata["object"])
	assert.Equal(t, int64(1234567890), ponchoChunk.Metadata["created"])
	assert.Equal(t, "glm-4.6", ponchoChunk.Metadata["model"])

	// Check delta
	assert.NotNil(t, ponchoChunk.Delta)
	assert.Equal(t, interfaces.PonchoRoleAssistant, ponchoChunk.Delta.Role)
	assert.Len(t, ponchoChunk.Delta.Content, 1)
	assert.Equal(t, interfaces.PonchoContentTypeText, ponchoChunk.Delta.Content[0].Type)
	assert.Equal(t, "Hello, world!", ponchoChunk.Delta.Content[0].Text)

	// Check finish reason
	assert.True(t, ponchoChunk.Done)
	assert.Equal(t, interfaces.PonchoFinishReasonStop, ponchoChunk.FinishReason)

	// Check usage
	assert.NotNil(t, ponchoChunk.Usage)
	assert.Equal(t, 10, ponchoChunk.Usage.PromptTokens)
	assert.Equal(t, 5, ponchoChunk.Usage.CompletionTokens)
	assert.Equal(t, 15, ponchoChunk.Usage.TotalTokens)
}

func TestConvertStreamChunkToPoncho_ToolCall(t *testing.T) {
	// Create a Z.AI stream response with tool call
	zaiChunk := &ZAIStreamResponse{
		ID:      "test-789",
		Object:  "chat.completion.chunk",
		Created: 1234567890,
		Model:   "glm-4.6",
		Choices: []ZAIStreamChoice{
			{
				Index: 0,
				Delta: ZAIStreamDelta{
					ToolCalls: []ZAIToolCall{
						{
							ID:   "call-123",
							Type: "function",
							Function: ZAIFunctionCall{
								Name:      "get_weather",
								Arguments: `{"location":"New York"}`,
							},
						},
					},
				},
			},
		},
	}

	ponchoChunk, err := ConvertStreamChunkToPoncho(zaiChunk)
	assert.NoError(t, err)
	assert.NotNil(t, ponchoChunk)

	// Check tool call in delta
	assert.Len(t, ponchoChunk.Delta.Content, 1)
	assert.Equal(t, interfaces.PonchoContentTypeTool, ponchoChunk.Delta.Content[0].Type)
	assert.NotNil(t, ponchoChunk.Delta.Content[0].Tool)
	assert.Equal(t, "call-123", ponchoChunk.Delta.Content[0].Tool.ID)
	assert.Equal(t, "get_weather", ponchoChunk.Delta.Content[0].Tool.Name)
	assert.NotNil(t, ponchoChunk.Delta.Content[0].Tool.Args)
	assert.Equal(t, "New York", ponchoChunk.Delta.Content[0].Tool.Args["location"])
}

func TestIsStreamComplete(t *testing.T) {
	tests := []struct {
		name     string
		chunk    *ZAIStreamResponse
		expected bool
	}{
		{
			name:     "complete chunk",
			chunk:    &ZAIStreamResponse{Choices: []ZAIStreamChoice{{FinishReason: stringPtr("stop")}}},
			expected: true,
		},
		{
			name:     "complete chunk with length",
			chunk:    &ZAIStreamResponse{Choices: []ZAIStreamChoice{{FinishReason: stringPtr("length")}}},
			expected: true,
		},
		{
			name:     "incomplete chunk",
			chunk:    &ZAIStreamResponse{Choices: []ZAIStreamChoice{{FinishReason: stringPtr("")}}},
			expected: false,
		},
		{
			name:     "nil finish reason",
			chunk:    &ZAIStreamResponse{Choices: []ZAIStreamChoice{{FinishReason: nil}}},
			expected: false,
		},
		{
			name:     "no choices",
			chunk:    &ZAIStreamResponse{Choices: []ZAIStreamChoice{}},
			expected: false,
		},
		{
			name:     "nil chunk",
			chunk:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsStreamComplete(tt.chunk)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStreamChunkID(t *testing.T) {
	chunk := &ZAIStreamResponse{ID: "test-123"}
	assert.Equal(t, "test-123", GetStreamChunkID(chunk))

	// Test nil chunk
	assert.Equal(t, "", GetStreamChunkID(nil))
}

func TestGetStreamChunkModel(t *testing.T) {
	chunk := &ZAIStreamResponse{Model: "glm-4.6v"}
	assert.Equal(t, "glm-4.6v", GetStreamChunkModel(chunk))

	// Test nil chunk
	assert.Equal(t, "", GetStreamChunkModel(nil))
}

func TestGetStreamChunkCreated(t *testing.T) {
	chunk := &ZAIStreamResponse{Created: 1234567890}
	assert.Equal(t, int64(1234567890), GetStreamChunkCreated(chunk))

	// Test nil chunk
	assert.Equal(t, int64(0), GetStreamChunkCreated(nil))
}

func TestProcessSSEStream_Integration(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY environment variable not set, skipping integration test")
	}

	// This test would require a real API call to test SSE processing
	// For now, we'll test the parsing logic with mock data
	t.Skip("SSE integration test requires real API setup")
}

// Helper function

func stringPtr(s string) *string {
	return &s
}
