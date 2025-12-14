package wildberries

import (
	"fmt"
	"time"
)

// WBResponse represents the standard Wildberries API response structure
type WBResponse struct {
	Error           bool   `json:"error"`
	ErrorText       string `json:"errorText"`
	AdditionalError string `json:"additionalErrors"`
}

// WBErrorResponse represents error response from Wildberries API
type WBErrorResponse struct {
	Title      string `json:"title"`
	Detail     string `json:"detail"`
	Status     int    `json:"status"`
	StatusText string `json:"statusText"`
	Timestamp  string `json:"timestamp"`
}

// WBError represents a structured error from Wildberries API
type WBError struct {
	Code    int
	Message string
	Details string
}

func (e *WBError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("WB API Error %d: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("WB API Error %d: %s", e.Code, e.Message)
}

// PingResponse represents the API ping response
type PingResponse struct {
	TS     string `json:"TS"`
	Status string `json:"Status"`
}

// ParentCategory represents a top-level product category
type ParentCategory struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Shard  string `json:"shard,omitempty"`
	Active bool   `json:"active,omitempty"`
}

// Subject represents a product subject (subcategory)
type Subject struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ParentID   int    `json:"parentID"`
	ObjectName string `json:"objectName"`
	Shard      string `json:"shard,omitempty"`
}

// GetSubjectsOptions represents options for getting subjects
type GetSubjectsOptions struct {
	Locale  string `json:"locale,omitempty"`
	Name    string `json:"name,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
	ParentID int   `json:"parentID,omitempty"`
}

// SubjectCharacteristic represents a characteristic for a subject
type SubjectCharacteristic struct {
	CharcID     int    `json:"charcID"`
	SubjectName string `json:"subjectName"`
	SubjectID   int    `json:"subjectID"`
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	UnitName    string `json:"unitName,omitempty"`
	MaxCount    int    `json:"maxCount"`
	Popular     bool   `json:"popular"`
	CharcType   int    `json:"charcType"`
}

// Brand represents a product brand
type Brand struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Logo     string `json:"logo,omitempty"`
	Popular  bool   `json:"popular,omitempty"`
}

// Color represents a color characteristic value
type Color struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	ColorID int    `json:"colorId"`
}

// Gender represents a gender characteristic value
type Gender struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Season represents a season characteristic value
type Season struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// VATRate represents a VAT rate value
type VATRate struct {
	ID    int     `json:"id"`
	Vat   float64 `json:"vat"`
	Name  string `json:"name"`
	Title string `json:"title"`
}

// Country represents a country of origin
type Country struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code,omitempty"`
}

// HSCode represents an HS code for customs
type HSCode struct {
	ID       int    `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	ParentID int    `json:"parentId,omitempty"`
}

// FashionProduct represents a fashion product with Wildberries-specific fields
type FashionProduct struct {
	// Basic product info
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Brand       string            `json:"brand"`
	SubjectID   int               `json:"subjectId"`

	// Fashion-specific characteristics
	Characteristics map[string]interface{} `json:"characteristics"`

	// Visual content
	Images []ProductImage `json:"images"`

	// Metadata
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ProductImage represents a product image
type ProductImage struct {
	URL       string `json:"url"`
	Name      string `json:"name,omitempty"`
	Main      bool   `json:"main"`
	Index     int    `json:"index"`
}

// FashionAnalysis represents the result of AI analysis of a fashion item
type FashionAnalysis struct {
	// Product classification
	SubjectID       int    `json:"subjectId,omitempty"`
	ParentCategory  string `json:"parentCategory,omitempty"`
	Subject         string `json:"subject,omitempty"`

	// Fashion characteristics
	Type        string `json:"type,omitempty"`
	Style       string `json:"style,omitempty"`
	Season      string `json:"season,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Material    string `json:"material,omitempty"`
	Color       string `json:"color,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
	Size        string `json:"size,omitempty"`

	// Descriptive content
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	// Metadata
	Confidence  float64 `json:"confidence,omitempty"`
	Model       string  `json:"model,omitempty"`
	ProcessedAt time.Time `json:"processedAt"`
}