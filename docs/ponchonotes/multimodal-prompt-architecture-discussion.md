# Multimodal Prompt Architecture Discussion

## Overview

Документ содержит полную переписку об архитектурных паттернах реализации мультимодальных промптов в PonchoFramework, включая анализ архитектурных решений и примеры реализации.

## Архитектурные паттерны PonchoFramework

### 1. **Центральный Оркестратор (Facade Pattern)**

**PonchoFramework** реализует паттерн Facade:
```go
type PonchoFramework struct {
    models  *PonchoModelRegistry    // Registry паттерн
    tools   *PonchoToolRegistry     // Registry паттерн  
    flows   *PonchoFlowRegistry     // Registry паттерн
    prompts *PonchoPromptManager    // Manager паттерн
    config  *PonchoFrameworkConfig  // Configuration паттерн
}
```

**Ключевые методы-оркестраторы:**
- `Generate()` - унифицированная генерация
- `ExecuteTool()` - выполнение инструментов
- `ExecuteFlow()` - оркестрация workflow

### 2. **Адаптерный паттерн для AI-моделей**

**PonchoModel** интерфейс обеспечивает унифицированный доступ:
```go
type PonchoModel interface {
    Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
    SupportsStreaming() bool
    SupportsTools() bool
    SupportsVision() bool
}
```

**Реализации адаптеров:**
- `DeepSeekModel` - адаптер для DeepSeek API
- `ZAIModel` - адаптер для Z.AI GLM моделей
- `VisionModel` - адаптер для vision-анализа

### 3. **Registry Паттерн для управления компонентами**

**PonchoModelRegistry**:
```go
type PonchoModelRegistry interface {
    Register(name string, model PonchoModel) error
    Get(name string) (PonchoModel, error)
    List() []string
}
```

Аналогично для `PonchoToolRegistry` и `PonchoFlowRegistry`.

### 4. **Strategy Паттерн для выполнения**

**PonchoTool** интерфейс:
```go
type PonchoTool interface {
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}
}
```

**Конкретные стратегии:**
- `ArticleImporterTool` - импорт данных из S3
- `BucketBrowserTool` - навигация по S3
- `WBCategoriesTool` - работа с Wildberries API

### 5. **Template Method Паттерн для Flow**

**PonchoFlow** определяет структуру выполнения:
```go
type PonchoFlow interface {
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}
}
```

### 6. **Builder Паттерн для запросов**

**PonchoModelRequest** строится через конфигурацию:
```go
type PonchoModelRequest struct {
    Model       string                 `json:"model"`
    Messages    []*PonchoMessage       `json:"messages"`
    Temperature float32                `json:"temperature"`
    MaxTokens   int                    `json:"max_tokens"`
    Tools       []*PonchoTool          `json:"tools,omitempty"`
    Media       []*PonchoMediaPart     `json:"media,omitempty"`
}
```

### 7. **Observer Паттерн для Streaming**

**PonchoStreamCallback**:
```go
type PonchoStreamCallback func(chunk *PonchoStreamChunk) error

type PonchoStreamChunk struct {
    Type      PonchoStreamChunkType  `json:"type"`
    Content   string                 `json:"content"`
    Delta     string                 `json:"delta,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Done      bool                   `json:"done"`
}
```

### 8. **Factory Паттерн для создания компонентов**

```go
// Factory методы в PonchoFramework
func (pf *PonchoFramework) RegisterModel(name string, model PonchoModel) error
func (pf *PonchoFramework) RegisterTool(name string, tool PonchoTool) error
func (pf *PonchoFramework) RegisterFlow(name string, flow PonchoFlow) error
```

### 9. **Configuration Паттерн**

**PonchoFrameworkConfig** централизует настройки:
```go
type PonchoFrameworkConfig struct {
    Models map[string]*PonchoModelConfig `json:"models"`
    Tools  map[string]*PonchoToolConfig  `json:"tools"`
    Flows  map[string]*PonchoFlowConfig  `json:"flows"`
    Prompts *PonchoPromptConfig          `json:"prompts"`
}
```

### 10. **Data Transfer Object (DTO) Паттерн**

**Структуры данных для обмена:**
- `PonchoMessage` - унифицированное сообщение
- `PonchoContentPart` - часть контента (текст, медиа, tool call)
- `PonchoMediaPart` - медиа контент для vision

## Сценарий: Реализация мультимодального промпта

### 1. **Структура промпта файла (.prompt формат)**

```yaml
# sketch_analysis.prompt
model: "glm-vision"
temperature: 0.7
maxTokens: 2000
responseFormat:
  type: "json"
  schema:
    type: "object"
    properties:
      analysis:
        type: "string"
      features:
        type: "array"
        items:
          type: "string"

---
{{role "system"}}
Ты - эксперт по анализу fashion-эскизов. Проанализируй изображение и верни структурированный ответ в JSON формате.

{{role "user"}}
Проанализируй этот эскиз одежды:
{{media url=imageUrl}}

Дополнительные требования: {{customPrompt}}

Верни ответ в формате JSON с полями:
- analysis: подробный анализ эскиза
- features: массив ключевых характеристик
```

### 2. **PonchoPromptManager - загрузка и парсинг**

```go
type PonchoPromptManager struct {
    prompts map[string]*PonchoPrompt
    cache   *PonchoCache
    loader  *PromptLoader
}

type PonchoPrompt struct {
    Name         string                 `json:"name"`
    Description  string                 `json:"description"`
    Version      string                 `json:"version"`
    Model        string                 `json:"model"`
    Category     string                 `json:"category"`
    Tags         []string               `json:"tags"`
    Config       *PonchoModelConfig     `json:"config"`
    Template     string                 `json:"template"`
    InputSchema  map[string]interface{} `json:"input_schema"`
    OutputSchema map[string]interface{} `json:"output_schema"`
    Roles        map[string]string      `json:"roles"` // system, user, model sections
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
}

// PonchoModelConfig содержит все опции модели включая topP, temperature, maxTokens
type PonchoModelConfig struct {
    Provider         string            `json:"provider"`
    ModelName        string            `json:"model_name"`
    APIKey           string            `json:"api_key"`
    Endpoint         string            `json:"endpoint"`
    Temperature      float32           `json:"temperature"`
    MaxTokens        int               `json:"max_tokens"`
    TopP             float32           `json:"top_p"`
    TopK             int               `json:"top_k"`
    FrequencyPenalty float32           `json:"frequency_penalty"`
    PresencePenalty  float32           `json:"presence_penalty"`
    Timeout          time.Duration     `json:"timeout"`
    Supports         struct {
        Vision   bool `json:"vision"`
        Tools    bool `json:"tools"`
        Stream   bool `json:"stream"`
        System   bool `json:"system"`
    } `json:"supports"`
    ResponseFormat   *PonchoResponseFormat `json:"response_format"`
    Custom          map[string]interface{} `json:"custom"`
}

type PonchoResponseFormat struct {
    Type   string                 `json:"type"`
    Schema map[string]interface{} `json:"schema,omitempty"`
}

// Как эти структуры работают вместе:
//
// 1. PonchoPrompt - основной контейнер для загруженного промпта
//    - Содержит метаданные (имя, версия, категория)
//    - Хранит конфигурацию модели в поле Config
//    - Включает шаблон для генерации запросов
//    - Определяет схемы входных/выходных данных
//
// 2. PonchoModelConfig - все опции модели
//    - Temperature (0.0-2.0): контролирует креативность ответов
//    - MaxTokens: максимальное количество токенов в ответе
//    - TopP (0.0-1.0): nucleus sampling, ограничивает выбор токенов
//    - TopK: количество наиболее вероятных токенов для выбора
//    - FrequencyPenalty: уменьшает повторение одинаковых фраз
//    - PresencePenalty: уменьшает упоминание уже использованных тем
//    - ResponseFormat: формат ответа (текст/JSON с схемой)
//
// 3. Пример использования:
//    prompt.Config.Temperature    // доступ к temperature
//    prompt.Config.MaxTokens       // доступ к maxTokens
//    prompt.Config.TopP           // доступ к topP
//    prompt.Config.ResponseFormat  // доступ к формату ответа
//
// 4. Валидация опций:
//    - Temperature: 0.0 ≤ temp ≤ 2.0
//    - TopP: 0.0 ≤ topP ≤ 1.0
//    - MaxTokens: 1 ≤ tokens ≤ моделиспецифичный лимит
//    - FrequencyPenalty: -2.0 ≤ penalty ≤ 2.0
//    - PresencePenalty: -2.0 ≤ penalty ≤ 2.0

func (pm *PonchoPromptManager) LoadPrompt(name string) (*PonchoPrompt, error) {
    // 1. Проверка кэша
    if prompt, hit := pm.cache.Get(name); hit {
        return prompt.(*PonchoPrompt), nil
    }
    
    // 2. Загрузка из файла
    filePath := filepath.Join("prompts", name+".prompt")
    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read prompt file: %w", err)
    }
    
    // 3. Парсинг YAML фронматтера и шаблона
    prompt, err := pm.parsePromptContent(content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse prompt: %w", err)
    }
    
    // 4. Кэширование
    pm.cache.Set(name, prompt, 5*time.Minute)
    
    return prompt, nil
}
```

### 3. **Построение запроса из промпта**

```go
func (pm *PonchoPromptManager) BuildRequest(prompt *PonchoPrompt, input map[string]interface{}) (*PonchoModelRequest, error) {
    // 1. Применение переменных к шаблону
    renderedTemplate, err := pm.renderTemplate(prompt.Template, input)
    if err != nil {
        return nil, fmt.Errorf("failed to render template: %w", err)
    }
    
    // 2. Построение сообщений из role секций
    messages, err := pm.buildMessagesFromRoles(renderedTemplate, input)
    if err != nil {
        return nil, fmt.Errorf("failed to build messages: %w", err)
    }
    
    // 3. Создание запроса
    request := &PonchoModelRequest{
        Model:       prompt.Model,
        Messages:    messages,
        Temperature: prompt.Config.Temperature,
        MaxTokens:   prompt.Config.MaxTokens,
        Config:      prompt.Config,
    }
    
    // 4. Добавление медиа если есть
    if imageUrl, ok := input["imageUrl"].(string); ok {
        request.Media = append(request.Media, &PonchoMediaPart{
            URL:         imageUrl,
            ContentType: "image/jpeg",
        })
    }
    
    return request, nil
}
```

### 4. **Выполнение через PonchoFramework**

```go
func (pf *PonchoFramework) ExecutePrompt(ctx context.Context, promptName string, input map[string]interface{}) (interface{}, error) {
    // 1. Загрузка промпта
    prompt, err := pf.prompts.LoadPrompt(promptName)
    if err != nil {
        return nil, fmt.Errorf("failed to load prompt: %w", err)
    }
    
    // 2. Валидация входных данных
    if err := pm.validateInput(prompt, input); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }
    
    // 3. Построение запроса
    request, err := pf.prompts.BuildRequest(prompt, input)
    if err != nil {
        return nil, fmt.Errorf("failed to build request: %w", err)
    }
    
    // 4. Выполнение генерации
    response, err := pf.Generate(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("failed to generate: %w", err)
    }
    
    // 5. Валидация и парсинг ответа
    result, err := pf.validateAndParseResponse(prompt, response)
    if err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    return result, nil
}
```

## Полный пример текстового файла промпта

```prompt
# prompts/fashion_sketch_analysis.prompt
name: "fashion_sketch_analysis"
description: "Анализ fashion-эскиза с генерацией структурированного описания"
version: "1.0.0"
model: "glm-vision"
category: "vision-analysis"
tags: ["fashion", "vision", "analysis", "json"]

# Конфигурация модели
temperature: 0.7
maxTokens: 2000
topP: 0.9
frequencyPenalty: 0.1
presencePenalty: 0.1

# Настройки ответа
responseFormat:
  type: "json"
  schema:
    type: "object"
    properties:
      analysis:
        type: "object"
        properties:
          overall_impression:
            type: "string"
          style_direction:
            type: "string"
          target_audience:
            type: "string"
      features:
        type: "array"
        items:
          type: "object"
          properties:
            feature_type:
              type: "string"
              description: "Тип характеристики (крой, ткань, детали)"
            value:
              type: "string"
            confidence:
              type: "number"
              minimum: 0
              maximum: 1
      recommendations:
        type: "array"
        items:
          type: "string"
    required: ["analysis", "features", "recommendations"]

# Схема входных данных
inputSchema:
  type: "object"
  properties:
    imageUrl:
      type: "string"
      description: "URL изображения эскиза"
    customPrompt:
      type: "string"
      description: "Дополнительные требования к анализу"
    analysisType:
      type: "string"
      enum: ["basic", "detailed", "technical"]
      default: "detailed"
    focusAreas:
      type: "array"
      items:
        type: "string"
      description: "Области для фокусировки анализа"
    brandStyle:
      type: "string"
      description: "Стиль бренда для адаптации анализа"
    targetMarket:
      type: "string"
      description: "Целевой рынок"
  required: ["imageUrl"]

---

{{role "config"}}
# Эта секция определяет параметры выполнения
model: {{model}}
temperature: {{temperature}}
maxTokens: {{maxTokens}}
topP: {{topP}}
responseFormat: {{responseFormat}}

{{role "system"}}
Ты - senior fashion-аналитик с 15-летним опытом работы в индустрии моды. 

Твоя экспертиза включает:
- Анализ технических эскизов одежды
- Определение трендов и стилей
- Понимание особенностей кроя и конструирования
- Знание текстильных материалов
- Опыт работы с маркетплейсами (Wildberries, Ozon)

Принципы анализа:
1. Будь объективным и основывайся на визуальных данных
2. Указывай уровень уверенности для каждой характеристики
3. Предоставляй практические рекомендации
4. Адаптируй анализ под целевой рынок: {{targetMarket}}
5. Стилевой ориентир: {{brandStyle}}

Тип анализа: {{analysisType}}
{{#if focusAreas}}
Фокусные области: {{join focusAreas ", "}}
{{/if}}

{{role "user"}}
Проанализируй представленный fashion-эскиз и предоставь детальную экспертизу.

{{media url=imageUrl}}

{{#if customPrompt}}
Дополнительные требования к анализу:
{{customPrompt}}
{{/if}}

{{#if focusAreas}}
Особое внимание удели следующим аспектам:
{{#each focusAreas}}
- {{this}}
{{/each}}
{{/if}}

Структура ответа:
1. **Общая оценка** - первое впечатление от эскиза
2. **Стилевое направление** - определение стиля и трендов
3. **Детальные характеристики** - крой, ткань, фурнитура, детали
4. **Рекомендации** - предложения по улучшению и позиционированию

Верни ответ строго в JSON формате согласно схеме.

{{role "model"}}
# Пример структуры ответа для reference:
{
  "analysis": {
    "overall_impression": "Современный кэжуал-стиль с элементами спорт-шика",
    "style_direction": "streetwear/casual",
    "target_audience": "молодежь 18-30 лет"
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

{{role "post-processing"}}
# Инструкции для пост-обработки результата
- Валидировать JSON согласно схеме
- Проверить полноту всех обязательных полей
- Убедиться в confidence значениях между 0 и 1
- Добавить метаданные обработки

{{role "error-handling"}}
# Обработка ошибочных ситуаций
Если изображение нечеткое или анализ невозможен:
1. Вернуть error: true в ответе
2. Указать причину в error_message
3. Предложить решение (улучшить качество изображения, изменить ракурс)

{{variables}}
# Дополнительные переменные для шаблонизации
current_date: {{date "2006-01-02"}}
analysis_version: "v2.1"
model_version: "glm-4.5v"
processing_mode: "multimodal"
```

## Объяснение всех секций и хелперов:

### **YAML Frontmatter (метаданные):**
- `name`, `description`, `version` - базовая информация
- `model`, `temperature`, `maxTokens` - параметры модели
- `responseFormat.schema` - JSON схема для валидации ответа
- `inputSchema`, `outputSchema` - схемы входных/выходных данных

### **Role-based секции:**
- `{{role "config"}}` - параметры выполнения
- `{{role "system"}}` - системные инструкции для LLM
- `{{role "user"}}` - пользовательский запрос с медиа
- `{{role "model"}}` - пример ответа (few-shot learning)
- `{{role "post-processing"}}` - инструкции для пост-обработки
- `{{role "error-handling"}}` - обработка ошибок

### **Хелперы шаблонизации:**
- `{{media url=imageUrl}}` - вставка медиа контента
- `{{#if condition}}...{{/if}}` - условные блоки
- `{{#each array}}...{{/each}}` - итерация по массивам
- `{{join array ", "}}` - объединение массива в строку
- `{{date "2006-01-02"}}` - текущая дата
- `{{variable}}` - простая подстановка переменных

### **Специальные секции:**
- `{{variables}}` - определение дополнительных переменных
- `---` - разделитель YAML и шаблона

## Ключевые архитектурные паттерны в этой реализации:

1. **Template Method** - PonchoPromptManager определяет алгоритм обработки промптов
2. **Strategy** - Разные модели обрабатывают запросы по-разному
3. **Builder** - Построение PonchoModelRequest из компонентов промпта
4. **Factory** - Создание сообщений из role-based секций
5. **Observer** - Валидация ответа согласно схеме из промпта

## Преимущества архитектуры:

1. **Модульность** - каждый компонент независим
2. **Расширяемость** - легкое добавление новых моделей/tools
3. **Тестируемость** - все интерфейсы легко мокаются
4. **Производительность** - оптимизация под конкретные use cases
5. **Совместимость** - плавная миграция с GenKit

Эта архитектура обеспечивает гибкость, переиспользование кода и строгую типизацию при работе с мультимодальными промптами.

---

*Документ создан: 11 декабря 2025 г.*
*Источник: Обсуждение архитектуры PonchoFramework*