# Сводка по унифицированному формату промптов PonchoFramework

## Обзор

Разработан унифицированный текстовый формат промптов для PonchoFramework с использованием `{{...}}` синтаксиса. Формат обеспечивает обратную совместимость с существующими промптами и предоставляет расширенные возможности для фешн-индустрии.

## Созданные документы

### 1. Спецификация формата
**Файл:** [`unified-prompt-format-specification.md`](unified-prompt-format-specification.md)

- Полное описание синтаксиса `{{...}}`
- Структура промпта и все элементы
- Примеры использования
- Fashion-specific возможности

### 2. Схема валидации
**Файл:** [`prompt-validation-schema.md`](prompt-validation-schema.md)

- JSON Schema для валидации промптов
- Extended schema для fashion промптов
- Правила валидации синтаксиса
- Алгоритмы валидации

### 3. Примеры миграции
**Файл:** [`prompt-migration-examples.md`](prompt-migration-examples.md)

- Детальные примеры миграции существующих промптов
- Сравнение старого и нового форматов
- Инструменты для автоматической миграции
- Расширенные примеры с фешн-спецификой

### 4. Руководство по парсингу
**Файл:** [`prompt-parsing-processing-guide.md`](prompt-parsing-processing-guide.md)

- Архитектура парсера и лексера
- Алгоритмы синтаксического анализа
- Template Engine и обработка хелперов
- Кэширование и оптимизация

### 5. Fashion-specific возможности
**Файл:** [`fashion-specific-capabilities.md`](fashion-specific-capabilities.md)

- Специализированные хелперы для фешн-индустрии
- Интеграция с Wildberries API
- Russian language поддержка
- Vision анализ для fashion

## Ключевые характеристики формата

### ✅ Единый синтаксис
Все элементы используют `{{...}}` синтаксис:
- `{{metadata ...}}` - метаданные
- `{{config ...}}` - конфигурация модели
- `{{system}}...{{/system}}` - ролевые секции
- `{{media ...}}` - мультимедийный контент
- `{{variable}}` - переменные

### ✅ Обратная совместимость
Существующие промпты продолжают работать:
- Автоматическая миграция старого синтаксиса
- Инструменты для конвертации
- Пошаговые примеры миграции

### ✅ Расширяемость
Поддержка продвинутых возможностей:
- Conditionals: `{{#if}}...{{/if}}`
- Loops: `{{#each}}...{{/each}}`
- Fashion-specific хелперы
- Russian language функции

### ✅ Fashion-specific специализация
Специализированные возможности:
- Интеграция с Wildberries API
- Анализ fashion изображений
- Russian language поддержка
- SEO оптимизация для маркетплейсов

### ✅ Multimodal поддержка
Полная поддержка медиа:
- `{{media url="..." type="image"}}`
- Vision анализ для GLM-4.6V
- Множественные изображения
- Видео и аудио поддержка

## Основные преимущества

### Для разработчиков
- **Простота**: Интуитивный `{{...}}` синтаксис
- **Мощь**: Conditionals, loops, helpers
- **Валидация**: JSON schema валидация
- **Кэширование**: Оптимизация производительности

### для фешн-индустрии
- **Специализация**: Fashion-specific хелперы
- **Wildberries**: Глубокая интеграция API
- **Russian**: Нативная поддержка языка
- **Vision**: Анализ fashion изображений

### Для бизнеса
- **SEO**: Оптимизация для маркетплейсов
- **Качество**: Валидация и контроль
- **Масштаб**: Поддержка больших объемов
- **Аналитика**: Метрики и мониторинг

## Пример полного промпта

```handlebars
{{metadata 
  name="wildberries_fashion_analyzer" 
  version="1.0.0" 
  category="vision-analysis" 
  tags="fashion,wildberries,russian,seo"
}}

{{config 
  temperature=0.5 
  max_tokens=3000 
  model="glm-vision"
}}

{{system}}
Ты - senior fashion-эксперт для Wildberries.
{{/system}}

{{user}}
Проанализируй fashion-изделие:
{{media url=image_url type="image"}}

Категория: {{wildberries_category category_id=category_id}}
SEO ключи: {{wb_seo_keywords category=category_name}}
{{/user}}
```

## Интеграция с PonchoFramework

### Архитектурная интеграция
```go
type PonchoFramework struct {
    // Существующие компоненты
    models  *PonchoModelRegistry
    tools   *PonchoToolRegistry
    flows   *PonchoFlowRegistry
    
    // Новый компонент
    prompts *PromptManager
}

type PromptManager struct {
    parser         *PromptParser
    validator      *PromptValidator
    templateEngine *TemplateEngine
    cache          *PromptCache
}
```

### Использование в коде
```go
// Загрузка и выполнение промпта
result, err := framework.ExecutePrompt(ctx, "fashion_analyzer", map[string]interface{}{
    "image_url": "https://example.com/image.jpg",
    "category_id": "12345",
    "category_name": "платья",
})
```

## Следующие шаги

### Phase 1: Implementation (2-3 недели)
1. **Core Parser** - Реализация лексера и парсера
2. **Template Engine** - Базовые хелперы и функции
3. **Validation** - JSON schema валидация
4. **Integration** - Встраивание в PonchoFramework

### Phase 2: Fashion Features (2-3 недели)
1. **Wildberries Integration** - API клиенты и хелперы
2. **Russian Language** - Морфология и терминология
3. **Vision Analysis** - Fashion-specific анализ изображений
4. **SEO Optimization** - Ключевые слова и описания

### Phase 3: Advanced Features (1-2 недели)
1. **Caching** - Многоуровневое кэширование
2. **Metrics** - Мониторинг и аналитика
3. **Migration Tools** - Автоматическая миграция
4. **Documentation** - Примеры и лучшие практики

## Заключение

Унифицированный формат промптов PonchoFramework предоставляет:

✅ **Complete solution** для фешн-индустрии  
✅ **Backward compatibility** с существующими промптами  
✅ **Extensible architecture** для будущих расширений  
✅ **Fashion specialization** с Wildberries интеграцией  
✅ **Russian language** нативная поддержка  
✅ **Production-ready** валидация и мониторинг  

Формат готов к реализации и может значительно ускорить разработку AI-решений для фешн-индустрии на российском рынке.

---

**Документация создана:** 12 декабря 2025  
**Версия формата:** 1.0.0  
**Статус:** Готов к реализации