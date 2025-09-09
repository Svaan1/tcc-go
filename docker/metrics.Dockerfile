# Multi-stage build for optimized Go benchmark client
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates for dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the benchmark binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o benchmark \
    ./cmd/benchmark/main.go

# Final stage: runtime image with AMD SMI
FROM ubuntu:22.04

# Install AMD SMI and required dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    wget \
    ca-certificates \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install AMD SMI
RUN wget -q https://repo.radeon.com/amdgpu-install/22.40.5/ubuntu/jammy/amdgpu-install_5.4.50405-1_all.deb && \
    apt-get update && \
    apt-get install -y ./amdgpu-install_5.4.50405-1_all.deb && \
    amdgpu-install -y --usecase=dkms,graphics,multimedia,opencl,hip,hiplibsdk,mlsdk,mllib --no-dkms && \
    rm amdgpu-install_5.4.50405-1_all.deb && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Alternative: Install ROCm for AMD SMI (comment above and uncomment below if preferred)
# RUN wget -q -O - https://repo.radeon.com/rocm/rocm.gpg.key | apt-key add - && \
#     echo 'deb [arch=amd64] https://repo.radeon.com/rocm/apt/debian/ ubuntu main' | tee /etc/apt/sources.list.d/rocm.list && \
#     apt-get update && \
#     apt-get install -y rocm-smi-lib && \
#     apt-get clean && \
#     rm -rf /var/lib/apt/lists/*

# Copy timezone data and certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the benchmark binary
COPY --from=builder /build/benchmark /benchmark

# Make sure amd-smi is available in PATH
ENV PATH="/opt/rocm/bin:${PATH}"

# Use the benchmark binary as entrypoint
ENTRYPOINT ["/benchmark"]