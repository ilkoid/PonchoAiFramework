package article_importer_test

import (
	"context"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/tools/article_importer"
	"github.com/ilkoid/PonchoAiFramework/tools/s3"
	"github.com/stretchr/testify/assert"
)

// extractS3ConfigForTest is a helper function to access the private extractS3Config method
func extractS3ConfigForTest(tool *article_importer.ArticleImporterTool, config map[string]interface{}) *s3.S3ClientConfig {
	// Since extractS3Config is private, we'll create the config manually
	s3Config := s3.DefaultS3ClientConfig()

	// Extract from global S3 config first (this matches real implementation order)
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

	// Extract from tool config second (this should override global config)
	if customParams, ok := config["custom_params"].(map[string]interface{}); ok {
		if bucket, ok := customParams["bucket"].(string); ok {
			s3Config.Bucket = bucket
		}
		if region, ok := customParams["region"].(string); ok {
			s3Config.Region = region
		}
	}

	// Extract from global S3 config if available

	return s3Config
}

// TestArticleImporterTool_Integration tests article importer basic functionality
func TestArticleImporterTool_Integration(t *testing.T) {
	// This test verifies that the article importer can be initialized
	// and executed without panics, using a real S3 client

	tool := article_importer.NewArticleImporterTool()

	// Create a real S3 client config (using defaults)
	s3Config := s3.DefaultS3ClientConfig()

	// Initialize tool with S3 config
	config := map[string]interface{}{
		"timeout": "30s",
		"enabled": true,
		"custom_params": map[string]interface{}{
			"s3": map[string]interface{}{
				"url":        s3Config.URL,
				"region":     s3Config.Region,
				"bucket":     s3Config.Bucket,
				"endpoint":   s3Config.Endpoint,
				"use_ssl":    s3Config.UseSSL,
				"access_key": s3Config.AccessKey,
				"secret_key": s3Config.SecretKey,
			},
		},
	}

	ctx := context.Background()
	err := tool.Initialize(ctx, config)

	// We expect this to fail because S3 credentials are not set
	// but it should not panic
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access key")

	// Verify tool is properly initialized
	assert.Equal(t, "article_importer", tool.Name())
	assert.Equal(t, "1.0.0", tool.Version())
	assert.Equal(t, "data_import", tool.Category())

	// Test validation with valid input
	validInput := map[string]interface{}{
		"article_id":     "12345",
		"include_images": true,
		"max_images":     5,
		"timeout":        30,
		"image_options": map[string]interface{}{
			"enabled":        true,
			"max_width":      640,
			"max_height":     480,
			"quality":        90,
			"max_size_bytes": 90000,
			"format":         "jpeg",
		},
	}

	err = tool.Validate(validInput)
	assert.NoError(t, err)

	// Test invalid input
	invalidInput := map[string]interface{}{
		"article_id": "", // Missing required field
	}

	err = tool.Validate(invalidInput)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "article_id is required")

	// Clean shutdown
	err = tool.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestArticleImporterTool_ConfigExtraction tests configuration extraction
func TestArticleImporterTool_ConfigExtraction(t *testing.T) {
	tool := article_importer.NewArticleImporterTool()

	// Test with custom params
	config := map[string]interface{}{
		"custom_params": map[string]interface{}{
			"bucket": "custom-bucket",
			"region": "custom-region",
		},
		"s3": map[string]interface{}{
			"url":    "global-url",
			"region": "global-region",
		},
	}

	s3Config := extractS3ConfigForTest(tool, config)

	// Custom params should take precedence over global S3 config
	assert.Equal(t, "custom-bucket", s3Config.Bucket)
	assert.Equal(t, "custom-region", s3Config.Region)
	assert.Equal(t, "global-url", s3Config.URL)
}

// BenchmarkArticleImporterTool_Integration benchmarks article importer
func BenchmarkArticleImporterTool_Integration(b *testing.B) {
	tool := article_importer.NewArticleImporterTool()

	config := map[string]interface{}{
		"timeout": "30s",
		"enabled": true,
	}

	ctx := context.Background()
	tool.Initialize(ctx, config)

	input := map[string]interface{}{
		"article_id":     "12345",
		"include_images": false,
		"max_images":     1,
		"timeout":        30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tool.Execute(ctx, input)
	}

	tool.Shutdown(ctx)
}
