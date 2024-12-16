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

# Check if the container exists
EXISTING_CONTAINER=$(docker ps -a --filter "name=${CONTAINER_NAME}" --format "{{.ID}}")

if [ -n "$EXISTING_CONTAINER" ]; then
  echo "Stopping Kafka container (${CONTAINER_NAME})..."
  docker stop "$CONTAINER_NAME"
else
  echo "Kafka container (${CONTAINER_NAME}) does not exist or is already stopped."
fi