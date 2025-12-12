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
	items   map[string]*cacheItem
	order   *list.List
	mutex   sync.RWMutex
	logger  interfaces.Logger
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
	return &PromptCacheImpl{
		maxSize: maxSize,
		items:   make(map[string]*cacheItem),
		order:   list.New(),
		logger:  logger,
	}
}

// GetTemplate gets a cached template
func (pc *PromptCacheImpl) GetTemplate(name string) (*interfaces.PromptTemplate, bool) {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	item, exists := pc.items[name]
	if !exists {
		pc.logger.Debug("Cache miss", "name", name)
		return nil, false
	}

	// Update access time and move to front
	item.accessAt = time.Now()
	pc.order.MoveToFront(item.element)

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

	return &interfaces.CacheStats{
		Hits:    0, // TODO: Implement hit/miss tracking
		Misses:  0,
		Size:    int64(len(pc.items)),
		MaxSize: int64(pc.maxSize),
		HitRate: 0.0,
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