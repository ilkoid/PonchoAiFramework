package media

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// Base64Converter converts between bytes and base64 format
type Base64Converter struct {
	logger interfaces.Logger
}

func NewBase64Converter() *Base64Converter {
	return &Base64Converter{
		logger: interfaces.NewDefaultLogger(),
	}
}

func (bc *Base64Converter) Name() string {
	return "base64"
}

func (bc *Base64Converter) SupportedFormats() []MediaFormat {
	return []MediaFormat{MediaFormatBase64, MediaFormatBytes}
}

func (bc *Base64Converter) CanConvert(source *MediaData, targetFormat MediaFormat) bool {
	if targetFormat != MediaFormatBase64 {
		return false
	}

	// Can convert if we have bytes or a URL
	return len(source.Bytes) > 0 || source.URL != ""
}

func (bc *Base64Converter) Convert(ctx context.Context, source *MediaData, targetFormat MediaFormat) (*MediaData, error) {
	if targetFormat != MediaFormatBase64 {
		return nil, fmt.Errorf("base64 converter only supports base64 target format")
	}

	result := source.Clone()
	result.Format = MediaFormatBase64

	// If we already have bytes, just mark as base64
	if len(source.Bytes) > 0 {
		return result, nil
	}

	// If we have a URL, need to download first
	if source.URL != "" {
		bc.logger.Debug("Downloading media for base64 conversion", "url", source.URL)

		downloader := NewDownloader()
		downloaded, err := downloader.Convert(ctx, source, MediaFormatBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to download for base64 conversion: %w", err)
		}

		result.Bytes = downloaded.Bytes
		result.Size = downloaded.Size
		result.MimeType = downloaded.MimeType
	}

	return result, nil
}

// URLConverter handles URL-based media
type URLConverter struct {
	logger interfaces.Logger
}

func NewURLConverter() *URLConverter {
	return &URLConverter{
		logger: interfaces.NewDefaultLogger(),
	}
}

func (uc *URLConverter) Name() string {
	return "url"
}

func (uc *URLConverter) SupportedFormats() []MediaFormat {
	return []MediaFormat{MediaFormatURL}
}

func (uc *URLConverter) CanConvert(source *MediaData, targetFormat MediaFormat) bool {
	if targetFormat != MediaFormatURL {
		return false
	}

	// Can create URL if we have bytes (upload) or already have URL
	return len(source.Bytes) > 0 || source.URL != ""
}

func (uc *URLConverter) Convert(ctx context.Context, source *MediaData, targetFormat MediaFormat) (*MediaData, error) {
	if targetFormat != MediaFormatURL {
		return nil, fmt.Errorf("URL converter only supports URL target format")
	}

	result := source.Clone()
	result.Format = MediaFormatURL

	// If we already have a URL, just use it
	if source.URL != "" {
		return result, nil
	}

	// If we have bytes, in a real implementation we would upload to storage
	// For now, create a data URL
	if len(source.Bytes) > 0 {
		encoded := base64.StdEncoding.EncodeToString(source.Bytes)
		result.URL = fmt.Sprintf("data:%s;base64,%s", source.MimeType, encoded)
		uc.logger.Debug("Created data URL for bytes", "mime_type", source.MimeType)
	}

	return result, nil
}

// ImageConverter handles image format conversions
type ImageConverter struct {
	logger interfaces.Logger
}

func NewImageConverter() *ImageConverter {
	return &ImageConverter{
		logger: interfaces.NewDefaultLogger(),
	}
}

func (ic *ImageConverter) Name() string {
	return "image"
}

func (ic *ImageConverter) SupportedFormats() []MediaFormat {
	return []MediaFormat{MediaFormatBytes, MediaFormatBase64}
}

func (ic *ImageConverter) CanConvert(source *MediaData, targetFormat MediaFormat) bool {
	// Only works with images
	if !source.IsImage() {
		return false
	}

	return targetFormat == MediaFormatBytes || targetFormat == MediaFormatBase64
}

func (ic *ImageConverter) Convert(ctx context.Context, source *MediaData, targetFormat MediaFormat) (*MediaData, error) {
	if !source.IsImage() {
		return nil, fmt.Errorf("source is not an image: %s", source.MimeType)
	}

	// For base64, delegate to Base64Converter
	if targetFormat == MediaFormatBase64 {
		converter := NewBase64Converter()
		return converter.Convert(ctx, source, targetFormat)
	}

	// For bytes, ensure we have valid image data
	result := source.Clone()
	result.Format = MediaFormatBytes

	// If we need to download first
	if len(source.Bytes) == 0 && source.URL != "" {
		downloader := NewDownloader()
		downloaded, err := downloader.Convert(ctx, source, MediaFormatBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to download image: %w", err)
		}
		result.Bytes = downloaded.Bytes
		result.Size = downloaded.Size
		result.MimeType = downloaded.MimeType
	}

	// Validate image data
	if len(result.Bytes) > 0 {
		if err := ic.validateImageData(result.Bytes); err != nil {
			return nil, fmt.Errorf("invalid image data: %w", err)
		}
		ic.logger.Debug("Image validated successfully", "mime_type", result.MimeType)
	}

	return result, nil
}

func (ic *ImageConverter) validateImageData(data []byte) error {
	// Simple validation - check if it's a valid image format
	// In a real implementation, you might use image.Decode to validate
	if len(data) < 8 {
		return fmt.Errorf("image data too small")
	}

	// Check common image signatures
	signatures := map[string][]byte{
		"image/jpeg": {0xFF, 0xD8, 0xFF},
		"image/png":  {0x89, 0x50, 0x4E, 0x47},
		"image/gif":  {0x47, 0x49, 0x46, 0x38},
		"image/webp": {0x52, 0x49, 0x46, 0x46},
	}

	for _, sig := range signatures {
		if len(data) >= len(sig) && matchesSignature(data[:len(sig)], sig) {
			return nil
		}
	}

	return fmt.Errorf("unrecognized image format")
}

func matchesSignature(data, signature []byte) bool {
	for i, b := range signature {
		if i >= len(data) || data[i] != b {
			return false
		}
	}
	return true
}

// Downloader handles downloading media from URLs
type Downloader struct {
	client *http.Client
	logger interfaces.Logger
}

func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: interfaces.NewDefaultLogger(),
	}
}

func (d *Downloader) Name() string {
	return "downloader"
}

func (d *Downloader) SupportedFormats() []MediaFormat {
	return []MediaFormat{MediaFormatBytes}
}

func (d *Downloader) CanConvert(source *MediaData, targetFormat MediaFormat) bool {
	if targetFormat != MediaFormatBytes {
		return false
	}

	return source.URL != ""
}

func (d *Downloader) Convert(ctx context.Context, source *MediaData, targetFormat MediaFormat) (*MediaData, error) {
	if source.URL == "" {
		return nil, fmt.Errorf("no URL provided for download")
	}

	d.logger.Debug("Downloading media", "url", source.URL)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Perform request
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Detect content type if not provided
	contentType := source.MimeType
	if contentType == "" {
		contentType = resp.Header.Get("Content-Type")
		if contentType == "" {
			// Try to detect from data
			contentType = http.DetectContentType(data)
		}
	}

	// Create result
	result := source.Clone()
	result.Bytes = data
	result.Size = int64(len(data))
	result.Format = MediaFormatBytes
	result.MimeType = contentType

	d.logger.Debug("Media downloaded successfully",
		"url", source.URL,
		"size", result.Size,
		"mime_type", result.MimeType,
	)

	return result, nil
}

// InMemoryCache provides a simple in-memory cache implementation
type InMemoryCache struct {
	data map[string]*cacheEntry
	mutex sync.RWMutex
}

type cacheEntry struct {
	media   *MediaData
	expires time.Time
}

func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		data: make(map[string]*cacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

func (c *InMemoryCache) Get(key string) (*MediaData, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.data[key]
	if !exists || time.Now().After(entry.expires) {
		return nil, false
	}

	return entry.media.Clone(), true
}

func (c *InMemoryCache) Set(key string, media *MediaData, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = &cacheEntry{
		media:   media.Clone(),
		expires: time.Now().Add(ttl),
	}

	return nil
}

func (c *InMemoryCache) Delete(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.data[key]; exists {
		delete(c.data, key)
		return true
	}
	return false
}

func (c *InMemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]*cacheEntry)
}

func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.expires) {
				delete(c.data, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Helper functions for common operations

// IsDataURL checks if a string is a data URL
func IsDataURL(s string) bool {
	return strings.HasPrefix(s, "data:")
}

// ExtractDataURL extracts the media type and data from a data URL
func ExtractDataURL(dataURL string) (mimeType string, data []byte, err error) {
	if !IsDataURL(dataURL) {
		return "", nil, fmt.Errorf("not a data URL")
	}

	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid data URL format")
	}

	// Extract mime type
	header := parts[0]
	if !strings.HasPrefix(header, "data:") {
		return "", nil, fmt.Errorf("invalid data URL header")
	}
	header = header[5:] // Remove "data:"

	if idx := strings.Index(header, ";"); idx != -1 {
		mimeType = header[:idx]
	} else {
		mimeType = header
	}

	// Decode data
	data, err = base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	return mimeType, data, nil
}