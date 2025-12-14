package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// FashionSketchAnalyzer is a specialized flow for analyzing fashion sketches
// It uses vision models to extract structured data from fashion sketches
// It provides detailed analysis including style, materials, and product characteristics
type FashionSketchAnalyzer struct {
	*base.PonchoBaseFlow
	modelRegistry interfaces.PonchoModelRegistry
	promptManager interfaces.PromptManager
}

// NewFashionSketchAnalyzer creates a new fashion sketch analyzer flow
func NewFashionSketchAnalyzer(
	modelRegistry interfaces.PonchoModelRegistry,
	promptManager interfaces.PromptManager,
	logger interfaces.Logger,
) *FashionSketchAnalyzer {
	flow := base.NewPonchoBaseFlow(
		"fashion_sketch_analyzer",
		"Analyzes fashion sketches and extracts structured product data",
		"1.0.0",
		"fashion",
	)

	// Set up input schema
	inputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"image_url": map[string]interface{}{
				"type":        "string",
				"description": "URL of the fashion sketch image",
			},
			"analysis_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"description", "creative", "both"},
				"default":     "description",
				"description": "Type of analysis to perform",
			},
			"style_context": map[string]interface{}{
				"type":        "string",
				"description": "Optional style context (casual, business, evening, etc.)",
			},
		},
		"required": []string{"image_url"},
	}
	flow.SetInputSchema(inputSchema)

	// Set up output schema
	outputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"analysis": map[string]interface{}{
				"type":        "object",
				"description": "Structured analysis of the fashion sketch",
			},
			"metadata": map[string]interface{}{
				"type":        "object",
				"description": "Processing metadata and timing",
			},
		},
	}
	flow.SetOutputSchema(outputSchema)

	// Add tags
	flow.AddTag("fashion")
	flow.AddTag("vision")
	flow.AddTag("analysis")
	flow.AddTag("wildberries")

	// Set dependencies
	flow.AddDependency("glm-4.6v-flash") // For image analysis
	flow.AddDependency("sketch_description.prompt")
	flow.AddDependency("sketch_creative.prompt")

	analyzer := &FashionSketchAnalyzer{
		PonchoBaseFlow: flow,
		modelRegistry:  modelRegistry,
		promptManager:  promptManager,
	}

	if logger != nil {
		analyzer.SetLogger(logger)
	}

	return analyzer
}

// Execute executes the fashion sketch analysis flow
func (f *FashionSketchAnalyzer) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	startTime := time.Now()

	// Validate and parse input
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("input must be a map[string]interface{}")
	}

	imageURL, ok := inputMap["image_url"].(string)
	if !ok {
		return nil, fmt.Errorf("image_url is required and must be a string")
	}

	analysisType, _ := inputMap["analysis_type"].(string)
	if analysisType == "" {
		analysisType = "description"
	}

	styleContext, _ := inputMap["style_context"].(string)

	f.GetLogger().Info("Starting fashion sketch analysis",
		"image_url", imageURL,
		"analysis_type", analysisType,
		"style_context", styleContext,
	)

	
	// Prepare the analysis request
	promptName := "sketch_description.prompt"
	if analysisType == "creative" {
		promptName = "sketch_creative.prompt"
	}

	// Prepare variables for the prompt
	variables := map[string]interface{}{
		"photoUrl": imageURL,
	}

	if styleContext != "" {
		variables["styleContext"] = styleContext
	}

	// Execute the prompt with the vision model
	resp, err := f.promptManager.ExecutePrompt(ctx, promptName, variables, "glm-4.6v-flash")
	if err != nil {
		return nil, fmt.Errorf("failed to generate analysis: %w", err)
	}

	// Parse the JSON response
	var analysis map[string]interface{}
	if resp.Message != nil && len(resp.Message.Content) > 0 {
		content := resp.Message.Content[0].Text
		if err := json.Unmarshal([]byte(content), &analysis); err != nil {
			// If JSON parsing fails, return the raw content
			analysis = map[string]interface{}{
				"raw_content": content,
				"parsing_error": err.Error(),
			}
		}
	} else {
		return nil, fmt.Errorf("no content in response")
	}

	// Prepare output with metadata
	duration := time.Since(startTime)
	output := map[string]interface{}{
		"analysis": analysis,
		"metadata": map[string]interface{}{
			"processing_time_ms": duration.Milliseconds(),
			"analysis_type":      analysisType,
			"model":             "glm-4.6v-flash",
			"prompt":            promptName,
			"image_url":         imageURL,
		},
	}

	f.GetLogger().Info("Fashion sketch analysis completed",
		"duration_ms", duration.Milliseconds(),
		"analysis_type", analysisType,
	)

	return output, nil
}

// ExecuteStreaming executes the flow with streaming support
func (f *FashionSketchAnalyzer) ExecuteStreaming(
	ctx context.Context,
	input interface{},
	callback interfaces.PonchoStreamCallback,
) error {
	startTime := time.Now()

	// Validate and parse input
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("input must be a map[string]interface{}")
	}

	imageURL, ok := inputMap["image_url"].(string)
	if !ok {
		return fmt.Errorf("image_url is required and must be a string")
	}

	analysisType, _ := inputMap["analysis_type"].(string)
	if analysisType == "" {
		analysisType = "description"
	}

	styleContext, _ := inputMap["style_context"].(string)

	f.GetLogger().Info("Starting streaming fashion sketch analysis",
		"image_url", imageURL,
		"analysis_type", analysisType,
		"style_context", styleContext,
	)

	// Send initial status
	callback(&interfaces.PonchoStreamChunk{
		Delta: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "Starting fashion sketch analysis...",
				},
			},
		},
		Done: false,
		Metadata: map[string]interface{}{"stage": "init"},
	})

	// Get the vision model
	callback(&interfaces.PonchoStreamChunk{
		Delta: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "Loading vision model...",
				},
			},
		},
		Done: false,
		Metadata: map[string]interface{}{"stage": "model_load"},
	})

	// Get the prompt template
	promptName := "sketch_description.prompt"
	if analysisType == "creative" {
		promptName = "sketch_creative.prompt"
	}

	// Prepare variables for the prompt
	variables := map[string]interface{}{
		"photoUrl": imageURL,
	}

	if styleContext != "" {
		variables["styleContext"] = styleContext
	}

	// Execute streaming analysis
	callback(&interfaces.PonchoStreamChunk{
		Delta: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: "Analyzing fashion sketch...",
				},
			},
		},
		Done: false,
		Metadata: map[string]interface{}{"stage": "analysis"},
	})

	var err error
	err = f.promptManager.ExecutePromptStreaming(ctx, promptName, variables, "glm-4.6v-flash", func(chunk *interfaces.PonchoStreamChunk) error {
		// Forward streaming response with additional metadata
		if chunk.Metadata == nil {
			chunk.Metadata = make(map[string]interface{})
		}
		chunk.Metadata["flow"] = f.Name()
		chunk.Metadata["stage"] = "analysis_streaming"
		return callback(chunk)
	})

	if err != nil {
		return fmt.Errorf("failed to generate streaming analysis: %w", err)
	}

	// Send completion
	duration := time.Since(startTime)
	callback(&interfaces.PonchoStreamChunk{
		Delta: &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: fmt.Sprintf("Analysis completed in %dms", duration.Milliseconds()),
				},
			},
		},
		Done:    true,
		Metadata: map[string]interface{}{
			"flow":               f.Name(),
			"stage":              "complete",
			"processing_time_ms": duration.Milliseconds(),
			"analysis_type":      analysisType,
			"prompt":            promptName,
		},
	})

	return nil
}