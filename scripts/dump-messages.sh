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
MAX_MESSAGES=${2:-10}
TIMEOUT=${3:-5000}

# Dump all messages from the beginning
  kafka-console-consumer --bootstrap-server localhost:${PORT} \
  --topic "${TOPIC_NAME}" --from-beginning --timeout-ms ${TIMEOUT} --max-messages ${MAX_MESSAGES} --partition 0