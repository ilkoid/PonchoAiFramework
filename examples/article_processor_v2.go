package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/context"
	"github.com/ilkoid/PonchoAiFramework/core/flow"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ArticleProcessorV2 показывает, как ваша задача решается с FlowContext
type ArticleProcessorV2 struct {
	flow interfaces.PonchoFlowV2
}

// NewArticleProcessorV2 создает ваш flow с новой архитектурой
func NewArticleProcessorV2(
	articleImporter interfaces.PonchoTool,
	visionModel interfaces.PonchoModel, // glm-4.6v-flash
	categoryTool interfaces.PonchoTool, // WB Categories
	textModel interfaces.PonchoModel, // deepseek-chat
	s3Tool interfaces.PonchoTool, // Для сохранения результатов
) (*ArticleProcessorV2, error) {

	// Ваш Fashion Product Enrichment Flow с FlowBuilder DSL
	flowBuilder := flow.NewFlowBuilder("fashion_product_enrichment").
		Description("Fashion product enrichment with AI analysis and SEO generation").
		Version("2.0.0").
		Category("fashion").
		RequiresVision(). // Говорим, что нужен vision model
		RequiresTool("article_importer").
		RequiresTool("wb_categories").
		RequiresTool("s3_storage").
		RequiresModel("vision_model").
		RequiresModel("text_model").
		Timeout(10 * time.Minute).
		EnableParallel(true). // Включаем параллельную обработку
		MaxConcurrency(3)

	// Step 1: Импорт данных из S3
	flowBuilder.
		Step("import_product_data").
		Tool(articleImporter, "s3_path").
		Input("s3_path").
		Output("product_data").
		Timeout(30 * time.Second).
		Continue()

	// Step 2: ПАРАЛЛЕЛЬНЫЙ анализ изображений (v2.0 преимущество!)
	flowBuilder.
		Step("analyze_images_parallel").
		Parallel().
		MaxConcurrency(5).
		FailFast(false). // Если одно изображение не analyze, продолжаем
		AddSubStep(&flow.CustomStep{
			name: "analyze_single_image",
			executor: func(ctx context.Context, flowCtx context.FlowContext) error {
				return analyzeProductImages(ctx, flowCtx, visionModel)
			},
		}).
		Requires("product_data.images").
		Provides("visual_features_list").
		Timeout(120 * time.Second).
		Continue()

	// Step 3: Классификация категории (может идти параллельно с анализом)
	flowBuilder.
		Step("classify_category").
		Tool(categoryTool, "product_data.name").
		Input("product_data.name").
		Output("category_id").
		Timeout(15 * time.Second).
		Continue()

	// Step 4: Генерация SEO описания (использует данные ИЗ всех предыдущих шагов!)
	flowBuilder.
		Step("generate_seo_description").
		Model(textModel, "seo_description_prompt").
		Inputs("product_data", "visual_features_list", "category_id").
		Output("seo_description").
		Temperature(0.7).
		MaxTokens(1000).
		Timeout(30 * time.Second).
		Continue()

	// Step 5: Сохранение enriched данных обратно в S3
	flowBuilder.
		Step("save_enriched_data").
		Tool(s3Tool, "enriched_product_data").
		Input("enriched_product_data").
		Timeout(30 * time.Second).
		CanFail(true). // Не прерываем flow если сохранение не удалось
		Continue()

	// Build flow
	builtFlow, err := flowBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build article processor v2: %w", err)
	}

	return &ArticleProcessorV2{
		flow: builtFlow,
	}, nil
}

// Execute запускает ваш Fashion Product Enrichment Flow
func (apv2 *ArticleProcessorV2) Execute(ctx context.Context, s3Path string) (*EnrichmentResult, error) {
	// 1. Создаем FlowContext - он будет жить на протяжении всего flow
	flowCtx := context.NewBaseFlowContext()

	// 2. Устанавливаем начальные данные в context
	if err := flowCtx.Set("s3_path", s3Path); err != nil {
		return nil, fmt.Errorf("failed to set s3_path: %w", err)
	}

	// 3. Выполняем flow (FlowContext автоматически передается между шагами)
	_, err := apv2.flow.Execute(ctx, s3Path, flowCtx)
	if err != nil {
		return nil, fmt.Errorf("flow execution failed: %w", err)
	}

	// 4. Собираем результаты из FlowContext
	result := &EnrichmentResult{
		S3Path:   s3Path,
		Success:  true,
		DateTime: time.Now(),
	}

	// Получаем все результаты из context (больше нет ручной передачи!)
	if productData, has := flowCtx.Get("product_data"); has {
		result.ProductData = productData
	}

	if visualFeatures, has := flowCtx.Get("visual_features_list"); has {
		result.VisualFeatures = visualFeatures
	}

	if categoryID, err := flowCtx.GetString("category_id"); err == nil {
		result.CategoryID = categoryID
	}

	if seoDescription, err := flowCtx.GetString("seo_description"); err == nil {
		result.SEODescription = seoDescription
	}

	// Метаданные выполнения
	if metadata, has := flowCtx.Get("execution_metadata"); has {
		result.Metadata = metadata
	}

	return result, nil
}

// EnrichmentResult представляет результат работы flow
type EnrichmentResult struct {
	S3Path         string      `json:"s3_path"`
	ProductData    interface{} `json:"product_data"`
	VisualFeatures interface{} `json:"visual_features"`
	CategoryID     string      `json:"category_id"`
	SEODescription string      `json:"seo_description"`
	Success        bool        `json:"success"`
	DateTime       time.Time   `json:"date_time"`
	Metadata       interface{} `json:"metadata"`
}

// analyzeProductImages - анализ изображений с использованием FlowContext
func analyzeProductImages(
	ctx context.Context,
	flowCtx context.FlowContext,
	visionModel interfaces.PonchoModel,
) error {
	// Получаем изображения из FlowContext (Step 1 их туда положил)
	productData, has := flowCtx.Get("product_data")
	if !has {
		return fmt.Errorf("product_data not found in context")
	}

	// Извлекаем изображения из product_data
	if productMap, ok := productData.(map[string]interface{}); ok {
		if imagesData, has := productMap["images"]; has {
			if imagesSlice, ok := imagesData.([]interface{}); ok {
				var allFeatures []string

				// Анализируем каждое изображение
				for i, imageData := range imagesSlice {
					if imageMap, ok := imageData.(map[string]interface{}); ok {
						if url, has := imageMap["url"].(string); has {
							// Анализируем одно изображение
							features, err := analyzeSingleImage(ctx, visionModel, url)
							if err != nil {
								// Log error but continue with other images
								fmt.Printf("Failed to analyze image %d: %v\n", i, err)
								continue
							}

							// Сохраняем анализ для каждого изображения
							imageKey := fmt.Sprintf("image_%d_analysis", i)
							flowCtx.Set(imageKey, features)
							allFeatures = append(allFeatures, features)
						}
					}
				}

				// Сохраняем список всех visual features в FlowContext
				// Step 4 (generate_seo_description) сможет их использовать!
				if err := flowCtx.Set("visual_features_list", allFeatures); err != nil {
					return fmt.Errorf("failed to store visual features: %w", err)
				}
			}
		}
	}

	return nil
}

// analyzeSingleImage - анализ одного изображения с vision model
func analyzeSingleImage(ctx context.Context, visionModel interfaces.PonchoModel, imageURL string) (string, error) {
	// Создаем prompt для анализа
	prompt := fmt.Sprintf("Analyze this fashion image: %s. Describe the item, style, materials, and visible features.", imageURL)

	maxTokens := 500
	temperature := float32(0.3)

	// Создаем request (v2.0 использует тот же интерфейс)
	req := &interfaces.PonchoModelRequest{
		Model: visionModel.Name(),
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
							MimeType: "image/jpeg", // MediaPipeline v2.0 определяет автоматически
						},
					},
				},
			},
		},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	}

	// Выполняем vision model
	resp, err := visionModel.Generate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("vision analysis failed: %w", err)
	}

	if resp != nil && resp.Message != nil && len(resp.Message.Content) > 0 {
		return resp.Message.Content[0].Text, nil
	}

	return "", fmt.Errorf("no content in vision response")
}

// createSEODescriptionPrompt создает промпт с учетом всех данных из FlowContext
func createSEODescriptionPrompt() string {
	// В реальной реализации этот промпт будет использовать переменные из FlowContext
	return `
	Сгенерируй привлекательное SEO-оптимизированное описание для товара модной индустрии.

	Используй следующую информацию:
	- Базовые данные продукта: {{.product_data}}
	- Визуальные особенности: {{.visual_features_list}}
	- Категория товара: {{.category_id}}

	Создай профессиональное описание для маркетплейса (Wildberries/Ozon), которое:
	- Включает описанные визуальные особенности из анализа изображений
	- Учитывает категорию товара для оптимизации под целевую аудиторию
	- Привлекает внимание покупателей
	- Подчеркивает ключевые преимущества
	- Оптимизировано для поиска

	Ответ предоставь только текст описания без дополнительных комментариев.
	`
}

// Пример использования:
func ExampleArticleProcessorV2() {
	// Инициализация компонентов (в реальном коде здесь будут ваши имплементации)
	var articleImporter interfaces.PonchoTool
	var visionModel interfaces.PonchoModel
	var categoryTool interfaces.PonchoTool
	var textModel interfaces.PonchoModel
	var s3Tool interfaces.PonchoTool

	// Создаем flow с новой архитектурой
	processor, err := NewArticleProcessorV2(
		articleImporter,
		visionModel,
		categoryTool,
		textModel,
		s3Tool,
	)
	if err != nil {
		fmt.Printf("Failed to create processor: %v\n", err)
		return
	}

	// Выполняем Fashion Product Enrichment
	ctx := context.Background()
	result, err := processor.Execute(ctx, "s3://fashion-products/product-123.json")
	if err != nil {
		fmt.Printf("Processing failed: %v\n", err)
		return
	}

	// Результат содержит данные ИЗ всех шагов
	fmt.Printf("✅ Product processed successfully!\n")
	fmt.Printf("📦 Product: %v\n", result.ProductData)
	fmt.Printf("👁️  Visual Features: %v\n", result.VisualFeatures)
	fmt.Printf("🏷️  Category: %s\n", result.CategoryID)
	fmt.Printf("📝 SEO Description: %s\n", result.SEODescription)
}

/*
КЛЮЧЕВЫЕ ПРЕИМУЩЕСТВА v2.0 для вашего случая:

1. ✅ РЕШЕНО: State Management
   FlowContext автоматически передает данные между шагами
   Step 4 получает product_data + visual_features + category_id из context

2. ✅ РЕШЕНО: Media Processing
   MediaPipeline автоматически конвертирует изображения для vision model
   Больше нет ручного base64 encoding

3. ✅ РЕШЕНО: Parallel Execution
   Анализ изображений может идти параллельно с категоризацией
   Ускорение обработки больших batch'ей товаров

4. ✅ РЕШЕНО: Complex Dependencies
   generate_seo_description автоматически получает данные из:
   - Step 1: product_data
   - Step 2: visual_features_list
   - Step 3: category_id

5. ✅ РЕШЕНО: Type Safety
   flowCtx.GetString("category_id") вместо interface{} type assertions
   Меньше runtime ошибок

6. ✅ РЕШЕНО: Error Handling
   Можно продолжать flow если один шаг failed
   CanFail(true) для non-critical operations
*/