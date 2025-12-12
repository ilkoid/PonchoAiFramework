package common

import (
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

func TestProviderConstants(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		expected string
	}{
		{"DeepSeek", ProviderDeepSeek, "deepseek"},
		{"ZAI", ProviderZAI, "zai"},
		{"OpenAI", ProviderOpenAI, "openai"},
		{"Custom", ProviderCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.provider) != tt.expected {
				t.Errorf("Provider %s = %s, want %s", tt.name, tt.provider, tt.expected)
			}
		})
	}
}

func TestModelTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		modelType ModelType
		expected string
	}{
		{"Text", ModelTypeText, "text"},
		{"Vision", ModelTypeVision, "vision"},
		{"Multimodal", ModelTypeMultimodal, "multimodal"},
		{"Embedding", ModelTypeEmbedding, "embedding"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.modelType) != tt.expected {
				t.Errorf("ModelType %s = %s, want %s", tt.name, tt.modelType, tt.expected)
			}
		})
	}
}

func TestToPonchoFinishReason(t *testing.T) {
	tests := []struct {
		name     string
		reason   FinishReason
		expected interfaces.PonchoFinishReason
	}{
		{"Stop", FinishReasonStop, interfaces.PonchoFinishReasonStop},
		{"Length", FinishReasonLength, interfaces.PonchoFinishReasonLength},
		{"Tool", FinishReasonTool, interfaces.PonchoFinishReasonTool},
		{"Error", FinishReasonError, interfaces.PonchoFinishReasonError},
		{"Unknown", FinishReason("unknown"), interfaces.PonchoFinishReasonStop},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToPonchoFinishReason(tt.reason)
			if result != tt.expected {
				t.Errorf("ToPonchoFinishReason(%s) = %s, want %s", tt.reason, result, tt.expected)
			}
		})
	}
}

func TestFromPonchoFinishReason(t *testing.T) {
	tests := []struct {
		name     string
		reason   interfaces.PonchoFinishReason
		expected FinishReason
	}{
		{"Stop", interfaces.PonchoFinishReasonStop, FinishReasonStop},
		{"Length", interfaces.PonchoFinishReasonLength, FinishReasonLength},
		{"Tool", interfaces.PonchoFinishReasonTool, FinishReasonTool},
		{"Error", interfaces.PonchoFinishReasonError, FinishReasonError},
		{"Unknown", interfaces.PonchoFinishReason("unknown"), FinishReasonStop},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromPonchoFinishReason(tt.reason)
			if result != tt.expected {
				t.Errorf("FromPonchoFinishReason(%s) = %s, want %s", tt.reason, result, tt.expected)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		errorCode string
		expected bool
	}{
		{"Rate Limit", ErrCodeRateLimit, true},
		{"Server Error", ErrCodeServerError, true},
		{"Timeout", ErrCodeTimeout, true},
		{"Network Error", ErrCodeNetworkError, true},
		{"Invalid Request", ErrCodeInvalidRequest, false},
		{"Invalid Auth", ErrCodeInvalidAuth, false},
		{"Content Filter", ErrCodeContentFilter, false},
		{"Unknown", "unknown_error", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.errorCode)
			if result != tt.expected {
				t.Errorf("IsRetryableError(%s) = %t, want %t", tt.errorCode, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultConfigForProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		expected CommonModelConfig
	}{
		{
			name:     "DeepSeek",
			provider: ProviderDeepSeek,
			expected: CommonModelConfig{
				Provider:    ProviderDeepSeek,
				Model:       DeepSeekDefaultModel,
				BaseURL:     DeepSeekDefaultBaseURL,
				MaxTokens:   4000,
				Temperature: 0.7,
				Timeout:     30 * time.Second,
			},
		},
		{
			name:     "ZAI",
			provider: ProviderZAI,
			expected: CommonModelConfig{
				Provider:    ProviderZAI,
				Model:       ZAIDefaultModel,
				BaseURL:     ZAIDefaultBaseURL,
				MaxTokens:   2000,
				Temperature: 0.5,
				Timeout:     60 * time.Second,
			},
		},
		{
			name:     "Custom",
			provider: ProviderCustom,
			expected: CommonModelConfig{
				Provider:    ProviderCustom,
				MaxTokens:   4000,
				Temperature: 0.7,
				Timeout:     30 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDefaultConfigForProvider(tt.provider)
			
			if result.Provider != tt.expected.Provider {
				t.Errorf("Provider = %s, want %s", result.Provider, tt.expected.Provider)
			}
			if result.Model != tt.expected.Model {
				t.Errorf("Model = %s, want %s", result.Model, tt.expected.Model)
			}
			if result.BaseURL != tt.expected.BaseURL {
				t.Errorf("BaseURL = %s, want %s", result.BaseURL, tt.expected.BaseURL)
			}
			if result.MaxTokens != tt.expected.MaxTokens {
				t.Errorf("MaxTokens = %d, want %d", result.MaxTokens, tt.expected.MaxTokens)
			}
			if result.Temperature != tt.expected.Temperature {
				t.Errorf("Temperature = %f, want %f", result.Temperature, tt.expected.Temperature)
			}
			if result.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout = %v, want %v", result.Timeout, tt.expected.Timeout)
			}
		})
	}
}

func TestDefaultConfigurations(t *testing.T) {
	// Test HTTP config
	if DefaultHTTPConfig.Timeout != 30*time.Second {
		t.Errorf("DefaultHTTPConfig.Timeout = %v, want %v", DefaultHTTPConfig.Timeout, 30*time.Second)
	}
	if DefaultHTTPConfig.MaxIdleConns != 100 {
		t.Errorf("DefaultHTTPConfig.MaxIdleConns = %d, want %d", DefaultHTTPConfig.MaxIdleConns, 100)
	}

	// Test Retry config
	if DefaultRetryConfig.MaxAttempts != 3 {
		t.Errorf("DefaultRetryConfig.MaxAttempts = %d, want %d", DefaultRetryConfig.MaxAttempts, 3)
	}
	if DefaultRetryConfig.BackoffType != BackoffTypeExponential {
		t.Errorf("DefaultRetryConfig.BackoffType = %s, want %s", DefaultRetryConfig.BackoffType, BackoffTypeExponential)
	}

	// Test Validation rules
	if len(DefaultValidationRules) == 0 {
		t.Error("DefaultValidationRules should not be empty")
	}
}

func TestErrorCodes(t *testing.T) {
	expectedCodes := []string{
		ErrCodeInvalidRequest,
		ErrCodeInvalidAuth,
		ErrCodeRateLimit,
		ErrCodeInsufficientQuota,
		ErrCodeModelNotFound,
		ErrCodeContentFilter,
		ErrCodeServerError,
		ErrCodeTimeout,
		ErrCodeNetworkError,
		ErrCodeParsingError,
		ErrCodeValidationError,
		ErrCodeUnknown,
	}

	for _, code := range expectedCodes {
		if code == "" {
			t.Error("Error code should not be empty")
		}
	}
}

func TestHeaders(t *testing.T) {
	expectedHeaders := []string{
		HeaderContentType,
		HeaderAccept,
		HeaderAuthorization,
		HeaderUserAgent,
		HeaderCacheControl,
		HeaderXRequestID,
		HeaderRetryAfter,
	}

	for _, header := range expectedHeaders {
		if header == "" {
			t.Error("Header should not be empty")
		}
	}
}

func TestMIMETypes(t *testing.T) {
	expectedMIMETypes := []string{
		MIMETypeJSON,
		MIMETypeEventStream,
		MIMETypeText,
		MIMETypeHTML,
	}

	for _, mimeType := range expectedMIMETypes {
		if mimeType == "" {
			t.Error("MIME type should not be empty")
		}
	}
}