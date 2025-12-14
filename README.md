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

| ID  | Type    | Subs Count | Pubs Count | Payload (KB) | Duration (Seconds) | Test Name |
| --- | ------- | ---------- | ---------- | ------------ | ------------------ | --------- |
| 1   | single  | 100        | 10         | 100          | 300                | single1   |
| 2   | single  | 100        | 10         | 100          | 300                | single2   |
| 3   | single  | 100        | 10         | 100          | 300                | single3   |
| 4   | single  | 250        | 25         | 100          | 300                | single4   |
| 5   | single  | 250        | 25         | 100          | 300                | single5   |
| 6   | single  | 250        | 25         | 100          | 300                | single6   |
| 7   | single  | 500        | 50         | 100          | 300                | single7   |
| 8   | single  | 500        | 50         | 100          | 300                | single8   |
| 9   | single  | 500        | 50         | 100          | 300                | single9   |
| 1   | cluster | 100        | 10         | 100          | 300                | cluster1  |
| 2   | cluster | 100        | 10         | 100          | 300                | cluster2  |
| 3   | cluster | 100        | 10         | 100          | 300                | cluster3  |
| 4   | cluster | 250        | 25         | 100          | 300                | cluster4  |
| 5   | cluster | 250        | 25         | 100          | 300                | cluster5  |
| 6   | cluster | 250        | 25         | 100          | 300                | cluster6  |
| 7   | cluster | 500        | 50         | 100          | 300                | cluster7  |
| 8   | cluster | 500        | 50         | 100          | 300                | cluster8  |
| 9   | cluster | 500        | 50         | 100          | 300                | cluster9  |

### Test Commands:

single1

```bash
go run main.go \
  --broker-host=127.0.0.1 --broker-port=1883 \
  --out-dir=./results --test-name=single1 \
  --fixed-subs=100 --fixed-pubs=10 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

single2

```bash
go run main.go \
  --broker-host=127.0.0.1 --broker-port=1883 \
  --out-dir=./results --test-name=single2 \
  --fixed-subs=100 --fixed-pubs=10 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

single3

```bash
go run main.go \
  --broker-host=127.0.0.1 --broker-port=1883 \
  --out-dir=./results --test-name=single3 \
  --fixed-subs=100 --fixed-pubs=10 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

single4

```bash
go run main.go \
  --broker-host=127.0.0.1 --broker-port=1883 \
  --out-dir=./results --test-name=single4 \
  --fixed-subs=250 --fixed-pubs=25 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

single5

```bash
go run main.go \
  --broker-host=127.0.0.1 --broker-port=1883 \
  --out-dir=./results --test-name=single5 \
  --fixed-subs=250 --fixed-pubs=25 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

single6

```bash
go run main.go \
  --broker-host=127.0.0.1 --broker-port=1883 \
  --out-dir=./results --test-name=single6 \
  --fixed-subs=250 --fixed-pubs=25 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

single7

```bash
go run main.go \
  --broker-host=127.0.0.1 --broker-port=1883 \
  --out-dir=./results --test-name=single7 \
  --fixed-subs=500 --fixed-pubs=50 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

single8

```bash
go run main.go \
  --broker-host=127.0.0.1 --broker-port=1883 \
  --out-dir=./results --test-name=single8 \
  --fixed-subs=500 --fixed-pubs=50 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

single9

```bash
go run main.go \
  --broker-host=127.0.0.1 --broker-port=1883 \
  --out-dir=./results --test-name=single9 \
  --fixed-subs=500 --fixed-pubs=50 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

cluster1

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster1 \
  --fixed-subs=100 --fixed-pubs=10 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

cluster2

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster2 \
  --fixed-subs=100 --fixed-pubs=10 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

cluster3

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster3 \
  --fixed-subs=100 --fixed-pubs=10 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

cluster4

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster4 \
  --fixed-subs=250 --fixed-pubs=25 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

cluster5

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster5 \
  --fixed-subs=250 --fixed-pubs=25 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

cluster6

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster6 \
  --fixed-subs=250 --fixed-pubs=25 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

cluster7

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster7 \
  --fixed-subs=500 --fixed-pubs=50 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

cluster8

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster8 \
  --fixed-subs=500 --fixed-pubs=50 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```

cluster9

```bash
go run main.go \
  --brokers-json=./brokers.json \
  --out-dir=./results --test-name=cluster9 \
  --fixed-subs=500 --fixed-pubs=50 \
  --payload-kb=100 --pub-rate=1 \
  --topic-count=10 \
  --max-duration-sec=300 --warmup-sec=10
```
