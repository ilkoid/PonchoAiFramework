# S3 Explorer - Утилита для анализа данных в S3

Простая, но мощная утилита для исследования и парсинга данных из S3 хранилища, специально адаптированная для работы с Wildberries и fashion данными.

## Установка

Скомпилируйте утилиту:
```bash
go build -o s3-explorer ./cmd/s3-explorer
```

## Настройка

Скопируйте и настройте `.env` файл:
```bash
cp .env.example .env
```

Обязательные переменные окружения:
```bash
export S3_ACCESS_KEY="your_yandex_cloud_access_key"
export S3_SECRET_KEY="your_yandex_cloud_secret_key"
export S3_BUCKET="plm-ai"
```

## Использование

### Базовые команды

**1. Просмотр списка объектов:**
```bash
# Показать все объекты
./s3-explorer list

# Показать объекты с префиксом
./s3-explorer list articles/
./s3-explorer list articles/12612003/

# Вывод в CSV формате
./s3-explorer list articles/ -format=csv
```

**2. Скачивание объекта:**
```bash
# Вывести в консоль
./s3-explorer get articles/12612003.json

# Сохранить в файл
./s3-explorer get articles/12612003.json ./data/article.json
```

**3. Парсинг и анализ данных:**
```bash
# Анализировать все JSON файлы в директории
./s3-explorer parse articles/

# Показать краткую сводку
./s3-explorer parse articles/ -format=summary

# Подробный вывод
./s3-explorer parse articles/ -verbose
```

**4. Статистика по данным:**
```bash
# Общая статистика
./s3-explorer stats articles/

# Детальная статистика
./s3-explorer stats articles/ -verbose
```

**5. Извлечение данных:**
```bash
# Извлечь все файлы с префиксом
./s3-explorer extract articles/

# Извлечь в конкретную директорию
./s3-extractor extract articles/12612003/ -output-dir=./extracted

# "Сухой" запуск (без скачивания)
./s3-extractor extract articles/ -dry-run
```

## Примеры использования для Fashion/Wildberries данных

### 1. Анализ артикулов Wildberries
```bash
# Получить статистику по артикулам
./s3-explorer stats articles/ -verbose

# Извлечь все данные конкретного артикула
./s3-extractor extract articles/12612003/ -output-dir=./article_12612003
```

### 2. Парсинг JSON данных артикулов
```bash
# Показать примеры структуры данных
./s3-explorer parse articles/ -format=summary

# Найти все JSON файлы и проанализировать
./s3-explorer parse articles/12612003/ -format=json
```

### 3. Работа с изображениями
```bash
# Список всех изображений
./s3-explorer list images/

# Статистика по размерам изображений
./s3-explorer stats images/ -verbose
```

### 4. Мониторинг изменений
```bash
# Показать последние файлы
./s3-explorer stats articles/ -verbose | grep "Latest File"

# Извлечь новые данные за последние N дней
./s3-extractor extract articles/2024/01/ -dry-run
```

## Форматы вывода

### JSON (по умолчанию)
```json
{
  "key": "articles/12612003.json",
  "size": 2048,
  "last_modified": "2024-01-15T10:30:00Z",
  "etag": "\"d41d8cd98f00b204e9800998ecf8427e\"",
  "is_folder": false
}
```

### CSV
```
key,size,last_modified,etag,storage_class,content_type,is_folder
articles/12612003.json,2048,2024-01-15T10:30:00Z,"d41d8...",STANDARD,application/json,false
```

### Summary
```
Parse Summary
=============
Total Files: 150

File Types:
  .json: 120
  .jpg: 25
  .png: 5

Size Distribution:
  <1KB: 10
  <1MB: 130
  <10MB: 10
```

## Production сценарии

### 1. Валидация структуры данных
```bash
# Проверить наличие всех необходимых полей в JSON
./s3-explorer parse articles/ | jq '.data | keys'
```

### 2. Мониторинг качества данных
```bash
# Найти пустые или поврежденные файлы
./s3-extractor extract articles/ -dry-run -verbose | grep "Failed"
```

### 3. Анализ использования хранилища
```bash
# Статистика по использованию
./s3-explorer stats / -verbose

# Крупные файлы
./s3-explorer stats / -verbose | grep "Largest File"
```

### 4. Резервное копирование
```bash
# Создать локальную копию важных данных
./s3-extractor extract articles/ -output-dir=./backup/$(date +%Y%m%d)
```

## Полезные флаги

- `-verbose`: Подробный вывод с логированием
- `-dry-run`: "Сухой" запуск без скачивания файлов
- `-format`: Формат вывода (json, csv, summary)
- `-output-dir`: Директория для извлеченных файлов

## Расширенное использование

### Комбинирование с другими утилитами

**Анализ данных с помощью jq:**
```bash
./s3-extractor get articles/12612003.json | jq '.data.subject_name'
```

**Интеграция с AI анализом:**
```bash
# Извлечь данные и проанализировать
./s3-extractor parse articles/ -format=json > articles_data.json
python analyze_fashion_data.py articles_data.json
```

**Автоматизация с bash скриптами:**
```bash
#!/bin/bash
# Мониторинг новых артикулов
echo "Checking for new articles..."
NEW_COUNT=$(./s3-explorer stats articles/ -format=json | jq '.total_objects')
echo "Total articles: $NEW_COUNT"

# Извлечь новые данные
if [ $NEW_COUNT -gt 1000 ]; then
    ./s3-extractor extract articles/ -output-dir=./latest
    echo "New articles extracted to ./latest"
fi
```

## Безопасность

- Никогда не храните ключи доступа в коде
- Используйте переменные окружения или `.env` файл
- Добавьте `.env` в `.gitignore`
- Для production используйте IAM роли с минимальными правами

## Проблемы и решения

### 1. Ошибка аутентификации
```
Failed to list objects: access denied
```
**Решение:** Проверьте S3_ACCESS_KEY и S3_SECRET_KEY

### 2. Медленная работа
**Решение:** Используйте префиксы для ограничения количества объектов

### 3. Большой объем данных
**Решение:** Используйте `-dry-run` для предварительной оценки

## Разработка

Для добавления новых команд:
1. Добавьте функцию `handleNewCommand` в `main.go`
2. Зарегистрийте команду в `switch`
3. Реализуйте необходимую логику

Для тестирования:
```bash
go test ./cmd/s3-explorer
```

## Поддержка

При возникновении проблем:
1. Проверьте конфигурацию в `.env`
2. Используйте флаг `-verbose` для детального логирования
3. Убедитесь в наличии прав доступа к бакету