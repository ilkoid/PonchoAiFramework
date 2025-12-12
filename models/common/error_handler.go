package common

import (
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ErrorHandler handles error processing for different providers
type ErrorHandler struct {
	provider Provider
	logger   interfaces.Logger
}

// NewErrorHandler creates a new error handler for a provider
func NewErrorHandler(provider Provider, logger interfaces.Logger) *ErrorHandler {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &ErrorHandler{
		provider: provider,
		logger:   logger,
	}
}

// IsRetryableError checks if an error is retryable for this provider
func (eh *ErrorHandler) IsRetryableError(err error) bool {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr.Retryable
	}
	
	return false
}

// GetRetryDelay returns provider-specific retry delay
func (eh *ErrorHandler) GetRetryDelay(err error, attempt int) time.Duration {
	// Default delay based on provider
	switch eh.provider {
	case ProviderDeepSeek:
		return eh.getDeepSeekRetryDelay(err, attempt)
	case ProviderZAI:
		return eh.getZAIRetryDelay(err, attempt)
	default:
		return time.Duration(attempt) * time.Second
	}
}

// getDeepSeekRetryDelay returns retry delay for DeepSeek
func (eh *ErrorHandler) getDeepSeekRetryDelay(err error, attempt int) time.Duration {
	if modelErr, ok := err.(*ModelError); ok {
		switch modelErr.Code {
		case ErrorCodeRateLimitError:
			// Exponential backoff for rate limits
			return time.Duration(1<<uint(attempt)) * time.Second
		case ErrorCodeServerError:
			// Longer backoff for server errors
			return time.Duration(attempt) * 5 * time.Second
		case ErrorCodeTimeoutError:
			// Moderate backoff for timeouts
			return time.Duration(attempt) * 2 * time.Second
		default:
			return time.Second
		}
	}
	
	return time.Second
}

// getZAIRetryDelay returns retry delay for Z.AI
func (eh *ErrorHandler) getZAIRetryDelay(err error, attempt int) time.Duration {
	if modelErr, ok := err.(*ModelError); ok {
		switch modelErr.Code {
		case ErrorCodeRateLimitError:
			// Exponential backoff for rate limits
			return time.Duration(1<<uint(attempt)) * 2 * time.Second
		case ErrorCodeServerError:
			// Longer backoff for server errors
			return time.Duration(attempt) * 3 * time.Second
		case ErrorCodeTimeoutError:
			// Moderate backoff for timeouts
			return time.Duration(attempt) * time.Second
		default:
			return time.Second
		}
	}
	
	return time.Second
}

// WrapError wraps an error with provider-specific context
func (eh *ErrorHandler) WrapError(err error, message string) *ModelError {
	return &ModelError{
		Code:      ErrorCodeInternalError,
		Message:   message,
		Provider:  string(eh.provider),
		Retryable: eh.IsRetryableError(err),
		Cause:     err,
	}
}

// LogError logs an error with provider context
func (eh *ErrorHandler) LogError(err error, context map[string]interface{}) {
	if modelErr, ok := err.(*ModelError); ok {
		eh.logger.Error("Model error occurred",
			"provider", modelErr.Provider,
			"model", modelErr.Model,
			"code", modelErr.Code,
			"message", modelErr.Message,
			"retryable", modelErr.Retryable,
			"context", context)
	} else {
		eh.logger.Error("Error occurred",
			"provider", eh.provider,
			"error", err.Error(),
			"context", context)
	}
}

// ConvertToModelError converts any error to ModelError
func (eh *ErrorHandler) ConvertToModelError(err error, provider, model string) *ModelError {
	if modelErr, ok := err.(*ModelError); ok {
		return modelErr
	}

	// Try to extract status code from HTTP errors
	if statusCode := extractStatusCode(err); statusCode > 0 {
		return ErrorFromHTTPStatus(statusCode, provider, model).WithCause(err)
	}

	// Default error
	return NewModelError(ErrorCodeUnknownError, err.Error(), provider, model).WithCause(err)
}

// extractStatusCode tries to extract HTTP status code from error
func extractStatusCode(err error) int {
	// This is a simplified implementation
	// In practice, you might want to check for specific error types
	// that contain status codes
	return 0
}