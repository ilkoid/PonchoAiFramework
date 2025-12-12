package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/prompts"
)

// SimpleLogger простая реализация логгера для примера
type SimpleLogger struct{}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, args)
}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, args)
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] %s %v\n", msg, args)
}

func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, args)
}

func main() {
	fmt.Println("=== Пример использования V1 парсера ===")
	fmt.Println()

	// Создаем логгер
	logger := &SimpleLogger{}

	// Создаем V1 интеграцию
	v1Integration := prompts.NewV1Integration(logger)

	// Пример 1: Парсинг простого V1 промпта
	fmt.Println("1. Парсинг простого V1 промпта:")
	simplePrompt := `{{role "config"}}
model: test-model
config:
  temperature: 0.7
{{role "system"}}
You are a helpful assistant.
{{role "user"}}
Analyze this image: {{media url=testImage}}`

	example1(v1Integration, simplePrompt)

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Пример 2: Парсинг реального файла sketch_description.prompt
	fmt.Println("2. Парсинг реального файла sketch_description.prompt:")
	example2(v1Integration)

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Пример 3: Демонстрация обработки ошибок
	fmt.Println("3. Демонстрация обработки ошибок:")
	example3(v1Integration)

	fmt.Println("\n=== Пример завершен ===")
}

func example1(v1Integration *prompts.V1Integration, content string) {
	// Проверяем формат
	isV1 := v1Integration.IsV1Format(content)
	fmt.Printf("Это V1 формат: %v\n", isV1)

	if !isV1 {
		fmt.Println("Контент не является V1 форматом")
		return
	}

	// Генерируем имя шаблона
	templateName := v1Integration.GenerateTemplateName(content)
	fmt.Printf("Сгенерированное имя шаблона: %s\n", templateName)

	// Парсим и конвертируем
	template, err := v1Integration.ParseAndConvert(content, templateName)
	if err != nil {
		fmt.Printf("Ошибка парсинга: %v\n", err)
		return
	}

	// Выводим результаты
	printTemplateInfo(template)
}

func example2(v1Integration *prompts.V1Integration) {
	// Путь к реальному файлу
	filePath := "../examples/test_data/prompts/sketch_description.prompt"
	
	// Читаем файл
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Ошибка чтения файла %s: %v\n", filePath, err)
		return
	}

	// Парсим
	templateName := "sketch_description_real"
	template, err := v1Integration.ParseAndConvert(string(content), templateName)
	if err != nil {
		fmt.Printf("Ошибка парсинга файла: %v\n", err)
		return
	}

	fmt.Printf("Файл: %s\n", filePath)
	printTemplateInfo(template)
}

func example3(v1Integration *prompts.V1Integration) {
	// Пример невалидного контента
	invalidContent := "Это не V1 формат - нет role директив"

	fmt.Printf("Контент: %s\n", invalidContent)
	
	isV1 := v1Integration.IsV1Format(invalidContent)
	fmt.Printf("Это V1 формат: %v\n", isV1)

	if !isV1 {
		fmt.Println("✅ Корректно определено, что это не V1 формат")
		
		// Пробуем все равно спарсить (должна быть ошибка)
		_, err := v1Integration.ParseAndConvert(invalidContent, "invalid")
		if err != nil {
			fmt.Printf("✅ Ожидаемая ошибка при парсинге: %v\n", err)
		}
	}
}

func printTemplateInfo(template *interfaces.PromptTemplate) {
	fmt.Printf("✅ Шаблон успешно создан:\n")
	fmt.Printf("  Имя: %s\n", template.Name)
	fmt.Printf("  Описание: %s\n", template.Description)
	fmt.Printf("  Версия: %s\n", template.Version)
	fmt.Printf("  Категория: %s\n", template.Category)
	fmt.Printf("  Теги: %v\n", template.Tags)
	fmt.Printf("  Количество частей: %d\n", len(template.Parts))
	fmt.Printf("  Количество переменных: %d\n", len(template.Variables))
	
	// Выводим информацию о частях
	fmt.Println("\n  Части шаблона:")
	for i, part := range template.Parts {
		contentPreview := part.Content
		if len(contentPreview) > 50 {
			contentPreview = contentPreview[:47] + "..."
		}
		fmt.Printf("    %d. Тип: %s, Контент: %s\n", i+1, part.Type, contentPreview)
		
		if part.Media != nil {
			fmt.Printf("       Media URL: %s\n", part.Media.URL)
		}
	}
	
	// Выводим информацию о переменных
	if len(template.Variables) > 0 {
		fmt.Println("\n  Переменные:")
		for i, variable := range template.Variables {
			fmt.Printf("    %d. Имя: %s, Тип: %s, Обязательная: %v\n", 
				i+1, variable.Name, variable.Type, variable.Required)
			if variable.DefaultValue != nil {
				fmt.Printf("       Значение по умолчанию: %v\n", variable.DefaultValue)
			}
		}
	}
	
	fmt.Println()
}