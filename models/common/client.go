package common

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// HTTPClient represents a configured HTTP client with retry capabilities
type HTTPClient struct {
	client      *http.Client
	config      *HTTPConfig
	retryConfig *RetryConfig
}

// NewHTTPClient creates a new HTTP client with given configuration
func NewHTTPClient(config *HTTPConfig, retryConfig RetryConfig) (*HTTPClient, error) {
	if config == nil {
		config = &DefaultHTTPConfig
	}

	if retryConfig.MaxAttempts == 0 {
		retryConfig = RetryConfig{
			MaxAttempts:     3,
			BaseDelay:       1 * time.Second,
			MaxDelay:        30 * time.Second,
			BackoffType:     BackoffTypeExponential,
			Jitter:          true,
			RetryableErrors: []string{"timeout", "rate_limit", "server_error", "network_error"},
		}
	}

	// Create transport with custom configuration
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		TLSHandshakeTimeout: config.TLSHandshakeTimeout,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		DisableKeepAlives:   config.DisableKeepAlives,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
	}

	// Create HTTP client
	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &HTTPClient{
		client:      client,
		config:      config,
		retryConfig: &retryConfig,
	}, nil
}

// Do executes an HTTP request with retry logic
func (c *HTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 1; attempt <= c.retryConfig.MaxAttempts; attempt++ {
		// Clone request for each attempt to avoid body consumption issues
		reqClone := c.cloneRequest(req)

		// Set user agent if not already set
		if reqClone.Header.Get("User-Agent") == "" {
			reqClone.Header.Set("User-Agent", c.config.UserAgent)
		}

		// Execute request
		resp, err := c.client.Do(reqClone.WithContext(ctx))
		if err == nil {
			// Check for non-retryable status codes
			if !c.isRetryableStatus(resp.StatusCode) {
				return resp, nil
			}
			resp.Body.Close() // Close body before retry
		}

		lastErr = err

		// If this is the last attempt, return the error
		if attempt == c.retryConfig.MaxAttempts {
			if err != nil {
				return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryConfig.MaxAttempts, err)
			}
			return nil, fmt.Errorf("request failed after %d attempts with status code: %d", c.retryConfig.MaxAttempts, resp.StatusCode)
		}

		// Calculate delay for next attempt
		delay := c.calculateDelay(attempt)

		// Wait before retry (with context cancellation)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return nil, lastErr
}

// Get executes a GET request with retry logic
func (c *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	return c.Do(ctx, req)
}

// Post executes a POST request with retry logic
func (c *HTTPClient) Post(ctx context.Context, url string, contentType string, body interface{}) (*http.Response, error) {
	req, err := c.createRequestWithBody(ctx, "POST", url, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	return c.Do(ctx, req)
}

// Put executes a PUT request with retry logic
func (c *HTTPClient) Put(ctx context.Context, url string, contentType string, body interface{}) (*http.Response, error) {
	req, err := c.createRequestWithBody(ctx, "PUT", url, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create PUT request: %w", err)
	}

	return c.Do(ctx, req)
}

// Delete executes a DELETE request with retry logic
func (c *HTTPClient) Delete(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DELETE request: %w", err)
	}

	return c.Do(ctx, req)
}

// Close closes the HTTP client and cleans up resources
func (c *HTTPClient) Close() error {
	// Close idle connections
	c.client.CloseIdleConnections()
	return nil
}

// GetConfig returns client configuration
func (c *HTTPClient) GetConfig() *HTTPConfig {
	return c.config
}

// GetRetryConfig returns retry configuration
func (c *HTTPClient) GetRetryConfig() *RetryConfig {
	return c.retryConfig
}

// cloneRequest creates a clone of the HTTP request
func (c *HTTPClient) cloneRequest(req *http.Request) *http.Request {
	// Create new request
	reqClone := new(http.Request)
	*reqClone = *req

	// Clone header
	reqClone.Header = make(http.Header, len(req.Header))
	for k, v := range req.Header {
		reqClone.Header[k] = append([]string(nil), v...)
	}

	return reqClone
}

// isRetryableStatus checks if the HTTP status code is retryable
func (c *HTTPClient) isRetryableStatus(statusCode int) bool {
	// Don't retry for client errors (4xx) except for specific cases
	if statusCode >= 400 && statusCode < 500 {
		switch statusCode {
		case 408, // Request Timeout
			429, // Too Many Requests
			449, // Retry With (proprietary)
			509: // Bandwidth Limit Exceeded
			return true
		default:
			return false
		}
	}

	// Retry for server errors (5xx)
	return statusCode >= 500
}

// calculateDelay calculates the delay before the next retry attempt
func (c *HTTPClient) calculateDelay(attempt int) time.Duration {
	var delay time.Duration

	switch c.retryConfig.BackoffType {
	case BackoffTypeLinear:
		delay = time.Duration(attempt) * c.retryConfig.BaseDelay
	case BackoffTypeExponential:
		delay = c.retryConfig.BaseDelay * time.Duration(1<<uint(attempt-1))
	default:
		delay = c.retryConfig.BaseDelay
	}

	// Apply jitter (±25% random variation)
	jitter := time.Duration(float64(delay) * 0.25 * (2.0*float64(time.Now().UnixNano()%1000)/1000.0 - 1.0))
	delay += jitter

	// Ensure delay doesn't exceed max delay
	if delay > c.retryConfig.MaxDelay {
		delay = c.retryConfig.MaxDelay
	}

	return delay
}

// createRequestWithBody creates an HTTP request with a body
func (c *HTTPClient) createRequestWithBody(ctx context.Context, method, url, contentType string, body interface{}) (*http.Request, error) {
	var req *http.Request
	var err error

	if body == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		// Handle different body types
		switch body := body.(type) {
		case string:
			req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
		case []byte:
			req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		default:
			// Try to marshal as JSON for complex types
			if contentType == MIMETypeJSON || contentType == "" {
				jsonBody, err := json.Marshal(body)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal body as JSON: %w", err)
				}
				req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(jsonBody))
			} else {
				return nil, fmt.Errorf("unsupported body type: %T for content type: %s", body, contentType)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

// HTTPClientFactory creates HTTP clients for different providers
type HTTPClientFactory struct {
	config *HTTPConfig
	logger interfaces.Logger
}

// NewHTTPClientFactory creates a new HTTP client factory
func NewHTTPClientFactory(config *HTTPConfig, logger interfaces.Logger) *HTTPClientFactory {
	if config == nil {
		config = &DefaultHTTPConfig
	}

	return &HTTPClientFactory{
		config: config,
		logger: logger,
	}
}

// GetClient returns an HTTP client for the specified provider
func (f *HTTPClientFactory) GetClient(provider Provider) *HTTPClient {
	retryConfig := DefaultRetryConfig

	client, err := NewHTTPClient(f.config, retryConfig)
	if err != nil {
		// In a real implementation, you might want to handle this error
		// For now, we'll return nil
		return nil
	}

	return client
}

// HTTPRequestBuilder helps build HTTP requests
type HTTPRequestBuilder struct {
	method  string
	url     string
	headers map[string]string
	query   map[string]string
	body    []byte
}

// NewHTTPRequestBuilder creates a new HTTP request builder
func NewHTTPRequestBuilder(method, url string) *HTTPRequestBuilder {
	return &HTTPRequestBuilder{
		method:  method,
		url:     url,
		headers: make(map[string]string),
		query:   make(map[string]string),
	}
}

// SetHeader sets a header
func (b *HTTPRequestBuilder) SetHeader(key, value string) *HTTPRequestBuilder {
	b.headers[key] = value
	return b
}

// SetHeaders sets multiple headers
func (b *HTTPRequestBuilder) SetHeaders(headers map[string]string) *HTTPRequestBuilder {
	for k, v := range headers {
		b.headers[k] = v
	}
	return b
}

// SetQuery sets a query parameter
func (b *HTTPRequestBuilder) SetQuery(key, value string) *HTTPRequestBuilder {
	b.query[key] = value
	return b
}

// SetBody sets the request body
func (b *HTTPRequestBuilder) SetBody(body []byte) *HTTPRequestBuilder {
	b.body = body
	return b
}

// Build builds the HTTP request
func (b *HTTPRequestBuilder) Build() (*http.Request, error) {
	req, err := http.NewRequest(b.method, b.url, bytes.NewReader(b.body))
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range b.headers {
		req.Header.Set(k, v)
	}

	// Set query parameters
	if len(b.query) > 0 {
		q := req.URL.Query()
		for k, v := range b.query {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

// HTTPResponse wraps http.Response with helper methods
type HTTPResponse struct {
	*http.Response
}

// IsSuccess returns true for successful status codes (2xx)
func (r *HTTPResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError returns true for client error status codes (4xx)
func (r *HTTPResponse) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError returns true for server error status codes (5xx)
func (r *HTTPResponse) IsServerError() bool {
	return r.StatusCode >= 500
}

// CreateJSONRequest creates a JSON HTTP request
func CreateJSONRequest(method, url string, body interface{}, headers map[string]string) (*http.Request, error) {
	builder := NewHTTPRequestBuilder(method, url)

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		builder.SetBody(bodyBytes)
	}

	if headers == nil {
		headers = make(map[string]string)
	}

	// Set default content type for JSON
	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/json"
	}

	builder.SetHeaders(headers)

	return builder.Build()
}
