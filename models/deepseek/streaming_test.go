package deepseek

import (
	"bytes"
	"context"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessStreamChunk(t *testing.T) {
	// Test valid stream chunk
	chunkJSON := `{
		"id": "chatcmpl-123",
		"object": "chat.completion.chunk",
		"created": 1677652288,
		"model": "deepseek-chat",
		"choices": [{
			"index": 0,
			"delta": {
				"role": "assistant",
				"content": "Hello"
			},
			"finish_reason": null
		}]
	}`

	chunk, err := ProcessStreamChunk([]byte(chunkJSON))
	require.NoError(t, err)
	assert.Equal(t, "chatcmpl-123", chunk.ID)
	assert.Equal(t, "chat.completion.chunk", chunk.Object)
	assert.Equal(t, int64(1677652288), chunk.Created)
	assert.Equal(t, "deepseek-chat", chunk.Model)
	assert.Len(t, chunk.Choices, 1)
	assert.Equal(t, "assistant", chunk.Choices[0].Delta.Role)
	assert.Equal(t, "Hello", chunk.Choices[0].Delta.Content)
	assert.Nil(t, chunk.Choices[0].FinishReason)
}

func TestConvertStreamChunkToPoncho(t *testing.T) {
	// Test conversion of stream chunk to Poncho format
	streamChunk := &DeepSeekStreamResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion.chunk",
		Created: 1677652288,
		Model:   "deepseek-chat",
		Choices: []DeepSeekStreamChoice{
			{
				Index: 0,
				Delta: DeepSeekStreamDelta{
					Role:    "assistant",
					Content: "Hello, world!",
				},
			},
		},
	}

	ponchoChunk, err := ConvertStreamChunkToPoncho(streamChunk)
	require.NoError(t, err)
	assert.NotNil(t, ponchoChunk)
	assert.False(t, ponchoChunk.Done)
	assert.Equal(t, "chatcmpl-123", ponchoChunk.Metadata["id"])
	assert.Equal(t, "chat.completion.chunk", ponchoChunk.Metadata["object"])
	assert.Equal(t, int64(1677652288), ponchoChunk.Metadata["created"])
	assert.Equal(t, "deepseek-chat", ponchoChunk.Metadata["model"])
	assert.NotNil(t, ponchoChunk.Delta)
	assert.Equal(t, interfaces.PonchoRoleAssistant, ponchoChunk.Delta.Role)
	assert.Len(t, ponchoChunk.Delta.Content, 1)
	assert.Equal(t, interfaces.PonchoContentTypeText, ponchoChunk.Delta.Content[0].Type)
	assert.Equal(t, "Hello, world!", ponchoChunk.Delta.Content[0].Text)
}

func TestConvertStreamChunkWithToolCalls(t *testing.T) {
	// Test conversion of stream chunk with tool calls
	toolCallArgs := `{"location": "Boston"}`
	streamChunk := &DeepSeekStreamResponse{
		ID:      "chatcmpl-456",
		Object:  "chat.completion.chunk",
		Created: 1677652289,
		Model:   "deepseek-chat",
		Choices: []DeepSeekStreamChoice{
			{
				Index: 0,
				Delta: DeepSeekStreamDelta{
					ToolCalls: []DeepSeekToolCall{
						{
							ID:   "call_123",
							Type: "function",
							Function: DeepSeekFunctionCall{
								Name:      "get_weather",
								Arguments: toolCallArgs,
							},
						},
					},
				},
			},
		},
	}

	ponchoChunk, err := ConvertStreamChunkToPoncho(streamChunk)
	require.NoError(t, err)
	assert.NotNil(t, ponchoChunk)
	assert.Len(t, ponchoChunk.Delta.Content, 1)
	assert.Equal(t, interfaces.PonchoContentTypeTool, ponchoChunk.Delta.Content[0].Type)
	assert.NotNil(t, ponchoChunk.Delta.Content[0].Tool)
	assert.Equal(t, "call_123", ponchoChunk.Delta.Content[0].Tool.ID)
	assert.Equal(t, "get_weather", ponchoChunk.Delta.Content[0].Tool.Name)

	// Parse arguments to verify they were parsed correctly
	args := ponchoChunk.Delta.Content[0].Tool.Args
	assert.Contains(t, args, "location")
	assert.Equal(t, "Boston", args["location"])
}

func TestConvertStreamChunkWithFinishReason(t *testing.T) {
	// Test conversion of stream chunk with finish reason
	finishReason := "stop"
	streamChunk := &DeepSeekStreamResponse{
		ID:      "chatcmpl-789",
		Object:  "chat.completion.chunk",
		Created: 1677652290,
		Model:   "deepseek-chat",
		Choices: []DeepSeekStreamChoice{
			{
				Index:        0,
				Delta:        DeepSeekStreamDelta{},
				FinishReason: &finishReason,
			},
		},
	}

	ponchoChunk, err := ConvertStreamChunkToPoncho(streamChunk)
	require.NoError(t, err)
	assert.True(t, ponchoChunk.Done)
	assert.Equal(t, interfaces.PonchoFinishReasonStop, ponchoChunk.FinishReason)
}

func TestIsStreamComplete(t *testing.T) {
	tests := []struct {
		name     string
		chunk    *DeepSeekStreamResponse
		expected bool
	}{
		{
			name:     "nil chunk",
			chunk:    nil,
			expected: false,
		},
		{
			name: "empty choices",
			chunk: &DeepSeekStreamResponse{
				Choices: []DeepSeekStreamChoice{},
			},
			expected: false,
		},
		{
			name: "no finish reason",
			chunk: &DeepSeekStreamResponse{
				Choices: []DeepSeekStreamChoice{
					{FinishReason: nil},
				},
			},
			expected: false,
		},
		{
			name: "empty finish reason",
			chunk: &DeepSeekStreamResponse{
				Choices: []DeepSeekStreamChoice{
					{FinishReason: func() *string { s := ""; return &s }()},
				},
			},
			expected: false,
		},
		{
			name: "finish reason set",
			chunk: &DeepSeekStreamResponse{
				Choices: []DeepSeekStreamChoice{
					{FinishReason: func() *string { s := "stop"; return &s }()},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsStreamComplete(tt.chunk)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessSSEStream(t *testing.T) {
	// Test SSE stream processing
	streamData := `data: {"id": "chunk1", "object": "chat.completion.chunk", "created": 1677652288, "model": "deepseek-chat", "choices": [{"index": 0, "delta": {"content": "Hello"}}]}

data: {"id": "chunk2", "object": "chat.completion.chunk", "created": 1677652289, "model": "deepseek-chat", "choices": [{"index": 0, "delta": {"content": " world"}}, {"index": 0, "delta": {"content": "!"}}]}

data: [DONE]

: this is a comment
`

	ctx := context.Background()
	body := bytes.NewReader([]byte(streamData))

	var processedChunks []*DeepSeekStreamResponse
	callback := func(chunk *DeepSeekStreamResponse) error {
		processedChunks = append(processedChunks, chunk)
		return nil
	}

	// Create a ReadCloser from the reader
	readCloser := &testReadCloser{Reader: body}

	err := ProcessSSEStream(ctx, readCloser, callback)
	require.NoError(t, err)
	assert.Len(t, processedChunks, 2)
	assert.Equal(t, "chunk1", processedChunks[0].ID)
	assert.Equal(t, "chunk2", processedChunks[1].ID)
}

func TestProcessSSEStreamWithContextCancellation(t *testing.T) {
	// Test SSE stream processing with context cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	streamData := `data: {"id": "chunk1", "object": "chat.completion.chunk"}`
	body := bytes.NewReader([]byte(streamData))
	readCloser := &testReadCloser{Reader: body}

	callback := func(chunk *DeepSeekStreamResponse) error {
		return nil
	}

	err := ProcessSSEStream(ctx, readCloser, callback)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestProcessSSEStreamWithCallbackError(t *testing.T) {
	// Test SSE stream processing when callback returns error
	ctx := context.Background()
	streamData := `data: {"id": "chunk1", "object": "chat.completion.chunk"}
data: {"id": "chunk2", "object": "chat.completion.chunk"}`

	body := bytes.NewReader([]byte(streamData))
	readCloser := &testReadCloser{Reader: body}

	callCount := 0
	callback := func(chunk *DeepSeekStreamResponse) error {
		callCount++
		if callCount >= 2 {
			return assert.AnError
		}
		return nil
	}

	err := ProcessSSEStream(ctx, readCloser, callback)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, 2, callCount) // Should process both chunks before error
}

// Test helper functions
func TestGetStreamChunkID(t *testing.T) {
	chunk := &DeepSeekStreamResponse{ID: "test-id"}
	assert.Equal(t, "test-id", GetStreamChunkID(chunk))
	assert.Equal(t, "", GetStreamChunkID(nil))
}

func TestGetStreamChunkModel(t *testing.T) {
	chunk := &DeepSeekStreamResponse{Model: "deepseek-chat"}
	assert.Equal(t, "deepseek-chat", GetStreamChunkModel(chunk))
	assert.Equal(t, "", GetStreamChunkModel(nil))
}

func TestGetStreamChunkCreated(t *testing.T) {
	chunk := &DeepSeekStreamResponse{Created: 1677652288}
	assert.Equal(t, int64(1677652288), GetStreamChunkCreated(chunk))
	assert.Equal(t, int64(0), GetStreamChunkCreated(nil))
}

// testReadCloser is a helper for testing ReadCloser interface
type testReadCloser struct {
	*bytes.Reader
}

func (trc *testReadCloser) Close() error {
	return nil
}

// Benchmark tests
func BenchmarkProcessStreamChunk(b *testing.B) {
	chunkJSON := []byte(`{
		"id": "chatcmpl-123",
		"object": "chat.completion.chunk",
		"created": 1677652288,
		"model": "deepseek-chat",
		"choices": [{
			"index": 0,
			"delta": {
				"role": "assistant",
				"content": "Hello"
			}
		}]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ProcessStreamChunk(chunkJSON)
	}
}

func BenchmarkConvertStreamChunkToPoncho(b *testing.B) {
	streamChunk := &DeepSeekStreamResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion.chunk",
		Created: 1677652288,
		Model:   "deepseek-chat",
		Choices: []DeepSeekStreamChoice{
			{
				Index: 0,
				Delta: DeepSeekStreamDelta{
					Role:    "assistant",
					Content: "Hello, world!",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertStreamChunkToPoncho(streamChunk)
	}
}
