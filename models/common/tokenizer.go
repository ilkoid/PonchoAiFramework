package common

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// Tokenizer handles token counting and usage tracking for different models
type Tokenizer struct {
	logger interfaces.Logger
}

// NewTokenizer creates a new tokenizer
func NewTokenizer(logger interfaces.Logger) *Tokenizer {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}
	return &Tokenizer{
		logger: logger,
	}
}


// ModelTokenConfig represents token configuration for a specific model
type ModelTokenConfig struct {
	ModelName            string  `json:"model_name"`
	MaxTokens           int     `json:"max_tokens"`
	TokensPerCharacter  float64 `json:"tokens_per_character"`
	TokensPerWord       float64 `json:"tokens_per_word"`
	SupportsVision      bool    `json:"supports_vision"`
	VisionTokenCost     int     `json:"vision_token_cost"`
	OverheadTokens      int     `json:"overhead_tokens"`
}

// Default token configurations for different models
var defaultTokenConfigs = map[Provider]map[string]ModelTokenConfig{
	ProviderDeepSeek: {
		"deepseek-chat": {
			ModelName:           "deepseek-chat",
			MaxTokens:          4000,
			TokensPerCharacter:  0.25,
			TokensPerWord:       1.3,
			SupportsVision:      false,
			VisionTokenCost:     0,
			OverheadTokens:      10,
		},
		"deepseek-coder": {
			ModelName:           "deepseek-coder",
			MaxTokens:          8000,
			TokensPerCharacter:  0.3,
			TokensPerWord:       1.4,
			SupportsVision:      false,
			VisionTokenCost:     0,
			OverheadTokens:      15,
		},
	},
	ProviderZAI: {
		"glm-4.6v": {
			ModelName:           "glm-4.6v",
			MaxTokens:          2000,
			TokensPerCharacter:  0.35,
			TokensPerWord:       1.5,
			SupportsVision:      true,
			VisionTokenCost:     85,
			OverheadTokens:      20,
		},
		"glm-4.6v-flash": {
			ModelName:           "glm-4.6v-flash",
			MaxTokens:          4000,
			TokensPerCharacter:  0.3,
			TokensPerWord:       1.4,
			SupportsVision:      true,
			VisionTokenCost:     65,
			OverheadTokens:      15,
		},
		"glm-4.6": {
			ModelName:           "glm-4.6",
			MaxTokens:          8000,
			TokensPerCharacter:  0.25,
			TokensPerWord:       1.3,
			SupportsVision:      false,
			VisionTokenCost:     0,
			OverheadTokens:      10,
		},
	},
	ProviderOpenAI: {
		"gpt-4": {
			ModelName:           "gpt-4",
			MaxTokens:          8000,
			TokensPerCharacter:  0.25,
			TokensPerWord:       1.3,
			SupportsVision:      false,
			VisionTokenCost:     0,
			OverheadTokens:      10,
		},
		"gpt-4-vision-preview": {
			ModelName:           "gpt-4-vision-preview",
			MaxTokens:          4000,
			TokensPerCharacter:  0.3,
			TokensPerWord:       1.4,
			SupportsVision:      true,
			VisionTokenCost:     85,
			OverheadTokens:      20,
		},
	},
}

// CountTokens counts tokens in text for a specific model
func (t *Tokenizer) CountTokens(text string, provider Provider, modelName string) (int, error) {
	if text == "" {
		return 0, nil
	}

	config, err := t.getModelConfig(provider, modelName)
	if err != nil {
		return 0, fmt.Errorf("failed to get model config: %w", err)
	}

	// Use multiple estimation methods for better accuracy
	charCount := utf8.RuneCountInString(text)
	wordCount := t.countWords(text)

	// Weighted average of character and word-based estimates
	charTokens := int(float64(charCount) * config.TokensPerCharacter)
	wordTokens := int(float64(wordCount) * config.TokensPerWord)

	// Use the higher estimate to be safe
	tokens := int(math.Max(float64(charTokens), float64(wordTokens)))

	// Add overhead tokens
	tokens += config.OverheadTokens

	t.logger.Debug("Token count estimated",
		"model", modelName,
		"provider", provider,
		"char_count", charCount,
		"word_count", wordCount,
		"char_tokens", charTokens,
		"word_tokens", wordTokens,
		"final_tokens", tokens)

	return tokens, nil
}

// CountMessageTokens counts tokens in a message for a specific model
func (t *Tokenizer) CountMessageTokens(message *interfaces.PonchoMessage, provider Provider, modelName string) (int, error) {
	config, err := t.getModelConfig(provider, modelName)
	if err != nil {
		return 0, fmt.Errorf("failed to get model config: %w", err)
	}

	totalTokens := config.OverheadTokens

	// Count tokens for each content part
	for _, part := range message.Content {
		partTokens, err := t.countContentPartTokens(part, provider, modelName)
		if err != nil {
			return 0, fmt.Errorf("failed to count content part tokens: %w", err)
		}
		totalTokens += partTokens
	}

	// Add tokens for role and name
	if message.Role != "" {
		totalTokens += 2 // Role tokens
	}
	if message.Name != nil && *message.Name != "" {
		nameTokens, err := t.CountTokens(*message.Name, provider, modelName)
		if err != nil {
			return 0, fmt.Errorf("failed to count name tokens: %w", err)
		}
		totalTokens += nameTokens + 1 // Name tokens + separator
	}

	return totalTokens, nil
}

// CountRequestTokens counts tokens in a request for a specific model
func (t *Tokenizer) CountRequestTokens(request *interfaces.PonchoModelRequest, provider Provider, modelName string) (*interfaces.PonchoUsage, error) {
	config, err := t.getModelConfig(provider, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}

	promptTokens := config.OverheadTokens

	// Count tokens for all messages
	for _, message := range request.Messages {
		messageTokens, err := t.CountMessageTokens(message, provider, modelName)
		if err != nil {
			return nil, fmt.Errorf("failed to count message tokens: %w", err)
		}
		promptTokens += messageTokens
	}

	// System messages should be included in the Messages array as PonchoRoleSystem
	// No separate System field to handle

	// Add tokens for tools if present
	if len(request.Tools) > 0 {
		toolTokens, err := t.countToolTokens(request.Tools, provider, modelName)
		if err != nil {
			return nil, fmt.Errorf("failed to count tool tokens: %w", err)
		}
		promptTokens += toolTokens
	}

	// Estimate completion tokens (rough estimate)
	completionTokens := t.estimateCompletionTokens(request, config)

	usage := &interfaces.PonchoUsage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}

	t.logger.Debug("Request token usage calculated",
		"model", modelName,
		"provider", provider,
		"prompt_tokens", usage.PromptTokens,
		"completion_tokens", usage.CompletionTokens,
		"total_tokens", usage.TotalTokens)

	return usage, nil
}

// ValidateTokenLimits validates if request is within token limits
func (t *Tokenizer) ValidateTokenLimits(request *interfaces.PonchoModelRequest, provider Provider, modelName string) error {
	config, err := t.getModelConfig(provider, modelName)
	if err != nil {
		return fmt.Errorf("failed to get model config: %w", err)
	}

	usage, err := t.CountRequestTokens(request, provider, modelName)
	if err != nil {
		return fmt.Errorf("failed to count request tokens: %w", err)
	}

	if usage.PromptTokens >= config.MaxTokens {
		return fmt.Errorf("request exceeds model token limit: %d >= %d", 
			usage.PromptTokens, config.MaxTokens)
	}

	// Check if we have enough tokens for completion
	if usage.TotalTokens > config.MaxTokens {
		return fmt.Errorf("total tokens (prompt + completion) exceeds model limit: %d > %d", 
			usage.TotalTokens, config.MaxTokens)
	}

	return nil
}

// EstimateCost estimates the cost of a request
func (t *Tokenizer) EstimateCost(usage *interfaces.PonchoUsage, provider Provider, modelName string) (float64, error) {
	// Cost per 1000 tokens for different models (example rates)
	costPer1K := map[Provider]map[string]float64{
		ProviderDeepSeek: {
			"deepseek-chat":   0.0014, // $0.0014 per 1K tokens
			"deepseek-coder":  0.0020, // $0.0020 per 1K tokens
		},
		ProviderZAI: {
			"glm-4.6v":         0.0050, // $0.0050 per 1K tokens
			"glm-4.6v-flash":   0.0030, // $0.0030 per 1K tokens
			"glm-4.6":          0.0025, // $0.0025 per 1K tokens
		},
		ProviderOpenAI: {
			"gpt-4":               0.0300, // $0.0300 per 1K tokens
			"gpt-4-vision-preview": 0.0500, // $0.0500 per 1K tokens
		},
	}

	providerCosts, exists := costPer1K[provider]
	if !exists {
		return 0, fmt.Errorf("no cost information available for provider: %s", provider)
	}

	cost, exists := providerCosts[modelName]
	if !exists {
		return 0, fmt.Errorf("no cost information available for model: %s", modelName)
	}

	totalCost := (float64(usage.TotalTokens) / 1000.0) * cost

	t.logger.Debug("Cost estimated",
		"model", modelName,
		"provider", provider,
		"total_tokens", usage.TotalTokens,
		"cost_per_1k", cost,
		"total_cost", totalCost)

	return totalCost, nil
}

// getModelConfig gets the token configuration for a specific model
func (t *Tokenizer) getModelConfig(provider Provider, modelName string) (*ModelTokenConfig, error) {
	providerConfigs, exists := defaultTokenConfigs[provider]
	if !exists {
		return nil, fmt.Errorf("no token configuration available for provider: %s", provider)
	}

	config, exists := providerConfigs[modelName]
	if !exists {
		return nil, fmt.Errorf("no token configuration available for model: %s", modelName)
	}

	return &config, nil
}

// countWords counts words in text (simple implementation)
func (t *Tokenizer) countWords(text string) int {
	// Remove punctuation and split on whitespace
	re := regexp.MustCompile(`\s+`)
	words := re.Split(strings.TrimSpace(text), -1)
	
	count := 0
	for _, word := range words {
		if word != "" {
			count++
		}
	}
	
	return count
}

// countContentPartTokens counts tokens in a content part
func (t *Tokenizer) countContentPartTokens(part *interfaces.PonchoContentPart, provider Provider, modelName string) (int, error) {
	switch part.Type {
	case interfaces.PonchoContentTypeText:
		return t.CountTokens(part.Text, provider, modelName)
	case interfaces.PonchoContentTypeMedia:
		config, err := t.getModelConfig(provider, modelName)
		if err != nil {
			return 0, err
		}
		if config.SupportsVision {
			return config.VisionTokenCost, nil
		}
		return 0, fmt.Errorf("model %s does not support vision content", modelName)
	case interfaces.PonchoContentTypeTool:
		if part.Tool == nil {
			return 0, nil
		}
		// Count tokens for tool call
		tokens := 10 // Base tool call overhead
		if part.Tool.ID != "" {
			idTokens, err := t.CountTokens(part.Tool.ID, provider, modelName)
			if err != nil {
				return 0, err
			}
			tokens += idTokens
		}
		if part.Tool.Name != "" {
			nameTokens, err := t.CountTokens(part.Tool.Name, provider, modelName)
			if err != nil {
				return 0, err
			}
			tokens += nameTokens
		}
		if len(part.Tool.Args) > 0 {
			argsStr := t.formatArgs(part.Tool.Args)
			argsTokens, err := t.CountTokens(argsStr, provider, modelName)
			if err != nil {
				return 0, err
			}
			tokens += argsTokens
		}
		return tokens, nil
	default:
		return 0, fmt.Errorf("unsupported content type: %s", part.Type)
	}
}

// countToolTokens counts tokens for tool definitions
func (t *Tokenizer) countToolTokens(tools []*interfaces.PonchoToolDef, provider Provider, modelName string) (int, error) {
	totalTokens := 0

	for _, tool := range tools {
		toolTokens := 20 // Base tool definition overhead
		
		// Add tokens for tool name and description
		if tool.Name != "" {
			nameTokens, err := t.CountTokens(tool.Name, provider, modelName)
			if err != nil {
				return 0, err
			}
			toolTokens += nameTokens
		}
		
		if tool.Description != "" {
			descTokens, err := t.CountTokens(tool.Description, provider, modelName)
			if err != nil {
				return 0, err
			}
			toolTokens += descTokens
		}
		
		// Add tokens for parameters (simplified)
		if tool.Parameters != nil {
			paramsStr := t.formatArgs(tool.Parameters)
			paramsTokens, err := t.CountTokens(paramsStr, provider, modelName)
			if err != nil {
				return 0, err
			}
			toolTokens += paramsTokens
		}
		
		totalTokens += toolTokens
	}

	return totalTokens, nil
}

// estimateCompletionTokens estimates completion tokens based on request
func (t *Tokenizer) estimateCompletionTokens(request *interfaces.PonchoModelRequest, config *ModelTokenConfig) int {
	// Simple heuristic: estimate based on request complexity
	baseEstimate := 150
	
	// Adjust based on request characteristics
	if len(request.Messages) > 5 {
		baseEstimate += 50
	}
	
	if request.MaxTokens != nil && *request.MaxTokens > 0 && *request.MaxTokens < baseEstimate {
		return *request.MaxTokens
	}
	
	return baseEstimate
}

// formatArgs formats arguments for token counting
func (t *Tokenizer) formatArgs(args interface{}) string {
	if args == nil {
		return ""
	}
	
	// Simple string representation for token counting
	return fmt.Sprintf("%v", args)
}

// GetModelConfig returns the token configuration for a model
func (t *Tokenizer) GetModelConfig(provider Provider, modelName string) (*ModelTokenConfig, error) {
	return t.getModelConfig(provider, modelName)
}

// SetModelConfig sets a custom token configuration for a model
func (t *Tokenizer) SetModelConfig(provider Provider, config ModelTokenConfig) {
	if defaultTokenConfigs[provider] == nil {
		defaultTokenConfigs[provider] = make(map[string]ModelTokenConfig)
	}
	defaultTokenConfigs[provider][config.ModelName] = config
	
	t.logger.Info("Custom model token configuration set",
		"provider", provider,
		"model", config.ModelName,
		"max_tokens", config.MaxTokens)
}

// ListSupportedModels returns all supported models with their configurations
func (t *Tokenizer) ListSupportedModels() map[Provider]map[string]ModelTokenConfig {
	// Return a copy to prevent modification
	result := make(map[Provider]map[string]ModelTokenConfig)
	
	for provider, models := range defaultTokenConfigs {
		result[provider] = make(map[string]ModelTokenConfig)
		for model, config := range models {
			result[provider][model] = config
		}
	}
	
	return result
}