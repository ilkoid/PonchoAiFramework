package s3

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

)

// S3Client provides S3-compatible storage client with image processing
type S3Client struct {
	config     *ClientConfig
	httpClient *http.Client
	logger     Logger
}

// NewS3Client creates a new S3 client instance
func NewS3Client(config *ClientConfig, logger Logger) (*S3Client, error) {
	if config == nil {
		config = DefaultClientConfig()
	}

	if logger == nil {
		logger = NewDefaultLogger()
	}

	// Validate configuration
	if err := validateS3Config(config); err != nil {
		return nil, fmt.Errorf("invalid S3 config: %w", err)
	}

	// Create HTTP client with timeout
	timeout := time.Duration(config.Timeout) * time.Second
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	client := &S3Client{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}

	logger.Info("S3 client created",
		"bucket", config.Bucket,
		"region", config.Region,
		"endpoint", config.Endpoint,
		"use_ssl", config.UseSSL,
	)

	return client, nil
}

// validateS3Config validates S3 configuration
func validateS3Config(config *ClientConfig) error {
	// Bucket is not required if custom URL is provided
	if config.Bucket == "" && config.URL == "" {
		return fmt.Errorf("bucket name is required when no custom URL is provided")
	}
	if config.AccessKey == "" {
		return fmt.Errorf("access key is required")
	}
	if config.SecretKey == "" {
		return fmt.Errorf("secret key is required")
	}
	if config.Region == "" {
		return fmt.Errorf("region is required")
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 // Default to 30 seconds
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 3 // Default to 3 retries
	}
	return nil
}

// DownloadArticle downloads complete article data including JSON and images
func (c *S3Client) DownloadArticle(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error) {
	startTime := time.Now()

	// Set defaults
	if req.ImageOptions == nil {
		req.ImageOptions = DefaultImageProcessingOptions()
	}
	if req.MaxImages <= 0 {
		req.MaxImages = 10
	}
	if req.Timeout <= 0 {
		req.Timeout = c.config.Timeout
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
	defer cancel()

	requestID := generateRequestID()

	c.logger.Info("Starting article download",
		"article_id", req.ArticleID,
		"request_id", requestID,
		"include_images", req.IncludeImages,
		"max_images", req.MaxImages,
	)

	// Initialize response
	response := &DownloadResponse{
		Success: false,
		Metadata: &ResponseMetadata{
			RequestID: requestID,
			Timestamp: startTime,
			Region:    c.config.Region,
			Bucket:    c.config.Bucket,
		},
	}

	// Download JSON data
	jsonData, err := c.downloadJSON(timeoutCtx, req.ArticleID)
	if err != nil {
		response.Error = &Error{
			Code:      "DOWNLOAD_JSON_FAILED",
			Message:   fmt.Sprintf("Failed to download JSON for article %s", req.ArticleID),
			Details:   err.Error(),
			Retryable: isRetryableError(err),
		}
		response.Metadata.Duration = time.Since(startTime).Milliseconds()
		return response, err
	}

	article := &ArticleData{
		ArticleID: req.ArticleID,
		JSONData:  jsonData,
		Images:    []*Image{},
		Metadata: &ArticleMetadata{
			DownloadTime: startTime,
			Bucket:       c.config.Bucket,
			Region:       c.config.Region,
		},
	}

	// Download images if requested
	if req.IncludeImages {
		images, err := c.downloadImages(timeoutCtx, req.ArticleID, req.ImageOptions, req.MaxImages)
		if err != nil {
			response.Error = &Error{
				Code:      "DOWNLOAD_IMAGES_FAILED",
				Message:   fmt.Sprintf("Failed to download images for article %s", req.ArticleID),
				Details:   err.Error(),
				Retryable: isRetryableError(err),
			}
			response.Metadata.Duration = time.Since(startTime).Milliseconds()
			return response, err
		}
		article.Images = images
		article.Metadata.ImageCount = len(images)
	}

	// Calculate total sizes
	totalSize := int64(len(jsonData))
	processedSize := totalSize
	for _, img := range article.Images {
		totalSize += img.Size
		processedSize += int64(len(img.Data))
	}
	article.Metadata.TotalSize = totalSize
	article.Metadata.ProcessedSize = processedSize
	article.Metadata.ProcessingTime = time.Since(startTime).Milliseconds()

	response.Success = true
	response.Article = article
	response.Metadata.Duration = time.Since(startTime).Milliseconds()

	c.logger.Info("Article download completed",
		"article_id", req.ArticleID,
		"request_id", requestID,
		"duration_ms", response.Metadata.Duration,
		"image_count", len(article.Images),
		"total_size", totalSize,
	)

	return response, nil
}

// downloadJSON downloads JSON data for an article
func (c *S3Client) downloadJSON(ctx context.Context, articleID string) (string, error) {
	jsonKey := fmt.Sprintf("%s/%s.json", articleID, articleID)
	url := c.buildObjectURL(jsonKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(data), nil
}

// downloadImages downloads and processes images for an article
func (c *S3Client) downloadImages(ctx context.Context, articleID string, options *ImageProcessingOptions, maxImages int) ([]*Image, error) {
	// List images in the article folder
	imagesFolder := fmt.Sprintf("%s/images/", articleID)
	url := c.buildListURL(imagesFolder)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list images, HTTP %d", resp.StatusCode)
	}

	// Parse response to get image list (this is simplified - in real implementation you'd parse XML)
	// For now, we'll assume we get a list of image URLs
	var imageNames []string
	// TODO: Parse actual S3 list response XML to get image names

	// For demo purposes, we'll try common image names
	commonNames := []string{
		fmt.Sprintf("%s_Detailed_color_sketch.png", articleID),
		fmt.Sprintf("%s_Technical_drawing.png", articleID),
		fmt.Sprintf("%s_Photo_1.jpg", articleID),
		fmt.Sprintf("%s_Photo_2.jpg", articleID),
	}

	for i, name := range commonNames {
		if i >= maxImages {
			break
		}
		imageNames = append(imageNames, name)
	}

	// Download and process each image
	var images []*Image
	for i, imageName := range imageNames {
		if i >= maxImages {
			break
		}

		image, err := c.downloadAndProcessImage(ctx, articleID, imageName, options)
		if err != nil {
			c.logger.Warn("Failed to download image",
				"article_id", articleID,
				"image_name", imageName,
				"error", err.Error(),
			)
			continue // Continue with other images
		}

		images = append(images, image)
	}

	return images, nil
}

// downloadAndProcessImage downloads a single image and applies processing
func (c *S3Client) downloadAndProcessImage(ctx context.Context, articleID, imageName string, options *ImageProcessingOptions) (*Image, error) {
	imageKey := fmt.Sprintf("%s/images/%s", articleID, imageName)
	url := c.buildObjectURL(imageKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create image request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image, HTTP %d", resp.StatusCode)
	}

	// Read image data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Process image if enabled
	processedData := data
	width, height := 0, 0
	resized := false

	if options.Enabled {
		processedData, width, height, err = c.processImage(data, options)
		if err != nil {
			c.logger.Warn("Image processing failed, using original",
				"article_id", articleID,
				"image_name", imageName,
				"error", err.Error(),
			)
			processedData = data // Use original data
		} else {
			resized = true
		}
	}

	// Encode to base64
	contentType := getContentType(imageName)
	base64Data := base64.StdEncoding.EncodeToString(processedData)

	return &Image{
		Filename:    imageName,
		Data:        base64Data,
		ContentType: contentType,
		Size:        int64(len(data)),
		Width:       width,
		Height:      height,
		URL:         url,
		Resized:     resized,
	}, nil
}

// processImage resizes and optimizes an image
func (c *S3Client) processImage(data []byte, options *ImageProcessingOptions) ([]byte, int, int, error) {
	// Decode image
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Calculate new dimensions if resizing is needed
	newWidth, newHeight := originalWidth, originalHeight
	needsResize := false

	if originalWidth > options.MaxWidth || originalHeight > options.MaxHeight {
		// Calculate aspect ratio preserving dimensions
		ratio := float64(originalWidth) / float64(originalHeight)

		if originalWidth > originalHeight {
			newWidth = options.MaxWidth
			newHeight = int(float64(options.MaxWidth) / ratio)
		} else {
			newHeight = options.MaxHeight
			newWidth = int(float64(options.MaxHeight) * ratio)
		}
		needsResize = true
	}

	// Resize if needed
	if needsResize {
		// Use simple nearest neighbor scaling for now
		// In production, you might want to use better scaling algorithms
		newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

		// Simple scaling (not the best quality, but functional)
		for y := 0; y < newHeight; y++ {
			for x := 0; x < newWidth; x++ {
				srcX := x * originalWidth / newWidth
				srcY := y * originalHeight / newHeight
				newImg.Set(x, y, img.At(srcX, srcY))
			}
		}
		img = newImg
	}

	// Encode to target format
	var buf bytes.Buffer
	targetFormat := options.Format
	if targetFormat == "" {
		targetFormat = format
	}

	switch targetFormat {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: options.Quality})
	case "png":
		err = png.Encode(&buf, img)
	default:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: options.Quality})
		targetFormat = "jpeg"
	}

	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to encode image: %w", err)
	}

	// Check size limit
	processedData := buf.Bytes()
	if len(processedData) > int(options.MaxSizeBytes) {
		// Try with lower quality
		for quality := options.Quality - 10; quality > 10; quality -= 10 {
			buf.Reset()
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
			if err != nil {
				continue
			}
			if buf.Len() <= int(options.MaxSizeBytes) {
				processedData = buf.Bytes()
				break
			}
		}
	}

	return processedData, newWidth, newHeight, nil
}

// ListArticles lists available articles in the S3 bucket
func (c *S3Client) ListArticles(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	startTime := time.Now()
	requestID := generateRequestID()

	// Set defaults
	if req.Bucket == "" {
		req.Bucket = c.config.Bucket
	}
	if req.Region == "" {
		req.Region = c.config.Region
	}

	response := &ListResponse{
		Success: false,
		Metadata: &ResponseMetadata{
			RequestID: requestID,
			Timestamp: startTime,
			Region:    req.Region,
			Bucket:    req.Bucket,
		},
	}

	// Build list URL
	prefix := req.Prefix
	if prefix == "" {
		prefix = "" // List all
	}

	url := c.buildListURL(prefix)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		response.Error = &Error{
			Code:      "CREATE_REQUEST_FAILED",
			Message:   "Failed to create list request",
			Details:   err.Error(),
			Retryable: false,
		}
		response.Metadata.Duration = time.Since(startTime).Milliseconds()
		return response, err
	}

	c.setAuthHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		response.Error = &Error{
			Code:      "LIST_REQUEST_FAILED",
			Message:   "Failed to execute list request",
			Details:   err.Error(),
			Retryable: true,
		}
		response.Metadata.Duration = time.Since(startTime).Milliseconds()
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		response.Error = &Error{
			Code:      "LIST_HTTP_ERROR",
			Message:   fmt.Sprintf("List request failed with HTTP %d", resp.StatusCode),
			Details:   string(body),
			Retryable: resp.StatusCode >= 500,
		}
		response.Metadata.Duration = time.Since(startTime).Milliseconds()
		return response, fmt.Errorf("list request failed: HTTP %d", resp.StatusCode)
	}

	// Parse response (simplified - in real implementation you'd parse XML)
	// For now, return empty list
	var articles []string

	response.Success = true
	response.Articles = articles
	response.Metadata.Duration = time.Since(startTime).Milliseconds()

	return response, nil
}

// Helper methods

func (c *S3Client) buildObjectURL(key string) string {
	// If custom URL is provided, use it directly
	if c.config.URL != "" {
		// For Yandex Cloud, use format: https://endpoint/bucket/key
		if strings.Contains(c.config.URL, "yandexcloud.net") {
			return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(c.config.URL, "/"), c.config.Bucket, key)
		}
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(c.config.URL, "/"), key)
	}

	scheme := "https"
	if !c.config.UseSSL {
		scheme = "http"
	}

	// For Yandex Cloud, use virtual-hosted style or path style
	if strings.Contains(c.config.Endpoint, "yandexcloud.net") {
		// Use path style for Yandex Cloud: https://endpoint/bucket/key
		return fmt.Sprintf("%s://%s/%s/%s", scheme, c.config.Endpoint, c.config.Bucket, key)
	}

	// Default virtual-hosted style: https://bucket.endpoint/key
	return fmt.Sprintf("%s://%s.%s/%s", scheme, c.config.Bucket, c.config.Endpoint, key)
}

func (c *S3Client) buildListURL(prefix string) string {
	var baseURL string

	// If custom URL is provided, use it directly
	if c.config.URL != "" {
		// For Yandex Cloud, use format: https://endpoint/bucket
		if strings.Contains(c.config.URL, "yandexcloud.net") {
			baseURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(c.config.URL, "/"), c.config.Bucket)
		} else {
			baseURL = strings.TrimSuffix(c.config.URL, "/")
		}
	} else {
		scheme := "https"
		if !c.config.UseSSL {
			scheme = "http"
		}

		// For Yandex Cloud, use path style: https://endpoint/bucket
		if strings.Contains(c.config.Endpoint, "yandexcloud.net") {
			baseURL = fmt.Sprintf("%s://%s/%s", scheme, c.config.Endpoint, c.config.Bucket)
		} else {
			// Default virtual-hosted style
			baseURL = fmt.Sprintf("%s://%s.%s", scheme, c.config.Bucket, c.config.Endpoint)
		}
	}

	if prefix != "" {
		return fmt.Sprintf("%s?list-type=2&prefix=%s", baseURL, url.QueryEscape(prefix))
	}
	return fmt.Sprintf("%s?list-type=2", baseURL)
}

func (c *S3Client) setAuthHeaders(req *http.Request) {
	// Use AWS Signature Version 4 for Yandex Cloud compatibility
	if err := c.setAuthHeadersV4(req); err != nil {
		// Fallback to simple auth for development/testing
		c.logger.Warn("Failed to set AWS V4 signature, using fallback", "error", err.Error())
		req.Header.Set("Authorization", fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s", c.config.AccessKey))
		req.Header.Set("Content-Type", "application/json")
	}
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func isRetryableError(err error) bool {
	// Simple heuristic for retryable errors
	errStr := strings.ToLower(err.Error())
	retryable := []string{
		"timeout", "connection", "network", "temporary", "rate limit",
	}

	for _, keyword := range retryable {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
}
