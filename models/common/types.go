// Package common provides shared types, constants, and configurations for AI model
// providers in PonchoFramework. This file defines the foundational types used
// across all model implementations with provider-specific configurations and defaults.
//
// Key Type Categories:
// - Provider Types: DeepSeek, Z.AI, OpenAI, Custom
// - Model Types: Text, Vision, Multimodal, Embedding
// - Configuration Types: HTTP, Retry, Validation, Model capabilities
// - Error Types: Standardized error codes and handling
// - Content Types: Text, Media, Tool, Image URL
//
// Provider Support:
// - DeepSeek: OpenAI-compatible, text generation focus
// - Z.AI: Custom API, vision and multimodal support
// - OpenAI: Standard OpenAI API compatibility
// - Custom: Extensible for new providers
//
// Configuration Management:
// - HTTP client settings with connection pooling
// - Retry strategies with backoff and jitter
// - Model capabilities and limitations
// - Default configurations for all providers
// - Validation rules and constraints
//
// Usage Examples:
//   config := GetDefaultConfigForProvider(ProviderDeepSeek)
//   client := NewHTTPClient(&DefaultHTTPConfig, DefaultRetryConfig)
//   validator := NewValidator(DefaultValidationRules, logger)
//
// Type Safety:
// - Strongly typed constants for all enums
// - Comprehensive validation rules
// - Provider-specific configurations
// - Extensible design for new models
package common

import (
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// Common types and constants shared across all model implementations

// Provider represents an AI model provider
type Provider string

const (
	ProviderDeepSeek Provider = "deepseek"
	ProviderZAI     Provider = "zai"
	ProviderOpenAI   Provider = "openai"
	ProviderCustom   Provider = "custom"
)

// ModelType represents the type of AI model
type ModelType string

const (
	ModelTypeText     ModelType = "text"
	ModelTypeVision   ModelType = "vision"
	ModelTypeMultimodal ModelType = "multimodal"
	ModelTypeEmbedding ModelType = "embedding"
)

// FinishReason represents why a model response finished
type FinishReason string

const (
	FinishReasonStop   FinishReason = "stop"
	FinishReasonLength FinishReason = "length"
	FinishReasonTool   FinishReason = "tool_calls"
	FinishReasonError  FinishReason = "error"
	FinishReasonFilter FinishReason = "content_filter"
)

// ToolChoice represents tool choice options
type ToolChoice string

const (
	ToolChoiceNone ToolChoice = "none"
	ToolChoiceAuto ToolChoice = "auto"
	ToolChoiceRequired ToolChoice = "required"
)

// ContentType represents content part types
type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeImageURL ContentType = "image_url"
	ContentTypeMedia    ContentType = "media"
	ContentTypeTool     ContentType = "tool"
)

// ResponseFormat represents response format options
type ResponseFormat string

const (
	ResponseFormatText       ResponseFormat = "text"
	ResponseFormatJSONObject ResponseFormat = "json_object"
)

// ThinkingType represents thinking mode options
type ThinkingType string

const (
	ThinkingEnabled  ThinkingType = "enabled"
	ThinkingDisabled ThinkingType = "disabled"
)

// CommonModelConfig represents shared configuration for all models
type CommonModelConfig struct {
	Provider         Provider        `json:"provider"`
	Model            string          `json:"model"`
	ModelType        ModelType       `json:"model_type"`
	BaseURL          string          `json:"base_url"`
	APIKey           string          `json:"api_key"`
	MaxTokens        int             `json:"max_tokens"`
	Temperature      float32         `json:"temperature"`
	Timeout          time.Duration   `json:"timeout"`
	TopP             *float32        `json:"top_p,omitempty"`
	FrequencyPenalty *float32        `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32        `json:"presence_penalty,omitempty"`
	Stop             interface{}     `json:"stop,omitempty"`
	ResponseFormat   *ResponseFormat `json:"response_format,omitempty"`
	Thinking         *ThinkingType   `json:"thinking,omitempty"`
	LogProbs         bool            `json:"logprobs,omitempty"`
	TopLogProbs      *int            `json:"top_logprobs,omitempty"`
}

// ModelCapabilities represents what a model can do
type ModelCapabilities struct {
	SupportsStreaming bool     `json:"supports_streaming"`
	SupportsTools     bool     `json:"supports_tools"`
	SupportsVision     bool     `json:"supports_vision"`
	SupportsSystem     bool     `json:"supports_system"`
	SupportsThinking   bool     `json:"supports_thinking"`
	SupportedTypes     []ContentType `json:"supported_types"`
	MaxInputTokens     int      `json:"max_input_tokens"`
	MaxOutputTokens    int      `json:"max_output_tokens"`
}

// ModelMetadata represents metadata about a model
type ModelMetadata struct {
	Provider      Provider         `json:"provider"`
	Model         string           `json:"model"`
	ModelType     ModelType        `json:"model_type"`
	Capabilities  ModelCapabilities `json:"capabilities"`
	Version       string           `json:"version"`
	Description   string           `json:"description"`
	CostPer1KTokens float64       `json:"cost_per_1k_tokens"`
	TrainedAt     *time.Time      `json:"trained_at,omitempty"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// RateLimitInfo represents rate limiting information
type RateLimitInfo struct {
	RequestsPerMinute int           `json:"requests_per_minute"`
	TokensPerMinute   int           `json:"tokens_per_minute"`
	RetryAfter        *time.Duration `json:"retry_after,omitempty"`
	ResetTime         *time.Time     `json:"reset_time,omitempty"`
}

// ErrorInfo represents standardized error information
type ErrorInfo struct {
	Code       string    `json:"code"`
	Type       string    `json:"type"`
	Message    string    `json:"message"`
	Provider   Provider  `json:"provider"`
	Retryable  bool      `json:"retryable"`
	Timestamp  time.Time `json:"timestamp"`
	RequestID  string    `json:"request_id,omitempty"`
	StatusCode int       `json:"status_code,omitempty"`
}

// RequestMetrics represents metrics for a request
type RequestMetrics struct {
	RequestID       string        `json:"request_id"`
	Provider        Provider      `json:"provider"`
	Model           string        `json:"model"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Duration        time.Duration `json:"duration"`
	TokenUsage      TokenUsage    `json:"token_usage"`
	Success         bool          `json:"success"`
	Error           *ErrorInfo    `json:"error,omitempty"`
	StreamChunks    int           `json:"stream_chunks,omitempty"`
	CacheHit        bool          `json:"cache_hit,omitempty"`
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Field     string      `json:"field"`
	Required  bool        `json:"required"`
	MinLength *int        `json:"min_length,omitempty"`
	MaxLength *int        `json:"max_length,omitempty"`
	MinValue  *float64    `json:"min_value,omitempty"`
	MaxValue  *float64    `json:"max_value,omitempty"`
	Pattern   *string     `json:"pattern,omitempty"`
	Enum      []string    `json:"enum,omitempty"`
	Type      string      `json:"type"`
	Custom    func(interface{}) error `json:"-"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	BaseDelay       time.Duration `json:"base_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffType     BackoffType   `json:"backoff_type"`
	Jitter          bool          `json:"jitter"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// BackoffType represents backoff strategy
type BackoffType string

const (
	BackoffTypeLinear      BackoffType = "linear"
	BackoffTypeExponential BackoffType = "exponential"
	BackoffTypeFixed       BackoffType = "fixed"
)

// HTTPConfig represents HTTP client configuration
type HTTPConfig struct {
	Timeout              time.Duration `json:"timeout"`
	MaxIdleConns         int           `json:"max_idle_conns"`
	MaxIdleConnsPerHost  int           `json:"max_idle_conns_per_host"`
	IdleConnTimeout      time.Duration `json:"idle_conn_timeout"`
	TLSHandshakeTimeout  time.Duration `json:"tls_handshake_timeout"`
	MaxConnsPerHost      int           `json:"max_conns_per_host"`
	MaxResponseHeader    int64         `json:"max_response_header"`
	ReadBufferSize       int           `json:"read_buffer_size"`
	WriteBufferSize      int           `json:"write_buffer_size"`
	DisableCompression   bool          `json:"disable_compression"`
	DisableKeepAlives    bool          `json:"disable_keep_alives"`
	DisableRedirects     bool          `json:"disable_redirects"`
	InsecureSkipVerify   bool          `json:"insecure_skip_verify"`
	ProxyURL            string        `json:"proxy_url,omitempty"`
	UserAgent            string        `json:"user_agent,omitempty"`
}

// Default configurations
var (
	DefaultHTTPConfig = HTTPConfig{
		Timeout:             30 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:      90 * time.Second,
		TLSHandshakeTimeout:  10 * time.Second,
		MaxConnsPerHost:      100,
		MaxResponseHeader:    1 << 20, // 1MB
		ReadBufferSize:       32 << 10, // 32KB
		WriteBufferSize:      32 << 10, // 32KB
		DisableCompression:   false,
		DisableKeepAlives:    false,
		DisableRedirects:     false,
		InsecureSkipVerify:   false,
		UserAgent:            "PonchoFramework/1.0",
	}

	DefaultRetryConfig = RetryConfig{
		MaxAttempts:     3,
		BaseDelay:       1 * time.Second,
		MaxDelay:        30 * time.Second,
		BackoffType:     BackoffTypeExponential,
		Jitter:          true,
		RetryableErrors: []string{"timeout", "rate_limit", "server_error", "network_error"},
	}

	DefaultValidationRules = []ValidationRule{
		{
			Field:    "model",
			Required: true,
			Type:     "string",
			MinLength: func() *int { i := 1; return &i }(),
		},
		{
			Field:    "messages",
			Required: true,
			Type:     "array",
			MinLength: func() *int { i := 1; return &i }(),
		},
		{
			Field:   "temperature",
			Type:    "number",
			MinValue: func() *float64 { f := 0.0; return &f }(),
			MaxValue: func() *float64 { f := 2.0; return &f }(),
		},
		{
			Field:   "max_tokens",
			Type:    "number",
			MinValue: func() *float64 { f := 1.0; return &f }(),
			MaxValue: func() *float64 { f := 32000.0; return &f }(),
		},
		{
			Field:   "top_p",
			Type:    "number",
			MinValue: func() *float64 { f := 0.0; return &f }(),
			MaxValue: func() *float64 { f := 1.0; return &f }(),
		},
	}
)

// Provider-specific constants
const (
	// DeepSeek
	DeepSeekDefaultBaseURL = "https://api.deepseek.com"
	DeepSeekDefaultModel   = "deepseek-chat"
	DeepSeekEndpoint       = "/chat/completions"

	// Z.AI
	ZAIDefaultBaseURL = "https://api.z.ai/api/paas/v4"
	ZAIDefaultModel   = "glm-4.6"
	ZAIEndpoint       = "/chat/completions"
	ZAIVisionModel    = "glm-4.6v"

	// Common headers
	HeaderContentType     = "Content-Type"
	HeaderAccept         = "Accept"
	HeaderAuthorization  = "Authorization"
	HeaderUserAgent      = "User-Agent"
	HeaderCacheControl   = "Cache-Control"
	HeaderXRequestID     = "X-Request-ID"
	HeaderRetryAfter     = "Retry-After"

	// MIME types
	MIMETypeJSON         = "application/json"
	MIMETypeEventStream  = "text/event-stream"
	MIMETypeText         = "text/plain"
	MIMETypeHTML        = "text/html"
)

// Common error codes
const (
	ErrCodeInvalidRequest    = "invalid_request"
	ErrCodeInvalidAuth       = "invalid_auth"
	ErrCodeRateLimit         = "rate_limit"
	ErrCodeInsufficientQuota = "insufficient_quota"
	ErrCodeModelNotFound     = "model_not_found"
	ErrCodeContentFilter     = "content_filter"
	ErrCodeServerError       = "server_error"
	ErrCodeTimeout           = "timeout"
	ErrCodeNetworkError      = "network_error"
	ErrCodeParsingError      = "parsing_error"
	ErrCodeValidationError   = "validation_error"
	ErrCodeUnknown          = "unknown_error"
)

// Helper functions for type conversions

// ToPonchoFinishReason converts common finish reason to PonchoFinishReason
func ToPonchoFinishReason(reason FinishReason) interfaces.PonchoFinishReason {
	switch reason {
	case FinishReasonStop:
		return interfaces.PonchoFinishReasonStop
	case FinishReasonLength:
		return interfaces.PonchoFinishReasonLength
	case FinishReasonTool:
		return interfaces.PonchoFinishReasonTool
	case FinishReasonError:
		return interfaces.PonchoFinishReasonError
	default:
		return interfaces.PonchoFinishReasonStop
	}
}

// FromPonchoFinishReason converts PonchoFinishReason to common finish reason
func FromPonchoFinishReason(reason interfaces.PonchoFinishReason) FinishReason {
	switch reason {
	case interfaces.PonchoFinishReasonStop:
		return FinishReasonStop
	case interfaces.PonchoFinishReasonLength:
		return FinishReasonLength
	case interfaces.PonchoFinishReasonTool:
		return FinishReasonTool
	case interfaces.PonchoFinishReasonError:
		return FinishReasonError
	default:
		return FinishReasonStop
	}
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(errorCode string) bool {
	retryableErrors := map[string]bool{
		ErrCodeRateLimit:         true,
		ErrCodeServerError:       true,
		ErrCodeTimeout:           true,
		ErrCodeNetworkError:      true,
	}
	return retryableErrors[errorCode]
}

// GetDefaultConfigForProvider returns default configuration for a provider
func GetDefaultConfigForProvider(provider Provider) CommonModelConfig {
	switch provider {
	case ProviderDeepSeek:
		return CommonModelConfig{
			Provider:         ProviderDeepSeek,
			Model:            DeepSeekDefaultModel,
			BaseURL:          DeepSeekDefaultBaseURL,
			MaxTokens:        4000,
			Temperature:      0.7,
			Timeout:          30 * time.Second,
		}
	case ProviderZAI:
		return CommonModelConfig{
			Provider:         ProviderZAI,
			Model:            ZAIDefaultModel,
			BaseURL:          ZAIDefaultBaseURL,
			MaxTokens:        2000,
			Temperature:      0.5,
			Timeout:          60 * time.Second,
		}
	default:
		return CommonModelConfig{
			Provider:         provider,
			MaxTokens:        4000,
			Temperature:      0.7,
			Timeout:          30 * time.Second,
		}
	}
}