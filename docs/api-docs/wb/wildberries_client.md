# Wildberries API Client

## Обзор

Wildberries API клиент предоставляет функционал для работы с API Wildberries с фокусом на управлении товарами. Клиент реализован в соответствии с архитектурой Poncho Tools и поддерживает все необходимые функции для работы с категориями, предметами и характеристиками товаров.

## Структура пакета

```
wildberries/
├── client.go           # Основной HTTP клиент
├── types.go            # Типы данных для API
├── products.go         # Функции работы с товарами
├── validation.go       # Валидация данных
└── errors.go          # Специфичные ошибки
```

## Основные компоненты

### WBClient

Основной клиент для работы с Wildberries API.

```go
type WBClient struct {
    config     *config.WildberriesConfig
    httpClient *http.Client
    logger     *log.Logger
}
```

#### Создание клиента

```go
import (
    "github.com/ilkoid/poncho-tools/config"
    "github.com/ilkoid/poncho-tools/wildberries"
)

cfg, err := config.LoadConfig()
if err != nil {
    log.Fatal(err)
}

wbClient, err := wildberries.NewWBClient(&cfg.Wildberries, log.New(os.Stdout, "[WB] ", log.LstdFlags))
if err != nil {
    log.Fatal(err)
}
defer wbClient.Close()
```

## API функции

### 1. Работа с категориями

#### GetParentCategories

Получение списка родительских категорий товаров.

```go
func (c *WBClient) GetParentCategories(ctx context.Context, params *APIRequestParams) (*CategoriesResponse, error)
```

**Параметры:**
- `ctx` - контекст выполнения
- `params` - параметры запроса (пагинация, поиск)

**Пример использования:**

```go
params := &wildberries.APIRequestParams{
    Page:    1,
    PerPage: 10,
    Search:  "одежда", // опционально
}

response, err := wbClient.GetParentCategories(ctx, params)
if err != nil {
    log.Printf("Error: %v", err)
    return
}

fmt.Printf("Найдено категорий: %d\n", len(response.Data))
for _, category := range response.Data {
    fmt.Printf("- %s (ID: %d)\n", category.Name, category.ID)
}
```

#### GetAllParentCategories

Получение всех родительских категорий с автоматической пагинацией.

```go
func (c *WBClient) GetAllParentCategories(ctx context.Context) ([]Category, error)
```

**Пример использования:**

```go
categories, err := wbClient.GetAllParentCategories(ctx)
if err != nil {
    log.Printf("Error: %v", err)
    return
}

fmt.Printf("Всего категорий: %d\n", len(categories))
```

### 2. Работа с предметами

#### GetSubjectsByCategory

Получение списка предметов для указанной категории.

```go
func (c *WBClient) GetSubjectsByCategory(ctx context.Context, categoryID int, params *APIRequestParams) (*SubjectsResponse, error)
```

**Параметры:**
- `categoryID` - ID родительской категории
- `params` - параметры запроса

**Пример использования:**

```go
params := &wildberries.APIRequestParams{
    Page:    1,
    PerPage: 20,
    IsActive: &[]bool{true}[0], // только активные предметы
}

response, err := wbClient.GetSubjectsByCategory(ctx, 1, params) // ID категории одежды
if err != nil {
    log.Printf("Error: %v", err)
    return
}

fmt.Printf("Найдено предметов: %d\n", len(response.Data))
for _, subject := range response.Data {
    fmt.Printf("- %s (ID: %d)\n", subject.Name, subject.ID)
}
```

#### GetAllSubjectsByCategory

Получение всех предметов категории с автоматической пагинацией.

```go
func (c *WBClient) GetAllSubjectsByCategory(ctx context.Context, categoryID int) ([]Subject, error)
```

#### SearchSubjects

Поиск предметов по названию.

```go
func (c *WBClient) SearchSubjects(ctx context.Context, searchQuery string, params *APIRequestParams) (*SubjectsResponse, error)
```

**Пример использования:**

```go
params := &wildberries.APIRequestParams{
    Page:    1,
    PerPage: 10,
}

response, err := wbClient.SearchSubjects(ctx, "платье", params)
if err != nil {
    log.Printf("Error: %v", err)
    return
}

fmt.Printf("Найдено платьев: %d\n", len(response.Data))
```

### 3. Работа с характеристиками

#### GetSubjectCharacteristics

Получение характеристик предмета по его ID.

```go
func (c *WBClient) GetSubjectCharacteristics(ctx context.Context, subjectID int, params *APIRequestParams) (*SubjectCharacteristicsResponse, error)
```

**Пример использования:**

```go
params := &wildberries.APIRequestParams{
    Page:    1,
    PerPage: 50,
}

response, err := wbClient.GetSubjectCharacteristics(ctx, 12345, params)
if err != nil {
    log.Printf("Error: %v", err)
    return
}

fmt.Printf("Характеристик предмета: %d\n", len(response.Data))
for _, char := range response.Data {
    fmt.Printf("- %s (ID: %d, Type: %d, Required: %t)\n", 
        char.Name, char.ID, char.Type, char.IsRequired)
}
```

#### GetAllSubjectCharacteristics

Получение всех характеристик предмета с автоматической пагинацией.

```go
func (c *WBClient) GetAllSubjectCharacteristics(ctx context.Context, subjectID int) ([]Characteristic, error)
```

## Типы данных

### APIRequestParams

Параметры для API запросов.

```go
type APIRequestParams struct {
    Page      int    `json:"page"`        // Номер страницы (по умолчанию: 1)
    PerPage   int    `json:"per_page"`    // Элементов на странице (по умолчанию: 100, max: 1000)
    Search    string `json:"search"`       // Строка поиска (опционально)
    SortBy    string `json:"sort_by"`     // Поле сортировки (опционально)
    SortOrder string `json:"sort_order"`   // Порядок сортировки (asc/desc, опционально)
    IsActive  *bool  `json:"is_active"`   // Фильтр по активности (опционально)
    CategoryID int   `json:"category"`     // ID категории для фильтрации
}
```

### Category

Информация о категории.

```go
type Category struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    IsActive bool   `json:"is_active"`
    ParentID *int  `json:"parent_id,omitempty"`
}
```

### Subject

Информация о предмете товара.

```go
type Subject struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    IsActive bool   `json:"is_active"`
    CategoryID int  `json:"category_id"`
}
```

### Characteristic

Информация о характеристике предмета.

```go
type Characteristic struct {
    ID          int      `json:"id"`
    Name        string   `json:"name"`
    Type        int      `json:"type"`         // 0=text, 1=number, 2=boolean, 3=list, 4=multilist
    IsRequired  bool     `json:"is_required"`
    IsMultiple  bool     `json:"is_multiple"`
    Values      []string `json:"values,omitempty"` // Для list/multilist типов
}
```

## Конфигурация

### WildberriesConfig

Конфигурация для подключения к Wildberries API.

```go
type WildberriesConfig struct {
    ContentAPIKey    string // API ключ для Content API
    AnalyticsAPIKey  string // API ключ для Analytics API
    MarketplaceAPIKey string // API ключ для Marketplace API
    BaseURL         string // Базовый URL API
    Timeout         int    // Таймаут запросов в секундах
}
```

**Переменные окружения:**

```bash
# Wildberries API Keys
WB_CONTENT_API_KEY=your_content_api_key_here      # Для контентных функций (категории, предметы, характеристики)
WB_ANALYTICS_API_KEY=your_analytics_api_key_here    # Для аналитики и статистики
WB_MARKETPLACE_API_KEY=your_marketplace_api_key_here  # Для маркетплейс функций

# Wildberries API Configuration
WB_BASE_URL=https://content-api.wildberries.ru
WB_TIMEOUT=30
```

**Важно:** Для функций работы с товарами (категории, предметы, характеристики) используется `WB_CONTENT_API_KEY`.

## Обработка ошибок

Клиент использует структурированную обработку ошибок с контекстом.

### Типы ошибок

- `ErrCodeUnauthorized` - ошибка авторизации (неверный API ключ)
- `ErrCodeForbidden` - недостаточно прав доступа
- `ErrCodeNotFound` - ресурс не найден
- `ErrCodeInternalError` - внутренняя ошибка сервера
- `ErrCodeJSONInvalid` - ошибка парсинга JSON ответа

### Пример обработки ошибок

```go
response, err := wbClient.GetParentCategories(ctx, params)
if err != nil {
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        switch appErr.Code {
        case errors.ErrCodeUnauthorized:
            log.Println("Ошибка авторизации: проверьте API ключ")
        case errors.ErrCodeForbidden:
            log.Println("Недостаточно прав доступа")
        case errors.ErrCodeNotFound:
            log.Println("Ресурс не найден")
        default:
            log.Printf("Ошибка: %s", appErr.Message)
        }
    }
    return
}
```

## Валидация

Клиент включает встроенную валидацию параметров:

### Валидация ID

```go
func (c *WBClient) ValidateCategoryID(categoryID int) error
func (c *WBClient) ValidateSubjectID(subjectID int) error
```

### Валидация параметров запроса

```go
func (c *WBClient) ValidateAPIRequestParams(params *APIRequestParams) error
```

## Логирование

Клиент логирует все запросы к API включая:

- URL и метод запроса
- Время выполнения
- Статус код ответа
- Размер ответа
- Ошибки (если есть)

**Пример лога:**

```
[WB] Making GET request to https://content-api.wildberries.ru/content/v1/object/categories/parent?page=1&per_page=10
[WB] Response received: status=200, body_size=2048 bytes
[WB] Successfully retrieved 5 parent categories (page 1 of 3)
```

## Пример использования

### Полный пример получения характеристик товара

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/ilkoid/poncho-tools/config"
    "github.com/ilkoid/poncho-tools/wildberries"
)

func main() {
    // Загрузка конфигурации
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Создание клиента
    wbClient, err := wildberries.NewWBClient(&cfg.Wildberries, log.New(os.Stdout, "[WB] ", log.LstdFlags))
    if err != nil {
        log.Fatal(err)
    }
    defer wbClient.Close()

    // Создание контекста
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 1. Получаем родительские категории
    categories, err := wbClient.GetAllParentCategories(ctx)
    if err != nil {
        log.Printf("Error getting categories: %v", err)
        return
    }

    // 2. Ищем категорию "Одежда"
    var clothingCategory *wildberries.Category
    for _, cat := range categories {
        if strings.Contains(strings.ToLower(cat.Name), "одежда") {
            clothingCategory = &cat
            break
        }
    }

    if clothingCategory == nil {
        log.Println("Категория 'Одежда' не найдена")
        return
    }

    // 3. Получаем предметы категории
    subjects, err := wbClient.GetAllSubjectsByCategory(ctx, clothingCategory.ID)
    if err != nil {
        log.Printf("Error getting subjects: %v", err)
        return
    }

    // 4. Берем первый предмет и получаем его характеристики
    if len(subjects) == 0 {
        log.Println("Предметы не найдены")
        return
    }

    subject := subjects[0]
    characteristics, err := wbClient.GetAllSubjectCharacteristics(ctx, subject.ID)
    if err != nil {
        log.Printf("Error getting characteristics: %v", err)
        return
    }

    // 5. Выводим результат
    fmt.Printf("Категория: %s\n", clothingCategory.Name)
    fmt.Printf("Предмет: %s\n", subject.Name)
    fmt.Printf("Характеристик: %d\n", len(characteristics))

    for _, char := range characteristics {
        required := ""
        if char.IsRequired {
            required = " (обязательная)"
        }
        fmt.Printf("- %s%s\n", char.Name, required)
    }
}
```

## Ограничения и рекомендации

### Rate limiting

- Соблюдайте ограничения API Wildberries
- Используйте пагинацию для больших объемов данных
- Реализуйте экспоненциальный backoff при ошибках 429

### Таймауты

- Установите разумные таймауты для контекста
- Рекомендуемый таймаут: 30 секунд
- Для больших запросов увеличьте до 60 секунд

### Кэширование

- Кэшируйте данные категорий и предметов (они меняются редко)
- Характеристики можно кэшировать на короткое время
- Используйте in-memory кэш для улучшения производительности

## Тестирование

Для тестирования функционала используйте программу:

```bash
go run cmd/test_wildberries_api/main.go
```

Программа проверяет:
- Получение родительских категорий
- Получение предметов категории
- Получение характеристик предмета
- Поиск предметов по названию

## Интеграция с Poncho Tools

Wildberries клиент интегрирован в общую архитектуру:

1. **Конфигурация** - использует общую систему конфигурации
2. **Логирование** - интегрировано с общим логгером
3. **Ошибки** - использует унифицированную систему ошибок
4. **Валидация** - встроенная валидация параметров

Это обеспечивает согласованность с остальными компонентами системы и упрощает поддержку и развитие.