#!/bin/bash

SCRIPT_DIR=$(dirname "$(realpath "${BASH_SOURCE[0]}")")
$SCRIPT_DIR/create-topic.sh order-received 3
$SCRIPT_DIR/create-topic.sh order-confirmed 3
$SCRIPT_DIR/create-topic.sh order-picked-packed 3
$SCRIPT_DIR/create-topic.sh order-notification 3
$SCRIPT_DIR/create-topic.sh order-error 3
