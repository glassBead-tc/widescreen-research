# Operations Guide

This document provides operational guidance for running and monitoring the widescreen-research coordinator and drones.

## HTTP Endpoints

### Coordinator
- `GET /` - Basic service information and status
- `GET /health` - Health/readiness check endpoint
- `POST /api/drones/register` - Drone registration endpoint
- (Future) `GET /ready` - Readiness probe (separate from liveness)
- (Future) `GET /live` - Liveness probe

### Drone
- `GET /health` - Health/readiness check endpoint  
- `POST /mcp` - MCP protocol endpoint (HTTP transport mode)

## Logging

### Log Levels
Control via `LOG_LEVEL` environment variable:
- `debug` - Verbose debugging information
- `info` - General operational information (default)
- `warn` - Warning conditions
- `error` - Error conditions only

### Structured Logging
Both coordinator and drones use structured JSON logging with fields:
- `timestamp` - RFC3339 timestamp
- `level` - Log level
- `msg` - Human-readable message
- `component` - Service component (coordinator/drone)
- `drone_id` - Drone identifier (for drone logs)
- `request_id` - Request correlation ID
- `error` - Error details (when applicable)

### Log Destinations
- **Local Development**: stdout/stderr
- **Cloud Run**: Google Cloud Logging (automatic)
- **Container**: stdout (captured by orchestrator)

## Monitoring and Observability

### Health Checks
```bash
# Coordinator health
curl -f http://localhost:8080/health

# Cloud Run health (via external URL)
curl -f https://widescreen-coordinator-xxxx.run.app/health
```

### Key Metrics to Monitor
Currently manual monitoring. Future OpenTelemetry integration will provide:

#### Coordinator Metrics
- Active drone count
- Drone registration rate
- Failed drone operations
- HTTP request latency and count
- Memory usage and CPU usage

#### Drone Metrics  
- MCP request processing time
- Research API call success/failure rates
- Cache hit/miss rates (if caching added)
- Memory usage per drone

### Tracing (Planned)
OpenTelemetry integration planned for:
- Distributed tracing across coordinator-drone communications
- Request flow from MCP client → coordinator → drone → external APIs
- Performance bottleneck identification
- Error propagation tracking

## Environment Configuration

### Required Environment Variables
- `GOOGLE_CLOUD_PROJECT` - GCP project ID (coordinator)
- `EXA_API_KEY` - Exa AI API key (research drones)

### Optional Environment Variables
- `LOG_LEVEL` - Logging verbosity (default: info)
- `PORT` - HTTP server port (default: 8080)
- `GCP_REGION` - GCP region (default: us-central1)
- `COORDINATOR_BASE_URL` - Coordinator callback URL (local dev)
- `DRONE_TYPE` - Drone capability type (research, scraper, etc.)

## Cloud Run Deployment

### Resource Allocation
- **CPU**: 1 vCPU (coordinator), 0.5 vCPU (drone)
- **Memory**: 512Mi (coordinator), 256Mi (drone)
- **Concurrency**: 100 (coordinator), 10 (drone)
- **Min instances**: 0 (cost optimization)
- **Max instances**: 10 (coordinator), 100 (drones)

### Auto-scaling Behavior
- Cold start: ~1-2 seconds (Go binaries)
- Scale-to-zero after 15 minutes of no traffic
- Scale-up based on concurrent requests and CPU/memory usage

### Secret Management
Secrets injected via Cloud Run environment from Secret Manager:
```bash
gcloud run deploy widescreen-coordinator \
  --set-secrets=EXA_API_KEY=exa_api_key:latest
```

## Troubleshooting

### Common Issues

#### Coordinator won't start
1. Check `GOOGLE_CLOUD_PROJECT` is set
2. Verify GCP authentication: `gcloud auth list`
3. Ensure required APIs are enabled
4. Check logs for specific error messages

#### Drone registration fails
1. Verify coordinator URL is reachable
2. Check network connectivity between services
3. Examine coordinator logs for registration errors
4. Confirm drone has correct `COORDINATOR_BASE_URL`

#### External API failures
1. Check `EXA_API_KEY` is correctly set and valid
2. Verify API rate limits aren't exceeded
3. Monitor external service status pages
4. Review structured logs for API error responses

### Log Analysis
```bash
# Filter coordinator logs by level
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=widescreen-coordinator" --limit=50 --format="value(textPayload)"

# Search for specific errors
gcloud logging read "resource.type=cloud_run_revision AND textPayload:ERROR" --limit=20

# Follow real-time logs
gcloud logging tail "resource.type=cloud_run_revision AND resource.labels.service_name=widescreen-coordinator"
```

## Performance Optimization

### Current State
- Minimal resource usage with distroless containers
- ~2MB container images for fast cold starts
- Go binaries with optimized build flags

### Future Improvements
- Connection pooling for external APIs
- Request caching (Redis/Memorystore)
- Metric-based auto-scaling
- Warm pool of drones for reduced latency

## Security Considerations

### Current Security Posture
- Non-root container execution
- Secrets via Secret Manager (not environment variables)
- Minimal container attack surface (distroless base)
- Cloud Run IAM-based access control

### Recommended Enhancements
- Implement authentication between coordinator and drones
- Network security policies (VPC, firewall rules)
- Regular security scanning of dependencies
- Audit logging for administrative operations