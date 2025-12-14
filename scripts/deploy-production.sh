#!/bin/bash
# Production Deployment Script for PonchoAI Framework
# This script handles the complete deployment process with safety checks

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.production.yml"
ENV_FILE="${PROJECT_ROOT}/.env.production"

# Version information
VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    # Check if Docker is installed and running
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed or not in PATH"
    fi

    if ! docker info &> /dev/null; then
        error "Docker daemon is not running"
    fi

    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is not installed or not in PATH"
    fi

    # Check if required files exist
    if [[ ! -f "$COMPOSE_FILE" ]]; then
        error "Docker Compose file not found: $COMPOSE_FILE"
    fi

    if [[ ! -f "$ENV_FILE" ]]; then
        warn "Environment file not found: $ENV_FILE"
        info "Creating template environment file..."
        create_env_template
    fi

    # Check if we're in a git repository
    if ! git rev-parse --git-dir &> /dev/null; then
        warn "Not in a git repository, version detection may be inaccurate"
    fi

    log "Prerequisites check completed"
}

# Create environment file template
create_env_template() {
    cat > "$ENV_FILE" << 'EOF'
# Production Environment Variables
# Copy this file and fill in your actual values

# Application Settings
GIN_MODE=release
LOG_LEVEL=info
CONFIG_PATH=/app/config/config-production.yaml

# API Keys (REQUIRED - Replace with your actual keys)
DEEPSEEK_API_KEY=your_deepseek_api_key_here
ZAI_API_KEY=your_zai_api_key_here

# AWS Configuration (Required for S3 tool)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_aws_access_key_here
AWS_SECRET_ACCESS_KEY=your_aws_secret_access_key_here
S3_BUCKET_NAME=your_s3_bucket_name_here

# Optional News API Keys
NEWS_API_KEY=your_news_api_key_here
GUARDIAN_API_KEY=your_guardian_api_key_here

# Redis Configuration
REDIS_PASSWORD=your_secure_redis_password_here

# Grafana Configuration
GRAFANA_PASSWORD=your_secure_grafana_password_here

# Monitoring Configuration
JAEGER_ENDPOINT=http://jaeger:14268/api/traces

# Alerting Configuration
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
SMTP_SERVER=smtp.yourcompany.com
SMTP_PORT=587
SMTP_USERNAME=your_smtp_username
SMTP_PASSWORD=your_smtp_password
EOF

    warn "Please edit $ENV_FILE with your actual values before deploying"
}

# Validate environment
validate_environment() {
    log "Validating environment configuration..."

    # Check for required environment variables
    local required_vars=(
        "DEEPSEEK_API_KEY"
        "ZAI_API_KEY"
        "REDIS_PASSWORD"
        "GRAFANA_PASSWORD"
    )

    local missing_vars=()
    for var in "${required_vars[@]}"; do
        if ! grep -q "^${var}=" "$ENV_FILE" || grep -q "${var}=your_" "$ENV_FILE"; then
            missing_vars+=("$var")
        fi
    done

    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        error "Missing or placeholder values for required variables: ${missing_vars[*]}"
    fi

    # Validate SSL certificates if they exist
    local ssl_dir="${PROJECT_ROOT}/nginx/ssl"
    if [[ -d "$ssl_dir" ]]; then
        if [[ -f "$ssl_dir/cert.pem" && -f "$ssl_dir/key.pem" ]]; then
            log "SSL certificates found and will be used"
        else
            warn "SSL directory exists but certificates are incomplete"
        fi
    else
        info "No SSL certificates found, HTTP only mode"
    fi

    log "Environment validation completed"
}

# Build Docker image
build_image() {
    log "Building Docker image..."

    cd "$PROJECT_ROOT"

    docker build \
        --file Dockerfile.production \
        --tag ponchoai-framework:${VERSION} \
        --tag ponchoai-framework:latest \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        .

    log "Docker image built successfully: ponchoai-framework:${VERSION}"
}

# Run security scan
run_security_scan() {
    log "Running security scan..."

    if command -v trivy &> /dev/null; then
        trivy image --exit-code 1 --severity HIGH,CRITICAL ponchoai-framework:latest
        log "Security scan passed"
    else
        warn "Trivy not found, skipping security scan"
        info "Install Trivy for container security scanning: https://github.com/aquasecurity/trivy"
    fi
}

# Deploy services
deploy_services() {
    log "Deploying services..."

    cd "$PROJECT_ROOT"

    # Pull latest images
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" pull

    # Start services
    docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d

    # Wait for services to be healthy
    wait_for_services

    log "Services deployed successfully"
}

# Wait for services to be healthy
wait_for_services() {
    log "Waiting for services to be healthy..."

    local services=(
        "ponchoai-app:8080"
        "redis:6379"
        "prometheus:9090"
        "grafana:3000"
    )

    local max_attempts=30
    local attempt=1

    while [[ $attempt -le $max_attempts ]]; do
        local all_healthy=true

        for service in "${services[@]}"; do
            local name=$(echo "$service" | cut -d: -f1)
            local port=$(echo "$service" | cut -d: -f2)

            if ! curl -f --max-time 5 "http://localhost:${port}/health" &>/dev/null; then
                warn "Service $name is not healthy yet (attempt $attempt/$max_attempts)"
                all_healthy=false
            fi
        done

        if [[ "$all_healthy" == true ]]; then
            log "All services are healthy"
            return 0
        fi

        sleep 10
        ((attempt++))
    done

    error "Services failed to become healthy within expected time"
}

# Run health checks
run_health_checks() {
    log "Running comprehensive health checks..."

    # Check main application
    if ! curl -f http://localhost:8080/health &>/dev/null; then
        error "Main application health check failed"
    fi

    # Check Prometheus
    if ! curl -f http://localhost:9090/-/healthy &>/dev/null; then
        error "Prometheus health check failed"
    fi

    # Check Grafana
    if ! curl -f http://localhost:3000/api/health &>/dev/null; then
        error "Grafana health check failed"
    fi

    log "All health checks passed"
}

# Display deployment summary
display_summary() {
    log "Deployment completed successfully!"

    echo
    echo "=== Deployment Summary ==="
    echo "Version: $VERSION"
    echo "Build Time: $BUILD_TIME"
    echo "Git Commit: $GIT_COMMIT"
    echo
    echo "=== Service URLs ==="
    echo "Main Application: http://localhost:8080"
    echo "Health Check: http://localhost:8080/health"
    echo "Prometheus: http://localhost:9090"
    echo "Grafana: http://localhost:3000"
    echo "Jaeger: http://localhost:16686"
    echo
    echo "=== Management Commands ==="
    echo "View logs: docker-compose -f $COMPOSE_FILE logs -f"
    echo "Stop services: docker-compose -f $COMPOSE_FILE down"
    echo "Restart services: docker-compose -f $COMPOSE_FILE restart"
    echo
    echo "=== Monitoring ==="
    echo "Default Grafana credentials: admin / $(grep GRAFANA_PASSWORD $ENV_FILE | cut -d= -f2)"
    echo "Prometheus metrics: http://localhost:9090/metrics"
    echo "Jaeger tracing: http://localhost:16686"
}

# Cleanup on error
cleanup() {
    if [[ $? -ne 0 ]]; then
        error "Deployment failed. Check logs for details:"
        echo "docker-compose -f $COMPOSE_FILE logs"
    fi
}

# Main execution
main() {
    log "Starting PonchoAI Framework production deployment..."
    echo "Version: $VERSION"
    echo "Build Time: $BUILD_TIME"
    echo "Git Commit: $GIT_COMMIT"
    echo

    # Set up error handling
    trap cleanup EXIT

    # Run deployment steps
    check_prerequisites
    validate_environment
    build_image
    run_security_scan
    deploy_services
    run_health_checks
    display_summary

    log "Deployment completed successfully! 🚀"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [VERSION]"
        echo "Deploy PonchoAI Framework to production"
        echo
        echo "Arguments:"
        echo "  VERSION    Version tag for the Docker image (default: git describe)"
        echo
        echo "Environment:"
        echo "  .env.production must exist with required variables set"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac