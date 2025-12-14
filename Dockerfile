FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /app

# Build + runtime deps for Mosquitto + plugin build + Go sidecar build
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates git build-essential cmake pkg-config \
    libssl-dev \
    libcjson-dev \
    uuid-dev \
    libhiredis-dev \
    # Go toolchain (Ubuntu 22.04 provides Go 1.18.x)
    golang-go \
    && rm -rf /var/lib/apt/lists/*

CMD ["bash"]
