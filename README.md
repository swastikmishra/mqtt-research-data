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

### Common Setup

| Subscriber Step    | 100   |
| ------------------ | ----- |
| Initital Subs      | 500   |
| Max Subs           | 20000 |
| Publisher Strategy | Fixed |
| Publishers         | 50    |
| Publish Rate / Pub | 1     |
| QoS                | 0     |
| Topic Count        | 10    |
| Warmup (s)         | 30    |
| Window (s)         | 60    |

### Test Scenarios

| Scenario ID | Scenario Name                                   | Broker Mode | Broker Scale Strategy     | Initial Brokers | Max Brokers | Payload (Bytes) |
| ----------- | ----------------------------------------------- | ----------- | ------------------------- | --------------- | ----------- | --------------- |
| S1-P10      | Single Broker – Sub Scaling (10B)               | Single      | None                      | 1               | 1           | 10              |
| S1-P100     | Single Broker – Sub Scaling (100B)              | Single      | None                      | 1               | 1           | 100             |
| S1-P1000    | Single Broker – Sub Scaling (1000B)             | Single      | None                      | 1               | 1           | 1000            |
| S2-P10      | Clustered – Fixed Brokers – Sub Scaling (10B)   | Cluster     | Fixed broker set          | 3               | 3           | 10              |
| S2-P100     | Clustered – Fixed Brokers – Sub Scaling (100B)  | Cluster     | Fixed broker set          | 3               | 3           | 100             |
| S2-P1000    | Clustered – Fixed Brokers – Sub Scaling (1000B) | Cluster     | Fixed broker set          | 3               | 3           | 1000            |
| S3-P10      | Clustered – Incremental Broker Scale (10B)      | Cluster     | Add broker on SLA failure | 1               | 3           | 10              |
| S3-P100     | Clustered – Incremental Broker Scale (100B)     | Cluster     | Add broker on SLA failure | 1               | 3           | 100             |
| S3-P1000    | Clustered – Incremental Broker Scale (1000B)    | Cluster     | Add broker on SLA failure | 1               | 3           | 1000            |

### SLA Metrics

| Scenario ID | Primary Stop SLA             | Secondary Stop SLAs                                     | Notes                |
| ----------- | ---------------------------- | ------------------------------------------------------- | -------------------- |
| S1-P10      | p95 latency > SLA            | delivery <99%, connected subs <99%, disconnects/min > X | Baseline capacity    |
| S1-P100     | p95 latency > SLA            | delivery <99%, connected subs <99%, disconnects/min > X | Payload impact       |
| S1-P1000    | p95 latency > SLA            | delivery <99%, connected subs <99%, disconnects/min > X | Large payload        |
| S2-P10      | p95 latency > SLA            | delivery <99%, connected subs <99%, disconnects/min > X | Distribution effects |
| S2-P100     | p95 latency > SLA            | delivery <99%, connected subs <99%, disconnects/min > X | Scaling vs payload   |
| S2-P1000    | p95 latency > SLA            | delivery <99%, connected subs <99%, disconnects/min > X | Upper bound          |
| S3-P10      | SLA fail w/ all brokers used | same as above                                           | Staircase scaling    |
| S3-P100     | SLA fail w/ all brokers used | same as above                                           | Capacity gains       |
| S3-P1000    | SLA fail w/ all brokers used | same as above                                           | Diminishing returns  |

### Test Commands:

#### S1 — Single Broker: Subscriber Scaling

##### S1-P10

```bash
go run main.go \
  --broker-host=43.205.176.30 --broker-port=1883 \
  --out-dir=./results --test-name=S1-P10_single_subscale_10B \
  --initial-subs=500 --sub-step=100 --max-subs=20000 \
  --publishers=50 --pub-rate=1 \
  --payload-bytes=10 --topic-count=10 --topic-prefix=bench/topic \
  --window-sec=60 --warmup-sec=10 \
  --sla-min-connected-sub-pct=99 --sla-min-delivery-pct=99 --sla-max-p95-ms=200 --sla-max-disc-per-min=50 \
  --sla-consecutive-breaches=2
```

##### S1-P100

```bash
go run main.go \
  --broker-host=43.205.176.30 --broker-port=1883 \
  --out-dir=./results --test-name=S1-P100_single_subscale_100B \
  --initial-subs=500 --sub-step=100 --max-subs=20000 \
  --publishers=50 --pub-rate=1 \
  --payload-bytes=100 --topic-count=10 --topic-prefix=bench/topic \
  --window-sec=60 --warmup-sec=10 \
  --sla-min-connected-sub-pct=99 --sla-min-delivery-pct=99 --sla-max-p95-ms=200 --sla-max-disc-per-min=50 \
  --sla-consecutive-breaches=2
```

##### S1-P1000

```bash
go run main.go \
  --broker-host=43.205.176.30 --broker-port=1883 \
  --out-dir=./results --test-name=S1-P1000_single_subscale_1000B \
  --initial-subs=500 --sub-step=100 --max-subs=20000 \
  --publishers=50 --pub-rate=1 \
  --payload-bytes=1000 --topic-count=10 --topic-prefix=bench/topic \
  --window-sec=60 --warmup-sec=10 \
  --sla-min-connected-sub-pct=99 --sla-min-delivery-pct=99 --sla-max-p95-ms=200 --sla-max-disc-per-min=50 \
  --sla-consecutive-breaches=2
```

#### S2 — Cluster (Fixed brokers.json): Subscriber Scaling

##### S2-P10

```bash
go run main.go \
  --brokers-json=./brokers.json --cluster-incremental=false \
  --out-dir=./results --test-name=S2-P10_cluster_fixed_subscale_10B \
  --initial-subs=500 --sub-step=100 --max-subs=20000 \
  --publishers=50 --pub-rate=1 \
  --payload-bytes=10 --topic-count=10 --topic-prefix=bench/topic \
  --window-sec=60 --warmup-sec=10 \
  --sla-min-connected-sub-pct=99 --sla-min-delivery-pct=99 --sla-max-p95-ms=200 --sla-max-disc-per-min=50 \
  --sla-consecutive-breaches=2
```

##### S2-P100

```bash
go run main.go \
  --brokers-json=./brokers.json --cluster-incremental=false \
  --out-dir=./results --test-name=S2-P100_cluster_fixed_subscale_100B \
  --initial-subs=500 --sub-step=100 --max-subs=20000 \
  --publishers=50 --pub-rate=1 \
  --payload-bytes=100 --topic-count=10 --topic-prefix=bench/topic \
  --window-sec=60 --warmup-sec=10 \
  --sla-min-connected-sub-pct=99 --sla-min-delivery-pct=99 --sla-max-p95-ms=200 --sla-max-disc-per-min=50 \
  --sla-consecutive-breaches=2
```

##### S2-P1000

```bash
go run main.go \
  --brokers-json=./brokers.json --cluster-incremental=false \
  --out-dir=./results --test-name=S2-P1000_cluster_fixed_subscale_1000B \
  --initial-subs=500 --sub-step=100 --max-subs=20000 \
  --publishers=50 --pub-rate=1 \
  --payload-bytes=1000 --topic-count=10 --topic-prefix=bench/topic \
  --window-sec=60 --warmup-sec=10 \
  --sla-min-connected-sub-pct=99 --sla-min-delivery-pct=99 --sla-max-p95-ms=200 --sla-max-disc-per-min=50 \
  --sla-consecutive-breaches=2
```

#### S3 — Cluster Incremental: Add broker when SLA fails

##### S3-P10

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --cluster-hot-add-new-clients=true \
  --out-dir=./results --test-name=S3-P10_cluster_hotadd_new_clients_only_10B \
  --initial-subs=500 --sub-step=100 --max-subs=20000 \
  --publishers=50 --pub-rate=1 \
  --payload-bytes=10 --topic-count=10 --topic-prefix=bench/topic \
  --window-sec=60 --warmup-sec=30 \
  --sla-min-connected-sub-pct=99 --sla-min-delivery-pct=99 --sla-max-p95-ms=200 --sla-max-disc-per-min=50 \
  --sla-consecutive-breaches=2
```

##### S3-P100

```bash
go run mqtt_scale_tester.go \
  --brokers-json=./brokers.json \
  --cluster-hot-add-new-clients=true \
  --out-dir=./results --test-name=S3-P100_cluster_hotadd_new_clients_only_100B \
  --initial-subs=500 --sub-step=100 --max-subs=20000 \
  --publishers=50 --pub-rate=1 \
  --payload-bytes=100 --topic-count=10 --topic-prefix=bench/topic \
  --window-sec=60 --warmup-sec=30 \
  --sla-min-connected-sub-pct=99 --sla-min-delivery-pct=99 --sla-max-p95-ms=200 --sla-max-disc-per-min=50 \
  --sla-consecutive-breaches=2
```

##### S3-P1000

```bash
go run mqtt_scale_tester.go \
  --brokers-json=./brokers.json \
  --cluster-hot-add-new-clients=true \
  --out-dir=./results --test-name=S3-P1000_cluster_hotadd_new_clients_only_1000B \
  --initial-subs=500 --sub-step=100 --max-subs=20000 \
  --publishers=50 --pub-rate=1 \
  --payload-bytes=1000 --topic-count=10 --topic-prefix=bench/topic \
  --window-sec=60 --warmup-sec=30 \
  --sla-min-connected-sub-pct=99 --sla-min-delivery-pct=99 --sla-max-p95-ms=200 --sla-max-disc-per-min=50 \
  --sla-consecutive-breaches=2
```
