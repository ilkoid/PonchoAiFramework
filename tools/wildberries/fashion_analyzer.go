package wildberries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// FashionAnalyzer integrates Wildberries API with AI models for fashion analysis
type FashionAnalyzer struct {
	client       *WBClient
	modelRegistry interfaces.PonchoModelRegistry
	logger       interfaces.Logger

	// Caching
	categories     []ParentCategory
	categoriesTime time.Time
	subjects       map[int][]Subject
	subjectsTime   time.Time
	characteristics map[int][]SubjectCharacteristic
	characteristicsTime time.Time
}

// NewFashionAnalyzer creates a new fashion analyzer
func NewFashionAnalyzer(client *WBClient, modelRegistry interfaces.PonchoModelRegistry, logger interfaces.Logger) *FashionAnalyzer {
	return &FashionAnalyzer{
		client:       client,
		modelRegistry: modelRegistry,
		logger:       logger,
		subjects:     make(map[int][]Subject),
		characteristics: make(map[int][]SubjectCharacteristic),
	}
}

// AnalyzeFashionItem analyzes a fashion item from image and text description
func (fa *FashionAnalyzer) AnalyzeFashionItem(ctx context.Context, imageURL, description string) (*FashionAnalysis, error) {
	fa.logger.Info("Starting fashion analysis", map[string]interface{}{
		"image_url":   imageURL,
		"description": description,
	})

	// Get vision model for image analysis
	model, err := fa.modelRegistry.Get("glm-4.6v-flash")
	if err != nil {
		return nil, fmt.Errorf("failed to get vision model: %w", err)
	}

	// Create analysis prompt
	prompt := fa.buildAnalysisPrompt(description)

	// Make request to vision model
	request := &interfaces.PonchoModelRequest{
		Model: model.Name(),
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleSystem,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "You are a professional fashion analyst specializing in product categorization for Wildberries marketplace. Analyze fashion items and extract structured information in Russian.",
					},
				},
			},
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
							URL: imageURL,
						},
					},
				},
			},
		},
		Temperature: &[]float32{0.1}[0], // Low temperature for consistent classification
		MaxTokens:   &[]int{2000}[0],
	}

	response, err := model.Generate(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate analysis: %w", err)
	}

	// Parse AI response
	content := ""
	if response.Message != nil && len(response.Message.Content) > 0 {
		content = response.Message.Content[0].Text
	}
	analysis, err := fa.parseAnalysisResponse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	// Enhance with Wildberries data
	if err := fa.enrichWithWBData(ctx, analysis); err != nil {
		fa.logger.Warn("Failed to enrich with Wildberries data", map[string]interface{}{
			"error": err.Error(),
		})
		// Don't fail the entire analysis if enrichment fails
	}

	analysis.ProcessedAt = time.Now()
	analysis.Model = model.Name()

	fa.logger.Info("Fashion analysis completed", map[string]interface{}{
		"subject":     analysis.Subject,
		"type":        analysis.Type,
		"confidence":  analysis.Confidence,
	})

	return analysis, nil
}

// FindSuitableCategory finds the best matching category for a fashion item
func (fa *FashionAnalyzer) FindSuitableCategory(ctx context.Context, keywords []string) (*ParentCategory, *Subject, error) {
	// Get categories (with caching)
	categories, err := fa.getParentCategories(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get parent categories: %w", err)
	}

	// Find best matching parent category
	var bestParent *ParentCategory
	var bestScore float64

	for _, category := range categories {
		score := fa.calculateMatchScore(category.Name, keywords)
		if score > bestScore {
			bestScore = score
			bestParent = &category
		}
	}

	if bestParent == nil {
		return nil, nil, fmt.Errorf("no suitable category found")
	}

	// Get subjects for this parent category
	subjects, err := fa.getSubjectsForParent(ctx, bestParent.ID)
	if err != nil {
		return bestParent, nil, fmt.Errorf("failed to get subjects: %w", err)
	}

	// Find best matching subject
	var bestSubject *Subject
	var bestSubjectScore float64

	for _, subject := range subjects {
		score := fa.calculateMatchScore(subject.Name, keywords)
		if score > bestSubjectScore {
			bestSubjectScore = score
			bestSubject = &subject
		}
	}

	if bestSubject == nil {
		return bestParent, nil, fmt.Errorf("no suitable subject found for category %s", bestParent.Name)
	}

	return bestParent, bestSubject, nil
}

// GetRequiredCharacteristics gets required characteristics for a subject
func (fa *FashionAnalyzer) GetRequiredCharacteristics(ctx context.Context, subjectID int) ([]SubjectCharacteristic, error) {
	// Check cache first
	if charcs, exists := fa.characteristics[subjectID]; exists && time.Since(fa.characteristicsTime) < time.Hour {
		return charcs, nil
	}

	characteristics, err := fa.client.GetSubjectCharacteristics(ctx, subjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subject characteristics: %w", err)
	}

	// Update cache
	fa.characteristics[subjectID] = characteristics
	fa.characteristicsTime = time.Now()

	return characteristics, nil
}

// GenerateProductSchema generates a product schema with required characteristics
func (fa *FashionAnalyzer) GenerateProductSchema(ctx context.Context, subjectID int) (map[string]interface{}, error) {
	characteristics, err := fa.GetRequiredCharacteristics(ctx, subjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get characteristics: %w", err)
	}

	schema := make(map[string]interface{})

	// Group characteristics by type
	for _, charc := range characteristics {
		fieldInfo := map[string]interface{}{
			"type":        fa.mapCharcType(charc.CharcType),
			"required":    charc.Required,
			"description": charc.Name,
		}

		if charc.UnitName != "" {
			fieldInfo["unit"] = charc.UnitName
		}

		if charc.MaxCount > 0 {
			fieldInfo["max_count"] = charc.MaxCount
		}

		// Use characteristic name as field key
		fieldKey := strings.ToLower(strings.ReplaceAll(charc.Name, " ", "_"))
		schema[fieldKey] = fieldInfo
	}

	return schema, nil
}

// Helper methods

func (fa *FashionAnalyzer) buildAnalysisPrompt(description string) string {
	return fmt.Sprintf(`Проанализируй изображение модного изделия и верни результат в формате JSON.

Описание товара: %s

Верни JSON со следующей структурой:
{
  "type": "тип изделия (платье, рубашка, джинсы, обувь и т.д.)",
  "style": "стиль (кэжуал, деловой, спортивный, вечерний и т.д.)",
  "season": "сезон (демисезон, лето, зима, круглогодичный)",
  "gender": "пол (мужской, женский, унисекс)",
  "material": "основной материал (хлопок, полиэстер, шерсть и т.д.)",
  "color": "основной цвет",
  "pattern": "узор (однотонный, полоска, клетка, цветной и т.д.)",
  "name": "рекомендуемое название для товара на Wildberries",
  "description": "рекомендуемое описание для товара",
  "tags": ["тег1", "тег2", "тег3"],
  "confidence": 0.95
}

Будь точен в классификации и используй русские термины для Wildberries.`, description)
}

func (fa *FashionAnalyzer) parseAnalysisResponse(content string) (*FashionAnalysis, error) {
	// Extract JSON from response
	var analysis FashionAnalysis

	// Find JSON in response
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")

	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := content[start : end+1]

	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &analysis, nil
}

func (fa *FashionAnalyzer) enrichWithWBData(ctx context.Context, analysis *FashionAnalysis) error {
	// Find matching category and subject
	keywords := []string{analysis.Type, analysis.Style, analysis.Gender}

	parent, subject, err := fa.FindSuitableCategory(ctx, keywords)
	if err != nil {
		return err
	}

	if parent != nil {
		analysis.ParentCategory = parent.Name
	}

	if subject != nil {
		analysis.Subject = subject.Name
		analysis.SubjectID = subject.ID
	}

	return nil
}

func (fa *FashionAnalyzer) getParentCategories(ctx context.Context) ([]ParentCategory, error) {
	if len(fa.categories) > 0 && time.Since(fa.categoriesTime) < 24*time.Hour {
		return fa.categories, nil
	}

	categories, err := fa.client.GetParentCategories(ctx)
	if err != nil {
		return nil, err
	}

	fa.categories = categories
	fa.categoriesTime = time.Now()

	return categories, nil
}

func (fa *FashionAnalyzer) getSubjectsForParent(ctx context.Context, parentID int) ([]Subject, error) {
	// Check cache
	if subjects, exists := fa.subjects[parentID]; exists && time.Since(fa.subjectsTime) < time.Hour {
		return subjects, nil
	}

	subjects, err := fa.client.GetSubjects(ctx, &GetSubjectsOptions{
		ParentID: parentID,
		Limit:    1000,
	})
	if err != nil {
		return nil, err
	}

	// Update cache
	fa.subjects[parentID] = subjects
	fa.subjectsTime = time.Now()

	return subjects, nil
}

func (fa *FashionAnalyzer) calculateMatchScore(text string, keywords []string) float64 {
	text = strings.ToLower(text)
	score := 0.0

	for _, keyword := range keywords {
		keyword = strings.ToLower(keyword)
		if strings.Contains(text, keyword) {
			score += 1.0
		}
		// Add partial matches
		if strings.Contains(keyword, text) || strings.Contains(text, keyword) {
			score += 0.5
		}
	}

	return score
}

func (fa *FashionAnalyzer) mapCharcType(charcType int) string {
	switch charcType {
	case 1:
		return "string"
	case 2:
		return "number"
	case 3:
		return "boolean"
	case 4:
		return "list"
	case 5:
		return "text"
	default:
		return "string"
	}
}