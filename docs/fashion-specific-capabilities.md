# Fashion-specific возможности унифицированного формата промптов

## Обзор

PonchoFramework предоставляет специализированные возможности для фешн-индустрии, включая интеграцию с Wildberries, Russian language поддержку, и продвинутый анализ изображений одежды.

## Fashion-specific хелперы

### 1. Анализ категорий и стилей

#### {{fashion_category}}

Получение информации о фешн-категории по ID или названию.

```handlebars
{{fashion_category category_id=123}}
{{fashion_category name="платья"}}
{{fashion_category category_id=123 language="ru"}}
```

**Атрибуты:**
- `category_id` (опциональный) - ID категории
- `name` (опциональный) - название категории
- `language` (опциональный, default="ru") - язык результата

**Возвращает:** Объект с информацией о категории (название, описание, характеристики)

#### {{fashion_style}}

Определение стиля и направление в моде.

```handlebars
{{fashion_style style="casual" audience="youth"}}
{{fashion_style style="business" season="winter"}}
{{fashion_style style="vintage" era="1980s"}}
```

**Атрибуты:**
- `style` (обязательный) - стиль (casual, business, sport, evening, vintage, streetwear, minimalist, romantic, bohemian, classic)
- `audience` (опциональный) - целевая аудитория
- `season` (опциональный) - сезонность
- `era` (опциональный) - эпоха для vintage стилей

### 2. Анализ материалов и тканей

#### {{material_analysis}}

Детальный анализ материала ткани.

```handlebars
{{material_analysis material="cotton" properties="breathability,texture"}}
{{material_analysis material="silk" properties="drape,lustre,weight"}}
{{material_analysis material="wool" properties="warmth,elasticity,texture"}}
```

**Атрибуты:**
- `material` (обязательный) - название материала
- `properties` (опциональный) - свойства для анализа через запятую

**Доступные свойства:**
- `breathability` - воздухопроницаемость
- `texture` - текстура
- `drape` - драпируемость
- `lustre` - блеск
- `weight` - вес
- `warmth` - теплота
- `elasticity` - эластичность
- `durability` - долговечность
- `care` - уход

#### {{material_translation}}

Перевод названий материалов между языками.

```handlebars
{{material_translation material="cotton" from="en" to="ru"}}
{{material_translation material="хлопок" from="ru" to="en"}}
{{material_translation material="silk" from="en" to="zh"}}
```

**Атрибуты:**
- `material` (обязательный) - название материала
- `from` (обязательный) - исходный язык
- `to` (обязательный) - целевой язык

### 3. Цветовой анализ

#### {{color_analysis}}

Анализ цвета и его фешн-интерпретации.

```handlebars
{{color_analysis hex="#FF5733" format="russian"}}
{{color_analysis rgb="255,87,51" format="pantone"}}
{{color_analysis name="красный" format="seasonal"}}
```

**Атрибуты:**
- `hex` (опциональный) - цвет в HEX формате
- `rgb` (опциональный) - цвет в RGB формате
- `name` (опциональный) - название цвета
- `format` (обязательный) - формат вывода

**Форматы вывода:**
- `russian` - русское название и ассоциации
- `pantone` - ближайший Pantone цвет
- `seasonal` - сезонная рекомендация
- `psychology` - психологическое воздействие
- `combinations` - сочетаемые цвета

#### {{color_compatibility}}

Определение совместимости цветов.

```handlebars
{{color_compatibility color1="#FF5733" color2="#3366FF" type="complementary"}}
{{color_compatibility color1="красный" color2="синий" type="harmony"}}
{{color_compatibility palette="warm" count=3}}
```

### 4. Размеры и посадка

#### {{size_conversion}}

Конвертация размеров между системами.

```handlebars
{{size_conversion size="M" from="international" to="eu"}}
{{size_conversion size="42" from="eu" to="us"}}
{{size_conversion size="S" from="international" to="ru"}}
{{size_conversion bust="92" waist="74" hip="100" system="ru"}}
```

**Атрибуты:**
- `size` (опциональный) - размер для конвертации
- `from` (обязательный) - исходная система
- `to` (обязательный) - целевая система
- `bust`, `waist`, `hip` (опциональные) - измерения для точной конвертации

**Системы размеров:**
- `eu` - европейская
- `us` - американская
- `uk` - британская
- `ru` - российская
- `international` - международная (S, M, L, XL)

#### {{fit_recommendation}}

Рекомендации по посадке.

```handlebars
{{fit_recommendation body_type="hourglass" garment_type="dress"}}
{{fit_recommendation height="170" weight="65" style="casual"}}
{{fit_recommendation measurements="92-74-100" preference="slim"}}
```

### 5. Сезонность и тренды

#### {{season_detection}}

Определение сезона для одежды.

```handlebars
{{season_detection material="wool" style="coat" hemisphere="north"}}
{{season_detection colors="orange,brown" style="sweater"}}
{{season_detection features="sleeves,heavy_fabric" month="12"}}
```

**Атрибуты:**
- `material` (опциональный) - материал изделия
- `style` (опциональный) - стиль изделия
- `colors` (опциональный) - цвета через запятую
- `features` (опциональный) - характеристики через запятую
- `month` (опциональный) - месяц (1-12)
- `hemisphere` (опциональный) - полушарие (north/south)

#### {{trend_analysis}}

Анализ трендов и популярности.

```handlebars
{{trend_analysis style="oversized" timeframe="current" market="russia"}}
{{trend_analysis color="pastel" season="spring" year="2025"}}
{{trend_analysis category="dresses" audience="gen_z"}}
```

## Wildberries интеграция

### 1. Работа с категориями Wildberries

#### {{wildberries_category}}

Получение информации о категории Wildberries.

```handlebars
{{wildberries_category category_id=12345}}
{{wildberries_category parent_id=0 recursive=true}}
{{wildberries_category search="платья" language="ru"}}
```

**Атрибуты:**
- `category_id` (опциональный) - ID категории
- `parent_id` (опциональный) - ID родительской категории
- `recursive` (опциональный, default=false) - рекурсивный поиск
- `search` (опциональный) - поиск по названию
- `language` (опциональный, default="ru") - язык

#### {{wildberries_characteristics}}

Получение характеристик для категории.

```handlebars
{{wildberries_characteristics subject_id=12345}}
{{wildberries_characteristics category_id=12345 required_only=true}}
{{wildberries_characteristics subject_id=12345 filter="size,color"}}
```

**Атрибуты:**
- `subject_id` (опциональный) - ID предмета
- `category_id` (опциональный) - ID категории
- `required_only` (опциональный, default=false) - только обязательные
- `filter` (опциональный) - фильтр характеристик

### 2. SEO оптимизация для Wildberries

#### {{wb_seo_keywords}}

Генерация SEO ключевых слов.

```handlebars
{{wb_seo_keywords category="платья" style="casual" material="cotton"}}
{{wb_seo_keywords product_name="вечернее платье" audience="women"}}
{{wb_seo_keywords characteristics="длинный,вечерний,черный" count=10}}
```

**Атрибуты:**
- `category` (опциональный) - категория товара
- `product_name` (опциональный) - название товара
- `style` (опциональный) - стиль
- `material` (опциональный) - материал
- `audience` (опциональный) - аудитория
- `characteristics` (опциональный) - характеристики через запятую
- `count` (опциональный, default=10) - количество ключевых слов

#### {{wb_description_optimizer}}

Оптимизация описания для Wildberries.

```handlebars
{{wb_description_optimizer text="базовое описание" category="платья"}}
{{wb_description_optimizer description="описание" keywords="ключевые слова"}}
{{wb_description_optimizer text="текст" max_length=3000 include_seo=true}}
```

**Атрибуты:**
- `text` или `description` (обязательный) - исходный текст
- `category` (опциональный) - категория для контекста
- `keywords` (опциональный) - ключевые слова
- `max_length` (опциональный, default=3000) - максимальная длина
- `include_seo` (опциональный, default=true) - включить SEO оптимизацию

### 3. Валидация под Wildberries

#### {{wb_validation}}

Проверка соответствия требованиям Wildberries.

```handlebars
{{wb_validation category_id=12345 characteristics=characteristics}}
{{wb_validation description="текст описания" rules="length,keywords,format"}}
{{wb_validation product_data=data validation_level="strict"}}
```

**Атрибуты:**
- `category_id` (опциональный) - ID категории
- `characteristics` (опциональный) - характеристики товара
- `description` (опциональный) - описание товара
- `rules` (опциональный) - правила валидации через запятую
- `validation_level` (опциональный, default="standard") - уровень валидации

## Russian language возможности

### 1. Морфология и грамматика

#### {{russian_case}}

Склонение слов по падежам.

```handlebars
{{russian_case word="платье" case="nominative"}}      // платье
{{russian_case word="платье" case="genitive"}}       // платья
{{russian_case word="платье" case="dative"}}        // платью
{{russian_case word="платье" case="accusative"}}     // платье
{{russian_case word="платье" case="instrumental"}}    // платьем
{{russian_case word="платье" case="prepositional"}}  // о платье
```

**Падежи:**
- `nominative` - именительный (кто? что?)
- `genitive` - родительный (кого? чего?)
- `dative` - дательный (кому? чему?)
- `accusative` - винительный (кого? что?)
- `instrumental` - творительный (кем? чем?)
- `prepositional` - предложный (о ком? о чем?)

#### {{pluralize_russian}}

Множественное число для русских слов.

```handlebars
{{pluralize_russian word="платье" count=1}}    // платье
{{pluralize_russian word="платье" count=2}}    // платья
{{pluralize_russian word="платье" count=5}}    // платьев
{{pluralize_russian word="платье" count=21}}   // платьев
```

#### {{russian_gender}}

Определение рода русского слова.

```handlebars
{{russian_gender word="платье"}}     // средний
{{russian_gender word="рубашка"}}   // женский
{{russian_gender word="пиджак"}}    // мужской
```

### 2. Фешн-терминология

#### {{fashion_terminology}}

Перевод фешн-терминов на русский.

```handlebars
{{fashion_terminology term="oversized" target="ru"}}
{{fashion_terminology term="A-line" target="ru"}}
{{fashion_terminology term="v-neck" target="ru"}}
{{fashion_terminology term="крой" target="en"}}
```

**Результаты:**
- `oversized` → `оверсайз`, `свободный крой`
- `A-line` → `А-силуэт`
- `v-neck` → `V-образный вырез`
- `крой` → `cut`, `pattern`

#### {{russian_adjective_form}}

Формирование прилагательных в нужной форме.

```handlebars
{{russian_adjective_form adjective="красивый" noun="платье" case="nominative"}}
{{russian_adjective_form adjective="красивый" noun="платья" case="genitive" count=2}}
{{russian_adjective_form adjective="красивый" noun="платьев" case="genitive" count=5}}
```

### 3. Транслитерация и локализация

#### {{transliterate}}

Транслитерация между алфавитами.

```handlebars
{{transliterate text="Красное платье" from="cyr" to="lat"}}     // Krasnoe plat'e
{{transliterate text="Fashion dress" from="lat" to="cyr"}}     // Фэшн дресс
{{transliterate text="Вечернее платье" from="cyr" to="pinyin"}} // Wechernie plat'ie
```

#### {{russian_currency}}

Форматирование валюты для русского рынка.

```handlebars
{{russian_currency amount=1500 currency="RUB"}}    // 1 500 ₽
{{russian_currency amount=99.99 currency="RUB"}}   // 99,99 ₽
{{russian_currency amount=100 currency="USD"}}      // 100 $ (≈ 9 000 ₽)
```

## Vision анализ для fashion

### 1. Детекция элементов одежды

#### {{fashion_object_detection}}

Детекция объектов одежды на изображении.

```handlebars
{{fashion_object_detection image_url=image_url confidence=0.7}}
{{fashion_object_detection image_url=image_url categories="dress,shirt,shoes"}}
{{fashion_object_detection image_url=image_url include_attributes=true}}
```

**Атрибуты:**
- `image_url` (обязательный) - URL изображения
- `confidence` (опциональный, default=0.7) - порог уверенности
- `categories` (опциональный) - фильтр категорий
- `include_attributes` (опциональный, default=false) - включить атрибуты

**Возвращает:** Массив объектов с типом, позицией, атрибутами

#### {{garment_analysis}}

Детальный анализ garments.

```handlebars
{{garment_analysis image_url=image_url analysis_type="technical"}}
{{garment_analysis image_url=image_url focus="construction,materials"}}
{{garment_analysis image_url=image_url include_measurements=true}}
```

**Типы анализа:**
- `technical` - технический анализ (крой, швы, фурнитура)
- `style` - стилистический анализ (силуэт, стиль, сезонность)
- `quality` - оценка качества (материалы, пошив, дефекты)
- `complete` - комплексный анализ

### 2. Цвет и паттерны

#### {{color_extraction}}

Извлечение цветовой палитры.

```handlebars
{{color_extraction image_url=image_url palette_count=5}}
{{color_extraction image_url=image_url format="pantone"}}
{{color_extraction image_url=image_url focus="dominant,accent"}}
```

#### {{pattern_recognition}}

Распознавание паттернов и принтов.

```handlebars
{{pattern_recognition image_url=image_url}}
{{pattern_recognition image_url=image_url categories="geometric,floral,abstract"}}
{{pattern_recognition image_url=image_url include_scale=true}}
```

**Типы паттернов:**
- `geometric` - геометрические
- `floral` - цветочные
- `abstract` - абстрактные
- `striped` - полосатые
- `checked` - клетчатые
- `solid` - однотонные

### 3. Качество и дефекты

#### {{quality_assessment}}

Оценка качества изображения и изделия.

```handlebars
{{quality_assessment image_url=image_url criteria="resolution,lighting,focus"}}
{{quality_assessment image_url=image_url include_defects=true}}
{{quality_assessment image_url=image_url quality_scale="1-10"}}
```

#### {{defect_detection}}

Детекция дефектов изделия.

```handlebars
{{defect_detection image_url=image_url defect_types="stains,tears,wrinkles"}}
{{defect_detection image_url=image_url severity_threshold=0.5}}
{{defect_detection image_url=image_url include_suggestions=true}}
```

## Продвинутые возможности

### 1. Комбинированные хелперы

#### {{complete_fashion_analysis}}

Комплексный анализ с использованием всех возможностей.

```handlebars
{{complete_fashion_analysis 
  image_url=image_url
  analysis_depth="detailed"
  include_style=true
  include_technical=true
  include_quality=true
  target_market="wildberries"
  language="ru"
}}
```

#### {{wb_product_optimization}}

Полная оптимизация для Wildberries.

```handlebars
{{wb_product_optimization
  image_url=image_url
  product_data=data
  category_id=12345
  optimize_seo=true
  validate_requirements=true
  generate_keywords=true
}}
```

### 2. Условные fashion-хелперы

```handlebars
{{#if_eq product_type "dress"}}
  {{fashion_style style="evening" audience="women"}}
{{else_if_eq product_type "shirt"}}
  {{fashion_style style="casual" audience="unisex"}}
{{/if_eq}}

{{#each colors}}
  {{color_analysis name=this format="russian"}}
{{/each}}

{{#with material}}
  {{material_analysis material=this properties="breathability,texture"}}
  {{material_translation material=this from="en" to="ru"}}
{{/with}}
```

### 3. Интеграция с внешними сервисами

#### {{trend_forecast}}

Прогноз трендов на основе данных.

```handlebars
{{trend_forecast season="spring" year="2025" market="russia"}}
{{trend_forecast category="dresses" timeframe="6months"}}
{{trend_forecast colors="pastel" styles="minimalist"}}
```

#### {{competitor_analysis}}

Анализ конкурентов на Wildberries.

```handlebars
{{competitor_analysis category="платья" price_range="1000-5000"}}
{{competitor_analysis product_type="вечерние платья" top_n=10}}
{{competitor_analysis keywords="коктейльное платье" analysis_depth="deep"}}
```

## Пример комплексного использования

```handlebars
{{metadata 
  name="fashion_product_complete_analysis" 
  version="1.0.0" 
  category="vision-analysis" 
  tags="fashion,wildberries,russian,complete"
}}

{{config temperature=0.5 max_tokens=4000 model="glm-vision"}}

{{system}}
Ты - senior fashion-эксперт с глубоким знанием Wildberries и Russian fashion market.
{{/system}}

{{user}}
Выполни комплексный анализ fashion-изделия:

ИЗОБРАЖЕНИЕ:
{{media url=image_url type="image"}}

БАЗОВЫЕ ДАННЫЕ:
Артикул: {{article_id}}
Название: {{product_name}}
Категория WB: {{wildberries_category category_id=category_id}}

АНАЛИЗ:
1. {{fashion_object_detection image_url=image_url include_attributes=true}}
2. {{garment_analysis image_url=image_url analysis_type="complete"}}
3. {{color_extraction image_url=image_url palette_count=5}}
4. {{pattern_recognition image_url=image_url}}

ОПТИМИЗАЦИЯ ДЛЯ WILDBERRIES:
{{wb_seo_keywords category=category_name product_name=product_name count=10}}
{{wb_description_optimizer text=base_description category=category_name max_length=3000}}

РУССКАЯ ЛОКАЛИЗАЦИЯ:
{{#each detected_materials}}
  {{material_translation material=this from="en" to="ru"}}
{{/each}}

{{#each detected_styles}}
  {{fashion_terminology term=this target="ru"}}
{{/each}}
{{/user}}
```

Эти fashion-specific возможности делают PonchoFramework мощным инструментом для фешн-индустрии с полной поддержкой Russian market и Wildberries интеграции.