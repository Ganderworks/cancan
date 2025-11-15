FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY main.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o canreplay .

# Runtime image
FROM alpine:latest

# Install SocketCAN utilities for testing/debugging
RUN apk add --no-cache iproute2 can-utils

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/canreplay .

# Copy CAN data
COPY can/ ./can/

# Setup script for vcan interface
RUN echo '#!/bin/sh' > /app/setup-vcan.sh && \
    echo 'ip link add dev vcan0 type vcan 2>/dev/null || true' >> /app/setup-vcan.sh && \
    echo 'ip link set up vcan0 2>/dev/null || true' >> /app/setup-vcan.sh && \
    echo 'ip link show vcan0' >> /app/setup-vcan.sh && \
    chmod +x /app/setup-vcan.sh

ENTRYPOINT ["/app/canreplay"]
CMD ["-csv", "can/rdu_onown_nomotor_on_thenoff_vehcan.csv", "-can", "vcan0"]