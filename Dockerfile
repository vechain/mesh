# Thor builder stage
FROM golang:1.24 AS thor-builder

ARG THOR_REPO=https://github.com/vechain/thor.git
ARG THOR_VERSION=v2.3.1

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    git \
    make \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Clone and build Thor
RUN git clone ${THOR_REPO} thor && \
    cd thor && \
    git checkout ${THOR_VERSION} && \
    make all

# Mesh builder stage
FROM golang:1.24-bullseye AS mesh-builder

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

# Install ca-certificates, wget for health checks, git, make, and build tools
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

# Copy Thor binary from thor-builder stage
COPY --from=thor-builder /app/thor/bin/thor ./thor

# Make Thor binary executable
RUN chmod +x ./thor

# Create and set permissions for Thor data directory
RUN mkdir -p /tmp/thor_data && chown -R mesh:mesh /tmp/thor_data

# Copy the binary from builder stage
COPY --from=mesh-builder /app/mesh-server .

# Copy configuration files
COPY --from=mesh-builder /app/config ./config

# Change ownership to mesh user
RUN chown -R mesh:mesh /app

# Switch to non-root user
USER mesh

# Expose ports
EXPOSE 8080 8669 11235 11235/udp

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./mesh-server"]
