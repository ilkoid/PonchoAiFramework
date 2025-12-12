package prompts

import (
	"time"
	
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PromptMetadata extends the interface definition with additional implementation details
type PromptMetadata struct {
	// Core metadata from interface
	interfaces.PromptMetadata
	
	// Implementation-specific fields
	FilePath    string                 `json:"file_path,omitempty"`
	FileSize    int64                  `json:"file_size,omitempty"`
	Hash        string                 `json:"hash,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// PromptConfig represents configuration for prompt system
type PromptConfig struct {
	// Template settings
	Templates struct {
		Directory     string   `yaml:"directory" json:"directory"`
		Extensions    []string `yaml:"extensions" json:"extensions"`
		AutoReload    bool     `yaml:"auto_reload" json:"auto_reload"`
		ReloadInterval string   `yaml:"reload_interval" json:"reload_interval"`
	} `yaml:"templates" json:"templates"`
	
	// Cache settings
	Cache struct {
		Enabled bool   `yaml:"enabled" json:"enabled"`
		Size    int    `yaml:"size" json:"size"`
		TTL     string `yaml:"ttl" json:"ttl"`
		Type    string `yaml:"type" json:"type"` // memory, redis
	} `yaml:"cache" json:"cache"`
	
	// Validation settings
	Validation struct {
		Strict          bool `yaml:"strict" json:"strict"`
		ValidateOnLoad  bool `yaml:"validate_on_load" json:"validate_on_load"`
		ValidateOnExec  bool `yaml:"validate_on_execute" json:"validate_on_execute"`
	} `yaml:"validation" json:"validation"`
	
	// Execution settings
	Execution struct {
		DefaultModel       string  `yaml:"default_model" json:"default_model"`
		DefaultMaxTokens   int     `yaml:"default_max_tokens" json:"default_max_tokens"`
		DefaultTemperature float32 `yaml:"default_temperature" json:"default_temperature"`
		Timeout           string  `yaml:"timeout" json:"timeout"`
		RetryAttempts     int     `yaml:"retry_attempts" json:"retry_attempts"`
		RetryDelay        string  `yaml:"retry_delay" json:"retry_delay"`
	} `yaml:"execution" json:"execution"`
	
	// Metrics settings
	Metrics struct {
		Enabled   bool   `yaml:"enabled" json:"enabled"`
		Interval  string `yaml:"interval" json:"interval"`
		Retention string `yaml:"retention" json:"retention"`
	} `yaml:"metrics" json:"metrics"`
	
	// Fashion-specific settings
	Fashion *FashionPromptConfig `yaml:"fashion,omitempty" json:"fashion,omitempty"`
}

// FashionPromptConfig represents fashion-specific configuration
type FashionPromptConfig struct {
	// Default context
	DefaultMarketplace string   `yaml:"default_marketplace" json:"default_marketplace"`
	DefaultLanguage    string   `yaml:"default_language" json:"default_language"`
	DefaultAudience    []string `yaml:"default_audience" json:"default_audience"`
	DefaultStyles      []string `yaml:"default_styles" json:"default_styles"`
	
	// Image analysis
	ImageAnalysis struct {
		DefaultFocus     []string `yaml:"default_focus" json:"default_focus"`
		DefaultFeatures  []string `yaml:"default_features" json:"default_features"`
		OutputFormat     string   `yaml:"output_format" json:"output_format"`
		MaxImageSize    int      `yaml:"max_image_size" json:"max_image_size"`
		SupportedTypes  []string `yaml:"supported_types" json:"supported_types"`
	} `yaml:"image_analysis" json:"image_analysis"`
	
	// Product classification
	Classification struct {
		CategoriesFile    string   `yaml:"categories_file" json:"categories_file"`
	CharacteristicsFile string   `yaml:"characteristics_file" json:"characteristics_file"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold" json:"confidence_threshold"`
	} `yaml:"classification" json:"classification"`
	
	// Content generation
	Generation struct {
		MaxDescriptionLength int      `yaml:"max_description_length" json:"max_description_length"`
		MinDescriptionLength int      `yaml:"min_description_length" json:"min_description_length"`
		SEOKeywords         []string `yaml:"seo_keywords" json:"seo_keywords"`
		ToneOfVoice        string   `yaml:"tone_of_voice" json:"tone_of_voice"`
	} `yaml:"generation" json:"generation"`
}

// PromptPart represents a part of a prompt template with additional implementation details
type PromptPart struct {
	// Core fields from interface
	interfaces.PromptPart
	
	// Implementation-specific fields
	Variables []string `json:"variables,omitempty"` // Variables used in this part
	Processed bool     `json:"processed,omitempty"` // Whether this part has been processed
	Index     int      `json:"index,omitempty"`     // Order index
}

// ValidationError represents a validation error with additional context
type ValidationError struct {
	// Core fields from interface
	interfaces.ValidationError
	
	// Additional context
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestion  string               `json:"suggestion,omitempty"`
	Severity    string               `json:"severity,omitempty"` // error, warning, info
}

// ValidationResult represents the result of validation with additional metrics
type ValidationResult struct {
	// Core fields from interface
	interfaces.ValidationResult
	
	// Additional metrics
	ValidationTime time.Duration `json:"validation_time,omitempty"`
	Validator     string        `json:"validator,omitempty"`
	Rules         []string      `json:"rules,omitempty"` // Validation rules applied
}

// TemplateContext represents execution context for a template
type TemplateContext struct {
	// Template information
	Template    *interfaces.PromptTemplate `json:"template"`
	Variables   map[string]interface{}     `json:"variables"`
	Parameters  map[string]interface{}     `json:"parameters,omitempty"`
	
	// Execution context
	Model       string    `json:"model"`
	RequestID   string    `json:"request_id,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	SessionID   string    `json:"session_id,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	
	// Fashion context
	Fashion     *interfaces.FashionContext `json:"fashion,omitempty"`
	
	// Execution options
	Options     *ExecutionOptions `json:"options,omitempty"`
}

// ExecutionOptions represents options for prompt execution
type ExecutionOptions struct {
	// Model overrides
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
	
	// Execution behavior
	Timeout     time.Duration `json:"timeout,omitempty"`
	Retry       *RetryOptions `json:"retry,omitempty"`
	Cache       *CacheOptions `json:"cache,omitempty"`
	
	// Validation
	Validate    bool `json:"validate,omitempty"`
	StrictMode  bool `json:"strict_mode,omitempty"`
	
	// Debugging
	Debug       bool   `json:"debug,omitempty"`
	TraceID     string `json:"trace_id,omitempty"`
}

// RetryOptions represents retry configuration for execution
type RetryOptions struct {
	MaxAttempts int           `json:"max_attempts"`
	Delay       time.Duration `json:"delay"`
	Backoff     string        `json:"backoff"` // linear, exponential
	MaxDelay    time.Duration `json:"max_delay,omitempty"`
}

// CacheOptions represents cache options for execution
type CacheOptions struct {
	Enabled    bool          `json:"enabled"`
	TTL        time.Duration `json:"ttl,omitempty"`
	Key        string        `json:"key,omitempty"`
	Invalidate bool          `json:"invalidate,omitempty"`
}

// ExecutionResult represents the result of prompt execution
type ExecutionResult struct {
	// Core response
	Response    *interfaces.PonchoModelResponse `json:"response"`
	
	// Execution metadata
	Context     *TemplateContext `json:"context"`
	Success     bool             `json:"success"`
	Error       *ExecutionError  `json:"error,omitempty"`
	
	// Performance metrics
	ExecutionTime time.Duration `json:"execution_time"`
	TokensUsed   int            `json:"tokens_used"`
	ModelUsed    string         `json:"model_used"`
	
	// Caching information
	CacheHit    bool          `json:"cache_hit"`
	CacheKey    string        `json:"cache_key,omitempty"`
	
	// Additional data
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionError represents an error during prompt execution
type ExecutionError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Type        string                 `json:"type"` // validation, execution, model, timeout
	Component   string                 `json:"component"` // parser, validator, executor, model
	Retryable   bool                   `json:"retryable"`
	Context     map[string]interface{} `json:"context,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
}

// TemplateStatistics represents statistics for a specific template
type TemplateStatistics struct {
	// Basic info
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Category    string    `json:"category"`
	
	// Usage statistics
	TotalExecutions     int64     `json:"total_executions"`
	SuccessfulExecutions int64    `json:"successful_executions"`
	FailedExecutions    int64     `json:"failed_executions"`
	SuccessRate         float64   `json:"success_rate"`
	
	// Performance metrics
	AvgExecutionTime    float64   `json:"avg_execution_time_ms"`
	MinExecutionTime    float64   `json:"min_execution_time_ms"`
	MaxExecutionTime    float64   `json:"max_execution_time_ms"`
	AvgTokensUsed      int       `json:"avg_tokens_used"`
	TotalTokensUsed    int64     `json:"total_tokens_used"`
	
	// Usage by model
	ByModel            map[string]*ModelUsageStats `json:"by_model"`
	
	// Usage by user
	ByUser             map[string]*UserUsageStats `json:"by_user,omitempty"`
	
	// Time-based statistics
	LastExecuted       time.Time `json:"last_executed"`
	FirstExecuted      time.Time `json:"first_executed,omitempty"`
	ExecutionsToday    int64     `json:"executions_today"`
	ExecutionsThisWeek int64     `json:"executions_this_week"`
	ExecutionsThisMonth int64    `json:"executions_this_month"`
	
	// Error statistics
	ErrorsByType       map[string]int64 `json:"errors_by_type"`
	LastError          *ExecutionError  `json:"last_error,omitempty"`
}

// ModelUsageStats represents usage statistics for a specific model
type ModelUsageStats struct {
	Model              string    `json:"model"`
	Executions         int64     `json:"executions"`
	SuccessRate        float64   `json:"success_rate"`
	AvgExecutionTime   float64   `json:"avg_execution_time_ms"`
	AvgTokensUsed      int       `json:"avg_tokens_used"`
	LastExecuted       time.Time `json:"last_executed"`
}

// UserUsageStats represents usage statistics for a specific user
type UserUsageStats struct {
	UserID             string    `json:"user_id"`
	Executions         int64     `json:"executions"`
	SuccessRate        float64   `json:"success_rate"`
	AvgExecutionTime   float64   `json:"avg_execution_time_ms"`
	TotalTokensUsed    int64     `json:"total_tokens_used"`
	LastExecuted       time.Time `json:"last_executed"`
}

// SystemMetrics represents overall prompt system metrics
type SystemMetrics struct {
	// Template metrics
	TotalTemplates     int64 `json:"total_templates"`
	LoadedTemplates    int64 `json:"loaded_templates"`
	CachedTemplates    int64 `json:"cached_templates"`
	
	// Execution metrics
	TotalExecutions    int64 `json:"total_executions"`
	SuccessfulExecutions int64 `json:"successful_executions"`
	FailedExecutions   int64 `json:"failed_executions"`
	AvgExecutionTime   float64 `json:"avg_execution_time_ms"`
	
	// Cache metrics
	CacheHits          int64   `json:"cache_hits"`
	CacheMisses        int64   `json:"cache_misses"`
	CacheHitRate       float64 `json:"cache_hit_rate"`
	
	// Error metrics
	TotalErrors        int64            `json:"total_errors"`
	ErrorsByType      map[string]int64 `json:"errors_by_type"`
	ErrorsByComponent map[string]int64 `json:"errors_by_component"`
	
	// Resource metrics
	MemoryUsage       int64 `json:"memory_usage_bytes"`
	CPUUsage          float64 `json:"cpu_usage_percent"`
	
	// Timestamps
	LastUpdated       time.Time `json:"last_updated"`
	StartTime         time.Time `json:"start_time"`
}

// VariableProcessor represents interface for processing template variables
type VariableProcessor interface {
	// ProcessVariables processes variables in template content
	ProcessVariables(content string, variables map[string]interface{}) (string, error)
	
	// ExtractVariables extracts variable references from content
	ExtractVariables(content string) ([]string, error)
	
	// ValidateVariables validates variables against template definition
	ValidateVariables(template *interfaces.PromptTemplate, variables map[string]interface{}) error
	
	// SetDefaults sets default values for missing variables
	SetDefaults(template *interfaces.PromptTemplate, variables map[string]interface{}) map[string]interface{}
}

// MediaProcessor represents interface for processing media content
type MediaProcessor interface {
	// ProcessMedia processes media content for prompts
	ProcessMedia(media *interfaces.MediaPart) (*interfaces.MediaPart, error)
	
	// ValidateMedia validates media content
	ValidateMedia(media *interfaces.MediaPart) error
	
	// OptimizeMedia optimizes media content for models
	OptimizeMedia(media *interfaces.MediaPart, model string) (*interfaces.MediaPart, error)
	
	// GetMediaInfo retrieves information about media content
	GetMediaInfo(url string) (*MediaInfo, error)
}

// MediaInfo represents information about media content
type MediaInfo struct {
	URL         string            `json:"url"`
	MimeType    string            `json:"mime_type"`
	Size        int64             `json:"size"`
	Width       int               `json:"width,omitempty"`
	Height      int               `json:"height,omitempty"`
	Duration    time.Duration     `json:"duration,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FashionData represents fashion-specific data structures
type FashionData struct {
	// Product information
	Product    *ProductInfo    `json:"product,omitempty"`
	
	// Image analysis results
	ImageAnalysis *ImageAnalysisResult `json:"image_analysis,omitempty"`
	
	// Marketplace data
	Marketplace *MarketplaceData `json:"marketplace,omitempty"`
	
	// Style and trends
	Style      *StyleInfo      `json:"style,omitempty"`
}

// ProductInfo represents fashion product information
type ProductInfo struct {
	ArticleID      string                 `json:"article_id"`
	Name           string                 `json:"name"`
	Category       string                 `json:"category"`
	Subcategory    string                 `json:"subcategory,omitempty"`
	Brand          string                 `json:"brand,omitempty"`
	Material       []string               `json:"material,omitempty"`
	Color          []string               `json:"color,omitempty"`
	Size           []string               `json:"size,omitempty"`
	Season         string                 `json:"season,omitempty"`
	Style          []string               `json:"style,omitempty"`
	TargetAudience []string               `json:"target_audience,omitempty"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty"`
	Price          *PriceInfo             `json:"price,omitempty"`
}

// ImageAnalysisResult represents results from image analysis
type ImageAnalysisResult struct {
	// Basic analysis
	Description    string                 `json:"description"`
	Objects        []string               `json:"objects"`
	Colors         []ColorInfo            `json:"colors"`
	Patterns       []string               `json:"patterns"`
	Textures       []string               `json:"textures"`
	
	// Fashion-specific
	GarmentType    string                 `json:"garment_type"`
	Style          []string               `json:"style"`
	Fit            string                 `json:"fit"`
	Occasion       []string               `json:"occasion"`
	Features       []string               `json:"features"`
	Quality        string                 `json:"quality"`
	
	// Technical details
	Confidence     float64                `json:"confidence"`
	Model          string                 `json:"model"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ColorInfo represents color information
type ColorInfo struct {
	Name       string  `json:"name"`
	Hex        string  `json:"hex"`
	RGB        []int   `json:"rgb"`
	Percentage float64 `json:"percentage"`
	Confidence float64 `json:"confidence"`
}

// MarketplaceData represents marketplace-specific data
type MarketplaceData struct {
	Marketplace    string                 `json:"marketplace"`
	CategoryID     int                    `json:"category_id,omitempty"`
	CategoryName   string                 `json:"category_name,omitempty"`
	SubjectID      int                    `json:"subject_id,omitempty"`
	SubjectName    string                 `json:"subject_name,omitempty"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty"`
	Requirements   map[string]interface{} `json:"requirements,omitempty"`
}

// StyleInfo represents style and trend information
type StyleInfo struct {
	PrimaryStyle    string   `json:"primary_style"`
	SecondaryStyles []string `json:"secondary_styles"`
	Season          string   `json:"season"`
	TrendLevel      string   `json:"trend_level"` // emerging, popular, classic
	TargetAudience  []string `json:"target_audience"`
	Occasions       []string `json:"occasions"`
	KeyFeatures     []string `json:"key_features"`
}

// PriceInfo represents price information
type PriceInfo struct {
	Currency     string  `json:"currency"`
	Amount       float64 `json:"amount"`
	Discount     float64 `json:"discount,omitempty"`
	OriginalAmount float64 `json:"original_amount,omitempty"`
}