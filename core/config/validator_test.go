package config

import (
	"fmt"
	"testing"
)

func TestConfigValidator_Validate_ValidConfig(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Валидная конфигурация
	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"deepseek-chat": map[string]interface{}{
					"provider":    "deepseek",
					"model_name":  "deepseek-chat",
					"api_key":     "test-key-that-is-long-enough-to-pass-validation",
					"max_tokens":  4000,
					"temperature": 0.7,
					"timeout":     "30s",
					"supports": map[string]interface{}{
						"streaming": true,
						"tools":     true,
						"vision":    false,
						"system":    true,
						"json_mode": true,
					},
				},
			},
			"tools": map[string]interface{}{
				"test_tool": map[string]interface{}{
					"timeout": "30s",
					"retry": map[string]interface{}{
						"max_attempts": 3,
					},
				},
			},
			"logger": map[string]interface{}{
				"level":  "info",
				"format": "text",
			},
		},
	}

	err := validator.Validate(configData)
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}
}

func TestConfigValidator_Validate_MissingRequiredField(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Конфигурация с отсутствующим обязательным полем
	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"deepseek-chat": map[string]interface{}{
					"provider":   "deepseek",
					"model_name": "deepseek-chat",
					// api_key отсутствует
					"max_tokens":  4000,
					"temperature": 0.7,
					"timeout":     "30s",
				},
			},
		},
	}

	err := validator.Validate(configData)
	if err == nil {
		t.Fatal("Expected validation error for missing required field")
	}

	expectedMsg := "required field 'models.deepseek-chat.api_key' is missing"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConfigValidator_Validate_InvalidEnumValue(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Конфигурация с невалидным значением enum
	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"test-model": map[string]interface{}{
					"provider":   "invalid-provider", // Не в enum
					"api_key":    "test-key",
					"max_tokens": 4000,
				},
			},
		},
	}

	err := validator.Validate(configData)
	if err == nil {
		t.Fatal("Expected validation error for invalid enum value")
	}

	expectedMsg := "value must be one of"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConfigValidator_Validate_InvalidType(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Конфигурация с неверным типом данных
	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"test-model": map[string]interface{}{
					"provider":   "deepseek",
					"model_name": "test-model",
					"api_key":    "test-key-that-is-long-enough-to-pass-validation",
					"max_tokens": "not-a-number", // Должен быть int
				},
			},
		},
	}

	err := validator.Validate(configData)
	if err == nil {
		t.Fatal("Expected validation error for invalid type")
	}

	expectedMsg := "expected integer"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConfigValidator_Validate_OutOfRange(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Конфигурация с выходом за пределы диапазона
	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"test-model": map[string]interface{}{
					"provider":   "deepseek",
					"model_name": "test-model",
					"api_key":    "test-key-that-is-long-enough-to-pass-validation",
					"max_tokens": 200000, // > 100000
				},
			},
		},
	}

	err := validator.Validate(configData)
	if err == nil {
		t.Fatal("Expected validation error for out of range value")
	}

	expectedMsg := "value must be at most 100000"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConfigValidator_ValidateSection_ValidSection(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"test-model": map[string]interface{}{
					"provider":    "deepseek",
					"model_name":  "test-model",
					"api_key":     "test-key-that-is-long-enough-to-pass-validation",
					"max_tokens":  4000,
					"temperature": 0.7,
					"timeout":     "30s",
				},
			},
			"tools": map[string]interface{}{
				"test_tool": map[string]interface{}{
					"timeout": "30s",
				},
			},
		},
	}

	// Валидируем только секцию models
	err := validator.ValidateSection(configData, "models")
	if err != nil {
		t.Errorf("Expected valid models section to pass validation, got error: %v", err)
	}
}

func TestConfigValidator_ValidateSection_InvalidSection(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"test-model": map[string]interface{}{
					"provider": "invalid-provider", // Невалидный
					"api_key":  "test-key",
				},
			},
		},
	}

	err := validator.ValidateSection(configData, "models")
	if err == nil {
		t.Fatal("Expected validation error for invalid models section")
	}

	expectedMsg := "section models validation failed"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConfigValidator_AddCustomRule(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Добавляем custom правило
	customRule := ConfigValidationRule{
		Name:     "custom_test_rule",
		Path:     "custom.field",
		Required: true,
		Type:     TypeString,
		CustomFunc: func(value interface{}) error {
			if str, ok := value.(string); ok && str == "forbidden" {
				return fmt.Errorf("forbidden value")
			}
			return nil
		},
	}

	validator.AddRule(customRule)

	// Тестируем валидное значение
	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"deepseek": map[string]interface{}{
					"provider":   "deepseek",
					"model_name": "deepseek-chat",
					"api_key":    "test-key-that-is-long-enough-to-pass-validation",
				},
			},
			"custom": map[string]interface{}{
				"field": "valid-value",
			},
		},
	}

	err := validator.Validate(configData)
	if err != nil {
		t.Errorf("Expected valid custom field to pass validation, got error: %v", err)
	}

	// Тестируем невалидное значение (custom function)
	configData.Data["custom"].(map[string]interface{})["field"] = "forbidden"
	err = validator.Validate(configData)
	if err == nil {
		t.Fatal("Expected validation error for forbidden custom value")
	}

	expectedMsg := "forbidden value"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConfigValidator_RemoveRule(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Добавляем и удаляем правило
	rule := ConfigValidationRule{
		Name:     "test_rule_to_remove",
		Path:     "test.field",
		Required: true,
		Type:     TypeString,
	}

	validator.AddRule(rule)
	validator.RemoveRule("test_rule_to_remove")

	// Теперь конфигурация должна пройти валидацию без обязательного поля
	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"deepseek": map[string]interface{}{
					"provider":   "deepseek",
					"model_name": "deepseek-chat",
					"api_key":    "test-key-that-is-long-enough-to-pass-validation",
				},
			},
			"test": map[string]interface{}{
				// field отсутствует, но правило удалено
			},
		},
	}

	err := validator.Validate(configData)
	if err != nil {
		t.Errorf("Expected validation to pass after rule removal, got error: %v", err)
	}
}

func TestConfigValidator_Validate_Duration(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Валидная длительность
	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"deepseek": map[string]interface{}{
					"provider":   "deepseek",
					"model_name": "deepseek-chat",
					"api_key":    "test-key-that-is-long-enough-to-pass-validation",
				},
			},
			"tools": map[string]interface{}{
				"test_tool": map[string]interface{}{
					"timeout": "30s",
				},
			},
		},
	}

	err := validator.Validate(configData)
	if err != nil {
		t.Errorf("Expected valid duration to pass validation, got error: %v", err)
	}

	// Невалидная длительность
	configData.Data["tools"].(map[string]interface{})["test_tool"].(map[string]interface{})["timeout"] = "invalid-duration"

	err = validator.Validate(configData)
	if err == nil {
		t.Fatal("Expected validation error for invalid duration")
	}

	expectedMsg := "invalid duration format"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConfigValidator_Validate_Boolean(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Добавляем правило для boolean поля
	validator.AddRule(ConfigValidationRule{
		Name: "test_boolean_rule",
		Path: "test.bool_field",
		Type: TypeBool,
	})

	// Тестируем различные представления boolean
	testCases := []struct {
		name  string
		value interface{}
		valid bool
	}{
		{"boolean true", true, true},
		{"boolean false", false, true},
		{"string true", "true", true},
		{"string false", "false", true},
		{"string TRUE", "TRUE", true},
		{"string False", "False", true},
		{"invalid string", "maybe", false},
		{"number", 1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configData := &ConfigData{
				Format: "yaml",
				Data: map[string]interface{}{
					"models": map[string]interface{}{
						"deepseek": map[string]interface{}{
							"provider":   "deepseek",
							"model_name": "deepseek-chat",
							"api_key":    "test-key-that-is-long-enough-to-pass-validation",
						},
					},
					"test": map[string]interface{}{
						"bool_field": tc.value,
					},
				},
			}

			err := validator.Validate(configData)
			if tc.valid && err != nil {
				t.Errorf("Expected valid boolean value %v to pass validation, got error: %v", tc.value, err)
			}
			if !tc.valid && err == nil {
				t.Errorf("Expected invalid boolean value %v to fail validation", tc.value)
			}
		})
	}
}

func TestConfigValidator_Validate_Array(t *testing.T) {
	logger := &MockLogger{}
	validator := NewConfigValidator(logger)

	// Добавляем правило для массива
	validator.AddRule(ConfigValidationRule{
		Name: "test_array_rule",
		Path: "test.array_field",
		Type: TypeArray,
		Min:  2,
		Max:  5,
	})

	// Валидный массив
	configData := &ConfigData{
		Format: "yaml",
		Data: map[string]interface{}{
			"models": map[string]interface{}{
				"deepseek": map[string]interface{}{
					"provider":   "deepseek",
					"model_name": "deepseek-chat",
					"api_key":    "test-key-that-is-long-enough-to-pass-validation",
				},
			},
			"test": map[string]interface{}{
				"array_field": []interface{}{"item1", "item2", "item3"},
			},
		},
	}

	err := validator.Validate(configData)
	if err != nil {
		t.Errorf("Expected valid array to pass validation, got error: %v", err)
	}

	// Слишком короткий массив
	configData.Data["test"].(map[string]interface{})["array_field"] = []interface{}{"item1"}

	err = validator.Validate(configData)
	if err == nil {
		t.Fatal("Expected validation error for array too short")
	}

	expectedMsg := "array length must be at least 2"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}

	// Не массив
	configData.Data["test"].(map[string]interface{})["array_field"] = "not-an-array"

	err = validator.Validate(configData)
	if err == nil {
		t.Fatal("Expected validation error for non-array type")
	}

	expectedMsg = "expected array"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

// Вспомогательная функция
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
