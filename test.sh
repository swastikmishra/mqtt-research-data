#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 4 ]]; then
  echo "Usage: $0 <broker_ip> <broker_port> <payload_sizes_kb> <test_name>"
  echo "Examples:"
  echo "  $0 203.0.113.10 1883 10 single-broker-5k-ramp"
  echo "  $0 203.0.113.10 1883 10,100,500 single-broker-5k-ramp"
  echo "  $0 203.0.113.10 1883 \"10 100 500\" single-broker-5k-ramp"
  exit 1
fi

BROKER_HOST="$1"
BROKER_PORT="$2"
PAYLOAD_SIZES_RAW="$3"
BASE_TEST_NAME="$4"

RESULTS_DIR="$(pwd)/results"
mkdir -p "$RESULTS_DIR"

# Normalize payload sizes:
# - convert commas to spaces
# - collapse multiple spaces
PAYLOAD_SIZES="$(echo "$PAYLOAD_SIZES_RAW" | tr ',' ' ' | xargs)"

if [[ -z "$PAYLOAD_SIZES" ]]; then
  echo "Error: payload_sizes_kb is empty"
  exit 1
fi

echo "=============================================="
echo " MQTT Load Test Batch"
echo "----------------------------------------------"
echo " Broker:        ${BROKER_HOST}:${BROKER_PORT}"
echo " Base test:     ${BASE_TEST_NAME}"
echo " Payload sizes: ${PAYLOAD_SIZES} (KB)"
echo " Started at:    $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
echo " Results:       ${RESULTS_DIR}"
echo "=============================================="

# Build image once
docker compose build loadtest

EXIT_CODE=0

for PAYLOAD_KB in $PAYLOAD_SIZES; do
  if ! [[ "$PAYLOAD_KB" =~ ^[0-9]+$ ]]; then
    echo "Skipping invalid payload size: '$PAYLOAD_KB' (must be an integer KB)"
    continue
  fi

  TEST_NAME="${BASE_TEST_NAME}__payload${PAYLOAD_KB}kb"

  echo "----------------------------------------------"
  echo " Running: ${TEST_NAME}"
  echo " Payload: ${PAYLOAD_KB} KB"
  echo " Started: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo "----------------------------------------------"

  # Run a single test (use docker compose run so it exits cleanly per payload)
  BROKER_HOST="${BROKER_HOST}" \
  BROKER_PORT="${BROKER_PORT}" \
  TEST_NAME="${TEST_NAME}" \
  docker compose run --rm loadtest \
    --broker-host "${BROKER_HOST}" \
    --broker-port "${BROKER_PORT}" \
    --test-name "${TEST_NAME}" \
    --payload-kb "${PAYLOAD_KB}" \
    --out-dir /app/results || EXIT_CODE=$?

  echo " Finished: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo " Outputs:"
  echo "   ${RESULTS_DIR}/${TEST_NAME}.csv"
  echo "   ${RESULTS_DIR}/${TEST_NAME}.json"
done

echo "=============================================="
echo " Batch finished at: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
echo " Exit code:        ${EXIT_CODE}"
echo "=============================================="

exit "${EXIT_CODE}"
