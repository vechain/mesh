FROM golang:1.24-bullseye AS builder

WORKDIR /app

# Clone and build mesh-cli
ARG MESH_CLI_VERSION=v0.10.4
RUN git clone https://github.com/coinbase/mesh-cli.git . && \
    git checkout ${MESH_CLI_VERSION} && \
    go mod download && \
    go build -o mesh-cli .

# Final stage
FROM ubuntu:24.04

# Install ca-certificates and jq for working with JSON
RUN apt-get update && apt-get install -y \
    ca-certificates \
    jq \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -s /bin/bash mesh

# Create data directory for mesh-cli
RUN mkdir -p /data

WORKDIR /app

# Copy mesh-cli binary from builder
COPY --from=builder /app/mesh-cli .

# Make it executable
RUN chmod +x mesh-cli

# Change ownership to mesh user
RUN chown -R mesh:mesh /app

# Switch to non-root user
USER mesh

# Set data directory as volume
VOLUME ["/data"]

# Default command shows help
ENTRYPOINT ["./mesh-cli"]
CMD ["--help"]

