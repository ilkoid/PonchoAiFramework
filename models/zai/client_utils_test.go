package zai

import (
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestZAIClient_GetLogger(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	client := &ZAIClient{
		logger: logger,
	}

	retrievedLogger := client.GetLogger()
	assert.NotNil(t, retrievedLogger)
	assert.Equal(t, logger, retrievedLogger)
}

func TestZAIError_Creation(t *testing.T) {
	errorDetail := ZAIErrorDetail{
		Message: "Bad Request",
		Type:    "invalid_request",
		Code:    "400",
	}

	err := &ZAIError{
		Error: errorDetail,
	}

	assert.Equal(t, "Bad Request", err.Error.Message)
	assert.Equal(t, "invalid_request", err.Error.Type)
	assert.Equal(t, "400", err.Error.Code)
}

func TestVisionConfig_DefaultValues(t *testing.T) {
	config := VisionConfig{
		MaxImageSize:   5 * 1024 * 1024,
		DefaultQuality: "high",
		DefaultDetail:  "high",
		Timeout:        30 * time.Second,
	}

	assert.Equal(t, 5*1024*1024, config.MaxImageSize)
	assert.Equal(t, "high", config.DefaultQuality)
	assert.Equal(t, "high", config.DefaultDetail)
	assert.Equal(t, 30*time.Second, config.Timeout)
}

func TestVisionConfig_Timeout(t *testing.T) {
	config := VisionConfig{
		Timeout: 45 * time.Second,
	}

	assert.Equal(t, 45*time.Second, config.Timeout)
}