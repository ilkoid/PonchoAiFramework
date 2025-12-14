package context

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ImageReference представляет ссылку на изображение без хранения binary данных
type ImageReference struct {
	// Идентификатор изображения в контексте
	ID string `json:"id"`

	// Источник изображения (URL, S3 path, file path)
	Source string `json:"source"`

	// Тип источника
	SourceType SourceType `json:"source_type"`

	// Метаданные
	MimeType string            `json:"mime_type"`
	Size     int64             `json:"size"`
	Width    int               `json:"width,omitempty"`
	Height   int               `json:"height,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Кэширование
	Cached     bool      `json:"cached"`
	CachedAt   time.Time `json:"cached_at,omitempty"`
	CacheExpiry time.Time `json:"cache_expiry,omitempty"`

	// Ленивая загрузка
	loader    ImageLoader `json:"-"`
	loadMutex sync.RWMutex `json:"-"`
	loaded    bool        `json:"-"`
	bytes     []byte      `json:"-"`
	error     error       `json:"-"`
}

// SourceType определяет тип источника изображения
type SourceType string

const (
	SourceTypeURL      SourceType = "url"
	SourceTypeS3       SourceType = "s3"
	SourceTypeFile     SourceType = "file"
	SourceTypeBase64   SourceType = "base64"
	SourceTypeMemory   SourceType = "memory" // Только для маленьких изображений
)

// ImageLoader определяет интерфейс для загрузки изображений
type ImageLoader interface {
	LoadImage(ctx context.Context, ref *ImageReference) ([]byte, error)
	LoadImageMetadata(ctx context.Context, ref *ImageReference) error
}

// NewImageFromURL создает ImageReference из URL
func NewImageFromURL(url string) (*ImageReference, error) {
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	return &ImageReference{
		ID:         generateImageID(),
		Source:     url,
		SourceType: SourceTypeURL,
		MimeType:   detectMimeTypeFromURL(url),
		loader:     NewURLImageLoader(),
	}, nil
}

// NewImageFromS3 создает ImageReference из S3 пути
func NewImageFromS3(bucket, key string) (*ImageReference, error) {
	if bucket == "" || key == "" {
		return nil, fmt.Errorf("bucket and key cannot be empty")
	}

	source := fmt.Sprintf("s3://%s/%s", bucket, key)

	return &ImageReference{
		ID:         generateImageID(),
		Source:     source,
		SourceType: SourceTypeS3,
		MimeType:   detectMimeTypeFromPath(key),
		loader:     NewS3ImageLoader(bucket),
	}, nil
}

// NewImageFromMemory создает ImageReference для бинарных данных (только маленькие изображения)
func NewImageFromMemory(data []byte, mimeType string) (*ImageReference, error) {
	if len(data) > 10*1024*1024 { // 10MB limit
		return nil, fmt.Errorf("image too large for memory storage (%d bytes)", len(data))
	}

	return &ImageReference{
		ID:         generateImageID(),
		Source:     "memory",
		SourceType: SourceTypeMemory,
		MimeType:   mimeType,
		Size:       int64(len(data)),
		Cached:     true,
		loader:     NewMemoryImageLoader(data),
	}, nil
}

// GetBytes загружает бинарные данные по требованию (lazy loading)
func (ir *ImageReference) GetBytes(ctx context.Context) ([]byte, error) {
	// Если уже загружено - возвращаем из кэша
	ir.loadMutex.RLock()
	if ir.loaded {
		ir.loadMutex.RUnlock()
		if ir.error != nil {
			return nil, ir.error
		}
		return ir.bytes, nil
	}
	ir.loadMutex.RUnlock()

	// Загружаем первый раз
	ir.loadMutex.Lock()
	defer ir.loadMutex.Unlock()

	// Double-check после получения lock
	if ir.loaded {
		if ir.error != nil {
			return nil, ir.error
		}
		return ir.bytes, nil
	}

	// Выполняем загрузку
	if ir.loader == nil {
		return nil, fmt.Errorf("no loader configured for image %s", ir.ID)
	}

	bytes, err := ir.loader.LoadImage(ctx, ir)
	if err != nil {
		ir.error = err
		ir.loaded = true
		return nil, fmt.Errorf("failed to load image %s: %w", ir.ID, err)
	}

	ir.bytes = bytes
	ir.loaded = true
	ir.Size = int64(len(bytes))
	ir.Cached = true
	ir.CachedAt = time.Now()
	ir.CacheExpiry = time.Now().Add(30 * time.Minute) // Кэш на 30 минут

	return ir.bytes, nil
}

// GetDataURL возвращает data URL (lazy loading)
func (ir *ImageReference) GetDataURL(ctx context.Context) (string, error) {
	if ir.SourceType == SourceTypeBase64 {
		return ir.Source, nil
	}

	bytes, err := ir.GetBytes(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("data:%s;base64,%x", ir.MimeType, bytes), nil
}

// IsCached проверяет, загружено ли изображение в память
func (ir *ImageReference) IsCached() bool {
	ir.loadMutex.RLock()
	defer ir.loadMutex.RUnlock()
	return ir.Cached && ir.loaded
}

// EvictFromCache очищает изображение из памяти
func (ir *ImageReference) EvictFromCache() {
	ir.loadMutex.Lock()
	defer ir.loadMutex.Unlock()

	ir.bytes = nil
	ir.loaded = false
	ir.Cached = false
	ir.error = nil
}

// EstimateMemoryUsage оценивает использование памяти
func (ir *ImageReference) EstimateMemoryUsage() int64 {
	if ir.IsCached() {
		return int64(len(ir.bytes))
	}
	return 0 // Не загружено в память
}

// ImageCollection управляет коллекцией изображений с контролем памяти
type ImageCollection struct {
	images      map[string]*ImageReference
	maxMemory   int64 // Максимальный размер памяти в байтах
	currentMem  int64
	mutex       sync.RWMutex
	logger      interfaces.Logger
}

// NewImageCollection создает новую коллекцию
func NewImageCollection(maxMemoryMB int64, logger interfaces.Logger) *ImageCollection {
	return &ImageCollection{
		images:     make(map[string]*ImageReference),
		maxMemory:  maxMemoryMB * 1024 * 1024,
		logger:     logger,
	}
}

// Add добавляет изображение в коллекцию
func (ic *ImageCollection) Add(ref *ImageReference) error {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	if _, exists := ic.images[ref.ID]; exists {
		return fmt.Errorf("image %s already exists", ref.ID)
	}

	// Проверяем лимит памяти
	if ic.currentMem >= ic.maxMemory {
		// Evict oldest cached images
		ic.evictOldestCached()
	}

	ic.images[ref.ID] = ref
	return nil
}

// Get получает изображение по ID
func (ic *ImageCollection) Get(id string) (*ImageReference, bool) {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	ref, exists := ic.images[id]
	return ref, exists
}

// LoadBytes загружает бинарные данные для изображения
func (ic *ImageCollection) LoadBytes(ctx context.Context, id string) ([]byte, error) {
	ic.mutex.RLock()
	ref, exists := ic.images[id]
	ic.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("image %s not found", id)
	}

	// Проверяем лимит памяти перед загрузкой
	if ic.currentMem >= ic.maxMemory {
		ic.evictOldestCached()
	}

	bytes, err := ref.GetBytes(ctx)
	if err != nil {
		return nil, err
	}

	// Обновляем счетчик памяти
	ic.mutex.Lock()
	ic.currentMem += int64(len(bytes))
	ic.mutex.Unlock()

	return bytes, nil
}

// evictOldestCached очищает самые старые кэшированные изображения
func (ic *ImageCollection) evictOldestCached() {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	var oldestTime time.Time
	var oldestID string

	for id, ref := range ic.images {
		if ref.IsCached() {
			if oldestTime.IsZero() || ref.CachedAt.Before(oldestTime) {
				oldestTime = ref.CachedAt
				oldestID = id
			}
		}
	}

	if oldestID != "" {
		ref := ic.images[oldestID]
		memSize := ref.EstimateMemoryUsage()
		ref.EvictFromCache()
		ic.currentMem -= memSize

		ic.logger.Debug("Evicted image from cache",
			"image_id", oldestID,
			"freed_bytes", memSize,
			"current_memory", ic.currentMem,
		)
	}
}

// GetMemoryUsage возвращает текущее использование памяти
func (ic *ImageCollection) GetMemoryUsage() int64 {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()
	return ic.currentMem
}

// Clear очищает всю коллекцию
func (ic *ImageCollection) Clear() {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	for _, ref := range ic.images {
		ref.EvictFromCache()
	}

	ic.images = make(map[string]*ImageReference)
	ic.currentMem = 0
}

// Helper functions
func generateImageID() string {
	return fmt.Sprintf("img_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func detectMimeTypeFromURL(url string) string {
	// Простая реализация - в реальной системе использовать http.DetectContentType
	if strings.Contains(url, ".jpg") || strings.Contains(url, ".jpeg") {
		return "image/jpeg"
	}
	if strings.Contains(url, ".png") {
		return "image/png"
	}
	if strings.Contains(url, ".webp") {
		return "image/webp"
	}
	return "image/jpeg" // Default
}

func detectMimeTypeFromPath(path string) string {
	if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		return "image/jpeg"
	}
	if strings.HasSuffix(path, ".png") {
		return "image/png"
	}
	if strings.HasSuffix(path, ".webp") {
		return "image/webp"
	}
	return "image/jpeg" // Default
}

// Реализации загрузчиков изображений

// URLImageLoader загружает изображения из URL
type URLImageLoader struct{}

func NewURLImageLoader() *URLImageLoader {
	return &URLImageLoader{}
}

func (l *URLImageLoader) LoadImage(ctx context.Context, ref *ImageReference) ([]byte, error) {
	// Заглушка - реализация должна загружать изображение по URL
	return nil, fmt.Errorf("URLImageLoader not implemented")
}

func (l *URLImageLoader) LoadImageMetadata(ctx context.Context, ref *ImageReference) error {
	// Заглушка - реализация должна загружать метаданные
	return fmt.Errorf("URLImageLoader metadata not implemented")
}

// S3ImageLoader загружает изображения из S3
type S3ImageLoader struct {
	bucket string
}

func NewS3ImageLoader(bucket string) *S3ImageLoader {
	return &S3ImageLoader{bucket: bucket}
}

func (l *S3ImageLoader) LoadImage(ctx context.Context, ref *ImageReference) ([]byte, error) {
	// Заглушка - реализация должна загружать изображение из S3
	return nil, fmt.Errorf("S3ImageLoader not implemented")
}

func (l *S3ImageLoader) LoadImageMetadata(ctx context.Context, ref *ImageReference) error {
	// Заглушка - реализация должна загружать метаданные из S3
	return fmt.Errorf("S3ImageLoader metadata not implemented")
}

// MemoryImageLoader хранит изображения в памяти
type MemoryImageLoader struct {
	data []byte
}

func NewMemoryImageLoader(data []byte) *MemoryImageLoader {
	return &MemoryImageLoader{data: data}
}

func (l *MemoryImageLoader) LoadImage(ctx context.Context, ref *ImageReference) ([]byte, error) {
	return l.data, nil
}

func (l *MemoryImageLoader) LoadImageMetadata(ctx context.Context, ref *ImageReference) error {
	ref.Size = int64(len(l.data))
	return nil
}