package s3


// DefaultClientConfig returns default S3 client configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		URL:        "https://storage.yandexcloud.net",
		Region:     "ru-central1",
		Bucket:     "plm-ai",
		Endpoint:   "storage.yandexcloud.net",
		UseSSL:     true,
		Timeout:    30,
		MaxRetries: 3,
	}
}

// DefaultImageProcessingOptions returns default image processing options
func DefaultImageProcessingOptions() *ImageProcessingOptions {
	return &ImageProcessingOptions{
		Enabled:          true,
		MaxWidth:         640,
		MaxHeight:        480,
		Quality:          90,
		MaxSizeBytes:     90000, // 90KB
		Format:           "jpeg",
		PreserveMetadata: false,
	}
}