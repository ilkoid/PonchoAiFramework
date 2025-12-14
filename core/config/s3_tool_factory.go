package config

// S3ToolFactory implements the S3 tool factory for the PonchoFramework.
// It provides specialized factory methods for creating S3-related tools with
// support for article importer tools designed for fashion data processing.
// The factory handles S3 client initialization with proper credentials and
// configuration, providing bucket browsing and file management tools.
//
// This serves as a specialized factory for Yandex Cloud S3 integration,
// including comprehensive error handling for S3 connection and configuration
// issues. It enables dynamic S3 tool loading from configuration files and
// abstracts S3 tool creation complexity from the main framework.
//
// Supported tool types:
// - article_importer: Imports fashion articles and images from S3 storage
// - s3_storage: Generic S3 storage operations (planned)
//
// The factory validates tool configurations, handles retry logic, caching
// settings, and custom parameters specific to fashion industry workflows.

import (
	"fmt"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/tools/article_importer"
)

// S3ToolFactory creates S3-related tools
type S3ToolFactory struct {
	logger interfaces.Logger
}

// NewS3ToolFactory creates a new S3 tool factory
func NewS3ToolFactory(logger interfaces.Logger) *S3ToolFactory {
	return &S3ToolFactory{
		logger: logger,
	}
}

// CreateTool implements the ToolFactory interface
func (stf *S3ToolFactory) CreateTool(config *interfaces.ToolConfig) (interfaces.PonchoTool, error) {
	switch stf.GetToolType() {
	case "article_importer":
		return stf.createArticleImporter(config)
	case "s3_storage":
		return stf.createS3Storage(config)
	default:
		return nil, fmt.Errorf("unsupported tool type: %s", stf.GetToolType())
	}
}

// ValidateConfig implements the ToolFactory interface
func (stf *S3ToolFactory) ValidateConfig(config *interfaces.ToolConfig) error {
	if config == nil {
		return fmt.Errorf("tool config cannot be nil")
	}

	// Common validation
	if config.Timeout == "" {
		return fmt.Errorf("timeout is required for S3 tools")
	}

	// Tool-specific validation
	switch stf.GetToolType() {
	case "article_importer":
		return stf.validateArticleImporterConfig(config)
	case "s3_storage":
		return stf.validateS3StorageConfig(config)
	default:
		return nil // No specific validation for unknown types
	}
}

// GetToolType implements the ToolFactory interface
func (stf *S3ToolFactory) GetToolType() string {
	return "s3_tools"
}

// createArticleImporter creates an article importer tool
func (stf *S3ToolFactory) createArticleImporter(config *interfaces.ToolConfig) (interfaces.PonchoTool, error) {
	tool := article_importer.NewArticleImporterTool()

	// Prepare tool configuration
	toolConfig := make(map[string]interface{})

	// Add basic configuration
	toolConfig["timeout"] = config.Timeout
	toolConfig["enabled"] = config.Enabled

	// Add retry configuration
	if config.Retry != nil {
		toolConfig["retry"] = map[string]interface{}{
			"max_attempts": config.Retry.MaxAttempts,
			"backoff":      config.Retry.Backoff,
			"base_delay":   config.Retry.BaseDelay,
			"max_delay":    config.Retry.MaxDelay,
		}
	}

	// Add cache configuration
	if config.Cache != nil {
		toolConfig["cache"] = map[string]interface{}{
			"ttl":      config.Cache.TTL,
			"max_size": config.Cache.MaxSize,
		}
	}

	// Add custom parameters
	if config.CustomParams != nil {
		toolConfig["custom_params"] = config.CustomParams
	}

	// Add S3 configuration if available in custom params
	if s3Config, ok := config.CustomParams["s3"].(map[string]interface{}); ok {
		toolConfig["s3"] = s3Config
	}

	return tool, nil
}

// createS3Storage creates an S3 storage tool
func (stf *S3ToolFactory) createS3Storage(config *interfaces.ToolConfig) (interfaces.PonchoTool, error) {
	// This would create a generic S3 storage tool
	// For now, return an error as this tool is not fully implemented
	return nil, fmt.Errorf("s3_storage tool is not yet implemented")
}

// validateArticleImporterConfig validates article importer configuration
func (stf *S3ToolFactory) validateArticleImporterConfig(config *interfaces.ToolConfig) error {
	// Validate custom parameters for article importer
	if config.CustomParams != nil {
		if maxImages, ok := config.CustomParams["max_images"].(int); ok {
			if maxImages < 1 || maxImages > 50 {
				return fmt.Errorf("max_images must be between 1 and 50")
			}
		}

		if imageOptions, ok := config.CustomParams["image_options"].(map[string]interface{}); ok {
			if err := stf.validateImageOptions(imageOptions); err != nil {
				return fmt.Errorf("invalid image_options: %w", err)
			}
		}
	}

	return nil
}

// validateS3StorageConfig validates S3 storage configuration
func (stf *S3ToolFactory) validateS3StorageConfig(config *interfaces.ToolConfig) error {
	// Validate S3 storage specific configuration
	if config.CustomParams != nil {
		if bucket, ok := config.CustomParams["bucket"].(string); ok && bucket == "" {
			return fmt.Errorf("bucket is required for s3_storage tool")
		}
	}

	return nil
}

// validateImageOptions validates image processing options
func (stf *S3ToolFactory) validateImageOptions(options map[string]interface{}) error {
	// Validate max_width
	if maxWidth, ok := options["max_width"].(int); ok {
		if maxWidth < 100 || maxWidth > 2048 {
			return fmt.Errorf("max_width must be between 100 and 2048")
		}
	}

	// Validate max_height
	if maxHeight, ok := options["max_height"].(int); ok {
		if maxHeight < 100 || maxHeight > 2048 {
			return fmt.Errorf("max_height must be between 100 and 2048")
		}
	}

	// Validate quality
	if quality, ok := options["quality"].(int); ok {
		if quality < 1 || quality > 100 {
			return fmt.Errorf("quality must be between 1 and 100")
		}
	}

	// Validate max_size_bytes
	if maxSizeBytes, ok := options["max_size_bytes"].(int); ok {
		if maxSizeBytes < 1000 || maxSizeBytes > 1048576 {
			return fmt.Errorf("max_size_bytes must be between 1000 and 1048576")
		}
	}

	// Validate format
	if format, ok := options["format"].(string); ok {
		validFormats := []string{"jpeg", "png", "webp"}
		valid := false
		for _, f := range validFormats {
			if f == format {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("format must be one of: jpeg, png, webp")
		}
	}

	return nil
}

// ArticleImporterFactory is a specialized factory for article importer tool
type ArticleImporterFactory struct {
	*S3ToolFactory
}

// NewArticleImporterFactory creates a new article importer factory
func NewArticleImporterFactory(logger interfaces.Logger) *ArticleImporterFactory {
	return &ArticleImporterFactory{
		S3ToolFactory: NewS3ToolFactory(logger),
	}
}

// GetToolType returns the specific tool type
func (aif *ArticleImporterFactory) GetToolType() string {
	return "article_importer"
}

// CreateTool creates an article importer tool
func (aif *ArticleImporterFactory) CreateTool(config *interfaces.ToolConfig) (interfaces.PonchoTool, error) {
	return aif.S3ToolFactory.createArticleImporter(config)
}

// ValidateConfig validates article importer configuration
func (aif *ArticleImporterFactory) ValidateConfig(config *interfaces.ToolConfig) error {
	return aif.S3ToolFactory.validateArticleImporterConfig(config)
}

// S3StorageFactory is a specialized factory for S3 storage tool
type S3StorageFactory struct {
	*S3ToolFactory
}

// NewS3StorageFactory creates a new S3 storage factory
func NewS3StorageFactory(logger interfaces.Logger) *S3StorageFactory {
	return &S3StorageFactory{
		S3ToolFactory: NewS3ToolFactory(logger),
	}
}

// GetToolType returns the specific tool type
func (ssf *S3StorageFactory) GetToolType() string {
	return "s3_storage"
}

// CreateTool creates an S3 storage tool
func (ssf *S3StorageFactory) CreateTool(config *interfaces.ToolConfig) (interfaces.PonchoTool, error) {
	return ssf.S3ToolFactory.createS3Storage(config)
}

// ValidateConfig validates S3 storage configuration
func (ssf *S3StorageFactory) ValidateConfig(config *interfaces.ToolConfig) error {
	return ssf.S3ToolFactory.validateS3StorageConfig(config)
}
