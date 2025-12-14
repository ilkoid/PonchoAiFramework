package wildberries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// WBClient implements a client for Wildberries Content API
type WBClient struct {
	client     *http.Client
	apiKey     string
	baseURL    string
	rateLimiter *RateLimiter
	logger     interfaces.Logger
}

// RateLimiter implements token bucket algorithm for API rate limiting
type RateLimiter struct {
	tokens       int
	maxTokens    int
	refillRate   time.Duration
	lastRefill   time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// WaitForToken waits for a token to be available
func (rl *RateLimiter) WaitForToken(ctx context.Context) error {
	for {
		rl.refill()

		if rl.tokens > 0 {
			rl.tokens--
			return nil
		}

		// Calculate wait time until next token
		waitTime := rl.refillRate - time.Since(rl.lastRefill)
		if waitTime <= 0 {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			continue
		}
	}
}

// refill adds tokens based on elapsed time
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	if elapsed >= rl.refillRate {
		// Add one token per refill interval
		tokensToAdd := int(elapsed / rl.refillRate)
		rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
		rl.lastRefill = now
	}
}

// NewWBClient creates a new Wildberries API client
func NewWBClient(apiKey string, baseURL string, logger interfaces.Logger) *WBClient {
	// Default to production URL if not specified
	if baseURL == "" {
		baseURL = "https://content-api.wildberries.ru"
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &WBClient{
		client:      client,
		apiKey:      apiKey,
		baseURL:     baseURL,
		rateLimiter: NewRateLimiter(5, 600*time.Millisecond), // 5 burst, 600ms refill
		logger:      logger,
	}
}

// Ping checks API connectivity and token validity
func (c *WBClient) Ping(ctx context.Context) (*PingResponse, error) {
	var result PingResponse
	err := c.makeRequest(ctx, "GET", "/ping", nil, &result)
	return &result, err
}

// GetParentCategories retrieves all parent product categories
func (c *WBClient) GetParentCategories(ctx context.Context) ([]ParentCategory, error) {
	var response struct {
		Data []ParentCategory `json:"data"`
		WBResponse
	}

	err := c.makeRequest(ctx, "GET", "/content/v2/object/parent/all", nil, &response)
	return response.Data, err
}

// GetSubjects retrieves all subjects with optional filtering
func (c *WBClient) GetSubjects(ctx context.Context, opts *GetSubjectsOptions) ([]Subject, error) {
	if opts == nil {
		opts = &GetSubjectsOptions{}
	}

	params := url.Values{}
	if opts.Locale != "" {
		params.Set("locale", opts.Locale)
	}
	if opts.Name != "" {
		params.Set("name", opts.Name)
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.ParentID > 0 {
		params.Set("parentID", strconv.Itoa(opts.ParentID))
	}

	var response struct {
		Data []Subject `json:"data"`
		WBResponse
	}

	endpoint := "/content/v2/object/all"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	err := c.makeRequest(ctx, "GET", endpoint, nil, &response)
	return response.Data, err
}

// GetSubjectCharacteristics retrieves characteristics for a specific subject
func (c *WBClient) GetSubjectCharacteristics(ctx context.Context, subjectID int) ([]SubjectCharacteristic, error) {
	endpoint := fmt.Sprintf("/content/v2/object/charcs/%d", subjectID)

	var response struct {
		Data []SubjectCharacteristic `json:"data"`
		WBResponse
	}

	err := c.makeRequest(ctx, "GET", endpoint, nil, &response)
	return response.Data, err
}

// GetBrands retrieves brands by subject ID
func (c *WBClient) GetBrands(ctx context.Context, subjectID int) ([]Brand, error) {
	endpoint := fmt.Sprintf("/api/content/v1/brands?subjectID=%d", subjectID)

	var response struct {
		Data []Brand `json:"data"`
		WBResponse
	}

	err := c.makeRequest(ctx, "GET", endpoint, nil, &response)
	return response.Data, err
}

// GetColors retrieves color characteristic values
func (c *WBClient) GetColors(ctx context.Context) ([]Color, error) {
	var response struct {
		Data []Color `json:"data"`
		WBResponse
	}

	err := c.makeRequest(ctx, "GET", "/content/v2/directory/colors", nil, &response)
	return response.Data, err
}

// GetGenders retrieves gender characteristic values
func (c *WBClient) GetGenders(ctx context.Context) ([]Gender, error) {
	var response struct {
		Data []Gender `json:"data"`
		WBResponse
	}

	err := c.makeRequest(ctx, "GET", "/content/v2/directory/kinds", nil, &response)
	return response.Data, err
}

// GetSeasons retrieves season characteristic values
func (c *WBClient) GetSeasons(ctx context.Context) ([]Season, error) {
	var response struct {
		Data []Season `json:"data"`
		WBResponse
	}

	err := c.makeRequest(ctx, "GET", "/content/v2/directory/seasons", nil, &response)
	return response.Data, err
}

// GetVATRates retrieves VAT rate values
func (c *WBClient) GetVATRates(ctx context.Context) ([]VATRate, error) {
	var response struct {
		Data []VATRate `json:"data"`
		WBResponse
	}

	err := c.makeRequest(ctx, "GET", "/content/v2/directory/vat", nil, &response)
	return response.Data, err
}

// makeRequest makes an HTTP request to the Wildberries API
func (c *WBClient) makeRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	// Wait for rate limiter
	if err := c.rateLimiter.WaitForToken(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}

	// Build request URL
	fullURL := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonBody))
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PonchoAiFramework/1.0")

	// Make request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.handleHTTPError(resp.StatusCode, respBody)
	}

	// Parse response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// handleHTTPError converts HTTP errors to structured errors
func (c *WBClient) handleHTTPError(statusCode int, body []byte) error {
	var errorResp WBErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		// If we can't parse error response, create generic error
		return &WBError{
			Code:    statusCode,
			Message: fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
		}
	}

	return &WBError{
		Code:    statusCode,
		Message: errorResp.Title,
		Details: errorResp.Detail,
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}