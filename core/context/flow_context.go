package context

import (
	"fmt"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// FlowContext provides a shared state storage for workflow execution
// It allows steps to accumulate and share data throughout the flow lifecycle
type FlowContext interface {
	// Basic state management
	Set(key string, value interface{}) error
	Get(key string) (interface{}, bool)
	Delete(key string) bool
	Has(key string) bool
	Clear()

	// Type-safe operations
	SetString(key, value string) error
	GetString(key string) (string, error)
	SetBytes(key string, value []byte) error
	GetBytes(key string) ([]byte, error)
	SetInt(key string, value int) error
	GetInt(key string) (int, error)
	SetFloat(key string, value float64) error
	GetFloat(key string) (float64, error)
	SetBool(key string, value bool) error
	GetBool(key string) (bool, error)

	// Array/List operations
	SetArray(key string, values []interface{}) error
	GetArray(key string) ([]interface{}, error)
	AppendToArray(key string, value interface{}) error
	GetArraySize(key string) (int, error)

	// Object operations
	SetObject(key string, obj interface{}) error
	GetObject(key string, target interface{}) error

	// Media-specific operations
	SetMedia(key string, media *MediaData) error
	GetMedia(key string) (*MediaData, error)
	GetAllMedia(prefix string) ([]*MediaData, error)
	AccumulateMedia(prefix string, mediaList []*MediaData) error

	// Metadata and utilities
	Keys() []string
	Size() int
	Clone() FlowContext
	Merge(other FlowContext) error

	// Serialization for persistence/debugging
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	ToJSON() (string, error)

	// Context lifecycle
	ID() string
	CreatedAt() time.Time
	Parent() FlowContext
	CreateChild() FlowContext

	// Logging and debugging
	SetLogger(logger interfaces.Logger)
	GetLogger() interfaces.Logger
	Dump() map[string]interface{}
	PrintState()
}

// MediaData represents media content with metadata
type MediaData struct {
	URL      string                 `json:"url,omitempty"`
	Bytes    []byte                 `json:"-"` // Don't serialize bytes
	MimeType string                 `json:"mime_type"`
	Size     int64                  `json:"size"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewMediaDataFromURL creates MediaData from URL
func NewMediaDataFromURL(url, mimeType string) *MediaData {
	return &MediaData{
		URL:      url,
		MimeType: mimeType,
		Metadata: make(map[string]interface{}),
	}
}

// NewMediaDataFromBytes creates MediaData from byte slice
func NewMediaDataFromBytes(bytes []byte, mimeType string) *MediaData {
	return &MediaData{
		Bytes:    bytes,
		MimeType: mimeType,
		Size:     int64(len(bytes)),
		Metadata: make(map[string]interface{}),
	}
}

// GetDataURL returns data URL for byte-based media
func (m *MediaData) GetDataURL() string {
	if len(m.Bytes) == 0 {
		return m.URL
	}
	// Simple base64 encoding - in real implementation use proper encoding
	return fmt.Sprintf("data:%s;base64,%x", m.MimeType, m.Bytes)
}

// ContextConfig provides configuration for FlowContext
type ContextConfig struct {
	ID                string
	MaxSize          int           // Maximum number of keys
	TTL              time.Duration // Time to live
	EnableSerialization bool
	Logger           interfaces.Logger
	Parent           FlowContext
}

// DefaultContextConfig returns default context configuration
func DefaultContextConfig() *ContextConfig {
	return &ContextConfig{
		ID:                generateID(),
		MaxSize:          1000,
		TTL:              time.Hour * 24,
		EnableSerialization: true,
		Logger:           interfaces.NewDefaultLogger(),
	}
}