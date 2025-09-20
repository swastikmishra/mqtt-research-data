# MQTT Research Data - Test Cases Overview

## 00 Single

### Local Device (Broker + Pub-Sub)

| Code            | Publishers | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| --------------- | ---------- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| SN-LOCAL-10K-Q0 | 10000      | 1000          | Q0  | 10                | 1                           | 5                    |
| SN-LOCAL-10K-Q1 | 10000      | 1000          | Q1  | 10                | 1                           | 5                    |
| SN-LOCAL-10K-Q2 | 10000      | 1000          | Q2  | 10                | 1                           | 5                    |
| SN-LOCAL-20K-Q0 | 20000      | 1000          | Q0  | 10                | 1                           | 5                    |
| SN-LOCAL-20K-Q1 | 20000      | 1000          | Q1  | 10                | 1                           | 5                    |
| SN-LOCAL-20K-Q2 | 20000      | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker + Pub-Sub)

| Code          | Publishers | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ------------- | ---------- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| SN-AWS-10K-Q0 | 10000      | 1000          | Q0  | 10                | 1                           | 5                    |
| SN-AWS-10K-Q1 | 10000      | 1000          | Q1  | 10                | 1                           | 5                    |
| SN-AWS-10K-Q2 | 10000      | 1000          | Q2  | 10                | 1                           | 5                    |
| SN-AWS-20K-Q0 | 20000      | 1000          | Q0  | 10                | 1                           | 5                    |
| SN-AWS-20K-Q1 | 20000      | 1000          | Q1  | 10                | 1                           | 5                    |
| SN-AWS-20K-Q2 | 20000      | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker) + Local (Pub-Sub)

| Code               | Publishers | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ------------------ | ---------- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| SN-AWSLOCAL-10K-Q0 | 10000      | 1000          | Q0  | 10                | 1                           | 5                    |
| SN-AWSLOCAL-10K-Q1 | 10000      | 1000          | Q1  | 10                | 1                           | 5                    |
| SN-AWSLOCAL-10K-Q2 | 10000      | 1000          | Q2  | 10                | 1                           | 5                    |
| SN-AWSLOCAL-20K-Q0 | 20000      | 1000          | Q0  | 10                | 1                           | 5                    |
| SN-AWSLOCAL-20K-Q1 | 20000      | 1000          | Q1  | 10                | 1                           | 5                    |
| SN-AWSLOCAL-20K-Q2 | 20000      | 1000          | Q2  | 10                | 1                           | 5                    |

---

## 01 Horizontal

### Local Device (Broker + Pub-Sub)

| Code            | Publishers | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| --------------- | ---------- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| HZ-LOCAL-10K-Q0 | 10000      | 1000          | Q0  | 10                | 1                           | 5                    |
| HZ-LOCAL-10K-Q1 | 10000      | 1000          | Q1  | 10                | 1                           | 5                    |
| HZ-LOCAL-10K-Q2 | 10000      | 1000          | Q2  | 10                | 1                           | 5                    |
| HZ-LOCAL-20K-Q0 | 20000      | 1000          | Q0  | 10                | 1                           | 5                    |
| HZ-LOCAL-20K-Q1 | 20000      | 1000          | Q1  | 10                | 1                           | 5                    |
| HZ-LOCAL-20K-Q2 | 20000      | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker + Pub-Sub)

| Code          | Publishers | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ------------- | ---------- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| HZ-AWS-10K-Q0 | 10000      | 1000          | Q0  | 10                | 1                           | 5                    |
| HZ-AWS-10K-Q1 | 10000      | 1000          | Q1  | 10                | 1                           | 5                    |
| HZ-AWS-10K-Q2 | 10000      | 1000          | Q2  | 10                | 1                           | 5                    |
| HZ-AWS-20K-Q0 | 20000      | 1000          | Q0  | 10                | 1                           | 5                    |
| HZ-AWS-20K-Q1 | 20000      | 1000          | Q1  | 10                | 1                           | 5                    |
| HZ-AWS-20K-Q2 | 20000      | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker) + Local (Pub-Sub)

| Code               | Publishers | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ------------------ | ---------- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| HZ-AWSLOCAL-10K-Q0 | 10000      | 1000          | Q0  | 10                | 1                           | 5                    |
| HZ-AWSLOCAL-10K-Q1 | 10000      | 1000          | Q1  | 10                | 1                           | 5                    |
| HZ-AWSLOCAL-10K-Q2 | 10000      | 1000          | Q2  | 10                | 1                           | 5                    |
| HZ-AWSLOCAL-20K-Q0 | 20000      | 1000          | Q0  | 10                | 1                           | 5                    |
| HZ-AWSLOCAL-20K-Q1 | 20000      | 1000          | Q1  | 10                | 1                           | 5                    |
| HZ-AWSLOCAL-20K-Q2 | 20000      | 1000          | Q2  | 10                | 1                           | 5                    |

---

## 02 Vertical

### Local Device (Broker + Pub-Sub)

| Code            | Publishers | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| --------------- | ---------- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| VT-LOCAL-10K-Q0 | 10000      | 1000          | Q0  | 10                | 1                           | 5                    |
| VT-LOCAL-10K-Q1 | 10000      | 1000          | Q1  | 10                | 1                           | 5                    |
| VT-LOCAL-10K-Q2 | 10000      | 1000          | Q2  | 10                | 1                           | 5                    |
| VT-LOCAL-20K-Q0 | 20000      | 1000          | Q0  | 10                | 1                           | 5                    |
| VT-LOCAL-20K-Q1 | 20000      | 1000          | Q1  | 10                | 1                           | 5                    |
| VT-LOCAL-20K-Q2 | 20000      | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker + Pub-Sub)

| Code          | Publishers | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ------------- | ---------- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| VT-AWS-10K-Q0 | 10000      | 1000          | Q0  | 10                | 1                           | 5                    |
| VT-AWS-10K-Q1 | 10000      | 1000          | Q1  | 10                | 1                           | 5                    |
| VT-AWS-10K-Q2 | 10000      | 1000          | Q2  | 10                | 1                           | 5                    |
| VT-AWS-20K-Q0 | 20000      | 1000          | Q0  | 10                | 1                           | 5                    |
| VT-AWS-20K-Q1 | 20000      | 1000          | Q1  | 10                | 1                           | 5                    |
| VT-AWS-20K-Q2 | 20000      | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker) + Local (Pub-Sub)

| Code               | Publishers | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ------------------ | ---------- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| VT-AWSLOCAL-10K-Q0 | 10000      | 1000          | Q0  | 10                | 1                           | 5                    |
| VT-AWSLOCAL-10K-Q1 | 10000      | 1000          | Q1  | 10                | 1                           | 5                    |
| VT-AWSLOCAL-10K-Q2 | 10000      | 1000          | Q2  | 10                | 1                           | 5                    |
| VT-AWSLOCAL-20K-Q0 | 20000      | 1000          | Q0  | 10                | 1                           | 5                    |
| VT-AWSLOCAL-20K-Q1 | 20000      | 1000          | Q1  | 10                | 1                           | 5                    |
| VT-AWSLOCAL-20K-Q2 | 20000      | 1000          | Q2  | 10                | 1                           | 5                    |

## How to run the tests

-   Make sure you have placed the correct `brokers.json` file in the root directory.
-   Default `brokers.json` for single/horizontal testing:
    ```json
    [
        {
            "name": "broker",
            "host": "localhost",
            "port": 1883
        }
    ]
    ```
-   Then run the go docker container:

    ```bash
      docker compose up -d --build
    ```

-   Enter into the docker container:

    ```bash
      docker compose exec app bash
    ```

-   Build the go application:
    ```bash
      go build -o main .
    ```

# Test commands for each use cases

## Single (SN) Tests

### Local Device Tests
```bash
# SN-LOCAL-10K-Q0
testCode=SN-LOCAL-10K-Q0 totalUsers=10000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-LOCAL-10K-Q1
testCode=SN-LOCAL-10K-Q1 totalUsers=10000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-LOCAL-10K-Q2
testCode=SN-LOCAL-10K-Q2 totalUsers=10000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-LOCAL-20K-Q0
testCode=SN-LOCAL-20K-Q0 totalUsers=20000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-LOCAL-20K-Q1
testCode=SN-LOCAL-20K-Q1 totalUsers=20000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-LOCAL-20K-Q2
testCode=SN-LOCAL-20K-Q2 totalUsers=20000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS Tests
```bash
# SN-AWS-10K-Q0
testCode=SN-AWS-10K-Q0 totalUsers=10000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWS-10K-Q1
testCode=SN-AWS-10K-Q1 totalUsers=10000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWS-10K-Q2
testCode=SN-AWS-10K-Q2 totalUsers=10000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWS-20K-Q0
testCode=SN-AWS-20K-Q0 totalUsers=20000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWS-20K-Q1
testCode=SN-AWS-20K-Q1 totalUsers=20000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWS-20K-Q2
testCode=SN-AWS-20K-Q2 totalUsers=20000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS-Local Tests
```bash
# SN-AWSLOCAL-10K-Q0
testCode=SN-AWSLOCAL-10K-Q0 totalUsers=10000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWSLOCAL-10K-Q1
testCode=SN-AWSLOCAL-10K-Q1 totalUsers=10000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWSLOCAL-10K-Q2
testCode=SN-AWSLOCAL-10K-Q2 totalUsers=10000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWSLOCAL-20K-Q0
testCode=SN-AWSLOCAL-20K-Q0 totalUsers=20000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWSLOCAL-20K-Q1
testCode=SN-AWSLOCAL-20K-Q1 totalUsers=20000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SN-AWSLOCAL-20K-Q2
testCode=SN-AWSLOCAL-20K-Q2 totalUsers=20000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

## Horizontal (HZ) Tests

### Local Device Tests
```bash
# HZ-LOCAL-10K-Q0
testCode=HZ-LOCAL-10K-Q0 totalUsers=10000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-LOCAL-10K-Q1
testCode=HZ-LOCAL-10K-Q1 totalUsers=10000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-LOCAL-10K-Q2
testCode=HZ-LOCAL-10K-Q2 totalUsers=10000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-LOCAL-20K-Q0
testCode=HZ-LOCAL-20K-Q0 totalUsers=20000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-LOCAL-20K-Q1
testCode=HZ-LOCAL-20K-Q1 totalUsers=20000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-LOCAL-20K-Q2
testCode=HZ-LOCAL-20K-Q2 totalUsers=20000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS Tests
```bash
# HZ-AWS-10K-Q0
testCode=HZ-AWS-10K-Q0 totalUsers=10000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWS-10K-Q1
testCode=HZ-AWS-10K-Q1 totalUsers=10000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWS-10K-Q2
testCode=HZ-AWS-10K-Q2 totalUsers=10000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWS-20K-Q0
testCode=HZ-AWS-20K-Q0 totalUsers=20000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWS-20K-Q1
testCode=HZ-AWS-20K-Q1 totalUsers=20000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWS-20K-Q2
testCode=HZ-AWS-20K-Q2 totalUsers=20000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS-Local Tests
```bash
# HZ-AWSLOCAL-10K-Q0
testCode=HZ-AWSLOCAL-10K-Q0 totalUsers=10000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWSLOCAL-10K-Q1
testCode=HZ-AWSLOCAL-10K-Q1 totalUsers=10000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWSLOCAL-10K-Q2
testCode=HZ-AWSLOCAL-10K-Q2 totalUsers=10000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWSLOCAL-20K-Q0
testCode=HZ-AWSLOCAL-20K-Q0 totalUsers=20000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWSLOCAL-20K-Q1
testCode=HZ-AWSLOCAL-20K-Q1 totalUsers=20000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HZ-AWSLOCAL-20K-Q2
testCode=HZ-AWSLOCAL-20K-Q2 totalUsers=20000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

## Vertical (VT) Tests

### Local Device Tests
```bash
# VT-LOCAL-10K-Q0
testCode=VT-LOCAL-10K-Q0 totalUsers=10000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-LOCAL-10K-Q1
testCode=VT-LOCAL-10K-Q1 totalUsers=10000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-LOCAL-10K-Q2
testCode=VT-LOCAL-10K-Q2 totalUsers=10000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-LOCAL-20K-Q0
testCode=VT-LOCAL-20K-Q0 totalUsers=20000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-LOCAL-20K-Q1
testCode=VT-LOCAL-20K-Q1 totalUsers=20000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-LOCAL-20K-Q2
testCode=VT-LOCAL-20K-Q2 totalUsers=20000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS Tests
```bash
# VT-AWS-10K-Q0
testCode=VT-AWS-10K-Q0 totalUsers=10000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWS-10K-Q1
testCode=VT-AWS-10K-Q1 totalUsers=10000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWS-10K-Q2
testCode=VT-AWS-10K-Q2 totalUsers=10000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWS-20K-Q0
testCode=VT-AWS-20K-Q0 totalUsers=20000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWS-20K-Q1
testCode=VT-AWS-20K-Q1 totalUsers=20000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWS-20K-Q2
testCode=VT-AWS-20K-Q2 totalUsers=20000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS-Local Tests
```bash
# VT-AWSLOCAL-10K-Q0
testCode=VT-AWSLOCAL-10K-Q0 totalUsers=10000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWSLOCAL-10K-Q1
testCode=VT-AWSLOCAL-10K-Q1 totalUsers=10000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWSLOCAL-10K-Q2
testCode=VT-AWSLOCAL-10K-Q2 totalUsers=10000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWSLOCAL-20K-Q0
testCode=VT-AWSLOCAL-20K-Q0 totalUsers=20000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWSLOCAL-20K-Q1
testCode=VT-AWSLOCAL-20K-Q1 totalUsers=20000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VT-AWSLOCAL-20K-Q2
testCode=VT-AWSLOCAL-20K-Q2 totalUsers=20000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

---

## Summary

Total test commands generated: **54**

-   **Single tests**: 18 commands (6 Local + 6 AWS + 6 AWS-Local)
-   **Horizontal tests**: 18 commands (6 Local + 6 AWS + 6 AWS-Local)
-   **Vertical tests**: 18 commands (6 Local + 6 AWS + 6 AWS-Local)

Each category tests 2 different user counts (10K and 20K users) with 3 different QoS levels (0, 1, 2).

### Command Format Parameters:

-   `testCode`: Unique identifier for each test case
-   `totalUsers`: Number of total users (10000 or 20000 for all tests)
-   `rampUpRate`: Users added per second (1000 for all tests)
-   `qos`: Quality of Service level (0, 1, or 2)
-   `messageSizeKB`: Message size in KB (10 for all tests)
-   `publishRatePerSec`: Messages published per second per user (1 for all tests)
-   `testDurationMins`: Test duration in minutes (5 for all tests)
