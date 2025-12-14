package context

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// BaseFlowContextV2 представляет улучшенную реализацию с type-safe геттерами
type BaseFlowContextV2 struct {
	mutex        sync.RWMutex
	data         map[string]interface{}
	imageRefs    map[string]*ImageReference
	config       *ContextConfig
	logger       interfaces.Logger
	parent       FlowContext
	children     map[string]FlowContext
	createdAt    time.Time
	lastAccess   time.Time
	imageCollection *ImageCollection
}

// NewBaseFlowContextV2 создает новый context с type-safe геттерами
func NewBaseFlowContextV2() *BaseFlowContextV2 {
	config := DefaultContextConfig()

	// Создаем image collection с лимитом 100MB
	imageCollection := NewImageCollection(100, config.Logger)

	return &BaseFlowContextV2{
		data:          make(map[string]interface{}),
		imageRefs:     make(map[string]*ImageReference),
		config:        config,
		logger:        config.Logger,
		parent:        config.Parent,
		children:      make(map[string]FlowContext),
		createdAt:     time.Now(),
		lastAccess:    time.Now(),
		imageCollection: imageCollection,
	}
}

// Type-Safe Getters с error handling

// GetString безопасно получает строковое значение
func (bfc *BaseFlowContextV2) GetString(key string) (string, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	value, exists := bfc.data[key]
	if !exists {
		return "", fmt.Errorf("key '%s' not found in context", key)
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case *string:
		if v != nil {
			return *v, nil
		}
		return "", fmt.Errorf("string pointer is nil for key '%s'", key)
	default:
		return "", fmt.Errorf("key '%s' exists but is not a string, got %T", key, value)
	}
}

// GetInt безопасно получает int значение
func (bfc *BaseFlowContextV2) GetInt(key string) (int, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	value, exists := bfc.data[key]
	if !exists {
		return 0, fmt.Errorf("key '%s' not found in context", key)
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("key '%s' contains string '%s' that cannot be converted to int: %w", key, v, err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("key '%s' exists but cannot be converted to int, got %T", key, value)
	}
}

// GetInt64 безопасно получает int64 значение
func (bfc *BaseFlowContextV2) GetInt64(key string) (int64, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	value, exists := bfc.data[key]
	if !exists {
		return 0, fmt.Errorf("key '%s' not found in context", key)
	}

	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("key '%s' contains string '%s' that cannot be converted to int64: %w", key, v, err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("key '%s' exists but cannot be converted to int64, got %T", key, value)
	}
}

// GetFloat64 безопасно получает float64 значение
func (bfc *BaseFlowContextV2) GetFloat64(key string) (float64, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	value, exists := bfc.data[key]
	if !exists {
		return 0, fmt.Errorf("key '%s' not found in context", key)
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("key '%s' contains string '%s' that cannot be converted to float64: %w", key, v, err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("key '%s' exists but cannot be converted to float64, got %T", key, value)
	}
}

// GetBool безопасно получает bool значение
func (bfc *BaseFlowContextV2) GetBool(key string) (bool, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	value, exists := bfc.data[key]
	if !exists {
		return false, fmt.Errorf("key '%s' not found in context", key)
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		switch strings.ToLower(v) {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		default:
			return false, fmt.Errorf("key '%s' contains string '%s' that cannot be converted to bool", key, v)
		}
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("key '%s' exists but cannot be converted to bool, got %T", key, value)
	}
}

// GetImageRef безопасно получает ImageReference
func (bfc *BaseFlowContextV2) GetImageRef(key string) (*ImageReference, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	// Проверяем в data map
	if value, exists := bfc.data[key]; exists {
		if ref, ok := value.(*ImageReference); ok {
			return ref, nil
		}
		return nil, fmt.Errorf("key '%s' exists but is not an ImageReference, got %T", key, value)
	}

	// Проверяем в imageRefs map
	if ref, exists := bfc.imageRefs[key]; exists {
		return ref, nil
	}

	return nil, fmt.Errorf("image reference '%s' not found in context", key)
}

// GetStringSlice безопасно получает []string
func (bfc *BaseFlowContextV2) GetStringSlice(key string) ([]string, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	value, exists := bfc.data[key]
	if !exists {
		return nil, fmt.Errorf("key '%s' not found in context", key)
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				result[i] = str
			} else {
				return nil, fmt.Errorf("item %d in slice for key '%s' is not a string, got %T", i, key, item)
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("key '%s' exists but is not a string slice, got %T", key, value)
	}
}

// GetMap безопасно получает map[string]interface{}
func (bfc *BaseFlowContextV2) GetMap(key string) (map[string]interface{}, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	value, exists := bfc.data[key]
	if !exists {
		return nil, fmt.Errorf("key '%s' not found in context", key)
	}

	if result, ok := value.(map[string]interface{}); ok {
		return result, nil
	}

	return nil, fmt.Errorf("key '%s' exists but is not a map[string]interface{}, got %T", key, value)
}

// GetObject безопасно десериализует объект в target
func (bfc *BaseFlowContextV2) GetObject(key string, target interface{}) error {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	value, exists := bfc.data[key]
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

// Setters с type checking

// SetString безопасно устанавливает строковое значение
func (bfc *BaseFlowContextV2) SetString(key, value string) error {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()
	bfc.checkSize()
	bfc.data[key] = value
	bfc.lastAccess = time.Now()
	return nil
}

// SetInt безопасно устанавливает int значение
func (bfc *BaseFlowContextV2) SetInt(key string, value int) error {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()
	bfc.checkSize()
	bfc.data[key] = value
	bfc.lastAccess = time.Now()
	return nil
}

// SetFloat64 безопасно устанавливает float64 значение
func (bfc *BaseFlowContextV2) SetFloat64(key string, value float64) error {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()
	bfc.checkSize()
	bfc.data[key] = value
	bfc.lastAccess = time.Now()
	return nil
}

// SetBool безопасно устанавливает bool значение
func (bfc *BaseFlowContextV2) SetBool(key string, value bool) error {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()
	bfc.checkSize()
	bfc.data[key] = value
	bfc.lastAccess = time.Now()
	return nil
}

// SetImageRef безопасно устанавливает ImageReference
func (bfc *BaseFlowContextV2) SetImageRef(key string, ref *ImageReference) error {
	if ref == nil {
		return fmt.Errorf("image reference cannot be nil")
	}

	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()
	bfc.checkSize()

	// Добавляем в image collection
	if err := bfc.imageCollection.Add(ref); err != nil {
		return fmt.Errorf("failed to add image to collection: %w", err)
	}

	bfc.imageRefs[key] = ref
	bfc.data[key] = ref // Также в data map для backward compatibility
	bfc.lastAccess = time.Now()

	bfc.logger.Debug("Image reference set in context",
		"key", key,
		"image_id", ref.ID,
		"source", ref.Source,
		"memory_usage", bfc.imageCollection.GetMemoryUsage(),
	)

	return nil
}

// SetImageFromURL создает ImageReference из URL и устанавливает его
func (bfc *BaseFlowContextV2) SetImageFromURL(key, url string) error {
	ref, err := NewImageFromURL(url)
	if err != nil {
		return fmt.Errorf("failed to create image reference from URL: %w", err)
	}
	return bfc.SetImageRef(key, ref)
}

// SetImageFromS3 создает ImageReference из S3 и устанавливает его
func (bfc *BaseFlowContextV2) SetImageFromS3(key, bucket, keyPath string) error {
	ref, err := NewImageFromS3(bucket, keyPath)
	if err != nil {
		return fmt.Errorf("failed to create image reference from S3: %w", err)
	}
	return bfc.SetImageRef(key, ref)
}

// Методы работы с изображениями

// LoadImageBytes загружает бинарные данные изображения по требованию
func (bfc *BaseFlowContextV2) LoadImageBytes(ctx context.Context, key string) ([]byte, error) {
	ref, err := bfc.GetImageRef(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get image reference: %w", err)
	}

	return bfc.imageCollection.LoadBytes(ctx, ref.ID)
}

// GetImageDataURL получает data URL для изображения
func (bfc *BaseFlowContextV2) GetImageDataURL(ctx context.Context, key string) (string, error) {
	ref, err := bfc.GetImageRef(key)
	if err != nil {
		return "", fmt.Errorf("failed to get image reference: %w", err)
	}

	return ref.GetDataURL(ctx)
}

// Memory Management

// GetMemoryUsage возвращает текущее использование памяти изображениями
func (bfc *BaseFlowContextV2) GetMemoryUsage() int64 {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return bfc.imageCollection.GetMemoryUsage()
}

// EvictImageCache очищает кэш для конкретного изображения
func (bfc *BaseFlowContextV2) EvictImageCache(key string) error {
	ref, err := bfc.GetImageRef(key)
	if err != nil {
		return err
	}

	ref.EvictFromCache()
	return nil
}

// EvictAllImageCache очищает кэш всех изображений
func (bfc *BaseFlowContextV2) EvictAllImageCache() {
	bfc.imageCollection.Clear()
}

// Required FlowContext interface implementation
// (Остальные методы аналогичны оригинальной реализации...)

func (bfc *BaseFlowContextV2) Set(key string, value interface{}) error {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()
	bfc.checkSize()
	bfc.data[key] = value
	bfc.lastAccess = time.Now()
	return nil
}

func (bfc *BaseFlowContextV2) Get(key string) (interface{}, bool) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()
	value, exists := bfc.data[key]
	return value, exists
}

func (bfc *BaseFlowContextV2) Delete(key string) bool {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	// Check data
	if _, exists := bfc.data[key]; exists {
		delete(bfc.data, key)
		bfc.lastAccess = time.Now()
		return true
	}

	// Check image refs
	if _, exists := bfc.imageRefs[key]; exists {
		delete(bfc.imageRefs, key)
		bfc.lastAccess = time.Now()
		return true
	}

	return false
}

func (bfc *BaseFlowContextV2) Has(key string) bool {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	bfc.lastAccess = time.Now()

	if _, exists := bfc.data[key]; exists {
		return true
	}
	_, exists := bfc.imageRefs[key]
	return exists
}

func (bfc *BaseFlowContextV2) Clear() {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	bfc.data = make(map[string]interface{})
	bfc.imageRefs = make(map[string]*ImageReference)
	bfc.imageCollection.Clear()
	bfc.lastAccess = time.Now()
}

func (bfc *BaseFlowContextV2) Keys() []string {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	keys := make([]string, 0, len(bfc.data)+len(bfc.imageRefs))
	for k := range bfc.data {
		keys = append(keys, k)
	}
	for k := range bfc.imageRefs {
		keys = append(keys, k)
	}
	return keys
}

func (bfc *BaseFlowContextV2) Size() int {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return len(bfc.data) + len(bfc.imageRefs)
}

// Другие required методы (ID, CreatedAt, Parent, и т.д.)...

func (bfc *BaseFlowContextV2) ID() string {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return bfc.config.ID
}

func (bfc *BaseFlowContextV2) CreatedAt() time.Time {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return bfc.createdAt
}

func (bfc *BaseFlowContextV2) Parent() FlowContext {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return bfc.parent
}

func (bfc *BaseFlowContextV2) CreateChild() FlowContext {
	childConfig := *bfc.config
	childConfig.ID = generateID()
	childConfig.Parent = bfc

	child := NewBaseFlowContextV2WithConfig(&childConfig)

	bfc.mutex.Lock()
	bfc.children[childConfig.ID] = child
	bfc.mutex.Unlock()

	bfc.logger.Debug("Child context created",
		"parent_id", bfc.config.ID,
		"child_id", childConfig.ID,
	)

	return child
}

func (bfc *BaseFlowContextV2) SetLogger(logger interfaces.Logger) {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()
	bfc.logger = logger
}

func (bfc *BaseFlowContextV2) GetLogger() interfaces.Logger {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()
	return bfc.logger
}

func (bfc *BaseFlowContextV2) AccumulateMedia(prefix string, mediaList []*MediaData) error {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	for i, media := range mediaList {
		key := fmt.Sprintf("%s_%d", prefix, i)
		if media != nil {
			ref := &ImageReference{
				ID:         key,
				Source:     media.URL,
				SourceType: determineSourceType(media.URL),
				MimeType:   media.MimeType,
				Size:       media.Size,
				Metadata:   media.Metadata,
			}
			bfc.imageRefs[key] = ref
		}
	}

	return nil
}

func (bfc *BaseFlowContextV2) AppendToArray(key string, value interface{}) error {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	// Get existing array or create new one
	var arr []interface{}
	if existing, ok := bfc.data[key]; ok {
		if existingArr, ok := existing.([]interface{}); ok {
			arr = existingArr
		} else {
			return fmt.Errorf("key %s exists but is not an array", key)
		}
	}

	// Append new value
	arr = append(arr, value)
	bfc.data[key] = arr
	bfc.lastAccess = time.Now()
	return nil
}

func determineSourceType(source string) SourceType {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return SourceTypeURL
	}
	if strings.HasPrefix(source, "s3://") {
		return SourceTypeS3
	}
	return SourceTypeFile
}

func (bfc *BaseFlowContextV2) Clone() FlowContext {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	// Create new context with same config but new ID
	childConfig := &ContextConfig{
		ID:                generateContextID(),
		MaxSize:           bfc.config.MaxSize,
		TTL:               bfc.config.TTL,
		EnableSerialization: bfc.config.EnableSerialization,
		Logger:            bfc.config.Logger,
		Parent:            bfc,
	}

	// Copy all data
	child := &BaseFlowContextV2{
		data:     make(map[string]interface{}),
		imageRefs: make(map[string]*ImageReference),
		config:   childConfig,
		logger:   childConfig.Logger,
		parent:   bfc,
		children: make(map[string]FlowContext),
		createdAt: time.Now(),
		lastAccess: time.Now(),
		imageCollection: NewImageCollection(100, childConfig.Logger), // Default 100MB memory
	}

	// Deep copy data
	for k, v := range bfc.data {
		child.data[k] = v
	}

	// Copy image references
	for k, v := range bfc.imageRefs {
		child.imageRefs[k] = v
	}

	bfc.children[childConfig.ID] = child

	bfc.logger.Debug("Context cloned",
		"parent_id", bfc.config.ID,
		"child_id", childConfig.ID,
	)

	return child
}

func generateContextID() string {
	return fmt.Sprintf("ctx_%d", time.Now().UnixNano())
}

func (bfc *BaseFlowContextV2) SetBytes(key string, value []byte) error {
	return bfc.Set(key, value)
}

func (bfc *BaseFlowContextV2) GetBytes(key string) ([]byte, error) {
	val, ok := bfc.Get(key)
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}
	if bytes, ok := val.([]byte); ok {
		return bytes, nil
	}
	return nil, fmt.Errorf("key %s is not byte array", key)
}

func (bfc *BaseFlowContextV2) SetFloat(key string, value float64) error {
	return bfc.Set(key, value)
}

func (bfc *BaseFlowContextV2) GetFloat(key string) (float64, error) {
	val, ok := bfc.Get(key)
	if !ok {
		return 0, fmt.Errorf("key %s not found", key)
	}
	if f, ok := val.(float64); ok {
		return f, nil
	}
	return 0, fmt.Errorf("key %s is not a float64", key)
}

func (bfc *BaseFlowContextV2) SetArray(key string, values []interface{}) error {
	return bfc.Set(key, values)
}

func (bfc *BaseFlowContextV2) GetArray(key string) ([]interface{}, error) {
	val, ok := bfc.Get(key)
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}
	if arr, ok := val.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("key %s is not an array", key)
}

func (bfc *BaseFlowContextV2) GetArraySize(key string) (int, error) {
	arr, err := bfc.GetArray(key)
	if err != nil {
		return 0, err
	}
	return len(arr), nil
}

func (bfc *BaseFlowContextV2) SetObject(key string, obj interface{}) error {
	return bfc.Set(key, obj)
}

func (bfc *BaseFlowContextV2) SetMedia(key string, media *MediaData) error {
	bfc.mutex.Lock()
	defer bfc.mutex.Unlock()

	ref := &ImageReference{
		ID:         key,
		Source:     media.URL,
		SourceType: determineSourceType(media.URL),
		MimeType:   media.MimeType,
		Size:       media.Size,
		Metadata:   media.Metadata,
	}
	bfc.imageRefs[key] = ref
	return nil
}

func (bfc *BaseFlowContextV2) GetMedia(key string) (*MediaData, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	ref, ok := bfc.imageRefs[key]
	if !ok {
		return nil, fmt.Errorf("media key %s not found", key)
	}

	return &MediaData{
		URL:      ref.Source,
		MimeType: ref.MimeType,
		Size:     ref.Size,
		Metadata: ref.Metadata,
	}, nil
}

func (bfc *BaseFlowContextV2) GetAllMedia(prefix string) ([]*MediaData, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	var mediaList []*MediaData
	for key, ref := range bfc.imageRefs {
		if strings.HasPrefix(key, prefix) {
			mediaList = append(mediaList, &MediaData{
				URL:      ref.Source,
				MimeType: ref.MimeType,
				Size:     ref.Size,
				Metadata: ref.Metadata,
			})
		}
	}
	return mediaList, nil
}

func (bfc *BaseFlowContextV2) Merge(other FlowContext) error {
	// Simple implementation - in production would need more sophisticated merging
	keys := other.Keys()
	for _, key := range keys {
		if val, ok := other.Get(key); ok {
			bfc.Set(key, val)
		}
	}
	return nil
}

func (bfc *BaseFlowContextV2) Serialize() ([]byte, error) {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	data := map[string]interface{}{
		"data": bfc.data,
	}
	return json.Marshal(data)
}

func (bfc *BaseFlowContextV2) Deserialize(data []byte) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if newData, ok := payload["data"].(map[string]interface{}); ok {
		bfc.mutex.Lock()
		bfc.data = newData
		bfc.mutex.Unlock()
	}
	return nil
}

func (bfc *BaseFlowContextV2) ToJSON() (string, error) {
	data, err := bfc.Serialize()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (bfc *BaseFlowContextV2) Dump() map[string]interface{} {
	bfc.mutex.RLock()
	defer bfc.mutex.RUnlock()

	dump := make(map[string]interface{})
	dump["data"] = bfc.data
	dump["id"] = bfc.config.ID
	dump["created_at"] = bfc.createdAt
	dump["last_access"] = bfc.lastAccess

	media := make(map[string]interface{})
	for k, ref := range bfc.imageRefs {
		media[k] = map[string]interface{}{
			"source":   ref.Source,
			"type":     ref.MimeType,
			"size":     ref.Size,
			"metadata": ref.Metadata,
		}
	}
	dump["media"] = media

	return dump
}

func (bfc *BaseFlowContextV2) PrintState() {
	dump := bfc.Dump()
	if jsonData, err := json.MarshalIndent(dump, "", "  "); err == nil {
		fmt.Println(string(jsonData))
	}
}

// Helper method
func (bfc *BaseFlowContextV2) checkSize() {
	if bfc.config.MaxSize > 0 && len(bfc.data)+len(bfc.imageRefs) >= bfc.config.MaxSize {
		bfc.logger.Warn("Context approaching maximum size",
			"current", len(bfc.data)+len(bfc.imageRefs),
			"max", bfc.config.MaxSize,
			"context_id", bfc.config.ID,
		)
	}
}

// Factory function with config
func NewBaseFlowContextV2WithConfig(config *ContextConfig) *BaseFlowContextV2 {
	if config == nil {
		config = DefaultContextConfig()
	}

	imageCollection := NewImageCollection(100, config.Logger)

	return &BaseFlowContextV2{
		data:          make(map[string]interface{}),
		imageRefs:     make(map[string]*ImageReference),
		config:        config,
		logger:        config.Logger,
		parent:        config.Parent,
		children:      make(map[string]FlowContext),
		createdAt:     time.Now(),
		lastAccess:    time.Now(),
		imageCollection: imageCollection,
	}
}