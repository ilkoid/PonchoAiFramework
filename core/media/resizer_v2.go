package media

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/config"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/disintegration/imaging"
)

// ResizerV2 implements comprehensive image resizing with YAML configuration support
type ResizerV2 struct {
	config   *config.ImageResizeConfig
	logger   interfaces.Logger
	cache    map[string][]byte
	cacheMux sync.RWMutex
	stats    *ResizeStats
}

// ResizeStats tracks resizing operation statistics
type ResizeStats struct {
	TotalProcessed    int64         `json:"total_processed"`
	TotalOriginalKB   int64         `json:"total_original_kb"`
	TotalResizedKB    int64         `json:"total_resized_kb"`
	CompressionRatio  float64       `json:"compression_ratio"`
	AvgProcessTime    time.Duration `json:"avg_process_time"`
	ErrorCount        int64         `json:"error_count"`
	CacheHits         int64         `json:"cache_hits"`
	CacheMisses       int64         `json:"cache_misses"`
	lastProcessTime   time.Duration
	processTimeCount  int64
	mux               sync.RWMutex
}

// ResizeResult contains the result of a resize operation
type ResizeResult struct {
	Data         []byte        `json:"-"`
	Format       string        `json:"format"`
	OriginalKB   int           `json:"original_kb"`
	ResizedKB    int           `json:"resized_kb"`
	ProcessTime  time.Duration `json:"process_time_ms"`
	StrategyName string        `json:"strategy_name"`
	CacheHit     bool          `json:"cache_hit"`
}

// BatchResizeRequest represents a batch resize request
type BatchResizeRequest struct {
	Images      map[string][]byte          `json:"images"`
	MimeTypes   map[string]string          `json:"mime_types"`
	Strategy    string                     `json:"strategy"`
	Quality     string                     `json:"quality_preset"`
	Parallel    bool                       `json:"parallel"`
	MaxConcurrency int                     `json:"max_concurrency"`
	Progress    chan *BatchResizeProgress  `json:"-"`
}

// BatchResizeProgress reports progress for batch operations
type BatchResizeProgress struct {
	Completed int   `json:"completed"`
	Total     int   `json:"total"`
	Current   string `json:"current"`
	Error     error `json:"error,omitempty"`
}

// NewResizerV2 creates a new ResizerV2 instance
func NewResizerV2(cfg *config.ImageResizeConfig, logger interfaces.Logger) *ResizerV2 {
	if cfg == nil {
		cfg = config.DefaultImageResizeConfig()
	}

	resizer := &ResizerV2{
		config: cfg,
		logger: logger,
		cache:  make(map[string][]byte),
		stats: &ResizeStats{
			CompressionRatio: 1.0,
		},
	}

	// Initialize default strategies if none provided
	if len(resizer.config.Strategies) == 0 {
		resizer.config.Strategies = resizer.getDefaultStrategies()
	}

	return resizer
}

// ResizeSingle resizes a single image using the specified strategy
func (r *ResizerV2) ResizeSingle(ctx context.Context, imageData []byte, mimeType string, modelName string) (*ResizeResult, error) {
	if !r.config.Enabled {
		return &ResizeResult{
			Data:       imageData,
			Format:     strings.Split(mimeType, "/")[1],
			CacheHit:   false,
		}, nil
	}

	startTime := time.Now()

	// Select strategy for model
	strategy := r.config.GetStrategyForModel(modelName)
	if strategy == nil || !strategy.Enabled {
		// Fallback to default strategy
		var err error
		strategy, err = r.config.GetStrategy(r.config.DefaultStrategy)
		if err != nil {
			strategy, _ = r.config.GetStrategy("vision_optimized")
		}
	}

	// Check cache if enabled
	cacheKey := r.generateCacheKey(imageData, strategy)
	if r.config.CacheSize > 0 {
		r.cacheMux.RLock()
		if cached, exists := r.cache[cacheKey]; exists {
			r.cacheMux.RUnlock()
			r.recordCacheHit()
			return &ResizeResult{
				Data:       cached,
				Format:     strategy.TargetFormat,
				CacheHit:   true,
			}, nil
		}
		r.cacheMux.RUnlock()
		r.recordCacheMiss()
	}

	// Check if resize is needed
	originalKB := len(imageData) / 1024
	if originalKB <= strategy.MaxFileSizeKB {
		return &ResizeResult{
			Data:       imageData,
			Format:     strings.Split(mimeType, "/")[1],
			OriginalKB: originalKB,
			ResizedKB:  originalKB,
			CacheHit:   false,
		}, nil
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		r.recordError()
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Apply resize strategy
	resizedImg, err := r.applyStrategy(img, strategy)
	if err != nil {
		r.recordError()
		return nil, fmt.Errorf("failed to apply resize strategy: %w", err)
	}

	// Encode to target format
	resizedData, err := r.encodeImage(resizedImg, strategy)
	if err != nil {
		r.recordError()
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Update cache
	if r.config.CacheSize > 0 {
		r.updateCache(cacheKey, resizedData)
	}

	processTime := time.Since(startTime)
	resizedKB := len(resizedData) / 1024

	// Update statistics
	r.updateStats(originalKB, resizedKB, processTime)

	result := &ResizeResult{
		Data:         resizedData,
		Format:       strategy.TargetFormat,
		OriginalKB:   originalKB,
		ResizedKB:    resizedKB,
		ProcessTime:  processTime,
		StrategyName: strategy.Name,
		CacheHit:     false,
	}

	if r.config.LogOperations {
		r.logger.Info("Image resized successfully",
			"strategy", strategy.Name,
			"original_kb", originalKB,
			"resized_kb", resizedKB,
			"compression_ratio", float64(resizedKB)/float64(originalKB),
			"process_time_ms", processTime.Milliseconds(),
		)
	}

	return result, nil
}

// ResizeBatch resizes multiple images in parallel or sequentially
func (r *ResizerV2) ResizeBatch(ctx context.Context, req *BatchResizeRequest, modelName string) (map[string]*ResizeResult, error) {
	if req == nil || len(req.Images) == 0 {
		return nil, fmt.Errorf("empty batch request")
	}

	results := make(map[string]*ResizeResult)
	errors := make(map[string]error)

	if req.Parallel && len(req.Images) > 1 {
		return r.resizeBatchParallel(ctx, req, modelName)
	}

	// Sequential processing
	for key, imageData := range req.Images {
		mimeType := req.MimeTypes[key]
		if mimeType == "" {
			mimeType = http.DetectContentType(imageData)
		}

		result, err := r.ResizeSingle(ctx, imageData, mimeType, modelName)
		if err != nil {
			errors[key] = err
			r.logger.Warn("Failed to resize image in batch",
				"key", key,
				"error", err,
			)
			continue
		}

		results[key] = result

		if req.Progress != nil {
			progress := &BatchResizeProgress{
				Completed: len(results) + len(errors),
				Total:     len(req.Images),
				Current:   key,
			}
			select {
			case req.Progress <- progress:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if len(errors) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("all images failed to resize")
	}

	return results, nil
}

// resizeBatchParallel processes images in parallel with worker pool
func (r *ResizerV2) resizeBatchParallel(ctx context.Context, req *BatchResizeRequest, modelName string) (map[string]*ResizeResult, error) {
	maxConcurrency := req.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = r.config.MaxConcurrency
	}

	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	results := make(map[string]*ResizeResult)
	resultsMux := sync.Mutex{}
	errors := make(map[string]error)
	errorsMux := sync.Mutex{}

	// completed := 0 // Пока не используется

	for key, imageData := range req.Images {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		wg.Add(1)
		go func(key string, data []byte) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			mimeType := req.MimeTypes[key]
			if mimeType == "" {
				mimeType = http.DetectContentType(data)
			}

			result, err := r.ResizeSingle(ctx, data, mimeType, modelName)
			if err != nil {
				errorsMux.Lock()
				errors[key] = err
				errorsMux.Unlock()
				return
			}

			resultsMux.Lock()
			results[key] = result
			resultsMux.Unlock()

			if req.Progress != nil {
				errorsMux.Lock()
				currentCompleted := len(results) + len(errors)
				errorsMux.Unlock()

				progress := &BatchResizeProgress{
					Completed: currentCompleted,
					Total:     len(req.Images),
					Current:   key,
				}
				select {
				case req.Progress <- progress:
				default:
				}
			}

		}(key, imageData)
	}

	wg.Wait()

	if len(errors) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("all images failed to resize")
	}

	return results, nil
}

// applyStrategy applies a resize strategy to an image
func (r *ResizerV2) applyStrategy(img image.Image, strategy *config.ResizeStrategy) (image.Image, error) {
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Check if resize is needed
	if originalWidth <= strategy.MaxWidth && originalHeight <= strategy.MaxHeight {
		return img, nil
	}

	// Calculate new dimensions maintaining aspect ratio
	widthRatio := float64(strategy.MaxWidth) / float64(originalWidth)
	heightRatio := float64(strategy.MaxHeight) / float64(originalHeight)

	var newWidth, newHeight int
	if widthRatio < heightRatio {
		newWidth = strategy.MaxWidth
		newHeight = int(float64(originalHeight) * widthRatio)
	} else {
		newHeight = strategy.MaxHeight
		newWidth = int(float64(originalWidth) * heightRatio)
	}

	// Apply smart crop if enabled
	if r.config.SmartCrop {
		img = imaging.Fill(img, newWidth, newHeight, imaging.Center, imaging.Lanczos)
	} else {
		img = imaging.Resize(img, newWidth, newHeight, r.getInterpolationFilter(strategy.Interpolation))
	}

	// Apply sharpening if configured
	if strategy.Sharpening != nil && strategy.Sharpening.Enabled {
		img = r.applySharpening(img, strategy.Sharpening)
	}

	// Apply noise reduction if configured
	if strategy.NoiseReduction != nil && strategy.NoiseReduction.Enabled {
		img = r.applyNoiseReduction(img, strategy.NoiseReduction)
	}

	return img, nil
}

// encodeImage encodes an image using the target format and quality
func (r *ResizerV2) encodeImage(img image.Image, strategy *config.ResizeStrategy) ([]byte, error) {
	var buf bytes.Buffer

	switch strings.ToLower(strategy.TargetFormat) {
	case "jpeg", "jpg":
		options := &jpeg.Options{Quality: strategy.Quality}
		err := jpeg.Encode(&buf, img, options)
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
		// Default to JPEG
		options := &jpeg.Options{Quality: strategy.Quality}
		err := jpeg.Encode(&buf, img, options)
		if err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}
	}

	return buf.Bytes(), nil
}

// getInterpolationFilter converts interpolation method to imaging filter
func (r *ResizerV2) getInterpolationFilter(method config.InterpolationMethod) imaging.ResampleFilter {
	switch method {
	case config.InterpolationNearestNeighbor:
		return imaging.NearestNeighbor
	case config.InterpolationBilinear:
		return imaging.Linear
	case config.InterpolationBicubic:
		return imaging.CatmullRom
	case config.InterpolationLanczos:
		return imaging.Lanczos
	default:
		return imaging.Lanczos
	}
}

// applySharpening applies sharpening to an image
func (r *ResizerV2) applySharpening(img image.Image, config *config.SharpeningConfig) image.Image {
	// Simple sharpening using imaging library
	return imaging.Sharpen(img, config.Amount)
}

// applyNoiseReduction applies noise reduction to an image
func (r *ResizerV2) applyNoiseReduction(img image.Image, config *config.NoiseReductionConfig) image.Image {
	// Simple blur for noise reduction
	if config.Smoothing {
		return imaging.Blur(img, config.Strength)
	}
	return img
}

// generateCacheKey creates a unique cache key for the image and strategy
func (r *ResizerV2) generateCacheKey(imageData []byte, strategy *config.ResizeStrategy) string {
	return fmt.Sprintf("%s_%dx%d_q%d_%s",
		hash(imageData),
		strategy.MaxWidth,
		strategy.MaxHeight,
		strategy.Quality,
		strategy.TargetFormat,
	)
}

// updateCache updates the image cache with size limit
func (r *ResizerV2) updateCache(key string, data []byte) {
	r.cacheMux.Lock()
	defer r.cacheMux.Unlock()

	// Check cache size limit
	cacheSize := int64(len(r.cache))
	if cacheSize >= r.config.CacheSize {
		// Simple LRU: remove oldest entry
		var oldestKey string
		for k := range r.cache {
			oldestKey = k
			break
		}
		if oldestKey != "" {
			delete(r.cache, oldestKey)
		}
	}

	r.cache[key] = data
}

// updateStats updates resize statistics
func (r *ResizerV2) updateStats(originalKB, resizedKB int, processTime time.Duration) {
	r.stats.mux.Lock()
	defer r.stats.mux.Unlock()

	r.stats.TotalProcessed++
	r.stats.TotalOriginalKB += int64(originalKB)
	r.stats.TotalResizedKB += int64(resizedKB)
	r.stats.CompressionRatio = float64(r.stats.TotalResizedKB) / float64(r.stats.TotalOriginalKB)

	// Update average process time
	r.stats.lastProcessTime = processTime
	r.stats.processTimeCount++
	r.stats.AvgProcessTime = time.Duration(
		(int64(r.stats.AvgProcessTime)* (r.stats.processTimeCount-1) + int64(processTime)) / r.stats.processTimeCount,
	)
}

// recordCacheHit increments cache hit counter
func (r *ResizerV2) recordCacheHit() {
	r.stats.mux.Lock()
	defer r.stats.mux.Unlock()
	r.stats.CacheHits++
}

// recordCacheMiss increments cache miss counter
func (r *ResizerV2) recordCacheMiss() {
	r.stats.mux.Lock()
	defer r.stats.mux.Unlock()
	r.stats.CacheMisses++
}

// recordError increments error counter
func (r *ResizerV2) recordError() {
	r.stats.mux.Lock()
	defer r.stats.mux.Unlock()
	r.stats.ErrorCount++
}

// GetStats returns current resize statistics
func (r *ResizerV2) GetStats() *ResizeStats {
	r.stats.mux.RLock()
	defer r.stats.mux.RUnlock()

	// Return a copy to avoid concurrent access issues
	return &ResizeStats{
		TotalProcessed:   r.stats.TotalProcessed,
		TotalOriginalKB:  r.stats.TotalOriginalKB,
		TotalResizedKB:   r.stats.TotalResizedKB,
		CompressionRatio: r.stats.CompressionRatio,
		AvgProcessTime:   r.stats.AvgProcessTime,
		ErrorCount:       r.stats.ErrorCount,
		CacheHits:        r.stats.CacheHits,
		CacheMisses:      r.stats.CacheMisses,
	}
}

// ClearCache clears the image cache
func (r *ResizerV2) ClearCache() {
	r.cacheMux.Lock()
	defer r.cacheMux.Unlock()
	r.cache = make(map[string][]byte)
}

// ResetStats resets resize statistics
func (r *ResizerV2) ResetStats() {
	r.stats.mux.Lock()
	defer r.stats.mux.Unlock()
	r.stats = &ResizeStats{}
}

// getDefaultStrategies returns default resize strategies
func (r *ResizerV2) getDefaultStrategies() map[string]*config.ResizeStrategy {
	return map[string]*config.ResizeStrategy{
		"vision_optimized": {
			Name:         "vision_optimized",
			Description:  "Optimized for vision models (1024x1024, high quality)",
			MaxWidth:     1024,
			MaxHeight:    1024,
			MaxFileSizeKB: 500,
			Quality:      85,
			TargetFormat: "jpeg",
			Enabled:      true,
			Priority:     100,
			Interpolation: config.InterpolationLanczos,
		},
		"text_optimized": {
			Name:         "text_optimized",
			Description:  "Optimized for text analysis (512x512, medium quality)",
			MaxWidth:     512,
			MaxHeight:    512,
			MaxFileSizeKB: 200,
			Quality:      75,
			TargetFormat: "jpeg",
			Enabled:      true,
			Priority:     50,
			Interpolation: config.InterpolationBicubic,
		},
		"code_optimized": {
			Name:         "code_optimized",
			Description:  "Optimized for code screenshots (800x600, high quality)",
			MaxWidth:     800,
			MaxHeight:    600,
			MaxFileSizeKB: 300,
			Quality:      90,
			TargetFormat: "png",
			Enabled:      true,
			Priority:     75,
			Interpolation: config.InterpolationLanczos,
		},
		"high_quality": {
			Name:         "high_quality",
			Description:  "High quality for detailed analysis (2048x2048, maximum quality)",
			MaxWidth:     2048,
			MaxHeight:    2048,
			MaxFileSizeKB: 1024,
			Quality:      95,
			TargetFormat: "png",
			Enabled:      true,
			Priority:     90,
			Interpolation: config.InterpolationLanczos,
		},
	}
}

// hash creates a simple hash for cache key generation
func hash(data []byte) string {
	var hash uint32
	for _, b := range data {
		hash = hash*31 + uint32(b)
	}
	return fmt.Sprintf("%x", hash)
}

// ResizeToTempFile resizes an image and saves it to a temporary file
func (r *ResizerV2) ResizeToTempFile(ctx context.Context, imageData []byte, mimeType string, modelName string) (string, error) {
	result, err := r.ResizeSingle(ctx, imageData, mimeType, modelName)
	if err != nil {
		return "", err
	}

	// Create temporary file with appropriate extension
	ext := "." + result.Format
	if result.Format == "jpg" {
		ext = ".jpeg"
	}

	tmpFile, err := os.CreateTemp("", "poncho_resized_*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	_, err = tmpFile.Write(result.Data)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

// GetOptimalFormat determines the best output format for the given content
func (r *ResizerV2) GetOptimalFormat(mimeType string) string {
	// For photographs, JPEG is usually best
	if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
		return "jpeg"
	}

	// For images with text or diagrams, PNG is better
	if strings.Contains(mimeType, "png") || strings.Contains(mimeType, "gif") {
		return "png"
	}

	// Default to JPEG for compatibility
	return "jpeg"
}

// ValidateResizeConfig validates the resize configuration
func (r *ResizerV2) ValidateResizeConfig() error {
	return r.config.Validate()
}