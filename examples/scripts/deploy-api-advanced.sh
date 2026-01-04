#!/bin/bash
# deploy-api-advanced.sh - Advanced API deployment with notifications and rollback
#
# This script demonstrates advanced usage of CD-Gun environment variables:
# - Conditional deployment based on changed files
# - Docker image building and pushing
# - Notifications to Slack and PagerDuty
# - Automatic rollback on error
#
# Required environment variables (set by CD-Gun):
#   CDGUN_REPO_NAME    - Repository name (e.g., "api-service")
#   CDGUN_REPO_PATH    - Local path to repository cache
#   CDGUN_REPO_URL     - Repository URL
#   CDGUN_BRANCH       - Branch name
#   CDGUN_CHANGED_FILES - Comma-separated list of changed files
#   CDGUN_OLD_HASH     - Previous commit hash
#   CDGUN_NEW_HASH     - Current commit hash
#
# Custom environment variables (from config.yaml):
#   DOCKER_REGISTRY    - Docker registry (e.g., "docker.company.com")
#   DEPLOY_ENV         - Environment (production, staging, etc.)
#   SLACK_WEBHOOK      - Slack webhook URL for notifications
#   PAGERDUTY_KEY      - PagerDuty integration key
#   ROLLBACK_ON_ERROR  - Enable/disable automatic rollback
#
# See docs/ENVIRONMENT_VARIABLES.md for full documentation

set -euo pipefail

# ==============================================================================
# Configuration
# ==============================================================================

REPO_NAME="${CDGUN_REPO_NAME}"
REPO_PATH="${CDGUN_REPO_PATH}"
BRANCH="${CDGUN_BRANCH}"
NEW_HASH="${CDGUN_NEW_HASH}"
OLD_HASH="${CDGUN_OLD_HASH:-unknown}"
CHANGED_FILES="${CDGUN_CHANGED_FILES}"

REGISTRY="${DOCKER_REGISTRY:-docker.io}"
ENV="${DEPLOY_ENV:-production}"
SLACK_WEBHOOK="${SLACK_WEBHOOK:-}"
PAGERDUTY_KEY="${PAGERDUTY_KEY:-}"
ROLLBACK_ENABLED="${ROLLBACK_ON_ERROR:-false}"

VERSION="${NEW_HASH:0:8}"
IMAGE_NAME="${REGISTRY}/${REPO_NAME}:${VERSION}"
IMAGE_LATEST="${REGISTRY}/${REPO_NAME}:latest"

LOG_FILE="/var/log/cd-gun/${REPO_NAME}-deployment.log"
mkdir -p "$(dirname "$LOG_FILE")"

# ==============================================================================
# Helper Functions
# ==============================================================================

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$LOG_FILE" >&2
}

log_success() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✓ $*" | tee -a "$LOG_FILE"
}

# Send notification to Slack
notify_slack() {
    local message="$1"
    local color="${2:-warning}"
    
    if [ -z "$SLACK_WEBHOOK" ]; then
        return 0
    fi
    
    curl -X POST "$SLACK_WEBHOOK" \
        -H 'Content-Type: application/json' \
        -d @- <<EOF >/dev/null 2>&1 || true
{
    "text": "$message",
    "attachments": [{
        "color": "$color",
        "fields": [
            {"title": "Repository", "value": "$REPO_NAME", "short": true},
            {"title": "Environment", "value": "$ENV", "short": true},
            {"title": "Branch", "value": "$BRANCH", "short": true},
            {"title": "Hash", "value": "$VERSION", "short": true},
            {"title": "Old Hash", "value": "${OLD_HASH:0:8}", "short": true},
            {"title": "Host", "value": "$(hostname)", "short": true}
        ]
    }]
}
EOF
}

# Send incident to PagerDuty
notify_pagerduty() {
    local severity="$1"
    local message="$2"
    
    if [ -z "$PAGERDUTY_KEY" ]; then
        return 0
    fi
    
    curl -X POST "https://events.pagerduty.com/v2/enqueue" \
        -H 'Content-Type: application/json' \
        -d @- <<EOF >/dev/null 2>&1 || true
{
    "routing_key": "$PAGERDUTY_KEY",
    "event_action": "trigger",
    "dedup_key": "${REPO_NAME}-${VERSION}",
    "payload": {
        "summary": "$message",
        "severity": "$severity",
        "source": "cd-gun",
        "custom_details": {
            "repository": "$REPO_NAME",
            "environment": "$ENV",
            "hash": "$VERSION"
        }
    }
}
EOF
}

# Cleanup and rollback on error
cleanup() {
    local exit_code=$?
    
    if [ $exit_code -ne 0 ]; then
        log_error "Deployment failed with exit code $exit_code"
        notify_slack "❌ Deployment FAILED: $REPO_NAME" "danger"
        notify_pagerduty "critical" "Deployment failed for $REPO_NAME on $ENV"
        
        if [ "$ROLLBACK_ENABLED" = "true" ]; then
            log "Attempting rollback to previous version..."
            if rollback_deployment; then
                log_success "Rollback completed successfully"
                notify_slack "🔄 Rollback completed for $REPO_NAME" "warning"
            else
                log_error "Rollback failed!"
                notify_slack "⚠️ Rollback FAILED for $REPO_NAME - MANUAL INTERVENTION REQUIRED" "danger"
            fi
        fi
        
        exit $exit_code
    fi
}

trap cleanup EXIT

# ==============================================================================
# Deployment Functions
# ==============================================================================

check_prerequisites() {
    log "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

should_rebuild_image() {
    # Rebuild if Dockerfile or dependencies changed
    if echo "$CHANGED_FILES" | grep -E "^(Dockerfile|go\.mod|go\.sum|requirements\.txt|package\.json)"; then
        return 0
    fi
    return 1
}

build_and_push_image() {
    log "Building Docker image: $IMAGE_NAME"
    
    cd "$REPO_PATH"
    git checkout "$BRANCH"
    git reset --hard "$NEW_HASH"
    
    docker build \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
        --build-arg VCS_REF="$NEW_HASH" \
        -t "$IMAGE_NAME" \
        -t "$IMAGE_LATEST" \
        .
    
    log_success "Image built successfully"
    
    log "Pushing image to registry: $REGISTRY"
    docker push "$IMAGE_NAME"
    docker push "$IMAGE_LATEST"
    
    log_success "Image pushed to registry"
}

deploy_to_environment() {
    log "Deploying to environment: $ENV"
    
    cd "$REPO_PATH"
    
    # Update docker-compose with new image version
    sed -i.bak \
        "s|image: .*|image: $IMAGE_NAME|g" \
        docker-compose.yml
    
    # Stop old containers
    docker-compose down || true
    
    # Start new containers
    docker-compose up -d
    
    # Wait for service to be healthy
    log "Waiting for service to become healthy..."
    for i in {1..30}; do
        if curl -f http://localhost:8080/health > /dev/null 2>&1; then
            log_success "Service is healthy"
            return 0
        fi
        log "Health check attempt $i/30..."
        sleep 2
    done
    
    log_error "Service did not become healthy"
    return 1
}

run_smoke_tests() {
    log "Running smoke tests..."
    
    cd "$REPO_PATH"
    
    # Run basic API tests
    if command -v pytest &> /dev/null; then
        pytest tests/smoke/ -v || return 1
    fi
    
    log_success "Smoke tests passed"
}

verify_deployment() {
    log "Verifying deployment..."
    
    # Check if new version is running
    running_version=$(docker-compose exec -T api cat /app/VERSION 2>/dev/null || echo "unknown")
    
    if [ "$running_version" = "$VERSION" ]; then
        log_success "Correct version is running"
        return 0
    else
        log_error "Expected version $VERSION, but running $running_version"
        return 1
    fi
}

rollback_deployment() {
    log "Rolling back to previous version..."
    
    cd "$REPO_PATH"
    
    # Get previous version from state
    PREV_IMAGE="${REGISTRY}/${REPO_NAME}:${OLD_HASH:0:8}"
    
    log "Starting containers with previous image: $PREV_IMAGE"
    
    sed -i \
        "s|image: .*|image: $PREV_IMAGE|g" \
        docker-compose.yml
    
    docker-compose down || true
    docker-compose up -d || return 1
    
    # Wait for service
    for i in {1..30}; do
        if curl -f http://localhost:8080/health > /dev/null 2>&1; then
            log_success "Rollback completed"
            return 0
        fi
        sleep 2
    done
    
    return 1
}

# ==============================================================================
# Main Deployment Flow
# ==============================================================================

main() {
    log "Starting deployment..."
    log "Repository: $REPO_NAME"
    log "Branch: $BRANCH"
    log "Old hash: ${OLD_HASH:0:8}"
    log "New hash: $NEW_HASH"
    log "Environment: $ENV"
    log "Changed files: $CHANGED_FILES"
    
    notify_slack "🚀 Deployment started for $REPO_NAME on $ENV" "warning"
    
    check_prerequisites
    
    if should_rebuild_image; then
        log "Changes detected, rebuilding Docker image"
        build_and_push_image
    else
        log "No code changes detected, skipping image build"
    fi
    
    deploy_to_environment
    
    run_smoke_tests
    
    verify_deployment
    
    log_success "Deployment completed successfully!"
    notify_slack "✅ Deployment successful: $REPO_NAME on $ENV" "good"
    
    log "Deployment summary:"
    log "  Repository: $REPO_NAME"
    log "  Environment: $ENV"
    log "  Version: $VERSION"
    log "  Duration: $(date -u +%s)s"
}

main
