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
