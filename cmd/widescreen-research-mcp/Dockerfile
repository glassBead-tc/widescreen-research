# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o widescreen-research ./cmd/widescreen-research-mcp

# Final stage
FROM gcr.io/distroless/static:nonroot

# Copy binary from builder
COPY --from=builder /app/widescreen-research /widescreen-research

# Use non-root user
USER nonroot:nonroot

# Expose port (if needed for health checks)
EXPOSE 8080

# Run the application
ENTRYPOINT ["/widescreen-research"]