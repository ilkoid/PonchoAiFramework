package examples

import (
	"fmt"
	"runtime"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/context"
)

// Демонстрация prevention memory bloat в нашей v2.0 реализации
func MemoryBloatPreventionExample() {
	fmt.Println("=== Memory Bloat Prevention Demo ===")

	// Создаем FlowContext с контролем памяти
	flowCtx := context.NewBaseFlowContextV2()

	// ✅ БЕЗОПАСНО: Строки, числа, маленькие структуры
	flowCtx.SetString("product_name", "Summer Dress")
	flowCtx.SetFloat64("price", 99.99)
	flowCtx.SetBool("in_stock", true)

	// Маленькая JSON структура
	productInfo := map[string]interface{}{
		"id":    12345,
		"name":  "Summer Dress",
		"brand": "FashionCo",
	}
	flowCtx.Set("product_info", productInfo)

	fmt.Println("✅ Safe data stored")
	fmt.Printf("Memory usage: %d MB\n", getMemoryUsageMB())

	// ❌ ОПАСНО: Большие бинарные данные - НЕЛЬЗЯ!
	/*
	// Это бы привело к OOM!
	largeImageData := make([]byte, 5*1024*1024) // 5MB
	for i := 0; i < 100; i++ { // 100 таких изображений = 500MB!
		flowCtx.SetBytes(fmt.Sprintf("image_%d", i), largeImageData)
	}
	*/

	// ✅ ПРАВИЛЬНО: ImageReference вместо binary данных
	imageURLs := []string{
		"https://example.com/dress1.jpg",  // URL только, не байты!
		"https://example.com/dress2.jpg",
		"https://example.com/dress3.jpg",
	}

	for i, url := range imageURLs {
		flowCtx.SetImageFromURL(fmt.Sprintf("image_%d", i), url)
	}

	fmt.Println("✅ Image references stored (no binary data)")
	fmt.Printf("Memory usage: %d MB\n", getMemoryUsageMB())

	// Мониторинг использования памяти изображений
	memoryUsage := flowCtx.GetMemoryUsage()
	fmt.Printf("Image memory usage: %d KB\n", memoryUsage/1024)

	// Ленивая загрузка - байты загружаются только при необходимости
	fmt.Println("\n=== Lazy Loading Demo ===")

	// Загрузка только ОДНОГО изображения по требованию
	ctx := context.Background()
	imageBytes, err := flowCtx.LoadImageBytes(ctx, "image_0")
	if err != nil {
		fmt.Printf("Failed to load image: %v\n", err)
	} else {
		fmt.Printf("Loaded image 0: %d bytes\n", len(imageBytes))
		fmt.Printf("Memory usage after load: %d MB\n", getMemoryUsageMB())
	}

	// Автоматическая очистка старого кэша
	fmt.Println("\n=== Automatic Cache Cleanup ===")
	fmt.Printf("Before cleanup: %d KB used\n", flowCtx.GetMemoryUsage()/1024)

	// Имитация загрузки множества изображений для триггера cleanup
	for i := 0; i < 150; i++ {
		flowCtx.SetImageFromURL(fmt.Sprintf("test_img_%d", i),
			fmt.Sprintf("https://example.com/test%d.jpg", i))
	}

	fmt.Printf("After adding 150 images: %d KB used\n", flowCtx.GetMemoryUsage()/1024)
	fmt.Println("Automatic cleanup triggered when limit exceeded")

	// Демонстрация безопасности параллельных контекстов
	fmt.Println("\n=== Parallel Contexts Demo ===")

	// 10 параллельных FlowContext для 10 пользователей
	startTime := time.Now()

	var contexts []*context.BaseFlowContextV2
	for i := 0; i < 10; i++ {
		ctx := context.NewBaseFlowContextV2()
		ctx.SetString("user_id", fmt.Sprintf("user_%d", i))
		ctx.SetString("product_name", fmt.Sprintf("Product %d", i))

		contexts = append(contexts, ctx)
	}

	fmt.Printf("Created 10 isolated contexts in %v\n", time.Since(startTime))
	fmt.Printf("Total memory usage: %d MB\n", getMemoryUsageMB())

	// Очистка
	for _, ctx := range contexts {
		ctx.EvictAllImageCache()
	}

	fmt.Println("=== Memory Management Best Practices ===")
	PrintMemoryBestPractices()
}

func PrintMemoryBestPractices() {
	fmt.Println("\n🚫 НЕ ДЕЛАТЬ (Memory Bloat Risks):")
	fmt.Println("❌ flowCtx.SetBytes('image', largeImageData) // Загрузка байтов в память")
	fmt.Println("❌ хранить 100+ изображений по 5MB в одном контексте")
	fmt.Println("❌ Игнорировать лимиты памяти при параллельных процессах")

	fmt.Println("\n✅ ДЕЛАТЬ (Memory Safe):")
	fmt.Println("✅ flowCtx.SetImageFromURL('image', url) // Только ссылка")
	fmt.Println("✅ flowCtx.GetImageBytes(ctx, key) // Lazy loading по требованию")
	fmt.Println("✅ flowCtx.GetMemoryUsage() // Мониторинг потребления")
	fmt.Println("✅ flowCtx.EvictImageCache(key) // Ручная очистка при необходимости")

	fmt.Println("\n📊 Memory Thresholds:")
	fmt.Println("• Small images (< 1MB): OK в ImageReference")
	fmt.Println("• Medium images (1-50MB): Temp file pattern")
	fmt.Println("• Large images (> 50MB): Stream processing")
	fmt.Println("• Parallel flows: 10-50MB контекст на поток")
	fmt.Println("• System limit: 70-80% от доступной RAM")
}

func getMemoryUsageMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

// Stress test для проверки memory safety
func MemoryStressTest() {
	fmt.Println("\n=== Memory Stress Test ===")

	baseMemory := getMemoryUsageMB()
	fmt.Printf("Base memory: %.2f MB\n", baseMemory)

	// Тест: 50 контекстов с изображениями
	contexts := make([]*context.BaseFlowContextV2, 50)

	for i := range contexts {
		contexts[i] = context.NewBaseFlowContextV2()
		contexts[i].SetString("flow_id", fmt.Sprintf("flow_%d", i))

		// Добавляем ImageReference (не байты!)
		for j := 0; j < 5; j++ {
			contexts[i].SetImageFromURL(
				fmt.Sprintf("img_%d", j),
				fmt.Sprintf("https://example.com/img_%d.jpg", j),
			)
		}
	}

	peakMemory := getMemoryUsageMB()
	fmt.Printf("Peak memory with 50 contexts: %.2f MB\n", peakMemory)
	fmt.Printf("Memory increase: %.2f MB\n", peakMemory-baseMemory)

	// Тест lazy loading
	for i, ctx := range contexts {
		for j := 0; j < 2; j++ {
			// Загружаем только несколько изображений для теста
			key := fmt.Sprintf("img_%d", j)
			if bytes, err := ctx.LoadImageBytes(context.Background(), key); err == nil {
				fmt.Printf("Context %d, Image %d: %d bytes\n", i, j, len(bytes))
			}
		}
	}

	loadedMemory := getMemoryUsageMB()
	fmt.Printf("Memory after lazy loading: %.2f MB\n", loadedMemory)

	// Очистка
	for _, ctx := range contexts {
		ctx.Clear()
		ctx.EvictAllImageCache()
	}

	finalMemory := getMemoryUsageMB()
	fmt.Printf("Memory after cleanup: %.2f MB\n", finalMemory)
	fmt.Printf("Memory reclaimed: %.2f MB\n", loadedMemory-finalMemory)
}

// Демонстрация разницы между подходами
func CompareMemoryApproaches() {
	fmt.Println("\n=== Memory Approach Comparison ===")

	// BAD: Хранение байтов в памяти (симуляция)
	fmt.Println("❌ BAD APPROACH: Store bytes in memory")
	badMemoryBefore := getMemoryUsageMB()

	// Симуляция хранения 10 изображений по 5MB каждое в памяти
	badData := make(map[string][]byte)
	for i := 0; i < 10; i++ {
		badData[fmt.Sprintf("img_%d", i)] = make([]byte, 5*1024*1024) // 5MB
	}

	badMemoryAfter := getMemoryUsageMB()
	fmt.Printf("Memory usage with 50MB of image data: %.2f MB\n", badMemoryAfter-badMemoryBefore)

	// Очистка
	for _, data := range badData {
		data = nil
	}

	// GOOD: ImageReference approach
	fmt.Println("\n✅ GOOD APPROACH: ImageReference with lazy loading")
	goodMemoryBefore := getMemoryUsageMB()

	flowCtx := context.NewBaseFlowContextV2()
	for i := 0; i < 10; i++ {
		flowCtx.SetImageFromURL(fmt.Sprintf("img_%d", i),
			fmt.Sprintf("https://example.com/img_%d.jpg", i))
	}

	goodMemoryAfter := getMemoryUsageMB()
	fmt.Printf("Memory usage with 10 image references: %.2f MB\n", goodMemoryAfter-goodMemoryBefore)

	// Lazy loading одного изображения
	if bytes, err := flowCtx.LoadImageBytes(context.Background(), "img_0"); err == nil {
		fmt.Printf("Loaded 1 image: %d bytes\n", len(bytes))
		finalMemory := getMemoryUsageMB()
		fmt.Printf("Memory after loading 1 image: %.2f MB\n", finalMemory-goodMemoryBefore)
	}
}

func RunAllMemoryDemos() {
	MemoryBloatPreventionExample()
	MemoryStressTest()
	CompareMemoryApproaches()
}