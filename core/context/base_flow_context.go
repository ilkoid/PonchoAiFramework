package context

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// BaseFlowContext provides a thread-safe implementation of FlowContext
type BaseFlowContext struct {
	mutex     sync.RWMutex
	data      map[string]interface{}
	media     map[string]*MediaData
	config    *ContextConfig
	logger    interfaces.Logger
	parent    FlowContext
	children  map[string]FlowContext
	createdAt time.Time
	lastAccess time.Time
}

// NewBaseFlowContext creates a new BaseFlowContext with default configuration
func NewBaseFlowContext() *BaseFlowContext {
	return NewBaseFlowContextWithConfig(DefaultContextConfig())
}

// NewBaseFlowContextWithConfig creates a new BaseFlowContext with custom configuration
func NewBaseFlowContextWithConfig(config *ContextConfig) *BaseFlowContext {
	if config == nil {
		config = DefaultContextConfig()
	}

	ctx := &BaseFlowContext{
		data:      make(map[string]interface{}),
		media:     make(map[string]*MediaData),
		config:    config,
		logger:    config.Logger,
		parent:    config.Parent,
		children:  make(map[string]FlowContext),
		createdAt: time.Now(),
		lastAccess: time.Now(),
	}

	// Set context ID in data for easy access
	ctx.data["_context_id"] = config.ID

	return ctx
}

// Basic state management
func (bfc *BaseFlowContext) Set(key string, value interface{}) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	bfc.checkSize()
	bfc.data[key] = value
	bfc.lastAccess = time.Now()

	bfc.logger.Debug("Value set in context",
		"key", key,
		"type", fmt.Sprintf("%T", value),
		"context_id", bfc.config.ID,
	)

	return nil
}

func (bfc *BaseFlowContext) Get(key string) (interface{}, bool) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	value, exists := bfc.data[key]
	bfc.lastAccess = time.Now()

	return value, exists
}

func (bfc *BaseFlowContext) Delete(key string) bool {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	if _, exists := bfc.data[key]; exists {
		delete(bfc.data, key)
		bfc.lastAccess = time.Now()

		bfc.logger.Debug("Value deleted from context",
			"key", key,
			"context_id", bfc.config.ID,
		)

		return true
	}

	return false
}

func (bfc *BaseFlowContext) Has(key string) bool {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	_, exists := bfc.data[key]
	return exists
}

func (bfc *BaseFlowContext) Clear() {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	bfc.data = make(map[string]interface{})
	bfc.media = make(map[string]*MediaData)
	bfc.lastAccess = time.Now()

	bfc.logger.Debug("Context cleared", "context_id", bfc.config.ID)
}

// Type-safe operations
func (bfc *BaseFlowContext) SetString(key, value string) error {
	return bfc.Set(key, value)
}

func (bfc *BaseFlowContext) GetString(key string) (string, error) {
	value, exists := bfc.Get(key)
	if !exists {
		return "", fmt.Errorf("key '%s' not found in context", key)
	}

	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("value at key '%s' is not a string, got %T", key, value)
}

func (bfc *BaseFlowContext) SetBytes(key string, value []byte) error {
	return bfc.Set(key, value)
}

func (bfc *BaseFlowContext) GetBytes(key string) ([]byte, error) {
	value, exists := bfc.Get(key)
	if !exists {
		return nil, fmt.Errorf("key '%s' not found in context", key)
	}

	if bytes, ok := value.([]byte); ok {
		return bytes, nil
	}

	return nil, fmt.Errorf("value at key '%s' is not []byte, got %T", key, value)
}

func (bfc *BaseFlowContext) SetInt(key string, value int) error {
	return bfc.Set(key, value)
}

func (bfc *BaseFlowContext) GetInt(key string) (int, error) {
	value, exists := bfc.Get(key)
	if !exists {
		return 0, fmt.Errorf("key '%s' not found in context", key)
	}

	if i, ok := value.(int); ok {
		return i, nil
	}

	return 0, fmt.Errorf("value at key '%s' is not int, got %T", key, value)
}

func (bfc *BaseFlowContext) SetFloat(key string, value float64) error {
	return bfc.Set(key, value)
}

func (bfc *BaseFlowContext) GetFloat(key string) (float64, error) {
	value, exists := bfc.Get(key)
	if !exists {
		return 0, fmt.Errorf("key '%s' not found in context", key)
	}

	if f, ok := value.(float64); ok {
		return f, nil
	}

	return 0, fmt.Errorf("value at key '%s' is not float64, got %T", key, value)
}

func (bfc *BaseFlowContext) SetBool(key string, value bool) error {
	return bfc.Set(key, value)
}

func (bfc *BaseFlowContext) GetBool(key string) (bool, error) {
	value, exists := bfc.Get(key)
	if !exists {
		return false, fmt.Errorf("key '%s' not found in context", key)
	}

	if b, ok := value.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("value at key '%s' is not bool, got %T", key, value)
}

// Array/List operations
func (bfc *BaseFlowContext) SetArray(key string, values []interface{}) error {
	return bfc.Set(key, values)
}

func (bfc *BaseFlowContext) GetArray(key string) ([]interface{}, error) {
	value, exists := bfc.Get(key)
	if !exists {
		return nil, fmt.Errorf("key '%s' not found in context", key)
	}

	if arr, ok := value.([]interface{}); ok {
		return arr, nil
	}

	return nil, fmt.Errorf("value at key '%s' is not []interface{}, got %T", key, value)
}

func (bfc *BaseFlowContext) AppendToArray(key string, value interface{}) error {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	var arr []interface{}
	if existingValue, exists := bfc.data[key]; exists {
		if existingArr, ok := existingValue.([]interface{}); ok {
			arr = existingArr
		} else {
			return fmt.Errorf("key '%s' exists but is not an array", key)
		}
	} else {
		arr = make([]interface{}, 0)
	}

	arr = append(arr, value)
	bfc.data[key] = arr
	bfc.lastAccess = time.Now()

	return nil
}

func (bfc *BaseFlowContext) GetArraySize(key string) (int, error) {
	arr, err := bfc.GetArray(key)
	if err != nil {
		return 0, err
	}
	return len(arr), nil
}

// Object operations
func (bfc *BaseFlowContext) SetObject(key string, obj interface{}) error {
	return bfc.Set(key, obj)
}

func (bfc *BaseFlowContext) GetObject(key string, target interface{}) error {
	value, exists := bfc.Get(key)
	if !exists {
		return fmt.Errorf("key '%s' not found in context", key)
	}

	// Use JSON marshaling/unmarshaling for deep copy
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value at key '%s': %w", key, err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target: %w", err)
	}

	return nil
}

// Media-specific operations
func (bfc *BaseFlowContext) SetMedia(key string, media *MediaData) error {
	if media == nil {
		return fmt.Errorf("media cannot be nil")
	}

	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	bfc.checkSize()
	bfc.media[key] = media
	bfc.lastAccess = time.Now()

	bfc.logger.Debug("Media set in context",
		"key", key,
		"mime_type", media.MimeType,
		"size", media.Size,
		"context_id", bfc.config.ID,
	)

	return nil
}

func (bfc *BaseFlowContext) GetMedia(key string) (*MediaData, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	media, exists := bfc.media[key]
	bfc.lastAccess = time.Now()

	if !exists {
		return nil, fmt.Errorf("media at key '%s' not found", key)
	}

	return media, nil
}

func (bfc *BaseFlowContext) GetAllMedia(prefix string) ([]*MediaData, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	var result []*MediaData
	for key, media := range bfc.media {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			result = append(result, media)
		}
	}

	bfc.lastAccess = time.Now()
	return result, nil
}

func (bfc *BaseFlowContext) AccumulateMedia(prefix string, mediaList []*MediaData) error {
	if mediaList == nil {
		return nil
	}

	for i, media := range mediaList {
		key := fmt.Sprintf("%s_%d", prefix, i)
		if err := bfc.SetMedia(key, media); err != nil {
			return fmt.Errorf("failed to store media at %s: %w", key, err)
		}
	}

	bfc.logger.Debug("Media accumulated",
		"prefix", prefix,
		"count", len(mediaList),
		"context_id", bfc.config.ID,
	)

	return nil
}

// Metadata and utilities
func (bfc *BaseFlowContext) Keys() []string {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	keys := make([]string, 0, len(bfc.data)+len(bfc.media))
	for k := range bfc.data {
		keys = append(keys, k)
	}
	for k := range bfc.media {
		keys = append(keys, k)
	}

	return keys
}

func (bfc *BaseFlowContext) Size() int {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	return len(bfc.data) + len(bfc.media)
}

func (bfc *BaseFlowContext) Clone() FlowContext {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	cloneData := make(map[string]interface{})
	for k, v := range bfc.data {
		cloneData[k] = v
	}

	cloneMedia := make(map[string]*MediaData)
	for k, v := range bfc.media {
		cloneMedia[k] = v
	}

	cloneConfig := *bfc.config
	cloneConfig.ID = generateID()

	clone := &BaseFlowContext{
		data:      cloneData,
		media:     cloneMedia,
		config:    &cloneConfig,
		logger:    bfc.logger,
		parent:    bfc,
		children:  make(map[string]FlowContext),
		createdAt: time.Now(),
		lastAccess: time.Now(),
	}

	bfc.children[cloneConfig.ID] = clone

	return clone
}

func (bfc *BaseFlowContext) Merge(other FlowContext) error {
	if other == nil {
		return fmt.Errorf("cannot merge with nil context")
	}

	otherKeys := other.Keys()
	for _, key := range otherKeys {
		value, exists := other.Get(key)
		if exists {
			if err := bfc.Set(key, value); err != nil {
				return fmt.Errorf("failed to set key '%s' during merge: %w", key, err)
			}
		}
	}

	return nil
}

// Serialization
func (bfc *BaseFlowContext) Serialize() ([]byte, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	if !bfc.config.EnableSerialization {
		return nil, fmt.Errorf("serialization is disabled")
	}

	data := map[string]interface{}{
		"data":       bfc.data,
		"config":     bfc.config,
		"created_at": bfc.createdAt,
		"accessed_at": bfc.lastAccess,
	}

	// Don't serialize media bytes directly, only metadata
	mediaMetadata := make(map[string]*MediaData)
	for k, v := range bfc.media {
		mediaCopy := *v
		mediaCopy.Bytes = nil // Don't serialize bytes
		mediaMetadata[k] = &mediaCopy
	}
	data["media"] = mediaMetadata

	return json.Marshal(data)
}

func (bfc *BaseFlowContext) Deserialize(data []byte) error {
	var serialized map[string]interface{}
	if err := json.Unmarshal(data, &serialized); err != nil {
		return fmt.Errorf("failed to deserialize data: %w", err)
	}

	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	if dataMap, ok := serialized["data"].(map[string]interface{}); ok {
		bfc.data = dataMap
	}

	// Media metadata only
	if mediaMap, ok := serialized["media"].(map[string]interface{}); ok {
		for k, v := range mediaMap {
			if mediaBytes, err := json.Marshal(v); err == nil {
				var media MediaData
				if err := json.Unmarshal(mediaBytes, &media); err == nil {
					bfc.media[k] = &media
				}
			}
		}
	}

	return nil
}

func (bfc *BaseFlowContext) ToJSON() (string, error) {
	data, err := bfc.Serialize()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Context lifecycle
func (bfc *BaseFlowContext) ID() string {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return bfc.config.ID
}

func (bfc *BaseFlowContext) CreatedAt() time.Time {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return bfc.createdAt
}

func (bfc *BaseFlowContext) Parent() FlowContext {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return bfc.parent
}

func (bfc *BaseFlowContext) CreateChild() FlowContext {
	childConfig := *bfc.config
	childConfig.ID = generateID()
	childConfig.Parent = bfc

	child := NewBaseFlowContextWithConfig(&childConfig)

	bfc.mutex.Lock()
	bfc.children[childConfig.ID] = child
	bfc.mutex.Unlock()

	bfc.logger.Debug("Child context created",
		"parent_id", bfc.config.ID,
		"child_id", childConfig.ID,
	)

	return child
}

// Logging and debugging
func (bfc *BaseFlowContext) SetLogger(logger interfaces.Logger) {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()
	bfc.logger = logger
}

func (bfc *BaseFlowContext) GetLogger() interfaces.Logger {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return bfc.logger
}

func (bfc *BaseFlowContext) Dump() map[string]interface{} {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	dump := make(map[string]interface{})
	for k, v := range bfc.data {
		dump[k] = v
	}

	// Add media metadata
	for k, v := range bfc.media {
		mediaCopy := map[string]interface{}{
			"url":       v.URL,
			"mime_type": v.MimeType,
			"size":      v.Size,
			"metadata":  v.Metadata,
		}
		dump["media_"+k] = mediaCopy
	}

	dump["_context_info"] = map[string]interface{}{
		"id":          bfc.config.ID,
		"created_at":  bfc.createdAt,
		"last_access": bfc.lastAccess,
		"size":        len(bfc.data) + len(bfc.media),
	}

	return dump
}

func (bfc *BaseFlowContext) PrintState() {
	dump := bfc.Dump()
	data, _ := json.MarshalIndent(dump, "", "  ")
	fmt.Printf("FlowContext State [%s]:\n%s\n", bfc.config.ID, string(data))
}

// Helper methods
func (bfc *BaseFlowContext) checkSize() {
	if bfc.config.MaxSize > 0 && len(bfc.data)+len(bfc.media) >= bfc.config.MaxSize {
		bfc.logger.Warn("Context approaching maximum size",
			"current", len(bfc.data)+len(bfc.media),
			"max", bfc.config.MaxSize,
			"context_id", bfc.config.ID,
		)
	}
}

// generateID creates a unique context ID
func generateID() string {
	return fmt.Sprintf("ctx_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}