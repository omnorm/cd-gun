#!/bin/bash
# deploy-web.sh - Example web application deployment script
# This script is called by cd-gun when changes are detected
#
# Available environment variables:
#   CDGUN_REPO_NAME    - Repository name
#   CDGUN_REPO_URL     - Repository URL
#   CDGUN_REPO_PATH    - Local path to repository cache
#   CDGUN_BRANCH       - Branch name being monitored
#   CDGUN_CHANGED_FILES - Comma-separated list of changed files
#   CDGUN_OLD_HASH     - Previous commit hash
#   CDGUN_NEW_HASH     - Current commit hash
#
# For more details, see docs/ENVIRONMENT_VARIABLES.md

set -e

echo "[$(date)] Starting web app deployment..."
echo "Repository: $CDGUN_REPO_NAME"
echo "Branch: $CDGUN_BRANCH"
echo "Changed files: $CDGUN_CHANGED_FILES"
echo "New hash: $CDGUN_NEW_HASH"

# Navigate to the repository
cd "$CDGUN_REPO_PATH"
git checkout "$CDGUN_BRANCH"
git reset --hard "$CDGUN_NEW_HASH"

# Install dependencies
echo "[$(date)] Installing dependencies..."
npm ci

# Build the application
echo "[$(date)] Building application..."
npm run build

# Deploy
echo "[$(date)] Deploying to production..."
DEPLOY_DIR="/var/www/myapp"
sudo cp -r dist/* "$DEPLOY_DIR/"

# Reload web server
echo "[$(date)] Reloading nginx..."
sudo systemctl reload nginx

echo "[$(date)] Deployment completed successfully!"
