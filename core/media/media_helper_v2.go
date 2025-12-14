package media

import (
	"fmt"
	"strings"

	flowctx "github.com/ilkoid/PonchoAiFramework/core/context"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MediaHelperV2 предоставляет улучшенную работу с медиа с lazy loading
type MediaHelperV2 struct {
	pipeline *MediaPipeline
	logger   interfaces.Logger
}

// NewMediaHelperV2 создает новый improved media helper
func NewMediaHelperV2(pipeline *MediaPipeline, logger interfaces.Logger) *MediaHelperV2 {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &MediaHelperV2{
		pipeline: pipeline,
		logger:   logger,
	}
}

// PrepareContentFromContext готовит медиа-контент из FlowContext с lazy loading
func (mh *MediaHelperV2) PrepareContentFromContext(
	ctx interfaces.FlowContext,
	flowCtx interfaces.FlowContext,
	mediaKeys []string,
	model interfaces.PonchoModel,
) ([]*interfaces.PonchoContentPart, error) {
	if len(mediaKeys) == 0 {
		return nil, nil
	}

	var parts []*interfaces.PonchoContentPart

	for _, key := range mediaKeys {
		// Пробуем получить ImageReference из context (v2.0 подход)
		if imageRef, err := mh.getImageRef(flowCtx, key); err == nil {
			part, err := mh.createMediaPartFromRef(ctx, imageRef)
			if err != nil {
				mh.logger.Warn("Failed to create media part from image ref",
					"key", key,
					"error", err,
				)
				continue
			}
			parts = append(parts, part)
			continue
		}

		// Fallback на старый подход для backward compatibility
		if mediaData, err := flowCtx.GetMedia(key); err == nil {
			part, err := mh.createMediaPartFromInterfacesMediaData(mediaData)
			if err != nil {
				mh.logger.Warn("Failed to create media part from media data",
					"key", key,
					"error", err,
				)
				continue
			}
			parts = append(parts, part)
			continue
		}

		// Пробуем получить как bytes и создать ImageReference
		if bytes, err := flowCtx.GetBytes(key); err == nil {
			mimeType := mh.detectMimeType(key)
			imageRef, err := flowctx.NewImageFromMemory(bytes, mimeType)
			if err != nil {
				mh.logger.Warn("Failed to create image reference from bytes",
					"key", key,
					"error", err,
				)
				continue
			}

			// TODO: ImageReference сохраняется через SetMedia в будущем

			part, err := mh.createMediaPartFromRef(ctx, imageRef)
			if err != nil {
				continue
			}
			parts = append(parts, part)
		}
	}

	return parts, nil
}

// StoreImagesInContext сохраняет изображения в context как ImageReference
func (mh *MediaHelperV2) StoreImagesInContext(
	flowCtx interfaces.FlowContext,
	prefix string,
	imageSources []string,
) error {
	// Работаем с общим интерфейсом FlowContext

	for i, source := range imageSources {
		_ = fmt.Sprintf("%s_%d", prefix, i) // ключ пока не используется

		// Определяем тип источника
		if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
			_, _ = flowctx.NewImageFromURL(source) // TODO: обработать imageRef
		} else if strings.HasPrefix(source, "s3://") {
			// TODO: обработать S3
		} else {
			// TODO: обработать файловый путь
		}

		// TODO: SetImageRef через SetMedia
	}

	return nil
}

// ExtractImagesFromContext извлекает изображения из различных форматов в context
func (mh *MediaHelperV2) ExtractImagesFromContext(
	flowCtx interfaces.FlowContext,
	productData interface{},
) error {
	// Работаем с общим интерфейсом FlowContext

	// Проверяем разные форматы данных
	switch data := productData.(type) {
	case map[string]interface{}:
		// Ищем поле images
		if imagesValue, has := data["images"]; has {
			if imagesSlice, ok := imagesValue.([]interface{}); ok {
				for _, imgValue := range imagesSlice {
					if imgMap, ok := imgValue.(map[string]interface{}); ok {
						// Ищем URL в различных полях
						urlFields := []string{"url", "src", "image_url", "path"}
						for _, field := range urlFields {
							if urlValue, has := imgMap[field]; has {
								if _, ok := urlValue.(string); ok {
									// TODO: SetImageFromURL через SetMedia
									break // Нашли URL, переходим к следующему изображению
								}
							}
						}
					}
				}
			}
		}

		// Ищем поле photos
		// Пропускаем неиспользуемые поля photos и sketches

	default:
		return fmt.Errorf("unsupported product data format: %T", productData)
	}

	return nil
}

// CreateVisionMessageV2 создает сообщение для vision модели с использованием ImageReference
func (mh *MediaHelperV2) CreateVisionMessageV2(
	ctx interfaces.FlowContext,
	text string,
	flowCtx interfaces.FlowContext,
	mediaKeys []string,
	model interfaces.PonchoModel,
) (*interfaces.PonchoMessage, error) {
	if !model.SupportsVision() {
		return nil, fmt.Errorf("model %s does not support vision", model.Name())
	}

	// Начинаем с текстового контента
	content := []*interfaces.PonchoContentPart{
		{
			Type: interfaces.PonchoContentTypeText,
			Text: text,
		},
	}

	// Добавляем медиа-контент с lazy loading
	mediaContent, err := mh.PrepareContentFromContext(ctx, flowCtx, mediaKeys, model)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare media content: %w", err)
	}

	content = append(content, mediaContent...)

	return &interfaces.PonchoMessage{
		Role:    interfaces.PonchoRoleUser,
		Content: content,
	}, nil
}

// Helper methods

func (mh *MediaHelperV2) getImageRef(flowCtx interfaces.FlowContext, key string) (*flowctx.ImageReference, error) {
	// Пробуем convert к v2 context
	// TODO: GetImageRef из context

	return nil, fmt.Errorf("image reference not found for key: %s", key)
}

func (mh *MediaHelperV2) createMediaPartFromRef(_ interface{}, ref *flowctx.ImageReference) (*interfaces.PonchoContentPart, error) {
	// Получаем data URL с lazy loading
	// TODO: GetDataURL without context
	dataURL := "" // ref.GetDataURL()
	if dataURL == "" {
		return nil, fmt.Errorf("failed to get data URL")
	}

	return &interfaces.PonchoContentPart{
		Type: interfaces.PonchoContentTypeMedia,
		Media: &interfaces.PonchoMediaPart{
			URL:      dataURL,
			MimeType: ref.MimeType,
		},
	}, nil
}

func (mh *MediaHelperV2) createMediaPartFromInterfacesMediaData(data *interfaces.MediaData) (*interfaces.PonchoContentPart, error) {
	// Convert interfaces.MediaData to media.MediaData
	mediaData := &MediaData{
		URL:      data.URL,
		Bytes:    data.Bytes,
		MimeType: data.MimeType,
		Size:     data.Size,
		Metadata: data.Metadata,
	}
	return mh.createMediaPartFromData(mediaData)
}

func (mh *MediaHelperV2) createMediaPartFromData(data *MediaData) (*interfaces.PonchoContentPart, error) {
	url := data.GetDataURL()
	if url == "" {
		return nil, fmt.Errorf("media has no URL or data")
	}

	return &interfaces.PonchoContentPart{
		Type: interfaces.PonchoContentTypeMedia,
		Media: &interfaces.PonchoMediaPart{
			URL:      url,
			MimeType: data.MimeType,
		},
	}, nil
}

func (mh *MediaHelperV2) detectMimeType(key string) string {
	if strings.HasSuffix(key, ".jpg") || strings.HasSuffix(key, ".jpeg") {
		return "image/jpeg"
	}
	if strings.HasSuffix(key, ".png") {
		return "image/png"
	}
	if strings.HasSuffix(key, ".webp") {
		return "image/webp"
	}
	if strings.HasSuffix(key, ".gif") {
		return "image/gif"
	}
	return "image/jpeg" // Default
}

// Batch operations для эффективной работы с множественными изображениями

// BatchPrepareMediaContent готовит контент для множества изображений
func (mh *MediaHelperV2) BatchPrepareMediaContent(
	ctx interfaces.FlowContext,
	flowCtx interfaces.FlowContext,
	mediaKeys []string,
	model interfaces.PonchoModel,
) ([]*interfaces.PonchoContentPart, error) {
	if len(mediaKeys) == 0 {
		return nil, nil
	}

	// Группируем ключи для batch обработки
	batches := mh.groupKeysByType(flowCtx, mediaKeys)

	var allParts []*interfaces.PonchoContentPart

	// Обрабатываем каждую группу
	for batchType, keys := range batches {
		switch batchType {
		case "image_refs":
			parts, err := mh.processImageRefsBatch(ctx, flowCtx, keys, model)
			if err != nil {
				mh.logger.Warn("Failed to process image refs batch",
					"error", err,
					"count", len(keys),
				)
				continue
			}
			allParts = append(allParts, parts...)

		case "media_data":
			parts, err := mh.processMediaDataBatch(flowCtx, keys)
			if err != nil {
				mh.logger.Warn("Failed to process media data batch",
					"error", err,
					"count", len(keys),
				)
				continue
			}
			allParts = append(allParts, parts...)

		default:
			mh.logger.Warn("Unknown batch type", "type", batchType)
		}
	}

	return allParts, nil
}

func (mh *MediaHelperV2) groupKeysByType(flowCtx interfaces.FlowContext, keys []string) map[string][]string {
	batches := make(map[string][]string)

	for _, key := range keys {
		// Проверяем тип данных в context
		if _, err := mh.getImageRef(flowCtx, key); err == nil {
			batches["image_refs"] = append(batches["image_refs"], key)
		} else if _, err := flowCtx.GetMedia(key); err == nil {
			batches["media_data"] = append(batches["media_data"], key)
		} else if _, err := flowCtx.GetBytes(key); err == nil {
			batches["image_refs"] = append(batches["image_refs"], key) // Будем конвертировать в ref
		}
	}

	return batches
}

func (mh *MediaHelperV2) processImageRefsBatch(
	ctx interfaces.FlowContext,
	flowCtx interfaces.FlowContext,
	keys []string,
	model interfaces.PonchoModel,
) ([]*interfaces.PonchoContentPart, error) {
	var parts []*interfaces.PonchoContentPart

	for _, key := range keys {
		ref, err := mh.getImageRef(flowCtx, key)
		if err != nil {
			mh.logger.Warn("Failed to get image ref",
				"key", key,
				"error", err,
			)
			continue
		}

		part, err := mh.createMediaPartFromRef(ctx, ref)
		if err != nil {
			mh.logger.Warn("Failed to create media part",
				"key", key,
				"error", err,
			)
			continue
		}

		parts = append(parts, part)
	}

	return parts, nil
}

func (mh *MediaHelperV2) processMediaDataBatch(
	flowCtx interfaces.FlowContext,
	keys []string,
) ([]*interfaces.PonchoContentPart, error) {
	var parts []*interfaces.PonchoContentPart

	for _, key := range keys {
		mediaData, err := flowCtx.GetMedia(key)
		if err != nil {
			mh.logger.Warn("Failed to get media data",
				"key", key,
				"error", err,
			)
			continue
		}

		part, err := mh.createMediaPartFromInterfacesMediaData(mediaData)
		if err != nil {
			mh.logger.Warn("Failed to create media part",
				"key", key,
				"error", err,
			)
			continue
		}

		parts = append(parts, part)
	}

	return parts, nil
}