package media

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/config"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResizerV2(t *testing.T) {
	logger := interfaces.NewDefaultLogger()

	// Test with default config
	resizer := NewResizerV2(nil, logger)
	assert.NotNil(t, resizer)
	assert.NotNil(t, resizer.config)
	assert.NotNil(t, resizer.stats)
	assert.True(t, len(resizer.config.Strategies) > 0)

	// Test with custom config
	cfg := config.DefaultImageResizeConfig()
	cfg.Enabled = false
	resizer2 := NewResizerV2(cfg, logger)
	assert.False(t, resizer2.config.Enabled)
}

func TestResizerV2_SingleResize(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	resizer := NewResizerV2(cfg, logger)
	ctx := context.Background()

	// Create test image (1MB)
	imageData := makeTestImage(1024 * 1024)
	mimeType := "image/jpeg"

	// Test resize enabled
	result, err := resizer.ResizeSingle(ctx, imageData, mimeType, "glm-vision")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, len(result.Data) > 0)
	assert.Equal(t, "jpeg", result.Format)
	assert.True(t, result.ResizedKB < result.OriginalKB) // Should be smaller

	// Test resize disabled
	cfg.Enabled = false
	resizerDisabled := NewResizerV2(cfg, logger)
	result2, err := resizerDisabled.ResizeSingle(ctx, imageData, mimeType, "glm-vision")
	require.NoError(t, err)
	assert.Equal(t, len(imageData), len(result2.Data)) // Should be unchanged
}

func TestResizerV2_BatchResize(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	resizer := NewResizerV2(cfg, logger)
	ctx := context.Background()

	// Create test images
	images := map[string][]byte{
		"img1": makeTestImage(2 * 1024 * 1024), // 2MB
		"img2": makeTestImage(3 * 1024 * 1024), // 3MB
		"img3": makeTestImage(1 * 1024 * 1024), // 1MB
	}
	mimeTypes := map[string]string{
		"img1": "image/jpeg",
		"img2": "image/jpeg",
		"img3": "image/jpeg",
	}

	// Test sequential batch resize
	req := &BatchResizeRequest{
		Images:    images,
		MimeTypes: mimeTypes,
		Parallel:  false,
	}

	results, err := resizer.ResizeBatch(ctx, req, "glm-vision")
	require.NoError(t, err)
	assert.Len(t, results, 3)

	for key, result := range results {
		assert.True(t, result.ResizedKB > 0)
		assert.Equal(t, "jpeg", result.Format)
		t.Logf("  %s: %d KB → %d KB", key, result.OriginalKB, result.ResizedKB)
	}
}

func TestResizerV2_ParallelBatchResize(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	cfg.MaxConcurrency = 2
	resizer := NewResizerV2(cfg, logger)
	ctx := context.Background()

	// Create test images
	images := map[string][]byte{
		"img1": makeTestImage(2 * 1024 * 1024), // 2MB
		"img2": makeTestImage(2 * 1024 * 1024), // 2MB
		"img3": makeTestImage(2 * 1024 * 1024), // 2MB
	}
	mimeTypes := map[string]string{
		"img1": "image/jpeg",
		"img2": "image/jpeg",
		"img3": "image/jpeg",
	}

	// Test parallel batch resize
	progress := make(chan *BatchResizeProgress)
	req := &BatchResizeRequest{
		Images:         images,
		MimeTypes:      mimeTypes,
		Parallel:       true,
		MaxConcurrency: 2,
		Progress:       progress,
	}

	resultChan := make(chan map[string]*ResizeResult)
	errorChan := make(chan error)

	go func() {
		results, err := resizer.ResizeBatch(ctx, req, "glm-vision")
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- results
	}()

	// Monitor progress
	progressCount := 0
	for p := range progress {
		progressCount++
		assert.True(t, p.Completed <= 3)
		t.Logf("  Progress: %d/%d (%s)", p.Completed, p.Total, p.Current)
	}

	select {
	case results := <-resultChan:
		assert.Len(t, results, 3)
		for key, result := range results {
			assert.True(t, result.ResizedKB > 0)
			t.Logf("  %s: %d KB → %d KB", key, result.OriginalKB, result.ResizedKB)
		}
	case err := <-errorChan:
		t.Errorf("Parallel batch resize failed: %v", err)
	case <-time.After(10 * time.Second):
		t.Error("Parallel batch resize timed out")
	}
}

func TestResizerV2_StrategySelection(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	resizer := NewResizerV2(cfg, logger)
	ctx := context.Background()

	imageData := makeTestImage(5 * 1024 * 1024) // 5MB
	mimeType := "image/jpeg"

	// Test different model strategies
	testCases := []struct {
		modelName       string
		expectedStrategy string
	}{
		{"glm-vision", "vision_optimized"},
		{"glm-4.6v-flash", "vision_optimized"},
		{"deepseek-chat", "text_optimized"},
		{"deepseek-coder", "code_optimized"},
		{"unknown-model", "vision_optimized"}, // Should fallback to default
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			result, err := resizer.ResizeSingle(ctx, imageData, mimeType, tc.modelName)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStrategy, result.StrategyName)
			t.Logf("  %s: used strategy %s", tc.modelName, result.StrategyName)
		})
	}
}

func TestResizerV2_Cache(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	cfg.CacheSize = 10 // 10MB cache
	resizer := NewResizerV2(cfg, logger)
	ctx := context.Background()

	imageData := makeTestImage(1024 * 1024) // 1MB
	mimeType := "image/jpeg"

	// First resize - should cache miss
	result1, err := resizer.ResizeSingle(ctx, imageData, mimeType, "glm-vision")
	require.NoError(t, err)

	stats1 := resizer.GetStats()
	assert.Equal(t, int64(0), stats1.CacheHits)
	assert.Equal(t, int64(1), stats1.CacheMisses)

	// Second resize of same image - should cache hit
	result2, err := resizer.ResizeSingle(ctx, imageData, mimeType, "glm-vision")
	require.NoError(t, err)
	assert.True(t, result2.CacheHit)

	stats2 := resizer.GetStats()
	assert.Equal(t, int64(1), stats2.CacheHits)
	assert.Equal(t, int64(1), stats2.CacheMisses)

	// Clear cache
	resizer.ClearCache()

	// Third resize - should cache miss again
	result3, err := resizer.ResizeSingle(ctx, imageData, mimeType, "glm-vision")
	require.NoError(t, err)
	assert.False(t, result3.CacheHit)

	stats3 := resizer.GetStats()
	assert.Equal(t, int64(1), stats3.CacheHits) // Still 1 from before
	assert.Equal(t, int64(2), stats3.CacheMisses)
}

func TestResizerV2_Statistics(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	resizer := NewResizerV2(cfg, logger)
	ctx := context.Background()

	// Reset stats
	resizer.ResetStats()
	stats := resizer.GetStats()
	assert.Equal(t, int64(0), stats.TotalProcessed)

	// Process some images
	for i := 0; i < 3; i++ {
		imageData := makeTestImage(1024 * 1024)
		_, err := resizer.ResizeSingle(ctx, imageData, "image/jpeg", "glm-vision")
		require.NoError(t, err)
	}

	// Check stats
	stats = resizer.GetStats()
	assert.Equal(t, int64(3), stats.TotalProcessed)
	assert.True(t, stats.TotalOriginalKB > 0)
	assert.True(t, stats.TotalResizedKB > 0)
	assert.True(t, stats.CompressionRatio < 1.0) // Should be compressed
	assert.True(t, stats.AvgProcessTime > 0)
	assert.Equal(t, int64(0), stats.ErrorCount)

	t.Logf("  Statistics:")
	t.Logf("    Total processed: %d", stats.TotalProcessed)
	t.Logf("    Original size: %d KB", stats.TotalOriginalKB)
	t.Logf("    Resized size: %d KB", stats.TotalResizedKB)
	t.Logf("    Compression ratio: %.2f", stats.CompressionRatio)
	t.Logf("    Average time: %v", stats.AvgProcessTime)
}

func TestResizerV2_TempFile(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	resizer := NewResizerV2(cfg, logger)
	ctx := context.Background()

	imageData := makeTestImage(2 * 1024 * 1024) // 2MB

	tempFile, err := resizer.ResizeToTempFile(ctx, imageData, "image/jpeg", "glm-vision")
	require.NoError(t, err)
	assert.NotEmpty(t, tempFile)
	assert.True(t, strings.Contains(tempFile, ".jpeg") || strings.Contains(tempFile, ".jpg"))

	t.Logf("  Created temp file: %s", tempFile)

	// Clean up
	defer os.Remove(tempFile)
}

func TestResizerV2_ConfigValidation(t *testing.T) {
	logger := interfaces.NewDefaultLogger()

	// Test invalid config
	cfg := &config.ImageResizeConfig{
		Enabled:     true,
		MaxMemoryMB: -1, // Invalid
	}

	resizer := NewResizerV2(cfg, logger)
	err := resizer.ValidateResizeConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max_memory_mb must be positive")

	// Test valid config
	cfg2 := config.DefaultImageResizeConfig()
	resizer2 := NewResizerV2(cfg2, logger)
	err = resizer2.ValidateResizeConfig()
	assert.NoError(t, err)
}

// makeTestImage creates test image data with a pattern
func makeTestImage(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// BenchmarkResizeSingle benchmarks single image resize
func BenchmarkResizeSingle(b *testing.B) {
	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	resizer := NewResizerV2(cfg, logger)
	ctx := context.Background()

	// 1MB test image
	imageData := makeTestImage(1024 * 1024)
	mimeType := "image/jpeg"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resizer.ResizeSingle(ctx, imageData, mimeType, "glm-vision")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBatchResize benchmarks batch image resize
func BenchmarkBatchResize(b *testing.B) {
	logger := interfaces.NewDefaultLogger()
	cfg := config.DefaultImageResizeConfig()
	resizer := NewResizerV2(cfg, logger)
	ctx := context.Background()

	// Create test images
	images := map[string][]byte{
		"img1": makeTestImage(512 * 1024),  // 512KB
		"img2": makeTestImage(512 * 1024),  // 512KB
		"img3": makeTestImage(512 * 1024),  // 512KB
	}
	mimeTypes := map[string]string{
		"img1": "image/jpeg",
		"img2": "image/jpeg",
		"img3": "image/jpeg",
	}

	req := &BatchResizeRequest{
		Images:    images,
		MimeTypes: mimeTypes,
		Parallel:  false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resizer.ResizeBatch(ctx, req, "glm-vision")
		if err != nil {
			b.Fatal(err)
		}
	}
}