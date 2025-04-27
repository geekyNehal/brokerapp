# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o brokerapp ./cmd/brokerapp

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary tools
RUN apk add --no-cache mysql-client

# Copy the binary from builder
COPY --from=builder /app/brokerapp .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Copy setup scripts
COPY --from=builder /app/setup_sample_data.sh .
COPY --from=builder /app/check_db.sh .

# Make scripts executable
RUN chmod +x setup_sample_data.sh check_db.sh

# Expose port
EXPOSE 8080

# Run the application
CMD ["./brokerapp"] 