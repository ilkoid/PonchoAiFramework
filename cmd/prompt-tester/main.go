package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/prompts"
)

func main() {
	// Parse command line arguments
	promptPath := flag.String("prompt", "", "Path to prompt file (required)")
	outputDir := flag.String("output", "./output", "Directory to save results")
	logFile := flag.String("log", "", "Path to log file (default stdout)")
	modelOverride := flag.String("model", "", "Override model from prompt config")
	saveRequest := flag.Bool("save", true, "Save request to file")
	flag.Parse()

	if *promptPath == "" {
		log.Fatal("Error: -prompt flag is required")
	}

	// Setup logging
	logger := setupLogger(*logFile)
	defer func() {
		if file, ok := logger.Writer().(*os.File); ok && file != os.Stdout {
			file.Close()
		}
	}()

	logger.Printf("Starting prompt tester with file: %s", *promptPath)

	// Load prompt file
	content, err := os.ReadFile(*promptPath)
	if err != nil {
		logger.Fatalf("Failed to read prompt file: %v", err)
	}

	// Parse prompt using V1 parser
	template, err := parsePromptV1(string(content), filepath.Base(*promptPath), logger)
	if err != nil {
		logger.Fatalf("Failed to parse prompt: %v", err)
	}

	logger.Printf("Parsed prompt template: %s (parts: %d)", template.Name, len(template.Parts))

	// Determine model
	modelName := determineModel(template, *modelOverride, logger)
	logger.Printf("Using model: %s", modelName)

	// Build request from template
	request, err := buildRequest(template, modelName)
	if err != nil {
		logger.Fatalf("Failed to build request: %v", err)
	}

	// Log request details
	logger.Printf("Built request for model: %s", request.Model)
	logger.Printf("Number of messages: %d", len(request.Messages))
	for i, msg := range request.Messages {
		logger.Printf("  Message %d: role=%s", i, msg.Role)
		for _, part := range msg.Content {
			if part.Type == interfaces.PonchoContentTypeText {
				// Truncate long text for logging
				text := part.Text
				if len(text) > 200 {
					text = text[:200] + "..."
				}
				logger.Printf("    Text: %s", text)
			}
		}
	}

	// Print request as JSON
	fmt.Println("\n=== REQUEST (JSON) ===")
	jsonBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		logger.Printf("Failed to marshal request: %v", err)
	} else {
		fmt.Println(string(jsonBytes))
	}

	// Save to file if requested
	if *saveRequest {
		if err := saveRequestToFile(*outputDir, template.Name, request, logger); err != nil {
			logger.Printf("Warning: failed to save request: %v", err)
		}
	}

	logger.Printf("Done.")
}

// setupLogger configures logging to file or stdout
func setupLogger(logPath string) *log.Logger {
	if logPath == "" {
		return log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	return log.New(file, "", log.LstdFlags|log.Lshortfile)
}

// parsePromptV1 uses V1 parser to convert prompt content to template
func parsePromptV1(content, name string, logger *log.Logger) (*interfaces.PromptTemplate, error) {
	// Create a simple logger that implements interfaces.Logger
	simpleLogger := &simpleLogger{logger: logger}
	v1Integration := prompts.NewV1Integration(simpleLogger)

	if !v1Integration.IsV1Format(content) {
		return nil, fmt.Errorf("file is not in V1 format")
	}

	templateName := v1Integration.GenerateTemplateName(content)
	template, err := v1Integration.ParseAndConvert(content, templateName)
	if err != nil {
		return nil, fmt.Errorf("V1 parse failed: %w", err)
	}
	return template, nil
}

// determineModel extracts model from config part or uses override
func determineModel(template *interfaces.PromptTemplate, override string, logger *log.Logger) string {
	if override != "" {
		return override
	}
	// Look for config part with model: line
	for _, part := range template.Parts {
		if part.Type == interfaces.PromptPartTypeSystem && strings.Contains(part.Content, "model:") {
			// Simple extraction: find line starting with "model:"
			lines := strings.Split(part.Content, "\n")
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "model:") {
					model := strings.TrimSpace(strings.TrimPrefix(line, "model:"))
					logger.Printf("Found model in config: %s", model)
					return model
				}
			}
		}
	}
	// Default
	return "deepseek-chat"
}

// buildRequest converts template to PonchoModelRequest
func buildRequest(template *interfaces.PromptTemplate, modelName string) (*interfaces.PonchoModelRequest, error) {
	messages := []*interfaces.PonchoMessage{}

	for _, part := range template.Parts {
		switch part.Type {
		case interfaces.PromptPartTypeSystem:
			messages = append(messages, &interfaces.PonchoMessage{
				Role: interfaces.PonchoRoleSystem,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: part.Content},
				},
			})
		case interfaces.PromptPartTypeUser:
			messages = append(messages, &interfaces.PonchoMessage{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{Type: interfaces.PonchoContentTypeText, Text: part.Content},
				},
			})
		case interfaces.PromptPartTypeMedia:
			// For now, skip media as DeepSeek doesn't support vision
			// Could be extended to support other models
		}
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no valid messages in template")
	}

	request := &interfaces.PonchoModelRequest{
		Model:    modelName,
		Messages: messages,
		// Use default temperature and max tokens
	}
	return request, nil
}

// saveRequestToFile writes request to a file
func saveRequestToFile(outputDir, templateName string, request *interfaces.PonchoModelRequest, logger *log.Logger) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Join(outputDir, fmt.Sprintf("%s-%s-request.json", templateName, timestamp))

	// Create a simple struct for serialization
	result := struct {
		Template  string                         `json:"template"`
		Timestamp string                         `json:"timestamp"`
		Request   *interfaces.PonchoModelRequest `json:"request"`
	}{
		Template:  templateName,
		Timestamp: time.Now().Format(time.RFC3339),
		Request:   request,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	logger.Printf("Request saved to %s", filename)
	return nil
}

// simpleLogger adapts log.Logger to interfaces.Logger
type simpleLogger struct {
	logger *log.Logger
}

func (s *simpleLogger) Debug(msg string, fields ...interface{}) {
	s.logger.Printf("[DEBUG] "+msg, fields...)
}

func (s *simpleLogger) Info(msg string, fields ...interface{}) {
	s.logger.Printf("[INFO] "+msg, fields...)
}

func (s *simpleLogger) Warn(msg string, fields ...interface{}) {
	s.logger.Printf("[WARN] "+msg, fields...)
}

func (s *simpleLogger) Error(msg string, fields ...interface{}) {
	s.logger.Printf("[ERROR] "+msg, fields...)
}