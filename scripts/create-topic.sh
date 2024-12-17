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

# Ensure PORT is set
if [ -z "${PORT}" ]; then
  echo "Error: PORT is not set in the .env file."
  exit 1
fi

# Check if a topic name was provided
if [ -z "$1" ]; then
  echo "Usage: $0 <topic-name> [retention-days]"
  echo "Example: $0 my-topic 3"
  exit 1
fi

TOPIC_NAME=$1
RETENTION_DAYS=${2:-7}  # Default retention is 7 days if not specified

# Convert retention time from days to milliseconds
RETENTION_MS=$((RETENTION_DAYS * 24 * 60 * 60 * 1000))

# Create the Kafka topic with retention.ms
echo "Creating Kafka topic '${TOPIC_NAME}' with a retention time of ${RETENTION_DAYS} day(s) (${RETENTION_MS} ms)..."

kafka-topics --create \
  --topic "${TOPIC_NAME}" \
  --bootstrap-server "localhost:${PORT}" \
  --config retention.ms="${RETENTION_MS}"

# Check if the topic was created successfully
if [ $? -eq 0 ]; then
  echo "✅ Topic '${TOPIC_NAME}' created successfully with retention time ${RETENTION_DAYS} day(s)."
else
  echo "❌ Failed to create topic '${TOPIC_NAME}'. Check Kafka logs for details."
fi