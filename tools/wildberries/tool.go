package wildberries

import (
	"context"
	"fmt"

	"github.com/ilkoid/PonchoAiFramework/core/base"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// WBTool implements PonchoTool interface for Wildberries marketplace integration
type WBTool struct {
	*base.PonchoBaseTool
	client        *WBClient
	fashionAnalyzer *FashionAnalyzer
	modelRegistry interfaces.PonchoModelRegistry
	logger       interfaces.Logger
}

// NewWBTool creates a new Wildberries tool
func NewWBTool(modelRegistry interfaces.PonchoModelRegistry, logger interfaces.Logger) interfaces.PonchoTool {
	tool := &WBTool{
		modelRegistry: modelRegistry,
		logger:       logger,
	}

	// Initialize base tool
	baseTool := base.NewPonchoBaseTool(
		"wildberries",
		"Wildberries marketplace integration tool",
		"1.0.0",
		"marketplace",
	)

	tool.PonchoBaseTool = baseTool
	baseTool.SetLogger(logger)
	// Set tags
	for _, tag := range []string{"marketplace", "wildberries", "fashion", "ecommerce"} {
		baseTool.AddTag(tag)
	}
	return tool
}

// Initialize initializes the tool with configuration
func (t *WBTool) Initialize(ctx context.Context, config map[string]interface{}) error {
	t.logger.Info("Initializing Wildberries tool", nil)

	// Extract API key from config
	apiKey, ok := config["api_key"].(string)
	if !ok {
		return fmt.Errorf("wildberries api_key is required")
	}

	// Extract base URL (optional)
	baseURL, _ := config["base_url"].(string)

	// Create Wildberries client
	t.client = NewWBClient(apiKey, baseURL, t.logger)

	// Test API connection
	if err := t.testConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to Wildberries API: %w", err)
	}

	// Create fashion analyzer
	t.fashionAnalyzer = NewFashionAnalyzer(t.client, t.modelRegistry, t.logger)

	t.logger.Info("Wildberries tool initialized successfully", nil)
	return nil
}

// Shutdown gracefully shuts down the tool
func (t *WBTool) Shutdown(ctx context.Context) error {
	t.logger.Info("Shutting down Wildberries tool", nil)
	// No specific cleanup needed
	return nil
}

// executeTool executes the tool with given input
func (t *WBTool) executeTool(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	action, ok := input["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action is required")
	}

	switch action {
	case "ping":
		return t.executePing(ctx)
	case "get_categories":
		return t.executeGetCategories(ctx)
	case "get_subjects":
		return t.executeGetSubjects(ctx, input)
	case "analyze_fashion":
		return t.executeAnalyzeFashion(ctx, input)
	case "find_category":
		return t.executeFindCategory(ctx, input)
	case "get_characteristics":
		return t.executeGetCharacteristics(ctx, input)
	case "generate_schema":
		return t.executeGenerateSchema(ctx, input)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Tool actions

func (t *WBTool) executePing(ctx context.Context) (interface{}, error) {
	response, err := t.client.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (t *WBTool) executeGetCategories(ctx context.Context) (interface{}, error) {
	categories, err := t.client.GetParentCategories(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	}, nil
}

func (t *WBTool) executeGetSubjects(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	opts := &GetSubjectsOptions{}

	if locale, ok := input["locale"].(string); ok {
		opts.Locale = locale
	}
	if name, ok := input["name"].(string); ok {
		opts.Name = name
	}
	if parentID, ok := input["parent_id"].(float64); ok {
		opts.ParentID = int(parentID)
	}
	if limit, ok := input["limit"].(float64); ok {
		opts.Limit = int(limit)
	}
	if offset, ok := input["offset"].(float64); ok {
		opts.Offset = int(offset)
	}

	subjects, err := t.client.GetSubjects(ctx, opts)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"subjects": subjects,
		"count":    len(subjects),
	}, nil
}

func (t *WBTool) executeAnalyzeFashion(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	imageURL, ok := input["image_url"].(string)
	if !ok {
		return nil, fmt.Errorf("image_url is required")
	}

	description, _ := input["description"].(string) // Optional

	analysis, err := t.fashionAnalyzer.AnalyzeFashionItem(ctx, imageURL, description)
	if err != nil {
		return nil, err
	}

	return analysis, nil
}

func (t *WBTool) executeFindCategory(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	keywordsInterface, ok := input["keywords"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("keywords is required and must be an array")
	}

	keywords := make([]string, len(keywordsInterface))
	for i, kw := range keywordsInterface {
		if kwStr, ok := kw.(string); ok {
			keywords[i] = kwStr
		} else {
			return nil, fmt.Errorf("keyword at index %d is not a string", i)
		}
	}

	parent, subject, err := t.fashionAnalyzer.FindSuitableCategory(ctx, keywords)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"parent_category": parent,
		"subject":         subject,
	}

	return result, nil
}

func (t *WBTool) executeGetCharacteristics(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	subjectIDFloat, ok := input["subject_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("subject_id is required")
	}

	subjectID := int(subjectIDFloat)

	characteristics, err := t.fashionAnalyzer.GetRequiredCharacteristics(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	// Separate required and optional
	required := []SubjectCharacteristic{}
	optional := []SubjectCharacteristic{}

	for _, charc := range characteristics {
		if charc.Required {
			required = append(required, charc)
		} else {
			optional = append(optional, charc)
		}
	}

	return map[string]interface{}{
		"characteristics": characteristics,
		"required":        required,
		"optional":        optional,
		"count":           len(characteristics),
	}, nil
}

func (t *WBTool) executeGenerateSchema(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	subjectIDFloat, ok := input["subject_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("subject_id is required")
	}

	subjectID := int(subjectIDFloat)

	schema, err := t.fashionAnalyzer.GenerateProductSchema(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"schema": schema,
		"subject_id": subjectID,
	}, nil
}

// Helper methods

func (t *WBTool) testConnection(ctx context.Context) error {
	_, err := t.client.Ping(ctx)
	return err
}

func (t *WBTool) getToolSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{
					"ping",
					"get_categories",
					"get_subjects",
					"analyze_fashion",
					"find_category",
					"get_characteristics",
					"generate_schema",
				},
				"description": "Action to perform",
			},
			"image_url": map[string]interface{}{
				"type":        "string",
				"description": "Image URL for fashion analysis (required for analyze_fashion)",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Optional description for fashion analysis",
			},
			"keywords": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Keywords for category matching (required for find_category)",
			},
			"locale": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"ru", "en", "zh"},
				"description": "Locale for subjects (optional)",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Search by subject name (optional)",
			},
			"parent_id": map[string]interface{}{
				"type":        "number",
				"description": "Filter subjects by parent ID (optional)",
			},
			"limit": map[string]interface{}{
				"type":        "number",
				"description": "Limit number of results (optional, max 1000)",
			},
			"offset": map[string]interface{}{
				"type":        "number",
				"description": "Pagination offset (optional)",
			},
			"subject_id": map[string]interface{}{
				"type":        "number",
				"description": "Subject ID for characteristics and schema (required)",
			},
		},
		"required": []string{"action"},
	}
}