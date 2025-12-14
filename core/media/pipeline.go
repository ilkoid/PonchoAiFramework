package media

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MediaType represents different media types
type MediaType string

const (
	MediaTypeJPEG  MediaType = "image/jpeg"
	MediaTypePNG   MediaType = "image/png"
	MediaTypeWEBP  MediaType = "image/webp"
	MediaTypeGIF   MediaType = "image/gif"
	MediaTypeURL   MediaType = "url"
	MediaTypeBytes MediaType = "bytes"
)

// MediaFormat represents how media is stored/transmitted
type MediaFormat string

const (
	MediaFormatURL    MediaFormat = "url"
	MediaFormatBase64 MediaFormat = "base64"
	MediaFormatBytes  MediaFormat = "bytes"
)

// MediaConverter handles conversion between different media formats
type MediaConverter interface {
	CanConvert(source *MediaData, targetFormat MediaFormat) bool
	Convert(ctx context.Context, source *MediaData, targetFormat MediaFormat) (*MediaData, error)
	SupportedFormats() []MediaFormat
	Name() string
}

// MediaPipeline provides automatic media conversion and processing
type MediaPipeline struct {
	converters  map[string]MediaConverter
	cache       MediaCache
	logger      interfaces.Logger
	maxSize     int64
	timeout     time.Duration
	concurrency int
	mutex       sync.RWMutex
}

// MediaData represents media content with metadata
type MediaData struct {
	URL      string                 `json:"url,omitempty"`
	Bytes    []byte                 `json:"-"` // Don't serialize bytes
	MimeType string                 `json:"mime_type"`
	Size     int64                  `json:"size"`
	Format   MediaFormat            `json:"format"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MediaCache provides caching for processed media
type MediaCache interface {
	Get(key string) (*MediaData, bool)
	Set(key string, data *MediaData, ttl time.Duration) error
	Delete(key string) bool
	Clear()
}

// PipelineConfig provides configuration for MediaPipeline
type PipelineConfig struct {
	MaxSize     int64         // Maximum file size in bytes (default: 50MB)
	Timeout     time.Duration // Request timeout (default: 30s)
	Concurrency int           // Max concurrent conversions (default: 5)
	Cache       MediaCache    // Optional cache implementation
	Logger      interfaces.Logger
}

// DefaultPipelineConfig returns default configuration
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		MaxSize:     50 * 1024 * 1024, // 50MB
		Timeout:     30 * time.Second,
		Concurrency: 5,
		Logger:      interfaces.NewDefaultLogger(),
	}
}

// NewMediaPipeline creates a new MediaPipeline with default converters
func NewMediaPipeline(config *PipelineConfig) *MediaPipeline {
	if config == nil {
		config = DefaultPipelineConfig()
	}

	pipeline := &MediaPipeline{
		converters:  make(map[string]MediaConverter),
		cache:       config.Cache,
		logger:      config.Logger,
		maxSize:     config.MaxSize,
		timeout:     config.Timeout,
		concurrency: config.Concurrency,
	}

	// Register default converters
	pipeline.RegisterConverter(NewBase64Converter())
	pipeline.RegisterConverter(NewURLConverter())
	pipeline.RegisterConverter(NewImageConverter())
	pipeline.RegisterConverter(NewDownloader())

	return pipeline
}

// RegisterConverter adds a new converter to the pipeline
func (mp *MediaPipeline) RegisterConverter(converter MediaConverter) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.converters[converter.Name()] = converter
}

// PrepareForModel prepares media for a specific model
func (mp *MediaPipeline) PrepareForModel(ctx context.Context, mediaList []*MediaData, model interfaces.PonchoModel) ([]*MediaData, error) {
	if len(mediaList) == 0 {
		return mediaList, nil
	}

	mp.logger.Debug("Preparing media for model",
		"model", model.Name(),
		"count", len(mediaList),
		"supports_vision", model.SupportsVision(),
	)

	if !model.SupportsVision() {
		return nil, fmt.Errorf("model %s doesn't support vision", model.Name())
	}

	// Determine target format based on model capabilities
	targetFormat := mp.getPreferredFormat(model)

	// Process media with concurrency control
	return mp.batchConvert(ctx, mediaList, targetFormat)
}

// batchConvert converts multiple media items with concurrency control
func (mp *MediaPipeline) batchConvert(ctx context.Context, mediaList []*MediaData, targetFormat MediaFormat) ([]*MediaData, error) {
	if len(mediaList) == 0 {
		return mediaList, nil
	}

	// Create buffered channel for results
	results := make(chan *ConversionResult, len(mediaList))
	semaphore := make(chan struct{}, mp.concurrency)

	// Start conversion goroutines
	var wg sync.WaitGroup
	for i, media := range mediaList {
		wg.Add(1)
		go func(index int, m *MediaData) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			converted, err := mp.Convert(ctx, m, targetFormat)
			results <- &ConversionResult{
				Index:     index,
				Media:     converted,
				Error:     err,
				Timestamp: time.Now(),
			}
		}(i, media)
	}

	// Wait for all conversions to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	output := make([]*MediaData, len(mediaList))
	var errors []error

	for result := range results {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("conversion %d failed: %w", result.Index, result.Error))
			continue
		}
		output[result.Index] = result.Media
	}

	if len(errors) > 0 {
		return output, fmt.Errorf("batch conversion completed with %d errors: %v", len(errors), errors)
	}

	return output, nil
}

// Convert converts media data to target format
func (mp *MediaPipeline) Convert(ctx context.Context, source *MediaData, targetFormat MediaFormat) (*MediaData, error) {
	if source == nil {
		return nil, fmt.Errorf("source media cannot be nil")
	}

	// Check cache first
	cacheKey := mp.generateCacheKey(source, targetFormat)
	if mp.cache != nil {
		if cached, found := mp.cache.Get(cacheKey); found {
			mp.logger.Debug("Media cache hit", "cache_key", cacheKey)
			return cached, nil
		}
	}

	// Find appropriate converter
	converter := mp.findConverter(source, targetFormat)
	if converter == nil {
		return nil, fmt.Errorf("no converter available for %s -> %s", source.Format, targetFormat)
	}

	// Perform conversion
	result, err := converter.Convert(ctx, source, targetFormat)
	if err != nil {
		return nil, fmt.Errorf("conversion failed: %w", err)
	}

	// Cache result
	if mp.cache != nil && result != nil {
		if cacheErr := mp.cache.Set(cacheKey, result, time.Hour); cacheErr != nil {
			mp.logger.Warn("Failed to cache media", "error", cacheErr)
		}
	}

	return result, nil
}

// findConverter finds the best converter for the conversion
func (mp *MediaPipeline) findConverter(source *MediaData, targetFormat MediaFormat) MediaConverter {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	var bestConverter MediaConverter
	var bestScore int = -1

	for _, converter := range mp.converters {
		if converter.CanConvert(source, targetFormat) {
			score := mp.scoreConverter(converter, source, targetFormat)
			if score > bestScore {
				bestScore = score
				bestConverter = converter
			}
		}
	}

	return bestConverter
}

// scoreConverter scores converters for selection (higher is better)
func (mp *MediaPipeline) scoreConverter(converter MediaConverter, source *MediaData, targetFormat MediaFormat) int {
	score := 0

	// Base64 converter is preferred for vision models
	if converter.Name() == "base64" && targetFormat == MediaFormatBase64 {
		score += 100
	}

	// URL converter is good for web resources
	if converter.Name() == "url" && source.URL != "" {
		score += 50
	}

	// Image converter for image formats
	if converter.Name() == "image" && strings.HasPrefix(source.MimeType, "image/") {
		score += 75
	}

	return score
}

// getPreferredFormat determines the best format for a model
func (mp *MediaPipeline) getPreferredFormat(model interfaces.PonchoModel) MediaFormat {
	// Most vision models prefer base64 data URLs
	if strings.Contains(model.Name(), "vision") || strings.Contains(model.Name(), "glm") {
		return MediaFormatBase64
	}

	// Default to base64 for compatibility
	return MediaFormatBase64
}

// generateCacheKey creates a unique key for caching
func (mp *MediaPipeline) generateCacheKey(media *MediaData, targetFormat MediaFormat) string {
	var key strings.Builder

	// Include source identifier
	if media.URL != "" {
		key.WriteString("url:")
		key.WriteString(media.URL)
	} else if len(media.Bytes) > 0 {
		key.WriteString("bytes:")
		key.WriteString(fmt.Sprintf("%x", media.Bytes[:min(32, len(media.Bytes))]))
	} else {
		key.WriteString("empty")
	}

	// Include format and mime type
	key.WriteString(fmt.Sprintf("_%s_%s_%s", media.Format, media.MimeType, targetFormat))

	return key.String()
}

// ConversionResult represents the result of a media conversion
type ConversionResult struct {
	Index     int
	Media     *MediaData
	Error     error
	Timestamp time.Time
}

// NewMediaDataFromURL creates MediaData from URL
func NewMediaDataFromURL(rawURL string) (*MediaData, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(parsedURL.Path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return &MediaData{
		URL:      rawURL,
		MimeType: mimeType,
		Format:   MediaFormatURL,
		Metadata: make(map[string]interface{}),
	}, nil
}

// NewMediaDataFromBytes creates MediaData from byte slice
func NewMediaDataFromBytes(data []byte, mimeType string) (*MediaData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	return &MediaData{
		Bytes:    data,
		MimeType: mimeType,
		Size:     int64(len(data)),
		Format:   MediaFormatBytes,
		Metadata: make(map[string]interface{}),
	}, nil
}

// GetDataURL returns a data URL for the media
func (md *MediaData) GetDataURL() string {
	switch md.Format {
	case MediaFormatBase64, MediaFormatBytes:
		if len(md.Bytes) == 0 {
			return md.URL
		}
		encoded := base64.StdEncoding.EncodeToString(md.Bytes)
		return fmt.Sprintf("data:%s;base64,%s", md.MimeType, encoded)
	case MediaFormatURL:
		return md.URL
	default:
		return ""
	}
}

// Validate checks if the media data is valid
func (md *MediaData) Validate(maxSize int64) error {
	if md.URL == "" && len(md.Bytes) == 0 {
		return fmt.Errorf("media has no URL or bytes")
	}

	if md.MimeType == "" {
		return fmt.Errorf("media has no mime type")
	}

	if len(md.Bytes) > 0 && int64(len(md.Bytes)) > maxSize {
		return fmt.Errorf("media size %d exceeds maximum %d", len(md.Bytes), maxSize)
	}

	return nil
}

// IsImage checks if media is an image
func (md *MediaData) IsImage() bool {
	return strings.HasPrefix(md.MimeType, "image/")
}

// GetExtension returns the file extension for the media type
func (md *MediaData) GetExtension() string {
	exts, err := mime.ExtensionsByType(md.MimeType)
	if err != nil || len(exts) == 0 {
		return ""
	}
	return exts[0]
}

// Clone creates a copy of the media data
func (md *MediaData) Clone() *MediaData {
	clone := &MediaData{
		URL:      md.URL,
		MimeType: md.MimeType,
		Size:     md.Size,
		Format:   md.Format,
		Metadata: make(map[string]interface{}),
	}

	// Copy bytes
	if len(md.Bytes) > 0 {
		clone.Bytes = make([]byte, len(md.Bytes))
		copy(clone.Bytes, md.Bytes)
	}

	// Copy metadata
	for k, v := range md.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}