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

#### Set 1: Single Node Scenarios

| ID  | Test Name            | Subs  | Pubs | Payload | Rate (per Pub) | Total Tput  | Duration | Test Goal                                                                  |
| --- | -------------------- | ----- | ---- | ------- | -------------- | ----------- | -------- | -------------------------------------------------------------------------- |
| S1  | single_latency_base  | 100   | 1    | 1 KB    | 10             | 10 msg/s    | 180s     | Baseline Latency. Minimal load to find "best case" response time.          |
| S2  | single_cpu_stress    | 500   | 50   | 100 B   | 50             | 2,500 msg/s | 180s     | High Throughput (CPU). High interrupt rate with small packets.             |
| S3  | single_fanout_stress | 1,000 | 10   | 100 B   | 100            | 1,000 msg/s | 180s     | Fan-Out Logic. High fan-out (1 pub → 1000 subs) to test dispatch speed.    |
| S4  | single_conn_max      | 5,000 | 1    | 1 KB    | 1              | 1 msg/s     | 300s     | Max Connections (RAM). Idle connection holding capacity (Longer duration). |
| S5  | single_bw_limit      | 200   | 10   | 50 KB   | 5              | 50 msg/s    | 180s     | Bandwidth Saturation. Tests network I/O limits (~25MB/s ingress).          |

#### Set 2: Clustered Node Scenarios

| ID  | Test Name             | Subs  | Pubs | Payload | Rate (per Pub) | Total Tput  | Duration | Test Goal                                                                        |
| --- | --------------------- | ----- | ---- | ------- | -------------- | ----------- | -------- | -------------------------------------------------------------------------------- |
| C1  | cluster_latency_base  | 100   | 1    | 1 KB    | 10             | 10 msg/s    | 180s     | Cluster Tax. Compare vs S1 to see overhead of cluster hops.                      |
| C2  | cluster_cpu_stress    | 500   | 50   | 100 B   | 50             | 2,500 msg/s | 180s     | CPU & Sync. High packet rate stresses internal cluster state sync.               |
| C3  | cluster_fanout_stress | 1,000 | 10   | 100 B   | 100            | 1,000 msg/s | 180s     | Routing Efficiency. Tests if messages reach subs on different nodes efficiently. |
| C4  | cluster_conn_max      | 5,000 | 1    | 1 KB    | 1              | 1 msg/s     | 300s     | Distributed RAM. Can the cluster hold more idle conns than Single?               |
| C5  | cluster_bw_limit      | 200   | 10   | 50 KB   | 5              | 50 msg/s    | 180s     | Network Dist. Does splitting traffic across nodes increase total BW capacity?    |

### Test Commands:

#### Single Node Commands

**S1: Baseline Latency**

```bash
go run main.go \
  --broker-host=43.205.176.30 --broker-port=1883 \
  --out-dir=./results --test-name=single_latency_base \
  --fixed-subs=100 --fixed-pubs=1 \
  --payload-kb=1 --pub-rate=10 \
  --max-duration-sec=180 --warmup-sec=10 \
  --topic-count=10 \
  --min-connected-sub-pct=90 --min-delivery-ratio-pct=99 --max-p95-latency-ms=200
```

**S2: CPU Stress (High Packet Rate)**

```bash
go run main.go \
  --broker-host=43.205.176.30 --broker-port=1883 \
  --out-dir=./results --test-name=single_cpu_stress \
  --fixed-subs=500 --fixed-pubs=50 \
  --payload-kb=0.1 --pub-rate=50 \
  --max-duration-sec=180 --warmup-sec=10 \
  --topic-count=10 \
  --min-connected-sub-pct=90 --min-delivery-ratio-pct=95 --max-p95-latency-ms=1000
```

**S3: Fan-Out Stress**

```bash
go run main.go \
  --broker-host=43.205.176.30 --broker-port=1883 \
  --out-dir=./results --test-name=single_fanout_stress \
  --fixed-subs=1000 --fixed-pubs=10 \
  --payload-kb=0.1 --pub-rate=100 \
  --max-duration-sec=180 --warmup-sec=10 \
  --topic-count=10 \
  --min-connected-sub-pct=90 --min-delivery-ratio-pct=95 --max-p95-latency-ms=2000
```

**S4: Max Connections (RAM)**

```bash
go run main.go \
  --broker-host=43.205.176.30 --broker-port=1883 \
  --out-dir=./results --test-name=single_conn_max \
  --fixed-subs=5000 --fixed-pubs=1 \
  --payload-kb=1 --pub-rate=1 \
  --max-duration-sec=300 --warmup-sec=20 \
  --topic-count=10 \
  --min-connected-sub-pct=95 --min-delivery-ratio-pct=99 --max-p95-latency-ms=5000
```

**S5: Bandwidth Limit**

```bash
go run main.go \
  --broker-host=43.205.176.30 --broker-port=1883 \
  --out-dir=./results --test-name=single_bw_limit \
  --fixed-subs=200 --fixed-pubs=10 \
  --payload-kb=50 --pub-rate=5 \
  --max-duration-sec=180 --warmup-sec=10 \
  --topic-count=10 \
  --min-connected-sub-pct=90 --min-delivery-ratio-pct=95 --max-p95-latency-ms=2000
```


#### Cluster Node Commands

**C1: Baseline Latency**

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster_latency_base \
  --fixed-subs=100 --fixed-pubs=1 \
  --payload-kb=1 --pub-rate=10 \
  --max-duration-sec=180 --warmup-sec=10 \
  --topic-count=10 \
  --min-connected-sub-pct=90 --min-delivery-ratio-pct=99 --max-p95-latency-ms=200
```

**C2: CPU Stress (High Packet Rate)**

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster_cpu_stress \
  --fixed-subs=500 --fixed-pubs=50 \
  --payload-kb=0.1 --pub-rate=50 \
  --max-duration-sec=180 --warmup-sec=10 \
  --topic-count=10 \
  --min-connected-sub-pct=90 --min-delivery-ratio-pct=95 --max-p95-latency-ms=1000
```

**C3: Fan-Out Stress**

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster_fanout_stress \
  --fixed-subs=1000 --fixed-pubs=10 \
  --payload-kb=0.1 --pub-rate=100 \
  --max-duration-sec=180 --warmup-sec=10 \
  --topic-count=10 \
  --min-connected-sub-pct=90 --min-delivery-ratio-pct=95 --max-p95-latency-ms=2000
```

**C4: Max Connections (RAM)**

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster_conn_max \
  --fixed-subs=5000 --fixed-pubs=1 \
  --payload-kb=1 --pub-rate=1 \
  --max-duration-sec=300 --warmup-sec=20 \
  --topic-count=10 \
  --min-connected-sub-pct=95 --min-delivery-ratio-pct=99 --max-p95-latency-ms=5000
```

**C5: Bandwidth Limit**

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster_bw_limit \
  --fixed-subs=200 --fixed-pubs=10 \
  --payload-kb=50 --pub-rate=5 \
  --max-duration-sec=180 --warmup-sec=10 \
  --topic-count=10 \
  --min-connected-sub-pct=90 --min-delivery-ratio-pct=95 --max-p95-latency-ms=2000
```