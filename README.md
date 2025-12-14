# MQTT Load Testing Tool (QoS 0)

This repository contains a **Go-based MQTT load testing tool** designed
to evaluate the scalability of Mosquitto brokers in **single-broker**
and **clustered** deployments.

The tool acts as a **smart client** that concurrently spawns publishers
and subscribers, ramps client counts over time, and records detailed
performance metrics for research and benchmarking purposes.

------------------------------------------------------------------------

## Features

-   MQTT **QoS 0** testing (no TLS)
-   Concurrent **publishers and subscribers**
-   **Ramped load** (adds clients every fixed interval)
-   Measures:
    -   successful vs failed connections
    -   message throughput
    -   end-to-end latency (p50/p95/p99)
-   Outputs results in **JSON (summary)** and **CSV (time series)**
    formats
-   Dockerized for **reproducible experiments**

------------------------------------------------------------------------

## Prerequisites

-   Docker ≥ 20.x
-   Docker Compose ≥ 2.x
-   Network access to the target MQTT broker

------------------------------------------------------------------------

## Directory Structure

    .
    ├── Dockerfile
    ├── docker-compose.yml
    ├── test.sh
    ├── results/
    │   ├── <test-name>.json
    │   └── <test-name>.csv
    ├── cmd/
    │   └── loadtest/
    │       └── main.go
    └── README.md

------------------------------------------------------------------------

## Running a Test

### 1. Build the load testing container

``` bash
docker compose build
```

### 2. Run a test (Single)

``` bash
./test.sh <broker_ip> <broker_port> <payload_kb> <test_name>
```

Example:

``` bash
./test.sh 13.201.32.55 1883 10 single-10kb
```

All clients will connect directly to the specified broker.

### 3. Run a test (Clustered)


``` bash
./test.sh <brokers_json_file> <payload_kb> <test_name>
```

Example:

``` bash
./test-clustered.sh brokers.json 10 cluster-10kb
```


------------------------------------------------------------------------

## Output Artifacts

After the test completes, results are written to the `results/`
directory.

### JSON Summary

    results/<test-name>.json

Contains: - test configuration - peak stable client count - failure
reason (if any) - latency percentiles - aggregate throughput

### CSV Time Series

    results/<test-name>.csv

Contains one row per measurement interval (e.g., per second or ramp
step), including: - connected clients - connection success/failure
counts - messages sent/received - delivery rate - latency percentiles

These files are suitable for direct import into plotting tools or
statistical analysis frameworks.

------------------------------------------------------------------------

## Experiment Methodology

Typical experiment flow: 1. Start with a low number of clients 2. Ramp
clients every fixed interval (e.g., every 10 seconds) 3. Maintain load
briefly at each step 4. Stop when stability criteria are violated (e.g.,
connection failures or excessive latency)

This enables precise measurement of: - maximum stable connections -
throughput under fan-out - latency degradation under load

------------------------------------------------------------------------

## Intended Use

This tool is designed for: - evaluating Mosquitto scalability limits -
comparing single-broker vs clustered architectures - academic and
industrial research on MQTT performance

It is **not** intended as a production monitoring or benchmarking
service.

------------------------------------------------------------------------

## License

MIT (or specify your preferred license)
