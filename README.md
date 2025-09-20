# MQTT Research Data - Test Cases Overview

## 00 Single

### Local Device (Broker + Pub-Sub)

| Code                  | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| --------------------- | ----- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| SINGLE-LOCAL-R10-Q0   | 1000  | 10            | Q0  | 10                | 1                           | 5                    |
| SINGLE-LOCAL-R10-Q1   | 1000  | 10            | Q1  | 10                | 1                           | 5                    |
| SINGLE-LOCAL-R10-Q2   | 1000  | 10            | Q2  | 10                | 1                           | 5                    |
| SINGLE-LOCAL-R100-Q0  | 1000  | 100           | Q0  | 10                | 1                           | 5                    |
| SINGLE-LOCAL-R100-Q1  | 1000  | 100           | Q1  | 10                | 1                           | 5                    |
| SINGLE-LOCAL-R100-Q2  | 1000  | 100           | Q2  | 10                | 1                           | 5                    |
| SINGLE-LOCAL-R1000-Q0 | 1000  | 1000          | Q0  | 10                | 1                           | 5                    |
| SINGLE-LOCAL-R1000-Q1 | 1000  | 1000          | Q1  | 10                | 1                           | 5                    |
| SINGLE-LOCAL-R1000-Q2 | 1000  | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker + Pub-Sub)

| Code                | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ------------------- | ----- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| SINGLE-AWS-R10-Q0   | 1000  | 10            | Q0  | 10                | 1                           | 5                    |
| SINGLE-AWS-R10-Q1   | 1000  | 10            | Q1  | 10                | 1                           | 5                    |
| SINGLE-AWS-R10-Q2   | 1000  | 10            | Q2  | 10                | 1                           | 5                    |
| SINGLE-AWS-R100-Q0  | 1000  | 100           | Q0  | 10                | 1                           | 5                    |
| SINGLE-AWS-R100-Q1  | 1000  | 100           | Q1  | 10                | 1                           | 5                    |
| SINGLE-AWS-R100-Q2  | 1000  | 100           | Q2  | 10                | 1                           | 5                    |
| SINGLE-AWS-R1000-Q0 | 1000  | 1000          | Q0  | 10                | 1                           | 5                    |
| SINGLE-AWS-R1000-Q1 | 1000  | 1000          | Q1  | 10                | 1                           | 5                    |
| SINGLE-AWS-R1000-Q2 | 1000  | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker) + Local (Pub-Sub)

| Code                     | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ------------------------ | ----- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| SINGLE-AWSLOCAL-R10-Q0   | 1000  | 10            | Q0  | 10                | 1                           | 5                    |
| SINGLE-AWSLOCAL-R10-Q1   | 1000  | 10            | Q1  | 10                | 1                           | 5                    |
| SINGLE-AWSLOCAL-R10-Q2   | 1000  | 10            | Q2  | 10                | 1                           | 5                    |
| SINGLE-AWSLOCAL-R100-Q0  | 1000  | 100           | Q0  | 10                | 1                           | 5                    |
| SINGLE-AWSLOCAL-R100-Q1  | 1000  | 100           | Q1  | 10                | 1                           | 5                    |
| SINGLE-AWSLOCAL-R100-Q2  | 1000  | 100           | Q2  | 10                | 1                           | 5                    |
| SINGLE-AWSLOCAL-R1000-Q0 | 1000  | 1000          | Q0  | 10                | 1                           | 5                    |
| SINGLE-AWSLOCAL-R1000-Q1 | 1000  | 1000          | Q1  | 10                | 1                           | 5                    |
| SINGLE-AWSLOCAL-R1000-Q2 | 1000  | 1000          | Q2  | 10                | 1                           | 5                    |

---

## 01 Horizontal

### Local Device (Broker + Pub-Sub)

| Code                      | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ------------------------- | ----- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| HORIZONTAL-LOCAL-R10-Q0   | 1000  | 10            | Q0  | 10                | 1                           | 5                    |
| HORIZONTAL-LOCAL-R10-Q1   | 1000  | 10            | Q1  | 10                | 1                           | 5                    |
| HORIZONTAL-LOCAL-R10-Q2   | 1000  | 10            | Q2  | 10                | 1                           | 5                    |
| HORIZONTAL-LOCAL-R100-Q0  | 1000  | 100           | Q0  | 10                | 1                           | 5                    |
| HORIZONTAL-LOCAL-R100-Q1  | 1000  | 100           | Q1  | 10                | 1                           | 5                    |
| HORIZONTAL-LOCAL-R100-Q2  | 1000  | 100           | Q2  | 10                | 1                           | 5                    |
| HORIZONTAL-LOCAL-R1000-Q0 | 1000  | 1000          | Q0  | 10                | 1                           | 5                    |
| HORIZONTAL-LOCAL-R1000-Q1 | 1000  | 1000          | Q1  | 10                | 1                           | 5                    |
| HORIZONTAL-LOCAL-R1000-Q2 | 1000  | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker + Pub-Sub)

| Code                    | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ----------------------- | ----- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| HORIZONTAL-AWS-R10-Q0   | 1000  | 10            | Q0  | 10                | 1                           | 5                    |
| HORIZONTAL-AWS-R10-Q1   | 1000  | 10            | Q1  | 10                | 1                           | 5                    |
| HORIZONTAL-AWS-R10-Q2   | 1000  | 10            | Q2  | 10                | 1                           | 5                    |
| HORIZONTAL-AWS-R100-Q0  | 1000  | 100           | Q0  | 10                | 1                           | 5                    |
| HORIZONTAL-AWS-R100-Q1  | 1000  | 100           | Q1  | 10                | 1                           | 5                    |
| HORIZONTAL-AWS-R100-Q2  | 1000  | 100           | Q2  | 10                | 1                           | 5                    |
| HORIZONTAL-AWS-R1000-Q0 | 1000  | 1000          | Q0  | 10                | 1                           | 5                    |
| HORIZONTAL-AWS-R1000-Q1 | 1000  | 1000          | Q1  | 10                | 1                           | 5                    |
| HORIZONTAL-AWS-R1000-Q2 | 1000  | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker) + Local (Pub-Sub)

| Code                         | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ---------------------------- | ----- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| HORIZONTAL-AWSLOCAL-R10-Q0   | 1000  | 10            | Q0  | 10                | 1                           | 5                    |
| HORIZONTAL-AWSLOCAL-R10-Q1   | 1000  | 10            | Q1  | 10                | 1                           | 5                    |
| HORIZONTAL-AWSLOCAL-R10-Q2   | 1000  | 10            | Q2  | 10                | 1                           | 5                    |
| HORIZONTAL-AWSLOCAL-R100-Q0  | 1000  | 100           | Q0  | 10                | 1                           | 5                    |
| HORIZONTAL-AWSLOCAL-R100-Q1  | 1000  | 100           | Q1  | 10                | 1                           | 5                    |
| HORIZONTAL-AWSLOCAL-R100-Q2  | 1000  | 100           | Q2  | 10                | 1                           | 5                    |
| HORIZONTAL-AWSLOCAL-R1000-Q0 | 1000  | 1000          | Q0  | 10                | 1                           | 5                    |
| HORIZONTAL-AWSLOCAL-R1000-Q1 | 1000  | 1000          | Q1  | 10                | 1                           | 5                    |
| HORIZONTAL-AWSLOCAL-R1000-Q2 | 1000  | 1000          | Q2  | 10                | 1                           | 5                    |

---

## 02 Vertical

### Local Device (Broker + Pub-Sub)

| Code                    | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| ----------------------- | ----- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| VERTICAL-LOCAL-R10-Q0   | 1000  | 10            | Q0  | 10                | 1                           | 5                    |
| VERTICAL-LOCAL-R10-Q1   | 1000  | 10            | Q1  | 10                | 1                           | 5                    |
| VERTICAL-LOCAL-R10-Q2   | 1000  | 10            | Q2  | 10                | 1                           | 5                    |
| VERTICAL-LOCAL-R100-Q0  | 1000  | 100           | Q0  | 10                | 1                           | 5                    |
| VERTICAL-LOCAL-R100-Q1  | 1000  | 100           | Q1  | 10                | 1                           | 5                    |
| VERTICAL-LOCAL-R100-Q2  | 1000  | 100           | Q2  | 10                | 1                           | 5                    |
| VERTICAL-LOCAL-R1000-Q0 | 1000  | 1000          | Q0  | 10                | 1                           | 5                    |
| VERTICAL-LOCAL-R1000-Q1 | 1000  | 1000          | Q1  | 10                | 1                           | 5                    |
| VERTICAL-LOCAL-R1000-Q2 | 1000  | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker + Pub-Sub)

| Code                  | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| --------------------- | ----- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| VERTICAL-AWS-R10-Q0   | 1000  | 10            | Q0  | 10                | 1                           | 5                    |
| VERTICAL-AWS-R10-Q1   | 1000  | 10            | Q1  | 10                | 1                           | 5                    |
| VERTICAL-AWS-R10-Q2   | 1000  | 10            | Q2  | 10                | 1                           | 5                    |
| VERTICAL-AWS-R100-Q0  | 1000  | 100           | Q0  | 10                | 1                           | 5                    |
| VERTICAL-AWS-R100-Q1  | 1000  | 100           | Q1  | 10                | 1                           | 5                    |
| VERTICAL-AWS-R100-Q2  | 1000  | 100           | Q2  | 10                | 1                           | 5                    |
| VERTICAL-AWS-R1000-Q0 | 1000  | 1000          | Q0  | 10                | 1                           | 5                    |
| VERTICAL-AWS-R1000-Q1 | 1000  | 1000          | Q1  | 10                | 1                           | 5                    |
| VERTICAL-AWS-R1000-Q2 | 1000  | 1000          | Q2  | 10                | 1                           | 5                    |

### AWS (Broker) + Local (Pub-Sub)

| Code                       | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
| -------------------------- | ----- | ------------- | --- | ----------------- | --------------------------- | -------------------- |
| VERTICAL-AWSLOCAL-R10-Q0   | 1000  | 10            | Q0  | 10                | 1                           | 5                    |
| VERTICAL-AWSLOCAL-R10-Q1   | 1000  | 10            | Q1  | 10                | 1                           | 5                    |
| VERTICAL-AWSLOCAL-R10-Q2   | 1000  | 10            | Q2  | 10                | 1                           | 5                    |
| VERTICAL-AWSLOCAL-R100-Q0  | 1000  | 100           | Q0  | 10                | 1                           | 5                    |
| VERTICAL-AWSLOCAL-R100-Q1  | 1000  | 100           | Q1  | 10                | 1                           | 5                    |
| VERTICAL-AWSLOCAL-R100-Q2  | 1000  | 100           | Q2  | 10                | 1                           | 5                    |
| VERTICAL-AWSLOCAL-R1000-Q0 | 1000  | 1000          | Q0  | 10                | 1                           | 5                    |
| VERTICAL-AWSLOCAL-R1000-Q1 | 1000  | 1000          | Q1  | 10                | 1                           | 5                    |
| VERTICAL-AWSLOCAL-R1000-Q2 | 1000  | 1000          | Q2  | 10                | 1                           | 5                    |

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

## 00 Single

### Local Device (Broker + Pub-Sub)

```bash
# SINGLE-LOCAL-R10-Q0
testCode=SINGLE-LOCAL-R10-Q0 totalUsers=1000 rampUpRate=10 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-LOCAL-R10-Q1
testCode=SINGLE-LOCAL-R10-Q1 totalUsers=1000 rampUpRate=10 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-LOCAL-R10-Q2
testCode=SINGLE-LOCAL-R10-Q2 totalUsers=1000 rampUpRate=10 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-LOCAL-R100-Q0
testCode=SINGLE-LOCAL-R100-Q0 totalUsers=1000 rampUpRate=100 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-LOCAL-R100-Q1
testCode=SINGLE-LOCAL-R100-Q1 totalUsers=1000 rampUpRate=100 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-LOCAL-R100-Q2
testCode=SINGLE-LOCAL-R100-Q2 totalUsers=1000 rampUpRate=100 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-LOCAL-R1000-Q0
testCode=SINGLE-LOCAL-R1000-Q0 totalUsers=1000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-LOCAL-R1000-Q1
testCode=SINGLE-LOCAL-R1000-Q1 totalUsers=1000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-LOCAL-R1000-Q2
testCode=SINGLE-LOCAL-R1000-Q2 totalUsers=1000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS (Broker + Pub-Sub)

```bash
# SINGLE-AWS-R10-Q0
testCode=SINGLE-AWS-R10-Q0 totalUsers=1000 rampUpRate=10 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWS-R10-Q1
testCode=SINGLE-AWS-R10-Q1 totalUsers=1000 rampUpRate=10 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWS-R10-Q2
testCode=SINGLE-AWS-R10-Q2 totalUsers=1000 rampUpRate=10 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWS-R100-Q0
testCode=SINGLE-AWS-R100-Q0 totalUsers=1000 rampUpRate=100 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWS-R100-Q1
testCode=SINGLE-AWS-R100-Q1 totalUsers=1000 rampUpRate=100 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWS-R100-Q2
testCode=SINGLE-AWS-R100-Q2 totalUsers=1000 rampUpRate=100 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWS-R1000-Q0
testCode=SINGLE-AWS-R1000-Q0 totalUsers=1000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWS-R1000-Q1
testCode=SINGLE-AWS-R1000-Q1 totalUsers=1000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWS-R1000-Q2
testCode=SINGLE-AWS-R1000-Q2 totalUsers=1000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS (Broker) + Local (Pub-Sub)

```bash
# SINGLE-AWSLOCAL-R10-Q0
testCode=SINGLE-AWSLOCAL-R10-Q0 totalUsers=1000 rampUpRate=10 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWSLOCAL-R10-Q1
testCode=SINGLE-AWSLOCAL-R10-Q1 totalUsers=1000 rampUpRate=10 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWSLOCAL-R10-Q2
testCode=SINGLE-AWSLOCAL-R10-Q2 totalUsers=1000 rampUpRate=10 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWSLOCAL-R100-Q0
testCode=SINGLE-AWSLOCAL-R100-Q0 totalUsers=1000 rampUpRate=100 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWSLOCAL-R100-Q1
testCode=SINGLE-AWSLOCAL-R100-Q1 totalUsers=1000 rampUpRate=100 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWSLOCAL-R100-Q2
testCode=SINGLE-AWSLOCAL-R100-Q2 totalUsers=1000 rampUpRate=100 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWSLOCAL-R1000-Q0
testCode=SINGLE-AWSLOCAL-R1000-Q0 totalUsers=1000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWSLOCAL-R1000-Q1
testCode=SINGLE-AWSLOCAL-R1000-Q1 totalUsers=1000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# SINGLE-AWSLOCAL-R1000-Q2
testCode=SINGLE-AWSLOCAL-R1000-Q2 totalUsers=1000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

---

## 01 Horizontal

### Local Device (Broker + Pub-Sub)

```bash
# HORIZONTAL-LOCAL-R10-Q0
testCode=HORIZONTAL-LOCAL-R10-Q0 totalUsers=1000 rampUpRate=10 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-LOCAL-R10-Q1
testCode=HORIZONTAL-LOCAL-R10-Q1 totalUsers=1000 rampUpRate=10 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-LOCAL-R10-Q2
testCode=HORIZONTAL-LOCAL-R10-Q2 totalUsers=1000 rampUpRate=10 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-LOCAL-R100-Q0
testCode=HORIZONTAL-LOCAL-R100-Q0 totalUsers=1000 rampUpRate=100 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-LOCAL-R100-Q1
testCode=HORIZONTAL-LOCAL-R100-Q1 totalUsers=1000 rampUpRate=100 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-LOCAL-R100-Q2
testCode=HORIZONTAL-LOCAL-R100-Q2 totalUsers=1000 rampUpRate=100 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-LOCAL-R1000-Q0
testCode=HORIZONTAL-LOCAL-R1000-Q0 totalUsers=1000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-LOCAL-R1000-Q1
testCode=HORIZONTAL-LOCAL-R1000-Q1 totalUsers=1000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-LOCAL-R1000-Q2
testCode=HORIZONTAL-LOCAL-R1000-Q2 totalUsers=1000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS (Broker + Pub-Sub)

```bash
# HORIZONTAL-AWS-R10-Q0
testCode=HORIZONTAL-AWS-R10-Q0 totalUsers=1000 rampUpRate=10 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWS-R10-Q1
testCode=HORIZONTAL-AWS-R10-Q1 totalUsers=1000 rampUpRate=10 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWS-R10-Q2
testCode=HORIZONTAL-AWS-R10-Q2 totalUsers=1000 rampUpRate=10 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWS-R100-Q0
testCode=HORIZONTAL-AWS-R100-Q0 totalUsers=1000 rampUpRate=100 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWS-R100-Q1
testCode=HORIZONTAL-AWS-R100-Q1 totalUsers=1000 rampUpRate=100 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWS-R100-Q2
testCode=HORIZONTAL-AWS-R100-Q2 totalUsers=1000 rampUpRate=100 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWS-R1000-Q0
testCode=HORIZONTAL-AWS-R1000-Q0 totalUsers=1000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWS-R1000-Q1
testCode=HORIZONTAL-AWS-R1000-Q1 totalUsers=1000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWS-R1000-Q2
testCode=HORIZONTAL-AWS-R1000-Q2 totalUsers=1000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS (Broker) + Local (Pub-Sub)

```bash
# HORIZONTAL-AWSLOCAL-R10-Q0
testCode=HORIZONTAL-AWSLOCAL-R10-Q0 totalUsers=1000 rampUpRate=10 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWSLOCAL-R10-Q1
testCode=HORIZONTAL-AWSLOCAL-R10-Q1 totalUsers=1000 rampUpRate=10 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWSLOCAL-R10-Q2
testCode=HORIZONTAL-AWSLOCAL-R10-Q2 totalUsers=1000 rampUpRate=10 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWSLOCAL-R100-Q0
testCode=HORIZONTAL-AWSLOCAL-R100-Q0 totalUsers=1000 rampUpRate=100 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWSLOCAL-R100-Q1
testCode=HORIZONTAL-AWSLOCAL-R100-Q1 totalUsers=1000 rampUpRate=100 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWSLOCAL-R100-Q2
testCode=HORIZONTAL-AWSLOCAL-R100-Q2 totalUsers=1000 rampUpRate=100 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWSLOCAL-R1000-Q0
testCode=HORIZONTAL-AWSLOCAL-R1000-Q0 totalUsers=1000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWSLOCAL-R1000-Q1
testCode=HORIZONTAL-AWSLOCAL-R1000-Q1 totalUsers=1000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# HORIZONTAL-AWSLOCAL-R1000-Q2
testCode=HORIZONTAL-AWSLOCAL-R1000-Q2 totalUsers=1000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

---

## 02 Vertical

### Local Device (Broker + Pub-Sub)

```bash
# VERTICAL-LOCAL-R10-Q0
testCode=VERTICAL-LOCAL-R10-Q0 totalUsers=1000 rampUpRate=10 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-LOCAL-R10-Q1
testCode=VERTICAL-LOCAL-R10-Q1 totalUsers=1000 rampUpRate=10 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-LOCAL-R10-Q2
testCode=VERTICAL-LOCAL-R10-Q2 totalUsers=1000 rampUpRate=10 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-LOCAL-R100-Q0
testCode=VERTICAL-LOCAL-R100-Q0 totalUsers=1000 rampUpRate=100 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-LOCAL-R100-Q1
testCode=VERTICAL-LOCAL-R100-Q1 totalUsers=1000 rampUpRate=100 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-LOCAL-R100-Q2
testCode=VERTICAL-LOCAL-R100-Q2 totalUsers=1000 rampUpRate=100 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-LOCAL-R1000-Q0
testCode=VERTICAL-LOCAL-R1000-Q0 totalUsers=1000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-LOCAL-R1000-Q1
testCode=VERTICAL-LOCAL-R1000-Q1 totalUsers=1000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-LOCAL-R1000-Q2
testCode=VERTICAL-LOCAL-R1000-Q2 totalUsers=1000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS (Broker + Pub-Sub)

```bash
# VERTICAL-AWS-R10-Q0
testCode=VERTICAL-AWS-R10-Q0 totalUsers=1000 rampUpRate=10 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWS-R10-Q1
testCode=VERTICAL-AWS-R10-Q1 totalUsers=1000 rampUpRate=10 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWS-R10-Q2
testCode=VERTICAL-AWS-R10-Q2 totalUsers=1000 rampUpRate=10 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWS-R100-Q0
testCode=VERTICAL-AWS-R100-Q0 totalUsers=1000 rampUpRate=100 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWS-R100-Q1
testCode=VERTICAL-AWS-R100-Q1 totalUsers=1000 rampUpRate=100 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWS-R100-Q2
testCode=VERTICAL-AWS-R100-Q2 totalUsers=1000 rampUpRate=100 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWS-R1000-Q0
testCode=VERTICAL-AWS-R1000-Q0 totalUsers=1000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWS-R1000-Q1
testCode=VERTICAL-AWS-R1000-Q1 totalUsers=1000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWS-R1000-Q2
testCode=VERTICAL-AWS-R1000-Q2 totalUsers=1000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

### AWS (Broker) + Local (Pub-Sub)

```bash
# VERTICAL-AWSLOCAL-R10-Q0
testCode=VERTICAL-AWSLOCAL-R10-Q0 totalUsers=1000 rampUpRate=10 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWSLOCAL-R10-Q1
testCode=VERTICAL-AWSLOCAL-R10-Q1 totalUsers=1000 rampUpRate=10 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWSLOCAL-R10-Q2
testCode=VERTICAL-AWSLOCAL-R10-Q2 totalUsers=1000 rampUpRate=10 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWSLOCAL-R100-Q0
testCode=VERTICAL-AWSLOCAL-R100-Q0 totalUsers=1000 rampUpRate=100 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWSLOCAL-R100-Q1
testCode=VERTICAL-AWSLOCAL-R100-Q1 totalUsers=1000 rampUpRate=100 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWSLOCAL-R100-Q2
testCode=VERTICAL-AWSLOCAL-R100-Q2 totalUsers=1000 rampUpRate=100 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWSLOCAL-R1000-Q0
testCode=VERTICAL-AWSLOCAL-R1000-Q0 totalUsers=1000 rampUpRate=1000 qos=0 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWSLOCAL-R1000-Q1
testCode=VERTICAL-AWSLOCAL-R1000-Q1 totalUsers=1000 rampUpRate=1000 qos=1 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main

# VERTICAL-AWSLOCAL-R1000-Q2
testCode=VERTICAL-AWSLOCAL-R1000-Q2 totalUsers=1000 rampUpRate=1000 qos=2 messageSizeKB=10 publishRatePerSec=1 testDurationMins=5 ./main
```

---

## Summary

Total test commands generated: **81**

-   **Single tests**: 27 commands (9 Local + 9 AWS + 9 AWS-Local)
-   **Horizontal tests**: 27 commands (9 Local + 9 AWS + 9 AWS-Local)
-   **Vertical tests**: 27 commands (9 Local + 9 AWS + 9 AWS-Local)

Each category tests 3 different ramp-up rates (10, 100, 1000) with 3 different QoS levels (0, 1, 2).

### Command Format Parameters:

-   `testCode`: Unique identifier for each test case
-   `totalUsers`: Number of total users (1000 for all tests)
-   `rampUpRate`: Users added per second (10, 100, or 1000)
-   `qos`: Quality of Service level (0, 1, or 2)
-   `messageSizeKB`: Message size in KB (10 for all tests)
-   `publishRatePerSec`: Messages published per second per user (1 for all tests)
-   `testDurationMins`: Test duration in minutes (5 for all tests)
