package config

// ConfigManager implements the configuration management system for the PonchoFramework.
// It provides the ConfigManager interface and its implementation, handling loading,
// validation, and management of configuration from YAML/JSON files.
//
// Key features:
// - Environment variable substitution for sensitive data using ${VAR_NAME} syntax
// - Configuration for models, tools, flows, and system settings
// - Hot-reload capabilities with configuration change watchers
// - Comprehensive configuration validation with detailed error reporting
// - Factory methods for creating and initializing components from configuration
// - Type-safe configuration access with automatic type conversion
// - Support for multiple configuration file formats and merging
//
// This serves as the central configuration hub for the entire framework,
// providing a unified interface for all configuration needs and ensuring
// consistency across all framework components.

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ConfigManager интерфейс для управления конфигурацией
type ConfigManager interface {
	// Load загружает конфигурацию
	Load() error

	// Reload перезагружает конфигурацию
	Reload() error

	// Get возвращает значение по ключу
	Get(key string) interface{}

	// GetString возвращает строковое значение
	GetString(key string) string

	// GetInt возвращает целочисленное значение
	GetInt(key string) int

	// GetFloat64 возвращает значение float64
	GetFloat64(key string) float64

	// GetBool возвращает булево значение
	GetBool(key string) bool

	// GetDuration возвращает длительность
	GetDuration(key string) time.Duration

	// GetStringSlice возвращает срез строк
	GetStringSlice(key string) []string

	// Set устанавливает значение
	Set(key string, value interface{})

	// Has проверяет наличие ключа
	Has(key string) bool

	// GetSection возвращает секцию конфигурации
	GetSection(section string) map[string]interface{}

	// Validate валидирует конфигурацию
	Validate() error

	// GetConfig возвращает полную конфигурацию
	GetConfig() *ConfigData

	// Watch отслеживает изменения конфигурации
	Watch(callback func(*ConfigData)) error

	// GetModelConfigs возвращает конфигурации моделей
	GetModelConfigs() (map[string]*interfaces.ModelConfig, error)

	// LoadAndInitializeModels загружает и инициализирует модели
	LoadAndInitializeModels() (map[string]interfaces.PonchoModel, error)
}

// ConfigManagerImpl реализация ConfigManager
type ConfigManagerImpl struct {
	loader          ConfigLoader
	validator       ConfigValidator
	logger          interfaces.Logger
	modelFactoryMgr  *ModelFactoryManager
	modelValidator  *ModelConfigValidator
	modelInitializer *ModelInitializer

	config    *ConfigData
	watchers  []func(*ConfigData)
	envPrefix string
	filePaths []string
}

// ConfigOptions опции для создания ConfigManager
type ConfigOptions struct {
	FilePaths []string
	EnvPrefix string
	Loader    ConfigLoader
	Validator ConfigValidator
	Logger    interfaces.Logger
}

// NewConfigManager создает новый экземпляр ConfigManager
func NewConfigManager(opts ConfigOptions) *ConfigManagerImpl {
	if opts.Loader == nil {
		opts.Loader = NewConfigLoader(opts.Logger)
	}

	if opts.Validator == nil {
		opts.Validator = NewConfigValidator(opts.Logger)
	}

	if opts.EnvPrefix == "" {
		opts.EnvPrefix = "PONCHO_"
	}

	// Initialize model-related components
	modelFactoryMgr := NewModelFactoryManager(opts.Logger)
	modelValidator := NewModelConfigValidator(opts.Logger)
	modelRegistry := NewModelRegistry(opts.Logger)
	modelInitializer := NewModelInitializer(modelRegistry, opts.Logger)

	return &ConfigManagerImpl{
		loader:          opts.Loader,
		validator:       opts.Validator,
		logger:          opts.Logger,
		modelFactoryMgr:  modelFactoryMgr,
		modelValidator:  modelValidator,
		modelInitializer: modelInitializer,
		envPrefix:       opts.EnvPrefix,
		filePaths:       opts.FilePaths,
		watchers:        make([]func(*ConfigData), 0),
	}
}

// Load загружает конфигурацию
func (cm *ConfigManagerImpl) Load() error {
	cm.logger.Info("Loading configuration", "files", cm.filePaths)

	var configData *ConfigData
	var err error

	// Загружаем из файлов
	if len(cm.filePaths) > 0 {
		configData, err = cm.loader.LoadMultiple(cm.filePaths)
		if err != nil {
			return fmt.Errorf("failed to load configuration files: %w", err)
		}
	} else {
		// Создаем пустую конфигурацию
		configData = &ConfigData{
			Source: "memory",
			Format: "yaml",
			Data:   make(map[string]interface{}),
		}
	}

	// Применяем environment variables
	cm.applyEnvironmentVariables(configData)

	// Валидируем конфигурацию
	if err := cm.validator.Validate(configData); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	cm.config = configData

	// Уведомляем watchers
	cm.notifyWatchers()

	cm.logger.Info("Configuration loaded successfully", "source", configData.Source)

	return nil
}

// Reload перезагружает конфигурацию
func (cm *ConfigManagerImpl) Reload() error {
	cm.logger.Info("Reloading configuration")

	// Очищаем кэши и watchers
	cm.config = nil

	// Загружаем заново
	return cm.Load()
}

// Get возвращает значение по ключу
func (cm *ConfigManagerImpl) Get(key string) interface{} {
	if cm.config == nil {
		return nil
	}

	value, _ := cm.getValueByPath(cm.config.Data, key)
	return value
}

// GetString возвращает строковое значение
func (cm *ConfigManagerImpl) GetString(key string) string {
	value := cm.Get(key)
	if value == nil {
		return ""
	}

	if str, ok := value.(string); ok {
		return str
	}

	return fmt.Sprintf("%v", value)
}

// GetInt возвращает целочисленное значение
func (cm *ConfigManagerImpl) GetInt(key string) int {
	value := cm.Get(key)
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return 0
}

// GetFloat64 возвращает значение float64
func (cm *ConfigManagerImpl) GetFloat64(key string) float64 {
	value := cm.Get(key)
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}

	return 0
}

// GetBool возвращает булево значение
func (cm *ConfigManagerImpl) GetBool(key string) bool {
	value := cm.Get(key)
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true"
	}

	return false
}

// GetDuration возвращает длительность
func (cm *ConfigManagerImpl) GetDuration(key string) time.Duration {
	value := cm.Get(key)
	if value == nil {
		return 0
	}

	if str, ok := value.(string); ok {
		if duration, err := time.ParseDuration(str); err == nil {
			return duration
		}
	}

	return 0
}

// GetStringSlice возвращает срез строк
func (cm *ConfigManagerImpl) GetStringSlice(key string) []string {
	value := cm.Get(key)
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				result[i] = str
			} else {
				result[i] = fmt.Sprintf("%v", item)
			}
		}
		return result
	case string:
		return strings.Split(v, ",")
	}

	return nil
}

// Set устанавливает значение
func (cm *ConfigManagerImpl) Set(key string, value interface{}) {
	if cm.config == nil {
		cm.config = &ConfigData{
			Source: "memory",
			Format: "yaml",
			Data:   make(map[string]interface{}),
		}
	}

	cm.setValueByPath(cm.config.Data, key, value)

	// Уведомляем watchers
	cm.notifyWatchers()
}

// Has проверяет наличие ключа
func (cm *ConfigManagerImpl) Has(key string) bool {
	return cm.Get(key) != nil
}

// GetSection возвращает секцию конфигурации
func (cm *ConfigManagerImpl) GetSection(section string) map[string]interface{} {
	if cm.config == nil {
		return nil
	}

	value, exists := cm.getValueByPath(cm.config.Data, section)
	if !exists {
		return nil
	}

	if sectionMap, ok := value.(map[string]interface{}); ok {
		return sectionMap
	}

	return nil
}

// Validate валидирует конфигурацию
func (cm *ConfigManagerImpl) Validate() error {
	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	return cm.validator.Validate(cm.config)
}

// GetConfig возвращает полную конфигурацию
func (cm *ConfigManagerImpl) GetConfig() *ConfigData {
	return cm.config
}

// Watch отслеживает изменения конфигурации
func (cm *ConfigManagerImpl) Watch(callback func(*ConfigData)) error {
	cm.watchers = append(cm.watchers, callback)
	return nil
}

// applyEnvironmentVariables применяет переменные окружения
func (cm *ConfigManagerImpl) applyEnvironmentVariables(config *ConfigData) {
	cm.logger.Debug("Applying environment variables", "prefix", cm.envPrefix)

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, cm.envPrefix) {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(parts[0][len(cm.envPrefix):])
		value := parts[1]

		// Конвертируем ключ из ENV_NOTATION в dot.notation
		// Заменяем только двойные подчеркивания на точки, чтобы сохранить подчеркивания в именах
		key = strings.ReplaceAll(key, "__", ".")

		// Определяем тип значения
		convertedValue := cm.convertEnvironmentValue(value)

		// Устанавливаем значение
		cm.setValueByPath(config.Data, key, convertedValue)

		cm.logger.Debug("Applied environment variable", "key", key, "value", value)
	}
}

// convertEnvironmentValue конвертирует значение из переменной окружения
func (cm *ConfigManagerImpl) convertEnvironmentValue(value string) interface{} {
	// Попытка конвертировать в bool
	if strings.ToLower(value) == "true" {
		return true
	}
	if strings.ToLower(value) == "false" {
		return false
	}

	// Попытка конвертировать в число
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}

	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	// Попытка конвертировать в duration
	if _, err := time.ParseDuration(value); err == nil {
		return value // Возвращаем исходную строку, а не конвертированную
	}

	// Возвращаем как строку
	return value
}

// getValueByPath получает значение по пути
func (cm *ConfigManagerImpl) getValueByPath(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			value, exists := current[part]
			return value, exists
		}

		next, exists := current[part]
		if !exists {
			return nil, false
		}

		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return nil, false
		}

		current = nextMap
	}

	return nil, false
}

// setValueByPath устанавливает значение по пути
func (cm *ConfigManagerImpl) setValueByPath(data map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}

		next, exists := current[part]
		if !exists {
			nextMap := make(map[string]interface{})
			current[part] = nextMap
			current = nextMap
		} else {
			nextMap, ok := next.(map[string]interface{})
			if !ok {
				// Заменяем на map если тип не совпадает
				nextMap = make(map[string]interface{})
				current[part] = nextMap
			}
			current = nextMap
		}
	}
}

// notifyWatchers уведомляет всех watchers об изменениях
func (cm *ConfigManagerImpl) notifyWatchers() {
	for _, watcher := range cm.watchers {
		go watcher(cm.config)
	}
}

// GetModelConfigs возвращает конфигурации моделей
func (cm *ConfigManagerImpl) GetModelConfigs() (map[string]*interfaces.ModelConfig, error) {
	if cm.config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	// Use the loader to extract model configurations
	loader, ok := cm.loader.(*ConfigLoaderImpl)
	if !ok {
		return nil, fmt.Errorf("loader does not support model configuration loading")
	}

	return loader.LoadModelConfigs(cm.config)
}

// LoadAndInitializeModels загружает и инициализирует модели
func (cm *ConfigManagerImpl) LoadAndInitializeModels() (map[string]interfaces.PonchoModel, error) {
	cm.logger.Info("Loading and initializing models")

	// Get model configurations
	modelConfigs, err := cm.GetModelConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to get model configurations: %w", err)
	}

	// Validate all model configurations
	for name, config := range modelConfigs {
		if err := cm.modelValidator.ValidateConfig(config); err != nil {
			cm.logger.Error("Model configuration validation failed",
				"name", name,
				"error", err)
			return nil, fmt.Errorf("invalid model configuration for %s: %w", name, err)
		}
	}

	// Initialize models using the factory manager
	models := make(map[string]interfaces.PonchoModel)
	var errors []error

	for name, config := range modelConfigs {
		cm.logger.Debug("Initializing model", "name", name, "provider", config.Provider, "model_name", config.ModelName)

		model, err := cm.modelFactoryMgr.CreateModel(config)
		if err != nil {
			cm.logger.Error("Failed to create model",
				"name", name,
				"provider", config.Provider,
				"error", err)
			errors = append(errors, fmt.Errorf("failed to create model %s: %w", name, err))
			continue
		}

		models[name] = model
		cm.logger.Info("Model initialized successfully",
			"name", name,
			"provider", config.Provider,
			"model_name", config.ModelName,
			"model", model.Name())
	}

	if len(errors) > 0 {
		return models, fmt.Errorf("some models failed to initialize: %v", errors)
	}

	cm.logger.Info("All models loaded and initialized successfully", "count", len(models))

	return models, nil
}

// GetModelFactoryManager returns the model factory manager
func (cm *ConfigManagerImpl) GetModelFactoryManager() *ModelFactoryManager {
	return cm.modelFactoryMgr
}

// GetModelValidator returns the model configuration validator
func (cm *ConfigManagerImpl) GetModelValidator() *ModelConfigValidator {
	return cm.modelValidator
}

// ReloadModels перезагружает только модели
func (cm *ConfigManagerImpl) ReloadModels() (map[string]interfaces.PonchoModel, error) {
	cm.logger.Info("Reloading models")

	// Reload configuration
	if err := cm.Reload(); err != nil {
		return nil, fmt.Errorf("failed to reload configuration: %w", err)
	}

	// Reinitialize models
	return cm.LoadAndInitializeModels()
}

// ValidateModelConfigs валидирует все конфигурации моделей
func (cm *ConfigManagerImpl) ValidateModelConfigs() error {
	modelConfigs, err := cm.GetModelConfigs()
	if err != nil {
		return fmt.Errorf("failed to get model configurations: %w", err)
	}

	var errors []error
	for name, config := range modelConfigs {
		if err := cm.modelValidator.ValidateConfig(config); err != nil {
			errors = append(errors, fmt.Errorf("model %s: %w", name, err))
		}

		// Also validate with factory
		if err := cm.modelFactoryMgr.ValidateConfig(config); err != nil {
			errors = append(errors, fmt.Errorf("model %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("model configuration validation failed: %v", errors)
	}

	return nil
}

// GetConfigAsStruct конвертирует конфигурацию в структуру
func (cm *ConfigManagerImpl) GetConfigAsStruct(target interface{}) error {
	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Используем рефлексию для конвертации
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must point to a struct")
	}

	return cm.convertMapToStruct(cm.config.Data, targetValue)
}

// convertMapToStruct рекурсивно конвертирует map в struct
func (cm *ConfigManagerImpl) convertMapToStruct(data map[string]interface{}, target reflect.Value) error {
	targetType := target.Type()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := target.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Получаем имя поля из тега yaml или json
		fieldName := field.Name
		if yamlTag := field.Tag.Get("yaml"); yamlTag != "" {
			parts := strings.Split(yamlTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		} else if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		value, exists := data[fieldName]
		if !exists {
			continue
		}

		if err := cm.convertValueToField(value, fieldValue); err != nil {
			return fmt.Errorf("failed to convert field %s: %w", fieldName, err)
		}
	}

	return nil
}

// convertValueToField конвертирует значение в поле структуры
func (cm *ConfigManagerImpl) convertValueToField(value interface{}, field reflect.Value) error {
	if value == nil {
		return nil
	}

	valueType := reflect.TypeOf(value)
	fieldType := field.Type()

	// Если типы совпадают, устанавливаем напрямую
	if valueType == fieldType {
		field.Set(reflect.ValueOf(value))
		return nil
	}

	// Конвертация между совместимыми типами
	if valueType.ConvertibleTo(fieldType) {
		field.Set(reflect.ValueOf(value).Convert(fieldType))
		return nil
	}

	// Специальные случаи для сложных типов
	switch fieldType.Kind() {
	case reflect.Map:
		if valueMap, ok := value.(map[string]interface{}); ok {
			newMap := reflect.MakeMap(fieldType)
			for k, v := range valueMap {
				keyValue := reflect.ValueOf(k)
				if !keyValue.Type().ConvertibleTo(fieldType.Key()) {
					continue
				}

				valueElement := reflect.New(fieldType.Elem()).Elem()
				if err := cm.convertValueToField(v, valueElement); err != nil {
					continue
				}

				newMap.SetMapIndex(keyValue.Convert(fieldType.Key()), valueElement)
			}
			field.Set(newMap)
		}
	case reflect.Slice:
		if valueSlice, ok := value.([]interface{}); ok {
			slice := reflect.MakeSlice(fieldType, len(valueSlice), len(valueSlice))
			for i, v := range valueSlice {
				if err := cm.convertValueToField(v, slice.Index(i)); err != nil {
					return err
				}
			}
			field.Set(slice)
		}
	case reflect.Struct:
		if valueMap, ok := value.(map[string]interface{}); ok {
			return cm.convertMapToStruct(valueMap, field)
		}
	}

	return nil
}
