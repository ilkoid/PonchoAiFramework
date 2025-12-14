package articleflow

import (
	"context"
	"testing"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
	"github.com/ilkoid/PonchoAiFramework/tools/wildberries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWBCache is a mock implementation of WBCache interface
type MockWBCache struct {
	mock.Mock
}

func (m *MockWBCache) GetParents(ctx context.Context) ([]wildberries.WBParentCategory, error) {
	args := m.Called(ctx)
	return args.Get(0).([]wildberries.WBParentCategory), args.Error(1)
}

func (m *MockWBCache) GetSubjects(ctx context.Context) ([]wildberries.WBSubject, error) {
	args := m.Called(ctx)
	return args.Get(0).([]wildberries.WBSubject), args.Error(1)
}

func (m *MockWBCache) GetCharacteristics(ctx context.Context, subjectID int) ([]wildberries.WBCharacteristic, error) {
	args := m.Called(ctx, subjectID)
	return args.Get(0).([]wildberries.WBCharacteristic), args.Error(1)
}

func (m *MockWBCache) Invalidate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockWBCache) InvalidateSubject(ctx context.Context, subjectID int) error {
	args := m.Called(ctx, subjectID)
	return args.Error(0)
}

func TestArticleFlowState(t *testing.T) {
	t.Run("NewArticleFlowState", func(t *testing.T) {
		articleID := "test-123"
		state := NewArticleFlowState(articleID)

		assert.Equal(t, articleID, state.ArticleID)
		assert.NotNil(t, state.TechAnalysisByImage)
		assert.NotNil(t, state.CreativeByImage)
		assert.NotZero(t, state.StartedAt)
		assert.Zero(t, state.FinishedAt)
		assert.Zero(t, state.Duration)
	})

	t.Run("AddImage", func(t *testing.T) {
		state := NewArticleFlowState("test")
		img := ImageRef{
			ID:       "img1",
			Filename: "test.jpg",
			URL:      "http://example.com/test.jpg",
		}

		state.AddImage(img)

		assert.Len(t, state.Images, 1)
		assert.Equal(t, img, state.Images[0])
	})

	t.Run("TechAnalysis", func(t *testing.T) {
		state := NewArticleFlowState("test")
		imageID := "img1"
		tech := &TechInfo{
			ImageID: imageID,
			Analysis: map[string]interface{}{
				"type": "dress",
			},
		}

		state.SetTechAnalysis(imageID, tech)

		// Test GetTechAnalysis
		retrieved, exists := state.GetTechAnalysis(imageID)
		require.True(t, exists)
		assert.Equal(t, tech, retrieved)

		// Test non-existent
		_, exists = state.GetTechAnalysis("non-existent")
		assert.False(t, exists)
	})

	t.Run("CreativeDescriptions", func(t *testing.T) {
		state := NewArticleFlowState("test")
		imageID := "img1"
		description := "A beautiful red dress"

		state.SetCreativeDescription(imageID, description)

		// Test GetCreativeDescription
		retrieved, exists := state.GetCreativeDescription(imageID)
		require.True(t, exists)
		assert.Equal(t, description, retrieved)

		// Test non-existent
		_, exists = state.GetCreativeDescription("non-existent")
		assert.False(t, exists)
	})

	t.Run("WildberriesData", func(t *testing.T) {
		state := NewArticleFlowState("test")

		parents := []wildberries.WBParentCategory{
			{ID: 1, Name: "Clothing"},
		}
		state.SetWBParents(parents)
		assert.Equal(t, parents, state.WBParents)

		subjects := []wildberries.WBSubject{
			{ID: 123, Name: "Dresses"},
		}
		state.SetWBSubjects(subjects)
		assert.Equal(t, subjects, state.WBSubjects)

		selected := &wildberries.WBSubject{ID: 123, Name: "Dresses"}
		state.SetSelectedSubject(selected)
		assert.Equal(t, selected, state.SelectedSubject)

		characteristics := []wildberries.WBCharacteristic{
			{ID: 1, Name: "Color", Type: "string"},
		}
		state.SetWBCharacteristics(characteristics)
		assert.Equal(t, characteristics, state.WBCharacteristics)
	})

	t.Run("FinalPayload", func(t *testing.T) {
		state := NewArticleFlowState("test")
		payload := []byte(`{"test": "data"}`)

		state.SetFinalWBPayload(payload)
		assert.Equal(t, payload, state.FinalWBPayload)
	})

	t.Run("MarkCompleted", func(t *testing.T) {
		state := NewArticleFlowState("test")
		time.Sleep(time.Millisecond) // Ensure some duration

		state.MarkCompleted()

		assert.NotZero(t, state.FinishedAt)
		assert.NotZero(t, state.Duration)
		assert.True(t, state.FinishedAt.After(state.StartedAt))
	})

	t.Run("GetImageIDs", func(t *testing.T) {
		state := NewArticleFlowState("test")
		state.Images = []ImageRef{
			{ID: "img1"},
			{ID: "img2"},
		}

		ids := state.GetImageIDs()
		assert.Equal(t, []string{"img1", "img2"}, ids)
	})

	t.Run("HasImage", func(t *testing.T) {
		state := NewArticleFlowState("test")
		state.Images = []ImageRef{
			{ID: "img1"},
			{ID: "img2"},
		}

		assert.True(t, state.HasImage("img1"))
		assert.False(t, state.HasImage("img3"))
	})

	t.Run("GetImagesWithCreative", func(t *testing.T) {
		state := NewArticleFlowState("test")
		state.Images = []ImageRef{
			{ID: "img1"},
			{ID: "img2"},
		}
		state.SetCreativeDescription("img1", "description 1")

		withCreative := state.GetImagesWithCreative()
		assert.Len(t, withCreative, 1)
		assert.Equal(t, "img1", withCreative[0].ID)
	})

	t.Run("GetImagesWithTechAnalysis", func(t *testing.T) {
		state := NewArticleFlowState("test")
		state.Images = []ImageRef{
			{ID: "img1"},
			{ID: "img2"},
		}
		state.SetTechAnalysis("img1", &TechInfo{ImageID: "img1"})

		withTech := state.GetImagesWithTechAnalysis()
		assert.Len(t, withTech, 1)
		assert.Equal(t, "img1", withTech[0].ID)
	})

	t.Run("ToJSON", func(t *testing.T) {
		state := NewArticleFlowState("test")
		state.ArticleID = "test-123"

		jsonData, err := state.ToJSON()
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "test-123")
	})
}

func TestProcessImagesConcurrently(t *testing.T) {
	t.Run("SuccessfulProcessing", func(t *testing.T) {
		ctx := context.Background()
		images := []ImageRef{
			{ID: "img1"},
			{ID: "img2"},
			{ID: "img3"},
		}

		worker := func(ctx context.Context, img ImageRef) (string, error) {
			return "result-" + img.ID, nil
		}

		results, err := ProcessImagesConcurrently(ctx, images, worker, 2)
		require.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, "result-img1", results["img1"])
		assert.Equal(t, "result-img2", results["img2"])
		assert.Equal(t, "result-img3", results["img3"])
	})

	t.Run("WithError", func(t *testing.T) {
		ctx := context.Background()
		images := []ImageRef{
			{ID: "img1"},
			{ID: "img2"},
		}

		worker := func(ctx context.Context, img ImageRef) (string, error) {
			if img.ID == "img2" {
				return "", fmt.Errorf("processing failed")
			}
			return "result-" + img.ID, nil
		}

		results, err := ProcessImagesConcurrently(ctx, images, worker, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "img2: processing failed")
		assert.Len(t, results, 1)
		assert.Equal(t, "result-img1", results["img1"])
	})

	t.Run("WithContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		images := []ImageRef{
			{ID: "img1"},
			{ID: "img2"},
		}

		worker := func(ctx context.Context, img ImageRef) (string, error) {
			if img.ID == "img1" {
				cancel() // Cancel after first image
			}
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "result-" + img.ID, nil
			}
		}

		results, err := ProcessImagesConcurrently(ctx, images, worker, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("ZeroConcurrency", func(t *testing.T) {
		ctx := context.Background()
		images := []ImageRef{{ID: "img1"}}

		worker := func(ctx context.Context, img ImageRef) (string, error) {
			return "result", nil
		}

		results, err := ProcessImagesConcurrently(ctx, images, worker, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestConvertFromS3Images(t *testing.T) {
	s3Images := []*s3.Image{
		{
			Filename:    "test1.jpg",
			URL:         "s3://bucket/test1.jpg",
			ContentType: "image/jpeg",
			Size:        1024,
			Width:       800,
			Height:      600,
		},
		{
			Filename:    "test2.png",
			URL:         "s3://bucket/test2.png",
			ContentType: "image/png",
			Size:        2048,
			Width:       1024,
			Height:      768,
		},
	}

	images := ConvertFromS3Images(s3Images)

	assert.Len(t, images, 2)

	// First image
	assert.Equal(t, "img_0", images[0].ID)
	assert.Equal(t, "test1.jpg", images[0].Filename)
	assert.Equal(t, "s3://bucket/test1.jpg", images[0].URL)
	assert.Equal(t, "s3://bucket/test1.jpg", images[0].S3Key)
	assert.Equal(t, int64(1024), images[0].Size)
	assert.Equal(t, 800, images[0].Width)
	assert.Equal(t, 600, images[0].Height)
	assert.Equal(t, "image/jpeg", images[0].MimeType)

	// Second image
	assert.Equal(t, "img_1", images[1].ID)
	assert.Equal(t, "test2.png", images[1].Filename)
	assert.Equal(t, "s3://bucket/test2.png", images[1].URL)
	assert.Equal(t, "s3://bucket/test2.png", images[1].S3Key)
	assert.Equal(t, int64(2048), images[1].Size)
	assert.Equal(t, 1024, images[1].Width)
	assert.Equal(t, 768, images[1].Height)
	assert.Equal(t, "image/png", images[1].MimeType)
}

func TestDefaultFlowConfig(t *testing.T) {
	config := DefaultFlowConfig()

	assert.NotNil(t, config)
	assert.NotNil(t, config.ImageResize)
	assert.True(t, config.ImageResize.Enabled)
	assert.Equal(t, 1024, config.ImageResize.MaxWidth)
	assert.Equal(t, 1024, config.ImageResize.MaxHeight)
	assert.Equal(t, 85, config.ImageResize.Quality)
	assert.Equal(t, 500, config.ImageResize.MaxFileSizeKB)
	assert.Equal(t, "jpeg", config.ImageResize.TargetFormat)

	assert.NotNil(t, config.Concurrency)
	assert.Equal(t, 3, config.Concurrency.VisionAnalysisWorkers)
	assert.Equal(t, 2, config.Concurrency.CreativeWorkers)

	assert.NotNil(t, config.Caching)
	assert.Equal(t, 24*time.Hour, config.Caching.WBParentsTTL)
	assert.Equal(t, 12*time.Hour, config.Caching.WBSubjectsTTL)
	assert.Equal(t, 6*time.Hour, config.Caching.WBCharsTTL)
	assert.Equal(t, 1000, config.Caching.MaxCacheSize)

	assert.NotNil(t, config.ModelParams)
	assert.Equal(t, "glm-4.6v-flash", config.ModelParams.VisionModel)
	assert.Equal(t, "deepseek-chat", config.ModelParams.TextModel)
	assert.Equal(t, float32(0.3), config.ModelParams.Temperature)
	assert.Equal(t, 2000, config.ModelParams.MaxTokens)
}