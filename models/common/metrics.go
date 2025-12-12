package common

import (
	"runtime"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// MetricsCollector handles performance monitoring and metrics collection
type MetricsCollector struct {
	logger interfaces.Logger
	mu     sync.RWMutex

	// Request metrics
	requestCount    int64
	successCount    int64
	errorCount      int64
	totalLatency    time.Duration
	totalTokens     int64

	// Provider-specific metrics
	providerMetrics map[Provider]*ProviderMetrics

	// Error tracking
	errorsByType    map[string]int64
	errorsByProvider map[Provider]int64

	// System metrics
	startTime time.Time

	// Performance thresholds
	latencyThreshold time.Duration
	errorRateThreshold float64
}

// ProviderMetrics represents metrics for a specific provider
type ProviderMetrics struct {
	Provider       Provider `json:"provider"`
	RequestCount   int64     `json:"request_count"`
	SuccessCount   int64     `json:"success_count"`
	ErrorCount     int64     `json:"error_count"`
	AvgLatency     float64   `json:"avg_latency_ms"`
	TotalTokens    int64     `json:"total_tokens"`
	LastRequest    time.Time `json:"last_request"`
	LastError      time.Time `json:"last_error,omitempty"`
}

// ModelRequestMetrics represents metrics for a single request
type ModelRequestMetrics struct {
	Provider      Provider        `json:"provider"`
	Model         string          `json:"model"`
	RequestType   string          `json:"request_type"` // generation, tool, flow
	Duration      time.Duration   `json:"duration_ms"`
	Success       bool            `json:"success"`
	TokenCount    int             `json:"token_count"`
	ErrorCode     string          `json:"error_code,omitempty"`
	ErrorMessage  string          `json:"error_message,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}

// MetricsConfig represents configuration for metrics collection
type MetricsConfig struct {
	Enabled           bool          `json:"enabled"`
	CollectionInterval time.Duration `json:"collection_interval"`
	RetentionPeriod   time.Duration `json:"retention_period"`
	LatencyThreshold  time.Duration `json:"latency_threshold"`
	ErrorRateThreshold float64      `json:"error_rate_threshold"`
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:           true,
		CollectionInterval: 30 * time.Second,
		RetentionPeriod:   24 * time.Hour,
		LatencyThreshold:  2 * time.Second,
		ErrorRateThreshold: 0.05, // 5%
	}
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config *MetricsConfig, logger interfaces.Logger) *MetricsCollector {
	if config == nil {
		config = DefaultMetricsConfig()
	}
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	collector := &MetricsCollector{
		logger:            logger,
		providerMetrics:    make(map[Provider]*ProviderMetrics),
		errorsByType:      make(map[string]int64),
		errorsByProvider:   make(map[Provider]int64),
		startTime:         time.Now(),
		latencyThreshold:  config.LatencyThreshold,
		errorRateThreshold: config.ErrorRateThreshold,
	}

	// Initialize provider metrics
	for _, provider := range []Provider{ProviderDeepSeek, ProviderZAI, ProviderOpenAI} {
		collector.providerMetrics[provider] = &ProviderMetrics{
			Provider: provider,
		}
	}

	if config.Enabled {
		collector.startCollection(config.CollectionInterval)
	}

	logger.Info("Metrics collector initialized",
		"enabled", config.Enabled,
		"latency_threshold", config.LatencyThreshold,
		"error_rate_threshold", config.ErrorRateThreshold)

	return collector
}

// RecordRequest records metrics for a completed request
func (mc *MetricsCollector) RecordRequest(metrics *ModelRequestMetrics) {
	if metrics == nil {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Update global metrics
	mc.requestCount++
	mc.totalLatency += metrics.Duration
	mc.totalTokens += int64(metrics.TokenCount)

	if metrics.Success {
		mc.successCount++
	} else {
		mc.errorCount++
		mc.errorsByProvider[metrics.Provider]++
		if metrics.ErrorCode != "" {
			mc.errorsByType[metrics.ErrorCode]++
		}
	}

	// Update provider-specific metrics
	providerMetrics := mc.providerMetrics[metrics.Provider]
	if providerMetrics == nil {
		providerMetrics = &ProviderMetrics{Provider: metrics.Provider}
		mc.providerMetrics[metrics.Provider] = providerMetrics
	}

	providerMetrics.RequestCount++
	providerMetrics.TotalTokens += int64(metrics.TokenCount)
	providerMetrics.LastRequest = metrics.Timestamp

	if metrics.Success {
		providerMetrics.SuccessCount++
	} else {
		providerMetrics.ErrorCount++
		providerMetrics.LastError = metrics.Timestamp
	}

	// Update average latency
	providerMetrics.AvgLatency = float64(metrics.Duration.Milliseconds())

	// Log if thresholds are exceeded
	mc.checkThresholds(metrics)

	mc.logger.Debug("Request metrics recorded",
		"provider", metrics.Provider,
		"model", metrics.Model,
		"duration_ms", metrics.Duration.Milliseconds(),
		"success", metrics.Success,
		"token_count", metrics.TokenCount)
}

// GetProviderMetrics returns metrics for a specific provider
func (mc *MetricsCollector) GetProviderMetrics(provider Provider) *ProviderMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := mc.providerMetrics[provider]
	if metrics == nil {
		return nil
	}

	// Return a copy to prevent modification
	return &ProviderMetrics{
		Provider:     metrics.Provider,
		RequestCount: metrics.RequestCount,
		SuccessCount: metrics.SuccessCount,
		ErrorCount:   metrics.ErrorCount,
		AvgLatency:   metrics.AvgLatency,
		TotalTokens:  metrics.TotalTokens,
		LastRequest:  metrics.LastRequest,
		LastError:    metrics.LastError,
	}
}

// GetAllProviderMetrics returns metrics for all providers
func (mc *MetricsCollector) GetAllProviderMetrics() map[Provider]*ProviderMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[Provider]*ProviderMetrics)
	for provider, metrics := range mc.providerMetrics {
		result[provider] = &ProviderMetrics{
			Provider:     metrics.Provider,
			RequestCount: metrics.RequestCount,
			SuccessCount: metrics.SuccessCount,
			ErrorCount:   metrics.ErrorCount,
			AvgLatency:   metrics.AvgLatency,
			TotalTokens:  metrics.TotalTokens,
			LastRequest:  metrics.LastRequest,
			LastError:    metrics.LastError,
		}
	}

	return result
}

// GetSystemMetrics returns current system metrics
func (mc *MetricsCollector) GetSystemMetrics() *interfaces.SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return &interfaces.SystemMetrics{
		MemoryUsage:    int64(m.Alloc),
		CPUUsage:       mc.getCPUUsage(),
		GoroutineCount: int64(runtime.NumGoroutine()),
		GCCount:        int64(m.NumGC),
		HeapSize:       int64(m.HeapSys),
		HeapAlloc:      int64(m.HeapAlloc),
	}
}

// GetGlobalMetrics returns global metrics across all providers
func (mc *MetricsCollector) GetGlobalMetrics() *interfaces.GenerationMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	avgLatency := float64(0)
	if mc.requestCount > 0 {
		avgLatency = float64(mc.totalLatency.Milliseconds()) / float64(mc.requestCount)
	}

	// Calculate by-model metrics
	byModel := make(map[string]*interfaces.ModelMetrics)
	for _, providerMetrics := range mc.providerMetrics {
		modelName := string(providerMetrics.Provider)
		successRate := float64(0)
		if providerMetrics.RequestCount > 0 {
			successRate = float64(providerMetrics.SuccessCount) / float64(providerMetrics.RequestCount)
		}

		byModel[modelName] = &interfaces.ModelMetrics{
			Requests:    providerMetrics.RequestCount,
			SuccessRate: successRate,
			AvgLatency:  providerMetrics.AvgLatency,
			TotalTokens: providerMetrics.TotalTokens,
		}
	}

	return &interfaces.GenerationMetrics{
		TotalRequests: mc.requestCount,
		SuccessCount:  mc.successCount,
		ErrorCount:    mc.errorCount,
		AvgLatency:    avgLatency,
		TotalTokens:   mc.totalTokens,
		ByModel:       byModel,
	}
}

// ResetMetrics resets all collected metrics
func (mc *MetricsCollector) ResetMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.requestCount = 0
	mc.successCount = 0
	mc.errorCount = 0
	mc.totalLatency = 0
	mc.totalTokens = 0
	mc.errorsByType = make(map[string]int64)
	mc.errorsByProvider = make(map[Provider]int64)
	mc.startTime = time.Now()

	for _, providerMetrics := range mc.providerMetrics {
		providerMetrics.RequestCount = 0
		providerMetrics.SuccessCount = 0
		providerMetrics.ErrorCount = 0
		providerMetrics.AvgLatency = 0
		providerMetrics.TotalTokens = 0
		providerMetrics.LastRequest = time.Time{}
		providerMetrics.LastError = time.Time{}
	}

	mc.logger.Info("Metrics reset")
}

// GetErrorSummary returns a summary of errors by type and provider
func (mc *MetricsCollector) GetErrorSummary() (map[string]int64, map[Provider]int64) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	errorsByType := make(map[string]int64)
	for k, v := range mc.errorsByType {
		errorsByType[k] = v
	}

	errorsByProvider := make(map[Provider]int64)
	for k, v := range mc.errorsByProvider {
		errorsByProvider[k] = v
	}

	return errorsByType, errorsByProvider
}

// GetHealthStatus returns the health status of the metrics collector
func (mc *MetricsCollector) GetHealthStatus() *interfaces.ComponentHealth {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	status := "healthy"
	message := "All systems operational"

	// Check error rate
	if mc.requestCount > 0 {
		errorRate := float64(mc.errorCount) / float64(mc.requestCount)
		if errorRate > mc.errorRateThreshold {
			status = "degraded"
			message = "High error rate detected"
		}
	}

	// Check average latency
	if mc.requestCount > 0 {
		avgLatency := mc.totalLatency / time.Duration(mc.requestCount)
		if avgLatency > mc.latencyThreshold {
			status = "degraded"
			message = "High latency detected"
		}
	}

	return &interfaces.ComponentHealth{
		Status:  status,
		Message: message,
		Details: map[string]interface{}{
			"total_requests":    mc.requestCount,
			"error_rate":        float64(mc.errorCount) / float64(mc.requestCount),
			"avg_latency_ms":    float64(mc.totalLatency.Milliseconds()) / float64(mc.requestCount),
			"uptime_seconds":    time.Since(mc.startTime).Seconds(),
		},
	}
}

// checkThresholds checks if metrics exceed configured thresholds
func (mc *MetricsCollector) checkThresholds(metrics *ModelRequestMetrics) {
	// Check latency threshold
	if metrics.Duration > mc.latencyThreshold {
		mc.logger.Warn("Request latency threshold exceeded",
			"provider", metrics.Provider,
			"duration_ms", metrics.Duration.Milliseconds(),
			"threshold_ms", mc.latencyThreshold.Milliseconds(),
			"model", metrics.Model)
	}

	// Check error rate for provider
	providerMetrics := mc.providerMetrics[metrics.Provider]
	if providerMetrics != nil && providerMetrics.RequestCount >= 10 {
		errorRate := float64(providerMetrics.ErrorCount) / float64(providerMetrics.RequestCount)
		if errorRate > mc.errorRateThreshold {
			mc.logger.Warn("Provider error rate threshold exceeded",
				"provider", metrics.Provider,
				"error_rate", errorRate,
				"threshold", mc.errorRateThreshold,
				"requests", providerMetrics.RequestCount,
				"errors", providerMetrics.ErrorCount)
		}
	}
}

// getCPUUsage estimates CPU usage (simplified implementation)
func (mc *MetricsCollector) getCPUUsage() float64 {
	// This is a simplified implementation
	// In a production environment, you'd want to use a proper CPU monitoring library
	return 0.0
}

// startCollection starts the metrics collection goroutine
func (mc *MetricsCollector) startCollection(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			mc.collectSystemMetrics()
		}
	}()
}

// collectSystemMetrics collects system metrics periodically
func (mc *MetricsCollector) collectSystemMetrics() {
	systemMetrics := mc.GetSystemMetrics()

	mc.logger.Debug("System metrics collected",
		"memory_usage_mb", systemMetrics.MemoryUsage/1024/1024,
		"goroutines", systemMetrics.GoroutineCount,
		"gc_count", systemMetrics.GCCount,
		"heap_size_mb", systemMetrics.HeapSize/1024/1024,
		"heap_alloc_mb", systemMetrics.HeapAlloc/1024/1024)
}

// ExportMetrics exports metrics in a format suitable for external monitoring systems
func (mc *MetricsCollector) ExportMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := map[string]interface{}{
		"timestamp": time.Now(),
		"global":    mc.GetGlobalMetrics(),
		"system":    mc.GetSystemMetrics(),
		"providers": mc.GetAllProviderMetrics(),
	}

	errorsByType, errorsByProvider := mc.GetErrorSummary()
	result["errors"] = map[string]interface{}{
		"by_type":    errorsByType,
		"by_provider": errorsByProvider,
	}

	return result
}

// Close gracefully shuts down the metrics collector
func (mc *MetricsCollector) Close() error {
	mc.logger.Info("Metrics collector shutting down")
	return nil
}