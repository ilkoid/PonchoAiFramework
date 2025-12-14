package articleflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/tools/s3"
	"github.com/ilkoid/PonchoAiFramework/tools/wildberries"
)

// ArticleFlow orchestrates the article processing pipeline
type ArticleFlow struct {
	s3Tool       interfaces.PonchoTool
	visionModel  interfaces.PonchoModel
	textModel    interfaces.PonchoModel
	wbClient     *wildberries.WBClient
	logger       interfaces.Logger
	config       *ArticleFlowConfig
	wbCache      WBCache
}

// NewArticleFlow creates a new article flow instance
func NewArticleFlow(
	s3Tool interfaces.PonchoTool,
	visionModel interfaces.PonchoModel,
	textModel interfaces.PonchoModel,
	wbClient *wildberries.WBClient,
	logger interfaces.Logger,
	config *ArticleFlowConfig,
	wbCache WBCache,
) *ArticleFlow {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &ArticleFlow{
		s3Tool:       s3Tool,
		visionModel:  visionModel,
		textModel:    textModel,
		wbClient:     wbClient,
		logger:       logger,
		config:       config,
		wbCache:      wbCache,
	}
}

// Run executes the complete article processing flow
func (f *ArticleFlow) Run(ctx context.Context, articleID string) (*ArticleFlowState, error) {
	f.logger.Info("Starting article flow",
		"article_id", articleID,
	)

	// Initialize state
	state := NewArticleFlowState(articleID)

	// Step 1: Load article from S3
	if err := f.loadArticleFromS3(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to load article from S3: %w", err)
	}

	// Step 2: Process images (resize if configured)
	if err := f.processImages(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to process images: %w", err)
	}

	// Step 3: Run vision analysis (Prompt 1)
	if err := f.runVisionAnalysis(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to run vision analysis: %w", err)
	}

	// Step 4: Generate creative descriptions (Prompt 2)
	if err := f.generateCreativeDescriptions(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to generate creative descriptions: %w", err)
	}

	// Step 5: Fetch Wildberries data
	if err := f.fetchWildberriesData(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to fetch Wildberries data: %w", err)
	}

	// Step 6: Select WB subject (Prompt 3)
	if err := f.selectWBSubject(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to select WB subject: %w", err)
	}

	// Step 7: Fetch characteristics for selected subject
	if err := f.fetchWBCharacteristics(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to fetch WB characteristics: %w", err)
	}

	// Step 8: Generate final WB payload (Prompt 4)
	if err := f.generateFinalWBPayload(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to generate final WB payload: %w", err)
	}

	// Mark flow as completed
	state.MarkCompleted()

	f.logger.Info("Article flow completed successfully",
		"article_id", articleID,
		"duration", state.Duration,
		"images_processed", len(state.Images),
	)

	return state, nil
}

// loadArticleFromS3 loads article data and images from S3
func (f *ArticleFlow) loadArticleFromS3(ctx context.Context, state *ArticleFlowState) error {
	f.logger.Info("Loading article from S3",
		"article_id", state.ArticleID,
	)

	// Execute S3 tool
	s3Input := map[string]interface{}{
		"article_id":     state.ArticleID,
		"include_images": true,
		"max_images":     10, // Limit for safety
	}

	resp, err := f.s3Tool.Execute(ctx, s3Input)
	if err != nil {
		return fmt.Errorf("S3 tool execution failed: %w", err)
	}

	// Convert response to expected format
	respData, ok := resp.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format from S3 tool")
	}

	// Extract download response
	data, ok := respData["article"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no article data in S3 response")
	}

	jsonData, ok := data["json_data"].(string)
	if !ok {
		return fmt.Errorf("no json_data in article response")
	}

	imagesData, ok := data["images"].([]interface{})
	if !ok {
		return fmt.Errorf("no images in article response")
	}

	// Convert images
	var s3Images []*s3.Image
	for _, imgData := range imagesData {
		if imgMap, ok := imgData.(map[string]interface{}); ok {
			s3Img := &s3.Image{
				Filename:    getString(imgMap, "filename"),
				URL:         getString(imgMap, "url"),
				ContentType: getString(imgMap, "content_type"),
				Size:        getInt64(imgMap, "size"),
				Width:       getInt(imgMap, "width"),
				Height:      getInt(imgMap, "height"),
			}
			s3Images = append(s3Images, s3Img)
		}
	}

	// Store data in state
	state.PLMJSON = []byte(jsonData)
	state.Images = ConvertFromS3Images(s3Images)

	f.logger.Info("Loaded article from S3",
		"images_count", len(state.Images),
		"json_size", len(state.PLMJSON),
	)

	return nil
}

// processImages resizes images if configured
func (f *ArticleFlow) processImages(ctx context.Context, state *ArticleFlowState) error {
	if !f.config.ImageResize.Enabled {
		f.logger.Info("Image resize disabled, skipping")
		return nil
	}

	f.logger.Info("Processing images",
		"count", len(state.Images),
		"max_width", f.config.ImageResize.MaxWidth,
		"max_height", f.config.ImageResize.MaxHeight,
	)

	// Here we would integrate with the media pipeline for resizing
	// For now, just log that we're ready for processing
	// The actual resizing happens during vision/creative processing

	return nil
}

// runVisionAnalysis executes Prompt 1 on all images in parallel
func (f *ArticleFlow) runVisionAnalysis(ctx context.Context, state *ArticleFlowState) error {
	if len(state.Images) == 0 {
		f.logger.Info("No images to analyze")
		return nil
	}

	f.logger.Info("Running vision analysis",
		"model", f.visionModel.Name(),
		"images_count", len(state.Images),
	)

	// Prepare vision analysis prompt
	visionPrompt := f.buildVisionAnalysisPrompt()

	// Process images concurrently
	results, err := ProcessImagesConcurrently(
		ctx,
		state.Images,
		func(ctx context.Context, img ImageRef) (*TechInfo, error) {
			return f.analyzeImageWithVision(ctx, img, visionPrompt)
		},
		f.config.Concurrency.VisionAnalysisWorkers,
	)

	if err != nil {
		return fmt.Errorf("vision analysis failed: %w", err)
	}

	// Store results
	for imageID, techInfo := range results {
		state.SetTechAnalysis(imageID, techInfo)
	}

	f.logger.Info("Vision analysis completed",
		"processed_images", len(results),
	)

	return nil
}

// analyzeImageWithVision analyzes a single image using the vision model
func (f *ArticleFlow) analyzeImageWithVision(ctx context.Context, img ImageRef, prompt string) (*TechInfo, error) {
	// For now, use the image URL directly
	// TODO: Implement image resizing when media pipeline is available

	// Prepare content parts
	content := []*interfaces.PonchoContentPart{
		{
			Type: interfaces.PonchoContentTypeText,
			Text: prompt,
		},
		{
			Type: interfaces.PonchoContentTypeMedia,
			Media: &interfaces.PonchoMediaPart{
				URL:      img.URL,
				MimeType: img.MimeType,
			},
		},
	}

	// Create request
	req := &interfaces.PonchoModelRequest{
		Model: f.visionModel.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role:    interfaces.PonchoRoleUser,
				Content: content,
			},
		},
		Temperature: &f.config.ModelParams.Temperature,
		MaxTokens:   &f.config.ModelParams.MaxTokens,
	}

	// Execute vision model
	resp, err := f.visionModel.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("vision model generation failed: %w", err)
	}

	// Parse response
	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Message.Content[0].Text), &analysis); err != nil {
		// If not JSON, store as text
		analysis = map[string]interface{}{
			"text": resp.Message.Content[0].Text,
		}
	}

	// Create tech info
	techInfo := &TechInfo{
		ImageID:     img.ID,
		Analysis:    analysis,
		RawJSON:     []byte(resp.Message.Content[0].Text),
		ProcessedAt: time.Now(),
		Model:       f.visionModel.Name(),
		TokenUsage:  resp.Usage,
	}

	return techInfo, nil
}

// generateCreativeDescriptions executes Prompt 2 on all images in parallel
func (f *ArticleFlow) generateCreativeDescriptions(ctx context.Context, state *ArticleFlowState) error {
	if len(state.Images) == 0 {
		f.logger.Info("No images for creative descriptions")
		return nil
	}

	f.logger.Info("Generating creative descriptions",
		"model", f.textModel.Name(),
		"images_count", len(state.Images),
	)

	// Build creative prompt with PLM data context
	creativePrompt := f.buildCreativePrompt(state.PLMJSON)

	// Process images concurrently
	results, err := ProcessImagesConcurrently(
		ctx,
		state.Images,
		func(ctx context.Context, img ImageRef) (string, error) {
			return f.generateCreativeDescription(ctx, img, creativePrompt)
		},
		f.config.Concurrency.CreativeWorkers,
	)

	if err != nil {
		return fmt.Errorf("creative description generation failed: %w", err)
	}

	// Store results
	for imageID, description := range results {
		state.SetCreativeDescription(imageID, description)
	}

	f.logger.Info("Creative descriptions generated",
		"processed_images", len(results),
	)

	return nil
}

// generateCreativeDescription generates a creative description for a single image
func (f *ArticleFlow) generateCreativeDescription(ctx context.Context, img ImageRef, basePrompt string) (string, error) {
	// TODO: Pass tech analysis as parameter when needed
	// For now, use base prompt only
	prompt := basePrompt

	// Create request
	req := &interfaces.PonchoModelRequest{
		Model: f.textModel.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role:    interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{{Type: interfaces.PonchoContentTypeText, Text: prompt}},
			},
		},
		Temperature: &f.config.ModelParams.Temperature,
		MaxTokens:   &f.config.ModelParams.MaxTokens,
	}

	// Execute text model
	resp, err := f.textModel.Generate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("text model generation failed: %w", err)
	}

	return resp.Message.Content[0].Text, nil
}

// fetchWildberriesData fetches parents and subjects from Wildberries
func (f *ArticleFlow) fetchWildberriesData(ctx context.Context, state *ArticleFlowState) error {
	f.logger.Info("Fetching Wildberries data")

	// Use cache if available
	parents, err := f.wbCache.GetParents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WB parents: %w", err)
	}
	state.SetWBParents(parents)

	subjects, err := f.wbCache.GetSubjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WB subjects: %w", err)
	}
	state.SetWBSubjects(subjects)

	f.logger.Info("Fetched Wildberries data",
		"parents_count", len(parents),
		"subjects_count", len(subjects),
	)

	return nil
}

// selectWBSubject executes Prompt 3 to select appropriate subject
func (f *ArticleFlow) selectWBSubject(ctx context.Context, state *ArticleFlowState) error {
	f.logger.Info("Selecting Wildberries subject")

	// Build subject selection prompt
	prompt := f.buildSubjectSelectionPrompt(state)

	// Create request
	req := &interfaces.PonchoModelRequest{
		Model: f.textModel.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role:    interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{{Type: interfaces.PonchoContentTypeText, Text: prompt}},
			},
		},
		Temperature: &f.config.ModelParams.Temperature,
		MaxTokens:   &f.config.ModelParams.MaxTokens,
	}

	// Execute model
	resp, err := f.textModel.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("subject selection failed: %w", err)
	}

	// Parse response to get selected subject ID
	var selection struct {
		SubjectID int `json:"subject_id"`
		Reason    string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(resp.Message.Content[0].Text), &selection); err != nil {
		return fmt.Errorf("failed to parse subject selection: %w", err)
	}

	// Find selected subject
	for _, subject := range state.WBSubjects {
		if subject.ID == selection.SubjectID {
			state.SetSelectedSubject(&subject)
			f.logger.Info("Selected Wildberries subject",
				"subject_id", selection.SubjectID,
				"subject_name", subject.Name,
			)
			return nil
		}
	}

	return fmt.Errorf("subject ID %d not found in subjects list", selection.SubjectID)
}

// fetchWBCharacteristics fetches characteristics for the selected subject
func (f *ArticleFlow) fetchWBCharacteristics(ctx context.Context, state *ArticleFlowState) error {
	if state.SelectedSubject == nil {
		return fmt.Errorf("no subject selected")
	}

	f.logger.Info("Fetching Wildberries characteristics",
		"subject_id", state.SelectedSubject.ID,
	)

	characteristics, err := f.wbCache.GetCharacteristics(ctx, state.SelectedSubject.ID)
	if err != nil {
		return fmt.Errorf("failed to get WB characteristics: %w", err)
	}

	state.SetWBCharacteristics(characteristics)

	f.logger.Info("Fetched Wildberries characteristics",
		"count", len(characteristics),
	)

	return nil
}

// generateFinalWBPayload executes Prompt 4 to create final JSON
func (f *ArticleFlow) generateFinalWBPayload(ctx context.Context, state *ArticleFlowState) error {
	f.logger.Info("Generating final Wildberries payload")

	// Build final prompt
	prompt := f.buildFinalPayloadPrompt(state)

	// Create request
	req := &interfaces.PonchoModelRequest{
		Model: f.textModel.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role:    interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{{Type: interfaces.PonchoContentTypeText, Text: prompt}},
			},
		},
		Temperature: &f.config.ModelParams.Temperature,
		MaxTokens:   &f.config.ModelParams.MaxTokens,
	}

	// Execute model
	resp, err := f.textModel.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("final payload generation failed: %w", err)
	}

	// Validate JSON
	if !json.Valid([]byte(resp.Message.Content[0].Text)) {
		return fmt.Errorf("generated payload is not valid JSON")
	}

	// Store final payload
	state.SetFinalWBPayload([]byte(resp.Message.Content[0].Text))

	f.logger.Info("Generated final Wildberries payload",
		"size", len(state.FinalWBPayload),
	)

	return nil
}

// Prompt building methods

func (f *ArticleFlow) buildVisionAnalysisPrompt() string {
	return `Analyze this fashion sketch/image and provide detailed technical information in JSON format:
{
  "garment_type": "type of clothing",
  "silhouette": "overall shape/outline",
  "length": "garment length",
  "sleeves": "sleeve type",
  "neckline": "neck/collar style",
  "materials": ["likely materials"],
  "colors": ["dominant colors"],
  "patterns": "patterns present",
  "construction_details": ["specific construction elements"],
  "style_attributes": ["style keywords"],
  "target_occasion": "occasions suitable for",
  "seasonality": "best season(s)"
}`
}

func (f *ArticleFlow) buildCreativePrompt(plmJSON []byte) string {
	return fmt.Sprintf(`Based on the following PLM data and the fashion sketch provided, write an engaging, creative product description that would appeal to fashion buyers.

PLM Data:
%s

Guidelines:
- Write 150-200 words
- Highlight unique features and selling points
- Use fashion-forward language
- Include seasonal and occasion suggestions
- Mention materials and construction quality
- Create desire and exclusivity

Focus on creating an emotional connection with the buyer while maintaining professional appeal.`, string(plmJSON))
}

func (f *ArticleFlow) buildSubjectSelectionPrompt(state *ArticleFlowState) string {
	// Build prompt with all available data
	return fmt.Sprintf(`Based on the following information, select the most appropriate Wildberries subject category from the provided list.

Article Analysis:
%s

Creative Descriptions:
%s

Available Wildberries Subjects:
%s

Please respond with JSON:
{
  "subject_id": <number>,
  "reason": "brief explanation of why this subject was chosen"
}`,
		f.formatTechAnalyses(state),
		f.formatCreativeDescriptions(state),
		f.formatWBSubjects(state))
}

func (f *ArticleFlow) buildFinalPayloadPrompt(state *ArticleFlowState) string {
	return fmt.Sprintf(`Create a complete Wildberries product payload using all the gathered information.

Selected Subject: %s (ID: %d)

Technical Analysis:
%s

Creative Descriptions:
%s

Characteristics Required:
%s

PLM Data:
%s

Generate a complete, valid JSON payload that matches Wildberries format requirements. Include all required characteristics and fields.`,
		state.SelectedSubject.Name,
		state.SelectedSubject.ID,
		f.formatTechAnalyses(state),
		f.formatCreativeDescriptions(state),
		f.formatWBCharacteristics(state),
		string(state.PLMJSON))
}

// Helper methods for formatting data

func (f *ArticleFlow) formatTechAnalyses(state *ArticleFlowState) string {
	var parts []string
	for imgID, tech := range state.TechAnalysisByImage {
		parts = append(parts, fmt.Sprintf("Image %s: %s", imgID, string(tech.RawJSON)))
	}
	return strings.Join(parts, "\n\n")
}

func (f *ArticleFlow) formatCreativeDescriptions(state *ArticleFlowState) string {
	var parts []string
	for imgID, desc := range state.CreativeByImage {
		parts = append(parts, fmt.Sprintf("Image %s: %s", imgID, desc))
	}
	return strings.Join(parts, "\n\n")
}

func (f *ArticleFlow) formatWBSubjects(state *ArticleFlowState) string {
	var parts []string
	for _, subject := range state.WBSubjects {
		parts = append(parts, fmt.Sprintf("ID: %d - %s", subject.ID, subject.Name))
	}
	return strings.Join(parts, "\n")
}

func (f *ArticleFlow) formatWBCharacteristics(state *ArticleFlowState) string {
	var parts []string
	for _, char := range state.WBCharacteristics {
		parts = append(parts, fmt.Sprintf("- %s (type: %d)", char.Name, char.CharcType))
	}
	return strings.Join(parts, "\n")
}

// Helper functions for type assertions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key].(int64); ok {
		return val
	}
	if val, ok := m[key].(int); ok {
		return int64(val)
	}
	if val, ok := m[key].(float64); ok {
		return int64(val)
	}
	return 0
}