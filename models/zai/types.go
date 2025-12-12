package zai

import "time"

// GLM-specific types for Z.AI API integration

// GLMMessage represents a message in GLM API format
type GLMMessage struct {
	Role      string        `json:"role"`
	Content   interface{}   `json:"content"` // string or []GLMContentPart
	Name      *string       `json:"name,omitempty"`
	ToolCalls []GLMToolCall `json:"tool_calls,omitempty"`
}

// GLMContentPart represents a content part in multimodal messages
type GLMContentPart struct {
	Type     string       `json:"type"` // "text", "image_url"
	Text     string       `json:"text,omitempty"`
	ImageURL *GLMImageURL `json:"image_url,omitempty"`
}

// GLMImageURL represents an image URL in GLM format
type GLMImageURL struct {
	URL string `json:"url"`
}

// GLMToolCall represents a tool call in GLM API format
type GLMToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"` // always "function"
	Function GLMFunctionCall `json:"function"`
}

// GLMFunctionCall represents a function call in GLM API format
type GLMFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// GLMTool represents a tool definition in GLM API format
type GLMTool struct {
	Type     string          `json:"type"` // always "function"
	Function GLMToolFunction `json:"function"`
}

// GLMToolFunction represents a tool function definition
type GLMToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GLMThinking represents thinking mode configuration (GLM-specific)
type GLMThinking struct {
	Type string `json:"type"` // "enabled" or "disabled"
}

// GLMRequest represents a request to GLM API
type GLMRequest struct {
	Model            string       `json:"model"`
	Messages         []GLMMessage `json:"messages"`
	Temperature      *float32     `json:"temperature,omitempty"`
	MaxTokens        *int         `json:"max_tokens,omitempty"`
	Stream           bool         `json:"stream,omitempty"`
	Tools            []GLMTool    `json:"tools,omitempty"`
	ToolChoice       interface{}  `json:"tool_choice,omitempty"` // "none", "auto", or specific tool
	TopP             *float32     `json:"top_p,omitempty"`
	FrequencyPenalty *float32     `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32     `json:"presence_penalty,omitempty"`
	Stop             interface{}  `json:"stop,omitempty"` // string or []string
	Thinking         *GLMThinking `json:"thinking,omitempty"`
}

// GLMChoice represents a choice in GLM API response
type GLMChoice struct {
	Index        int        `json:"index"`
	Message      GLMMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

// GLMUsage represents token usage information
type GLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GLMResponse represents a response from GLM API
type GLMResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"` // "chat.completion"
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []GLMChoice `json:"choices"`
	Usage   GLMUsage    `json:"usage"`
}

// GLMStreamChoice represents a choice in streaming response
type GLMStreamChoice struct {
	Index        int            `json:"index"`
	Delta        GLMStreamDelta `json:"delta"`
	FinishReason *string        `json:"finish_reason,omitempty"`
}

// GLMStreamDelta represents a delta in streaming response
type GLMStreamDelta struct {
	Role      string        `json:"role,omitempty"`
	Content   string        `json:"content,omitempty"`
	ToolCalls []GLMToolCall `json:"tool_calls,omitempty"`
}

// GLMStreamResponse represents a streaming response from GLM API
type GLMStreamResponse struct {
	ID      string            `json:"id"`
	Object  string            `json:"object"` // "chat.completion.chunk"
	Created int64             `json:"created"`
	Model   string            `json:"model"`
	Choices []GLMStreamChoice `json:"choices"`
	Usage   *GLMUsage         `json:"usage,omitempty"`
}

// GLMError represents an error response from GLM API
type GLMError struct {
	Error GLMErrorDetail `json:"error"`
}

// GLMErrorDetail represents error details
type GLMErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// GLMConfig represents GLM-specific configuration
type GLMConfig struct {
	BaseURL          string        `json:"base_url"`
	APIKey           string        `json:"api_key"`
	Model            string        `json:"model"`
	MaxTokens        int           `json:"max_tokens"`
	Temperature      float32       `json:"temperature"`
	Timeout          time.Duration `json:"timeout"`
	TopP             *float32      `json:"top_p,omitempty"`
	FrequencyPenalty *float32      `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32      `json:"presence_penalty,omitempty"`
	Stop             interface{}   `json:"stop,omitempty"`
	Thinking         *GLMThinking  `json:"thinking,omitempty"`
}

// Constants for GLM API
const (
	GLMDefaultBaseURL = "https://api.z.ai/api/paas/v4"
	GLMDefaultModel   = "glm-4.6"
	GLMEndpoint       = "/chat/completions"
	GLMVisionModel    = "glm-4.6v"

	// Finish reasons
	GLMFinishReasonStop   = "stop"
	GLMFinishReasonLength = "length"
	GLMFinishReasonTool   = "tool_calls"
	GLMFinishReasonError  = "error"

	// Tool choices
	GLMToolChoiceNone = "none"
	GLMToolChoiceAuto = "auto"

	// Content types
	GLMContentTypeText     = "text"
	GLMContentTypeImageURL = "image_url"

	// Thinking types
	GLMThinkingEnabled  = "enabled"
	GLMThinkingDisabled = "disabled"
)

// Vision-specific types for fashion image analysis

// FashionAnalysis represents the result of fashion image analysis
type FashionAnalysis struct {
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Style       string                 `json:"style"`
	Materials   []string               `json:"materials"`
	Colors      []string               `json:"colors"`
	Season      string                 `json:"season"`
	Gender      string                 `json:"gender"`
	Features    map[string]interface{} `json:"features"`
	Coordinates []BoundingCoordinates  `json:"coordinates,omitempty"`
	Confidence  float64                `json:"confidence"`
}

// BoundingCoordinates represents bounding box coordinates
type BoundingCoordinates struct {
	XMin  float64 `json:"xmin"`
	YMin  float64 `json:"ymin"`
	XMax  float64 `json:"xmax"`
	YMax  float64 `json:"ymax"`
	Label string  `json:"label,omitempty"`
}

// ImageProcessingConfig represents configuration for image processing
type ImageProcessingConfig struct {
	MaxWidth       int     `json:"max_width"`
	MaxHeight      int     `json:"max_height"`
	Quality        int     `json:"quality"`
	MaxSizeBytes   int64   `json:"max_size_bytes"`
	EnableAnalysis bool    `json:"enable_analysis"`
	Confidence     float64 `json:"confidence_threshold"`
}

// Default image processing configuration
var DefaultImageProcessingConfig = ImageProcessingConfig{
	MaxWidth:       640,
	MaxHeight:      480,
	Quality:        90,
	MaxSizeBytes:   90000, // 90KB
	EnableAnalysis: true,
	Confidence:     0.7,
}
