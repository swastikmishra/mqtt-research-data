#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 4 ]]; then
  echo "Usage: $0 <brokers_json> <payload_kb> <test_name> <max_duration_sec=300>"
  echo "Example:"
  echo "  $0 brokers.json 10 clustered-10kb 300"
  echo "Default max_duration_sec is 300 seconds"
  exit 1
fi

BROKERS_JSON="$1"
PAYLOAD_KB="$2"
TEST_NAME="$3"
MAX_DURATION_SEC="$4"

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
echo " Max duration: ${MAX_DURATION_SEC} sec"
echo "=============================================="

# Build once
go mod tidy
go build -o loadtest

# Run test
./loadtest \
  --brokers-json "${BROKERS_JSON}" \
  --payload-kb "${PAYLOAD_KB}" \
  --test-name "${TEST_NAME}" \
  --max-duration-sec "${MAX_DURATION_SEC:-300}"
