# MQTT Research Data - Test Cases Overview

## 00 Single

### Local Device (Broker + Pub-Sub)
**Test Config:** [tests.json](test-cases/00%20Single/1.%20Local%20Device%20(Broker%20+%20Pub-Sub)/tests.json)

| Code | Category | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
|------|----------|-------|---------------|-----|-------------------|---------------------------|---------------------|
| LOCAL-HS-A-Q0 | SINGLE | 1000 | 10 | Q0 | 1 | 1 | 10 |
| LOCAL-HS-A-Q1 | SINGLE | 1000 | 10 | Q1 | 1 | 1 | 10 |
| LOCAL-HS-A-Q2 | SINGLE | 1000 | 10 | Q2 | 1 | 1 | 10 |
| LOCAL-HS-B-Q0 | SINGLE | 1000 | 100 | Q0 | 10 | 1 | 10 |
| LOCAL-HS-B-Q1 | SINGLE | 1000 | 100 | Q1 | 10 | 1 | 10 |
| LOCAL-HS-B-Q2 | SINGLE | 1000 | 100 | Q2 | 10 | 1 | 10 |

### AWS (Broker + Pub-Sub)
**Test Config:** [tests.json](test-cases/00%20Single/2.%20AWS%20(Broker%20+%20Pub-Sub)/tests.json)

| Code | Category | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
|------|----------|-------|---------------|-----|-------------------|---------------------------|---------------------|
| AWS-HS-A-Q0 | SINGLE | 1000 | 10 | Q0 | 1 | 1 | 10 |
| AWS-HS-A-Q1 | SINGLE | 1000 | 10 | Q1 | 1 | 1 | 10 |
| AWS-HS-A-Q2 | SINGLE | 1000 | 10 | Q2 | 1 | 1 | 10 |
| AWS-HS-B-Q0 | SINGLE | 1000 | 100 | Q0 | 10 | 1 | 10 |
| AWS-HS-B-Q1 | SINGLE | 1000 | 100 | Q1 | 10 | 1 | 10 |
| AWS-HS-B-Q2 | SINGLE | 1000 | 100 | Q2 | 10 | 1 | 10 |

### AWS (Broker) + Local (Pub-Sub)
**Test Config:** [tests.json](test-cases/00%20Single/3.%20AWS%20(Broker)%20+%20Local%20(Pub-Sub)/tests.json)

| Code | Category | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
|------|----------|-------|---------------|-----|-------------------|---------------------------|---------------------|
| AWS-LC-HS-A-Q0 | SINGLE | 1000 | 10 | Q0 | 1 | 1 | 10 |
| AWS-LC-HS-A-Q1 | SINGLE | 1000 | 10 | Q1 | 1 | 1 | 10 |
| AWS-LC-HS-A-Q2 | SINGLE | 1000 | 10 | Q2 | 1 | 1 | 10 |
| AWS-LC-HS-B-Q0 | SINGLE | 1000 | 100 | Q0 | 10 | 1 | 10 |
| AWS-LC-HS-B-Q1 | SINGLE | 1000 | 100 | Q1 | 10 | 1 | 10 |
| AWS-LC-HS-B-Q2 | SINGLE | 1000 | 100 | Q2 | 10 | 1 | 10 |

---

## 01 Horizontal

### Local Device (Broker + Pub-Sub)
**Test Config:** [tests.json](test-cases/01%20Horizontal/1.%20Local%20Device%20(Broker%20+%20Pub-Sub)/tests.json)

| Code | Category | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
|------|----------|-------|---------------|-----|-------------------|---------------------------|---------------------|
| LOCAL-HM-A-Q0 | HORIZONTAL | 1000 | 10 | Q0 | 1 | 1 | 10 |
| LOCAL-HM-A-Q1 | HORIZONTAL | 1000 | 10 | Q1 | 1 | 1 | 10 |
| LOCAL-HM-A-Q2 | HORIZONTAL | 1000 | 10 | Q2 | 1 | 1 | 10 |
| LOCAL-HM-B-Q0 | HORIZONTAL | 1000 | 100 | Q0 | 10 | 1 | 10 |
| LOCAL-HM-B-Q1 | HORIZONTAL | 1000 | 100 | Q1 | 10 | 1 | 10 |
| LOCAL-HM-B-Q2 | HORIZONTAL | 1000 | 100 | Q2 | 10 | 1 | 10 |

### AWS (Broker + Pub-Sub)
**Test Config:** [tests.json](test-cases/01%20Horizontal/2.%20AWS%20(Broker%20+%20Pub-Sub)/tests.json)

| Code | Category | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
|------|----------|-------|---------------|-----|-------------------|---------------------------|---------------------|
| AWS-HM-A-Q0 | HORIZONTAL | 1000 | 10 | Q0 | 1 | 1 | 10 |
| AWS-HM-A-Q1 | HORIZONTAL | 1000 | 10 | Q1 | 1 | 1 | 10 |
| AWS-HM-A-Q2 | HORIZONTAL | 1000 | 10 | Q2 | 1 | 1 | 10 |
| AWS-HM-B-Q0 | HORIZONTAL | 1000 | 100 | Q0 | 10 | 1 | 10 |
| AWS-HM-B-Q1 | HORIZONTAL | 1000 | 100 | Q1 | 10 | 1 | 10 |
| AWS-HM-B-Q2 | HORIZONTAL | 1000 | 100 | Q2 | 10 | 1 | 10 |

### AWS (Broker) + Local (Pub-Sub)
**Test Config:** [tests.json](test-cases/01%20Horizontal/3.%20AWS%20(Broker)%20+%20Local%20(Pub-Sub)/tests.json)

| Code | Category | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
|------|----------|-------|---------------|-----|-------------------|---------------------------|---------------------|
| AWS-LC-HM-A-Q0 | HORIZONTAL | 1000 | 10 | Q0 | 1 | 1 | 10 |
| AWS-LC-HM-A-Q1 | HORIZONTAL | 1000 | 10 | Q1 | 1 | 1 | 10 |
| AWS-LC-HM-A-Q2 | HORIZONTAL | 1000 | 10 | Q2 | 1 | 1 | 10 |
| AWS-LC-HM-B-Q0 | HORIZONTAL | 1000 | 100 | Q0 | 10 | 1 | 10 |
| AWS-LC-HM-B-Q1 | HORIZONTAL | 1000 | 100 | Q1 | 10 | 1 | 10 |
| AWS-LC-HM-B-Q2 | HORIZONTAL | 1000 | 100 | Q2 | 10 | 1 | 10 |

---

## 02 Vertical

### Local Device (Broker + Pub-Sub)
**Test Config:** [tests.json](test-cases/02%20Vertical/1.%20Local%20Device%20(Broker%20+%20Pub-Sub)/tests.json)

| Code | Category | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
|------|----------|-------|---------------|-----|-------------------|---------------------------|---------------------|
| LOCAL-VM-A-Q0 | VERTICAL | 1000 | 10 | Q0 | 1 | 1 | 10 |
| LOCAL-VM-A-Q1 | VERTICAL | 1000 | 10 | Q1 | 1 | 1 | 10 |
| LOCAL-VM-A-Q2 | VERTICAL | 1000 | 10 | Q2 | 1 | 1 | 10 |
| LOCAL-VM-B-Q0 | VERTICAL | 1000 | 100 | Q0 | 10 | 1 | 10 |
| LOCAL-VM-B-Q1 | VERTICAL | 1000 | 100 | Q1 | 10 | 1 | 10 |
| LOCAL-VM-B-Q2 | VERTICAL | 1000 | 100 | Q2 | 10 | 1 | 10 |

### AWS (Broker + Pub-Sub)
**Test Config:** [tests.json](test-cases/02%20Vertical/2.%20AWS%20(Broker%20+%20Pub-Sub)/tests.json)

| Code | Category | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
|------|----------|-------|---------------|-----|-------------------|---------------------------|---------------------|
| AWS-VM-A-Q0 | VERTICAL | 1000 | 10 | Q0 | 1 | 1 | 10 |
| AWS-VM-A-Q1 | VERTICAL | 1000 | 10 | Q1 | 1 | 1 | 10 |
| AWS-VM-A-Q2 | VERTICAL | 1000 | 10 | Q2 | 1 | 1 | 10 |
| AWS-VM-B-Q0 | VERTICAL | 1000 | 100 | Q0 | 10 | 1 | 10 |
| AWS-VM-B-Q1 | VERTICAL | 1000 | 100 | Q1 | 10 | 1 | 10 |
| AWS-VM-B-Q2 | VERTICAL | 1000 | 100 | Q2 | 10 | 1 | 10 |

### AWS (Broker) + Local (Pub-Sub)
**Test Config:** [tests.json](test-cases/02%20Vertical/3.%20AWS%20(Broker)%20+%20Local%20(Pub-Sub)/tests.json)

| Code | Category | Users | Ramp up / sec | QoS | Message Size (KB) | Publish Rate (msg/sec/user) | Test Duration (mins) |
|------|----------|-------|---------------|-----|-------------------|---------------------------|---------------------|
| AWS-LC-VM-A-Q0 | VERTICAL | 1000 | 10 | Q0 | 1 | 1 | 10 |
| AWS-LC-VM-A-Q1 | VERTICAL | 1000 | 10 | Q1 | 1 | 1 | 10 |
| AWS-LC-VM-A-Q2 | VERTICAL | 1000 | 10 | Q2 | 1 | 1 | 10 |
| AWS-LC-VM-B-Q0 | VERTICAL | 1000 | 100 | Q0 | 10 | 1 | 10 |
| AWS-LC-VM-B-Q1 | VERTICAL | 1000 | 100 | Q1 | 10 | 1 | 10 |
| AWS-LC-VM-B-Q2 | VERTICAL | 1000 | 100 | Q2 | 10 | 1 | 10 |
