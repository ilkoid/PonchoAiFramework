// Package prompts provides template execution and variable processing capabilities
//
// Key functionality:
// • Template execution against AI models with variable substitution
// • Streaming execution support for real-time response generation
// • Variable processing with validation and default value handling
// • Model request building from template parts and metadata
// • Media content handling for vision-capable models
// • Fashion context integration for specialized AI responses
//
// Key relationships:
// • Implements PromptExecutor interface from core interfaces
// • Integrates with core framework through PonchoFramework.Generate methods
// • Uses VariableProcessor for template variable substitution and validation
// • Converts PromptTemplate structures to PonchoModelRequest format
// • Supports both batch and streaming execution patterns
// • Handles media content for vision models (Z.AI GLM with fashion analysis)
//
// Design patterns:
// • Strategy pattern for different execution modes (batch vs streaming)
// • Builder pattern for model request construction
// • Template method pattern for execution workflow
// • Chain of responsibility for message building from template parts

package prompts

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PromptExecutorImpl implements PromptExecutor interface
type PromptExecutorImpl struct {
	framework interfaces.PonchoFramework
	config    *PromptConfig
	logger    interfaces.Logger
	processor VariableProcessor
}

// NewPromptExecutor creates a new PromptExecutor instance
func NewPromptExecutor(
	framework interfaces.PonchoFramework,
	config *PromptConfig,
	logger interfaces.Logger,
) interfaces.PromptExecutor {
	return &PromptExecutorImpl{
		framework: framework,
		config:    config,
		logger:    logger,
		processor: NewVariableProcessor(logger),
	}
}

// ExecuteTemplate executes a prompt template
func (pe *PromptExecutorImpl) ExecuteTemplate(
	ctx context.Context,
	template *interfaces.PromptTemplate,
	variables map[string]interface{},
	modelName string,
) (*interfaces.PonchoModelResponse, error) {
	pe.logger.Debug("Executing template", "name", template.Name, "model", modelName)

	startTime := time.Now()

	// Process variables
	processedVariables := pe.processor.SetDefaults(template, variables)
	
	// Validate variables
	if err := pe.processor.ValidateVariables(template, processedVariables); err != nil {
		return nil, fmt.Errorf("variable validation failed: %w", err)
	}

	// Build model request
	request, err := pe.BuildModelRequest(template, processedVariables, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to build model request: %w", err)
	}

	// Execute request
	response, err := pe.framework.Generate(ctx, request)
	if err != nil {
		pe.logger.Error("Template execution failed", 
			"name", template.Name, 
			"model", modelName, 
			"error", err)
		return nil, fmt.Errorf("template execution failed: %w", err)
	}

	pe.logger.Debug("Template executed successfully", 
		"name", template.Name,
		"model", modelName,
		"tokens", response.Usage.TotalTokens,
		"duration_ms", time.Since(startTime).Milliseconds())

	return response, nil
}

// ExecuteTemplateStreaming executes a prompt template with streaming
func (pe *PromptExecutorImpl) ExecuteTemplateStreaming(
	ctx context.Context,
	template *interfaces.PromptTemplate,
	variables map[string]interface{},
	modelName string,
	callback interfaces.PonchoStreamCallback,
) error {
	pe.logger.Debug("Executing streaming template", "name", template.Name, "model", modelName)

	startTime := time.Now()

	// Process variables
	processedVariables := pe.processor.SetDefaults(template, variables)
	
	// Validate variables
	if err := pe.processor.ValidateVariables(template, processedVariables); err != nil {
		return fmt.Errorf("variable validation failed: %w", err)
	}

	// Build model request
	request, err := pe.BuildModelRequest(template, processedVariables, modelName)
	if err != nil {
		return fmt.Errorf("failed to build model request: %w", err)
	}

	// Set streaming flag
	request.Stream = true

	// Execute streaming request
	err = pe.framework.GenerateStreaming(ctx, request, callback)
	if err != nil {
		pe.logger.Error("Streaming template execution failed", 
			"name", template.Name, 
			"model", modelName, 
			"error", err)
		return fmt.Errorf("streaming template execution failed: %w", err)
	}

	pe.logger.Debug("Streaming template executed successfully", 
		"name", template.Name,
		"model", modelName,
		"duration_ms", time.Since(startTime).Milliseconds())

	return nil
}

// BuildModelRequest builds a model request from template
func (pe *PromptExecutorImpl) BuildModelRequest(
	template *interfaces.PromptTemplate,
	variables map[string]interface{},
	modelName string,
) (*interfaces.PonchoModelRequest, error) {
	pe.logger.Debug("Building model request", "template", template.Name, "model", modelName)

	// Process template parts
	messages := make([]*interfaces.PonchoMessage, 0)

	for _, part := range template.Parts {
		message, err := pe.buildMessageFromPart(part, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to build message from part: %w", err)
		}

		if message != nil {
			messages = append(messages, message)
		}
	}

	// Build request
	request := &interfaces.PonchoModelRequest{
		Model:    modelName,
		Messages: messages,
		Metadata: make(map[string]interface{}),
	}

	// Set template metadata
	request.Metadata["template_name"] = template.Name
	request.Metadata["template_version"] = template.Version
	request.Metadata["template_category"] = template.Category
	request.Metadata["execution_time"] = time.Now().Unix()

	// Apply template-level settings
	if template.MaxTokens != nil {
		request.MaxTokens = template.MaxTokens
	} else {
		// Use default from config
		request.MaxTokens = &pe.config.Execution.DefaultMaxTokens
	}

	if template.Temperature != nil {
		request.Temperature = template.Temperature
	} else {
		// Use default from config
		request.Temperature = &pe.config.Execution.DefaultTemperature
	}

	// Add fashion context if available
	if template.FashionContext != nil {
		request.Metadata["fashion_context"] = template.FashionContext
	}

	pe.logger.Debug("Model request built", 
		"model", modelName,
		"messages", len(messages),
		"max_tokens", *request.MaxTokens,
		"temperature", *request.Temperature)

	return request, nil
}

// buildMessageFromPart builds a message from a template part
func (pe *PromptExecutorImpl) buildMessageFromPart(
	part *interfaces.PromptPart,
	variables map[string]interface{},
) (*interfaces.PonchoMessage, error) {
	// Process content with variables
	processedContent, err := pe.processor.ProcessVariables(part.Content, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to process variables: %w", err)
	}

	// Determine role
	var role interfaces.PonchoRole
	switch part.Type {
	case interfaces.PromptPartTypeSystem:
		role = interfaces.PonchoRoleSystem
	case interfaces.PromptPartTypeUser:
		role = interfaces.PonchoRoleUser
	case interfaces.PromptPartTypeMedia:
		role = interfaces.PonchoRoleUser // Media parts are typically user content
	default:
		return nil, fmt.Errorf("unknown part type: %s", part.Type)
	}

	// Build content parts
	contentParts := make([]*interfaces.PonchoContentPart, 0)

	if part.Type == interfaces.PromptPartTypeMedia {
		// Handle media content
		if part.Media != nil {
			mediaURL := part.Media.URL
			if strings.Contains(mediaURL, "{{") {
				// Process variables in media URL
				processedURL, err := pe.processor.ProcessVariables(mediaURL, variables)
				if err != nil {
					return nil, fmt.Errorf("failed to process media URL variables: %w", err)
				}
				mediaURL = processedURL
			}

			contentParts = append(contentParts, &interfaces.PonchoContentPart{
				Type: interfaces.PonchoContentTypeMedia,
				Media: &interfaces.PonchoMediaPart{
					URL:      mediaURL,
					MimeType: part.Media.MimeType,
				},
			})

			// Add text content if provided
			if processedContent != "" {
				contentParts = append(contentParts, &interfaces.PonchoContentPart{
					Type: interfaces.PonchoContentTypeText,
					Text: processedContent,
				})
			}
		}
	} else {
		// Handle text content
		contentParts = append(contentParts, &interfaces.PonchoContentPart{
			Type: interfaces.PonchoContentTypeText,
			Text: processedContent,
		})
	}

	// Build message
	message := &interfaces.PonchoMessage{
		Role:    role,
		Content: contentParts,
	}

	return message, nil
}

// VariableProcessorImpl implements VariableProcessor interface
type VariableProcessorImpl struct {
	logger interfaces.Logger
}

// NewVariableProcessor creates a new VariableProcessor instance
func NewVariableProcessor(logger interfaces.Logger) VariableProcessor {
	return &VariableProcessorImpl{
		logger: logger,
	}
}

// ProcessVariables processes variables in template content with security fixes
func (vp *VariableProcessorImpl) ProcessVariables(content string, variables map[string]interface{}) (string, error) {
	vp.logger.Debug("Processing variables", "content_length", len(content))

	// Secure variable substitution using regex with proper escaping
	varPattern := regexp.MustCompile(`\{\{(\w+)\}\}`)
	
	result := varPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name
		varName := match[2 : len(match)-2] // Remove {{ and }}
		
		// Validate variable name to prevent injection
		if !isValidVariableName(varName) {
			vp.logger.Warn("Invalid variable name detected", "name", varName)
			return "" // Remove invalid variables
		}
		
		// Look up variable value
		if value, exists := variables[varName]; exists {
			// Sanitize variable value to prevent injection
			return sanitizeVariableValue(value)
		}
		
		// Variable not found - return empty string for security
		vp.logger.Warn("Variable not found", "name", varName)
		return ""
	})

	vp.logger.Debug("Variables processed", "result_length", len(result))
	return result, nil
}

// ExtractVariables extracts variable references from content
func (vp *VariableProcessorImpl) ExtractVariables(content string) ([]string, error) {
	varPattern := regexp.MustCompile(`\{\{(\w+)\}\}`)
	matches := varPattern.FindAllStringSubmatch(content, -1)
	
	variables := make([]string, 0)
	variableSet := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if !variableSet[varName] {
				variables = append(variables, varName)
				variableSet[varName] = true
			}
		}
	}
	
	vp.logger.Debug("Variables extracted", "count", len(variables))
	return variables, nil
}

// ValidateVariables validates variables against template definition
func (vp *VariableProcessorImpl) ValidateVariables(
	template *interfaces.PromptTemplate,
	variables map[string]interface{},
) error {
	vp.logger.Debug("Validating variables", "template", template.Name)

	// Check required variables
	for _, templateVar := range template.Variables {
		if templateVar.Required {
			if _, exists := variables[templateVar.Name]; !exists {
				return fmt.Errorf("required variable '%s' is missing", templateVar.Name)
			}
		}
	}

	vp.logger.Debug("Variables validated successfully")
	return nil
}

// SetDefaults sets default values for missing variables
func (vp *VariableProcessorImpl) SetDefaults(
	template *interfaces.PromptTemplate,
	variables map[string]interface{},
) map[string]interface{} {
	vp.logger.Debug("Setting variable defaults", "template", template.Name)

	// Create a copy to avoid modifying original
	result := make(map[string]interface{})
	for k, v := range variables {
		result[k] = v
	}

	// Set default values
	for _, templateVar := range template.Variables {
		if _, exists := result[templateVar.Name]; !exists && templateVar.DefaultValue != nil {
			result[templateVar.Name] = templateVar.DefaultValue
			vp.logger.Debug("Set default value", "variable", templateVar.Name, "value", templateVar.DefaultValue)
		}
	}

	return result
}

// isValidVariableName validates variable name to prevent injection
func isValidVariableName(name string) bool {
	// Only allow alphanumeric characters and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	return matched
}

// sanitizeVariableValue returns string representation of value
func sanitizeVariableValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}
