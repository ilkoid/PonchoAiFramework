# Спецификация унифицированного формата промптов PonchoFramework

## Обзор

PonchoFramework использует унифицированный текстовый формат промптов с синтаксисом `{{...}}` для всех элементов. Формат обеспечивает обратную совместимость с существующими промптами и поддерживает расширенные возможности для фешн-индустрии.

## Ключевые принципы

1. **Единый синтаксис**: Все элементы используют `{{...}}` синтаксис
2. **Обратная совместимость**: Существующие промпты продолжают работать
3. **Расширяемость**: Поддержка conditionals, loops, helpers
4. **Fashion-specific**: Специализированные возможности для фешн-домена
5. **Multimodal**: Полная поддержка медиа-контента

## Синтаксис элементов

### 1. Метаданные промпта

```handlebars
{{metadata name="prompt_name" version="1.0.0" category="fashion" tags="vision,analysis"}}
```

**Атрибуты:**
- `name` (обязательный) - уникальное имя промпта
- `version` (обязательный) - версия в семантическом формате
- `category` (обязательный) - категория промпта
- `tags` (опциональный) - теги через запятую
- `description` (опциональный) - описание промпта
- `author` (опциональный) - автор промпта
- `created` (опциональный) - дата создания

### 2. Конфигурация модели

```handlebars
{{config temperature=0.7 max_tokens=2000 model="glm-vision" top_p=0.9}}
```

**Параметры:**
- `temperature` (0.0-2.0) - креативность ответов
- `max_tokens` (1-8000) - максимальное количество токенов
- `model` - имя модели для выполнения
- `top_p` (0.0-1.0) - nucleus sampling
- `top_k` - количество наиболее вероятных токенов
- `frequency_penalty` (-2.0-2.0) - штраф за повторы
- `presence_penalty` (-2.0-2.0) - штраф за упоминания
- `timeout` - таймаут выполнения в секундах

### 3. Ролевые секции

```handlebars
{{system}}
Ты - эксперт по анализу fashion-эскизов.
{{/system}}

{{user}}
Проанализируй этот эскиз: {{media url=image_url}}
{{/user}}

{{model}}
{
  "analysis": "Пример структурированного ответа"
}
{{/model}}
```

**Доступные роли:**
- `{{system}}...{{/system}}` - системные инструкции
- `{{user}}...{{/user}}` - пользовательский запрос
- `{{model}}...{{/model}}` - пример ответа (few-shot)
- `{{assistant}}...{{/assistant}}` - альтернатива для model
- `{{config}}...{{/config}}` - конфигурация выполнения

### 4. Медиа контент

```handlebars
{{media url="https://example.com/image.jpg" type="image"}}
{{media url=image_url type="image" width=640 height=480}}
{{media url=video_url type="video" thumbnail=thumb_url}}
```

**Атрибуты:**
- `url` (обязательный) - URL медиа ресурса
- `type` (обязательный) - тип медиа (image, video, audio)
- `width` (опциональный) - ширина для изображений
- `height` (опциональный) - высота для изображений
- `thumbnail` (опциональный) - миниатюра для видео
- `format` (опциональный) - формат файла (jpeg, png, mp4)

### 5. Переменные и шаблонизация

```handlebars
{{variable_name}}
{{product_name}}
{{analysis_type}}

{{date "2006-01-02"}}
{{timestamp}}
{{uuid}}
```

**Встроенные переменные:**
- `{{date "format"}}` - текущая дата
- `{{timestamp}}` - Unix timestamp
- `{{uuid}}` - уникальный идентификатор
- `{{model_name}}` - имя текущей модели
- `{{prompt_version}}` - версия промпта

### 6. Условные конструкции

```handlebars
{{#if include_details}}
Детальный анализ эскиза:
{{/if}}

{{#unless skip_validation}}
Валидация результатов обязательна.
{{/unless}}

{{#if_eq analysis_type "detailed"}}
Выполнить детальный анализ с характеристиками ткани.
{{else_if_eq analysis_type "basic"}}
Базовый анализ без технических деталей.
{{/if}}
```

### 7. Циклы и итерации

```handlebars
{{#each focus_areas}}
- Анализ области: {{this}}
{{/each}}

{{#each characteristics}}
- {{feature_type}}: {{value}} (уверенность: {{confidence}})
{{/each}}

{{#with product}}
Артикул: {{article_id}}
Название: {{name}}
{{/with}}
```

### 8. Fashion-specific хелперы

```handlebars
{{fashion_category category_id}}
{{wildberries_category parent_id=123}}
{{material_translation material="cotton" target="ru"}}
{{size_conversion size="M" system="eu"}}
{{season_detection month=12 hemisphere="north"}}

{{fashion_style style="casual" audience="youth"}}
{{color_analysis hex="#FF5733" format="russian"}}
{{pattern_type pattern="floral" complexity="medium"}}
```

### 9. Russian language хелперы

```handlebars
{{russian_case word="платье" case="genitive"}}
{{pluralize_russian word="платье" count=5}}
{{transliterate text="Krasnoe plat'ye" from="lat" to="cyr"}}
{{russian_currency amount=1500 currency="RUB"}}
```

## Полная структура промпта

```handlebars
{{metadata 
  name="fashion_sketch_analysis" 
  version="2.1.0" 
  category="vision-analysis" 
  tags="fashion,vision,json,russian"
  description="Анализ fashion-эскиза с генерацией структурированного описания"
  author="Poncho Team"
  created="2025-12-12"
}}

{{config 
  temperature=0.7 
  max_tokens=2000 
  model="glm-vision" 
  top_p=0.9
  frequency_penalty=0.1
  timeout=60
}}

{{system}}
Ты - senior fashion-аналитик с 15-летним опытом работы в индустрии моды.

Экспертиза:
- Анализ технических эскизов одежды
- Определение трендов и стилей
- Понимание особенностей кроя и конструирования
- Знание текстильных материалов
- Опыт работы с маркетплейсами Wildberries

Принципы анализа:
1. Объективность на основе визуальных данных
2. Указание уровня уверенности для каждой характеристики
3. Практические рекомендации
4. Адаптация под целевой рынок: {{target_market}}
5. Стилевой ориентир: {{brand_style}}
{{/system}}

{{user}}
Проанализируй представленный fashion-эскиз и предоставь детальную экспертизу.

{{media url=image_url type="image" width=640 height=480}}

{{#if custom_prompt}}
Дополнительные требования:
{{custom_prompt}}
{{/if}}

{{#if focus_areas}}
Особое внимание удели:
{{#each focus_areas}}
- {{this}}
{{/each}}
{{/if}}

Тип анализа: {{analysis_type}}
Целевой рынок: {{target_market}}
{{/user}}

{{model}}
{
  "analysis": {
    "overall_impression": "Современный кэжуал-стиль с элементами спорт-шика",
    "style_direction": "streetwear/casual",
    "target_audience": "молодежь 18-30 лет",
    "season": "весна-лето"
  },
  "features": [
    {
      "feature_type": "крой",
      "value": "оверсайз, свободный силуэт",
      "confidence": 0.9
    },
    {
      "feature_type": "ткань",
      "value": "хлопок с эластичными волокнами",
      "confidence": 0.7
    }
  ],
  "recommendations": [
    "Добавить контрастные швы для акцентирования формы",
    "Рассмотреть варианты декатировки для улучшения драпировки"
  ]
}
{{/model}}

{{#if_eq analysis_type "technical"}}
{{system}}
Дополнительные технические требования:
- Указать типы швов и их расположение
- Описать фурнитуру и крепления
- Определить сложность конструирования (1-10)
{{/system}}
{{/if_eq}}
```

## Схема валидации

### JSON Schema для валидации промптов

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "metadata": {
      "type": "object",
      "required": ["name", "version", "category"],
      "properties": {
        "name": {"type": "string", "pattern": "^[a-zA-Z0-9_]+$"},
        "version": {"type": "string", "pattern": "^\\d+\\.\\d+\\.\\d+$"},
        "category": {"type": "string", "enum": ["vision-analysis", "description", "categorization", "validation"]},
        "tags": {"type": "array", "items": {"type": "string"}},
        "description": {"type": "string"},
        "author": {"type": "string"},
        "created": {"type": "string", "format": "date"}
      }
    },
    "config": {
      "type": "object",
      "properties": {
        "temperature": {"type": "number", "minimum": 0, "maximum": 2},
        "max_tokens": {"type": "integer", "minimum": 1, "maximum": 8000},
        "model": {"type": "string"},
        "top_p": {"type": "number", "minimum": 0, "maximum": 1},
        "frequency_penalty": {"type": "number", "minimum": -2, "maximum": 2},
        "timeout": {"type": "integer", "minimum": 1}
      }
    },
    "roles": {
      "type": "object",
      "properties": {
        "system": {"type": "string"},
        "user": {"type": "string"},
        "model": {"type": "string"},
        "assistant": {"type": "string"}
      }
    },
    "variables": {
      "type": "object",
      "patternProperties": {
        "^[a-zA-Z_][a-zA-Z0-9_]*$": {}
      }
    }
  },
  "required": ["metadata"]
}
```

## Примеры миграции

### Миграция из существующего формата

**Было (старый формат):**
```handlebars
{{role "config"}}
model: zai-vision/glm-4.6v-flash
config:
  temperature: 0.7
  maxOutputTokens: 4000
{{role "system"}}
You are a precise fashion sketch analyzer.
{{role "user}}
ЗАДАЧА: Создай креативное описание модного изделия.
{{media url=photoUrl}}
```

**Стало (новый формат):**
```handlebars
{{metadata name="sketch_creative" version="1.0.0" category="vision-analysis" tags="fashion,creative"}}
{{config temperature=0.7 max_tokens=4000 model="glm-vision"}}
{{system}}
Ты - эксперт по анализу fashion-эскизов. Создавай креативные описания.
{{/system}}
{{user}}
ЗАДАЧА: Создай креативное описание модного изделия.
{{media url=photoUrl type="image"}}
{{/user}}
```

### Расширенный пример с фешн-спецификой

```handlebars
{{metadata 
  name="wildberries_product_description" 
  version="1.2.0" 
  category="description" 
  tags="wildberries,seo,russian,fashion"
}}

{{config 
  temperature=0.5 
  max_tokens=3000 
  model="deepseek-chat"
  top_p=0.8
}}

{{system}}
Ты - эксперт по написанию описаний для Wildberries с глубоким пониманием SEO и фешн-индустрии.

Правила для Wildberries:
- Максимальная длина: 5000 символов
- Ключевые слова в начале описания
- Структурированный формат с абзацами
- Русский язык без англицизмов
- SEO-оптимизация под поисковые запросы

Категория: {{wildberries_category category_id}}
{{/system}}

{{user}}
Создай SEO-оптимизированное описание для товара:

Артикул: {{article_id}}
Название: {{product_name}}
Категория: {{category_name}}
Характеристики:
{{#each characteristics}}
- {{name}}: {{value}}
{{/each}}

{{media url=main_image type="image"}}

{{#if additional_images}}
Дополнительные изображения:
{{#each additional_images}}
{{media url=this type="image"}}
{{/each}}
{{/if}}

Требования:
- Длина до 5000 символов
- Включить ключевые слова: {{join keywords ", "}}
- Учесть сезонность: {{season}}
- Целевая аудитория: {{target_audience}}
{{/user}}
```

## Рекомендации по парсингу и обработке

### Алгоритм парсинга

1. **Извлечение метаданных** - парсинг `{{metadata ...}}`
2. **Извлечение конфигурации** - парсинг `{{config ...}}`
3. **Извлечение ролевых секций** - поиск `{{role}}...{{/role}}`
4. **Обработка медиа** - поиск `{{media ...}}`
5. **Обработка переменных** - поиск `{{variable}}`
6. **Обработка conditionals** - парсинг `{{#if}}...{{/if}}`
7. **Обработка циклов** - парсинг `{{#each}}...{{/each}}`
8. **Валидация** - проверка по JSON schema

### Обработка ошибок

```go
type PromptParseError struct {
    Line    int
    Column  int
    Message string
    Context string
}

type PromptValidator struct {
    schema *jsonschema.Schema
}

func (pv *PromptValidator) Validate(prompt string) (*ValidatedPrompt, error) {
    // 1. Парсинг с отслеживанием позиций
    parsed, err := pv.parseWithPositions(prompt)
    if err != nil {
        return nil, err
    }
    
    // 2. Валидация структуры
    if err := pv.validateStructure(parsed); err != nil {
        return nil, err
    }
    
    // 3. Валидация по JSON schema
    if err := pv.validateAgainstSchema(parsed); err != nil {
        return nil, err
    }
    
    return parsed, nil
}
```

### Кэширование и оптимизация

```go
type PromptCache struct {
    templates map[string]*template.Template
    metadata  map[string]*PromptMetadata
    mutex     sync.RWMutex
}

func (pc *PromptCache) GetCompiled(name string) (*template.Template, error) {
    pc.mutex.RLock()
    if tmpl, exists := pc.templates[name]; exists {
        pc.mutex.RUnlock()
        return tmpl, nil
    }
    pc.mutex.RUnlock()
    
    // Компиляция и кэширование
    tmpl, err := pc.compile(name)
    if err != nil {
        return nil, err
    }
    
    pc.mutex.Lock()
    pc.templates[name] = tmpl
    pc.mutex.Unlock()
    
    return tmpl, nil
}
```

## Fashion-specific возможности

### Специализированные хелперы

1. **Анализ изображений**: `{{fashion_analysis image_url}}`
2. **Категоризация**: `{{wildberries_category features}}`
3. **Материалы**: `{{material_analysis material_name}}`
4. **Сезонность**: `{{season_detection product_features}}`
5. **Цвета**: `{{color_analysis hex_code}}`
6. **Размеры**: `{{size_conversion size system}}`

### Интеграция с Wildberries

```handlebars
{{wildberries_api endpoint="categories" parent_id=123}}
{{wildberries_characteristics subject_id=456}}
{{wb_seo_keywords category="платья" gender="женский"}}
```

### Russian language поддержка

```handlebars
{{russian_grammar_check text=description}}
{{fashion_terminology term="oversize" target="ru"}}
{{brand_name_transliteration name="Fashion House"}}
```

## Заключение

Унифицированный формат промптов PonchoFramework обеспечивает:

- **Простоту использования** - интуитивный `{{...}}` синтаксис
- **Мощь** - поддержка conditionals, loops, helpers
- **Специализацию** - fashion-specific возможности
- **Совместимость** - обратная совместимость с существующими промптами
- **Расширяемость** - легкое добавление новых хелперов и функций

Формат оптимизирован для фешн-индустрии и Russian language, обеспечивая максимальную эффективность для PonchoFramework.