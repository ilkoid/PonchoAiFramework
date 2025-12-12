package article_importer

import (
	"context"
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/tools/s3"
)

// ArticleImporterTool imports article data from S3 storage
type ArticleImporterTool struct {
	*base.PonchoBaseTool
	s3Client *s3.S3Client
}

// NewArticleImporterTool creates a new article importer tool
func NewArticleImporterTool() *ArticleImporterTool {
	tool := base.NewPonchoBaseTool(
		"article_importer",
		"Imports article data from S3 storage with image processing capabilities",
		"1.0.0",
		"data_import",
	)

	// Add tags
	tool.AddTag("s3")
	tool.AddTag("article")
	tool.AddTag("image_processing")
	tool.AddTag("fashion")

	// Set input schema
	inputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"article_id": map[string]interface{}{
				"type":        "string",
				"description": "Unique identifier of the article to import",
				"pattern":     "^[0-9]+$",
			},
			"include_images": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to download and process images",
				"default":     true,
			},
			"max_images": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of images to download",
				"default":     10,
				"minimum":     1,
				"maximum":     50,
			},
			"image_options": map[string]interface{}{
				"type":        "object",
				"description": "Image processing options",
				"properties": map[string]interface{}{
					"enabled": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable image processing",
						"default":     true,
					},
					"max_width": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum image width",
						"default":     640,
						"minimum":     100,
						"maximum":     2048,
					},
					"max_height": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum image height",
						"default":     480,
						"minimum":     100,
						"maximum":     2048,
					},
					"quality": map[string]interface{}{
						"type":        "integer",
						"description": "JPEG quality (1-100)",
						"default":     90,
						"minimum":     1,
						"maximum":     100,
					},
					"max_size_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum file size in bytes",
						"default":     90000,
						"minimum":     1000,
						"maximum":     1048576, // 1MB
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Output image format",
						"default":     "jpeg",
						"enum":        []string{"jpeg", "png", "webp"},
					},
				},
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds",
				"default":     30,
				"minimum":     5,
				"maximum":     300,
			},
		},
		"required": []string{"article_id"},
	}
	tool.SetInputSchema(inputSchema)

	// Set output schema
	outputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type": "boolean",
			},
			"article": map[string]interface{}{
				"type":        "object",
				"description": "Imported article data",
			},
			"error": map[string]interface{}{
				"type":        "object",
				"description": "Error details if import failed",
			},
			"metadata": map[string]interface{}{
				"type":        "object",
				"description": "Import metadata",
			},
		},
	}
	tool.SetOutputSchema(outputSchema)

	return &ArticleImporterTool{
		PonchoBaseTool: tool,
	}
}

// Execute implements the PonchoTool interface
func (t *ArticleImporterTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	startTime := time.Now()

	// Parse input
	request, err := t.parseInput(input)
	if err != nil {
		return t.PrepareError(err, map[string]interface{}{
			"input_type": fmt.Sprintf("%T", input),
		}), nil
	}

	t.GetLogger().Info("Starting article import",
		"article_id", request.ArticleID,
		"include_images", request.IncludeImages,
		"max_images", request.MaxImages,
	)

	// Validate article ID
	if request.ArticleID == "" {
		err := fmt.Errorf("article_id is required")
		return t.PrepareError(err, map[string]interface{}{
			"validation": "missing_required_field",
		}), nil
	}

	// Create S3 download request
	s3Request := &s3.S3DownloadRequest{
		ArticleID:     request.ArticleID,
		IncludeImages: request.IncludeImages,
		ImageOptions:  request.ImageOptions,
		MaxImages:     request.MaxImages,
		Timeout:       request.Timeout,
	}

	// Execute S3 download
	s3Response, err := t.s3Client.DownloadArticle(ctx, s3Request)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		t.LogExecution(ctx, input, nil, err, duration)

		return t.PrepareError(err, map[string]interface{}{
			"article_id":    request.ArticleID,
			"s3_error_code": s3Response.Error.Code,
			"retryable":     s3Response.Error.Retryable,
		}), nil
	}

	// Prepare response
	response := map[string]interface{}{
		"success": s3Response.Success,
		"metadata": map[string]interface{}{
			"request_id":  s3Response.Metadata.RequestID,
			"duration_ms": s3Response.Metadata.Duration,
			"bucket":      s3Response.Metadata.Bucket,
			"region":      s3Response.Metadata.Region,
			"retry_count": s3Response.Metadata.RetryCount,
		},
	}

	if s3Response.Success && s3Response.Article != nil {
		response["article"] = s3Response.Article

		t.GetLogger().Info("Article import completed successfully",
			"article_id", request.ArticleID,
			"image_count", len(s3Response.Article.Images),
			"total_size", s3Response.Article.Metadata.TotalSize,
		)
	} else if s3Response.Error != nil {
		response["error"] = s3Response.Error
	}

	duration := time.Since(startTime).Milliseconds()
	t.LogExecution(ctx, input, response, nil, duration)

	return t.PrepareOutput(response, map[string]interface{}{
		"article_id":    request.ArticleID,
		"image_count":   len(s3Response.Article.Images),
		"processing_ms": s3Response.Article.Metadata.ProcessingTime,
	}), nil
}

// Initialize implements the PonchoTool interface
func (t *ArticleImporterTool) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Initialize base tool
	if err := t.PonchoBaseTool.Initialize(ctx, config); err != nil {
		return err
	}

	// Create S3 client configuration
	s3Config := t.extractS3Config(config)

	// Create S3 client
	s3Client, err := s3.NewS3Client(s3Config, t.GetLogger())
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	t.s3Client = s3Client
	t.GetLogger().Info("Article importer tool initialized",
		"s3_bucket", s3Config.Bucket,
		"s3_region", s3Config.Region,
	)

	return nil
}

// Shutdown implements the PonchoTool interface
func (t *ArticleImporterTool) Shutdown(ctx context.Context) error {
	t.s3Client = nil
	return t.PonchoBaseTool.Shutdown(ctx)
}

// parseInput parses and validates the input
func (t *ArticleImporterTool) parseInput(input interface{}) (*ImportRequest, error) {
	// Convert input to map
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("input must be a map[string]interface{}")
	}

	request := &ImportRequest{
		ImageOptions: s3.DefaultImageProcessingOptions(),
	}

	// Parse required fields
	if articleID, ok := inputMap["article_id"].(string); ok {
		request.ArticleID = articleID
	}

	// Parse optional fields
	if includeImages, ok := inputMap["include_images"].(bool); ok {
		request.IncludeImages = includeImages
	} else {
		request.IncludeImages = true // Default
	}

	if maxImages, ok := inputMap["max_images"].(int); ok {
		request.MaxImages = maxImages
	} else {
		request.MaxImages = 10 // Default
	}

	if timeout, ok := inputMap["timeout"].(int); ok {
		request.Timeout = timeout
	} else {
		request.Timeout = 30 // Default
	}

	// Parse image options
	if imageOptionsMap, ok := inputMap["image_options"].(map[string]interface{}); ok {
		request.ImageOptions = t.parseImageOptions(imageOptionsMap)
	}

	return request, nil
}

// parseImageOptions parses image processing options
func (t *ArticleImporterTool) parseImageOptions(options map[string]interface{}) *s3.ImageProcessingOptions {
	imageOptions := s3.DefaultImageProcessingOptions()

	if enabled, ok := options["enabled"].(bool); ok {
		imageOptions.Enabled = enabled
	}

	if maxWidth, ok := options["max_width"].(int); ok {
		imageOptions.MaxWidth = maxWidth
	}

	if maxHeight, ok := options["max_height"].(int); ok {
		imageOptions.MaxHeight = maxHeight
	}

	if quality, ok := options["quality"].(int); ok {
		imageOptions.Quality = quality
	}

	if maxSizeBytes, ok := options["max_size_bytes"].(int); ok {
		imageOptions.MaxSizeBytes = int64(maxSizeBytes)
	}

	if format, ok := options["format"].(string); ok {
		imageOptions.Format = format
	}

	return imageOptions
}

// extractS3Config extracts S3 configuration from tool config
func (t *ArticleImporterTool) extractS3Config(config map[string]interface{}) *s3.S3ClientConfig {
	s3Config := s3.DefaultS3ClientConfig()

	// Extract from tool config first
	if customParams, ok := config["custom_params"].(map[string]interface{}); ok {
		if bucket, ok := customParams["bucket"].(string); ok {
			s3Config.Bucket = bucket
		}
		if region, ok := customParams["region"].(string); ok {
			s3Config.Region = region
		}
	}

	// Extract from global S3 config if available
	if s3ConfigMap, ok := config["s3"].(map[string]interface{}); ok {
		if url, ok := s3ConfigMap["url"].(string); ok {
			s3Config.URL = url
		}
		if region, ok := s3ConfigMap["region"].(string); ok {
			s3Config.Region = region
		}
		if bucket, ok := s3ConfigMap["bucket"].(string); ok {
			s3Config.Bucket = bucket
		}
		if endpoint, ok := s3ConfigMap["endpoint"].(string); ok {
			s3Config.Endpoint = endpoint
		}
		if useSSL, ok := s3ConfigMap["use_ssl"].(bool); ok {
			s3Config.UseSSL = useSSL
		}
		if accessKey, ok := s3ConfigMap["access_key"].(string); ok {
			s3Config.AccessKey = accessKey
		}
		if secretKey, ok := s3ConfigMap["secret_key"].(string); ok {
			s3Config.SecretKey = secretKey
		}
		if timeout, ok := s3ConfigMap["timeout"].(int); ok {
			s3Config.Timeout = timeout
		}
		if maxRetries, ok := s3ConfigMap["max_retries"].(int); ok {
			s3Config.MaxRetries = maxRetries
		}
	}

	return s3Config
}

// ImportRequest represents the internal request structure
type ImportRequest struct {
	ArticleID     string                     `json:"article_id"`
	IncludeImages bool                       `json:"include_images"`
	MaxImages     int                        `json:"max_images"`
	Timeout       int                        `json:"timeout"`
	ImageOptions  *s3.ImageProcessingOptions `json:"image_options"`
}

// Validate implements custom validation for the article importer
func (t *ArticleImporterTool) Validate(input interface{}) error {
	// Use base validation first
	if err := t.PonchoBaseTool.Validate(input); err != nil {
		return err
	}

	request, err := t.parseInput(input)
	if err != nil {
		return err
	}

	// Custom validation rules
	if request.ArticleID == "" {
		return fmt.Errorf("article_id is required")
	}

	if len(request.ArticleID) > 50 {
		return fmt.Errorf("article_id too long (max 50 characters)")
	}

	if request.MaxImages < 1 || request.MaxImages > 50 {
		return fmt.Errorf("max_images must be between 1 and 50")
	}

	if request.Timeout < 5 || request.Timeout > 300 {
		return fmt.Errorf("timeout must be between 5 and 300 seconds")
	}

	// Validate image options
	if request.ImageOptions != nil {
		if request.ImageOptions.MaxWidth < 100 || request.ImageOptions.MaxWidth > 2048 {
			return fmt.Errorf("image max_width must be between 100 and 2048")
		}
		if request.ImageOptions.MaxHeight < 100 || request.ImageOptions.MaxHeight > 2048 {
			return fmt.Errorf("image max_height must be between 100 and 2048")
		}
		if request.ImageOptions.Quality < 1 || request.ImageOptions.Quality > 100 {
			return fmt.Errorf("image quality must be between 1 and 100")
		}
		if request.ImageOptions.MaxSizeBytes < 1000 || request.ImageOptions.MaxSizeBytes > 1048576 {
			return fmt.Errorf("image max_size_bytes must be between 1000 and 1048576")
		}
	}

	return nil
}
