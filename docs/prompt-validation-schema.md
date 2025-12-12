# Схема валидации унифицированного формата промптов

## JSON Schema для валидации промптов

### Основная схема (prompt-schema.json)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://poncho.ai/schemas/prompt-schema.json",
  "title": "PonchoFramework Prompt Schema",
  "description": "Схема валидации унифицированного формата промптов PonchoFramework",
  "type": "object",
  "properties": {
    "metadata": {
      "$ref": "#/definitions/metadata"
    },
    "config": {
      "$ref": "#/definitions/config"
    },
    "roles": {
      "$ref": "#/definitions/roles"
    },
    "variables": {
      "$ref": "#/definitions/variables"
    },
    "validation": {
      "$ref": "#/definitions/validation"
    },
    "fashion_specific": {
      "$ref": "#/definitions/fashion_specific"
    }
  },
  "required": ["metadata"],
  "additionalProperties": false,
  "definitions": {
    "metadata": {
      "type": "object",
      "title": "Prompt Metadata",
      "description": "Метаданные промпта",
      "required": ["name", "version", "category"],
      "properties": {
        "name": {
          "type": "string",
          "pattern": "^[a-zA-Z0-9_-]+$",
          "minLength": 1,
          "maxLength": 100,
          "description": "Уникальное имя промпта (только латиница, цифры, дефис, подчеркивание)"
        },
        "version": {
          "type": "string",
          "pattern": "^\\d+\\.\\d+\\.\\d+(-[a-zA-Z0-9]+)?$",
          "description": "Версия в семантическом формате (например, 1.0.0 или 2.1.0-beta)"
        },
        "category": {
          "type": "string",
          "enum": [
            "vision-analysis",
            "description-generation", 
            "categorization",
            "validation",
            "translation",
            "seo-optimization",
            "quality-check"
          ],
          "description": "Категория промпта"
        },
        "tags": {
          "type": "array",
          "items": {
            "type": "string",
            "pattern": "^[a-zA-Z0-9_-]+$",
            "minLength": 1,
            "maxLength": 50
          },
          "uniqueItems": true,
          "maxItems": 10,
          "description": "Теги для классификации и поиска"
        },
        "description": {
          "type": "string",
          "minLength": 1,
          "maxLength": 500,
          "description": "Краткое описание назначения промпта"
        },
        "author": {
          "type": "string",
          "pattern": "^[a-zA-Z0-9_\\s-]+$",
          "minLength": 1,
          "maxLength": 100,
          "description": "Автор промпта"
        },
        "created": {
          "type": "string",
          "format": "date",
          "description": "Дата создания в формате YYYY-MM-DD"
        },
        "updated": {
          "type": "string",
          "format": "date",
          "description": "Дата последнего обновления"
        },
        "deprecated": {
          "type": "boolean",
          "default": false,
          "description": "Флаг устаревания промпта"
        },
        "deprecated_version": {
          "type": "string",
          "pattern": "^\\d+\\.\\d+\\.\\d+$",
          "description": "Версия, с которой промпт считается устаревшим"
        }
      },
      "additionalProperties": false
    },
    "config": {
      "type": "object",
      "title": "Model Configuration",
      "description": "Конфигурация модели для выполнения промпта",
      "properties": {
        "model": {
          "type": "string",
          "enum": [
            "deepseek-chat",
            "glm-vision", 
            "glm-4.6v",
            "glm-4.6v-flash",
            "custom-model"
          ],
          "description": "Имя модели для выполнения"
        },
        "temperature": {
          "type": "number",
          "minimum": 0,
          "maximum": 2,
          "default": 0.7,
          "description": "Креативность ответов (0.0 - детерминированный, 2.0 - максимальная креативность)"
        },
        "max_tokens": {
          "type": "integer",
          "minimum": 1,
          "maximum": 8000,
          "default": 2000,
          "description": "Максимальное количество токенов в ответе"
        },
        "top_p": {
          "type": "number",
          "minimum": 0,
          "maximum": 1,
          "default": 0.9,
          "description": "Nucleus sampling параметр"
        },
        "top_k": {
          "type": "integer",
          "minimum": 1,
          "maximum": 100,
          "description": "Количество наиболее вероятных токенов для выбора"
        },
        "frequency_penalty": {
          "type": "number",
          "minimum": -2,
          "maximum": 2,
          "default": 0,
          "description": "Штраф за повторение одинаковых фраз"
        },
        "presence_penalty": {
          "type": "number", 
          "minimum": -2,
          "maximum": 2,
          "default": 0,
          "description": "Штраф за упоминание уже использованных тем"
        },
        "timeout": {
          "type": "integer",
          "minimum": 1,
          "maximum": 300,
          "default": 60,
          "description": "Таймаут выполнения в секундах"
        },
        "retry_attempts": {
          "type": "integer",
          "minimum": 0,
          "maximum": 5,
          "default": 3,
          "description": "Количество попыток повтора при ошибке"
        },
        "response_format": {
          "$ref": "#/definitions/response_format"
        }
      },
      "additionalProperties": false
    },
    "response_format": {
      "type": "object",
      "title": "Response Format",
      "description": "Формат ответа модели",
      "properties": {
        "type": {
          "type": "string",
          "enum": ["text", "json", "markdown"],
          "default": "text",
          "description": "Тип формата ответа"
        },
        "schema": {
          "$ref": "#/definitions/json_schema",
          "description": "JSON схема для валидации ответа (только для type=json)"
        },
        "strict": {
          "type": "boolean",
          "default": false,
          "description": "Строгое следование схеме"
        }
      },
      "required": ["type"],
      "additionalProperties": false
    },
    "json_schema": {
      "type": "object",
      "description": "JSON схема для валидации",
      "properties": {
        "type": {
          "type": "string",
          "enum": ["object", "array", "string", "number", "boolean", "null"]
        },
        "properties": {
          "type": "object",
          "additionalProperties": {
            "$ref": "#/definitions/json_schema"
          }
        },
        "items": {
          "$ref": "#/definitions/json_schema"
        },
        "required": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "enum": {
          "type": "array",
          "items": {}
        }
      }
    },
    "roles": {
      "type": "object",
      "title": "Role Sections",
      "description": "Ролевые секции промпта",
      "properties": {
        "system": {
          "type": "string",
          "minLength": 1,
          "maxLength": 10000,
          "description": "Системные инструкции для модели"
        },
        "user": {
          "type": "string", 
          "minLength": 1,
          "maxLength": 10000,
          "description": "Пользовательский запрос"
        },
        "model": {
          "type": "string",
          "minLength": 1,
          "maxLength": 10000,
          "description": "Пример ответа модели (few-shot learning)"
        },
        "assistant": {
          "type": "string",
          "minLength": 1,
          "maxLength": 10000,
          "description": "Альтернативное название для model роли"
        }
      },
      "additionalProperties": false
    },
    "variables": {
      "type": "object",
      "title": "Template Variables",
      "description": "Переменные для шаблонизации",
      "patternProperties": {
        "^[a-zA-Z_][a-zA-Z0-9_]*$": {
          "oneOf": [
            {"type": "string"},
            {"type": "number"},
            {"type": "boolean"},
            {
              "type": "array",
              "items": {}
            },
            {"type": "object"}
          ]
        }
      },
      "additionalProperties": false
    },
    "validation": {
      "type": "object",
      "title": "Validation Rules",
      "description": "Правила валидации промпта",
      "properties": {
        "required_variables": {
          "type": "array",
          "items": {
            "type": "string",
            "pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
          },
          "description": "Список обязательных переменных"
        },
        "forbidden_variables": {
          "type": "array",
          "items": {
            "type": "string",
            "pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
          },
          "description": "Список запрещенных переменных"
        },
        "max_media_count": {
          "type": "integer",
          "minimum": 0,
          "maximum": 10,
          "default": 5,
          "description": "Максимальное количество медиа элементов"
        },
        "required_roles": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": ["system", "user", "model", "assistant"]
          },
          "description": "Обязательные ролевые секции"
        },
        "max_response_length": {
          "type": "integer",
          "minimum": 1,
          "maximum": 50000,
          "description": "Максимальная длина ответа в символах"
        }
      },
      "additionalProperties": false
    },
    "fashion_specific": {
      "type": "object",
      "title": "Fashion Specific Configuration",
      "description": "Специфичные настройки для фешн-индустрии",
      "properties": {
        "supported_languages": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": ["ru", "en", "zh", "fr", "de", "it", "es"]
          },
          "default": ["ru"],
          "description": "Поддерживаемые языки"
        },
        "target_marketplace": {
          "type": "string",
          "enum": ["wildberries", "ozon", "lamoda", "generic"],
          "default": "wildberries",
          "description": "Целевой маркетплейс"
        },
        "product_categories": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": [
              "clothing",
              "shoes", 
              "accessories",
              "bags",
              "jewelry",
              "underwear",
              "sportswear"
            ]
          },
          "description": "Поддерживаемые категории товаров"
        },
        "analysis_depth": {
          "type": "string",
          "enum": ["basic", "detailed", "technical"],
          "default": "detailed",
          "description": "Глубина анализа"
        },
        "size_systems": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": ["eu", "us", "uk", "ru", "international"]
          },
          "default": ["ru", "eu"],
          "description": "Поддерживаемые системы размеров"
        },
        "color_standards": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": ["pantone", "ral", "rgb", "hex", "cmyk"]
          },
          "default": ["hex", "rgb"],
          "description": "Поддерживаемые стандарты цветов"
        }
      },
      "additionalProperties": false
    }
  }
}
```

## Extended Schema для Fashion промптов

### Fashion-specific schema (fashion-prompt-schema.json)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://poncho.ai/schemas/fashion-prompt-schema.json",
  "title": "Fashion Prompt Schema",
  "description": "Расширенная схема для фешн-специфичных промптов",
  "allOf": [
    {
      "$ref": "prompt-schema.json"
    },
    {
      "type": "object",
      "properties": {
        "fashion_metadata": {
          "$ref": "#/definitions/fashion_metadata"
        },
        "wildberries_integration": {
          "$ref": "#/definitions/wildberries_integration"
        },
        "vision_analysis": {
          "$ref": "#/definitions/vision_analysis"
        },
        "russian_language": {
          "$ref": "#/definitions/russian_language"
        }
      }
    }
  ],
  "definitions": {
    "fashion_metadata": {
      "type": "object",
      "title": "Fashion Metadata",
      "properties": {
        "product_type": {
          "type": "string",
          "enum": [
            "sketch_analysis",
            "product_description", 
            "category_classification",
            "material_identification",
            "style_analysis",
            "size_recommendation",
            "color_analysis",
            "trend_analysis"
          ],
          "description": "Тип фешн-операции"
        },
        "target_audience": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": [
              "children",
              "teenagers", 
              "young_adults",
              "adults",
              "seniors",
              "luxury",
              "mass_market",
              "sports"
            ]
          },
          "description": "Целевая аудитория"
        },
        "seasonality": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": ["spring", "summer", "autumn", "winter", "all-season"]
          },
          "description": "Сезонность"
        },
        "style_directions": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": [
              "casual",
              "business",
              "sport",
              "evening",
              "vintage",
              "streetwear",
              "minimalist",
              "romantic",
              "bohemian",
              "classic"
            ]
          },
          "description": "Стилевые направления"
        }
      },
      "additionalProperties": false
    },
    "wildberries_integration": {
      "type": "object",
      "title": "Wildberries Integration",
      "properties": {
        "api_version": {
          "type": "string",
          "pattern": "^v\\d+$",
          "default": "v2",
          "description": "Версия API Wildberries"
        },
        "category_mapping": {
          "type": "object",
          "patternProperties": {
            "^[0-9]+$": {
              "type": "string"
            }
          },
          "description": "Маппинг категорий Wildberries"
        },
        "characteristics_required": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "description": "Обязательные характеристики для Wildberries"
        },
        "seo_optimization": {
          "type": "boolean",
          "default": true,
          "description": "Включить SEO-оптимизацию для Wildberries"
        },
        "description_length": {
          "type": "object",
          "properties": {
            "min": {
              "type": "integer",
              "minimum": 50,
              "default": 200
            },
            "max": {
              "type": "integer", 
              "maximum": 5000,
              "default": 3000
            }
          },
          "description": "Ограничения длины описания"
        }
      },
      "additionalProperties": false
    },
    "vision_analysis": {
      "type": "object",
      "title": "Vision Analysis Configuration",
      "properties": {
        "image_requirements": {
          "type": "object",
          "properties": {
            "min_width": {
              "type": "integer",
              "minimum": 100,
              "default": 640
            },
            "min_height": {
              "type": "integer",
              "minimum": 100,
              "default": 480
            },
            "max_file_size": {
              "type": "integer",
              "minimum": 1024,
              "maximum": 10485760,
              "default": 5242880
            },
            "supported_formats": {
              "type": "array",
              "items": {
                "type": "string",
                "enum": ["jpeg", "jpg", "png", "webp", "bmp"]
              },
              "default": ["jpeg", "jpg", "png"]
            }
          },
          "description": "Требования к изображениям"
        },
        "analysis_features": {
          "type": "array",
          "items": {
            "type": "string",
            "enum": [
              "color_detection",
              "style_classification",
              "material_identification",
              "pattern_recognition",
              "silhouette_analysis",
              "detail_extraction",
              "quality_assessment"
            ]
          },
          "description": "Функции анализа изображений"
        },
        "confidence_threshold": {
          "type": "number",
          "minimum": 0,
          "maximum": 1,
          "default": 0.7,
          "description": "Порог уверенности для детекции"
        },
        "multiple_objects": {
          "type": "boolean",
          "default": true,
          "description": "Поддержка анализа нескольких объектов"
        }
      },
      "additionalProperties": false
    },
    "russian_language": {
      "type": "object",
      "title": "Russian Language Configuration",
      "properties": {
        "grammar_check": {
          "type": "boolean",
          "default": true,
          "description": "Включить проверку грамматики"
        },
        "terminology_standard": {
          "type": "string",
          "enum": ["gost", "industry", "custom"],
          "default": "industry",
          "description": "Стандарт терминологии"
        },
        "forbidden_words": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "description": "Запрещенные слова и англицизмы"
        },
        "required_terms": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "description": "Обязательные термины"
        },
        "character_encoding": {
          "type": "string",
          "enum": ["utf-8", "cp1251"],
          "default": "utf-8",
          "description": "Кодировка символов"
        }
      },
      "additionalProperties": false
    }
  }
}
```

## Правила валидации синтаксиса

### Регулярные выражения для парсинга

```javascript
// Основные элементы синтаксиса
const SYNTAX_PATTERNS = {
  // Метаданные
  METADATA: /\{\{\s*metadata\s+([^}]+)\s*\}\}/g,
  
  // Конфигурация
  CONFIG: /\{\{\s*config\s+([^}]+)\s*\}\}/g,
  
  // Ролевые секции
  ROLE_START: /\{\{\s*(system|user|model|assistant)\s*\}\}/g,
  ROLE_END: /\{\{\s*\/(system|user|model|assistant)\s*\}\}/g,
  
  // Медиа контент
  MEDIA: /\{\{\s*media\s+([^}]+)\s*\}\}/g,
  
  // Переменные
  VARIABLE: /\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}/g,
  
  // Условные конструкции
  IF_START: /\{\{\s*#if\s+([^}]+)\s*\}\}/g,
  IF_END: /\{\{\s*\/if\s*\}\}/g,
  UNLESS_START: /\{\{\s*#unless\s+([^}]+)\s*\}\}/g,
  UNLESS_END: /\{\{\s*\/unless\s*\}\}/g,
  
  // Циклы
  EACH_START: /\{\{\s*#each\s+([^}]+)\s*\}\}/g,
  EACH_END: /\{\{\s*\/each\s*\}\}/g,
  WITH_START: /\{\{\s*#with\s+([^}]+)\s*\}\}/g,
  WITH_END: /\{\{\s*\/with\s*\}\}/g,
  
  // Fashion-specific хелперы
  FASHION_HELPER: /\{\{\s*(fashion_|wildberries_|russian_)[^}]+\s*\}\}/g,
  
  // Встроенные функции
  BUILTIN_HELPER: /\{\{\s*(date|timestamp|uuid|model_name|prompt_version)\s*([^}]*)\s*\}\}/g
};

// Валидация атрибутов
const ATTRIBUTE_PATTERNS = {
  // Ключ=значение
  KEY_VALUE: /([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*["']([^"']+)["']/g,
  
  // Ключ без кавычек
  KEY_UNQUOTED: /([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*([a-zA-Z0-9_\-\.]+)/g,
  
  // Флаги (булевые значения)
  FLAG: /([a-zA-Z_][a-zA-Z0-9_]*)/g
};
```

### Алгоритм валидации

```go
type PromptValidator struct {
    schema           *jsonschema.Schema
    fashionSchema    *jsonschema.Schema
    syntaxPatterns   map[string]*regexp.Regexp
    allowedModels    []string
    allowedCategories []string
}

func (pv *PromptValidator) ValidatePrompt(content string) (*ValidationResult, error) {
    result := &ValidationResult{
        Valid: true,
        Errors: []ValidationError{},
        Warnings: []ValidationWarning{},
    }
    
    // 1. Синтаксическая валидация
    if err := pv.validateSyntax(content, result); err != nil {
        return result, err
    }
    
    // 2. Извлечение и валидация метаданных
    metadata, err := pv.extractAndValidateMetadata(content, result)
    if err != nil {
        return result, err
    }
    
    // 3. Валидация конфигурации
    if err := pv.validateConfig(content, result); err != nil {
        return result, err
    }
    
    // 4. Валидация ролевых секций
    if err := pv.validateRoles(content, result); err != nil {
        return result, err
    }
    
    // 5. Валидация медиа контента
    if err := pv.validateMedia(content, result); err != nil {
        return result, err
    }
    
    // 6. Fashion-specific валидация
    if pv.isFashionPrompt(metadata) {
        if err := pv.validateFashionSpecific(content, result); err != nil {
            return result, err
        }
    }
    
    // 7. Валидация по JSON schema
    promptData := pv.extractPromptData(content)
    if err := pv.validateAgainstSchema(promptData, result); err != nil {
        return result, err
    }
    
    return result, nil
}

func (pv *PromptValidator) validateSyntax(content string, result *ValidationResult) error {
    lines := strings.Split(content, "\n")
    
    for i, line := range lines {
        lineNumber := i + 1
        
        // Проверка незакрытых тегов
        if strings.Contains(line, "{{") && !strings.Contains(line, "}}") {
            result.Errors = append(result.Errors, ValidationError{
                Line:    lineNumber,
                Column:  strings.Index(line, "{{") + 1,
                Message: "Unclosed template tag",
                Context: line,
            })
            result.Valid = false
        }
        
        // Проверка несбалансированных ролевых секций
        if strings.Contains(line, "{{system}}") || strings.Contains(line, "{{user}}") || 
           strings.Contains(line, "{{model}}") || strings.Contains(line, "{{assistant}}") {
            // Проверка наличия закрывающего тега
            roleName := extractRoleName(line)
            closingTag := fmt.Sprintf("{{/%s}}", roleName)
            if !strings.Contains(content, closingTag) {
                result.Errors = append(result.Errors, ValidationError{
                    Line:    lineNumber,
                    Column:  strings.Index(line, "{{") + 1,
                    Message: fmt.Sprintf("Missing closing tag for %s role", roleName),
                    Context: line,
                })
                result.Valid = false
            }
        }
    }
    
    return nil
}
```

## Примеры валидации

### Валидный промпт

```handlebars
{{metadata name="fashion_analysis" version="1.0.0" category="vision-analysis" tags="fashion,vision"}}
{{config temperature=0.7 max_tokens=2000 model="glm-vision"}}
{{system}}
Ты - эксперт по анализу fashion-эскизов.
{{/system}}
{{user}}
Проанализируй эскиз: {{media url=image_url type="image"}}
{{/user}}
```

**Результат валидации:** ✅ Valid

### Невалидный промпт

```handlebars
{{metadata name="invalid prompt" version="1.0" category="invalid"}}
{{config temperature=3.0 max_tokens=10000}}
{{system}}
Ты - эксперт.
{{user}}
Пропущен закрывающий тег
```

**Ошибки валидации:**
1. ❌ Имя содержит пробелы
2. ❌ Версия не в семантическом формате
3. ❌ Недопустимая категория
4. ❌ Temperature вне диапазона (0-2)
5. ❌ Max_tokens превышает лимит (8000)
6. ❌ Отсутствует закрывающий тег `{{/system}}`

## Интеграция с PonchoFramework

### Использование в коде

```go
// Создание валидатора
validator, err := NewPromptValidator(
    "schemas/prompt-schema.json",
    "schemas/fashion-prompt-schema.json",
)
if err != nil {
    return err
}

// Валидация промпта
result, err := validator.ValidatePrompt(promptContent)
if err != nil {
    return fmt.Errorf("validation error: %w", err)
}

if !result.Valid {
    // Обработка ошибок
    for _, validationError := range result.Errors {
        log.Printf("Error at line %d: %s", validationError.Line, validationError.Message)
    }
    return fmt.Errorf("prompt validation failed")
}

// Предупреждения
for _, warning := range result.Warnings {
    log.Printf("Warning: %s", warning.Message)
}

// Использование валидированного промпта
prompt, err := promptManager.LoadValidatedPrompt(promptContent)
if err != nil {
    return err
}
```

Эта схема валидации обеспечивает комплексную проверку промптов PonchoFramework с учетом специфики фешн-индустрии и Russian language поддержки.