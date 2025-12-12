package common

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// RetryState represents the state of a retry operation
type RetryState struct {
	Attempt       int           `json:"attempt"`
	MaxAttempts   int           `json:"max_attempts"`
	LastError    error         `json:"last_error"`
	TotalDelay    time.Duration `json:"total_delay"`
	NextDelay     time.Duration `json:"next_delay"`
	StartTime     time.Time     `json:"start_time"`
	ShouldRetry   bool          `json:"should_retry"`
}

// RetryExecutor handles retry logic for model operations
type RetryExecutor struct {
	config      RetryConfig
	logger      interfaces.Logger
	errorHandler *ErrorHandler
	randSource  rand.Source
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(config RetryConfig, logger interfaces.Logger) *RetryExecutor {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	// Use default config if not provided
	if config.MaxAttempts == 0 {
		config = DefaultRetryConfig
	}

	return &RetryExecutor{
		config:      config,
		logger:      logger,
		errorHandler: NewErrorHandler(ProviderCustom, logger),
		randSource:  rand.NewSource(time.Now().UnixNano()),
	}
}

// Execute executes a function with retry logic
func (re *RetryExecutor) Execute(ctx context.Context, operation func() error) error {
	state := &RetryState{
		Attempt:     0,
		MaxAttempts: re.config.MaxAttempts,
		StartTime:   time.Now(),
		ShouldRetry:  true,
	}

	for state.ShouldRetry && state.Attempt < state.MaxAttempts {
		state.Attempt++

		// Execute the operation
		err := operation()
		if err == nil {
			re.logger.Debug("Operation succeeded",
				"attempt", state.Attempt,
				"total_delay", state.TotalDelay,
				"duration", time.Since(state.StartTime))
			return nil
		}

		state.LastError = err

		// Check if error is retryable
		if !re.errorHandler.IsRetryableError(err) {
			re.logger.Debug("Error is not retryable",
				"attempt", state.Attempt,
				"error", err.Error())
			return err
		}

		// Calculate delay for next attempt
		state.NextDelay = re.calculateDelay(state.Attempt, err)
		state.TotalDelay += state.NextDelay
		state.ShouldRetry = state.Attempt < state.MaxAttempts

		re.logger.Debug("Operation failed, scheduling retry",
			"attempt", state.Attempt,
			"max_attempts", state.MaxAttempts,
			"error", err.Error(),
			"next_delay", state.NextDelay,
			"should_retry", state.ShouldRetry)

		// If this is the last attempt, don't wait
		if !state.ShouldRetry {
			break
		}

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		case <-time.After(state.NextDelay):
			// Continue to next attempt
		}
	}

	// All attempts exhausted
	return fmt.Errorf("operation failed after %d attempts, last error: %w", state.Attempt, state.LastError)
}

// Note: Generic methods not supported in current Go version
// Use Execute method and handle results in calling code

// calculateDelay calculates the delay for the next retry attempt
func (re *RetryExecutor) calculateDelay(attempt int, err error) time.Duration {
	baseDelay := re.config.BaseDelay
	maxDelay := re.config.MaxDelay

	var delay time.Duration

	switch re.config.BackoffType {
	case BackoffTypeLinear:
		delay = time.Duration(attempt) * baseDelay
	case BackoffTypeFixed:
		delay = baseDelay
	case BackoffTypeExponential:
		delay = time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
	default:
		delay = baseDelay
	}

	// Apply jitter if enabled
	if re.config.Jitter {
		delay = re.applyJitter(delay)
	}

	// Ensure delay doesn't exceed maximum
	if delay > maxDelay {
		delay = maxDelay
	}

	// Use provider-specific delay adjustment if available
	providerDelay := re.errorHandler.GetRetryDelay(err, attempt)
	if providerDelay > delay {
		delay = providerDelay
	}

	return delay
}

// applyJitter applies random jitter to delay
func (re *RetryExecutor) applyJitter(delay time.Duration) time.Duration {
	// Add ±25% random jitter
	jitterFactor := 0.75 + (rand.New(re.randSource).Float64() * 0.5)
	return time.Duration(float64(delay) * jitterFactor)
}

// GetRetryState returns the current retry state
func (re *RetryExecutor) GetRetryState(attempt int, lastError error) *RetryState {
	return &RetryState{
		Attempt:     attempt,
		MaxAttempts: re.config.MaxAttempts,
		LastError:  lastError,
		NextDelay:   re.calculateDelay(attempt, lastError),
		ShouldRetry: attempt < re.config.MaxAttempts && re.errorHandler.IsRetryableError(lastError),
	}
}

// CircuitBreaker implements circuit breaker pattern for retries
type CircuitBreaker struct {
	maxFailures   int
	resetTimeout  time.Duration
	failures      int
	lastFailure   time.Time
	state         CircuitState
	logger        interfaces.Logger
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitStateClosed CircuitState = iota
	CircuitStateOpen
	CircuitStateHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, logger interfaces.Logger) *CircuitBreaker {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitStateClosed,
		logger:       logger,
	}
}

// Execute executes an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation func() error) error {
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is %v, operation rejected", cb.state)
	}

	err := operation()
	cb.recordResult(err)
	return err
}

// Note: Generic methods not supported in current Go version
// Use Execute method and handle results in calling code

// canExecute checks if operation can be executed based on circuit state
func (cb *CircuitBreaker) canExecute() bool {
	switch cb.state {
	case CircuitStateClosed:
		return true
	case CircuitStateOpen:
		// Check if reset timeout has passed
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitStateHalfOpen
			cb.logger.Info("Circuit breaker transitioning to half-open")
			return true
		}
		return false
	case CircuitStateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(err error) {
	if err == nil {
		// Success - reset failure count and close circuit
		cb.failures = 0
		if cb.state != CircuitStateClosed {
			cb.state = CircuitStateClosed
			cb.logger.Info("Circuit breaker closed after successful operation")
		}
	} else {
		// Failure - increment failure count
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = CircuitStateOpen
			cb.logger.Warn("Circuit breaker opened after too many failures",
				"failures", cb.failures,
				"max_failures", cb.maxFailures)
		}
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}

// GetFailures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	return cb.failures
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.failures = 0
	cb.state = CircuitStateClosed
	cb.logger.Info("Circuit breaker manually reset")
}

// RetryWithCircuitBreaker combines retry logic with circuit breaker
type RetryWithCircuitBreaker struct {
	retryExecutor   *RetryExecutor
	circuitBreaker *CircuitBreaker
	logger         interfaces.Logger
}

// NewRetryWithCircuitBreaker creates a combined retry and circuit breaker
func NewRetryWithCircuitBreaker(retryConfig RetryConfig, circuitMaxFailures int, circuitResetTimeout time.Duration, logger interfaces.Logger) *RetryWithCircuitBreaker {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &RetryWithCircuitBreaker{
		retryExecutor:   NewRetryExecutor(retryConfig, logger),
		circuitBreaker: NewCircuitBreaker(circuitMaxFailures, circuitResetTimeout, logger),
		logger:         logger,
	}
}

// Execute executes an operation with both retry and circuit breaker logic
func (rcb *RetryWithCircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	return rcb.retryExecutor.Execute(ctx, func() error {
		return rcb.circuitBreaker.Execute(operation)
	})
}

// Note: Generic methods not supported in current Go version
// Use Execute method and handle results in calling code

// GetCircuitBreakerState returns the current circuit breaker state
func (rcb *RetryWithCircuitBreaker) GetCircuitBreakerState() CircuitState {
	return rcb.circuitBreaker.GetState()
}

// ResetCircuitBreaker resets the circuit breaker
func (rcb *RetryWithCircuitBreaker) ResetCircuitBreaker() {
	rcb.circuitBreaker.Reset()
}

// Helper functions

// IsRetryableErrorWithConfig checks if an error is retryable based on configuration
func IsRetryableErrorWithConfig(err error, config RetryConfig) bool {
	if modelErr, ok := err.(*ModelError); ok {
		for _, retryableErr := range config.RetryableErrors {
			if string(modelErr.Code) == retryableErr {
				return true
			}
		}
		return modelErr.Retryable
	}
	return false
}

// CalculateExponentialBackoff calculates exponential backoff delay
func CalculateExponentialBackoff(attempt int, baseDelay time.Duration, maxDelay time.Duration, jitter bool) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay

	if jitter {
		// Add ±25% random jitter
		jitterFactor := 0.75 + (rand.Float64() * 0.5)
		delay = time.Duration(float64(delay) * jitterFactor)
	}

	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// CalculateLinearBackoff calculates linear backoff delay
func CalculateLinearBackoff(attempt int, baseDelay time.Duration, maxDelay time.Duration, jitter bool) time.Duration {
	delay := time.Duration(attempt) * baseDelay

	if jitter {
		jitterFactor := 0.75 + (rand.Float64() * 0.5)
		delay = time.Duration(float64(delay) * jitterFactor)
	}

	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// GetRetryDelayForError returns appropriate retry delay for different error types
func GetRetryDelayForError(err error, attempt int) time.Duration {
	if modelErr, ok := err.(*ModelError); ok {
		switch modelErr.Code {
		case ErrorCodeRateLimitError:
			// Exponential backoff for rate limits
			return CalculateExponentialBackoff(attempt, 2*time.Second, 60*time.Second, true)
		case ErrorCodeServerError, ErrorCodeServiceUnavailable:
			// Longer backoff for server errors
			return CalculateExponentialBackoff(attempt, 5*time.Second, 120*time.Second, true)
		case ErrorCodeNetworkError, ErrorCodeTimeoutError, ErrorCodeConnectionError:
			// Moderate backoff for network issues
			return CalculateLinearBackoff(attempt, 1*time.Second, 30*time.Second, true)
		default:
			// Default backoff
			return time.Duration(attempt) * time.Second
		}
	}

	// Default delay for unknown errors
	return time.Second
}

// RetryConfigBuilder helps build retry configurations
type RetryConfigBuilder struct {
	config RetryConfig
}

// NewRetryConfigBuilder creates a new retry configuration builder
func NewRetryConfigBuilder() *RetryConfigBuilder {
	return &RetryConfigBuilder{
		config: DefaultRetryConfig,
	}
}

// WithMaxAttempts sets the maximum number of retry attempts
func (b *RetryConfigBuilder) WithMaxAttempts(maxAttempts int) *RetryConfigBuilder {
	b.config.MaxAttempts = maxAttempts
	return b
}

// WithBaseDelay sets the base delay between retries
func (b *RetryConfigBuilder) WithBaseDelay(baseDelay time.Duration) *RetryConfigBuilder {
	b.config.BaseDelay = baseDelay
	return b
}

// WithMaxDelay sets the maximum delay between retries
func (b *RetryConfigBuilder) WithMaxDelay(maxDelay time.Duration) *RetryConfigBuilder {
	b.config.MaxDelay = maxDelay
	return b
}

// WithBackoffType sets the backoff type
func (b *RetryConfigBuilder) WithBackoffType(backoffType BackoffType) *RetryConfigBuilder {
	b.config.BackoffType = backoffType
	return b
}

// WithJitter enables or disables jitter
func (b *RetryConfigBuilder) WithJitter(jitter bool) *RetryConfigBuilder {
	b.config.Jitter = jitter
	return b
}

// WithRetryableErrors sets the list of retryable error codes
func (b *RetryConfigBuilder) WithRetryableErrors(retryableErrors []string) *RetryConfigBuilder {
	b.config.RetryableErrors = retryableErrors
	return b
}

// Build creates the final retry configuration
func (b *RetryConfigBuilder) Build() RetryConfig {
	return b.config
}