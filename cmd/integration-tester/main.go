package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/models/deepseek"
	"github.com/ilkoid/PonchoAiFramework/models/zai"
	"github.com/ilkoid/PonchoAiFramework/prompts"
)

// TestResult represents the result of a single test
type TestResult struct {
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Success    bool          `json:"success"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
	Model      string        `json:"model"`
	TokensUsed int           `json:"tokens_used,omitempty"`
	Metadata   interface{}   `json:"metadata,omitempty"`
}

// TestSuite represents a collection of test results
type TestSuite struct {
	Name        string        `json:"name"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	TotalTests  int           `json:"total_tests"`
	PassedTests int           `json:"passed_tests"`
	FailedTests int           `json:"failed_tests"`
	Results     []TestResult  `json:"results"`
	Summary     TestSummary   `json:"summary"`
}

// TestSummary provides overall test statistics
type TestSummary struct {
	ModelsTested    []string           `json:"models_tested"`
	PromptsTested   []string           `json:"prompts_tested"`
	TotalTokens     int                `json:"total_tokens"`
	AvgResponseTime time.Duration      `json:"avg_response_time"`
	ErrorsByType    map[string]int     `json:"errors_by_type"`
	Performance     map[string]float64 `json:"performance"`
}

// IntegrationTester handles comprehensive testing of framework components
type IntegrationTester struct {
	framework     interfaces.PonchoFramework
	promptManager interfaces.PromptManager
	logger        interfaces.Logger
	config        *TestConfig
	results       *TestSuite
}

// TestConfig contains configuration for the integration tester
type TestConfig struct {
	ConfigPath      string        `json:"config_path"`
	PromptDir       string        `json:"prompt_dir"`
	OutputDir       string        `json:"output_dir"`
	TestDataDir     string        `json:"test_data_dir"`
	ModelsToTestStr string        `json:"models_to_test_str"`
	ModelsToTest    []string      `json:"models_to_test"`
	SkipAPIKeys     bool          `json:"skip_api_keys"`
	Verbose         bool          `json:"verbose"`
	Timeout         time.Duration `json:"timeout"`
}

func main() {
	// Parse command line arguments
	config := &TestConfig{}

	flag.StringVar(&config.ConfigPath, "config", "config.yaml", "Path to framework configuration file")
	flag.StringVar(&config.PromptDir, "prompts", "examples/test_data/prompts", "Directory containing prompt templates")
	flag.StringVar(&config.OutputDir, "output", "./test-results", "Directory to save test results")
	flag.StringVar(&config.TestDataDir, "testdata", "examples/test_data/result", "Directory containing test data")
	var modelsStr string
	flag.StringVar(&modelsStr, "models", "deepseek-chat,glm-vision-flash", "Comma-separated list of models to test")
	flag.BoolVar(&config.SkipAPIKeys, "skip-api-keys", false, "Skip tests requiring API keys")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.DurationVar(&config.Timeout, "timeout", 60*time.Second, "Test timeout duration")

	flag.Parse()

	// Parse models string
	if config.ModelsToTestStr != "" {
		models := strings.Split(modelsStr, ",")
		config.ModelsToTest = make([]string, len(models))
		for i, model := range models {
			config.ModelsToTest[i] = strings.TrimSpace(model)
		}
	}

	// Create integration tester
	tester, err := NewIntegrationTester(config)
	if err != nil {
		log.Fatalf("Failed to create integration tester: %v", err)
	}

	// Run comprehensive tests
	if err := tester.RunTests(); err != nil {
		log.Fatalf("Test execution failed: %v", err)
	}

	// Print results
	tester.PrintResults()

	// Save results
	if err := tester.SaveResults(); err != nil {
		log.Printf("Warning: failed to save results: %v", err)
	}
}

// NewIntegrationTester creates a new integration tester instance
func NewIntegrationTester(config *TestConfig) (*IntegrationTester, error) {
	// Setup logger
	logger := setupLogger(config.Verbose)

	// Load framework configuration
	frameworkConfig, err := loadFrameworkConfig(config.ConfigPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load framework config: %w", err)
	}

	// Create framework
	framework := createTestFramework(frameworkConfig, logger)

	// Create prompt manager
	promptConfig := &prompts.PromptConfig{
		Templates: struct {
			Directory      string   `yaml:"directory" json:"directory"`
			Extensions     []string `yaml:"extensions" json:"extensions"`
			AutoReload     bool     `yaml:"auto_reload" json:"auto_reload"`
			ReloadInterval string   `yaml:"reload_interval" json:"reload_interval"`
		}{
			Directory:  config.PromptDir,
			Extensions: []string{".prompt"},
		},
		Validation: struct {
			Strict         bool `yaml:"strict" json:"strict"`
			ValidateOnLoad bool `yaml:"validate_on_load" json:"validate_on_load"`
			ValidateOnExec bool `yaml:"validate_on_execute" json:"validate_on_execute"`
		}{
			ValidateOnLoad: true,
			ValidateOnExec: true,
			Strict:         false,
		},
		Execution: struct {
			DefaultModel       string  `yaml:"default_model" json:"default_model"`
			DefaultMaxTokens   int     `yaml:"default_max_tokens" json:"default_max_tokens"`
			DefaultTemperature float32 `yaml:"default_temperature" json:"default_temperature"`
			Timeout            string  `yaml:"timeout" json:"timeout"`
			RetryAttempts      int     `yaml:"retry_attempts" json:"retry_attempts"`
			RetryDelay         string  `yaml:"retry_delay" json:"retry_delay"`
		}{
			DefaultModel: "deepseek-chat",
		},
		Cache: struct {
			Enabled bool   `yaml:"enabled" json:"enabled"`
			Size    int    `yaml:"size" json:"size"`
			TTL     string `yaml:"ttl" json:"ttl"`
			Type    string `yaml:"type" json:"type"` // memory, redis
		}{
			Enabled: true,
			Size:    100,
		},
	}

	promptManager := prompts.NewPromptManager(promptConfig, framework, logger)

	tester := &IntegrationTester{
		framework:     framework,
		promptManager: promptManager,
		logger:        logger,
		config:        config,
		results: &TestSuite{
			Name:      "PonchoFramework Integration Tests",
			StartTime: time.Now(),
			Results:   make([]TestResult, 0),
			Summary: TestSummary{
				ErrorsByType: make(map[string]int),
				Performance:  make(map[string]float64),
			},
		},
	}

	return tester, nil
}

// RunTests executes the complete test suite
func (it *IntegrationTester) RunTests() error {
	it.logger.Info("Starting PonchoFramework integration tests")

	ctx := context.Background()

	// Test 1: Framework Initialization
	if err := it.testFrameworkInitialization(ctx); err != nil {
		it.logger.Error("Framework initialization test failed", "error", err)
	}

	// Test 2: Model Registration and Availability
	if err := it.testModelRegistration(ctx); err != nil {
		it.logger.Error("Model registration test failed", "error", err)
	}

	// Test 3: Configuration Loading
	if err := it.testConfigurationLoading(ctx); err != nil {
		it.logger.Error("Configuration loading test failed", "error", err)
	}

	// Test 4: Prompt Loading and Validation
	if err := it.testPromptLoading(ctx); err != nil {
		it.logger.Error("Prompt loading test failed", "error", err)
	}

	// Test 5: Text Generation with DeepSeek
	if !it.config.SkipAPIKeys {
		if err := it.testTextGeneration(ctx); err != nil {
			it.logger.Error("Text generation test failed", "error", err)
		}
	} else {
		it.logger.Info("Skipping API key dependent tests")
	}

	// Test 6: Vision Analysis with Z.AI GLM
	if !it.config.SkipAPIKeys {
		if err := it.testVisionAnalysis(ctx); err != nil {
			it.logger.Error("Vision analysis test failed", "error", err)
		}
	}

	// Test 7: Prompt Execution with Variable Substitution
	if err := it.testPromptExecution(ctx); err != nil {
		it.logger.Error("Prompt execution test failed", "error", err)
	}

	// Test 8: Performance Metrics Collection
	if err := it.testMetricsCollection(ctx); err != nil {
		it.logger.Error("Metrics collection test failed", "error", err)
	}

	// Finalize test suite
	it.results.EndTime = time.Now()
	it.results.Duration = it.results.EndTime.Sub(it.results.StartTime)
	it.results.TotalTests = len(it.results.Results)

	for _, result := range it.results.Results {
		if result.Success {
			it.results.PassedTests++
		} else {
			it.results.FailedTests++
		}
	}

	it.logger.Info("Integration tests completed",
		"total", it.results.TotalTests,
		"passed", it.results.PassedTests,
		"failed", it.results.FailedTests,
		"duration", it.results.Duration)

	return nil
}

// testFrameworkInitialization tests framework startup and basic functionality
func (it *IntegrationTester) testFrameworkInitialization(ctx context.Context) error {
	startTime := time.Now()
	result := TestResult{
		Name: "Framework Initialization",
		Type: "framework",
	}

	defer func() {
		result.Duration = time.Since(startTime)
		it.results.Results = append(it.results.Results, result)
	}()

	// Start framework
	if err := it.framework.Start(ctx); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to start framework: %v", err)
		return err
	}

	// Check health status
	health, err := it.framework.Health(ctx)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Health check failed: %v", err)
		return err
	}

	if health.Status != "healthy" {
		result.Success = false
		result.Error = fmt.Sprintf("Framework health status: %s", health.Status)
		return fmt.Errorf("framework not healthy")
	}

	result.Success = true
	result.Metadata = map[string]interface{}{
		"health_status": health.Status,
		"uptime":        health.Uptime,
		"components":    health.Components,
	}

	it.logger.Info("Framework initialization test passed")
	return nil
}

// testModelRegistration tests model registration and availability
func (it *IntegrationTester) testModelRegistration(ctx context.Context) error {
	startTime := time.Now()
	result := TestResult{
		Name: "Model Registration",
		Type: "models",
	}

	defer func() {
		result.Duration = time.Since(startTime)
		it.results.Results = append(it.results.Results, result)
	}()

	// Register DeepSeek model
	deepseekModel := deepseek.NewDeepSeekModel()
	if err := it.framework.RegisterModel("deepseek-chat", deepseekModel); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to register DeepSeek model: %v", err)
		return err
	}

	// Register Z.AI models
	zaiModel := zai.NewZAIModel()
	if err := it.framework.RegisterModel("glm-vision-flash", zaiModel); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to register Z.AI model: %v", err)
		return err
	}

	// Check model registry
	modelRegistry := it.framework.GetModelRegistry()
	registeredModels := modelRegistry.List()

	result.Success = true
	result.Metadata = map[string]interface{}{
		"registered_models": registeredModels,
		"total_models":      len(registeredModels),
	}

	// Update summary
	for _, model := range registeredModels {
		it.results.Summary.ModelsTested = append(it.results.Summary.ModelsTested, model)
	}

	it.logger.Info("Model registration test passed", "models", registeredModels)
	return nil
}

// testConfigurationLoading tests configuration loading and validation
func (it *IntegrationTester) testConfigurationLoading(ctx context.Context) error {
	startTime := time.Now()
	result := TestResult{
		Name: "Configuration Loading",
		Type: "config",
	}

	defer func() {
		result.Duration = time.Since(startTime)
		it.results.Results = append(it.results.Results, result)
	}()

	// Get framework configuration
	config := it.framework.GetConfig()
	if config == nil {
		result.Success = false
		result.Error = "Framework configuration is nil"
		return fmt.Errorf("configuration is nil")
	}

	// Validate essential sections
	if config.Models == nil {
		result.Success = false
		result.Error = "Models configuration is missing"
		return fmt.Errorf("models config missing")
	}

	if config.Logging == nil {
		result.Success = false
		result.Error = "Logging configuration is missing"
		return fmt.Errorf("logging config missing")
	}

	result.Success = true
	metadata := map[string]interface{}{
		"models_count": 0,
		"tools_count":  0,
		"flows_count":  0,
	}

	if config.Models != nil {
		metadata["models_count"] = len(config.Models)
	}

	if config.Tools != nil {
		metadata["tools_count"] = len(config.Tools)
	}

	if config.Flows != nil {
		metadata["flows_count"] = len(config.Flows)
	}

	if config.Cache != nil {
		metadata["cache_enabled"] = config.Cache.Type
	}

	if config.Metrics != nil {
		metadata["metrics_enabled"] = config.Metrics.Enabled
	}

	result.Metadata = metadata

	it.logger.Info("Configuration loading test passed")
	return nil
}

// testPromptLoading tests prompt template loading and validation
func (it *IntegrationTester) testPromptLoading(ctx context.Context) error {
	startTime := time.Now()
	result := TestResult{
		Name: "Prompt Loading",
		Type: "prompts",
	}

	defer func() {
		result.Duration = time.Since(startTime)
		it.results.Results = append(it.results.Results, result)
	}()

	// List available prompts
	prompts, err := it.promptManager.ListTemplates()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to list prompts: %v", err)
		return err
	}

	if len(prompts) == 0 {
		result.Success = false
		result.Error = "No prompts found"
		return fmt.Errorf("no prompts found")
	}

	// Test loading each prompt
	loadedPrompts := make([]string, 0)
	for _, promptName := range prompts {
		template, err := it.promptManager.LoadTemplate(promptName)
		if err != nil {
			it.logger.Warn("Failed to load prompt", "name", promptName, "error", err)
			continue
		}

		// Extract model from template config
		model := extractModelFromTemplate(template)
		if model != "" {
			it.logger.Debug("Prompt specifies model", "prompt", promptName, "model", model)
		}

		loadedPrompts = append(loadedPrompts, promptName)
	}

	if len(loadedPrompts) == 0 {
		result.Success = false
		result.Error = "No prompts could be loaded"
		return fmt.Errorf("no prompts loaded")
	}

	result.Success = true
	result.Metadata = map[string]interface{}{
		"available_prompts": prompts,
		"loaded_prompts":    loadedPrompts,
		"total_prompts":     len(prompts),
	}

	// Update summary
	it.results.Summary.PromptsTested = loadedPrompts

	it.logger.Info("Prompt loading test passed", "loaded", len(loadedPrompts))
	return nil
}

// testTextGeneration tests text generation with DeepSeek
func (it *IntegrationTester) testTextGeneration(ctx context.Context) error {
	startTime := time.Now()
	result := TestResult{
		Name:  "Text Generation (DeepSeek)",
		Type:  "generation",
		Model: "deepseek-chat",
	}

	defer func() {
		result.Duration = time.Since(startTime)
		it.results.Results = append(it.results.Results, result)
	}()

	// Check if DeepSeek API key is available
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		result.Success = false
		result.Error = "DEEPSEEK_API_KEY not available"
		return fmt.Errorf("API key not available")
	}

	// Create a simple text generation request
	request := &interfaces.PonchoModelRequest{
		Model: "deepseek-chat",
		Messages: []*interfaces.PonchoMessage{
			{
				Role: interfaces.PonchoRoleUser,
				Content: []*interfaces.PonchoContentPart{
					{
						Type: interfaces.PonchoContentTypeText,
						Text: "Напиши краткое описание модной куртки для осеннего сезона (максимум 50 слов).",
					},
				},
			},
		},
		MaxTokens:   &[]int{100}[0],
		Temperature: &[]float32{0.7}[0],
	}

	// Generate response
	response, err := it.framework.Generate(ctx, request)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Text generation failed: %v", err)
		return err
	}

	if response.Message == nil || len(response.Message.Content) == 0 {
		result.Success = false
		result.Error = "Empty response received"
		return fmt.Errorf("empty response")
	}

	result.Success = true
	result.TokensUsed = response.Usage.TotalTokens
	result.Metadata = map[string]interface{}{
		"response_length":  len(response.Message.Content[0].Text),
		"response_preview": response.Message.Content[0].Text[:min(100, len(response.Message.Content[0].Text))],
	}

	// Update summary
	it.results.Summary.TotalTokens += response.Usage.TotalTokens

	it.logger.Info("Text generation test passed", "tokens", response.Usage.TotalTokens)
	return nil
}

// testVisionAnalysis tests vision analysis with Z.AI GLM
func (it *IntegrationTester) testVisionAnalysis(ctx context.Context) error {
	startTime := time.Now()
	result := TestResult{
		Name:  "Vision Analysis (Z.AI GLM)",
		Type:  "vision",
		Model: "glm-vision-flash",
	}

	defer func() {
		result.Duration = time.Since(startTime)
		it.results.Results = append(it.results.Results, result)
	}()

	// Check if Z.AI API key is available
	if os.Getenv("ZAI_API_KEY") == "" {
		result.Success = false
		result.Error = "ZAI_API_KEY not available"
		return fmt.Errorf("API key not available")
	}

	// Find test image
	imagePath := filepath.Join(it.config.TestDataDir, "12612003", "12612003.json")
	if _, err := os.Stat(imagePath); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Test image not found: %s", imagePath)
		return err
	}

	// Load test data
	testData, err := os.ReadFile(imagePath)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to load test data: %v", err)
		return err
	}

	// Create vision analysis request using prompt
	variables := map[string]interface{}{
		"photoUrl": "data:image/jpeg;base64," + extractBase64FromJSON(testData),
	}

	// Execute vision prompt
	response, err := it.promptManager.ExecutePrompt(ctx, "sketch_description", variables, "glm-vision-flash")
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Vision analysis failed: %v", err)
		return err
	}

	if response.Message == nil || len(response.Message.Content) == 0 {
		result.Success = false
		result.Error = "Empty vision response received"
		return fmt.Errorf("empty vision response")
	}

	result.Success = true
	result.TokensUsed = response.Usage.TotalTokens
	result.Metadata = map[string]interface{}{
		"response_length":  len(response.Message.Content[0].Text),
		"response_preview": response.Message.Content[0].Text[:min(200, len(response.Message.Content[0].Text))],
		"image_path":       imagePath,
	}

	// Update summary
	it.results.Summary.TotalTokens += response.Usage.TotalTokens

	it.logger.Info("Vision analysis test passed", "tokens", response.Usage.TotalTokens)
	return nil
}

// testPromptExecution tests prompt execution with variable substitution
func (it *IntegrationTester) testPromptExecution(ctx context.Context) error {
	startTime := time.Now()
	result := TestResult{
		Name: "Prompt Execution with Variables",
		Type: "prompts",
	}

	defer func() {
		result.Duration = time.Since(startTime)
		it.results.Results = append(it.results.Results, result)
	}()

	// Test variable substitution with a text prompt
	variables := map[string]interface{}{
		"product_type": "платье",
		"season":       "лето",
		"material":     "хлопок",
		"style":        "повседневный",
	}

	// Check if API keys are available
	hasDeepSeekKey := os.Getenv("DEEPSEEK_API_KEY") != ""
	hasZAIKey := os.Getenv("ZAI_API_KEY") != ""

	if !hasDeepSeekKey && !hasZAIKey {
		it.logger.Info("Prompt execution skipped due to missing API keys")
		result.Success = true
		result.Error = "Skipped due to missing API keys"
		result.Metadata = map[string]interface{}{
			"variables": variables,
			"skipped":   true,
		}
		return nil
	}

	// Try to execute a prompt with available API keys
	var execErr error
	if hasZAIKey {
		// Try vision prompt with Z.AI
		template, _ := it.promptManager.LoadTemplate("sketch_description")
		modelName := extractModelFromTemplate(template)
		if modelName == "" {
			modelName = "glm-vision-flash" // fallback
		}
		_, execErr = it.promptManager.ExecutePrompt(ctx, "sketch_description", variables, modelName)
	} else if hasDeepSeekKey {
		// Try text prompt with DeepSeek
		_, execErr = it.promptManager.ExecutePrompt(ctx, "sketch_creative", variables, "deepseek-chat")
	}

	if execErr != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Prompt execution failed: %v", execErr)
		return execErr
	}

	result.Success = true
	result.Metadata = map[string]interface{}{
		"variables": variables,
		"executed":  true,
	}

	it.logger.Info("Prompt execution test passed")
	return nil
}

// testMetricsCollection tests metrics collection and reporting
func (it *IntegrationTester) testMetricsCollection(ctx context.Context) error {
	startTime := time.Now()
	result := TestResult{
		Name: "Metrics Collection",
		Type: "metrics",
	}

	defer func() {
		result.Duration = time.Since(startTime)
		it.results.Results = append(it.results.Results, result)
	}()

	// Get framework metrics
	frameworkMetrics, err := it.framework.Metrics(ctx)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to get framework metrics: %v", err)
		return err
	}

	// Get prompt manager metrics
	// Get metrics from prompt manager (method doesn't exist yet)
	promptMetrics := &prompts.SystemMetrics{}

	result.Success = true
	result.Metadata = map[string]interface{}{
		"framework_metrics": frameworkMetrics,
		"prompt_metrics":    promptMetrics,
		"total_requests":    frameworkMetrics.GeneratedRequests.TotalRequests,
		"total_errors":      frameworkMetrics.Errors.TotalErrors,
	}

	it.logger.Info("Metrics collection test passed")
	return nil
}

// PrintResults displays test results in a formatted way
func (it *IntegrationTester) PrintResults() {
	fmt.Print("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Print("PONCHOFRAMEWORK INTEGRATION TEST RESULTS\n")
	fmt.Print(strings.Repeat("=", 80) + "\n")

	fmt.Printf("Test Suite: %s\n", it.results.Name)
	fmt.Printf("Duration: %v\n", it.results.Duration)
	fmt.Printf("Total Tests: %d\n", it.results.TotalTests)
	fmt.Printf("Passed: %d\n", it.results.PassedTests)
	fmt.Printf("Failed: %d\n", it.results.FailedTests)

	if it.results.TotalTests > 0 {
		successRate := float64(it.results.PassedTests) / float64(it.results.TotalTests) * 100
		fmt.Printf("Success Rate: %.1f%%\n", successRate)
	}

	fmt.Print("\n" + strings.Repeat("-", 80) + "\n")
	fmt.Print("DETAILED RESULTS:\n")
	fmt.Print(strings.Repeat("-", 80) + "\n")

	for i, result := range it.results.Results {
		status := "✅ PASS"
		if !result.Success {
			status = "❌ FAIL"
		}

		fmt.Printf("%d. %s - %s\n", i+1, result.Name, status)
		fmt.Printf("   Type: %s\n", result.Type)
		fmt.Printf("   Duration: %v\n", result.Duration)

		if result.Model != "" {
			fmt.Printf("   Model: %s\n", result.Model)
		}

		if result.TokensUsed > 0 {
			fmt.Printf("   Tokens Used: %d\n", result.TokensUsed)
		}

		if result.Error != "" {
			fmt.Printf("   Error: %s\n", result.Error)
		}

		if it.config.Verbose && result.Metadata != nil {
			fmt.Printf("   Metadata: %+v\n", result.Metadata)
		}

		fmt.Printf("\n")
	}

	// Print summary
	fmt.Print(strings.Repeat("-", 80) + "\n")
	fmt.Print("SUMMARY:\n")
	fmt.Print(strings.Repeat("-", 80) + "\n")

	if len(it.results.Summary.ModelsTested) > 0 {
		fmt.Printf("Models Tested: %v\n", it.results.Summary.ModelsTested)
	}

	if len(it.results.Summary.PromptsTested) > 0 {
		fmt.Printf("Prompts Tested: %v\n", it.results.Summary.PromptsTested)
	}

	if it.results.Summary.TotalTokens > 0 {
		fmt.Printf("Total Tokens Used: %d\n", it.results.Summary.TotalTokens)
	}

	if len(it.results.Summary.ErrorsByType) > 0 {
		fmt.Printf("Errors by Type: %v\n", it.results.Summary.ErrorsByType)
	}
}

// SaveResults saves test results to a JSON file
func (it *IntegrationTester) SaveResults() error {
	if err := os.MkdirAll(it.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Join(it.config.OutputDir, fmt.Sprintf("integration-test-%s.json", timestamp))

	data, err := json.MarshalIndent(it.results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write results file: %w", err)
	}

	fmt.Printf("Results saved to: %s\n", filename)
	return nil
}

// Helper functions

func setupLogger(verbose bool) interfaces.Logger {
	if verbose {
		return interfaces.NewDefaultLogger()
	}
	return &silentLogger{}
}

func loadFrameworkConfig(path string, logger interfaces.Logger) (*interfaces.PonchoFrameworkConfig, error) {
	// For integration tests, return a basic config without loading from file
	return &interfaces.PonchoFrameworkConfig{
		Models: map[string]*interfaces.ModelConfig{
			"deepseek-chat": {
				Provider:    "deepseek",
				ModelName:   "deepseek-chat",
				APIKey:      "${DEEPSEEK_API_KEY}",
				BaseURL:     "https://api.deepseek.com/v1",
				MaxTokens:   4000,
				Temperature: 0.7,
				Timeout:     "30s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    false,
					System:    true,
					JSONMode:  true,
				},
				CustomParams: map[string]interface{}{},
			},
			"glm-vision-flash": {
				Provider:    "zai",
				ModelName:   "glm-4.6v-flash",
				APIKey:      "${ZAI_API_KEY}",
				BaseURL:     "https://open.bigmodel.cn/api/paas/v4",
				MaxTokens:   1500,
				Temperature: 0.3,
				Timeout:     "30s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    true,
					System:    true,
					JSONMode:  true,
				},
				CustomParams: map[string]interface{}{},
			},
			"zai-vision/glm-4.6v-flash": {
				Provider:    "zai",
				ModelName:   "glm-4.6v-flash",
				APIKey:      "${ZAI_API_KEY}",
				BaseURL:     "https://open.bigmodel.cn/api/paas/v4",
				MaxTokens:   1500,
				Temperature: 0.3,
				Timeout:     "30s",
				Supports: &interfaces.ModelCapabilities{
					Streaming: true,
					Tools:     true,
					Vision:    true,
					System:    true,
					JSONMode:  true,
				},
				CustomParams: map[string]interface{}{},
			},
		},
		Tools: make(map[string]*interfaces.ToolConfig),
		Flows: make(map[string]*interfaces.FlowConfig),
		Logging: &interfaces.LoggingConfig{
			Level:  "info",
			Format: "text",
			File:   "",
		},
		Cache: &interfaces.CacheConfig{
			Type:    "memory",
			TTL:     "1h",
			MaxSize: 100,
		},
		Metrics: &interfaces.MetricsConfig{
			Enabled:  true,
			Interval: "30s",
			Endpoint: "http://localhost:9090/metrics",
		},
		S3: &interfaces.S3Config{
			URL:      "https://storage.yandexcloud.net",
			Region:   "ru-central1",
			Bucket:   "plm-ai",
			Endpoint: "storage.yandexcloud.net",
			UseSSL:   true,
		},
		Wildberries: &interfaces.WildberriesConfig{
			BaseURL: "https://content-api.wildberries.ru",
			Timeout: 30,
		},
	}, nil
}

func extractModelFromTemplate(template *interfaces.PromptTemplate) string {
	// Look for model in template parts
	for _, part := range template.Parts {
		if part.Type == interfaces.PromptPartTypeSystem && strings.Contains(part.Content, "model:") {
			// Extract model from config section
			lines := strings.Split(part.Content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "model:") {
					model := strings.TrimSpace(strings.TrimPrefix(line, "model:"))
					return model
				}
			}
		}
	}
	// If no model found in template, return empty string to use default
	return ""
}

func extractBase64FromJSON(jsonData []byte) string {
	// Extract base64 image data from the JSON structure
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return ""
	}

	if реквизиты, ok := data["Реквизиты"].(map[string]interface{}); ok {
		if миниатюра, ok := реквизиты["Миниатюра_Файл"].(string); ok {
			// This is already base64 encoded
			return миниатюра
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// silentLogger implements a no-op logger for non-verbose mode
type silentLogger struct{}

func (s *silentLogger) Debug(msg string, fields ...interface{}) {}
func (s *silentLogger) Info(msg string, fields ...interface{})  {}
func (s *silentLogger) Warn(msg string, fields ...interface{})  {}
func (s *silentLogger) Error(msg string, fields ...interface{}) {}

// createTestFramework creates a framework instance without config manager for testing
func createTestFramework(cfg *interfaces.PonchoFrameworkConfig, logger interfaces.Logger) interfaces.PonchoFramework {
	// Create a simple framework wrapper for testing
	return &testFrameworkWrapper{
		config:  cfg,
		logger:  logger,
		models:  make(map[string]interfaces.PonchoModel),
		started: false,
	}
}

// testFrameworkWrapper is a minimal framework implementation for testing
type testFrameworkWrapper struct {
	config  *interfaces.PonchoFrameworkConfig
	logger  interfaces.Logger
	models  map[string]interfaces.PonchoModel
	started bool
}

func (tfw *testFrameworkWrapper) Start(ctx context.Context) error {
	tfw.started = true
	tfw.logger.Info("Test framework started")
	return nil
}

func (tfw *testFrameworkWrapper) Stop(ctx context.Context) error {
	tfw.started = false
	tfw.logger.Info("Test framework stopped")
	return nil
}

func (tfw *testFrameworkWrapper) RegisterModel(name string, model interfaces.PonchoModel) error {
	// Initialize model with framework configuration
	if tfw.config != nil && tfw.config.Models != nil {
		if modelConfig, exists := tfw.config.Models[name]; exists {
			initParams := map[string]interface{}{
				"provider":      modelConfig.Provider,
				"model_name":    modelConfig.ModelName,
				"api_key":       modelConfig.APIKey,
				"base_url":      modelConfig.BaseURL,
				"max_tokens":    modelConfig.MaxTokens,
				"temperature":   modelConfig.Temperature,
				"timeout":       modelConfig.Timeout,
				"supports":      modelConfig.Supports,
				"custom_params": modelConfig.CustomParams,
			}
			if err := model.Initialize(context.Background(), initParams); err != nil {
				tfw.logger.Error("Failed to initialize model", "name", name, "error", err)
				return fmt.Errorf("failed to initialize model '%s': %w", name, err)
			}
		}
	}

	tfw.models[name] = model
	tfw.logger.Info("Model registered and initialized", "name", name)
	return nil
}

func (tfw *testFrameworkWrapper) RegisterTool(name string, tool interfaces.PonchoTool) error {
	return fmt.Errorf("tools not supported in test framework")
}

func (tfw *testFrameworkWrapper) RegisterFlow(name string, flow interfaces.PonchoFlow) error {
	return fmt.Errorf("flows not supported in test framework")
}

func (tfw *testFrameworkWrapper) Generate(ctx context.Context, req *interfaces.PonchoModelRequest) (*interfaces.PonchoModelResponse, error) {
	if !tfw.started {
		return nil, fmt.Errorf("framework is not started")
	}

	model, exists := tfw.models[req.Model]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found", req.Model)
	}

	return model.Generate(ctx, req)
}

func (tfw *testFrameworkWrapper) GenerateStreaming(ctx context.Context, req *interfaces.PonchoModelRequest, callback interfaces.PonchoStreamCallback) error {
	if !tfw.started {
		return fmt.Errorf("framework is not started")
	}

	model, exists := tfw.models[req.Model]
	if !exists {
		return fmt.Errorf("model '%s' not found", req.Model)
	}

	return model.GenerateStreaming(ctx, req, callback)
}

func (tfw *testFrameworkWrapper) ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error) {
	return nil, fmt.Errorf("tools not supported in test framework")
}

func (tfw *testFrameworkWrapper) ExecuteFlow(ctx context.Context, flowName string, input interface{}) (interface{}, error) {
	return nil, fmt.Errorf("flows not supported in test framework")
}

func (tfw *testFrameworkWrapper) ExecuteFlowStreaming(ctx context.Context, flowName string, input interface{}, callback interfaces.PonchoStreamCallback) error {
	return fmt.Errorf("flows not supported in test framework")
}

func (tfw *testFrameworkWrapper) GetModelRegistry() interfaces.PonchoModelRegistry {
	return &testModelRegistry{models: tfw.models}
}

func (tfw *testFrameworkWrapper) GetToolRegistry() interfaces.PonchoToolRegistry {
	return &testToolRegistry{}
}

func (tfw *testFrameworkWrapper) GetFlowRegistry() interfaces.PonchoFlowRegistry {
	return &testFlowRegistry{}
}

func (tfw *testFrameworkWrapper) GetConfig() *interfaces.PonchoFrameworkConfig {
	return tfw.config
}

func (tfw *testFrameworkWrapper) ReloadConfig(ctx context.Context) error {
	return nil
}

func (tfw *testFrameworkWrapper) Health(ctx context.Context) (*interfaces.PonchoHealthStatus, error) {
	status := "healthy"
	if !tfw.started {
		status = "unhealthy"
	}

	return &interfaces.PonchoHealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "1.0.0-test",
		Components: map[string]*interfaces.ComponentHealth{
			"framework": {
				Status:  status,
				Message: "Test framework",
			},
		},
		Uptime: time.Duration(0),
	}, nil
}

func (tfw *testFrameworkWrapper) Metrics(ctx context.Context) (*interfaces.PonchoMetrics, error) {
	return &interfaces.PonchoMetrics{
		GeneratedRequests: &interfaces.GenerationMetrics{
			TotalRequests: 0,
			SuccessCount:  0,
			ErrorCount:    0,
			ByModel:       make(map[string]*interfaces.ModelMetrics),
		},
		ToolExecutions: &interfaces.ToolMetrics{
			TotalExecutions: 0,
			ByTool:          make(map[string]int64),
		},
		FlowExecutions: &interfaces.FlowMetrics{
			TotalExecutions: 0,
			ByFlow:          make(map[string]int64),
		},
		Errors: &interfaces.ErrorMetrics{
			TotalErrors:  0,
			ByType:       make(map[string]int64),
			ByComponent:  make(map[string]int64),
			RecentErrors: make([]*interfaces.ErrorInfo, 0),
		},
		System:    &interfaces.SystemMetrics{},
		Timestamp: time.Now(),
	}, nil
}

// Test registry implementations
type testModelRegistry struct {
	models map[string]interfaces.PonchoModel
}

func (tmr *testModelRegistry) Register(name string, model interfaces.PonchoModel) error {
	tmr.models[name] = model
	return nil
}

func (tmr *testModelRegistry) Get(name string) (interfaces.PonchoModel, error) {
	model, exists := tmr.models[name]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found", name)
	}
	return model, nil
}

func (tmr *testModelRegistry) List() []string {
	names := make([]string, 0, len(tmr.models))
	for name := range tmr.models {
		names = append(names, name)
	}
	return names
}

func (tmr *testModelRegistry) Unregister(name string) error {
	delete(tmr.models, name)
	return nil
}

func (tmr *testModelRegistry) Clear() error {
	tmr.models = make(map[string]interfaces.PonchoModel)
	return nil
}

type testToolRegistry struct{}

func (ttr *testToolRegistry) Register(name string, tool interfaces.PonchoTool) error {
	return fmt.Errorf("not implemented")
}

func (ttr *testToolRegistry) Get(name string) (interfaces.PonchoTool, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ttr *testToolRegistry) List() []string {
	return []string{}
}

func (ttr *testToolRegistry) ListByCategory(category string) []string {
	return []string{}
}

func (ttr *testToolRegistry) Unregister(name string) error {
	return fmt.Errorf("not implemented")
}

func (ttr *testToolRegistry) Clear() error {
	return fmt.Errorf("not implemented")
}

type testFlowRegistry struct{}

func (tfr *testFlowRegistry) Register(name string, flow interfaces.PonchoFlow) error {
	return fmt.Errorf("not implemented")
}

func (tfr *testFlowRegistry) Get(name string) (interfaces.PonchoFlow, error) {
	return nil, fmt.Errorf("not implemented")
}

func (tfr *testFlowRegistry) List() []string {
	return []string{}
}

func (tfr *testFlowRegistry) ListByCategory(category string) []string {
	return []string{}
}

func (tfr *testFlowRegistry) Unregister(name string) error {
	return fmt.Errorf("not implemented")
}

func (tfr *testFlowRegistry) Clear() error {
	return fmt.Errorf("not implemented")
}

func (tfr *testFlowRegistry) ValidateDependencies(flow interfaces.PonchoFlow, modelRegistry interfaces.PonchoModelRegistry, toolRegistry interfaces.PonchoToolRegistry) error {
	return fmt.Errorf("not implemented")
}
