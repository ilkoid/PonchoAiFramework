package deepseek

import "time"

// DeepSeek-specific types for API integration

// DeepSeekMessage represents a message in DeepSeek API format
type DeepSeekMessage struct {
	Role      string             `json:"role"`
	Content   string             `json:"content"`
	Name      *string            `json:"name,omitempty"`
	ToolCalls []DeepSeekToolCall `json:"tool_calls,omitempty"`
}

// DeepSeekToolCall represents a tool call in DeepSeek API format
type DeepSeekToolCall struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"` // always "function"
	Function DeepSeekFunctionCall `json:"function"`
}

// DeepSeekFunctionCall represents a function call in DeepSeek API format
type DeepSeekFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// DeepSeekTool represents a tool definition in DeepSeek API format
type DeepSeekTool struct {
	Type     string               `json:"type"` // always "function"
	Function DeepSeekToolFunction `json:"function"`
}

// DeepSeekToolFunction represents a tool function definition
type DeepSeekToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// DeepSeekRequest represents a request to DeepSeek API
type DeepSeekRequest struct {
	Model            string                  `json:"model"`
	Messages         []DeepSeekMessage       `json:"messages"`
	Temperature      *float32                `json:"temperature,omitempty"`
	MaxTokens        *int                    `json:"max_tokens,omitempty"`
	Stream           bool                    `json:"stream,omitempty"`
	Tools            []DeepSeekTool          `json:"tools,omitempty"`
	ToolChoice       interface{}             `json:"tool_choice,omitempty"` // "none", "auto", or specific tool
	TopP             *float32                `json:"top_p,omitempty"`
	FrequencyPenalty *float32                `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32                `json:"presence_penalty,omitempty"`
	Stop             interface{}             `json:"stop,omitempty"` // string or []string
	ResponseFormat   *DeepSeekResponseFormat `json:"response_format,omitempty"`
	Thinking         *DeepSeekThinking       `json:"thinking,omitempty"`
	LogProbs         bool                    `json:"logprobs,omitempty"`
	TopLogProbs      *int                    `json:"top_logprobs,omitempty"`
}

// DeepSeekResponseFormat represents response format configuration
type DeepSeekResponseFormat struct {
	Type string `json:"type"` // "text" or "json_object"
}

// DeepSeekThinking represents thinking mode configuration
type DeepSeekThinking struct {
	Type string `json:"type"` // "enabled" or "disabled"
}

// DeepSeekChoice represents a choice in DeepSeek API response
type DeepSeekChoice struct {
	Index        int               `json:"index"`
	Message      DeepSeekMessage   `json:"message"`
	FinishReason string            `json:"finish_reason"`
	LogProbs     *DeepSeekLogProbs `json:"logprobs,omitempty"`
}

// DeepSeekLogProbs represents log probabilities in response
type DeepSeekLogProbs struct {
	Content          []DeepSeekTokenLogProb `json:"content,omitempty"`
	ReasoningContent []DeepSeekTokenLogProb `json:"reasoning_content,omitempty"`
}

// DeepSeekTokenLogProb represents token log probability information
type DeepSeekTokenLogProb struct {
	Token       string               `json:"token"`
	LogProb     float64              `json:"logprob"`
	Bytes       []int                `json:"bytes,omitempty"`
	TopLogProbs []DeepSeekTopLogProb `json:"top_logprobs,omitempty"`
}

// DeepSeekTopLogProb represents top log probability information
type DeepSeekTopLogProb struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []int   `json:"bytes,omitempty"`
}

// DeepSeekUsage represents token usage information
type DeepSeekUsage struct {
	PromptTokens            int                        `json:"prompt_tokens"`
	CompletionTokens        int                        `json:"completion_tokens"`
	TotalTokens             int                        `json:"total_tokens"`
	PromptCacheHitTokens    int                        `json:"prompt_cache_hit_tokens,omitempty"`
	PromptCacheMissTokens   int                        `json:"prompt_cache_miss_tokens,omitempty"`
	CompletionTokensDetails *DeepSeekCompletionDetails `json:"completion_tokens_details,omitempty"`
}

// DeepSeekCompletionDetails represents completion token details
type DeepSeekCompletionDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// DeepSeekResponse represents a response from DeepSeek API
type DeepSeekResponse struct {
	ID                string           `json:"id"`
	Object            string           `json:"object"` // "chat.completion"
	Created           int64            `json:"created"`
	Model             string           `json:"model"`
	SystemFingerprint string           `json:"system_fingerprint,omitempty"`
	Choices           []DeepSeekChoice `json:"choices"`
	Usage             DeepSeekUsage    `json:"usage"`
}

// DeepSeekStreamChoice represents a choice in streaming response
type DeepSeekStreamChoice struct {
	Index        int                 `json:"index"`
	Delta        DeepSeekStreamDelta `json:"delta"`
	FinishReason *string             `json:"finish_reason,omitempty"`
	LogProbs     *DeepSeekLogProbs   `json:"logprobs,omitempty"`
}

// DeepSeekStreamDelta represents a delta in streaming response
type DeepSeekStreamDelta struct {
	Role             string             `json:"role,omitempty"`
	Content          string             `json:"content,omitempty"`
	ToolCalls        []DeepSeekToolCall `json:"tool_calls,omitempty"`
	ReasoningContent string             `json:"reasoning_content,omitempty"`
}

// DeepSeekStreamResponse represents a streaming response from DeepSeek API
type DeepSeekStreamResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"` // "chat.completion.chunk"
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
	Choices           []DeepSeekStreamChoice `json:"choices"`
	Usage             *DeepSeekUsage         `json:"usage,omitempty"`
}

// DeepSeekError represents an error response from DeepSeek API
type DeepSeekError struct {
	Error DeepSeekErrorDetail `json:"error"`
}

// DeepSeekErrorDetail represents error details
type DeepSeekErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// DeepSeekConfig represents DeepSeek-specific configuration
type DeepSeekConfig struct {
	BaseURL          string                  `json:"base_url"`
	APIKey           string                  `json:"api_key"`
	Model            string                  `json:"model"`
	MaxTokens        int                     `json:"max_tokens"`
	Temperature      float32                 `json:"temperature"`
	Timeout          time.Duration           `json:"timeout"`
	TopP             *float32                `json:"top_p,omitempty"`
	FrequencyPenalty *float32                `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32                `json:"presence_penalty,omitempty"`
	Stop             interface{}             `json:"stop,omitempty"`
	ResponseFormat   *DeepSeekResponseFormat `json:"response_format,omitempty"`
	Thinking         *DeepSeekThinking       `json:"thinking,omitempty"`
	LogProbs         bool                    `json:"logprobs,omitempty"`
	TopLogProbs      *int                    `json:"top_logprobs,omitempty"`
}

// Constants for DeepSeek API
const (
	DeepSeekDefaultBaseURL = "https://api.deepseek.com"
	DeepSeekDefaultModel   = "deepseek-chat"
	DeepSeekEndpoint       = "/chat/completions"

	// Finish reasons
	DeepSeekFinishReasonStop   = "stop"
	DeepSeekFinishReasonLength = "length"
	DeepSeekFinishReasonTool   = "tool_calls"
	DeepSeekFinishReasonError  = "error"

	// Tool choices
	DeepSeekToolChoiceNone = "none"
	DeepSeekToolChoiceAuto = "auto"

	// Response formats
	DeepSeekResponseFormatText       = "text"
	DeepSeekResponseFormatJSONObject = "json_object"

	// Thinking types
	DeepSeekThinkingEnabled  = "enabled"
	DeepSeekThinkingDisabled = "disabled"
)
