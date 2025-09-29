# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Overview

This repository implements a Go-first coordinator-worker Model Context Protocol (MCP) architecture for distributed AI orchestration on Google Cloud Platform. Node.js implementations are provided as examples only.

**Core Components:**

- **Coordinator** (`cmd/coordinator`): HTTP server orchestrating drone workers on Cloud Run
- **Drone** (`cmd/drone`): Go worker executing research/analysis tasks
- **MCP Coordinator** (`cmd/mcp-coordinator`): stdio MCP server exposing coordinator tools
- **Simple MCP** (`cmd/simple-mcp`): Standalone stdio MCP server for testing
- **Widescreen Research MCP** (`cmd/widescreen-research-mcp`): Specialized research MCP server
- **Node.js Examples** (`examples/node/`): Reference implementations for MCP interoperability

## Architecture

### Coordinator (`cmd/coordinator`)

HTTP server that orchestrates drone workers on GCP Cloud Run.

- **Entry Point**: `cmd/coordinator/main.go`
- **Transport**: HTTP only (port 8080)
- **Requirements**: `GOOGLE_CLOUD_PROJECT` (required), `GOOGLE_CLOUD_REGION` (optional, defaults to us-central1)
- **Core Services**:
  - GCP client (`pkg/gcp/`) for Cloud Run management
  - Coordinator server (`pkg/coordinator/`) for drone orchestration and campaign management
  - HTTP endpoints for health checks and API operations

### MCP Coordinator (`cmd/mcp-coordinator`)

stdio MCP server that wraps coordinator functionality for MCP clients.

- **Entry Point**: `cmd/mcp-coordinator/main.go`
- **Transport**: stdio only
- **MCP Tools** (via `pkg/mcp/server.go`):
  - `spawn_drone_server(drone_type, region)` - Launch new drone on Cloud Run
  - `list_active_drones()` - Show active drone fleet
  - `execute_distributed_task(task_type, description, max_drones)` - Distribute work
  - `get_drone_status(drone_id)` - Query specific drone
  - `terminate_drone(drone_id)` - Shutdown drone
  - `plan_campaign(spec_json)` - Campaign orchestration
  - `launch_fleet(run_id, target_workers)` - Fleet provisioning
  - `fleet_status(run_id)` - Monitor campaign progress
  - `abort(run_id)` - Cancel campaign run
  - `export_graph(mem0_space, format)` - Export collected data

### Drone (`cmd/drone`)

Go worker that executes research and analysis tasks.

- **Entry Point**: `cmd/drone/main.go`
- **Transport**: HTTP (port 8080)
- **Implementation**: `pkg/drone/researcher.go`, `pkg/drone/http_worker.go`
- **Capabilities**: Research operations, data processing

### Simple MCP (`cmd/simple-mcp`)

Standalone stdio MCP server for local testing without GCP dependencies.

- **Entry Point**: `cmd/simple-mcp/main.go`
- **Transport**: stdio only
- **Purpose**: Mock coordinator for development and testing
- **Tools**: Similar to MCP Coordinator but with in-memory state

### Node.js Examples (`examples/node/`)

Reference MCP implementations demonstrating interoperability:

- **drone-mcp-template**: Full-featured example with stdio/HTTP transport
- **exa-mcp-server**: Research-focused MCP server using Exa API
- **spawn-mcp**: Minimal spawning example

**Why Node.js examples exist:**
- Demonstrate MCP protocol interoperability across runtimes
- Provide quick prototyping templates
- Aid in testing MCP client tooling (many tools are Node-centric)
- **Note**: These are examples only, not part of the production control plane

## Build & Run Commands

### Go Components

```bash
# Build all Go binaries
make build

# Build individual binaries
go build -o bin/coordinator ./cmd/coordinator
go build -o bin/drone ./cmd/drone
go build -o bin/mcp-coordinator ./cmd/mcp-coordinator
go build -o bin/simple-mcp ./cmd/simple-mcp
go build -o bin/widescreen-research-mcp ./cmd/widescreen-research-mcp

# Run coordinator (requires GCP project)
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_CLOUD_REGION=us-central1  # optional
go run ./cmd/coordinator
# or with make
make run-coordinator

# Run drone worker
LOG_LEVEL=debug DRONE_TYPE=research EXA_API_KEY=dummy go run ./cmd/drone
# or with make
make run-drone

# Run MCP coordinator (stdio MCP server)
go run ./cmd/mcp-coordinator

# Run simple MCP (no GCP needed)
go run ./cmd/simple-mcp

# Run widescreen research MCP
go run ./cmd/widescreen-research-mcp
```

### Testing & Quality

```bash
# Run tests with race detection and coverage
make test
# or directly
go test -race -cover ./...

# Run linters
make lint-go
# or directly
golangci-lint run --fix

# Run all quality checks before pushing
make pre-push

# Install git hooks
make install-hooks

# Security scanning
make security-scan
```

### Node.js Examples

```bash
# Navigate to example
cd examples/node/drone-mcp-template

# Install dependencies
npm install

# Run with HTTP transport (Cloud Run mode)
PORT=8080 MCP_TRANSPORT=http DRONE_TYPE=research node index.js

# Run with stdio transport (desktop MCP client mode)
MCP_TRANSPORT=stdio DRONE_TYPE=research node index.js

# With coordinator registration
COORDINATOR_URL=http://localhost:8080 PORT=8081 MCP_TRANSPORT=http DRONE_TYPE=research node index.js

# Run tests
npm test
```

### Docker

```bash
# Build coordinator image (from repo root)
docker build -t widescreen/coordinator:dev .
# or with make
make docker

# Build drone image
docker build -f cmd/drone/Dockerfile -t widescreen/drone:dev .

# Build widescreen-research-mcp image
docker build -f cmd/widescreen-research-mcp/Dockerfile -t widescreen/research-mcp:dev .

# Build Node.js example
cd examples/node/drone-mcp-template
docker build -t drone-mcp-example .

# Run coordinator container
docker run --rm -p 8080:8080 \
  -e GOOGLE_CLOUD_PROJECT=your-project-id \
  widescreen/coordinator:dev
```

### Makefile Targets

```bash
make help              # Show all available targets
make build             # Build Go binaries
make run-coordinator   # Run coordinator locally
make run-drone        # Run drone locally
make test             # Run tests with race detection
make docker           # Build Docker image
make install-hooks   # Install Git hooks
make lint            # Run all linters
make lint-go        # Run Go linters
make lint-docker    # Run Docker linters
make security-scan   # Run security scans
make pre-push        # Run all checks before pushing
make clean-hooks     # Remove Git hooks
```

## Environment Variables

### Coordinator (Go)

- `GOOGLE_CLOUD_PROJECT` - GCP project ID (required)
- `GOOGLE_CLOUD_REGION` - Deployment region (default: us-central1)
- `PORT` - HTTP server port (default: 8080)
- `LOG_LEVEL` - Log verbosity (debug, info, warn, error)

### Drone (Go)

- `PORT` - HTTP server port (default: 8080)
- `LOG_LEVEL` - Log verbosity (debug, info, warn, error)
- `DRONE_TYPE` - Drone type for capabilities (research, analyst, etc.)
- `EXA_API_KEY` - Exa API key for research operations

### MCP Coordinator (Go)

- `GOOGLE_CLOUD_PROJECT` - GCP project ID (optional, skips GCP if not set)
- `GOOGLE_CLOUD_REGION` - Deployment region (default: us-central1)

### Node.js Examples

- `PORT` - HTTP server port (default: 8080)
- `DRONE_TYPE` - Drone capabilities (research, scraper, processor, analyzer, generic)
- `MCP_TRANSPORT` - Transport type (stdio or http)
- `COORDINATOR_URL` - Coordinator endpoint for registration (optional)
- `GOOGLE_CLOUD_PROJECT` - GCP project (optional, auto-detected on GCP)
- `EXA_API_KEY` - Required for research drone type
- `K_SERVICE`, `K_REVISION` - Cloud Run metadata (auto-set)

## Development Workflows

### Local Testing Without GCP

```bash
# Terminal 1: Run simple MCP server
go run ./cmd/simple-mcp

# Terminal 2: Run drone with HTTP transport
cd examples/node/drone-mcp-template
PORT=8080 MCP_TRANSPORT=http DRONE_TYPE=research node index.js

# Terminal 3: Test with MCP inspector
npx @modelcontextprotocol/inspector node examples/node/drone-mcp-template/index.js
```

### GCP Development

```bash
# Authenticate with GCP
gcloud auth application-default login
gcloud config set project $GOOGLE_CLOUD_PROJECT

# Terminal 1: Run coordinator
export GOOGLE_CLOUD_PROJECT=your-project-id
go run ./cmd/coordinator

# Terminal 2: Run drone with coordinator registration
cd examples/node/drone-mcp-template
COORDINATOR_URL=http://localhost:8080 PORT=8081 MCP_TRANSPORT=http DRONE_TYPE=research node index.js
```

## Troubleshooting & Tips

### Common Issues

1. **Coordinator not exposing MCP tools via stdio**
   - `cmd/coordinator` runs HTTP services only by default
   - Use `cmd/mcp-coordinator` for stdio MCP testing
   - Use `cmd/simple-mcp` for standalone testing without GCP

2. **Drone registration failing**
   - Check `COORDINATOR_URL` is accessible
   - For local dev: coordinator may need to accept unauthenticated requests
   - GCP metadata authentication works only on GCP infrastructure

3. **Transport mode confusion**
   - `stdio`: For desktop MCP clients (Claude Desktop, CLI tools)
   - `http`: For Cloud Run deployment or HTTP-based integration
   - MCP endpoint with HTTP: `http://localhost:{PORT}/mcp`

4. **Missing capabilities**
   - Capabilities depend on `DRONE_TYPE` (see `getDroneCapabilities()`)
   - Research drone needs `EXA_API_KEY` for full functionality

### Performance Notes

- Pre-compiled binaries have faster startup than `go run`
- Node drones have 100ms-1s cold start on optimized Cloud Run containers
- Use distroless or Alpine base images for smaller containers (~2-20MB)

### Monitoring & Debugging

```bash
# Check drone health (HTTP transport)
curl http://localhost:8080/health

# View coordinator logs
go run ./cmd/coordinator 2>&1 | grep -E "(Spawning|Terminating|Executing)"

# Debug MCP communication
MCP_TRANSPORT=stdio node examples/node/drone-mcp-template/index.js 2>&1 | jq .
```

## Quick Reference

```bash
# GCP auth
gcloud auth application-default login

# Coordinator with GCP
export GOOGLE_CLOUD_PROJECT=my-project && go run ./cmd/coordinator

# Simple MCP (no GCP)
go run ./cmd/simple-mcp

# Drone HTTP mode
cd examples/node/drone-mcp-template && PORT=8080 MCP_TRANSPORT=http DRONE_TYPE=research node index.js

# Drone stdio mode
cd examples/node/drone-mcp-template && MCP_TRANSPORT=stdio DRONE_TYPE=research node index.js

# Build Go binary
go build -o bin/coordinator ./cmd/coordinator

# Test everything
go test ./... && cd examples/node/drone-mcp-template && npm test
```
