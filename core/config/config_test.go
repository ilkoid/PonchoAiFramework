package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigManager_Load_FromFile(t *testing.T) {
	// Создаем временный конфигурационный файл
	configContent := `
models:
  deepseek:
    provider: "deepseek"
    model_name: "deepseek-chat"
    api_key: "test-key-that-is-long-enough-to-pass-validation"
    max_tokens: 4000
    temperature: 0.7
    timeout: "30s"
    supports:
      streaming: true
      tools: true
      vision: false
      system: true
      json_mode: true

tools:
  test_tool:
    enabled: true
    timeout: "30s"

logger:
  level: "info"
  format: "text"
`

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Создаем менеджер конфигурации
	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		FilePaths: []string{tmpFile.Name()},
		Logger:    logger,
	})

	// Загружаем конфигурацию
	err = manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Проверяем значения
	if manager.GetString("models.deepseek.provider") != "deepseek" {
		t.Errorf("Expected provider 'deepseek', got '%s'", manager.GetString("models.deepseek.provider"))
	}

	if manager.GetInt("models.deepseek.max_tokens") != 4000 {
		t.Errorf("Expected max_tokens 4000, got %d", manager.GetInt("models.deepseek.max_tokens"))
	}

	if manager.GetFloat64("models.deepseek.temperature") != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", manager.GetFloat64("models.deepseek.temperature"))
	}

	if manager.GetBool("tools.test_tool.enabled") != true {
		t.Errorf("Expected enabled true, got %v", manager.GetBool("tools.test_tool.enabled"))
	}

	if manager.GetDuration("tools.test_tool.timeout") != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", manager.GetDuration("tools.test_tool.timeout"))
	}
}

func TestConfigManager_Load_WithEnvironmentVariables(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("PONCHO_MODELS__DEEPSEEK__API_KEY", "environment-key-that-is-long-enough")
	os.Setenv("PONCHO_MODELS__DEEPSEEK__MAX_TOKENS", "5000")
	os.Setenv("PONCHO_TOOLS__TEST_TOOL__ENABLED", "false")
	os.Setenv("PONCHO_TOOLS__TEST_TOOL__TIMEOUT", "60s")
	defer func() {
		os.Unsetenv("PONCHO_MODELS__DEEPSEEK__API_KEY")
		os.Unsetenv("PONCHO_MODELS__DEEPSEEK__MAX_TOKENS")
		os.Unsetenv("PONCHO_TOOLS__TEST_TOOL__ENABLED")
		os.Unsetenv("PONCHO_TOOLS__TEST_TOOL__TIMEOUT")
	}()

	// Базовая конфигурация
	configContent := "models:\n" +
		"  deepseek:\n" +
		"    provider: \"deepseek\"\n" +
		"    model_name: \"deepseek-chat\"\n" +
		"    api_key: \"default-key-that-is-long-enough-to-pass-validation\"\n" +
		"    max_tokens: 4000\n" +
		"    temperature: 0.7\n" +
		"    timeout: \"30s\"\n" +
		"\n" +
		"tools:\n" +
		"  test_tool:\n" +
		"    enabled: true\n" +
		"    timeout: \"30s\"\n"

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		FilePaths: []string{tmpFile.Name()},
		Logger:    logger,
		EnvPrefix: "PONCHO_",
	})

	err = manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Debug: выводим все значения из конфигурации
	t.Logf("Full config data: %+v", manager.GetConfig().Data)
	t.Logf("models.deepseek.api_key = %v (type: %T)", manager.Get("models.deepseek.api_key"), manager.Get("models.deepseek.api_key"))
	t.Logf("models.deepseek.max_tokens = %v (type: %T)", manager.Get("models.deepseek.max_tokens"), manager.Get("models.deepseek.max_tokens"))
	t.Logf("tools.test_tool.enabled = %v (type: %T)", manager.Get("tools.test_tool.enabled"), manager.Get("tools.test_tool.enabled"))
	t.Logf("tools.test_tool.timeout = %v (type: %T)", manager.Get("tools.test_tool.timeout"), manager.Get("tools.test_tool.timeout"))

	// Проверяем, что environment variables перекрывают файловые значения
	if manager.GetString("models.deepseek.api_key") != "environment-key-that-is-long-enough" {
		t.Errorf("Expected api_key from env 'environment-key-that-is-long-enough', got '%s'", manager.GetString("models.deepseek.api_key"))
	}

	if manager.GetInt("models.deepseek.max_tokens") != 5000 {
		t.Errorf("Expected max_tokens from env 5000, got %d", manager.GetInt("models.deepseek.max_tokens"))
	}

	if manager.GetBool("tools.test_tool.enabled") != false {
		t.Errorf("Expected enabled from env false, got %v", manager.GetBool("tools.test_tool.enabled"))
	}

	if manager.GetDuration("tools.test_tool.timeout") != 60*time.Second {
		t.Errorf("Expected timeout from env 60s, got %v", manager.GetDuration("tools.test_tool.timeout"))
	}
}

func TestConfigManager_Get_DefaultValues(t *testing.T) {
	// Создаем минимальную конфигурацию, чтобы пройти валидацию
	configContent := "models:\n" +
		"  deepseek:\n" +
		"    provider: \"deepseek\"\n" +
		"    model_name: \"deepseek-chat\"\n" +
		"    api_key: \"dummy-key-that-is-long-enough-to-pass-validation\"\n" +
		"    max_tokens: 4000\n" +
		"    temperature: 0.7\n" +
		"    timeout: \"30s\"\n" +
		"\n" +
		"tools:\n" +
		"  dummy_tool:\n" +
		"    enabled: true\n"
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		FilePaths: []string{tmpFile.Name()},
		Logger:    logger,
	})

	// Загружаем конфигурацию
	err = manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Проверяем значения по умолчанию для несуществующих ключей
	if manager.GetString("nonexistent.key") != "" {
		t.Errorf("Expected empty string for nonexistent key, got '%s'", manager.GetString("nonexistent.key"))
	}

	if manager.GetInt("nonexistent.key") != 0 {
		t.Errorf("Expected 0 for nonexistent int key, got %d", manager.GetInt("nonexistent.key"))
	}

	if manager.GetFloat64("nonexistent.key") != 0.0 {
		t.Errorf("Expected 0.0 for nonexistent float64 key, got %f", manager.GetFloat64("nonexistent.key"))
	}

	if manager.GetBool("nonexistent.key") != false {
		t.Errorf("Expected false for nonexistent bool key, got %v", manager.GetBool("nonexistent.key"))
	}

	if manager.GetDuration("nonexistent.key") != 0 {
		t.Errorf("Expected 0 for nonexistent duration key, got %v", manager.GetDuration("nonexistent.key"))
	}

	if manager.GetStringSlice("nonexistent.key") != nil {
		t.Errorf("Expected nil for nonexistent slice key, got %v", manager.GetStringSlice("nonexistent.key"))
	}
}

func TestConfigManager_SetAndGet(t *testing.T) {
	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		Logger: logger,
	})

	// Устанавливаем значения
	manager.Set("test.string", "test-value")
	manager.Set("test.int", 42)
	manager.Set("test.float", 3.14)
	manager.Set("test.bool", true)
	manager.Set("test.duration", "5m")
	manager.Set("test.slice", []string{"item1", "item2"})

	// Проверяем установленные значения
	if manager.GetString("test.string") != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", manager.GetString("test.string"))
	}

	if manager.GetInt("test.int") != 42 {
		t.Errorf("Expected 42, got %d", manager.GetInt("test.int"))
	}

	if manager.GetFloat64("test.float") != 3.14 {
		t.Errorf("Expected 3.14, got %f", manager.GetFloat64("test.float"))
	}

	if manager.GetBool("test.bool") != true {
		t.Errorf("Expected true, got %v", manager.GetBool("test.bool"))
	}

	if manager.GetDuration("test.duration") != 5*time.Minute {
		t.Errorf("Expected 5m, got %v", manager.GetDuration("test.duration"))
	}

	slice := manager.GetStringSlice("test.slice")
	if len(slice) != 2 || slice[0] != "item1" || slice[1] != "item2" {
		t.Errorf("Expected [item1, item2], got %v", slice)
	}
}

func TestConfigManager_Has(t *testing.T) {
	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		Logger: logger,
	})

	manager.Set("existing.key", "value")

	if !manager.Has("existing.key") {
		t.Error("Expected Has to return true for existing key")
	}

	if manager.Has("nonexistent.key") {
		t.Error("Expected Has to return false for nonexistent key")
	}
}

func TestConfigManager_GetSection(t *testing.T) {
	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		Logger: logger,
	})

	// Устанавливаем секцию
	manager.Set("models.deepseek.provider", "deepseek")
	manager.Set("models.deepseek.api_key", "test-key")
	manager.Set("models.glm.provider", "zai")

	// Получаем секцию
	models := manager.GetSection("models")
	if models == nil {
		t.Fatal("Expected models section to exist")
	}

	deepseek, ok := models["deepseek"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected deepseek to be a map")
	}

	if deepseek["provider"] != "deepseek" {
		t.Errorf("Expected provider 'deepseek', got %v", deepseek["provider"])
	}

	// Проверяем несуществующую секцию
	nonexistent := manager.GetSection("nonexistent")
	if nonexistent != nil {
		t.Error("Expected nil for nonexistent section")
	}
}

func TestConfigManager_Watch(t *testing.T) {
	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		Logger: logger,
	})

	callbackCalled := false
	var receivedConfig *ConfigData

	// Добавляем watcher
	err := manager.Watch(func(config *ConfigData) {
		callbackCalled = true
		receivedConfig = config
	})
	if err != nil {
		t.Fatalf("Failed to add watcher: %v", err)
	}

	// Изменяем конфигурацию
	manager.Set("test.key", "new-value")

	// Ждем немного для асинхронного вызова
	time.Sleep(100 * time.Millisecond)

	if !callbackCalled {
		t.Error("Expected watcher callback to be called")
	}

	if receivedConfig == nil {
		t.Error("Expected config to be passed to watcher")
	}
}

func TestConfigManager_Reload(t *testing.T) {
	configContent := "models:\n" +
		"  deepseek:\n" +
		"    provider: \"deepseek\"\n" +
		"    model_name: \"deepseek-chat\"\n" +
		"    api_key: \"dummy-key-that-is-long-enough-to-pass-validation\"\n" +
		"    max_tokens: 4000\n" +
		"    temperature: 0.7\n" +
		"    timeout: \"30s\"\n" +
		"initial:\n" +
		"  value: \"original\"\n"

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(configContent)
	tmpFile.Close()

	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		FilePaths: []string{tmpFile.Name()},
		Logger:    logger,
	})

	// Первоначальная загрузка
	err = manager.Load()
	if err != nil {
		t.Fatalf("Failed to load initial configuration: %v", err)
	}

	if manager.GetString("initial.value") != "original" {
		t.Errorf("Expected 'original', got '%s'", manager.GetString("initial.value"))
	}

	// Изменяем файл
	newContent := "models:\n" +
		"  deepseek:\n" +
		"    provider: \"deepseek\"\n" +
		"    model_name: \"deepseek-chat\"\n" +
		"    api_key: \"dummy-key-that-is-long-enough-to-pass-validation\"\n" +
		"    max_tokens: 4000\n" +
		"    temperature: 0.7\n" +
		"    timeout: \"30s\"\n" +
		"initial:\n" +
		"  value: \"updated\"\n" +
		"new:\n" +
		"  field: \"added\"\n"

	err = os.WriteFile(tmpFile.Name(), []byte(newContent), 0644)
	if err != nil {
		t.Fatalf("Failed to update config file: %v", err)
	}

	// Перезагружаем
	err = manager.Reload()
	if err != nil {
		t.Fatalf("Failed to reload configuration: %v", err)
	}

	if manager.GetString("initial.value") != "updated" {
		t.Errorf("Expected 'updated', got '%s'", manager.GetString("initial.value"))
	}

	if manager.GetString("new.field") != "added" {
		t.Errorf("Expected 'added', got '%s'", manager.GetString("new.field"))
	}
}

func TestConfigManager_Validate(t *testing.T) {
	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		Logger: logger,
	})

	// Валидная конфигурация
	manager.Set("models.test.provider", "deepseek")
	manager.Set("models.test.model_name", "test-model")
	manager.Set("models.test.api_key", "test-key-that-is-long-enough-to-pass-validation")
	manager.Set("models.test.max_tokens", 4000)
	manager.Set("models.test.temperature", 0.7)
	manager.Set("models.test.timeout", "30s")

	err := manager.Validate()
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// Невалидная конфигурация (отсутствует api_key)
	manager.Set("models.test2.provider", "deepseek")
	// api_key не установлен

	err = manager.Validate()
	if err == nil {
		t.Error("Expected validation error for missing api_key")
	}
}

func TestConfigManager_GetConfigAsStruct(t *testing.T) {
	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		Logger: logger,
	})

	// Устанавливаем тестовые данные
	manager.Set("models.deepseek.provider", "deepseek")
	manager.Set("models.deepseek.max_tokens", 4000)
	manager.Set("tools.test.enabled", true)

	// Тестовая структура
	type TestModel struct {
		Provider  string `yaml:"provider"`
		MaxTokens int    `yaml:"max_tokens"`
	}

	type TestTools struct {
		Test struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"test"`
	}

	type TestConfig struct {
		Models map[string]TestModel `yaml:"models"`
		Tools  TestTools            `yaml:"tools"`
	}

	var config TestConfig
	err := manager.GetConfigAsStruct(&config)
	if err != nil {
		t.Fatalf("Failed to convert config to struct: %v", err)
	}

	// Проверяем значения
	if config.Models["deepseek"].Provider != "deepseek" {
		t.Errorf("Expected provider 'deepseek', got '%s'", config.Models["deepseek"].Provider)
	}

	if config.Models["deepseek"].MaxTokens != 4000 {
		t.Errorf("Expected MaxTokens 4000, got %d", config.Models["deepseek"].MaxTokens)
	}

	if config.Tools.Test.Enabled != true {
		t.Errorf("Expected Enabled true, got %v", config.Tools.Test.Enabled)
	}
}

func TestConfigManager_EnvironmentVariableTypes(t *testing.T) {
	// Тестируем конвертацию различных типов из environment variables
	os.Setenv("PONCHO_TEST__STRING", "test-string")
	os.Setenv("PONCHO_TEST__INT", "123")
	os.Setenv("PONCHO_TEST__FLOAT", "3.14")
	os.Setenv("PONCHO_TEST__BOOL", "true")
	os.Setenv("PONCHO_TEST__DURATION", "5m")
	defer func() {
		os.Unsetenv("PONCHO_TEST__STRING")
		os.Unsetenv("PONCHO_TEST__INT")
		os.Unsetenv("PONCHO_TEST__FLOAT")
		os.Unsetenv("PONCHO_TEST__BOOL")
		os.Unsetenv("PONCHO_TEST__DURATION")
	}()

	// Создаем минимальную конфигурацию с секцией models
	configContent := "models:\n" +
		"  deepseek:\n" +
		"    provider: \"deepseek\"\n" +
		"    model_name: \"deepseek-chat\"\n" +
		"    api_key: \"dummy-key-that-is-long-enough-to-pass-validation\"\n" +
		"    max_tokens: 4000\n" +
		"    temperature: 0.7\n" +
		"    timeout: \"30s\"\n"
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		FilePaths: []string{tmpFile.Name()},
		Logger:    logger,
		EnvPrefix: "PONCHO_",
	})

	err = manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	if manager.GetString("test.string") != "test-string" {
		t.Errorf("Expected 'test-string', got '%s'", manager.GetString("test.string"))
	}

	if manager.GetInt("test.int") != 123 {
		t.Errorf("Expected 123, got %d", manager.GetInt("test.int"))
	}

	if manager.GetFloat64("test.float") != 3.14 {
		t.Errorf("Expected 3.14, got %f", manager.GetFloat64("test.float"))
	}

	if manager.GetBool("test.bool") != true {
		t.Errorf("Expected true, got %v", manager.GetBool("test.bool"))
	}

	if manager.GetDuration("test.duration") != 5*time.Minute {
		t.Errorf("Expected 5m, got %v", manager.GetDuration("test.duration"))
	}
}

func TestConfigManager_MultipleFilesMerge(t *testing.T) {
	// Базовый файл
	baseContent := "models:\n" +
		"  deepseek:\n" +
		"    provider: \"deepseek\"\n" +
		"    api_key: \"dummy\"\n" +
		"base:\n" +
		"  value: \"base\"\n" +
		"  override_me: \"base_value\"\n"

	// Файл перекрытия
	overrideContent := "models:\n" +
		"  deepseek:\n" +
		"    provider: \"deepseek\"\n" +
		"    model_name: \"deepseek-chat\"\n" +
		"    api_key: \"dummy-key-that-is-long-enough-to-pass-validation\"\n" +
		"    max_tokens: 4000\n" +
		"    temperature: 0.7\n" +
		"    timeout: \"30s\"\n" +
		"override:\n" +
		"  value: \"override\"\n" +
		"base:\n" +
		"  override_me: \"override_value\"\n" +
		"  new_field: \"new_value\"\n"

	baseFile, err := os.CreateTemp("", "base_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create base temp file: %v", err)
	}
	defer os.Remove(baseFile.Name())
	baseFile.WriteString(baseContent)
	baseFile.Close()

	overrideFile, err := os.CreateTemp("", "override_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create override temp file: %v", err)
	}
	defer os.Remove(overrideFile.Name())
	overrideFile.WriteString(overrideContent)
	overrideFile.Close()

	logger := &MockLogger{}
	manager := NewConfigManager(ConfigOptions{
		FilePaths: []string{baseFile.Name(), overrideFile.Name()},
		Logger:    logger,
	})

	err = manager.Load()
	if err != nil {
		t.Fatalf("Failed to load multiple files: %v", err)
	}

	// Проверяем слияние
	if manager.GetString("base.value") != "base" {
		t.Errorf("Expected 'base', got '%s'", manager.GetString("base.value"))
	}

	if manager.GetString("override.value") != "override" {
		t.Errorf("Expected 'override', got '%s'", manager.GetString("override.value"))
	}

	// override файл должен перекрыть base
	if manager.GetString("base.override_me") != "override_value" {
		t.Errorf("Expected 'override_value', got '%s'", manager.GetString("base.override_me"))
	}

	// Новое поле из override
	if manager.GetString("base.new_field") != "new_value" {
		t.Errorf("Expected 'new_value', got '%s'", manager.GetString("base.new_field"))
	}
}
