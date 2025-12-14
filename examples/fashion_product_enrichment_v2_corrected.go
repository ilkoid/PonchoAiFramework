package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/context"
	"github.com/ilkoid/PonchoAiFramework/core/flow"
	"github.com/ilkoid/PonchoAiFramework/core/media"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// FashionProductEnrichmentV2 показывает правильную реализацию с type-safe getters и lazy loading
type FashionProductEnrichmentV2 struct {
	flow         interfaces.PonchoFlowV2
	mediaHelper  *media.MediaHelperV2
}

// NewFashionProductEnrichmentV2 создает corrected flow
func NewFashionProductEnrichmentV2(
	articleImporter interfaces.PonchoTool,
	visionModel interfaces.PonchoModel,
	categoryTool interfaces.PonchoTool,
	textModel interfaces.PonchoModel,
	s3Tool interfaces.PonchoTool,
) (*FashionProductEnrichmentV2, error) {

	// Media pipeline с поддержкой lazy loading
	mediaPipeline := media.NewMediaPipeline(nil)
	mediaHelper := media.NewMediaHelperV2(mediaPipeline, nil)

	// FlowBuilder с type-safe operations
	flowBuilder := flow.NewFlowBuilder("fashion_product_enrichment_v2").
		Description("Fashion product enrichment with type-safe operations and lazy loading").
		Version("2.1.0").
		Category("fashion").
		RequiresVision().
		Timeout(15 * time.Minute).
		EnableParallel(true).
		MaxConcurrency(3)

	// Step 1: Import product data (type-safe)
	flowBuilder.
		Step("import_product_data").
		Custom(func(ctx context.Context, flowCtx interfaces.FlowContext) error {
			return importProductDataV2(ctx, flowCtx, articleImporter, mediaHelper)
		}).
		Requires("s3_path").
		Provides("product_data", "product_images").
		Timeout(60 * time.Second).
		Continue()

	// Step 2: Parallel image analysis with lazy loading
	flowBuilder.
		Step("analyze_images_parallel").
		Parallel().
		MaxConcurrency(5).
		FailFast(false).
		AddSubStep(&flow.CustomStep{
			name: "analyze_single_image",
			executor: func(ctx context.Context, flowCtx interfaces.FlowContext) error {
				return analyzeImagesV2(ctx, flowCtx, visionModel, mediaHelper)
			},
		}).
		Requires("product_images").
		Provides("visual_features_list").
		Timeout(180 * time.Second).
		Continue()

	// Step 3: Product categorization with type-safe data access
	flowBuilder.
		Step("categorize_product").
		Custom(func(ctx context.Context, flowCtx interfaces.FlowContext) error {
			return categorizeProductV2(ctx, flowCtx, categoryTool)
		}).
		Requires("product_data").
		Provides("category_id", "category_info").
		Timeout(30 * time.Second).
		Continue()

	// Step 4: SEO description generation using ALL previous data
	flowBuilder.
		Step("generate_seo_description").
		Custom(func(ctx context.Context, flowCtx interfaces.FlowContext) error {
			return generateSEODescriptionV2(ctx, flowCtx, textModel, mediaHelper)
		}).
		Requires("product_data", "visual_features_list", "category_id").
		Provides("seo_description", "seo_metadata").
		Timeout(60 * time.Second).
		Continue()

	// Step 5: Save results back to S3
	flowBuilder.
		Step("save_enriched_data").
		Custom(func(ctx context.Context, flowCtx interfaces.FlowContext) error {
			return saveEnrichedDataV2(ctx, flowCtx, s3Tool)
		}).
		Requires("product_data", "visual_features_list", "category_id", "seo_description").
		Timeout(60 * time.Second).
		CanFail(true).
		Continue()

	// Build flow
	builtFlow, err := flowBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build fashion enrichment flow v2: %w", err)
	}

	return &FashionProductEnrichmentV2{
		flow:        builtFlow,
		mediaHelper: mediaHelper,
	}, nil
}

// Execute с type-safe result extraction
func (fpev2 *FashionProductEnrichmentV2) Execute(ctx context.Context, s3Path string) (*EnrichmentResultV2, error) {
	// Create type-safe FlowContext
	flowCtx := context.NewBaseFlowContextV2()

	// Set initial input
	if err := flowCtx.Set("s3_path", s3Path); err != nil {
		return nil, fmt.Errorf("failed to set s3_path: %w", err)
	}

	// Execute flow
	_, err := fpev2.flow.Execute(ctx, s3Path, flowCtx)
	if err != nil {
		return nil, fmt.Errorf("flow execution failed: %w", err)
	}

	// Type-safe result extraction
	result := &EnrichmentResultV2{
		S3Path:       s3Path,
		Success:      true,
		ExecutedAt:   time.Now(),
		MemoryUsage:  flowCtx.GetMemoryUsage(),
	}

	// Type-safe getters с error handling
	if productData, err := flowCtx.GetMap("product_data"); err == nil {
		result.ProductData = productData
	}

	if visualFeatures, err := flowCtx.GetStringSlice("visual_features_list"); err == nil {
		result.VisualFeatures = visualFeatures
	}

	if categoryID, err := flowCtx.GetString("category_id"); err == nil {
		result.CategoryID = categoryID
	}

	if seoDescription, err := flowCtx.GetString("seo_description"); err == nil {
		result.SEODescription = seoDescription
	}

	if price, err := flowCtx.GetFloat64("price"); err == nil {
		result.Price = price
	}

	if inStock, err := flowCtx.GetBool("in_stock"); err == nil {
		result.InStock = inStock
	}

	return result, nil
}

// Step implementations with type-safe operations

func importProductDataV2(
	ctx context.Context,
	flowCtx interfaces.FlowContext,
	importer interfaces.PonchoTool,
	mediaHelper *media.MediaHelperV2,
) error {
	// Type-safe get s3_path
	s3Path, err := flowCtx.GetString("s3_path")
	if err != nil {
		return fmt.Errorf("s3_path required: %w", err)
	}

	// Execute tool
	result, err := importer.Execute(ctx, s3Path)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	// Type-safe store product data
	if err := flowCtx.Set("product_data", result); err != nil {
		return fmt.Errorf("failed to store product data: %w", err)
	}

	// Extract images and store as ImageReference (lazy loading)
	if err := mediaHelper.ExtractImagesFromContext(flowCtx, result); err != nil {
		return fmt.Errorf("failed to extract images: %w", err)
	}

	return nil
}

func analyzeImagesV2(
	ctx context.Context,
	flowCtx interfaces.FlowContext,
	visionModel interfaces.PonchoModel,
	mediaHelper *media.MediaHelperV2,
) error {
	// Create type-safe flow context
	ctxV2, ok := flowCtx.(*context.BaseFlowContextV2)
	if !ok {
		return fmt.Errorf("flow context v2 required")
	}

	// Get all image keys from context
	var imageKeys []string
	for _, key := range flowCtx.Keys() {
		if (strings.Contains(key, "image") || strings.Contains(key, "photo")) &&
		   !strings.Contains(key, "_analysis") {
			imageKeys = append(imageKeys, key)
		}
	}

	if len(imageKeys) == 0 {
		return fmt.Errorf("no images found in context")
	}

	// Create vision message with lazy loading
	prompt := "Analyze this fashion image. Describe the style, materials, colors, and key features."

	message, err := mediaHelper.CreateVisionMessageV2(
		ctx,
		prompt,
		flowCtx,
		imageKeys,
		visionModel,
	)
	if err != nil {
		return fmt.Errorf("failed to create vision message: %w", err)
	}

	// Execute vision model
	maxTokens := 500
	temperature := float32(0.3)

	req := &interfaces.PonchoModelRequest{
		Model:    visionModel.Name(),
		Messages: []*interfaces.PonchoMessage{message},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	resp, err := visionModel.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("vision model failed: %w", err)
	}

	// Extract and store results
	if resp != nil && resp.Message != nil && len(resp.Message.Content) > 0 {
		analysis := resp.Message.Content[0].Text

		// Store analysis results
		if err := flowCtx.Set("latest_image_analysis", analysis); err != nil {
			return fmt.Errorf("failed to store analysis: %w", err)
		}

		// Accumulate in features list
		var features []string
		if existingFeatures, err := flowCtx.GetStringSlice("visual_features_list"); err == nil {
			features = existingFeatures
		}
		features = append(features, analysis)

		if err := flowCtx.Set("visual_features_list", features); err != nil {
			return fmt.Errorf("failed to store features list: %w", err)
		}
	}

	return nil
}

func categorizeProductV2(
	ctx context.Context,
	flowCtx interfaces.FlowContext,
	categoryTool interfaces.PonchoTool,
) error {
	// Type-safe get product data
	productData, err := flowCtx.GetMap("product_data")
	if err != nil {
		return fmt.Errorf("product_data required: %w", err)
	}

	// Extract product name safely
	productName := ""
	if nameValue, has := productData["name"]; has {
		if name, ok := nameValue.(string); ok {
			productName = name
		}
	}

	if productName == "" {
		return fmt.Errorf("product name not found in product_data")
	}

	// Execute category tool
	result, err := categoryTool.Execute(ctx, productName)
	if err != nil {
		return fmt.Errorf("categorization failed: %w", err)
	}

	// Type-safe store results
	if err := flowCtx.Set("category_result", result); err != nil {
		return fmt.Errorf("failed to store category result: %w", err)
	}

	// Extract category ID if available
	if resultMap, ok := result.(map[string]interface{}); ok {
		if categoryID, has := resultMap["category_id"]; has {
			if id, ok := categoryID.(string); ok {
				if err := flowCtx.SetString("category_id", id); err != nil {
					return fmt.Errorf("failed to store category_id: %w", err)
				}
			}
		}
	}

	return nil
}

func generateSEODescriptionV2(
	ctx context.Context,
	flowCtx interfaces.FlowContext,
	textModel interfaces.PonchoModel,
	mediaHelper *media.MediaHelperV2,
) error {
	// Type-safe get all required data
	productData, err := flowCtx.GetMap("product_data")
	if err != nil {
		return fmt.Errorf("product_data required: %w", err)
	}

	visualFeatures, err := flowCtx.GetStringSlice("visual_features_list")
	if err != nil {
		return fmt.Errorf("visual_features_list required: %w", err)
	}

	categoryID, err := flowCtx.GetString("category_id")
	if err != nil {
		return fmt.Errorf("category_id required: %w", err)
	}

	// Build comprehensive prompt using all data
	prompt := buildSEOProductPromptV2(productData, visualFeatures, categoryID)

	// Execute text model
	maxTokens := 1000
	temperature := float32(0.7)

	req := &interfaces.PonchoModelRequest{
		Model: textModel.Name(),
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

	resp, err := textModel.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("description generation failed: %w", err)
	}

	// Store results
	if resp != nil && resp.Message != nil && len(resp.Message.Content) > 0 {
		description := resp.Message.Content[0].Text

		if err := flowCtx.SetString("seo_description", description); err != nil {
			return fmt.Errorf("failed to store seo description: %w", err)
		}

		// Store metadata
		metadata := map[string]interface{}{
			"generated_at": time.Now(),
			"model":        textModel.Name(),
			"word_count":   len(strings.Fields(description)),
			"visual_features_count": len(visualFeatures),
		}

		if err := flowCtx.Set("seo_metadata", metadata); err != nil {
			return fmt.Errorf("failed to store seo metadata: %w", err)
		}
	}

	return nil
}

func saveEnrichedDataV2(
	ctx context.Context,
	flowCtx interfaces.FlowContext,
	s3Tool interfaces.PonchoTool,
) error {
	// Create enriched product data
	enrichedData := map[string]interface{}{
		"enriched_at": time.Now(),
		"version":     "2.0",
	}

	// Type-safe collection of all data
	keys := []string{"product_data", "visual_features_list", "category_id", "seo_description", "seo_metadata"}

	for _, key := range keys {
		if value, has := flowCtx.Get(key); has {
			enrichedData[key] = value
		}
	}

	// Save to S3
	_, err := s3Tool.Execute(ctx, enrichedData)
	if err != nil {
		return fmt.Errorf("save to S3 failed: %w", err)
	}

	return nil
}

// EnrichmentResultV2 представляет type-safe результат
type EnrichmentResultV2 struct {
	S3Path         string                 `json:"s3_path"`
	ProductData    map[string]interface{} `json:"product_data,omitempty"`
	VisualFeatures []string               `json:"visual_features,omitempty"`
	CategoryID     string                 `json:"category_id,omitempty"`
	SEODescription string                 `json:"seo_description,omitempty"`
	Price          float64                `json:"price,omitempty"`
	InStock        bool                   `json:"in_stock,omitempty"`
	Success        bool                   `json:"success"`
	ExecutedAt     time.Time              `json:"executed_at"`
	MemoryUsage    int64                  `json:"memory_usage,omitempty"`
}

// Helper function
func buildSEOProductPromptV2(productData map[string]interface{}, visualFeatures []string, categoryID string) string {
	var prompt strings.Builder

	prompt.WriteString("Generate an SEO-optimized product description for a fashion item.\n\n")

	prompt.WriteString("PRODUCT INFORMATION:\n")
	if name, has := productData["name"]; has {
		prompt.WriteString(fmt.Sprintf("- Name: %v\n", name))
	}
	if desc, has := productData["description"]; has {
		prompt.WriteString(fmt.Sprintf("- Description: %v\n", desc))
	}
	if material, has := productData["material"]; has {
		prompt.WriteString(fmt.Sprintf("- Material: %v\n", material))
	}
	if price, has := productData["price"]; has {
		prompt.WriteString(fmt.Sprintf("- Price: %v\n", price))
	}

	prompt.WriteString(fmt.Sprintf("\nVISUAL ANALYSIS:\n%s\n", strings.Join(visualFeatures, "\n")))
	prompt.WriteString(fmt.Sprintf("\nCATEGORY: %s\n", categoryID))

	prompt.WriteString("\nREQUIREMENTS:\n")
	prompt.WriteString("- Include visual features from the analysis\n")
	prompt.WriteString("- Mention the category for target audience\n")
	prompt.WriteString("- Optimize for fashion marketplace search\n")
	prompt.WriteString("- Be engaging but professional\n")
	prompt.WriteString("- Include relevant keywords naturally\n")

	prompt.WriteString("\nDescription:")

	return prompt.String()
}

// Example usage demonstrating type safety
func ExampleTypeSafety() {
	// Create context
	flowCtx := context.NewBaseFlowContextV2()

	// Type-safe operations
	flowCtx.SetString("product_name", "Summer Dress")
	flowCtx.SetFloat64("price", 99.99)
	flowCtx.SetBool("in_stock", true)

	// Safe extraction with error handling
	name, err := flowCtx.GetString("product_name")
	if err != nil {
		fmt.Printf("Error getting product name: %v\n", err)
		return
	}

	price, err := flowCtx.GetFloat64("price")
	if err != nil {
		fmt.Printf("Error getting price: %v\n", err)
		return
	}

	fmt.Printf("Product: %s, Price: $%.2f\n", name, price)

	// This will NOT panic, will return error:
	invalidPrice, err := flowCtx.GetInt("product_name")
	if err != nil {
		fmt.Printf("Expected error: %v\n", err) // "key 'product_name' exists but is not an int, got string"
	}
	fmt.Printf("Invalid price: %d (error: %v)\n", invalidPrice, err)
}

/*
КЛЮЧЕВЫЕ УЛУЧШЕНИЯ:

1. ✅ Type Safety:
   - GetString(key) (string, error) вместо interface{}.(string)
   - GetFloat64(key) (float64, error) вместо рискованных type assertions
   - Никаких runtime паник!

2. ✅ Memory Management:
   - ImageReference хранит только URL/путь, не binary данные
   - Lazy loading - байты загружаются только при использовании
   - ImageCollection с контролем памяти (100MB limit)
   - Автоматическая очистка старых кэшей

3. ✅ Performance:
   - Параллельная обработка изображений
   - Кэширование загруженных изображений
   - Batch операции для множественных медиа

4. ✅ Error Handling:
   - Все операции возвращают ошибки
   - Graceful degradation при ошибках
   - CanFail() для non-critical operations
*/