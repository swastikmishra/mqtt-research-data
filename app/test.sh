#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 4 ]]; then
  echo "Usage: $0 <broker_host> <broker_port> <payload_kb> <test_name>"
  echo "Example:"
  echo "  $0 203.0.113.10 1883 10 baseline-10kb"
  exit 1
fi

BROKER_HOST="$1"
BROKER_PORT="$2"
PAYLOAD_KB="$3"
TEST_NAME="$4"

# Optional sanity checks
if ! [[ "$BROKER_PORT" =~ ^[0-9]+$ ]]; then
  echo "Error: broker_port must be a number"
  exit 1
fi

if ! [[ "$PAYLOAD_KB" =~ ^[0-9]+$ ]]; then
  echo "Error: payload_kb must be a number"
  exit 1
fi

echo "=============================================="
echo " Running MQTT Load Test"
echo "----------------------------------------------"
echo " Broker:     ${BROKER_HOST}:${BROKER_PORT}"
echo " Payload:    ${PAYLOAD_KB} KB"
echo " Test name:  ${TEST_NAME}"
echo "=============================================="

# Build once
go mod tidy
go build -o loadtest

# Run test
./loadtest \
  --broker-host "${BROKER_HOST}" \
  --broker-port "${BROKER_PORT}" \
  --payload-kb "${PAYLOAD_KB}" \
  --test-name "${TEST_NAME}"

