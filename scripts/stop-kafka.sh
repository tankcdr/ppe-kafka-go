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
if [ -z "$COMPOSE_FILE" ] || [ -z "$PORT" ]; then
  echo "Required environment variables (COMPOSE_FILE, PORT) are not set in .env"
  exit 1
fi

# Change to the root directory containing docker-compose.yml
ROOT_DIR=$(dirname "$(realpath "${BASH_SOURCE[0]}")")/..
cd "$ROOT_DIR" || { echo "Failed to change directory to root"; exit 1; }

# Stop Kafka and Zookeeper using Docker Compose
echo "Stopping Kafka and Zookeeper using Docker Compose..."
docker compose -f $COMPOSE_FILE down

# Check if the services are stopoped
if [ $? -eq 0 ]; then
  echo "✅ Kafka and Zookeeper have been stopped."
else
  echo "❌ Failed to stop Kafka and Zookeeper."
  exit 1
fi