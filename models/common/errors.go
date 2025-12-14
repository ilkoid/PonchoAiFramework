// Package common provides unified error types and handling for AI model providers
// in PonchoFramework. This file implements a comprehensive error system
// with provider-specific error codes, retry logic, and detailed error context.
//
// Key Features:
// - Unified error type for all model providers
// - Provider-specific error codes and messages
// - Automatic retryability determination
// - HTTP status code to error mapping
// - Error wrapping with cause preservation
// - Structured error information for debugging
//
// Error Categories:
// - Configuration Errors: Invalid config, missing API keys
// - Network Errors: Timeouts, connection issues, rate limits
// - API Errors: Invalid requests, authorization, server errors
// - Model-Specific Errors: Model not found, token limits, content filtering
// - Streaming Errors: Stream interruptions, parsing errors
//
// Error Codes:
// - INVALID_CONFIG: Configuration validation failed
// - RATE_LIMIT_ERROR: API rate limit exceeded (retryable)
// - TOKEN_LIMIT_EXCEEDED: Model token limit exceeded
// - CONTENT_FILTERED: Content blocked by safety filters
// - SERVER_ERROR: Internal server error (retryable)
//
// Retry Logic:
// - Automatic retryability determination based on error type
// - HTTP status code analysis for retry decisions
// - Provider-specific retry strategies
// - Circuit breaker integration support
//
// Usage Example:
//   err := NewRateLimitError("Too many requests", "deepseek", "deepseek-chat")
//   if IsRetryableModelError(err) {
//       // implement retry logic
//   }
//
// Error Context:
// - Provider and model information
// - HTTP status codes
// - Retryability flags
// - Detailed error messages
// - Optional error details and causes
package common

import (
	"fmt"
	"net/http"
)

// ModelError represents a unified error type for all models
type ModelError struct {
	Code        ModelErrorCode `json:"code"`
	Message     string        `json:"message"`
	Provider    string        `json:"provider"`
	Model       string        `json:"model"`
	StatusCode  int           `json:"status_code,omitempty"`
	Retryable   bool          `json:"retryable"`
	Details     interface{}   `json:"details,omitempty"`
	Cause       error         `json:"cause,omitempty"`
}

// ModelErrorCode represents different types of model errors
type ModelErrorCode string

const (
	// Configuration errors
	ErrorCodeInvalidConfig     ModelErrorCode = "INVALID_CONFIG"
	ErrorCodeMissingConfig     ModelErrorCode = "MISSING_CONFIG"
	ErrorCodeInvalidAPIKey    ModelErrorCode = "INVALID_API_KEY"
	ErrorCodeMissingAPIKey    ModelErrorCode = "MISSING_API_KEY"

	// Network errors
	ErrorCodeNetworkError      ModelErrorCode = "NETWORK_ERROR"
	ErrorCodeTimeoutError     ModelErrorCode = "TIMEOUT_ERROR"
	ErrorCodeConnectionError  ModelErrorCode = "CONNECTION_ERROR"
	ErrorCodeRateLimitError   ModelErrorCode = "RATE_LIMIT_ERROR"

	// API errors
	ErrorCodeInvalidRequest   ModelErrorCode = "INVALID_REQUEST"
	ErrorCodeUnauthorized     ModelErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden        ModelErrorCode = "FORBIDDEN"
	ErrorCodeNotFound        ModelErrorCode = "NOT_FOUND"
	ErrorCodeServerError      ModelErrorCode = "SERVER_ERROR"
	ErrorCodeServiceUnavailable ModelErrorCode = "SERVICE_UNAVAILABLE"

	// Model-specific errors
	ErrorCodeModelNotFound    ModelErrorCode = "MODEL_NOT_FOUND"
	ErrorCodeInvalidModel     ModelErrorCode = "INVALID_MODEL"
	ErrorCodeTokenLimitExceeded ModelErrorCode = "TOKEN_LIMIT_EXCEEDED"
	ErrorCodeContentFiltered  ModelErrorCode = "CONTENT_FILTERED"

	// Streaming errors
	ErrorCodeStreamError      ModelErrorCode = "STREAM_ERROR"
	ErrorCodeStreamInterrupted ModelErrorCode = "STREAM_INTERRUPTED"

	// Generic errors
	ErrorCodeInternalError    ModelErrorCode = "INTERNAL_ERROR"
	ErrorCodeUnknownError     ModelErrorCode = "UNKNOWN_ERROR"
)

// Error implements the error interface
func (e *ModelError) Error() string {
	if e.Provider != "" && e.Model != "" {
		return fmt.Sprintf("[%s:%s] %s: %s", e.Provider, e.Model, e.Code, e.Message)
	} else if e.Provider != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *ModelError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns whether the error is retryable
func (e *ModelError) IsRetryable() bool {
	return e.Retryable
}

// WithCause adds a cause to the error
func (e *ModelError) WithCause(cause error) *ModelError {
	e.Cause = cause
	return e
}

// WithDetails adds details to the error
func (e *ModelError) WithDetails(details interface{}) *ModelError {
	e.Details = details
	return e
}

// NewModelError creates a new model error
func NewModelError(code ModelErrorCode, message string, provider, model string) *ModelError {
	return &ModelError{
		Code:       code,
		Message:    message,
		Provider:   provider,
		Model:      model,
		Retryable:  isRetryableErrorCode(code),
	}
}

// NewModelErrorWithStatus creates a new model error with HTTP status code
func NewModelErrorWithStatus(code ModelErrorCode, message string, provider, model string, statusCode int) *ModelError {
	return &ModelError{
		Code:       code,
		Message:    message,
		Provider:   provider,
		Model:      model,
		StatusCode: statusCode,
		Retryable:  isRetryableErrorCode(code) && isRetryableStatus(statusCode),
	}
}

// WrapError wraps an existing error as a model error
func WrapError(err error, code ModelErrorCode, message string, provider, model string) *ModelError {
	return &ModelError{
		Code:      code,
		Message:   message,
		Provider:  provider,
		Model:     model,
		Retryable: isRetryableErrorCode(code),
		Cause:     err,
	}
}

// isRetryableErrorCode checks if an error code is retryable
func isRetryableErrorCode(code ModelErrorCode) bool {
	switch code {
	case ErrorCodeNetworkError,
		 ErrorCodeTimeoutError,
		 ErrorCodeConnectionError,
		 ErrorCodeRateLimitError,
		 ErrorCodeServerError,
		 ErrorCodeServiceUnavailable,
		 ErrorCodeStreamError,
		 ErrorCodeStreamInterrupted:
		return true
	default:
		return false
	}
}

// isRetryableStatus checks if an HTTP status code is retryable
func isRetryableStatus(statusCode int) bool {
	// Don't retry for client errors (4xx) except for specific cases
	if statusCode >= 400 && statusCode < 500 {
		switch statusCode {
		case 408, // Request Timeout
			 429, // Too Many Requests
			 449, // Retry With (proprietary)
			 509: // Bandwidth Limit Exceeded
			return true
		default:
			return false
		}
	}
	
	// Retry for server errors (5xx)
	return statusCode >= 500
}

// ErrorFromHTTPStatus creates a model error from HTTP status code
func ErrorFromHTTPStatus(statusCode int, provider, model string) *ModelError {
	var code ModelErrorCode
	var message string
	var retryable bool

	switch statusCode {
	case http.StatusBadRequest:
		code = ErrorCodeInvalidRequest
		message = "Invalid request format or parameters"
		retryable = false
	case http.StatusUnauthorized:
		code = ErrorCodeUnauthorized
		message = "Invalid or missing API key"
		retryable = false
	case http.StatusForbidden:
		code = ErrorCodeForbidden
		message = "Access forbidden - insufficient permissions"
		retryable = false
	case http.StatusNotFound:
		code = ErrorCodeNotFound
		message = "Requested resource not found"
		retryable = false
	case http.StatusTooManyRequests:
		code = ErrorCodeRateLimitError
		message = "Rate limit exceeded"
		retryable = true
	case http.StatusInternalServerError:
		code = ErrorCodeServerError
		message = "Internal server error"
		retryable = true
	case http.StatusBadGateway:
		code = ErrorCodeServerError
		message = "Bad gateway"
		retryable = true
	case http.StatusServiceUnavailable:
		code = ErrorCodeServiceUnavailable
		message = "Service temporarily unavailable"
		retryable = true
	case http.StatusGatewayTimeout:
		code = ErrorCodeTimeoutError
		message = "Gateway timeout"
		retryable = true
	default:
		if statusCode >= 400 && statusCode < 500 {
			code = ErrorCodeInvalidRequest
			message = fmt.Sprintf("Client error: %d", statusCode)
			retryable = false
		} else if statusCode >= 500 {
			code = ErrorCodeServerError
			message = fmt.Sprintf("Server error: %d", statusCode)
			retryable = true
		} else {
			code = ErrorCodeUnknownError
			message = fmt.Sprintf("Unexpected status code: %d", statusCode)
			retryable = false
		}
	}

	return &ModelError{
		Code:       code,
		Message:    message,
		Provider:   provider,
		Model:      model,
		StatusCode: statusCode,
		Retryable:  retryable,
	}
}

// IsModelError checks if an error is a ModelError
func IsModelError(err error) (*ModelError, bool) {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr, true
	}
	return nil, false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ModelErrorCode {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr.Code
	}
	return ErrorCodeUnknownError
}

// IsRetryableModelError checks if an error is retryable
func IsRetryableModelError(err error) bool {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr.Retryable
	}
	return false
}

// Common error constructors

// NewConfigError creates a configuration error
func NewConfigError(message, provider, model string) *ModelError {
	return NewModelError(ErrorCodeInvalidConfig, message, provider, model)
}

// NewAPIKeyError creates an API key error
func NewAPIKeyError(message, provider, model string) *ModelError {
	return NewModelError(ErrorCodeInvalidAPIKey, message, provider, model)
}

// NewNetworkError creates a network error
func NewNetworkError(message, provider, model string) *ModelError {
	return NewModelError(ErrorCodeNetworkError, message, provider, model)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message, provider, model string) *ModelError {
	return NewModelError(ErrorCodeTimeoutError, message, provider, model)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message, provider, model string) *ModelError {
	return NewModelError(ErrorCodeRateLimitError, message, provider, model)
}

// NewTokenLimitError creates a token limit error
func NewTokenLimitError(message, provider, model string) *ModelError {
	return NewModelError(ErrorCodeTokenLimitExceeded, message, provider, model)
}

// NewContentFilteredError creates a content filtered error
func NewContentFilteredError(message, provider, model string) *ModelError {
	return NewModelError(ErrorCodeContentFiltered, message, provider, model)
}

// NewInternalError creates an internal error
func NewInternalError(message, provider, model string) *ModelError {
	return NewModelError(ErrorCodeInternalError, message, provider, model)
}