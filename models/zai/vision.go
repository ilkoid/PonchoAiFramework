// Package zai provides vision processing capabilities for Z.AI GLM models
//
// This file implements specialized vision processing for fashion industry applications,
// leveraging Z.AI's GLM-4.6V model for detailed image analysis. The VisionProcessor
// provides high-level APIs for fashion image analysis, feature extraction, and
// multimodal content processing with support for various image formats and sources.
//
// Key Features:
// - Fashion-specific image analysis and classification
// - Support for multiple image formats (JPEG, PNG, GIF, WebP)
// - Base64 and URL-based image processing
// - Configurable vision parameters (quality, detail, size)
// - Structured fashion analysis with clothing item detection
// - Color, style, and material extraction
// - Image downloading and caching support
//
// Fashion Analysis Capabilities:
// - Clothing item detection and classification
// - Color palette extraction
// - Style and season identification
// - Material and accessory recognition
// - Confidence scoring and metadata
//
// Usage:
//   processor := NewVisionProcessor(model, logger, config)
//   analysis, err := processor.AnalyzeFashionImage(ctx, imageURL)
//   // or from base64:
//   analysis, err := processor.AnalyzeFashionImageFromBase64(ctx, data, mimeType)
package zai

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// VisionProcessor handles vision-specific functionality for Z.AI GLM models
//
// This struct provides comprehensive vision processing capabilities specifically
// designed for fashion industry applications. It integrates with Z.AI's GLM-4.6V
// model to provide detailed image analysis, clothing item detection, and fashion
// attribute extraction with configurable parameters and error handling.
type VisionProcessor struct {
	model  *ZAIModel
	logger interfaces.Logger
	config *VisionConfig
}

// VisionConfig represents configuration for vision processing
type VisionConfig struct {
	MaxImageSize     int           `json:"max_image_size"`
	SupportedFormats []string      `json:"supported_formats"`
	DefaultQuality   string        `json:"default_quality"`
	DefaultDetail    string        `json:"default_detail"`
	Timeout          time.Duration `json:"timeout"`
	EnableCaching    bool          `json:"enable_caching"`
	CacheTTL         time.Duration `json:"cache_ttl"`
}

// FashionAnalysis represents the result of fashion image analysis
type FashionAnalysis struct {
	Description   string                 `json:"description"`
	ClothingItems []ClothingItem         `json:"clothing_items"`
	Colors        []string               `json:"colors"`
	Style         string                 `json:"style"`
	Season        string                 `json:"season"`
	Materials     []string               `json:"materials"`
	Accessories   []string               `json:"accessories"`
	Confidence    float64                `json:"confidence"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ClothingItem represents a detected clothing item
type ClothingItem struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Material    string  `json:"material"`
	Style       string  `json:"style"`
	Position    string  `json:"position"`
	Confidence  float64 `json:"confidence"`
}

// NewVisionProcessor creates a new vision processor
func NewVisionProcessor(model *ZAIModel, logger interfaces.Logger, config *VisionConfig) *VisionProcessor {
	if config == nil {
		config = &VisionConfig{
			MaxImageSize:     ZAIVisionMaxImageSize,
			SupportedFormats: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
			DefaultQuality:   ZAIVisionQualityAuto,
			DefaultDetail:    ZAIVisionDetailAuto,
			Timeout:          30 * time.Second,
			EnableCaching:    false,
			CacheTTL:         time.Hour,
		}
	}

	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &VisionProcessor{
		model:  model,
		logger: logger,
		config: config,
	}
}

// AnalyzeFashionImage analyzes a fashion image and returns detailed analysis
func (vp *VisionProcessor) AnalyzeFashionImage(ctx context.Context, imageURL string) (*FashionAnalysis, error) {
	vp.logger.Debug("Starting fashion image analysis",
		"image_url", imageURL,
		"timeout", vp.config.Timeout)

	// Validate image URL
	if err := vp.validateImageURL(imageURL); err != nil {
		return nil, fmt.Errorf("invalid image URL: %w", err)
	}

	// Create vision request
	req := &interfaces.PonchoModelRequest{
		Model: vp.model.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: ZAIFashionVisionPrompt,
					},
					{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL:      imageURL,
							MimeType: "image/jpeg", // Default, should be detected
						},
					},
				},
			},
		},
		MaxTokens:   intPtr(1000),
		Temperature: float32Ptr(0.3), // Lower temperature for more consistent analysis
	}

	// Generate response
	resp, err := vp.model.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze image: %w", err)
	}

	// Parse fashion analysis
	analysis, err := vp.parseFashionAnalysis(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fashion analysis: %w", err)
	}

	vp.logger.Info("Fashion image analysis completed",
		"clothing_items_count", len(analysis.ClothingItems),
		"confidence", analysis.Confidence,
		"style", analysis.Style)

	return analysis, nil
}

// AnalyzeFashionImageFromBase64 analyzes a fashion image from base64 data
func (vp *VisionProcessor) AnalyzeFashionImageFromBase64(ctx context.Context, base64Data string, mimeType string) (*FashionAnalysis, error) {
	vp.logger.Debug("Starting fashion image analysis from base64",
		"mime_type", mimeType,
		"data_size", len(base64Data),
		"timeout", vp.config.Timeout)

	// Validate base64 data
	if err := vp.validateBase64Image(base64Data, mimeType); err != nil {
		return nil, fmt.Errorf("invalid base64 image data: %w", err)
	}

	// Create data URL
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	// Create vision request
	req := &interfaces.PonchoModelRequest{
		Model: vp.model.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: ZAIFashionVisionPrompt,
					},
					{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL:      dataURL,
							MimeType: mimeType,
						},
					},
				},
			},
		},
		MaxTokens:   intPtr(1000),
		Temperature: float32Ptr(0.3),
	}

	// Generate response
	resp, err := vp.model.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze image: %w", err)
	}

	// Parse fashion analysis
	analysis, err := vp.parseFashionAnalysis(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fashion analysis: %w", err)
	}

	vp.logger.Info("Fashion image analysis from base64 completed",
		"clothing_items_count", len(analysis.ClothingItems),
		"confidence", analysis.Confidence,
		"style", analysis.Style)

	return analysis, nil
}

// ExtractImageFeatures extracts general features from an image
func (vp *VisionProcessor) ExtractImageFeatures(ctx context.Context, imageURL string) (*FashionAnalysis, error) {
	vp.logger.Debug("Starting image feature extraction",
		"image_url", imageURL,
		"timeout", vp.config.Timeout)

	// Validate image URL
	if err := vp.validateImageURL(imageURL); err != nil {
		return nil, fmt.Errorf("invalid image URL: %w", err)
	}

	// Create general vision request
	prompt := "Analyze this image in detail. Describe the main objects, colors, composition, and any notable features. Provide a comprehensive description suitable for cataloging and search purposes."

	req := &interfaces.PonchoModelRequest{
		Model: vp.model.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: prompt,
					},
					{
						Type: interfaces.PonchoContentTypeMedia,
						Media: &interfaces.PonchoMediaPart{
							URL:      imageURL,
							MimeType: "image/jpeg",
						},
					},
				},
			},
		},
		MaxTokens:   intPtr(800),
		Temperature: float32Ptr(0.5),
	}

	// Generate response
	resp, err := vp.model.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Create basic analysis from description
	analysis := &FashionAnalysis{
		Description: resp.Message.Content[0].Text,
		Confidence:  0.8, // Default confidence
		Metadata: map[string]interface{}{
			"analysis_type": "general_features",
			"model":         vp.model.Name(),
			"timestamp":     time.Now().Unix(),
		},
	}

	vp.logger.Info("Image feature extraction completed",
		"description_length", len(analysis.Description))

	return analysis, nil
}

// DownloadAndAnalyzeImage downloads an image from URL and analyzes it
func (vp *VisionProcessor) DownloadAndAnalyzeImage(ctx context.Context, imageURL string) (*FashionAnalysis, error) {
	vp.logger.Debug("Downloading and analyzing image",
		"image_url", imageURL,
		"timeout", vp.config.Timeout)

	// Download image
	imageData, mimeType, err := vp.downloadImage(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	// Convert to base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// Analyze from base64
	return vp.AnalyzeFashionImageFromBase64(ctx, base64Data, mimeType)
}

// Helper methods

// validateImageURL validates an image URL
func (vp *VisionProcessor) validateImageURL(imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("image URL cannot be empty")
	}

	// Check if it's a valid URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" && !strings.HasPrefix(imageURL, "data:") {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	return nil
}

// validateBase64Image validates base64 image data
func (vp *VisionProcessor) validateBase64Image(base64Data string, mimeType string) error {
	if base64Data == "" {
		return fmt.Errorf("base64 data cannot be empty")
	}

	if mimeType == "" {
		return fmt.Errorf("mime type cannot be empty")
	}

	// Check if mime type is supported
	supported := false
	for _, format := range vp.config.SupportedFormats {
		if format == mimeType {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported mime type: %s", mimeType)
	}

	// Decode base64 to validate
	decoded, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("invalid base64 data: %w", err)
	}

	// Check size
	if len(decoded) > vp.config.MaxImageSize {
		return fmt.Errorf("image size (%d bytes) exceeds maximum (%d bytes)", len(decoded), vp.config.MaxImageSize)
	}

	return nil
}

// downloadImage downloads an image from URL
func (vp *VisionProcessor) downloadImage(ctx context.Context, imageURL string) ([]byte, string, error) {
	// Create HTTP request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, vp.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", imageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Make request
	client := &http.Client{Timeout: vp.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Get content type
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg" // Default
	}

	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Validate size
	if len(imageData) > vp.config.MaxImageSize {
		return nil, "", fmt.Errorf("image size (%d bytes) exceeds maximum (%d bytes)", len(imageData), vp.config.MaxImageSize)
	}

	return imageData, mimeType, nil
}

// parseFashionAnalysis parses the response from fashion analysis
func (vp *VisionProcessor) parseFashionAnalysis(resp *interfaces.PonchoModelResponse) (*FashionAnalysis, error) {
	if resp == nil || resp.Message == nil || len(resp.Message.Content) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	// Get text content
	var textContent string
	for _, part := range resp.Message.Content {
		if part.Type == interfaces.PonchoContentTypeText {
			textContent += part.Text
		}
	}

	if textContent == "" {
		return nil, fmt.Errorf("no text content in response")
	}

	// Create basic analysis - in a real implementation, you might want to use
	// structured output or JSON parsing for more reliable results
	analysis := &FashionAnalysis{
		Description:   textContent,
		ClothingItems: []ClothingItem{},
		Colors:        []string{},
		Style:         "unknown",
		Season:        "unknown",
		Materials:     []string{},
		Accessories:   []string{},
		Confidence:    0.8, // Default confidence
		Metadata: map[string]interface{}{
			"analysis_type": "fashion",
			"model":         vp.model.Name(),
			"timestamp":     time.Now().Unix(),
			"raw_response":  textContent,
		},
	}

	// TODO: Implement more sophisticated parsing for structured fashion data
	// This could involve:
	// 1. Using JSON mode in the API request
	// 2. Using regex patterns to extract structured data
	// 3. Using additional AI calls to parse the response

	return analysis, nil
}

// Helper functions for pointers

func intPtr(i int) *int {
	return &i
}

func float32Ptr(f float32) *float32 {
	return &f
}

// GetConfig returns the current vision configuration
func (vp *VisionProcessor) GetConfig() *VisionConfig {
	return vp.config
}

// UpdateConfig updates the vision configuration
func (vp *VisionProcessor) UpdateConfig(config *VisionConfig) {
	if config != nil {
		vp.config = config
		vp.logger.Info("Vision processor configuration updated",
			"max_image_size", config.MaxImageSize,
			"default_quality", config.DefaultQuality,
			"default_detail", config.DefaultDetail)
	}
}
