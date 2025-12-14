// Package prompts provides comprehensive prompt management system for PonchoFramework
//
// Key functionality:
// • Central orchestration of prompt template lifecycle (load, validate, execute, cache)
// • Integration with framework's model registry for AI model execution
// • Support for both batch and streaming prompt execution
// • Template validation with configurable strictness levels
// • Metrics collection for performance monitoring and debugging
// • Thread-safe operations with concurrent access protection
//
// Key relationships:
// • Depends on interfaces package for core type definitions and framework integration
// • Uses parser package for template loading and V1 format support
// • Uses validator package for template and variable validation
// • Uses executor package for actual prompt execution against AI models
// • Uses cache package for template caching and performance optimization
// • Integrates with core framework through PonchoFramework interface
//
// Design patterns:
// • Registry pattern for template storage and management
// • Strategy pattern for different template formats (V1, YAML, JSON)
// • Observer pattern for metrics collection and monitoring
// • Builder pattern for request construction and execution

package prompts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PromptManagerImpl implements the PromptManager interface
type PromptManagerImpl struct {
	// Core components
	loader     interfaces.PromptTemplateLoader
	validator  interfaces.PromptValidator
	executor   interfaces.PromptExecutor
	cache      interfaces.PromptCache
	logger     interfaces.Logger
	
	// Configuration
	config     *PromptConfig
	
	// Template storage
	templates  map[string]*interfaces.PromptTemplate
	mutex      sync.RWMutex
	
	// Metrics
	metrics    *SystemMetrics
	
	// Framework reference
	framework  interfaces.PonchoFramework
}

// NewPromptManager creates a new PromptManager instance
func NewPromptManager(
	config *PromptConfig,
	framework interfaces.PonchoFramework,
	logger interfaces.Logger,
) *PromptManagerImpl {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	pm := &PromptManagerImpl{
		config:    config,
		templates: make(map[string]*interfaces.PromptTemplate),
		logger:    logger,
		framework: framework,
		metrics:   &SystemMetrics{
			StartTime: time.Now(),
		},
	}

	// Initialize components
	pm.loader = NewPromptTemplateLoader(config, logger)
	pm.validator = NewPromptValidator(config, logger)
	pm.executor = NewPromptExecutor(framework, config, logger)
	
	if config.Cache.Enabled {
		pm.cache = NewPromptCache(config.Cache.Size, logger)
	}

	return pm
}

// LoadTemplate loads a prompt template from file or cache
func (pm *PromptManagerImpl) LoadTemplate(name string) (*interfaces.PromptTemplate, error) {
	pm.mutex.RLock()
	
	// Check cache first
	if pm.cache != nil {
		if template, found := pm.cache.GetTemplate(name); found {
			pm.mutex.RUnlock()
			pm.metrics.CacheHits++
			pm.logger.Debug("Template loaded from cache", "name", name)
			return template, nil
		}
		pm.metrics.CacheMisses++
	}
	
	// Check in-memory storage
	if template, exists := pm.templates[name]; exists {
		pm.mutex.RUnlock()
		pm.logger.Debug("Template loaded from memory", "name", name)
		return template, nil
	}
	pm.mutex.RUnlock()

	// Load from file system
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Double-check after acquiring write lock
	if template, exists := pm.templates[name]; exists {
		return template, nil
	}

	// Try to find the template file
	templatePath, err := pm.findTemplateFile(name)
	if err != nil {
		pm.logger.Error("Failed to find template file", "name", name, "error", err)
		pm.metrics.TotalErrors++
		if pm.metrics.ErrorsByType == nil {
			pm.metrics.ErrorsByType = make(map[string]int64)
		}
		pm.metrics.ErrorsByType["template_not_found"]++
		return nil, fmt.Errorf("failed to find template '%s': %w", name, err)
	}

	template, err := pm.loader.LoadFromFile(templatePath)
	if err != nil {
		pm.logger.Error("Failed to load template", "name", name, "path", templatePath, "error", err)
		pm.metrics.TotalErrors++
		if pm.metrics.ErrorsByType == nil {
			pm.metrics.ErrorsByType = make(map[string]int64)
		}
		pm.metrics.ErrorsByType["template_load_failed"]++
		return nil, fmt.Errorf("failed to load template '%s': %w", name, err)
	}

	// Validate template if enabled
	if pm.config.Validation.ValidateOnLoad {
		if validationResult, err := pm.validator.ValidateTemplate(template); err != nil {
			pm.logger.Error("Template validation failed", "name", name, "error", err)
			return nil, fmt.Errorf("template validation failed: %w", err)
		} else if !validationResult.Valid {
			pm.logger.Warn("Template validation warnings", "name", name, "warnings", len(validationResult.Warnings))
			if pm.config.Validation.Strict {
				return nil, fmt.Errorf("template validation failed with %d errors", len(validationResult.Errors))
			}
		}
	}

	// Store template
	pm.templates[name] = template
	pm.metrics.LoadedTemplates++

	// Cache template
	if pm.cache != nil {
		pm.cache.SetTemplate(name, template)
		pm.metrics.CachedTemplates++
	}

	pm.logger.Info("Template loaded successfully", "name", name, "path", templatePath)
	return template, nil
}

// ExecutePrompt executes a prompt template with variables
func (pm *PromptManagerImpl) ExecutePrompt(
	ctx context.Context,
	name string,
	variables map[string]interface{},
	model string,
) (*interfaces.PonchoModelResponse, error) {
	startTime := time.Now()
	defer func() {
		pm.metrics.TotalExecutions++
		executionTime := time.Since(startTime).Milliseconds()
		if pm.metrics.TotalExecutions == 1 {
			pm.metrics.AvgExecutionTime = float64(executionTime)
		} else {
			pm.metrics.AvgExecutionTime = (pm.metrics.AvgExecutionTime + float64(executionTime)) / 2.0
		}
	}()

	pm.logger.Debug("Executing prompt", "name", name, "model", model)

	// Load template
	template, err := pm.LoadTemplate(name)
	if err != nil {
		pm.metrics.FailedExecutions++
		return nil, err
	}

	// Set default model if not specified
	if model == "" {
		model = pm.config.Execution.DefaultModel
		if model == "" {
			model = template.Model
		}
	}

	// Validate variables if enabled
	if pm.config.Validation.ValidateOnExec {
		if validationResult, err := pm.validator.ValidateVariables(template, variables); err != nil {
			pm.metrics.FailedExecutions++
			pm.logger.Error("Variable validation failed", "name", name, "error", err)
			return nil, fmt.Errorf("variable validation failed: %w", err)
		} else if !validationResult.Valid && pm.config.Validation.Strict {
			pm.metrics.FailedExecutions++
			return nil, fmt.Errorf("variable validation failed with %d errors", len(validationResult.Errors))
		}
	}

	// Execute template
	response, err := pm.executor.ExecuteTemplate(ctx, template, variables, model)
	if err != nil {
		pm.metrics.FailedExecutions++
		pm.logger.Error("Template execution failed", "name", name, "model", model, "error", err)
		
		// Update error metrics
		if pm.metrics.ErrorsByType == nil {
			pm.metrics.ErrorsByType = make(map[string]int64)
		}
		if pm.metrics.ErrorsByComponent == nil {
			pm.metrics.ErrorsByComponent = make(map[string]int64)
		}
		pm.metrics.ErrorsByType["execution_failed"]++
		pm.metrics.ErrorsByComponent["executor"]++
		
		return nil, fmt.Errorf("template execution failed: %w", err)
	}

	pm.metrics.SuccessfulExecutions++
	pm.logger.Debug("Prompt executed successfully", 
		"name", name, 
		"model", model, 
		"tokens", response.Usage.TotalTokens,
		"duration_ms", time.Since(startTime).Milliseconds())

	return response, nil
}

// ExecutePromptStreaming executes a prompt template with streaming
func (pm *PromptManagerImpl) ExecutePromptStreaming(
	ctx context.Context,
	name string,
	variables map[string]interface{},
	model string,
	callback interfaces.PonchoStreamCallback,
) error {
	startTime := time.Now()
	defer func() {
		pm.metrics.TotalExecutions++
		executionTime := time.Since(startTime).Milliseconds()
		if pm.metrics.TotalExecutions == 1 {
			pm.metrics.AvgExecutionTime = float64(executionTime)
		} else {
			pm.metrics.AvgExecutionTime = (pm.metrics.AvgExecutionTime + float64(executionTime)) / 2.0
		}
	}()

	pm.logger.Debug("Executing streaming prompt", "name", name, "model", model)

	// Load template
	template, err := pm.LoadTemplate(name)
	if err != nil {
		pm.metrics.FailedExecutions++
		return err
	}

	// Set default model if not specified
	if model == "" {
		model = pm.config.Execution.DefaultModel
		if model == "" {
			model = template.Model
		}
	}

	// Validate variables if enabled
	if pm.config.Validation.ValidateOnExec {
		if validationResult, err := pm.validator.ValidateVariables(template, variables); err != nil {
			pm.metrics.FailedExecutions++
			pm.logger.Error("Variable validation failed", "name", name, "error", err)
			return fmt.Errorf("variable validation failed: %w", err)
		} else if !validationResult.Valid && pm.config.Validation.Strict {
			pm.metrics.FailedExecutions++
			return fmt.Errorf("variable validation failed with %d errors", len(validationResult.Errors))
		}
	}

	// Execute template with streaming
	err = pm.executor.ExecuteTemplateStreaming(ctx, template, variables, model, callback)
	if err != nil {
		pm.metrics.FailedExecutions++
		pm.logger.Error("Streaming template execution failed", "name", name, "model", model, "error", err)
		
		// Update error metrics
		if pm.metrics.ErrorsByType == nil {
			pm.metrics.ErrorsByType = make(map[string]int64)
		}
		if pm.metrics.ErrorsByComponent == nil {
			pm.metrics.ErrorsByComponent = make(map[string]int64)
		}
		pm.metrics.ErrorsByType["streaming_execution_failed"]++
		pm.metrics.ErrorsByComponent["executor"]++
		
		return fmt.Errorf("streaming template execution failed: %w", err)
	}

	pm.metrics.SuccessfulExecutions++
	pm.logger.Debug("Streaming prompt executed successfully", 
		"name", name, 
		"model", model,
		"duration_ms", time.Since(startTime).Milliseconds())

	return nil
}

// ValidatePrompt validates a prompt template
func (pm *PromptManagerImpl) ValidatePrompt(template *interfaces.PromptTemplate) (*interfaces.ValidationResult, error) {
	pm.logger.Debug("Validating prompt template", "name", template.Name)
	
	result, err := pm.validator.ValidateTemplate(template)
	if err != nil {
		pm.logger.Error("Template validation error", "name", template.Name, "error", err)
		return nil, fmt.Errorf("template validation error: %w", err)
	}

	if result.Valid {
		pm.logger.Debug("Template validation passed", "name", template.Name)
	} else {
		pm.logger.Warn("Template validation failed", 
			"name", template.Name, 
			"errors", len(result.Errors),
			"warnings", len(result.Warnings))
	}

	return result, nil
}

// ListTemplates returns list of available prompt templates
func (pm *PromptManagerImpl) ListTemplates() ([]string, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// If templates are not loaded, load them from directory
	if len(pm.templates) == 0 {
		pm.mutex.RUnlock()
		if err := pm.loadAllTemplates(); err != nil {
			return nil, fmt.Errorf("failed to load templates: %w", err)
		}
		pm.mutex.RLock()
	}

	// Get template names
	names := make([]string, 0, len(pm.templates))
	for name := range pm.templates {
		names = append(names, name)
	}

	pm.logger.Debug("Listed templates", "count", len(names))
	return names, nil
}

// ReloadTemplates reloads all prompt templates
func (pm *PromptManagerImpl) ReloadTemplates() error {
	pm.logger.Info("Reloading all prompt templates")

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Clear existing templates
	pm.templates = make(map[string]*interfaces.PromptTemplate)
	
	// Clear cache
	if pm.cache != nil {
		pm.cache.Clear()
	}

	// Reload all templates
	if err := pm.loadAllTemplates(); err != nil {
		pm.logger.Error("Failed to reload templates", "error", err)
		return fmt.Errorf("failed to reload templates: %w", err)
	}

	pm.logger.Info("Templates reloaded successfully")
	return nil
}

// GetMetrics returns system metrics
func (pm *PromptManagerImpl) GetMetrics() *SystemMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// Update cache hit rate
	if pm.cache != nil {
		cacheStats := pm.cache.Stats()
		pm.metrics.CacheHits = cacheStats.Hits
		pm.metrics.CacheMisses = cacheStats.Misses
		if cacheStats.Hits+cacheStats.Misses > 0 {
			pm.metrics.CacheHitRate = float64(cacheStats.Hits) / float64(cacheStats.Hits+cacheStats.Misses)
		}
	}

	// Update timestamps
	pm.metrics.LastUpdated = time.Now()
	pm.metrics.TotalTemplates = int64(len(pm.templates))
	pm.metrics.LoadedTemplates = int64(len(pm.templates))

	return pm.metrics
}

// GetTemplate returns a specific template without loading from disk
func (pm *PromptManagerImpl) GetTemplate(name string) (*interfaces.PromptTemplate, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	template, exists := pm.templates[name]
	return template, exists
}

// loadAllTemplates loads all templates from the configured directory
func (pm *PromptManagerImpl) loadAllTemplates() error {
	templates, err := pm.loader.LoadFromDirectory(pm.config.Templates.Directory)
	if err != nil {
		return fmt.Errorf("failed to load templates from directory: %w", err)
	}

	// Validate and store templates
	for name, template := range templates {
		// Validate template if enabled
		if pm.config.Validation.ValidateOnLoad {
			if validationResult, err := pm.validator.ValidateTemplate(template); err != nil {
				pm.logger.Error("Template validation failed during bulk load", 
					"name", name, "error", err)
				if pm.config.Validation.Strict {
					return fmt.Errorf("template '%s' validation failed: %w", name, err)
				}
				continue
			} else if !validationResult.Valid {
				pm.logger.Warn("Template validation warnings during bulk load", 
					"name", name, "warnings", len(validationResult.Warnings))
				if pm.config.Validation.Strict {
					return fmt.Errorf("template '%s' validation failed with %d errors", 
						name, len(validationResult.Errors))
				}
			}
		}

		// Store template
		pm.templates[name] = template
		
		// Cache template
		if pm.cache != nil {
			pm.cache.SetTemplate(name, template)
		}
	}

	pm.logger.Info("Templates loaded from directory", 
		"directory", pm.config.Templates.Directory,
		"count", len(templates))

	return nil
}

// validateTemplateName validates template name format
func (pm *PromptManagerImpl) validateTemplateName(name string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("template name contains invalid character: %s", char)
		}
	}

	// Check extension
	validExtensions := pm.config.Templates.Extensions
	if len(validExtensions) == 0 {
		validExtensions = []string{".prompt", ".yaml", ".yml", ".json"}
	}

	hasValidExtension := false
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(name), strings.ToLower(ext)) {
			hasValidExtension = true
			break
		}
	}

	if !hasValidExtension {
		return fmt.Errorf("template name must have one of these extensions: %v", validExtensions)
	}

	return nil
}

// findTemplateFile finds the template file by trying different extensions
func (pm *PromptManagerImpl) findTemplateFile(name string) (string, error) {
	// If name already has a valid extension, try it directly
	validExtensions := pm.config.Templates.Extensions
	if len(validExtensions) == 0 {
		validExtensions = []string{".prompt", ".yaml", ".yml", ".json"}
	}

	// Check if name already has a valid extension
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(name), strings.ToLower(ext)) {
			templatePath := filepath.Join(pm.config.Templates.Directory, name)
			if _, err := os.Stat(templatePath); err == nil {
				return templatePath, nil
			}
		}
	}

	// Try each valid extension
	for _, ext := range validExtensions {
		templatePath := filepath.Join(pm.config.Templates.Directory, name+ext)
		if _, err := os.Stat(templatePath); err == nil {
			return templatePath, nil
		}
	}

	return "", fmt.Errorf("template file not found for '%s' (tried extensions: %v)", name, validExtensions)
}

// updateTemplateMetrics updates metrics for a specific template
func (pm *PromptManagerImpl) updateTemplateMetrics(name string, success bool, executionTime time.Duration, tokensUsed int) {
	// This would be implemented to track per-template metrics
	// For now, we just log the information
	pm.logger.Debug("Template execution metrics", 
		"name", name,
		"success", success,
		"duration_ms", executionTime.Milliseconds(),
		"tokens_used", tokensUsed)
}

// handleError handles errors and updates metrics
func (pm *PromptManagerImpl) handleError(component, errorType string, err error) {
	pm.metrics.TotalErrors++
	
	if pm.metrics.ErrorsByType == nil {
		pm.metrics.ErrorsByType = make(map[string]int64)
	}
	if pm.metrics.ErrorsByComponent == nil {
		pm.metrics.ErrorsByComponent = make(map[string]int64)
	}
	
	pm.metrics.ErrorsByType[errorType]++
	pm.metrics.ErrorsByComponent[component]++
	
	pm.logger.Error("Prompt manager error", 
		"component", component,
		"type", errorType,
		"error", err)
}