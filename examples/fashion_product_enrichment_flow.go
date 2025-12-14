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

// FashionProductEnrichmentFlow demonstrates the new v2 flow architecture
// This example shows how the FlowContext and FlowBuilder solve the original problem
type FashionProductEnrichmentFlow struct {
	flow interfaces.PonchoFlowV2
	mediaPipeline *media.MediaPipeline
}

// NewFashionProductEnrichmentFlow creates the fashion product enrichment flow
func NewFashionProductEnrichmentFlow(
	articleImporter interfaces.PonchoTool,
	visionModel interfaces.PonchoModel,
	categoryTool interfaces.PonchoTool,
	textModel interfaces.PonchoModel,
	s3Tool interfaces.PonchoTool,
) (*FashionProductEnrichmentFlow, error) {

	// Create media pipeline for image processing
	mediaPipeline := media.NewMediaPipeline(nil)

	// Build the flow using the DSL
	flowBuilder := flow.NewFlowBuilder("fashion_product_enrichment").
		Description("Enriches fashion products with AI analysis and categorization").
		Version("2.0.0").
		Category("fashion").
		RequiresVision().
		RequiresTool("article_importer").
		RequiresTool("category_classifier").
		RequiresTool("s3_storage").
		RequiresModel("vision_model").
		RequiresModel("text_model").
		Timeout(10 * time.Minute)

	// Step 1: Import product data and images from S3
	flowBuilder.
		Step("import_data").
		Tool(articleImporter, "s3_path").
		Input("s3_path").
		Output("product_data").
		Timeout(30 * time.Second).
		Continue()

	// Step 2: Extract images for vision analysis (custom step)
	flowBuilder.
		Step("extract_images").
		Custom(extractImagesFromProductData).
		Requires("product_data").
		Provides("images_for_analysis").
		Timeout(10 * time.Second).
		Continue();

	// Step 3: Parallel vision analysis of all images
	flowBuilder.
		Step("analyze_images").
		Parallel().
		MaxConcurrency(3).
		AddSubStep(&flow.CustomStep{
			name: "analyze_vision",
			executor: func(ctx context.Context, flowCtx interfaces.FlowContext) error {
				return analyzeImagesWithVision(ctx, flowCtx, visionModel, mediaPipeline)
			},
		}).
		Requires("images_for_analysis").
		Provides("visual_features").
		Timeout(60 * time.Second).
		Continue();

	// Step 4: Classify product category
	flowBuilder.
		Step("classify_category").
		Tool(categoryTool, "product_data.name").
		Input("product_data.name").
		Output("category_id").
		Timeout(15 * time.Second).
		Continue();

	// Step 5: Generate SEO description (requires data from steps 1, 3, 4)
	flowBuilder.
		Step("generate_description").
		Model(textModel, "description_prompt").
		Input("description_prompt").
		Inputs("product_data", "visual_features", "category_id").
		Output("seo_description").
		Temperature(0.7).
		MaxTokens(1000).
		Timeout(30 * time.Second).
		Continue();

	// Step 6: Save enhanced data back to S3
	flowBuilder.
		Step("save_results").
		Tool(s3Tool, "enhanced_product_data").
		Input("enhanced_product_data").
		Timeout(30 * time.Second).
		CanFail(true). // Don't fail the entire flow if save fails
		Continue();

	// Build the flow
	builtFlow, err := flowBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build fashion enrichment flow: %w", err)
	}

	return &FashionProductEnrichmentFlow{
		flow: builtFlow,
		mediaPipeline: mediaPipeline,
	}, nil
}

// Execute runs the fashion product enrichment flow
func (fpef *FashionProductEnrichmentFlow) Execute(ctx context.Context, s3Path string) (*EnrichmentResult, error) {
	// Create flow context
	flowCtx := context.NewBaseFlowContext()

	// Set initial input
	if err := flowCtx.Set("s3_path", s3Path); err != nil {
		return nil, fmt.Errorf("failed to set initial input: %w", err)
	}

	// Execute the flow
	_, err := fpef.flow.Execute(ctx, s3Path, flowCtx)
	if err != nil {
		return nil, fmt.Errorf("flow execution failed: %w", err)
	}

	// Collect results from context
	result := &EnrichmentResult{
		S3Path: s3Path,
		Success: true,
	}

	// Extract all results from context
	if productData, has := flowCtx.Get("product_data"); has {
		result.ProductData = productData
	}

	if visualFeatures, has := flowCtx.Get("visual_features"); has {
		result.VisualFeatures = visualFeatures
	}

	if categoryID, has := flowCtx.GetString("category_id"); has == nil {
		result.CategoryID = categoryID
	}

	if seoDescription, has := flowCtx.GetString("seo_description"); has == nil {
		result.SEODescription = seoDescription
	}

	return result, nil
}

// EnrichmentResult represents the output of the fashion enrichment flow
type EnrichmentResult struct {
	S3Path         string      `json:"s3_path"`
	ProductData    interface{} `json:"product_data"`
	VisualFeatures interface{} `json:"visual_features"`
	CategoryID     string      `json:"category_id"`
	SEODescription string      `json:"seo_description"`
	Success        bool        `json:"success"`
	Error          string      `json:"error,omitempty"`
	Timestamp      time.Time   `json:"timestamp"`
}

// Custom step implementations

// extractImagesFromProductData extracts images from imported product data
func extractImagesFromProductData(ctx context.Context, flowCtx interfaces.FlowContext) error {
	productData, has := flowCtx.Get("product_data")
	if !has {
		return fmt.Errorf("product_data not found in context")
	}

	// Extract images from product data (implementation depends on your data structure)
	// This is a simplified example
	if productMap, ok := productData.(map[string]interface{}); ok {
		if images, has := productMap["images"]; has {
			// Convert to []byte format for vision model
			if imageList, ok := images.([]interface{}); ok {
				var processedImages []*media.MediaData
				for i, img := range imageList {
					// Convert image to MediaData
					if imgBytes, ok := img.([]byte); ok {
						mediaData, err := media.NewMediaDataFromBytes(imgBytes, "image/jpeg")
						if err != nil {
							return fmt.Errorf("failed to process image %d: %w", i, err)
						}
						processedImages = append(processedImages, mediaData)
					}
				}

				// Store in context for vision analysis
				if err := flowCtx.SetMedia("images_for_analysis", processedImages[0]); err != nil {
					return fmt.Errorf("failed to store images in context: %w", err)
				}
			}
		}
	}

	return nil
}

// analyzeImagesWithVision performs vision analysis using the vision model
func analyzeImagesWithVision(
	ctx context.Context,
	flowCtx interfaces.FlowContext,
	visionModel interfaces.PonchoModel,
	mediaPipeline *media.MediaPipeline,
) error {
	// Get images from context
	images, err := flowCtx.GetMedia("images_for_analysis")
	if err != nil {
		return fmt.Errorf("images not found in context: %w", err)
	}

	// Prepare media for vision model
	mediaList := []*media.MediaData{images}
	preparedMedia, err := mediaPipeline.PrepareForModel(ctx, mediaList, visionModel)
	if err != nil {
		return fmt.Errorf("failed to prepare media for vision model: %w", err)
	}

	// Create vision analysis request
	prompt := "Analyze this fashion item and describe its visual features, style, materials, and characteristics."

	message := &interfaces.PonchoMessage{
		Role: interfaces.PonchoRoleUser,
		Content: []*interfaces.PonchoContentPart{
			{
				Type: interfaces.PonchoContentTypeText,
				Text: prompt,
			},
		},
	}

	// Add media to message
	for _, mediaData := range preparedMedia {
		message.Content = append(message.Content, &interfaces.PonchoContentPart{
			Type: interfaces.PonchoContentTypeMedia,
			Media: &interfaces.PonchoMediaPart{
				URL:      mediaData.GetDataURL(),
				MimeType: mediaData.MimeType,
			},
		})
	}

	// Execute vision model
	maxTokens := 500
	temperature := float32(0.3)

	req := &interfaces.PonchoModelRequest{
		Model: visionModel.Name(),
		Messages: []*interfaces.PonchoMessage{message},
		MaxTokens: &maxTokens,
		Temperature: &temperature,
	}

	resp, err := visionModel.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("vision model execution failed: %w", err)
	}

	// Store results in context
	if resp != nil && resp.Message != nil && len(resp.Message.Content) > 0 {
		visualAnalysis := resp.Message.Content[0].Text
		if err := flowCtx.Set("visual_features", visualAnalysis); err != nil {
			return fmt.Errorf("failed to store visual features: %w", err)
		}
	}

	return nil
}

// Example usage
func ExampleFashionProductEnrichmentFlow() {
	// This would be initialized with actual implementations
	var articleImporter interfaces.PonchoTool
	var visionModel interfaces.PonchoModel
	var categoryTool interfaces.PonchoTool
	var textModel interfaces.PonchoModel
	var s3Tool interfaces.PonchoTool

	// Create the flow
	flow, err := NewFashionProductEnrichmentFlow(
		articleImporter,
		visionModel,
		categoryTool,
		textModel,
		s3Tool,
	)
	if err != nil {
		fmt.Printf("Failed to create flow: %v\n", err)
		return
	}

	// Execute the flow
	ctx := context.Background()
	result, err := flow.Execute(ctx, "s3://fashion-products/product-123.json")
	if err != nil {
		fmt.Printf("Flow execution failed: %v\n", err)
		return
	}

	fmt.Printf("Flow completed successfully!\n")
	fmt.Printf("Product: %+v\n", result.ProductData)
	fmt.Printf("Visual Features: %s\n", result.VisualFeatures)
	fmt.Printf("Category: %s\n", result.CategoryID)
	fmt.Printf("SEO Description: %s\n", result.SEODescription)
}

// Comparison with old approach:

/*
OLD APPROACH (without FlowContext):

func (f *FashionFlow) Execute(input interface{}) (interface{}, error) {
    // Step 1
    importResult := f.articleImporter.Execute(input)
    product := importResult.(map[string]interface{})["product"]
    images := importResult.(map[string]interface{})["images"]

    // Step 2 - Must pass images explicitly
    visionResults := make([]string, len(images))
    for i, img := range images {
        result := f.visionModel.Execute(img)  // Manual handling
        visionResults[i] = result
    }

    // Step 3 - Must pass product name explicitly
    category := f.categoryTool.Execute(product["name"])

    // Step 4 - Must combine all previous results manually
    combinedData := map[string]interface{}{
        "product": product,
        "features": visionResults,
        "category": category,
    }
    description := f.textModel.Execute(combinedData)

    return description
}

NEW APPROACH (with FlowContext and FlowBuilder):

flow := flow.NewFlowBuilder("fashion_enrichment").
    Step("import").Tool(articleImporter, "s3_path").Output("product_data").Continue().
    Step("analyze").Model(visionModel, "prompt").Inputs("product_data.images").Output("visual_features").Continue().
    Step("classify").Tool(categoryTool, "product_data.name").Output("category").Continue().
    Step("generate").Model(textModel, "prompt").Inputs("product_data", "visual_features", "category").Output("description").Continue().
    Build()

// All state management is automatic!
// Each step reads/writes to shared context.
// No manual parameter passing between steps.
*/