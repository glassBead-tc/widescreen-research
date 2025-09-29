# Build stage
FROM golang:1.23 as build

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the coordinator binary
# Embed version info if desired: pass --build-arg VERSION
ARG VERSION=dev
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w -X main.version=${VERSION}" -o /out/coordinator ./cmd/coordinator

# Minimal runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary from build stage
COPY --from=build /out/coordinator /coordinator

# Set default environment
ENV PORT=8080

# Expose the port
EXPOSE 8080

# Use non-root user for security
USER nonroot:nonroot

# Run the coordinator
ENTRYPOINT ["/coordinator"]
