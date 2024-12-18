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

# Usage function
function usage() {
  echo "Usage: $0 <topic-name> [--message \"<message>\" | --file <file-path>]"
  echo "Options:"
  echo "  --message   Specify the message as a string."
  echo "  --file      Specify a file containing the message contents."
  echo "Example:"
  echo "  $0 my-topic --message \"Hello, Kafka!\""
  echo "  $0 my-topic --file message.txt"
  exit 1
}

# Check if a topic name was provided
if [ -z "$1" ]; then
  usage
fi

TOPIC_NAME=$1
MESSAGE=""
FILE_PATH=""

# Parse the message or file argument
if [[ "$2" == "--message" && -n "$3" ]]; then
  MESSAGE="$3"
elif [[ "$2" == "--file" && -n "$3" ]]; then
  FILE_PATH="$3"
  if [ ! -f "$FILE_PATH" ]; then
    echo "Error: File '$FILE_PATH' not found."
    exit 1
  fi
else
  echo "Error: You must provide either a message or a file."
  usage
fi

# Determine message source
if [ -n "$MESSAGE" ]; then
  # Send the message provided as a string
  echo "Publishing message to topic '${TOPIC_NAME}'..."
  kafka-console-producer --bootstrap-server localhost:${PORT} --topic "${TOPIC_NAME}" <<< "${MESSAGE}"
elif [ -n "$FILE_PATH" ]; then
  # Send the contents of the file as a message
  echo "Publishing contents of file '${FILE_PATH}' to topic '${TOPIC_NAME}'..."
  kafka-console-producer --bootstrap-server localhost:${PORT} --topic "${TOPIC_NAME}" <<< "$(tr '\n' ' ' < "${FILE_PATH}")"
fi

# Check if the command succeeded
if [ $? -eq 0 ]; then
  if [ -n "$MESSAGE" ]; then
    echo "✅ Message published to topic '${TOPIC_NAME}': ${MESSAGE}"
  else
    echo "✅ File contents published to topic '${TOPIC_NAME}' from '${FILE_PATH}'."
  fi
else
  echo "❌ Failed to publish to topic '${TOPIC_NAME}'."
  exit 1
fi