#!/bin/bash
# deploy-api.sh - Example API service deployment script
#
# Available environment variables:
#   CDGUN_REPO_NAME    - Repository name
#   CDGUN_REPO_PATH    - Local path to repository cache
#   CDGUN_BRANCH       - Branch name being monitored
#   CDGUN_CHANGED_FILES - Comma-separated list of changed files
#   CDGUN_NEW_HASH     - Current commit hash
#
# Custom variables (if configured in config.yaml):
#   DOCKER_REGISTRY    - Docker registry address
#   DEPLOY_ENV        - Deployment environment
#
# For more details, see docs/ENVIRONMENT_VARIABLES.md

set -e

echo "[$(date)] Starting API service deployment..."
echo "Repository: $CDGUN_REPO_NAME"
echo "Branch: $CDGUN_BRANCH"
echo "Hash: $CDGUN_NEW_HASH"

# Navigate to the repository
cd "$CDGUN_REPO_PATH"
git checkout "$CDGUN_BRANCH"
git reset --hard "$CDGUN_NEW_HASH"

# Build Docker image
echo "[$(date)] Building Docker image..."
docker build -t myapi:$CDGUN_NEW_HASH .

# Stop old container
echo "[$(date)] Stopping old container..."
docker-compose down || true

# Start new container
echo "[$(date)] Starting new container..."
docker-compose up -d

# Verify deployment
echo "[$(date)] Verifying deployment..."
sleep 5
if curl -f http://localhost:8080/health > /dev/null; then
    echo "[$(date)] Health check passed!"
else
    echo "[$(date)] Health check failed!"
    docker-compose logs
    exit 1
fi

echo "[$(date)] Deployment completed successfully!"
