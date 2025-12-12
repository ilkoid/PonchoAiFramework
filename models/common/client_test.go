package common

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

func TestHTTPClientFactory_GetClient(t *testing.T) {
	logger := interfaces.NewDefaultLogger()
	factory := NewHTTPClientFactory(DefaultHTTPConfig, logger)

	tests := []struct {
		name     string
		provider Provider
		wantErr  bool
	}{
		{
			name:     "DeepSeek",
			provider: ProviderDeepSeek,
			wantErr:  false,
		},
		{
			name:     "ZAI",
			provider: ProviderZAI,
			wantErr:  false,
		},
		{
			name:     "OpenAI",
			provider: ProviderOpenAI,
			wantErr:  false,
		},
		{
			name:     "Custom",
			provider: ProviderCustom,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := factory.GetClient(tt.provider)
			
			if tt.wantErr && client == nil {
				t.Error("GetClient() returned nil client when error expected")
			}
			
			if !tt.wantErr && client == nil {
				t.Error("GetClient() returned nil client")
			}
		})
	}
}

func TestHTTPRequestBuilder_SetHeader(t *testing.T) {
	builder := NewHTTPRequestBuilder("GET", "https://api.example.com/test")

	tests := []struct {
		name   string
		key    string
		value  string
	}{
		{"Content-Type", "Content-Type", "application/json"},
		{"Authorization", "Authorization", "Bearer token123"},
		{"User-Agent", "User-Agent", "TestAgent/1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder.SetHeader(tt.key, tt.value)
			
			req, err := builder.Build()
			if err != nil {
				t.Fatalf("Build() failed: %v", err)
			}
			
			actualValue := req.Header.Get(tt.key)
			if actualValue != tt.value {
				t.Errorf("SetHeader() = %s, want %s", actualValue, tt.value)
			}
		})
	}
}

func TestHTTPRequestBuilder_SetBody(t *testing.T) {
	builder := NewHTTPRequestBuilder("POST", "https://api.example.com/test")

	body := map[string]string{"key": "value"}
	bodyBytes, _ := json.Marshal(body)

	builder.SetBody(bodyBytes)

	req, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if req == nil {
		t.Error("Build() returned nil request")
	}
}

func TestHTTPRequestBuilder_SetQuery(t *testing.T) {
	builder := NewHTTPRequestBuilder("GET", "https://api.example.com/test")

	builder.SetQuery("param1", "value1")
	builder.SetQuery("param2", "value2")

	req, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	expectedQuery := "param1=value1&param2=value2"
	if req.URL.RawQuery != expectedQuery {
		t.Errorf("Query = %s, want %s", req.URL.RawQuery, expectedQuery)
	}
}

func TestHTTPRequestBuilder_Build(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		url         string
		body        interface{}
		headers     map[string]string
		wantErr     bool
		wantHeaders map[string]string
	}{
		{
			name:    "Simple GET",
			method:  "GET",
			url:     "https://api.example.com/test",
			body:    nil,
			wantErr: false,
		},
		{
			name:   "POST with body",
			method: "POST",
			url:    "https://api.example.com/test",
			body:   map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "With headers",
			method:  "GET",
			url:     "https://api.example.com/test",
			body:    nil,
			headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
			wantErr: false,
			wantHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
		},
		{
			name:    "Invalid URL",
			method:  "GET",
			url:     "invalid-url",
			body:    nil,
			wantErr: false, // Build() doesn't return error for invalid URL, only during URL parsing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewHTTPRequestBuilder(tt.method, tt.url)
			
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				builder.SetBody(bodyBytes)
			}
			if tt.headers != nil {
				builder.SetHeaders(tt.headers)
			}
			
			req, err := builder.Build()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if req == nil {
					t.Error("Build() returned nil request")
					return
				}
				
				if req.Method != tt.method {
					t.Errorf("Method = %s, want %s", req.Method, tt.method)
				}
				
				if req.URL.String() != tt.url {
					t.Errorf("URL = %s, want %s", req.URL.String(), tt.url)
				}
				
				// Check headers
				for key, expectedValue := range tt.wantHeaders {
					actualValue := req.Header.Get(key)
					if actualValue != expectedValue {
						t.Errorf("Header %s = %s, want %s", key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestHTTPResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, true},
		{"201 Created", 201, true},
		{"204 No Content", 204, true},
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"404 Not Found", 404, false},
		{"500 Internal Server Error", 500, false},
		{"502 Bad Gateway", 502, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode}
			response := &HTTPResponse{Response: resp}
			
			result := response.IsSuccess()
			if result != tt.want {
				t.Errorf("IsSuccess() = %t, want %t", result, tt.want)
			}
		})
	}
}

func TestHTTPResponse_IsClientError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, false},
		{"400 Bad Request", 400, true},
		{"401 Unauthorized", 401, true},
		{"403 Forbidden", 403, true},
		{"404 Not Found", 404, true},
		{"500 Internal Server Error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode}
			response := &HTTPResponse{Response: resp}
			
			result := response.IsClientError()
			if result != tt.want {
				t.Errorf("IsClientError() = %t, want %t", result, tt.want)
			}
		})
	}
}

func TestHTTPResponse_IsServerError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, false},
		{"400 Bad Request", 400, false},
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode}
			response := &HTTPResponse{Response: resp}
			
			result := response.IsServerError()
			if result != tt.want {
				t.Errorf("IsServerError() = %t, want %t", result, tt.want)
			}
		})
	}
}

func TestCreateJSONRequest(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		url     string
		body    interface{}
		headers map[string]string
		wantErr bool
	}{
		{
			name:    "Simple GET",
			method:  "GET",
			url:     "https://api.example.com/test",
			body:    nil,
			wantErr: false,
		},
		{
			name:   "POST with body",
			method: "POST",
			url:    "https://api.example.com/test",
			body:   map[string]string{"key": "value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := CreateJSONRequest(tt.method, tt.url, tt.body, tt.headers)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateJSONRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if req == nil {
					t.Error("CreateJSONRequest() returned nil request")
					return
				}
				
				if req.Method != tt.method {
					t.Errorf("Method = %s, want %s", req.Method, tt.method)
				}
				
				if req.URL.String() != tt.url {
					t.Errorf("URL = %s, want %s", req.URL.String(), tt.url)
				}
				
				// Check Content-Type header
				contentType := req.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Content-Type = %s, want application/json", contentType)
				}
			}
		})
	}
}

func TestDefaultHTTPConfig(t *testing.T) {
	config := DefaultHTTPConfig

	if config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", config.Timeout, 30*time.Second)
	}
	if config.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %d, want %d", config.MaxIdleConns, 100)
	}
	if config.MaxIdleConnsPerHost != 10 {
		t.Errorf("MaxIdleConnsPerHost = %d, want %d", config.MaxIdleConnsPerHost, 10)
	}
	if config.IdleConnTimeout != 90*time.Second {
		t.Errorf("IdleConnTimeout = %v, want %v", config.IdleConnTimeout, 90*time.Second)
	}
	if config.TLSHandshakeTimeout != 10*time.Second {
		t.Errorf("TLSHandshakeTimeout = %v, want %v", config.TLSHandshakeTimeout, 10*time.Second)
	}
	if config.UserAgent != "PonchoFramework/1.0" {
		t.Errorf("UserAgent = %s, want %s", config.UserAgent, "PonchoFramework/1.0")
	}
}

// Benchmark tests
func BenchmarkHTTPClientFactory_GetClient(b *testing.B) {
	logger := interfaces.NewDefaultLogger()
	factory := NewHTTPClientFactory(DefaultHTTPConfig, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = factory.GetClient(ProviderCustom)
	}
}

func BenchmarkHTTPRequestBuilder_Build(b *testing.B) {
	builder := NewHTTPRequestBuilder("GET", "https://api.example.com/test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.Build()
	}
}