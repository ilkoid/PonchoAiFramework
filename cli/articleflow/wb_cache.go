package articleflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/tools/wildberries"
)

// WBCache defines the interface for Wildberries data caching
type WBCache interface {
	GetParents(ctx context.Context) ([]wildberries.ParentCategory, error)
	GetSubjects(ctx context.Context) ([]wildberries.Subject, error)
	GetCharacteristics(ctx context.Context, subjectID int) ([]wildberries.SubjectCharacteristic, error)
	Invalidate(ctx context.Context) error
	InvalidateSubject(ctx context.Context, subjectID int) error
}

// WBClientInterface defines the interface for Wildberries client operations
type WBClientInterface interface {
	GetParentCategories(ctx context.Context) ([]wildberries.ParentCategory, error)
	GetSubjects(ctx context.Context, opts *wildberries.GetSubjectsOptions) ([]wildberries.Subject, error)
	GetSubjectCharacteristics(ctx context.Context, subjectID int) ([]wildberries.SubjectCharacteristic, error)
}

// WBMemoryCache provides an in-memory cache for Wildberries data
type WBMemoryCache struct {
	client       *wildberries.WBClient
	logger       interfaces.Logger
	config       *CachingConfig

	// Cache storage
	parents         *cachedItem[[]wildberries.ParentCategory]
	subjects        *cachedItem[[]wildberries.Subject]
	characteristics map[int]*cachedItem[[]wildberries.SubjectCharacteristic]

	// Synchronization
	mu sync.RWMutex
}

// cachedItem represents a cached item with TTL
type cachedItem[T any] struct {
	data      T
	expiresAt time.Time
}

// isValid checks if the cached item is still valid
func (c *cachedItem[T]) isValid() bool {
	return time.Now().Before(c.expiresAt)
}

// NewWBMemoryCache creates a new in-memory Wildberries cache
func NewWBMemoryCache(
	client *wildberries.WBClient,
	logger interfaces.Logger,
	config *CachingConfig,
) *WBMemoryCache {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	if config == nil {
		config = DefaultCachingConfig()
	}

	return &WBMemoryCache{
		client:         client,
		logger:         logger,
		config:         config,
		characteristics: make(map[int]*cachedItem[[]wildberries.SubjectCharacteristic]),
	}
}

// DefaultCachingConfig returns default caching configuration
func DefaultCachingConfig() *CachingConfig {
	return &CachingConfig{
		WBParentsTTL:    24 * time.Hour,  // Parents rarely change
		WBSubjectsTTL:   12 * time.Hour,  // Subjects occasionally change
		WBCharsTTL:      6 * time.Hour,   // Characteristics might change more often
		MaxCacheSize:    1000,            // Max number of subject characteristics to cache
	}
}

// GetParents returns parent categories, using cache if available
func (c *WBMemoryCache) GetParents(ctx context.Context) ([]wildberries.ParentCategory, error) {
	c.mu.RLock()
	if c.parents != nil && c.parents.isValid() {
		defer c.mu.RUnlock()
		c.logger.Debug("Cache hit: WB parents")
		return c.parents.data, nil
	}
	c.mu.RUnlock()

	// Cache miss - fetch from API
	c.logger.Info("Cache miss: fetching WB parents from API")
	parents, err := c.client.GetParentCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch parent categories: %w", err)
	}

	// Update cache
	c.mu.Lock()
	c.parents = &cachedItem[[]wildberries.ParentCategory]{
		data:      parents,
		expiresAt: time.Now().Add(c.config.WBParentsTTL),
	}
	c.mu.Unlock()

	c.logger.Info("Cached WB parents",
		"count", len(parents),
		"ttl", c.config.WBParentsTTL,
	)

	return parents, nil
}

// GetSubjects returns subjects, using cache if available
func (c *WBMemoryCache) GetSubjects(ctx context.Context) ([]wildberries.Subject, error) {
	c.mu.RLock()
	if c.subjects != nil && c.subjects.isValid() {
		defer c.mu.RUnlock()
		c.logger.Debug("Cache hit: WB subjects")
		return c.subjects.data, nil
	}
	c.mu.RUnlock()

	// Cache miss - fetch from API
	c.logger.Info("Cache miss: fetching WB subjects from API")
	subjects, err := c.client.GetSubjects(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subjects: %w", err)
	}

	// Update cache
	c.mu.Lock()
	c.subjects = &cachedItem[[]wildberries.Subject]{
		data:      subjects,
		expiresAt: time.Now().Add(c.config.WBSubjectsTTL),
	}
	c.mu.Unlock()

	c.logger.Info("Cached WB subjects",
		"count", len(subjects),
		"ttl", c.config.WBSubjectsTTL,
	)

	return subjects, nil
}

// GetCharacteristics returns characteristics for a subject, using cache if available
func (c *WBMemoryCache) GetCharacteristics(ctx context.Context, subjectID int) ([]wildberries.SubjectCharacteristic, error) {
	// Check cache first
	c.mu.RLock()
	if cached, exists := c.characteristics[subjectID]; exists && cached.isValid() {
		defer c.mu.RUnlock()
		c.logger.Debug("Cache hit: WB characteristics",
			"subject_id", subjectID,
		)
		return cached.data, nil
	}
	c.mu.RUnlock()

	// Cache miss - fetch from API
	c.logger.Info("Cache miss: fetching WB characteristics from API",
		"subject_id", subjectID,
	)

	characteristics, err := c.client.GetSubjectCharacteristics(ctx, subjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subject characteristics: %w", err)
	}

	// Check cache size limit
	c.mu.Lock()
	if len(c.characteristics) >= c.config.MaxCacheSize {
		// Remove oldest entries (simple LRU-like behavior)
		c.evictOldestCharacteristics()
	}

	// Update cache
	c.characteristics[subjectID] = &cachedItem[[]wildberries.SubjectCharacteristic]{
		data:      characteristics,
		expiresAt: time.Now().Add(c.config.WBCharsTTL),
	}
	c.mu.Unlock()

	c.logger.Info("Cached WB characteristics",
		"subject_id", subjectID,
		"count", len(characteristics),
		"ttl", c.config.WBCharsTTL,
	)

	return characteristics, nil
}

// Invalidate clears all cached data
func (c *WBMemoryCache) Invalidate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.parents = nil
	c.subjects = nil
	c.characteristics = make(map[int]*cachedItem[[]wildberries.SubjectCharacteristic])

	c.logger.Info("Invalidated all WB cache entries")
	return nil
}

// InvalidateSubject clears cached data for a specific subject
func (c *WBMemoryCache) InvalidateSubject(ctx context.Context, subjectID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.characteristics, subjectID)

	c.logger.Info("Invalidated WB cache for subject",
		"subject_id", subjectID,
	)
	return nil
}

// evictOldestCharacteristics removes the oldest cached characteristics
func (c *WBMemoryCache) evictOldestCharacteristics() {
	// Simple strategy: find and remove the entry with earliest expiration
	var oldestSubjectID int
	var oldestTime time.Time
	first := true

	for subjectID, cached := range c.characteristics {
		if first || cached.expiresAt.Before(oldestTime) {
			oldestSubjectID = subjectID
			oldestTime = cached.expiresAt
			first = false
		}
	}

	if !first {
		delete(c.characteristics, oldestSubjectID)
		c.logger.Debug("Evicted oldest WB characteristics from cache",
			"subject_id", oldestSubjectID,
		)
	}
}

// GetCacheStats returns cache statistics for monitoring
func (c *WBMemoryCache) GetCacheStats(ctx context.Context) map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]interface{}{
		"parents_cached":     c.parents != nil && c.parents.isValid(),
		"subjects_cached":    c.subjects != nil && c.subjects.isValid(),
		"characteristics_count": len(c.characteristics),
	}

	// Add expiration times if cached
	if c.parents != nil {
		stats["parents_expires_at"] = c.parents.expiresAt
	}
	if c.subjects != nil {
		stats["subjects_expires_at"] = c.subjects.expiresAt
	}

	return stats
}

// CleanupExpired removes expired entries from the cache
func (c *WBMemoryCache) CleanupExpired(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check parents
	if c.parents != nil && !c.parents.isValid() {
		c.parents = nil
		c.logger.Debug("Cleaned up expired parents cache")
	}

	// Check subjects
	if c.subjects != nil && !c.subjects.isValid() {
		c.subjects = nil
		c.logger.Debug("Cleaned up expired subjects cache")
	}

	// Check characteristics
	var expiredSubjects []int
	for subjectID, cached := range c.characteristics {
		if !cached.isValid() {
			expiredSubjects = append(expiredSubjects, subjectID)
		}
	}

	for _, subjectID := range expiredSubjects {
		delete(c.characteristics, subjectID)
	}

	if len(expiredSubjects) > 0 {
		c.logger.Debug("Cleaned up expired characteristics cache",
			"count", len(expiredSubjects),
		)
	}
}

// StartCleanupWorker starts a background worker that periodically cleans up expired entries
func (c *WBMemoryCache) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				c.logger.Info("WB cache cleanup worker stopped")
				return
			case <-ticker.C:
				c.CleanupExpired(ctx)
			}
		}
	}()

	c.logger.Info("WB cache cleanup worker started",
		"interval", interval,
	)
}