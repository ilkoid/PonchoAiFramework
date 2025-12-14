package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ImageResizeConfig представляет полную конфигурацию ресайза изображений
type ImageResizeConfig struct {
	// Глобальные настройки
	Enabled    bool             `yaml:"enabled" json:"enabled"`
	MaxMemoryMB int64            `yaml:"max_memory_mb" json:"max_memory_mb"`
	CacheSize  int64            `yaml:"cache_size_mb" json:"cache_size_mb"`

	// Стратегии ресайза по умолчанию
	DefaultStrategy string          `yaml:"default_strategy" json:"default_strategy"`
	Strategies    map[string]*ResizeStrategy `yaml:"strategies" json:"strategies"`

	// Автоматическое применение для моделей
	AutoResize     bool             `yaml:"auto_resize" json:"auto_resize"`
	ModelStrategies map[string]string `yaml:"model_strategies" json:"model_strategies"`

	// Batch processing
	BatchSize      int              `yaml:"batch_size" json:"batch_size"`
	ParallelResize bool            `yaml:"parallel_resize" json:"parallel_resize"`
	MaxConcurrency  int              `yaml:"max_concurrency" json:"max_concurrency"`

	// Quality настройки
	QualityPresets  map[string]*ResizeStrategy `yaml:"quality_presets" json:"quality_presets"`
	DefaultPreset   string          `yaml:"default_preset" json:"default_preset"`

	// Format настройки
	DefaultFormat   string          `yaml:"default_format" json:"default_format"`
	ConvertFormats   map[string]string `yaml:"convert_formats" json:"convert_formats"`

	// Thresholds и лимиты
	SizeThresholds  *SizeThresholds   `yaml:"size_thresholds" json:"size_thresholds"`
	Limits          *ResizeLimits     `yaml:"limits" json:"limits"`

	// Advanced options
	SmartCrop        bool            `yaml:"smart_crop" json:"smart_crop"`
	PreserveMetadata bool            `yaml:"preserve_metadata" json:"preserve_metadata"`
	ProgressiveJPEG bool            `yaml:"progressive_jpeg" json:"progressive_jpeg"`

	// Monitoring
	Monitoring      bool            `yaml:"monitoring" json:"monitoring"`
	LogOperations   bool            `yaml:"log_operations" json:"log_operations"`
	Profiling       bool            `yaml:"profiling" json:"profiling"`

	// Настройки для flow
	Flows           map[string]*FlowResizeConfig `yaml:"flows" json:"flows"`
}

// ResizeStrategy определяет стратегию ресайза изображений
type ResizeStrategy struct {
	Name         string  `yaml:"name" json:"name"`
	Description  string  `yaml:"description" json:"description"`
	MaxWidth     int     `yaml:"max_width" json:"max_width"`
	MaxHeight    int     `yaml:"max_height" json:"max_height"`
	MaxFileSizeKB int     `yaml:"max_file_size_kb" json:"max_file_size_kb"`
	Quality      int     `yaml:"quality" json:"quality"`
	TargetFormat  string  `yaml:"target_format" json:"target_format"`
	Enabled      bool    `yaml:"enabled" json:"enabled"`
	Priority     int     `yaml:"priority" json:"priority"` // Для выбора оптимальной стратегии

	// Продвинутые настройки
	Interpolation InterpolationMethod `yaml:"interpolation" json:"interpolation"`
	Sharpening    *SharpeningConfig     `yaml:"sharpening" json:"sharpening"`
	NoiseReduction *NoiseReductionConfig `yaml:"noise_reduction" json:"noise_reduction"`
}

// InterpolationMethod определяет метод интерполяции
type InterpolationMethod string

const (
	InterpolationNearestNeighbor InterpolationMethod = "nearest"
	InterpolationBilinear      InterpolationMethod = "bilinear"
	InterpolationBicubic        InterpolationMethod = "bicubic"
	InterpolationLanczos         InterpolationMethod = "lanczos"
)

// SharpeningConfig определяет настройки резкости
type SharpeningConfig struct {
	Enabled   bool    `yaml:"enabled" json:"enabled"`
	Amount    float64 `yaml:"amount" json:"amount"`
	Radius    float64 `yaml:"radius" json:"radius"`
	Threshold float64 `yaml:"threshold" json:"threshold"`
}

// NoiseReductionConfig определяет настройки шумоподавления
type NoiseReductionConfig struct {
	Enabled   bool    `yaml:"enabled" json:"enabled"`
	Strength  float64 `yaml:"strength" json:"strength"`
	Smoothing bool    `yaml:"smoothing" json:"smoothing"`
}

// SizeThresholds определяет пороги для принятия решений
type SizeThresholds struct {
	SmallImageKB   int   `yaml:"small_image_kb" json:"small_image_kb"`   // < порога - не ресайзить
	MediumImageKB  int   `yaml:"medium_image_kb" json:"medium_image_kb"`  // Средние - стандартный ресайз
	LargeImageKB   int   `yaml:"large_image_kb" json:"large_image_kb"`    // Большие - агрессивный ресайз
	HugeImageKB    int   `yaml:"huge_image_kb" json:"huge_image_kb"`      // Огромные - temp files
}

// ResizeLimits определяет лимиты безопасности
type ResizeLimits struct {
	MaxImagesPerContext    int     `yaml:"max_images_per_context" json:"max_images_per_context"`
	MaxMemoryPerContextMB   int64   `yaml:"max_memory_per_context_mb" json:"max_memory_per_context_mb"`
	MaxConcurrentResizes    int     `yaml:"max_concurrent_resizes" json:"max_concurrent_resizes"`
	ResizeTimeoutSeconds      int     `yaml:"resize_timeout_seconds" json:"resize_timeout_seconds"`
	MaxBatchSize              int     `yaml:"max_batch_size" json:"max_batch_size"`
}

// FlowResizeConfig определяет настройки ресайза для конкретного flow
type FlowResizeConfig struct {
	Enabled         bool   `yaml:"enabled" json:"enabled"`
	StrategyName    string `yaml:"strategy_name" json:"strategy_name"`
	QualityPreset   string `yaml:"quality_preset" json:"quality_preset"`
	MaxImages       int    `yaml:"max_images" json:"max_images"`
	MemoryLimitMB    int64  `yaml:"memory_limit_mb" json:"memory_limit_mb"`
}

// DefaultImageResizeConfig возвращает конфигурацию по умолчанию
func DefaultImageResizeConfig() *ImageResizeConfig {
	return &ImageResizeConfig{
	Enabled:    true,
		MaxMemoryMB: 100,
		CacheSize:  50,

		DefaultStrategy: "vision_optimized",
		AutoResize:     true,
		ModelStrategies: map[string]string{
			"glm-vision":      "vision_optimized",
			"glm-4.6v-flash":  "vision_optimized",
			"deepseek-chat":    "text_optimized",
			"deepseek-coder":   "code_optimized",
		},

		BatchSize:     10,
		ParallelResize: false,
		MaxConcurrency: 3,

		DefaultPreset: "high_quality",
		DefaultFormat:  "jpeg",

		SizeThresholds: &SizeThresholds{
			SmallImageKB:  100,   // < 100KB - не трогать
			MediumImageKB: 1024,  // 100KB-1MB - стандартный
			LargeImageKB:  5120, // 1-5MB - агрессивный
			HugeImageKB: 10240, // > 5MB - temp files
		},

		Limits: &ResizeLimits{
			MaxImagesPerContext:    50,
			MaxMemoryPerContextMB:   50,
			MaxConcurrentResizes:    5,
			ResizeTimeoutSeconds:      30,
			MaxBatchSize:              20,
		},

		SmartCrop:        false,
		PreserveMetadata: false,
		ProgressiveJPEG:   false,

		Monitoring:       true,
		LogOperations:    false,
		Profiling:        false,
	}
}

// GetStrategy получает стратегию по имени
func (irc *ImageResizeConfig) GetStrategy(name string) (*ResizeStrategy, error) {
	strategy, exists := irc.Strategies[name]
	if !exists {
		return nil, fmt.Errorf("resize strategy '%s' not found", name)
	}
	return strategy, nil
}

// GetStrategyForModel получает стратегию для модели
func (irc *ImageResizeConfig) GetStrategyForModel(modelName string) *ResizeStrategy {
	strategyName, exists := irc.ModelStrategies[modelName]
	if !exists {
		// Используем стратегию по умолчанию
		strategyName = irc.DefaultStrategy
	}

	strategy, err := irc.GetStrategy(strategyName)
	if err != nil {
		// Fallback к vision_optimized
		strategy, _ = irc.GetStrategy("vision_optimized")
	}

	return strategy
}

// GetPreset получает пресет по имени
func (irc *ImageResizeConfig) GetPreset(name string) (*ResizeStrategy, error) {
	preset, exists := irc.QualityPresets[name]
	if !exists {
		return nil, fmt.Errorf("quality preset '%s' not found", name)
	}
	return preset, nil
}

// GetFlowConfig получает конфигурацию для flow
func (irc *ImageResizeConfig) GetFlowConfig(flowName string) (*FlowResizeConfig, error) {
	config, exists := irc.Flows[flowName]
	if !exists {
		return &FlowResizeConfig{
			Enabled:      false,
			StrategyName: irc.DefaultStrategy,
			QualityPreset: irc.DefaultPreset,
			MaxImages:   irc.Limits.MaxImagesPerContext,
		}, nil
	}
	return config, nil
}

// Validate проверяет конфигурацию на валидность
func (irc *ImageResizeConfig) Validate() error {
	if irc.MaxMemoryMB <= 0 {
		return fmt.Errorf("max_memory_mb must be positive")
	}

	if irc.Limits.MaxImagesPerContext <= 0 {
		return fmt.Errorf("max_images_per_context must be positive")
	}

	if irc.SizeThresholds.SmallImageKB <= 0 {
		return fmt.Errorf("small_image_kb must be positive")
	}

	// Проверяем стратегии
	for name, strategy := range irc.Strategies {
		if err := strategy.Validate(); err != nil {
			return fmt.Errorf("strategy '%s' is invalid: %w", name, err)
		}
	}

	return nil
}

// GetEffectiveLimits вычисляет эффективные лимиты на основе конфигурации
func (irc *ImageResizeConfig) GetEffectiveLimits() *EffectiveLimits {
	return &EffectiveLimits{
		MaxImagesPerContext: irc.Limits.MaxImagesPerContext,
		MaxMemoryPerContextMB: irc.Limits.MaxMemoryPerContextMB,
		SafeImageKB:        irc.SizeThresholds.SmallImageKB,
		StandardResizeKB:   irc.SizeThresholds.MediumImageKB,
		AggressiveResizeKB: irc.SizeThresholds.LargeImageKB,
		TempFileThresholdKB: irc.SizeThresholds.HugeImageKB,
	}
}

// EffectiveLimits представляет эффективные лимиты
type EffectiveLimits struct {
	MaxImagesPerContext    int
	MaxMemoryPerContextMB   int64
	SafeImageKB            int
	StandardResizeKB       int
	AggressiveResizeKB     int
	TempFileThresholdKB     int
}

// FromConfig создает конфигурацию из interfaces.PonchoFrameworkConfig
func FromConfig(config *interfaces.PonchoFrameworkConfig) *ImageResizeConfig {
	irc := DefaultImageResizeConfig()

	// Базовые настройки
	if config.ImageOptimization != nil {
		irc.DefaultStrategy = "custom"
		irc.Strategies["custom"] = &ResizeStrategy{
			Name:        "custom",
			Description: "Custom configuration from main config",
			MaxWidth:    config.ImageOptimization.MaxWidth,
			MaxHeight:   config.ImageOptimization.MaxHeight,
			MaxFileSizeKB: int(config.ImageOptimization.MaxSizeBytes / 1024),
			Quality:     config.ImageOptimization.Quality,
			TargetFormat: config.ImageOptimization.Format,
			Enabled:     config.ImageOptimization.Enabled,
			Priority:    50,
		}
	}

	// Расширенная конфигурация (если есть)
	if config.CustomParams != nil {
		if resizeConfigData, ok := config.CustomParams["image_resize"].(map[string]interface{}); ok {
			irc.loadFromCustomParams(resizeConfigData)
		}
	}

	return irc
}

// loadFromCustomParams загружает конфигурацию из custom_params
func (irc *ImageResizeConfig) loadFromCustomParams(params map[string]interface{}) {
	// Стратегии
	if strategiesData, ok := params["strategies"].(map[string]interface{}); ok {
		for name, strategyData := range strategiesData {
			if strategyMap, ok := strategyData.(map[string]interface{}); ok {
				strategy := &ResizeStrategy{
					Name: name,
				}
				if name, ok := strategyMap["name"].(string); ok {
					strategy.Name = name
				}
				if description, ok := strategyMap["description"].(string); ok {
					strategy.Description = description
				}
				if maxWidth, ok := strategyMap["max_width"].(int); ok {
					strategy.MaxWidth = maxWidth
				}
				if maxHeight, ok := strategyMap["max_height"].(int); ok {
					strategy.MaxHeight = maxHeight
				}
				if maxFileSizeKB, ok := strategyMap["max_file_size_kb"].(int); ok {
					strategy.MaxFileSizeKB = maxFileSizeKB
				}
				if quality, ok := strategyMap["quality"].(int); ok {
					strategy.Quality = quality
				}
				if targetFormat, ok := strategyMap["target_format"].(string); ok {
					strategy.TargetFormat = targetFormat
				}
				if enabled, ok := strategyMap["enabled"].(bool); ok {
					strategy.Enabled = enabled
				}
				if priority, ok := strategyMap["priority"].(int); ok {
					strategy.Priority = priority
				}

				irc.Strategies[name] = strategy
			}
		}
	}

	// Presets
	if presetsData, ok := params["quality_presets"].(map[string]interface{}); ok {
		for name, presetData := range presetsData {
			if presetMap, ok := presetData.(map[string]interface{}); ok {
				preset := &ResizeStrategy{
					Name:        name + "_preset",
					Description: "Custom preset: " + name,
				}
				if maxWidth, ok := presetMap["max_width"].(int); ok {
					preset.MaxWidth = maxWidth
				}
				if maxHeight, ok := presetMap["max_height"].(int); ok {
					preset.MaxHeight = maxHeight
				}
				if maxFileSizeKB, ok := presetMap["max_file_size_kb"].(int); ok {
					preset.MaxFileSizeKB = maxFileSizeKB
				}
				if quality, ok := presetMap["quality"].(int); ok {
					preset.Quality = quality
				}
				if targetFormat, ok := presetMap["target_format"].(string); ok {
					preset.TargetFormat = targetFormat
				}
				if enabled, ok := presetMap["enabled"].(bool); ok {
					preset.Enabled = enabled
				}

				irc.QualityPresets[name] = preset
			}
		}
	}

	// Flow конфигурации
	if flowsData, ok := params["flows"].(map[string]interface{}); ok {
		irc.Flows = make(map[string]*FlowResizeConfig)
		for flowName, flowData := range flowsData {
			if flowMap, ok := flowData.(map[string]interface{}); ok {
				config := &FlowResizeConfig{}
				if enabled, ok := flowMap["enabled"].(bool); ok {
					config.Enabled = enabled
				}
				if strategyName, ok := flowMap["strategy_name"].(string); ok {
					config.StrategyName = strategyName
				}
				if qualityPreset, ok := flowMap["quality_preset"].(string); ok {
					config.QualityPreset = qualityPreset
				}
				if maxImages, ok := flowMap["max_images"].(int); ok {
					config.MaxImages = maxImages
				}
				if memoryLimitMB, ok := flowMap["memory_limit_mb"].(int64); ok {
					config.MemoryLimitMB = memoryLimitMB
				}

				irc.Flows[flowName] = config
			}
		}
	}
}

// Validate проверяет стратегию на валидность
func (rs *ResizeStrategy) Validate() error {
	if rs.MaxWidth <= 0 || rs.MaxWidth > 8192 {
		return fmt.Errorf("max_width must be between 1 and 8192, got %d", rs.MaxWidth)
	}
	if rs.MaxHeight <= 0 || rs.MaxHeight > 8192 {
		return fmt.Errorf("max_height must be between 1 and 8192, got %d", rs.MaxHeight)
	}
	if rs.MaxFileSizeKB <= 0 || rs.MaxFileSizeKB > 10240 {
		return fmt.Errorf("max_file_size_kb must be between 1 and 10240, got %d", rs.MaxFileSizeKB)
	}
	if rs.Quality < 1 || rs.Quality > 100 {
		return fmt.Errorf("quality must be between 1 and 100, got %d", rs.Quality)
	}
	if !isValidFormat(rs.TargetFormat) {
		return fmt.Errorf("invalid target format: %s", rs.TargetFormat)
	}
	return nil
}

func isValidFormat(format string) bool {
	validFormats := []string{"jpeg", "jpg", "png", "webp", "bmp", "tiff"}
	for _, valid := range validFormats {
		if strings.EqualFold(strings.ToLower(format), valid) {
			return true
		}
	}
	return false
}

// ParseDuration парсит duration string в time.Duration
func ParseDuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 0, nil
	}

	// Простой парсер для формата "30s", "5m", "1h"
	if strings.HasSuffix(durationStr, "s") {
		seconds, err := strconv.Atoi(strings.TrimSuffix(durationStr, "s"))
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %v", err)
		}
		return time.Duration(seconds) * time.Second, nil
	}

	if strings.HasSuffix(durationStr, "m") {
		minutes, err := strconv.Atoi(strings.TrimSuffix(durationStr, "m"))
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %v", err)
		}
		return time.Duration(minutes) * time.Minute, nil
	}

	if strings.HasSuffix(durationStr, "h") {
		hours, err := strconv.Atoi(strings.TrimSuffix(durationStr, "h"))
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %v", err)
		}
		return time.Duration(hours) * time.Hour, nil
	}

	return 0, fmt.Errorf("invalid duration format: %s", durationStr)
}

// ParseSizeMB парсит размер в MB
func ParseSizeMB(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}

	// Поддержка "100MB", "500KB", "2GB"
	if strings.HasSuffix(sizeStr, "MB") {
		mb, err := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "MB"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid MB size: %v", err)
		}
		return int64(mb), nil
	}

	if strings.HasSuffix(sizeStr, "KB") {
		kb, err := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "KB"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid KB size: %v", err)
		}
		return int64(kb / 1024), nil
	}

	if strings.HasSuffix(sizeStr, "GB") {
		gb, err := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "GB"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid GB size: %v", err)
		}
		return int64(gb * 1024), nil
	}

	return 0, fmt.Errorf("invalid size format: %s", sizeStr)
}