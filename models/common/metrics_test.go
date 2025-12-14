package common

import (
	"fmt"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

func TestMetricsCollector_RecordRequest(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := DefaultMetricsConfig()
	collector := NewMetricsCollector(config, logger)

	tests := []struct {
		name    string
		metrics *ModelRequestMetrics
	}{
		{
			name: "successful request",
			metrics: &ModelRequestMetrics{
				Provider:    ProviderDeepSeek,
				Model:       "deepseek-chat",
				RequestType: "generation",
				Duration:    100 * time.Millisecond,
				Success:     true,
				TokenCount:  50,
				Timestamp:   time.Now(),
			},
		},
		{
			name: "failed request",
			metrics: &ModelRequestMetrics{
				Provider:     ProviderZAI,
				Model:        "glm-4.6v",
				RequestType:  "vision",
				Duration:     200 * time.Millisecond,
				Success:      false,
				TokenCount:   0,
				ErrorCode:    "rate_limit",
				ErrorMessage: "Rate limit exceeded",
				Timestamp:    time.Now(),
			},
		},
		{
			name: "nil metrics",
			metrics: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset metrics before each test
			collector.ResetMetrics()

			collector.RecordRequest(tt.metrics)

			if tt.metrics == nil {
				// Should not panic with nil metrics
				return
			}

			// Check global metrics
			globalMetrics := collector.GetGlobalMetrics()
			if globalMetrics.TotalRequests != 1 {
				t.Errorf("Expected 1 total request, got %d", globalMetrics.TotalRequests)
			}

			if tt.metrics.Success {
				if globalMetrics.SuccessCount != 1 {
					t.Errorf("Expected 1 success, got %d", globalMetrics.SuccessCount)
				}
				if globalMetrics.ErrorCount != 0 {
					t.Errorf("Expected 0 errors, got %d", globalMetrics.ErrorCount)
				}
			} else {
				if globalMetrics.SuccessCount != 0 {
					t.Errorf("Expected 0 successes, got %d", globalMetrics.SuccessCount)
				}
				if globalMetrics.ErrorCount != 1 {
					t.Errorf("Expected 1 error, got %d", globalMetrics.ErrorCount)
				}
			}

			// Check provider-specific metrics
			providerMetrics := collector.GetProviderMetrics(tt.metrics.Provider)
			if providerMetrics == nil {
				t.Error("Expected provider metrics to be created")
			} else {
				if providerMetrics.RequestCount != 1 {
					t.Errorf("Expected 1 provider request, got %d", providerMetrics.RequestCount)
				}
				if providerMetrics.TotalTokens != int64(tt.metrics.TokenCount) {
					t.Errorf("Expected %d tokens, got %d", tt.metrics.TokenCount, providerMetrics.TotalTokens)
				}
			}
		})
	}
}

func TestMetricsCollector_SuccessRateCalculation(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := DefaultMetricsConfig()
	collector := NewMetricsCollector(config, logger)

	// Reset metrics
	collector.ResetMetrics()

	// Record multiple requests with mixed success/failure
	testMetrics := []*ModelRequestMetrics{
		{
			Provider: ProviderDeepSeek,
			Model:    "deepseek-chat",
			Duration: 100 * time.Millisecond,
			Success:  true,
			TokenCount: 50,
			Timestamp: time.Now(),
		},
		{
			Provider: ProviderDeepSeek,
			Model:    "deepseek-chat",
			Duration: 150 * time.Millisecond,
			Success:  true,
			TokenCount: 75,
			Timestamp: time.Now(),
		},
		{
			Provider: ProviderDeepSeek,
			Model:    "deepseek-chat",
			Duration: 200 * time.Millisecond,
			Success:  false,
			TokenCount: 0,
			ErrorCode: "timeout",
			Timestamp: time.Now(),
		},
		{
			Provider: ProviderZAI,
			Model:    "glm-4.6v",
			Duration: 300 * time.Millisecond,
			Success:  true,
			TokenCount: 100,
			Timestamp: time.Now(),
		},
		{
			Provider: ProviderZAI,
			Model:    "glm-4.6v",
			Duration: 400 * time.Millisecond,
			Success:  false,
			TokenCount: 0,
			ErrorCode: "server_error",
			Timestamp: time.Now(),
		},
	}

	// Record all metrics
	for _, metric := range testMetrics {
		collector.RecordRequest(metric)
	}

	// Check global metrics
	globalMetrics := collector.GetGlobalMetrics()
	
	// Total requests should be 5
	if globalMetrics.TotalRequests != 5 {
		t.Errorf("Expected 5 total requests, got %d", globalMetrics.TotalRequests)
	}

	// Success count should be 3
	if globalMetrics.SuccessCount != 3 {
		t.Errorf("Expected 3 successes, got %d", globalMetrics.SuccessCount)
	}

	// Error count should be 2
	if globalMetrics.ErrorCount != 2 {
		t.Errorf("Expected 2 errors, got %d", globalMetrics.ErrorCount)
	}

	// Check by-model metrics
	deepseekMetrics, exists := globalMetrics.ByModel["deepseek"]
	if !exists {
		t.Error("Expected DeepSeek model metrics")
	} else {
		// DeepSeek: 2 successes, 1 failure = 66.67% success rate
		expectedSuccessRate := float64(2) / float64(3)
		if deepseekMetrics.SuccessRate != expectedSuccessRate {
			t.Errorf("DeepSeek success rate: expected %f, got %f", expectedSuccessRate, deepseekMetrics.SuccessRate)
		}
		if deepseekMetrics.Requests != 3 {
			t.Errorf("DeepSeek requests: expected 3, got %d", deepseekMetrics.Requests)
		}
	}

	zaiMetrics, exists := globalMetrics.ByModel["zai"]
	if !exists {
		t.Error("Expected ZAI model metrics")
	} else {
		// ZAI: 1 success, 1 failure = 50% success rate
		expectedSuccessRate := float64(1) / float64(2)
		if zaiMetrics.SuccessRate != expectedSuccessRate {
			t.Errorf("ZAI success rate: expected %f, got %f", expectedSuccessRate, zaiMetrics.SuccessRate)
		}
		if zaiMetrics.Requests != 2 {
			t.Errorf("ZAI requests: expected 2, got %d", zaiMetrics.Requests)
		}
	}
}

func TestMetricsCollector_AverageLatencyCalculation(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := DefaultMetricsConfig()
	collector := NewMetricsCollector(config, logger)

	// Reset metrics
	collector.ResetMetrics()

	// Record requests with different latencies
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
	}

	for i, duration := range durations {
		metric := &ModelRequestMetrics{
			Provider:    ProviderDeepSeek,
			Model:       "deepseek-chat",
			RequestType: "generation",
			Duration:    duration,
			Success:     true,
			TokenCount:  50,
			Timestamp:   time.Now(),
		}
		collector.RecordRequest(metric)

		// Check average latency after each request
		globalMetrics := collector.GetGlobalMetrics()
		expectedAvg := float64(0)
		for j := 0; j <= i; j++ {
			expectedAvg += float64(durations[j].Milliseconds())
		}
		expectedAvg /= float64(i + 1)

		if globalMetrics.AvgLatency != expectedAvg {
			t.Errorf("After %d requests: expected avg latency %f, got %f", i+1, expectedAvg, globalMetrics.AvgLatency)
		}
	}
}

func TestMetricsCollector_ConcurrentAccess(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := DefaultMetricsConfig()
	collector := NewMetricsCollector(config, logger)

	// Reset metrics
	collector.ResetMetrics()

	const numGoroutines = 20
	const numRequests = 50

	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numRequests; j++ {
				metric := &ModelRequestMetrics{
					Provider:    ProviderDeepSeek,
					Model:       "deepseek-chat",
					RequestType: "generation",
					Duration:    time.Duration(id*10+j) * time.Millisecond,
					Success:     j%3 != 0, // 2/3 success rate
					TokenCount:  (id + j) % 100,
					Timestamp:   time.Now(),
				}

				collector.RecordRequest(metric)

				// Also test concurrent reads
				globalMetrics := collector.GetGlobalMetrics()
				providerMetrics := collector.GetProviderMetrics(ProviderDeepSeek)
				systemMetrics := collector.GetSystemMetrics()

				if globalMetrics == nil || providerMetrics == nil || systemMetrics == nil {
					errors <- fmt.Errorf("goroutine %d: nil metrics returned", id)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Error(err)
	}

	// Verify final metrics
	globalMetrics := collector.GetGlobalMetrics()
	expectedTotalRequests := numGoroutines * numRequests
	if globalMetrics.TotalRequests != int64(expectedTotalRequests) {
		t.Errorf("Expected %d total requests, got %d", expectedTotalRequests, globalMetrics.TotalRequests)
	}

	// Verify error summary
	_, errorsByProvider := collector.GetErrorSummary()
	if errorsByProvider[ProviderDeepSeek] == 0 {
		t.Error("Expected some errors for DeepSeek provider")
	}
}

func TestMetricsCollector_ThresholdChecking(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := &MetricsConfig{
		Enabled:            true,
		CollectionInterval:  30 * time.Second,
		RetentionPeriod:    24 * time.Hour,
		LatencyThreshold:   50 * time.Millisecond,
		ErrorRateThreshold: 0.1, // 10%
	}
	collector := NewMetricsCollector(config, logger)

	// Reset metrics
	collector.ResetMetrics()

	// Record a request that exceeds latency threshold
	slowMetric := &ModelRequestMetrics{
		Provider:    ProviderDeepSeek,
		Model:       "deepseek-chat",
		RequestType: "generation",
		Duration:    100 * time.Millisecond, // Exceeds 50ms threshold
		Success:     true,
		TokenCount:  50,
		Timestamp:   time.Now(),
	}

	collector.RecordRequest(slowMetric)

	// Record multiple requests to trigger error rate threshold
	for i := 0; i < 10; i++ {
		metric := &ModelRequestMetrics{
			Provider:    ProviderDeepSeek,
			Model:       "deepseek-chat",
			RequestType: "generation",
			Duration:    30 * time.Millisecond,
			Success:     i < 2, // Only 2 successes out of 10 = 20% error rate
			TokenCount:  50,
			Timestamp:   time.Now(),
		}
		collector.RecordRequest(metric)
	}

	// Check health status - should be degraded due to high error rate
	health := collector.GetHealthStatus()
	if health.Status == "healthy" {
		t.Error("Expected degraded health status due to high error rate")
	}
}

func TestMetricsCollector_ResetMetrics(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := DefaultMetricsConfig()
	collector := NewMetricsCollector(config, logger)

	// Record some metrics
	metric := &ModelRequestMetrics{
		Provider:    ProviderDeepSeek,
		Model:       "deepseek-chat",
		RequestType: "generation",
		Duration:    100 * time.Millisecond,
		Success:     true,
		TokenCount:  50,
		Timestamp:   time.Now(),
	}

	collector.RecordRequest(metric)
	collector.RecordRequest(metric)
	collector.RecordRequest(metric)

	// Verify metrics are recorded
	globalMetrics := collector.GetGlobalMetrics()
	if globalMetrics.TotalRequests != 3 {
		t.Errorf("Expected 3 requests before reset, got %d", globalMetrics.TotalRequests)
	}

	// Reset metrics
	collector.ResetMetrics()

	// Verify metrics are cleared
	globalMetrics = collector.GetGlobalMetrics()
	if globalMetrics.TotalRequests != 0 {
		t.Errorf("Expected 0 requests after reset, got %d", globalMetrics.TotalRequests)
	}
	if globalMetrics.SuccessCount != 0 {
		t.Errorf("Expected 0 successes after reset, got %d", globalMetrics.SuccessCount)
	}
	if globalMetrics.ErrorCount != 0 {
		t.Errorf("Expected 0 errors after reset, got %d", globalMetrics.ErrorCount)
	}

	// Check provider metrics are also reset
	providerMetrics := collector.GetProviderMetrics(ProviderDeepSeek)
	if providerMetrics.RequestCount != 0 {
		t.Errorf("Expected 0 provider requests after reset, got %d", providerMetrics.RequestCount)
	}

	// Check error summary is also reset
	errorsByType, errorsByProvider := collector.GetErrorSummary()
	if len(errorsByType) != 0 || len(errorsByProvider) != 0 {
		t.Error("Expected error summary to be empty after reset")
	}
}

func TestMetricsCollector_ExportMetrics(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	config := DefaultMetricsConfig()
	collector := NewMetricsCollector(config, logger)

	// Record some metrics
	metric := &ModelRequestMetrics{
		Provider:    ProviderDeepSeek,
		Model:       "deepseek-chat",
		RequestType: "generation",
		Duration:    100 * time.Millisecond,
		Success:     true,
		TokenCount:  50,
		Timestamp:   time.Now(),
	}

	collector.RecordRequest(metric)

	// Export metrics
	exported := collector.ExportMetrics()

	// Check exported structure
	if exported == nil {
		t.Error("Exported metrics should not be nil")
	}

	// Check required fields exist
	requiredFields := []string{"timestamp", "global", "system", "providers", "errors"}
	for _, field := range requiredFields {
		if _, exists := exported[field]; !exists {
			t.Errorf("Missing required field in exported metrics: %s", field)
		}
	}

	// Check global metrics in export
	globalMetrics, ok := exported["global"].(*interfaces.GenerationMetrics)
	if !ok {
		t.Error("Global metrics should be of type GenerationMetrics")
	}
	if globalMetrics.TotalRequests != 1 {
		t.Errorf("Exported global metrics should have 1 request, got %d", globalMetrics.TotalRequests)
	}
}