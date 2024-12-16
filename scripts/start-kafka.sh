#!/bin/bash

# Get the directory of the current script
SCRIPT_DIR=$(dirname "$(realpath "${BASH_SOURCE[0]}")")

# Load environment variables from .env
if [ -f "${SCRIPT_DIR}/.env" ]; then
  export $(grep -v '^#' "${SCRIPT_DIR}/.env" | xargs)
else
  echo ".env file not found!"
  exit 1
fi

# Validate required variables
if [ -z "$TAG" ] || [ -z "$IMAGE" ] || [ -z "$PORT" ]; then
  echo "Required environment variables (TAG, IMAGE, PORT) are not set in .env"
  exit 1
fi

# Check if the container already exists

EXISTING_CONTAINER=$(docker ps -a --filter "name=${CONTAINER_NAME}" --format "{{.ID}}")

if [ -n "$EXISTING_CONTAINER" ]; then
  # Start the existing container if it's stopped
  echo "Starting existing Kafka container (${CONTAINER_NAME})..."
  docker start "$CONTAINER_NAME"
else
  # Run a new container and name it
  echo "Creating and starting a new Kafka container (${CONTAINER_NAME})..."
  docker run -d \
      --name "$CONTAINER_NAME" \
      -p "${PORT}:${PORT}" \
      "${IMAGE}:${TAG}"
fi

# Start Schema Registry container
EXISTING_SCHEMA_REGISTRY_CONTAINER=$(docker ps -a --filter "name=${SCHEMA_REGISTRY_CONTAINER_NAME}" --format "{{.ID}}")
if [ -n "$EXISTING_SCHEMA_REGISTRY_CONTAINER" ]; then
  echo "Starting existing Schema Registry container (${SCHEMA_REGISTRY_CONTAINER_NAME})..."
  docker start "$SCHEMA_REGISTRY_CONTAINER_NAME"
else
  echo "Creating and starting a new Schema Registry container (${SCHEMA_REGISTRY_CONTAINER_NAME})..."
  docker run -d \
      --name "$SCHEMA_REGISTRY_CONTAINER_NAME" \
      -p 8081:8081 \
      -e SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS=PLAINTEXT://localhost:${PORT} \
      -e SCHEMA_REGISTRY_HOST_NAME=schema-registry \
      confluentinc/cp-schema-registry:latest
fi

echo "Kafka and Schema Registry are now running."