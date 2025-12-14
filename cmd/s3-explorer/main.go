package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ilkoid/PonchoAiFramework/tools/s3"
)

// Config holds the application configuration
type Config struct {
	S3 struct {
		AccessKey string `env:"S3_ACCESS_KEY"`
		SecretKey string `env:"S3_SECRET_KEY"`
		Endpoint   string `env:"S3_ENDPOINT" default:"https://storage.yandexcloud.net"`
		Region     string `env:"S3_REGION" default:"ru-central1"`
		Bucket     string `env:"S3_BUCKET" default:"plm-ai"`
		URL        string `env:"S3_URL" default:"https://storage.yandexcloud.net"`
		UseSSL     bool   `env:"S3_USE_SSL" default:"true"`
	}

	Output struct {
		Dir     string `flag:"output-dir" default:"./s3-extract"`
		Verbose bool   `flag:"verbose"`
		DryRun  bool   `flag:"dry-run"`
	}
}

// S3Explorer handles S3 exploration and data extraction
type S3Explorer struct {
	config *Config
	client *s3.S3Client
	logger s3.Logger
}

func main() {
	// Parse command line flags
	config := parseFlags()

	// Load environment variables
	loadEnvVars(config)

	// Create explorer
	explorer, err := NewS3Explorer(config)
	if err != nil {
		log.Fatalf("Failed to create explorer: %v", err)
	}

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "list":
		handleListCommand(explorer, os.Args[2:])
	case "get":
		handleGetCommand(explorer, os.Args[2:])
	case "stats":
		handleStatsCommand(explorer, os.Args[2:])
	case "extract":
		handleExtractCommand(explorer, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.Output.Dir, "output-dir", "./s3-extract", "Output directory for extracted files")
	flag.BoolVar(&config.Output.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&config.Output.DryRun, "dry-run", false, "Dry run - don't download files")

	flag.Parse()
	return config
}

func loadEnvVars(config *Config) {
	if config.S3.AccessKey == "" {
		config.S3.AccessKey = os.Getenv("S3_ACCESS_KEY")
	}
	if config.S3.SecretKey == "" {
		config.S3.SecretKey = os.Getenv("S3_SECRET_KEY")
	}
	if config.S3.Endpoint == "" {
		config.S3.Endpoint = os.Getenv("S3_ENDPOINT")
	}
	if config.S3.Region == "" {
		config.S3.Region = os.Getenv("S3_REGION")
	}
	if config.S3.Bucket == "" {
		config.S3.Bucket = os.Getenv("S3_BUCKET")
	}
	if config.S3.URL == "" {
		config.S3.URL = os.Getenv("S3_URL")
	}
	if envUseSSL := os.Getenv("S3_USE_SSL"); envUseSSL != "" {
		if useSSL, err := strconv.ParseBool(envUseSSL); err == nil {
			config.S3.UseSSL = useSSL
		}
	}
}

func NewS3Explorer(config *Config) (*S3Explorer, error) {
	// Validate required fields
	if config.S3.AccessKey == "" || config.S3.SecretKey == "" {
		return nil, fmt.Errorf("S3 credentials are required (S3_ACCESS_KEY and S3_SECRET_KEY)")
	}
	if config.S3.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required (S3_BUCKET)")
	}

	// Create logger
	logger := s3.NewDefaultLogger()

	// Create S3 client configuration
	s3Config := &s3.ClientConfig{
		AccessKey:  config.S3.AccessKey,
		SecretKey:  config.S3.SecretKey,
		Endpoint:   config.S3.Endpoint,
		Region:     config.S3.Region,
		Bucket:     config.S3.Bucket,
		URL:        config.S3.URL,
		UseSSL:     config.S3.UseSSL,
		Timeout:    30,
		MaxRetries: 3,
	}

	// Create S3 client
	client, err := s3.NewS3Client(s3Config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	return &S3Explorer{
		config: config,
		client: client,
		logger: logger,
	}, nil
}

func printUsage() {
	fmt.Printf("PonchoAiFramework S3 Explorer\n\n")
	fmt.Printf("Usage: s3-explorer [flags] <command> [arguments]\n\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("  list [prefix]        List articles with given prefix\n")
	fmt.Printf("  get <article_id>     Download article data (JSON + images)\n")
	fmt.Printf("  stats [prefix]      Show statistics about articles\n")
	fmt.Printf("  extract <article_id> Extract article to output directory\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  s3-explorer list\n")
	fmt.Printf("  s3-explorer get 12612003\n")
	fmt.Printf("  s3-explorer stats\n")
	fmt.Printf("  s3-extractor extract 12612003 -output-dir=./data\n\n")
	fmt.Printf("Flags:\n")
	flag.PrintDefaults()
}

func handleListCommand(explorer *S3Explorer, args []string) {
	prefix := ""
	if len(args) > 0 {
		prefix = args[0]
	}

	explorer.logger.Info("Listing articles", "prefix", prefix)

	ctx := context.Background()
	listReq := &s3.ListRequest{
		Prefix:   prefix,
		MaxItems: 1000,
	}

	response, err := explorer.client.ListArticles(ctx, listReq)
	if err != nil {
		log.Fatalf("Failed to list articles: %v", err)
	}

	if !response.Success {
		log.Fatalf("Failed to list articles: %s", response.Error.Message)
	}

	fmt.Printf("Found %d articles:\n", len(response.Articles))
	for i, articleID := range response.Articles {
		fmt.Printf("%4d. %s\n", i+1, articleID)
	}

	if response.Metadata != nil {
		fmt.Printf("\nMetadata:\n")
		fmt.Printf("  Request ID: %s\n", response.Metadata.RequestID)
		fmt.Printf("  Duration: %d ms\n", response.Metadata.Duration)
		fmt.Printf("  Bucket: %s\n", response.Metadata.Bucket)
		fmt.Printf("  Region: %s\n", response.Metadata.Region)
	}
}

func handleGetCommand(explorer *S3Explorer, args []string) {
	if len(args) == 0 {
		log.Fatal("Please provide article ID")
	}

	articleID := args[0]
	explorer.logger.Info("Getting article", "article_id", articleID)

	ctx := context.Background()
	downloadReq := &s3.DownloadRequest{
		ArticleID:     articleID,
		IncludeImages: true,
		MaxImages:     10,
		ImageOptions:  s3.DefaultImageProcessingOptions(),
	}

	response, err := explorer.client.DownloadArticle(ctx, downloadReq)
	if err != nil {
		log.Fatalf("Failed to download article: %v", err)
	}

	if !response.Success {
		log.Fatalf("Failed to download article: %s", response.Error.Message)
	}

	// Output results
	if response.Article != nil {
		// Save to file or output to console
		outputFile := fmt.Sprintf("%s/%s.json", explorer.config.Output.Dir, articleID)

		if err := os.MkdirAll(explorer.config.Output.Dir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}

		jsonData, err := json.MarshalIndent(response.Article, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}

		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}

		fmt.Printf("Article saved: %s\n", outputFile)
		fmt.Printf("JSON data size: %d bytes\n", len(response.Article.JSONData))
		fmt.Printf("Images count: %d\n", len(response.Article.Images))
		fmt.Printf("Total size: %d bytes\n", response.Article.Metadata.TotalSize)

		// Save images if any
		if len(response.Article.Images) > 0 {
			imagesDir := fmt.Sprintf("%s/%s/images", explorer.config.Output.Dir, articleID)
			if err := os.MkdirAll(imagesDir, 0755); err == nil {
				for _, img := range response.Article.Images {
					imgFile := fmt.Sprintf("%s/%s", imagesDir, img.Filename)
					if data, err := base64.StdEncoding.DecodeString(img.Data); err == nil {
						if err := os.WriteFile(imgFile, data, 0644); err == nil {
							fmt.Printf("  Saved image: %s (%d bytes)\n", imgFile, len(data))
						}
					}
				}
			}
		}
	}
}

func handleStatsCommand(explorer *S3Explorer, args []string) {
	prefix := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		prefix = args[0]
	}

	explorer.logger.Info("Getting statistics", "prefix", prefix)

	ctx := context.Background()
	listReq := &s3.ListRequest{
		Prefix:   prefix,
		MaxItems: 10000, // Large number for stats
	}

	response, err := explorer.client.ListArticles(ctx, listReq)
	if err != nil {
		log.Fatalf("Failed to list articles: %v", err)
	}

	if !response.Success {
		log.Fatalf("Failed to list articles: %s", response.Error.Message)
	}

	// Download sample articles for size analysis
	sampleSize := min(10, len(response.Articles))
	totalSize := int64(0)
	totalImages := 0
	hasImages := 0

	if explorer.config.Output.Verbose {
		fmt.Printf("Sampling %d articles for size analysis...\n", sampleSize)
	}

	for idx := 0; idx < sampleSize && idx < len(response.Articles); idx++ {
		articleID := response.Articles[idx]

		downloadReq := &s3.DownloadRequest{
			ArticleID:     articleID,
			IncludeImages: false, // No images for stats
		}

		downloadResponse, err := explorer.client.DownloadArticle(ctx, downloadReq)
		if err != nil {
			if explorer.config.Output.Verbose {
				fmt.Printf("Failed to sample article %s: %v\n", articleID, err)
			}
			continue
		}

		if downloadResponse.Success && downloadResponse.Article != nil {
			totalSize += downloadResponse.Article.Metadata.TotalSize
			totalImages += downloadResponse.Article.Metadata.ImageCount
			if downloadResponse.Article.Metadata.ImageCount > 0 {
				hasImages++
			}
		}
	}

	// Calculate estimates
	avgSize := int64(0)
	if sampleSize > 0 {
		avgSize = totalSize / int64(sampleSize)
	}

	// Output statistics
	fmt.Printf("S3 Bucket Statistics\n")
	fmt.Printf("===================\n")
	fmt.Printf("Total Articles: %d\n", len(response.Articles))
	fmt.Printf("Sample Size: %d articles\n", sampleSize)
	fmt.Printf("Average Size: %s per article\n", formatBytes(avgSize))
	fmt.Printf("Total Estimated Size: %s\n", formatBytes(int64(len(response.Articles))*avgSize))
	fmt.Printf("Articles with Images: %d (%.1f%%)\n", hasImages, float64(hasImages)/float64(sampleSize)*100)
	fmt.Printf("Average Images per Article: %.1f\n", float64(totalImages)/float64(sampleSize))

	if response.Metadata != nil {
		fmt.Printf("\nRequest Details:\n")
		fmt.Printf("  Request ID: %s\n", response.Metadata.RequestID)
		fmt.Printf("  Duration: %d ms\n", response.Metadata.Duration)
		fmt.Printf("  Bucket: %s\n", response.Metadata.Bucket)
		fmt.Printf("  Region: %s\n", response.Metadata.Region)
	}
}

func handleExtractCommand(explorer *S3Explorer, args []string) {
	if len(args) == 0 {
		log.Fatal("Please provide article ID")
	}

	articleID := args[0]
	explorer.logger.Info("Extracting article", "article_id", articleID, "dry-run", explorer.config.Output.DryRun)

	if explorer.config.Output.DryRun {
		fmt.Printf("DRY RUN - Would extract article: %s\n", articleID)
		fmt.Printf("Output directory: %s\n", explorer.config.Output.Dir)
		return
	}

	ctx := context.Background()
	downloadReq := &s3.DownloadRequest{
		ArticleID:     articleID,
		IncludeImages: true,
		MaxImages:     20,
		ImageOptions:  s3.DefaultImageProcessingOptions(),
	}

	response, err := explorer.client.DownloadArticle(ctx, downloadReq)
	if err != nil {
		log.Fatalf("Failed to download article: %v", err)
	}

	if !response.Success {
		log.Fatalf("Failed to download article: %s", response.Error.Message)
	}

	if response.Article == nil {
		log.Fatal("No article data received")
	}

	// Create output directory
	outputDir := fmt.Sprintf("%s/%s", explorer.config.Output.Dir, articleID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Save JSON data
	jsonFile := fmt.Sprintf("%s/%s.json", outputDir, articleID)
	jsonData, err := json.MarshalIndent(response.Article, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		log.Fatalf("Failed to write JSON file: %v", err)
	}

	extractedImages := 0
	if len(response.Article.Images) > 0 {
		imagesDir := fmt.Sprintf("%s/images", outputDir)
		if err := os.MkdirAll(imagesDir, 0755); err == nil {
			for _, img := range response.Article.Images {
				imgFile := fmt.Sprintf("%s/%s", imagesDir, img.Filename)
				if data, err := base64.StdEncoding.DecodeString(img.Data); err == nil {
					if err := os.WriteFile(imgFile, data, 0644); err == nil {
						extractedImages++
					}
				}
			}
		}
	}

	fmt.Printf("Extraction completed:\n")
	fmt.Printf("  Article ID: %s\n", articleID)
	fmt.Printf("  JSON saved: %s\n", jsonFile)
	fmt.Printf("  Images extracted: %d/%d\n", extractedImages, len(response.Article.Images))
	fmt.Printf("  Total size: %s\n", formatBytes(response.Article.Metadata.TotalSize))
	fmt.Printf("  Output directory: %s\n", outputDir)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}