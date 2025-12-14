package articleflow

import (
	"context"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/tools/wildberries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWBClient is a mock implementation of wildberries.Client
type MockWBClient struct {
	mock.Mock
}

func (m *MockWBClient) GetParentCategories(ctx context.Context) ([]wildberries.WBParentCategory, error) {
	args := m.Called(ctx)
	return args.Get(0).([]wildberries.WBParentCategory), args.Error(1)
}

func (m *MockWBClient) GetSubjects(ctx context.Context) ([]wildberries.WBSubject, error) {
	args := m.Called(ctx)
	return args.Get(0).([]wildberries.WBSubject), args.Error(1)
}

func (m *MockWBClient) GetSubjectCharacteristics(ctx context.Context, subjectID int) ([]wildberries.WBCharacteristic, error) {
	args := m.Called(ctx, subjectID)
	return args.Get(0).([]wildberries.WBCharacteristic), args.Error(1)
}

func TestWBMemoryCache(t *testing.T) {
	t.Run("GetParents - Cache Miss", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(MockWBClient)
		logger := interfaces.NewDefaultLogger()
		config := &CachingConfig{
			WBParentsTTL: time.Hour,
		}

		parents := []wildberries.WBParentCategory{
			{ID: 1, Name: "Clothing"},
			{ID: 2, Name: "Shoes"},
		}

		mockClient.On("GetParentCategories", ctx).Return(parents, nil)

		cache := NewWBMemoryCache(mockClient, logger, config)

		result, err := cache.GetParents(ctx)
		require.NoError(t, err)
		assert.Equal(t, parents, result)

		mockClient.AssertExpectations(t)

		// Verify it's cached
		result2, err := cache.GetParents(ctx)
		require.NoError(t, err)
		assert.Equal(t, parents, result2)

		// Should not call client again
		mockClient.AssertExpectations(t)
	})

	t.Run("GetParents - Cache Expired", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(MockWBClient)
		logger := interfaces.NewDefaultLogger()
		config := &CachingConfig{
			WBParentsTTL: 10 * time.Millisecond, // Very short TTL
		}

		parents1 := []wildberries.WBParentCategory{
			{ID: 1, Name: "Clothing"},
		}
		parents2 := []wildberries.WBParentCategory{
			{ID: 2, Name: "Shoes"},
		}

		// First call
		mockClient.On("GetParentCategories", ctx).Return(parents1, nil).Once()

		cache := NewWBMemoryCache(mockClient, logger, config)

		// First call should hit the API
		result1, err := cache.GetParents(ctx)
		require.NoError(t, err)
		assert.Equal(t, parents1, result1)

		// Wait for cache to expire
		time.Sleep(20 * time.Millisecond)

		// Second call should hit the API again
		mockClient.On("GetParentCategories", ctx).Return(parents2, nil).Once()

		result2, err := cache.GetParents(ctx)
		require.NoError(t, err)
		assert.Equal(t, parents2, result2)

		mockClient.AssertExpectations(t)
	})

	t.Run("GetSubjects - Cache Hit and Miss", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(MockWBClient)
		logger := interfaces.NewDefaultLogger()
		config := &CachingConfig{
			WBSubjectsTTL: time.Hour,
		}

		subjects := []wildberries.WBSubject{
			{ID: 123, Name: "Dresses"},
			{ID: 456, Name: "Shirts"},
		}

		mockClient.On("GetSubjects", ctx).Return(subjects, nil).Once()

		cache := NewWBMemoryCache(mockClient, logger, config)

		// First call - cache miss
		result, err := cache.GetSubjects(ctx)
		require.NoError(t, err)
		assert.Equal(t, subjects, result)

		// Second call - cache hit
		result2, err := cache.GetSubjects(ctx)
		require.NoError(t, err)
		assert.Equal(t, subjects, result2)

		mockClient.AssertExpectations(t)
	})

	t.Run("GetCharacteristics", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(MockWBClient)
		logger := interfaces.NewDefaultLogger()
		config := &CachingConfig{
			WBCharsTTL: time.Hour,
		}

		characteristics := []wildberries.WBCharacteristic{
			{ID: 1, Name: "Color", Type: "string"},
			{ID: 2, Name: "Size", Type: "string"},
		}

		mockClient.On("GetSubjectCharacteristics", ctx, 123).Return(characteristics, nil).Once()

		cache := NewWBMemoryCache(mockClient, logger, config)

		// First call - cache miss
		result, err := cache.GetCharacteristics(ctx, 123)
		require.NoError(t, err)
		assert.Equal(t, characteristics, result)

		// Second call - cache hit
		result2, err := cache.GetCharacteristics(ctx, 123)
		require.NoError(t, err)
		assert.Equal(t, characteristics, result2)

		mockClient.AssertExpectations(t)
	})

	t.Run("Invalidate", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(MockWBClient)
		logger := interfaces.NewDefaultLogger()
		config := DefaultCachingConfig()

		parents := []wildberries.WBParentCategory{{ID: 1, Name: "Clothing"}}
		subjects := []wildberries.WBSubject{{ID: 123, Name: "Dresses"}}
		characteristics := []wildberries.WBCharacteristic{{ID: 1, Name: "Color", Type: "string"}}

		mockClient.On("GetParentCategories", ctx).Return(parents, nil).Once()
		mockClient.On("GetSubjects", ctx).Return(subjects, nil).Once()
		mockClient.On("GetSubjectCharacteristics", ctx, 123).Return(characteristics, nil).Once()

		cache := NewWBMemoryCache(mockClient, logger, config)

		// Populate cache
		_, err := cache.GetParents(ctx)
		require.NoError(t, err)
		_, err = cache.GetSubjects(ctx)
		require.NoError(t, err)
		_, err = cache.GetCharacteristics(ctx, 123)
		require.NoError(t, err)

		// Invalidate all
		err = cache.Invalidate(ctx)
		require.NoError(t, err)

		// Second calls should hit API again
		mockClient.On("GetParentCategories", ctx).Return(parents, nil).Once()
		mockClient.On("GetSubjects", ctx).Return(subjects, nil).Once()
		mockClient.On("GetSubjectCharacteristics", ctx, 123).Return(characteristics, nil).Once()

		_, err = cache.GetParents(ctx)
		require.NoError(t, err)
		_, err = cache.GetSubjects(ctx)
		require.NoError(t, err)
		_, err = cache.GetCharacteristics(ctx, 123)
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("InvalidateSubject", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(MockWBClient)
		logger := interfaces.NewDefaultLogger()
		config := DefaultCachingConfig()

		characteristics := []wildberries.WBCharacteristic{{ID: 1, Name: "Color", Type: "string"}}

		mockClient.On("GetSubjectCharacteristics", ctx, 123).Return(characteristics, nil).Once()

		cache := NewWBMemoryCache(mockClient, logger, config)

		// Populate cache
		_, err := cache.GetCharacteristics(ctx, 123)
		require.NoError(t, err)

		// Invalidate specific subject
		err = cache.InvalidateSubject(ctx, 123)
		require.NoError(t, err)

		// Second call should hit API again
		mockClient.On("GetSubjectCharacteristics", ctx, 123).Return(characteristics, nil).Once()

		_, err = cache.GetCharacteristics(ctx, 123)
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("CleanupExpired", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(MockWBClient)
		logger := interfaces.NewDefaultLogger()
		config := &CachingConfig{
			WBCharsTTL: 10 * time.Millisecond, // Very short TTL
		}

		characteristics := []wildberries.WBCharacteristic{{ID: 1, Name: "Color", Type: "string"}}

		mockClient.On("GetSubjectCharacteristics", ctx, 123).Return(characteristics, nil).Once()

		cache := NewWBMemoryCache(mockClient, logger, config)

		// Populate cache
		_, err := cache.GetCharacteristics(ctx, 123)
		require.NoError(t, err)

		// Verify item is cached
		stats := cache.GetCacheStats(ctx)
		assert.Equal(t, 1, stats["characteristics_count"])

		// Wait for expiration
		time.Sleep(20 * time.Millisecond)

		// Run cleanup
		cache.CleanupExpired(ctx)

		// Verify item was cleaned up
		stats = cache.GetCacheStats(ctx)
		assert.Equal(t, 0, stats["characteristics_count"])

		mockClient.AssertExpectations(t)
	})

	t.Run("MaxCacheSize", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(MockWBClient)
		logger := interfaces.NewDefaultLogger()
		config := &CachingConfig{
			WBCharsTTL:   time.Hour,
			MaxCacheSize: 2, // Very small cache
		}

		characteristics := []wildberries.WBCharacteristic{{ID: 1, Name: "Color", Type: "string"}}

		// Setup expectations for multiple calls
		mockClient.On("GetSubjectCharacteristics", ctx, 123).Return(characteristics, nil).Once()
		mockClient.On("GetSubjectCharacteristics", ctx, 456).Return(characteristics, nil).Once()
		mockClient.On("GetSubjectCharacteristics", ctx, 789).Return(characteristics, nil).Once()

		cache := NewWBMemoryCache(mockClient, logger, config)

		// Fill cache to capacity
		_, err := cache.GetCharacteristics(ctx, 123)
		require.NoError(t, err)
		_, err = cache.GetCharacteristics(ctx, 456)
		require.NoError(t, err)

		// Add one more - should evict oldest
		_, err = cache.GetCharacteristics(ctx, 789)
		require.NoError(t, err)

		stats := cache.GetCacheStats(ctx)
		// Should still be at max capacity
		assert.Equal(t, 2, stats["characteristics_count"])

		mockClient.AssertExpectations(t)
	})

	t.Run("DefaultCachingConfig", func(t *testing.T) {
		config := DefaultCachingConfig()

		assert.Equal(t, 24*time.Hour, config.WBParentsTTL)
		assert.Equal(t, 12*time.Hour, config.WBSubjectsTTL)
		assert.Equal(t, 6*time.Hour, config.WBCharsTTL)
		assert.Equal(t, 1000, config.MaxCacheSize)
	})

	t.Run("GetCacheStats", func(t *testing.T) {
		ctx := context.Background()
		mockClient := new(MockWBClient)
		logger := interfaces.NewDefaultLogger()
		config := DefaultCachingConfig()

		parents := []wildberries.WBParentCategory{{ID: 1, Name: "Clothing"}}
		subjects := []wildberries.WBSubject{{ID: 123, Name: "Dresses"}}

		mockClient.On("GetParentCategories", ctx).Return(parents, nil).Once()
		mockClient.On("GetSubjects", ctx).Return(subjects, nil).Once()

		cache := NewWBMemoryCache(mockClient, logger, config)

		// Populate cache
		_, err := cache.GetParents(ctx)
		require.NoError(t, err)
		_, err = cache.GetSubjects(ctx)
		require.NoError(t, err)

		stats := cache.GetCacheStats(ctx)

		assert.True(t, stats["parents_cached"])
		assert.True(t, stats["subjects_cached"])
		assert.Equal(t, 0, stats["characteristics_count"])
		assert.NotNil(t, stats["parents_expires_at"])
		assert.NotNil(t, stats["subjects_expires_at"])

		mockClient.AssertExpectations(t)
	})
}