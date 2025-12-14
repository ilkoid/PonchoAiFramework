package s3

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// AWS V4 Signature implementation for Yandex Cloud S3 compatibility
// Based on AWS Signature Version 4 algorithm

const (
	algorithm     = "AWS4-HMAC-SHA256"
	service       = "s3"
	requestType   = "aws4_request"
)

// getSigningKey returns the derived signing key for AWS Signature V4
func getSigningKey(secretKey, region, dateStamp string) []byte {
	// HMAC-SHA256(AWS4 + SECRET_KEY, dateStamp)
	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)

	// HMAC-SHA256(kDate, region)
	kRegion := hmacSHA256(kDate, region)

	// HMAC-SHA256(kRegion, service)
	kService := hmacSHA256(kRegion, service)

	// HMAC-SHA256(kService, requestType)
	kSigning := hmacSHA256(kService, requestType)

	return kSigning
}

// hmacSHA256 returns HMAC-SHA256 of the given key and message
func hmacSHA256(key []byte, message string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return h.Sum(nil)
}

// createCanonicalRequest creates the canonical request for AWS Signature V4
func createCanonicalRequest(method, path, query, headers, payload string) string {
	var canonical strings.Builder

	// Method
	canonical.WriteString(method + "\n")

	// Canonical URI (encoded path)
	canonical.WriteString(path + "\n")

	// Canonical Query String
	canonical.WriteString(query + "\n")

	// Canonical Headers
	canonical.WriteString(headers + "\n")

	// Signed Headers (from createCanonicalHeaders)
	headerLines := strings.Split(headers, "\n")
	signedHeaders := ""
	for _, line := range headerLines {
		if line != "" {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				if signedHeaders != "" {
					signedHeaders += ";"
				}
				signedHeaders += strings.ToLower(strings.TrimSpace(parts[0]))
			}
		}
	}
	canonical.WriteString(signedHeaders + "\n")

	// Payload hash
	canonical.WriteString(payload)

	return canonical.String()
}

// createCanonicalHeaders creates the canonical headers string
func createCanonicalHeaders(headers map[string]string) string {
	var canonical strings.Builder

	// Sort headers by lowercase name
	var sortedHeaders []string
	for name := range headers {
		sortedHeaders = append(sortedHeaders, strings.ToLower(name))
	}
	sort.Strings(sortedHeaders)

	for _, name := range sortedHeaders {
		value := headers[name]
		canonical.WriteString(fmt.Sprintf("%s:%s\n", name, value))
	}

	return canonical.String()
}

// createStringToSign creates the string to sign for AWS Signature V4
func createStringToSign(algorithm, dateStamp, scope, canonicalRequestHash string) string {
	var stringToSign strings.Builder

	stringToSign.WriteString(algorithm + "\n")
	stringToSign.WriteString(dateStamp + "\n")
	stringToSign.WriteString(scope + "\n")
	stringToSign.WriteString(canonicalRequestHash)

	return stringToSign.String()
}

// calculateSignature calculates the signature for AWS Signature V4
func calculateSignature(signingKey []byte, stringToSign string) string {
	signature := hmacSHA256(signingKey, stringToSign)
	return hex.EncodeToString(signature)
}

// setAuthHeadersV4 sets AWS Signature Version 4 headers
func (c *S3Client) setAuthHeadersV4(req *http.Request) error {
	// Get current time
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	timeStamp := now.Format("20060102T150405Z")

	// Set required headers
	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("X-Amz-Date", timeStamp)

	// For GET requests, don't set Content-Type unless explicitly needed
	if req.Method == "GET" && req.Header.Get("Content-Type") == "application/octet-stream" {
		req.Header.Del("Content-Type")
	} else if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	// Get payload hash (empty string for GET requests)
	payloadHash := sha256Hash("")

	// Create canonical query string
	query := req.URL.Query().Encode()
	if query == "" {
		query = ""
	}

	// Create canonical URI - ensure it starts with /
	path := req.URL.Path
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Create canonical headers - only include headers that are actually needed
	headers := make(map[string]string)
	for name, values := range req.Header {
		if len(values) > 0 {
			// Normalize header names to lowercase and trim values
			lowerName := strings.ToLower(name)
			value := strings.TrimSpace(values[0])

			// Only include headers that are part of AWS signature
			if lowerName == "host" || lowerName == "x-amz-date" ||
			   lowerName == "content-type" || lowerName == "x-amz-content-sha256" {
				headers[lowerName] = value
			}
		}
	}

	// Ensure host is included
	if headers["host"] == "" {
		headers["host"] = req.URL.Host
	}

	canonicalHeaders := createCanonicalHeaders(headers)

	// Create canonical request
	canonicalRequest := createCanonicalRequest(
		req.Method,
		path,
		query,
		canonicalHeaders,
		payloadHash,
	)

	// Create scope
	scope := fmt.Sprintf("%s/%s/%s/%s", dateStamp, c.config.Region, service, requestType)

	// Create string to sign
	canonicalRequestHash := sha256Hash(canonicalRequest)
	stringToSign := createStringToSign(algorithm, timeStamp, scope, canonicalRequestHash)

	// Calculate signature
	signingKey := getSigningKey(c.config.SecretKey, c.config.Region, dateStamp)
	signature := calculateSignature(signingKey, stringToSign)

	// Set signed headers - same order as in canonical headers
	var signedHeadersList []string
	for name := range headers {
		signedHeadersList = append(signedHeadersList, name)
	}
	sort.Strings(signedHeadersList)
	signedHeaders := strings.Join(signedHeadersList, ";")

	// Set authorization header
	credential := fmt.Sprintf("%s/%s", c.config.AccessKey, scope)
	authHeader := fmt.Sprintf("%s Credential=%s, SignedHeaders=%s, Signature=%s",
		algorithm, credential, signedHeaders, signature)

	req.Header.Set("Authorization", authHeader)

	return nil
}

// sha256Hash returns SHA256 hash of the given string
func sha256Hash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// createPresignedURL creates a presigned URL for S3 objects (alternative auth method)
func (c *S3Client) createPresignedURL(objectKey string, expiresIn int64) (string, error) {
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	timeStamp := now.Format("20060102T150405Z")

	// Build base URL
	baseURL := c.buildObjectURL(objectKey)
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add X-Amz-Algorithm parameter
	q := u.Query()
	q.Set("X-Amz-Algorithm", algorithm)
	q.Set("X-Amz-Credential", fmt.Sprintf("%s/%s/%s/%s/%s",
		c.config.AccessKey, dateStamp, c.config.Region, service, requestType))
	q.Set("X-Amz-Date", timeStamp)
	q.Set("X-Amz-Expires", fmt.Sprintf("%d", expiresIn))
	q.Set("X-Amz-SignedHeaders", "host")

	u.RawQuery = q.Encode()

	// Create canonical request for presigned URL
	canonicalRequest := createCanonicalRequest(
		"GET",
		u.Path,
		u.RawQuery,
		"host:"+u.Host+"\n",
		"e3b0c44298fc1c149afbf4c8996fb924", // SHA256 of empty string
	)

	// Create string to sign
	scope := fmt.Sprintf("%s/%s/%s/%s", dateStamp, c.config.Region, service, requestType)
	canonicalRequestHash := sha256Hash(canonicalRequest)
	stringToSign := createStringToSign(algorithm, timeStamp, scope, canonicalRequestHash)

	// Calculate signature
	signingKey := getSigningKey(c.config.SecretKey, c.config.Region, dateStamp)
	signature := calculateSignature(signingKey, stringToSign)

	// Add signature to query
	q.Set("X-Amz-Signature", signature)
	u.RawQuery = q.Encode()

	return u.String(), nil
}