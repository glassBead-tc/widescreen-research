# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Overview

This repository implements a coordinator-worker Model Context Protocol (MCP) architecture for distributed AI orchestration on Google Cloud Platform.

**Core Components:**

- **Go Control Plane** (`cmd/coordinator`, `pkg/`): Orchestrates spawning and coordinating MCP "drone" workers on Cloud Run
- **Node.js Drone Servers** (`drone-mcp-template/`): Research-focused MCP servers with stdio/HTTP transport
- **Simple Local MCP** (`cmd/simple-mcp`): Minimal Go stdio MCP server for testing without GCP

## Architecture

### Coordinator (Go)

- **Entry Point**: `cmd/coordinator/main.go`
- **Requirements**: `GOOGLE_CLOUD_PROJECT` env var (required), `GOOGLE_CLOUD_REGION` (optional, defaults to us-central1)
- **Core Services**:
  - GCP client (`pkg/gcp/`) for Cloud Run management
  - Coordinator server (`pkg/coordinator/`) for drone orchestration
  - MCP wrapper (`pkg/mcp/server.go`) exposes tools via mark3labs/mcp-go

**Available MCP Tools** (defined in `pkg/mcp/server.go`):

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

### Drone MCP Servers (Node.js)

- **Entry Point**: `drone-mcp-template/index.js`
- **Transport Modes**:
  - `MCP_TRANSPORT=stdio`: Desktop MCP clients (Claude, etc.)
  - `MCP_TRANSPORT=http`: Cloud Run deployment (MCP at `/mcp` endpoint)
- **Self-Registration**: If `COORDINATOR_URL` set, posts to `{url}/api/drones/register`
- **Capabilities by Type** (`getDroneCapabilities()`):
  - `generic`: echo, ping
  - `research`: web_search, research_papers, company_research, crawl_url, find_competitors, linkedin_search, wikipedia_search, github_search
  - `scraper`: fetch_url, extract_data
  - `processor`: transform_data, validate_data
  - `analyzer`: analyze_text, sentiment_analysis

## Build & Run Commands

### Go Components

```bash
# Build coordinator binary
go build -o bin/coordinator ./cmd/coordinator

# Run coordinator (requires GCP project)
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_CLOUD_REGION=us-central1  # optional
go run ./cmd/coordinator

# Run simple stdio MCP server (no GCP needed)
go run ./cmd/simple-mcp

# Run tests
go test ./...

# Lint code
go vet ./...
```

### Node.js Drone

```bash
# Install dependencies
cd drone-mcp-template
npm install

# Run with HTTP transport (Cloud Run mode)
PORT=8080 MCP_TRANSPORT=http DRONE_TYPE=research node index.js

# Run with stdio transport (desktop MCP client mode)
MCP_TRANSPORT=stdio DRONE_TYPE=research node index.js

# With coordinator registration
COORDINATOR_URL=http://localhost:8080 PORT=8081 MCP_TRANSPORT=http DRONE_TYPE=research node index.js

# Run tests
npm test

# Lint
npm run lint
```

### Docker Images

```bash
# Build coordinator image
docker build -f cmd/coordinator/Dockerfile -t coordinator .

# Build drone image
docker build -f drone-mcp-template/Dockerfile -t drone-mcp .

# Build specific drone type
cd cmd/drone
docker build -t drone-researcher .
```

## Environment Variables

### Coordinator (Go)

- `GOOGLE_CLOUD_PROJECT` - GCP project ID (required)
- `GOOGLE_CLOUD_REGION` - Deployment region (default: us-central1)
- `PORT` - HTTP server port (default: 8080)

### Drone (Node.js)

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
cd drone-mcp-template
PORT=8080 MCP_TRANSPORT=http DRONE_TYPE=research node index.js

# Terminal 3: Test with MCP inspector
npx @modelcontextprotocol/inspector node drone-mcp-template/index.js
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
cd drone-mcp-template
COORDINATOR_URL=http://localhost:8080 PORT=8081 MCP_TRANSPORT=http DRONE_TYPE=research node index.js
```

## Troubleshooting & Tips

### Common Issues

1. **Coordinator not exposing MCP tools via stdio**
   - `cmd/coordinator` runs HTTP services only by default
   - Use `cmd/simple-mcp` for stdio MCP testing
   - To add stdio: wire `pkg/mcp.NewMCPServer()` into coordinator main

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
MCP_TRANSPORT=stdio node drone-mcp-template/index.js 2>&1 | jq .
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
cd drone-mcp-template && PORT=8080 MCP_TRANSPORT=http DRONE_TYPE=research node index.js

# Drone stdio mode
cd drone-mcp-template && MCP_TRANSPORT=stdio DRONE_TYPE=research node index.js

# Build Go binary
go build -o bin/coordinator ./cmd/coordinator

# Test everything
go test ./... && cd drone-mcp-template && npm test
```
