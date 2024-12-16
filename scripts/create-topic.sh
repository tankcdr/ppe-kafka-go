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

# Check if a topic name was provided
if [ -z "$1" ]; then
  echo "Usage: $0 <topic-name>"
  exit 1
fi

TOPIC_NAME=$1

# Execute the create topic command inside the container
kafka-topics  --create --topic ${TOPIC_NAME} --bootstrap-server localhost:${PORT} 
if [ $? -eq 0 ]; then
  echo "Topic '${TOPIC_NAME}' created successfully."
else
  echo "Failed to create topic '${TOPIC_NAME}'."
fi