package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/ilkoid/PonchoAiFramework/core/media"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MediaHelper provides utilities for working with media in AI models
type MediaHelper struct {
	pipeline *media.MediaPipeline
	logger   interfaces.Logger
}

// NewMediaHelper creates a new media helper
func NewMediaHelper(pipeline *media.MediaPipeline, logger interfaces.Logger) *MediaHelper {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &MediaHelper{
		pipeline: pipeline,
		logger:   logger,
	}
}

// PrepareMediaContent converts media data to PonchoContentPart for model requests
func (mh *MediaHelper) PrepareMediaContent(ctx context.Context, mediaList []*interfaces.MediaData, model interfaces.PonchoModel) ([]*interfaces.PonchoContentPart, error) {
	if len(mediaList) == 0 {
		return nil, nil
	}

	// Prepare media for the specific model
	// TODO: Преобразовать interfaces.MediaData в media.MediaData
	// preparedMedia, err := mh.pipeline.PrepareForModel(ctx, mediaList, model)
	// TODO: Преобразование типов
	_ = mediaList // временно

	// Convert to content parts
	var parts []*interfaces.PonchoContentPart
	for i, mediaItem := range mediaList {
		part, err := mh.mediaToContentPart(mediaItem, i)
		if err != nil {
			return nil, fmt.Errorf("failed to convert media item %d to content part: %w", i, err)
		}
		parts = append(parts, part)
	}

	return parts, nil
}

// mediaToContentPart converts MediaData to PonchoContentPart
func (mh *MediaHelper) mediaToContentPart(mediaData *interfaces.MediaData, index int) (*interfaces.PonchoContentPart, error) {
	// Get the URL
	url := mediaData.URL
	if url == "" && len(mediaData.Bytes) > 0 {
		// Create data URL from bytes
		url = fmt.Sprintf("data:%s;base64,%x", mediaData.MimeType, mediaData.Bytes)
	}
	if url == "" {
		return nil, fmt.Errorf("media has no URL or data")
	}

	return &interfaces.PonchoContentPart{
		Type: interfaces.PonchoContentTypeMedia,
		Media: &interfaces.PonchoMediaPart{
			URL:      url,
			MimeType: mediaData.MimeType,
		},
	}, nil
}

// CreateVisionMessage creates a message with text and media content for vision models
func (mh *MediaHelper) CreateVisionMessage(
	ctx context.Context,
	text string,
	mediaList []*interfaces.MediaData,
	model interfaces.PonchoModel,
) (*interfaces.PonchoMessage, error) {
	if !model.SupportsVision() {
		return nil, fmt.Errorf("model %s does not support vision", model.Name())
	}

	// Start with text content
	content := []*interfaces.PonchoContentPart{
		{
			Type: interfaces.PonchoContentTypeText,
			Text: text,
		},
	}

	// Add media content
	mediaContent, err := mh.PrepareMediaContent(ctx, mediaList, model)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare media content: %w", err)
	}

	content = append(content, mediaContent...)

	return &interfaces.PonchoMessage{
		Role:    interfaces.PonchoRoleUser,
		Content: content,
	}, nil
}

// ExtractMediaFromContext extracts media from flow context
func (mh *MediaHelper) ExtractMediaFromContext(
	ctx interfaces.FlowContext,
	keys []string,
) ([]*interfaces.MediaData, error) {
	var mediaList []*interfaces.MediaData

	for _, key := range keys {
		// Try to get media directly
		if mediaData, err := ctx.GetMedia(key); err == nil {
			mediaList = append(mediaList, mediaData)
			continue
		}

		// Try to get as bytes and convert
		if bytes, err := ctx.GetBytes(key); err == nil {
			// Try to detect mime type
			mimeType := "image/jpeg" // default
			if strings.HasSuffix(key, ".png") {
				mimeType = "image/png"
			} else if strings.HasSuffix(key, ".webp") {
				mimeType = "image/webp"
			}

			mediaData := &interfaces.MediaData{
				URL:      "",
				Bytes:    bytes,
				MimeType: mimeType,
				Size:     int64(len(bytes)),
				Metadata: make(map[string]interface{}),
			}
			mediaList = append(mediaList, mediaData)
		}
	}

	return mediaList, nil
}

// StoreMediaInContext stores media data in flow context
func (mh *MediaHelper) StoreMediaInContext(
	ctx interfaces.FlowContext,
	prefix string,
	mediaList []*interfaces.MediaData,
) error {
	return ctx.SetMedia(prefix, mediaList[0]) // Store first for now
}

// CreateBatchVisionRequest creates a request for analyzing multiple images
func (mh *MediaHelper) CreateBatchVisionRequest(
	ctx context.Context,
	basePrompt string,
	mediaList []*interfaces.MediaData,
	model interfaces.PonchoModel,
	options *VisionOptions,
) (*interfaces.PonchoModelRequest, error) {
	if options == nil {
		options = DefaultVisionOptions()
	}

	// Build detailed prompt for batch analysis
	prompt := mh.buildBatchPrompt(basePrompt, mediaList, options)

	// Create message with media
	message, err := mh.CreateVisionMessage(ctx, prompt, mediaList, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create vision message: %w", err)
	}

	// Configure request parameters
	req := &interfaces.PonchoModelRequest{
		Model:    model.Name(),
		Messages: []*interfaces.PonchoMessage{message},
	}

	// Apply options
	if options.MaxTokens != nil {
		req.MaxTokens = options.MaxTokens
	}

	if options.Temperature != nil {
		req.Temperature = options.Temperature
	}

	return req, nil
}

// VisionOptions provides options for vision model requests
type VisionOptions struct {
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`
	Detail      string   `json:"detail,omitempty"`        // "low", "high", "auto"
	Format      string   `json:"format,omitempty"`       // "text", "json"
	BatchMode   bool     `json:"batch_mode,omitempty"`   // Analyze all images together
	Individual  bool     `json:"individual,omitempty"`   // Analyze each image separately
	Fields      []string `json:"fields,omitempty"`       // Specific fields to extract
}

// DefaultVisionOptions returns default vision options
func DefaultVisionOptions() *VisionOptions {
	maxTokens := 1000
	temperature := float32(0.3)

	return &VisionOptions{
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
		Detail:      "auto",
		Format:      "text",
		BatchMode:   true,
		Individual:  false,
	}
}

// buildBatchPrompt creates a prompt for batch image analysis
func (mh *MediaHelper) buildBatchPrompt(
	basePrompt string,
	mediaList []*interfaces.MediaData,
	options *VisionOptions,
) string {
	prompt := basePrompt

	if options.BatchMode && len(mediaList) > 1 {
		prompt += fmt.Sprintf("\n\nAnalyze the %d provided images as a collection.", len(mediaList))
	}

	if options.Individual && len(mediaList) > 1 {
		prompt += "\n\nPlease analyze each image individually and provide results for each."
	}

	if len(options.Fields) > 0 {
		prompt += fmt.Sprintf("\n\nFocus on extracting: %s", strings.Join(options.Fields, ", "))
	}

	if options.Format == "json" {
		prompt += "\n\nPlease provide the response in JSON format."
	}

	return prompt
}

// ProcessVisionResponse processes a vision model response
func (mh *MediaHelper) ProcessVisionResponse(
	resp *interfaces.PonchoModelResponse,
	options *VisionOptions,
) (*VisionAnalysisResult, error) {
	if resp == nil || resp.Message == nil || len(resp.Message.Content) == 0 {
		return nil, fmt.Errorf("empty response from vision model")
	}

	text := resp.Message.Content[0].Text
	result := &VisionAnalysisResult{
		RawText: text,
		Success: true,
	}

	// Try to parse as JSON if requested
	if options != nil && options.Format == "json" {
		// In a real implementation, you would parse JSON here
		result.StructuredData = map[string]interface{}{
			"raw_json": text,
		}
	}

	// Extract usage information if available
	if resp.Usage != nil {
		result.TokenUsage = &TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return result, nil
}

// VisionAnalysisResult represents the result of vision analysis
type VisionAnalysisResult struct {
	RawText        string                 `json:"raw_text"`
	StructuredData map[string]interface{} `json:"structured_data,omitempty"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
	TokenUsage     *TokenUsage            `json:"token_usage,omitempty"`
}

// TokenUsage определен в types.go

// Helper functions for common vision patterns

// FashionVisionPrompts provides common prompts for fashion image analysis
var FashionVisionPrompts = struct {
	GeneralAnalysis    string
	ColorAnalysis      string
	StyleAnalysis      string
	MaterialAnalysis   string
	DetailedFeatures   string
}{
	GeneralAnalysis:    "Analyze this fashion item and describe its key visual characteristics.",
	ColorAnalysis:      "Extract and describe all visible colors, patterns, and color combinations in this fashion item.",
	StyleAnalysis:      "Determine the style category (casual, formal, sport, etc.) and aesthetic characteristics of this fashion item.",
	MaterialAnalysis:   "Identify the apparent materials and textures visible in this fashion item.",
	DetailedFeatures:   "Provide a detailed analysis of this fashion item including style, colors, materials, and notable features.",
}

// CreateFashionAnalysisRequest creates a request for fashion-specific analysis
func (mh *MediaHelper) CreateFashionAnalysisRequest(
	ctx context.Context,
	analysisType string,
	mediaList []*interfaces.MediaData,
	model interfaces.PonchoModel,
) (*interfaces.PonchoModelRequest, error) {
	var prompt string

	switch analysisType {
	case "general":
		prompt = FashionVisionPrompts.GeneralAnalysis
	case "colors":
		prompt = FashionVisionPrompts.ColorAnalysis
	case "style":
		prompt = FashionVisionPrompts.StyleAnalysis
	case "materials":
		prompt = FashionVisionPrompts.MaterialAnalysis
	case "detailed":
		prompt = FashionVisionPrompts.DetailedFeatures
	default:
		prompt = FashionVisionPrompts.GeneralAnalysis
	}

	return mh.CreateBatchVisionRequest(ctx, prompt, mediaList, model, DefaultVisionOptions())
}

// MultiModelMediaHelper supports media operations across different model providers
type MultiModelMediaHelper struct {
	helpers map[string]*MediaHelper
	logger  interfaces.Logger
}

// NewMultiModelMediaHelper creates a helper for multiple model providers
func NewMultiModelMediaHelper(pipeline *media.MediaPipeline, logger interfaces.Logger) *MultiModelMediaHelper {
	return &MultiModelMediaHelper{
		helpers: make(map[string]*MediaHelper),
		logger:  logger,
	}
}

// GetHelper gets or creates a helper for a specific model provider
func (mmh *MultiModelMediaHelper) GetHelper(provider string) *MediaHelper {
	if helper, exists := mmh.helpers[provider]; exists {
		return helper
	}

	// Create new helper for provider
	helper := NewMediaHelper(nil, mmh.logger)
	mmh.helpers[provider] = helper
	return helper
}