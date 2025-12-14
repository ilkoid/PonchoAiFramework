package examples

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/config"
	"github.com/ilkoid/PonchoAiFramework/core/context"
	"github.com/ilkoid/PonchoAiFramework/core/media"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ResizeIntegrationExample demonstrates how to use the new resize functionality
func ResizeIntegrationExample() {
	fmt.Println("=== Resize Integration Example ===")

	// Create logger
	logger := interfaces.NewDefaultLogger()

	// Load configuration from YAML
	fmt.Println("\n📋 Loading resize configuration...")
	cfg := config.DefaultImageResizeConfig()

	// Override with custom settings for demonstration
	cfg.AutoResize = true
	cfg.Monitoring = true
	cfg.ModelStrategies = map[string]string{
		"glm-vision":      "vision_optimized",
		"glm-4.6v-flash":  "vision_optimized",
		"deepseek-chat":    "text_optimized",
		"deepseek-coder":   "code_optimized",
	}

	// Create resizer
	fmt.Println("🔧 Creating ResizerV2...")
	resizer := media.NewResizerV2(cfg, logger)

	// Create FlowContext with resize capabilities
	fmt.Println("📦 Creating FlowContextV2 with resize support...")
	flowCtx := context.NewBaseFlowContextV2()
	ctx := context.Background()

	// Example 1: Single image resize with vision model
	fmt.Println("\n🔍 Example 1: Single image resize for vision model")
	singleImageExample(ctx, resizer, flowCtx, "glm-vision")

	// Example 2: Batch resize for multiple images
	fmt.Println("\n📚 Example 2: Batch resize for multiple images")
	batchResizeExample(ctx, resizer, flowCtx, "deepseek-chat")

	// Example 3: Flow-specific resize configuration
	fmt.Println("\n🎯 Example 3: Flow-specific resize configuration")
	flowSpecificExample(ctx, resizer, cfg, "article_processor")

	// Example 4: Memory-safe resize with ImageReference
	fmt.Println("\n🛡️ Example 4: Memory-safe resize with ImageReference")
	memorySafeExample(ctx, resizer, flowCtx)

	// Display statistics
	fmt.Println("\n📊 Resize Statistics:")
	stats := resizer.GetStats()
	fmt.Printf("  Total processed: %d\n", stats.TotalProcessed)
	fmt.Printf("  Original size: %d KB\n", stats.TotalOriginalKB)
	fmt.Printf("  Resized size: %d KB\n", stats.TotalResizedKB)
	fmt.Printf("  Compression ratio: %.2fx\n", stats.CompressionRatio)
	fmt.Printf("  Average processing time: %v\n", stats.AvgProcessTime)
	fmt.Printf("  Cache hits: %d\n", stats.CacheHits)
	fmt.Printf("  Cache misses: %d\n", stats.CacheMisses)
	fmt.Printf("  Errors: %d\n", stats.ErrorCount)

	// Clean up
	resizer.ClearCache()
	flowCtx.Clear()
}

func singleImageExample(ctx context.Context, resizer *media.ResizerV2, flowCtx *context.BaseFlowContextV2, modelName string) {
	// Simulate image data (5MB JPEG)
	imageData := make([]byte, 5*1024*1024)
	for i := range imageData {
		imageData[i] = byte(i % 256)
	}
	mimeType := "image/jpeg"

	fmt.Printf("  Original image: %d KB\n", len(imageData)/1024)

	// Resize using ResizerV2
	result, err := resizer.ResizeSingle(ctx, imageData, mimeType, modelName)
	if err != nil {
		fmt.Printf("  ❌ Resize failed: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Resized image: %d KB (%s)\n", len(result.Data)/1024, result.Format)
	fmt.Printf("  📐 Strategy: %s\n", result.StrategyName)
	fmt.Printf("  ⏱️  Processing time: %v\n", result.ProcessTime)
	fmt.Printf("  📦 Compression: %.2fx\n", float64(result.OriginalKB)/float64(result.ResizedKB))

	// Store in FlowContext using safe approach
	err = flowCtx.SetBytes("resized_image", result.Data)
	if err != nil {
		fmt.Printf("  ❌ Failed to store in FlowContext: %v\n", err)
		return
	}

	fmt.Printf("  💾 Stored in FlowContext with key: 'resized_image'\n")
}

func batchResizeExample(ctx context.Context, resizer *media.ResizerV2, flowCtx *context.BaseFlowContextV2, modelName string) {
	// Create multiple test images
	images := make(map[string][]byte)
	mimeTypes := make(map[string]string)

	sizes := []int{2, 3, 4, 5, 8] // Different sizes in MB
	for i, size := range sizes {
		key := fmt.Sprintf("image_%d", i)
		images[key] = make([]byte, size*1024*1024)
		mimeTypes[key] = "image/jpeg"

		// Fill with pattern
		for j := range images[key] {
			images[key][j] = byte((i + j) % 256)
		}
	}

	fmt.Printf("  Created %d test images (total: %d KB)\n", len(images), calculateTotalSize(images)/1024)

	// Create batch resize request
	req := &media.BatchResizeRequest{
		Images:      images,
		MimeTypes:   mimeTypes,
		Strategy:    "text_optimized",
		Parallel:    true,
		MaxConcurrency: 3,
		Progress:    make(chan *media.BatchResizeProgress),
	}

	// Process in goroutine to handle progress
	startTime := time.Now()
	resultChan := make(chan map[string]*media.ResizeResult)
	errorChan := make(chan error)

	go func() {
		results, err := resizer.ResizeBatch(ctx, req, modelName)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- results
	}()

	// Monitor progress
	for progress := range req.Progress {
		if progress.Error != nil {
			fmt.Printf("  ⚠️  Error processing %s: %v\n", progress.Current, progress.Error)
		} else {
			fmt.Printf("  📈 Progress: %d/%d completed (%s)\n", progress.Completed, progress.Total, progress.Current)
		}
	}

	// Get results
	select {
	case results := <-resultChan:
		duration := time.Since(startTime)
		totalOriginal := 0
		totalResized := 0

		for key, result := range results {
			totalOriginal += result.OriginalKB
			totalResized += result.ResizedKB

			// Store in FlowContext
			err := flowCtx.SetBytes(key+"_resized", result.Data)
			if err != nil {
				fmt.Printf("  ❌ Failed to store %s: %v\n", key, err)
				continue
			}
		}

		fmt.Printf("  ✅ Batch completed in %v\n", duration)
		fmt.Printf("  📊 Total: %d KB → %d KB (%.2fx reduction)\n",
			totalOriginal, totalResized, float64(totalOriginal)/float64(totalResized))

	case err := <-errorChan:
		fmt.Printf("  ❌ Batch failed: %v\n", err)
	}

	close(req.Progress)
	close(resultChan)
	close(errorChan)
}

func flowSpecificExample(ctx context.Context, resizer *media.ResizerV2, cfg *config.ImageResizeConfig, flowName string) {
	// Get flow-specific configuration
	flowConfig, err := cfg.GetFlowConfig(flowName)
	if err != nil {
		fmt.Printf("  ❌ Failed to get flow config: %v\n", err)
		return
	}

	fmt.Printf("  📋 Flow '%s' configuration:\n", flowName)
	fmt.Printf("    Enabled: %t\n", flowConfig.Enabled)
	fmt.Printf("    Strategy: %s\n", flowConfig.StrategyName)
	fmt.Printf("    Quality preset: %s\n", flowConfig.QualityPreset)
	fmt.Printf("    Max images: %d\n", flowConfig.MaxImages)
	fmt.Printf("    Memory limit: %d MB\n", flowConfig.MemoryLimitMB)

	// Simulate flow processing
	if flowConfig.Enabled {
		// Get the strategy for this flow
		strategy, err := cfg.GetStrategy(flowConfig.StrategyName)
		if err != nil {
			fmt.Printf("  ❌ Failed to get strategy: %v\n", err)
			return
		}

		// Create test image for this flow
		testImage := make([]byte, 3*1024*1024) // 3MB
		for i := range testImage {
			testImage[i] = byte(i % 128)
		}

		// Resize with flow-specific strategy
		result, err := resizer.ResizeSingle(ctx, testImage, "image/jpeg", "glm-vision")
		if err != nil {
			fmt.Printf("  ❌ Flow resize failed: %v\n", err)
			return
		}

		fmt.Printf("  ✅ Flow resize completed: %d KB → %d KB\n",
			result.OriginalKB, result.ResizedKB)
		fmt.Printf("  🎯 Used strategy: %s\n", strategy.Description)
	}
}

func memorySafeExample(ctx context.Context, resizer *media.ResizerV2, flowCtx *context.BaseFlowContextV2) {
	// Create ImageReference instead of storing binary data
	imageURLs := []string{
		"https://example.com/fashion1.jpg",
		"https://example.com/fashion2.jpg",
		"https://example.com/fashion3.jpg",
	}

	fmt.Printf("  📷 Creating ImageReferences for %d URLs\n", len(imageURLs))

	for i, url := range imageURLs {
		key := fmt.Sprintf("fashion_image_%d", i)

		// Create ImageReference (stores only URL, not binary data)
		flowCtx.SetImageFromURL(key, url)
		fmt.Printf("    ✅ Created ImageReference: %s\n", key)
	}

	// Show memory usage is minimal
	memoryUsage := flowCtx.GetMemoryUsage()
	fmt.Printf("  💾 Memory usage after storing ImageReferences: %d KB\n", memoryUsage/1024)

	// Lazy loading example
	fmt.Println("  🔄 Lazy loading example:")
	for i := 0; i < len(imageURLs); i++ {
		key := fmt.Sprintf("fashion_image_%d", i)

		// This would normally load the image, but since we don't have real URLs,
		// we'll simulate with a test image
		testImageData := make([]byte, 1024*1024) // 1MB

		// Resize the loaded image
		result, err := resizer.ResizeSingle(ctx, testImageData, "image/jpeg", "glm-vision")
		if err != nil {
			fmt.Printf("    ❌ Failed to resize %s: %v\n", key, err)
			continue
		}

		fmt.Printf("    ✅ Loaded and resized %s: %d KB\n", key, len(result.Data)/1024)

		// Store the resized image back to context
		err = flowCtx.SetBytes(key+"_resized", result.Data)
		if err != nil {
			fmt.Printf("    ❌ Failed to store resized %s: %v\n", key, err)
		}
	}

	finalMemory := flowCtx.GetMemoryUsage()
	fmt.Printf("  💾 Final memory usage: %d KB\n", finalMemory/1024)
}

func calculateTotalSize(images map[string][]byte) int {
	total := 0
	for _, data := range images {
		total += len(data)
	}
	return total
}

// AdvancedUsageExample demonstrates advanced features
func AdvancedUsageExample() {
	fmt.Println("\n=== Advanced Resize Usage ===")

	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	resizer := media.NewResizerV2(cfg, logger)
	ctx := context.Background()

	// Example 1: Custom strategy selection
	fmt.Println("\n🎯 Custom Strategy Selection:")
	strategies := []string{"vision_optimized", "text_optimized", "code_optimized", "high_quality"}
	testImage := make([]byte, 4*1024*1024) // 4MB

	for _, strategyName := range strategies {
		result, err := resizer.ResizeSingle(ctx, testImage, "image/jpeg", "glm-vision")
		if err != nil {
			fmt.Printf("  ❌ Strategy %s failed: %v\n", strategyName, err)
			continue
		}

		fmt.Printf("  ✅ %s: %d KB → %d KB (%.2fx)\n",
			strategyName, result.OriginalKB, result.ResizedKB,
			float64(result.OriginalKB)/float64(result.ResizedKB))
	}

	// Example 2: Temp file usage for large images
	fmt.Println("\n💾 Temp File Usage for Large Images:")
	largeImage := make([]byte, 15*1024*1024) // 15MB

	tempFile, err := resizer.ResizeToTempFile(ctx, largeImage, "image/jpeg", "glm-vision")
	if err != nil {
		fmt.Printf("  ❌ Temp file creation failed: %v\n", err)
	} else {
		fmt.Printf("  ✅ Created temp file: %s\n", tempFile)
		// Remember to clean up temp files!
		os.Remove(tempFile)
	}

	// Example 3: Format conversion
	fmt.Println("\n🔄 Format Conversion Example:")
	pngImage := make([]byte, 2*1024*1024) // 2MB PNG

	result, err := resizer.ResizeSingle(ctx, pngImage, "image/png", "deepseek-coder")
	if err != nil {
		fmt.Printf("  ❌ Format conversion failed: %v\n", err)
	} else {
		fmt.Printf("  ✅ PNG → %s: %d KB → %d KB\n",
			result.Format, result.OriginalKB, result.ResizedKB)
	}
}

// PerformanceBenchmark runs performance tests
func PerformanceBenchmark() {
	fmt.Println("\n=== Performance Benchmark ===")

	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	cfg.Monitoring = true
	cfg.LogOperations = true
	resizer := media.NewResizerV2(cfg, logger)
	ctx := context.Background()

	// Test with different image sizes
	imageSizes := []int{1, 2, 5, 10, 20} // MB
	strategy := "vision_optimized"

	fmt.Printf("\n📊 Performance by Image Size (%s strategy):\n", strategy)
	fmt.Printf("%-10s | %-15s | %-15s | %-10s\n", "Size(MB)", "Original(KB)", "Resized(KB)", "Time(ms)")
	fmt.Println(strings.Repeat("-", 55))

	for _, sizeMB := range imageSizes {
		imageData := make([]byte, sizeMB*1024*1024)

		startTime := time.Now()
		result, err := resizer.ResizeSingle(ctx, imageData, "image/jpeg", "glm-vision")
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("%-10d | %-15s | %-15s | %-10s\n", sizeMB, "ERROR", "ERROR", "ERROR")
			continue
		}

		fmt.Printf("%-10d | %-15d | %-15d | %-10d\n",
			sizeMB, result.OriginalKB, result.ResizedKB, duration.Milliseconds())
	}

	// Show final statistics
	stats := resizer.GetStats()
	fmt.Printf("\n📈 Benchmark Summary:\n")
	fmt.Printf("  Total images processed: %d\n", stats.TotalProcessed)
	fmt.Printf("  Average processing time: %v\n", stats.AvgProcessTime)
	fmt.Printf("  Overall compression ratio: %.2fx\n", stats.CompressionRatio)
	fmt.Printf("  Cache efficiency: %.1f%%\n",
		float64(stats.CacheHits)/float64(stats.CacheHits+stats.CacheMisses)*100)
}

func RunAllResizeExamples() {
	ResizeIntegrationExample()
	AdvancedUsageExample()
	PerformanceBenchmark()
}