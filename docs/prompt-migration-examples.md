# Примеры миграции промптов в унифицированный формат

## Обзор

Документ содержит подробные примеры миграции существующих промптов PonchoFramework в новый унифицированный формат с использованием `{{...}}` синтаксиса.

## Пример 1: Базовая миграция sketch_creative.prompt

### Исходный формат (старый)

```handlebars
{{role "config"}}
model: zai-vision/glm-4.6v-flash
config:
  temperature: 0.7
  maxOutputTokens: 4000
{{role "system"}}
You are a precise fashion sketch analyzer and fashion assistant.  
The image is a sketch (not a photo), so be focused to collect as much data as possible to describe sketch view in a creative manner.

HARD FORMAT RULES:
- Output ONLY a single valid JSON object.
- Do NOT add any text before or after the JSON.
- Do NOT use Markdown, code blocks, or HTML.
- The answer MUST start with '{' and end with '}'.
- The only key is 'описание' and it must be in Russian ONLY.
- Value is you creative description and must be in Russian ONLY.
- NEVER use English keys or values under any circumstances.
- If you use English keys, the entire analysis will be rejected.
- There must be NO Chinese characters in the output!
- Use double quotes for all keys and string values.
- Do NOT use unescaped quotes inside values.

{{role "user"}}
ЗАДАЧА:
Создай креативное описание модного изделия на основе эскиза.

Это эскиз одежды (а не фотография). Тебе нужно:
- Найти все видимые элементы одежды или аксессуара на эскизе.
- Выделить важные конструктивные детали, которые характеризуют именно этот эскиз.
- Креативно описать только то, что реально видно или логично вытекает из рисунка.
- Придумать атмосферу, эмоции которые эта одежда лучше всего подчеркнёт. 
- Не придумывать сюжет и фон.

ФОРМАТ ВЫВОДА:
- Сформируй один JSON-объект.
- Используй только ключ 'описание' .
- Значение ключа должно быть на русском языке.

ПРИМЕР ПРАВИЛЬНОГО ФОРМАТА (пример, НЕ обязанная схема):
{
  "описание": "Чудесное платье для летней погоды",
}
ИЗОБРАЖЕНИЕ ДЛЯ АНАЛИЗА:
{{media url=photoUrl}}

ВЕРНИ ТОЛЬКО JSON-ОБЪЕКТ, БЕЗ КАКИХ-ЛИБО ДОПОЛНИТЕЛЬНЫХ СИМВОЛОВ ИЛИ ТЕКСТА.
```

### Новый формат (мигрированный)

```handlebars
{{metadata 
  name="sketch_creative" 
  version="2.0.0" 
  category="vision-analysis" 
  tags="fashion,creative,russian,json"
  description="Креативное описание fashion-эскиза на русском языке"
  author="Poncho Team"
  created="2025-12-12"
}}

{{config 
  temperature=0.7 
  max_tokens=4000 
  model="glm-vision"
  response_format='{"type": "json", "schema": {"type": "object", "properties": {"описание": {"type": "string"}}, "required": ["описание"]}}'
  strict=true
}}

{{system}}
Ты - эксперт по анализу fashion-эскизов и креативный описатель.

ПРИНЦИПЫ РАБОТЫ:
- Изображение является эскизом (не фотографией)
- Собираешь максимум визуальной информации для креативного описания
- Создаешь атмосферу и эмоциональный контекст
- Описываешь только видимые элементы и логичные выводы

ПРАВИЛА ФОРМАТИРОВАНИЯ:
- Возвращаешь ТОЛЬКО валидный JSON объект
- Ноль текста до или после JSON
- Используешь только ключ 'описание'
- Значение должно быть на русском языке
- Никаких английских ключей или значений
- Никаких китайских символов
- Двойные кавычки для всех строк
{{/system}}

{{user}}
ЗАДАЧА: Создай креативное описание модного изделия на основе эскиза.

ТРЕБОВАНИЯ К АНАЛИЗУ:
- Найти все видимые элементы одежды или аксессуара
- Выделить важные конструктивные детали эскиза
- Креативно описать видимые элементы
- Создать атмосферу и эмоции, которые подчеркивает одежда
- Не придумывать сюжет и фон

ИЗОБРАЖЕНИЕ ДЛЯ АНАЛИЗА:
{{media url=photoUrl type="image" width=640 height=480}}

ВЕРНИ ТОЛЬКО JSON-ОБЪЕКТ С КЛЮЧОМ 'описание'.
{{/user}}

{{model}}
{
  "описание": "Элегантное летнее платье в романтичном стиле с воздушным силуэтом, идеально подходящее для теплых вечеров и особых случаев"
}
{{/model}}
```

### Ключевые изменения в миграции

1. **Метаданные**: Добавлены структурированные метаданные
2. **Конфигурация**: Улучшена с указанием JSON schema
3. **Роли**: Используются `{{role}}...{{/role}}` вместо `{{role "role"}}`
4. **Медиа**: Добавлены атрибуты типа и размеров
5. **Few-shot**: Добавлен пример ответа в секции `{{model}}`

## Пример 2: Миграция sketch_description.prompt

### Исходный формат (старый)

```handlebars
{{role "config"}}
model: zai-vision/glm-4.6v-flash
config:
  temperature: 0.1
  maxOutputTokens: 2000
{{role "system"}}
You are a precise fashion sketch analyzer. You are NOT an assistant. You do not speak. Act as a strict JSON generator. Your only task is to convert visual information from clothing sketches into structured JSON. 
The image is a sketch (not a photo), so focus on garment design and construction details, not mood or atmosphere.
HARD FORMAT RULES:
- Output ONLY a single valid JSON object.
- Do NOT add any text before or after the JSON.
- Do NOT use Markdown, code blocks, or HTML.
- The answer MUST start with '{' and end with '}'.
- Keys must be in Russian (snake_case) ONLY. Example: "тип_изделия", "цвет",
- Values must be in Russian ONLY.
- NEVER use English keys under any circumstances.
- If you use English keys, the entire analysis will be rejected.
- Use double quotes for all keys and string values.
- Do NOT use unescaped quotes inside values.
- If something is unclear on the sketch, use values like "неопределено" or "не видно" in Russian.
{{role "user"}}
ЗАДАЧА:
Проанализируй эскиз одежды и опиши все элементы конструкции, извлеки все важные характеристики товара.
ИЗОБРАЖЕНИЕ ДЛЯ АНАЛИЗА:
{{media url=photoUrl}}

Это эскиз одежды (а не фотография). Тебе нужно:
- Найти и перечислить все видимые элементы одежды на эскизе.
- Выделить важные конструктивные детали, которые характеризуют именно этот эскиз.
- Описывать только то, что реально видно или логично вытекает из рисунка.
- Не придумывать атмосферу, эмоции, сюжет и фон.
ОБЯЗАТЕЛЬНЫЕ АСПЕКТЫ АНАЛИЗА (если они различимы на эскизе):
- Тип изделия (платье, юбка, брюки, жакет, пальто, костюм и т.п.).
- Силуэт и форма (приталенный, оверсайз, А-силуэт, прямой, облегающий и т.п.).
- Линия верха: воротник/горловина (тип, форма, глубина), капюшон, лацканы.
- Рукава (есть/нет, длина, форма — реглан, втачной, рубашечный и т.п.).
- Длина изделия (по талию, до бедра, миди, макси и т.д.).
- Тип пояса/линии талии (завышенная, заниженная, естественная, пояс, кулиска).
- Застёжки (молния, пуговицы, кнопки, запах; расположение — спереди, сзади, сбоку).
- Карманы (тип, количество, расположение).
- Кокетки, вытачки, рельефы, складки, драпировки, разрезы.
- Декоративные элементы (оборки, воланы, декоративные строчки, аппликации, принт, погоны и т.п.).
- Наличие дополнительных слоёв (подъюбник, жилет поверх, многослойность и т.д.).
- Вид (фронтальный, спинка, бок, комбинация видов).
- Вид утеплителя.
- Фурнитура.
- Украшения (стразы и т.д).
- Описание принта (однотонный, конкретное изображение и т.д.)
- Элементы декора (принты, шевроны и т.д.)
- Пол.
- Цвет.
- Любые другие элементы изделия, различимые на эскизе, всё, что может быть полезно для описания товара
ФОРМАТ ВЫВОДА:
- Сформируй один JSON-объект.
- Используй произвольный набор ключей, но только на русском (snake_case).
- Значения должны быть на русском языке.
- Можно использовать вложенные объекты и массивы.
- При наличии нескольких предметов одежды на эскизе можешь описать их в массиве.
ПРИМЕР ПРАВИЛЬНОГО ФОРМАТА (пример, НЕ обязанная схема):
{
  "тип_изделия": "платье",
  "цвет": "красный",
  "материал": "хлопок",
  "стиль": "летний",
  "узор": "цветочный"
}

ВЕРНИ ТОЛЬКО JSON-ОБЪЕКТ, БЕЗ КАКИХ-ЛИБО ДОПОЛНИТЕЛЬНЫХ СИМВОЛОВ ИЛИ ТЕКСТА.
```

### Новый формат (мигрированный)

```handlebars
{{metadata 
  name="sketch_description" 
  version="2.0.0" 
  category="vision-analysis" 
  tags="fashion,technical,structured,russian"
  description="Технический анализ fashion-эскиза с детальными характеристиками"
  author="Poncho Team"
  created="2025-12-12"
}}

{{config 
  temperature=0.1 
  max_tokens=2000 
  model="glm-vision"
  response_format='{"type": "json", "strict": true}'
}}

{{system}}
Ты - точный анализатор fashion-эскизов. Твоя единственная задача - преобразование визуальной информации в структурированный JSON.

ПРИНЦИПЫ:
- Изображение является эскизом (не фотографией)
- Фокус на конструктивных деталях, а не на атмосфере
- Только объективная информация из эскиза
- Никаких предположений о настроении или сюжете

ПРАВИЛА JSON:
- ТОЛЬКО валидный JSON объект
- Ключи на русском языке в snake_case
- Значения на русском языке
- Никаких английских ключей
- Двойные кавычки для всех строк
- Для нечетких элементов использовать "неопределено" или "не видно"
{{/system}}

{{user}}
ЗАДАЧА: Проанализируй эскиз одежды и извлеки все технические характеристики.

ИЗОБРАЖЕНИЕ ДЛЯ АНАЛИЗА:
{{media url=photoUrl type="image" width=640 height=480}}

ОБЯЗАТЕЛЬНЫЕ АСПЕКТЫ АНАЛИЗА:
{{#if analysis_focus}}
Фокус анализа: {{analysis_focus}}
{{/if}}

БАЗОВЫЕ ХАРАКТЕРИСТИКИ:
- Тип изделия (платье, юбка, брюки, жакет, пальто, костюм)
- Силуэт и форма (приталенный, оверсайз, А-силуэт, прямой)
- Линия верха (воротник, горловина, капюшон, лацканы)
- Рукава (наличие, длина, форма - реглан, втачной, рубашечный)
- Длина изделия (по талию, до бедра, миди, макси)
- Пояс/линия талии (завышенная, заниженная, естественная)

КОНСТРУКТИВНЫЕ ДЕТАЛИ:
- Застёжки (молния, пуговицы, кнопки, запах)
- Карманы (тип, количество, расположение)
- Кокетки, вытачки, рельефы, складки, драпировки, разрезы
- Декоративные элементы (оборки, воланы, строчки, аппликации)
- Дополнительные слои (подъюбник, жилет, многослойность)
- Фурнитура и украшения (стразы, пуговицы, заклепки)

ВИЗУАЛЬНЫЕ ХАРАКТЕРИСТИКИ:
- Вид эскиза (фронтальный, спинка, бок, комбинация)
- Принт (однотонный, узор, изображение)
- Цветовая гамма
- Пол (мужской, женский, унисекс)
- Материал (если можно определить)

ДОПОЛНИТЕЛЬНЫЕ ЭЛЕМЕНТЫ:
{{#if additional_focus}}
{{#each additional_focus}}
- {{this}}
{{/each}}
{{/if}}

ФОРМАТ ВЫВОДА:
- Структурированный JSON с произвольными ключами
- Ключи на русском языке (snake_case)
- Значения на русском языке
- Поддержка вложенных объектов и массивов
{{/user}}

{{model}}
{
  "тип_изделия": "платье",
  "силуэт": "приталенный",
  "длина": "миди",
  "линия_верха": {
    "тип": "горловина",
    "форма": "круглая"
  },
  "рукава": {
    "наличие": true,
    "длина": "короткие",
    "тип": "втачной"
  },
  "материал": "хлопок",
  "цвет": "синий",
  "декоративные_элементы": [
    "кокетка на спинке",
    "разрез по боку"
  ],
  "пол": "женский"
}
{{/model}}
```

## Пример 3: Расширенный фешн-промпт с Wildberries интеграцией

### Новый формат (с нуля)

```handlebars
{{metadata 
  name="wildberries_product_optimizer" 
  version="1.0.0" 
  category="description-generation" 
  tags="wildberries,seo,russian,fashion,ecommerce"
  description="Оптимизация описания товара для Wildberries с SEO и фешн-экспертизой"
  author="Poncho Team"
  created="2025-12-12"
}}

{{fashion_specific 
  product_type="product_description"
  target_marketplace="wildberries"
  target_audience=["young_adults", "adults"]
  seasonality=["spring", "summer", "autumn", "winter"]
  style_directions=["casual", "business", "sport"]
}}

{{config 
  temperature=0.5 
  max_tokens=3000 
  model="deepseek-chat"
  top_p=0.8
  frequency_penalty=0.1
  response_format='{"type": "json", "schema": {"type": "object", "properties": {"title": {"type": "string", "maxLength": 100}, "description": {"type": "string", "maxLength": 5000}, "keywords": {"type": "array", "items": {"type": "string"}, "maxItems": 10}, "characteristics": {"type": "object"}}, "required": ["title", "description", "keywords"]}}'
}}

{{wildberries_integration 
  api_version="v2"
  seo_optimization=true
  description_length='{"min": 200, "max": 3000}'
}}

{{system}}
Ты - эксперт по написанию описаний для Wildberries с глубоким пониманием:
- SEO-оптимизации для маркетплейсов
- Фешн-индустрии и трендов
- Русского языка и маркетинга
- Требований Wildberries к контенту

ПРИНЦИПЫ ДЛЯ WILDBERRIES:
- Заголовок до 100 символов с ключевыми словами
- Описание 200-3000 символов, структурированное по абзацам
- Ключевые слова в начале текста
- Отсутствие англицизмов и сленга
- SEO-оптимизация под поисковые запросы

КАТЕГОРИЯ: {{wildberries_category category_id}}
ЦЕЛЕВАЯ АУДИТОРИЯ: {{join target_audience ", "}}
СТИЛЕВОЕ НАПРАВЛЕНИЕ: {{style_direction}}
{{/system}}

{{user}}
Создай оптимизированное описание для товара на Wildberries:

ОСНОВНАЯ ИНФОРМАЦИЯ:
Артикул: {{article_id}}
Название бренда: {{brand_name}}
Базовое название: {{product_name}}
Категория WB: {{category_name}} (ID: {{category_id}})

ХАРАКТЕРИСТИКИ ТОВАРА:
{{#each characteristics}}
- {{name}}: {{value}}
{{/each}}

ИЗОБРАЖЕНИЯ:
{{#if main_image}}
Основное изображение:
{{media url=main_image type="image" width=800 height=600}}
{{/if}}

{{#if additional_images}}
Дополнительные изображения:
{{#each additional_images}}
{{media url=this type="image" width=400 height=300}}
{{/each}}
{{/if}}

СЕО-ТРЕБОВАНИЯ:
Ключевые слова: {{join seo_keywords ", "}}
Сезонность: {{season}}
Пол: {{gender}}
Возрастная группа: {{age_group}}

ДОПОЛНИТЕЛЬНЫЕ ТРЕБОВАНИЯ:
{{#if custom_requirements}}
{{custom_requirements}}
{{/if}}

СТРУКТУРА ОТВЕТА:
1. **title** - Оптимизированный заголовок (до 100 символов)
2. **description** - SEO-оптимизированное описание (200-3000 символов)
3. **keywords** - Дополнительные ключевые слова (массив)
4. **characteristics** - Структурированные характеристики для WB

ФОКУС НА:
{{#if focus_areas}}
{{#each focus_areas}}
- {{this}}
{{/each}}
{{else}}
- Уникальности товара
- Преимуществах для покупателя
- Сезонности и стиле
- Технических характеристиках
{{/if}}
{{/user}}

{{model}}
{
  "title": "Женское платье оверсайз с V-образным вырезом",
  "description": "Элегантное женское платье свободного кроя из натурального хлопка. Идеально подходит для создания повседневных и деловых образов. V-образный вырез визуально удлиняет шею, а оверсайз силуэт обеспечивает комфорт в течение всего дня. Платье легко сочетается с кроссовками для кэжуал-стиля или с каблуками для более формального вида. Материал приятен к телу, не вызывает раздражения и сохраняет форму после стирок.",
  "keywords": ["платье женское", "платье оверсайз", "платье хлопок", "платье повседневное", "платье деловое", "платье V-вырез"],
  "characteristics": {
    "пол": "женский",
    "материал": "хлопок 100%",
    "силуэт": "оверсайз",
    "длина": "миди",
    "рукав": "короткий",
    "вырез": "V-образный",
    "сезон": "весна-лето-осень",
    "уход": "деликатная стирка при 30°C"
  }
}
{{/model}}
```

## Пример 4: Мультимодальный промпт с анализом нескольких изображений

### Новый формат (расширенный)

```handlebars
{{metadata 
  name="multi_image_fashion_analysis" 
  version="1.0.0" 
  category="vision-analysis" 
  tags="fashion,multimodal,comparison,technical"
  description="Анализ нескольких изображений fashion-изделия с сравнительной оценкой"
}}

{{config 
  temperature=0.3 
  max_tokens=4000 
  model="glm-vision"
  top_p=0.9
  timeout=120
}}

{{vision_analysis 
  image_requirements='{"min_width": 640, "min_height": 480, "max_file_size": 5242880, "supported_formats": ["jpeg", "jpg", "png"]}'
  analysis_features=["color_detection", "style_classification", "material_identification", "quality_assessment", "detail_extraction"]
  confidence_threshold=0.7
  multiple_objects=true
}}

{{system}}
Ты - senior fashion-эксперт с многолетним опытом анализа изображений. Твоя задача - комплексный анализ нескольких изображений одного изделия.

ПРОТОКОЛ АНАЛИЗА:
1. Анализ каждого изображения индивидуально
2. Сравнительный анализ между изображениями
3. Выявление противоречий и несоответствий
4. Формирование целостного представления об изделии

КРИТЕРИИ КАЧЕСТВА:
- Уверенность детекции: {{confidence_threshold}}
- Учет ракурсов и освещения
- Внимание к деталям и качеству изображений
{{/system}}

{{user}}
ЗАДАЧА: Проанализируй несколько изображений fashion-изделия.

ИНФОРМАЦИЯ ОБ ИЗДЕЛИИ:
Артикул: {{article_id}}
Название: {{product_name}}
Категория: {{category_name}}
Предполагаемый материал: {{material_type}}

ИЗОБРАЖЕНИЯ ДЛЯ АНАЛИЗА:
{{#if front_image}}
ПЕРЕДНИЙ ВИД:
{{media url=front_image type="image" width=800 height=600}}
{{/if}}

{{#if back_image}}
ЗАДНИЙ ВИД:
{{media url=back_image type="image" width=800 height=600}}
{{/if}}

{{#if side_image}}
БОКОВОЙ ВИД:
{{media url=side_image type="image" width=800 height=600}}
{{/if}}

{{#if detail_images}}
ДЕТАЛЬНЫЕ ИЗОБРАЖЕНИЯ:
{{#each detail_images}}
{{media url=this type="image" width=400 height=300}}
{{/each}}
{{/if}}

ТРЕБОВАНИЯ К АНАЛИЗУ:
{{#if analysis_requirements}}
{{#each analysis_requirements}}
- {{this}}
{{/each}}
{{else}}
- Определить тип изделия и стиль
- Проанализировать конструктивные особенности
- Оценить качество материалов и пошива
- Выявить возможные дефекты или особенности
- Сравнить данные с указанной информацией
{{/if}}

ФОРМАТ ОТВЕТА:
- Структурированный JSON с анализом каждого изображения
- Сравнительная таблица характеристик
- Итоговая оценка качества и рекомендации
{{/user}}

{{model}}
{
  "individual_analysis": {
    "front_view": {
      "visible_features": {
        "тип_изделия": "платье",
        "силуэт": "приталенный",
        "длина": "миди",
        "рукава": "длинные втачные",
        "вырез": "круглый",
        "материал": "вискоза",
        "цвет": "темно-синий"
      },
      "quality_indicators": {
        "качество_пошива": "хорошее",
        "соответствие_размеру": "стандартное",
        "дефекты": "не выявлены"
      },
      "confidence": 0.85
    },
    "back_view": {
      "visible_features": {
        "застежка": "потайная молния",
        "вытачки": "поясные",
        "отделка": "кант по швам"
      },
      "quality_indicators": {
        "качество_пошива": "хорошее",
        "аккуратность_строчек": "высокая"
      },
      "confidence": 0.82
    }
  },
  "comparative_analysis": {
    "consistency_check": {
      "цвет_соответствует": true,
      "материал_одинаков": true,
      "конструкция_логична": true
    },
    "discrepancies": [],
    "overall_quality_score": 0.84
  },
  "final_assessment": {
    "общая_оценка": "отлично",
    "рекомендации": ["Изделие соответствует описанию", "Качество пошива высокое", "Рекомендуется для продажи"],
    "potential_issues": []
  }
}
{{/model}}
```

## Пример 5: Russian language специализированный промпт

### Новый формат (с Russian language хелперами)

```handlebars
{{metadata 
  name="russian_fashion_description" 
  version="1.0.0" 
  category="description-generation" 
  tags="russian,fashion,grammar,terminology"
  description="Грамматически корректное описание на русском языке с фешн-терминологией"
}}

{{russian_language 
  grammar_check=true
  terminology_standard="industry"
  character_encoding="utf-8"
  forbidden_words=["крутой", "топовый", "лук", "аутфит"]
  required_terms=["изделие", "материал", "композиция"]
}}

{{config 
  temperature=0.4 
  max_tokens=2500 
  model="deepseek-chat"
}}

{{system}}
Ты - эксперт по русской фешн-терминологии и стилистике.

ГРАММАТИЧЕСКИЕ ТРЕБОВАНИЯ:
- Безупречная русская грамматика
- Отсутствие англицизмов и сленга
- Правильное использование падежей
- Корректное согласование слов

ТЕРМИНОЛОГИЧЕСКИЙ СТАНДАРТ: {{terminology_standard}}
ЗАПРЕЩЕННЫЕ СЛОВА: {{join forbidden_words ", "}}
ОБЯЗАТЕЛЬНЫЕ ТЕРМИНЫ: {{join required_terms ", "}}
{{/system}}

{{user}}
Создай грамматически корректное описание изделия:

БАЗОВАЯ ИНФОРМАЦИЯ:
{{#if product_name}}
Название: {{product_name}}
{{/if}}

{{#if material}}
Материал: {{material}}
{{/if}}

{{#if style}}
Стиль: {{style}}
{{/if}}

ХАРАКТЕРИСТИКИ:
{{#each characteristics}}
{{russian_case word=name case="nominative"}}: {{russian_case word=value case="nominative"}}
{{/each}}

ДОПОЛНИТЕЛЬНО:
{{#if additional_info}}
{{additional_info}}
{{/if}}

ТРЕБОВАНИЯ:
- Использовать только русскую лексику
- Применять фешн-терминологию корректно
- Обеспечить грамматическую правильность
- Избегать калькирования с английского

КОЛИЧЕСТВО: {{#if_eq quantity 1}}{{pluralize_russian word="изделие" count=1}}{{else}}{{pluralize_russian word="изделие" count=quantity}}{{/if_eq}}
{{/user}}

{{model}}
Изделие представляет собой элегантное платье из натурального шелка. Композиция выполнена в классическом стиле с приталенным силуэтом. Материал обладает превосходными тактильными свойствами и драпируется красивыми фалдами. Длина до колена создает универсальный силуэт, подходящий для различных ситуаций. Отделка выполнена с использованием ручной вышивки, что подчеркивает премиальный характер изделия.
{{/model}}
```

## Инструменты для миграции

### Автоматический мигратор

```go
type PromptMigrator struct {
    oldPatterns map[string]*regexp.Regexp
    newTemplates map[string]string
    validator *PromptValidator
}

func (pm *PromptMigrator) MigratePrompt(oldContent string) (*MigrationResult, error) {
    result := &MigrationResult{
        OriginalContent: oldContent,
        MigratedContent: "",
        Changes: []MigrationChange{},
        Warnings: []string{},
    }
    
    // 1. Извлечение метаданных из старого формата
    oldMetadata := pm.extractOldMetadata(oldContent)
    
    // 2. Преобразование в новый формат
    newMetadata := pm.convertMetadata(oldMetadata)
    
    // 3. Преобразование ролевых секций
    newRoles := pm.convertRoles(oldContent)
    
    // 4. Преобразование медиа элементов
    newMedia := pm.convertMedia(oldContent)
    
    // 5. Сборка нового контента
    newContent := pm.buildNewPrompt(newMetadata, newRoles, newMedia)
    
    // 6. Валидация результата
    if err := pm.validator.ValidatePrompt(newContent); err != nil {
        return result, fmt.Errorf("migration validation failed: %w", err)
    }
    
    result.MigratedContent = newContent
    result.Success = true
    
    return result, nil
}
```

### Скрипт массовой миграции

```bash
#!/bin/bash

# Массовая миграция промптов
PROMPTS_DIR="examples/test_data/prompts"
OUTPUT_DIR="examples/migrated/prompts"
MIGRATOR="./prompt-migrator"

mkdir -p "$OUTPUT_DIR"

for prompt_file in "$PROMPTS_DIR"/*.prompt; do
    filename=$(basename "$prompt_file")
    echo "Migrating $filename..."
    
    $MIGRATOR migrate "$prompt_file" > "$OUTPUT_DIR/$filename"
    
    if [ $? -eq 0 ]; then
        echo "✅ $filename migrated successfully"
    else
        echo "❌ $filename migration failed"
    fi
done

echo "Migration completed. Check $OUTPUT_DIR for results."
```

Эти примеры демонстрируют полный процесс миграции с сохранением функциональности и добавлением новых возможностей унифицированного формата.