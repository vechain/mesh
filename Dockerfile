# Build stage
FROM golang:1.24-bullseye AS builder

# Set working directory
WORKDIR /app

# Install git, ca-certificates and build dependencies
RUN apt-get update && apt-get install -y \
    git \
    ca-certificates \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o mesh-server .

# Final stage
FROM ubuntu:24.04

# Install ca-certificates, wget for health checks, git for networkhub, make, build tools for Thor compilation, and Go
RUN apt-get update && apt-get install -y \
    ca-certificates \
    wget \
    git \
    make \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -s /bin/bash mesh

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/mesh-server .

# Copy Go from builder stage for networkhub to compile Thor
COPY --from=builder /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

# Copy configuration files
COPY --from=builder /app/config ./config

# Change ownership to mesh user
RUN chown -R mesh:mesh /app

# Create and set permissions for Thor cache directory
RUN mkdir -p /tmp/thor_master_reusable && chown -R mesh:mesh /tmp/thor_master_reusable

# Switch to non-root user
USER mesh

# Expose ports
EXPOSE 8080 8669 11235 11235/udp

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./mesh-server"]
