package articleflow

import "time"

// DefaultFlowConfig returns default configuration for the article flow
func DefaultFlowConfig() *ArticleFlowConfig {
	return &ArticleFlowConfig{
		ImageResize: &ImageResizeConfig{
			Enabled:       true,
			MaxWidth:      1024,
			MaxHeight:     1024,
			Quality:       85,
			MaxFileSizeKB: 500,
			TargetFormat:  "jpeg",
		},
		Concurrency: &ConcurrencyConfig{
			VisionAnalysisWorkers: 3,
			CreativeWorkers:       2,
		},
		Caching: &CachingConfig{
			WBParentsTTL:    24 * time.Hour,
			WBSubjectsTTL:   12 * time.Hour,
			WBCharsTTL:      6 * time.Hour,
			MaxCacheSize:    1000,
		},
		ModelParams: &ModelParamsConfig{
			VisionModel: "glm-4.6v-flash",
			TextModel:   "deepseek-chat",
			Temperature: 0.3,
			MaxTokens:   2000,
		},
	}
}