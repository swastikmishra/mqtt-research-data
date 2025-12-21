# MQTT Mosquitto Load Test Tool (Go)

A Go-based MQTT load-testing tool for benchmarking **single Mosquitto broker** vs **clustered brokers** (with deterministic client-to-broker mapping).  
Designed for research-style experiments where **subscriber count**, **publisher count**, and **payload size** are controlled, and results focus on **delivery correctness** and **latency under load**.

---

### Building the tool

```bash
go mod tidy
go build -o loadtest
```

## What this tool measures (paper-friendly)

### 1) Delivery & Correctness (Data Plane)

Because every subscriber subscribes to **all topics** (by default), each published message is expected to be delivered to **every connected subscriber**.

Per second:

-   **Published (Δpubs_sent)**: number of successful publishes in that second
-   **Received (Δmsgs_recv)**: total messages received by subscribers in that second
-   **Expected deliveries (Δexpected)**:
    -   `Δexpected = Δpubs_sent × connected_subs`
-   **Delivery ratio**:
    -   `delivery_ratio = Δmsgs_recv / Δexpected` (when Δexpected > 0)
-   **Loss / Duplicate / Reorder (estimated)**:
    -   Uses `(pub_id, seq)` embedded in payload to detect:
        -   **lost deliveries** (sequence gaps)
        -   **duplicates**
        -   **reordered deliveries**

> Note: Loss/dup/reorder are tracked from the subscriber’s perspective and are most meaningful when the topic/subscription model stays fixed.

### 2) Latency (End-to-End)

Per message sample:

-   **publish_ts_unix_nano** (written by publisher)
-   **recv_ts_unix_nano** (measured at subscriber)
-   **e2e_latency_ms = (recv - pub) / 1e6**

Reported as:

-   p50 / p95 / p99 (approx bucket-midpoint quantiles)
-   mean / stddev / min / max

### 3) Client Threshold / SLA-style Stop Conditions (Optional)

Runs can stop automatically if quality degrades for consecutive windows:

-   connected subscribers fall below a % of target
-   delivery ratio falls below a % threshold
-   p95 latency exceeds a threshold

---

## Output files

Each run produces:

1. **Traffic & delivery metrics** (per second)

-   `results/<testname>.traffic.csv`

Includes per-second deltas:

-   conn/disconnect deltas (pub + sub split)
-   pubs_sent / msgs_recv / expected deliveries
-   delivery ratio
-   lost/dup/reorder deltas
-   latency p50/p95/p99 (global so far)

2. **Raw latency samples** (one row per received message sample)

-   `results/<testname>.latency.csv`

Columns include:

-   `pub_ts_unix_nano`, `recv_ts_unix_nano`, `pub_id`, `seq`, `latency_ms`, `topic`

3. **Final summary JSON**

-   `results/<testname>.json`

Includes:

-   totals, peaks, overall delivery ratio
-   latency quantiles + mean/std/min/max
-   stop reason and config snapshot

---

## Modes

### A) Single broker mode

Connect all clients to one broker:

-   `--broker-host`
-   `--broker-port`

### B) Cluster mode

Distribute clients across brokers (deterministic):

-   `--brokers-json=./brokers.json`

Client assignment is deterministic by hashing the client ID.

---

## brokers.json format (cluster mode)

Example:

```json
{
    "mode": "cluster",
    "brokers": [
        { "id": "b1", "host": "10.0.0.11", "port": 1883 },
        { "id": "b2", "host": "10.0.0.12", "port": 1883 },
        { "id": "b3", "host": "10.0.0.13", "port": 1883 }
    ]
}
```

## Tests

### Common Setup (Global Defaults)

| Parameter          | Value                   | Flag                            |
| ------------------ | ----------------------- | ------------------------------- |
| Subscriber Step    | 500                     | --sub-step=500                  |
| Initial Subs       | 500                     | --initial-subs=500              |
| Max Subs           | 20000                   | --max-subs=20000                |
| Publisher Strategy | Fixed (50 Pubs)         | --publishers=50                 |
| Publish Rate       | 1 msg/sec               | --pub-rate=1                    |
| Topic Strategy     | 10 Topics (Partitioned) | --topic-count=10                |
| Timings            | 30s Warmup / 60s Window | --warmup-sec=30 --window-sec=60 |

### Test Scenarios

#### Single

| Test Type                | Scenario ID | Payload Size | Goal                                                  | Primary Constraint       | Success Criteria                                |
| ------------------------ | ----------- | ------------ | ----------------------------------------------------- | ------------------------ | ----------------------------------------------- |
| Type A (Max Connections) | SP10        | 10 Bytes     | Determine max connection count before CPU saturation. | CPU / File Descriptors   | Reach >20k subs with P95 Latency < 200ms        |
| Type B (Stability)       | SP100       | 100 Bytes    | Measure latency jitter at medium load.                | Mixed CPU & Network      | Maintain steady P95 < 200ms for full 60s window |
| Type C (Throughput)      | SP1000      | 1 KB         | Determine max network bandwidth (MB/s).               | Network Interface (Mbps) | Delivery Ratio > 99% at max publisher rate      |

#### Cluster

| Test Type                | Scenario ID | Payload Size | Goal                                          | Scaling Strategy                            | Success Criteria                               |
| ------------------------ | ----------- | ------------ | --------------------------------------------- | ------------------------------------------- | ---------------------------------------------- |
| Type A (Max Connections) | CP10        | 10 Bytes     | Exceed single-node connection limit by 2x-3x. | Switch broker when Connection Count > Limit | Total Cluster Subs > (Single Node Max \* 2)    |
| Type B (Stability)       | CP100       | 100 Bytes    | Flatten latency curve during high load.       | Switch broker when P95 Latency > 200ms      | P95 remains < 200ms during the "Switch" event  |
| Type C (Throughput)      | CP1000      | 1 KB         | Maximize aggregate cluster throughput.        | Switch broker when Delivery Ratio < 99%     | Cluster MPS (Msg/Sec) > (Single Node MPS \* 2) |

### SLA Metrics

| Scenario      | Primary Trigger (Stop or Failover) | Flags to Set                                      | Notes                                                           |
| ------------- | ---------------------------------- | ------------------------------------------------- | --------------------------------------------------------------- |
| Single (All)  | Latency > 1000ms OR Error Rate     | --sla-max-p95-ms=1000--sla-consecutive-breaches=2 | Stops test after 3 failing windows.                             |
| Cluster (All) | Latency > 1000ms                   | --sla-max-p95-ms=1000--sla-consecutive-breaches=2 | Action: Triggers ActivateNextBroker().Stop: If no brokers left. |

### Test Commands

#### Single Broker Tests (Baseline)

##### Scenario SP10: Type A (Max Connections)

> Goal: Stress CPU/Connection limits with small packets.

```bash
go run main.go \
 --test-name=SP10_MaxConn \
 --broker-host=43.205.176.30 --broker-port=1883 \
 --cluster-hot-add-new-clients=false \
 --payload-bytes=10 \
 --initial-subs=500 --sub-step=500 --max-subs=20000 \
 --publishers=50 --pub-rate=1 \
 --topic-count=10 \
 --warmup-sec=30 --window-sec=60 \
 --sla-max-p95-ms=1000 --sla-consecutive-breaches=2
```

##### Scenario SP100: Type B (Stability)

> Goal: Measure jitter with medium payloads.

```bash
go run main.go \
 --test-name=SP100_Stability \
 --broker-host=43.205.176.30 --broker-port=1883 \
 --cluster-hot-add-new-clients=false \
 --payload-bytes=100 \
 --initial-subs=500 --sub-step=500 --max-subs=20000 \
 --publishers=50 --pub-rate=1 \
 --topic-count=10 \
 --warmup-sec=30 --window-sec=60 \
 --sla-max-p95-ms=200 --sla-consecutive-breaches=2
```

##### Scenario SP1000: Type C (Throughput)

> Goal: Saturation of Network Bandwidth (1KB payloads).

```bash
go run main.go \
 --test-name=SP1000_Throughput \
--broker-host=43.205.176.30 --broker-port=1883 \
 --cluster-hot-add-new-clients=false \
 --payload-bytes=1000 \
 --initial-subs=500 --sub-step=500 --max-subs=20000 \
 --publishers=50 --pub-rate=1 \
 --topic-count=10 \
 --warmup-sec=30 --window-sec=60 \
 --sla-min-delivery-pct=99.0 --sla-consecutive-breaches=2
```

#### 2. Cluster Tests (Scalability)

##### Scenario CP10: Type A (Max Connections)

> Goal: Validate if cluster scales connections linearly.

```bash
go run main.go \
 --test-name=CP10_ClusterConn \
 --brokers-json=./brokers2.json \
 --cluster-hot-add-new-clients=true \
 --payload-bytes=10 \
 --initial-subs=500 --sub-step=500 --max-subs=40000 \
 --publishers=50 --pub-rate=1 \
 --topic-count=10 \
 --warmup-sec=30 --window-sec=60 \
 --sla-max-p95-ms=1000 --sla-consecutive-breaches=2
```

##### Scenario CP100: Type B (Stability)

> Goal: Validate latency stability during broker "Hot Add" events.

```bash
go run main.go \
 --test-name=CP100_ClusterStability \
 --brokers-json=./brokers2.json \
 --cluster-hot-add-new-clients=true \
 --payload-bytes=100 \
 --initial-subs=500 --sub-step=500 --max-subs=20000 \
 --publishers=50 --pub-rate=1 \
 --topic-count=10 \
 --warmup-sec=30 --window-sec=60 \
 --sla-max-p95-ms=200 --sla-consecutive-breaches=2
```

##### Scenario CP1000: Type C (Throughput)

> Goal: Validate total cluster bandwidth capacity.

```bash
go run main.go \
 --test-name=CP1000_ClusterThroughput \
 --brokers-json=./brokers2.json \
 --cluster-hot-add-new-clients=true \
 --payload-bytes=1000 \
 --initial-subs=500 --sub-step=500 --max-subs=20000 \
 --publishers=50 --pub-rate=1 \
 --topic-count=10 \
 --warmup-sec=30 --window-sec=60 \
 --sla-min-delivery-pct=99.0 --sla-consecutive-breaches=2
```
