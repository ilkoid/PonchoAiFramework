package zai

import "time"

// Z.AI-specific types for API integration

// ZAIMessage represents a message in Z.AI API format
type ZAIMessage struct {
	Role      string        `json:"role"`
	Content   interface{}   `json:"content"` // Can be string or []ZAIContentPart
	Name      *string       `json:"name,omitempty"`
	ToolCalls []ZAIToolCall `json:"tool_calls,omitempty"`
}

// ZAIContentPart represents a content part in Z.AI API format
type ZAIContentPart struct {
	Type     string       `json:"type"` // "text", "image_url", "media"
	Text     string       `json:"text,omitempty"`
	ImageURL *ZAIImageURL `json:"image_url,omitempty"`
	Media    *ZAIMedia    `json:"media,omitempty"`
}

// ZAIImageURL represents an image URL in Z.AI API format
type ZAIImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // "auto", "low", "high"
}

// ZAIMedia represents media content in Z.AI API format
type ZAIMedia struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type,omitempty"`
	Data     string `json:"data,omitempty"` // Base64 encoded data
}

// ZAIToolCall represents a tool call in Z.AI API format
type ZAIToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"` // always "function"
	Function ZAIFunctionCall `json:"function"`
}

// ZAIFunctionCall represents a function call in Z.AI API format
type ZAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ZAITool represents a tool definition in Z.AI API format
type ZAITool struct {
	Type     string          `json:"type"` // always "function"
	Function ZAIToolFunction `json:"function"`
}

// ZAIToolFunction represents a tool function definition
type ZAIToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ZAIRequest represents a request to Z.AI API
type ZAIRequest struct {
	Model            string             `json:"model"`
	Messages         []ZAIMessage       `json:"messages"`
	Temperature      *float32           `json:"temperature,omitempty"`
	MaxTokens        *int               `json:"max_tokens,omitempty"`
	Stream           bool               `json:"stream,omitempty"`
	Tools            []ZAITool          `json:"tools,omitempty"`
	ToolChoice       interface{}        `json:"tool_choice,omitempty"` // "none", "auto", or specific tool
	TopP             *float32           `json:"top_p,omitempty"`
	FrequencyPenalty *float32           `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32           `json:"presence_penalty,omitempty"`
	Stop             interface{}        `json:"stop,omitempty"` // string or []string
	ResponseFormat   *ZAIResponseFormat `json:"response_format,omitempty"`
	LogProbs         bool               `json:"logprobs,omitempty"`
	TopLogProbs      *int               `json:"top_logprobs,omitempty"`
}

// ZAIResponseFormat represents response format configuration
type ZAIResponseFormat struct {
	Type string `json:"type"` // "text" or "json_object"
}

// ZAIChoice represents a choice in Z.AI API response
type ZAIChoice struct {
	Index        int          `json:"index"`
	Message      ZAIMessage   `json:"message"`
	FinishReason string       `json:"finish_reason"`
	LogProbs     *ZAILogProbs `json:"logprobs,omitempty"`
}

// ZAILogProbs represents log probabilities in response
type ZAILogProbs struct {
	Content []ZAITokenLogProb `json:"content,omitempty"`
}

// ZAITokenLogProb represents token log probability information
type ZAITokenLogProb struct {
	Token       string          `json:"token"`
	LogProb     float64         `json:"logprob"`
	Bytes       []int           `json:"bytes,omitempty"`
	TopLogProbs []ZAITopLogProb `json:"top_logprobs,omitempty"`
}

// ZAITopLogProb represents top log probability information
type ZAITopLogProb struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []int   `json:"bytes,omitempty"`
}

// ZAIUsage represents token usage information
type ZAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ZAIResponse represents a response from Z.AI API
type ZAIResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"` // "chat.completion"
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []ZAIChoice `json:"choices"`
	Usage   ZAIUsage    `json:"usage"`
}

// ZAIStreamChoice represents a choice in streaming response
type ZAIStreamChoice struct {
	Index        int            `json:"index"`
	Delta        ZAIStreamDelta `json:"delta"`
	FinishReason *string        `json:"finish_reason,omitempty"`
	LogProbs     *ZAILogProbs   `json:"logprobs,omitempty"`
}

// ZAIStreamDelta represents a delta in streaming response
type ZAIStreamDelta struct {
	Role      string        `json:"role,omitempty"`
	Content   string        `json:"content,omitempty"`
	ToolCalls []ZAIToolCall `json:"tool_calls,omitempty"`
}

// ZAIStreamResponse represents a streaming response from Z.AI API
type ZAIStreamResponse struct {
	ID      string            `json:"id"`
	Object  string            `json:"object"` // "chat.completion.chunk"
	Created int64             `json:"created"`
	Model   string            `json:"model"`
	Choices []ZAIStreamChoice `json:"choices"`
	Usage   *ZAIUsage         `json:"usage,omitempty"`
}

// ZAIError represents an error response from Z.AI API
type ZAIError struct {
	Error ZAIErrorDetail `json:"error"`
}

// ZAIErrorDetail represents error details
type ZAIErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// ZAIConfig represents Z.AI-specific configuration
type ZAIConfig struct {
	BaseURL          string             `json:"base_url"`
	APIKey           string             `json:"api_key"`
	Model            string             `json:"model"`
	MaxTokens        int                `json:"max_tokens"`
	Temperature      float32            `json:"temperature"`
	Timeout          time.Duration      `json:"timeout"`
	TopP             *float32           `json:"top_p,omitempty"`
	FrequencyPenalty *float32           `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32           `json:"presence_penalty,omitempty"`
	Stop             interface{}        `json:"stop,omitempty"`
	ResponseFormat   *ZAIResponseFormat `json:"response_format,omitempty"`
	LogProbs         bool               `json:"logprobs,omitempty"`
	TopLogProbs      *int               `json:"top_logprobs,omitempty"`
	VisionConfig     *ZAIVisionConfig   `json:"vision_config,omitempty"`
}

// ZAIVisionConfig represents vision-specific configuration
type ZAIVisionConfig struct {
	MaxImageSize   int      `json:"max_image_size"`  // Maximum image size in pixels
	SupportedTypes []string `json:"supported_types"` // Supported image formats
	Quality        string   `json:"quality"`         // "auto", "low", "high"
	Detail         string   `json:"detail"`          // "auto", "low", "high"
}

// Constants for Z.AI API
const (
	ZAIDefaultBaseURL = "https://api.z.ai/api/paas/v4"
	ZAIDefaultModel   = "glm-4.6"
	ZAIVisionModel    = "glm-4.6v"
	ZAIEndpoint       = "/chat/completions"

	// Finish reasons
	ZAIFinishReasonStop   = "stop"
	ZAIFinishReasonLength = "length"
	ZAIFinishReasonTool   = "tool_calls"
	ZAIFinishReasonError  = "error"
	ZAIFinishReasonFilter = "content_filter"

	// Tool choices
	ZAIToolChoiceNone     = "none"
	ZAIToolChoiceAuto     = "auto"
	ZAIToolChoiceRequired = "required"

	// Response formats
	ZAIResponseFormatText       = "text"
	ZAIResponseFormatJSONObject = "json_object"

	// Content types
	ZAIContentTypeText     = "text"
	ZAIContentTypeImageURL = "image_url"
	ZAIContentTypeMedia    = "media"

	// Vision configuration
	ZAIVisionMaxImageSize = 10 * 1024 * 1024 // 10MB
	ZAIVisionQualityAuto  = "auto"
	ZAIVisionQualityLow   = "low"
	ZAIVisionQualityHigh  = "high"
	ZAIVisionDetailAuto   = "auto"
	ZAIVisionDetailLow    = "low"
	ZAIVisionDetailHigh   = "high"

	// Supported image formats
	ZAIVisionSupportedFormats = "image/jpeg,image/png,image/gif,image/webp"

	// Fashion-specific constants
	ZAIFashionVisionPrompt = "Analyze this fashion image in detail. Focus on clothing items, accessories, materials, colors, style, and fashion elements. Provide detailed descriptions suitable for fashion industry applications."
)
