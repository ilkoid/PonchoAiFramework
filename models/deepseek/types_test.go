package deepseek

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepSeekConstants(t *testing.T) {
	assert.Equal(t, "https://api.deepseek.com", DeepSeekDefaultBaseURL)
	assert.Equal(t, "deepseek-chat", DeepSeekDefaultModel)
	assert.Equal(t, "/chat/completions", DeepSeekEndpoint)

	assert.Equal(t, "stop", DeepSeekFinishReasonStop)
	assert.Equal(t, "length", DeepSeekFinishReasonLength)
	assert.Equal(t, "tool_calls", DeepSeekFinishReasonTool)
	assert.Equal(t, "error", DeepSeekFinishReasonError)

	assert.Equal(t, "none", DeepSeekToolChoiceNone)
	assert.Equal(t, "auto", DeepSeekToolChoiceAuto)

	assert.Equal(t, "text", DeepSeekResponseFormatText)
	assert.Equal(t, "json_object", DeepSeekResponseFormatJSONObject)

	assert.Equal(t, "enabled", DeepSeekThinkingEnabled)
	assert.Equal(t, "disabled", DeepSeekThinkingDisabled)
}

func TestDeepSeekMessage_JSONSerialization(t *testing.T) {
	msg := DeepSeekMessage{
		Role:    "user",
		Content: "Hello, world!",
		Name:    func() *string { s := "test"; return &s }(),
		ToolCalls: []DeepSeekToolCall{
			{
				ID:   "call_123",
				Type: "function",
				Function: DeepSeekFunctionCall{
					Name:      "test_function",
					Arguments: `{"param1": "value1"}`,
				},
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"role":"user"`)
	assert.Contains(t, string(data), `"content":"Hello, world!"`)
	assert.Contains(t, string(data), `"name":"test"`)
	assert.Contains(t, string(data), `"id":"call_123"`)
	assert.Contains(t, string(data), `"name":"test_function"`)

	// Test JSON unmarshaling
	var unmarshaled DeepSeekMessage
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, msg.Role, unmarshaled.Role)
	assert.Equal(t, msg.Content, unmarshaled.Content)
	assert.Equal(t, *msg.Name, *unmarshaled.Name)
	assert.Len(t, unmarshaled.ToolCalls, 1)
	assert.Equal(t, msg.ToolCalls[0].ID, unmarshaled.ToolCalls[0].ID)
	assert.Equal(t, msg.ToolCalls[0].Type, unmarshaled.ToolCalls[0].Type)
	assert.Equal(t, msg.ToolCalls[0].Function.Name, unmarshaled.ToolCalls[0].Function.Name)
	assert.Equal(t, msg.ToolCalls[0].Function.Arguments, unmarshaled.ToolCalls[0].Function.Arguments)
}

func TestDeepSeekTool_JSONSerialization(t *testing.T) {
	tool := DeepSeekTool{
		Type: "function",
		Function: DeepSeekToolFunction{
			Name:        "test_tool",
			Description: "A test tool",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param1": map[string]interface{}{
						"type":        "string",
						"description": "First parameter",
					},
				},
				"required": []string{"param1"},
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(tool)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"type":"function"`)
	assert.Contains(t, string(data), `"name":"test_tool"`)
	assert.Contains(t, string(data), `"description":"A test tool"`)

	// Test JSON unmarshaling
	var unmarshaled DeepSeekTool
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, tool.Type, unmarshaled.Type)
	assert.Equal(t, tool.Function.Name, unmarshaled.Function.Name)
	assert.Equal(t, tool.Function.Description, unmarshaled.Function.Description)
	assert.Equal(t, tool.Function.Parameters["type"], unmarshaled.Function.Parameters["type"])
}

func TestDeepSeekRequest_JSONSerialization(t *testing.T) {
	maxTokens := 1000
	temperature := float32(0.8)
	topP := float32(0.9)
	freqPenalty := float32(0.1)
	presPenalty := float32(0.2)
	logProbs := true
	topLogProbs := 5

	req := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{Role: "system", Content: "You are helpful assistant"},
			{Role: "user", Content: "Hello"},
		},
		Temperature:      &temperature,
		MaxTokens:        &maxTokens,
		TopP:             &topP,
		FrequencyPenalty: &freqPenalty,
		PresencePenalty:  &presPenalty,
		Stop:             []string{"stop1", "stop2"},
		ResponseFormat:   &DeepSeekResponseFormat{Type: "text"},
		Thinking:         &DeepSeekThinking{Type: "disabled"},
		LogProbs:         logProbs,
		TopLogProbs:      &topLogProbs,
		Tools: []DeepSeekTool{
			{
				Type: "function",
				Function: DeepSeekToolFunction{
					Name:        "test_tool",
					Description: "Test tool",
					Parameters:  map[string]interface{}{"type": "object"},
				},
			},
		},
		ToolChoice: DeepSeekToolChoiceAuto,
	}

	// Test JSON marshaling
	data, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"model":"deepseek-chat"`)
	assert.Contains(t, string(data), `"temperature":0.8`)
	assert.Contains(t, string(data), `"max_tokens":1000`)
	assert.Contains(t, string(data), `"top_p":0.9`)
	assert.Contains(t, string(data), `"frequency_penalty":0.1`)
	assert.Contains(t, string(data), `"presence_penalty":0.2`)
	assert.Contains(t, string(data), `"logprobs":true`)
	assert.Contains(t, string(data), `"top_logprobs":5`)
	assert.Contains(t, string(data), `"tool_choice":"auto"`)

	// Test JSON unmarshaling
	var unmarshaled DeepSeekRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, req.Model, unmarshaled.Model)
	assert.Len(t, unmarshaled.Messages, 2)
	assert.Equal(t, *req.Temperature, *unmarshaled.Temperature)
	assert.Equal(t, *req.MaxTokens, *unmarshaled.MaxTokens)
	assert.Equal(t, *req.TopP, *unmarshaled.TopP)
	assert.Equal(t, *req.FrequencyPenalty, *unmarshaled.FrequencyPenalty)
	assert.Equal(t, *req.PresencePenalty, *unmarshaled.PresencePenalty)
	assert.Equal(t, req.LogProbs, unmarshaled.LogProbs)
	assert.Equal(t, *req.TopLogProbs, *unmarshaled.TopLogProbs)
	assert.Equal(t, req.ToolChoice, unmarshaled.ToolChoice)
}

func TestDeepSeekResponse_JSONSerialization(t *testing.T) {
	resp := DeepSeekResponse{
		ID:                "resp_123",
		Object:            "chat.completion",
		Created:           1234567890,
		Model:             "deepseek-chat",
		SystemFingerprint: "fp_456",
		Choices: []DeepSeekChoice{
			{
				Index: 0,
				Message: DeepSeekMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
					ToolCalls: []DeepSeekToolCall{
						{
							ID:   "call_789",
							Type: "function",
							Function: DeepSeekFunctionCall{
								Name:      "test_function",
								Arguments: `{"result": "success"}`,
							},
						},
					},
				},
				FinishReason: "tool_calls",
			},
		},
		Usage: DeepSeekUsage{
			PromptTokens:          20,
			CompletionTokens:      15,
			TotalTokens:           35,
			PromptCacheHitTokens:  5,
			PromptCacheMissTokens: 15,
			CompletionTokensDetails: &DeepSeekCompletionDetails{
				ReasoningTokens: 8,
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"id":"resp_123"`)
	assert.Contains(t, string(data), `"object":"chat.completion"`)
	assert.Contains(t, string(data), `"model":"deepseek-chat"`)
	assert.Contains(t, string(data), `"finish_reason":"tool_calls"`)
	assert.Contains(t, string(data), `"prompt_tokens":20`)
	assert.Contains(t, string(data), `"completion_tokens":15`)
	assert.Contains(t, string(data), `"total_tokens":35`)

	// Test JSON unmarshaling
	var unmarshaled DeepSeekResponse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, resp.ID, unmarshaled.ID)
	assert.Equal(t, resp.Object, unmarshaled.Object)
	assert.Equal(t, resp.Created, unmarshaled.Created)
	assert.Equal(t, resp.Model, unmarshaled.Model)
	assert.Equal(t, resp.SystemFingerprint, unmarshaled.SystemFingerprint)
	assert.Len(t, unmarshaled.Choices, 1)
	assert.Equal(t, resp.Choices[0].Index, unmarshaled.Choices[0].Index)
	assert.Equal(t, resp.Choices[0].FinishReason, unmarshaled.Choices[0].FinishReason)
	assert.Equal(t, resp.Usage.PromptTokens, unmarshaled.Usage.PromptTokens)
	assert.Equal(t, resp.Usage.CompletionTokens, unmarshaled.Usage.CompletionTokens)
	assert.Equal(t, resp.Usage.TotalTokens, unmarshaled.Usage.TotalTokens)
	assert.Equal(t, resp.Usage.PromptCacheHitTokens, unmarshaled.Usage.PromptCacheHitTokens)
	assert.Equal(t, resp.Usage.PromptCacheMissTokens, unmarshaled.Usage.PromptCacheMissTokens)
	assert.NotNil(t, unmarshaled.Usage.CompletionTokensDetails)
	assert.Equal(t, resp.Usage.CompletionTokensDetails.ReasoningTokens, unmarshaled.Usage.CompletionTokensDetails.ReasoningTokens)
}

func TestDeepSeekStreamResponse_JSONSerialization(t *testing.T) {
	resp := DeepSeekStreamResponse{
		ID:                "stream_123",
		Object:            "chat.completion.chunk",
		Created:           1234567890,
		Model:             "deepseek-chat",
		SystemFingerprint: "fp_stream",
		Choices: []DeepSeekStreamChoice{
			{
				Index: 0,
				Delta: DeepSeekStreamDelta{
					Role:             "assistant",
					Content:          "Hello",
					ReasoningContent: "thinking...",
					ToolCalls: []DeepSeekToolCall{
						{
							ID:   "call_stream",
							Type: "function",
							Function: DeepSeekFunctionCall{
								Name:      "stream_tool",
								Arguments: `{"data": "streaming"}`,
							},
						},
					},
				},
				FinishReason: func() *string { s := "stop"; return &s }(),
			},
		},
		Usage: &DeepSeekUsage{
			PromptTokens:     25,
			CompletionTokens: 10,
			TotalTokens:      35,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"id":"stream_123"`)
	assert.Contains(t, string(data), `"object":"chat.completion.chunk"`)
	assert.Contains(t, string(data), `"role":"assistant"`)
	assert.Contains(t, string(data), `"content":"Hello"`)
	assert.Contains(t, string(data), `"reasoning_content":"thinking..."`)
	assert.Contains(t, string(data), `"finish_reason":"stop"`)

	// Test JSON unmarshaling
	var unmarshaled DeepSeekStreamResponse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, resp.ID, unmarshaled.ID)
	assert.Equal(t, resp.Object, unmarshaled.Object)
	assert.Equal(t, resp.Model, unmarshaled.Model)
	assert.Len(t, unmarshaled.Choices, 1)
	assert.Equal(t, resp.Choices[0].Delta.Role, unmarshaled.Choices[0].Delta.Role)
	assert.Equal(t, resp.Choices[0].Delta.Content, unmarshaled.Choices[0].Delta.Content)
	assert.Equal(t, resp.Choices[0].Delta.ReasoningContent, unmarshaled.Choices[0].Delta.ReasoningContent)
	assert.NotNil(t, unmarshaled.Choices[0].FinishReason)
	assert.Equal(t, *resp.Choices[0].FinishReason, *unmarshaled.Choices[0].FinishReason)
	assert.NotNil(t, unmarshaled.Usage)
	assert.Equal(t, resp.Usage.PromptTokens, unmarshaled.Usage.PromptTokens)
}

func TestDeepSeekError_JSONSerialization(t *testing.T) {
	errResp := DeepSeekError{
		Error: DeepSeekErrorDetail{
			Message: "Invalid API key",
			Type:    "authentication_error",
			Code:    "invalid_api_key",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(errResp)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"message":"Invalid API key"`)
	assert.Contains(t, string(data), `"type":"authentication_error"`)
	assert.Contains(t, string(data), `"code":"invalid_api_key"`)

	// Test JSON unmarshaling
	var unmarshaled DeepSeekError
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, errResp.Error.Message, unmarshaled.Error.Message)
	assert.Equal(t, errResp.Error.Type, unmarshaled.Error.Type)
	assert.Equal(t, errResp.Error.Code, unmarshaled.Error.Code)
}
