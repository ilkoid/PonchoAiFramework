package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"gopkg.in/yaml.v3"
)

// ConfigLoader интерфейс для загрузки конфигурации
type ConfigLoader interface {
	// LoadFromFile загружает конфигурацию из файла
	LoadFromFile(filename string) (*ConfigData, error)

	// LoadFromBytes загружает конфигурацию из байтов
	LoadFromBytes(data []byte, format string) (*ConfigData, error)

	// LoadFromReader загружает конфигурацию из Reader
	LoadFromReader(reader io.Reader, format string) (*ConfigData, error)

	// LoadMultiple загружает и объединяет несколько конфигурационных файлов
	LoadMultiple(filenames []string) (*ConfigData, error)

	// SupportedFormats возвращает поддерживаемые форматы
	SupportedFormats() []string
}

// ConfigLoaderImpl реализация ConfigLoader
type ConfigLoaderImpl struct {
	logger interfaces.Logger
}

// NewConfigLoader создает новый экземпляр ConfigLoader
func NewConfigLoader(logger interfaces.Logger) *ConfigLoaderImpl {
	return &ConfigLoaderImpl{
		logger: logger,
	}
}

// ConfigData представляет загруженные данные конфигурации
type ConfigData struct {
	Source string                 // Источник конфигурации (файл, env и т.д.)
	Format string                 // Формат (yaml, json)
	Data   map[string]interface{} // Сырые данные
}

// LoadFromFile загружает конфигурацию из файла
func (cl *ConfigLoaderImpl) LoadFromFile(filename string) (*ConfigData, error) {
	cl.logger.Debug("Loading configuration from file", "filename", filename)

	// Проверяем существование файла
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, os.ErrNotExist
	}

	// Определяем формат по расширению файла
	format := cl.detectFormat(filename)
	if format == "" {
		return nil, fmt.Errorf("unsupported file format: %s", filename)
	}

	// Читаем файл
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Загружаем конфигурацию
	configData, err := cl.LoadFromBytes(data, format)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration from %s: %w", filename, err)
	}

	configData.Source = filename
	cl.logger.Info("Configuration loaded successfully", "filename", filename, "format", format)

	return configData, nil
}

// LoadFromBytes загружает конфигурацию из байтов
func (cl *ConfigLoaderImpl) LoadFromBytes(data []byte, format string) (*ConfigData, error) {
	var config map[string]interface{}
	var err error

	switch strings.ToLower(format) {
	case "yaml", "yml":
		err = yaml.Unmarshal(data, &config)
	case "json":
		err = json.Unmarshal(data, &config)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", format, err)
	}

	return &ConfigData{
		Format: format,
		Data:   config,
	}, nil
}

// LoadFromReader загружает конфигурацию из Reader
func (cl *ConfigLoaderImpl) LoadFromReader(reader io.Reader, format string) (*ConfigData, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	return cl.LoadFromBytes(data, format)
}

// SupportedFormats возвращает поддерживаемые форматы
func (cl *ConfigLoaderImpl) SupportedFormats() []string {
	return []string{"yaml", "yml", "json"}
}

// detectFormat определяет формат файла по расширению
func (cl *ConfigLoaderImpl) detectFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	default:
		return ""
	}
}

// LoadMultiple загружает и объединяет несколько конфигурационных файлов
func (cl *ConfigLoaderImpl) LoadMultiple(filenames []string) (*ConfigData, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("no configuration files provided")
	}

	merged := make(map[string]interface{})
	var sources []string
	var format string

	for i, filename := range filenames {
		configData, err := cl.LoadFromFile(filename)
		if err != nil {
			// Продолжаем загрузку, если файл не найден (для опциональных файлов)
			if os.IsNotExist(err) {
				if i > 0 {
					cl.logger.Warn("Optional configuration file not found", "filename", filename)
				} else {
					cl.logger.Info("Main configuration file not found, using empty config", "filename", filename)
				}
				continue
			}
			return nil, err
		}

		// Объединяем конфигурации
		merged = cl.mergeMaps(merged, configData.Data)
		sources = append(sources, filename)

		// Используем формат первого файла
		if i == 0 {
			format = configData.Format
		}
	}

	// Handle case where no files were loaded
	if len(sources) == 0 {
		return &ConfigData{
			Source: "memory",
			Format: "yaml",
			Data:   make(map[string]interface{}),
		}, nil
	}

	return &ConfigData{
		Source: strings.Join(sources, ", "),
		Format: format,
		Data:   merged,
	}, nil
}

// mergeMaps рекурсивно объединяет два map
func (cl *ConfigLoaderImpl) mergeMaps(base, overlay map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Копируем базовые значения
	for k, v := range base {
		result[k] = v
	}

	// Накладываем значения из overlay
	for k, v := range overlay {
		if baseValue, exists := result[k]; exists {
			// Если оба значения - map, рекурсивно объединяем
			if baseMap, ok := baseValue.(map[string]interface{}); ok {
				if overlayMap, ok := v.(map[string]interface{}); ok {
					result[k] = cl.mergeMaps(baseMap, overlayMap)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}

// LoadModelConfigs extracts and loads model configurations from config data
func (cl *ConfigLoaderImpl) LoadModelConfigs(configData *ConfigData) (map[string]*interfaces.ModelConfig, error) {
	if configData == nil || configData.Data == nil {
		return nil, fmt.Errorf("config data is nil")
	}

	cl.logger.Debug("Loading model configurations", "source", configData.Source)

	// Extract models section from config data
	modelsData, exists := configData.Data["models"]
	if !exists {
		cl.logger.Info("No models section found in configuration")
		return make(map[string]*interfaces.ModelConfig), nil
	}

	// Convert to map[string]interface{}
	modelsMap, ok := modelsData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("models section must be a map/object")
	}

	modelConfigs := make(map[string]*interfaces.ModelConfig)
	var errors []error

	// Process each model configuration
	for name, modelData := range modelsMap {
		cl.logger.Debug("Processing model configuration", "name", name)

		modelConfig, err := cl.parseModelConfig(name, modelData)
		if err != nil {
			cl.logger.Error("Failed to parse model configuration",
				"name", name,
				"error", err)
			errors = append(errors, fmt.Errorf("failed to parse model %s: %w", name, err))
			continue
		}

		// Perform environment variable substitution
		if err := cl.substituteEnvVars(modelConfig); err != nil {
			cl.logger.Error("Failed to substitute environment variables for model",
				"name", name,
				"error", err)
			errors = append(errors, fmt.Errorf("failed to substitute env vars for model %s: %w", name, err))
			continue
		}

		modelConfigs[name] = modelConfig
		cl.logger.Info("Model configuration loaded successfully",
			"name", name,
			"provider", modelConfig.Provider,
			"model_name", modelConfig.ModelName)
	}

	if len(errors) > 0 {
		return modelConfigs, fmt.Errorf("some model configurations failed to load: %v", errors)
	}

	cl.logger.Info("Model configurations loaded successfully",
		"count", len(modelConfigs),
		"source", configData.Source)

	return modelConfigs, nil
}

// parseModelConfig parses a single model configuration
func (cl *ConfigLoaderImpl) parseModelConfig(name string, data interface{}) (*interfaces.ModelConfig, error) {
	// Convert to map[string]interface{}
	modelData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("model configuration must be a map/object")
	}

	config := &interfaces.ModelConfig{}

	// Parse required fields
	if provider, ok := modelData["provider"].(string); ok {
		config.Provider = provider
	} else {
		return nil, fmt.Errorf("provider is required and must be a string")
	}

	if modelName, ok := modelData["model_name"].(string); ok {
		config.ModelName = modelName
	} else {
		return nil, fmt.Errorf("model_name is required and must be a string")
	}

	if apiKey, ok := modelData["api_key"].(string); ok {
		config.APIKey = apiKey
	} else {
		return nil, fmt.Errorf("api_key is required and must be a string")
	}

	// Parse optional fields with defaults
	if baseURL, ok := modelData["base_url"].(string); ok {
		config.BaseURL = baseURL
	}

	if maxTokens, ok := modelData["max_tokens"]; ok {
		switch v := maxTokens.(type) {
		case int:
			config.MaxTokens = v
		case float64:
			config.MaxTokens = int(v)
		default:
			return nil, fmt.Errorf("max_tokens must be a number")
		}
	} else {
		config.MaxTokens = 4000 // Default value
	}

	if temperature, ok := modelData["temperature"]; ok {
		switch v := temperature.(type) {
		case float64:
			config.Temperature = float32(v)
		case float32:
			config.Temperature = v
		case int:
			config.Temperature = float32(v)
		default:
			return nil, fmt.Errorf("temperature must be a number")
		}
	} else {
		config.Temperature = 0.7 // Default value
	}

	if timeout, ok := modelData["timeout"].(string); ok {
		config.Timeout = timeout
	} else {
		config.Timeout = "30s" // Default value
	}

	// Parse supports section
	if supportsData, ok := modelData["supports"]; ok {
		supportsMap, ok := supportsData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("supports section must be a map/object")
		}

		config.Supports = &interfaces.ModelCapabilities{
			Streaming: cl.getBoolOrDefault(supportsMap, "stream", false),
			Tools:     cl.getBoolOrDefault(supportsMap, "tools", false),
			Vision:    cl.getBoolOrDefault(supportsMap, "vision", false),
			System:    cl.getBoolOrDefault(supportsMap, "system", false),
		}
	} else {
		// Default capabilities
		config.Supports = &interfaces.ModelCapabilities{
			Streaming: true,
			Tools:     true,
			Vision:    false,
			System:    true,
		}
	}

	// Parse custom parameters
	if customParams, ok := modelData["custom_params"]; ok {
		customMap, ok := customParams.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("custom_params must be a map/object")
		}
		config.CustomParams = customMap
	}

	return config, nil
}

// getBoolOrDefault gets a boolean value from map with default
func (cl *ConfigLoaderImpl) getBoolOrDefault(m map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := m[key]; ok {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

// substituteEnvVars substitutes environment variables in configuration
func (cl *ConfigLoaderImpl) substituteEnvVars(config *interfaces.ModelConfig) error {
	// Substitute API key
	if strings.HasPrefix(config.APIKey, "${") && strings.HasSuffix(config.APIKey, "}") {
		envVar := config.APIKey[2 : len(config.APIKey)-1]
		envValue := os.Getenv(envVar)
		if envValue == "" {
			return fmt.Errorf("environment variable %s is not set", envVar)
		}
		config.APIKey = envValue
	}

	// Substitute base URL if it contains environment variables
	if config.BaseURL != "" && strings.HasPrefix(config.BaseURL, "${") && strings.HasSuffix(config.BaseURL, "}") {
		envVar := config.BaseURL[2 : len(config.BaseURL)-1]
		envValue := os.Getenv(envVar)
		if envValue == "" {
			return fmt.Errorf("environment variable %s is not set", envVar)
		}
		config.BaseURL = envValue
	}

	// Substitute environment variables in custom parameters
	if config.CustomParams != nil {
		if err := cl.substituteEnvVarsInMap(config.CustomParams); err != nil {
			return fmt.Errorf("failed to substitute env vars in custom_params: %w", err)
		}
	}

	return nil
}

// substituteEnvVarsInMap recursively substitutes environment variables in a map
func (cl *ConfigLoaderImpl) substituteEnvVarsInMap(m map[string]interface{}) error {
	for key, value := range m {
		if strValue, ok := value.(string); ok {
			if strings.HasPrefix(strValue, "${") && strings.HasSuffix(strValue, "}") {
				envVar := strValue[2 : len(strValue)-1]
				envValue := os.Getenv(envVar)
				if envValue == "" {
					return fmt.Errorf("environment variable %s is not set", envVar)
				}
				m[key] = envValue
			}
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			if err := cl.substituteEnvVarsInMap(nestedMap); err != nil {
				return err
			}
		}
	}
	return nil
}

// SaveModelConfigs saves model configurations to a file
func (cl *ConfigLoaderImpl) SaveModelConfigs(modelConfigs map[string]*interfaces.ModelConfig, filename string) error {
	if modelConfigs == nil {
		return fmt.Errorf("model configs cannot be nil")
	}

	// Convert to map for serialization
	modelsData := make(map[string]interface{})
	for name, config := range modelConfigs {
		modelsData[name] = cl.modelConfigToMap(config)
	}

	// Create full config structure
	fullConfig := map[string]interface{}{
		"models": modelsData,
	}

	// Determine format
	format := cl.detectFormat(filename)
	if format == "" {
		format = "yaml" // Default to YAML
	}

	var data []byte
	var err error

	switch format {
	case "yaml", "yml":
		data, err = yaml.Marshal(fullConfig)
	case "json":
		data, err = json.MarshalIndent(fullConfig, "", "  ")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal model configs: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	cl.logger.Info("Model configurations saved successfully",
		"filename", filename,
		"format", format,
		"count", len(modelConfigs))

	return nil
}

// modelConfigToMap converts ModelConfig to map for serialization
func (cl *ConfigLoaderImpl) modelConfigToMap(config *interfaces.ModelConfig) map[string]interface{} {
	result := map[string]interface{}{
		"provider":    config.Provider,
		"model_name":  config.ModelName,
		"api_key":     config.APIKey,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
		"timeout":     config.Timeout,
		"supports": map[string]interface{}{
			"stream": config.Supports.Streaming,
			"tools":   config.Supports.Tools,
			"vision":  config.Supports.Vision,
			"system":  config.Supports.System,
		},
	}

	if config.BaseURL != "" {
		result["base_url"] = config.BaseURL
	}

	if config.CustomParams != nil {
		result["custom_params"] = config.CustomParams
	}

	return result
}

