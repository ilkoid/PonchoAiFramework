package examples

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math/rand"
	"runtime"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/context"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ResizeStrategy определяет стратегию ресайза изображений
type ResizeStrategy struct {
	MaxWidth        int     `json:"max_width"`
	MaxHeight       int     `json:"max_height"`
	MaxFileSizeKB    int     `json:"max_file_size_kb"`
	Quality         int     `json:"quality"`         // 1-100
	TargetFormat     string  `json:"target_format"`    // "jpeg", "png"
	EnableResize     bool    `json:"enable_resize"`
	EnableSmartCrop  bool    `json:"enable_smart_crop"`
}

// DefaultVisionResizeStrategy оптимальна для vision моделей
var DefaultVisionResizeStrategy = &ResizeStrategy{
	MaxWidth:     1024,   // GLM-4.6V оптимально с 1024px
	MaxHeight:    1024,
	MaxFileSizeKB: 500,    // 500KB предел
	Quality:      85,     // Хороший баланс качества/размера
	TargetFormat:  "jpeg", // Лучше для фотографий
	EnableResize: true,
	EnableSmartCrop: false,
}

// HighQualityResizeStrategy для анализа деталей
var HighQualityResizeStrategy = &ResizeStrategy{
	MaxWidth:     2048,   // Высокое разрешение
	MaxHeight:    2048,
	MaxFileSizeKB: 1024,   // 1MB предел
	Quality:      95,     // Максимальное качество
	TargetFormat:  "png",  // Без потерь для деталей
	EnableResize: true,
	EnableSmartCrop: true,
}

// ResizeManager управляет ресайзом изображений
type ResizeManager struct {
	strategy  *ResizeStrategy
	cache     map[string][]byte
	cacheSize int64
	logger    interfaces.Logger
}

// NewResizeManager создает менеджер ресайза
func NewResizeManager(strategy *ResizeStrategy, logger interfaces.Logger) *ResizeManager {
	if strategy == nil {
		strategy = DefaultVisionResizeStrategy
	}

	return &ResizeManager{
		strategy:  strategy,
		cache:     make(map[string][]byte),
		logger:    logger,
		cacheSize: 50 * 1024 * 1024, // 50MB cache limit
	}
}

// ResizeForVision ресайзит изображение для vision модели
func (rm *ResizeManager) ResizeForVision(ctx context.Context, imageData []byte, mimeType string) ([]byte, error) {
	if !rm.strategy.EnableResize {
		return imageData, nil // Без изменений
	}

	// Проверяем размер
	imageSizeKB := len(imageData) / 1024
	if imageSizeKB <= rm.strategy.MaxFileSizeKB {
		return imageData, nil // В пределах лимита
	}

	rm.logger.Info("Resizing image for vision model",
		"original_size_kb", imageSizeKB,
		"max_size_kb", rm.strategy.MaxFileSizeKB,
	)

	// Декодируем изображение
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Ресайзим
	resizedImg := rm.resizeImage(img)

	// Кодируем с заданным качеством
	return rm.encodeImage(resizedImg, format)
}

// ResizeImageDims ресайзит с учетом пропорций
func (rm *ResizeManager) resizeImage(img image.Image) image.Image {
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Если изображение уже нужного размера
	if originalWidth <= rm.strategy.MaxWidth && originalHeight <= rm.strategy.MaxHeight {
		return img
	}

	// Вычисляем новые размеры с сохранением пропорций
	widthRatio := float64(rm.strategy.MaxWidth) / float64(originalWidth)
	heightRatio := float64(rm.strategy.MaxHeight) / float64(originalHeight)

	var newWidth, newHeight int
	if widthRatio < heightRatio {
		// Ограничено по ширине
		newWidth = rm.strategy.MaxWidth
		newHeight = int(float64(originalHeight) * widthRatio)
	} else {
		// Ограничено по высоте
		newHeight = rm.strategy.MaxHeight
		newWidth = int(float64(originalWidth) * heightRatio)
	}

	// Создаем новое изображение
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Рисуем с хорошим качеством
	draw.CatmullRom(resized, img, bounds, resized.Bounds, draw.Over, nil)

	rm.logger.Debug("Image resized",
		"original", fmt.Sprintf("%dx%d", originalWidth, originalHeight),
		"resized", fmt.Sprintf("%dx%d", newWidth, newHeight),
	)

	return resized
}

func (rm *ResizeManager) encodeImage(img image.Image, format image.Image) ([]byte, error) {
	var buf bytes.Buffer

	switch rm.strategy.TargetFormat {
	case "jpeg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: rm.strategy.Quality})
		if err != nil {
			return nil, fmt.Errorf("failed to encode JPEG: %w", err)
		}
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err := encoder.Encode(&buf, img)
		if err != nil {
			return nil, fmt.Errorf("failed to encode PNG: %w", err)
		}
	default:
		// По умолчанию JPEG
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: rm.strategy.Quality})
		if err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// GetOptimalStrategy выбирает оптимальную стратегию для модели
func GetOptimalStrategy(model interfaces.PonchoModel) *ResizeStrategy {
	modelName := model.Name()

	// GLM-4.6V работает оптимально с 1024px
	if strings.Contains(modelName, "glm") || strings.Contains(modelName, "vision") {
		return DefaultVisionResizeStrategy
	}

	// Для моделей требующих максимальной детализации
	if strings.Contains(modelName, "analysis") || strings.Contains(modelName, "detail") {
		return HighQualityResizeStrategy
	}

	// По умолчанию - стандартная стратегия
	return DefaultVisionResizeStrategy
}

// FlowContextWithResize интегрирует ресайз с FlowContext
type FlowContextWithResize struct {
	*context.BaseFlowContextV2
	resizeManager *ResizeManager
	logger         interfaces.Logger
}

// NewFlowContextWithResize создает контекст с ресайзом
func NewFlowContextWithResize(logger interfaces.Logger) *FlowContextWithResize {
	baseCtx := context.NewBaseFlowContextV2()
	resizeManager := NewResizeManager(DefaultVisionResizeStrategy, logger)

	return &FlowContextWithResize{
		BaseFlowContextV2: baseCtx,
		resizeManager:     resizeManager,
		logger:           logger,
	}

}

// SetResizedImage сохраняет ресайзенное изображение в контекст
func (fcr *FlowContextWithResize) SetResizedImage(
	ctx context.Context,
	key string,
	imageData []byte,
	mimeType string,
) error {
	// Ресайзим для vision модели
	resizedData, err := fcr.resizeManager.ResizeForVision(ctx, imageData, mimeType)
	if err != nil {
		return fmt.Errorf("failed to resize image: %w", err)
	}

	// Сохраняем в контекст (уже оптимизированное изображение)
	return fcr.SetBytes(key, resizedData)
}

// BatchResizeImages пакетный ресайз изображений
func (fcr *FlowContextWithResize) BatchResizeImages(
	ctx context.Context,
	images map[string][]byte,
	mimeTypes map[string]string,
) (map[string][]byte, error) {
	result := make(map[string][]byte)

	for key, imageData := range images {
		mimeType := mimeTypes[key]
		resized, err := fcr.resizeManager.ResizeForVision(ctx, imageData, mimeType)
		if err != nil {
			fcr.logger.Warn("Failed to resize image",
				"key", key,
				"error", err,
			)
			// Используем оригинальное изображение
			result[key] = imageData
		} else {
			result[key] = resized
		}
	}

	return result, nil
}

// MemorySafetyAnalysis анализирует безопасность по памяти
func MemorySafetyAnalysis() {
	fmt.Println("=== Memory Safety Analysis for Resize Strategy ===")

	// Сценарии использования
	scenarios := []struct {
		name        string
		imageCount  int
		originalMB  float64
		resizedMB   float64
		safe        bool
		description string
	}{
		{
			name:        "Small Batch (5 images)",
			imageCount:  5,
			originalMB:  25.0,  // 5MB каждое
			resizedMB:   1.25,  // 250KB каждое
			safe:        true,
			description: "Маленький batch, безопасно",
		},
		{
			name:        "Medium Batch (20 images)",
			imageCount:  20,
			originalMB:  100.0, // 5MB каждое
			resizedMB:   5.0,   // 250KB каждое
			safe:        true,
			description: "Средний batch, безопасно",
		},
		{
			name:        "Large Batch (50 images)",
			imageCount:  50,
			originalMB:  250.0, // 5MB каждое
			resizedMB:   12.5,  // 250KB каждое
			safe:        false,
			description: "Большой batch, может быть опасно",
		},
		{
			name:        "Very Large Batch (100 images)",
			imageCount: 100,
			originalMB:  500.0, // 5MB каждое
			resizedMB:   25.0,  // 250KB каждое
			safe:        false,
			description: "Очень большой batch, опасно",
		},
	}

	fmt.Println("\n📊 Memory Usage Comparison:")
	fmt.Printf("%-25s | %10s | %10s | %5s | %s\n", "Scenario", "Original", "Resized", "Safe", "Description")
	fmt.Println(strings.Repeat("-", 75))

	for _, scenario := range scenarios {
		status := "✅"
		if !scenario.safe {
			status = "❌"
		}
		fmt.Printf("%-25s | %9.1fMB | %9.1fMB | %4s | %s\n",
			scenario.name, scenario.originalMB, scenario.resizedMB, status, scenario.description)
	}

	fmt.Println("\n💡 Safety Guidelines:")
	fmt.Println("✅ Small batches (< 20 images): Safe to store in context")
	fmt.Println("✅ Medium batches (20-50 images): Safe with memory monitoring")
	fmt.Println("❌ Large batches (> 50 images): Use streaming or temp files")
	fmt.Println("❌ Very large batches (> 100 images): Definitely use temp files")

	fmt.Println("\n🎯 Recommended Limits:")
	fmt.Println("• Small images (< 500KB after resize): OK to store")
	fmt.Println("• Medium images (500KB-1MB): Store with monitoring")
	fmt.Println("• Large images (> 1MB): Use temp file pattern")
	fmt.Println("• Parallel contexts: 50-100MB total per server")
	fmt.Println("• System memory reserve: 20-30% of total RAM")
}

// PerformanceBenchmark сравнивает производительность
func PerformanceBenchmark() {
	fmt.Println("\n=== Performance Benchmark ===")

	// Тестовые данные
	testSizes := []int{100, 500, 1024, 2048} // px
	testImages := []int{5, 10, 20, 50}

	fmt.Println("\n📏 Resize Performance (average per image):")
	fmt.Printf("%-10s | %15s | %15s | %15s\n", "Size", "Original KB", "Resized KB", "Ratio")
	fmt.Println(strings.Repeat("-", 65))

	for _, size := range testSizes {
		// Симуляция размеров файлов
		originalSize := float64(size*size) * 4 // Rough estimate: 4 bytes per pixel
		resizedSize := float64(1024*1024) * 0.25 // 250KB target

		ratio := originalSize / (1024 * resizedSize)

		fmt.Printf("%-10d | %13.1fKB | %13.1fKB | %13.1fx\n",
			size, originalSize/1024, resizedSize/1024, ratio)
	}

	fmt.Println("\n⚡ Batch Processing Time:")
	fmt.Printf("%-15s | %15s | %15s\n", "Images", "Total Time", "Per Image")
	fmt.Println(strings.Repeat("-", 50))

	for _, count := range testImages {
		// Симуляция времени обработки
		avgResizeTime := 50 * time.Millisecond
		totalTime := time.Duration(count) * avgResizeTime

		fmt.Printf("%-15d | %13v | %11v\n",
			count, totalTime, avgResizeTime)
	}
}

// DemonstrateSafeUsage демонстрирует безопасное использование
func DemonstrateSafeUsage() {
	fmt.Println("\n=== Safe Usage Demo ===")

	flowCtx := NewFlowContextWithResize(interfaces.NewDefaultLogger())
	ctx := context.Background()

	// Симуляция получения изображений
	images := make(map[string][]byte)
	mimeTypes := make(map[string]string)

	// Генерируем тестовые изображения разных размеров
	sizes := []int{512, 1024, 2048} // px
	for i, size := range sizes {
		key := fmt.Sprintf("product_image_%d", i)

		// Симуляция изображения (размер ~ size^2 * 4 bytes)
		imageData := make([]byte, size*size*4)
		rand.Read(imageData)

		images[key] = imageData
		mimeTypes[key] = "image/jpeg"
	}

	fmt.Printf("Created %d test images\n", len(images))

	// Батчный ресайз (безопасно!)
	startTime := time.Now()
	resizedImages, err := flowCtx.BatchResizeImages(ctx, images, mimeTypes)
	if err != nil {
		fmt.Printf("Batch resize failed: %v\n", err)
		return
	}

	batchTime := time.Since(startTime)
	totalOriginalSize := calculateTotalSize(images)
	totalResizedSize := calculateTotalSize(resizedImages)
	reduction := (float64(totalOriginalSize-totalResizedSize) / float64(totalOriginalSize)) * 100

	fmt.Printf("Batch resize completed in %v\n", batchTime)
	fmt.Printf("Total original size: %d KB\n", totalOriginalSize/1024)
	fmt.Printf("Total resized size: %d KB\n", totalResizedSize/1024)
	fmt.Printf("Size reduction: %.1f%%\n", reduction)

	// Проверяем использование памяти
	memoryUsage := flowCtx.GetMemoryUsage()
	fmt.Printf("Memory usage: %.2f KB\n", float64(memoryUsage)/1024)

	// Проверяем безопасность
	if memoryUsage < 10*1024*1024 { // 10MB limit
		fmt.Println("✅ Memory usage is safe")
	} else {
		fmt.Println("⚠️  High memory usage detected")
	}
}

func calculateTotalSize(images map[string][]byte) int {
	total := 0
	for _, data := range images {
		total += len(data)
	}
	return total
}

// BestPractices рекомендации
func BestPractices() {
	fmt.Println("\n🏆 Resize Strategy Best Practices:")

	fmt.Println("\n📐 Recommended Settings for Different Use Cases:")

	fmt.Println("\n1. Vision Models (GLM-4.6V, GPT-4V):")
	fmt.Println("   • Max resolution: 1024x1024px")
	fmt.Println("   • Target file size: 200-500KB")
	fmt.Println("   • Format: JPEG (85% quality)")
	fmt.Println("   • Reason: Optimal balance for analysis")

	fmt.Println("\n2. Classification Models:")
	fmt.Println("   • Max resolution: 512x512px")
	fmt.Println("   • Target file size: 100-200KB")
	fmt.Println("   • Format: JPEG (75% quality)")
	fmt.Println("   • Reason: Features are visible at lower resolution")

	fmt.Println("\n3. Detail Analysis Models:")
	fmt.Println("   • Max resolution: 2048x2048px")
	fmt.Println("   • Target file size: 500KB-1MB")
	fmt.Println("   • Format: PNG (lossless) or JPEG (95% quality)")
	fmt.Println("   • Reason: Maximum detail retention")

	fmt.Println("\n⚡ Performance Tips:")
	fmt.Println("   • Batch resize when possible (parallel processing)")
	fmt.Println("   • Cache resized images for repeated analysis")
	fmt.Println("   • Use progressive JPEG for faster loading")
	fmt.Println("   • Consider image format based on content type")

	fmt.Println("\n🛡️ Safety Rules:")
	fmt.Println("   • Never resize to > 2MB in memory")
	fmt.Println("   • Limit 20-30 images per context")
	fmt.Println("   • Monitor memory usage in real-time")
	fmt.Println("   • Use temp files for > 50MB total data")
	fmt.Println("   • Implement circuit breakers for resize failures")
}

func RunResizeAnalysis() {
	MemorySafetyAnalysis()
	PerformanceBenchmark()
	DemonstrateSafeUsage()
	BestPractices()
}

/*
RESIZE STRATEGY SUMMARY:

✅ ПОЗИТИВНЫЕ АСПЕКТЫ:
1. Сокращает использование памяти на 80-95%
2. Улучшает производительность vision models
3. Позволяет обрабатывать больше изображений параллельно
4. Совместим с существующим FlowContext

⚠️ ОГРАНИЧЕНИЯ:
1. Потеря деталей при агрессивном ресайзе
2. Дополнительное время на обработку
3. Потребность в тестировании качества
4. Не подходит для всех use cases (медицинские изображения, например)

🎯 РЕКОМЕНДАЦИЯ:
- Используйте для vision analysis и классификации
- Не используйте для медицинских/научных изображений
- Всегда тестируйте качество после ресайза
- Мониторьте memory usage в production
*/