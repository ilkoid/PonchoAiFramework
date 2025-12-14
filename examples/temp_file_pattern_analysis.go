package examples

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/core/context"
	"github.com/ilkoid/PonchoAiFramework/core/flow"
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// TempFileManager управляет временными файлами с автоматической очисткой
type TempFileManager struct {
	tempDir    string
	createdFiles map[string]*TempFileInfo
	mutex      sync.RWMutex
	cleanupInterval time.Duration
	logger     interfaces.Logger
}

// TempFileInfo содержит информацию о временном файле
type TempFileInfo struct {
	Path         string    `json:"path"`
	CreatedAt    time.Time `json:"created_at"`
	Size         int64     `json:"size"`
	AccessCount  int       `json:"access_count"`
	LastAccess   time.Time `json:"last_access"`
	AutoCleanup  bool      `json:"auto_cleanup"`
}

// NewTempFileManager создает менеджер временных файлов
func NewTempFileManager(logger interfaces.Logger) *TempFileManager {
	tempDir := filepath.Join(os.TempDir(), "poncho-framework", fmt.Sprintf("session_%d", time.Now().UnixNano()))

	// Create temp directory
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.Error("Failed to create temp directory", "path", tempDir, "error", err)
	}

	tfm := &TempFileManager{
		tempDir:        tempDir,
		createdFiles:   make(map[string]*TempFileInfo),
		cleanupInterval: 30 * time.Minute,
		logger:         logger,
	}

	// Start cleanup goroutine
	go tfm.startCleanupRoutine()

	tfm.logger.Info("Temp file manager initialized", "temp_dir", tempDir)
	return tfm
}

// CreateTempFile создает временный файл и возвращает путь
func (tfm *TempFileManager) CreateTempFile(prefix, extension string) (string, error) {
	tfm.mutex.Lock()
	defer tfm.mutex.Unlock()

	filename := fmt.Sprintf("%s_%d%s", prefix, time.Now().UnixNano(), extension)
	filePath := filepath.Join(tfm.tempDir, filename)

	// Create empty file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	file.Close()

	// Track file
	tfm.createdFiles[filePath] = &TempFileInfo{
		Path:        filePath,
		CreatedAt:   time.Now(),
		LastAccess:  time.Now(),
		AccessCount: 0,
		AutoCleanup: true,
	}

	tfm.logger.Debug("Temp file created", "path", filePath)
	return filePath, nil
}

// WriteToTempFile пишет данные во временный файл
func (tfm *TempFileManager) WriteToTempFile(data []byte, prefix, extension string) (string, error) {
	filePath, err := tfm.CreateTempFile(prefix, extension)
	if err != nil {
		return "", err
	}

	// Write data
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		tfm.CleanupFile(filePath) // Cleanup on error
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// Update file info
	tfm.mutex.Lock()
	if fileInfo, exists := tfm.createdFiles[filePath]; exists {
		fileInfo.Size = int64(len(data))
		fileInfo.LastAccess = time.Now()
	}
	tfm.mutex.Unlock()

	tfm.logger.Debug("Data written to temp file",
		"path", filePath,
		"size", len(data))

	return filePath, nil
}

// LoadFromFile загружает данные из временного файла
func (tfm *TempFileManager) LoadFromFile(filePath string) ([]byte, error) {
	tfm.mutex.Lock()
	defer tfm.mutex.Unlock()

	// Check if file exists in our tracking
	fileInfo, exists := tfm.createdFiles[filePath]
	if !exists {
		return nil, fmt.Errorf("temp file not tracked: %s", filePath)
	}

	// Update access info
	fileInfo.AccessCount++
	fileInfo.LastAccess = time.Now()

	// Load file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}

	tfm.logger.Debug("Temp file loaded",
		"path", filePath,
		"size", len(data),
		"access_count", fileInfo.AccessCount)

	return data, nil
}

// CleanupFile удаляет временный файл
func (tfm *TempFileManager) CleanupFile(filePath string) error {
	tfm.mutex.Lock()
	defer tfm.mutex.Unlock()

	// Remove from tracking
	delete(tfm.createdFiles, filePath)

	// Delete file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to cleanup temp file: %w", err)
	}

	tfm.logger.Debug("Temp file cleaned up", "path", filePath)
	return nil
}

// CleanupAll удаляет все временные файлы
func (tfm *TempFileManager) CleanupAll() error {
	tfm.mutex.Lock()
	defer tfm.mutex.Unlock()

	var errors []error

	for filePath := range tfm.createdFiles {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, err)
			tfm.logger.Warn("Failed to cleanup temp file", "path", filePath, "error", err)
		}
	}

	// Clear tracking
	tfm.createdFiles = make(map[string]*TempFileInfo)

	// Remove temp directory
	if err := os.RemoveAll(tfm.tempDir); err != nil {
		tfm.logger.Warn("Failed to remove temp directory", "path", tfm.tempDir, "error", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup completed with %d errors", len(errors))
	}

	tfm.logger.Info("All temp files cleaned up", "temp_dir", tfm.tempDir)
	return nil
}

// GetStats возвращает статистику использования
func (tfm *TempFileManager) GetStats() map[string]interface{} {
	tfm.mutex.RLock()
	defer tfm.mutex.RUnlock()

	stats := map[string]interface{}{
		"temp_dir":      tfm.tempDir,
		"file_count":    len(tfm.createdFiles),
		"total_size":    int64(0),
		"oldest_file":   "",
		"newest_file":   "",
	}

	var oldestTime, newestTime time.Time
	var totalSize int64

	for _, fileInfo := range tfm.createdFiles {
		totalSize += fileInfo.Size

		if oldestTime.IsZero() || fileInfo.CreatedAt.Before(oldestTime) {
			oldestTime = fileInfo.CreatedAt
			stats["oldest_file"] = fileInfo.Path
		}

		if newestTime.IsZero() || fileInfo.CreatedAt.After(newestTime) {
			newestTime = fileInfo.CreatedAt
			stats["newest_file"] = fileInfo.Path
		}
	}

	stats["total_size"] = totalSize
	return stats
}

// startCleanupRoutine запускает периодическую очистку старых файлов
func (tfm *TempFileManager) startCleanupRoutine() {
	ticker := time.NewTicker(tfm.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		tfm.cleanupOldFiles()
	}
}

// cleanupOldFiles удаляет файлы старше 1 часа
func (tfm *TempFileManager) cleanupOldFiles() {
	tfm.mutex.Lock()
	defer tfm.mutex.Unlock()

	cutoff := time.Now().Add(-time.Hour)
	var cleanedCount int

	for filePath, fileInfo := range tfm.createdFiles {
		if fileInfo.AutoCleanup && fileInfo.CreatedAt.Before(cutoff) {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				tfm.logger.Warn("Failed to cleanup old temp file",
					"path", filePath,
					"error", err)
			} else {
				delete(tfm.createdFiles, filePath)
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		tfm.logger.Info("Cleaned up old temp files", "count", cleanedCount)
	}
}

// Improved FlowContext with Temp File Integration
type FlowContextWithTempFiles struct {
	*context.BaseFlowContextV2
	tempFileManager *TempFileManager
}

// NewFlowContextWithTempFiles создает контекст с поддержкой временных файлов
func NewFlowContextWithTempFiles(logger interfaces.Logger) *FlowContextWithTempFiles {
	baseCtx := context.NewBaseFlowContextV2()
	tempFileManager := NewTempFileManager(logger)

	return &FlowContextWithTempFiles{
		BaseFlowContextV2: baseCtx,
		tempFileManager:   tempFileManager,
	}
}

// SaveDataToTempFile сохраняет данные во временный файл
func (fctf *FlowContextWithTempFiles) SaveDataToTempFile(data []byte, prefix string) (string, error) {
	filePath, err := fctf.tempFileManager.WriteToTempFile(data, prefix, ".tmp")
	if err != nil {
		return "", fmt.Errorf("failed to save data to temp file: %w", err)
	}

	// Store reference in context
	key := fmt.Sprintf("%s_temp_path", prefix)
	if err := fctf.SetString(key, filePath); err != nil {
		// Cleanup file if context storage fails
		fctf.tempFileManager.CleanupFile(filePath)
		return "", fmt.Errorf("failed to store temp file path in context: %w", err)
	}

	return filePath, nil
}

// LoadDataFromTempFile загружает данные из временного файла
func (fctf *FlowContextWithTempFiles) LoadDataFromTempFile(prefix string) ([]byte, error) {
	key := fmt.Sprintf("%s_temp_path", prefix)

	filePath, err := fctf.GetString(key)
	if err != nil {
		return nil, fmt.Errorf("temp file path not found for %s: %w", prefix, err)
	}

	return fctf.tempFileManager.LoadFromFile(filePath)
}

// CleanupTempFile удаляет временный файл
func (fctf *FlowContextWithTempFiles) CleanupTempFile(prefix string) error {
	key := fmt.Sprintf("%s_temp_path", prefix)

	filePath, err := fctf.GetString(key)
	if err != nil {
		return nil // File doesn't exist, nothing to cleanup
	}

	// Remove from context
	fctf.Delete(key)

	// Cleanup file
	return fctf.tempFileManager.CleanupFile(filePath)
}

// CleanupTempFiles удаляет все временные файлы из контекста
func (fctf *FlowContextWithTempFiles) CleanupTempFiles() error {
	var errors []error

	// Find all temp file keys
	for _, key := range fctf.Keys() {
		if strings.Contains(key, "_temp_path") {
			if filePath, err := fctf.GetString(key); err == nil {
				if err := fctf.tempFileManager.CleanupFile(filePath); err != nil {
					errors = append(errors, err)
				}
			}
			fctf.Delete(key)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup completed with %d errors", len(errors))
	}

	return nil
}

// GetTempFileManager возвращает менеджер временных файлов
func (fctf *FlowContextWithTempFiles) GetTempFileManager() *TempFileManager {
	return fctf.tempFileManager
}

// Example implementation of the improved pattern
func ImprovedFashionFlowExample() {
	// Initialize flow context with temp file support
	flowCtx := NewFlowContextWithTempFiles(interfaces.NewDefaultLogger())
	defer flowCtx.tempFileManager.CleanupAll() // Ensure cleanup

	// Step 1: Download from S3 to temp file
	downloadStep := func(ctx context.Context, flowCtx interfaces.FlowContext) error {
		// Simulate S3 download
		imageData := []byte("fake image data") // In real: download from S3

		// Save to temp file with automatic tracking
		tempPath, err := flowCtx.(*FlowContextWithTempFiles).SaveDataToTempFile(imageData, "product_image")
		if err != nil {
			return fmt.Errorf("failed to save image to temp file: %w", err)
		}

		fmt.Printf("Image saved to: %s\n", tempPath)
		return nil
	}

	// Step 2: Load from temp file, analyze, cleanup
	analyzeStep := func(ctx context.Context, flowCtx interfaces.FlowContext) error {
		// Load image from temp file
		imageData, err := flowCtx.(*FlowContextWithTempFiles).LoadDataFromTempFile("product_image")
		if err != nil {
			return fmt.Errorf("failed to load image from temp file: %w", err)
		}

		// Simulate vision analysis
		tags := []string{"dress", "summer", "floral", "casual"}

		// Store results
		if err := flowCtx.Set("image_tags", tags); err != nil {
			return fmt.Errorf("failed to store tags: %w", err)
		}

		// Cleanup temp file
		if err := flowCtx.(*FlowContextWithTempFiles).CleanupTempFile("product_image"); err != nil {
			flowCtx.GetLogger().Warn("Failed to cleanup temp file", "error", err)
		}

		fmt.Printf("Image analyzed, tags: %v\n", tags)
		return nil
	}

	// Step 3: Generate description
	generateStep := func(ctx context.Context, flowCtx interfaces.FlowContext) error {
		tags, err := flowCtx.GetStringSlice("image_tags")
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}

		description := fmt.Sprintf("Beautiful %s dress with %s pattern",
			tags[2], tags[1]) // floral + summer

		if err := flowCtx.SetString("description", description); err != nil {
			return fmt.Errorf("failed to store description: %w", err)
		}

		fmt.Printf("Generated description: %s\n", description)
		return nil
	}

	// Execute steps
	ctx := context.Background()

	fmt.Println("=== Step 1: Download ===")
	if err := downloadStep(ctx, flowCtx); err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	fmt.Println("\n=== Step 2: Analyze ===")
	if err := analyzeStep(ctx, flowCtx); err != nil {
		fmt.Printf("Analysis failed: %v\n", err)
		return
	}

	fmt.Println("\n=== Step 3: Generate ===")
	if err := generateStep(ctx, flowCtx); err != nil {
		fmt.Printf("Generation failed: %v\n", err)
		return
	}

	// Show stats
	stats := flowCtx.(*FlowContextWithTempFiles).GetTempFileManager().GetStats()
	fmt.Printf("\n=== Temp File Stats ===\n")
	fmt.Printf("Temp directory: %v\n", stats["temp_dir"])
	fmt.Printf("Files tracked: %v\n", stats["file_count"])
	fmt.Printf("Total size: %v bytes\n", stats["total_size"])
}

// Comparison with original pattern
func ComparePatterns() {
	fmt.Println("=== Pattern Comparison ===")

	fmt.Println("\nORIGINAL PATTERN PROBLEMS:")
	fmt.Println("❌ Race conditions: Multiple flows can overwrite /tmp/img1.jpg")
	fmt.Println("❌ Resource leaks: Errors may skip os.Remove()")
	fmt.Println("❌ Disk I/O bottleneck: Every step reads/writes disk")
	fmt.Println("❌ No error recovery: Failed step loses file reference")
	fmt.Println("❌ Manual tracking: Developer must manage all temp files")

	fmt.Println("\nIMPROVED PATTERN SOLUTIONS:")
	fmt.Println("✅ Unique temp directory per session: /tmp/poncho-framework/session_123456/")
	fmt.Println("✅ Automatic cleanup: TempFileManager tracks all files")
	fmt.Println("✅ Error safety: Cleanup runs even on errors")
	fmt.Println("✅ Memory efficiency: Files loaded only when needed")
	fmt.Println("✅ Statistics and monitoring: Built-in file tracking")
	fmt.Println("✅ Graceful degradation: Failed cleanup doesn't crash flow")
}

/*
AUDIT RESULTS:

RECOMMENDATIONS:
1. ✅ Use TempFileManager for production systems
2. ✅ Always defer cleanup or use automatic cleanup
3. ✅ Use unique session-based temp directories
4. ✅ Add file access tracking for debugging
5. ✅ Implement size limits and cleanup policies

WHEN TO USE TEMP FILE PATTERN:
✅ Large images (>50MB) that shouldn't be in memory
✅ Integration with legacy tools that require file paths
✅ Multi-step processing where memory is limited
✅ Debugging scenarios where you need to inspect files

WHEN TO USE IMAGE REFERENCE PATTERN:
✅ Small to medium images (<50MB)
✅ Frequent access to the same images
✅ Parallel processing of many images
✅ Cloud-based deployments where disk I/O is expensive
*/