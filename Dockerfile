FROM golang:1.21-alpine

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Create config directory
RUN mkdir -p /app/config

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary and config
COPY --from=0 /app/bin/light-client .
COPY --from=0 /app/config ./config

# Create data and log directories
RUN mkdir -p /app/data/headers \
    && mkdir -p /app/data/checkpoints \
    && mkdir -p /app/logs

# Set environment variables
ENV CONFIG_FILE=/app/config/config.yaml
ENV LOG_FILE=/app/logs/light-client.log

# Expose metrics port
EXPOSE 9090

# Run the application
CMD ["./light-client", "--config", "/app/config/config.yaml"] 