#!/bin/bash

# S3 Explorer Demo Script
# Demonstrates common operations with S3 data

echo "=== PonchoAiFramework S3 Explorer Demo ==="
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Creating .env file from template..."
    cp .env.example .env
    echo "Please edit .env file with your S3 credentials"
    echo "Required variables:"
    echo "  S3_ACCESS_KEY"
    echo "  S3_SECRET_KEY"
    echo "  S3_BUCKET (default: plm-ai)"
    echo ""
    exit 1
fi

# Load environment variables
source .env

# Check required variables
if [ -z "$S3_ACCESS_KEY" ] || [ -z "$S3_SECRET_KEY" ]; then
    echo "Error: S3 credentials not found in .env"
    echo "Please set S3_ACCESS_KEY and S3_SECRET_KEY"
    exit 1
fi

# Build the utility
echo "Building s3-explorer..."
go build -o s3-explorer ./cmd/s3-explorer
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful!"
echo ""

# Create output directory
OUTPUT_DIR="./demo-output"
mkdir -p $OUTPUT_DIR

echo "=== Demo Commands ==="
echo ""

# Demo 1: List root objects
echo "1. Listing root objects..."
./s3-explorer list "" -format=summary | head -20
echo ""
echo "---"
echo ""

# Demo 2: Search for articles
echo "2. Searching for articles..."
./s3-explorer list "articles/" 2>/dev/null | head -10
if [ $? -ne 0 ]; then
    echo "No 'articles/' prefix found or access denied"
fi
echo "---"
echo ""

# Demo 3: Try common fashion data prefixes
echo "3. Searching for common fashion data prefixes..."
for prefix in "fashion/" "clothing/" "products/" "wildberries/" "wb/" "data/"; do
    echo "  Checking '$prefix'..."
    ./s3-explorer list "$prefix" 2>/dev/null | head -3 | grep -v "No objects found" || echo "    (not found or no access)"
done
echo "---"
echo ""

# Demo 4: Show bucket statistics
echo "4. Bucket statistics (dry run)..."
./s3-explorer stats "" -verbose 2>/dev/null | head -20 || echo "Stats failed - check permissions"
echo "---"
echo ""

# Demo 5: Parse any found data
echo "5. Attempting to parse JSON data..."
# Try to find and parse a JSON file
JSON_FILE=$(./s3-explorer list "" 2>/dev/null | grep ".json" | head -1 | cut -d'"' -f2)
if [ ! -z "$JSON_FILE" ]; then
    echo "Found JSON file: $JSON_FILE"
    echo "Content preview:"
    ./s3-explorer get "$JSON_FILE" 2>/dev/null | head -20 || echo "Failed to get file content"
else
    echo "No JSON files found in root"
fi
echo "---"
echo ""

# Demo 6: Show available commands
echo "6. Available commands for manual testing:"
echo "  ./s3-explorer list <prefix>                    - List objects"
echo "  ./s3-explorer get <key>                       - Download object"
echo "  ./s3-explorer parse <prefix>                   - Parse and analyze"
echo "  ./s3-explorer stats <prefix>                   - Show statistics"
echo "  ./s3-extractor extract <prefix> -output-dir=./data - Extract files"
echo ""
echo "Examples:"
echo "  ./s3-explorer list articles/"
echo "  ./s3-explorer get articles/12612003.json"
echo "  ./s3-extractor extract articles/ -dry-run"
echo ""

# Cleanup
echo "Cleaning up..."
rm -f s3-explorer
echo "Demo completed!"
echo ""
echo "Notes:"
echo "- This demo uses production S3 credentials"
echo "- Some operations may fail due to permissions"
echo "- Check the logs for detailed error messages"
echo "- Use -verbose flag for more detailed output"