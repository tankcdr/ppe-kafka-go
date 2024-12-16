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
  echo "Usage: $0 <topic-name> <message>"
  exit 1
fi

# Check if a message was provided
if [ -z "$2" ]; then
  echo "Error: No message provided."
  echo "Usage: $0 <topic-name> <message>"
  exit 1
fi


TOPIC_NAME=$1
MESSAGE=$2

# Publish the message
kafka-console-producer --bootstrap-server localhost:${PORT} --topic "${TOPIC_NAME}" <<< "${MESSAGE}"

if [ $? -eq 0 ]; then
  echo "Message published to topic '${TOPIC_NAME}': ${MESSAGE}"
else
  echo "Failed to publish message to topic '${TOPIC_NAME}'."
fi