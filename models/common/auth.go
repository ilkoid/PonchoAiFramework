package common

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// AuthType represents different authentication types
type AuthType string

const (
	AuthTypeBearer    AuthType = "bearer"
	AuthTypeAPIKey    AuthType = "api_key"
	AuthTypeBasic     AuthType = "basic"
	AuthTypeCustom    AuthType = "custom"
	AuthTypeNone      AuthType = "none"
)

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type        AuthType `json:"type"`
	APIKey      string   `json:"api_key,omitempty"`
	SecretKey   string   `json:"secret_key,omitempty"`
	Username    string   `json:"username,omitempty"`
	Password    string   `json:"password,omitempty"`
	HeaderName  string   `json:"header_name,omitempty"`
	QueryParam  string   `json:"query_param,omitempty"`
	CustomFunc  func(*http.Request) error `json:"-"`
}

// TokenInfo represents information about an authentication token
type TokenInfo struct {
	Token      string    `json:"token"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	TokenType  string    `json:"token_type"`
	Scopes     []string  `json:"scopes,omitempty"`
}

// AuthManager manages authentication for different providers
type AuthManager struct {
	config    AuthConfig
	provider  Provider
	logger    interfaces.Logger
	token     *TokenInfo
	mutex     sync.RWMutex
	tokenFunc  func(ctx context.Context, config *AuthConfig) (*TokenInfo, error)
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(provider Provider, config AuthConfig, logger interfaces.Logger) *AuthManager {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}

	// Set default header name for API key auth
	if config.Type == AuthTypeAPIKey && config.HeaderName == "" {
		config.HeaderName = "Authorization"
	}

	return &AuthManager{
		config:   config,
		provider: provider,
		logger:   logger,
	}
}

// SetTokenRefreshFunction sets a function to refresh tokens
func (am *AuthManager) SetTokenRefreshFunction(tokenFunc func(ctx context.Context, config *AuthConfig) (*TokenInfo, error)) {
	am.tokenFunc = tokenFunc
}

// Authenticate applies authentication to an HTTP request
func (am *AuthManager) Authenticate(req *http.Request) error {
	switch am.config.Type {
	case AuthTypeBearer:
		return am.applyBearerAuth(req)
	case AuthTypeAPIKey:
		return am.applyAPIKeyAuth(req)
	case AuthTypeBasic:
		return am.applyBasicAuth(req)
	case AuthTypeCustom:
		return am.applyCustomAuth(req)
	case AuthTypeNone:
		return nil
	default:
		return fmt.Errorf("unsupported authentication type: %s", am.config.Type)
	}
}

// applyBearerAuth applies Bearer token authentication
func (am *AuthManager) applyBearerAuth(req *http.Request) error {
	if am.config.APIKey == "" {
		return fmt.Errorf("API key is required for Bearer authentication")
	}

	headerValue := fmt.Sprintf("Bearer %s", am.config.APIKey)
	req.Header.Set(HeaderAuthorization, headerValue)

	am.logger.Debug("Applied Bearer authentication", "provider", am.provider)
	return nil
}

// applyAPIKeyAuth applies API key authentication
func (am *AuthManager) applyAPIKeyAuth(req *http.Request) error {
	if am.config.APIKey == "" {
		return fmt.Errorf("API key is required for API key authentication")
	}

	if am.config.HeaderName != "" {
		// Use custom header name
		if strings.ToLower(am.config.HeaderName) == strings.ToLower(HeaderAuthorization) {
			// Standard Authorization header
			req.Header.Set(am.config.HeaderName, am.config.APIKey)
		} else {
			// Custom header (e.g., "X-API-Key")
			req.Header.Set(am.config.HeaderName, am.config.APIKey)
		}
	} else if am.config.QueryParam != "" {
		// Add as query parameter
		if req.URL.Query() == nil {
			req.URL.RawQuery = fmt.Sprintf("%s=%s", am.config.QueryParam, am.config.APIKey)
		} else {
			query := req.URL.Query()
			query.Set(am.config.QueryParam, am.config.APIKey)
			req.URL.RawQuery = query.Encode()
		}
	} else {
		// Default to Authorization header
		req.Header.Set(HeaderAuthorization, am.config.APIKey)
	}

	am.logger.Debug("Applied API key authentication", "provider", am.provider, "header", am.config.HeaderName)
	return nil
}

// applyBasicAuth applies basic authentication
func (am *AuthManager) applyBasicAuth(req *http.Request) error {
	if am.config.Username == "" || am.config.Password == "" {
		return fmt.Errorf("username and password are required for Basic authentication")
	}

	req.SetBasicAuth(am.config.Username, am.config.Password)

	am.logger.Debug("Applied Basic authentication", "provider", am.provider)
	return nil
}

// applyCustomAuth applies custom authentication
func (am *AuthManager) applyCustomAuth(req *http.Request) error {
	if am.config.CustomFunc != nil {
		return am.config.CustomFunc(req)
	}
	return fmt.Errorf("custom authentication function not configured")
}

// GetToken returns the current authentication token
func (am *AuthManager) GetToken() *TokenInfo {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.token
}

// SetToken sets the authentication token
func (am *AuthManager) SetToken(token *TokenInfo) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.token = token

	am.logger.Debug("Updated authentication token", "provider", am.provider, "expires_at", token.ExpiresAt)
}

// RefreshToken refreshes the authentication token
func (am *AuthManager) RefreshToken(ctx context.Context) error {
	if am.tokenFunc == nil {
		return fmt.Errorf("token refresh function not configured")
	}

	am.logger.Debug("Refreshing authentication token", "provider", am.provider)

	token, err := am.tokenFunc(ctx, &am.config)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	am.SetToken(token)
	return nil
}

// IsTokenExpired checks if the current token is expired
func (am *AuthManager) IsTokenExpired() bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	if am.token == nil || am.token.ExpiresAt == nil {
		return false // No expiration info
	}

	// Add 5-minute buffer before expiration
	return time.Now().Add(5 * time.Minute).After(*am.token.ExpiresAt)
}

// ValidateConfig validates the authentication configuration
func (am *AuthManager) ValidateConfig() error {
	switch am.config.Type {
	case AuthTypeBearer, AuthTypeAPIKey:
		if am.config.APIKey == "" {
			return fmt.Errorf("API key is required for %s authentication", am.config.Type)
		}
	case AuthTypeBasic:
		if am.config.Username == "" || am.config.Password == "" {
			return fmt.Errorf("username and password are required for Basic authentication")
		}
	case AuthTypeCustom:
		if am.config.CustomFunc == nil {
			return fmt.Errorf("custom authentication function is required for Custom authentication")
		}
	case AuthTypeNone:
		// No validation needed
	default:
		return fmt.Errorf("unsupported authentication type: %s", am.config.Type)
	}
	return nil
}

// GetAuthHeaders returns authentication headers
func (am *AuthManager) GetAuthHeaders() map[string]string {
	headers := make(map[string]string)

	switch am.config.Type {
	case AuthTypeBearer:
		if am.config.APIKey != "" {
			headers[HeaderAuthorization] = fmt.Sprintf("Bearer %s", am.config.APIKey)
		}
	case AuthTypeAPIKey:
		if am.config.APIKey != "" {
			if am.config.HeaderName != "" {
				headers[am.config.HeaderName] = am.config.APIKey
			} else {
				headers[HeaderAuthorization] = am.config.APIKey
			}
		}
	}

	return headers
}

// Provider-specific authentication configurations

// GetDeepSeekAuthConfig returns default authentication configuration for DeepSeek
func GetDeepSeekAuthConfig(apiKey string) AuthConfig {
	return AuthConfig{
		Type:   AuthTypeBearer,
		APIKey: apiKey,
	}
}

// GetZAIAuthConfig returns default authentication configuration for Z.AI
func GetZAIAuthConfig(apiKey string) AuthConfig {
	return AuthConfig{
		Type:   AuthTypeBearer,
		APIKey: apiKey,
	}
}

// GetOpenAIAuthConfig returns default authentication configuration for OpenAI
func GetOpenAIAuthConfig(apiKey string) AuthConfig {
	return AuthConfig{
		Type:   AuthTypeBearer,
		APIKey: apiKey,
	}
}

// GetCustomAuthConfig returns custom authentication configuration
func GetCustomAuthConfig(headerName, apiKey string) AuthConfig {
	return AuthConfig{
		Type:       AuthTypeAPIKey,
		APIKey:     apiKey,
		HeaderName: headerName,
	}
}

// GetBasicAuthConfig returns basic authentication configuration
func GetBasicAuthConfig(username, password string) AuthConfig {
	return AuthConfig{
		Type:     AuthTypeBasic,
		Username: username,
		Password: password,
	}
}

// AuthMiddleware represents middleware for authentication
type AuthMiddleware struct {
	authManager *AuthManager
	logger      interfaces.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authManager *AuthManager, logger interfaces.Logger) *AuthMiddleware {
	if logger == nil {
		logger = interfaces.NewDefaultLogger()
	}
	return &AuthMiddleware{
		authManager: authManager,
		logger:      logger,
	}
}

// Wrap wraps an HTTP client with authentication middleware
func (m *AuthMiddleware) Wrap(client *http.Client) *http.Client {
	originalTransport := client.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	authTransport := &authTransport{
		originalTransport: originalTransport,
		authManager:      m.authManager,
		logger:           m.logger,
	}

	client.Transport = authTransport
	return client
}

// authTransport is a custom transport that adds authentication
type authTransport struct {
	originalTransport http.RoundTripper
	authManager      *AuthManager
	logger           interfaces.Logger
}

// RoundTrip executes a single HTTP transaction with authentication
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	authReq := req.Clone(req.Context())

	// Apply authentication
	if err := t.authManager.Authenticate(authReq); err != nil {
		t.logger.Error("Failed to apply authentication", "error", err)
		return nil, err
	}

	// Execute the request
	return t.originalTransport.RoundTrip(authReq)
}

// Helper functions

// MaskAPIKey masks an API key for logging
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}

// ValidateAPIKey validates an API key format
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	if len(apiKey) < 10 {
		return fmt.Errorf("API key appears to be too short")
	}
	return nil
}

// GetAuthTypeFromString converts string to AuthType
func GetAuthTypeFromString(authType string) AuthType {
	switch strings.ToLower(authType) {
	case "bearer":
		return AuthTypeBearer
	case "api_key", "apikey":
		return AuthTypeAPIKey
	case "basic":
		return AuthTypeBasic
	case "custom":
		return AuthTypeCustom
	case "none", "":
		return AuthTypeNone
	default:
		return AuthTypeNone
	}
}