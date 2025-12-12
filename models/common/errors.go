package common

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "low"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityCritical ErrorSeverity = "critical"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	ErrorCategoryAuthentication ErrorCategory = "authentication"
	ErrorCategoryAuthorization  ErrorCategory = "authorization"
	ErrorCategoryRateLimit     ErrorCategory = "rate_limit"
	ErrorCategoryValidation    ErrorCategory = "validation"
	ErrorCategoryNetwork       ErrorCategory = "network"
	ErrorCategoryServer       ErrorCategory = "server"
	ErrorCategoryClient       ErrorCategory = "client"
	ErrorCategoryParsing      ErrorCategory = "parsing"
	ErrorCategoryTimeout      ErrorCategory = "timeout"
	ErrorCategoryQuota        ErrorCategory = "quota"
	ErrorCategoryContent      ErrorCategory = "content_filter"
	ErrorCategoryUnknown      ErrorCategory = "unknown"
)

// ModelError represents a standardized model error
type ModelError struct {
	Code         string         `json:"code"`
	Type         string         `json:"type"`
	Message      string         `json:"message"`
	Provider     Provider       `json:"provider"`
	Category     ErrorCategory  `json:"category"`
	Severity     ErrorSeverity  `json:"severity"`
	Retryable    bool           `json:"retryable"`
	Timestamp    time.Time      `json:"timestamp"`
	RequestID    string         `json:"request_id,omitempty"`
	StatusCode   int            `json:"status_code,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Cause        error          `json:"-"`
}

// Error implements the error interface
func (e *ModelError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("[%s] %s: %s (RequestID: %s)", e.Provider, e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *ModelError) Unwrap() error {
	return e.Cause
}

// WithCause adds a cause to the error
func (e *ModelError) WithCause(cause error) *ModelError {
	e.Cause = cause
	return e
}

// WithDetail adds a detail to the error
func (e *ModelError) WithDetail(key string, value interface{}) *ModelError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithRequestID adds a request ID to the error
func (e *ModelError) WithRequestID(requestID string) *ModelError {
	e.RequestID = requestID
	return e
}

// ErrorClassifier classifies errors from different providers
type ErrorClassifier struct {
	provider Provider
	logger   interfaces.Logger
}

// NewErrorClassifier creates a new error classifier
func NewErrorClassifier(provider Provider, logger interfaces.Logger) *ErrorClassifier {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}
	return &ErrorClassifier{
		provider: provider,
		logger:   logger,
	}
}

// ClassifyError classifies an error into a standardized ModelError
func (ec *ErrorClassifier) ClassifyError(err error, statusCode int, responseBody []byte) *ModelError {
	if err == nil {
		return nil
	}

	// Try to parse provider-specific error response first
	if len(responseBody) > 0 {
		if parsedErr := ec.parseProviderError(responseBody, statusCode); parsedErr != nil {
			return parsedErr
		}
	}

	// Classify based on HTTP status code
	if statusCode > 0 {
		if classifiedErr := ec.classifyByStatusCode(statusCode); classifiedErr != nil {
			return classifiedErr
		}
	}

	// Classify based on error message
	return ec.classifyByMessage(err)
}

// parseProviderError attempts to parse provider-specific error response
func (ec *ErrorClassifier) parseProviderError(responseBody []byte, statusCode int) *ModelError {
	var errorResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
		return nil
	}

	switch ec.provider {
	case ProviderDeepSeek:
		return ec.parseDeepSeekError(errorResponse, statusCode)
	case ProviderZAI:
		return ec.parseZAIError(errorResponse, statusCode)
	case ProviderOpenAI:
		return ec.parseOpenAIError(errorResponse, statusCode)
	default:
		return nil
	}
}

// parseDeepSeekError parses DeepSeek-specific error format
func (ec *ErrorClassifier) parseDeepSeekError(errorResponse map[string]interface{}, statusCode int) *ModelError {
	errorObj, hasError := errorResponse["error"].(map[string]interface{})
	if !hasError {
		return nil
	}

	message, _ := errorObj["message"].(string)
	errorType, _ := errorObj["type"].(string)
	code, _ := errorObj["code"].(string)

	return ec.createModelError(code, errorType, message, statusCode)
}

// parseZAIError parses Z.AI-specific error format
func (ec *ErrorClassifier) parseZAIError(errorResponse map[string]interface{}, statusCode int) *ModelError {
	errorObj, hasError := errorResponse["error"].(map[string]interface{})
	if !hasError {
		return nil
	}

	message, _ := errorObj["message"].(string)
	errorType, _ := errorObj["type"].(string)
	code, _ := errorObj["code"].(string)

	return ec.createModelError(code, errorType, message, statusCode)
}

// parseOpenAIError parses OpenAI-specific error format
func (ec *ErrorClassifier) parseOpenAIError(errorResponse map[string]interface{}, statusCode int) *ModelError {
	errorObj, hasError := errorResponse["error"].(map[string]interface{})
	if !hasError {
		return nil
	}

	message, _ := errorObj["message"].(string)
	errorType, _ := errorObj["type"].(string)
	code, _ := errorObj["code"].(string)

	return ec.createModelError(code, errorType, message, statusCode)
}

// classifyByStatusCode classifies errors based on HTTP status code
func (ec *ErrorClassifier) classifyByStatusCode(statusCode int) *ModelError {
	switch {
	case statusCode == 401:
		return ec.createModelError(ErrCodeInvalidAuth, "authentication_error", "Invalid authentication credentials", statusCode)
	case statusCode == 403:
		return ec.createModelError(ErrCodeInvalidAuth, "authorization_error", "Access forbidden", statusCode)
	case statusCode == 429:
		return ec.createModelError(ErrCodeRateLimit, "rate_limit_error", "Rate limit exceeded", statusCode)
	case statusCode >= 400 && statusCode < 500:
		return ec.createModelError(ErrCodeInvalidRequest, "client_error", "Client error", statusCode)
	case statusCode >= 500:
		return ec.createModelError(ErrCodeServerError, "server_error", "Server error", statusCode)
	default:
		return ec.createModelError(ErrCodeUnknown, "unknown_error", "Unknown error occurred", statusCode)
	}
}

// classifyByMessage classifies errors based on error message content
func (ec *ErrorClassifier) classifyByMessage(err error) *ModelError {
	message := strings.ToLower(err.Error())

	// Network-related errors
	if strings.Contains(message, "connection refused") ||
		strings.Contains(message, "network is unreachable") ||
		strings.Contains(message, "no such host") ||
		strings.Contains(message, "connection timeout") {
		return ec.createModelError(ErrCodeNetworkError, "network_error", err.Error(), 0)
	}

	// Timeout errors
	if strings.Contains(message, "timeout") ||
		strings.Contains(message, "deadline exceeded") {
		return ec.createModelError(ErrCodeTimeout, "timeout_error", err.Error(), 0)
	}

	// JSON parsing errors
	if strings.Contains(message, "json") ||
		strings.Contains(message, "unmarshal") ||
		strings.Contains(message, "invalid character") {
		return ec.createModelError(ErrCodeParsingError, "parsing_error", err.Error(), 0)
	}

	// Default classification
	return ec.createModelError(ErrCodeUnknown, "unknown_error", err.Error(), 0)
}

// createModelError creates a standardized ModelError
func (ec *ErrorClassifier) createModelError(code, errorType, message string, statusCode int) *ModelError {
	category := ec.determineCategory(code, statusCode)
	severity := ec.determineSeverity(category, statusCode)
	retryable := ec.isRetryable(code, statusCode)

	return &ModelError{
		Code:       code,
		Type:       errorType,
		Message:    message,
		Provider:   ec.provider,
		Category:   category,
		Severity:   severity,
		Retryable:  retryable,
		Timestamp:  time.Now(),
		StatusCode: statusCode,
	}
}

// determineCategory determines the error category based on error code and status
func (ec *ErrorClassifier) determineCategory(code string, statusCode int) ErrorCategory {
	switch code {
	case ErrCodeInvalidAuth:
		return ErrorCategoryAuthentication
	case ErrCodeRateLimit:
		return ErrorCategoryRateLimit
	case ErrCodeInsufficientQuota:
		return ErrorCategoryQuota
	case ErrCodeModelNotFound:
		return ErrorCategoryValidation
	case ErrCodeContentFilter:
		return ErrorCategoryContent
	case ErrCodeServerError:
		return ErrorCategoryServer
	case ErrCodeTimeout:
		return ErrorCategoryTimeout
	case ErrCodeNetworkError:
		return ErrorCategoryNetwork
	case ErrCodeParsingError:
		return ErrorCategoryParsing
	case ErrCodeValidationError:
		return ErrorCategoryValidation
	default:
		// Fallback to status code classification
		switch statusCode {
		case 401:
			return ErrorCategoryAuthentication
		case 403:
			return ErrorCategoryAuthorization
		case 429:
			return ErrorCategoryRateLimit
		case 400, 422:
			return ErrorCategoryValidation
		case 404:
			return ErrorCategoryClient
		default:
			if statusCode >= 500 {
				return ErrorCategoryServer
			} else if statusCode >= 400 {
				return ErrorCategoryClient
			}
			return ErrorCategoryUnknown
		}
	}
}

// determineSeverity determines the error severity
func (ec *ErrorClassifier) determineSeverity(category ErrorCategory, statusCode int) ErrorSeverity {
	switch category {
	case ErrorCategoryAuthentication, ErrorCategoryAuthorization:
		return ErrorSeverityHigh
	case ErrorCategoryRateLimit:
		return ErrorSeverityMedium
	case ErrorCategoryQuota:
		return ErrorSeverityHigh
	case ErrorCategoryContent:
		return ErrorSeverityLow
	case ErrorCategoryServer:
		return ErrorSeverityCritical
	case ErrorCategoryNetwork, ErrorCategoryTimeout:
		return ErrorSeverityMedium
	case ErrorCategoryValidation:
		return ErrorSeverityLow
	default:
		// Fallback to status code
		if statusCode >= 500 {
			return ErrorSeverityCritical
		} else if statusCode >= 400 {
			return ErrorSeverityMedium
		}
		return ErrorSeverityLow
	}
}

// isRetryable determines if an error is retryable
func (ec *ErrorClassifier) isRetryable(code string, statusCode int) bool {
	switch code {
	case ErrCodeRateLimit, ErrCodeServerError, ErrCodeTimeout, ErrCodeNetworkError:
		return true
	default:
		// Fallback to status code
		return statusCode == 429 || statusCode >= 500
	}
}

// ErrorHandler provides centralized error handling utilities
type ErrorHandler struct {
	classifier *ErrorClassifier
	logger     interfaces.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(provider Provider, logger interfaces.Logger) *ErrorHandler {
	return &ErrorHandler{
		classifier: NewErrorClassifier(provider, logger),
		logger:     logger,
	}
}

// HandleError handles and logs an error
func (eh *ErrorHandler) HandleError(err error, statusCode int, responseBody []byte, requestID string) *ModelError {
	if err == nil {
		return nil
	}

	modelErr := eh.classifier.ClassifyError(err, statusCode, responseBody)
	if modelErr != nil && requestID != "" {
		modelErr.WithRequestID(requestID)
	}

	// Log the error
	eh.logger.Error("Model error occurred",
		"provider", modelErr.Provider,
		"code", modelErr.Code,
		"type", modelErr.Type,
		"message", modelErr.Message,
		"category", modelErr.Category,
		"severity", modelErr.Severity,
		"retryable", modelErr.Retryable,
		"status_code", modelErr.StatusCode,
		"request_id", modelErr.RequestID)

	return modelErr
}

// IsRetryableError checks if an error is retryable
func (eh *ErrorHandler) IsRetryableError(err error) bool {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr.Retryable
	}
	return false
}

// GetRetryDelay calculates retry delay based on error type
func (eh *ErrorHandler) GetRetryDelay(err error, attempt int) time.Duration {
	if modelErr, ok := err.(*ModelError); ok {
		switch modelErr.Category {
		case ErrorCategoryRateLimit:
			// Exponential backoff for rate limits
			return time.Duration(attempt*attempt) * time.Second
		case ErrorCategoryServer:
			// Longer backoff for server errors
			return time.Duration(attempt*attempt*2) * time.Second
		case ErrorCategoryNetwork, ErrorCategoryTimeout:
			// Moderate backoff for network issues
			return time.Duration(attempt) * time.Second
		default:
			// Default backoff
			return time.Duration(attempt) * 500 * time.Millisecond
		}
	}
	return time.Second
}

// Helper functions

// CreateModelError creates a new ModelError
func CreateModelError(provider Provider, code, errorType, message string, statusCode int) *ModelError {
	classifier := &ErrorClassifier{provider: provider}
	return classifier.createModelError(code, errorType, message, statusCode)
}

// IsAuthenticationError checks if error is authentication-related
func IsAuthenticationError(err error) bool {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr.Category == ErrorCategoryAuthentication
	}
	return false
}

// IsRateLimitError checks if error is rate limit related
func IsRateLimitError(err error) bool {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr.Category == ErrorCategoryRateLimit
	}
	return false
}

// IsServerError checks if error is server-related
func IsServerError(err error) bool {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr.Category == ErrorCategoryServer
	}
	return false
}

// IsValidationError checks if error is validation-related
func IsValidationError(err error) bool {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr.Category == ErrorCategoryValidation
	}
	return false
}

// ExtractErrorCode extracts error code from various error formats
func ExtractErrorCode(err error) string {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr.Code
	}

	// Try to extract from error message using regex
	patterns := []struct {
		pattern *regexp.Regexp
		code    string
	}{
		{regexp.MustCompile(`(?i)invalid.*auth`), ErrCodeInvalidAuth},
		{regexp.MustCompile(`(?i)rate.*limit`), ErrCodeRateLimit},
		{regexp.MustCompile(`(?i)quota`), ErrCodeInsufficientQuota},
		{regexp.MustCompile(`(?i)timeout`), ErrCodeTimeout},
		{regexp.MustCompile(`(?i)network`), ErrCodeNetworkError},
		{regexp.MustCompile(`(?i)server.*error`), ErrCodeServerError},
	}

	message := strings.ToLower(err.Error())
	for _, p := range patterns {
		if p.pattern.MatchString(message) {
			return p.code
		}
	}

	return ErrCodeUnknown
}