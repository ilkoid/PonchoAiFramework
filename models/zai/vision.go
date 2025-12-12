package zai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// VisionProcessor handles image processing and vision-specific operations
type VisionProcessor struct {
	client *GLMClient
	logger interfaces.Logger
	config *ImageProcessingConfig
}

// NewVisionProcessor creates a new vision processor
func NewVisionProcessor(client *GLMClient, logger interfaces.Logger, config *ImageProcessingConfig) *VisionProcessor {
	if config == nil {
		config = &DefaultImageProcessingConfig
	}

	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &VisionProcessor{
		client: client,
		logger: logger,
		config: config,
	}
}

// ProcessImage processes an image URL and returns base64 encoded data
func (vp *VisionProcessor) ProcessImage(ctx context.Context, imageURL string) (*ProcessedImageData, error) {
	// Validate URL
	if imageURL == "" {
		return nil, fmt.Errorf("image URL cannot be empty")
	}

	// Check if it's already a base64 data URL
	if strings.HasPrefix(imageURL, "data:image/") {
		return vp.processBase64Image(imageURL)
	}

	// Download image from URL
	imageData, contentType, err := vp.downloadImage(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	// Process and optimize image
	optimizedData, err := vp.optimizeImage(imageData, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize image: %w", err)
	}

	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(optimizedData)

	return &ProcessedImageData{
		Base64Data:  base64Data,
		ContentType: contentType,
		Size:        len(optimizedData),
		OriginalURL: imageURL,
		ProcessedAt: time.Now(),
	}, nil
}

// AnalyzeFashionImage performs fashion-specific analysis on an image
func (vp *VisionProcessor) AnalyzeFashionImage(ctx context.Context, imageURL string, prompt string) (*FashionAnalysis, error) {
	if !vp.config.EnableAnalysis {
		return nil, fmt.Errorf("fashion analysis is disabled")
	}

	// Process the image
	imageData, err := vp.ProcessImage(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Create fashion-specific prompt
	fashionPrompt := vp.buildFashionPrompt(prompt)

	// Create GLM request with vision
	req := &GLMRequest{
		Model: GLMVisionModel,
		Messages: []GLMMessage{
			{
				Role: "user",
				Content: []GLMContentPart{
					{
						Type: GLMContentTypeImageURL,
						ImageURL: &GLMImageURL{
							URL: fmt.Sprintf("data:%s;base64,%s", imageData.ContentType, imageData.Base64Data),
						},
					},
					{
						Type: GLMContentTypeText,
						Text: fashionPrompt,
					},
				},
			},
		},
		MaxTokens:   &vp.client.config.MaxTokens,
		Temperature: &vp.client.config.Temperature,
		Thinking: &GLMThinking{
			Type: GLMThinkingEnabled,
		},
	}

	// Call GLM API
	response, err := vp.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze image: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Parse fashion analysis from response
	contentStr, ok := response.Choices[0].Message.Content.(string)
	if !ok {
		return nil, fmt.Errorf("message content is not a string")
	}

	analysis, err := vp.parseFashionAnalysis(contentStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fashion analysis: %w", err)
	}

	// Add metadata
	analysis.Confidence = vp.calculateConfidence(contentStr)
	analysis.Coordinates = vp.extractCoordinates(contentStr)

	return analysis, nil
}

// ExtractProductFeatures extracts specific product features from an image
func (vp *VisionProcessor) ExtractProductFeatures(ctx context.Context, imageURL string, features []string) (map[string]interface{}, error) {
	if !vp.config.EnableAnalysis {
		return nil, fmt.Errorf("feature extraction is disabled")
	}

	// Process the image
	imageData, err := vp.ProcessImage(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// Build feature extraction prompt
	prompt := vp.buildFeatureExtractionPrompt(features)

	// Create GLM request
	req := &GLMRequest{
		Model: GLMVisionModel,
		Messages: []GLMMessage{
			{
				Role: "user",
				Content: []GLMContentPart{
					{
						Type: GLMContentTypeImageURL,
						ImageURL: &GLMImageURL{
							URL: fmt.Sprintf("data:%s;base64,%s", imageData.ContentType, imageData.Base64Data),
						},
					},
					{
						Type: GLMContentTypeText,
						Text: prompt,
					},
				},
			},
		},
		MaxTokens:   &vp.client.config.MaxTokens,
		Temperature: &vp.client.config.Temperature,
		Thinking: &GLMThinking{
			Type: GLMThinkingEnabled,
		},
	}

	// Call GLM API
	response, err := vp.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Parse features
	contentStr, ok := response.Choices[0].Message.Content.(string)
	if !ok {
		return nil, fmt.Errorf("message content is not a string")
	}

	featuresMap, err := vp.parseFeatures(contentStr, features)
	if err != nil {
		return nil, fmt.Errorf("failed to parse features: %w", err)
	}

	return featuresMap, nil
}

// ProcessedImageData represents processed image information
type ProcessedImageData struct {
	Base64Data  string    `json:"base64_data"`
	ContentType string    `json:"content_type"`
	Size        int       `json:"size"`
	OriginalURL string    `json:"original_url"`
	ProcessedAt time.Time `json:"processed_at"`
}

// downloadImage downloads an image from URL
func (vp *VisionProcessor) downloadImage(ctx context.Context, imageURL string) ([]byte, string, error) {
	// Validate URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return nil, "", fmt.Errorf("invalid image URL: %w", err)
	}

	if !strings.HasPrefix(parsedURL.Scheme, "http") {
		return nil, "", fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "PonchoFramework/1.0")

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	// Read data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Get content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from data
		contentType = http.DetectContentType(data)
	}

	// Validate content type
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", fmt.Errorf("invalid content type: %s", contentType)
	}

	return data, contentType, nil
}

// processBase64Image processes an already base64 encoded image
func (vp *VisionProcessor) processBase64Image(dataURL string) (*ProcessedImageData, error) {
	// Parse data URL
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data URL format")
	}

	// Extract content type
	header := parts[0]
	if !strings.HasPrefix(header, "data:image/") {
		return nil, fmt.Errorf("invalid data URL header")
	}

	contentType := strings.TrimPrefix(header, "data:")
	if semicolonIndex := strings.Index(contentType, ";"); semicolonIndex != -1 {
		contentType = contentType[:semicolonIndex]
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	return &ProcessedImageData{
		Base64Data:  parts[1],
		ContentType: contentType,
		Size:        len(data),
		OriginalURL: dataURL,
		ProcessedAt: time.Now(),
	}, nil
}

// optimizeImage optimizes image for API usage
func (vp *VisionProcessor) optimizeImage(data []byte, contentType string) ([]byte, error) {
	// Check size limit
	if int64(len(data)) <= vp.config.MaxSizeBytes {
		return data, nil
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, err // Return original if decode fails
	}

	// Resize if needed
	bounds := img.Bounds()
	if bounds.Dx() <= vp.config.MaxWidth && bounds.Dy() <= vp.config.MaxHeight {
		return data, nil
	}

	// Calculate new dimensions
	newWidth, newHeight := vp.calculateDimensions(bounds.Dx(), bounds.Dy())

	// Create new image
	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple resize (would use better library in production)
	// This is a placeholder - in production use a proper image resizing library
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := x * bounds.Dx() / newWidth
			srcY := y * bounds.Dy() / newHeight
			newImg.Set(x, y, img.At(srcX, srcY))
		}
	}

	// Encode back to bytes
	var buf bytes.Buffer
	switch contentType {
	case "image/png":
		err = png.Encode(&buf, newImg)
	case "image/jpeg":
		err = jpeg.Encode(&buf, newImg, &jpeg.Options{Quality: vp.config.Quality})
	default:
		err = jpeg.Encode(&buf, newImg, &jpeg.Options{Quality: vp.config.Quality})
	}

	if err != nil {
		return data, err // Return original if encode fails
	}

	return buf.Bytes(), nil
}

// calculateDimensions calculates new dimensions maintaining aspect ratio
func (vp *VisionProcessor) calculateDimensions(width, height int) (int, int) {
	if width <= vp.config.MaxWidth && height <= vp.config.MaxHeight {
		return width, height
	}

	ratio := float64(width) / float64(height)
	if width > height {
		newWidth := vp.config.MaxWidth
		newHeight := int(float64(newWidth) / ratio)
		return newWidth, newHeight
	} else {
		newHeight := vp.config.MaxHeight
		newWidth := int(float64(newHeight) * ratio)
		return newWidth, newHeight
	}
}

// buildFashionPrompt builds a fashion-specific analysis prompt
func (vp *VisionProcessor) buildFashionPrompt(customPrompt string) string {
	basePrompt := `You are a fashion expert specializing in clothing and accessory analysis. Analyze the provided fashion image and provide detailed information.

Please analyze and describe:
1. Item type and category (e.g., dress, shirt, pants, shoes, bag)
2. Style (e.g., casual, formal, sporty, elegant)
3. Materials visible (e.g., cotton, silk, denim, leather)
4. Color palette and dominant colors
5. Season suitability (spring, summer, fall, winter, all-season)
6. Target gender (men, women, unisex)
7. Notable features and details
8. Occasion suitability

If coordinates are requested, provide them in [[xmin,ymin,xmax,ymax]] format.`

	if customPrompt != "" {
		return basePrompt + "\n\nAdditional specific request: " + customPrompt
	}

	return basePrompt
}

// buildFeatureExtractionPrompt builds a prompt for specific feature extraction
func (vp *VisionProcessor) buildFeatureExtractionPrompt(features []string) string {
	featureList := strings.Join(features, ", ")
	return fmt.Sprintf(`Extract the following specific features from the fashion image: %s

Please provide the results in JSON format with the feature names as keys and the extracted values as values. Be specific and accurate in your analysis.`, featureList)
}

// parseFashionAnalysis parses fashion analysis from response text
func (vp *VisionProcessor) parseFashionAnalysis(responseText string) (*FashionAnalysis, error) {
	// This is a simplified parser - in production, use more sophisticated parsing
	analysis := &FashionAnalysis{
		Features: make(map[string]interface{}),
	}

	// Extract basic information (simplified)
	lines := strings.Split(responseText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "category") {
			analysis.Category = vp.extractValue(line)
		} else if strings.Contains(strings.ToLower(line), "style") {
			analysis.Style = vp.extractValue(line)
		} else if strings.Contains(strings.ToLower(line), "material") {
			analysis.Materials = vp.extractList(line)
		} else if strings.Contains(strings.ToLower(line), "color") {
			analysis.Colors = vp.extractList(line)
		} else if strings.Contains(strings.ToLower(line), "season") {
			analysis.Season = vp.extractValue(line)
		} else if strings.Contains(strings.ToLower(line), "gender") {
			analysis.Gender = vp.extractValue(line)
		}
	}

	// Store full description
	analysis.Description = responseText

	return analysis, nil
}

// parseFeatures parses specific features from response
func (vp *VisionProcessor) parseFeatures(responseText string, requestedFeatures []string) (map[string]interface{}, error) {
	features := make(map[string]interface{})

	// Try to parse as JSON first
	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(responseText), &jsonResult); err == nil {
		return jsonResult, nil
	}

	// Fallback to text parsing
	lines := strings.Split(responseText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		for _, feature := range requestedFeatures {
			if strings.Contains(strings.ToLower(line), strings.ToLower(feature)) {
				features[feature] = vp.extractValue(line)
			}
		}
	}

	return features, nil
}

// extractValue extracts a value from a line of text
func (vp *VisionProcessor) extractValue(line string) string {
	// Look for patterns like "Category: dress" or "Category - dress"
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}

	parts = strings.SplitN(line, "-", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}

	return line
}

// extractList extracts a list of values from a line of text
func (vp *VisionProcessor) extractList(line string) []string {
	value := vp.extractValue(line)
	// Split by common separators
	separators := []string{",", ";", "/", "|"}

	for _, sep := range separators {
		if strings.Contains(value, sep) {
			items := strings.Split(value, sep)
			result := make([]string, len(items))
			for i, item := range items {
				result[i] = strings.TrimSpace(item)
			}
			return result
		}
	}

	return []string{value}
}

// calculateConfidence calculates confidence score from response
func (vp *VisionProcessor) calculateConfidence(responseText string) float64 {
	// Simple heuristic based on response length and detail
	words := strings.Fields(responseText)
	if len(words) < 10 {
		return 0.3
	} else if len(words) < 50 {
		return 0.7
	} else {
		return 0.9
	}
}

// extractCoordinates extracts bounding box coordinates from response
func (vp *VisionProcessor) extractCoordinates(responseText string) []BoundingCoordinates {
	var coords []BoundingCoordinates

	// Look for coordinate patterns [[xmin,ymin,xmax,ymax]]
	pattern := "[["
	start := strings.Index(responseText, pattern)
	for start != -1 {
		end := strings.Index(responseText[start:], "]]")
		if end == -1 {
			break
		}
		end += start + 2

		coordStr := responseText[start+2 : end-2]
		values := strings.Split(coordStr, ",")
		if len(values) == 4 {
			coord := BoundingCoordinates{
				Label: "detected_object",
			}
			fmt.Sscanf(coordStr, "%f,%f,%f,%f", &coord.XMin, &coord.YMin, &coord.XMax, &coord.YMax)
			coords = append(coords, coord)
		}

		// Look for next occurrence
		start = strings.Index(responseText[end:], pattern)
		if start != -1 {
			start += end
		}
	}

	return coords
}
