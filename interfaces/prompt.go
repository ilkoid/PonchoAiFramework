// Package interfaces provides prompt management system for PonchoFramework
//
// Key responsibilities:
// - Define prompt template management interfaces
// - Support fashion-specific prompt contexts
// - Enable template validation and execution
//
// Core interfaces:
// - PromptManager: Template loading and execution orchestration
// - PromptExecutor: Template execution with variable substitution
// - PromptValidator: Template syntax and semantic validation
// - PromptCache: LRU caching for template performance
//
// Implementation notes:
// - Supports V1 format with {{role "..."}} syntax
// - Fashion-specific context for industry workflows
// - Thread-safe caching with configurable TTL
// - Comprehensive validation with detailed error reporting
package interfaces

import (
	"context"
	"time"
)

// PromptManager defines the interface for managing prompts
type PromptManager interface {
	// LoadTemplate loads a prompt template from file or cache
	LoadTemplate(name string) (*PromptTemplate, error)
	
	// ExecutePrompt executes a prompt template with variables
	ExecutePrompt(ctx context.Context, name string, variables map[string]interface{}, model string) (*PonchoModelResponse, error)
	
	// ExecutePromptStreaming executes a prompt template with streaming
	ExecutePromptStreaming(ctx context.Context, name string, variables map[string]interface{}, model string, callback PonchoStreamCallback) error
	
	// ValidatePrompt validates a prompt template
	ValidatePrompt(template *PromptTemplate) (*ValidationResult, error)
	
	// ListTemplates returns list of available prompt templates
	ListTemplates() ([]string, error)
	
	// ReloadTemplates reloads all prompt templates
	ReloadTemplates() error
}

// PromptTemplate represents a prompt template with metadata
type PromptTemplate struct {
	// Basic metadata
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	
	// Template content
	Parts       []*PromptPart     `json:"parts"`
	Variables   []*PromptVariable `json:"variables"`
	
	// Execution settings
	Model       string            `json:"model,omitempty"`
	MaxTokens   *int              `json:"max_tokens,omitempty"`
	Temperature *float32          `json:"temperature,omitempty"`
	
	// Fashion-specific settings
	FashionContext *FashionContext `json:"fashion_context,omitempty"`
	
	// Metadata
	Metadata    *PromptMetadata   `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// PromptPart represents a part of a prompt template
type PromptPart struct {
	Type    PromptPartType `json:"type"`
	Content string         `json:"content"`
	Media   *MediaPart    `json:"media,omitempty"`
}

// PromptPartType represents the type of prompt part
type PromptPartType string

const (
	PromptPartTypeSystem PromptPartType = "system"
	PromptPartTypeUser   PromptPartType = "user"
	PromptPartTypeMedia  PromptPartType = "media"
)

// MediaPart represents media content in prompts
type MediaPart struct {
	URL      string            `json:"url"`
	MimeType string            `json:"mime_type,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// PromptVariable represents a variable definition in template
type PromptVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`         // string, number, boolean, array, object
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Validation   *VariableValidation `json:"validation,omitempty"`
}

// VariableValidation represents validation rules for variables
type VariableValidation struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
}

// FashionContext represents fashion-specific context for prompts
type FashionContext struct {
	// Target audience
	TargetAudience []string `json:"target_audience,omitempty"`
	
	// Product type
	ProductTypes   []string `json:"product_types,omitempty"`
	
	// Style preferences
	Styles         []string `json:"styles,omitempty"`
	
	// Seasonal context
	Seasons        []string `json:"seasons,omitempty"`
	
	// Market context
	Marketplace    string   `json:"marketplace,omitempty"`
	Language       string   `json:"language,omitempty"`
	
	// Image analysis context
	ImageAnalysis  *ImageAnalysisContext `json:"image_analysis,omitempty"`
}

// ImageAnalysisContext represents context for image analysis
type ImageAnalysisContext struct {
	FocusAreas     []string `json:"focus_areas,omitempty"`     // fabric, style, fit, details, etc.
	ExtractFeatures []string `json:"extract_features,omitempty"` // color, pattern, texture, etc.
	OutputFormat   string   `json:"output_format,omitempty"`    // json, structured, narrative
}

// PromptMetadata represents additional metadata for prompts
type PromptMetadata struct {
	Author      string            `json:"author,omitempty"`
	License     string            `json:"license,omitempty"`
	Source      string            `json:"source,omitempty"`
	Usage       map[string]int    `json:"usage,omitempty"`        // usage count by date
	Performance *PerformanceMetrics `json:"performance,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// PerformanceMetrics represents performance metrics for prompts
type PerformanceMetrics struct {
	AvgResponseTime   float64 `json:"avg_response_time_ms"`
	AvgTokensUsed     int     `json:"avg_tokens_used"`
	SuccessRate       float64 `json:"success_rate"`
	LastExecuted      time.Time `json:"last_executed"`
	ExecutionCount   int64   `json:"execution_count"`
}

// PromptExecutor defines the interface for executing prompts
type PromptExecutor interface {
	// ExecuteTemplate executes a prompt template
	ExecuteTemplate(ctx context.Context, template *PromptTemplate, variables map[string]interface{}, modelName string) (*PonchoModelResponse, error)
	
	// ExecuteTemplateStreaming executes a prompt template with streaming
	ExecuteTemplateStreaming(ctx context.Context, template *PromptTemplate, variables map[string]interface{}, modelName string, callback PonchoStreamCallback) error
	
	// BuildModelRequest builds a model request from template
	BuildModelRequest(template *PromptTemplate, variables map[string]interface{}, modelName string) (*PonchoModelRequest, error)
}

// PromptValidator defines the interface for validating prompts
type PromptValidator interface {
	// ValidateTemplate validates a prompt template
	ValidateTemplate(template *PromptTemplate) (*ValidationResult, error)
	
	// ValidateVariables validates template variables
	ValidateVariables(template *PromptTemplate, variables map[string]interface{}) (*ValidationResult, error)
	
	// ValidateSyntax validates template syntax
	ValidateSyntax(content string) (*ValidationResult, error)
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid   bool               `json:"valid"`
	Errors  []*ValidationError `json:"errors,omitempty"`
	Warnings []*ValidationWarning `json:"warnings,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// PromptCache defines the interface for caching prompts
type PromptCache interface {
	// GetTemplate gets a cached template
	GetTemplate(name string) (*PromptTemplate, bool)
	
	// SetTemplate sets a template in cache
	SetTemplate(name string, template *PromptTemplate)
	
	// InvalidateTemplate invalidates a cached template
	InvalidateTemplate(name string)
	
	// Clear clears all cached templates
	Clear()
	
	// Stats returns cache statistics
	Stats() *CacheStats
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Size        int64 `json:"size"`
	MaxSize     int64 `json:"max_size"`
	HitRate     float64 `json:"hit_rate"`
}

// PromptTemplateLoader defines the interface for loading templates
type PromptTemplateLoader interface {
	// LoadFromFile loads template from file
	LoadFromFile(filePath string) (*PromptTemplate, error)
	
	// LoadFromDirectory loads all templates from directory
	LoadFromDirectory(dirPath string) (map[string]*PromptTemplate, error)
	
	// SaveToFile saves template to file
	SaveToFile(template *PromptTemplate, filePath string) error
	
	// ValidateFile validates template file format
	ValidateFile(filePath string) error
}

// PromptMetrics represents metrics for prompt system
type PromptMetrics struct {
	// Execution metrics
	TotalExecutions    int64             `json:"total_executions"`
	SuccessfulExecutions int64            `json:"successful_executions"`
	FailedExecutions   int64             `json:"failed_executions"`
	AvgExecutionTime   float64           `json:"avg_execution_time_ms"`
	
	// Template metrics
	TemplatesLoaded    int64             `json:"templates_loaded"`
	TemplatesCached    int64             `json:"templates_cached"`
	CacheHits         int64             `json:"cache_hits"`
	CacheMisses       int64             `json:"cache_misses"`
	
	// Performance by template
	ByTemplate        map[string]*TemplateMetrics `json:"by_template"`
	
	// Performance by model
	ByModel          map[string]*ModelPromptMetrics `json:"by_model"`
	
	// Timestamp
	LastUpdated       time.Time         `json:"last_updated"`
}

// TemplateMetrics represents metrics for a specific template
type TemplateMetrics struct {
	Executions       int64   `json:"executions"`
	SuccessRate      float64 `json:"success_rate"`
	AvgResponseTime  float64 `json:"avg_response_time_ms"`
	AvgTokensUsed    int     `json:"avg_tokens_used"`
	LastExecuted     time.Time `json:"last_executed"`
}

// ModelPromptMetrics represents metrics for prompt execution by model
type ModelPromptMetrics struct {
	Executions       int64   `json:"executions"`
	SuccessRate      float64 `json:"success_rate"`
	AvgResponseTime  float64 `json:"avg_response_time_ms"`
	AvgTokensUsed    int     `json:"avg_tokens_used"`
	LastExecuted     time.Time `json:"last_executed"`
}

// PromptSystemConfig represents configuration for prompt system
type PromptSystemConfig struct {
	// Template loading
	TemplatesDirectory    string   `json:"templates_directory"`
	TemplateExtensions    []string `json:"template_extensions"`
	AutoReload           bool     `json:"auto_reload"`
	ReloadInterval       string   `json:"reload_interval"`
	
	// Caching
	CacheEnabled         bool     `json:"cache_enabled"`
	CacheSize           int      `json:"cache_size"`
	CacheTTL            string   `json:"cache_ttl"`
	
	// Validation
	StrictValidation     bool     `json:"strict_validation"`
	ValidateOnLoad      bool     `json:"validate_on_load"`
	ValidateOnExecute   bool     `json:"validate_on_execute"`
	
	// Execution
	DefaultModel        string   `json:"default_model"`
	DefaultMaxTokens    int      `json:"default_max_tokens"`
	DefaultTemperature  float32  `json:"default_temperature"`
	ExecutionTimeout   string   `json:"execution_timeout"`
	
	// Metrics
	MetricsEnabled      bool     `json:"metrics_enabled"`
	MetricsInterval    string   `json:"metrics_interval"`
	
	// Fashion-specific
	FashionDefaults     *FashionDefaults `json:"fashion_defaults,omitempty"`
}

// FashionDefaults represents default fashion context settings
type FashionDefaults struct {
	Marketplace        string   `json:"marketplace"`
	Language           string   `json:"language"`
	TargetAudience     []string `json:"target_audience"`
	DefaultStyles      []string `json:"default_styles"`
	SeasonalContext    []string `json:"seasonal_context"`
	ImageAnalysisFocus []string `json:"image_analysis_focus"`
}