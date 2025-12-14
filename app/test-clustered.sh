#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "Usage: $0 <brokers_json> <payload_kb> <test_name>"
  echo "Example:"
  echo "  $0 brokers.json 10 clustered-10kb"
  exit 1
fi

BROKERS_JSON="$1"
PAYLOAD_KB="$2"
TEST_NAME="$3"

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
echo " Brokers:     ${BROKERS_JSON}"
echo " Payload:    ${PAYLOAD_KB} KB"
echo " Test name:  ${TEST_NAME}"
echo "=============================================="

# Build once
go mod tidy
go build -o loadtest

# Run test
./loadtest \
  --brokers-json "${BROKERS_JSON}" \
  --payload-kb "${PAYLOAD_KB}" \
  --test-name "${TEST_NAME}"
