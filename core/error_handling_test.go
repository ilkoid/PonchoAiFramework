package core

import (
	"context"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFrameworkErrorHandling tests basic error scenarios
func TestFrameworkErrorHandling(t *testing.T) {
	logger := interfaces.NewNoOpLogger()
	config := &interfaces.PonchoFrameworkConfig{}
	framework := NewPonchoFramework(config, logger)

	ctx := context.Background()

	// Test operations on stopped framework
	t.Run("Operations on stopped framework", func(t *testing.T) {
		// Test generation on stopped framework
		req := &interfaces.PonchoModelRequest{
			Model: "test-model",
			Messages: []*interfaces.PonchoMessage{
				{
					Role: interfaces.PonchoRoleUser,
					Content: []*interfaces.PonchoContentPart{
						{Type: interfaces.PonchoContentTypeText, Text: "Hello"},
					},
				},
			},
		}

		resp, err := framework.Generate(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "framework is not started")

		// Test tool execution on stopped framework
		result, err := framework.ExecuteTool(ctx, "test-tool", map[string]interface{}{"key": "value"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "framework is not started")
	})

	// Start framework for further tests
	err := framework.Start(ctx)
	require.NoError(t, err)

	// Test operations with non-existent components
	t.Run("Operations with non-existent components", func(t *testing.T) {
		req := &interfaces.PonchoModelRequest{
			Model: "non-existent-model",
			Messages: []*interfaces.PonchoMessage{
				{
					Role: interfaces.PonchoRoleUser,
					Content: []*interfaces.PonchoContentPart{
						{Type: interfaces.PonchoContentTypeText, Text: "Hello"},
					},
				},
			},
		}

		// Test generation with non-existent model
		resp, err := framework.Generate(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "not found")

		// Test tool execution with non-existent tool
		result, err := framework.ExecuteTool(ctx, "non-existent-tool", map[string]interface{}{"key": "value"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})

	// Test framework lifecycle errors
	t.Run("Framework lifecycle errors", func(t *testing.T) {
		// Test double start
		err := framework.Start(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already started")

		// Test stop of not started framework (create new one)
		newFramework := NewPonchoFramework(config, logger)
		err = newFramework.Stop(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not started")
	})
}