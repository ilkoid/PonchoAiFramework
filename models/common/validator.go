package common

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidationError represents a specific validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
	Rule    string `json:"rule"`
}

// Validator provides validation for model requests and responses
type Validator struct {
	rules []ValidationRule
	logger interfaces.Logger
}

// NewValidator creates a new validator with rules
func NewValidator(rules []ValidationRule, logger interfaces.Logger) *Validator {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	// Use default rules if none provided
	if len(rules) == 0 {
		rules = DefaultValidationRules
	}

	return &Validator{
		rules: rules,
		logger: logger,
	}
}

// ValidateRequest validates a PonchoModelRequest
func (v *Validator) ValidateRequest(req *interfaces.PonchoModelRequest) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	if req == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "request",
			Message: "request cannot be nil",
			Rule:    "required",
		})
		return result
	}

	// Apply validation rules
	for _, rule := range v.rules {
		if err := v.applyRule(req, rule); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, *err)
		}
	}

	// Custom validations for model requests
	v.validateModelRequest(req, result)

	return result
}

// ValidateResponse validates a PonchoModelResponse
func (v *Validator) ValidateResponse(resp *interfaces.PonchoModelResponse) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	if resp == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "response",
			Message: "response cannot be nil",
			Rule:    "required",
		})
		return result
	}

	// Validate message
	if resp.Message == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "message",
			Message: "message cannot be nil",
			Rule:    "required",
		})
	}

	// Validate usage if present
	if resp.Usage != nil {
		v.validateUsage(resp.Usage, result)
	}

	return result
}

// applyRule applies a single validation rule
func (v *Validator) applyRule(req *interfaces.PonchoModelRequest, rule ValidationRule) *ValidationError {
	value := v.getFieldValue(req, rule.Field)
	if value == nil {
		if rule.Required {
			return &ValidationError{
				Field:   rule.Field,
				Message: fmt.Sprintf("field %s is required", rule.Field),
				Rule:    "required",
			}
		}
		return nil
	}

	// Type validation
	if err := v.validateType(value, rule.Type); err != nil {
		return &ValidationError{
			Field:   rule.Field,
			Message: err.Error(),
			Value:   value,
			Rule:    rule.Type,
		}
	}

	// Length validation
	if rule.MinLength != nil || rule.MaxLength != nil {
		if strValue, ok := value.(string); ok {
			if err := v.validateLength(strValue, rule.MinLength, rule.MaxLength); err != nil {
				return &ValidationError{
					Field:   rule.Field,
					Message: err.Error(),
					Value:   value,
					Rule:    "length",
				}
			}
		}
	}

	// Range validation
	if rule.MinValue != nil || rule.MaxValue != nil {
		if numValue, ok := v.toFloat64(value); ok {
			if err := v.validateRange(numValue, rule.MinValue, rule.MaxValue); err != nil {
				return &ValidationError{
					Field:   rule.Field,
					Message: err.Error(),
					Value:   value,
					Rule:    "range",
				}
			}
		}
	}

	// Pattern validation
	if rule.Pattern != nil {
		if strValue, ok := value.(string); ok {
			if err := v.validatePattern(strValue, *rule.Pattern); err != nil {
				return &ValidationError{
					Field:   rule.Field,
					Message: err.Error(),
					Value:   value,
					Rule:    "pattern",
				}
			}
		}
	}

	// Enum validation
	if len(rule.Enum) > 0 {
		if err := v.validateEnum(value, rule.Enum); err != nil {
			return &ValidationError{
				Field:   rule.Field,
				Message: err.Error(),
				Value:   value,
				Rule:    "enum",
			}
		}
	}

	// Custom validation
	if rule.Custom != nil {
		if err := rule.Custom(value); err != nil {
			return &ValidationError{
				Field:   rule.Field,
				Message: err.Error(),
				Value:   value,
				Rule:    "custom",
			}
		}
	}

	return nil
}

// getFieldValue extracts field value from request using reflection
func (v *Validator) getFieldValue(req *interfaces.PonchoModelRequest, fieldName string) interface{} {
	val := reflect.ValueOf(req).Elem()
	field := val.FieldByName(fieldName)
	
	if !field.IsValid() {
		return nil
	}

	return field.Interface()
}

// validateType validates the type of a value
func (v *Validator) validateType(value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		if !v.isNumber(value) {
			return fmt.Errorf("expected number, got %T", value)
		}
	case "array":
		if !v.isArray(value) {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "object":
		if !v.isObject(value) {
			return fmt.Errorf("expected object, got %T", value)
		}
	}
	return nil
}

// validateLength validates string length
func (v *Validator) validateLength(value string, minLen, maxLen *int) error {
	length := len(value)

	if minLen != nil && length < *minLen {
		return fmt.Errorf("length %d is less than minimum %d", length, *minLen)
	}

	if maxLen != nil && length > *maxLen {
		return fmt.Errorf("length %d exceeds maximum %d", length, *maxLen)
	}

	return nil
}

// validateRange validates numeric range
func (v *Validator) validateRange(value float64, minVal, maxVal *float64) error {
	if minVal != nil && value < *minVal {
		return fmt.Errorf("value %g is less than minimum %g", value, *minVal)
	}

	if maxVal != nil && value > *maxVal {
		return fmt.Errorf("value %g exceeds maximum %g", value, *maxVal)
	}

	return nil
}

// validatePattern validates string against regex pattern
func (v *Validator) validatePattern(value string, pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	if !regex.MatchString(value) {
		return fmt.Errorf("value %s does not match pattern %s", value, pattern)
	}

	return nil
}

// validateEnum validates value against allowed enum values
func (v *Validator) validateEnum(value interface{}, enum []string) error {
	valueStr := fmt.Sprintf("%v", value)
	
	for _, allowed := range enum {
		if valueStr == allowed {
			return nil
		}
	}

	return fmt.Errorf("value %s is not in allowed values: %v", valueStr, enum)
}

// validateModelRequest performs model-specific validations
func (v *Validator) validateModelRequest(req *interfaces.PonchoModelRequest, result *ValidationResult) {
	// Validate model name
	if req.Model == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "model",
			Message: "model name is required",
			Rule:    "required",
		})
	}

	// Validate messages
	if len(req.Messages) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "messages",
			Message: "at least one message is required",
			Rule:    "required",
		})
	} else {
		// Validate each message
		for i, msg := range req.Messages {
			v.validateMessage(msg, i, result)
		}
	}

	// Validate temperature if provided
	if req.Temperature != nil {
		temp := float64(*req.Temperature)
		if temp < 0.0 || temp > 2.0 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "temperature",
				Message: "temperature must be between 0.0 and 2.0",
				Value:   temp,
				Rule:    "range",
			})
		}
	}

	// Validate max_tokens if provided
	if req.MaxTokens != nil {
		maxTokens := *req.MaxTokens
		if maxTokens <= 0 || maxTokens > 32000 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "max_tokens",
				Message: "max_tokens must be between 1 and 32000",
				Value:   maxTokens,
				Rule:    "range",
			})
		}
	}

	// Note: top_p field not available in PonchoModelRequest
	// This validation would be added if the field is added to the interface
}

// validateMessage validates a single message
func (v *Validator) validateMessage(msg *interfaces.PonchoMessage, index int, result *ValidationResult) {
	// Validate role
	if msg.Role == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fmt.Sprintf("messages[%d].role", index),
			Message: "message role is required",
			Rule:    "required",
		})
	} else {
		// Validate role value
		validRoles := []string{"system", "user", "assistant", "tool"}
		if !v.contains(validRoles, string(msg.Role)) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("messages[%d].role", index),
				Message: fmt.Sprintf("invalid role: %s", msg.Role),
				Value:   string(msg.Role),
				Rule:    "enum",
			})
		}
	}

	// Validate content
	if len(msg.Content) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fmt.Sprintf("messages[%d].content", index),
			Message: "message content is required",
			Rule:    "required",
		})
	} else {
		// Validate each content part
		for j, part := range msg.Content {
			v.validateContentPart(part, index, j, result)
		}
	}
}

// validateContentPart validates a single content part
func (v *Validator) validateContentPart(part *interfaces.PonchoContentPart, msgIndex, partIndex int, result *ValidationResult) {
	fieldName := fmt.Sprintf("messages[%d].content[%d]", msgIndex, partIndex)

	// Validate type
	if part.Type == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName + ".type",
			Message: "content part type is required",
			Rule:    "required",
		})
	} else {
		// Validate type value
		validTypes := []string{"text", "media", "tool"}
		if !v.contains(validTypes, string(part.Type)) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName + ".type",
				Message: fmt.Sprintf("invalid content type: %s", part.Type),
				Value:   string(part.Type),
				Rule:    "enum",
			})
		}
	}

	// Type-specific validations
	switch part.Type {
	case interfaces.PonchoContentTypeText:
		if part.Text == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName + ".text",
				Message: "text content is required for text type",
				Rule:    "required",
			})
		}
	case interfaces.PonchoContentTypeMedia:
		if part.Media == nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName + ".media",
				Message: "media content is required for media type",
				Rule:    "required",
			})
		} else {
			// Validate media URL
			if part.Media.URL == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   fieldName + ".media.url",
					Message: "media URL is required",
					Rule:    "required",
				})
			} else {
				// Validate URL format
				if err := v.validateURL(part.Media.URL); err != nil {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fieldName + ".media.url",
						Message: err.Error(),
						Value:   part.Media.URL,
						Rule:    "url",
					})
				}
			}
		}
	case interfaces.PonchoContentTypeTool:
		if part.Tool == nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName + ".tool",
				Message: "tool content is required for tool type",
				Rule:    "required",
			})
		} else {
			// Validate tool ID
			if part.Tool.ID == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   fieldName + ".tool.id",
					Message: "tool ID is required",
					Rule:    "required",
				})
			}
			// Validate tool name
			if part.Tool.Name == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   fieldName + ".tool.name",
					Message: "tool name is required",
					Rule:    "required",
				})
			}
		}
	}
}

// validateUsage validates token usage information
func (v *Validator) validateUsage(usage *interfaces.PonchoUsage, result *ValidationResult) {
	if usage.PromptTokens < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "usage.prompt_tokens",
			Message: "prompt_tokens must be non-negative",
			Value:   usage.PromptTokens,
			Rule:    "range",
		})
	}

	if usage.CompletionTokens < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "usage.completion_tokens",
			Message: "completion_tokens must be non-negative",
			Value:   usage.CompletionTokens,
			Rule:    "range",
		})
	}

	if usage.TotalTokens < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "usage.total_tokens",
			Message: "total_tokens must be non-negative",
			Value:   usage.TotalTokens,
			Rule:    "range",
		})
	}

	// Validate total consistency
	expectedTotal := usage.PromptTokens + usage.CompletionTokens
	if usage.TotalTokens != expectedTotal {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "usage.total_tokens",
			Message: fmt.Sprintf("total_tokens (%d) must equal prompt_tokens (%d) + completion_tokens (%d)", 
				usage.TotalTokens, usage.PromptTokens, usage.CompletionTokens),
			Value:   usage.TotalTokens,
			Rule:    "consistency",
		})
	}
}

// validateURL validates URL format
func (v *Validator) validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Check if it's a valid URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must have a scheme (http/https)")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	return nil
}

// Helper functions for type checking

func (v *Validator) isNumber(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}

func (v *Validator) isArray(value interface{}) bool {
	kind := reflect.TypeOf(value).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

func (v *Validator) isObject(value interface{}) bool {
	kind := reflect.TypeOf(value).Kind()
	return kind == reflect.Map || kind == reflect.Struct
}

func (v *Validator) toFloat64(value interface{}) (float64, bool) {
	switch val := value.(type) {
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	case string:
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func (v *Validator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidationRuleBuilder helps build validation rules
type ValidationRuleBuilder struct {
	rule ValidationRule
}

// NewValidationRuleBuilder creates a new validation rule builder
func NewValidationRuleBuilder(field string) *ValidationRuleBuilder {
	return &ValidationRuleBuilder{
		rule: ValidationRule{
			Field: field,
		},
	}
}

// Required marks the field as required
func (b *ValidationRuleBuilder) Required() *ValidationRuleBuilder {
	b.rule.Required = true
	return b
}

// Type sets the expected type
func (b *ValidationRuleBuilder) Type(typeName string) *ValidationRuleBuilder {
	b.rule.Type = typeName
	return b
}

// MinLength sets minimum length for strings
func (b *ValidationRuleBuilder) MinLength(min int) *ValidationRuleBuilder {
	b.rule.MinLength = &min
	return b
}

// MaxLength sets maximum length for strings
func (b *ValidationRuleBuilder) MaxLength(max int) *ValidationRuleBuilder {
	b.rule.MaxLength = &max
	return b
}

// MinValue sets minimum value for numbers
func (b *ValidationRuleBuilder) MinValue(min float64) *ValidationRuleBuilder {
	b.rule.MinValue = &min
	return b
}

// MaxValue sets maximum value for numbers
func (b *ValidationRuleBuilder) MaxValue(max float64) *ValidationRuleBuilder {
	b.rule.MaxValue = &max
	return b
}

// Pattern sets regex pattern for validation
func (b *ValidationRuleBuilder) Pattern(pattern string) *ValidationRuleBuilder {
	b.rule.Pattern = &pattern
	return b
}

// Enum sets allowed values
func (b *ValidationRuleBuilder) Enum(values ...string) *ValidationRuleBuilder {
	b.rule.Enum = values
	return b
}

// Custom sets a custom validation function
func (b *ValidationRuleBuilder) Custom(fn func(interface{}) error) *ValidationRuleBuilder {
	b.rule.Custom = fn
	return b
}

// Build creates the final validation rule
func (b *ValidationRuleBuilder) Build() ValidationRule {
	return b.rule
}

// Helper functions for common validation rules

// RequiredRule creates a required field rule
func RequiredRule(field string) ValidationRule {
	return NewValidationRuleBuilder(field).Required().Build()
}

// StringTypeRule creates a string type rule
func StringTypeRule(field string) ValidationRule {
	return NewValidationRuleBuilder(field).Type("string").Build()
}

// NumberTypeRule creates a number type rule
func NumberTypeRule(field string) ValidationRule {
	return NewValidationRuleBuilder(field).Type("number").Build()
}

// ArrayTypeRule creates an array type rule
func ArrayTypeRule(field string) ValidationRule {
	return NewValidationRuleBuilder(field).Type("array").Build()
}

// RangeRule creates a range validation rule
func RangeRule(field string, min, max float64) ValidationRule {
	return NewValidationRuleBuilder(field).Type("number").MinValue(min).MaxValue(max).Build()
}

// LengthRule creates a length validation rule
func LengthRule(field string, min, max int) ValidationRule {
	return NewValidationRuleBuilder(field).Type("string").MinLength(min).MaxLength(max).Build()
}

// EnumRule creates an enum validation rule
func EnumRule(field string, values ...string) ValidationRule {
	return NewValidationRuleBuilder(field).Enum(values...).Build()
}

// PatternRule creates a pattern validation rule
func PatternRule(field string, pattern string) ValidationRule {
	return NewValidationRuleBuilder(field).Pattern(pattern).Build()
}