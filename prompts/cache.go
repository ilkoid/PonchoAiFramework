// Package prompts provides LRU caching system for prompt templates
//
// Key functionality:
// • Thread-safe LRU cache with configurable size limits
// • Template caching with access time tracking and eviction
// • Cache statistics and performance monitoring
// • Template invalidation and cache clearing operations
// • Integration with prompt manager for transparent caching
// • Memory-efficient storage with automatic cleanup
//
// Key relationships:
// • Implements PromptCache interface from core interfaces
// • Integrates with manager package for template lifecycle management
// • Provides caching layer between parser and template storage
// • Supports cache statistics collection for performance monitoring
// • Used by executor package for cached template access
// • Designed for high-throughput template access scenarios
//
// Design patterns:
// • Cache pattern with LRU eviction strategy
// • Proxy pattern for transparent template access
// • Observer pattern for cache statistics collection
// • Singleton pattern for cache instance management
// • Template method pattern for cache operations workflow

package prompts

import (
	"container/list"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// PromptCacheImpl implements a simple LRU cache for prompt templates
type PromptCacheImpl struct {
	maxSize int
	ttl     time.Duration
	items   map[string]*cacheItem
	order   *list.List
	mutex   sync.RWMutex
	logger  interfaces.Logger
	hits    int64
	misses  int64
}

// cacheItem represents an item in the cache
type cacheItem struct {
	key       string
	template  *interfaces.PromptTemplate
	createdAt time.Time
	accessAt  time.Time
	element   *list.Element
}

// NewPromptCache creates a new prompt cache
func NewPromptCache(maxSize int, logger interfaces.Logger) interfaces.PromptCache {
	return NewPromptCacheWithTTL(maxSize, 0, logger) // 0 TTL means no expiration
}

// NewPromptCacheWithTTL creates a new prompt cache with TTL
func NewPromptCacheWithTTL(maxSize int, ttl time.Duration, logger interfaces.Logger) interfaces.PromptCache {
	return &PromptCacheImpl{
		maxSize: maxSize,
		ttl:     ttl,
		items:   make(map[string]*cacheItem),
		order:   list.New(),
		logger:  logger,
	}
}

// GetTemplate gets a cached template
func (pc *PromptCacheImpl) GetTemplate(name string) (*interfaces.PromptTemplate, bool) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	item, exists := pc.items[name]
	if !exists {
		pc.misses++
		pc.logger.Debug("Cache miss", "name", name)
		return nil, false
	}

	// Check if item has expired
	if pc.ttl > 0 && time.Since(item.createdAt) > pc.ttl {
		// Remove expired item
		pc.order.Remove(item.element)
		delete(pc.items, name)
		pc.misses++
		pc.logger.Debug("Cache expired", "name", name, "age_seconds", time.Since(item.createdAt).Seconds())
		return nil, false
	}

	// Update access time and move to front
	item.accessAt = time.Now()
	pc.order.MoveToFront(item.element)
	pc.hits++

	pc.logger.Debug("Cache hit", "name", name, "age_seconds", time.Since(item.createdAt).Seconds())
	return item.template, true
}

// SetTemplate sets a template in cache
func (pc *PromptCacheImpl) SetTemplate(name string, template *interfaces.PromptTemplate) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	now := time.Now()

	// Check if item already exists
	if item, exists := pc.items[name]; exists {
		// Update existing item
		item.template = template
		item.accessAt = now
		pc.order.MoveToFront(item.element)
		pc.logger.Debug("Cache updated", "name", name)
		return
	}

	// Create new item
	item := &cacheItem{
		key:       name,
		template:  template,
		createdAt: now,
		accessAt:  now,
	}

	// Add to front of list
	item.element = pc.order.PushFront(name)
	pc.items[name] = item

	// Evict if over capacity
	if len(pc.items) > pc.maxSize {
		pc.evictLRU()
	}

	pc.logger.Debug("Cache set", "name", name, "total_items", len(pc.items))
}

// InvalidateTemplate invalidates a cached template
func (pc *PromptCacheImpl) InvalidateTemplate(name string) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	item, exists := pc.items[name]
	if !exists {
		return
	}

	// Remove from list and map
	pc.order.Remove(item.element)
	delete(pc.items, name)

	pc.logger.Debug("Cache invalidated", "name", name, "total_items", len(pc.items))
}

// Clear removes all items from cache
func (pc *PromptCacheImpl) Clear() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	pc.items = make(map[string]*cacheItem)
	pc.order = list.New()

	pc.logger.Debug("Cache cleared")
}

// Stats returns cache statistics
func (pc *PromptCacheImpl) Stats() *interfaces.CacheStats {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	totalRequests := pc.hits + pc.misses
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(pc.hits) / float64(totalRequests)
	}

	return &interfaces.CacheStats{
		Hits:    pc.hits,
		Misses:  pc.misses,
		Size:    int64(len(pc.items)),
		MaxSize: int64(pc.maxSize),
		HitRate: hitRate,
	}
}

// evictLRU removes the least recently used item
func (pc *PromptCacheImpl) evictLRU() {
	if pc.order.Len() == 0 {
		return
	}

	// Get last element (least recently used)
	element := pc.order.Back()
	if element != nil {
		name := element.Value.(string)
		
		// Remove from map and list
		delete(pc.items, name)
		pc.order.Remove(element)
		
		pc.logger.Debug("Cache evicted LRU", "name", name)
	}
}