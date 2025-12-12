package prompts

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PromptValidatorImpl implements PromptValidator interface
type PromptValidatorImpl struct {
	config *PromptConfig
	logger interfaces.Logger
}

// NewPromptValidator creates a new PromptValidator instance
func NewPromptValidator(config *PromptConfig, logger interfaces.Logger) interfaces.PromptValidator {
	return &PromptValidatorImpl{
		config: config,
		logger: logger,
	}
}

// ValidateTemplate validates a prompt template
func (pv *PromptValidatorImpl) ValidateTemplate(template *interfaces.PromptTemplate) (*interfaces.ValidationResult, error) {
	pv.logger.Debug("Validating template", "name", template.Name)

	_ = time.Now() // startTime placeholder for future timing implementation
	result := &interfaces.ValidationResult{
		Valid:    true,
		Errors:   make([]*interfaces.ValidationError, 0),
		Warnings: make([]*interfaces.ValidationWarning, 0),
	}

	// Basic validation
	pv.validateBasicFields(template, result)
	
	// Parts validation
	pv.validateParts(template, result)
	
	// Variables validation
	pv.validateVariables(template, result)
	
	// Fashion context validation
	pv.validateFashionContext(template, result)
	
	// Syntax validation
	for _, part := range template.Parts {
		if part.Content != "" {
			pv.validateSyntax(part.Content, result)
		}
	}

	// Note: ValidationResult from interfaces doesn't have Validator/ValidationTime fields
	// These would be in the extended ValidationResult type in types.go

	pv.logger.Debug("Template validation completed", 
		"name", template.Name,
		"valid", result.Valid,
		"errors", len(result.Errors),
		"warnings", len(result.Warnings))

	return result, nil
}

// ValidateVariables validates template variables
func (pv *PromptValidatorImpl) ValidateVariables(template *interfaces.PromptTemplate, variables map[string]interface{}) (*interfaces.ValidationResult, error) {
	pv.logger.Debug("Validating variables", "template", template.Name)

	_ = time.Now() // startTime placeholder for future timing implementation
	result := &interfaces.ValidationResult{
		Valid:    true,
		Errors:   make([]*interfaces.ValidationError, 0),
		Warnings: make([]*interfaces.ValidationWarning, 0),
	}

	// Check required variables
	for _, templateVar := range template.Variables {
		value, exists := variables[templateVar.Name]
		
		if !exists && templateVar.Required {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "MISSING_REQUIRED_VARIABLE",
				Message: fmt.Sprintf("Required variable '%s' is missing", templateVar.Name),
				Field:   templateVar.Name,
			})
			result.Valid = false
			continue
		}

		if exists {
			// Type validation
			if err := pv.validateVariableType(templateVar, value); err != nil {
				result.Errors = append(result.Errors, &interfaces.ValidationError{
					Code:    "INVALID_VARIABLE_TYPE",
					Message: fmt.Sprintf("Variable '%s' has invalid type: %v", templateVar.Name, err),
					Field:   templateVar.Name,
				})
				result.Valid = false
				continue
			}

			// Validation rules
			if err := pv.validateVariableRules(templateVar, value); err != nil {
				result.Errors = append(result.Errors, &interfaces.ValidationError{
					Code:    "VARIABLE_VALIDATION_FAILED",
					Message: fmt.Sprintf("Variable '%s' validation failed: %v", templateVar.Name, err),
					Field:   templateVar.Name,
				})
				result.Valid = false
			}
		}
	}

	// Check for extra variables
	for varName := range variables {
		found := false
		for _, templateVar := range template.Variables {
			if templateVar.Name == varName {
				found = true
				break
			}
		}
		
		if !found {
			result.Warnings = append(result.Warnings, &interfaces.ValidationWarning{
				Code:    "UNKNOWN_VARIABLE",
				Message: fmt.Sprintf("Variable '%s' is not defined in template", varName),
				Field:   varName,
			})
		}
	}

	// Note: ValidationResult from interfaces doesn't have Validator/ValidationTime fields
	// These would be in the extended ValidationResult type in types.go

	pv.logger.Debug("Variable validation completed", 
		"template", template.Name,
		"valid", result.Valid,
		"errors", len(result.Errors),
		"warnings", len(result.Warnings))

	return result, nil
}

// ValidateSyntax validates template syntax
func (pv *PromptValidatorImpl) ValidateSyntax(content string) (*interfaces.ValidationResult, error) {
	pv.logger.Debug("Validating syntax")

	_ = time.Now() // startTime placeholder for future timing implementation
	result := &interfaces.ValidationResult{
		Valid:    true,
		Errors:   make([]*interfaces.ValidationError, 0),
		Warnings: make([]*interfaces.ValidationWarning, 0),
	}

	pv.validateSyntax(content, result)

	// Note: ValidationResult from interfaces doesn't have Validator/ValidationTime fields
	// These would be in the extended ValidationResult type in types.go

	pv.logger.Debug("Syntax validation completed", 
		"valid", result.Valid,
		"errors", len(result.Errors),
		"warnings", len(result.Warnings))

	return result, nil
}

// validateBasicFields validates basic template fields
func (pv *PromptValidatorImpl) validateBasicFields(template *interfaces.PromptTemplate, result *interfaces.ValidationResult) {
	if template.Name == "" {
		result.Errors = append(result.Errors, &interfaces.ValidationError{
			Code:    "MISSING_NAME",
			Message: "Template name is required",
		})
		result.Valid = false
	}

	// Validate name format
	if template.Name != "" {
		if err := pv.validateName(template.Name); err != nil {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "INVALID_NAME",
				Message: fmt.Sprintf("Invalid template name: %v", err),
				Field:   "name",
			})
			result.Valid = false
		}
	}

	if len(template.Parts) == 0 {
		result.Errors = append(result.Errors, &interfaces.ValidationError{
			Code:    "NO_PARTS",
			Message: "Template must have at least one part",
		})
		result.Valid = false
	}

	// Validate version format if provided
	if template.Version != "" {
		if !pv.isValidVersion(template.Version) {
			result.Warnings = append(result.Warnings, &interfaces.ValidationWarning{
				Code:    "INVALID_VERSION_FORMAT",
				Message: "Version should follow semantic versioning (e.g., 1.0.0)",
				Field:   "version",
			})
		}
	}
}

// validateParts validates template parts
func (pv *PromptValidatorImpl) validateParts(template *interfaces.PromptTemplate, result *interfaces.ValidationResult) {
	for i, part := range template.Parts {
		if part.Type == "" {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "MISSING_PART_TYPE",
				Message: fmt.Sprintf("Part %d is missing type", i),
				Field:   fmt.Sprintf("parts[%d].type", i),
			})
			result.Valid = false
			continue
		}

		// Validate part type
		validTypes := []interfaces.PromptPartType{
			interfaces.PromptPartTypeSystem,
			interfaces.PromptPartTypeUser,
			interfaces.PromptPartTypeMedia,
		}

		isValidType := false
		for _, validType := range validTypes {
			if part.Type == validType {
				isValidType = true
				break
			}
		}

		if !isValidType {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "INVALID_PART_TYPE",
				Message: fmt.Sprintf("Part %d has invalid type: %s", i, part.Type),
				Field:   fmt.Sprintf("parts[%d].type", i),
			})
			result.Valid = false
		}

		// Validate media parts
		if part.Type == interfaces.PromptPartTypeMedia {
			if part.Media == nil || part.Media.URL == "" {
				result.Errors = append(result.Errors, &interfaces.ValidationError{
					Code:    "MISSING_MEDIA_URL",
					Message: fmt.Sprintf("Media part %d is missing URL", i),
					Field:   fmt.Sprintf("parts[%d].media.url", i),
				})
				result.Valid = false
			} else if !pv.isValidURL(part.Media.URL) {
				result.Warnings = append(result.Warnings, &interfaces.ValidationWarning{
					Code:    "INVALID_MEDIA_URL",
					Message: fmt.Sprintf("Media part %d has potentially invalid URL: %s", i, part.Media.URL),
					Field:   fmt.Sprintf("parts[%d].media.url", i),
				})
			}
		}

		// Validate content for non-media parts
		if part.Type != interfaces.PromptPartTypeMedia && part.Content == "" {
			result.Warnings = append(result.Warnings, &interfaces.ValidationWarning{
				Code:    "EMPTY_PART_CONTENT",
				Message: fmt.Sprintf("Part %d has empty content", i),
				Field:   fmt.Sprintf("parts[%d].content", i),
			})
		}
	}
}

// validateVariables validates template variables
func (pv *PromptValidatorImpl) validateVariables(template *interfaces.PromptTemplate, result *interfaces.ValidationResult) {
	variableNames := make(map[string]bool)

	for i, variable := range template.Variables {
		if variable.Name == "" {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "MISSING_VARIABLE_NAME",
				Message: fmt.Sprintf("Variable %d is missing name", i),
				Field:   fmt.Sprintf("variables[%d].name", i),
			})
			result.Valid = false
			continue
		}

		// Check for duplicate variable names
		if variableNames[variable.Name] {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "DUPLICATE_VARIABLE_NAME",
				Message: fmt.Sprintf("Duplicate variable name: %s", variable.Name),
				Field:   fmt.Sprintf("variables[%d].name", i),
			})
			result.Valid = false
		}
		variableNames[variable.Name] = true

		// Validate variable name format
		if err := pv.validateVariableName(variable.Name); err != nil {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "INVALID_VARIABLE_NAME",
				Message: fmt.Sprintf("Invalid variable name '%s': %v", variable.Name, err),
				Field:   fmt.Sprintf("variables[%d].name", i),
			})
			result.Valid = false
		}

		// Validate variable type
		if !pv.isValidVariableType(variable.Type) {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "INVALID_VARIABLE_TYPE",
				Message: fmt.Sprintf("Variable '%s' has invalid type: %s", variable.Name, variable.Type),
				Field:   fmt.Sprintf("variables[%d].type", i),
			})
			result.Valid = false
		}

		// Validate validation rules
		if variable.Validation != nil {
			pv.validateVariableValidationRules(variable, result, i)
		}
	}
}

// validateFashionContext validates fashion-specific context
func (pv *PromptValidatorImpl) validateFashionContext(template *interfaces.PromptTemplate, result *interfaces.ValidationResult) {
	if template.FashionContext == nil {
		return
	}

	fc := template.FashionContext

	// Validate marketplace
	if fc.Marketplace != "" {
		validMarketplaces := []string{"wildberries", "ozon", "lamoda", "tsum"}
		isValid := false
		for _, mp := range validMarketplaces {
			if strings.ToLower(fc.Marketplace) == mp {
				isValid = true
				break
			}
		}
		if !isValid {
			result.Warnings = append(result.Warnings, &interfaces.ValidationWarning{
				Code:    "UNKNOWN_MARKETPLACE",
				Message: fmt.Sprintf("Unknown marketplace: %s", fc.Marketplace),
				Field:   "fashion_context.marketplace",
			})
		}
	}

	// Validate language
	if fc.Language != "" {
		validLanguages := []string{"ru", "en", "zh"}
		isValid := false
		for _, lang := range validLanguages {
			if strings.ToLower(fc.Language) == lang {
				isValid = true
				break
			}
		}
		if !isValid {
			result.Warnings = append(result.Warnings, &interfaces.ValidationWarning{
				Code:    "UNKNOWN_LANGUAGE",
				Message: fmt.Sprintf("Unknown language: %s", fc.Language),
				Field:   "fashion_context.language",
			})
		}
	}

	// Validate image analysis context
	if fc.ImageAnalysis != nil {
		ia := fc.ImageAnalysis
		if ia.OutputFormat != "" {
			validFormats := []string{"json", "structured", "narrative"}
			isValid := false
			for _, format := range validFormats {
				if strings.ToLower(ia.OutputFormat) == format {
					isValid = true
					break
				}
			}
			if !isValid {
				result.Warnings = append(result.Warnings, &interfaces.ValidationWarning{
					Code:    "INVALID_IMAGE_OUTPUT_FORMAT",
					Message: fmt.Sprintf("Invalid image analysis output format: %s", ia.OutputFormat),
					Field:   "fashion_context.image_analysis.output_format",
				})
			}
		}
	}
}

// validateSyntax validates template syntax
func (pv *PromptValidatorImpl) validateSyntax(content string, result *interfaces.ValidationResult) {
	// Check for unclosed variable brackets
	openCount := strings.Count(content, "{{")
	closeCount := strings.Count(content, "}}")
	
	if openCount != closeCount {
		result.Errors = append(result.Errors, &interfaces.ValidationError{
			Code:    "UNCLOSED_VARIABLE_BRACKETS",
			Message: fmt.Sprintf("Unclosed variable brackets: %d open, %d close", openCount, closeCount),
		})
		result.Valid = false
	}

	// Check for invalid variable syntax
	varPattern := regexp.MustCompile(`\{\{[^}]*\}\}`)
	matches := varPattern.FindAllString(content, -1)
	
	for _, match := range matches {
		if !strings.HasPrefix(match, "{{") || !strings.HasSuffix(match, "}}") {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "INVALID_VARIABLE_SYNTAX",
				Message: fmt.Sprintf("Invalid variable syntax: %s", match),
			})
			result.Valid = false
		}
	}
}

// Helper validation functions

// validateName validates template name
func (pv *PromptValidatorImpl) validateName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return fmt.Errorf("name must be between 1 and 100 characters")
	}
	
	// Check for invalid characters
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	if !matched {
		return fmt.Errorf("name can only contain letters, numbers, underscores, and hyphens")
	}
	
	return nil
}

// isValidVersion checks if version follows semantic versioning
func (pv *PromptValidatorImpl) isValidVersion(version string) bool {
	// Simple semantic versioning check
	semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`)
	return semverPattern.MatchString(version)
}

// isValidURL checks if URL is valid (basic check)
func (pv *PromptValidatorImpl) isValidURL(url string) bool {
	if url == "" {
		return false
	}
	
	// Basic URL pattern check
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlPattern.MatchString(url) || strings.HasPrefix(url, "/") || strings.HasPrefix(url, "data:")
}

// validateVariableName validates variable name format
func (pv *PromptValidatorImpl) validateVariableName(name string) error {
	if len(name) == 0 || len(name) > 50 {
		return fmt.Errorf("variable name must be between 1 and 50 characters")
	}
	
	// Check for valid characters (letters, numbers, underscores)
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	if !matched {
		return fmt.Errorf("variable name must start with letter or underscore and contain only letters, numbers, and underscores")
	}
	
	return nil
}

// isValidVariableType checks if variable type is valid
func (pv *PromptValidatorImpl) isValidVariableType(varType string) bool {
	validTypes := []string{"string", "number", "boolean", "array", "object"}
	for _, validType := range validTypes {
		if varType == validType {
			return true
		}
	}
	return false
}

// validateVariableType validates variable value against expected type
func (pv *PromptValidatorImpl) validateVariableType(variable *interfaces.PromptVariable, value interface{}) error {
	switch variable.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			if _, ok := value.(int); !ok {
				return fmt.Errorf("expected number, got %T", value)
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	default:
		return fmt.Errorf("unknown variable type: %s", variable.Type)
	}
	
	return nil
}

// validateVariableRules validates variable value against validation rules
func (pv *PromptValidatorImpl) validateVariableRules(variable *interfaces.PromptVariable, value interface{}) error {
	if variable.Validation == nil {
		return nil
	}

	rules := variable.Validation

	// String validation
	if str, ok := value.(string); ok {
		if rules.MinLength != nil && len(str) < *rules.MinLength {
			return fmt.Errorf("string too short, minimum length: %d", *rules.MinLength)
		}
		if rules.MaxLength != nil && len(str) > *rules.MaxLength {
			return fmt.Errorf("string too long, maximum length: %d", *rules.MaxLength)
		}
		if rules.Pattern != "" {
			matched, _ := regexp.MatchString(rules.Pattern, str)
			if !matched {
				return fmt.Errorf("string does not match pattern: %s", rules.Pattern)
			}
		}
	}

	// Number validation
	if num, ok := value.(float64); ok {
		if rules.Min != nil && num < *rules.Min {
			return fmt.Errorf("number too small, minimum: %f", *rules.Min)
		}
		if rules.Max != nil && num > *rules.Max {
			return fmt.Errorf("number too large, maximum: %f", *rules.Max)
		}
	}

	// Enum validation
	if len(rules.Enum) > 0 {
		strValue := fmt.Sprintf("%v", value)
		found := false
		for _, enumValue := range rules.Enum {
			if enumValue == strValue {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value not in allowed enum: %v", rules.Enum)
		}
	}

	return nil
}

// validateVariableValidationRules validates the validation rules themselves
func (pv *PromptValidatorImpl) validateVariableValidationRules(variable *interfaces.PromptVariable, result *interfaces.ValidationResult, index int) {
	rules := variable.Validation

	// Validate min/max consistency
	if rules.Min != nil && rules.Max != nil && *rules.Min > *rules.Max {
		result.Errors = append(result.Errors, &interfaces.ValidationError{
			Code:    "INVALID_MIN_MAX",
			Message: fmt.Sprintf("Variable '%s' has min > max", variable.Name),
			Field:   fmt.Sprintf("variables[%d].validation", index),
		})
		result.Valid = false
	}

	// Validate min/max length consistency
	if rules.MinLength != nil && rules.MaxLength != nil && *rules.MinLength > *rules.MaxLength {
		result.Errors = append(result.Errors, &interfaces.ValidationError{
			Code:    "INVALID_MIN_MAX_LENGTH",
			Message: fmt.Sprintf("Variable '%s' has min_length > max_length", variable.Name),
			Field:   fmt.Sprintf("variables[%d].validation", index),
		})
		result.Valid = false
	}

	// Validate regex pattern
	if rules.Pattern != "" {
		if _, err := regexp.Compile(rules.Pattern); err != nil {
			result.Errors = append(result.Errors, &interfaces.ValidationError{
				Code:    "INVALID_REGEX_PATTERN",
				Message: fmt.Sprintf("Variable '%s' has invalid regex pattern: %v", variable.Name, err),
				Field:   fmt.Sprintf("variables[%d].validation.pattern", index),
			})
			result.Valid = false
		}
	}
}