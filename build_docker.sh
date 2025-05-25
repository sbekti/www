#!/bin/bash

# Docker build script that automatically sets the build date
# Usage: ./build_docker.sh [image_name] [tag]

set -e  # Exit on any error

# Default values
IMAGE_NAME=${1:-"www"}
TAG=${2:-"latest"}
WORKERS=${3:-4}

# Get current date in UTC
BUILD_DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

echo "Building Docker image with the following parameters:"
echo "  Image name: $IMAGE_NAME"
echo "  Tag: $TAG"
echo "  Build date: $BUILD_DATE"
echo "  Gunicorn workers: $WORKERS"
echo ""

# Build the Docker image with build arguments
docker build \
    --build-arg BUILD_DATE="$BUILD_DATE" \
    --build-arg GUNICORN_WORKERS="$WORKERS" \
    -t "$IMAGE_NAME:$TAG" \
    .

echo ""
echo "✅ Docker image built successfully: $IMAGE_NAME:$TAG"
echo "   Build date: $BUILD_DATE"
echo ""
echo "To run the container:"
echo "  docker run -p 8000:8000 $IMAGE_NAME:$TAG"
echo ""
echo "To run with custom environment variables:"
echo "  docker run -p 8000:8000 \\"
echo "    -e DB_USERNAME=your_user \\"
echo "    -e DB_PASSWORD=your_password \\"
echo "    -e DB_HOST=your_host \\"
echo "    -e DB_PORT=5432 \\"
echo "    -e DB_NAME=your_db \\"
echo "    $IMAGE_NAME:$TAG" 