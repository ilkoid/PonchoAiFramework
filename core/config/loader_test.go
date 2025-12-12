package config

import (
	"os"
	"strings"
	"testing"
)

// MockLogger для тестов
type MockLogger struct {
	DebugMessages []string
	InfoMessages  []string
	WarnMessages  []string
	ErrorMessages []string
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.DebugMessages = append(m.DebugMessages, msg)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.InfoMessages = append(m.InfoMessages, msg)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.WarnMessages = append(m.WarnMessages, msg)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.ErrorMessages = append(m.ErrorMessages, msg)
}

func TestConfigLoader_LoadFromFile_YAML(t *testing.T) {
	// Создаем временный YAML файл
	yamlContent := `
models:
  deepseek:
    provider: "deepseek"
    api_key: "test-key"
    max_tokens: 4000
    temperature: 0.7

tools:
  test_tool:
    enabled: true
    timeout: "30s"
`

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Тестируем загрузку
	logger := &MockLogger{}
	loader := NewConfigLoader(logger)

	configData, err := loader.LoadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Проверяем базовые поля
	if configData.Source != tmpFile.Name() {
		t.Errorf("Expected source %s, got %s", tmpFile.Name(), configData.Source)
	}

	if configData.Format != "yaml" {
		t.Errorf("Expected format yaml, got %s", configData.Format)
	}

	// Проверяем содержимое
	models, ok := configData.Data["models"].(map[string]interface{})
	if !ok {
		t.Fatal("Models section not found or not a map")
	}

	deepseek, ok := models["deepseek"].(map[string]interface{})
	if !ok {
		t.Fatal("Deepseek model not found")
	}

	if deepseek["provider"] != "deepseek" {
		t.Errorf("Expected provider deepseek, got %v", deepseek["provider"])
	}

	if deepseek["max_tokens"] != 4000 {
		t.Errorf("Expected max_tokens 4000, got %v", deepseek["max_tokens"])
	}
}

func TestConfigLoader_LoadFromFile_JSON(t *testing.T) {
	jsonContent := `{
		"models": {
			"glm": {
				"provider": "zai",
				"api_key": "json-key",
				"supports": {
					"vision": true,
					"tools": false
				}
			}
		}
	}`

	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	logger := &MockLogger{}
	loader := NewConfigLoader(logger)

	configData, err := loader.LoadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if configData.Format != "json" {
		t.Errorf("Expected format json, got %s", configData.Format)
	}

	// Проверяем вложенные структуры
	models := configData.Data["models"].(map[string]interface{})
	glm := models["glm"].(map[string]interface{})
	supports := glm["supports"].(map[string]interface{})

	if supports["vision"] != true {
		t.Errorf("Expected vision true, got %v", supports["vision"])
	}
}

func TestConfigLoader_LoadFromFile_NotFound(t *testing.T) {
	logger := &MockLogger{}
	loader := NewConfigLoader(logger)

	_, err := loader.LoadFromFile("nonexistent_file.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}

	expectedMsg := "file does not exist"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConfigLoader_LoadFromFile_InvalidFormat(t *testing.T) {
	// Создаем файл с неподдерживаемым расширением
	tmpFile, err := os.CreateTemp("", "test_config_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := &MockLogger{}
	loader := NewConfigLoader(logger)

	_, err = loader.LoadFromFile(tmpFile.Name())
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}

	expectedMsg := "unsupported file format"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConfigLoader_LoadFromBytes_InvalidYAML(t *testing.T) {
	logger := &MockLogger{}
	loader := NewConfigLoader(logger)

	// Используем YAML с табами, которые запрещены спецификацией
	invalidYAML := `
models:
  deepseek:
    provider: "deepseek"
    api_key: "test"
    max_tokens: 4000
	invalid_indent: missing_parent  # таб перед этой строкой
`

	_, err := loader.LoadFromBytes([]byte(invalidYAML), "yaml")
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}

	if !strings.Contains(err.Error(), "failed to parse yaml") {
		t.Errorf("Expected YAML parsing error, got: %v", err)
	}
}

func TestConfigLoader_LoadMultiple_Merge(t *testing.T) {
	// Первый файл
	baseContent := `
models:
  deepseek:
    provider: "deepseek"
    api_key: "base-key"
    max_tokens: 4000

tools:
  base_tool:
    enabled: true
    timeout: "30s"
`

	// Второй файл (должен перекрыть значения)
	overrideContent := `
models:
  deepseek:
    api_key: "override-key"
    temperature: 0.5

tools:
  override_tool:
    enabled: false
    timeout: "60s"
`

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
	loader := NewConfigLoader(logger)

	configData, err := loader.LoadMultiple([]string{baseFile.Name(), overrideFile.Name()})
	if err != nil {
		t.Fatalf("Failed to load multiple configs: %v", err)
	}

	// Проверяем слияние
	models := configData.Data["models"].(map[string]interface{})
	deepseek := models["deepseek"].(map[string]interface{})

	// API ключ должен быть из override файла
	if deepseek["api_key"] != "override-key" {
		t.Errorf("Expected override api_key, got %v", deepseek["api_key"])
	}

	// max_tokens должен остаться из base файла
	if deepseek["max_tokens"] != 4000 {
		t.Errorf("Expected base max_tokens, got %v", deepseek["max_tokens"])
	}

	// temperature должно быть из override файла
	if deepseek["temperature"] != 0.5 {
		t.Errorf("Expected override temperature, got %v", deepseek["temperature"])
	}

	// Проверяем, что оба инструмента присутствуют
	tools := configData.Data["tools"].(map[string]interface{})
	if _, exists := tools["base_tool"]; !exists {
		t.Error("base_tool should exist after merge")
	}
	if _, exists := tools["override_tool"]; !exists {
		t.Error("override_tool should exist after merge")
	}
}

func TestConfigLoader_LoadMultiple_OptionalFileMissing(t *testing.T) {
	// Создаем только первый файл
	content := `
models:
  deepseek:
    provider: "deepseek"
    api_key: "test-key"
`

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	logger := &MockLogger{}
	loader := NewConfigLoader(logger)

	// Второй файл не существует
	configData, err := loader.LoadMultiple([]string{tmpFile.Name(), "nonexistent.yaml"})
	if err != nil {
		t.Fatalf("Should not fail with optional missing file: %v", err)
	}

	// Проверяем, что конфигурация загружена из существующего файла
	models := configData.Data["models"].(map[string]interface{})
	if _, exists := models["deepseek"]; !exists {
		t.Error("deepseek model should exist")
	}

	// Проверяем, что было предупреждение о недостающем файле
	found := false
	for _, msg := range logger.WarnMessages {
		if strings.Contains(msg, "Optional configuration file not found") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected warning about missing optional file")
	}
}

func TestConfigLoader_SupportedFormats(t *testing.T) {
	logger := &MockLogger{}
	loader := NewConfigLoader(logger)

	formats := loader.SupportedFormats()
	expected := []string{"yaml", "yml", "json"}

	if len(formats) != len(expected) {
		t.Fatalf("Expected %d formats, got %d", len(expected), len(formats))
	}

	for i, format := range expected {
		if formats[i] != format {
			t.Errorf("Expected format %s at index %d, got %s", format, i, formats[i])
		}
	}
}

func TestConfigLoader_LoadFromReader(t *testing.T) {
	yamlContent := `
test:
  value: "from_reader"
  number: 42
`

	logger := &MockLogger{}
	loader := NewConfigLoader(logger)

	configData, err := loader.LoadFromReader(strings.NewReader(yamlContent), "yaml")
	if err != nil {
		t.Fatalf("Failed to load from reader: %v", err)
	}

	if configData.Source != "" {
		t.Errorf("Expected empty source for reader, got %s", configData.Source)
	}

	test := configData.Data["test"].(map[string]interface{})
	if test["value"] != "from_reader" {
		t.Errorf("Expected 'from_reader', got %v", test["value"])
	}

	if test["number"] != 42 {
		t.Errorf("Expected 42, got %v", test["number"])
	}
}
