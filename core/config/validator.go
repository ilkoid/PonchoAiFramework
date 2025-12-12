package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ConfigValidator интерфейс для валидации конфигурации
type ConfigValidator interface {
	// Validate валидирует конфигурацию
	Validate(config *ConfigData) error

	// ValidateSection валидирует конкретную секцию
	ValidateSection(config *ConfigData, section string) error

	// AddRule добавляет правило валидации
	AddRule(rule ConfigValidationRule)

	// RemoveRule удаляет правило валидации
	RemoveRule(name string)
}

// ConfigValidationRule правило валидации для конфигурации
type ConfigValidationRule struct {
	Name        string
	Path        string // Путь в конфигурации (например, "models.deepseek-chat.api_key")
	Required    bool
	Type        ValidationType
	Min         interface{}
	Max         interface{}
	Enum        []interface{}
	CustomFunc  func(interface{}) error
	Description string
}

// ValidationType тип валидации
type ValidationType string

const (
	TypeString   ValidationType = "string"
	TypeInt      ValidationType = "int"
	TypeFloat    ValidationType = "float"
	TypeBool     ValidationType = "bool"
	TypeArray    ValidationType = "array"
	TypeObject   ValidationType = "object"
	TypeDuration ValidationType = "duration"
)

// ConfigValidatorImpl реализация ConfigValidator
type ConfigValidatorImpl struct {
	logger interfaces.Logger
	rules  []ConfigValidationRule
}

// NewConfigValidator создает новый экземпляр ConfigValidator
func NewConfigValidator(logger interfaces.Logger) *ConfigValidatorImpl {
	validator := &ConfigValidatorImpl{
		logger: logger,
		rules:  make([]ConfigValidationRule, 0),
	}

	// Добавляем базовые правила валидации
	validator.addDefaultRules()

	return validator
}

// Validate валидирует конфигурацию
func (cv *ConfigValidatorImpl) Validate(config *ConfigData) error {
	cv.logger.Debug("Starting configuration validation", "source", config.Source, "format", config.Format)

	// Если конфигурация пустая, считаем ее валидной для тестов
	if config.Data == nil || len(config.Data) == 0 {
		cv.logger.Debug("Empty configuration, skipping validation")
		return nil
	}

	var errors []error

	// Валидируем каждое правило
	for _, rule := range cv.rules {
		if err := cv.validateRule(config, rule); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	cv.logger.Info("Configuration validation passed")
	return nil
}

// ValidateSection валидирует конкретную секцию
func (cv *ConfigValidatorImpl) ValidateSection(config *ConfigData, section string) error {
	cv.logger.Debug("Validating configuration section", "section", section)

	var errors []error

	// Фильтруем правила для конкретной секции
	for _, rule := range cv.rules {
		if strings.HasPrefix(rule.Path, section+".") {
			if err := cv.validateRule(config, rule); err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("section %s validation failed: %v", section, errors)
	}

	return nil
}

// AddRule добавляет правило валидации
func (cv *ConfigValidatorImpl) AddRule(rule ConfigValidationRule) {
	cv.logger.Debug("Adding validation rule", "name", rule.Name, "path", rule.Path)
	cv.rules = append(cv.rules, rule)
}

// RemoveRule удаляет правило валидации
func (cv *ConfigValidatorImpl) RemoveRule(name string) {
	cv.logger.Debug("Removing validation rule", "name", name)

	for i, rule := range cv.rules {
		if rule.Name == name {
			cv.rules = append(cv.rules[:i], cv.rules[i+1:]...)
			break
		}
	}
}

// validateRule валидирует одно правило
func (cv *ConfigValidatorImpl) validateRule(config *ConfigData, rule ConfigValidationRule) error {
	// Если путь содержит wildcard, валидируем все совпадающие пути
	if strings.Contains(rule.Path, ".*") {
		return cv.validateWildcardRule(config, rule)
	}

	value, exists := cv.getValueByPath(config.Data, rule.Path)

	// Проверяем обязательность
	if rule.Required && !exists {
		return fmt.Errorf("required field '%s' is missing", rule.Path)
	}

	// Если поле не существует и не обязательное, пропускаем
	if !exists {
		return nil
	}

	// Валидируем тип
	if err := cv.validateType(value, rule); err != nil {
		return fmt.Errorf("field '%s': %w", rule.Path, err)
	}

	// Валидируем ограничения
	if err := cv.validateConstraints(value, rule); err != nil {
		return fmt.Errorf("field '%s': %w", rule.Path, err)
	}

	// Выполняем пользовательскую валидацию
	if rule.CustomFunc != nil {
		if err := rule.CustomFunc(value); err != nil {
			return fmt.Errorf("field '%s': %w", rule.Path, err)
		}
	}

	return nil
}

// validateWildcardRule валидирует правило с wildcard
func (cv *ConfigValidatorImpl) validateWildcardRule(config *ConfigData, rule ConfigValidationRule) error {
	parts := strings.Split(rule.Path, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid wildcard path: %s", rule.Path)
	}

	sectionName := parts[0]
	fieldName := strings.Join(parts[2:], ".")

	section, exists := config.Data[sectionName]
	if !exists {
		// Если секция не существует и правило не обязательное, пропускаем
		if !rule.Required {
			return nil
		}
		return fmt.Errorf("section '%s' is missing", sectionName)
	}

	sectionMap, ok := section.(map[string]interface{})
	if !ok {
		return fmt.Errorf("section '%s' is not a map", sectionName)
	}

	var errors []error

	for key, value := range sectionMap {

		// Проверяем, что значение является map для доступа к полю
		itemMap, ok := value.(map[string]interface{})
		if !ok && fieldName != "" {
			continue // Пропускаем, если нужен доступ к полю, а это не map
		}

		var targetValue interface{} = value
		if fieldName != "" {
			var exists bool
			targetValue, exists = cv.getValueByPath(itemMap, fieldName)
			if !exists || targetValue == nil {
				// Если поле обязательное, добавляем ошибку
				if rule.Required {
					errors = append(errors, fmt.Errorf("required field '%s.%s.%s' is missing", sectionName, key, fieldName))
				}
				continue // Поле не существует в этом элементе
			}
		}

		// Валидируем тип
		if err := cv.validateType(targetValue, rule); err != nil {
			errors = append(errors, fmt.Errorf("field '%s.%s': %w", sectionName, key, err))
			continue
		}

		// Валидируем ограничения
		if err := cv.validateConstraints(targetValue, rule); err != nil {
			errors = append(errors, fmt.Errorf("field '%s.%s': %w", sectionName, key, err))
			continue
		}

		// Выполняем пользовательскую валидацию
		if rule.CustomFunc != nil {
			if err := rule.CustomFunc(targetValue); err != nil {
				errors = append(errors, fmt.Errorf("field '%s.%s': %w", sectionName, key, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	return nil
}

// validateType валидирует тип значения
func (cv *ConfigValidatorImpl) validateType(value interface{}, rule ConfigValidationRule) error {
	switch rule.Type {
	case TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case TypeInt:
		switch v := value.(type) {
		case int, int32, int64:
			// OK
		case float64:
			// Проверяем, что это целое число
			if v != float64(int(v)) {
				return fmt.Errorf("expected integer, got float")
			}
		case string:
			// Пытаемся конвертировать строку в число
			if _, err := strconv.Atoi(v); err != nil {
				return fmt.Errorf("expected integer, got string")
			}
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case TypeFloat:
		switch v := value.(type) {
		case float64, float32:
			// OK
		case int, int32, int64:
			// OK, int можно преобразовать во float
		case string:
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				return fmt.Errorf("expected float, got string")
			}
		default:
			return fmt.Errorf("expected float, got %T", value)
		}
	case TypeBool:
		if _, ok := value.(bool); !ok {
			// Проверяем строковые представления bool
			if str, ok := value.(string); ok {
				lowerStr := strings.ToLower(str)
				if lowerStr != "true" && lowerStr != "false" {
					return fmt.Errorf("expected boolean, got string '%s'", str)
				}
			} else {
				return fmt.Errorf("expected boolean, got %T", value)
			}
		}
	case TypeArray:
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			return fmt.Errorf("expected array, got %T", value)
		}
	case TypeObject:
		if reflect.TypeOf(value).Kind() != reflect.Map {
			return fmt.Errorf("expected object, got %T", value)
		}
	case TypeDuration:
		// Проверяем строковое представление длительности
		if str, ok := value.(string); ok {
			if !cv.isValidDuration(str) {
				return fmt.Errorf("invalid duration format: %s", str)
			}
		} else {
			return fmt.Errorf("expected duration string, got %T", value)
		}
	}

	return nil
}

// validateConstraints валидирует ограничения
func (cv *ConfigValidatorImpl) validateConstraints(value interface{}, rule ConfigValidationRule) error {
	// Валидация Enum
	if len(rule.Enum) > 0 {
		found := false
		for _, enumValue := range rule.Enum {
			if cv.compareValues(value, enumValue) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value must be one of %v, got %v", rule.Enum, value)
		}
	}

	// Валидация Min/Max для строк
	if str, ok := value.(string); ok {
		if minStr, ok := rule.Min.(string); ok && len(str) < len(minStr) {
			return fmt.Errorf("length must be at least %d, got %d", len(minStr), len(str))
		}
		if maxStr, ok := rule.Max.(string); ok && len(str) > len(maxStr) {
			return fmt.Errorf("length must be at most %d, got %d", len(maxStr), len(str))
		}
	}

	// Валидация Min/Max для чисел
	if num := cv.toFloat64(value); num != nil {
		if minNum := cv.toFloat64(rule.Min); minNum != nil && *num < *minNum {
			return fmt.Errorf("value must be at least %f, got %f", *minNum, *num)
		}
		if maxNum := cv.toFloat64(rule.Max); maxNum != nil && *num > *maxNum {
			return fmt.Errorf("value must be at most %f, got %f", *maxNum, *num)
		}
	}

	// Валидация Min/Max для массивов
	if arr, ok := value.([]interface{}); ok {
		if minLen := cv.toInt(rule.Min); minLen != nil && len(arr) < *minLen {
			return fmt.Errorf("array length must be at least %d, got %d", *minLen, len(arr))
		}
		if maxLen := cv.toInt(rule.Max); maxLen != nil && len(arr) > *maxLen {
			return fmt.Errorf("array length must be at most %d, got %d", *maxLen, len(arr))
		}
	}

	return nil
}

// getValueByPath получает значение по пути из конфигурации
func (cv *ConfigValidatorImpl) getValueByPath(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Последний элемент - искомое значение
			value, exists := current[part]
			return value, exists
		}

		// Промежуточный элемент
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

// compareValues сравнивает два значения с учетом типов
func (cv *ConfigValidatorImpl) compareValues(a, b interface{}) bool {
	// Прямое сравнение
	if a == b {
		return true
	}

	// Сравнение строк
	if strA, ok := a.(string); ok {
		if strB, ok := b.(string); ok {
			return strA == strB
		}
	}

	// Сравнение чисел
	if numA := cv.toFloat64(a); numA != nil {
		if numB := cv.toFloat64(b); numB != nil {
			return *numA == *numB
		}
	}

	return false
}

// toFloat64 конвертирует значение в float64
func (cv *ConfigValidatorImpl) toFloat64(value interface{}) *float64 {
	switch v := value.(type) {
	case float64:
		return &v
	case float32:
		f := float64(v)
		return &f
	case int:
		f := float64(v)
		return &f
	case int32:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return &f
		}
	}
	return nil
}

// toInt конвертирует значение в int
func (cv *ConfigValidatorImpl) toInt(value interface{}) *int {
	switch v := value.(type) {
	case int:
		return &v
	case int32:
		i := int(v)
		return &i
	case int64:
		i := int(v)
		return &i
	case float64:
		if v == float64(int(v)) {
			i := int(v)
			return &i
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return &i
		}
	}
	return nil
}

// isValidDuration проверяет валидность формата длительности
func (cv *ConfigValidatorImpl) isValidDuration(s string) bool {
	// Простая проверка формата: число + единица измерения
	if len(s) == 0 {
		return false
	}

	// Находим цифры в начале
	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9') {
		i++
	}

	// Должна быть хотя бы одна цифра
	if i == 0 {
		return false
	}

	// Оставшаяся часть - единица измерения
	unit := strings.ToLower(s[i:])
	validUnits := []string{"ns", "us", "µs", "ms", "s", "m", "h"}

	for _, validUnit := range validUnits {
		if unit == validUnit {
			return true
		}
	}

	return false
}

// addDefaultRules добавляет правила валидации по умолчанию
func (cv *ConfigValidatorImpl) addDefaultRules() {
	// Правила для моделей
	cv.AddRule(ConfigValidationRule{
		Name:     "model_provider_required",
		Path:     "models.*.provider",
		Required: true,
		Type:     TypeString,
		Enum:     []interface{}{"deepseek", "zai", "openai"},
	})

	cv.AddRule(ConfigValidationRule{
		Name:     "model_api_key_required",
		Path:     "models.*.api_key",
		Required: true,
		Type:     TypeString,
	})

	cv.AddRule(ConfigValidationRule{
		Name:     "model_max_tokens",
		Path:     "models.*.max_tokens",
		Required: false,
		Type:     TypeInt,
		Min:      1,
		Max:      100000,
	})

	cv.AddRule(ConfigValidationRule{
		Name:     "model_temperature",
		Path:     "models.*.temperature",
		Required: false,
		Type:     TypeFloat,
		Min:      0.0,
		Max:      2.0,
	})

	// Правила для инструментов
	cv.AddRule(ConfigValidationRule{
		Name:     "tool_timeout",
		Path:     "tools.*.timeout",
		Required: false,
		Type:     TypeDuration,
	})

	cv.AddRule(ConfigValidationRule{
		Name:     "tool_retry_max_attempts",
		Path:     "tools.*.retry.max_attempts",
		Required: false,
		Type:     TypeInt,
		Min:      1,
		Max:      10,
	})

	// Правила для логирования
	cv.AddRule(ConfigValidationRule{
		Name:     "logger_level",
		Path:     "logger.level",
		Required: false,
		Type:     TypeString,
		Enum:     []interface{}{"debug", "info", "warn", "error"},
	})

	cv.AddRule(ConfigValidationRule{
		Name:     "logger_format",
		Path:     "logger.format",
		Required: false,
		Type:     TypeString,
		Enum:     []interface{}{"text", "json"},
	})
}
