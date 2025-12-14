#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 3 ]]; then
  echo "Usage: $0 <broker_ip> <broker_port> <test_name>"
  echo "Example: $0 203.0.113.10 1883 single-broker-5k-ramp"
  exit 1
fi

BROKER_HOST="$1"
BROKER_PORT="$2"
TEST_NAME="$3"

RESULTS_DIR="$(pwd)/results"
mkdir -p "$RESULTS_DIR"

echo "=============================================="
echo " MQTT Load Test"
echo "----------------------------------------------"
echo " Broker:     ${BROKER_HOST}:${BROKER_PORT}"
echo " Test name:  ${TEST_NAME}"
echo " Started at: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
echo " Results:    ${RESULTS_DIR}"
echo "=============================================="

# Build image if not present / changed
docker compose build loadtest

# Run test
BROKER_HOST="${BROKER_HOST}" \
BROKER_PORT="${BROKER_PORT}" \
TEST_NAME="${TEST_NAME}" \
docker compose up --abort-on-container-exit loadtest

EXIT_CODE=$?

echo "----------------------------------------------"
echo " Finished at: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
echo " Exit code:  ${EXIT_CODE}"
echo "=============================================="

exit "${EXIT_CODE}"
