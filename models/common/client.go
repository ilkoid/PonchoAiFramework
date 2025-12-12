package common

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// HTTPClientFactory creates and manages HTTP clients for model providers
type HTTPClientFactory struct {
	config  HTTPConfig
	logger  interfaces.Logger
	clients map[string]*http.Client
	mutex   sync.RWMutex
}

// NewHTTPClientFactory creates a new HTTP client factory
func NewHTTPClientFactory(config HTTPConfig, logger interfaces.Logger) *HTTPClientFactory {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	// Use default config if not provided
	if config.Timeout == 0 {
		config = DefaultHTTPConfig
	}

	return &HTTPClientFactory{
		config:  config,
		logger:  logger,
		clients: make(map[string]*http.Client),
	}
}

// GetClient returns an HTTP client for the specified provider
func (f *HTTPClientFactory) GetClient(provider Provider) *http.Client {
	f.mutex.RLock()
	client, exists := f.clients[string(provider)]
	f.mutex.RUnlock()

	if exists {
		return client
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Double-check after acquiring write lock
	if client, exists := f.clients[string(provider)]; exists {
		return client
	}

	// Create new client
	client = f.createClient()
	f.clients[string(provider)] = client

	f.logger.Debug("Created new HTTP client for provider", "provider", provider)
	return client
}

// createClient creates a new HTTP client with the factory configuration
func (f *HTTPClientFactory) createClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        f.config.MaxIdleConns,
		MaxIdleConnsPerHost: f.config.MaxIdleConnsPerHost,
		IdleConnTimeout:     f.config.IdleConnTimeout,
		TLSHandshakeTimeout:  f.config.TLSHandshakeTimeout,
		MaxConnsPerHost:      f.config.MaxConnsPerHost,
		ReadBufferSize:       f.config.ReadBufferSize,
		WriteBufferSize:      f.config.WriteBufferSize,
		DisableCompression:   f.config.DisableCompression,
		DisableKeepAlives:    f.config.DisableKeepAlives,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: f.config.InsecureSkipVerify,
		},
	}

	// Configure proxy if provided
	if f.config.ProxyURL != "" {
		if proxyURL, err := url.Parse(f.config.ProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		} else {
			f.logger.Warn("Invalid proxy URL", "url", f.config.ProxyURL, "error", err)
		}
	}

	// Configure dial context for connection timeouts
	transport.DialContext = (&net.Dialer{
		Timeout:   f.config.Timeout,
		KeepAlive: 30 * time.Second,
	}).DialContext

	return &http.Client{
		Timeout:   f.config.Timeout,
		Transport: transport,
	}
}

// UpdateConfig updates the factory configuration
func (f *HTTPClientFactory) UpdateConfig(config HTTPConfig) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.config = config

	// Close existing clients
	for provider, client := range f.clients {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
		delete(f.clients, provider)
	}

	f.logger.Info("Updated HTTP client factory configuration")
}

// Close closes all HTTP clients and cleans up resources
func (f *HTTPClientFactory) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for provider, client := range f.clients {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
		delete(f.clients, provider)
	}

	f.logger.Info("Closed all HTTP clients")
	return nil
}

// GetConfig returns the current configuration
func (f *HTTPClientFactory) GetConfig() HTTPConfig {
	return f.config
}

// HTTPClient represents a wrapper around http.Client with additional functionality
type HTTPClient struct {
	client    *http.Client
	config    HTTPConfig
	provider  Provider
	logger    interfaces.Logger
	requestID string
}

// NewHTTPClient creates a new HTTP client wrapper
func NewHTTPClient(client *http.Client, config HTTPConfig, provider Provider, logger interfaces.Logger) *HTTPClient {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	return &HTTPClient{
		client:   client,
		config:   config,
		provider: provider,
		logger:   logger,
	}
}

// Do executes an HTTP request with logging and metrics
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	// Add common headers
	c.addCommonHeaders(req)

	// Log request
	c.logger.Debug("Sending HTTP request",
		"method", req.Method,
		"url", req.URL.String(),
		"provider", c.provider,
		"request_id", c.requestID)

	// Execute request
	resp, err := c.client.Do(req)
	duration := time.Since(startTime)

	// Log response
	if err != nil {
		c.logger.Error("HTTP request failed",
			"error", err,
			"duration", duration,
			"provider", c.provider,
			"request_id", c.requestID)
	} else {
		c.logger.Debug("HTTP request completed",
			"status_code", resp.StatusCode,
			"duration", duration,
			"provider", c.provider,
			"request_id", c.requestID)
	}

	return resp, err
}

// DoWithContext executes an HTTP request with context
func (c *HTTPClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return c.Do(req)
}

// addCommonHeaders adds common headers to the request
func (c *HTTPClient) addCommonHeaders(req *http.Request) {
	if req.Header.Get(HeaderUserAgent) == "" {
		req.Header.Set(HeaderUserAgent, c.config.UserAgent)
	}

	if c.requestID != "" {
		req.Header.Set(HeaderXRequestID, c.requestID)
	}
}

// SetRequestID sets the request ID for tracking
func (c *HTTPClient) SetRequestID(requestID string) {
	c.requestID = requestID
}

// GetRequestID returns the current request ID
func (c *HTTPClient) GetRequestID() string {
	return c.requestID
}

// Close closes the HTTP client
func (c *HTTPClient) Close() error {
	if transport, ok := c.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// HTTPRequestBuilder helps build HTTP requests with common patterns
type HTTPRequestBuilder struct {
	method  string
	url     string
	headers map[string]string
	body    []byte
	query   map[string]string
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

// SetBody sets the request body
func (b *HTTPRequestBuilder) SetBody(body []byte) *HTTPRequestBuilder {
	b.body = body
	return b
}

// SetQuery sets a query parameter
func (b *HTTPRequestBuilder) SetQuery(key, value string) *HTTPRequestBuilder {
	b.query[key] = value
	return b
}

// SetQueryParams sets multiple query parameters
func (b *HTTPRequestBuilder) SetQueryParams(params map[string]string) *HTTPRequestBuilder {
	for k, v := range params {
		b.query[k] = v
	}
	return b
}

// Build builds the HTTP request
func (b *HTTPRequestBuilder) Build() (*http.Request, error) {
	// Add query parameters to URL
	if len(b.query) > 0 {
		parsedURL, err := url.Parse(b.url)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}

		query := parsedURL.Query()
		for k, v := range b.query {
			query.Set(k, v)
		}
		parsedURL.RawQuery = query.Encode()
		b.url = parsedURL.String()
	}

	// Create request
	var req *http.Request
	var err error

	if b.body != nil {
		req, err = http.NewRequest(b.method, b.url, bytes.NewReader(b.body))
	} else {
		req, err = http.NewRequest(b.method, b.url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range b.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// HTTPResponse represents a wrapped HTTP response with utility methods
type HTTPResponse struct {
	Response *http.Response
	Body     []byte
	Duration time.Duration
}

// NewHTTPResponse creates a new HTTP response wrapper
func NewHTTPResponse(resp *http.Response, body []byte, duration time.Duration) *HTTPResponse {
	return &HTTPResponse{
		Response: resp,
		Body:     body,
		Duration: duration,
	}
}

// IsSuccess returns true if the response status code indicates success
func (r *HTTPResponse) IsSuccess() bool {
	return r.Response.StatusCode >= 200 && r.Response.StatusCode < 300
}

// IsClientError returns true if the response status code indicates a client error
func (r *HTTPResponse) IsClientError() bool {
	return r.Response.StatusCode >= 400 && r.Response.StatusCode < 500
}

// IsServerError returns true if the response status code indicates a server error
func (r *HTTPResponse) IsServerError() bool {
	return r.Response.StatusCode >= 500
}

// GetRateLimitInfo extracts rate limit information from response headers
func (r *HTTPResponse) GetRateLimitInfo() *RateLimitInfo {
	info := &RateLimitInfo{}

	// Extract common rate limit headers
	if retryAfter := r.Response.Header.Get(HeaderRetryAfter); retryAfter != "" {
		if duration, err := time.ParseDuration(retryAfter + "s"); err == nil {
			info.RetryAfter = &duration
		}
	}

	// Note: Different providers use different headers for rate limiting
	// This can be extended based on specific provider requirements

	return info
}

// GetHeader returns the value of a response header
func (r *HTTPResponse) GetHeader(key string) string {
	return r.Response.Header.Get(key)
}

// GetHeaders returns all response headers
func (r *HTTPResponse) GetHeaders() map[string][]string {
	return r.Response.Header
}

// Helper functions

// CreateJSONRequest creates a JSON HTTP request
func CreateJSONRequest(method, url string, body interface{}, headers map[string]string) (*http.Request, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	builder := NewHTTPRequestBuilder(method, url).
		SetBody(bodyBytes).
		SetHeader(HeaderContentType, MIMETypeJSON).
		SetHeader(HeaderAccept, MIMETypeJSON)

	if headers != nil {
		builder.SetHeaders(headers)
	}

	return builder.Build()
}

// CreateStreamingRequest creates a streaming HTTP request
func CreateStreamingRequest(method, url string, body interface{}, headers map[string]string) (*http.Request, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	builder := NewHTTPRequestBuilder(method, url).
		SetBody(bodyBytes).
		SetHeader(HeaderContentType, MIMETypeJSON).
		SetHeader(HeaderAccept, MIMETypeEventStream).
		SetHeader(HeaderCacheControl, "no-cache")

	if headers != nil {
		builder.SetHeaders(headers)
	}

	return builder.Build()
}