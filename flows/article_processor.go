package flows

import (
	"context"
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ArticleProcessor is a specialized flow for processing fashion articles
// It handles article import, analysis, and preparation for marketplace integration
type ArticleProcessor struct {
	*base.PonchoBaseFlow
	modelRegistry interfaces.PonchoModelRegistry
	toolRegistry  interfaces.PonchoToolRegistry
	articleImporter interfaces.PonchoTool
	s3Storage     interfaces.PonchoTool
}

// NewArticleProcessor creates a new article processor flow
func NewArticleProcessor(
	modelRegistry interfaces.PonchoModelRegistry,
	toolRegistry interfaces.PonchoToolRegistry,
	logger interfaces.Logger,
) *ArticleProcessor {
	flow := base.NewPonchoBaseFlow(
		"article_processor",
		"Processes fashion articles for marketplace integration",
		"1.0.0",
		"processing",
	)

	// Set up input schema
	inputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"article_data": map[string]interface{}{
				"type":        "object",
				"description": "Raw article data from supplier or manufacturer",
			},
			"processing_options": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"generate_description": map[string]interface{}{
						"type":    "boolean",
						"default": true,
					},
					"analyze_images": map[string]interface{}{
						"type":    "boolean",
						"default": true,
					},
					"categorize": map[string]interface{}{
						"type":    "boolean",
						"default": true,
					},
					"validate_data": map[string]interface{}{
						"type":    "boolean",
						"default": true,
					},
				},
			},
		},
		"required": []string{"article_data"},
	}
	flow.SetInputSchema(inputSchema)

	// Set up output schema
	outputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"processed_article": map[string]interface{}{
				"type":        "object",
				"description": "Processed article ready for marketplace",
			},
			"processing_metadata": map[string]interface{}{
				"type":        "object",
				"description": "Metadata about processing steps and results",
			},
			"warnings": map[string]interface{}{
				"type":        "array",
				"description": "Any warnings during processing",
			},
			"errors": map[string]interface{}{
				"type":        "array",
				"description": "Any errors during processing",
			},
		},
	}
	flow.SetOutputSchema(outputSchema)

	// Add tags
	flow.AddTag("processing")
	flow.AddTag("article")
	flow.AddTag("wildberries")
	flow.AddTag("import")

	// Set dependencies
	flow.AddDependency("deepseek-chat")      // For text processing
	flow.AddDependency("glm-4.6v-flash")     // For image analysis
	flow.AddDependency("article_importer")   // For import operations
	flow.AddDependency("s3_storage")         // For image storage

	processor := &ArticleProcessor{
		PonchoBaseFlow: flow,
		modelRegistry:  modelRegistry,
		toolRegistry:   toolRegistry,
	}

	if logger != nil {
		processor.SetLogger(logger)
	}

	return processor
}

// Execute executes the article processing flow
func (ap *ArticleProcessor) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	startTime := time.Now()

	// Validate and parse input
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("input must be a map[string]interface{}")
	}

	articleData, ok := inputMap["article_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("article_data is required and must be an object")
	}

	processingOptions, _ := inputMap["processing_options"].(map[string]interface{})
	if processingOptions == nil {
		processingOptions = make(map[string]interface{})
	}

	ap.GetLogger().Info("Starting article processing",
		"article_id", articleData["id"],
		"options", processingOptions,
	)

	processedArticle := make(map[string]interface{})
	var warnings []interface{}
	var errors []interface{}

	// Step 1: Validate article data
	if shouldValidate(processingOptions) {
		ap.GetLogger().Debug("Validating article data")
		if err := ap.validateArticleData(articleData); err != nil {
			errors = append(errors, map[string]interface{}{
				"step":  "validation",
				"error": err.Error(),
			})
		}
	}

	// Step 2: Process images
	if shouldAnalyzeImages(processingOptions) {
		ap.GetLogger().Debug("Processing article images")
		images, imageWarnings := ap.processImages(ctx, articleData)
		if len(images) > 0 {
			processedArticle["images"] = images
		}
		warnings = append(warnings, imageWarnings...)
	}

	// Step 3: Generate description
	if shouldGenerateDescription(processingOptions) {
		ap.GetLogger().Debug("Generating article description")
		description, err := ap.generateDescription(ctx, articleData)
		if err != nil {
			errors = append(errors, map[string]interface{}{
				"step":  "description_generation",
				"error": err.Error(),
			})
		} else {
			processedArticle["generated_description"] = description
		}
	}

	// Step 4: Categorize article
	if shouldCategorize(processingOptions) {
		ap.GetLogger().Debug("Categorizing article")
		categories, err := ap.categorizeArticle(ctx, articleData)
		if err != nil {
			errors = append(errors, map[string]interface{}{
				"step":  "categorization",
				"error": err.Error(),
			})
		} else {
			processedArticle["categories"] = categories
		}
	}

	// Copy original article data
	for k, v := range articleData {
		if _, exists := processedArticle[k]; !exists {
			processedArticle[k] = v
		}
	}

	// Prepare metadata
	duration := time.Since(startTime)
	metadata := map[string]interface{}{
		"processing_time_ms": duration.Milliseconds(),
		"steps_performed":    ap.getStepsPerformed(processingOptions),
		"images_processed":   ap.countImages(articleData),
		"status":             ap.determineStatus(len(errors), len(warnings)),
	}

	output := map[string]interface{}{
		"processed_article":   processedArticle,
		"processing_metadata": metadata,
		"warnings":           warnings,
		"errors":             errors,
	}

	ap.GetLogger().Info("Article processing completed",
		"duration_ms", duration.Milliseconds(),
		"status", metadata["status"],
	)

	return output, nil
}

// validateArticleData validates the article data structure
func (ap *ArticleProcessor) validateArticleData(articleData map[string]interface{}) error {
	// Check required fields
	requiredFields := []string{"id", "title", "vendor_code"}
	for _, field := range requiredFields {
		if _, exists := articleData[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate data types
	if _, ok := articleData["id"].(string); !ok {
		return fmt.Errorf("article id must be a string")
	}

	if _, ok := articleData["title"].(string); !ok {
		return fmt.Errorf("article title must be a string")
	}

	return nil
}

// processImages processes article images using vision models
func (ap *ArticleProcessor) processImages(ctx context.Context, articleData map[string]interface{}) ([]interface{}, []interface{}) {
	var images []interface{}
	var warnings []interface{}

	imagesData, exists := articleData["images"]
	if !exists {
		return images, warnings
	}

	imagesSlice, ok := imagesData.([]interface{})
	if !ok {
		warnings = append(warnings, "images field is not an array")
		return images, warnings
	}

	// Get vision model
	model, err := ap.modelRegistry.Get("glm-4.6v-flash")
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to get vision model: %v", err))
		return images, warnings
	}

	for i, imageData := range imagesSlice {
		imageMap, ok := imageData.(map[string]interface{})
		if !ok {
			warnings = append(warnings, fmt.Sprintf("image %d is not an object", i))
			continue
		}

		imageURL, ok := imageMap["url"].(string)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("image %d has no url", i))
			continue
		}

		// Analyze image using vision model
		maxTokens := 500
		temperature := float32(0.3)

		req := &interfaces.PonchoModelRequest{
			Model: "glm-4.6v-flash",
			Messages: []*interfaces.PonchoMessage{
				{
					Role: interfaces.PonchoRoleUser,
					Content: []*interfaces.PonchoContentPart{
						{
							Type: interfaces.PonchoContentTypeText,
							Text: fmt.Sprintf("Analyze this fashion image: %s. Describe the item, style, materials, and visible features.", imageURL),
						},
					},
				},
			},
			MaxTokens:   &maxTokens,
			Temperature: &temperature,
		}

		resp, err := model.Generate(ctx, req)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to analyze image %d: %v", i, err))
			continue
		}

		// Add analysis to image data
		if resp.Message != nil && len(resp.Message.Content) > 0 {
			imageMap["ai_analysis"] = resp.Message.Content[0].Text
		} else {
			warnings = append(warnings, fmt.Sprintf("no content in response for image %d", i))
		}
		images = append(images, imageMap)
	}

	return images, warnings
}

// generateDescription generates article description using AI
func (ap *ArticleProcessor) generateDescription(ctx context.Context, articleData map[string]interface{}) (string, error) {
	model, err := ap.modelRegistry.Get("deepseek-chat")
	if err != nil {
		return "", fmt.Errorf("failed to get text model: %w", err)
	}

	title, _ := articleData["title"].(string)
	description, _ := articleData["description"].(string)
	material, _ := articleData["material"].(string)

	prompt := fmt.Sprintf(`
	Сгенерируй привлекательное описание для товара модной индустрии.

	Название: %s
	Описание: %s
	Материал: %s

	Создай профессиональное описание для маркетплейса (Wildberries/Ozon), которое:
	- Привлекает внимание покупателей
	- Подчеркивает ключевые преимущества
	- Описывает материал и качество
	- Указывает на сезонность и стиль
	- Оптимизировано для поиска

	Ответ предоставь только текст описания без дополнительных комментариев.
	`, title, description, material)

	maxTokens := 1000
	temperature := float32(0.7)

	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: prompt,
					},
				},
			},
		},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	resp, err := model.Generate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate description: %w", err)
	}

	if resp.Message != nil && len(resp.Message.Content) > 0 {
		return resp.Message.Content[0].Text, nil
	}
	return "", fmt.Errorf("no content in response")
}

// categorizeArticle determines the best categories for the article
func (ap *ArticleProcessor) categorizeArticle(ctx context.Context, articleData map[string]interface{}) (map[string]interface{}, error) {
	model, err := ap.modelRegistry.Get("deepseek-chat")
	if err != nil {
		return nil, fmt.Errorf("failed to get text model: %w", err)
	}

	title, _ := articleData["title"].(string)
	description, _ := articleData["description"].(string)

	prompt := fmt.Sprintf(`
	Определи категории для товара модной индустрии.

	Название: %s
	Описание: %s

	Верни JSON с категориями в формате:
	{
		"category": "основная категория (одежда, обувь, аксессуары)",
		"subcategory": "подкатегория (платья, рубашки, кроссовки)",
		"gender": "пол (мужской, женский, унисекс)",
		"season": "сезон (лето, зима, демисезон, круглогодичный)",
		"style": "стиль (casual, business, sport, evening)"
	}

	Ответ предоставь только JSON без дополнительных комментариев.
	`, title, description)

	maxTokens := 500
	temperature := float32(0.3)

	req := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: prompt,
					},
				},
			},
		},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	resp, err := model.Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to categorize article: %w", err)
	}

	// For now, return a simple map with the raw response
	// In production, you'd parse JSON properly
	if resp.Message != nil && len(resp.Message.Content) > 0 {
		return map[string]interface{}{
			"raw_analysis": resp.Message.Content[0].Text,
		}, nil
	}
	return nil, fmt.Errorf("no content in response")
}

// Helper functions
func shouldValidate(options map[string]interface{}) bool {
	if val, ok := options["validate_data"].(bool); ok {
		return val
	}
	return true // Default behavior
}

func shouldAnalyzeImages(options map[string]interface{}) bool {
	if val, ok := options["analyze_images"].(bool); ok {
		return val
	}
	return true // Default behavior
}

func shouldGenerateDescription(options map[string]interface{}) bool {
	if val, ok := options["generate_description"].(bool); ok {
		return val
	}
	return true // Default behavior
}

func shouldCategorize(options map[string]interface{}) bool {
	if val, ok := options["categorize"].(bool); ok {
		return val
	}
	return true // Default behavior
}

func (ap *ArticleProcessor) getStepsPerformed(options map[string]interface{}) []string {
	var steps []string
	if shouldValidate(options) {
		steps = append(steps, "validation")
	}
	if shouldAnalyzeImages(options) {
		steps = append(steps, "image_analysis")
	}
	if shouldGenerateDescription(options) {
		steps = append(steps, "description_generation")
	}
	if shouldCategorize(options) {
		steps = append(steps, "categorization")
	}
	return steps
}

func (ap *ArticleProcessor) countImages(articleData map[string]interface{}) int {
	images, exists := articleData["images"]
	if !exists {
		return 0
	}
	if slice, ok := images.([]interface{}); ok {
		return len(slice)
	}
	return 0
}

func (ap *ArticleProcessor) determineStatus(errorCount, warningCount int) string {
	if errorCount > 0 {
		return "failed"
	}
	if warningCount > 0 {
		return "completed_with_warnings"
	}
	return "success"
}

// ExecuteStreaming implements streaming execution (placeholder)
func (ap *ArticleProcessor) ExecuteStreaming(
	ctx context.Context,
	input interface{},
	callback interfaces.PonchoStreamCallback,
) error {
	// For now, just execute non-streaming and send result as single message
	result, err := ap.Execute(ctx, input)
	if err != nil {
		return err
	}

	resultJSON := fmt.Sprintf("%v", result)
	callback(&interfaces.PonchoStreamChunk{
		Delta: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: resultJSON,
				},
			},
		},
		Done: true,
		Metadata: map[string]interface{}{
			"flow": ap.Name(),
		},
	})

	return nil
}