package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// TestV1ParserWithRealPrompts тестирует V1 парсер с реальными промптами
func TestV1ParserWithRealPrompts(t *testing.T) {
	logger := createTestLogger(t)
	v1Parser := NewV1Parser()
	v1Integration := NewV1Integration(logger)

	// Тестовые файлы
	testFiles := []struct {
		name           string
		filename       string
		expectedRoles  []string
		expectedMedia  []string
	}{
		{
			name:          "sketch_description",
			filename:      "sketch_description.prompt",
			expectedRoles: []string{"config", "system", "user"},
			expectedMedia: []string{"photoUrl"},
		},
		{
			name:          "sketch_creative",
			filename:      "sketch_creative.prompt",
			expectedRoles: []string{"config", "system", "user"},
			expectedMedia: []string{"photoUrl"},
		},
	}

	for _, test := range testFiles {
		t.Run(test.name, func(t *testing.T) {
			// Загружаем тестовый файл
			filePath := filepath.Join("..", "examples", "test_data", "prompts", test.filename)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", filePath, err)
			}

			// Проверяем формат V1
			if !v1Integration.IsV1Format(string(content)) {
				t.Errorf("File %s is not recognized as V1 format", test.filename)
			}

			// Парсим содержимое
			v1Data, err := v1Parser.Parse(string(content))
			if err != nil {
				t.Fatalf("Failed to parse V1 content: %v", err)
			}

			// Проверяем извлечение ролей
			t.Run("role_extraction", func(t *testing.T) {
				if v1Data.Config == "" {
					t.Error("Config section is empty")
				}
				if v1Data.System == "" {
					t.Error("System section is empty")
				}
				if v1Data.User == "" {
					t.Error("User section is empty")
				}

				// Проверяем содержимое секций
				if len(v1Data.Config) < 10 {
					t.Errorf("Config section too short: %d characters", len(v1Data.Config))
				}
				if len(v1Data.System) < 10 {
					t.Errorf("System section too short: %d characters", len(v1Data.System))
				}
				if len(v1Data.User) < 10 {
					t.Errorf("User section too short: %d characters", len(v1Data.User))
				}
			})

			// Проверяем обработку media переменных
			t.Run("media_variables", func(t *testing.T) {
				for _, expectedMedia := range test.expectedMedia {
					if _, exists := v1Data.Media[expectedMedia]; !exists {
						t.Errorf("Media variable %s not found", expectedMedia)
					}
				}

				if len(v1Data.Media) != len(test.expectedMedia) {
					t.Errorf("Expected %d media variables, got %d", 
						len(test.expectedMedia), len(v1Data.Media))
				}
			})

			// Конвертируем в PromptTemplate
			template, err := v1Integration.ParseAndConvert(string(content), test.name)
			if err != nil {
				t.Fatalf("Failed to convert to PromptTemplate: %v", err)
			}

			// Проверяем PromptTemplate
			t.Run("prompt_template_conversion", func(t *testing.T) {
				if template.Name != test.name {
					t.Errorf("Expected template name %s, got %s", test.name, template.Name)
				}

				if len(template.Parts) == 0 {
					t.Error("Template has no parts")
				}

				// Проверяем наличие системных и пользовательских частей
				hasSystemPart := false
				hasUserPart := false
				for _, part := range template.Parts {
					if part.Type == interfaces.PromptPartTypeSystem {
						hasSystemPart = true
					}
					if part.Type == interfaces.PromptPartTypeUser {
						hasUserPart = true
					}
				}

				if !hasSystemPart {
					t.Error("Template missing system part")
				}
				if !hasUserPart {
					t.Error("Template missing user part")
				}

				// Проверяем переменные
				if len(template.Variables) == 0 {
					t.Error("Template has no variables")
				}

				for _, expectedMedia := range test.expectedMedia {
					found := false
					for _, variable := range template.Variables {
						if variable.Name == expectedMedia {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Variable %s not found in template", expectedMedia)
					}
				}
			})

			// Логируем результаты для отладки
			t.Logf("File: %s", test.filename)
			t.Logf("Config length: %d", len(v1Data.Config))
			t.Logf("System length: %d", len(v1Data.System))
			t.Logf("User length: %d", len(v1Data.User))
			t.Logf("Media variables: %v", v1Data.Media)
			t.Logf("Template parts: %d", len(template.Parts))
			t.Logf("Template variables: %d", len(template.Variables))
		})
	}
}

// TestV1ParserMediaVariableExtraction тестирует извлечение media переменных
func TestV1ParserMediaVariableExtraction(t *testing.T) {
	v1Parser := NewV1Parser()

	testCases := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name: "single_media_variable",
			content: `{{role "user"}}
This is a test with {{media url=photoUrl}} variable.
`,
			expected: map[string]string{"photoUrl": "photoUrl"},
		},
		{
			name: "multiple_media_variables",
			content: `{{role "user"}}
Here are two variables: {{media url=image1}} and {{media url=image2}}.
`,
			expected: map[string]string{"image1": "image1", "image2": "image2"},
		},
		{
			name: "media_with_spaces",
			content: `{{role "user"}}
Media with spaces: {{media url=my photo URL}}.
`,
			expected: map[string]string{"my photo URL": "my photo URL"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v1Data, err := v1Parser.Parse(tc.content)
			if err != nil {
				t.Fatalf("Failed to parse content: %v", err)
			}

			if len(v1Data.Media) != len(tc.expected) {
				t.Errorf("Expected %d media variables, got %d", 
					len(tc.expected), len(v1Data.Media))
			}

			for expectedKey, expectedValue := range tc.expected {
				if actualValue, exists := v1Data.Media[expectedKey]; !exists {
					t.Errorf("Media variable %s not found", expectedKey)
				} else if actualValue != expectedValue {
					t.Errorf("Media variable %s has wrong value: expected %s, got %s",
						expectedKey, expectedValue, actualValue)
				}
			}
		})
	}
}

// TestV1ParserValidation тестирует валидацию формата V1
func TestV1ParserValidation(t *testing.T) {
	v1Parser := NewV1Parser()

	testCases := []struct {
		name     string
		content  string
		valid    bool
		errorMsg string
	}{
		{
			name:     "valid_v1_format",
			content:  `{{role "config"}}\ntest\n{{role "system"}}\ntest\n{{role "user"}}\ntest`,
			valid:    true,
			errorMsg: "",
		},
		{
			name:     "missing_roles",
			content:  "Just plain text without roles",
			valid:    false,
			errorMsg: "no role delimiters found",
		},
		{
			name:     "invalid_role_type",
			content:  `{{role "invalid"}}\ntest`,
			valid:    false,
			errorMsg: "invalid role type",
		},
		{
			name:     "empty_content",
			content:  "",
			valid:    false,
			errorMsg: "no role delimiters found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v1Parser.ValidateFormat(tc.content)
			
			if tc.valid {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tc.errorMsg != "" && !contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tc.errorMsg, err)
				}
			}
		})
	}
}

// TestV1ParserWithActualFiles тестирует парсер с реальными файлами
func TestV1ParserWithActualFiles(t *testing.T) {
	logger := createTestLogger(t)
	v1Parser := NewV1Parser()
	v1Integration := NewV1Integration(logger)

	// Путь к директории с тестовыми промптами
	promptsDir := filepath.Join("..", "examples", "test_data", "prompts")

	// Проверяем, что директория существует
	if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
		t.Fatalf("Test prompts directory does not exist: %s", promptsDir)
	}

	// Находим все .prompt файлы
	files, err := os.ReadDir(promptsDir)
	if err != nil {
		t.Fatalf("Failed to read prompts directory: %v", err)
	}

	promptFiles := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".prompt" {
			promptFiles = append(promptFiles, file.Name())
		}
	}

	if len(promptFiles) == 0 {
		t.Error("No .prompt files found in test directory")
	}

	// Тестируем каждый файл
	for _, filename := range promptFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join(promptsDir, filename)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", filePath, err)
			}

			// Валидируем формат
			if err := v1Parser.ValidateFormat(string(content)); err != nil {
				t.Errorf("File %s has invalid V1 format: %v", filename, err)
			}

			// Парсим содержимое
			v1Data, err := v1Parser.Parse(string(content))
			if err != nil {
				t.Fatalf("Failed to parse file %s: %v", filename, err)
			}

			// Проверяем основные требования
			if v1Data.Config == "" {
				t.Errorf("File %s has empty config section", filename)
			}
			if v1Data.System == "" {
				t.Errorf("File %s has empty system section", filename)
			}
			if v1Data.User == "" {
				t.Errorf("File %s has empty user section", filename)
			}

			// Конвертируем и проверяем PromptTemplate
			templateName := v1Integration.GenerateTemplateName(string(content))
			template, err := v1Integration.ParseAndConvert(string(content), templateName)
			if err != nil {
				t.Fatalf("Failed to convert file %s to PromptTemplate: %v", filename, err)
			}

			// Проверяем структуру PromptTemplate
			if template.Name == "" {
				t.Errorf("File %s generated empty template name", filename)
			}
			if len(template.Parts) == 0 {
				t.Errorf("File %s generated template with no parts", filename)
			}

			t.Logf("Successfully processed file: %s", filename)
			t.Logf("  Config length: %d", len(v1Data.Config))
			t.Logf("  System length: %d", len(v1Data.System))
			t.Logf("  User length: %d", len(v1Data.User))
			t.Logf("  Media variables: %v", v1Data.Media)
			t.Logf("  Template parts: %d", len(template.Parts))
			t.Logf("  Template variables: %d", len(template.Variables))
		})
	}
}

// SimpleExampleV1Parser демонстрирует простой пример использования парсера
func SimpleExampleV1Parser() {
	fmt.Println("=== Пример использования V1 парсера ===")

	// Создаем парсер
	logger := createTestLogger(nil)
	v1Parser := NewV1Parser()
	v1Integration := NewV1Integration(logger)

	// Тестовый контент
	testContent := `{{role "config"}}
model: test-model
config:
  temperature: 0.7
{{role "system"}}
You are a test assistant.
{{role "user"}}
Analyze this image: {{media url=testImage}}`

	fmt.Println("Исходный контент:")
	fmt.Println(testContent)
	fmt.Println()

	// Проверяем формат
	isV1 := v1Integration.IsV1Format(testContent)
	fmt.Printf("Это V1 формат: %v\n", isV1)

	// Парсим
	v1Data, err := v1Parser.Parse(testContent)
	if err != nil {
		fmt.Printf("Ошибка парсинга: %v\n", err)
		return
	}

	fmt.Println("\nРезультаты парсинга:")
	fmt.Printf("Config: %d символов\n", len(v1Data.Config))
	fmt.Printf("System: %d символов\n", len(v1Data.System))
	fmt.Printf("User: %d символов\n", len(v1Data.User))
	fmt.Printf("Media переменные: %v\n", v1Data.Media)

	// Конвертируем в PromptTemplate
	template, err := v1Integration.ParseAndConvert(testContent, "example")
	if err != nil {
		fmt.Printf("Ошибка конвертации: %v\n", err)
		return
	}

	fmt.Println("\nPromptTemplate:")
	fmt.Printf("Имя: %s\n", template.Name)
	fmt.Printf("Описание: %s\n", template.Description)
	fmt.Printf("Количество частей: %d\n", len(template.Parts))
	fmt.Printf("Количество переменных: %d\n", len(template.Variables))

	fmt.Println("\nПример завершен успешно!")
}

// Вспомогательные функции

func createTestLogger(t *testing.T) interfaces.Logger {
	// Простой тестовый логгер
	return &TestLogger{t: t}
}

type TestLogger struct {
	t *testing.T
}

func (l *TestLogger) Debug(msg string, args ...interface{}) {
	if l.t != nil {
		l.t.Logf("DEBUG: "+msg, args...)
	}
}

func (l *TestLogger) Info(msg string, args ...interface{}) {
	if l.t != nil {
		l.t.Logf("INFO: "+msg, args...)
	}
}

func (l *TestLogger) Warn(msg string, args ...interface{}) {
	if l.t != nil {
		l.t.Logf("WARN: "+msg, args...)
	}
}

func (l *TestLogger) Error(msg string, args ...interface{}) {
	if l.t != nil {
		l.t.Logf("ERROR: "+msg, args...)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}