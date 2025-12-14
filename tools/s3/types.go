package s3

import (
	"time"
)

// S3Client defines the interface for S3-compatible storage clients

// DownloadRequest represents a request to download article data
type DownloadRequest struct {
	ArticleID     string                  `json:"article_id"`
	Bucket        string                  `json:"bucket,omitempty"`
	Region        string                  `json:"region,omitempty"`
	IncludeImages bool                    `json:"include_images"`
	ImageOptions  *ImageProcessingOptions `json:"image_options,omitempty"`
	MaxImages     int                     `json:"max_images,omitempty"`
	Timeout       int                     `json:"timeout,omitempty"` // Timeout in seconds
}

// DownloadResponse represents the response from downloading article data
type DownloadResponse struct {
	Success  bool              `json:"success"`
	Article  *ArticleData      `json:"article,omitempty"`
	Error    *Error            `json:"error,omitempty"`
	Metadata *ResponseMetadata `json:"metadata"`
}

// ListRequest represents a request to list articles
type ListRequest struct {
	Bucket   string `json:"bucket,omitempty"`
	Region   string `json:"region,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	MaxItems int    `json:"max_items,omitempty"`
}

// ListResponse represents the response from listing articles
type ListResponse struct {
	Success  bool              `json:"success"`
	Articles []string          `json:"articles,omitempty"` // List of article IDs
	Error    *Error            `json:"error,omitempty"`
	Metadata *ResponseMetadata `json:"metadata"`
}

// ArticleData represents the complete data structure for an article from S3
type ArticleData struct {
	ArticleID string         `json:"article_id"`
	JSONData  string         `json:"json_data"` // Raw JSON string from S3
	Images    []*Image       `json:"images"`
	Metadata  *ArticleMetadata `json:"metadata"`
}

// Image represents an image downloaded from S3
type Image struct {
	Filename    string `json:"filename"`
	Data        string `json:"data"` // Base64 encoded image data
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	URL         string `json:"url"`     // Original S3 URL
	Resized     bool   `json:"resized"` // Whether image was resized
}

// ArticleMetadata represents metadata about the article
type ArticleMetadata struct {
	DownloadTime   time.Time `json:"download_time"`
	TotalSize      int64     `json:"total_size"`
	ImageCount     int       `json:"image_count"`
	ProcessedSize  int64     `json:"processed_size"` // Size after optimization
	Bucket         string    `json:"bucket"`
	Region         string    `json:"region"`
	ProcessingTime int64     `json:"processing_time_ms"`
}

// ResponseMetadata represents metadata about the response
type ResponseMetadata struct {
	RequestID  string    `json:"request_id"`
	Timestamp  time.Time `json:"timestamp"`
	Duration   int64     `json:"duration_ms"`
	Region     string    `json:"region"`
	Bucket     string    `json:"bucket"`
	RetryCount int       `json:"retry_count"`
}

// Error represents an S3-specific error
type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Retryable bool   `json:"retryable"`
}

// ImageProcessingOptions represents options for image processing
type ImageProcessingOptions struct {
	Enabled          bool   `json:"enabled"`
	MaxWidth         int    `json:"max_width"`
	MaxHeight        int    `json:"max_height"`
	Quality          int    `json:"quality"`        // 1-100
	MaxSizeBytes     int64  `json:"max_size_bytes"` // Maximum file size in bytes
	Format           string `json:"format"`         // jpeg, png, webp
	PreserveMetadata bool   `json:"preserve_metadata"`
}

// ClientConfig represents configuration for S3 client
type ClientConfig struct {
	URL        string `json:"url"`
	Region     string `json:"region"`
	Bucket     string `json:"bucket"`
	Endpoint   string `json:"endpoint,omitempty"`
	UseSSL     bool   `json:"use_ssl"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	Timeout    int    `json:"timeout"` // Timeout in seconds
	MaxRetries int    `json:"max_retries"`
}