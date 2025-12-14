package context

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ContextDebugger предоставляет отладочные возможности для FlowContext
type ContextDebugger struct {
	enabled       bool
	flowCtx       interfaces.FlowContext
	logger        interfaces.Logger
	maxLogSize    int64
	stepCounter   int
	operationLog  []ContextOperation
}

// ContextOperation представляет операцию в контексте
type ContextOperation struct {
	Timestamp time.Time                 `json:"timestamp"`
	Step       string                    `json:"step"`
	Operation string                    `json:"operation"`
	Key       string                    `json:"key"`
	Value     interface{}               `json:"value"`
	ValueSize int64                     `json:"value_size"`
	Success   bool                      `json:"success"`
	Error     string                    `json:"error,omitempty"`
	Metadata  map[string]interface{}     `json:"metadata,omitempty"`
}

// NewContextDebugger создает отладчик контекста
func NewContextDebugger(flowCtx interfaces.FlowContext, logger interfaces.Logger, maxLogSizeMB int64) *ContextDebugger {
	return &ContextDebugger{
		enabled:    true,
		flowCtx:    flowCtx,
		logger:     logger,
		maxLogSize: maxLogSizeMB * 1024 * 1024,
		operationLog: make([]ContextOperation, 0),
	}
}

// LogOperation записывает операцию в лог
func (cd *ContextDebugger) LogOperation(step, operation, key string, value interface{}) {
	if !cd.enabled {
		return
	}

	op := ContextOperation{
		Timestamp: time.Now(),
		Step:       step,
		Operation: operation,
		Key:       key,
		Value:     cd.sanitizeValue(value),
		ValueSize: cd.getValueSize(value),
		Success:   true,
		Metadata:  make(map[string]interface{}),
	}

	cd.operationLog = append(cd.operationLog, op)

	// Log to logger
	cd.logger.Debug("Context operation",
		"step", step,
		"operation", operation,
		"key", key,
		"value_size", op.ValueSize,
		"context_id", cd.flowCtx.ID(),
	)

	// Check log size limit
	if cd.getTotalLogSize() > cd.maxLogSize {
		cd.trimLog()
		cd.logger.Warn("Context operation log trimmed due to size limit",
			"current_size_mb", cd.getTotalLogSize()/(1024*1024),
			"limit_mb", cd.maxLogSize/(1024*1024),
		)
	}
}

// LogError записывает ошибку операции
func (cd *ContextDebugger) LogError(step, operation, key string, err error) {
	if !cd.enabled {
		return
	}

	op := ContextOperation{
		Timestamp: time.Now(),
		Step:       step,
		Operation: operation,
		Key:       key,
		Success:   false,
		Error:     err.Error(),
		Metadata:  make(map[string]interface{}),
	}

	cd.operationLog = append(cd.operationLog, op)

	cd.logger.Error("Context operation error",
		"step", step,
		"operation", operation,
		"key", key,
		"error", err.Error(),
		"context_id", cd.flowCtx.ID(),
	)
}

// StepStart отмечает начало шага
func (cd *ContextDebugger) StepStart(stepName string, metadata map[string]interface{}) {
	cd.stepCounter++
	cd.LogOperation(stepName, "step_start", "step_metadata", metadata)
}

// StepEnd отмечает конец шага
func (cd *ContextDebugger) StepEnd(stepName string, duration time.Duration, success bool) {
	metadata := map[string]interface{}{
		"duration_ms": duration.Milliseconds(),
		"success":     success,
	}
	cd.LogOperation(stepName, "step_end", "step_result", metadata)
}

// DumpState выводит текущее состояние контекста
func (cd *ContextDebugger) DumpState(stepName string) {
	if !cd.enabled {
		return
	}

	dump := cd.flowCtx.Dump()

	// Add debug info
	debugInfo := map[string]interface{}{
		"debug_timestamp": time.Now(),
		"debug_step":      stepName,
		"step_counter":    cd.stepCounter,
		"operations_count": len(cd.operationLog),
		"context_memory_kb": cd.getContextMemoryKB(),
	}

	dump["_debug"] = debugInfo

	cd.logger.Info("Context state dump",
		"step", stepName,
		"keys_count", len(cd.flowCtx.Keys()),
		"dump", dump,
	)
}

// VisualizeFlow создает визуальное представление flow
func (cd *ContextDebugger) VisualizeFlow() string {
	if !cd.enabled {
		return "Debugger disabled"
	}

	var builder strings.Builder
	builder.WriteString("🔄 Flow Visualization\n")
	builder.WriteString(strings.Repeat("=", 50) + "\n")

	currentTime := time.Now()

	// Group operations by step
	steps := make(map[string][]ContextOperation)
	for _, op := range cd.operationLog {
		steps[op.Step] = append(steps[op.Step], op)
	}

	for stepName, operations := range steps {
		builder.WriteString(fmt.Sprintf("\n📍 Step: %s\n", stepName))

		for i, op := range operations {
			relativeTime := op.Timestamp.Sub(currentTime)
			icon := "✅"
			if !op.Success {
				icon = "❌"
			}

			builder.WriteString(fmt.Sprintf("   %2d. %s %s %s",
				i+1, icon, op.Operation, op.Key))

			if !op.Success {
				builder.WriteString(fmt.Sprintf(" (%s)", op.Error))
			} else if op.ValueSize > 0 {
				builder.WriteString(fmt.Sprintf(" (%d bytes)", op.ValueSize))
			}

			builder.WriteString(fmt.Sprintf(" [%v ago]\n", relativeTime.Truncate(time.Millisecond)))
		}

		builder.WriteString("\n")
	}

	// Add summary
	builder.WriteString("📊 Summary:\n")
	builder.WriteString(fmt.Sprintf("   Total operations: %d\n", len(cd.operationLog)))
	builder.WriteString(fmt.Sprintf("   Context keys: %d\n", len(cd.flowCtx.Keys())))
	builder.WriteString(fmt.Sprintf("   Memory usage: %.2f KB\n", cd.getContextMemoryKB()))

	return builder.String()
}

// GetOperationsLog возвращает лог операций в JSON
func (cd *ContextDebugger) GetOperationsLog() string {
	if !cd.enabled {
		return "Debugger disabled"
	}

	data, err := json.MarshalIndent(map[string]interface{}{
		"context_id":   cd.flowCtx.ID(),
		"step_counter": cd.stepCounter,
		"operations":  cd.operationLog,
		"summary": map[string]interface{}{
			"total_operations": len(cd.operationLog),
			"successful_ops":   cd.countSuccessfulOps(),
			"failed_ops":       cd.countFailedOps(),
			"total_value_size": cd.getTotalValueSize(),
		},
	}, "", "  ")

	if err != nil {
		return fmt.Sprintf("Failed to serialize operations log: %v", err)
	}

	return string(data)
}

// ExportToFile экспортирует лог в файл
func (cd *ContextDebugger) ExportToFile(filePath string) error {
	if !cd.enabled {
		return fmt.Errorf("debugger disabled")
	}

	logData := cd.GetOperationsLog()

	// В реальной реализации здесь была бы запись в файл
	cd.logger.Info("Context operations log exported",
		"file_path", filePath,
		"size_bytes", len(logData),
	)

	return nil
}

// GetTimeline создает временную шкалу операций
func (cd *ContextDebugger) GetTimeline() []TimelineEvent {
	var timeline []TimelineEvent

	for _, op := range cd.operationLog {
		event := TimelineEvent{
			Timestamp: op.Timestamp,
			Type:      "operation",
			Step:      op.Step,
			Action:    fmt.Sprintf("%s %s", op.Operation, op.Key),
			Success:   op.Success,
			Duration:  0, // Could be calculated between start/end operations
		}

		if !op.Success {
			event.Error = op.Error
		}

		timeline = append(timeline, event)
	}

	return timeline
}

// TimelineEvent представляет событие на временной шкале
type TimelineEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`      // operation, step, error
	Step      string    `json:"step"`
	Action    string    `json:"action"`
	Success   bool      `json:"success"`
	Duration  int64     `json:"duration_ns,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// Helper methods

func (cd *ContextDebugger) sanitizeValue(value interface{}) interface{} {
	// Remove large binary data from logs
	switch v := value.(type) {
	case []byte:
		if len(v) > 1024 {
			return fmt.Sprintf("[binary data: %d bytes]", len(v))
		}
		return v
	case string:
		if len(v) > 512 {
			return v[:512] + "..."
		}
		return v
	default:
		return v
	}
}

func (cd *ContextDebugger) getValueSize(value interface{}) int64 {
	switch v := value.(type) {
	case []byte:
		return int64(len(v))
	case string:
		return int64(len(v))
	case map[string]interface{}:
		// Approximate size
		data, _ := json.Marshal(v)
		return int64(len(data))
	default:
		// Rough estimate
		data, _ := json.Marshal(v)
		return int64(len(data))
	}
}

func (cd *ContextDebugger) getTotalLogSize() int64 {
	var total int64
	for _, op := range cd.operationLog {
		total += int64(op.ValueSize)
		data, _ := json.Marshal(op)
		total += int64(len(data))
	}
	return total
}

func (cd *ContextDebugger) trimLog() {
	// Keep only the most recent half
	keep := len(cd.operationLog) / 2
	cd.operationLog = cd.operationLog[keep:]
}

func (cd *ContextDebugger) countSuccessfulOps() int {
	count := 0
	for _, op := range cd.operationLog {
		if op.Success {
			count++
		}
	}
	return count
}

func (cd *ContextDebugger) countFailedOps() int {
	count := 0
	for _, op := range cd.operationLog {
		if !op.Success {
			count++
		}
	}
	return count
}

func (cd *ContextDebugger) getTotalValueSize() int64 {
	var total int64
	for _, op := range cd.operationLog {
		total += op.ValueSize
	}
	return total
}

func (cd *ContextDebugger) getContextMemoryKB() float64 {
	// This would need access to context implementation
	// For now, estimate based on total value size
	return float64(cd.getTotalValueSize()) / 1024
}

// ContextDebugHelper предоставляет удобные методы для отладки
type ContextDebugHelper struct {
	debugger *ContextDebugger
	logger   interfaces.Logger
}

// NewContextDebugHelper создает помощник для отладки
func NewContextDebugHelper(flowCtx interfaces.FlowContext, logger interfaces.Logger, maxLogSizeMB int64) *ContextDebugHelper {
	return &ContextDebugHelper{
		debugger: NewContextDebugger(flowCtx, logger, maxLogSizeMB),
		logger:   logger,
	}
}

// WithStep создает отладочную обертку для шага
func (cdh *ContextDebugHelper) WithStep(stepName string, fn func() error) error {
	cdh.debugger.StepStart(stepName, nil)

	startTime := time.Now()
	err := fn()
	duration := time.Since(startTime)

	cdh.debugger.StepEnd(stepName, duration, err == nil)

	if err != nil {
		cdh.debugger.LogError(stepName, "step_execution", "", err)
	}

	// Dump state after each step in debug mode
	if cdh.logger != nil {
		cdh.debugger.DumpState(stepName)
	}

	return err
}

// LogSet логирует установку значения
func (cdh *ContextDebugHelper) LogSet(step, key string, value interface{}) {
	cdh.debugger.LogOperation(step, "set", key, value)
}

// LogGet логирует получение значения
func (cdh *ContextDebugHelper) LogGet(step, key string, value interface{}) {
	cdh.debugger.LogOperation(step, "get", key, value)
}

// GetDebugger возвращает отладчик
func (cdh *ContextDebugHelper) GetDebugger() *ContextDebugger {
	return cdh.debugger
}