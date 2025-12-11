# PonchoFramework - Проектный бриф

## 1. Общее описание проекта

PonchoFramework - это кастомный AI-фреймворк, разработанный для полной замены Firebase GenKit в проекте Poncho Tools. Фреймворк специализируется на работе с фешн-индустрией и предоставляет унифицированный API для работы с AI-моделями, инструментами и workflow.

### Ключевые цели проекта:
- **Независимость от GenKit**: Полный контроль над AI-стеком и устранение зависимостей от Firebase
- **Специализация под фешн**: Оптимизация для работы с одеждой, аксессуарами и эскизами
- **Интеграция с Wildberries**: Глубокая интеграция с API крупнейшего российского маркетплейса
- **Производительность**: Улучшение скорости выполнения на 30% по сравнению с GenKit
- **Масштабируемость**: Поддержка роста бизнеса и увеличения нагрузки

## 2. Специфика фешн-индустрии

### Основные сущности:
- **Артикулы** - уникальные идентификаторы товаров
- **Эскизы** - дизайнерские зарисовки и технические рисунки
- **Изображения** - фотографии товаров в высоком разрешении
- **Спецификации** - технические характеристики товаров
- **Категории** - иерархическая классификация товаров
- **Характеристики** - атрибуты товаров (материал, размер, цвет и т.д.)

### Бизнес-процессы:
1. **Импорт данных** из PLM-систем в S3 хранилище
2. **AI-анализ** изображений и эскизов с помощью vision моделей
3. **Генерация описаний** для маркетплейсов
4. **Классификация** товаров по категориям Wildberries
5. **Валидация** характеристик товаров

### Уникальные требования:
- **Мультимодальность**: Одновременная работа с текстом, изображениями и структурированными данными
- **Vision-анализ**: Распознавание деталей одежды, материалов, стилей
- **Русский язык**: Полная поддержка русского языка для описаний и характеристик
- **Высокая точность**: Требование к качеству генерации описаний для коммерческого использования

## 3. Технические требования

### Язык и стек:
- **Основной язык**: Go 1.21+
- **AI-модели**: GLM-4.6/DeepSeek через API
- **Минимум зависимостей**: Собственная реализация ключевых компонентов
- **Кросс-платформенность**: Linux, macOS, Windows
- **Хранилище**: S3-совместимое облако + SQLite
- **Формат данных**: JSON
- **Vision-анализ**: GLM-4.6V семейство моделей для fashion-специфичного анализа изображений

### Архитектурные требования:
- **Модульность**: Независимые компоненты с четкими интерфейсами
- **Расширяемость**: Легкое добавление новых моделей и инструментов
- **Конфигурируемость**: Управление через YAML/JSON конфигурацию
- **Отказоустойчивость**: Graceful degradation при недоступности компонентов

## 4. Детальная интеграция с Wildberries

### API интеграция:

#### 4.1. Content API
```go
// Пример реализации из существующего кода
type WBClient struct {
    config *WildberriesConfig
    client *http.Client
    logger *log.Logger
}

// Получение родительских категорий
func (wb *WBClient) GetParentCategories(ctx context.Context, params *APIRequestParams) (*CategoriesResponse, error) {
    url := fmt.Sprintf("%s/content/v2/object/parent/all", wb.config.BaseURL)
    return wb.makeRequest(ctx, "GET", url, params)
}

// Получение предметов категории
func (wb *WBClient) GetSubjects(ctx context.Context, params *SubjectRequestParams) (*SubjectsResponse, error) {
    url := fmt.Sprintf("%s/content/v2/object/all", wb.config.BaseURL)
    return wb.makeRequest(ctx, "GET", url, params)
}

// Получение характеристик предмета
func (wb *WBClient) GetSubjectCharacteristics(ctx context.Context, params *APIRequestParams) (*CharacteristicsResponse, error) {
    url := fmt.Sprintf("%s/content/v2/object/charc/all", wb.config.BaseURL)
    return wb.makeRequest(ctx, "GET", url, params)
}
```

#### 4.2. Структуры данных
```go
// Категория Wildberries
type Category struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Parent int    `json:"parent"`
    Active bool   `json:"is_active"`
}

// Предмет товара
type SubjectItem struct {
    ID         int                    `json:"id"`
    Name       string                 `json:"name"`
    ObjectID   int                    `json:"object_id"`
    Active     bool                   `json:"is_active"`
    Photo      string                 `json:"photo"`
}

// Характеристика предмета
type SubjectCharacteristic struct {
    ID           int    `json:"id"`
    Name         string `json:"name"`
    Type         int    `json:"type"`
    IsRequired   bool   `json:"is_required"`
    IsMultiple   bool   `json:"is_multiple"`
    SortOrder    int    `json:"sort_order"`
}
```

#### 4.3. Интеграция в PonchoFramework
```go
// Tool для работы с категориями Wildberries
type WBCategoriesTool struct {
    *PonchoBaseTool
    wbClient *wildberries.WBClient
}

func (t *WBCategoriesTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    params := &wildberries.APIRequestParams{
        Page:    1,
        PerPage: 100,
        Locale:  "ru",
    }
    
    categories, err := t.wbClient.GetParentCategories(ctx, params)
    if err != nil {
        return nil, fmt.Errorf("failed to get categories: %w", err)
    }
    
    return map[string]interface{}{
        "categories": categories.Data,
        "total":      len(categories.Data),
    }, nil
}
```

## 5. Детальная интеграция с S3

### Архитектура S3 интеграции:

#### 5.1. S3 клиент
```go
// Основная структура S3 клиента
type S3Client struct {
    client *minio.Client
    config *config.S3Config
    logger *log.Logger
}

// Конфигурация S3
type S3Config struct {
    URL       string `json:"url"`
    Region    string `json:"region"`
    Bucket    string `json:"bucket"`
    AccessKey string `json:"access_key"`
    SecretKey string `json:"secret_key"`
    Endpoint  string `json:"endpoint,omitempty"`
    UseSSL    bool   `json:"use_ssl"`
}
```

#### 5.2. Структуры данных для товаров
```go
// Унифицированные данные продукта из S3
type S3ProductData struct {
    ArticleID              string                           `json:"article_id"`
    JSONString             string                           `json:"json_string"`
    Images                 []S3Image                        `json:"images,omitempty"`
    JSONFileName           string                           `json:"json_file_name"`
    RawContent             []byte                           `json:"-"`
    Metadata               *S3Metadata                      `json:"metadata,omitempty"`
    ImageAnalysisJSON      []map[string]interface{}         `json:"image_analysis_json,omitempty"`
    PLMCharacteristics    *S3PLMCharachteristics         `json:"plm_characteristics,omitempty"`
}

// Изображение продукта
type S3Image struct {
    Filename     string    `json:"filename"`
    Data         []byte    `json:"data"`
    ContentType  string    `json:"content_type"`
    Size         int64     `json:"size"`
    LastModified time.Time `json:"last_modified"`
    URL          string    `json:"url,omitempty"`
    LocalPath    string    `json:"local_path,omitempty"`
}
```

#### 5.3. Основные операции с S3
```go
// Загрузка данных товара с изображениями
func (s *S3Client) DownloadProductWithContent(ctx context.Context, articleID string) (*S3ProductData, error) {
    // Загрузка JSON данных
    jsonResult, err := s.ImportJSONFromArticle(ctx, articleID)
    if err != nil {
        return nil, fmt.Errorf("failed to import JSON: %w", err)
    }
    
    // Загрузка изображений
    images, err := s.DownloadAllImages(ctx, articleID)
    if err != nil {
        s.logger.Printf("Warning: Failed to download images for article %s: %v", articleID, err)
        images = []S3Image{}
    }
    
    return &S3ProductData{
        ArticleID:    articleID,
        JSONString:   jsonResult.ProductData.JSONString,
        Images:       images,
        JSONFileName:  jsonResult.ProductData.JSONFileName,
        Metadata:     jsonResult.ProductData.Metadata,
    }, nil
}

// Список артикулов в бакете
func (s *S3Client) ListArticles(ctx context.Context) (*S3ArticlesList, error) {
    objectsCh := make(chan minio.ObjectInfo)
    
    go func() {
        defer close(objectsCh)
        for object := range s.client.ListObjects(ctx, s.config.Bucket, minio.ListObjectsOptions{}) {
            objectsCh <- object
        }
    }()
    
    articles := make(map[string]*S3ArticleInfo)
    for object := range objectsCh {
        if s.shouldSkipObject(object) {
            continue
        }
        
        s.extractArticleInfo(object, articles)
    }
    
    return &S3ArticlesList{
        Articles:    s.sortArticles(articles),
        TotalCount:  len(articles),
        LastScanned: time.Now(),
        Bucket:      s.config.Bucket,
    }, nil
}
```

#### 5.4. Интеграция в PonchoFramework
```go
// Tool для импорта данных из S3
type ArticleImporterTool struct {
    *PonchoBaseTool
    s3Client       *s3.S3Client
    visionAnalyzer models.VisionAnalyzer
}

func (t *ArticleImporterTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    inputMap, ok := input.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid input type")
    }
    
    articleID, ok := inputMap["article_id"].(string)
    if !ok {
        return nil, fmt.Errorf("article_id is required")
    }
    
    // Загрузка данных из S3
    productData, err := t.s3Client.DownloadProductWithContent(ctx, articleID)
    if err != nil {
        return nil, fmt.Errorf("failed to download product: %w", err)
    }
    
    // Опциональный AI-анализ изображений
    var visionAnalysis *models.FashionAnalysis
    if includeImages, _ := inputMap["include_images"].(bool); includeImages && t.visionAnalyzer != nil {
        analysis, err := t.visionAnalyzer.ExtractProductFeatures(ctx, productData.Images)
        if err != nil {
            t.logger.Printf("Warning: Vision analysis failed: %v", err)
        } else {
            visionAnalysis = analysis
        }
    }
    
    return map[string]interface{}{
        "article_id":      articleID,
        "status":          "imported",
        "product_data":    productData,
        "vision_analysis": visionAnalysis,
    }, nil
}
```

## 6. Архитектура фреймворка

### 6.1. Core компоненты

#### PonchoFramework (Main Class)
```go
type PonchoFramework struct {
    // Core registries
    models  *PonchoModelRegistry
    tools   *PonchoToolRegistry
    flows   *PonchoFlowRegistry
    prompts *PonchoPromptManager
    
    // Configuration
    config  *PonchoFrameworkConfig
    logger  *PonchoLogger
    
    // Runtime state
    started bool
    mutex   sync.RWMutex
    metrics *PonchoMetrics
}

// Основной метод генерации - замена genkit.Generate()
func (pf *PonchoFramework) Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)

// Управление компонентами
func (pf *PonchoFramework) RegisterModel(name string, model PonchoModel) error
func (pf *PonchoFramework) RegisterTool(name string, tool PonchoTool) error
func (pf *PonchoFramework) RegisterFlow(name string, flow PonchoFlow) error
```

#### PonchoModel (Interface)
```go
type PonchoModel interface {
    Generate(ctx context.Context, req *PonchoModelRequest) (*PonchoModelResponse, error)
    GenerateStreaming(ctx context.Context, req *PonchoModelRequest, callback PonchoStreamCallback) error
    
    SupportsStreaming() bool
    SupportsTools() bool
    SupportsVision() bool
    SupportsSystemRole() bool
    
    Name() string
    Provider() string
    MaxTokens() int
    DefaultTemperature() float32
}
```

#### PonchoTool (Interface)
```go
type PonchoTool interface {
    Name() string
    Description() string
    Version() string
    
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}
    
    Validate(input interface{}) error
    
    Category() string
    Tags() []string
    Dependencies() []string
}
```

#### PonchoFlow (Interface)
```go
type PonchoFlow interface {
    Name() string
    Description() string
    Version() string
    
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    ExecuteStreaming(ctx context.Context, input interface{}, callback PonchoStreamCallback) error
    
    InputSchema() map[string]interface{}
    OutputSchema() map[string]interface{}
    
    Category() string
    Tags() []string
    Dependencies() []string
}
```

### 6.2. Архитектурные слои

```
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                      │
│  ┌─────────────┬─────────────┬─────────────────────┐ │
│  │   CLI Apps  │   Web API   │   Background Jobs   │ │
│  └─────────────┴─────────────┴─────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                PonchoFramework                         │
│  ┌─────────────┬─────────────┬─────────────────────┐ │
│  │  Models     │    Tools    │       Flows         │ │
│  │             │             │                     │ │
│  │ DeepSeek    │ S3 Import   │ Article Importer    │ │
│  │ Z.AI        │ WB API      │ Mini-Agent          │ │
│  │ Vision      │ Vision      │ Description Gen     │ │
│  │ Custom      │ Custom      │ Custom              │ │
│  └─────────────┴─────────────┴─────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│              Infrastructure Layer                      │
│  ┌─────────────┬─────────────┬─────────────────────┐ │
│  │     S3      │ Wildberries  │     Config          │ │
│  │   Storage   │    API       │   Management        │ │
│  └─────────────┴─────────────┴─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 7. Этапы реализации PonchoFramework

### Phase 1: Foundation
**Цель:** Создание базовой архитектуры и core интерфейсов

**Ключевые компоненты:**
- Core интерфейсы: `PonchoModel`, `PonchoTool`, `PonchoFlow`
- Базовые классы: `PonchoBaseModel`, `PonchoBaseTool`, `PonchoBaseFlow`
- Основные структуры данных: `PonchoModelRequest`, `PonchoModelResponse`
- Реестры: `PonchoModelRegistry`, `PonchoToolRegistry`, `PonchoFlowRegistry`
- Система конфигурации и логирования

### Phase 2: Model Integration
**Цель:** Интеграция AI-моделей GLM-4.6 и DeepSeek

**Ключевые компоненты:**
- Адаптеры для GLM-4.6 с поддержкой vision capabilities
- Интеграция DeepSeek для текстовой генерации и tools
- Единый интерфейс для работы с разными провайдерами
- Поддержка streaming и batch обработки
- Валидация и мониторинг моделей

### Phase 3: Tool System
**Цель:** Разработка инструментов для S3 и Wildberries

**Ключевые компоненты:**
- S3 инструменты: импорт данных, управление файлами, загрузка изображений
- Wildberries API: категории, характеристики, предметы
- Vision инструменты: анализ изображений, извлечение характеристик
- Валидация данных и обработка ошибок
- Кэширование результатов выполнения

### Phase 4: Flow System
**Цель:** Создание workflow для фешн-процессов

**Ключевые компоненты:**
- Орестрация сложных бизнес-процессов
- Последовательное и параллельное выполнение шагов
- Управление зависимостями между компонентами
- Обработка ошибок и retry механизмы
- Мониторинг выполнения workflow

### Phase 5: Prompt Management
**Цель:** Система управления промптами

**Ключевые компоненты:**
- Загрузка и валидация промптов из файлов
- Поддержка Handlebars templating
- Версионирование промптов
- Multimodal промпты для vision задач
- Оптимизация и кэширование

### Phase 6: Production Features
**Цель:** Продуктовые функции для production использования

**Ключевые компоненты:**
- Мониторинг производительности и метрики
- Многоуровневое кэширование
- Rate limiting и управление нагрузкой
- Graceful degradation и fault tolerance
- Оптимизация производительности

## 8. Преимущества и метрики

### 8.1. Технические преимущества перед GenKit

| Аспект | GenKit | PonchoFramework | Преимущество |
|--------|--------|-----------------|---------------|
| **Кастомизация** | Ограниченная | Полная | 100% контроль над функциональностью |
| **Производительность** | Базовая | Оптимизированная | +30% скорость выполнения |
| **Мониторинг** | Базовое | Продвинутое | Полная visibility системы |
| **Конфигурация** | Статическая | Динамическая | Hot-reload без перезапуска |
| **Кэширование** | Отсутствует | Многоуровневое | Снижение нагрузки на API |
| **Streaming** | Базовый | Расширенный | Прогресс-индикаторы |
| **Russian Support** | Ограниченная | Полная | Нативная поддержка русского |

### 8.2. Бизнес-преимущества

#### Производительность:
- **Response time**: < 2 секунд vs 3+ секунд у GenKit
- **Throughput**: 100+ запросов/секунду vs 70 у GenKit
- **Resource utilization**: 30% улучшение использования CPU/memory

#### Стоимость:
- **Infrastructure**: -20% за счет оптимизации
- **API calls**: -15% за счет интеллектуального кэширования
- **Development**: +50% скорость разработки новых фич

#### Надежность:
- **Uptime**: 99.9% vs 95% у GenKit
- **Error handling**: Graceful degradation vs cascade failures
- **Recovery**: < 5 минут MTTR vs 30 минут у GenKit

### 8.3. Метрики качества

#### Технические метрики:
- **Test Coverage**: > 90%
- **Code Quality**: Cyclomatic complexity < 10
- **Documentation**: 100% coverage public API
- **Performance**: p95 < 2 секунд

#### Бизнес-метрики:
- **Migration Success**: 100% функциональность сохранена
- **Developer Satisfaction**: > 4.5/5 по опросам
- **Incident Reduction**: 60% меньше инцидентов
- **Feature Velocity**: +40% фич в квартал

## 9. Практические примеры интеграции

### 9.1. Пример полного workflow с Wildberries и S3

```go
// Полный процесс обработки товара
func ProcessFashionItem(ctx context.Context, framework *PonchoFramework, articleID string) error {
    
    // Шаг 1: Импорт данных из S3
    importResult, err := framework.ExecuteTool(ctx, "articleImporter", map[string]interface{}{
        "article_id":     articleID,
        "include_images": true,
        "max_images":     5,
    })
    if err != nil {
        return fmt.Errorf("failed to import article: %w", err)
    }
    
    // Шаг 2: Анализ изображений с AI
    visionResult, err := framework.Generate(ctx, &PonchoModelRequest{
        Model: "glm-vision",
        Messages: []*PonchoMessage{
            {
                Role: PonchoRoleUser,
                Content: []*PonchoContentPart{
                    {
                        Type: PonchoContentTypeMedia,
                        Media: &PonchoMediaPart{
                            URL: importResult.ProductData.Images[0].URL,
                        },
                    },
                    {
                        Type: PonchoContentTypeText,
                        Text: "Проанализируй это изделие одежды: определи тип, материал, стиль, сезон",
                    },
                },
            },
        },
    })
    if err != nil {
        return fmt.Errorf("vision analysis failed: %w", err)
    }
    
    // Шаг 3: Определение категории Wildberries
    categoryResult, err := framework.ExecuteTool(ctx, "wbCategories", map[string]interface{}{
        "search_query": visionResult.Message.Content[0].Text,
    })
    if err != nil {
        return fmt.Errorf("category detection failed: %w", err)
    }
    
    // Шаг 4: Генерация описания для маркетплейса
    descriptionResult, err := framework.Generate(ctx, &PonchoModelRequest{
        Model: "deepseek-chat",
        Messages: []*PonchoMessage{
            {
                Role: PonchoRoleSystem,
                Content: []*PonchoContentPart{
                    {
                        Type: PonchoContentTypeText,
                        Text: "Ты - эксперт по написанию описаний для Wildberries. Создай продающее описание с учетом SEO.",
                    },
                },
            },
            {
                Role: PonchoRoleUser,
                Content: []*PonchoContentPart{
                    {
                        Type: PonchoContentTypeText,
                        Text: fmt.Sprintf("Создай описание для товара: %s\nКатегория: %s\nХарактеристики: %v", 
                            importResult.ProductData.JSONString,
                            categoryResult.Categories[0].Name,
                            visionResult.Message.Content[0].Text),
                    },
                },
            },
        },
    })
    if err != nil {
        return fmt.Errorf("description generation failed: %w", err)
    }
    
    // Шаг 5: Сохранение результатов
    finalResult := map[string]interface{}{
        "article_id":     articleID,
        "import_data":    importResult,
        "vision_analysis": visionResult.Message.Content[0].Text,
        "category":       categoryResult.Categories[0],
        "description":    descriptionResult.Message.Content[0].Text,
        "processed_at":   time.Now(),
    }
    
    // Сохранение в S3 или базу данных
    return saveProcessingResult(ctx, finalResult)
}
```

### 9.2. Пример интеграции с существующим кодом

```go
// Адаптер для совместимости с существующим GenKit кодом
type GenKitCompatibilityAdapter struct {
    framework *PonchoFramework
}

// Замена genkit.Generate()
func (a *GenKitCompatibilityAdapter) Generate(ctx context.Context, model string, messages []string) (string, error) {
    ponchoMessages := make([]*PonchoMessage, len(messages))
    for i, msg := range messages {
        ponchoMessages[i] = &PonchoMessage{
            Role: PonchoRoleUser,
            Content: []*PonchoContentPart{
                {
                    Type: PonchoContentTypeText,
                    Text: msg,
                },
            },
        }
    }
    
    response, err := a.framework.Generate(ctx, &PonchoModelRequest{
        Model:    model,
        Messages: ponchoMessages,
    })
    
    if err != nil {
        return "", err
    }
    
    return response.Message.Content[0].Text, nil
}

// Замена genkit.DefineTool()
func (a *GenKitCompatibilityAdapter) DefineTool(name string, description string, handler func(interface{}) (interface{}, error)) {
    tool := &CompatibilityTool{
        name:        name,
        description: description,
        handler:     handler,
    }
    
    a.framework.RegisterTool(name, tool)
}

type CompatibilityTool struct {
    *PonchoBaseTool
    handler func(interface{}) (interface{}, error)
}

func (t *CompatibilityTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    return t.handler(input)
}
```

### 9.3. Пример конфигурации для production

```yaml
# config.yaml - production конфигурация
models:
  deepseek-chat:
    provider: "deepseek"
    model_name: "deepseek-chat"
    api_key: "${DEEPSEEK_API_KEY}"
    max_tokens: 4000
    temperature: 0.7
    timeout: 30s
    supports:
      vision: false
      tools: true
      stream: true
      system: true
  
  glm-vision:
    provider: "zai"
    model_name: "glm-4.5v"
    api_key: "${ZAI_API_KEY}"
    max_tokens: 2000
    temperature: 0.5
    timeout: 60s
    supports:
      vision: true
      tools: true
      stream: true
      system: true

tools:
  article_importer:
    enabled: true
    timeout: 30s
    retry:
      max_attempts: 3
      backoff: "exponential"
      base_delay: 1s
    cache:
      ttl: 300s
      max_size: 1000
  
  wb_categories:
    enabled: true
    timeout: 15s
    retry:
      max_attempts: 2
      backoff: "linear"
      base_delay: 2s

flows:
  article_processor:
    enabled: true
    timeout: 120s
    parallel: false
    dependencies:
      - "article_importer"
      - "wb_categories"
      - "glm-vision"

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  interval: 30s
  endpoint: "http://prometheus:9090"

cache:
  type: "redis"
  redis_url: "${REDIS_URL}"
  ttl: 3600s
  max_size: 10000

security:
  api_keys:
    - "${DEEPSEEK_API_KEY}"
    - "${ZAI_API_KEY}"
  rate_limiting:
    requests_per_minute: 100
  encryption:
    enabled: true
    algorithm: "AES-256-GCM"
```

## 10. Заключение

PonchoFramework представляет собой комплексное решение для замены GenKit с улучшенными характеристиками, специализированное под фешн-индустрию. Фреймворк обеспечивает:

✅ **Полную совместимость** с существующим кодом через адаптеры  
✅ **Улучшенную производительность** на 30% за счет оптимизации  
✅ **Глубокую интеграцию** с Wildberries API и S3 хранилищем  
✅ **Специализированную поддержку** фешн-домена (vision анализ, классификация)  
✅ **Масштабируемую архитектуру** для роста бизнеса  
✅ **Продвинутое кэширование** и мониторинг для production использования  

**Следующие шаги:**
1. Начать с Phase 1: Foundation implementation
2. Создать repository для PonchoFramework
3. Настроить CI/CD pipeline
4. Начать миграцию по фазам с минимальным влиянием на бизнес-процессы

