package articleflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/tools/s3"
	"github.com/ilkoid/PonchoAiFramework/tools/wildberries"
)

// ImageRef represents a lightweight reference to an image
type ImageRef struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
	S3Key    string `json:"s3_key"`
	Size     int64  `json:"size"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	MimeType string `json:"mime_type"`
}

// TechInfo represents the technical analysis result from Prompt 1
type TechInfo struct {
	ImageID      string                 `json:"image_id"`
	Analysis     map[string]interface{} `json:"analysis"`
	RawJSON      []byte                 `json:"raw_json"`
	ProcessedAt  time.Time              `json:"processed_at"`
	Model        string                 `json:"model"`
	TokenUsage   *interfaces.PonchoUsage `json:"token_usage,omitempty"`
}

// ArticleFlowState holds all state for the article processing flow
type ArticleFlowState struct {
	// Input
	ArticleID string `json:"article_id"`

	// Raw data from S3
	PLMJSON    []byte     `json:"plm_json"`
	Images     []ImageRef `json:"images"`

	// Vision model results (Prompt 1)
	TechAnalysisByImage map[string]*TechInfo `json:"tech_analysis_by_image"`

	// Creative descriptions (Prompt 2)
	CreativeByImage map[string]string `json:"creative_by_image"`

	// Wildberries data
	WBParents          []wildberries.ParentCategory     `json:"wb_parents"`
	WBSubjects         []wildberries.Subject            `json:"wb_subjects"`
	SelectedSubject    *wildberries.Subject             `json:"selected_subject"`
	WBCharacteristics  []wildberries.SubjectCharacteristic `json:"wb_characteristics"`

	// Final output
	FinalWBPayload []byte `json:"final_wb_payload"`

	// Metadata
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Duration   time.Duration `json:"duration"`
}

// NewArticleFlowState creates a new flow state
func NewArticleFlowState(articleID string) *ArticleFlowState {
	return &ArticleFlowState{
		ArticleID:           articleID,
		TechAnalysisByImage: make(map[string]*TechInfo),
		CreativeByImage:     make(map[string]string),
		StartedAt:           time.Now(),
	}
}

// AddImage adds an image reference to the state
func (s *ArticleFlowState) AddImage(img ImageRef) {
	s.Images = append(s.Images, img)
}

// SetTechAnalysis stores technical analysis for an image
func (s *ArticleFlowState) SetTechAnalysis(imageID string, tech *TechInfo) {
	s.TechAnalysisByImage[imageID] = tech
}

// GetTechAnalysis retrieves technical analysis for an image
func (s *ArticleFlowState) GetTechAnalysis(imageID string) (*TechInfo, bool) {
	tech, exists := s.TechAnalysisByImage[imageID]
	return tech, exists
}

// SetCreativeDescription stores creative description for an image
func (s *ArticleFlowState) SetCreativeDescription(imageID, description string) {
	s.CreativeByImage[imageID] = description
}

// GetCreativeDescription retrieves creative description for an image
func (s *ArticleFlowState) GetCreativeDescription(imageID string) (string, bool) {
	desc, exists := s.CreativeByImage[imageID]
	return desc, exists
}

// SetWBParents stores Wildberries parent categories
func (s *ArticleFlowState) SetWBParents(parents []wildberries.ParentCategory) {
	s.WBParents = parents
}

// SetWBSubjects stores Wildberries subjects
func (s *ArticleFlowState) SetWBSubjects(subjects []wildberries.Subject) {
	s.WBSubjects = subjects
}

// SetSelectedSubject stores the selected Wildberries subject
func (s *ArticleFlowState) SetSelectedSubject(subject *wildberries.Subject) {
	s.SelectedSubject = subject
}

// SetWBCharacteristics stores characteristics for the selected subject
func (s *ArticleFlowState) SetWBCharacteristics(characteristics []wildberries.SubjectCharacteristic) {
	s.WBCharacteristics = characteristics
}

// SetFinalWBPayload stores the final Wildberries payload
func (s *ArticleFlowState) SetFinalWBPayload(payload []byte) {
	s.FinalWBPayload = payload
}

// MarkCompleted marks the flow as completed
func (s *ArticleFlowState) MarkCompleted() {
	s.FinishedAt = time.Now()
	s.Duration = s.FinishedAt.Sub(s.StartedAt)
}

// GetImageIDs returns all image IDs
func (s *ArticleFlowState) GetImageIDs() []string {
	ids := make([]string, len(s.Images))
	for i, img := range s.Images {
		ids[i] = img.ID
	}
	return ids
}

// HasImage checks if an image exists in the state
func (s *ArticleFlowState) HasImage(imageID string) bool {
	for _, img := range s.Images {
		if img.ID == imageID {
			return true
		}
	}
	return false
}

// GetImagesWithCreative returns images that have creative descriptions
func (s *ArticleFlowState) GetImagesWithCreative() []ImageRef {
	var result []ImageRef
	for _, img := range s.Images {
		if _, has := s.CreativeByImage[img.ID]; has {
			result = append(result, img)
		}
	}
	return result
}

// GetImagesWithTechAnalysis returns images that have technical analysis
func (s *ArticleFlowState) GetImagesWithTechAnalysis() []ImageRef {
	var result []ImageRef
	for _, img := range s.Images {
		if _, has := s.TechAnalysisByImage[img.ID]; has {
			result = append(result, img)
		}
	}
	return result
}

// ToJSON serializes the state to JSON
func (s *ArticleFlowState) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// ProcessImagesConcurrently processes images concurrently with a worker function
func ProcessImagesConcurrently[T any](
	ctx context.Context,
	images []ImageRef,
	worker func(ctx context.Context, img ImageRef) (T, error),
	maxConcurrency int,
) (map[string]T, error) {
	if maxConcurrency <= 0 {
		maxConcurrency = 3 // Default concurrency
	}

	results := make(map[string]T)
	errors := make(map[string]error)
	var mu sync.Mutex

	// Create semaphore for concurrency control
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	// Process each image
	for _, img := range images {
		wg.Add(1)
		go func(imgRef ImageRef) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Check context
			if ctx.Err() != nil {
				mu.Lock()
				errors[imgRef.ID] = ctx.Err()
				mu.Unlock()
				return
			}

			// Run worker
			result, err := worker(ctx, imgRef)

			mu.Lock()
			if err != nil {
				errors[imgRef.ID] = err
			} else {
				results[imgRef.ID] = result
			}
			mu.Unlock()
		}(img)
	}

	// Wait for all workers to complete
	wg.Wait()

	// Return combined results
	if len(errors) > 0 {
		// Create combined error message
		errMsg := fmt.Sprintf("errors processing %d images:", len(errors))
		for imgID, err := range errors {
			errMsg += fmt.Sprintf("\n  %s: %v", imgID, err)
		}
		return results, fmt.Errorf(errMsg)
	}

	return results, nil
}

// ConvertFromS3Images converts S3 Image structs to ImageRef
func ConvertFromS3Images(s3Images []*s3.Image) []ImageRef {
	images := make([]ImageRef, len(s3Images))
	for i, s3Img := range s3Images {
		images[i] = ImageRef{
			ID:       fmt.Sprintf("img_%d", i),
			Filename: s3Img.Filename,
			URL:      s3Img.URL,
			S3Key:    s3Img.URL, // Assuming URL contains the S3 key
			Size:     s3Img.Size,
			Width:    s3Img.Width,
			Height:   s3Img.Height,
			MimeType: s3Img.ContentType,
		}
	}
	return images
}

// ArticleFlowConfig holds configuration for the article flow
type ArticleFlowConfig struct {
	ImageResize    *ImageResizeConfig    `yaml:"image_resize" json:"image_resize"`
	Concurrency     *ConcurrencyConfig    `yaml:"concurrency" json:"concurrency"`
	Caching         *CachingConfig        `yaml:"caching" json:"caching"`
	ModelParams     *ModelParamsConfig    `yaml:"model_params" json:"model_params"`
}

// ImageResizeConfig holds image resize configuration
type ImageResizeConfig struct {
	Enabled         bool   `yaml:"enabled" json:"enabled"`
	MaxWidth        int    `yaml:"max_width" json:"max_width"`
	MaxHeight       int    `yaml:"max_height" json:"max_height"`
	Quality         int    `yaml:"quality" json:"quality"`
	MaxFileSizeKB   int    `yaml:"max_file_size_kb" json:"max_file_size_kb"`
	TargetFormat    string `yaml:"target_format" json:"target_format"`
}

// ConcurrencyConfig holds concurrency settings
type ConcurrencyConfig struct {
	VisionAnalysisWorkers int `yaml:"vision_analysis_workers" json:"vision_analysis_workers"`
	CreativeWorkers       int `yaml:"creative_workers" json:"creative_workers"`
}

// CachingConfig holds caching settings
type CachingConfig struct {
	WBParentsTTL    time.Duration `yaml:"wb_parents_ttl" json:"wb_parents_ttl"`
	WBSubjectsTTL   time.Duration `yaml:"wb_subjects_ttl" json:"wb_subjects_ttl"`
	WBCharsTTL      time.Duration `yaml:"wb_characteristics_ttl" json:"wb_characteristics_ttl"`
	MaxCacheSize    int           `yaml:"max_cache_size" json:"max_cache_size"`
}

// ModelParamsConfig holds model parameters
type ModelParamsConfig struct {
	VisionModel    string  `yaml:"vision_model" json:"vision_model"`
	TextModel      string  `yaml:"text_model" json:"text_model"`
	Temperature    float32 `yaml:"temperature" json:"temperature"`
	MaxTokens      int     `yaml:"max_tokens" json:"max_tokens"`
}