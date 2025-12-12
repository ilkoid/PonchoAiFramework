package interfaces

import "time"

// Core data types for PonchoFramework

// PonchoRole represents the role in a conversation
type PonchoRole string

const (
	PonchoRoleSystem    PonchoRole = "system"
	PonchoRoleUser      PonchoRole = "user"
	PonchoRoleAssistant PonchoRole = "assistant"
	PonchoRoleTool      PonchoRole = "tool"
)

// PonchoContentType represents the type of content in a message
type PonchoContentType string

const (
	PonchoContentTypeText  PonchoContentType = "text"
	PonchoContentTypeMedia PonchoContentType = "media"
	PonchoContentTypeTool  PonchoContentType = "tool_call"
)

// PonchoFinishReason represents the reason why generation finished
type PonchoFinishReason string

const (
	PonchoFinishReasonStop   PonchoFinishReason = "stop"
	PonchoFinishReasonLength PonchoFinishReason = "length"
	PonchoFinishReasonTool   PonchoFinishReason = "tool_calls"
	PonchoFinishReasonError  PonchoFinishReason = "error"
)

// PonchoModelRequest represents a request to an AI model
type PonchoModelRequest struct {
	Model       string                 `json:"model"`
	Messages    []*PonchoMessage       `json:"messages"`
	Temperature *float32               `json:"temperature,omitempty"`
	MaxTokens   *int                   `json:"max_tokens,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Tools       []*PonchoToolDef       `json:"tools,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PonchoModelResponse represents a response from an AI model
type PonchoModelResponse struct {
	Message      *PonchoMessage         `json:"message"`
	Usage        *PonchoUsage           `json:"usage"`
	FinishReason PonchoFinishReason     `json:"finish_reason"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// PonchoMessage represents a message in a conversation
type PonchoMessage struct {
	Role    PonchoRole           `json:"role"`
	Content []*PonchoContentPart `json:"content"`
	Name    *string              `json:"name,omitempty"`
}

// PonchoContentPart represents a part of message content
type PonchoContentPart struct {
	Type  PonchoContentType `json:"type"`
	Text  string            `json:"text,omitempty"`
	Media *PonchoMediaPart  `json:"media,omitempty"`
	Tool  *PonchoToolPart   `json:"tool,omitempty"`
}

// PonchoMediaPart represents media content (images, videos, etc.)
type PonchoMediaPart struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type,omitempty"`
}

// PonchoToolPart represents a tool call in content
type PonchoToolPart struct {
	ID   string                 `json:"id"`
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// PonchoToolDef represents a tool definition for model
type PonchoToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// PonchoUsage represents token usage information
type PonchoUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// PonchoStreamChunk represents a chunk in streaming response
type PonchoStreamChunk struct {
	Delta        *PonchoMessage         `json:"delta"`
	Usage        *PonchoUsage           `json:"usage,omitempty"`
	FinishReason PonchoFinishReason     `json:"finish_reason,omitempty"`
	Done         bool                   `json:"done"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// PonchoFrameworkConfig represents the main framework configuration
type PonchoFrameworkConfig struct {
	Models      map[string]*ModelConfig `json:"models"`
	Tools       map[string]*ToolConfig  `json:"tools"`
	Flows       map[string]*FlowConfig  `json:"flows"`
	Logging     *LoggingConfig          `json:"logging"`
	Metrics     *MetricsConfig          `json:"metrics"`
	Cache       *CacheConfig            `json:"cache"`
	Security    *SecurityConfig         `json:"security"`
	S3          *S3Config               `json:"s3"`
	Wildberries *WildberriesConfig      `json:"wildberries"`
	
}

// ModelConfig represents configuration for a specific model
type ModelConfig struct {
	Provider     string                 `json:"provider"`
	ModelName    string                 `json:"model_name"`
	APIKey       string                 `json:"api_key"`
	BaseURL      string                 `json:"base_url,omitempty"`
	MaxTokens    int                    `json:"max_tokens"`
	Temperature  float32                `json:"temperature"`
	Timeout      string                 `json:"timeout"`
	Retry        *RetryConfig           `json:"retry,omitempty"`
	Supports     *ModelCapabilities     `json:"supports"`
	CustomParams map[string]interface{} `json:"custom_params,omitempty"`
}

// ModelCapabilities represents model capabilities
type ModelCapabilities struct {
	Streaming bool `json:"streaming"`
	Tools     bool `json:"tools"`
	Vision    bool `json:"vision"`
	System    bool `json:"system"`
}

// ToolConfig represents configuration for a specific tool
type ToolConfig struct {
	Enabled      bool                   `json:"enabled"`
	Timeout      string                 `json:"timeout"`
	Retry        *RetryConfig           `json:"retry,omitempty"`
	Cache        *ToolCacheConfig       `json:"cache,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	CustomParams map[string]interface{} `json:"custom_params,omitempty"`
}

// FlowConfig represents configuration for a specific flow
type FlowConfig struct {
	Enabled      bool                   `json:"enabled"`
	Timeout      string                 `json:"timeout"`
	Parallel     bool                   `json:"parallel"`
	Dependencies []string               `json:"dependencies,omitempty"`
	CustomParams map[string]interface{} `json:"custom_params,omitempty"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxAttempts int    `json:"max_attempts"`
	Backoff     string `json:"backoff"` // linear, exponential
	BaseDelay   string `json:"base_delay"`
	MaxDelay    string `json:"max_delay,omitempty"`
}

// ToolCacheConfig represents cache configuration for tools
type ToolCacheConfig struct {
	TTL     string `json:"ttl"`
	MaxSize int    `json:"max_size"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // text, json
	File   string `json:"file"`   // empty = stdout
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled  bool   `json:"enabled"`
	Interval string `json:"interval"`
	Endpoint string `json:"endpoint,omitempty"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Type     string `json:"type"` // memory, redis
	RedisURL string `json:"redis_url,omitempty"`
	TTL      string `json:"ttl"`
	MaxSize  int    `json:"max_size"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	APIKeys      []string          `json:"api_keys"`
	RateLimiting *RateLimitConfig  `json:"rate_limiting,omitempty"`
	Encryption   *EncryptionConfig `json:"encryption,omitempty"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int `json:"requests_per_minute"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Enabled   bool   `json:"enabled"`
	Algorithm string `json:"algorithm"`
}

// S3Config represents S3 configuration
type S3Config struct {
	URL      string `json:"url"`
	Region   string `json:"region"`
	Bucket   string `json:"bucket"`
	Endpoint string `json:"endpoint,omitempty"`
	UseSSL   bool   `json:"use_ssl"`
}

// WildberriesConfig represents Wildberries API configuration
type WildberriesConfig struct {
	BaseURL string `json:"base_url"`
	Timeout int    `json:"timeout"`
}

// PonchoHealthStatus represents health status of framework
type PonchoHealthStatus struct {
	Status     string                      `json:"status"`
	Timestamp  time.Time                   `json:"timestamp"`
	Version    string                      `json:"version"`
	Components map[string]*ComponentHealth `json:"components"`
	Uptime     time.Duration               `json:"uptime"`
}

// ComponentHealth represents health of a specific component
type ComponentHealth struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// PonchoMetrics represents framework metrics
type PonchoMetrics struct {
	GeneratedRequests *GenerationMetrics `json:"generated_requests"`
	ToolExecutions    *ToolMetrics       `json:"tool_executions"`
	FlowExecutions    *FlowMetrics       `json:"flow_executions"`
	Errors            *ErrorMetrics      `json:"errors"`
	System            *SystemMetrics     `json:"system"`
	Timestamp         time.Time          `json:"timestamp"`
}

// GenerationMetrics represents generation-related metrics
type GenerationMetrics struct {
	TotalRequests int64                    `json:"total_requests"`
	SuccessCount  int64                    `json:"success_count"`
	ErrorCount    int64                    `json:"error_count"`
	AvgLatency    float64                  `json:"avg_latency_ms"`
	TotalTokens   int64                    `json:"total_tokens"`
	ByModel       map[string]*ModelMetrics `json:"by_model"`
}

// ModelMetrics represents metrics for a specific model
type ModelMetrics struct {
	Requests    int64   `json:"requests"`
	SuccessRate float64 `json:"success_rate"`
	AvgLatency  float64 `json:"avg_latency_ms"`
	TotalTokens int64   `json:"total_tokens"`
}

// ToolMetrics represents tool execution metrics
type ToolMetrics struct {
	TotalExecutions int64            `json:"total_executions"`
	SuccessCount    int64            `json:"success_count"`
	ErrorCount      int64            `json:"error_count"`
	AvgLatency      float64          `json:"avg_latency_ms"`
	ByTool          map[string]int64 `json:"by_tool"`
}

// FlowMetrics represents flow execution metrics
type FlowMetrics struct {
	TotalExecutions int64            `json:"total_executions"`
	SuccessCount    int64            `json:"success_count"`
	ErrorCount      int64            `json:"error_count"`
	AvgLatency      float64          `json:"avg_latency_ms"`
	ByFlow          map[string]int64 `json:"by_flow"`
}

// ErrorMetrics represents error-related metrics
type ErrorMetrics struct {
	TotalErrors  int64            `json:"total_errors"`
	ByType       map[string]int64 `json:"by_type"`
	ByComponent  map[string]int64 `json:"by_component"`
	RecentErrors []*ErrorInfo     `json:"recent_errors"`
}

// ErrorInfo represents information about an error
type ErrorInfo struct {
	Timestamp time.Time `json:"timestamp"`
	Component string    `json:"component"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
}

// SystemMetrics represents system-related metrics
type SystemMetrics struct {
	MemoryUsage    int64   `json:"memory_usage_bytes"`
	CPUUsage       float64 `json:"cpu_usage_percent"`
	GoroutineCount int64   `json:"goroutine_count"`
	GCCount        int64   `json:"gc_count"`
	HeapSize       int64   `json:"heap_size_bytes"`
	HeapAlloc      int64   `json:"heap_alloc_bytes"`
}
