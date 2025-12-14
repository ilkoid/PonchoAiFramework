package context

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ContextMetrics собирает метрики использования FlowContext
type ContextMetrics struct {
	TotalContexts      int64     `json:"total_contexts"`
	ActiveContexts      int64     `json:"active_contexts"`
	AvgMemoryPerContext float64   `json:"avg_memory_per_context_mb"`
	PeakMemoryUsage     int64     `json:"peak_memory_usage_mb"`
	TotalOperations     int64     `json:"total_operations"`
	Errors              int64     `json:"errors"`

	mutex sync.RWMutex
}

// Global metrics collector
var globalMetrics = &ContextMetrics{
	mutex: sync.RWMutex{},
}

// ContextMonitor предоставляет мониторинг для FlowContext
type ContextMonitor struct {
	contextID    string
	startTime    time.Time
	memoryStats  *MemoryStats
	operations   map[string]int
	errorCount   int
	mutex        sync.RWMutex
	logger       interfaces.Logger
}

// MemoryStats отслеживает статистику памяти
type MemoryStats struct {
	InitialMB    float64 `json:"initial_mb"`
	CurrentMB    float64 `json:"current_mb"`
	PeakMB       float64 `json:"peak_mb"`
	ImageCacheMB int64   `json:"image_cache_mb"`
}

// NewContextMonitor создает монитор для контекста
func NewContextMonitor(contextID string, logger interfaces.Logger) *ContextMonitor {
	initialMemory := getMemoryUsageMB()

	monitor := &ContextMonitor{
		contextID:   contextID,
		startTime:   time.Now(),
		memoryStats: &MemoryStats{
			InitialMB: initialMemory,
			CurrentMB: initialMemory,
			PeakMB:    initialMemory,
		},
		operations: make(map[string]int),
		logger:     logger,
	}

	// Update global metrics
	globalMetrics.mutex.Lock()
	globalMetrics.TotalContexts++
	globalMetrics.ActiveContexts++
	globalMetrics.mutex.Unlock()

	return monitor
}

// RecordOperation записывает операцию
func (cm *ContextMonitor) RecordOperation(operation string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.operations[operation]++
	cm.updateMemoryStats()
}

// RecordError записывает ошибку
func (cm *ContextMonitor) RecordError(err error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.errorCount++
	globalMetrics.mutex.Lock()
	globalMetrics.Errors++
	globalMetrics.mutex.Unlock()

	cm.logger.Error("Context operation error",
		"context_id", cm.contextID,
		"error", err.Error(),
		"total_errors", cm.errorCount,
	)
}

// RecordMemoryUsage записывает использование памяти
func (cm *ContextMonitor) RecordMemoryUsage(imageCacheMB int64) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.memoryStats.ImageCacheMB = imageCacheMB
	cm.updateMemoryStats()

	// Update global peak if this context has higher usage
	globalMemoryMB := getMemoryUsageMB()
	if int64(globalMemoryMB) > globalMetrics.PeakMemoryUsage {
		globalMetrics.mutex.Lock()
		globalMetrics.PeakMemoryUsage = int64(globalMemoryMB)
		globalMetrics.mutex.Unlock()
	}
}

// updateMemoryStats обновляет статистику памяти
func (cm *ContextMonitor) updateMemoryStats() {
	currentMemory := getMemoryUsageMB()
	cm.memoryStats.CurrentMB = currentMemory

	if currentMemory > cm.memoryStats.PeakMB {
		cm.memoryStats.PeakMB = currentMemory
	}
}

// GetStats возвращает статистику
func (cm *ContextMonitor) GetStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	duration := time.Since(cm.startTime)

	return map[string]interface{}{
		"context_id":           cm.contextID,
		"duration_ms":          duration.Milliseconds(),
		"duration_formatted":    duration.String(),
		"operations":           cm.operations,
		"operation_count":      len(cm.operations),
		"error_count":          cm.errorCount,
		"memory":               cm.memoryStats,
		"operations_per_second": float64(len(cm.operations)) / duration.Seconds(),
	}
}

// Close завершает мониторинг и обновляет глобальные метрики
func (cm *ContextMonitor) Close() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	duration := time.Since(cm.startTime)

	// Update global metrics
	globalMetrics.mutex.Lock()
	globalMetrics.ActiveContexts--

	// Update average memory per context
	totalMemory := globalMetrics.AvgMemoryPerContext * float64(globalMetrics.TotalContexts-1)
	currentMemory := cm.memoryStats.CurrentMB
	globalMetrics.AvgMemoryPerContext = (totalMemory + currentMemory) / float64(globalMetrics.TotalContexts)

	globalMetrics.mutex.Unlock()

	cm.logger.Info("Context monitor closed",
		"context_id", cm.contextID,
		"duration", duration.String(),
		"operations", len(cm.operations),
		"memory_peak_mb", cm.memoryStats.PeakMB,
		"errors", cm.errorCount,
	)
}

// GetGlobalMetrics возвращает глобальные метрики
func GetGlobalMetrics() map[string]interface{} {
	globalMetrics.mutex.RLock()
	defer globalMetrics.mutex.RUnlock()

	return map[string]interface{}{
		"total_contexts":        globalMetrics.TotalContexts,
		"active_contexts":       globalMetrics.ActiveContexts,
		"avg_memory_per_context_mb": globalMetrics.AvgMemoryPerContext,
		"peak_memory_usage_mb":   globalMetrics.PeakMemoryUsage,
		"total_operations":       globalMetrics.TotalOperations,
		"total_errors":           globalMetrics.Errors,
		"system_memory_mb":       getMemoryUsageMB(),
	}
}

// resetGlobalMetrics сбрасывает глобальные метрики
func resetGlobalMetrics() {
	globalMetrics.mutex.Lock()
	defer globalMetrics.mutex.Unlock()

	*globalMetrics = ContextMetrics{}
}

// getMemoryUsageMB получает использование памяти в MB
func getMemoryUsageMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

// ContextHealthChecker проверяет здоровье контекстов
type ContextHealthChecker struct {
	warningThresholdMB float64
	errorThresholdMB   float64
	checkInterval      time.Duration
	logger             interfaces.Logger
}

// NewContextHealthChecker создает health checker
func NewContextHealthChecker(logger interfaces.Logger) *ContextHealthChecker {
	return &ContextHealthChecker{
		warningThresholdMB: 50.0,  // Предупреждение при 50MB
		errorThresholdMB:   100.0, // Ошибка при 100MB
		checkInterval:      30 * time.Second,
		logger:             logger,
	}
}

// CheckContextHealth проверяет здоровье контекста
func (chc *ContextHealthChecker) CheckContextHealth(flowCtx interfaces.FlowContext) error {
	currentMemory := getMemoryUsageMB()

	if currentMemory > chc.errorThresholdMB {
		return fmt.Errorf("critical memory usage: %.2f MB (threshold: %.2f MB)",
			currentMemory, chc.errorThresholdMB)
	}

	if currentMemory > chc.warningThresholdMB {
		chc.logger.Warn("High memory usage detected",
			"current_mb", currentMemory,
			"warning_threshold_mb", chc.warningThresholdMB,
		)
	}

	return nil
}

// MonitorActiveContexts мониторит активные контексты
func (chc *ContextHealthChecker) MonitorActiveContexts() {
	metrics := GetGlobalMetrics()

	activeContexts := metrics["active_contexts"].(int64)
	avgMemory := metrics["avg_memory_per_context_mb"].(float64)
	peakMemory := float64(metrics["peak_memory_usage_mb"].(int64))

	estimatedTotalMemory := float64(activeContexts) * avgMemory

	if estimatedTotalMemory > chc.warningThresholdMB*2 {
		chc.logger.Warn("High estimated memory usage across contexts",
			"active_contexts", activeContexts,
			"avg_memory_per_context_mb", avgMemory,
			"estimated_total_mb", estimatedTotalMemory,
			"peak_mb", peakMemory,
		)
	}
}

// ContextProfiler создает профилирование контекста
type ContextProfiler struct {
	contextID string
	startTime time.Time
	events    []ProfileEvent
	mutex     sync.RWMutex
	logger    interfaces.Logger
}

// ProfileEvent представляет событие в профиле
type ProfileEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`      // "set", "get", "load_image", etc.
	Key       string    `json:"key"`
	Duration  int64     `json:"duration_ns"`
	MemoryMB  float64   `json:"memory_mb"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// NewContextProfiler создает профилировщик
func NewContextProfiler(contextID string, logger interfaces.Logger) *ContextProfiler {
	return &ContextProfiler{
		contextID: contextID,
		startTime: time.Now(),
		events:    make([]ProfileEvent, 0),
		logger:    logger,
	}
}

// ProfileOperation профилирует операцию
func (cp *ContextProfiler) ProfileOperation(operation, key string, duration time.Duration, memoryMB float64, err error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	event := ProfileEvent{
		Timestamp: time.Now(),
		Type:      operation,
		Key:       key,
		Duration:  duration.Nanoseconds(),
		MemoryMB:  memoryMB,
		Success:   err == nil,
	}

	if err != nil {
		event.Error = err.Error()
	}

	cp.events = append(cp.events, event)
}

// GetProfile возвращает профиль
func (cp *ContextProfiler) GetProfile() map[string]interface{} {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	totalDuration := time.Duration(0)
	totalMemory := float64(0)
	operations := make(map[string]struct {
		Count     int     `json:"count"`
		TotalTime int64   `json:"total_time_ns"`
		AvgTime   float64 `json:"avg_time_ms"`
		MaxTime   int64   `json:"max_time_ns"`
		Errors    int     `json:"errors"`
	})

	for _, event := range cp.events {
		totalDuration += time.Duration(event.Duration)
		totalMemory += event.MemoryMB

		op := operations[event.Type]
		op.Count++
		op.TotalTime += event.Duration
		if event.Duration > op.MaxTime {
			op.MaxTime = event.Duration
		}
		if !event.Success {
			op.Errors++
		}
	}

	// Calculate averages
	for opType, op := range operations {
		if op.Count > 0 {
			op.AvgTime = float64(op.TotalTime) / float64(op.Count) / 1e6 // Convert to ms
			delete(operations, opType)
			operations[opType] = op
		}
	}

	return map[string]interface{}{
		"context_id":       cp.contextID,
		"duration":         time.Since(cp.startTime).String(),
		"total_events":     len(cp.events),
		"total_duration":   totalDuration.String(),
		"total_memory_mb":  totalMemory,
		"operations":       operations,
		"operations_per_second": float64(len(cp.events)) / time.Since(cp.startTime).Seconds(),
	}
}